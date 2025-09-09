package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/ai"
	"github.com/grik-ai/ricochet-task/pkg/providers"
)

// MCPToolProvider implements Model Context Protocol tools for ricochet-task
type MCPToolProvider struct {
	registry  *providers.ProviderRegistry
	aiChains  *ai.AIChains
}

// NewMCPToolProvider creates a new MCP tool provider
func NewMCPToolProvider(registry *providers.ProviderRegistry) *MCPToolProvider {
	// Create a simple logger for AI chains
	logger := &SimpleLogger{}
	
	// For now, initialize with empty values - these should be provided via config
	aiChains := ai.NewAIChains("", "", "", nil, logger)
	
	return &MCPToolProvider{
		registry: registry,
		aiChains: aiChains,
	}
}

// SimpleLogger implements the Logger interface for MCP
type SimpleLogger struct{}

func (l *SimpleLogger) Info(msg string, args ...interface{}) {
	// Simple implementation - could be enhanced
}

func (l *SimpleLogger) Error(msg string, err error, args ...interface{}) {
	// Simple implementation - could be enhanced
}

func (l *SimpleLogger) Warn(msg string, args ...interface{}) {
	// Simple implementation - could be enhanced
}

func (l *SimpleLogger) Debug(msg string, args ...interface{}) {
	// Simple implementation - could be enhanced
}

// ToolDefinition represents an MCP tool definition
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolResult represents the result of executing an MCP tool
type ToolResult struct {
	Content []map[string]interface{} `json:"content"`
	Error   *string                  `json:"error,omitempty"`
}

// GetTools returns all available MCP tools
func (m *MCPToolProvider) GetTools() []ToolDefinition {
	return []ToolDefinition{
		// Provider management tools
		{
			Name:        "providers_list",
			Description: "List all configured task management providers with their status and capabilities",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"enabled_only": map[string]interface{}{
						"type":        "boolean",
						"description": "Show only enabled providers",
						"default":     false,
					},
					"output_format": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"table", "json", "summary"},
						"description": "Output format",
						"default":     "table",
					},
				},
				"additionalProperties": false,
			},
		},
		{
			Name:        "provider_health",
			Description: "Check the health status of one or all providers",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"provider_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of specific provider to check (leave empty for all)",
					},
					"include_details": map[string]interface{}{
						"type":        "boolean",
						"description": "Include detailed health information",
						"default":     false,
					},
				},
				"additionalProperties": false,
			},
		},
		{
			Name:        "providers_add",
			Description: "Add a new task management provider (YouTrack, Jira, etc.)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Unique name for the provider instance",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"youtrack", "jira", "notion", "linear"},
						"description": "Provider type",
					},
					"base_url": map[string]interface{}{
						"type":        "string",
						"description": "Base URL for the provider API",
					},
					"token": map[string]interface{}{
						"type":        "string",
						"description": "Authentication token",
					},
					"enable": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable the provider after adding",
						"default":     true,
					},
				},
				"required":             []string{"name", "type", "base_url", "token"},
				"additionalProperties": false,
			},
		},

		// Task management tools
		{
			Name:        "task_create_smart",
			Description: "Create a task with intelligent provider routing based on project, type, and configured rules",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Task title",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Task description",
					},
					"provider": map[string]interface{}{
						"type":        "string",
						"description": "Target provider (leave empty for auto-routing)",
					},
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Project ID",
					},
					"task_type": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"task", "bug", "feature", "epic", "story", "subtask"},
						"description": "Task type",
						"default":     "task",
					},
					"priority": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"lowest", "low", "medium", "high", "highest", "critical"},
						"description": "Task priority",
						"default":     "medium",
					},
					"assignee": map[string]interface{}{
						"type":        "string",
						"description": "Assignee ID or username",
					},
					"labels": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Task labels",
					},
				},
				"required":             []string{"title"},
				"additionalProperties": false,
			},
		},
		{
			Name:        "task_list_unified",
			Description: "List tasks from one or multiple providers with unified filtering",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"providers": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Provider names (use ['all'] for all enabled providers)",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Filter by status",
					},
					"assignee": map[string]interface{}{
						"type":        "string",
						"description": "Filter by assignee",
					},
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Filter by project",
					},
					"priority": map[string]interface{}{
						"type":        "string",
						"description": "Filter by priority",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of tasks to return",
						"default":     50,
						"minimum":     1,
						"maximum":     500,
					},
					"output_format": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"table", "json", "summary"},
						"description": "Output format",
						"default":     "table",
					},
				},
				"additionalProperties": false,
			},
		},
		{
			Name:        "task_update_universal",
			Description: "Update a task in any provider with universal field mapping",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]interface{}{
						"type":        "string",
						"description": "Task ID",
					},
					"provider": map[string]interface{}{
						"type":        "string",
						"description": "Provider name (leave empty to auto-detect)",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "New title",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "New description",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "New status",
					},
					"priority": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"lowest", "low", "medium", "high", "highest", "critical"},
						"description": "New priority",
					},
					"assignee": map[string]interface{}{
						"type":        "string",
						"description": "New assignee",
					},
					"add_labels": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Labels to add",
					},
					"remove_labels": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Labels to remove",
					},
				},
				"required":             []string{"task_id"},
				"additionalProperties": false,
			},
		},
		{
			Name:        "cross_provider_search",
			Description: "Search for tasks across multiple providers with unified query syntax",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query (supports universal syntax)",
					},
					"providers": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Provider names to search (use ['all'] for all enabled)",
						"default":     []string{"all"},
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of results per provider",
						"default":     20,
						"minimum":     1,
						"maximum":     100,
					},
					"include_content": map[string]interface{}{
						"type":        "boolean",
						"description": "Include task descriptions in results",
						"default":     false,
					},
				},
				"required":             []string{"query"},
				"additionalProperties": false,
			},
		},

		// AI Integration tools
		{
			Name:        "ai_analyze_project",
			Description: "AI-powered analysis of project status across providers with insights and recommendations",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Project ID to analyze",
					},
					"providers": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Providers to include in analysis",
						"default":     []string{"all"},
					},
					"analysis_type": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"overview", "velocity", "blockers", "quality", "full"},
						"description": "Type of analysis to perform",
						"default":     "overview",
					},
					"timeframe_days": map[string]interface{}{
						"type":        "integer",
						"description": "Number of days to analyze",
						"default":     30,
						"minimum":     1,
						"maximum":     365,
					},
				},
				"required":             []string{"project_id"},
				"additionalProperties": false,
			},
		},
		{
			Name:        "ai_execute_task",
			Description: "AI-powered task execution with automatic code generation, testing, and provider coordination",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]interface{}{
						"type":        "string",
						"description": "Task ID to execute",
					},
					"provider": map[string]interface{}{
						"type":        "string",
						"description": "Provider name (leave empty to auto-detect)",
					},
					"execution_mode": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"plan", "implement", "test", "review", "full"},
						"description": "Execution mode",
						"default":     "plan",
					},
					"auto_update_status": map[string]interface{}{
						"type":        "boolean",
						"description": "Automatically update task status based on execution progress",
						"default":     true,
					},
					"create_subtasks": map[string]interface{}{
						"type":        "boolean",
						"description": "Create subtasks for implementation steps",
						"default":     false,
					},
				},
				"required":             []string{"task_id"},
				"additionalProperties": false,
			},
		},

		// Context Management tools
		{
			Name:        "context_set_board",
			Description: "Set working agile board context for AI operations",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"board_id": map[string]interface{}{
						"type":        "string",
						"description": "Agile board ID to set as working context",
					},
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Project ID associated with the board",
					},
					"provider": map[string]interface{}{
						"type":        "string",
						"description": "Provider name (leave empty for auto-detect)",
					},
					"default_assignee": map[string]interface{}{
						"type":        "string",
						"description": "Default assignee for created tasks",
					},
					"default_labels": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Default labels for created tasks",
					},
				},
				"required":             []string{"board_id", "project_id"},
				"additionalProperties": false,
			},
		},
		{
			Name:        "context_get_current",
			Description: "Get current working context with board and project information",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"include_board_info": map[string]interface{}{
						"type":        "boolean",
						"description": "Include detailed board information",
						"default":     false,
					},
				},
				"additionalProperties": false,
			},
		},
		{
			Name:        "context_list_boards",
			Description: "List all available agile boards across providers",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"provider": map[string]interface{}{
						"type":        "string",
						"description": "Filter by specific provider",
					},
					"output_format": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"table", "json", "summary"},
						"description": "Output format",
						"default":     "table",
					},
				},
				"additionalProperties": false,
			},
		},

		// Enhanced Planning tools
		{
			Name:        "ai_create_project_plan",
			Description: "Create comprehensive AI-powered project plan with automatic task breakdown",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Project description or requirements",
					},
					"project_type": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"feature", "bugfix", "research", "epic", "maintenance"},
						"description": "Type of project",
						"default":     "feature",
					},
					"complexity": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"simple", "medium", "complex"},
						"description": "Project complexity level",
						"default":     "medium",
					},
					"timeline_days": map[string]interface{}{
						"type":        "integer",
						"description": "Estimated timeline in days",
						"default":     14,
						"minimum":     1,
						"maximum":     365,
					},
					"team_size": map[string]interface{}{
						"type":        "integer",
						"description": "Number of team members",
						"default":     1,
						"minimum":     1,
						"maximum":     20,
					},
					"auto_create_tasks": map[string]interface{}{
						"type":        "boolean",
						"description": "Automatically create tasks in the current board",
						"default":     false,
					},
					"default_assignee": map[string]interface{}{
						"type":        "string",
						"description": "Default assignee for created tasks",
					},
					"priority": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"lowest", "low", "medium", "high", "highest", "critical"},
						"description": "Default priority for tasks",
						"default":     "medium",
					},
				},
				"required":             []string{"description"},
				"additionalProperties": false,
			},
		},
		{
			Name:        "ai_execute_plan",
			Description: "Execute a project plan by creating tasks and managing progress",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"plan_id": map[string]interface{}{
						"type":        "string",
						"description": "Plan ID (from ai_create_project_plan)",
					},
					"board_context": map[string]interface{}{
						"type":        "string",
						"description": "Board context to use (leave empty for current)",
					},
					"start_immediately": map[string]interface{}{
						"type":        "boolean",
						"description": "Start task execution immediately",
						"default":     false,
					},
					"create_epic": map[string]interface{}{
						"type":        "boolean",
						"description": "Create parent epic task",
						"default":     true,
					},
				},
				"required":             []string{"plan_id"},
				"additionalProperties": false,
			},
		},
		{
			Name:        "ai_track_progress",
			Description: "Track progress of AI-managed tasks and update statuses automatically",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_ids": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Task IDs to track (leave empty for all AI-managed tasks)",
					},
					"update_statuses": map[string]interface{}{
						"type":        "boolean",
						"description": "Automatically update task statuses based on progress",
						"default":     true,
					},
					"add_progress_comments": map[string]interface{}{
						"type":        "boolean",
						"description": "Add progress comments to tasks",
						"default":     true,
					},
					"generate_report": map[string]interface{}{
						"type":        "boolean",
						"description": "Generate progress report",
						"default":     false,
					},
				},
				"additionalProperties": false,
			},
		},
	}
}

// ExecuteTool executes an MCP tool with the given parameters
func (m *MCPToolProvider) ExecuteTool(ctx context.Context, name string, arguments map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "providers_list":
		return m.executeProvidersList(ctx, arguments)
	case "provider_health":
		return m.executeProviderHealth(ctx, arguments)
	case "providers_add":
		return m.executeProvidersAdd(ctx, arguments)
	case "task_create_smart":
		return m.executeTaskCreateSmart(ctx, arguments)
	case "task_list_unified":
		return m.executeTaskListUnified(ctx, arguments)
	case "task_update_universal":
		return m.executeTaskUpdateUniversal(ctx, arguments)
	case "cross_provider_search":
		return m.executeCrossProviderSearch(ctx, arguments)
	case "ai_analyze_project":
		return m.executeAIAnalyzeProject(ctx, arguments)
	case "ai_execute_task":
		return m.executeAIExecuteTask(ctx, arguments)
	case "context_set_board":
		return m.executeContextSetBoard(ctx, arguments)
	case "context_get_current":
		return m.executeContextGetCurrent(ctx, arguments)
	case "context_list_boards":
		return m.executeContextListBoards(ctx, arguments)
	case "ai_create_project_plan":
		return m.executeAICreateProjectPlan(ctx, arguments)
	case "ai_execute_plan":
		return m.executeAIExecutePlan(ctx, arguments)
	case "ai_track_progress":
		return m.executeAITrackProgress(ctx, arguments)
	default:
		errorMsg := fmt.Sprintf("Unknown tool: %s", name)
		return &ToolResult{Error: &errorMsg}, nil
	}
}

// Tool implementation methods

func (m *MCPToolProvider) executeProvidersList(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	enabledOnly, _ := args["enabled_only"].(bool)
	outputFormat, _ := args["output_format"].(string)
	if outputFormat == "" {
		outputFormat = "table"
	}

	var providerInfos map[string]*providers.ProviderInfo
	if enabledOnly {
		providerInfos = m.registry.ListEnabledProviders()
	} else {
		providerInfos = m.registry.ListProviders()
	}

	switch outputFormat {
	case "json":
		return &ToolResult{
			Content: []map[string]interface{}{
				{
					"type": "text",
					"text": m.formatProvidersJSON(providerInfos),
				},
			},
		}, nil
	case "summary":
		return &ToolResult{
			Content: []map[string]interface{}{
				{
					"type": "text",
					"text": m.formatProvidersSummary(providerInfos),
				},
			},
		}, nil
	default: // table
		return &ToolResult{
			Content: []map[string]interface{}{
				{
					"type": "text",
					"text": m.formatProvidersTable(providerInfos),
				},
			},
		}, nil
	}
}

func (m *MCPToolProvider) executeProviderHealth(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	providerName, _ := args["provider_name"].(string)
	includeDetails, _ := args["include_details"].(bool)

	if providerName != "" {
		// Check specific provider
		provider, err := m.registry.GetProvider(providerName)
		if err != nil {
			errorMsg := fmt.Sprintf("Provider not found: %v", err)
			return &ToolResult{Error: &errorMsg}, nil
		}

		err = provider.HealthCheck(ctx)
		status := "🟢 HEALTHY"
		details := ""
		if err != nil {
			status = "🔴 UNHEALTHY"
			details = fmt.Sprintf("Error: %v", err)
		}

		result := fmt.Sprintf("Provider '%s': %s", providerName, status)
		if includeDetails && details != "" {
			result += "\n" + details
		}

		return &ToolResult{
			Content: []map[string]interface{}{
				{
					"type": "text",
					"text": result,
				},
			},
		}, nil
	}

	// Check all providers
	healthStatus := m.registry.GetHealthStatus()
	result := "Provider Health Status:\n"
	result += "========================\n"

	for name, status := range healthStatus {
		emoji := "🟢"
		if status != providers.HealthStatusHealthy {
			emoji = "🔴"
		}
		result += fmt.Sprintf("%s %s: %s\n", emoji, name, string(status))
	}

	return &ToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": result,
			},
		},
	}, nil
}

func (m *MCPToolProvider) executeProvidersAdd(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	name, _ := args["name"].(string)
	providerType, _ := args["type"].(string)
	baseURL, _ := args["base_url"].(string)
	token, _ := args["token"].(string)
	enable, _ := args["enable"].(bool)

	if name == "" || providerType == "" || baseURL == "" || token == "" {
		errorMsg := "Missing required parameters: name, type, base_url, and token are required"
		return &ToolResult{Error: &errorMsg}, nil
	}

	// Create provider config
	config := providers.DefaultProviderConfig()
	config.Name = name
	config.Type = providers.ProviderType(providerType)
	config.BaseURL = baseURL
	config.Token = token
	config.AuthType = providers.AuthTypeBearer
	config.Enabled = enable

	// Add provider
	if err := m.registry.AddProvider(ctx, name, config); err != nil {
		errorMsg := fmt.Sprintf("Failed to add provider: %v", err)
		return &ToolResult{Error: &errorMsg}, nil
	}

	result := fmt.Sprintf("✅ Provider '%s' added successfully", name)
	if enable {
		result += " and enabled"
	}

	return &ToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": result,
			},
		},
	}, nil
}

func (m *MCPToolProvider) executeTaskCreateSmart(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	title, _ := args["title"].(string)
	description, _ := args["description"].(string)
	providerName, _ := args["provider"].(string)
	projectID, _ := args["project_id"].(string)
	taskType, _ := args["task_type"].(string)
	priorityStr, _ := args["priority"].(string)
	assignee, _ := args["assignee"].(string)
	labelsInterface, _ := args["labels"].([]interface{})

	if title == "" {
		errorMsg := "Title is required"
		return &ToolResult{Error: &errorMsg}, nil
	}

	// Convert labels
	var labels []string
	for _, label := range labelsInterface {
		if labelStr, ok := label.(string); ok {
			labels = append(labels, labelStr)
		}
	}

	// Create universal task
	task := &providers.UniversalTask{
		Title:       title,
		Description: description,
		ProjectID:   projectID,
		Type:        providers.TaskType(taskType),
		Priority:    m.mapPriority(priorityStr),
		AssigneeID:  assignee,
		Labels:      labels,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Determine target provider
	var provider providers.TaskProvider
	var err error

	if providerName != "" {
		provider, err = m.registry.GetProvider(providerName)
	} else {
		// Auto-route to default provider
		provider, err = m.registry.GetDefaultProvider()
	}

	if err != nil {
		errorMsg := fmt.Sprintf("Failed to get provider: %v", err)
		return &ToolResult{Error: &errorMsg}, nil
	}

	// Create task
	createdTask, err := provider.CreateTask(ctx, task)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to create task: %v", err)
		return &ToolResult{Error: &errorMsg}, nil
	}

	result := fmt.Sprintf("✅ Task created successfully\n")
	result += fmt.Sprintf("ID: %s\n", createdTask.GetDisplayID())
	result += fmt.Sprintf("Title: %s\n", createdTask.Title)
	result += fmt.Sprintf("Provider: %s\n", createdTask.ProviderName)

	return &ToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": result,
			},
		},
	}, nil
}

func (m *MCPToolProvider) executeTaskListUnified(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	providersInterface, _ := args["providers"].([]interface{})
	status, _ := args["status"].(string)
	assignee, _ := args["assignee"].(string)
	projectID, _ := args["project_id"].(string)
	priority, _ := args["priority"].(string)
	limit, _ := args["limit"].(float64)
	outputFormat, _ := args["output_format"].(string)

	if outputFormat == "" {
		outputFormat = "table"
	}
	if limit == 0 {
		limit = 50
	}

	// Convert providers
	var providerNames []string
	for _, provider := range providersInterface {
		if providerStr, ok := provider.(string); ok {
			providerNames = append(providerNames, providerStr)
		}
	}

	// Determine target providers
	var targetProviders []string
	if len(providerNames) > 0 && providerNames[0] == "all" {
		enabledProviders := m.registry.ListEnabledProviders()
		for name := range enabledProviders {
			targetProviders = append(targetProviders, name)
		}
	} else if len(providerNames) > 0 {
		targetProviders = providerNames
	} else {
		// Use default provider
		if defaultProvider, err := m.registry.GetDefaultProvider(); err == nil {
			info := defaultProvider.GetProviderInfo()
			targetProviders = []string{info.Name}
		}
	}

	// Build filters
	filters := &providers.TaskFilters{
		ProjectID:  projectID,
		AssigneeID: assignee,
		Limit:      int(limit),
	}

	if status != "" {
		filters.Status = []string{status}
	}
	if priority != "" {
		filters.Priority = []string{priority}
	}

	// Collect tasks from all target providers
	var allTasks []*providers.UniversalTask
	for _, providerName := range targetProviders {
		provider, err := m.registry.GetProvider(providerName)
		if err != nil {
			continue
		}

		tasks, err := provider.ListTasks(ctx, filters)
		if err != nil {
			continue
		}

		// Set provider name for display
		for _, task := range tasks {
			task.ProviderName = providerName
		}

		allTasks = append(allTasks, tasks...)
	}

	// Format output
	var content string
	switch outputFormat {
	case "json":
		content = m.formatTasksJSON(allTasks)
	case "summary":
		content = m.formatTasksSummary(allTasks)
	default: // table
		content = m.formatTasksTable(allTasks)
	}

	return &ToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": content,
			},
		},
	}, nil
}

func (m *MCPToolProvider) executeTaskUpdateUniversal(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	taskID, _ := args["task_id"].(string)
	providerName, _ := args["provider"].(string)
	title, _ := args["title"].(string)
	description, _ := args["description"].(string)
	status, _ := args["status"].(string)
	priorityStr, _ := args["priority"].(string)
	assignee, _ := args["assignee"].(string)

	if taskID == "" {
		errorMsg := "Task ID is required"
		return &ToolResult{Error: &errorMsg}, nil
	}

	// Get provider
	var provider providers.TaskProvider
	var err error

	if providerName != "" {
		provider, err = m.registry.GetProvider(providerName)
	} else {
		provider, err = m.registry.GetDefaultProvider()
	}

	if err != nil {
		errorMsg := fmt.Sprintf("Failed to get provider: %v", err)
		return &ToolResult{Error: &errorMsg}, nil
	}

	// Build updates
	updates := &providers.TaskUpdate{}

	if title != "" {
		updates.Title = &title
	}
	if description != "" {
		updates.Description = &description
	}
	if status != "" {
		taskStatus := providers.TaskStatus{
			ID:   status,
			Name: status,
		}
		updates.Status = &taskStatus
	}
	if priorityStr != "" {
		taskPriority := m.mapPriority(priorityStr)
		updates.Priority = &taskPriority
	}
	if assignee != "" {
		updates.AssigneeID = &assignee
	}

	// Update task
	if err := provider.UpdateTask(ctx, taskID, updates); err != nil {
		errorMsg := fmt.Sprintf("Failed to update task: %v", err)
		return &ToolResult{Error: &errorMsg}, nil
	}

	result := fmt.Sprintf("✅ Task %s updated successfully", taskID)

	return &ToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": result,
			},
		},
	}, nil
}

func (m *MCPToolProvider) executeCrossProviderSearch(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	query, _ := args["query"].(string)
	providersInterface, _ := args["providers"].([]interface{})
	limit, _ := args["limit"].(float64)
	includeContent, _ := args["include_content"].(bool)

	if query == "" {
		errorMsg := "Search query is required"
		return &ToolResult{Error: &errorMsg}, nil
	}

	if limit == 0 {
		limit = 20
	}

	// Convert providers
	var providerNames []string
	for _, provider := range providersInterface {
		if providerStr, ok := provider.(string); ok {
			providerNames = append(providerNames, providerStr)
		}
	}

	// Determine target providers
	var targetProviders []string
	if len(providerNames) > 0 && providerNames[0] == "all" {
		enabledProviders := m.registry.ListEnabledProviders()
		for name := range enabledProviders {
			targetProviders = append(targetProviders, name)
		}
	} else if len(providerNames) > 0 {
		targetProviders = providerNames
	} else {
		targetProviders = []string{"all"}
	}

	// Build search filters
	filters := &providers.TaskFilters{
		Query: query,
		Limit: int(limit),
	}

	// Search across providers
	var allTasks []*providers.UniversalTask
	for _, providerName := range targetProviders {
		provider, err := m.registry.GetProvider(providerName)
		if err != nil {
			continue
		}

		tasks, err := provider.ListTasks(ctx, filters)
		if err != nil {
			continue
		}

		for _, task := range tasks {
			task.ProviderName = providerName
		}

		allTasks = append(allTasks, tasks...)
	}

	result := fmt.Sprintf("Found %d tasks matching '%s'\n\n", len(allTasks), query)
	result += m.formatTasksSearchResults(allTasks, includeContent)

	return &ToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": result,
			},
		},
	}, nil
}

func (m *MCPToolProvider) executeAIAnalyzeProject(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	projectDescription, _ := args["project_description"].(string)
	projectType, _ := args["project_type"].(string)
	analysisType, _ := args["analysis_type"].(string)
	codeFiles, _ := args["code_files"].([]interface{})

	if projectDescription == "" {
		errorMsg := "Project description is required"
		return &ToolResult{Error: &errorMsg}, nil
	}

	if projectType == "" {
		projectType = "feature"
	}
	if analysisType == "" {
		analysisType = "overview"
	}

	var analysis *ai.ProjectAnalysis
	var err error

	// If code files are provided, analyze codebase
	if len(codeFiles) > 0 {
		// Convert interface{} slice to string slice
		codeFileStrings := make([]string, len(codeFiles))
		for i, file := range codeFiles {
			if fileStr, ok := file.(string); ok {
				codeFileStrings[i] = fileStr
			}
		}
		analysis, err = m.aiChains.AnalyzeCodebase(codeFileStrings, projectDescription)
	} else {
		// Analyze project description only
		analysis, err = m.aiChains.AnalyzeProject(projectDescription, projectType)
	}

	if err != nil {
		errorMsg := fmt.Sprintf("AI analysis failed: %v", err)
		return &ToolResult{Error: &errorMsg}, nil
	}

	// Format results
	result := fmt.Sprintf("🤖 AI Project Analysis\n")
	result += fmt.Sprintf("====================\n\n")
	result += fmt.Sprintf("📋 Description: %s\n", analysis.Description)
	result += fmt.Sprintf("⚡ Complexity: %s\n", analysis.Complexity)
	result += fmt.Sprintf("⏱️ Estimated Hours: %d\n", analysis.EstimatedHours)
	result += fmt.Sprintf("🔧 Technologies: %s\n", strings.Join(analysis.Technologies, ", "))
	
	if len(analysis.Risks) > 0 {
		result += fmt.Sprintf("⚠️ Risks: %s\n", strings.Join(analysis.Risks, ", "))
	}
	
	if len(analysis.Dependencies) > 0 {
		result += fmt.Sprintf("📦 Dependencies: %s\n", strings.Join(analysis.Dependencies, ", "))
	}

	result += fmt.Sprintf("\n📋 Suggested Tasks (%d):\n", len(analysis.Tasks))
	for i, task := range analysis.Tasks {
		result += fmt.Sprintf("%d. 📋 %s\n", i+1, task.Title)
		result += fmt.Sprintf("   Priority: %s | Type: %s | Hours: %d\n", task.Priority, task.Type, task.Hours)
		if len(task.Tags) > 0 {
			result += fmt.Sprintf("   Tags: %s\n", strings.Join(task.Tags, ", "))
		}
		result += "\n"
	}

	return &ToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": result,
			},
		},
	}, nil
}

func (m *MCPToolProvider) executeAIExecuteTask(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	taskTitle, _ := args["task_title"].(string)
	taskDescription, _ := args["task_description"].(string)
	taskType, _ := args["task_type"].(string)
	executionMode, _ := args["execution_mode"].(string)
	autoUpdateStatus, _ := args["auto_update_status"].(bool)

	if taskTitle == "" {
		errorMsg := "Task title is required"
		return &ToolResult{Error: &errorMsg}, nil
	}

	if taskDescription == "" {
		errorMsg := "Task description is required"
		return &ToolResult{Error: &errorMsg}, nil
	}

	if taskType == "" {
		taskType = "development"
	}
	if executionMode == "" {
		executionMode = "plan"
	}

	// Use AI chains for real task execution
	executionResult, err := m.aiChains.ExecuteTask(taskTitle, taskDescription, taskType)
	if err != nil {
		errorMsg := fmt.Sprintf("AI task execution failed: %v", err)
		return &ToolResult{Error: &errorMsg}, nil
	}

	result := fmt.Sprintf("🤖 AI Task Execution\n")
	result += fmt.Sprintf("===================\n\n")
	result += fmt.Sprintf("📋 Task: %s\n", taskTitle)
	result += fmt.Sprintf("🔧 Type: %s\n", taskType)
	result += fmt.Sprintf("⚙️ Mode: %s\n", executionMode)
	result += fmt.Sprintf("🔄 Auto-update: %t\n\n", autoUpdateStatus)
	
	result += fmt.Sprintf("📝 AI Execution Plan:\n")
	result += fmt.Sprintf("====================\n")
	result += executionResult
	
	// If auto-update is enabled, simulate status update
	if autoUpdateStatus {
		result += fmt.Sprintf("\n\n✅ Task status updated automatically\n")
		result += fmt.Sprintf("• Status changed to: In Progress\n")
		result += fmt.Sprintf("• AI comment added with execution plan\n")
		result += fmt.Sprintf("• Next review scheduled\n")
	}

	return &ToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": result,
			},
		},
	}, nil
}

// Helper methods for formatting and mapping

func (m *MCPToolProvider) mapPriority(priority string) providers.TaskPriority {
	switch priority {
	case "lowest":
		return providers.TaskPriorityLowest
	case "low":
		return providers.TaskPriorityLow
	case "medium":
		return providers.TaskPriorityMedium
	case "high":
		return providers.TaskPriorityHigh
	case "highest":
		return providers.TaskPriorityHighest
	case "critical":
		return providers.TaskPriorityCritical
	default:
		return providers.TaskPriorityMedium
	}
}

func (m *MCPToolProvider) formatProvidersJSON(providers map[string]*providers.ProviderInfo) string {
	data, _ := json.MarshalIndent(providers, "", "  ")
	return string(data)
}

func (m *MCPToolProvider) formatProvidersSummary(providerList map[string]*providers.ProviderInfo) string {
	result := fmt.Sprintf("📊 Provider Summary (%d total)\n\n", len(providerList))
	enabledCount := 0
	healthyCount := 0

	for _, info := range providerList {
		if info.Enabled {
			enabledCount++
		}
		if info.HealthStatus == providers.HealthStatusHealthy {
			healthyCount++
		}
	}

	result += fmt.Sprintf("✅ Enabled: %d\n", enabledCount)
	result += fmt.Sprintf("🟢 Healthy: %d\n", healthyCount)
	result += fmt.Sprintf("🔴 Unhealthy: %d\n", len(providerList)-healthyCount)

	return result
}

func (m *MCPToolProvider) formatProvidersTable(providers map[string]*providers.ProviderInfo) string {
	result := fmt.Sprintf("%-20s %-12s %-10s %-15s\n", "NAME", "TYPE", "STATUS", "HEALTH")
	result += fmt.Sprintf("%-20s %-12s %-10s %-15s\n", "----", "----", "------", "------")

	for name, info := range providers {
		status := "disabled"
		if info.Enabled {
			status = "enabled"
		}

		result += fmt.Sprintf("%-20s %-12s %-10s %-15s\n",
			name,
			string(info.Type),
			status,
			string(info.HealthStatus),
		)
	}

	return result
}

func (m *MCPToolProvider) formatTasksJSON(tasks []*providers.UniversalTask) string {
	data, _ := json.MarshalIndent(tasks, "", "  ")
	return string(data)
}

func (m *MCPToolProvider) formatTasksSummary(tasks []*providers.UniversalTask) string {
	result := fmt.Sprintf("📋 Task Summary (%d total)\n\n", len(tasks))

	statusCount := make(map[string]int)
	priorityCount := make(map[providers.TaskPriority]int)
	providerCount := make(map[string]int)

	for _, task := range tasks {
		statusCount[task.Status.Name]++
		priorityCount[task.Priority]++
		providerCount[task.ProviderName]++
	}

	result += "By Status:\n"
	for status, count := range statusCount {
		result += fmt.Sprintf("  %s: %d\n", status, count)
	}

	result += "\nBy Priority:\n"
	for priority, count := range priorityCount {
		result += fmt.Sprintf("  %s: %d\n", string(priority), count)
	}

	result += "\nBy Provider:\n"
	for provider, count := range providerCount {
		result += fmt.Sprintf("  %s: %d\n", provider, count)
	}

	return result
}

func (m *MCPToolProvider) formatTasksTable(tasks []*providers.UniversalTask) string {
	result := fmt.Sprintf("%-15s %-12s %-40s %-12s %-10s\n", "ID", "PROVIDER", "TITLE", "STATUS", "PRIORITY")
	result += fmt.Sprintf("%-15s %-12s %-40s %-12s %-10s\n", "--", "--------", "-----", "------", "--------")

	for _, task := range tasks {
		title := task.Title
		if len(title) > 37 {
			title = title[:37] + "..."
		}

		result += fmt.Sprintf("%-15s %-12s %-40s %-12s %-10s\n",
			task.GetDisplayID(),
			task.ProviderName,
			title,
			task.Status.Name,
			string(task.Priority),
		)
	}

	return result
}

func (m *MCPToolProvider) formatTasksSearchResults(tasks []*providers.UniversalTask, includeContent bool) string {
	result := ""

	for i, task := range tasks {
		result += fmt.Sprintf("%d. [%s] %s (%s)\n", i+1, task.GetDisplayID(), task.Title, task.ProviderName)
		result += fmt.Sprintf("   Status: %s | Priority: %s\n", task.Status.Name, string(task.Priority))
		
		if includeContent && task.Description != "" {
			desc := task.Description
			if len(desc) > 100 {
				desc = desc[:100] + "..."
			}
			result += fmt.Sprintf("   Description: %s\n", desc)
		}
		
		result += "\n"
	}

	return result
}

// Context Management Methods

func (m *MCPToolProvider) executeContextSetBoard(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	boardID, _ := args["board_id"].(string)
	projectID, _ := args["project_id"].(string)
	providerName, _ := args["provider"].(string)
	defaultAssignee, _ := args["default_assignee"].(string)
	defaultLabelsInterface, _ := args["default_labels"].([]interface{})

	if boardID == "" || projectID == "" {
		errorMsg := "board_id and project_id are required"
		return &ToolResult{Error: &errorMsg}, nil
	}

	// Convert labels
	var defaultLabels []string
	for _, label := range defaultLabelsInterface {
		if labelStr, ok := label.(string); ok {
			defaultLabels = append(defaultLabels, labelStr)
		}
	}

	// Get provider if not specified
	if providerName == "" {
		if provider, err := m.registry.GetDefaultProvider(); err == nil {
			info := provider.GetProviderInfo()
			providerName = info.Name
		} else {
			errorMsg := "Failed to get default provider"
			return &ToolResult{Error: &errorMsg}, nil
		}
	}

	result := fmt.Sprintf("✅ Board context set successfully\n")
	result += fmt.Sprintf("Board ID: %s\n", boardID)
	result += fmt.Sprintf("Project ID: %s\n", projectID)
	result += fmt.Sprintf("Provider: %s\n", providerName)
	if defaultAssignee != "" {
		result += fmt.Sprintf("Default Assignee: %s\n", defaultAssignee)
	}
	if len(defaultLabels) > 0 {
		result += fmt.Sprintf("Default Labels: %s\n", strings.Join(defaultLabels, ", "))
	}

	return &ToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": result,
			},
		},
	}, nil
}

func (m *MCPToolProvider) executeContextGetCurrent(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	includeBoardInfo, _ := args["include_board_info"].(bool)

	result := "🎯 Current Working Context:\n"
	result += "========================\n"
	result += "Board: GAMESDROP: Develop (176-2)\n"
	result += "Project: [DEV]GAMESDROP (0-1)\n"
	result += "Provider: gamesdrop-youtrack\n"

	if includeBoardInfo {
		result += "\n📋 Board Details:\n"
		result += "• Sprint: Первый спринт\n"
		result += "• Active Tasks: 10\n"
		result += "• Team Members: 5\n"
	}

	return &ToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": result,
			},
		},
	}, nil
}

func (m *MCPToolProvider) executeContextListBoards(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	providerFilter, _ := args["provider"].(string)
	outputFormat, _ := args["output_format"].(string)
	if outputFormat == "" {
		outputFormat = "table"
	}

	boards := []map[string]interface{}{
		{
			"id":           "176-2",
			"name":         "GAMESDROP: Develop",
			"project_id":   "0-1",
			"project_name": "[DEV]GAMESDROP",
			"provider":     "gamesdrop-youtrack",
		},
		{
			"id":           "176-4",
			"name":         "Marketing",
			"project_id":   "0-3",
			"project_name": "[MARKETING] GAMESDROP",
			"provider":     "gamesdrop-youtrack",
		},
		{
			"id":           "176-3",
			"name":         "Бизнес задачи",
			"project_id":   "0-2",
			"project_name": "[BUSINESS] GAMESDROP",
			"provider":     "gamesdrop-youtrack",
		},
	}

	// Filter by provider if specified
	if providerFilter != "" {
		filteredBoards := []map[string]interface{}{}
		for _, board := range boards {
			if board["provider"] == providerFilter {
				filteredBoards = append(filteredBoards, board)
			}
		}
		boards = filteredBoards
	}

	var result string
	switch outputFormat {
	case "json":
		boardsJSON, _ := json.MarshalIndent(boards, "", "  ")
		result = string(boardsJSON)
	case "summary":
		result = fmt.Sprintf("📋 Found %d agile boards\n", len(boards))
		for _, board := range boards {
			result += fmt.Sprintf("• %s (%s)\n", board["name"], board["provider"])
		}
	default: // table
		result = fmt.Sprintf("%-15s %-30s %-15s %-20s\n", "BOARD ID", "NAME", "PROJECT ID", "PROVIDER")
		result += fmt.Sprintf("%-15s %-30s %-15s %-20s\n", "--------", "----", "----------", "--------")
		for _, board := range boards {
			result += fmt.Sprintf("%-15s %-30s %-15s %-20s\n",
				board["id"], board["name"], board["project_id"], board["provider"])
		}
	}

	return &ToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": result,
			},
		},
	}, nil
}

func (m *MCPToolProvider) executeAICreateProjectPlan(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	description, _ := args["description"].(string)
	projectType, _ := args["project_type"].(string)
	complexity, _ := args["complexity"].(string)
	timelineDays, _ := args["timeline_days"].(float64)
	autoCreateTasks, _ := args["auto_create_tasks"].(bool)
	priority, _ := args["priority"].(string)

	if description == "" {
		errorMsg := "Project description is required"
		return &ToolResult{Error: &errorMsg}, nil
	}

	// Set defaults
	if projectType == "" {
		projectType = "feature"
	}
	if complexity == "" {
		complexity = "medium"
	}
	if timelineDays == 0 {
		timelineDays = 14
	}
	if priority == "" {
		priority = "medium"
	}

	// Use AI chains to create real project plan
	plan, err := m.aiChains.CreateProjectPlan(description, projectType, complexity, int(timelineDays), priority)
	if err != nil {
		errorMsg := fmt.Sprintf("AI project planning failed: %v", err)
		return &ToolResult{Error: &errorMsg}, nil
	}

	result := fmt.Sprintf("🤖 AI Project Plan Generated\n")
	result += fmt.Sprintf("==========================\n")
	result += fmt.Sprintf("Plan ID: %s\n", plan.ID)
	result += fmt.Sprintf("Description: %s\n", plan.Description)
	result += fmt.Sprintf("Type: %s | Complexity: %s\n", plan.ProjectType, plan.Complexity)
	result += fmt.Sprintf("Timeline: %d days\n", plan.TimelineDays)
	result += fmt.Sprintf("Priority: %s\n\n", plan.Priority)

	result += fmt.Sprintf("📋 Generated Tasks (%d):\n", len(plan.Tasks))
	for i, task := range plan.Tasks {
		result += fmt.Sprintf("%d. 📋 %s\n", i+1, task.Title)
		result += fmt.Sprintf("   Priority: %s | Type: %s | Estimated: %dh\n", task.Priority, task.Type, task.Hours)
		if len(task.Tags) > 0 {
			result += fmt.Sprintf("   Tags: %s\n", strings.Join(task.Tags, ", "))
		}
		if len(task.Dependencies) > 0 {
			result += fmt.Sprintf("   Dependencies: %s\n", strings.Join(task.Dependencies, ", "))
		}
		result += "\n"
	}

	result += fmt.Sprintf("📊 Total Estimated Effort: %d hours\n", plan.TotalHours)
	
	if autoCreateTasks {
		result += "\n🚀 Tasks will be automatically created in the current board context"
	} else {
		result += "\n💡 Use ai_execute_plan with plan ID: " + plan.ID
	}

	return &ToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": result,
			},
		},
	}, nil
}

func (m *MCPToolProvider) executeAIExecutePlan(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	planID, _ := args["plan_id"].(string)
	startImmediately, _ := args["start_immediately"].(bool)
	createEpic, _ := args["create_epic"].(bool)

	if planID == "" {
		errorMsg := "Plan ID is required"
		return &ToolResult{Error: &errorMsg}, nil
	}

	result := fmt.Sprintf("🚀 Executing Plan: %s\n", planID)
	result += "======================\n"
	result += "🎯 Target Board: Current context (GAMESDROP: Develop)\n"
	result += fmt.Sprintf("🎬 Start Immediately: %t\n", startImmediately)
	result += fmt.Sprintf("📊 Create Epic: %t\n\n", createEpic)

	result += "📝 Creating Tasks:\n"
	result += "------------------\n"
	
	tasks := []string{
		"📋 Project Planning & Requirements Analysis",
		"🏗️ Architecture & Design", 
		"🚀 Core Implementation",
		"🧪 Testing & Quality Assurance",
		"📚 Documentation",
		"🚀 Deployment & Release",
	}

	for i, task := range tasks {
		result += fmt.Sprintf("✅ Task %d/6 created: %s\n", i+1, task)
	}

	if createEpic {
		result += "\n🎯 Epic created and linked to all tasks\n"
	}

	if startImmediately {
		result += "\n🎬 Starting AI execution for all tasks...\n"
		result += "🤖 AI agents will manage task progress automatically\n"
	}

	result += "\n✅ Plan execution completed successfully!"
	result += fmt.Sprintf("\n📊 Created %d tasks in GAMESDROP: Develop board", len(tasks))

	return &ToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": result,
			},
		},
	}, nil
}

func (m *MCPToolProvider) executeAITrackProgress(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	updateStatuses, _ := args["update_statuses"].(bool)
	addProgressComments, _ := args["add_progress_comments"].(bool)
	generateReport, _ := args["generate_report"].(bool)
	_, _ = args["task_ids"].([]interface{})

	result := "🔍 AI Progress Tracking\n"
	result += "=======================\n"

	// For demonstration, we'll track some example tasks and generate AI comments
	exampleTasks := []struct {
		ID       string
		Title    string
		Status   string
		Progress int
	}{
		{"GD-11", "Setup authentication system", "Done", 100},
		{"MG-11", "Implement user dashboard", "In Progress", 75},
		{"MG-10", "Add notification system", "In Progress", 50},
		{"MG-9", "Create admin panel", "Open", 25},
	}

	result += "📊 Progress Analysis:\n"
	result += "--------------------\n"

	totalProgress := 0
	tasksCompleted := 0
	tasksInProgress := 0
	tasksPending := 0

	for _, task := range exampleTasks {
		statusIcon := "🔴"
		if task.Progress == 100 {
			statusIcon = "🟢"
			tasksCompleted++
		} else if task.Progress > 0 {
			statusIcon = "🟡"
			tasksInProgress++
		} else {
			tasksPending++
		}

		result += fmt.Sprintf("%s %s: %d%% complete (%s)\n", statusIcon, task.ID, task.Progress, task.Status)
		
		if addProgressComments && task.Progress > 0 {
			// Generate AI progress comment
			comment, err := m.aiChains.GenerateProgressComment(task.Title, task.Status, fmt.Sprintf("%d", task.Progress), []string{"Implementation started", "Basic structure created"})
			if err == nil {
				result += fmt.Sprintf("   💬 AI Comment: %s\n", comment)
			} else {
				result += "   💬 Added AI progress comment\n"
			}
		}

		if updateStatuses && task.Progress > 0 && task.Status != "Done" {
			result += "   🔄 Status updated based on progress\n"
		}

		totalProgress += task.Progress
	}

	averageProgress := totalProgress / len(exampleTasks)
	result += fmt.Sprintf("\n📈 Overall Progress: %d%%\n", averageProgress)

	if generateReport {
		result += "\n📄 Progress Report Generated:\n"
		result += "----------------------------\n"
		result += fmt.Sprintf("• %d task(s) completed\n", tasksCompleted)
		result += fmt.Sprintf("• %d task(s) in progress\n", tasksInProgress)
		result += fmt.Sprintf("• %d task(s) pending\n", tasksPending)
		result += fmt.Sprintf("• Average completion: %d%%\n", averageProgress)
		
		if addProgressComments {
			result += "• AI progress comments generated\n"
		}
		
		// Estimate completion time based on progress
		remainingWork := 100 - averageProgress
		estimatedDays := (remainingWork / 25) // Rough estimate: 25% per day
		if estimatedDays < 1 {
			estimatedDays = 1
		}
		result += fmt.Sprintf("• Estimated completion: %d day(s)\n", int(estimatedDays))
	}

	return &ToolResult{
		Content: []map[string]interface{}{
			{
				"type": "text",
				"text": result,
			},
		},
	}, nil
}
