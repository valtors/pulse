package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingReturnsDefaults(t *testing.T) {
	c, err := Load("/nonexistent/path/config.json")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.LLM.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("expected default base url, got %s", c.LLM.BaseURL)
	}
	if c.LLM.Model != "gpt-4o-mini" {
		t.Errorf("expected default model, got %s", c.LLM.Model)
	}
	if c.Connectors == nil {
		t.Error("expected non-nil connectors")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	c := &Config{}
	c.Connectors = make(map[string]struct {
		Token string `json:"token"`
	})
	c.LLM.BaseURL = "http://localhost:8080"
	c.LLM.Model = "test-model"
	c.SetConnector("github", "ghp_token123")
	if err := c.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.LLM.BaseURL != "http://localhost:8080" {
		t.Errorf("expected custom base url, got %s", loaded.LLM.BaseURL)
	}
	if loaded.LLM.Model != "test-model" {
		t.Errorf("expected test-model, got %s", loaded.LLM.Model)
	}
	if loaded.GetConnector("github") != "ghp_token123" {
		t.Errorf("expected ghp_token123, got %s", loaded.GetConnector("github"))
	}
}

func TestSetGetConnector(t *testing.T) {
	c := &Config{}
	c.Connectors = make(map[string]struct {
		Token string `json:"token"`
	})
	c.SetConnector("slack", "xoxb_token")
	if c.GetConnector("slack") != "xoxb_token" {
		t.Errorf("expected xoxb_token, got %s", c.GetConnector("slack"))
	}
}

func TestGetMissingConnector(t *testing.T) {
	c := &Config{}
	c.Connectors = make(map[string]struct {
		Token string `json:"token"`
	})
	if c.GetConnector("nope") != "" {
		t.Errorf("expected empty, got %s", c.GetConnector("nope"))
	}
}

func TestRemoveConnector(t *testing.T) {
	c := &Config{}
	c.Connectors = make(map[string]struct {
		Token string `json:"token"`
	})
	c.SetConnector("test", "val")
	c.RemoveConnector("test")
	if c.GetConnector("test") != "" {
		t.Errorf("expected empty after remove, got %s", c.GetConnector("test"))
	}
}

func TestDefaultPath(t *testing.T) {
	p := DefaultPath()
	if p == "" {
		t.Error("expected non-empty default path")
	}
}

func TestSaveCreatesDir(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "deep", "nested", "config.json")
	c := &Config{}
	c.Connectors = make(map[string]struct {
		Token string `json:"token"`
	})
	if err := c.Save(nested); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(nested); err != nil {
		t.Errorf("expected file at %s", nested)
	}
}

func TestLoadDefaultsWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	os.WriteFile(path, []byte(`{}`), 0600)
	c, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.LLM.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("expected default base url, got %s", c.LLM.BaseURL)
	}
	if c.LLM.Model != "gpt-4o-mini" {
		t.Errorf("expected default model, got %s", c.LLM.Model)
	}
}
