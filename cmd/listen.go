package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/sudesh856/gobaitr/pkg/listener"
	"github.com/sudesh856/gobaitr/pkg/store"
)

var (
	listenPort    int
	listenWebhook string
	listenQuiet   bool
	listenTLS     bool
	listenCert    string
	listenKey     string
)

var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Start the canary listener HTTP server",
	Run: func(cmd *cobra.Command, args []string) {
		st, err := store.New()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer st.Close()

		cfg := listener.Config{
			Port:     listenPort,
			Webhook:  listenWebhook,
			Quiet:    listenQuiet,
			TLS:      listenTLS,
			CertFile: listenCert,
			KeyFile:  listenKey,
		}

		if err := listener.Start(cfg, st.GetDB()); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	listenCmd.Flags().IntVar(&listenPort, "port", 8080, "Local port to bind")
	listenCmd.Flags().StringVar(&listenWebhook, "webhook", "", "URL to POST alert JSON on trigger")
	listenCmd.Flags().BoolVar(&listenQuiet, "quiet", false, "Suppress terminal output")
	listenCmd.Flags().BoolVar(&listenTLS, "tls", false, "Enable TLS")
	listenCmd.Flags().StringVar(&listenCert, "cert", "", "Path to TLS certificate file")
	listenCmd.Flags().StringVar(&listenKey, "key", "", "Path to TLS private key file")
	rootCmd.AddCommand(listenCmd)
}
