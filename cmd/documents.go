package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/julianfbeck/paperless-cli/internal/api"
	"github.com/spf13/cobra"
)

var documentsCmd = &cobra.Command{
	Use:     "documents",
	Aliases: []string{"docs", "doc"},
	Short:   "Manage documents",
	Long:    `List, search, upload, download, and manage documents in Paperless.`,
}

var docsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List documents",
	Long: `List documents with optional filters.

Example:
  paperless documents list
  paperless documents list --query "invoice"
  paperless documents list --tag bills --limit 10`,
	RunE: runDocsList,
}

var docsSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search documents",
	Long: `Full-text search across all documents.

Example:
  paperless documents search "invoice 2024"
  paperless documents search "contract" --limit 5`,
	Args: cobra.ExactArgs(1),
	RunE: runDocsSearch,
}

var docsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get document details",
	Long: `Get detailed information about a document.

Example:
  paperless documents get 123`,
	Args: cobra.ExactArgs(1),
	RunE: runDocsGet,
}

var docsUploadCmd = &cobra.Command{
	Use:   "upload <file>...",
	Short: "Upload document(s)",
	Long: `Upload one or more documents to Paperless.

Example:
  paperless documents upload invoice.pdf
  paperless documents upload *.pdf --title "January Invoices"
  paperless documents upload doc.pdf --tag bills --correspondent "ACME"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runDocsUpload,
}

var docsDownloadCmd = &cobra.Command{
	Use:   "download <id>",
	Short: "Download document",
	Long: `Download a document file.

Example:
  paperless documents download 123
  paperless documents download 123 -o ~/Downloads/doc.pdf
  paperless documents download 123 --original`,
	Args: cobra.ExactArgs(1),
	RunE: runDocsDownload,
}

var docsEditCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit document metadata",
	Long: `Edit a document's metadata.

Example:
  paperless documents edit 123 --title "New Title"
  paperless documents edit 123 --add-tag important
  paperless documents edit 123 --correspondent "New Corp"`,
	Args: cobra.ExactArgs(1),
	RunE: runDocsEdit,
}

var docsDeleteCmd = &cobra.Command{
	Use:   "delete <id>...",
	Short: "Delete document(s)",
	Long: `Delete one or more documents.

Example:
  paperless documents delete 123
  paperless documents delete 123 456 789 --force`,
	Args: cobra.MinimumNArgs(1),
	RunE: runDocsDelete,
}

var docsContentCmd = &cobra.Command{
	Use:   "content <id>",
	Short: "Get document text content",
	Long: `Get the extracted text content of a document.

Example:
  paperless documents content 123`,
	Args: cobra.ExactArgs(1),
	RunE: runDocsContent,
}

// Flags
var (
	listQuery         string
	listTags          []string
	listCorrespondent string
	listDocType       string
	listCreatedAfter  string
	listCreatedBefore string
	listLimit         int
	listPage          int

	uploadTitle         string
	uploadCorrespondent string
	uploadDocType       string
	uploadTags          []string

	downloadOutput   string
	downloadOriginal bool

	editTitle            string
	editCorrespondent    string
	editDocType          string
	editAddTags          []string
	editRemoveTags       []string
	editASN              int

	deleteForce bool
)

func init() {
	rootCmd.AddCommand(documentsCmd)
	documentsCmd.AddCommand(docsListCmd)
	documentsCmd.AddCommand(docsSearchCmd)
	documentsCmd.AddCommand(docsGetCmd)
	documentsCmd.AddCommand(docsUploadCmd)
	documentsCmd.AddCommand(docsDownloadCmd)
	documentsCmd.AddCommand(docsEditCmd)
	documentsCmd.AddCommand(docsDeleteCmd)
	documentsCmd.AddCommand(docsContentCmd)

	// List flags
	docsListCmd.Flags().StringVar(&listQuery, "query", "", "search query")
	docsListCmd.Flags().StringArrayVar(&listTags, "tag", nil, "filter by tag (repeatable)")
	docsListCmd.Flags().StringVar(&listCorrespondent, "correspondent", "", "filter by correspondent")
	docsListCmd.Flags().StringVar(&listDocType, "type", "", "filter by document type")
	docsListCmd.Flags().StringVar(&listCreatedAfter, "created-after", "", "filter by creation date (YYYY-MM-DD)")
	docsListCmd.Flags().StringVar(&listCreatedBefore, "created-before", "", "filter by creation date (YYYY-MM-DD)")
	docsListCmd.Flags().IntVar(&listLimit, "limit", 25, "max results")
	docsListCmd.Flags().IntVar(&listPage, "page", 1, "page number")

	// Search flags
	docsSearchCmd.Flags().IntVar(&listLimit, "limit", 25, "max results")

	// Upload flags
	docsUploadCmd.Flags().StringVar(&uploadTitle, "title", "", "document title")
	docsUploadCmd.Flags().StringVar(&uploadCorrespondent, "correspondent", "", "correspondent name or ID")
	docsUploadCmd.Flags().StringVar(&uploadDocType, "type", "", "document type name or ID")
	docsUploadCmd.Flags().StringArrayVar(&uploadTags, "tag", nil, "tag name or ID (repeatable)")

	// Download flags
	docsDownloadCmd.Flags().StringVarP(&downloadOutput, "output", "o", "", "output path")
	docsDownloadCmd.Flags().BoolVar(&downloadOriginal, "original", false, "download original file")

	// Edit flags
	docsEditCmd.Flags().StringVar(&editTitle, "title", "", "new title")
	docsEditCmd.Flags().StringVar(&editCorrespondent, "correspondent", "", "set correspondent")
	docsEditCmd.Flags().StringVar(&editDocType, "type", "", "set document type")
	docsEditCmd.Flags().StringArrayVar(&editAddTags, "add-tag", nil, "add tag (repeatable)")
	docsEditCmd.Flags().StringArrayVar(&editRemoveTags, "remove-tag", nil, "remove tag (repeatable)")
	docsEditCmd.Flags().IntVar(&editASN, "asn", 0, "archive serial number")

	// Delete flags
	docsDeleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "skip confirmation")
}

func runDocsList(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	params := api.DocumentListParams{
		Query:         listQuery,
		Tags:          listTags,
		Correspondent: listCorrespondent,
		DocumentType:  listDocType,
		CreatedAfter:  listCreatedAfter,
		CreatedBefore: listCreatedBefore,
		Limit:         listLimit,
		Page:          listPage,
		Ordering:      "-created",
	}

	result, err := client.ListDocuments(params)
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(result)
	}

	if len(result.Results) == 0 {
		fmt.Println("No documents found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tCREATED\tTAGS")
	for _, doc := range result.Results {
		tagStr := fmt.Sprintf("%d tags", len(doc.Tags))
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", doc.ID, truncate(doc.Title, 40), doc.CreatedDate, tagStr)
	}
	w.Flush()

	if !isQuiet() {
		fmt.Fprintf(os.Stderr, "\nShowing %d of %d documents\n", len(result.Results), result.Count)
	}

	return nil
}

func runDocsSearch(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	params := api.DocumentListParams{
		Query:    args[0],
		Limit:    listLimit,
		Ordering: "-created",
	}

	result, err := client.ListDocuments(params)
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(result)
	}

	if len(result.Results) == 0 {
		fmt.Println("No documents found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tCREATED")
	for _, doc := range result.Results {
		fmt.Fprintf(w, "%d\t%s\t%s\n", doc.ID, truncate(doc.Title, 50), doc.CreatedDate)
	}
	w.Flush()

	if !isQuiet() {
		fmt.Fprintf(os.Stderr, "\nFound %d documents\n", result.Count)
	}

	return nil
}

func runDocsGet(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid document ID: %s", args[0])
	}

	doc, err := client.GetDocument(id)
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(doc)
	}

	fmt.Printf("ID:           %d\n", doc.ID)
	fmt.Printf("Title:        %s\n", doc.Title)
	fmt.Printf("Created:      %s\n", doc.CreatedDate)
	fmt.Printf("Added:        %s\n", doc.Added.Format("2006-01-02 15:04:05"))
	fmt.Printf("Modified:     %s\n", doc.Modified.Format("2006-01-02 15:04:05"))
	fmt.Printf("Original:     %s\n", doc.OriginalFileName)
	if doc.ArchiveSerialNumber != nil {
		fmt.Printf("ASN:          %d\n", *doc.ArchiveSerialNumber)
	}
	if doc.Correspondent != nil {
		fmt.Printf("Correspondent: %d\n", *doc.Correspondent)
	}
	if doc.DocumentType != nil {
		fmt.Printf("Type:         %d\n", *doc.DocumentType)
	}
	if len(doc.Tags) > 0 {
		fmt.Printf("Tags:         %v\n", doc.Tags)
	}

	return nil
}

func runDocsUpload(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	// Resolve correspondent ID
	var correspondentID *int
	if uploadCorrespondent != "" {
		if id, err := strconv.Atoi(uploadCorrespondent); err == nil {
			correspondentID = &id
		} else {
			corr, err := client.FindCorrespondentByName(uploadCorrespondent)
			if err != nil {
				return fmt.Errorf("correspondent not found: %s", uploadCorrespondent)
			}
			correspondentID = &corr.ID
		}
	}

	// Resolve document type ID
	var docTypeID *int
	if uploadDocType != "" {
		if id, err := strconv.Atoi(uploadDocType); err == nil {
			docTypeID = &id
		} else {
			dt, err := client.FindDocumentTypeByName(uploadDocType)
			if err != nil {
				return fmt.Errorf("document type not found: %s", uploadDocType)
			}
			docTypeID = &dt.ID
		}
	}

	// Resolve tag IDs
	var tagIDs []int
	for _, tagArg := range uploadTags {
		if id, err := strconv.Atoi(tagArg); err == nil {
			tagIDs = append(tagIDs, id)
		} else {
			tag, err := client.FindTagByName(tagArg)
			if err != nil {
				return fmt.Errorf("tag not found: %s", tagArg)
			}
			tagIDs = append(tagIDs, tag.ID)
		}
	}

	for _, filePath := range args {
		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", filePath)
		}

		title := uploadTitle
		if title == "" {
			// Use filename without extension as title
			title = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		}

		if !isQuiet() {
			fmt.Fprintf(os.Stderr, "Uploading %s...\n", filepath.Base(filePath))
		}

		taskID, err := client.UploadDocument(filePath, title, correspondentID, docTypeID, tagIDs)
		if err != nil {
			return fmt.Errorf("upload failed for %s: %w", filePath, err)
		}

		if isJSON() {
			printJSON(map[string]string{"file": filePath, "task_id": taskID})
		} else if !isQuiet() {
			fmt.Printf("Uploaded %s (task: %s)\n", filepath.Base(filePath), taskID)
		}
	}

	return nil
}

func runDocsDownload(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid document ID: %s", args[0])
	}

	data, filename, err := client.DownloadDocument(id, downloadOriginal)
	if err != nil {
		return err
	}

	outputPath := downloadOutput
	if outputPath == "" {
		outputPath = filename
		if outputPath == "" {
			outputPath = fmt.Sprintf("document_%d.pdf", id)
		}
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if !isQuiet() {
		fmt.Printf("Downloaded to %s (%d bytes)\n", outputPath, len(data))
	}

	return nil
}

func runDocsEdit(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid document ID: %s", args[0])
	}

	// Get current document to modify tags
	doc, err := client.GetDocument(id)
	if err != nil {
		return err
	}

	updates := make(map[string]interface{})

	if editTitle != "" {
		updates["title"] = editTitle
	}

	if editCorrespondent != "" {
		if editCorrespondent == "-" || editCorrespondent == "none" {
			updates["correspondent"] = nil
		} else if corrID, err := strconv.Atoi(editCorrespondent); err == nil {
			updates["correspondent"] = corrID
		} else {
			corr, err := client.FindCorrespondentByName(editCorrespondent)
			if err != nil {
				return fmt.Errorf("correspondent not found: %s", editCorrespondent)
			}
			updates["correspondent"] = corr.ID
		}
	}

	if editDocType != "" {
		if editDocType == "-" || editDocType == "none" {
			updates["document_type"] = nil
		} else if dtID, err := strconv.Atoi(editDocType); err == nil {
			updates["document_type"] = dtID
		} else {
			dt, err := client.FindDocumentTypeByName(editDocType)
			if err != nil {
				return fmt.Errorf("document type not found: %s", editDocType)
			}
			updates["document_type"] = dt.ID
		}
	}

	if editASN > 0 {
		updates["archive_serial_number"] = editASN
	}

	// Handle tag modifications
	if len(editAddTags) > 0 || len(editRemoveTags) > 0 {
		tags := make(map[int]bool)
		for _, t := range doc.Tags {
			tags[t] = true
		}

		// Add tags
		for _, tagArg := range editAddTags {
			if tagID, err := strconv.Atoi(tagArg); err == nil {
				tags[tagID] = true
			} else {
				tag, err := client.FindTagByName(tagArg)
				if err != nil {
					return fmt.Errorf("tag not found: %s", tagArg)
				}
				tags[tag.ID] = true
			}
		}

		// Remove tags
		for _, tagArg := range editRemoveTags {
			if tagID, err := strconv.Atoi(tagArg); err == nil {
				delete(tags, tagID)
			} else {
				tag, err := client.FindTagByName(tagArg)
				if err != nil {
					// Tag doesn't exist, nothing to remove
					continue
				}
				delete(tags, tag.ID)
			}
		}

		var newTags []int
		for t := range tags {
			newTags = append(newTags, t)
		}
		updates["tags"] = newTags
	}

	if len(updates) == 0 {
		return fmt.Errorf("no changes specified")
	}

	updatedDoc, err := client.UpdateDocument(id, updates)
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(updatedDoc)
	}

	if !isQuiet() {
		fmt.Printf("Updated document %d\n", id)
	}

	return nil
}

func runDocsDelete(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	var ids []int
	for _, arg := range args {
		id, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("invalid document ID: %s", arg)
		}
		ids = append(ids, id)
	}

	if !deleteForce {
		msg := fmt.Sprintf("Delete %d document(s)?", len(ids))
		if !confirmAction(msg) {
			fmt.Println("Cancelled")
			return nil
		}
	}

	for _, id := range ids {
		if err := client.DeleteDocument(id); err != nil {
			return fmt.Errorf("failed to delete document %d: %w", id, err)
		}
		if !isQuiet() {
			fmt.Printf("Deleted document %d\n", id)
		}
	}

	return nil
}

func runDocsContent(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid document ID: %s", args[0])
	}

	doc, err := client.GetDocument(id)
	if err != nil {
		return err
	}

	if isJSON() {
		return printJSON(map[string]string{"id": args[0], "content": doc.Content})
	}

	fmt.Println(doc.Content)
	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
