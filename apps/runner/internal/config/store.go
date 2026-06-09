package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const (
	defaultConfigDir  = ".runlet"
	defaultConfigFile = "settings.json"
	defaultPIDFile    = "runner.pid"
	defaultLogFile    = "runner.log"
)

type settingsFile struct {
	APIURL                   string            `json:"apiUrl,omitempty"`
	Token                    string            `json:"token,omitempty"`
	RunnerID                 string            `json:"runnerId,omitempty"`
	Name                     string            `json:"name,omitempty"`
	Concurrency              int               `json:"concurrency,omitempty"`
	PollIntervalSeconds      int               `json:"pollIntervalSeconds,omitempty"`
	HeartbeatIntervalSeconds int               `json:"heartbeatIntervalSeconds,omitempty"`
	DefaultTimeoutSeconds    int               `json:"defaultTimeoutSeconds,omitempty"`
	DefaultWorkspace         string            `json:"workspace,omitempty"`
	Shell                    string            `json:"shell,omitempty"`
	Labels                   map[string]string `json:"labels,omitempty"`
}

func DefaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, defaultConfigDir), nil
}

func DefaultPath() (string, error) {
	dir, err := DefaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, defaultConfigFile), nil
}

func DefaultPIDPath() (string, error) {
	dir, err := DefaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, defaultPIDFile), nil
}

func DefaultLogPath() (string, error) {
	dir, err := DefaultDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, defaultLogFile), nil
}

func LoadSeed(path string) (Seed, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Seed{}, nil
		}
		return Seed{}, err
	}

	var stored settingsFile
	if err := json.Unmarshal(data, &stored); err != nil {
		return Seed{}, err
	}
	return seedFromSettings(stored), nil
}

func SaveSeed(path string, seed Seed) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(settingsFromSeed(seed), "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return os.WriteFile(path, data, 0o600)
}

func settingsFromSeed(seed Seed) settingsFile {
	return settingsFile{
		APIURL:                   seed.APIURL,
		Token:                    seed.Token,
		RunnerID:                 seed.RunnerID,
		Name:                     seed.Name,
		Concurrency:              seed.Concurrency,
		PollIntervalSeconds:      seed.PollIntervalSeconds,
		HeartbeatIntervalSeconds: seed.HeartbeatIntervalSeconds,
		DefaultTimeoutSeconds:    seed.DefaultTimeoutSeconds,
		DefaultWorkspace:         seed.DefaultWorkspace,
		Shell:                    seed.Shell,
		Labels:                   seed.Labels,
	}
}

func seedFromSettings(stored settingsFile) Seed {
	return Seed{
		APIURL:                   stored.APIURL,
		Token:                    stored.Token,
		RunnerID:                 stored.RunnerID,
		Name:                     stored.Name,
		Concurrency:              stored.Concurrency,
		PollIntervalSeconds:      stored.PollIntervalSeconds,
		HeartbeatIntervalSeconds: stored.HeartbeatIntervalSeconds,
		DefaultTimeoutSeconds:    stored.DefaultTimeoutSeconds,
		DefaultWorkspace:         stored.DefaultWorkspace,
		Shell:                    stored.Shell,
		Labels:                   stored.Labels,
	}
}
