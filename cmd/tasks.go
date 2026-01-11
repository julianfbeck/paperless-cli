package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Manage tasks",
	Long:  `Check status of background tasks (e.g., document processing).`,
}

var tasksStatusCmd = &cobra.Command{
	Use:   "status <task-id>",
	Short: "Check task status",
	Long: `Check the status of a background task.

Example:
  paperless tasks status abc-123-def`,
	Args: cobra.ExactArgs(1),
	RunE: runTasksStatus,
}

func init() {
	rootCmd.AddCommand(tasksCmd)
	tasksCmd.AddCommand(tasksStatusCmd)
}

func runTasksStatus(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	task, err := client.GetTask(args[0])
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(task)
	}

	fmt.Printf("Task ID:     %s\n", task.TaskID)
	fmt.Printf("Status:      %s\n", task.Status)
	fmt.Printf("Type:        %s\n", task.Type)
	fmt.Printf("File:        %s\n", task.TaskFileName)
	fmt.Printf("Created:     %s\n", task.DateCreated)
	if task.DateDone != "" {
		fmt.Printf("Completed:   %s\n", task.DateDone)
	}
	if task.Result != "" {
		fmt.Printf("Result:      %s\n", task.Result)
	}
	if task.RelatedDoc != "" {
		fmt.Printf("Document:    %s\n", task.RelatedDoc)
	}

	return nil
}
