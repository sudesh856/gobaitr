package store

import (
	"database/sql"
	"time"
)

type Token struct {
	ID          string
	Type        string
	Note        string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
	Triggered   bool
	EventCount  int
}

type Event struct {
	ID        int64
	TokenID   string
	FiredAt   time.Time
	RemoteIP  string
	UserAgent string
	Headers   string
}

func (db *DB) GetToken(id string) (*Token, error) {
	row := db.QueryRow(
		`SELECT id, type, note, created_at, expires_at, triggered FROM tokens WHERE id = ?`, id,
	)
	return scanToken(row)
}

func scanToken(row *sql.Row) (*Token, error) {
	t := &Token{}
	var expiresAt sql.NullString
	var createdStr string
	err := row.Scan(&t.ID, &t.Type, &t.Note, &createdStr, &expiresAt, &t.Triggered)
	if err != nil {
		return nil, err
	}
	t.CreatedAt, _ = time.Parse("2006-01-02T15:04:05Z", createdStr)
	if expiresAt.Valid {
		exp, _ := time.Parse("2006-01-02T15:04:05Z", expiresAt.String)
		t.ExpiresAt = &exp
	}
	return t, nil
}

func (db *DB) GetEvents(tokenID string) ([]Event, error) {
	rows, err := db.Query(
		`SELECT id, token_id, fired_at, remote_ip, user_agent, headers FROM events WHERE token_id = ? ORDER BY fired_at DESC`,
		tokenID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var firedStr string
		if err := rows.Scan(&e.ID, &e.TokenID, &firedStr, &e.RemoteIP, &e.UserAgent, &e.Headers); err != nil {
			continue
		}
		e.FiredAt, _ = time.Parse("2006-01-02T15:04:05Z", firedStr)
		events = append(events, e)
	}
	return events, rows.Err()
}

func (db *DB) ListTokens(filterTriggered bool, filterType string) ([]Token, error) {
	query := `SELECT id, type, note, created_at, expires_at, triggered,
		(SELECT COUNT(*) FROM events WHERE events.token_id = tokens.id) as event_count
		FROM tokens WHERE 1=1`
	args := []interface{}{}

	if filterTriggered {
		query += " AND triggered = 1"
	}
	if filterType != "" {
		query += " AND type = ?"
		args = append(args, filterType)
	}
	query += " ORDER BY created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []Token
	for rows.Next() {
		var t Token
		var expiresAt sql.NullString
		var createdStr string
		if err := rows.Scan(&t.ID, &t.Type, &t.Note, &createdStr, &expiresAt, &t.Triggered, &t.EventCount); err != nil {
			continue
		}
		t.CreatedAt, _ = time.Parse("2006-01-02T15:04:05Z", createdStr)
		if expiresAt.Valid {
			exp, _ := time.Parse("2006-01-02T15:04:05Z", expiresAt.String)
			t.ExpiresAt = &exp
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}

func (db *DB) DeleteToken(id string) (int64, error) {
	var count int64
	db.QueryRow(`SELECT COUNT(*) FROM events WHERE token_id = ?`, id).Scan(&count)

	_, err := db.Exec(`DELETE FROM events WHERE token_id = ?`, id)
	if err != nil {
		return 0, err
	}
	_, err = db.Exec(`DELETE FROM tokens WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (db *DB) EventCount(tokenID string) int {
	var count int
	db.QueryRow(`SELECT COUNT(*) FROM events WHERE token_id = ?`, tokenID).Scan(&count)
	return count
}