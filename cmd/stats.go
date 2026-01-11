package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show system statistics",
	Long: `Display system statistics from Paperless.

Example:
  paperless stats
  paperless stats --json`,
	RunE: runStats,
}

func init() {
	rootCmd.AddCommand(statsCmd)
}

func runStats(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	stats, err := client.GetStatistics()
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(stats)
	}

	if docTotal, ok := stats["documents_total"]; ok {
		fmt.Printf("Documents:        %.0f\n", docTotal)
	}
	if docInbox, ok := stats["documents_inbox"]; ok {
		fmt.Printf("In Inbox:         %.0f\n", docInbox)
	}
	if charTotal, ok := stats["character_count"]; ok {
		fmt.Printf("Characters:       %.0f\n", charTotal)
	}
	if tagCount, ok := stats["tag_count"]; ok {
		fmt.Printf("Tags:             %.0f\n", tagCount)
	}
	if corrCount, ok := stats["correspondent_count"]; ok {
		fmt.Printf("Correspondents:   %.0f\n", corrCount)
	}
	if dtCount, ok := stats["document_type_count"]; ok {
		fmt.Printf("Document Types:   %.0f\n", dtCount)
	}
	if spCount, ok := stats["storage_path_count"]; ok {
		fmt.Printf("Storage Paths:    %.0f\n", spCount)
	}

	return nil
}
