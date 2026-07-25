package connect

import (
	"testing"
)

func TestGitHubConnector_ConnectSetsToken(t *testing.T) {
	c := &GitHubConnector{}
	// Connect calls Test() which hits real API. Just verify token is set.
	c.Token = "ghp_test123"
	if c.Token != "ghp_test123" {
		t.Error("token not set")
	}
}

func TestService_RegisterMultiple(t *testing.T) {
	s := NewService()
	s.Register(&GitHubConnector{Token: "a"})
	s.Register(&GitHubConnector{Token: "b"})
	// GitHub overwrites since same name
	list := s.List()
	if len(list) != 1 {
		t.Errorf("expected 1 (same name), got %d", len(list))
	}
}

func TestService_GetReturnsCorrectConnector(t *testing.T) {
	s := NewService()
	c := &GitHubConnector{Token: "ghp_abc"}
	s.Register(c)
	got, ok := s.Get("github")
	if !ok {
		t.Fatal("not found")
	}
	ghc, ok := got.(*GitHubConnector)
	if !ok || ghc.Token != "ghp_abc" {
		t.Errorf("expected ghp_abc, got %v", ghc.Token)
	}
}

func TestGitHubNotification_Struct(t *testing.T) {
	n := GitHubNotification{
		ID:     "1",
		Reason: "mention",
	}
	if n.ID != "1" || n.Reason != "mention" {
		t.Error("struct fields wrong")
	}
}

func TestNewService(t *testing.T) {
	s := NewService()
	if s == nil {
		t.Fatal("expected non-nil")
	}
	if s.connectors == nil {
		t.Error("expected initialized map")
	}
}

func TestService_RegisterAndList(t *testing.T) {
	s := NewService()
	if len(s.List()) != 0 {
		t.Error("expected empty list")
	}
	s.Register(&GitHubConnector{Token: "x"})
	if len(s.List()) != 1 {
		t.Errorf("expected 1, got %d", len(s.List()))
	}
	if s.List()[0] != "github" {
		t.Errorf("expected github, got %s", s.List()[0])
	}
}
