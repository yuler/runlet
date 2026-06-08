package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadSeed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runner.conf")
	seed := DefaultSeed()
	seed.APIURL = "http://localhost:3000/acme"
	seed.Token = "setup-token"
	seed.Labels = map[string]string{"kind": "desktop", "project": "runlet"}

	if err := SaveSeed(path, seed); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected 0600 permissions, got %v", info.Mode().Perm())
	}

	loaded, err := LoadSeed(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.APIURL != "http://localhost:3000/acme" {
		t.Fatalf("unexpected api url %q", loaded.APIURL)
	}
	if loaded.Token != "setup-token" {
		t.Fatalf("unexpected token %q", loaded.Token)
	}
	if loaded.Labels["project"] != "runlet" {
		t.Fatalf("unexpected labels %#v", loaded.Labels)
	}
}

func TestLoadSeedMissingFileUsesEmptySeed(t *testing.T) {
	seed, err := LoadSeed(filepath.Join(t.TempDir(), "missing.conf"))
	if err != nil {
		t.Fatal(err)
	}
	if seed.Token != "" || seed.APIURL != "" {
		t.Fatalf("expected empty seed, got %#v", seed)
	}
}
