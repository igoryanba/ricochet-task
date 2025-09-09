package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/ai"
)

// CompleteWorkflowEngine полная реализация движка workflow с всеми компонентами
type CompleteWorkflowEngine struct {
	// Основные компоненты
	eventBus         *EventBus
	ruleEngine       *RuleEngine
	progressTracker  *ProgressTracker
	autoAssignment   *AutoAssignment
	notifications    *SmartNotificationEngine
	mcpIntegration   *MCPIntegration

	// Системы
	workflows        map[string]*WorkflowDefinition
	runningWorkflows map[string]*WorkflowInstance
	logger           Logger
	aiChains         *ai.AIChains
	config           *CompleteEngineConfig

	mutex sync.RWMutex
}

// CompleteEngineConfig конфигурация полного движка
type CompleteEngineConfig struct {
	EventBusConfig       *EventBusConfig        `json:"event_bus"`
	RuleEngineConfig     *RuleEngineConfig      `json:"rule_engine"`
	ProgressConfig       *ProgressTrackingConfig `json:"progress_tracking"`
	AutoAssignmentConfig *AutoAssignmentConfig  `json:"auto_assignment"`
	NotificationConfig   *MCPConfig             `json:"notifications"`
	MCPConfig           *MCPConfig             `json:"mcp_integration"`
	MaxConcurrentWorkflows int                  `json:"max_concurrent_workflows"`
	DefaultTimeout      time.Duration          `json:"default_timeout"`
	EnableMetrics       bool                   `json:"enable_metrics"`
	EnableAuditLog      bool                   `json:"enable_audit_log"`
}

// WorkflowMetrics метрики workflow
type WorkflowMetrics struct {
	TotalWorkflows    int64             `json:"total_workflows"`
	ActiveWorkflows   int64             `json:"active_workflows"`
	CompletedWorkflows int64            `json:"completed_workflows"`
	FailedWorkflows   int64             `json:"failed_workflows"`
	AverageExecutionTime time.Duration   `json:"average_execution_time"`
	WorkflowsByStage  map[string]int64  `json:"workflows_by_stage"`
	TaskMetrics       *TaskMetrics      `json:"task_metrics"`
	NotificationMetrics *NotificationMetrics `json:"notification_metrics"`
	MCPToolUsage      map[string]int64  `json:"mcp_tool_usage"`
}

// TaskMetrics метрики задач
type TaskMetrics struct {
	TotalTasks        int64             `json:"total_tasks"`
	CompletedTasks    int64             `json:"completed_tasks"`
	OverdueTasks      int64             `json:"overdue_tasks"`
	AverageTaskTime   time.Duration     `json:"average_task_time"`
	TasksByPriority   map[string]int64  `json:"tasks_by_priority"`
	AutoAssignedTasks int64             `json:"auto_assigned_tasks"`
}

// NewCompleteWorkflowEngine создает полный движок workflow
func NewCompleteWorkflowEngine(aiChains *ai.AIChains, config *CompleteEngineConfig, logger Logger) (*CompleteWorkflowEngine, error) {
	if config == nil {
		config = GetDefaultCompleteConfig()
	}

	engine := &CompleteWorkflowEngine{
		workflows:        make(map[string]*WorkflowDefinition),
		runningWorkflows: make(map[string]*WorkflowInstance),
		logger:           logger,
		aiChains:         aiChains,
		config:           config,
	}

	// Инициализируем компоненты
	if err := engine.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}

	// Запускаем системы
	engine.startBackgroundSystems()

	logger.Info("Complete Workflow Engine initialized", 
		"max_concurrent", config.MaxConcurrentWorkflows,
		"metrics_enabled", config.EnableMetrics)

	return engine, nil
}

// initializeComponents инициализирует все компоненты
func (cwe *CompleteWorkflowEngine) initializeComponents() error {
	// Event Bus
	cwe.eventBus = NewEventBus(cwe.logger)

	// Rule Engine
	cwe.ruleEngine = NewRuleEngine(cwe.logger)

	// Progress Tracker
	var err error
	cwe.progressTracker, err = NewProgressTracker(cwe.config.ProgressConfig, cwe.logger)
	if err != nil {
		return fmt.Errorf("failed to create progress tracker: %w", err)
	}

	// Auto Assignment
	cwe.autoAssignment = NewAutoAssignment(cwe.aiChains, cwe.eventBus, cwe.config.AutoAssignmentConfig, cwe.logger)

	// Smart Notifications
	cwe.notifications = NewSmartNotificationEngine(cwe.aiChains, cwe.logger)

	// MCP Integration
	cwe.mcpIntegration = NewMCPIntegration(nil, cwe.aiChains, cwe.eventBus, cwe.config.MCPConfig, cwe.logger)

	// Регистрируем event handlers
	cwe.registerEventHandlers()

	return nil
}

// registerEventHandlers регистрирует обработчики событий
func (cwe *CompleteWorkflowEngine) registerEventHandlers() {
	// Progress tracking events
	cwe.eventBus.Subscribe("workflow.task.completed", &WorkflowEventHandler{
		handler: func(ctx context.Context, event Event) error {
			return cwe.progressTracker.HandleEvent(ctx, event)
		},
	})

	// Auto assignment events
	cwe.eventBus.Subscribe("workflow.task.created", &WorkflowEventHandler{
		handler: func(ctx context.Context, event Event) error {
			return cwe.autoAssignment.ProcessEvent(ctx, event)
		},
	})

	// Notification events
	cwe.eventBus.Subscribe("workflow.stage.changed", &WorkflowEventHandler{
		handler: func(ctx context.Context, event Event) error {
			return cwe.notifications.ProcessEvent(ctx, event)
		},
	})

	// Метрики и аудит
	if cwe.config.EnableMetrics || cwe.config.EnableAuditLog {
		cwe.eventBus.Subscribe("*", &WorkflowEventHandler{
			handler: cwe.handleMetricsAndAudit,
		})
	}
}

// WorkflowEventHandler адаптер для обработчиков событий
type WorkflowEventHandler struct {
	handler func(ctx context.Context, event Event) error
}

func (h *WorkflowEventHandler) CanHandle(eventType string) bool {
	return true
}

func (h *WorkflowEventHandler) Handle(ctx context.Context, event Event) error {
	return h.handler(ctx, event)
}

// startBackgroundSystems запускает фоновые системы
func (cwe *CompleteWorkflowEngine) startBackgroundSystems() {
	// Запускаем периодические задачи
	go cwe.runPeriodicTasks()
	
	// Запускаем мониторинг workflow
	go cwe.runWorkflowMonitoring()
	
	// Запускаем сборщик метрик
	if cwe.config.EnableMetrics {
		go cwe.runMetricsCollection()
	}
}

// CreateWorkflow создает новый workflow
func (cwe *CompleteWorkflowEngine) CreateWorkflow(ctx context.Context, definition *WorkflowDefinition) (*WorkflowInstance, error) {
	cwe.mutex.Lock()
	defer cwe.mutex.Unlock()

	// Проверяем лимиты
	if len(cwe.runningWorkflows) >= cwe.config.MaxConcurrentWorkflows {
		return nil, fmt.Errorf("maximum concurrent workflows limit reached: %d", cwe.config.MaxConcurrentWorkflows)
	}

	// Получаем первую стадию
	var firstStage string
	for stageName := range definition.Stages {
		firstStage = stageName
		break
	}

	// Создаем экземпляр workflow
	instance := &WorkflowInstance{
		ID:           fmt.Sprintf("wf-%d", time.Now().UnixNano()),
		Definition:   definition,
		Status:       "created",
		CreatedAt:    time.Now(),
		CurrentStage: firstStage,
		Context:      make(map[string]interface{}),
		Tasks:        make(map[string]*TaskInstance),
	}

	// Сохраняем workflow
	cwe.workflows[definition.Name] = definition
	cwe.runningWorkflows[instance.ID] = instance

	// Публикуем событие создания
	event := &WorkflowEvent{
		Type:       "workflow.created",
		Timestamp:  time.Now(),
		Source:     "complete_engine",
		WorkflowID: instance.ID,
		Data: map[string]interface{}{
			"definition_name": definition.Name,
			"stages_count":    len(definition.Stages),
		},
	}

	cwe.eventBus.Publish(ctx, event)

	cwe.logger.Info("Workflow created", 
		"workflow_id", instance.ID,
		"definition", definition.Name)

	return instance, nil
}

// ExecuteWorkflow выполняет workflow
func (cwe *CompleteWorkflowEngine) ExecuteWorkflow(ctx context.Context, workflowID string) error {
	cwe.mutex.RLock()
	instance, exists := cwe.runningWorkflows[workflowID]
	cwe.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("workflow %s not found", workflowID)
	}

	// Запускаем выполнение в отдельной горутине
	go func() {
		if err := cwe.executeWorkflowInternal(ctx, instance); err != nil {
			cwe.logger.Error("Workflow execution failed", err, "workflow_id", workflowID)
			
			// Публикуем событие ошибки
			errorEvent := &WorkflowEvent{
				Type:       "workflow.failed",
				Timestamp:  time.Now(),
				Source:     "complete_engine",
				WorkflowID: workflowID,
				Data: map[string]interface{}{
					"error": err.Error(),
				},
			}
			cwe.eventBus.Publish(context.Background(), errorEvent)
		}
	}()

	return nil
}

// executeWorkflowInternal внутренняя логика выполнения workflow
func (cwe *CompleteWorkflowEngine) executeWorkflowInternal(ctx context.Context, instance *WorkflowInstance) error {
	instance.Status = "running"
	instance.StartedAt = &[]time.Time{time.Now()}[0]

	defer func() {
		if instance.Status == "running" {
			instance.Status = "completed"
			completedAt := time.Now()
			instance.CompletedAt = &completedAt
		}
		
		// Публикуем событие завершения
		event := &WorkflowEvent{
			Type:       "workflow.completed",
			Timestamp:  time.Now(),
			Source:     "complete_engine",
			WorkflowID: instance.ID,
			Data: map[string]interface{}{
				"status":   instance.Status,
				"duration": time.Since(*instance.StartedAt),
			},
		}
		cwe.eventBus.Publish(context.Background(), event)
	}()

	// Выполняем стадии
	for stageName, stage := range instance.Definition.Stages {
		if err := cwe.executeStage(ctx, instance, stageName, stage); err != nil {
			instance.Status = "failed"
			return fmt.Errorf("stage %s failed: %w", stageName, err)
		}

		// Переходим к следующей стадии
		previousStage := instance.CurrentStage
		instance.CurrentStage = stageName
		
		// Публикуем событие смены стадии
		stageEvent := &WorkflowEvent{
			Type:       "workflow.stage.changed",
			Timestamp:  time.Now(),
			Source:     "complete_engine",
			WorkflowID: instance.ID,
			Data: map[string]interface{}{
				"stage_name":     stageName,
				"stage_actions":  len(stage.Actions),
				"previous_stage": previousStage,
			},
		}
		cwe.eventBus.Publish(ctx, stageEvent)
	}

	return nil
}

// executeStage выполняет стадию workflow
func (cwe *CompleteWorkflowEngine) executeStage(ctx context.Context, instance *WorkflowInstance, stageName string, stage *StageDefinition) error {
	cwe.logger.Info("Executing workflow stage", 
		"workflow_id", instance.ID,
		"stage", stageName,
		"actions_count", len(stage.Actions))

	// Создаем задачи для стадии из actions
	for _, action := range stage.Actions {
		task := &TaskInstance{
			ID:         fmt.Sprintf("task-%d", time.Now().UnixNano()),
			Definition: nil, // Создадим простую задачу
			Status:     "created",
			CreatedAt:  time.Now(),
			Context:    make(map[string]interface{}),
		}

		instance.Tasks[task.ID] = task

		// Публикуем событие создания задачи
		taskEvent := &WorkflowEvent{
			Type:       "workflow.task.created",
			Timestamp:  time.Now(),
			Source:     "complete_engine",
			WorkflowID: instance.ID,
			Data: map[string]interface{}{
				"task_id":     task.ID,
				"action_type": action.Type,
				"stage_name":  stageName,
			},
		}
		cwe.eventBus.Publish(ctx, taskEvent)

		// Автоназначение с использованием AI для некоторых типов действий
		if action.Type == "manual" || action.Type == "approval" {
			assignEvent := &WorkflowEvent{
				Type:       "workflow.task.auto_assign_requested",
				Timestamp:  time.Now(),
				Source:     "complete_engine",
				WorkflowID: instance.ID,
				Data: map[string]interface{}{
					"task_id":     task.ID,
					"action_type": action.Type,
				},
			}
			cwe.eventBus.Publish(ctx, assignEvent)
		}
	}

	// Ждем завершения всех задач стадии
	return cwe.waitForStageCompletion(ctx, instance, stageName, stage)
}

// waitForStageCompletion ждет завершения всех задач стадии
func (cwe *CompleteWorkflowEngine) waitForStageCompletion(ctx context.Context, instance *WorkflowInstance, stageName string, stage *StageDefinition) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := cwe.config.DefaultTimeout
	if timeout == 0 {
		timeout = 30 * time.Minute
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("stage %s timed out", stageName)
		case <-ticker.C:
			// Простая логика - если есть задачи для этой стадии и они не завершены
			stageTaskCount := 0
			completedTaskCount := 0
			
			for _, task := range instance.Tasks {
				// Проверяем принадлежность задачи к этой стадии
				if taskStageName, ok := task.Context["stage_name"].(string); ok && taskStageName == stageName {
					stageTaskCount++
					if task.Status == "completed" || task.Status == "skipped" {
						completedTaskCount++
					}
				}
			}

			// Если нет задач для стадии, считаем стадию завершенной
			if stageTaskCount == 0 {
				return nil
			}

			// Если все задачи завершены
			if completedTaskCount >= stageTaskCount {
				return nil
			}

			// Продолжаем цикл
		}
	}
}

// GetWorkflowStatus возвращает статус workflow
func (cwe *CompleteWorkflowEngine) GetWorkflowStatus(workflowID string) (*WorkflowInstance, error) {
	cwe.mutex.RLock()
	defer cwe.mutex.RUnlock()

	instance, exists := cwe.runningWorkflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow %s not found", workflowID)
	}

	return instance, nil
}

// GetMetrics возвращает метрики движка
func (cwe *CompleteWorkflowEngine) GetMetrics() *WorkflowMetrics {
	cwe.mutex.RLock()
	defer cwe.mutex.RUnlock()

	metrics := &WorkflowMetrics{
		TotalWorkflows:    int64(len(cwe.workflows)),
		ActiveWorkflows:   int64(len(cwe.runningWorkflows)),
		WorkflowsByStage:  make(map[string]int64),
		TaskMetrics:       &TaskMetrics{},
		MCPToolUsage:      make(map[string]int64),
	}

	// Подсчитываем статистику по running workflows
	for _, instance := range cwe.runningWorkflows {
		metrics.WorkflowsByStage[instance.CurrentStage]++
		
		switch instance.Status {
		case "completed":
			metrics.CompletedWorkflows++
		case "failed":
			metrics.FailedWorkflows++
		}

		// Подсчитываем задачи
		for _, task := range instance.Tasks {
			metrics.TaskMetrics.TotalTasks++
			if task.Status == "completed" {
				metrics.TaskMetrics.CompletedTasks++
			}
		}
	}

	// Получаем метрики уведомлений
	if cwe.notifications != nil {
		metrics.NotificationMetrics = cwe.notifications.analytics.GetMetrics()
	}

	return metrics
}

// Utility methods

func (cwe *CompleteWorkflowEngine) runPeriodicTasks() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cwe.cleanupCompletedWorkflows()
			cwe.checkWorkflowTimeouts()
		}
	}
}

func (cwe *CompleteWorkflowEngine) runWorkflowMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cwe.monitorWorkflowHealth()
		}
	}
}

func (cwe *CompleteWorkflowEngine) runMetricsCollection() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics := cwe.GetMetrics()
			cwe.logger.Debug("Workflow metrics collected", 
				"active_workflows", metrics.ActiveWorkflows,
				"total_tasks", metrics.TaskMetrics.TotalTasks)
		}
	}
}

func (cwe *CompleteWorkflowEngine) cleanupCompletedWorkflows() {
	cwe.mutex.Lock()
	defer cwe.mutex.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour) // Удаляем workflow старше 24 часов

	for id, instance := range cwe.runningWorkflows {
		if (instance.Status == "completed" || instance.Status == "failed") &&
		   instance.CompletedAt != nil && instance.CompletedAt.Before(cutoff) {
			delete(cwe.runningWorkflows, id)
			cwe.logger.Debug("Cleaned up completed workflow", "workflow_id", id)
		}
	}
}

func (cwe *CompleteWorkflowEngine) checkWorkflowTimeouts() {
	// Проверяем workflow на превышение таймаутов
	// Реализация проверки таймаутов
}

func (cwe *CompleteWorkflowEngine) monitorWorkflowHealth() {
	// Мониторинг здоровья workflow
	// Проверка зависших задач, ресурсов и т.д.
}

func (cwe *CompleteWorkflowEngine) handleMetricsAndAudit(ctx context.Context, event Event) error {
	if cwe.config.EnableAuditLog {
		cwe.logger.Info("Audit log", 
			"event_type", event.GetType(),
			"source", event.GetSource(),
			"timestamp", event.GetTimestamp())
	}

	// Дополнительная обработка метрик
	return nil
}

// GetDefaultCompleteConfig возвращает конфигурацию по умолчанию
func GetDefaultCompleteConfig() *CompleteEngineConfig {
	return &CompleteEngineConfig{
		MaxConcurrentWorkflows: 100,
		DefaultTimeout:         30 * time.Minute,
		EnableMetrics:          true,
		EnableAuditLog:         true,
		EventBusConfig: &EventBusConfig{
			MaxQueueSize:    1000,
			WorkerPoolSize:  10,
			EnableMetrics:   true,
		},
		ProgressConfig: &ProgressTrackingConfig{
			EnableGitIntegration: true,
			EnableMetrics:        true,
			UpdateInterval:       time.Minute,
		},
		AutoAssignmentConfig: &AutoAssignmentConfig{
			EnableAI:           true,
			MaxRetries:         3,
			AssignmentTimeout:  5 * time.Minute,
		},
		MCPConfig: &MCPConfig{
			MaxConcurrentOps: 10,
			Timeout:          30 * time.Second,
			EnableAutoTools:  true,
			SecurityPolicy: &MCPSecurityPolicy{
				SandboxMode:      true,
				TimeoutPerTool:   10 * time.Second,
				MaxResourceUsage: 100 * 1024 * 1024, // 100MB
			},
		},
	}
}