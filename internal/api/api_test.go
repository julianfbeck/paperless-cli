//go:build local

package api

import (
	"os"
	"testing"
)

// These tests require PAPERLESS_URL and PAPERLESS_TOKEN environment variables
// Run with: go test -tags=local -v ./internal/api/

func getTestClient(t *testing.T) *Client {
	url := os.Getenv("PAPERLESS_URL")
	token := os.Getenv("PAPERLESS_TOKEN")

	if url == "" || token == "" {
		t.Skip("PAPERLESS_URL and PAPERLESS_TOKEN must be set for integration tests")
	}

	return NewClient(url, token)
}

// ==================== Document Tests ====================

func TestListDocuments(t *testing.T) {
	client := getTestClient(t)

	result, err := client.ListDocuments(DocumentListParams{Limit: 5})
	if err != nil {
		t.Fatalf("ListDocuments failed: %v", err)
	}

	if result.Count == 0 {
		t.Log("Warning: No documents found in instance")
	}

	t.Logf("Found %d documents (showing %d)", result.Count, len(result.Results))
	for _, doc := range result.Results {
		t.Logf("  - [%d] %s (%s)", doc.ID, doc.Title, doc.CreatedDate)
	}
}

func TestSearchDocuments(t *testing.T) {
	client := getTestClient(t)

	result, err := client.ListDocuments(DocumentListParams{
		Query: "test",
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	t.Logf("Search for 'test' returned %d results", result.Count)
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

	docID := result.Results[0].ID
	doc, err := client.GetDocument(docID)
	if err != nil {
		t.Fatalf("GetDocument failed: %v", err)
	}

	t.Logf("Document %d: %s", doc.ID, doc.Title)
	t.Logf("  Created: %s", doc.CreatedDate)
	t.Logf("  Original: %s", doc.OriginalFileName)
	t.Logf("  Content length: %d chars", len(doc.Content))
}

func TestGetSimilarDocuments(t *testing.T) {
	client := getTestClient(t)

	// First get a document ID
	result, err := client.ListDocuments(DocumentListParams{Limit: 1})
	if err != nil {
		t.Fatalf("ListDocuments failed: %v", err)
	}

	if len(result.Results) == 0 {
		t.Skip("No documents available for testing")
	}

	docID := result.Results[0].ID
	similar, err := client.GetSimilarDocuments(docID, 5)
	if err != nil {
		t.Fatalf("GetSimilarDocuments failed: %v", err)
	}

	t.Logf("Found %d similar documents to %d", len(similar.Results), docID)
	for _, doc := range similar.Results {
		t.Logf("  - [%d] %s", doc.ID, doc.Title)
	}
}

func TestDownloadDocument(t *testing.T) {
	client := getTestClient(t)

	// First get a document ID
	result, err := client.ListDocuments(DocumentListParams{Limit: 1})
	if err != nil {
		t.Fatalf("ListDocuments failed: %v", err)
	}

	if len(result.Results) == 0 {
		t.Skip("No documents available for testing")
	}

	docID := result.Results[0].ID
	data, filename, err := client.DownloadDocument(docID, false)
	if err != nil {
		t.Fatalf("DownloadDocument failed: %v", err)
	}

	t.Logf("Downloaded document %d: %s (%d bytes)", docID, filename, len(data))
}

func TestGetDocumentThumb(t *testing.T) {
	client := getTestClient(t)

	// First get a document ID
	result, err := client.ListDocuments(DocumentListParams{Limit: 1})
	if err != nil {
		t.Fatalf("ListDocuments failed: %v", err)
	}

	if len(result.Results) == 0 {
		t.Skip("No documents available for testing")
	}

	docID := result.Results[0].ID
	data, err := client.GetDocumentThumb(docID)
	if err != nil {
		t.Fatalf("GetDocumentThumb failed: %v", err)
	}

	t.Logf("Got thumbnail for document %d: %d bytes", docID, len(data))
}

// ==================== Upload Test ====================

func TestUploadDocument(t *testing.T) {
	client := getTestClient(t)

	testFile := "../../testdata/test_upload.pdf"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Test PDF not found at testdata/test_upload.pdf")
	}

	taskID, err := client.UploadDocument(testFile, "API Test Upload", nil, nil, nil)
	if err != nil {
		t.Fatalf("UploadDocument failed: %v", err)
	}

	t.Logf("Upload task ID: %s", taskID)

	// Check task status
	task, err := client.GetTask(taskID)
	if err != nil {
		t.Logf("Warning: Could not get task status: %v", err)
	} else {
		t.Logf("Task status: %s", task.Status)
	}
}

// ==================== Tag Tests ====================

func TestListTags(t *testing.T) {
	client := getTestClient(t)

	result, err := client.ListTags()
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	t.Logf("Found %d tags", len(result.Results))
	for _, tag := range result.Results {
		t.Logf("  - [%d] %s (color: %s, docs: %d)", tag.ID, tag.Name, tag.Color, tag.DocumentCount)
	}
}

func TestCreateAndDeleteTag(t *testing.T) {
	client := getTestClient(t)

	// Create a test tag
	tag, err := client.CreateTag("test-cli-tag", "#ff0000")
	if err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}

	t.Logf("Created tag: [%d] %s", tag.ID, tag.Name)

	// Get the tag
	gotTag, err := client.GetTag(tag.ID)
	if err != nil {
		t.Fatalf("GetTag failed: %v", err)
	}
	if gotTag.Name != "test-cli-tag" {
		t.Errorf("Tag name mismatch: got %s, want test-cli-tag", gotTag.Name)
	}

	// Delete the tag
	err = client.DeleteTag(tag.ID)
	if err != nil {
		t.Fatalf("DeleteTag failed: %v", err)
	}

	t.Logf("Deleted tag %d", tag.ID)
}

// ==================== Correspondent Tests ====================

func TestListCorrespondents(t *testing.T) {
	client := getTestClient(t)

	result, err := client.ListCorrespondents()
	if err != nil {
		t.Fatalf("ListCorrespondents failed: %v", err)
	}

	t.Logf("Found %d correspondents", len(result.Results))
	for _, corr := range result.Results {
		t.Logf("  - [%d] %s (docs: %d)", corr.ID, corr.Name, corr.DocumentCount)
	}
}

func TestCreateAndDeleteCorrespondent(t *testing.T) {
	client := getTestClient(t)

	// Create a test correspondent
	corr, err := client.CreateCorrespondent("Test CLI Correspondent")
	if err != nil {
		t.Fatalf("CreateCorrespondent failed: %v", err)
	}

	t.Logf("Created correspondent: [%d] %s", corr.ID, corr.Name)

	// Get the correspondent
	gotCorr, err := client.GetCorrespondent(corr.ID)
	if err != nil {
		t.Fatalf("GetCorrespondent failed: %v", err)
	}
	if gotCorr.Name != "Test CLI Correspondent" {
		t.Errorf("Correspondent name mismatch: got %s", gotCorr.Name)
	}

	// Delete the correspondent
	err = client.DeleteCorrespondent(corr.ID)
	if err != nil {
		t.Fatalf("DeleteCorrespondent failed: %v", err)
	}

	t.Logf("Deleted correspondent %d", corr.ID)
}

// ==================== Document Type Tests ====================

func TestListDocumentTypes(t *testing.T) {
	client := getTestClient(t)

	result, err := client.ListDocumentTypes()
	if err != nil {
		t.Fatalf("ListDocumentTypes failed: %v", err)
	}

	t.Logf("Found %d document types", len(result.Results))
	for _, dt := range result.Results {
		t.Logf("  - [%d] %s (docs: %d)", dt.ID, dt.Name, dt.DocumentCount)
	}
}

func TestCreateAndDeleteDocumentType(t *testing.T) {
	client := getTestClient(t)

	// Create a test document type
	dt, err := client.CreateDocumentType("Test CLI DocType")
	if err != nil {
		t.Fatalf("CreateDocumentType failed: %v", err)
	}

	t.Logf("Created document type: [%d] %s", dt.ID, dt.Name)

	// Get the document type
	gotDT, err := client.GetDocumentType(dt.ID)
	if err != nil {
		t.Fatalf("GetDocumentType failed: %v", err)
	}
	if gotDT.Name != "Test CLI DocType" {
		t.Errorf("Document type name mismatch: got %s", gotDT.Name)
	}

	// Delete the document type
	err = client.DeleteDocumentType(dt.ID)
	if err != nil {
		t.Fatalf("DeleteDocumentType failed: %v", err)
	}

	t.Logf("Deleted document type %d", dt.ID)
}

// ==================== Storage Path Tests ====================

func TestListStoragePaths(t *testing.T) {
	client := getTestClient(t)

	result, err := client.ListStoragePaths()
	if err != nil {
		t.Fatalf("ListStoragePaths failed: %v", err)
	}

	t.Logf("Found %d storage paths", len(result.Results))
	for _, sp := range result.Results {
		t.Logf("  - [%d] %s: %s (docs: %d)", sp.ID, sp.Name, sp.Path, sp.DocumentCount)
	}
}

func TestCreateAndDeleteStoragePath(t *testing.T) {
	client := getTestClient(t)

	// Create a test storage path
	sp, err := client.CreateStoragePath("Test CLI Path", "test/{{ created_year }}")
	if err != nil {
		t.Fatalf("CreateStoragePath failed: %v", err)
	}

	t.Logf("Created storage path: [%d] %s", sp.ID, sp.Name)

	// Get the storage path
	gotSP, err := client.GetStoragePath(sp.ID)
	if err != nil {
		t.Fatalf("GetStoragePath failed: %v", err)
	}
	if gotSP.Name != "Test CLI Path" {
		t.Errorf("Storage path name mismatch: got %s", gotSP.Name)
	}

	// Delete the storage path
	err = client.DeleteStoragePath(sp.ID)
	if err != nil {
		t.Fatalf("DeleteStoragePath failed: %v", err)
	}

	t.Logf("Deleted storage path %d", sp.ID)
}

// ==================== Saved View Tests ====================

func TestListSavedViews(t *testing.T) {
	client := getTestClient(t)

	result, err := client.ListSavedViews()
	if err != nil {
		t.Fatalf("ListSavedViews failed: %v", err)
	}

	t.Logf("Found %d saved views", len(result.Results))
	for _, sv := range result.Results {
		t.Logf("  - [%d] %s (dashboard: %t, sidebar: %t)", sv.ID, sv.Name, sv.ShowOnDashboard, sv.ShowInSidebar)
	}
}

// ==================== Statistics Test ====================

func TestGetStatistics(t *testing.T) {
	client := getTestClient(t)

	stats, err := client.GetStatistics()
	if err != nil {
		t.Fatalf("GetStatistics failed: %v", err)
	}

	t.Log("Statistics:")
	for key, value := range stats {
		t.Logf("  %s: %v", key, value)
	}
}

// ==================== Task Tests ====================

func TestGetTask(t *testing.T) {
	// This test requires a known task ID - skip if we don't have one
	// Tasks are created during upload, so we'd need to upload first
	t.Skip("Task test requires a valid task ID from a recent upload")
}

// ==================== Find By Name Tests ====================

func TestFindByName(t *testing.T) {
	client := getTestClient(t)

	// Test finding non-existent tag
	_, err := client.FindTagByName("nonexistent-tag-12345")
	if err == nil {
		t.Error("Expected error for non-existent tag")
	}

	// Test finding non-existent correspondent
	_, err = client.FindCorrespondentByName("nonexistent-correspondent-12345")
	if err == nil {
		t.Error("Expected error for non-existent correspondent")
	}

	// Test finding non-existent document type
	_, err = client.FindDocumentTypeByName("nonexistent-doctype-12345")
	if err == nil {
		t.Error("Expected error for non-existent document type")
	}

	// Test finding non-existent storage path
	_, err = client.FindStoragePathByName("nonexistent-path-12345")
	if err == nil {
		t.Error("Expected error for non-existent storage path")
	}

	t.Log("FindByName tests passed - all returned expected errors for non-existent items")
}
