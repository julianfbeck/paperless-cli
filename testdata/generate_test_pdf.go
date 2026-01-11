//go:build ignore

package main

import (
	"log"
	"time"

	"github.com/jung-kurt/gofpdf"
)

func main() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set font
	pdf.SetFont("Arial", "B", 16)

	// Title
	pdf.Cell(190, 10, "Paperless CLI Test Document")
	pdf.Ln(15)

	// Body text
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(190, 7, "This is a test document for the Paperless CLI.\n\n"+
		"Created on: "+time.Now().Format("2006-01-02 15:04:05")+"\n\n"+
		"This document is used to test the upload and download functionality "+
		"of the paperless-cli tool. It contains some sample text that can be "+
		"extracted and verified by the PDF reading functionality.\n\n"+
		"Test Keywords: invoice, receipt, contract, important, paperless, cli, test", "", "", false)

	pdf.Ln(10)

	// Add a table
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(60, 10, "Item")
	pdf.Cell(60, 10, "Description")
	pdf.Cell(40, 10, "Value")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	items := [][]string{
		{"Test Item 1", "First test entry", "$100.00"},
		{"Test Item 2", "Second test entry", "$250.00"},
		{"Test Item 3", "Third test entry", "$75.50"},
	}

	for _, item := range items {
		pdf.Cell(60, 8, item[0])
		pdf.Cell(60, 8, item[1])
		pdf.Cell(40, 8, item[2])
		pdf.Ln(8)
	}

	// Save
	err := pdf.OutputFileAndClose("test_upload.pdf")
	if err != nil {
		log.Fatalf("Failed to create PDF: %v", err)
	}

	log.Println("Created testdata/test_upload.pdf")
}
