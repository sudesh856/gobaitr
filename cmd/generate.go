package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/sudesh856/gobaitr/pkg/store"
	"github.com/sudesh856/gobaitr/pkg/token"
)

var (
	generatePort int
	generateNote string
)

var generateCmd = &cobra.Command{
	Use:   "generate [url|file|env]",
	Short: "Generate a new canary token",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tokenType := args[0]
		if tokenType != "url" && tokenType != "file" && tokenType != "env" {
			fmt.Fprint(os.Stderr, "Error: type must be url, file, or env")
			os.Exit(1)
		}

		t, err := token.Generate(tokenType, generatePort, generateNote)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		st, err := store.New()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		defer st.Close()

		if err := st.Insert(t.ID, t.Type, t.Secret, t.CallbackURL, t.Note, t.CreatedAt); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		green := color.New(color.FgGreen, color.Bold)

		switch tokenType {
		case "env":
			green.Printf("✓ Token %s created (env)\n", t.ID)
			fmt.Printf("  Add this to your shell or .env file:\n")
			fmt.Printf("  export API_KEY=\"%s\"\n", t.CallbackURL)
		case "file":
			green.Printf("✓ Token %s created (file)\n", t.ID)
			fmt.Printf("  Callback: %s\n", t.CallbackURL)
			fmt.Printf("  Use 'gobaitr embed --token %s --target <file>' to inject it.\n", t.ID)
		default:
			green.Printf("✓ Token %s created (url)\n", t.ID)
			fmt.Printf("  Callback: %s\n", t.CallbackURL)
		}
		if t.Note != "" {
			fmt.Printf("  Note:     %s\n", t.Note)
		}
	},
}

func init() {
	generateCmd.Flags().IntVar(&generatePort, "port", 8080, "Listener port for callback URL")
	generateCmd.Flags().StringVar(&generateNote, "note", "", "Optional label for this token")
	rootCmd.AddCommand(generateCmd)
}
