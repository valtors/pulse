package connect

import (
	"testing"
)

func TestGitHubConnector_Name(t *testing.T) {
	c := &GitHubConnector{}
	if c.Name() != "github" {
		t.Errorf("expected github, got %s", c.Name())
	}
}

func TestGitHubConnector_TestEmptyToken(t *testing.T) {
	c := &GitHubConnector{Token: ""}
	if err := c.Test(); err == nil {
		t.Error("expected error for empty token")
	}
}

func TestGitHubConnector_ConnectEmpty(t *testing.T) {
	c := &GitHubConnector{}
	if err := c.Connect(""); err == nil {
		t.Error("expected error for empty token")
	}
}

func TestService_RegisterAndGet(t *testing.T) {
	s := NewService()
	c := &GitHubConnector{Token: "ghp_x"}
	s.Register(c)

	got, ok := s.Get("github")
	if !ok {
		t.Fatal("expected to find github connector")
	}
	if got.Name() != "github" {
		t.Errorf("expected github, got %s", got.Name())
	}
}

func TestService_GetMissing(t *testing.T) {
	s := NewService()
	_, ok := s.Get("nonexistent")
	if ok {
		t.Error("expected not found")
	}
}

func TestService_ListEmpty(t *testing.T) {
	s := NewService()
	if len(s.List()) != 0 {
		t.Errorf("expected 0, got %d", len(s.List()))
	}
}

func TestService_ListWithConnectors(t *testing.T) {
	s := NewService()
	s.Register(&GitHubConnector{Token: "ghp_x"})
	if len(s.List()) != 1 {
		t.Errorf("expected 1, got %d", len(s.List()))
	}
}
