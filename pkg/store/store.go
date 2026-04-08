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

type DB struct {
	*sql.DB
}


func (s *Store) GetDB() *sql.DB {
    return s.db
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

	if _, err := db.Exec("PRAGMA busy_timeout=5000;"); err != nil {
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
	expires_at DATETIME,
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

func (s *Store) Insert(id, typ, secret, callbackURL, note string, createdAt time.Time, expiresAt *time.Time) error {
	var exp interface{}
	if expiresAt != nil {
		exp = expiresAt.Format(time.RFC3339)
	}
	_, err := s.db.Exec(
		`INSERT INTO tokens (id, type, secret, callback_url, note, created_at, triggered, expires_at) VALUES (?, ?, ?, ?, ?, ?, 0, ?)`,
		id, typ, secret, callbackURL, note, createdAt.Format(time.RFC3339), exp,
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

func (s *Store) GetEvents(tokenID string)  ([]map[string]interface{}, error) {
	rows, err := s.db.Query(

		`SELECT id, token_id, fired_at, remote_ip, user_agent, headers
		FROM events WHERE token_id = ? ORDER BY fired_at DESC`,
		tokenID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var id int64
		var tokenID, firedAt string
		var remoteIP, userAgent, headers sql.NullString
		if err := rows.Scan(&id, &tokenID, &firedAt, &remoteIP, &userAgent, &headers); err != nil {
			return nil, err
		}

		events = append(events, map[string]interface{}{
			"id": id,
			"tokenID": tokenID,
			"firedAt": firedAt,
			"remoteIP": remoteIP.String,
			"userAgent": userAgent.String,
			"headers": headers.String,

		})
	}
	return events, rows.Err()
}

func (s *Store) ListFiltered(onlyTriggered bool, filterType string) ([]map[string]any, error) {
	query := `SELECT id, type, note, created_at, expires_at, triggered,
		(SELECT COUNT(*) FROM events WHERE events.token_id = tokens.id) as event_count
		FROM tokens WHERE 1=1`
	args := []any{}

	if onlyTriggered {
		query += " AND triggered = 1"
	}
	if filterType != "" {
		query += " AND type = ?"
		args = append(args, filterType)
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, typ, createdAt string
		var note, expiresAt sql.NullString
		var triggered, eventCount int
		if err := rows.Scan(&id, &typ, &note, &createdAt, &expiresAt, &triggered, &eventCount); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"id":         id,
			"type":       typ,
			"note":       note.String,
			"createdAt":  createdAt,
			"expiresAt":  expiresAt.String,
			"triggered":  triggered == 1,
			"eventCount": eventCount,
		})
	}
	return results, rows.Err()
}

func (s *Store) EventCount(tokenID string) int {
	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM events WHERE token_id = ?`, tokenID).Scan(&count)
	return count
}

func (s *Store) DeleteToken(id string) (int, error) {
	count := s.EventCount(id)
	if _, err := s.db.Exec(`DELETE FROM events WHERE token_id = ?`, id); err != nil {
		return 0, err
	}
	if _, err := s.db.Exec(`DELETE FROM tokens WHERE id = ?`, id); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) MarkTriggered(tokenID, remoteIP, userAgent, headersJSON string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339)

	if _, err := tx.Exec(
		`UPDATE tokens SET triggered = 1, triggered_at = ?, triggered_by = ? WHERE id = ?`,
		now, remoteIP, tokenID,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(
		`INSERT INTO events (token_id, fired_at, remote_ip, user_agent, headers) VALUES (?, ?, ?, ?, ?)`,
		tokenID, now, remoteIP, userAgent, headersJSON,
	); err != nil {
		return err
	}

	return tx.Commit()
}
