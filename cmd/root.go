package cmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	quietMode  bool
	noColor    bool
	urlFlag    string
	version    = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "paperless",
	Short: "CLI for Paperless-ngx document management",
	Long: `A command-line interface for managing documents in Paperless-ngx.

Set PAPERLESS_URL and PAPERLESS_TOKEN environment variables for authentication,
or use 'paperless config set-url' and 'paperless config set-token' to save them.`,
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	rootCmd.PersistentFlags().BoolVarP(&quietMode, "quiet", "q", false, "suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")
	rootCmd.PersistentFlags().StringVarP(&urlFlag, "url", "u", "", "Paperless server URL (overrides env/config)")
}

func isJSON() bool {
	return jsonOutput
}

func isQuiet() bool {
	return quietMode
}

func printJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
