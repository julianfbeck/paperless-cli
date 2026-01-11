// +build local

package api

import (
	"os"
	"testing"
)

// These tests require PAPERLESS_URL and PAPERLESS_TOKEN environment variables
// Run with: go test -tags=local ./...

func getTestClient(t *testing.T) *Client {
	url := os.Getenv("PAPERLESS_URL")
	token := os.Getenv("PAPERLESS_TOKEN")

	if url == "" || token == "" {
		t.Skip("PAPERLESS_URL and PAPERLESS_TOKEN must be set for integration tests")
	}

	return NewClient(url, token)
}

func TestListDocuments(t *testing.T) {
	client := getTestClient(t)

	result, err := client.ListDocuments(DocumentListParams{Limit: 5})
	if err != nil {
		t.Fatalf("ListDocuments failed: %v", err)
	}

	t.Logf("Found %d documents (showing %d)", result.Count, len(result.Results))
	for _, doc := range result.Results {
		t.Logf("  - [%d] %s", doc.ID, doc.Title)
	}
}

func TestListTags(t *testing.T) {
	client := getTestClient(t)

	result, err := client.ListTags()
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	t.Logf("Found %d tags", len(result.Results))
	for _, tag := range result.Results {
		t.Logf("  - [%d] %s (%d docs)", tag.ID, tag.Name, tag.DocumentCount)
	}
}

func TestListCorrespondents(t *testing.T) {
	client := getTestClient(t)

	result, err := client.ListCorrespondents()
	if err != nil {
		t.Fatalf("ListCorrespondents failed: %v", err)
	}

	t.Logf("Found %d correspondents", len(result.Results))
	for _, corr := range result.Results {
		t.Logf("  - [%d] %s (%d docs)", corr.ID, corr.Name, corr.DocumentCount)
	}
}

func TestListDocumentTypes(t *testing.T) {
	client := getTestClient(t)

	result, err := client.ListDocumentTypes()
	if err != nil {
		t.Fatalf("ListDocumentTypes failed: %v", err)
	}

	t.Logf("Found %d document types", len(result.Results))
	for _, dt := range result.Results {
		t.Logf("  - [%d] %s (%d docs)", dt.ID, dt.Name, dt.DocumentCount)
	}
}

func TestSearchDocuments(t *testing.T) {
	client := getTestClient(t)

	// Search for a common term
	result, err := client.ListDocuments(DocumentListParams{
		Query: "test",
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	t.Logf("Search returned %d results", result.Count)
}

func TestGetDocument(t *testing.T) {
	client := getTestClient(t)

	// First get a list to find a valid ID
	result, err := client.ListDocuments(DocumentListParams{Limit: 1})
	if err != nil {
		t.Fatalf("ListDocuments failed: %v", err)
	}

	if len(result.Results) == 0 {
		t.Skip("No documents available for testing")
	}

	doc, err := client.GetDocument(result.Results[0].ID)
	if err != nil {
		t.Fatalf("GetDocument failed: %v", err)
	}

	t.Logf("Document: %s (ID: %d)", doc.Title, doc.ID)
	t.Logf("  Created: %s", doc.CreatedDate)
	t.Logf("  Content length: %d chars", len(doc.Content))
}

func TestUploadAndDownload(t *testing.T) {
	client := getTestClient(t)

	// Create a test file
	testFile := "../../testdata/test_upload.pdf"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test PDF not found at testdata/test_upload.pdf")
	}

	// Upload the test file
	taskID, err := client.UploadDocument(testFile, "CLI Test Upload", nil, nil, nil)
	if err != nil {
		t.Fatalf("UploadDocument failed: %v", err)
	}

	t.Logf("Upload task ID: %s", taskID)

	// Note: The document won't be immediately available, so we just verify the upload was accepted
	// The task status can be checked with GetTask
}
