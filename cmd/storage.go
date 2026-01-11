package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var storageCmd = &cobra.Command{
	Use:     "storage",
	Aliases: []string{"paths", "storage-paths"},
	Short:   "Manage storage paths",
	Long:    `List, create, and delete storage paths.`,
}

var storageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all storage paths",
	Long: `List all storage paths in Paperless.

Example:
  paperless storage list
  paperless storage list --json`,
	RunE: runStorageList,
}

var storageGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get storage path details",
	Long: `Get detailed information about a storage path.

Example:
  paperless storage get 5`,
	Args: cobra.ExactArgs(1),
	RunE: runStorageGet,
}

var storageCreateCmd = &cobra.Command{
	Use:   "create <name> <path>",
	Short: "Create a new storage path",
	Long: `Create a new storage path.

Example:
  paperless storage create "Archive" "archive/{{ created_year }}"`,
	Args: cobra.ExactArgs(2),
	RunE: runStorageCreate,
}

var storageDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a storage path",
	Long: `Delete a storage path.

Example:
  paperless storage delete 5
  paperless storage delete 5 --force`,
	Args: cobra.ExactArgs(1),
	RunE: runStorageDelete,
}

var storageForce bool

func init() {
	rootCmd.AddCommand(storageCmd)
	storageCmd.AddCommand(storageListCmd)
	storageCmd.AddCommand(storageGetCmd)
	storageCmd.AddCommand(storageCreateCmd)
	storageCmd.AddCommand(storageDeleteCmd)

	storageDeleteCmd.Flags().BoolVarP(&storageForce, "force", "f", false, "skip confirmation")
}

func runStorageList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	result, err := client.ListStoragePaths()
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(result)
	}

	if len(result.Results) == 0 {
		fmt.Println("No storage paths found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tPATH\tDOCS")
	for _, sp := range result.Results {
		fmt.Fprintf(w, "%d\t%s\t%s\t%d\n", sp.ID, sp.Name, truncate(sp.Path, 40), sp.DocumentCount)
	}
	w.Flush()

	return nil
}

func runStorageGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid storage path ID: %s", args[0])
	}

	sp, err := client.GetStoragePath(id)
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(sp)
	}

	fmt.Printf("ID:        %d\n", sp.ID)
	fmt.Printf("Name:      %s\n", sp.Name)
	fmt.Printf("Path:      %s\n", sp.Path)
	fmt.Printf("Slug:      %s\n", sp.Slug)
	fmt.Printf("Documents: %d\n", sp.DocumentCount)

	return nil
}

func runStorageCreate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	sp, err := client.CreateStoragePath(args[0], args[1])
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(sp)
	}

	if !isQuiet() {
		fmt.Printf("Created storage path %d: %s\n", sp.ID, sp.Name)
	}

	return nil
}

func runStorageDelete(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid storage path ID: %s", args[0])
	}

	if !storageForce {
		if !confirmAction(fmt.Sprintf("Delete storage path %d?", id)) {
			fmt.Println("Cancelled")
			return nil
		}
	}

	if err := client.DeleteStoragePath(id); err != nil {
		return err
	}

	if !isQuiet() {
		fmt.Printf("Deleted storage path %d\n", id)
	}

	return nil
}
