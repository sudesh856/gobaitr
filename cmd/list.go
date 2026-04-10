package cmd

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/sudesh856/gobaitr/pkg/store"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all deployed tokens with status",
	RunE:  runList,
}

var (
	listTriggered bool
	listType      string
)

func init() {
	listCmd.Flags().BoolVar(&listTriggered, "triggered", false, "Show only triggered tokens")
	listCmd.Flags().StringVar(&listType, "type", "", "Filter by type: url|file|env")
}

func runList(cmd *cobra.Command, args []string) error {
	s, err := store.New()
	if err != nil {
		return fmt.Errorf("failed to open store: %w", err)
	}
	defer s.Close()

	tokens, err := s.ListFiltered(listTriggered, listType)
	if err != nil {
		return fmt.Errorf("error listing tokens: %w", err)
	}

	if len(tokens) == 0 {
		fmt.Println("No tokens found. Run `gobaitr generate` to create one.")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTYPE\tNOTE\tCREATED\tEXPIRES\tSTATUS")
	fmt.Fprintln(w,
		strings.Repeat("─", 14)+"\t"+
			strings.Repeat("─", 6)+"\t"+
			strings.Repeat("─", 16)+"\t"+
			strings.Repeat("─", 10)+"\t"+
			strings.Repeat("─", 10)+"\t"+
			strings.Repeat("─", 22))

	now := time.Now()

	for _, t := range tokens {
		id := fmt.Sprintf("%v", t["id"])
		if len(id) > 12 {
			id = id[:12] + "..."
		}

		note := fmt.Sprintf("%v", t["note"])
		if note == "" || note == "<nil>" {
			note = "-"
		}
		if len(note) > 16 {
			note = note[:13] + "..."
		}

		createdAt := fmt.Sprintf("%v", t["createdAt"])
		created, _ := time.Parse(time.RFC3339, createdAt)
		createdStr := created.Format("2006-01-02")

		expiresStr := "never"
		isExpired := false
		expiresAt := fmt.Sprintf("%v", t["expiresAt"])
		if expiresAt != "" && expiresAt != "<nil>" {
			exp, err := time.Parse(time.RFC3339, expiresAt)
			if err == nil {
				if exp.Before(now) {
					expiresStr = "EXPIRED"
					isExpired = true
				} else {
					expiresStr = exp.Format("2006-01-02")
				}
			}
		}

		triggered, _ := t["triggered"].(bool)
		eventCount, _ := t["eventCount"].(int)

		var status string
		switch {
		case isExpired:
			status = "⚪ expired"
		case triggered:
			status = colorAlert.Sprintf("🔴 TRIGGERED (%d events)", eventCount)
		default:
			status = colorSuccess.Sprintf("🟢 clean")
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			id, t["type"], note, createdStr, expiresStr, status)
	}

	return w.Flush()
}
