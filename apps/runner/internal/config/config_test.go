package config

import (
	"strings"
	"testing"
)

func TestFromSeedAppliesDefaults(t *testing.T) {
	cfg, err := FromSeed(DefaultSeed())
	if err != nil {
		t.Fatal(err)
	}

	if cfg.APIURL != "http://localhost:3000" {
		t.Fatalf("expected local api url, got %q", cfg.APIURL)
	}
	if cfg.Token != "dev-token" {
		t.Fatalf("expected dev token, got %q", cfg.Token)
	}
	if cfg.Concurrency != 1 {
		t.Fatalf("expected concurrency 1, got %d", cfg.Concurrency)
	}
	if cfg.Labels["os"] == "" || cfg.Labels["arch"] == "" {
		t.Fatalf("expected os and arch labels, got %#v", cfg.Labels)
	}
}

func TestInquireCanAcceptAllDefaultValues(t *testing.T) {
	input := strings.NewReader(strings.Repeat("\n", 10))
	var output strings.Builder

	cfg, err := Inquire(input, &output, DefaultSeed())
	if err != nil {
		t.Fatal(err)
	}

	if cfg.APIURL != "http://localhost:3000" {
		t.Fatalf("expected default api url, got %q", cfg.APIURL)
	}
	if cfg.Token != "dev-token" {
		t.Fatalf("expected default token, got %q", cfg.Token)
	}
	if cfg.Labels["project"] != "runlet" {
		t.Fatalf("expected default project label, got %#v", cfg.Labels)
	}
	if !strings.Contains(output.String(), "Runner token [dev-token]") {
		t.Fatalf("expected token default in prompt, got %q", output.String())
	}
}

func TestInquireUsesDefaultsAndInput(t *testing.T) {
	input := strings.NewReader(strings.Join([]string{
		"",
		"prompt-token",
		"",
		"Prompt Runner",
		"/tmp/runlet",
		"/bin/zsh",
		"",
		"",
		"120",
		"kind=desktop,project=runlet",
		"",
	}, "\n"))
	var output strings.Builder

	cfg, err := Inquire(input, &output, Seed{
		APIURL: "http://localhost:3000",
		Token:  "seed-token",
	})
	if err != nil {
		t.Fatal(err)
	}

	if cfg.APIURL != "http://localhost:3000" {
		t.Fatalf("expected default api url, got %q", cfg.APIURL)
	}
	if cfg.Token != "prompt-token" {
		t.Fatalf("expected prompted token, got %q", cfg.Token)
	}
	if cfg.Name != "Prompt Runner" {
		t.Fatalf("expected prompted name, got %q", cfg.Name)
	}
	if cfg.DefaultTimeoutSeconds != 120 {
		t.Fatalf("expected timeout 120, got %d", cfg.DefaultTimeoutSeconds)
	}
	if cfg.Labels["project"] != "runlet" {
		t.Fatalf("expected project label, got %#v", cfg.Labels)
	}
	if !strings.Contains(output.String(), "Core API URL") {
		t.Fatalf("expected prompt output, got %q", output.String())
	}
}

func TestParseLabels(t *testing.T) {
	labels := ParseLabels("kind=desktop, project=runlet,invalid")
	if labels["kind"] != "desktop" || labels["project"] != "runlet" {
		t.Fatalf("unexpected labels %#v", labels)
	}
	if _, ok := labels["invalid"]; ok {
		t.Fatalf("invalid label should be ignored: %#v", labels)
	}
}

func TestFormatLabelsSortsKeys(t *testing.T) {
	got := FormatLabels(map[string]string{
		"project": "runlet",
		"kind":    "desktop",
	})
	if got != "kind=desktop,project=runlet" {
		t.Fatalf("unexpected label format %q", got)
	}
}

func TestValidateRejectsUnsupportedConcurrency(t *testing.T) {
	cfg := Config{
		APIURL:      "http://localhost:3000",
		Token:       "token",
		Concurrency: 2,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected unsupported concurrency error")
	}
}
