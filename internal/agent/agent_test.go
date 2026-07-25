package agent

import (
	"testing"

	"github.com/valtors/pulse/internal/memory"
)

func newTestAgent(t *testing.T) *Agent {
	t.Helper()
	mem, err := memory.New(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("memory.New: %v", err)
	}
	t.Cleanup(func() { mem.Close() })
	return New(mem, nil)
}

func TestAgent_New(t *testing.T) {
	a := newTestAgent(t)
	if a == nil {
		t.Fatal("expected non-nil agent")
	}
}

func TestAgent_Connected_Empty(t *testing.T) {
	a := newTestAgent(t)
	if len(a.Connected()) != 0 {
		t.Error("expected empty connected for new agent")
	}
}

func TestAgent_Memory(t *testing.T) {
	a := newTestAgent(t)
	a.Memory().Remember("test_key", "test_value", "test")
	val, err := a.Memory().Recall("test_key")
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if val != "test_value" {
		t.Errorf("expected test_value, got %s", val)
	}
}

func TestAgent_ConnectGitHub_InvalidToken(t *testing.T) {
	a := newTestAgent(t)
	err := a.ConnectGitHub("invalid_token_12345")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestAgent_ConnectGmail_InvalidToken(t *testing.T) {
	a := newTestAgent(t)
	err := a.ConnectGmail("invalid_token_12345")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestAgent_ConnectCalendar_InvalidToken(t *testing.T) {
	a := newTestAgent(t)
	err := a.ConnectCalendar("invalid_token_12345")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestAgent_Do_NoConnections(t *testing.T) {
	a := newTestAgent(t)
	task, err := a.Do("what's up?")
	if err != nil {
		t.Logf("Do returned error (expected without connections): %v", err)
	}
	if task == nil {
		t.Log("Do returned nil task (expected without connections)")
	}
}

func TestAgent_Do_RememberCommand(t *testing.T) {
	a := newTestAgent(t)
	task, _ := a.Do("remember that I like pizza")
	if task != nil {
		t.Log("Do returned non-nil task")
	}
}

func TestAgent_Do_StatusQuery(t *testing.T) {
	a := newTestAgent(t)
	a.Do("remember that I like pizza")
	task, _ := a.Do("what do you know about me?")
	if task != nil {
		t.Log("Do returned non-nil task for status query")
	}
}

func TestAgent_Do_EmptyInput(t *testing.T) {
	a := newTestAgent(t)
	task, err := a.Do("")
	if err != nil {
		t.Logf("Do returned error for empty input: %v", err)
	}
	_ = task
}
