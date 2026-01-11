package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var typesCmd = &cobra.Command{
	Use:     "types",
	Aliases: []string{"type", "doctypes"},
	Short:   "Manage document types",
	Long:    `List, create, edit, and delete document types.`,
}

var typesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all document types",
	Long: `List all document types in Paperless.

Example:
  paperless types list
  paperless types list --json`,
	RunE: runTypesList,
}

var typesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get document type details",
	Long: `Get detailed information about a document type.

Example:
  paperless types get 5`,
	Args: cobra.ExactArgs(1),
	RunE: runTypesGet,
}

var typesCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new document type",
	Long: `Create a new document type.

Example:
  paperless types create "Invoice"`,
	Args: cobra.ExactArgs(1),
	RunE: runTypesCreate,
}

var typesEditCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit a document type",
	Long: `Edit a document type's properties.

Example:
  paperless types edit 5 --name "New Name"`,
	Args: cobra.ExactArgs(1),
	RunE: runTypesEdit,
}

var typesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a document type",
	Long: `Delete a document type.

Example:
  paperless types delete 5
  paperless types delete 5 --force`,
	Args: cobra.ExactArgs(1),
	RunE: runTypesDelete,
}

var (
	typeName  string
	typeForce bool
)

func init() {
	rootCmd.AddCommand(typesCmd)
	typesCmd.AddCommand(typesListCmd)
	typesCmd.AddCommand(typesGetCmd)
	typesCmd.AddCommand(typesCreateCmd)
	typesCmd.AddCommand(typesEditCmd)
	typesCmd.AddCommand(typesDeleteCmd)

	typesEditCmd.Flags().StringVar(&typeName, "name", "", "new name")
	typesDeleteCmd.Flags().BoolVarP(&typeForce, "force", "f", false, "skip confirmation")
}

func runTypesList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	result, err := client.ListDocumentTypes()
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(result)
	}

	if len(result.Results) == 0 {
		fmt.Println("No document types found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tDOCS")
	for _, dt := range result.Results {
		fmt.Fprintf(w, "%d\t%s\t%d\n", dt.ID, dt.Name, dt.DocumentCount)
	}
	w.Flush()

	return nil
}

func runTypesGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid document type ID: %s", args[0])
	}

	dt, err := client.GetDocumentType(id)
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(dt)
	}

	fmt.Printf("ID:        %d\n", dt.ID)
	fmt.Printf("Name:      %s\n", dt.Name)
	fmt.Printf("Slug:      %s\n", dt.Slug)
	fmt.Printf("Documents: %d\n", dt.DocumentCount)

	return nil
}

func runTypesCreate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	dt, err := client.CreateDocumentType(args[0])
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(dt)
	}

	if !isQuiet() {
		fmt.Printf("Created document type %d: %s\n", dt.ID, dt.Name)
	}

	return nil
}

func runTypesEdit(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid document type ID: %s", args[0])
	}

	updates := make(map[string]interface{})
	if typeName != "" {
		updates["name"] = typeName
	}

	if len(updates) == 0 {
		return fmt.Errorf("no changes specified")
	}

	dt, err := client.UpdateDocumentType(id, updates)
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(dt)
	}

	if !isQuiet() {
		fmt.Printf("Updated document type %d\n", id)
	}

	return nil
}

func runTypesDelete(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid document type ID: %s", args[0])
	}

	if !typeForce {
		if !confirmAction(fmt.Sprintf("Delete document type %d?", id)) {
			fmt.Println("Cancelled")
			return nil
		}
	}

	if err := client.DeleteDocumentType(id); err != nil {
		return err
	}

	if !isQuiet() {
		fmt.Printf("Deleted document type %d\n", id)
	}

	return nil
}
