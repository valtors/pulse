package memory

import (
	"os"
	"path/filepath"
	"testing"
)

func testStore(t *testing.T) *Store {
	dir := t.TempDir()
	s, err := New(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestRememberRecall(t *testing.T) {
	s := testStore(t)
	if err := s.Remember("os", "linux", "system"); err != nil {
		t.Fatalf("Remember: %v", err)
	}
	val, err := s.Recall("os")
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if val != "linux" {
		t.Errorf("expected linux, got %s", val)
	}
}

func TestRecallMissing(t *testing.T) {
	s := testStore(t)
	val, err := s.Recall("nonexistent")
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if val != "" {
		t.Errorf("expected empty string, got %s", val)
	}
}

func TestRememberUpdates(t *testing.T) {
	s := testStore(t)
	s.Remember("key", "old", "cat")
	s.Remember("key", "new", "cat2")
	val, _ := s.Recall("key")
	if val != "new" {
		t.Errorf("expected new, got %s", val)
	}
}

func TestRememberDefaultCategory(t *testing.T) {
	s := testStore(t)
	s.Remember("k", "v", "")
	mems, _ := s.All()
	if len(mems) != 1 {
		t.Fatalf("expected 1, got %d", len(mems))
	}
	if mems[0].Category != "general" {
		t.Errorf("expected general, got %s", mems[0].Category)
	}
}

func TestAll(t *testing.T) {
	s := testStore(t)
	s.Remember("a", "1", "x")
	s.Remember("b", "2", "y")
	mems, _ := s.All()
	if len(mems) != 2 {
		t.Errorf("expected 2, got %d", len(mems))
	}
}

func TestByCategory(t *testing.T) {
	s := testStore(t)
	s.Remember("a", "1", "work")
	s.Remember("b", "2", "personal")
	s.Remember("c", "3", "work")
	mems, _ := s.ByCategory("work")
	if len(mems) != 2 {
		t.Errorf("expected 2, got %d", len(mems))
	}
}

func TestForget(t *testing.T) {
	s := testStore(t)
	s.Remember("a", "1", "x")
	if err := s.Forget("a"); err != nil {
		t.Fatalf("Forget: %v", err)
	}
	val, _ := s.Recall("a")
	if val != "" {
		t.Errorf("expected empty after forget, got %s", val)
	}
}

func TestForgetMissing(t *testing.T) {
	s := testStore(t)
	if err := s.Forget("nonexistent"); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestNewCreatesSchema(t *testing.T) {
	s := testStore(t)
	mems, err := s.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(mems) != 0 {
		t.Errorf("expected empty store, got %d", len(mems))
	}
}

func TestNewCreatesDBFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	s, err := New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	s.Close()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected db file at %s, got %v", path, err)
	}
}
