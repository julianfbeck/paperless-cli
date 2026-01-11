package cmd

import (
	"fmt"
	"os"

	"github.com/julianfbeck/paperless-cli/internal/api"
	"github.com/julianfbeck/paperless-cli/internal/config"
)

// getClient returns an authenticated API client
func getClient() (*api.Client, error) {
	url := urlFlag
	if url == "" {
		url = config.GetURL()
	}
	if url == "" {
		return nil, fmt.Errorf("no server URL configured. Set PAPERLESS_URL or run 'paperless config set-url <url>'")
	}

	token := config.GetToken()
	if token == "" {
		return nil, fmt.Errorf("no API token configured. Set PAPERLESS_TOKEN or run 'paperless config set-token <token>'")
	}

	return api.NewClient(url, token), nil
}

// confirmAction asks for user confirmation
func confirmAction(message string) bool {
	if quietMode {
		return false
	}
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", message)
	var response string
	fmt.Scanln(&response)
	return response == "y" || response == "Y" || response == "yes" || response == "Yes"
}
