package config

import (
	"strings"
	"testing"
	"time"
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

func TestValidateRequiresAPIURL(t *testing.T) {
	cfg := Config{Token: "token", Concurrency: 1}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected missing api url error")
	}
}

func TestValidateRequiresToken(t *testing.T) {
	cfg := Config{APIURL: "http://localhost:3000", Concurrency: 1}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected missing token error")
	}
}

func TestFromSeedFailsWhenRequiredValuesMissing(t *testing.T) {
	if _, err := FromSeed(Seed{Concurrency: 1}); err == nil {
		t.Fatal("expected error when api url and token are empty")
	}
}

func TestSeedMergeLayersOverrides(t *testing.T) {
	base := Seed{
		APIURL:                   "http://default",
		Token:                    "default-token",
		Name:                     "default",
		PollIntervalSeconds:      5,
		HeartbeatIntervalSeconds: 15,
		Labels:                   map[string]string{"a": "1"},
	}
	override := Seed{
		APIURL: "http://override",
		Labels: map[string]string{"b": "2", "a": "x"},
	}

	merged := base.Merge(override)

	if merged.APIURL != "http://override" {
		t.Fatalf("expected override api url, got %q", merged.APIURL)
	}
	if merged.Token != "default-token" {
		t.Fatalf("expected default token to survive, got %q", merged.Token)
	}
	if merged.PollIntervalSeconds != 5 {
		t.Fatalf("expected default poll interval to survive, got %d", merged.PollIntervalSeconds)
	}
	if merged.Labels["a"] != "x" || merged.Labels["b"] != "2" {
		t.Fatalf("expected merged labels, got %#v", merged.Labels)
	}
}

func TestSeedFromEnvReadsKnownVars(t *testing.T) {
	t.Setenv("RUNLET_API_URL", "http://env-host")
	t.Setenv("RUNLET_TOKEN", "env-token")
	t.Setenv("RUNLET_CONCURRENCY", "1")
	t.Setenv("RUNLET_LABELS", "kind=ci,project=runlet")

	seed := SeedFromEnv()

	if seed.APIURL != "http://env-host" {
		t.Fatalf("expected api url from env, got %q", seed.APIURL)
	}
	if seed.Token != "env-token" {
		t.Fatalf("expected token from env, got %q", seed.Token)
	}
	if seed.Concurrency != 1 {
		t.Fatalf("expected concurrency 1, got %d", seed.Concurrency)
	}
	if seed.Labels["kind"] != "ci" || seed.Labels["project"] != "runlet" {
		t.Fatalf("unexpected env labels: %#v", seed.Labels)
	}
}

func TestSeedFromEnvIgnoresInvalidInts(t *testing.T) {
	t.Setenv("RUNLET_POLL_INTERVAL_SECONDS", "notnum")

	seed := SeedFromEnv()
	if seed.PollIntervalSeconds != 0 {
		t.Fatalf("expected invalid int to fall back to 0, got %d", seed.PollIntervalSeconds)
	}
}

func TestApplyDefaultsPopulatesOSAndArchLabels(t *testing.T) {
	cfg, err := FromSeed(Seed{APIURL: "http://localhost:3000", Token: "t"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Labels["os"] == "" || cfg.Labels["arch"] == "" {
		t.Fatalf("expected os/arch labels, got %#v", cfg.Labels)
	}
}

func TestParseLabelsEmptyReturnsNil(t *testing.T) {
	if got := ParseLabels(""); got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
}

func TestParseLabelsTrimsWhitespace(t *testing.T) {
	got := ParseLabels(" kind = desktop ,  project=runlet ")
	if got["kind"] != "desktop" || got["project"] != "runlet" {
		t.Fatalf("unexpected labels %#v", got)
	}
}

func TestFormatLabelsEmpty(t *testing.T) {
	if got := FormatLabels(nil); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestPollAndHeartbeatDurations(t *testing.T) {
	cfg := Config{PollIntervalSeconds: 3, HeartbeatIntervalSeconds: 7, DefaultTimeoutSeconds: 60}

	if cfg.PollInterval() != 3*time.Second {
		t.Fatalf("unexpected poll interval %v", cfg.PollInterval())
	}
	if cfg.HeartbeatInterval() != 7*time.Second {
		t.Fatalf("unexpected heartbeat interval %v", cfg.HeartbeatInterval())
	}
	if cfg.DefaultTimeout() != 60*time.Second {
		t.Fatalf("unexpected timeout %v", cfg.DefaultTimeout())
	}
}
