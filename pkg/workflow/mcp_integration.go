package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/ai"
)

// MCPIntegration интеграция с Model Context Protocol
type MCPIntegration struct {
	tools           map[string]MCPTool
	workflows       *WorkflowEngine
	aiChains        *ai.AIChains
	logger          Logger
	eventBus        *EventBus
	config          *MCPConfig
	resourceManager *MCPResourceManager
	mutex           sync.RWMutex
}

// MCPConfig конфигурация MCP интеграции
type MCPConfig struct {
	ToolsDirectory    string                 `json:"tools_directory"`
	MaxConcurrentOps  int                    `json:"max_concurrent_ops"`
	Timeout           time.Duration          `json:"timeout"`
	RetryAttempts     int                    `json:"retry_attempts"`
	EnableAutoTools   bool                   `json:"enable_auto_tools"`
	SecurityPolicy    *MCPSecurityPolicy     `json:"security_policy"`
	ProviderSettings  map[string]interface{} `json:"provider_settings"`
}

// MCPSecurityPolicy политика безопасности для MCP
type MCPSecurityPolicy struct {
	AllowedTools      []string `json:"allowed_tools"`
	BlockedTools      []string `json:"blocked_tools"`
	RequireApproval   bool     `json:"require_approval"`
	SandboxMode       bool     `json:"sandbox_mode"`
	MaxResourceUsage  int64    `json:"max_resource_usage"`
	TimeoutPerTool    time.Duration `json:"timeout_per_tool"`
}

// MCPTool интерфейс для MCP инструментов
type MCPTool interface {
	GetName() string
	GetDescription() string
	GetSchema() *MCPToolSchema
	Execute(ctx context.Context, input *MCPToolInput) (*MCPToolOutput, error)
	ValidateInput(input *MCPToolInput) error
	GetCapabilities() []string
}

// MCPToolSchema схема MCP инструмента
type MCPToolSchema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
	OutputSchema map[string]interface{} `json:"output_schema"`
	Parameters  []MCPParameter         `json:"parameters"`
	Capabilities []string              `json:"capabilities"`
	Version     string                 `json:"version"`
}

// MCPParameter параметр MCP инструмента
type MCPParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Pattern     string      `json:"pattern,omitempty"`
}

// MCPToolInput входные данные для MCP инструмента
type MCPToolInput struct {
	ToolName   string                 `json:"tool_name"`
	Parameters map[string]interface{} `json:"parameters"`
	Context    *MCPExecutionContext   `json:"context"`
	RequestID  string                 `json:"request_id"`
	Timestamp  time.Time              `json:"timestamp"`
}

// MCPToolOutput выходные данные MCP инструмента
type MCPToolOutput struct {
	Success   bool                   `json:"success"`
	Result    interface{}            `json:"result"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
	Resources []MCPResource          `json:"resources,omitempty"`
	Duration  time.Duration          `json:"duration"`
	RequestID string                 `json:"request_id"`
}

// MCPExecutionContext контекст выполнения MCP
type MCPExecutionContext struct {
	WorkflowID     string                 `json:"workflow_id"`
	TaskID         string                 `json:"task_id"`
	UserID         string                 `json:"user_id"`
	Environment    string                 `json:"environment"`
	Permissions    []string               `json:"permissions"`
	Resources      map[string]interface{} `json:"resources"`
	AIChainContext map[string]interface{} `json:"ai_chain_context"`
}

// MCPResource ресурс MCP
type MCPResource struct {
	URI         string                 `json:"uri"`
	Type        string                 `json:"type"`
	Content     interface{}            `json:"content"`
	Metadata    map[string]interface{} `json:"metadata"`
	AccessLevel string                 `json:"access_level"`
	TTL         time.Duration          `json:"ttl,omitempty"`
}

// MCPResourceManager менеджер ресурсов MCP
type MCPResourceManager struct {
	resources map[string]*MCPResource
	access    map[string][]string // resource -> allowed users
	mutex     sync.RWMutex
	logger    Logger
}

// NewMCPIntegration создает новую MCP интеграцию
func NewMCPIntegration(workflows *WorkflowEngine, aiChains *ai.AIChains, eventBus *EventBus, config *MCPConfig, logger Logger) *MCPIntegration {
	if config == nil {
		config = &MCPConfig{
			MaxConcurrentOps: 10,
			Timeout:          30 * time.Second,
			RetryAttempts:    3,
			EnableAutoTools:  true,
			SecurityPolicy: &MCPSecurityPolicy{
				RequireApproval:  false,
				SandboxMode:      true,
				MaxResourceUsage: 1024 * 1024 * 100, // 100MB
				TimeoutPerTool:   10 * time.Second,
			},
		}
	}

	mcp := &MCPIntegration{
		tools:           make(map[string]MCPTool),
		workflows:       workflows,
		aiChains:        aiChains,
		logger:          logger,
		eventBus:        eventBus,
		config:          config,
		resourceManager: NewMCPResourceManager(logger),
	}

	// Регистрируем стандартные MCP инструменты
	mcp.registerBuiltinTools()

	// Подписываемся на события workflow
	mcp.subscribeToWorkflowEvents()

	return mcp
}

// NewMCPResourceManager создает новый менеджер ресурсов
func NewMCPResourceManager(logger Logger) *MCPResourceManager {
	return &MCPResourceManager{
		resources: make(map[string]*MCPResource),
		access:    make(map[string][]string),
		logger:    logger,
	}
}

// registerBuiltinTools регистрирует встроенные MCP инструменты
func (mcp *MCPIntegration) registerBuiltinTools() {
	// AI Analysis Tool
	mcp.RegisterTool(NewAIAnalysisTool(mcp.aiChains, mcp.logger))
	
	// Workflow Control Tool
	mcp.RegisterTool(NewWorkflowControlTool(mcp.workflows, mcp.logger))
	
	// Resource Management Tool
	mcp.RegisterTool(NewResourceManagementTool(mcp.resourceManager, mcp.logger))
	
	// Code Analysis Tool
	mcp.RegisterTool(NewCodeAnalysisTool(mcp.aiChains, mcp.logger))
	
	// Notification Tool  
	mcp.RegisterTool(NewNotificationTool(nil, mcp.logger))

	mcp.logger.Info("Built-in MCP tools registered", "count", len(mcp.tools))
}

// subscribeToWorkflowEvents подписывается на события workflow
func (mcp *MCPIntegration) subscribeToWorkflowEvents() {
	mcp.eventBus.Subscribe("workflow.task.created", &MCPEventHandler{handler: mcp.handleTaskCreated})
	mcp.eventBus.Subscribe("workflow.task.completed", &MCPEventHandler{handler: mcp.handleTaskCompleted})
	mcp.eventBus.Subscribe("workflow.stage.changed", &MCPEventHandler{handler: mcp.handleStageChanged})
	mcp.eventBus.Subscribe("mcp.tool.requested", &MCPEventHandler{handler: mcp.handleToolRequested})
}

// RegisterTool регистрирует MCP инструмент
func (mcp *MCPIntegration) RegisterTool(tool MCPTool) error {
	mcp.mutex.Lock()
	defer mcp.mutex.Unlock()

	// Проверяем политику безопасности
	if !mcp.isToolAllowed(tool.GetName()) {
		return fmt.Errorf("tool %s is not allowed by security policy", tool.GetName())
	}

	mcp.tools[tool.GetName()] = tool
	mcp.logger.Info("MCP tool registered", "name", tool.GetName(), "description", tool.GetDescription())

	// Публикуем событие
	event := &WorkflowEvent{
		Type:       "mcp.tool.registered",
		Timestamp:  time.Now(),
		Source:     "mcp_integration",
		Data: map[string]interface{}{
			"tool_name":    tool.GetName(),
			"capabilities": tool.GetCapabilities(),
		},
	}
	mcp.eventBus.Publish(context.Background(), event)

	return nil
}

// ExecuteTool выполняет MCP инструмент
func (mcp *MCPIntegration) ExecuteTool(ctx context.Context, input *MCPToolInput) (*MCPToolOutput, error) {
	mcp.mutex.RLock()
	tool, exists := mcp.tools[input.ToolName]
	mcp.mutex.RUnlock()

	if !exists {
		return &MCPToolOutput{
			Success:   false,
			Error:     fmt.Sprintf("tool %s not found", input.ToolName),
			RequestID: input.RequestID,
		}, fmt.Errorf("tool %s not found", input.ToolName)
	}

	// Валидация входных данных
	if err := tool.ValidateInput(input); err != nil {
		return &MCPToolOutput{
			Success:   false,
			Error:     fmt.Sprintf("input validation failed: %v", err),
			RequestID: input.RequestID,
		}, err
	}

	// Проверка разрешений
	if !mcp.hasPermission(input.Context, tool.GetName()) {
		return &MCPToolOutput{
			Success:   false,
			Error:     "insufficient permissions",
			RequestID: input.RequestID,
		}, fmt.Errorf("insufficient permissions for tool %s", tool.GetName())
	}

	// Создаем контекст с таймаутом
	toolCtx, cancel := context.WithTimeout(ctx, mcp.config.SecurityPolicy.TimeoutPerTool)
	defer cancel()

	// Выполняем инструмент
	startTime := time.Now()
	output, err := tool.Execute(toolCtx, input)
	duration := time.Since(startTime)

	if output == nil {
		output = &MCPToolOutput{
			Success:   false,
			Error:     "tool returned nil output",
			RequestID: input.RequestID,
			Duration:  duration,
		}
	} else {
		output.Duration = duration
		output.RequestID = input.RequestID
	}

	// Логируем выполнение
	mcp.logger.Info("MCP tool executed", 
		"tool", input.ToolName,
		"success", output.Success,
		"duration", duration,
		"request_id", input.RequestID)

	// Публикуем событие выполнения
	event := &WorkflowEvent{
		Type:      "mcp.tool.executed",
		Timestamp: time.Now(),
		Source:    "mcp_integration",
		Data: map[string]interface{}{
			"tool_name":  input.ToolName,
			"success":    output.Success,
			"duration":   duration,
			"request_id": input.RequestID,
		},
	}
	mcp.eventBus.Publish(context.Background(), event)

	return output, err
}

// GetAvailableTools возвращает список доступных инструментов
func (mcp *MCPIntegration) GetAvailableTools() []MCPToolSchema {
	mcp.mutex.RLock()
	defer mcp.mutex.RUnlock()

	var schemas []MCPToolSchema
	for _, tool := range mcp.tools {
		schemas = append(schemas, *tool.GetSchema())
	}

	return schemas
}

// AIToolIntegration интеграция MCP с AI цепочками
func (mcp *MCPIntegration) AIToolIntegration(ctx context.Context, prompt string, toolSuggestions []string) (*MCPToolOutput, error) {
	// Генерируем план использования инструментов с помощью AI
	planPrompt := fmt.Sprintf(`Given this user request: "%s"

Available MCP tools: %v

Generate an execution plan that specifies:
1. Which tools to use in what order
2. How to pass data between tools
3. Expected output format

Respond in JSON format with the execution plan.`, prompt, toolSuggestions)

	aiResponse, err := mcp.aiChains.ExecuteTask("MCP Tool Planning", planPrompt, "planning")
	if err != nil {
		return nil, fmt.Errorf("AI planning failed: %w", err)
	}

	// Парсим план выполнения
	var executionPlan MCPExecutionPlan
	if err := json.Unmarshal([]byte(aiResponse), &executionPlan); err != nil {
		// Если не удалось распарсить JSON, используем простую стратегию
		return mcp.executeSimpleToolChain(ctx, toolSuggestions, prompt)
	}

	// Выполняем план
	return mcp.executeToolPlan(ctx, &executionPlan)
}

// MCPExecutionPlan план выполнения MCP инструментов
type MCPExecutionPlan struct {
	Steps []MCPExecutionStep `json:"steps"`
	Goal  string             `json:"goal"`
}

// MCPExecutionStep шаг выполнения
type MCPExecutionStep struct {
	ToolName   string                 `json:"tool_name"`
	Parameters map[string]interface{} `json:"parameters"`
	DependsOn  []string               `json:"depends_on"`
	Output     string                 `json:"output_var"`
}

// executeToolPlan выполняет план инструментов
func (mcp *MCPIntegration) executeToolPlan(ctx context.Context, plan *MCPExecutionPlan) (*MCPToolOutput, error) {
	results := make(map[string]*MCPToolOutput)
	
	for _, step := range plan.Steps {
		// Проверяем зависимости
		for _, dep := range step.DependsOn {
			if _, exists := results[dep]; !exists {
				return nil, fmt.Errorf("dependency %s not satisfied for step %s", dep, step.ToolName)
			}
		}

		// Подготавливаем входные данные с результатами предыдущих шагов
		input := &MCPToolInput{
			ToolName:   step.ToolName,
			Parameters: step.Parameters,
			Context: &MCPExecutionContext{
				Environment: "ai_integration",
				Permissions: []string{"read", "execute"},
			},
			RequestID: fmt.Sprintf("ai-plan-%d", time.Now().UnixNano()),
			Timestamp: time.Now(),
		}

		// Обогащаем параметры результатами предыдущих шагов
		for depName, depResult := range results {
			input.Parameters[fmt.Sprintf("prev_%s", depName)] = depResult.Result
		}

		// Выполняем шаг
		output, err := mcp.ExecuteTool(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to execute step %s: %w", step.ToolName, err)
		}

		results[step.Output] = output

		if !output.Success {
			return output, fmt.Errorf("step %s failed: %s", step.ToolName, output.Error)
		}
	}

	// Возвращаем результат последнего шага
	if len(plan.Steps) > 0 {
		lastStep := plan.Steps[len(plan.Steps)-1]
		return results[lastStep.Output], nil
	}

	return &MCPToolOutput{
		Success: true,
		Result:  "Plan executed successfully",
	}, nil
}

// executeSimpleToolChain выполняет простую цепочку инструментов
func (mcp *MCPIntegration) executeSimpleToolChain(ctx context.Context, tools []string, prompt string) (*MCPToolOutput, error) {
	if len(tools) == 0 {
		return &MCPToolOutput{
			Success: false,
			Error:   "no tools specified",
		}, nil
	}

	// Выполняем первый инструмент с исходным промптом
	input := &MCPToolInput{
		ToolName: tools[0],
		Parameters: map[string]interface{}{
			"prompt": prompt,
			"task":   "analyze",
		},
		Context: &MCPExecutionContext{
			Environment: "ai_simple_chain",
			Permissions: []string{"read", "execute"},
		},
		RequestID: fmt.Sprintf("simple-chain-%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
	}

	return mcp.ExecuteTool(ctx, input)
}

// Event handlers

func (mcp *MCPIntegration) handleTaskCreated(event Event) {
	// Автоматически анализируем новые задачи с помощью MCP инструментов
	if !mcp.config.EnableAutoTools {
		return
	}

	taskData := event.GetData()
	if taskTitle, ok := taskData["title"].(string); ok {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), mcp.config.Timeout)
			defer cancel()

			// Используем AI анализ для новой задачи
			input := &MCPToolInput{
				ToolName: "ai_analysis",
				Parameters: map[string]interface{}{
					"text":         taskTitle,
					"analysis_type": "task_classification",
				},
				Context: &MCPExecutionContext{
					WorkflowID:  taskData["workflow_id"].(string),
					TaskID:      taskData["task_id"].(string),
					Environment: "auto_analysis",
				},
				RequestID: fmt.Sprintf("auto-task-%s", taskData["task_id"]),
				Timestamp: time.Now(),
			}

			_, err := mcp.ExecuteTool(ctx, input)
			if err != nil {
				mcp.logger.Error("Auto task analysis failed", err)
			}
		}()
	}
}

func (mcp *MCPIntegration) handleTaskCompleted(event Event) {
	// Анализируем завершенные задачи для извлечения инсайтов
	if !mcp.config.EnableAutoTools {
		return
	}

	// Аналогично handleTaskCreated, но для завершенных задач
}

func (mcp *MCPIntegration) handleStageChanged(event Event) {
	// Реагируем на изменения стадий workflow
}

func (mcp *MCPIntegration) handleToolRequested(event Event) {
	// Обрабатываем запросы на выполнение инструментов
	toolData := event.GetData()
	
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), mcp.config.Timeout)
		defer cancel()

		input := &MCPToolInput{
			ToolName:   toolData["tool_name"].(string),
			Parameters: toolData["parameters"].(map[string]interface{}),
			RequestID:  toolData["request_id"].(string),
			Timestamp:  time.Now(),
		}

		_, err := mcp.ExecuteTool(ctx, input)
		if err != nil {
			mcp.logger.Error("Requested tool execution failed", err)
		}
	}()
}

// Utility methods

func (mcp *MCPIntegration) isToolAllowed(toolName string) bool {
	policy := mcp.config.SecurityPolicy
	
	// Проверяем блокированные инструменты
	for _, blocked := range policy.BlockedTools {
		if blocked == toolName {
			return false
		}
	}

	// Если есть список разрешенных и инструмент не в нем
	if len(policy.AllowedTools) > 0 {
		for _, allowed := range policy.AllowedTools {
			if allowed == toolName {
				return true
			}
		}
		return false
	}

	return true
}

func (mcp *MCPIntegration) hasPermission(ctx *MCPExecutionContext, toolName string) bool {
	// Упрощенная проверка разрешений
	if ctx == nil {
		return false
	}

	// Все пользователи могут использовать базовые инструменты
	basicTools := []string{"ai_analysis", "resource_management"}
	for _, basic := range basicTools {
		if toolName == basic {
			return true
		}
	}

	// Для других инструментов нужны специальные разрешения
	return mcpContains(ctx.Permissions, "admin") || mcpContains(ctx.Permissions, "tool_execution")
}

// Resource Manager methods

func (rm *MCPResourceManager) AddResource(uri string, resource *MCPResource) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	rm.resources[uri] = resource
	rm.logger.Debug("MCP resource added", "uri", uri, "type", resource.Type)
}

func (rm *MCPResourceManager) GetResource(uri string, userID string) (*MCPResource, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	resource, exists := rm.resources[uri]
	if !exists {
		return nil, fmt.Errorf("resource %s not found", uri)
	}

	// Проверяем права доступа
	if allowedUsers, exists := rm.access[uri]; exists {
		if !mcpContains(allowedUsers, userID) {
			return nil, fmt.Errorf("access denied to resource %s", uri)
		}
	}

	return resource, nil
}

func (rm *MCPResourceManager) GrantAccess(uri string, userID string) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if _, exists := rm.access[uri]; !exists {
		rm.access[uri] = []string{}
	}

	if !mcpContains(rm.access[uri], userID) {
		rm.access[uri] = append(rm.access[uri], userID)
	}
}

func (rm *MCPResourceManager) ListResources(userID string) []string {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	var accessible []string
	for uri, allowedUsers := range rm.access {
		if mcpContains(allowedUsers, userID) {
			accessible = append(accessible, uri)
		}
	}

	return accessible
}

// MCPEventHandler адаптер для событий MCP
type MCPEventHandler struct {
	handler func(event Event)
}

func (h *MCPEventHandler) CanHandle(eventType string) bool {
	return true
}

func (h *MCPEventHandler) Handle(ctx context.Context, event Event) error {
	h.handler(event)
	return nil
}

// mcpContains проверяет содержание элемента в слайсе
func mcpContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}