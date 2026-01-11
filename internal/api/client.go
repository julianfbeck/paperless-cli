package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Client is the Paperless API client
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new API client
func NewClient(baseURL, token string) *Client {
	// Ensure baseURL doesn't have trailing slash
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// request makes an authenticated request to the API
func (c *Client) request(method, path string, body io.Reader, contentType string) (*http.Response, error) {
	url := c.baseURL + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Token "+c.token)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", "application/json; version=5")

	return c.httpClient.Do(req)
}

// get makes a GET request
func (c *Client) get(path string) (*http.Response, error) {
	return c.request("GET", path, nil, "")
}

// post makes a POST request with JSON body
func (c *Client) post(path string, data interface{}) (*http.Response, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return c.request("POST", path, bytes.NewReader(body), "application/json")
}

// patch makes a PATCH request with JSON body
func (c *Client) patch(path string, data interface{}) (*http.Response, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return c.request("PATCH", path, bytes.NewReader(body), "application/json")
}

// delete makes a DELETE request
func (c *Client) delete(path string) (*http.Response, error) {
	return c.request("DELETE", path, nil, "")
}

// PaginatedResponse is the generic paginated response
type PaginatedResponse[T any] struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []T    `json:"results"`
	All      []int  `json:"all,omitempty"`
}

// Document represents a Paperless document
type Document struct {
	ID                  int       `json:"id"`
	Correspondent       *int      `json:"correspondent"`
	DocumentType        *int      `json:"document_type"`
	StoragePath         *int      `json:"storage_path"`
	Title               string    `json:"title"`
	Content             string    `json:"content"`
	Tags                []int     `json:"tags"`
	Created             time.Time `json:"created"`
	CreatedDate         string    `json:"created_date"`
	Modified            time.Time `json:"modified"`
	Added               time.Time `json:"added"`
	ArchiveSerialNumber *int      `json:"archive_serial_number"`
	OriginalFileName    string    `json:"original_file_name"`
	ArchivedFileName    string    `json:"archived_file_name"`
}

// Tag represents a Paperless tag
type Tag struct {
	ID             int    `json:"id"`
	Slug           string `json:"slug"`
	Name           string `json:"name"`
	Color          string `json:"color"`
	TextColor      string `json:"text_color"`
	Match          string `json:"match"`
	MatchingAlgo   int    `json:"matching_algorithm"`
	IsInsensitive  bool   `json:"is_insensitive"`
	IsInboxTag     bool   `json:"is_inbox_tag"`
	DocumentCount  int    `json:"document_count"`
}

// Correspondent represents a Paperless correspondent
type Correspondent struct {
	ID              int    `json:"id"`
	Slug            string `json:"slug"`
	Name            string `json:"name"`
	Match           string `json:"match"`
	MatchingAlgo    int    `json:"matching_algorithm"`
	IsInsensitive   bool   `json:"is_insensitive"`
	DocumentCount   int    `json:"document_count"`
	LastCorrespond  string `json:"last_correspondence"`
}

// DocumentType represents a Paperless document type
type DocumentType struct {
	ID            int    `json:"id"`
	Slug          string `json:"slug"`
	Name          string `json:"name"`
	Match         string `json:"match"`
	MatchingAlgo  int    `json:"matching_algorithm"`
	IsInsensitive bool   `json:"is_insensitive"`
	DocumentCount int    `json:"document_count"`
}

// Task represents a Paperless task
type Task struct {
	ID           int    `json:"id"`
	TaskID       string `json:"task_id"`
	TaskFileName string `json:"task_file_name"`
	DateCreated  string `json:"date_created"`
	DateDone     string `json:"date_done"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	Result       string `json:"result"`
	Acknowledged bool   `json:"acknowledged"`
	RelatedDoc   string `json:"related_document"`
}

// DocumentListParams contains parameters for listing documents
type DocumentListParams struct {
	Query         string
	Tags          []string
	Correspondent string
	DocumentType  string
	CreatedAfter  string
	CreatedBefore string
	Limit         int
	Page          int
	Ordering      string
}

// ListDocuments lists documents with optional filters
func (c *Client) ListDocuments(params DocumentListParams) (*PaginatedResponse[Document], error) {
	query := url.Values{}

	if params.Query != "" {
		query.Set("query", params.Query)
	}
	for _, tag := range params.Tags {
		query.Add("tags__name__iexact", tag)
	}
	if params.Correspondent != "" {
		query.Set("correspondent__name__iexact", params.Correspondent)
	}
	if params.DocumentType != "" {
		query.Set("document_type__name__iexact", params.DocumentType)
	}
	if params.CreatedAfter != "" {
		query.Set("created__date__gt", params.CreatedAfter)
	}
	if params.CreatedBefore != "" {
		query.Set("created__date__lt", params.CreatedBefore)
	}
	if params.Limit > 0 {
		query.Set("page_size", strconv.Itoa(params.Limit))
	}
	if params.Page > 0 {
		query.Set("page", strconv.Itoa(params.Page))
	}
	if params.Ordering != "" {
		query.Set("ordering", params.Ordering)
	}

	path := "/api/documents/"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	resp, err := c.get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result PaginatedResponse[Document]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetDocument gets a single document by ID
func (c *Client) GetDocument(id int) (*Document, error) {
	resp, err := c.get(fmt.Sprintf("/api/documents/%d/", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("document %d not found", id)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var doc Document
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}

	return &doc, nil
}

// UploadDocument uploads a document file
func (c *Client) UploadDocument(filePath string, title string, correspondent *int, docType *int, tags []int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the file
	part, err := writer.CreateFormFile("document", filepath.Base(filePath))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}

	// Add optional fields
	if title != "" {
		writer.WriteField("title", title)
	}
	if correspondent != nil {
		writer.WriteField("correspondent", strconv.Itoa(*correspondent))
	}
	if docType != nil {
		writer.WriteField("document_type", strconv.Itoa(*docType))
	}
	for _, tag := range tags {
		writer.WriteField("tags", strconv.Itoa(tag))
	}

	writer.Close()

	resp, err := c.request("POST", "/api/documents/post_document/", body, writer.FormDataContentType())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed %d: %s", resp.StatusCode, string(respBody))
	}

	// The response contains a task ID
	var result string
	respBody, _ := io.ReadAll(resp.Body)
	// Response is just a task UUID string
	result = strings.Trim(string(respBody), "\" \n")
	return result, nil
}

// DownloadDocument downloads a document file
func (c *Client) DownloadDocument(id int, original bool) ([]byte, string, error) {
	path := fmt.Sprintf("/api/documents/%d/download/", id)
	if original {
		path += "?original=true"
	}

	resp, err := c.get(path)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("download failed %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	// Extract filename from Content-Disposition header
	filename := ""
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if idx := strings.Index(cd, "filename="); idx != -1 {
			filename = strings.Trim(cd[idx+9:], "\"")
		}
	}

	return data, filename, nil
}

// UpdateDocument updates a document's metadata
func (c *Client) UpdateDocument(id int, updates map[string]interface{}) (*Document, error) {
	resp, err := c.patch(fmt.Sprintf("/api/documents/%d/", id), updates)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("update failed %d: %s", resp.StatusCode, string(body))
	}

	var doc Document
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}

	return &doc, nil
}

// DeleteDocument deletes a document
func (c *Client) DeleteDocument(id int) error {
	resp, err := c.delete(fmt.Sprintf("/api/documents/%d/", id))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListTags lists all tags
func (c *Client) ListTags() (*PaginatedResponse[Tag], error) {
	resp, err := c.get("/api/tags/?page_size=1000")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result PaginatedResponse[Tag]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTag gets a single tag by ID
func (c *Client) GetTag(id int) (*Tag, error) {
	resp, err := c.get(fmt.Sprintf("/api/tags/%d/", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("tag %d not found", id)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var tag Tag
	if err := json.NewDecoder(resp.Body).Decode(&tag); err != nil {
		return nil, err
	}

	return &tag, nil
}

// CreateTag creates a new tag
func (c *Client) CreateTag(name, color string) (*Tag, error) {
	data := map[string]interface{}{
		"name": name,
	}
	if color != "" {
		data["color"] = color
	}

	resp, err := c.post("/api/tags/", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create failed %d: %s", resp.StatusCode, string(body))
	}

	var tag Tag
	if err := json.NewDecoder(resp.Body).Decode(&tag); err != nil {
		return nil, err
	}

	return &tag, nil
}

// UpdateTag updates a tag
func (c *Client) UpdateTag(id int, updates map[string]interface{}) (*Tag, error) {
	resp, err := c.patch(fmt.Sprintf("/api/tags/%d/", id), updates)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("update failed %d: %s", resp.StatusCode, string(body))
	}

	var tag Tag
	if err := json.NewDecoder(resp.Body).Decode(&tag); err != nil {
		return nil, err
	}

	return &tag, nil
}

// DeleteTag deletes a tag
func (c *Client) DeleteTag(id int) error {
	resp, err := c.delete(fmt.Sprintf("/api/tags/%d/", id))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListCorrespondents lists all correspondents
func (c *Client) ListCorrespondents() (*PaginatedResponse[Correspondent], error) {
	resp, err := c.get("/api/correspondents/?page_size=1000")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result PaginatedResponse[Correspondent]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetCorrespondent gets a single correspondent by ID
func (c *Client) GetCorrespondent(id int) (*Correspondent, error) {
	resp, err := c.get(fmt.Sprintf("/api/correspondents/%d/", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("correspondent %d not found", id)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var corr Correspondent
	if err := json.NewDecoder(resp.Body).Decode(&corr); err != nil {
		return nil, err
	}

	return &corr, nil
}

// CreateCorrespondent creates a new correspondent
func (c *Client) CreateCorrespondent(name string) (*Correspondent, error) {
	data := map[string]interface{}{
		"name": name,
	}

	resp, err := c.post("/api/correspondents/", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create failed %d: %s", resp.StatusCode, string(body))
	}

	var corr Correspondent
	if err := json.NewDecoder(resp.Body).Decode(&corr); err != nil {
		return nil, err
	}

	return &corr, nil
}

// UpdateCorrespondent updates a correspondent
func (c *Client) UpdateCorrespondent(id int, updates map[string]interface{}) (*Correspondent, error) {
	resp, err := c.patch(fmt.Sprintf("/api/correspondents/%d/", id), updates)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("update failed %d: %s", resp.StatusCode, string(body))
	}

	var corr Correspondent
	if err := json.NewDecoder(resp.Body).Decode(&corr); err != nil {
		return nil, err
	}

	return &corr, nil
}

// DeleteCorrespondent deletes a correspondent
func (c *Client) DeleteCorrespondent(id int) error {
	resp, err := c.delete(fmt.Sprintf("/api/correspondents/%d/", id))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListDocumentTypes lists all document types
func (c *Client) ListDocumentTypes() (*PaginatedResponse[DocumentType], error) {
	resp, err := c.get("/api/document_types/?page_size=1000")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result PaginatedResponse[DocumentType]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetDocumentType gets a single document type by ID
func (c *Client) GetDocumentType(id int) (*DocumentType, error) {
	resp, err := c.get(fmt.Sprintf("/api/document_types/%d/", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("document type %d not found", id)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var dt DocumentType
	if err := json.NewDecoder(resp.Body).Decode(&dt); err != nil {
		return nil, err
	}

	return &dt, nil
}

// CreateDocumentType creates a new document type
func (c *Client) CreateDocumentType(name string) (*DocumentType, error) {
	data := map[string]interface{}{
		"name": name,
	}

	resp, err := c.post("/api/document_types/", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create failed %d: %s", resp.StatusCode, string(body))
	}

	var dt DocumentType
	if err := json.NewDecoder(resp.Body).Decode(&dt); err != nil {
		return nil, err
	}

	return &dt, nil
}

// UpdateDocumentType updates a document type
func (c *Client) UpdateDocumentType(id int, updates map[string]interface{}) (*DocumentType, error) {
	resp, err := c.patch(fmt.Sprintf("/api/document_types/%d/", id), updates)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("update failed %d: %s", resp.StatusCode, string(body))
	}

	var dt DocumentType
	if err := json.NewDecoder(resp.Body).Decode(&dt); err != nil {
		return nil, err
	}

	return &dt, nil
}

// DeleteDocumentType deletes a document type
func (c *Client) DeleteDocumentType(id int) error {
	resp, err := c.delete(fmt.Sprintf("/api/document_types/%d/", id))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetTask gets a task by ID
func (c *Client) GetTask(taskID string) (*Task, error) {
	resp, err := c.get(fmt.Sprintf("/api/tasks/?task_id=%s", taskID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var tasks []Task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	return &tasks[0], nil
}

// FindTagByName finds a tag by name
func (c *Client) FindTagByName(name string) (*Tag, error) {
	tags, err := c.ListTags()
	if err != nil {
		return nil, err
	}
	for _, tag := range tags.Results {
		if strings.EqualFold(tag.Name, name) {
			return &tag, nil
		}
	}
	return nil, fmt.Errorf("tag not found: %s", name)
}

// FindCorrespondentByName finds a correspondent by name
func (c *Client) FindCorrespondentByName(name string) (*Correspondent, error) {
	corrs, err := c.ListCorrespondents()
	if err != nil {
		return nil, err
	}
	for _, corr := range corrs.Results {
		if strings.EqualFold(corr.Name, name) {
			return &corr, nil
		}
	}
	return nil, fmt.Errorf("correspondent not found: %s", name)
}

// FindDocumentTypeByName finds a document type by name
func (c *Client) FindDocumentTypeByName(name string) (*DocumentType, error) {
	types, err := c.ListDocumentTypes()
	if err != nil {
		return nil, err
	}
	for _, dt := range types.Results {
		if strings.EqualFold(dt.Name, name) {
			return &dt, nil
		}
	}
	return nil, fmt.Errorf("document type not found: %s", name)
}

// StoragePath represents a Paperless storage path
type StoragePath struct {
	ID            int    `json:"id"`
	Slug          string `json:"slug"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	Match         string `json:"match"`
	MatchingAlgo  int    `json:"matching_algorithm"`
	IsInsensitive bool   `json:"is_insensitive"`
	DocumentCount int    `json:"document_count"`
}

// SavedView represents a Paperless saved view
type SavedView struct {
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	ShowOnDashboard    bool   `json:"show_on_dashboard"`
	ShowInSidebar      bool   `json:"show_in_sidebar"`
	SortField          string `json:"sort_field"`
	SortReverse        bool   `json:"sort_reverse"`
	FilterRules        []any  `json:"filter_rules"`
}

// GlobalSearchResult represents results from global search
type GlobalSearchResult struct {
	Documents      []Document      `json:"documents"`
	SavedViews     []SavedView     `json:"saved_views"`
	Correspondents []Correspondent `json:"correspondents"`
	DocumentTypes  []DocumentType  `json:"document_types"`
	StoragePaths   []StoragePath   `json:"storage_paths"`
	Tags           []Tag           `json:"tags"`
}

// ListStoragePaths lists all storage paths
func (c *Client) ListStoragePaths() (*PaginatedResponse[StoragePath], error) {
	resp, err := c.get("/api/storage_paths/?page_size=1000")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result PaginatedResponse[StoragePath]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetStoragePath gets a single storage path by ID
func (c *Client) GetStoragePath(id int) (*StoragePath, error) {
	resp, err := c.get(fmt.Sprintf("/api/storage_paths/%d/", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("storage path %d not found", id)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var sp StoragePath
	if err := json.NewDecoder(resp.Body).Decode(&sp); err != nil {
		return nil, err
	}

	return &sp, nil
}

// CreateStoragePath creates a new storage path
func (c *Client) CreateStoragePath(name, path string) (*StoragePath, error) {
	data := map[string]interface{}{
		"name": name,
		"path": path,
	}

	resp, err := c.post("/api/storage_paths/", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create failed %d: %s", resp.StatusCode, string(body))
	}

	var sp StoragePath
	if err := json.NewDecoder(resp.Body).Decode(&sp); err != nil {
		return nil, err
	}

	return &sp, nil
}

// DeleteStoragePath deletes a storage path
func (c *Client) DeleteStoragePath(id int) error {
	resp, err := c.delete(fmt.Sprintf("/api/storage_paths/%d/", id))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListSavedViews lists all saved views
func (c *Client) ListSavedViews() (*PaginatedResponse[SavedView], error) {
	resp, err := c.get("/api/saved_views/?page_size=1000")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result PaginatedResponse[SavedView]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetSavedView gets a single saved view by ID
func (c *Client) GetSavedView(id int) (*SavedView, error) {
	resp, err := c.get(fmt.Sprintf("/api/saved_views/%d/", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("saved view %d not found", id)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var sv SavedView
	if err := json.NewDecoder(resp.Body).Decode(&sv); err != nil {
		return nil, err
	}

	return &sv, nil
}

// GlobalSearch performs a global search across all objects
func (c *Client) GlobalSearch(query string) (*GlobalSearchResult, error) {
	resp, err := c.get(fmt.Sprintf("/api/search/?query=%s", url.QueryEscape(query)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result GlobalSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetSimilarDocuments finds documents similar to the given one
func (c *Client) GetSimilarDocuments(docID int, limit int) (*PaginatedResponse[Document], error) {
	path := fmt.Sprintf("/api/documents/?more_like_id=%d", docID)
	if limit > 0 {
		path += fmt.Sprintf("&page_size=%d", limit)
	}

	resp, err := c.get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result PaginatedResponse[Document]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetDocumentPreview gets the preview/thumbnail URL of a document
func (c *Client) GetDocumentPreview(id int) ([]byte, error) {
	resp, err := c.get(fmt.Sprintf("/api/documents/%d/preview/", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("preview failed %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// GetDocumentThumb gets the thumbnail of a document
func (c *Client) GetDocumentThumb(id int) ([]byte, error) {
	resp, err := c.get(fmt.Sprintf("/api/documents/%d/thumb/", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("thumbnail failed %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// GetStatistics gets system statistics
func (c *Client) GetStatistics() (map[string]any, error) {
	resp, err := c.get("/api/statistics/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// FindStoragePathByName finds a storage path by name
func (c *Client) FindStoragePathByName(name string) (*StoragePath, error) {
	paths, err := c.ListStoragePaths()
	if err != nil {
		return nil, err
	}
	for _, sp := range paths.Results {
		if strings.EqualFold(sp.Name, name) {
			return &sp, nil
		}
	}
	return nil, fmt.Errorf("storage path not found: %s", name)
}
