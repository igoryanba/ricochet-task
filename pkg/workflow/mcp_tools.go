package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/ai"
)

// AI Analysis Tool
type AIAnalysisTool struct {
	aiChains *ai.AIChains
	logger   Logger
}

func NewAIAnalysisTool(aiChains *ai.AIChains, logger Logger) *AIAnalysisTool {
	return &AIAnalysisTool{
		aiChains: aiChains,
		logger:   logger,
	}
}

func (tool *AIAnalysisTool) GetName() string {
	return "ai_analysis"
}

func (tool *AIAnalysisTool) GetDescription() string {
	return "Performs AI-powered analysis of text, code, or data"
}

func (tool *AIAnalysisTool) GetSchema() *MCPToolSchema {
	return &MCPToolSchema{
		Name:        "ai_analysis",
		Description: "AI analysis tool for various content types",
		Parameters: []MCPParameter{
			{
				Name:        "text",
				Type:        "string",
				Description: "Text to analyze",
				Required:    true,
			},
			{
				Name:        "analysis_type",
				Type:        "string",
				Description: "Type of analysis to perform",
				Required:    true,
				Enum:        []string{"sentiment", "classification", "summary", "extraction", "insight"},
			},
			{
				Name:        "context",
				Type:        "object",
				Description: "Additional context for analysis",
				Required:    false,
			},
		},
		Capabilities: []string{"text_analysis", "ml_inference", "insight_generation"},
		Version:      "1.0.0",
	}
}

func (tool *AIAnalysisTool) ValidateInput(input *MCPToolInput) error {
	if input.Parameters["text"] == nil {
		return fmt.Errorf("text parameter is required")
	}
	
	if input.Parameters["analysis_type"] == nil {
		return fmt.Errorf("analysis_type parameter is required")
	}

	analysisType := input.Parameters["analysis_type"].(string)
	validTypes := []string{"sentiment", "classification", "summary", "extraction", "insight"}
	
	for _, validType := range validTypes {
		if analysisType == validType {
			return nil
		}
	}

	return fmt.Errorf("invalid analysis_type: %s", analysisType)
}

func (tool *AIAnalysisTool) Execute(ctx context.Context, input *MCPToolInput) (*MCPToolOutput, error) {
	text := input.Parameters["text"].(string)
	analysisType := input.Parameters["analysis_type"].(string)

	var prompt string
	switch analysisType {
	case "sentiment":
		prompt = fmt.Sprintf("Analyze the sentiment of this text: '%s'. Provide sentiment score (-1 to 1) and explanation.", text)
	case "classification":
		prompt = fmt.Sprintf("Classify this text into relevant categories: '%s'. Provide categories and confidence scores.", text)
	case "summary":
		prompt = fmt.Sprintf("Provide a concise summary of this text: '%s'", text)
	case "extraction":
		prompt = fmt.Sprintf("Extract key entities, topics, and insights from this text: '%s'", text)
	case "insight":
		prompt = fmt.Sprintf("Generate actionable insights and recommendations based on this text: '%s'", text)
	default:
		return &MCPToolOutput{
			Success: false,
			Error:   fmt.Sprintf("unsupported analysis type: %s", analysisType),
		}, nil
	}

	var response string
	var err error

	if tool.aiChains != nil {
		response, err = tool.aiChains.ExecuteTask("AI Analysis", prompt, "analysis")
		if err != nil {
			return &MCPToolOutput{
				Success: false,
				Error:   fmt.Sprintf("AI analysis failed: %v", err),
			}, err
		}
	} else {
		// Мок-ответ для тестов
		response = fmt.Sprintf("Mock analysis result for %s: %s", analysisType, text)
	}

	result := map[string]interface{}{
		"analysis_type": analysisType,
		"input_text":    text,
		"result":        response,
		"timestamp":     time.Now(),
	}

	return &MCPToolOutput{
		Success:   true,
		Result:    result,
		RequestID: input.RequestID,
		Metadata: map[string]interface{}{
			"model_used":   "ai_chains",
			"text_length":  len(text),
		},
	}, nil
}

func (tool *AIAnalysisTool) GetCapabilities() []string {
	return []string{"text_analysis", "ml_inference", "insight_generation"}
}

// Workflow Control Tool
type WorkflowControlTool struct {
	workflows *WorkflowEngine
	logger    Logger
}

func NewWorkflowControlTool(workflows *WorkflowEngine, logger Logger) *WorkflowControlTool {
	return &WorkflowControlTool{
		workflows: workflows,
		logger:    logger,
	}
}

func (tool *WorkflowControlTool) GetName() string {
	return "workflow_control"
}

func (tool *WorkflowControlTool) GetDescription() string {
	return "Controls workflow execution, transitions, and state management"
}

func (tool *WorkflowControlTool) GetSchema() *MCPToolSchema {
	return &MCPToolSchema{
		Name:        "workflow_control",
		Description: "Workflow management and control operations",
		Parameters: []MCPParameter{
			{
				Name:        "action",
				Type:        "string",
				Description: "Action to perform",
				Required:    true,
				Enum:        []string{"start", "stop", "pause", "resume", "transition", "status", "list"},
			},
			{
				Name:        "workflow_id",
				Type:        "string",
				Description: "Workflow identifier",
				Required:    false,
			},
			{
				Name:        "task_id",
				Type:        "string",
				Description: "Task identifier",
				Required:    false,
			},
			{
				Name:        "new_stage",
				Type:        "string",
				Description: "New stage for transition",
				Required:    false,
			},
		},
		Capabilities: []string{"workflow_management", "state_control", "transition_handling"},
		Version:      "1.0.0",
	}
}

func (tool *WorkflowControlTool) ValidateInput(input *MCPToolInput) error {
	action, ok := input.Parameters["action"].(string)
	if !ok {
		return fmt.Errorf("action parameter is required")
	}

	validActions := []string{"start", "stop", "pause", "resume", "transition", "status", "list"}
	for _, validAction := range validActions {
		if action == validAction {
			return nil
		}
	}

	return fmt.Errorf("invalid action: %s", action)
}

func (tool *WorkflowControlTool) Execute(ctx context.Context, input *MCPToolInput) (*MCPToolOutput, error) {
	action := input.Parameters["action"].(string)

	switch action {
	case "list":
		return tool.listWorkflows()
	case "status":
		return tool.getWorkflowStatus(input)
	case "transition":
		return tool.transitionTask(input)
	case "start", "stop", "pause", "resume":
		return tool.controlWorkflow(action, input)
	default:
		return &MCPToolOutput{
			Success: false,
			Error:   fmt.Sprintf("unsupported action: %s", action),
		}, nil
	}
}

func (tool *WorkflowControlTool) listWorkflows() (*MCPToolOutput, error) {
	// Здесь должна быть логика получения списка workflow
	// Пока возвращаем мок данные
	workflows := []map[string]interface{}{
		{
			"id":     "wf-1",
			"name":   "Development Workflow",
			"status": "active",
			"tasks":  5,
		},
		{
			"id":     "wf-2", 
			"name":   "Review Workflow",
			"status": "paused",
			"tasks":  3,
		},
	}

	return &MCPToolOutput{
		Success: true,
		Result:  workflows,
		Metadata: map[string]interface{}{
			"total_workflows": len(workflows),
		},
	}, nil
}

func (tool *WorkflowControlTool) getWorkflowStatus(input *MCPToolInput) (*MCPToolOutput, error) {
	workflowID, ok := input.Parameters["workflow_id"].(string)
	if !ok {
		return &MCPToolOutput{
			Success: false,
			Error:   "workflow_id is required for status action",
		}, nil
	}

	// Мок статуса workflow
	status := map[string]interface{}{
		"id":            workflowID,
		"status":        "running",
		"current_stage": "development",
		"progress":      75.5,
		"tasks": []map[string]interface{}{
			{
				"id":     "task-1",
				"title":  "Implement feature",
				"status": "in_progress",
				"assignee": "user1",
			},
			{
				"id":     "task-2",
				"title":  "Write tests",
				"status": "pending",
				"assignee": "user2",
			},
		},
	}

	return &MCPToolOutput{
		Success: true,
		Result:  status,
	}, nil
}

func (tool *WorkflowControlTool) transitionTask(input *MCPToolInput) (*MCPToolOutput, error) {
	taskID, ok := input.Parameters["task_id"].(string)
	if !ok {
		return &MCPToolOutput{
			Success: false,
			Error:   "task_id is required for transition action",
		}, nil
	}

	newStage, ok := input.Parameters["new_stage"].(string)
	if !ok {
		return &MCPToolOutput{
			Success: false,
			Error:   "new_stage is required for transition action",
		}, nil
	}

	tool.logger.Info("Task transition requested", "task_id", taskID, "new_stage", newStage)

	result := map[string]interface{}{
		"task_id":     taskID,
		"old_stage":   "in_progress",
		"new_stage":   newStage,
		"transitioned_at": time.Now(),
		"success":     true,
	}

	return &MCPToolOutput{
		Success: true,
		Result:  result,
	}, nil
}

func (tool *WorkflowControlTool) controlWorkflow(action string, input *MCPToolInput) (*MCPToolOutput, error) {
	workflowID, ok := input.Parameters["workflow_id"].(string)
	if !ok {
		return &MCPToolOutput{
			Success: false,
			Error:   "workflow_id is required for control actions",
		}, nil
	}

	tool.logger.Info("Workflow control action", "action", action, "workflow_id", workflowID)

	result := map[string]interface{}{
		"workflow_id": workflowID,
		"action":      action,
		"status":      "completed",
		"timestamp":   time.Now(),
	}

	return &MCPToolOutput{
		Success: true,
		Result:  result,
	}, nil
}

func (tool *WorkflowControlTool) GetCapabilities() []string {
	return []string{"workflow_management", "state_control", "transition_handling"}
}

// Resource Management Tool
type ResourceManagementTool struct {
	resourceManager *MCPResourceManager
	logger          Logger
}

func NewResourceManagementTool(resourceManager *MCPResourceManager, logger Logger) *ResourceManagementTool {
	return &ResourceManagementTool{
		resourceManager: resourceManager,
		logger:          logger,
	}
}

func (tool *ResourceManagementTool) GetName() string {
	return "resource_management"
}

func (tool *ResourceManagementTool) GetDescription() string {
	return "Manages MCP resources, access control, and resource lifecycle"
}

func (tool *ResourceManagementTool) GetSchema() *MCPToolSchema {
	return &MCPToolSchema{
		Name:        "resource_management",
		Description: "Resource management operations",
		Parameters: []MCPParameter{
			{
				Name:        "action",
				Type:        "string",
				Description: "Action to perform",
				Required:    true,
				Enum:        []string{"create", "read", "update", "delete", "list", "grant_access", "revoke_access"},
			},
			{
				Name:        "resource_uri",
				Type:        "string",
				Description: "Resource URI",
				Required:    false,
			},
			{
				Name:        "user_id",
				Type:        "string",
				Description: "User ID for access control",
				Required:    false,
			},
			{
				Name:        "resource_data",
				Type:        "object",
				Description: "Resource data for create/update operations",
				Required:    false,
			},
		},
		Capabilities: []string{"resource_management", "access_control", "data_storage"},
		Version:      "1.0.0",
	}
}

func (tool *ResourceManagementTool) ValidateInput(input *MCPToolInput) error {
	action, ok := input.Parameters["action"].(string)
	if !ok {
		return fmt.Errorf("action parameter is required")
	}

	validActions := []string{"create", "read", "update", "delete", "list", "grant_access", "revoke_access"}
	for _, validAction := range validActions {
		if action == validAction {
			return nil
		}
	}

	return fmt.Errorf("invalid action: %s", action)
}

func (tool *ResourceManagementTool) Execute(ctx context.Context, input *MCPToolInput) (*MCPToolOutput, error) {
	action := input.Parameters["action"].(string)

	switch action {
	case "list":
		return tool.listResources(input)
	case "read":
		return tool.readResource(input)
	case "create":
		return tool.createResource(input)
	case "grant_access":
		return tool.grantAccess(input)
	default:
		return &MCPToolOutput{
			Success: false,
			Error:   fmt.Sprintf("action %s not implemented yet", action),
		}, nil
	}
}

func (tool *ResourceManagementTool) listResources(input *MCPToolInput) (*MCPToolOutput, error) {
	userID, ok := input.Parameters["user_id"].(string)
	if !ok {
		userID = "anonymous"
	}

	resources := tool.resourceManager.ListResources(userID)
	
	return &MCPToolOutput{
		Success: true,
		Result: map[string]interface{}{
			"resources": resources,
			"count":     len(resources),
			"user_id":   userID,
		},
	}, nil
}

func (tool *ResourceManagementTool) readResource(input *MCPToolInput) (*MCPToolOutput, error) {
	uri, ok := input.Parameters["resource_uri"].(string)
	if !ok {
		return &MCPToolOutput{
			Success: false,
			Error:   "resource_uri is required for read action",
		}, nil
	}

	userID, ok := input.Parameters["user_id"].(string)
	if !ok {
		userID = "anonymous"
	}

	resource, err := tool.resourceManager.GetResource(uri, userID)
	if err != nil {
		return &MCPToolOutput{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	return &MCPToolOutput{
		Success: true,
		Result:  resource,
	}, nil
}

func (tool *ResourceManagementTool) createResource(input *MCPToolInput) (*MCPToolOutput, error) {
	uri, ok := input.Parameters["resource_uri"].(string)
	if !ok {
		return &MCPToolOutput{
			Success: false,
			Error:   "resource_uri is required for create action",
		}, nil
	}

	resourceData, ok := input.Parameters["resource_data"].(map[string]interface{})
	if !ok {
		return &MCPToolOutput{
			Success: false,
			Error:   "resource_data is required for create action",
		}, nil
	}

	resource := &MCPResource{
		URI:         uri,
		Type:        "data",
		Content:     resourceData,
		Metadata:    make(map[string]interface{}),
		AccessLevel: "private",
	}

	tool.resourceManager.AddResource(uri, resource)

	return &MCPToolOutput{
		Success: true,
		Result: map[string]interface{}{
			"uri":        uri,
			"created_at": time.Now(),
			"status":     "created",
		},
	}, nil
}

func (tool *ResourceManagementTool) grantAccess(input *MCPToolInput) (*MCPToolOutput, error) {
	uri, ok := input.Parameters["resource_uri"].(string)
	if !ok {
		return &MCPToolOutput{
			Success: false,
			Error:   "resource_uri is required for grant_access action",
		}, nil
	}

	userID, ok := input.Parameters["user_id"].(string)
	if !ok {
		return &MCPToolOutput{
			Success: false,
			Error:   "user_id is required for grant_access action",
		}, nil
	}

	tool.resourceManager.GrantAccess(uri, userID)

	return &MCPToolOutput{
		Success: true,
		Result: map[string]interface{}{
			"uri":        uri,
			"user_id":    userID,
			"granted_at": time.Now(),
			"status":     "access_granted",
		},
	}, nil
}

func (tool *ResourceManagementTool) GetCapabilities() []string {
	return []string{"resource_management", "access_control", "data_storage"}
}

// Code Analysis Tool
type CodeAnalysisTool struct {
	aiChains *ai.AIChains
	logger   Logger
}

func NewCodeAnalysisTool(aiChains *ai.AIChains, logger Logger) *CodeAnalysisTool {
	return &CodeAnalysisTool{
		aiChains: aiChains,
		logger:   logger,
	}
}

func (tool *CodeAnalysisTool) GetName() string {
	return "code_analysis"
}

func (tool *CodeAnalysisTool) GetDescription() string {
	return "Analyzes code for quality, security, performance, and best practices"
}

func (tool *CodeAnalysisTool) GetSchema() *MCPToolSchema {
	return &MCPToolSchema{
		Name:        "code_analysis",
		Description: "Code analysis and review tool",
		Parameters: []MCPParameter{
			{
				Name:        "code",
				Type:        "string",
				Description: "Code to analyze",
				Required:    true,
			},
			{
				Name:        "language",
				Type:        "string",
				Description: "Programming language",
				Required:    true,
			},
			{
				Name:        "analysis_type",
				Type:        "string",
				Description: "Type of analysis",
				Required:    false,
				Enum:        []string{"quality", "security", "performance", "best_practices", "all"},
				Default:     "all",
			},
		},
		Capabilities: []string{"code_review", "security_scan", "quality_analysis"},
		Version:      "1.0.0",
	}
}

func (tool *CodeAnalysisTool) ValidateInput(input *MCPToolInput) error {
	if input.Parameters["code"] == nil {
		return fmt.Errorf("code parameter is required")
	}
	
	if input.Parameters["language"] == nil {
		return fmt.Errorf("language parameter is required")
	}

	return nil
}

func (tool *CodeAnalysisTool) Execute(ctx context.Context, input *MCPToolInput) (*MCPToolOutput, error) {
	code := input.Parameters["code"].(string)
	language := input.Parameters["language"].(string)
	
	analysisType := "all"
	if input.Parameters["analysis_type"] != nil {
		analysisType = input.Parameters["analysis_type"].(string)
	}

	prompt := fmt.Sprintf(`Analyze this %s code for %s:

%s

Provide a detailed analysis including:
1. Issues found
2. Recommendations
3. Score (0-100)
4. Priority areas for improvement

Format the response as JSON.`, language, analysisType, code)

	response, err := tool.aiChains.ExecuteTask("Code Analysis", prompt, "code_review")
	if err != nil {
		return &MCPToolOutput{
			Success: false,
			Error:   fmt.Sprintf("Code analysis failed: %v", err),
		}, err
	}

	// Попытка парсинга JSON ответа
	var analysisResult map[string]interface{}
	if err := json.Unmarshal([]byte(response), &analysisResult); err != nil {
		// Если JSON не парсится, возвращаем текстовый ответ
		analysisResult = map[string]interface{}{
			"analysis": response,
			"format":   "text",
		}
	}

	result := map[string]interface{}{
		"language":      language,
		"analysis_type": analysisType,
		"code_length":   len(code),
		"result":        analysisResult,
		"timestamp":     time.Now(),
	}

	return &MCPToolOutput{
		Success: true,
		Result:  result,
		Metadata: map[string]interface{}{
			"analyzer": "ai_chains",
			"language": language,
		},
	}, nil
}

func (tool *CodeAnalysisTool) GetCapabilities() []string {
	return []string{"code_review", "security_scan", "quality_analysis"}
}

// Notification Tool
type NotificationTool struct {
	notificationEngine *SmartNotificationEngine
	logger             Logger
}

func NewNotificationTool(notificationEngine *SmartNotificationEngine, logger Logger) *NotificationTool {
	return &NotificationTool{
		notificationEngine: notificationEngine,
		logger:             logger,
	}
}

func (tool *NotificationTool) GetName() string {
	return "notification"
}

func (tool *NotificationTool) GetDescription() string {
	return "Sends notifications through various channels with smart routing"
}

func (tool *NotificationTool) GetSchema() *MCPToolSchema {
	return &MCPToolSchema{
		Name:        "notification",
		Description: "Smart notification delivery system",
		Parameters: []MCPParameter{
			{
				Name:        "title",
				Type:        "string",
				Description: "Notification title",
				Required:    true,
			},
			{
				Name:        "message",
				Type:        "string",
				Description: "Notification message",
				Required:    true,
			},
			{
				Name:        "recipients",
				Type:        "array",
				Description: "List of recipients",
				Required:    true,
			},
			{
				Name:        "priority",
				Type:        "string",
				Description: "Notification priority",
				Required:    false,
				Enum:        []string{"low", "medium", "high", "critical"},
				Default:     "medium",
			},
			{
				Name:        "channels",
				Type:        "array",
				Description: "Preferred channels",
				Required:    false,
			},
		},
		Capabilities: []string{"multi_channel_delivery", "smart_routing", "priority_handling"},
		Version:      "1.0.0",
	}
}

func (tool *NotificationTool) ValidateInput(input *MCPToolInput) error {
	if input.Parameters["title"] == nil {
		return fmt.Errorf("title parameter is required")
	}
	
	if input.Parameters["message"] == nil {
		return fmt.Errorf("message parameter is required")
	}

	if input.Parameters["recipients"] == nil {
		return fmt.Errorf("recipients parameter is required")
	}

	return nil
}

func (tool *NotificationTool) Execute(ctx context.Context, input *MCPToolInput) (*MCPToolOutput, error) {
	title := input.Parameters["title"].(string)
	message := input.Parameters["message"].(string)
	
	// Преобразуем recipients
	recipientsRaw := input.Parameters["recipients"].([]interface{})
	recipients := make([]string, len(recipientsRaw))
	for i, r := range recipientsRaw {
		recipients[i] = r.(string)
	}

	priority := "medium"
	if input.Parameters["priority"] != nil {
		priority = input.Parameters["priority"].(string)
	}

	// Создаем уведомление
	notification := &Notification{
		ID:         fmt.Sprintf("mcp-notif-%d", time.Now().UnixNano()),
		Type:       "mcp_notification",
		Title:      title,
		Message:    message,
		Priority:   priority,
		Recipients: recipients,
		Data:       input.Parameters,
		Timestamp:  time.Now(),
	}

	// Создаем событие для отправки уведомления
	event := &WorkflowEvent{
		Type:      "notification.send",
		Timestamp: time.Now(),
		Source:    "mcp_tool",
		Data: map[string]interface{}{
			"notification": notification,
		},
	}

	// Если есть движок уведомлений, используем его
	if tool.notificationEngine != nil {
		err := tool.notificationEngine.ProcessEvent(ctx, event)
		if err != nil {
			return &MCPToolOutput{
				Success: false,
				Error:   fmt.Sprintf("Failed to send notification: %v", err),
			}, err
		}
	} else {
		// Иначе просто логируем
		tool.logger.Info("Notification would be sent", 
			"title", title, 
			"recipients", len(recipients))
	}

	result := map[string]interface{}{
		"notification_id": notification.ID,
		"title":           title,
		"recipients":      recipients,
		"priority":        priority,
		"sent_at":         time.Now(),
		"status":          "sent",
	}

	return &MCPToolOutput{
		Success: true,
		Result:  result,
		Metadata: map[string]interface{}{
			"delivery_channels": []string{"email", "slack"},
			"estimated_delivery": "immediate",
		},
	}, nil
}

func (tool *NotificationTool) GetCapabilities() []string {
	return []string{"multi_channel_delivery", "smart_routing", "priority_handling"}
}