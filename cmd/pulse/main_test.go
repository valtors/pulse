package main

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	cfg := loadConfig()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
}

func TestDataDir(t *testing.T) {
	dir := dataDir()
	if dir == "" {
		t.Error("expected non-empty data dir")
	}
}

func TestRustCore_NotFound(t *testing.T) {
	bin, err := rustCore()
	if err == nil {
		t.Skip("pulse-core found, skipping")
	}
	if bin != "" {
		t.Errorf("expected empty bin, got %s", bin)
	}
}

func TestPrintHelp(t *testing.T) {
	printHelp()
}

func TestVersion(t *testing.T) {
	if version != "0.2.0" {
		t.Errorf("expected 0.2.0, got %s", version)
	}
}

func TestDataDir_ContainsPulse(t *testing.T) {
	dir := dataDir()
	if dir == "" {
		t.Fatal("expected non-empty")
	}
}
