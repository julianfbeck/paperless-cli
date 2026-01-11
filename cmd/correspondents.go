package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var correspondentsCmd = &cobra.Command{
	Use:     "correspondents",
	Aliases: []string{"corr", "correspondent"},
	Short:   "Manage correspondents",
	Long:    `List, create, edit, and delete correspondents.`,
}

var corrListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all correspondents",
	Long: `List all correspondents in Paperless.

Example:
  paperless correspondents list
  paperless correspondents list --json`,
	RunE: runCorrList,
}

var corrGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get correspondent details",
	Long: `Get detailed information about a correspondent.

Example:
  paperless correspondents get 5`,
	Args: cobra.ExactArgs(1),
	RunE: runCorrGet,
}

var corrCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new correspondent",
	Long: `Create a new correspondent.

Example:
  paperless correspondents create "ACME Corp"`,
	Args: cobra.ExactArgs(1),
	RunE: runCorrCreate,
}

var corrEditCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit a correspondent",
	Long: `Edit a correspondent's properties.

Example:
  paperless correspondents edit 5 --name "New Name"`,
	Args: cobra.ExactArgs(1),
	RunE: runCorrEdit,
}

var corrDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a correspondent",
	Long: `Delete a correspondent.

Example:
  paperless correspondents delete 5
  paperless correspondents delete 5 --force`,
	Args: cobra.ExactArgs(1),
	RunE: runCorrDelete,
}

var (
	corrName  string
	corrForce bool
)

func init() {
	rootCmd.AddCommand(correspondentsCmd)
	correspondentsCmd.AddCommand(corrListCmd)
	correspondentsCmd.AddCommand(corrGetCmd)
	correspondentsCmd.AddCommand(corrCreateCmd)
	correspondentsCmd.AddCommand(corrEditCmd)
	correspondentsCmd.AddCommand(corrDeleteCmd)

	corrEditCmd.Flags().StringVar(&corrName, "name", "", "new name")
	corrDeleteCmd.Flags().BoolVarP(&corrForce, "force", "f", false, "skip confirmation")
}

func runCorrList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	result, err := client.ListCorrespondents()
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(result)
	}

	if len(result.Results) == 0 {
		fmt.Println("No correspondents found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tDOCS")
	for _, corr := range result.Results {
		fmt.Fprintf(w, "%d\t%s\t%d\n", corr.ID, corr.Name, corr.DocumentCount)
	}
	w.Flush()

	return nil
}

func runCorrGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid correspondent ID: %s", args[0])
	}

	corr, err := client.GetCorrespondent(id)
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(corr)
	}

	fmt.Printf("ID:        %d\n", corr.ID)
	fmt.Printf("Name:      %s\n", corr.Name)
	fmt.Printf("Slug:      %s\n", corr.Slug)
	fmt.Printf("Documents: %d\n", corr.DocumentCount)

	return nil
}

func runCorrCreate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	corr, err := client.CreateCorrespondent(args[0])
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(corr)
	}

	if !isQuiet() {
		fmt.Printf("Created correspondent %d: %s\n", corr.ID, corr.Name)
	}

	return nil
}

func runCorrEdit(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid correspondent ID: %s", args[0])
	}

	updates := make(map[string]interface{})
	if corrName != "" {
		updates["name"] = corrName
	}

	if len(updates) == 0 {
		return fmt.Errorf("no changes specified")
	}

	corr, err := client.UpdateCorrespondent(id, updates)
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(corr)
	}

	if !isQuiet() {
		fmt.Printf("Updated correspondent %d\n", id)
	}

	return nil
}

func runCorrDelete(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid correspondent ID: %s", args[0])
	}

	if !corrForce {
		if !confirmAction(fmt.Sprintf("Delete correspondent %d?", id)) {
			fmt.Println("Cancelled")
			return nil
		}
	}

	if err := client.DeleteCorrespondent(id); err != nil {
		return err
	}

	if !isQuiet() {
		fmt.Printf("Deleted correspondent %d\n", id)
	}

	return nil
}
