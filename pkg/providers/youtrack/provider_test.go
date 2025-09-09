package youtrack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grik-ai/ricochet-task/pkg/providers"
)

// createTestProvider creates a YouTrack provider for testing
func createTestProvider(baseURL, token string) (*YouTrackProvider, error) {
	config := &providers.ProviderConfig{
		Name:     "test-youtrack",
		Type:     providers.ProviderTypeYouTrack,
		BaseURL:  baseURL,
		Token:    token,
		AuthType: providers.AuthTypeBearer,
		Timeout:  30 * time.Second,
		Enabled:  true,
	}

	return NewYouTrackProvider(config)
}

// createMockServer creates a mock YouTrack server for testing
func createMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
			return
		}

		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == "POST" && r.URL.Path == "/api/issues":
			// Create issue
			var issue YouTrackIssue
			if err := json.NewDecoder(r.Body).Decode(&issue); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Return created issue with ID
			issue.ID = "created-123"
			issue.IDReadable = "PROJ-123"
			issue.Created = time.Now().Unix() * 1000
			issue.Updated = time.Now().Unix() * 1000

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(issue)

		case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/api/issues/"):
			// Get issue
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
				Project: &YouTrackProject{
					ID:   "PROJ",
					Name: "Test Project",
				},
				State: &YouTrackState{
					ID:   "open",
					Name: "Open",
				},
				Priority: &YouTrackPriority{
					ID:   "normal",
					Name: "Normal",
				},
			}

			json.NewEncoder(w).Encode(issue)

		case r.Method == "POST" && strings.Contains(r.URL.Path, "/commands"):
			// Update issue (YouTrack uses commands for updates)
			w.WriteHeader(http.StatusOK)

		case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, "/api/issues/"):
			// Delete issue
			w.WriteHeader(http.StatusOK)

		case r.Method == "GET" && r.URL.Path == "/api/issues":
			// List issues
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

			json.NewEncoder(w).Encode(issues)

		case r.Method == "GET" && r.URL.Path == "/api/admin/users/me":
			// Health check endpoint
			user := map[string]interface{}{
				"id":    "test-user",
				"login": "testuser",
				"name":  "Test User",
			}
			json.NewEncoder(w).Encode(user)

		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Endpoint not found"})
		}
	}))
}

// TestNewYouTrackProvider tests provider creation
func TestNewYouTrackProvider(t *testing.T) {
	t.Run("Valid configuration", func(t *testing.T) {
		server := createMockServer()
		defer server.Close()

		provider, err := createTestProvider(server.URL, "test-token")
		require.NoError(t, err)
		assert.NotNil(t, provider)

		info := provider.GetProviderInfo()
		assert.Equal(t, "test-youtrack", info.Name)
		assert.Equal(t, providers.ProviderTypeYouTrack, info.Type)
		assert.True(t, info.Enabled)
	})

	t.Run("Missing base URL", func(t *testing.T) {
		config := &providers.ProviderConfig{
			Type:  providers.ProviderTypeYouTrack,
			Token: "test-token",
		}

		provider, err := NewYouTrackProvider(config)
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "base URL")
	})

	t.Run("Missing token", func(t *testing.T) {
		config := &providers.ProviderConfig{
			Type:    providers.ProviderTypeYouTrack,
			BaseURL: "https://test.youtrack.cloud",
		}

		provider, err := NewYouTrackProvider(config)
		assert.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "token")
	})
}

// TestCreateTask tests task creation
func TestCreateTask(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	provider, err := createTestProvider(server.URL, "test-token")
	require.NoError(t, err)

	t.Run("Successful creation", func(t *testing.T) {
		task := &providers.UniversalTask{
			Title:       "Test Task",
			Description: "Test Description",
			ProjectID:   "PROJ",
			Type:        providers.TaskTypeTask,
			Priority:    providers.TaskPriorityMedium,
		}

		ctx := context.Background()
		createdTask, err := provider.CreateTask(ctx, task)

		assert.NoError(t, err)
		assert.NotNil(t, createdTask)
		assert.Equal(t, "created-123", createdTask.ID)
		assert.Equal(t, "PROJ-123", createdTask.Key)
		assert.Equal(t, "Test Task", createdTask.Title)
		assert.Equal(t, "Test Description", createdTask.Description)
	})

	t.Run("Empty title", func(t *testing.T) {
		task := &providers.UniversalTask{
			Description: "Test Description",
		}

		ctx := context.Background()
		createdTask, err := provider.CreateTask(ctx, task)

		assert.Error(t, err)
		assert.Nil(t, createdTask)
		assert.Contains(t, err.Error(), "title")
	})
}

// TestGetTask tests task retrieval
func TestGetTask(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	provider, err := createTestProvider(server.URL, "test-token")
	require.NoError(t, err)

	t.Run("Existing task", func(t *testing.T) {
		ctx := context.Background()
		task, err := provider.GetTask(ctx, "test-123")

		assert.NoError(t, err)
		assert.NotNil(t, task)
		assert.Equal(t, "test-123", task.ID)
		assert.Equal(t, "PROJ-test-123", task.Key)
		assert.Equal(t, "Test Issue", task.Title)
		assert.Equal(t, "Test Description", task.Description)
	})

	t.Run("Non-existing task", func(t *testing.T) {
		ctx := context.Background()
		task, err := provider.GetTask(ctx, "not-found")

		assert.Error(t, err)
		assert.Nil(t, task)
		assert.True(t, providers.IsNotFoundError(err))
	})

	t.Run("Empty task ID", func(t *testing.T) {
		ctx := context.Background()
		task, err := provider.GetTask(ctx, "")

		assert.Error(t, err)
		assert.Nil(t, task)
		assert.Contains(t, err.Error(), "task ID")
	})
}

// TestUpdateTask tests task updates
func TestUpdateTask(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	provider, err := createTestProvider(server.URL, "test-token")
	require.NoError(t, err)

	t.Run("Successful update", func(t *testing.T) {
		newTitle := "Updated Title"
		newStatus := providers.TaskStatus{ID: "in_progress", Name: "In Progress"}

		updates := &providers.TaskUpdate{
			Title:  &newTitle,
			Status: &newStatus,
		}

		ctx := context.Background()
		err := provider.UpdateTask(ctx, "test-123", updates)

		assert.NoError(t, err)
	})

	t.Run("Empty task ID", func(t *testing.T) {
		updates := &providers.TaskUpdate{}

		ctx := context.Background()
		err := provider.UpdateTask(ctx, "", updates)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task ID")
	})

	t.Run("Nil updates", func(t *testing.T) {
		ctx := context.Background()
		err := provider.UpdateTask(ctx, "test-123", nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "updates")
	})
}

// TestDeleteTask tests task deletion
func TestDeleteTask(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	provider, err := createTestProvider(server.URL, "test-token")
	require.NoError(t, err)

	t.Run("Successful deletion", func(t *testing.T) {
		ctx := context.Background()
		err := provider.DeleteTask(ctx, "test-123")

		assert.NoError(t, err)
	})

	t.Run("Empty task ID", func(t *testing.T) {
		ctx := context.Background()
		err := provider.DeleteTask(ctx, "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task ID")
	})
}

// TestListTasks tests task listing
func TestListTasks(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	provider, err := createTestProvider(server.URL, "test-token")
	require.NoError(t, err)

	t.Run("List all tasks", func(t *testing.T) {
		ctx := context.Background()
		filters := &providers.TaskFilters{}

		tasks, err := provider.ListTasks(ctx, filters)

		assert.NoError(t, err)
		assert.Len(t, tasks, 2)
		assert.Equal(t, "PROJ-001", tasks[0].Key)
		assert.Equal(t, "First Issue", tasks[0].Title)
		assert.Equal(t, "PROJ-002", tasks[1].Key)
		assert.Equal(t, "Second Issue", tasks[1].Title)
	})

	t.Run("List with filters", func(t *testing.T) {
		ctx := context.Background()
		filters := &providers.TaskFilters{
			ProjectID: "PROJ",
			Status:    []string{"open"},
			Limit:     10,
		}

		tasks, err := provider.ListTasks(ctx, filters)

		assert.NoError(t, err)
		assert.NotNil(t, tasks)
	})

	t.Run("Nil filters", func(t *testing.T) {
		ctx := context.Background()

		tasks, err := provider.ListTasks(ctx, nil)

		assert.NoError(t, err)
		assert.NotNil(t, tasks)
	})
}

// TestSearchTasks tests task searching
func TestSearchTasks(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	provider, err := createTestProvider(server.URL, "test-token")
	require.NoError(t, err)

	t.Run("Search with query", func(t *testing.T) {
		ctx := context.Background()
		filters := &providers.TaskFilters{
			Query: "test issue",
		}

		tasks, err := provider.SearchTasks(ctx, "test issue", filters)

		assert.NoError(t, err)
		assert.NotNil(t, tasks)
	})

	t.Run("Empty query", func(t *testing.T) {
		ctx := context.Background()
		filters := &providers.TaskFilters{}

		tasks, err := provider.SearchTasks(ctx, "", filters)

		assert.Error(t, err)
		assert.Nil(t, tasks)
		assert.Contains(t, err.Error(), "query")
	})
}

// TestHealthCheck tests provider health checking
func TestHealthCheck(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	provider, err := createTestProvider(server.URL, "test-token")
	require.NoError(t, err)

	t.Run("Healthy provider", func(t *testing.T) {
		ctx := context.Background()
		err := provider.HealthCheck(ctx)

		assert.NoError(t, err)
	})

	t.Run("Unhealthy provider", func(t *testing.T) {
		// Create provider with invalid URL
		unhealthyProvider, err := createTestProvider("http://invalid-url", "test-token")
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err = unhealthyProvider.HealthCheck(ctx)
		assert.Error(t, err)
	})
}

// TestProviderInfo tests provider information
func TestProviderInfo(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	provider, err := createTestProvider(server.URL, "test-token")
	require.NoError(t, err)

	info := provider.GetProviderInfo()

	assert.Equal(t, "test-youtrack", info.Name)
	assert.Equal(t, providers.ProviderTypeYouTrack, info.Type)
	assert.Equal(t, "1.0.0", info.Version)
	assert.True(t, info.Enabled)
	assert.Equal(t, providers.HealthStatusUnknown, info.HealthStatus)

	// Test capabilities
	expectedCapabilities := []providers.Capability{
		providers.CapabilityTasks,
		providers.CapabilityBoards,
		providers.CapabilityRealTimeSync,
		providers.CapabilityCustomFields,
		providers.CapabilityWorkflows,
		providers.CapabilityTimeTracking,
		providers.CapabilityHierarchicalTasks,
		providers.CapabilityReporting,
		providers.CapabilityAdvancedSearch,
		providers.CapabilityWebhooks,
	}

	for _, capability := range expectedCapabilities {
		assert.Contains(t, info.Capabilities, capability)
	}
}

// TestClose tests provider cleanup
func TestClose(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	provider, err := createTestProvider(server.URL, "test-token")
	require.NoError(t, err)

	err = provider.Close()
	assert.NoError(t, err)
}

// TestConcurrentOperations tests concurrent access to provider
func TestConcurrentOperations(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	provider, err := createTestProvider(server.URL, "test-token")
	require.NoError(t, err)

	t.Run("Concurrent task creation", func(t *testing.T) {
		ctx := context.Background()
		concurrency := 5
		errors := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(id int) {
				task := &providers.UniversalTask{
					Title: fmt.Sprintf("Concurrent Task %d", id),
				}
				_, err := provider.CreateTask(ctx, task)
				errors <- err
			}(i)
		}

		// Collect results
		for i := 0; i < concurrency; i++ {
			err := <-errors
			assert.NoError(t, err)
		}
	})

	t.Run("Concurrent health checks", func(t *testing.T) {
		ctx := context.Background()
		concurrency := 3
		errors := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			go func() {
				errors <- provider.HealthCheck(ctx)
			}()
		}

		// Collect results
		for i := 0; i < concurrency; i++ {
			err := <-errors
			assert.NoError(t, err)
		}
	})
}

// TestContextCancellation tests context cancellation handling
func TestContextCancellation(t *testing.T) {
	server := createMockServer()
	defer server.Close()

	provider, err := createTestProvider(server.URL, "test-token")
	require.NoError(t, err)

	t.Run("Cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		task := &providers.UniversalTask{
			Title: "Test Task",
		}

		_, err := provider.CreateTask(ctx, task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context")
	})

	t.Run("Timeout context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Ensure timeout

		task := &providers.UniversalTask{
			Title: "Test Task",
		}

		_, err := provider.CreateTask(ctx, task)
		assert.Error(t, err)
	})
}

// BenchmarkCreateTask benchmarks task creation
func BenchmarkCreateTask(b *testing.B) {
	server := createMockServer()
	defer server.Close()

	provider, err := createTestProvider(server.URL, "test-token")
	require.NoError(b, err)

	task := &providers.UniversalTask{
		Title:       "Benchmark Task",
		Description: "Benchmark Description",
		ProjectID:   "PROJ",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.CreateTask(ctx, task)
		if err != nil {
			b.Fatalf("CreateTask failed: %v", err)
		}
	}
}

// BenchmarkGetTask benchmarks task retrieval
func BenchmarkGetTask(b *testing.B) {
	server := createMockServer()
	defer server.Close()

	provider, err := createTestProvider(server.URL, "test-token")
	require.NoError(b, err)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.GetTask(ctx, "test-123")
		if err != nil {
			b.Fatalf("GetTask failed: %v", err)
		}
	}
}

// BenchmarkListTasks benchmarks task listing
func BenchmarkListTasks(b *testing.B) {
	server := createMockServer()
	defer server.Close()

	provider, err := createTestProvider(server.URL, "test-token")
	require.NoError(b, err)

	filters := &providers.TaskFilters{
		Limit: 10,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.ListTasks(ctx, filters)
		if err != nil {
			b.Fatalf("ListTasks failed: %v", err)
		}
	}
}
