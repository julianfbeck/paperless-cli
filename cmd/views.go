package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var viewsCmd = &cobra.Command{
	Use:     "views",
	Aliases: []string{"saved-views"},
	Short:   "Manage saved views",
	Long:    `List and view saved views.`,
}

var viewsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved views",
	Long: `List all saved views in Paperless.

Example:
  paperless views list
  paperless views list --json`,
	RunE: runViewsList,
}

var viewsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get saved view details",
	Long: `Get detailed information about a saved view.

Example:
  paperless views get 5`,
	Args: cobra.ExactArgs(1),
	RunE: runViewsGet,
}

func init() {
	rootCmd.AddCommand(viewsCmd)
	viewsCmd.AddCommand(viewsListCmd)
	viewsCmd.AddCommand(viewsGetCmd)
}

func runViewsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	result, err := client.ListSavedViews()
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(result)
	}

	if len(result.Results) == 0 {
		fmt.Println("No saved views found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tDASHBOARD\tSIDEBAR")
	for _, sv := range result.Results {
		dashboard := ""
		sidebar := ""
		if sv.ShowOnDashboard {
			dashboard = "yes"
		}
		if sv.ShowInSidebar {
			sidebar = "yes"
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", sv.ID, sv.Name, dashboard, sidebar)
	}
	w.Flush()

	return nil
}

func runViewsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid saved view ID: %s", args[0])
	}

	sv, err := client.GetSavedView(id)
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(sv)
	}

	fmt.Printf("ID:        %d\n", sv.ID)
	fmt.Printf("Name:      %s\n", sv.Name)
	fmt.Printf("Dashboard: %t\n", sv.ShowOnDashboard)
	fmt.Printf("Sidebar:   %t\n", sv.ShowInSidebar)
	fmt.Printf("Sort:      %s (reverse: %t)\n", sv.SortField, sv.SortReverse)

	return nil
}
