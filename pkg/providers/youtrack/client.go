package youtrack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"
	"github.com/grik-ai/ricochet-task/pkg/providers"
)

// YouTrackClient handles HTTP communication with YouTrack API
type YouTrackClient struct {
	baseURL     string
	token       string
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	userAgent   string
}

// YouTrackError represents an error from YouTrack API
type YouTrackError struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
}

func (e *YouTrackError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("YouTrack API error %d: %s - %s", e.StatusCode, e.Message, e.Details)
	}
	return fmt.Sprintf("YouTrack API error %d: %s", e.StatusCode, e.Message)
}

// NewYouTrackClient creates a new YouTrack client
func NewYouTrackClient(config *providers.ProviderConfig) (*YouTrackClient, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("YouTrack base URL is required")
	}

	if config.Token == "" {
		return nil, fmt.Errorf("YouTrack token is required")
	}

	// Setup rate limiter
	var rateLimiter *rate.Limiter
	if config.RateLimit != nil {
		rateLimiter = rate.NewLimiter(
			rate.Limit(config.RateLimit.RequestsPerSecond),
			config.RateLimit.BurstSize,
		)
	} else {
		// Default rate limit: 10 requests per second
		rateLimiter = rate.NewLimiter(rate.Limit(10), 20)
	}

	// Setup HTTP client
	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:       100,
			IdleConnTimeout:    90 * time.Second,
			DisableCompression: true,
		},
	}

	client := &YouTrackClient{
		baseURL:     strings.TrimSuffix(config.BaseURL, "/"),
		token:       config.Token,
		httpClient:  httpClient,
		rateLimiter: rateLimiter,
		userAgent:   "ricochet-task/1.0.0",
	}

	return client, nil
}

// CreateIssue creates a new issue in YouTrack
func (c *YouTrackClient) CreateIssue(ctx context.Context, issue *YouTrackIssue) (*YouTrackIssue, error) {
	body, err := json.Marshal(issue)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal issue: %w", err)
	}

	resp, err := c.makeRequest(ctx, "POST", "/api/issues", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.handleErrorResponse(resp)
	}

	var createdIssue YouTrackIssue
	if err := json.NewDecoder(resp.Body).Decode(&createdIssue); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createdIssue, nil
}

// GetIssue retrieves an issue by ID
func (c *YouTrackClient) GetIssue(ctx context.Context, id string) (*YouTrackIssue, error) {
	path := fmt.Sprintf("/api/issues/%s", url.PathEscape(id))
	params := url.Values{
		"fields": {"id,idReadable,summary,description,project(id,name),state(id,name),assignee(id,name),reporter(id,name),priority(id,name),type(id,name),created,updated,resolved,customFields(id,name,value),comments(id,text,author(id,name),created),attachments(id,name,url,size)"},
	}

	resp, err := c.makeRequest(ctx, "GET", path+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &YouTrackError{StatusCode: 404, Message: "Issue not found"}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var issue YouTrackIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &issue, nil
}

// GetIssueByKey retrieves an issue by its key (e.g., PROJ-123)
func (c *YouTrackClient) GetIssueByKey(ctx context.Context, key string) (*YouTrackIssue, error) {
	return c.GetIssue(ctx, key) // YouTrack accepts both ID and key
}

// UpdateIssue updates an existing issue
func (c *YouTrackClient) UpdateIssue(ctx context.Context, id string, updates *YouTrackIssueUpdate) error {
	body, err := json.Marshal(updates)
	if err != nil {
		return fmt.Errorf("failed to marshal updates: %w", err)
	}

	path := fmt.Sprintf("/api/issues/%s", url.PathEscape(id))
	resp, err := c.makeRequest(ctx, "POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &YouTrackError{StatusCode: 404, Message: "Issue not found"}
	}

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// UpdateIssueStatus updates the status of an issue
func (c *YouTrackClient) UpdateIssueStatus(ctx context.Context, issueID, status string) error {
	updates := &YouTrackIssueUpdate{
		State: &YouTrackState{
			Name: status,
		},
	}

	return c.UpdateIssue(ctx, issueID, updates)
}

// DeleteIssue deletes an issue
func (c *YouTrackClient) DeleteIssue(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/issues/%s", url.PathEscape(id))
	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &YouTrackError{StatusCode: 404, Message: "Issue not found"}
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// ListIssues lists issues with filters
func (c *YouTrackClient) ListIssues(ctx context.Context, filters *YouTrackIssueFilters) ([]*YouTrackIssue, error) {
	params := url.Values{
		"fields": {"id,idReadable,summary,description,project(id,name),state(id,name),assignee(id,name),reporter(id,name),priority(id,name),type(id,name),created,updated,resolved"},
	}

	// Build query string from filters
	query := c.buildQueryFromFilters(filters)
	if query != "" {
		params.Set("query", query)
	}

	if filters.Top > 0 {
		params.Set("$top", strconv.Itoa(filters.Top))
	}

	if filters.Skip > 0 {
		params.Set("$skip", strconv.Itoa(filters.Skip))
	}

	resp, err := c.makeRequest(ctx, "GET", "/api/issues?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var issues []*YouTrackIssue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return issues, nil
}

// GetProjectStatuses gets available statuses for a project
func (c *YouTrackClient) GetProjectStatuses(ctx context.Context, projectID string) ([]*YouTrackState, error) {
	path := fmt.Sprintf("/api/admin/projects/%s/states", url.PathEscape(projectID))
	params := url.Values{
		"fields": {"id,name,description,isResolved"},
	}

	resp, err := c.makeRequest(ctx, "GET", path+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var states []*YouTrackState
	if err := json.NewDecoder(resp.Body).Decode(&states); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return states, nil
}

// BulkCreateIssues creates multiple issues in YouTrack
func (c *YouTrackClient) BulkCreateIssues(ctx context.Context, issues []*YouTrackIssue) ([]*YouTrackIssue, error) {
	// YouTrack doesn't have native bulk create, so we create issues one by one
	// In a production implementation, you might want to implement concurrent creation with rate limiting
	createdIssues := make([]*YouTrackIssue, len(issues))

	for i, issue := range issues {
		createdIssue, err := c.CreateIssue(ctx, issue)
		if err != nil {
			return nil, fmt.Errorf("failed to create issue %d: %w", i, err)
		}
		createdIssues[i] = createdIssue
	}

	return createdIssues, nil
}

// BulkUpdateIssues updates multiple issues in YouTrack
func (c *YouTrackClient) BulkUpdateIssues(ctx context.Context, updates map[string]*YouTrackIssueUpdate) error {
	// YouTrack doesn't have native bulk update, so we update issues one by one
	for id, update := range updates {
		if err := c.UpdateIssue(ctx, id, update); err != nil {
			return fmt.Errorf("failed to update issue %s: %w", id, err)
		}
	}

	return nil
}

// AddComment adds a comment to an issue
func (c *YouTrackClient) AddComment(ctx context.Context, issueID string, comment *YouTrackComment) error {
	body, err := json.Marshal(comment)
	if err != nil {
		return fmt.Errorf("failed to marshal comment: %w", err)
	}

	path := fmt.Sprintf("/api/issues/%s/comments", url.PathEscape(issueID))
	resp, err := c.makeRequest(ctx, "POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &YouTrackError{StatusCode: 404, Message: "Issue not found"}
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// GetComments gets comments for an issue
func (c *YouTrackClient) GetComments(ctx context.Context, issueID string) ([]*YouTrackComment, error) {
	path := fmt.Sprintf("/api/issues/%s/comments", url.PathEscape(issueID))
	params := url.Values{
		"fields": {"id,text,author(id,name),created,updated"},
	}

	resp, err := c.makeRequest(ctx, "GET", path+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &YouTrackError{StatusCode: 404, Message: "Issue not found"}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var comments []*YouTrackComment
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return comments, nil
}

// HealthCheck performs a health check by getting server configuration
func (c *YouTrackClient) HealthCheck(ctx context.Context) error {
	resp, err := c.makeRequest(ctx, "GET", "/api/config", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// Close closes the client and cleans up resources
func (c *YouTrackClient) Close() error {
	// Close HTTP client connections
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}

// makeRequest makes an HTTP request to YouTrack API
func (c *YouTrackClient) makeRequest(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
	// Rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Create request
	url := c.baseURL + path
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// handleErrorResponse handles error responses from YouTrack API
func (c *YouTrackClient) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &YouTrackError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to read error response",
		}
	}

	// Try to parse YouTrack error format
	var ytError struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
		ErrorCode        string `json:"error_code"`
		Message          string `json:"message"`
	}

	if err := json.Unmarshal(body, &ytError); err == nil {
		message := ytError.Error
		if message == "" {
			message = ytError.Message
		}
		if message == "" {
			message = resp.Status
		}

		return &YouTrackError{
			StatusCode: resp.StatusCode,
			Message:    message,
			Details:    ytError.ErrorDescription,
		}
	}

	// Fallback to generic error
	return &YouTrackError{
		StatusCode: resp.StatusCode,
		Message:    resp.Status,
		Details:    string(body),
	}
}

// buildQueryFromFilters builds YouTrack query string from filters
func (c *YouTrackClient) buildQueryFromFilters(filters *YouTrackIssueFilters) string {
	var parts []string

	if filters.ProjectID != "" {
		parts = append(parts, fmt.Sprintf("project: %s", filters.ProjectID))
	}

	if filters.State != "" {
		parts = append(parts, fmt.Sprintf("State: {%s}", filters.State))
	}

	if filters.Assignee != "" {
		parts = append(parts, fmt.Sprintf("Assignee: %s", filters.Assignee))
	}

	if filters.Type != "" {
		parts = append(parts, fmt.Sprintf("Type: {%s}", filters.Type))
	}

	if filters.Priority != "" {
		parts = append(parts, fmt.Sprintf("Priority: {%s}", filters.Priority))
	}

	if filters.CreatedAfter != nil {
		parts = append(parts, fmt.Sprintf("created: %s .. *", filters.CreatedAfter.Format("2006-01-02")))
	}

	if filters.CreatedBefore != nil {
		parts = append(parts, fmt.Sprintf("created: * .. %s", filters.CreatedBefore.Format("2006-01-02")))
	}

	if filters.UpdatedAfter != nil {
		parts = append(parts, fmt.Sprintf("updated: %s .. *", filters.UpdatedAfter.Format("2006-01-02")))
	}

	if filters.UpdatedBefore != nil {
		parts = append(parts, fmt.Sprintf("updated: * .. %s", filters.UpdatedBefore.Format("2006-01-02")))
	}

	if filters.Query != "" {
		parts = append(parts, filters.Query)
	}

	return strings.Join(parts, " and ")
}

// Agile Board methods

// GetAgileBoard retrieves a specific agile board
func (c *YouTrackClient) GetAgileBoard(ctx context.Context, boardID string) (*YouTrackBoardInfo, error) {
	path := fmt.Sprintf("/api/agiles/%s?fields=id,name,projects(id,name),created,updated", boardID)
	
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var board YouTrackBoardInfo
	if err := json.NewDecoder(resp.Body).Decode(&board); err != nil {
		return nil, fmt.Errorf("failed to decode board response: %w", err)
	}

	return &board, nil
}

// ListAgileBoards retrieves all agile boards for a project
func (c *YouTrackClient) ListAgileBoards(ctx context.Context, projectID string) ([]*YouTrackBoardInfo, error) {
	path := "/api/agiles?fields=id,name,projects(id,name),created,updated"
	
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var boards []*YouTrackBoardInfo
	if err := json.NewDecoder(resp.Body).Decode(&boards); err != nil {
		return nil, fmt.Errorf("failed to decode boards response: %w", err)
	}

	// Filter by project if specified
	if projectID != "" {
		var filteredBoards []*YouTrackBoardInfo
		for _, board := range boards {
			for _, project := range board.Projects {
				if project.ID == projectID {
					filteredBoards = append(filteredBoards, board)
					break
				}
			}
		}
		return filteredBoards, nil
	}

	return boards, nil
}

// CreateAgileBoard creates a new agile board
func (c *YouTrackClient) CreateAgileBoard(ctx context.Context, request *YouTrackCreateBoardRequest) (*YouTrackBoardInfo, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal create board request: %w", err)
	}

	path := "/api/agiles?fields=id,name,projects(id,name),created,updated"
	
	resp, err := c.makeRequest(ctx, "POST", path, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.handleErrorResponse(resp)
	}

	var board YouTrackBoardInfo
	if err := json.NewDecoder(resp.Body).Decode(&board); err != nil {
		return nil, fmt.Errorf("failed to decode created board response: %w", err)
	}

	return &board, nil
}

// UpdateAgileBoard updates an existing agile board
func (c *YouTrackClient) UpdateAgileBoard(ctx context.Context, boardID string, request *YouTrackUpdateBoardRequest) error {
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal update board request: %w", err)
	}

	path := fmt.Sprintf("/api/agiles/%s", boardID)
	
	resp, err := c.makeRequest(ctx, "POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// DeleteAgileBoard deletes an agile board
func (c *YouTrackClient) DeleteAgileBoard(ctx context.Context, boardID string) error {
	path := fmt.Sprintf("/api/agiles/%s", boardID)
	
	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// GetBoardColumns retrieves columns for a board
func (c *YouTrackClient) GetBoardColumns(ctx context.Context, boardID string) ([]*YouTrackColumnInfo, error) {
	path := fmt.Sprintf("/api/agiles/%s/columns?fields=id,name,presentation(id,name)", boardID)
	
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var columns []*YouTrackColumnInfo
	if err := json.NewDecoder(resp.Body).Decode(&columns); err != nil {
		return nil, fmt.Errorf("failed to decode columns response: %w", err)
	}

	return columns, nil
}

// MoveTaskBetweenColumns moves a task between board columns
func (c *YouTrackClient) MoveTaskBetweenColumns(ctx context.Context, taskID, fromColumn, toColumn string) error {
	// YouTrack uses state changes to move tasks between columns
	// This is a simplified implementation
	updateRequest := map[string]interface{}{
		"customFields": []map[string]interface{}{
			{
				"name":  "State",
				"value": map[string]interface{}{"name": toColumn},
			},
		},
	}

	body, err := json.Marshal(updateRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal move task request: %w", err)
	}

	path := fmt.Sprintf("/api/issues/%s", taskID)
	
	resp, err := c.makeRequest(ctx, "POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	return nil
}