package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/runlet/runlet/apps/runner/internal/api"
	"github.com/runlet/runlet/apps/runner/internal/config"
	"github.com/runlet/runlet/apps/runner/internal/daemon"
	"github.com/runlet/runlet/apps/runner/internal/runner"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "setup" {
		if err := runSetup(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "setup failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	var once bool
	var nonInteractive bool
	var flagSeed config.Seed
	var labels labelFlags
	var configPath string

	flag.StringVar(&flagSeed.APIURL, "api-url", "", "Runlet Core API URL")
	flag.StringVar(&flagSeed.Token, "token", "", "runner authentication token")
	flag.StringVar(&flagSeed.RunnerID, "runner-id", "", "existing runner id")
	flag.StringVar(&flagSeed.Name, "name", "", "runner display name")
	flag.StringVar(&flagSeed.DefaultWorkspace, "workspace", "", "default execution workspace")
	flag.StringVar(&flagSeed.Shell, "shell", "", "shell used to execute commands")
	flag.StringVar(&configPath, "config", "", "runner configuration file")
	flag.IntVar(&flagSeed.PollIntervalSeconds, "poll-interval", 0, "poll interval in seconds")
	flag.IntVar(&flagSeed.HeartbeatIntervalSeconds, "heartbeat-interval", 0, "heartbeat interval in seconds")
	flag.IntVar(&flagSeed.DefaultTimeoutSeconds, "timeout", 0, "default run timeout in seconds")
	flag.Var(&labels, "label", "runner label as key=value; may be repeated")
	flag.BoolVar(&once, "once", false, "claim and execute at most one run")
	flag.BoolVar(&nonInteractive, "non-interactive", false, "read configuration only from flags and environment")
	flag.Parse()
	flagSeed.Labels = labels.Map()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	seed, err := runtimeSeed(configPath, flagSeed)
	if err != nil {
		logger.Error("failed to load runner configuration", "error", err)
		os.Exit(1)
	}
	cfg, err := loadConfig(seed, nonInteractive)
	if err != nil {
		logger.Error("failed to configure runner", "error", err)
		os.Exit(1)
	}

	client, err := api.NewClient(cfg.APIURL, cfg.Token)
	if err != nil {
		logger.Error("failed to create api client", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	service := runner.New(cfg, client, logger)
	if err := service.Start(ctx, runner.Options{Once: once}); err != nil {
		fmt.Fprintf(os.Stderr, "runner failed: %v\n", err)
		os.Exit(1)
	}
}

func runSetup(args []string) error {
	var flagSeed config.Seed
	var labels labelFlags
	var configPath string
	var foreground bool

	flags := flag.NewFlagSet("setup", flag.ContinueOnError)
	flags.StringVar(&flagSeed.APIURL, "api-url", "", "Runlet Core API URL")
	flags.StringVar(&flagSeed.Token, "token", "", "runner authentication token")
	flags.StringVar(&flagSeed.RunnerID, "runner-id", "", "existing runner id")
	flags.StringVar(&flagSeed.Name, "name", "", "runner display name")
	flags.StringVar(&flagSeed.DefaultWorkspace, "workspace", "", "default execution workspace")
	flags.StringVar(&flagSeed.Shell, "shell", "", "shell used to execute commands")
	flags.StringVar(&configPath, "config", "", "runner configuration file")
	flags.IntVar(&flagSeed.PollIntervalSeconds, "poll-interval", 0, "poll interval in seconds")
	flags.IntVar(&flagSeed.HeartbeatIntervalSeconds, "heartbeat-interval", 0, "heartbeat interval in seconds")
	flags.IntVar(&flagSeed.DefaultTimeoutSeconds, "timeout", 0, "default run timeout in seconds")
	flags.Var(&labels, "label", "runner label as key=value; may be repeated")
	flags.BoolVar(&foreground, "foreground", false, "save configuration without starting the background runner")
	positionalToken := ""
	parseArgs := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		positionalToken = args[0]
		parseArgs = args[1:]
	}
	if err := flags.Parse(parseArgs); err != nil {
		return err
	}
	if positionalToken != "" && flagSeed.Token == "" {
		flagSeed.Token = positionalToken
	} else if flags.NArg() > 0 && flagSeed.Token == "" {
		flagSeed.Token = flags.Arg(0)
	}
	flagSeed.Labels = labels.Map()

	seed := config.DefaultSeed().Merge(config.SeedFromEnv()).Merge(flagSeed)
	if flagSeed.RunnerID == "" {
		seed.RunnerID = ""
	}
	cfg, err := config.FromSeed(seed)
	if err != nil {
		return err
	}

	path, err := resolvedConfigPath(configPath)
	if err != nil {
		return err
	}
	if err := config.SaveSeed(path, config.Seed(cfg)); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Runlet runner configured at %s\n", path)
	if foreground {
		fmt.Fprintln(os.Stdout, "Start it with: runlet-runner")
		return nil
	}

	pidPath, err := config.DefaultPIDPath()
	if err != nil {
		return err
	}
	logPath, err := config.DefaultLogPath()
	if err != nil {
		return err
	}

	executable, err := os.Executable()
	if err != nil {
		return err
	}

	runnerArgs := []string{"-non-interactive"}
	if configPath != "" {
		runnerArgs = append(runnerArgs, "-config", configPath)
	}

	pid, err := daemon.Start(daemon.Options{
		Executable: executable,
		Args:       runnerArgs,
		PIDPath:    pidPath,
		LogPath:    logPath,
		WorkDir:    cfg.DefaultWorkspace,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Running in the background (pid %d). Logs: %s\n", pid, logPath)
	return nil
}

func runtimeSeed(configPath string, flagSeed config.Seed) (config.Seed, error) {
	path, err := resolvedConfigPath(configPath)
	if err != nil {
		return config.Seed{}, err
	}
	storedSeed, err := config.LoadSeed(path)
	if err != nil {
		return config.Seed{}, err
	}
	return config.DefaultSeed().Merge(storedSeed).Merge(config.SeedFromEnv()).Merge(flagSeed), nil
}

func resolvedConfigPath(path string) (string, error) {
	if path != "" {
		return path, nil
	}
	return config.DefaultPath()
}

func loadConfig(seed config.Seed, nonInteractive bool) (config.Config, error) {
	if nonInteractive || !isTerminal(os.Stdin) {
		return config.FromSeed(seed)
	}
	return config.Inquire(os.Stdin, os.Stdout, seed)
}

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&fs.ModeCharDevice != 0
}

type labelFlags []string

func (l *labelFlags) String() string {
	return config.FormatLabels(l.Map())
}

func (l *labelFlags) Set(value string) error {
	if _, _, ok := strings.Cut(value, "="); !ok {
		return fmt.Errorf("label must be key=value")
	}
	*l = append(*l, value)
	return nil
}

func (l labelFlags) Map() map[string]string {
	return config.ParseLabels(strings.Join(l, ","))
}
