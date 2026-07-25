package main

import (
	"testing"
)

func TestDisconnectCmd(t *testing.T) {
	disconnectCmd("github")
}

func TestConfigLLMCmd(t *testing.T) {
	configLLMCmd("http://localhost:8080", "test-key", "gpt-4")
	cfg := loadConfig()
	if cfg.LLM.BaseURL != "http://localhost:8080" {
		t.Errorf("expected http://localhost:8080, got %s", cfg.LLM.BaseURL)
	}
	if cfg.LLM.APIKey != "test-key" {
		t.Errorf("expected test-key, got %s", cfg.LLM.APIKey)
	}
	if cfg.LLM.Model != "gpt-4" {
		t.Errorf("expected gpt-4, got %s", cfg.LLM.Model)
	}
}
