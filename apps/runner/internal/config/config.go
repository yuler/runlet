package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	APIURL                   string
	Token                    string
	RunnerID                 string
	Name                     string
	Concurrency              int
	PollIntervalSeconds      int
	HeartbeatIntervalSeconds int
	DefaultTimeoutSeconds    int
	DefaultWorkspace         string
	Shell                    string
	Labels                   map[string]string
}

type Seed struct {
	APIURL                   string
	Token                    string
	RunnerID                 string
	Name                     string
	Concurrency              int
	PollIntervalSeconds      int
	HeartbeatIntervalSeconds int
	DefaultTimeoutSeconds    int
	DefaultWorkspace         string
	Shell                    string
	Labels                   map[string]string
}

func DefaultSeed() Seed {
	workspace, _ := os.Getwd()
	return Seed{
		APIURL:                   "http://localhost:3000",
		Token:                    "dev-token",
		Name:                     "local-runner",
		Concurrency:              1,
		PollIntervalSeconds:      5,
		HeartbeatIntervalSeconds: 15,
		DefaultTimeoutSeconds:    900,
		DefaultWorkspace:         workspace,
		Labels: map[string]string{
			"kind":    "desktop",
			"project": "runlet",
		},
	}
}

func FromSeed(seed Seed) (Config, error) {
	cfg := Config(seed)
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Inquire(reader io.Reader, writer io.Writer, seed Seed) (Config, error) {
	cfg := Config(seed)
	cfg.applyDefaults()

	prompt := newPrompt(reader, writer)
	var err error

	cfg.APIURL, err = prompt.ask("Core API URL", cfg.APIURL)
	if err != nil {
		return Config{}, err
	}
	cfg.Token, err = prompt.ask("Runner token", cfg.Token)
	if err != nil {
		return Config{}, err
	}
	cfg.RunnerID, err = prompt.ask("Runner ID", cfg.RunnerID)
	if err != nil {
		return Config{}, err
	}
	cfg.Name, err = prompt.ask("Runner name", cfg.Name)
	if err != nil {
		return Config{}, err
	}
	cfg.DefaultWorkspace, err = prompt.ask("Default workspace", cfg.DefaultWorkspace)
	if err != nil {
		return Config{}, err
	}
	cfg.Shell, err = prompt.ask("Shell", cfg.Shell)
	if err != nil {
		return Config{}, err
	}
	cfg.PollIntervalSeconds, err = prompt.askInt("Poll interval seconds", cfg.PollIntervalSeconds)
	if err != nil {
		return Config{}, err
	}
	cfg.HeartbeatIntervalSeconds, err = prompt.askInt("Heartbeat interval seconds", cfg.HeartbeatIntervalSeconds)
	if err != nil {
		return Config{}, err
	}
	cfg.DefaultTimeoutSeconds, err = prompt.askInt("Default timeout seconds", cfg.DefaultTimeoutSeconds)
	if err != nil {
		return Config{}, err
	}
	cfg.Labels, err = prompt.askLabels("Labels", cfg.Labels)
	if err != nil {
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func SeedFromEnv() Seed {
	return Seed{
		APIURL:                   os.Getenv("RUNLET_API_URL"),
		Token:                    os.Getenv("RUNLET_TOKEN"),
		RunnerID:                 os.Getenv("RUNLET_RUNNER_ID"),
		Name:                     os.Getenv("RUNLET_RUNNER_NAME"),
		Concurrency:              envInt("RUNLET_CONCURRENCY"),
		PollIntervalSeconds:      envInt("RUNLET_POLL_INTERVAL_SECONDS"),
		HeartbeatIntervalSeconds: envInt("RUNLET_HEARTBEAT_INTERVAL_SECONDS"),
		DefaultTimeoutSeconds:    envInt("RUNLET_DEFAULT_TIMEOUT_SECONDS"),
		DefaultWorkspace:         os.Getenv("RUNLET_WORKSPACE"),
		Shell:                    os.Getenv("RUNLET_SHELL"),
		Labels:                   ParseLabels(os.Getenv("RUNLET_LABELS")),
	}
}

func (s Seed) Merge(other Seed) Seed {
	out := s
	if other.APIURL != "" {
		out.APIURL = other.APIURL
	}
	if other.Token != "" {
		out.Token = other.Token
	}
	if other.RunnerID != "" {
		out.RunnerID = other.RunnerID
	}
	if other.Name != "" {
		out.Name = other.Name
	}
	if other.Concurrency != 0 {
		out.Concurrency = other.Concurrency
	}
	if other.PollIntervalSeconds != 0 {
		out.PollIntervalSeconds = other.PollIntervalSeconds
	}
	if other.HeartbeatIntervalSeconds != 0 {
		out.HeartbeatIntervalSeconds = other.HeartbeatIntervalSeconds
	}
	if other.DefaultTimeoutSeconds != 0 {
		out.DefaultTimeoutSeconds = other.DefaultTimeoutSeconds
	}
	if other.DefaultWorkspace != "" {
		out.DefaultWorkspace = other.DefaultWorkspace
	}
	if other.Shell != "" {
		out.Shell = other.Shell
	}
	if len(other.Labels) > 0 {
		if out.Labels == nil {
			out.Labels = map[string]string{}
		}
		for key, value := range other.Labels {
			out.Labels[key] = value
		}
	}
	return out
}

func (c *Config) applyDefaults() {
	if c.Name == "" {
		hostname, _ := os.Hostname()
		if hostname == "" {
			hostname = "runlet-runner"
		}
		c.Name = hostname
	}
	if c.Concurrency <= 0 {
		c.Concurrency = 1
	}
	if c.PollIntervalSeconds <= 0 {
		c.PollIntervalSeconds = 5
	}
	if c.HeartbeatIntervalSeconds <= 0 {
		c.HeartbeatIntervalSeconds = 15
	}
	if c.DefaultTimeoutSeconds <= 0 {
		c.DefaultTimeoutSeconds = 900
	}
	if c.Labels == nil {
		c.Labels = map[string]string{}
	}
	if _, ok := c.Labels["os"]; !ok {
		c.Labels["os"] = runtime.GOOS
	}
	if _, ok := c.Labels["arch"]; !ok {
		c.Labels["arch"] = runtime.GOARCH
	}
}

func (c Config) Validate() error {
	if c.APIURL == "" {
		return errors.New("apiUrl is required")
	}
	if c.Token == "" {
		return errors.New("token is required")
	}
	if c.Concurrency != 1 {
		return fmt.Errorf("only concurrency=1 is supported in this runner version, got %d", c.Concurrency)
	}
	return nil
}

func ParseLabels(value string) map[string]string {
	labels := map[string]string{}
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, val, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		labels[key] = strings.TrimSpace(val)
	}
	if len(labels) == 0 {
		return nil
	}
	return labels
}

func (c Config) PollInterval() time.Duration {
	return time.Duration(c.PollIntervalSeconds) * time.Second
}

func (c Config) HeartbeatInterval() time.Duration {
	return time.Duration(c.HeartbeatIntervalSeconds) * time.Second
}

func (c Config) DefaultTimeout() time.Duration {
	return time.Duration(c.DefaultTimeoutSeconds) * time.Second
}

type prompt struct {
	scanner *bufio.Scanner
	writer  io.Writer
}

func newPrompt(reader io.Reader, writer io.Writer) prompt {
	return prompt{
		scanner: bufio.NewScanner(reader),
		writer:  writer,
	}
}

func (p prompt) ask(label, defaultValue string) (string, error) {
	if defaultValue == "" {
		fmt.Fprintf(p.writer, "%s: ", label)
	} else {
		fmt.Fprintf(p.writer, "%s [%s]: ", label, defaultValue)
	}
	if !p.scanner.Scan() {
		if err := p.scanner.Err(); err != nil {
			return "", err
		}
		return defaultValue, nil
	}
	value := strings.TrimSpace(p.scanner.Text())
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func (p prompt) askInt(label string, defaultValue int) (int, error) {
	value, err := p.ask(label, strconv.Itoa(defaultValue))
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", label, err)
	}
	return parsed, nil
}

func (p prompt) askLabels(label string, defaults map[string]string) (map[string]string, error) {
	value, err := p.ask(label+" (key=value,key=value)", FormatLabels(defaults))
	if err != nil {
		return nil, err
	}
	labels := ParseLabels(value)
	if labels == nil {
		labels = map[string]string{}
	}
	return labels, nil
}

func FormatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+labels[key])
	}
	return strings.Join(parts, ",")
}

func envInt(key string) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsed
}
