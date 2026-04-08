package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/sudesh856/gobaitr/pkg/store"
)

var verifyCmd = &cobra.Command{
	Use:   "verify <token_id>",
	Short: "Check if a token has been triggered and show full event history",
	Args:  cobra.ExactArgs(1),
	RunE:  runVerify,
}

var (
	verifyJSON bool
	verifyAll  bool
)

func init() {
	verifyCmd.Flags().BoolVar(&verifyJSON, "json", false, "Output as JSON instead of table")
	verifyCmd.Flags().BoolVar(&verifyAll, "all", false, "Show full headers JSON for each event")
}

func runVerify(cmd *cobra.Command, args []string) error {
	tokenID := args[0]

	s, err := store.New()
	if err != nil {
		return fmt.Errorf("failed to open store: %w", err)
	}

	defer s.Close()

	token, err := s.GetByID(tokenID)
	if err != nil {
		return fmt.Errorf("token %s not found", &tokenID)
	}

	events, err := s.GetEvents(tokenID)
	if err != nil {
		return fmt.Errorf("error fetching events: %w", err)
	}

	triggered, _ := token["triggered"].(bool)

	if verifyJSON {
		type eventOut struct {
			FiredAt   string            `json:"fired_at"`
			RemoteIP  string            `json:"remote_ip"`
			UserAgent string            `json:"user_agent"`
			Headers   map[string]string `json:"headers,omitempty"`
		}

		type out struct {
			TokenID    string     `json:"token_id"`
			Triggered  bool       `json:"triggered"`
			EventCount int        `json:"event_count"`
			Events     []eventOut `json:"events"`
		}

		o := out{
			TokenID:    tokenID,
			Triggered:  triggered,
			EventCount: len(events),
			Events:     []eventOut{},
		}

		for _, e := range events {
			eo := eventOut{
				FiredAt:   fmt.Sprintf("%v", e["firedAt"]),
				RemoteIP:  fmt.Sprintf("%v", e["remoteIP"]),
				UserAgent: fmt.Sprintf("%v", e["userAgent"]),
			}

			if verifyAll {
				var h map[string]string
				json.Unmarshal([]byte(fmt.Sprintf("%v", e["headers"])), &h)
				eo.Headers = h
			}
			o.Events = append(o.Events, eo)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(o)
		if triggered {
			os.Exit(1)
		}
		return nil
	}

	if !triggered {
		fmt.Printf("\033[32m CLEAN — token %s has not been triggered\033[0m\n", shortID(tokenID))
		os.Exit(0)
	}

	fmt.Printf("\033[31m TRIGGERED — %d event(s) recorded for token %s\033[0m\n", len(events), shortID(tokenID))
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "FIRED AT\tREMOTE IP\tUSER AGENT")
	fmt.Fprintln(w, strings.Repeat("─", 19)+"\t"+strings.Repeat("─", 15)+"\t"+strings.Repeat("─", 30))
	for _, e := range events {
		ua := fmt.Sprintf("%v", e["userAgent"])
		if !verifyAll && len(ua) > 40 {
			ua = ua[:37] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			e["firedAt"],
			e["remoteIP"],
			ua,
		)
		if verifyAll && e["headers"] != "" {
			fmt.Fprintf(w, "  headers: %s\n", e["headers"])
		}
	}
	w.Flush()

	os.Exit(1)
	return nil
}

func shortID(id string) string {
	if len(id) > 12 {
		return id[:12] + "..."
	}
	return id
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
