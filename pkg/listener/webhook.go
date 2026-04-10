package listener

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
)

type WebhookEvent struct {
	Event          string            `json:"event"`
	TokenID        string            `json:"token_id"`
	TokenType      string            `json:"token_type"`
	TokenNote      string            `json:"token_note"`
	TriggeredAt    string            `json:"triggered_at"`
	RemoteIP       string            `json:"remote_ip"`
	UserAgent      string            `json:"user_agent"`
	Headers        map[string]string `json:"headers"`
	GobaitrVersion string            `json:"gobaitr_version"`
}

const GobaitrVersion = "1.0.0"

func DispatchWebhook(url, tokenID, tokenType, tokenNote, remoteIP, userAgent string, headers http.Header) {
	if url == "" {
		return
	}

	flatHeaders := make(map[string]string)
	for k, v := range headers {
		flatHeaders[k] = v[0]
	}

	payload := WebhookEvent{
		Event:          "token_triggered",
		TokenID:        tokenID,
		TokenType:      tokenType,
		TokenNote:      tokenNote,
		TriggeredAt:    time.Now().UTC().Format(time.RFC3339),
		RemoteIP:       remoteIP,
		UserAgent:      userAgent,
		Headers:        flatHeaders,
		GobaitrVersion: GobaitrVersion,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[webhook] marshal error: %v\n", err)
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}

	attempt := func() error {
		resp, err := client.Post(url, "application/json", bytes.NewReader(body))
		if err != nil {
			return err
		}
		resp.Body.Close()
		if resp.StatusCode >= 400 {
			return fmt.Errorf("server returned %d", resp.StatusCode)
		}
		return nil
	}

	if err := attempt(); err != nil {
		time.Sleep(2 * time.Second)
		if err := attempt(); err != nil {
			color.New(color.FgYellow).Fprintf(os.Stderr, "Warning: webhook delivery failed (attempt 2/2): %s\n", url)
			return
		}
	}
	fmt.Fprintf(os.Stderr, "[webhook] delivered for token %s\n", tokenID)
}
