package providers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTaskProvider is a mock implementation of TaskProvider for testing
type MockTaskProvider struct {
	mock.Mock
}

func (m *MockTaskProvider) CreateTask(ctx context.Context, task *UniversalTask) (*UniversalTask, error) {
	args := m.Called(ctx, task)
	return args.Get(0).(*UniversalTask), args.Error(1)
}

func (m *MockTaskProvider) GetTask(ctx context.Context, id string) (*UniversalTask, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UniversalTask), args.Error(1)
}

func (m *MockTaskProvider) UpdateTask(ctx context.Context, id string, updates *TaskUpdate) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockTaskProvider) DeleteTask(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskProvider) ListTasks(ctx context.Context, filters *TaskFilters) ([]*UniversalTask, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*UniversalTask), args.Error(1)
}

func (m *MockTaskProvider) SearchTasks(ctx context.Context, query string, filters *TaskFilters) ([]*UniversalTask, error) {
	args := m.Called(ctx, query, filters)
	return args.Get(0).([]*UniversalTask), args.Error(1)
}

func (m *MockTaskProvider) GetProviderInfo() *ProviderInfo {
	args := m.Called()
	return args.Get(0).(*ProviderInfo)
}

func (m *MockTaskProvider) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTaskProvider) Close() error {
	args := m.Called()
	return args.Error(0)
}

// TestUniversalTask tests the UniversalTask model
func TestUniversalTask(t *testing.T) {
	t.Run("GetDisplayID with Key", func(t *testing.T) {
		task := &UniversalTask{
			ID:  "internal-123",
			Key: "PROJ-456",
		}
		assert.Equal(t, "PROJ-456", task.GetDisplayID())
	})

	t.Run("GetDisplayID without Key", func(t *testing.T) {
		task := &UniversalTask{
			ID: "internal-123",
		}
		assert.Equal(t, "internal-123", task.GetDisplayID())
	})

	t.Run("IsCompleted with final status", func(t *testing.T) {
		task := &UniversalTask{
			Status: TaskStatus{
				ID:      "done",
				Name:    "Done",
				IsFinal: true,
			},
		}
		assert.True(t, task.IsCompleted())
	})

	t.Run("IsCompleted with non-final status", func(t *testing.T) {
		task := &UniversalTask{
			Status: TaskStatus{
				ID:      "open",
				Name:    "Open",
				IsFinal: false,
			},
		}
		assert.False(t, task.IsCompleted())
	})

	t.Run("GetAge calculation", func(t *testing.T) {
		now := time.Now()
		task := &UniversalTask{
			CreatedAt: now.Add(-24 * time.Hour), // 1 day ago
		}
		age := task.GetAge()
		assert.True(t, age >= 23*time.Hour && age <= 25*time.Hour)
	})

	t.Run("HasLabel", func(t *testing.T) {
		task := &UniversalTask{
			Labels: []string{"bug", "critical", "backend"},
		}
		assert.True(t, task.HasLabel("bug"))
		assert.True(t, task.HasLabel("critical"))
		assert.False(t, task.HasLabel("frontend"))
	})
}

// TestTaskStatus tests the TaskStatus model
func TestTaskStatus(t *testing.T) {
	t.Run("IsCategory", func(t *testing.T) {
		status := TaskStatus{
			Category: StatusCategoryInProgress,
		}
		assert.True(t, status.IsCategory(StatusCategoryInProgress))
		assert.False(t, status.IsCategory(StatusCategoryDone))
	})

	t.Run("Status categories", func(t *testing.T) {
		assert.Equal(t, "todo", string(StatusCategoryTodo))
		assert.Equal(t, "in_progress", string(StatusCategoryInProgress))
		assert.Equal(t, "review", string(StatusCategoryReview))
		assert.Equal(t, "done", string(StatusCategoryDone))
		assert.Equal(t, "cancelled", string(StatusCategoryCancelled))
		assert.Equal(t, "blocked", string(StatusCategoryBlocked))
	})
}

// TestTaskPriority tests the TaskPriority model
func TestTaskPriority(t *testing.T) {
	t.Run("Priority ordering", func(t *testing.T) {
		priorities := []TaskPriority{
			TaskPriorityLowest,
			TaskPriorityLow,
			TaskPriorityMedium,
			TaskPriorityHigh,
			TaskPriorityHighest,
			TaskPriorityCritical,
		}

		expected := []string{"lowest", "low", "medium", "high", "highest", "critical"}
		for i, priority := range priorities {
			assert.Equal(t, expected[i], string(priority))
		}
	})
}

// TestTaskType tests the TaskType model
func TestTaskType(t *testing.T) {
	t.Run("Task types", func(t *testing.T) {
		types := map[TaskType]string{
			TaskTypeTask:        "task",
			TaskTypeBug:         "bug",
			TaskTypeFeature:     "feature",
			TaskTypeEpic:        "epic",
			TaskTypeStory:       "story",
			TaskTypeSubtask:     "subtask",
			TaskTypeImprovement: "improvement",
			TaskTypeResearch:    "research",
		}

		for taskType, expected := range types {
			assert.Equal(t, expected, string(taskType))
		}
	})
}

// TestTaskFilters tests the TaskFilters model
func TestTaskFilters(t *testing.T) {
	t.Run("Empty filters", func(t *testing.T) {
		filters := &TaskFilters{}
		assert.Empty(t, filters.Status)
		assert.Empty(t, filters.Priority)
		assert.Empty(t, filters.Type)
		assert.Equal(t, 0, filters.Limit)
		assert.Equal(t, 0, filters.Offset)
	})

	t.Run("Filters with values", func(t *testing.T) {
		filters := &TaskFilters{
			ProjectID:  "PROJ",
			AssigneeID: "user123",
			Status:     []string{"open", "in_progress"},
			Priority:   []string{"high", "critical"},
			Type:       []string{"bug", "feature"},
			Labels:     []string{"backend"},
			Limit:      50,
			Offset:     10,
		}

		assert.Equal(t, "PROJ", filters.ProjectID)
		assert.Equal(t, "user123", filters.AssigneeID)
		assert.Equal(t, []string{"open", "in_progress"}, filters.Status)
		assert.Equal(t, []string{"high", "critical"}, filters.Priority)
		assert.Equal(t, []string{"bug", "feature"}, filters.Type)
		assert.Equal(t, []string{"backend"}, filters.Labels)
		assert.Equal(t, 50, filters.Limit)
		assert.Equal(t, 10, filters.Offset)
	})
}

// TestTaskUpdate tests the TaskUpdate model
func TestTaskUpdate(t *testing.T) {
	t.Run("Empty update", func(t *testing.T) {
		update := &TaskUpdate{}
		assert.Nil(t, update.Title)
		assert.Nil(t, update.Description)
		assert.Nil(t, update.Status)
		assert.Nil(t, update.Priority)
		assert.Nil(t, update.AssigneeID)
	})

	t.Run("Update with values", func(t *testing.T) {
		title := "New Title"
		description := "New Description"
		status := TaskStatus{ID: "done", Name: "Done"}
		priority := TaskPriorityHigh
		assigneeID := "user456"

		update := &TaskUpdate{
			Title:       &title,
			Description: &description,
			Status:      &status,
			Priority:    &priority,
			AssigneeID:  &assigneeID,
			Labels:      []string{"updated"},
		}

		assert.Equal(t, "New Title", *update.Title)
		assert.Equal(t, "New Description", *update.Description)
		assert.Equal(t, "done", update.Status.ID)
		assert.Equal(t, TaskPriorityHigh, *update.Priority)
		assert.Equal(t, "user456", *update.AssigneeID)
		assert.Equal(t, []string{"updated"}, update.Labels)
	})
}

// TestProviderInfo tests the ProviderInfo model
func TestProviderInfo(t *testing.T) {
	t.Run("Provider info creation", func(t *testing.T) {
		info := &ProviderInfo{
			Name:         "test-provider",
			Type:         ProviderTypeYouTrack,
			Version:      "1.0.0",
			Enabled:      true,
			HealthStatus: HealthStatusHealthy,
			Capabilities: []Capability{CapabilityTasks, CapabilityBoards},
			LastHealthCheck: time.Now(),
		}

		assert.Equal(t, "test-provider", info.Name)
		assert.Equal(t, ProviderTypeYouTrack, info.Type)
		assert.Equal(t, "1.0.0", info.Version)
		assert.True(t, info.Enabled)
		assert.Equal(t, HealthStatusHealthy, info.HealthStatus)
		assert.Len(t, info.Capabilities, 2)
		assert.Contains(t, info.Capabilities, CapabilityTasks)
		assert.Contains(t, info.Capabilities, CapabilityBoards)
	})

	t.Run("HasCapability", func(t *testing.T) {
		info := &ProviderInfo{
			Capabilities: []Capability{CapabilityTasks, CapabilityCustomFields},
		}

		assert.True(t, info.HasCapability(CapabilityTasks))
		assert.True(t, info.HasCapability(CapabilityCustomFields))
		assert.False(t, info.HasCapability(CapabilityWebhooks))
	})
}

// TestProviderTypes tests the provider type constants
func TestProviderTypes(t *testing.T) {
	t.Run("Provider types", func(t *testing.T) {
		types := map[ProviderType]string{
			ProviderTypeYouTrack: "youtrack",
			ProviderTypeJira:     "jira",
			ProviderTypeNotion:   "notion",
			ProviderTypeLinear:   "linear",
			ProviderTypeGitHub:   "github",
			ProviderTypeCustom:   "custom",
		}

		for providerType, expected := range types {
			assert.Equal(t, expected, string(providerType))
		}
	})
}

// TestCapabilities tests the capability constants
func TestCapabilities(t *testing.T) {
	t.Run("Capabilities", func(t *testing.T) {
		capabilities := map[Capability]string{
			CapabilityTasks:            "tasks",
			CapabilityBoards:           "boards",
			CapabilityRealTimeSync:     "real_time_sync",
			CapabilityCustomFields:     "custom_fields",
			CapabilityWorkflows:        "workflows",
			CapabilityTimeTracking:     "time_tracking",
			CapabilityHierarchicalTasks: "hierarchical_tasks",
			CapabilityReporting:        "reporting",
			CapabilityAdvancedSearch:   "advanced_search",
			CapabilityWebhooks:         "webhooks",
		}

		for capability, expected := range capabilities {
			assert.Equal(t, expected, string(capability))
		}
	})
}

// TestHealthStatus tests the health status constants
func TestHealthStatus(t *testing.T) {
	t.Run("Health statuses", func(t *testing.T) {
		statuses := map[ProviderHealthStatus]string{
			HealthStatusHealthy:     "healthy",
			HealthStatusUnhealthy:   "unhealthy",
			HealthStatusDegraded:    "degraded",
			HealthStatusUnknown:     "unknown",
		}

		for status, expected := range statuses {
			assert.Equal(t, expected, string(status))
		}
	})
}

// TestErrorTypes tests the error type constants
func TestErrorTypes(t *testing.T) {
	t.Run("Error types", func(t *testing.T) {
		errorTypes := map[ErrorType]string{
			ErrorTypeNotFound:       "not_found",
			ErrorTypeUnauthorized:   "unauthorized",
			ErrorTypeForbidden:      "forbidden",
			ErrorTypeRateLimit:      "rate_limit",
			ErrorTypeValidation:     "validation",
			ErrorTypeNetwork:        "network",
			ErrorTypeInternal:       "internal",
			ErrorTypeConfiguration:  "configuration",
		}

		for errorType, expected := range errorTypes {
			assert.Equal(t, expected, string(errorType))
		}
	})
}

// TestProviderError tests the ProviderError
func TestProviderError(t *testing.T) {
	t.Run("Provider error creation", func(t *testing.T) {
		err := &ProviderError{
			Type:    ErrorTypeNotFound,
			Message: "Task not found",
			Context: map[string]interface{}{
				"task_id": "PROJ-123",
				"provider": "youtrack",
			},
		}

		assert.Equal(t, ErrorTypeNotFound, err.Type)
		assert.Equal(t, "Task not found", err.Message)
		assert.Equal(t, "PROJ-123", err.Context["task_id"])
		assert.Equal(t, "youtrack", err.Context["provider"])
		assert.Contains(t, err.Error(), "Task not found")
	})

	t.Run("Error type checking", func(t *testing.T) {
		notFoundErr := &ProviderError{Type: ErrorTypeNotFound}
		rateLimit := &ProviderError{Type: ErrorTypeRateLimit}

		assert.True(t, IsNotFoundError(notFoundErr))
		assert.False(t, IsNotFoundError(rateLimit))

		assert.True(t, IsRateLimitError(rateLimit))
		assert.False(t, IsRateLimitError(notFoundErr))
	})
}

// TestMockTaskProvider tests the mock implementation
func TestMockTaskProvider(t *testing.T) {
	t.Run("Mock task provider operations", func(t *testing.T) {
		mockProvider := new(MockTaskProvider)
		ctx := context.Background()

		// Setup expectations
		task := &UniversalTask{
			ID:    "test-123",
			Title: "Test Task",
		}

		mockProvider.On("CreateTask", ctx, mock.AnythingOfType("*providers.UniversalTask")).Return(task, nil)
		mockProvider.On("GetTask", ctx, "test-123").Return(task, nil)
		mockProvider.On("HealthCheck", ctx).Return(nil)

		// Test operations
		createdTask, err := mockProvider.CreateTask(ctx, task)
		assert.NoError(t, err)
		assert.Equal(t, "test-123", createdTask.ID)

		retrievedTask, err := mockProvider.GetTask(ctx, "test-123")
		assert.NoError(t, err)
		assert.Equal(t, "Test Task", retrievedTask.Title)

		err = mockProvider.HealthCheck(ctx)
		assert.NoError(t, err)

		// Verify all expectations were met
		mockProvider.AssertExpectations(t)
	})
}

// BenchmarkUniversalTask benchmarks UniversalTask operations
func BenchmarkUniversalTask(b *testing.B) {
	task := &UniversalTask{
		ID:        "bench-123",
		Key:       "PROJ-456",
		Title:     "Benchmark Task",
		Labels:    []string{"performance", "test", "benchmark"},
		CreatedAt: time.Now().Add(-time.Hour),
	}

	b.Run("GetDisplayID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = task.GetDisplayID()
		}
	})

	b.Run("HasLabel", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = task.HasLabel("performance")
		}
	})

	b.Run("GetAge", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = task.GetAge()
		}
	})
}
