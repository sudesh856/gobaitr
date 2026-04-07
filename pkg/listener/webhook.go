package listener

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type WebhookEvent struct {
	Event string 				`json:"event"`
	TokenID string				`json:"token_id"`
	TriggeredAt string			`json:"triggered_at"`
	RemoteIP string				`json:"remote_ip"`
	UserAgent string			`json:"user_agent"`
	Headers map[string][]string `json:"headers"`
}


func dispatchWebhook(url, tokenID, remoteIP, userAgent string, headers http.Header) {
	payload := WebhookEvent{
		Event: "token_triggered",
		TokenID: tokenID,
		TriggeredAt: time.Now().UTC().Format(time.RFC3339),
		RemoteIP: remoteIP,
		UserAgent: userAgent,
		Headers: headers,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for attempt := 1; attempt <= 2; attempt++ {
		resp, err := client.Post(url, "application/json", bytes.NewReader(body))
		if err == nil {
			resp.Body.Close()
			return
		}
		if attempt == 1 {
			time.Sleep(2 * time.Second)
		}
	}
}