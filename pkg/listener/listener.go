package listener

import (
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"golang.org/x/time/rate"
)

type Config struct {
	Port     int
	Webhook  string
	Quiet    bool
	TLS      bool
	CertFile string
	KeyFile  string
}

var (
	ipLimiters = make(map[string]*rate.Limiter)
	mu         sync.Mutex
)

func getIPLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()
	if l, ok := ipLimiters[ip]; ok {
		return l
	}

	l := rate.NewLimiter(rate.Every(10*time.Second), 5)
	ipLimiters[ip] = l
	return l
}

func Start(cfg Config, db *sql.DB) error {
	mux := buildMux(cfg, db)

	if !cfg.TLS {
		color.New(color.FgYellow, color.Bold).Fprintf(os.Stderr, "WARNING: listener bound to :%d without TLS — secrets visible on wire\n", cfg.Port)
	}

	if cfg.TLS {
		color.New(color.FgGreen, color.Bold).Printf("● gobaitr listening on :%d (TLS)\n", cfg.Port)
		err := http.ListenAndServeTLS(fmt.Sprintf(":%d", cfg.Port), cfg.CertFile, cfg.KeyFile, mux)
		if err != nil {
			if strings.Contains(err.Error(), "address already in use") || strings.Contains(err.Error(), "Only one usage") || strings.Contains(err.Error(), "Only one usage") {
				color.New(color.FgRed).Fprintf(os.Stderr, "Error: port %d is already in use. Use --port to specify a different port.\n", cfg.Port)
				os.Exit(1)
			}
			return err
		}
		return nil
	}

	color.New(color.FgGreen, color.Bold).Printf("● gobaitr listening on :%d\n", cfg.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), mux)
	if err != nil {
		if strings.Contains(err.Error(), "address already in use") || strings.Contains(err.Error(), "Only one usage") || strings.Contains(err.Error(), "Only one usage") {
			color.New(color.FgRed).Fprintf(os.Stderr, "Error: port %d is already in use. Use --port to specify a different port.\n", cfg.Port)
			os.Exit(1)
		}
		return err
	}
	return nil
}

func buildMux(cfg Config, db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/t/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/t/"), "/")
		if len(parts) != 2 {
			w.WriteHeader(http.StatusOK)
			return
		}

		tokenID := parts[0]
		requestSecret := parts[1]

		remoteIP := r.RemoteAddr
		if idx := strings.LastIndex(remoteIP, ":"); idx != -1 {
			remoteIP = remoteIP[:idx]
		}

		if !getIPLimiter(remoteIP).Allow() {
			w.WriteHeader(http.StatusOK)
			return
		}

		row := db.QueryRow(`SELECT secret, expires_at, triggered FROM tokens WHERE id = ?`, tokenID)
		var storedSecret string
		var expiresAt sql.NullString
		var triggered int
		if err := row.Scan(&storedSecret, &expiresAt, &triggered); err != nil {
			w.WriteHeader(http.StatusOK)
			return
		}

		if subtle.ConstantTimeCompare([]byte(storedSecret), []byte(requestSecret)) != 1 {
			w.WriteHeader(http.StatusOK)
			return
		}

		if expiresAt.Valid && expiresAt.String != "" {
			exp, err := time.Parse(time.RFC3339, expiresAt.String)
			if err == nil && time.Now().UTC().After(exp) {
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		now := time.Now().UTC()
		db.Exec(`UPDATE tokens SET triggered=1, triggered_at=?, triggered_by=? WHERE id=?`,
			now.Format(time.RFC3339), remoteIP, tokenID)

		headers, _ := json.Marshal(r.Header)
		db.Exec(`INSERT INTO events (token_id, fired_at, remote_ip, user_agent, headers) VALUES (?, ?, ?, ?, ?)`,
			tokenID, now.Format(time.RFC3339), remoteIP, r.UserAgent(), string(headers))

		if !cfg.Quiet {
			color.New(color.FgRed, color.Bold).Printf("\n🔴 TRIGGERED — token %s — %s\n\n", tokenID, remoteIP)
		}

		if cfg.Webhook != "" {
			var tokenType, tokenNote string
			db.QueryRow(`SELECT type, note FROM tokens WHERE id = ?`, tokenID).Scan(&tokenType, &tokenNote)
			go DispatchWebhook(cfg.Webhook, tokenID, tokenType, tokenNote, remoteIP, r.UserAgent(), r.Header)
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		var count int
		db.QueryRow(`SELECT COUNT(*) FROM tokens`).Scan(&count)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok", "tokens_deployed":%d}`, count)
	})

	return mux
}