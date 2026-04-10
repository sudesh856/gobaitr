package listener

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec(`CREATE TABLE IF NOT EXISTS tokens (
	id TEXT PRIMARY KEY, type TEXT, secret TEXT, callback_url TEXT,
	note TEXT, created_at DATETIME, expires_at DATETIME,
	triggered INTEGER DEFAULT 0, triggered_at DATETIME, triggered_by TEXT
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS events (
	id INTEGER PRIMARY KEY AUTOINCREMENT , token_id TEXT,
	fired_at DATETIME, remote_ip TEXT, user_agent TEXT, headers TEXT
	)`)

	return db
}

func TestListenerHit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Exec(`INSERT INTO tokens (id, type, secret, callback_url, note, created_at, triggered)
	VALUES ('tok1', 'url', 'secret333', 'http://x', '', datetime('now'), 0)`)

	cfg := Config{Port: 0}
	mux := buildMux(cfg, db)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/t/tok1/secret333")
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var triggered int
	db.QueryRow(`SELECT triggered FROM tokens WHERE id='tok1'`).Scan(&triggered)
	if triggered != 1 {
		t.Error("expected token to be triggered")
	}
}

func TestListenerBadSecret(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Exec(`INSERT INTO tokens (id, type, secret, callback_url, note, created_at, triggered)
		VALUES ('tok2', 'url', 'correctsecret', 'http://x', '', datetime('now'), 0)`)

	cfg := Config{Port: 0, Quiet: true}
	mux := buildMux(cfg, db)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/t/tok2/wrongsecret")
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 (stealth), got %d", resp.StatusCode)
	}

	var triggered int
	db.QueryRow(`SELECT triggered FROM tokens WHERE id='tok2'`).Scan(&triggered)
	if triggered != 0 {
		t.Error("token should NOT be triggered with wrong secret")
	}
}

func TestListenerRateLimit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.Exec(`INSERT INTO tokens (id, type, secret, callback_url, note, created_at, triggered)
		VALUES ('tok3', 'url', 'sec', 'http://x', '', datetime('now'), 0)`)

	cfg := Config{Port: 0, Quiet: true}
	mux := buildMux(cfg, db)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	for i := 0; i < 10; i++ {
		resp, _ := http.Get(srv.URL + "/t/tok3/sec")
		if resp.StatusCode != 200 {
			t.Errorf("hit %d: expected 200, got %d", i, resp.StatusCode)
		}
	}
}

func TestTokenExpiry(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	expired := time.Now().UTC().Add(-1 * time.Second).Format(time.RFC3339)
	db.Exec(`INSERT INTO tokens (id, type, secret, callback_url, note, created_at, expires_at, triggered)
		VALUES ('tok4', 'url', 'sec', 'http://x', '', datetime('now'), ?, 0)`, expired)

	cfg := Config{Port: 0, Quiet: true}
	mux := buildMux(cfg, db)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	http.Get(srv.URL + "/t/tok4/sec")

	var triggered int
	db.QueryRow(`SELECT triggered FROM tokens WHERE id='tok4'`).Scan(&triggered)
	if triggered != 0 {
		t.Error("expired token should NOT be triggered")
	}
}

func TestListenerConstantTimeCompare(t *testing.T) {
	src, err := os.ReadFile("listener.go")
	if err != nil {
		t.Skip("listener.go not found")
	}
	if !strings.Contains(string(src), "subtle.ConstantTimeCompare") {
		t.Error("listener.go must use crypto/subtle.ConstantTimeCompare, not strings.Equal")
	}
}
