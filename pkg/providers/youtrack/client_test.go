package youtrack

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grik-ai/ricochet-task/pkg/providers"
)

// TestNewYouTrackClient tests client creation
func TestNewYouTrackClient(t *testing.T) {
	t.Run("Valid configuration", func(t *testing.T) {
		config := &providers.ProviderConfig{
			BaseURL: "https://test.youtrack.cloud",
			Token:   "test-token",
			Timeout: 30 * time.Second,
			RateLimit: &providers.RateLimitConfig{
				RequestsPerSecond: 5,
				BurstSize:         10,
			},
		}

		client, err := NewYouTrackClient(config)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "https://test.youtrack.cloud", client.baseURL)
		assert.Equal(t, "test-token", client.token)
		assert.Equal(t, "ricochet-task/1.0.0", client.userAgent)
	})

	t.Run("Missing base URL", func(t *testing.T) {
		config := &providers.ProviderConfig{
			Token: "test-token",
		}

		client, err := NewYouTrackClient(config)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "base URL")
	})

	t.Run("Missing token", func(t *testing.T) {
		config := &providers.ProviderConfig{
			BaseURL: "https://test.youtrack.cloud",
		}

		client, err := NewYouTrackClient(config)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "token")
	})

	t.Run("Default rate limit", func(t *testing.T) {
		config := &providers.ProviderConfig{
			BaseURL: "https://test.youtrack.cloud",
			Token:   "test-token",
			// No rate limit config
		}

		client, err := NewYouTrackClient(config)
		require.NoError(t, err)
		assert.NotNil(t, client.rateLimiter)
	})

	t.Run("Trailing slash removal", func(t *testing.T) {
		config := &providers.ProviderConfig{
			BaseURL: "https://test.youtrack.cloud/",
			Token:   "test-token",
		}

		client, err := NewYouTrackClient(config)
		require.NoError(t, err)
		assert.Equal(t, "https://test.youtrack.cloud", client.baseURL)
	})
}

// TestCreateIssue tests issue creation
func TestCreateIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/issues", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "ricochet-task/1.0.0", r.Header.Get("User-Agent"))

		// Parse request body
		var issue YouTrackIssue
		err := json.NewDecoder(r.Body).Decode(&issue)
		require.NoError(t, err)

		// Return created issue
		issue.ID = "created-123"
		issue.IDReadable = "PROJ-123"
		issue.Created = time.Now().Unix() * 1000
		issue.Updated = time.Now().Unix() * 1000

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(issue)
	}))
	defer server.Close()

	config := &providers.ProviderConfig{
		BaseURL: server.URL,
		Token:   "test-token",
	}

	client, err := NewYouTrackClient(config)
	require.NoError(t, err)

	t.Run("Successful creation", func(t *testing.T) {
		issue := &YouTrackIssue{
			Summary:     "Test Issue",
			Description: "Test Description",
		}

		ctx := context.Background()
		createdIssue, err := client.CreateIssue(ctx, issue)

		assert.NoError(t, err)
		assert.NotNil(t, createdIssue)
		assert.Equal(t, "created-123", createdIssue.ID)
		assert.Equal(t, "PROJ-123", createdIssue.IDReadable)
		assert.Equal(t, "Test Issue", createdIssue.Summary)
	})
}

// TestGetIssue tests issue retrieval
func TestGetIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.True(t, strings.HasPrefix(r.URL.Path, "/api/issues/"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		issueID := strings.TrimPrefix(r.URL.Path, "/api/issues/")

		if issueID == "not-found" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Issue not found"})
			return
		}

		issue := YouTrackIssue{
			ID:         issueID,
			IDReadable: "PROJ-" + issueID,
			Summary:    "Test Issue",
			Description: "Test Description",
			Created:    time.Now().Add(-time.Hour).Unix() * 1000,
			Updated:    time.Now().Unix() * 1000,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(issue)
	}))
	defer server.Close()

	config := &providers.ProviderConfig{
		BaseURL: server.URL,
		Token:   "test-token",
	}

	client, err := NewYouTrackClient(config)
	require.NoError(t, err)

	t.Run("Existing issue", func(t *testing.T) {
		ctx := context.Background()
		issue, err := client.GetIssue(ctx, "test-123")

		assert.NoError(t, err)
		assert.NotNil(t, issue)
		assert.Equal(t, "test-123", issue.ID)
		assert.Equal(t, "PROJ-test-123", issue.IDReadable)
		assert.Equal(t, "Test Issue", issue.Summary)
	})

	t.Run("Non-existing issue", func(t *testing.T) {
		ctx := context.Background()
		issue, err := client.GetIssue(ctx, "not-found")

		assert.Error(t, err)
		assert.Nil(t, issue)
		assert.IsType(t, &YouTrackError{}, err)
		youtrackErr := err.(*YouTrackError)
		assert.Equal(t, 404, youtrackErr.StatusCode)
	})
}

// TestUpdateIssue tests issue updates
func TestUpdateIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.True(t, strings.Contains(r.URL.Path, "/commands"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &providers.ProviderConfig{
		BaseURL: server.URL,
		Token:   "test-token",
	}

	client, err := NewYouTrackClient(config)
	require.NoError(t, err)

	t.Run("Successful update", func(t *testing.T) {
		updates := &YouTrackIssueUpdate{
			Summary: stringPtr("Updated Summary"),
		}

		ctx := context.Background()
		err := client.UpdateIssue(ctx, "test-123", updates)

		assert.NoError(t, err)
	})
}

// TestListIssues tests issue listing
func TestListIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/issues", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		issues := []*YouTrackIssue{
			{
				ID:         "list-1",
				IDReadable: "PROJ-001",
				Summary:    "First Issue",
				Created:    time.Now().Add(-2*time.Hour).Unix() * 1000,
				Updated:    time.Now().Add(-time.Hour).Unix() * 1000,
			},
			{
				ID:         "list-2",
				IDReadable: "PROJ-002",
				Summary:    "Second Issue",
				Created:    time.Now().Add(-time.Hour).Unix() * 1000,
				Updated:    time.Now().Unix() * 1000,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(issues)
	}))
	defer server.Close()

	config := &providers.ProviderConfig{
		BaseURL: server.URL,
		Token:   "test-token",
	}

	client, err := NewYouTrackClient(config)
	require.NoError(t, err)

	t.Run("List issues", func(t *testing.T) {
		filters := &YouTrackIssueFilters{
			Top: 10,
		}

		ctx := context.Background()
		issues, err := client.ListIssues(ctx, filters)

		assert.NoError(t, err)
		assert.Len(t, issues, 2)
		assert.Equal(t, "PROJ-001", issues[0].IDReadable)
		assert.Equal(t, "First Issue", issues[0].Summary)
	})
}

// TestDeleteIssue tests issue deletion
func TestDeleteIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.True(t, strings.HasPrefix(r.URL.Path, "/api/issues/"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &providers.ProviderConfig{
		BaseURL: server.URL,
		Token:   "test-token",
	}

	client, err := NewYouTrackClient(config)
	require.NoError(t, err)

	t.Run("Successful deletion", func(t *testing.T) {
		ctx := context.Background()
		err := client.DeleteIssue(ctx, "test-123")

		assert.NoError(t, err)
	})
}

// TestBulkOperations tests bulk operations
func TestBulkOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/api/issues":
			// Create issue
			var issue YouTrackIssue
			json.NewDecoder(r.Body).Decode(&issue)
			issue.ID = "bulk-created"
			issue.IDReadable = "PROJ-BULK"
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(issue)

		case r.Method == "POST" && strings.Contains(r.URL.Path, "/commands"):
			// Update issue
			w.WriteHeader(http.StatusOK)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := &providers.ProviderConfig{
		BaseURL: server.URL,
		Token:   "test-token",
	}

	client, err := NewYouTrackClient(config)
	require.NoError(t, err)

	t.Run("Bulk create issues", func(t *testing.T) {
		issues := []*YouTrackIssue{
			{Summary: "Bulk Issue 1"},
			{Summary: "Bulk Issue 2"},
		}

		ctx := context.Background()
		createdIssues, err := client.BulkCreateIssues(ctx, issues)

		assert.NoError(t, err)
		assert.Len(t, createdIssues, 2)
		for _, issue := range createdIssues {
			assert.Equal(t, "bulk-created", issue.ID)
		}
	})

	t.Run("Bulk update issues", func(t *testing.T) {
		updates := map[string]*YouTrackIssueUpdate{
			"issue-1": {Summary: stringPtr("Updated 1")},
			"issue-2": {Summary: stringPtr("Updated 2")},
		}

		ctx := context.Background()
		err := client.BulkUpdateIssues(ctx, updates)

		assert.NoError(t, err)
	})
}

// TestAddComment tests comment adding
func TestAddComment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.True(t, strings.Contains(r.URL.Path, "/comments"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	config := &providers.ProviderConfig{
		BaseURL: server.URL,
		Token:   "test-token",
	}

	client, err := NewYouTrackClient(config)
	require.NoError(t, err)

	t.Run("Add comment", func(t *testing.T) {
		comment := &YouTrackComment{
			Text: "Test comment",
		}

		ctx := context.Background()
		err := client.AddComment(ctx, "test-123", comment)

		assert.NoError(t, err)
	})
}

// TestErrorHandling tests error handling
func TestErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/issues/unauthorized":
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Unauthorized",
				"error_description": "Invalid token",
			})

		case "/api/issues/rate-limit":
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Too Many Requests",
			})

		case "/api/issues/server-error":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := &providers.ProviderConfig{
		BaseURL: server.URL,
		Token:   "test-token",
	}

	client, err := NewYouTrackClient(config)
	require.NoError(t, err)

	t.Run("Unauthorized error", func(t *testing.T) {
		ctx := context.Background()
		_, err := client.GetIssue(ctx, "unauthorized")

		assert.Error(t, err)
		ytError, ok := err.(*YouTrackError)
		require.True(t, ok)
		assert.Equal(t, 401, ytError.StatusCode)
		assert.Equal(t, "Unauthorized", ytError.Message)
		assert.Equal(t, "Invalid token", ytError.Details)
	})

	t.Run("Rate limit error", func(t *testing.T) {
		ctx := context.Background()
		_, err := client.GetIssue(ctx, "rate-limit")

		assert.Error(t, err)
		ytError, ok := err.(*YouTrackError)
		require.True(t, ok)
		assert.Equal(t, 429, ytError.StatusCode)
		assert.Equal(t, "Too Many Requests", ytError.Message)
	})

	t.Run("Server error", func(t *testing.T) {
		ctx := context.Background()
		_, err := client.GetIssue(ctx, "server-error")

		assert.Error(t, err)
		ytError, ok := err.(*YouTrackError)
		require.True(t, ok)
		assert.Equal(t, 500, ytError.StatusCode)
	})
}

// TestQueryBuilder tests query building from filters
func TestQueryBuilder(t *testing.T) {
	config := &providers.ProviderConfig{
		BaseURL: "https://test.youtrack.cloud",
		Token:   "test-token",
	}

	client, err := NewYouTrackClient(config)
	require.NoError(t, err)

	t.Run("Empty filters", func(t *testing.T) {
		filters := &YouTrackIssueFilters{}
		query := client.buildQueryFromFilters(filters)
		assert.Empty(t, query)
	})

	t.Run("Single filter", func(t *testing.T) {
		filters := &YouTrackIssueFilters{
			ProjectID: "PROJ",
		}
		query := client.buildQueryFromFilters(filters)
		assert.Equal(t, "project: PROJ", query)
	})

	t.Run("Multiple filters", func(t *testing.T) {
		filters := &YouTrackIssueFilters{
			ProjectID: "PROJ",
			State:     "Open",
			Assignee:  "testuser",
		}
		query := client.buildQueryFromFilters(filters)
		expected := "project: PROJ and State: {Open} and Assignee: testuser"
		assert.Equal(t, expected, query)
	})

	t.Run("Date filters", func(t *testing.T) {
		now := time.Now()
		yesterday := now.Add(-24 * time.Hour)
		filters := &YouTrackIssueFilters{
			CreatedAfter:  &yesterday,
			CreatedBefore: &now,
		}
		query := client.buildQueryFromFilters(filters)
		assert.Contains(t, query, "created:")
		assert.Contains(t, query, "..")
	})

	t.Run("Custom query", func(t *testing.T) {
		filters := &YouTrackIssueFilters{
			ProjectID: "PROJ",
			Query:     "assignee: me",
		}
		query := client.buildQueryFromFilters(filters)
		expected := "project: PROJ and assignee: me"
		assert.Equal(t, expected, query)
	})
}

// TestRateLimiting tests rate limiting functionality
func TestRateLimiting(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "test"})
	}))
	defer server.Close()

	t.Run("Rate limiting works", func(t *testing.T) {
		config := &providers.ProviderConfig{
			BaseURL: server.URL,
			Token:   "test-token",
			RateLimit: &providers.RateLimitConfig{
				RequestsPerSecond: 1, // Very low rate limit
				BurstSize:         1,
			},
		}

		client, err := NewYouTrackClient(config)
		require.NoError(t, err)

		ctx := context.Background()

		// First request should succeed immediately
		start := time.Now()
		_, err = client.GetIssue(ctx, "test")
		assert.NoError(t, err)
		firstDuration := time.Since(start)

		// Second request should be rate limited
		start = time.Now()
		_, err = client.GetIssue(ctx, "test")
		assert.NoError(t, err)
		secondDuration := time.Since(start)

		// Second request should take longer due to rate limiting
		assert.True(t, secondDuration > firstDuration)
		assert.True(t, secondDuration > 500*time.Millisecond)
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

// BenchmarkClient benchmarks client operations
func BenchmarkClient(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(YouTrackIssue{
			ID:      "bench-123",
			Summary: "Benchmark Issue",
		})
	}))
	defer server.Close()

	config := &providers.ProviderConfig{
		BaseURL: server.URL,
		Token:   "test-token",
	}

	client, err := NewYouTrackClient(config)
	require.NoError(b, err)

	ctx := context.Background()

	b.Run("GetIssue", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := client.GetIssue(ctx, "test-123")
			if err != nil {
				b.Fatalf("GetIssue failed: %v", err)
			}
		}
	})

	b.Run("CreateIssue", func(b *testing.B) {
		issue := &YouTrackIssue{
			Summary: "Benchmark Issue",
		}

		for i := 0; i < b.N; i++ {
			_, err := client.CreateIssue(ctx, issue)
			if err != nil {
				b.Fatalf("CreateIssue failed: %v", err)
			}
		}
	})
}
