use rusqlite::{params, Connection};
use serde::{Deserialize, Serialize};
use std::path::Path;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Memory {
    pub id: i64,
    pub key: String,
    pub value: String,
    pub category: String,
    pub created: String,
    pub accessed: String,
}

pub struct Store {
    db: Connection,
}

impl Store {
    pub fn open<P: AsRef<Path>>(path: P) -> Result<Self, String> {
        let db = Connection::open(path).map_err(|e| e.to_string())?;
        db.execute_batch(
            "CREATE TABLE IF NOT EXISTS memory (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                key TEXT NOT NULL UNIQUE,
                value TEXT NOT NULL,
                category TEXT NOT NULL DEFAULT 'general',
                created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
                accessed DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
            );
            CREATE INDEX IF NOT EXISTS idx_memory_category ON memory(category);
            CREATE INDEX IF NOT EXISTS idx_memory_key ON memory(key);",
        )
        .map_err(|e| e.to_string())?;
        Ok(Store { db })
    }

    pub fn remember(&self, key: &str, value: &str, category: &str) -> Result<(), String> {
        let cat = if category.is_empty() { "general" } else { category };
        self.db
            .execute(
                "INSERT INTO memory (key, value, category) VALUES (?, ?, ?)
                 ON CONFLICT(key) DO UPDATE SET value = ?, category = ?, accessed = CURRENT_TIMESTAMP",
                params![key, value, cat, value, cat],
            )
            .map_err(|e| e.to_string())?;
        Ok(())
    }

    pub fn recall(&self, key: &str) -> Result<Option<String>, String> {
        let mut stmt = self
            .db
            .prepare("SELECT value FROM memory WHERE key = ?")
            .map_err(|e| e.to_string())?;
        let result = stmt
            .query_row(params![key], |row| row.get::<_, String>(0))
            .optional()
            .map_err(|e| e.to_string())?;
        if result.is_some() {
            self.db
                .execute(
                    "UPDATE memory SET accessed = CURRENT_TIMESTAMP WHERE key = ?",
                    params![key],
                )
                .ok();
        }
        Ok(result)
    }

    pub fn all(&self) -> Result<Vec<Memory>, String> {
        let mut stmt = self
            .db
            .prepare("SELECT id, key, value, category, created, accessed FROM memory ORDER BY accessed DESC")
            .map_err(|e| e.to_string())?;
        let items = stmt
            .query_map([], |row| {
                Ok(Memory {
                    id: row.get(0)?,
                    key: row.get(1)?,
                    value: row.get(2)?,
                    category: row.get(3)?,
                    created: row.get(4)?,
                    accessed: row.get(5)?,
                })
            })
            .map_err(|e| e.to_string())?;
        let mut memories = Vec::new();
        for item in items {
            memories.push(item.map_err(|e| e.to_string())?);
        }
        Ok(memories)
    }

    pub fn forget(&self, key: &str) -> Result<(), String> {
        self.db
            .execute("DELETE FROM memory WHERE key = ?", params![key])
            .map_err(|e| e.to_string())?;
        Ok(())
    }
}

use rusqlite::OptionalExtension;
