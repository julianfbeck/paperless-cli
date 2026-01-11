package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/spf13/cobra"
)

var pdfCmd = &cobra.Command{
	Use:   "pdf",
	Short: "PDF utilities",
	Long:  `Local PDF utilities for reading and extracting text.`,
}

var pdfReadCmd = &cobra.Command{
	Use:   "read <file>",
	Short: "Extract text from a PDF",
	Long: `Extract and display text content from a local PDF file.

Example:
  paperless pdf read document.pdf
  paperless pdf read invoice.pdf --json`,
	Args: cobra.ExactArgs(1),
	RunE: runPDFRead,
}

var pdfInfoCmd = &cobra.Command{
	Use:   "info <file>",
	Short: "Show PDF information",
	Long: `Show metadata and information about a PDF file.

Example:
  paperless pdf info document.pdf`,
	Args: cobra.ExactArgs(1),
	RunE: runPDFInfo,
}

func init() {
	rootCmd.AddCommand(pdfCmd)
	pdfCmd.AddCommand(pdfReadCmd)
	pdfCmd.AddCommand(pdfInfoCmd)
}

func runPDFRead(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	content, err := extractPDFText(filePath)
	if err != nil {
		return fmt.Errorf("failed to read PDF: %w", err)
	}

	if isJSON() {
		return printJSON(map[string]string{
			"file":    filePath,
			"content": content,
		})
	}

	fmt.Println(content)
	return nil
}

func runPDFInfo(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// Check if file exists
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	f, r, err := pdf.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	numPages := r.NumPage()

	if isJSON() {
		return printJSON(map[string]interface{}{
			"file":       filePath,
			"size_bytes": info.Size(),
			"pages":      numPages,
		})
	}

	fmt.Printf("File:   %s\n", filePath)
	fmt.Printf("Size:   %d bytes\n", info.Size())
	fmt.Printf("Pages:  %d\n", numPages)

	return nil
}

// extractPDFText extracts text content from a PDF file
func extractPDFText(filePath string) (string, error) {
	f, r, err := pdf.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var textBuilder strings.Builder
	numPages := r.NumPage()

	for pageNum := 1; pageNum <= numPages; pageNum++ {
		page := r.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			// Skip pages with errors
			continue
		}

		if textBuilder.Len() > 0 {
			textBuilder.WriteString("\n\n--- Page " + fmt.Sprintf("%d", pageNum) + " ---\n\n")
		}
		textBuilder.WriteString(text)
	}

	return textBuilder.String(), nil
}
