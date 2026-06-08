package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const defaultConfigFile = "runner.conf"

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "runlet", defaultConfigFile), nil
}

func LoadSeed(path string) (Seed, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Seed{}, nil
		}
		return Seed{}, err
	}
	defer file.Close()

	return readSeed(file)
}

func SaveSeed(path string, seed Seed) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()

	return writeSeed(file, seed)
}

func readSeed(reader io.Reader) (Seed, error) {
	var seed Seed
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, rawValue, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		value, err := strconv.Unquote(strings.TrimSpace(rawValue))
		if err != nil {
			value = strings.TrimSpace(rawValue)
		}
		applyValue(&seed, strings.TrimSpace(key), value)
	}
	if err := scanner.Err(); err != nil {
		return Seed{}, err
	}
	return seed, nil
}

func writeSeed(writer io.Writer, seed Seed) error {
	lines := []struct {
		key   string
		value string
	}{
		{"RUNLET_API_URL", seed.APIURL},
		{"RUNLET_TOKEN", seed.Token},
		{"RUNLET_RUNNER_ID", seed.RunnerID},
		{"RUNLET_RUNNER_NAME", seed.Name},
		{"RUNLET_CONCURRENCY", intString(seed.Concurrency)},
		{"RUNLET_POLL_INTERVAL_SECONDS", intString(seed.PollIntervalSeconds)},
		{"RUNLET_HEARTBEAT_INTERVAL_SECONDS", intString(seed.HeartbeatIntervalSeconds)},
		{"RUNLET_DEFAULT_TIMEOUT_SECONDS", intString(seed.DefaultTimeoutSeconds)},
		{"RUNLET_WORKSPACE", seed.DefaultWorkspace},
		{"RUNLET_SHELL", seed.Shell},
		{"RUNLET_LABELS", FormatLabels(seed.Labels)},
	}

	for _, line := range lines {
		if line.value == "" {
			continue
		}
		if _, err := fmt.Fprintf(writer, "%s=%s\n", line.key, strconv.Quote(line.value)); err != nil {
			return err
		}
	}
	return nil
}

func applyValue(seed *Seed, key, value string) {
	switch key {
	case "RUNLET_API_URL":
		seed.APIURL = value
	case "RUNLET_TOKEN":
		seed.Token = value
	case "RUNLET_RUNNER_ID":
		seed.RunnerID = value
	case "RUNLET_RUNNER_NAME":
		seed.Name = value
	case "RUNLET_CONCURRENCY":
		seed.Concurrency = parseInt(value)
	case "RUNLET_POLL_INTERVAL_SECONDS":
		seed.PollIntervalSeconds = parseInt(value)
	case "RUNLET_HEARTBEAT_INTERVAL_SECONDS":
		seed.HeartbeatIntervalSeconds = parseInt(value)
	case "RUNLET_DEFAULT_TIMEOUT_SECONDS":
		seed.DefaultTimeoutSeconds = parseInt(value)
	case "RUNLET_WORKSPACE":
		seed.DefaultWorkspace = value
	case "RUNLET_SHELL":
		seed.Shell = value
	case "RUNLET_LABELS":
		seed.Labels = ParseLabels(value)
	}
}

func intString(value int) string {
	if value == 0 {
		return ""
	}
	return strconv.Itoa(value)
}

func parseInt(value string) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return parsed
}
