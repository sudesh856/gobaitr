package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/sudesh856/gobaitr/pkg/embed"
	"github.com/sudesh856/gobaitr/pkg/store"
)

var (

	embedToken string
	embedTarget string
	embedDryRun bool

)

var embedCmd = &cobra.Command{
	Use: "embed",
	Short: "Embed a canary token into a file",
	Run: func(cmd *cobra.Command , args []string) {
		if embedToken == "" || embedTarget == "" {
			fmt.Fprintln(os.Stderr, "Error: --token and --target are required!")
			os.Exit(1)
		}

		st, err := store.New()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		defer st.Close()

		token, err := st.GetByID(embedToken)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: token not found: %s. Run 'gobaitr list' to see all the available tokens.\n", embedToken)
			os.Exit(1)
		}

		callbackURL := token["callbackURL"].(string)
		ext := strings.ToLower(filepath.Ext(embedTarget))

		if embedDryRun {
			fmt.Printf("Dry run -- would embed into %s (token: %s)\n", embedTarget, embedToken)
			fmt.Printf("callback URL: %s\n",callbackURL)
			return
		}

		switch ext {
		case ".env":
			err = embed.EmbedEnv(embedTarget, callbackURL)
		case ".json":
			err = embed.EmbedJSON(embedTarget, callbackURL)
		default:
			err = embed.EmbedText(embedTarget, callbackURL)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		green := color.New(color.FgGreen, color.Bold)
		green.Printf("canary embedded in %s (token: %s)", embedTarget, embedToken)
	},
}

func init() {
	embedCmd.Flags().StringVar(&embedToken, "token", "", "Token ID to embed (required)")
	embedCmd.Flags().StringVar(&embedTarget, "target", "", "Target file path (required)")
	embedCmd.Flags().BoolVar(&embedDryRun, "dry-run", false, "Preview change without writing to disk")
	rootCmd.AddCommand(embedCmd)
}