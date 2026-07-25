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

func TestStore_New_InvalidPath(t *testing.T) {
	_, err := New("/nonexistent/dir/that/does/not/exist/db.sqlite")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestStore_All_Empty(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	items, err := s.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestStore_All_Multiple(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	s.Remember("key1", "val1", "cat1")
	s.Remember("key2", "val2", "cat2")
	s.Remember("key3", "val3", "cat1")
	items, err := s.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}
}

func TestStore_Recall_NotFound(t *testing.T) {
	s := testStore(t)
	s.Close()
	val, err := s.Recall("nonexistent")
	_ = val
	_ = err
}

func TestStore_Forget_Nonexistent(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	err := s.Forget("nonexistent")
	if err != nil {
		t.Errorf("expected nil for nonexistent key, got %v", err)
	}
}

func TestStore_Remember_Update(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	s.Remember("key", "val1", "cat")
	s.Remember("key", "val2", "cat")
	val, err := s.Recall("key")
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if val != "val2" {
		t.Errorf("expected val2, got %s", val)
	}
}

func TestStore_All_WithCategory(t *testing.T) {
	s := testStore(t)
	defer s.Close()
	s.Remember("a", "1", "food")
	s.Remember("b", "2", "work")
	s.Remember("c", "3", "food")
	items, _ := s.All()
	count := 0
	for _, item := range items {
		if item.Category == "food" {
			count++
		}
	}
	if count != 2 {
		t.Errorf("expected 2 food items, got %d", count)
	}
}
