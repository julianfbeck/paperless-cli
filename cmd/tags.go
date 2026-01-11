package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage tags",
	Long:  `List, create, edit, and delete tags.`,
}

var tagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tags",
	Long: `List all tags in Paperless.

Example:
  paperless tags list
  paperless tags list --json`,
	RunE: runTagsList,
}

var tagsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get tag details",
	Long: `Get detailed information about a tag.

Example:
  paperless tags get 5`,
	Args: cobra.ExactArgs(1),
	RunE: runTagsGet,
}

var tagsCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new tag",
	Long: `Create a new tag.

Example:
  paperless tags create "receipts"
  paperless tags create "important" --color "#ff0000"`,
	Args: cobra.ExactArgs(1),
	RunE: runTagsCreate,
}

var tagsEditCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit a tag",
	Long: `Edit a tag's properties.

Example:
  paperless tags edit 5 --name "new name"
  paperless tags edit 5 --color "#00ff00"`,
	Args: cobra.ExactArgs(1),
	RunE: runTagsEdit,
}

var tagsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a tag",
	Long: `Delete a tag.

Example:
  paperless tags delete 5
  paperless tags delete 5 --force`,
	Args: cobra.ExactArgs(1),
	RunE: runTagsDelete,
}

var (
	tagColor      string
	tagName       string
	tagForce      bool
)

func init() {
	rootCmd.AddCommand(tagsCmd)
	tagsCmd.AddCommand(tagsListCmd)
	tagsCmd.AddCommand(tagsGetCmd)
	tagsCmd.AddCommand(tagsCreateCmd)
	tagsCmd.AddCommand(tagsEditCmd)
	tagsCmd.AddCommand(tagsDeleteCmd)

	tagsCreateCmd.Flags().StringVar(&tagColor, "color", "", "tag color (hex, e.g. #ff0000)")
	tagsEditCmd.Flags().StringVar(&tagName, "name", "", "new name")
	tagsEditCmd.Flags().StringVar(&tagColor, "color", "", "new color (hex)")
	tagsDeleteCmd.Flags().BoolVarP(&tagForce, "force", "f", false, "skip confirmation")
}

func runTagsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	result, err := client.ListTags()
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(result)
	}

	if len(result.Results) == 0 {
		fmt.Println("No tags found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tCOLOR\tDOCS")
	for _, tag := range result.Results {
		fmt.Fprintf(w, "%d\t%s\t%s\t%d\n", tag.ID, tag.Name, tag.Color, tag.DocumentCount)
	}
	w.Flush()

	return nil
}

func runTagsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid tag ID: %s", args[0])
	}

	tag, err := client.GetTag(id)
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(tag)
	}

	fmt.Printf("ID:        %d\n", tag.ID)
	fmt.Printf("Name:      %s\n", tag.Name)
	fmt.Printf("Slug:      %s\n", tag.Slug)
	fmt.Printf("Color:     %s\n", tag.Color)
	fmt.Printf("Documents: %d\n", tag.DocumentCount)
	fmt.Printf("Inbox:     %t\n", tag.IsInboxTag)

	return nil
}

func runTagsCreate(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	tag, err := client.CreateTag(args[0], tagColor)
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(tag)
	}

	if !isQuiet() {
		fmt.Printf("Created tag %d: %s\n", tag.ID, tag.Name)
	}

	return nil
}

func runTagsEdit(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid tag ID: %s", args[0])
	}

	updates := make(map[string]interface{})
	if tagName != "" {
		updates["name"] = tagName
	}
	if tagColor != "" {
		updates["color"] = tagColor
	}

	if len(updates) == 0 {
		return fmt.Errorf("no changes specified")
	}

	tag, err := client.UpdateTag(id, updates)
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(tag)
	}

	if !isQuiet() {
		fmt.Printf("Updated tag %d\n", id)
	}

	return nil
}

func runTagsDelete(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid tag ID: %s", args[0])
	}

	if !tagForce {
		if !confirmAction(fmt.Sprintf("Delete tag %d?", id)) {
			fmt.Println("Cancelled")
			return nil
		}
	}

	if err := client.DeleteTag(id); err != nil {
		return err
	}

	if !isQuiet() {
		fmt.Printf("Deleted tag %d\n", id)
	}

	return nil
}
