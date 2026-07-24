package memory

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

type Memory struct {
	ID       int64     `json:"id"`
	Key      string    `json:"key"`
	Value    string    `json:"value"`
	Category string    `json:"category"`
	Created  time.Time `json:"created"`
	Accessed time.Time `json:"accessed"`
}

func New(path string) (*Store, error) {
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	schema := `
	CREATE TABLE IF NOT EXISTS memory (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL UNIQUE,
		value TEXT NOT NULL,
		category TEXT NOT NULL DEFAULT 'general',
		created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		accessed DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_memory_category ON memory(category);
	CREATE INDEX IF NOT EXISTS idx_memory_key ON memory(key);
	`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Remember(key, value, category string) error {
	if category == "" {
		category = "general"
	}
	_, err := s.db.Exec(`
		INSERT INTO memory (key, value, category) VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = ?, category = ?, accessed = CURRENT_TIMESTAMP
	`, key, value, category, value, category)
	return err
}

func (s *Store) Recall(key string) (string, error) {
	var value string
	err := s.db.QueryRow(`SELECT value FROM memory WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	s.db.Exec(`UPDATE memory SET accessed = CURRENT_TIMESTAMP WHERE key = ?`, key)
	return value, err
}

func (s *Store) All() ([]Memory, error) {
	rows, err := s.db.Query(`SELECT id, key, value, category, created, accessed FROM memory ORDER BY accessed DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var memories []Memory
	for rows.Next() {
		var m Memory
		if err := rows.Scan(&m.ID, &m.Key, &m.Value, &m.Category, &m.Created, &m.Accessed); err != nil {
			return nil, err
		}
		memories = append(memories, m)
	}
	return memories, rows.Err()
}

func (s *Store) ByCategory(category string) ([]Memory, error) {
	rows, err := s.db.Query(`SELECT id, key, value, category, created, accessed FROM memory WHERE category = ? ORDER BY accessed DESC`, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var memories []Memory
	for rows.Next() {
		var m Memory
		if err := rows.Scan(&m.ID, &m.Key, &m.Value, &m.Category, &m.Created, &m.Accessed); err != nil {
			return nil, err
		}
		memories = append(memories, m)
	}
	return memories, rows.Err()
}

func (s *Store) Forget(key string) error {
	_, err := s.db.Exec(`DELETE FROM memory WHERE key = ?`, key)
	return err
}

func (s *Store) Close() error {
	return s.db.Close()
}
