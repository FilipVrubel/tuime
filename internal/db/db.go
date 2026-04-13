package db

import (
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func Open(path string) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}

	return db, nil
}

func Migrate(db *sqlx.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS activities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			created_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_id INTEGER,
			started_at DATETIME NOT NULL,
			ended_at DATETIME NOT NULL,
			duration INTEGER NOT NULL,
			type TEXT NOT NULL CHECK(type IN ('pomodoro', 'tracker')),
			FOREIGN KEY(activity_id) REFERENCES activities(id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_activity ON sessions(activity_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_started ON sessions(started_at)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}
