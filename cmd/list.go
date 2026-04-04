package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/sudesh856/gobaitr/pkg/store"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all deployed canary tokens",
	Run: func(cmd *cobra.Command, args []string) {
		st, err := store.New()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening store: %v\n", err)
			os.Exit(1)
		}
		defer st.Close()

		tokens, err := st.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing: %v\n", err)
			os.Exit(1)
		}

		if len(tokens) == 0 {
			fmt.Println("No tokens deployed yet. Run 'gobaitr generate url' to create one.")
			return
		}

		cyan := color.New(color.FgCyan)
		green := color.New(color.FgGreen, color.Bold)
		red := color.New(color.FgRed, color.Bold)

		cyan.Printf("%-38s %-6s %-20s %-25s %s\n", "ID", "TYPE", "NOTE", "CREATED", "STATUS")
		fmt.Println("----------------------------------------------------------------------------------------------------------------------")

		for _, t := range tokens {
			status := green.Sprint("clean")
			if t["triggered"].(bool) {
				status = red.Sprint("Triggered")
			}
			fmt.Printf("%-38s %-6s %-20s %-25s %s\n",
				t["id"], t["type"], t["note"], t["createdAt"], status)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}