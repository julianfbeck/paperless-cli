package cmd

import (
	"fmt"

	"github.com/julianfbeck/paperless-cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long:  `Manage paperless-cli configuration settings.`,
}

var configSetURLCmd = &cobra.Command{
	Use:   "set-url <url>",
	Short: "Set the Paperless server URL",
	Long: `Set the default Paperless server URL.

Example:
  paperless config set-url https://paperless.example.com`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigSetURL,
}

var configSetTokenCmd = &cobra.Command{
	Use:   "set-token <token>",
	Short: "Set the API token",
	Long: `Set the API authentication token.

Example:
  paperless config set-token abc123def456`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigSetToken,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long: `Show the current configuration settings.

Example:
  paperless config show`,
	RunE: runConfigShow,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetURLCmd)
	configCmd.AddCommand(configSetTokenCmd)
	configCmd.AddCommand(configShowCmd)
}

func runConfigSetURL(cmd *cobra.Command, args []string) error {
	if err := config.SetURL(args[0]); err != nil {
		return fmt.Errorf("failed to save URL: %w", err)
	}

	if !isQuiet() {
		fmt.Printf("URL set to: %s\n", args[0])
	}

	return nil
}

func runConfigSetToken(cmd *cobra.Command, args []string) error {
	if err := config.SetToken(args[0]); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	if !isQuiet() {
		fmt.Println("Token saved")
	}

	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if isJSON() {
		return printJSON(map[string]string{
			"url":   cfg.URL,
			"token": maskToken(cfg.Token),
		})
	}

	fmt.Printf("URL:   %s\n", cfg.URL)
	fmt.Printf("Token: %s\n", maskToken(cfg.Token))

	// Show env overrides
	if envURL := config.GetURL(); envURL != cfg.URL && envURL != "" {
		fmt.Printf("\n(URL overridden by PAPERLESS_URL: %s)\n", envURL)
	}
	if envToken := config.GetToken(); envToken != cfg.Token && envToken != "" {
		fmt.Println("(Token overridden by PAPERLESS_TOKEN)")
	}

	return nil
}

func maskToken(token string) string {
	if token == "" {
		return "(not set)"
	}
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
