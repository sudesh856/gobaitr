package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(home, ".gobaitr")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", filepath.Join(dir, "tokens.db"))
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`

	CREATE TABLE IF NOT EXISTS tokens (
	id TEXT PRIMARY KEY,
	type TEXT NOT NULL,
	secret TEXT NOT NULL,
	callback_url TEXT NOT NULL,
	note TEXT,
	created_at DATETIME NOT NULL,
	triggered INTEGER DEFAULT 0,
	triggered_at DATETIME,
	triggered_by TEXT
	);
	CREATE TABLE IF NOT EXISTS events (

	id INTEGER PRIMARY KEY AUTOINCREMENT,
	token_id TEXT NOT NULL,
	fired_at DATETIME,
	remote_ip TEXT,
	user_agent TEXT,
	headers TEXT
	);
	`)
	return err
}

func (s *Store) Save(t interface {
	GetFields() (string, string, string, string, string, time.Time)
}) error {
	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Insert(id, typ, secret, callbackURL, note string, createdAt time.Time) error {
	_, err := s.db.Exec(
		`INSERT INTO tokens (id, type, secret, callback_url, note, created_at, triggered) VALUES (?, ?, ?, ?, ?, ?, 0)`,
		id, typ, secret, callbackURL, note, createdAt.Format(time.RFC3339),
	)
	return err
}

func (s *Store) List() ([]map[string]interface{}, error) {
	rows, err := s.db.Query(`SELECT id, type, note, created_at, triggered FROM tokens ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, typ, createdAt string
		var note sql.NullString
		var triggered int
		if err := rows.Scan(&id, &typ, &note, &createdAt, &triggered); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"id":        id,
			"type":      typ,
			"note":      note.String,
			"createdAt": createdAt,
			"triggered": triggered == 1,
		})
	}
	return results, nil
}

func (s *Store) GetByID(id string) (map[string]interface{}, error) {
	row := s.db.QueryRow(`SELECT id, type, secret, callback_url, note, created_at, triggered FROM tokens WHERE id = ?`, id)
	var tid, typ, secret, callbackURL, createdAt string
	var note sql.NullString
	var triggered int
	if err := row.Scan(&tid, &typ, &secret, &callbackURL, &note, &createdAt, &triggered); err != nil {
		return nil, fmt.Errorf("token not found: %s", id)
	}

	return map[string]interface{}{
		"id":          tid,
		"type":        typ,
		"secret":      secret,
		"callbackURL": callbackURL,
		"note":        note.String,
		"createdAt":   createdAt,
		"triggered":   triggered == 1,
	}, nil
}
