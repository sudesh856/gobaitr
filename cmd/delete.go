package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/sudesh856/gobaitr/pkg/store"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <token_id>",
	Short: "Delete a token and all its events",
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

var deleteForce bool

func init() {
	deleteCmd.Flags().BoolVar(&deleteForce, "force", false, "Skip confirmation prompt (for scripting)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	tokenID := args[0]

	s, err := store.New()
	if err != nil {
		return fmt.Errorf("failed to open store: %w", err)
	}
	defer s.Close()

	_, err = s.GetByID(tokenID)
	if err != nil {
		red := color.New(color.FgRed)
		red.Fprintf(os.Stderr, "Error: token %s not found.\n", tokenID)
		os.Exit(1)
return nil
	}

	eventCount := s.EventCount(tokenID)

	if !deleteForce {
		fmt.Printf("Delete token %s and all %d event(s)? [y/N] ", shortID(tokenID), eventCount)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "y" && input != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	if _, err := s.DeleteToken(tokenID); err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	fmt.Printf("\033[32m✓ Token %s deleted.\033[0m\n", shortID(tokenID))
	return nil
}