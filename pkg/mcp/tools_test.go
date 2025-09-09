package mcp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/sirupsen/logrus"

	"github.com/grik-ai/ricochet-task/pkg/providers"
)

// MockProviderRegistry is a mock implementation of ProviderRegistry for testing
type MockProviderRegistry struct {
	mock.Mock
}

func (m *MockProviderRegistry) ListProviders() map[string]*providers.ProviderInfo {
	args := m.Called()
	return args.Get(0).(map[string]*providers.ProviderInfo)
}

func (m *MockProviderRegistry) ListEnabledProviders() map[string]*providers.ProviderInfo {
	args := m.Called()
	return args.Get(0).(map[string]*providers.ProviderInfo)
}

func (m *MockProviderRegistry) GetProvider(name string) (providers.TaskProvider, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(providers.TaskProvider), args.Error(1)
}

func (m *MockProviderRegistry) GetDefaultProvider() (providers.TaskProvider, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(providers.TaskProvider), args.Error(1)
}

func (m *MockProviderRegistry) AddProvider(ctx context.Context, name string, config *providers.ProviderConfig) error {
	args := m.Called(ctx, name, config)
	return args.Error(0)
}

func (m *MockProviderRegistry) RemoveProvider(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockProviderRegistry) EnableProvider(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockProviderRegistry) DisableProvider(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockProviderRegistry) SetDefaultProvider(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockProviderRegistry) GetHealthStatus() map[string]providers.HealthStatus {
	args := m.Called()
	return args.Get(0).(map[string]providers.HealthStatus)
}

func (m *MockProviderRegistry) Initialize(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockProviderRegistry) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockTaskProvider for testing
type MockTaskProvider struct {
	mock.Mock
}

func (m *MockTaskProvider) CreateTask(ctx context.Context, task *providers.UniversalTask) (*providers.UniversalTask, error) {
	args := m.Called(ctx, task)
	return args.Get(0).(*providers.UniversalTask), args.Error(1)
}

func (m *MockTaskProvider) GetTask(ctx context.Context, id string) (*providers.UniversalTask, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*providers.UniversalTask), args.Error(1)
}

func (m *MockTaskProvider) UpdateTask(ctx context.Context, id string, updates *providers.TaskUpdate) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockTaskProvider) DeleteTask(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTaskProvider) ListTasks(ctx context.Context, filters *providers.TaskFilters) ([]*providers.UniversalTask, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*providers.UniversalTask), args.Error(1)
}

func (m *MockTaskProvider) SearchTasks(ctx context.Context, query string, filters *providers.TaskFilters) ([]*providers.UniversalTask, error) {
	args := m.Called(ctx, query, filters)
	return args.Get(0).([]*providers.UniversalTask), args.Error(1)
}

func (m *MockTaskProvider) GetProviderInfo() *providers.ProviderInfo {
	args := m.Called()
	return args.Get(0).(*providers.ProviderInfo)
}

func (m *MockTaskProvider) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTaskProvider) Close() error {
	args := m.Called()
	return args.Error(0)
}

// createTestToolProvider creates a test MCP tool provider
func createTestToolProvider() (*MCPToolProvider, *MockProviderRegistry) {
	mockRegistry := new(MockProviderRegistry)
	toolProvider := NewMCPToolProvider(mockRegistry)
	return toolProvider, mockRegistry
}

// TestNewMCPToolProvider tests tool provider creation
func TestNewMCPToolProvider(t *testing.T) {
	mockRegistry := new(MockProviderRegistry)
	toolProvider := NewMCPToolProvider(mockRegistry)

	assert.NotNil(t, toolProvider)
	assert.Equal(t, mockRegistry, toolProvider.registry)
}

// TestGetTools tests tool definitions retrieval
func TestGetTools(t *testing.T) {
	toolProvider, _ := createTestToolProvider()

	tools := toolProvider.GetTools()

	assert.Len(t, tools, 9) // We have 9 tools defined

	// Check that all expected tools are present
	expectedTools := []string{
		"providers_list",
		"provider_health",
		"providers_add",
		"task_create_smart",
		"task_list_unified",
		"task_update_universal",
		"cross_provider_search",
		"ai_analyze_project",
		"ai_execute_task",
	}

	toolMap := make(map[string]ToolDefinition)
	for _, tool := range tools {
		toolMap[tool.Name] = tool
	}

	for _, expectedTool := range expectedTools {
		assert.Contains(t, toolMap, expectedTool)
		assert.NotEmpty(t, toolMap[expectedTool].Description)
		assert.NotNil(t, toolMap[expectedTool].InputSchema)
	}
}

// TestExecuteProvidersList tests providers_list tool
func TestExecuteProvidersList(t *testing.T) {
	toolProvider, mockRegistry := createTestToolProvider()

	t.Run("List all providers", func(t *testing.T) {
		mockProviders := map[string]*providers.ProviderInfo{
			"youtrack-prod": {
				Name:         "youtrack-prod",
				Type:         providers.ProviderTypeYouTrack,
				Enabled:      true,
				HealthStatus: providers.HealthStatusHealthy,
			},
			"jira-dev": {
				Name:         "jira-dev",
				Type:         providers.ProviderTypeJira,
				Enabled:      false,
				HealthStatus: providers.HealthStatusUnhealthy,
			},
		}

		mockRegistry.On("ListProviders").Return(mockProviders)

		ctx := context.Background()
		args := map[string]interface{}{
			"enabled_only":   false,
			"output_format": "table",
		}

		result, err := toolProvider.ExecuteTool(ctx, "providers_list", args)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Error)
		assert.Len(t, result.Content, 1)
		assert.Equal(t, "text", result.Content[0]["type"])
		assert.Contains(t, result.Content[0]["text"], "youtrack-prod")
		assert.Contains(t, result.Content[0]["text"], "jira-dev")

		mockRegistry.AssertExpectations(t)
	})

	t.Run("List enabled providers only", func(t *testing.T) {
		mockEnabledProviders := map[string]*providers.ProviderInfo{
			"youtrack-prod": {
				Name:         "youtrack-prod",
				Type:         providers.ProviderTypeYouTrack,
				Enabled:      true,
				HealthStatus: providers.HealthStatusHealthy,
			},
		}

		mockRegistry.On("ListEnabledProviders").Return(mockEnabledProviders)

		ctx := context.Background()
		args := map[string]interface{}{
			"enabled_only":   true,
			"output_format": "summary",
		}

		result, err := toolProvider.ExecuteTool(ctx, "providers_list", args)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Error)
		assert.Contains(t, result.Content[0]["text"], "Provider Summary")

		mockRegistry.AssertExpectations(t)
	})
}

// TestExecuteProviderHealth tests provider_health tool
func TestExecuteProviderHealth(t *testing.T) {
	toolProvider, mockRegistry := createTestToolProvider()

	t.Run("Check specific provider health", func(t *testing.T) {
		mockProvider := new(MockTaskProvider)
		mockProvider.On("HealthCheck", mock.AnythingOfType("*context.timerCtx")).Return(nil)

		mockRegistry.On("GetProvider", "youtrack-prod").Return(mockProvider, nil)

		ctx := context.Background()
		args := map[string]interface{}{
			"provider_name":    "youtrack-prod",
			"include_details": false,
		}

		result, err := toolProvider.ExecuteTool(ctx, "provider_health", args)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Error)
		assert.Contains(t, result.Content[0]["text"], "youtrack-prod")
		assert.Contains(t, result.Content[0]["text"], "🟢 HEALTHY")

		mockRegistry.AssertExpectations(t)
		mockProvider.AssertExpectations(t)
	})

	t.Run("Check all providers health", func(t *testing.T) {
		mockHealthStatus := map[string]providers.HealthStatus{
			"youtrack-prod": providers.HealthStatusHealthy,
			"jira-dev":      providers.HealthStatusUnhealthy,
		}

		mockRegistry.On("GetHealthStatus").Return(mockHealthStatus)

		ctx := context.Background()
		args := map[string]interface{}{}

		result, err := toolProvider.ExecuteTool(ctx, "provider_health", args)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Error)
		assert.Contains(t, result.Content[0]["text"], "Provider Health Status")
		assert.Contains(t, result.Content[0]["text"], "youtrack-prod")
		assert.Contains(t, result.Content[0]["text"], "jira-dev")

		mockRegistry.AssertExpectations(t)
	})
}

// TestExecuteProvidersAdd tests providers_add tool
func TestExecuteProvidersAdd(t *testing.T) {
	toolProvider, mockRegistry := createTestToolProvider()

	t.Run("Add provider successfully", func(t *testing.T) {
		mockRegistry.On("AddProvider", mock.AnythingOfType("*context.timerCtx"), "test-youtrack", mock.AnythingOfType("*providers.ProviderConfig")).Return(nil)

		ctx := context.Background()
		args := map[string]interface{}{
			"name":     "test-youtrack",
			"type":     "youtrack",
			"base_url": "https://test.youtrack.cloud",
			"token":    "test-token",
			"enable":   true,
		}

		result, err := toolProvider.ExecuteTool(ctx, "providers_add", args)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Error)
		assert.Contains(t, result.Content[0]["text"], "✅")
		assert.Contains(t, result.Content[0]["text"], "test-youtrack")

		mockRegistry.AssertExpectations(t)
	})

	t.Run("Missing required parameters", func(t *testing.T) {
		ctx := context.Background()
		args := map[string]interface{}{
			"name": "test-youtrack",
			// Missing type, base_url, token
		}

		result, err := toolProvider.ExecuteTool(ctx, "providers_add", args)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Error)
		assert.Contains(t, *result.Error, "Missing required parameters")
	})
}

// TestExecuteTaskCreateSmart tests task_create_smart tool
func TestExecuteTaskCreateSmart(t *testing.T) {
	toolProvider, mockRegistry := createTestToolProvider()

	t.Run("Create task successfully", func(t *testing.T) {
		mockProvider := new(MockTaskProvider)
		createdTask := &providers.UniversalTask{
			ID:    "created-123",
			Key:   "PROJ-123",
			Title: "Test Task",
			ProviderName: "youtrack-prod",
		}

		mockProvider.On("CreateTask", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("*providers.UniversalTask")).Return(createdTask, nil)
		mockRegistry.On("GetDefaultProvider").Return(mockProvider, nil)

		ctx := context.Background()
		args := map[string]interface{}{
			"title":       "Test Task",
			"description": "Test Description",
			"task_type":   "task",
			"priority":    "medium",
		}

		result, err := toolProvider.ExecuteTool(ctx, "task_create_smart", args)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Error)
		assert.Contains(t, result.Content[0]["text"], "✅")
		assert.Contains(t, result.Content[0]["text"], "PROJ-123")
		assert.Contains(t, result.Content[0]["text"], "Test Task")

		mockRegistry.AssertExpectations(t)
		mockProvider.AssertExpectations(t)
	})

	t.Run("Missing title", func(t *testing.T) {
		ctx := context.Background()
		args := map[string]interface{}{
			"description": "Test Description",
		}

		result, err := toolProvider.ExecuteTool(ctx, "task_create_smart", args)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Error)
		assert.Contains(t, *result.Error, "Title is required")
	})
}

// TestExecuteTaskListUnified tests task_list_unified tool
func TestExecuteTaskListUnified(t *testing.T) {
	toolProvider, mockRegistry := createTestToolProvider()

	t.Run("List tasks from all providers", func(t *testing.T) {
		mockProvider := new(MockTaskProvider)
		mockTasks := []*providers.UniversalTask{
			{
				ID:    "task-1",
				Key:   "PROJ-001",
				Title: "First Task",
				Status: providers.TaskStatus{Name: "Open"},
				Priority: providers.TaskPriorityMedium,
			},
			{
				ID:    "task-2",
				Key:   "PROJ-002",
				Title: "Second Task",
				Status: providers.TaskStatus{Name: "In Progress"},
				Priority: providers.TaskPriorityHigh,
			},
		}

		mockEnabledProviders := map[string]*providers.ProviderInfo{
			"youtrack-prod": {
				Name: "youtrack-prod",
			},
		}

		mockProvider.On("ListTasks", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("*providers.TaskFilters")).Return(mockTasks, nil)
		mockRegistry.On("ListEnabledProviders").Return(mockEnabledProviders)
		mockRegistry.On("GetProvider", "youtrack-prod").Return(mockProvider, nil)

		ctx := context.Background()
		args := map[string]interface{}{
			"providers":     []interface{}{"all"},
			"limit":        50,
			"output_format": "table",
		}

		result, err := toolProvider.ExecuteTool(ctx, "task_list_unified", args)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Error)
		assert.Contains(t, result.Content[0]["text"], "PROJ-001")
		assert.Contains(t, result.Content[0]["text"], "First Task")

		mockRegistry.AssertExpectations(t)
		mockProvider.AssertExpectations(t)
	})
}

// TestExecuteTaskUpdateUniversal tests task_update_universal tool
func TestExecuteTaskUpdateUniversal(t *testing.T) {
	toolProvider, mockRegistry := createTestToolProvider()

	t.Run("Update task successfully", func(t *testing.T) {
		mockProvider := new(MockTaskProvider)
		mockProvider.On("UpdateTask", mock.AnythingOfType("*context.timerCtx"), "PROJ-123", mock.AnythingOfType("*providers.TaskUpdate")).Return(nil)
		mockRegistry.On("GetDefaultProvider").Return(mockProvider, nil)

		ctx := context.Background()
		args := map[string]interface{}{
			"task_id":  "PROJ-123",
			"title":    "Updated Title",
			"status":   "in_progress",
			"priority": "high",
		}

		result, err := toolProvider.ExecuteTool(ctx, "task_update_universal", args)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Error)
		assert.Contains(t, result.Content[0]["text"], "✅")
		assert.Contains(t, result.Content[0]["text"], "PROJ-123")

		mockRegistry.AssertExpectations(t)
		mockProvider.AssertExpectations(t)
	})

	t.Run("Missing task ID", func(t *testing.T) {
		ctx := context.Background()
		args := map[string]interface{}{
			"title": "Updated Title",
		}

		result, err := toolProvider.ExecuteTool(ctx, "task_update_universal", args)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Error)
		assert.Contains(t, *result.Error, "Task ID is required")
	})
}

// TestExecuteCrossProviderSearch tests cross_provider_search tool
func TestExecuteCrossProviderSearch(t *testing.T) {
	toolProvider, mockRegistry := createTestToolProvider()

	t.Run("Search across providers", func(t *testing.T) {
		mockProvider := new(MockTaskProvider)
		mockTasks := []*providers.UniversalTask{
			{
				ID:    "search-1",
				Key:   "PROJ-100",
				Title: "Authentication Bug",
				Description: "Fix authentication issue",
			},
		}

		mockEnabledProviders := map[string]*providers.ProviderInfo{
			"youtrack-prod": {Name: "youtrack-prod"},
		}

		mockProvider.On("ListTasks", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("*providers.TaskFilters")).Return(mockTasks, nil)
		mockRegistry.On("ListEnabledProviders").Return(mockEnabledProviders)
		mockRegistry.On("GetProvider", "youtrack-prod").Return(mockProvider, nil)

		ctx := context.Background()
		args := map[string]interface{}{
			"query":           "authentication",
			"providers":       []interface{}{"all"},
			"limit":           20,
			"include_content": true,
		}

		result, err := toolProvider.ExecuteTool(ctx, "cross_provider_search", args)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Error)
		assert.Contains(t, result.Content[0]["text"], "Found 1 tasks")
		assert.Contains(t, result.Content[0]["text"], "Authentication Bug")

		mockRegistry.AssertExpectations(t)
		mockProvider.AssertExpectations(t)
	})

	t.Run("Empty query", func(t *testing.T) {
		ctx := context.Background()
		args := map[string]interface{}{}

		result, err := toolProvider.ExecuteTool(ctx, "cross_provider_search", args)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Error)
		assert.Contains(t, *result.Error, "Search query is required")
	})
}

// TestExecuteAIAnalyzeProject tests ai_analyze_project tool
func TestExecuteAIAnalyzeProject(t *testing.T) {
	toolProvider, _ := createTestToolProvider()

	t.Run("AI project analysis placeholder", func(t *testing.T) {
		ctx := context.Background()
		args := map[string]interface{}{
			"project_id":     "BACKEND",
			"analysis_type":  "overview",
			"timeframe_days": 30,
		}

		result, err := toolProvider.ExecuteTool(ctx, "ai_analyze_project", args)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Error)
		assert.Contains(t, result.Content[0]["text"], "🤖 AI Project Analysis")
		assert.Contains(t, result.Content[0]["text"], "BACKEND")
		assert.Contains(t, result.Content[0]["text"], "🚧 AI analysis feature is under development")
	})

	t.Run("Missing project ID", func(t *testing.T) {
		ctx := context.Background()
		args := map[string]interface{}{
			"analysis_type": "overview",
		}

		result, err := toolProvider.ExecuteTool(ctx, "ai_analyze_project", args)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Error)
		assert.Contains(t, *result.Error, "Project ID is required")
	})
}

// TestExecuteAIExecuteTask tests ai_execute_task tool
func TestExecuteAIExecuteTask(t *testing.T) {
	toolProvider, _ := createTestToolProvider()

	t.Run("AI task execution placeholder", func(t *testing.T) {
		ctx := context.Background()
		args := map[string]interface{}{
			"task_id":           "PROJ-123",
			"execution_mode":    "plan",
			"auto_update_status": true,
			"create_subtasks":   false,
		}

		result, err := toolProvider.ExecuteTool(ctx, "ai_execute_task", args)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Error)
		assert.Contains(t, result.Content[0]["text"], "🤖 AI Task Execution")
		assert.Contains(t, result.Content[0]["text"], "PROJ-123")
		assert.Contains(t, result.Content[0]["text"], "🚧 AI execution feature is under development")
	})

	t.Run("Missing task ID", func(t *testing.T) {
		ctx := context.Background()
		args := map[string]interface{}{
			"execution_mode": "plan",
		}

		result, err := toolProvider.ExecuteTool(ctx, "ai_execute_task", args)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Error)
		assert.Contains(t, *result.Error, "Task ID is required")
	})
}

// TestExecuteUnknownTool tests unknown tool handling
func TestExecuteUnknownTool(t *testing.T) {
	toolProvider, _ := createTestToolProvider()

	ctx := context.Background()
	result, err := toolProvider.ExecuteTool(ctx, "unknown_tool", map[string]interface{}{})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Error)
	assert.Contains(t, *result.Error, "Unknown tool: unknown_tool")
}

// TestMapPriority tests priority mapping
func TestMapPriority(t *testing.T) {
	toolProvider, _ := createTestToolProvider()

	tests := []struct {
		input    string
		expected providers.TaskPriority
	}{
		{"lowest", providers.TaskPriorityLowest},
		{"low", providers.TaskPriorityLow},
		{"medium", providers.TaskPriorityMedium},
		{"high", providers.TaskPriorityHigh},
		{"highest", providers.TaskPriorityHighest},
		{"critical", providers.TaskPriorityCritical},
		{"invalid", providers.TaskPriorityMedium}, // Default fallback
		{"", providers.TaskPriorityMedium},        // Empty fallback
	}

	for _, test := range tests {
		actual := toolProvider.mapPriority(test.input)
		assert.Equal(t, test.expected, actual, "Priority mapping for '%s'", test.input)
	}
}

// TestFormatMethods tests various formatting methods
func TestFormatMethods(t *testing.T) {
	toolProvider, _ := createTestToolProvider()

	t.Run("Format providers table", func(t *testing.T) {
		providers := map[string]*providers.ProviderInfo{
			"test-provider": {
				Name:         "test-provider",
				Type:         providers.ProviderTypeYouTrack,
				Enabled:      true,
				HealthStatus: providers.HealthStatusHealthy,
			},
		}

		result := toolProvider.formatProvidersTable(providers)
		assert.Contains(t, result, "NAME")
		assert.Contains(t, result, "TYPE")
		assert.Contains(t, result, "STATUS")
		assert.Contains(t, result, "test-provider")
	})

	t.Run("Format tasks table", func(t *testing.T) {
		tasks := []*providers.UniversalTask{
			{
				ID:       "test-1",
				Key:      "PROJ-001",
				Title:    "Test Task",
				Status:   providers.TaskStatus{Name: "Open"},
				Priority: providers.TaskPriorityMedium,
				ProviderName: "test-provider",
			},
		}

		result := toolProvider.formatTasksTable(tasks)
		assert.Contains(t, result, "ID")
		assert.Contains(t, result, "PROVIDER")
		assert.Contains(t, result, "TITLE")
		assert.Contains(t, result, "PROJ-001")
		assert.Contains(t, result, "Test Task")
	})

	t.Run("Format tasks summary", func(t *testing.T) {
		tasks := []*providers.UniversalTask{
			{
				Status:   providers.TaskStatus{Name: "Open"},
				Priority: providers.TaskPriorityHigh,
				ProviderName: "youtrack",
			},
			{
				Status:   providers.TaskStatus{Name: "Done"},
				Priority: providers.TaskPriorityMedium,
				ProviderName: "jira",
			},
		}

		result := toolProvider.formatTasksSummary(tasks)
		assert.Contains(t, result, "Task Summary")
		assert.Contains(t, result, "(2 total)")
		assert.Contains(t, result, "By Status:")
		assert.Contains(t, result, "By Priority:")
		assert.Contains(t, result, "By Provider:")
	})

	t.Run("Format search results", func(t *testing.T) {
		tasks := []*providers.UniversalTask{
			{
				Key:         "PROJ-001",
				Title:       "Search Result",
				Description: "Long description that should be truncated because it's very long and exceeds the limit",
				Status:      providers.TaskStatus{Name: "Open"},
				Priority:    providers.TaskPriorityHigh,
				ProviderName: "test-provider",
			},
		}

		result := toolProvider.formatTasksSearchResults(tasks, true)
		assert.Contains(t, result, "1. [PROJ-001] Search Result")
		assert.Contains(t, result, "Status: Open")
		assert.Contains(t, result, "Priority: high")
		assert.Contains(t, result, "Description:")
		assert.Contains(t, result, "...")
	})
}

// TestConcurrentToolExecution tests concurrent tool execution
func TestConcurrentToolExecution(t *testing.T) {
	toolProvider, mockRegistry := createTestToolProvider()

	// Setup mocks for concurrent execution
	mockProviders := map[string]*providers.ProviderInfo{
		"test-provider": {
			Name:         "test-provider",
			HealthStatus: providers.HealthStatusHealthy,
		},
	}

	mockRegistry.On("ListProviders").Return(mockProviders)

	// Execute tools concurrently
	concurrency := 5
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			ctx := context.Background()
			args := map[string]interface{}{
				"enabled_only":   false,
				"output_format": "table",
			}
			_, err := toolProvider.ExecuteTool(ctx, "providers_list", args)
			errors <- err
		}()
	}

	// Collect results
	for i := 0; i < concurrency; i++ {
		err := <-errors
		assert.NoError(t, err)
	}

	mockRegistry.AssertExpectations(t)
}

// BenchmarkToolExecution benchmarks tool execution
func BenchmarkToolExecution(b *testing.B) {
	toolProvider, mockRegistry := createTestToolProvider()

	mockProviders := map[string]*providers.ProviderInfo{
		"bench-provider": {
			Name:         "bench-provider",
			HealthStatus: providers.HealthStatusHealthy,
		},
	}

	mockRegistry.On("ListProviders").Return(mockProviders)

	ctx := context.Background()
	args := map[string]interface{}{
		"enabled_only":   false,
		"output_format": "table",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := toolProvider.ExecuteTool(ctx, "providers_list", args)
		if err != nil {
			b.Fatalf("Tool execution failed: %v", err)
		}
	}
}
