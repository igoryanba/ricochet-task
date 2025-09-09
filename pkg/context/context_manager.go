package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/providers"
)

// ContextManager управляет контекстом работы с проектами и досками
type ContextManager struct {
	mu           sync.RWMutex
	contexts     map[string]*WorkingContext
	activeID     string
	persistPath  string
	logger       Logger
}

// WorkingContext содержит контекст работы с конкретным проектом
type WorkingContext struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	IsActive     bool                   `json:"is_active"`
	
	// Board context
	BoardID      string                 `json:"board_id"`
	ProjectID    string                 `json:"project_id"`
	ProviderName string                 `json:"provider_name"`
	
	// Default settings
	DefaultAssignee string              `json:"default_assignee"`
	DefaultLabels   []string            `json:"default_labels"`
	DefaultPriority providers.TaskPriority `json:"default_priority"`
	
	// Project settings
	ProjectType     string              `json:"project_type"`     // feature, bugfix, research, maintenance
	Complexity      string              `json:"complexity"`       // simple, medium, complex
	Timeline        int                 `json:"timeline"`         // days
	TeamSize        int                 `json:"team_size"`
	
	// Workflow preferences
	WorkflowType    string              `json:"workflow_type"`    // agile, kanban, waterfall
	AutoAssignment  bool                `json:"auto_assignment"`
	AutoProgress    bool                `json:"auto_progress"`
	AIEnabled       bool                `json:"ai_enabled"`
	
	// Custom fields
	CustomFields    map[string]interface{} `json:"custom_fields"`
	
	// Multi-project support
	SubProjects     []string            `json:"sub_projects"`
	Dependencies    []string            `json:"dependencies"`
	
	// Statistics
	Stats           *ContextStats       `json:"stats"`
}

// ContextStats содержит статистику работы в контексте
type ContextStats struct {
	TasksCreated    int       `json:"tasks_created"`
	TasksCompleted  int       `json:"tasks_completed"`
	PlansGenerated  int       `json:"plans_generated"`
	LastActivity    time.Time `json:"last_activity"`
	TotalTimeSpent  int64     `json:"total_time_spent"` // seconds
	SuccessRate     float64   `json:"success_rate"`     // 0.0 - 1.0
}

// MultiProjectConfig конфигурация для мульти-проектной работы
type MultiProjectConfig struct {
	EnableCrossProject   bool     `json:"enable_cross_project"`
	SyncBetweenContexts bool     `json:"sync_between_contexts"`
	SharedContexts      []string `json:"shared_contexts"`
	AutoSwitching       bool     `json:"auto_switching"`
}

// Logger интерфейс для логирования
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, err error, keysAndValues ...interface{})
}

// NewContextManager создает новый менеджер контекстов
func NewContextManager(persistPath string, logger Logger) *ContextManager {
	if persistPath == "" {
		homeDir, _ := os.UserHomeDir()
		persistPath = filepath.Join(homeDir, ".ricochet", "contexts.json")
	}

	cm := &ContextManager{
		contexts:    make(map[string]*WorkingContext),
		persistPath: persistPath,
		logger:      logger,
	}

	// Загружаем сохраненные контексты
	cm.loadContexts()

	return cm
}

// CreateContext создает новый рабочий контекст
func (cm *ContextManager) CreateContext(name, description string, config *ContextConfig) (*WorkingContext, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	id := fmt.Sprintf("ctx_%d", time.Now().UnixNano())
	
	ctx := &WorkingContext{
		ID:              id,
		Name:            name,
		Description:     description,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		IsActive:        false,
		CustomFields:    make(map[string]interface{}),
		Stats:           &ContextStats{},
	}

	// Применяем конфигурацию если передана
	if config != nil {
		ctx.BoardID = config.BoardID
		ctx.ProjectID = config.ProjectID
		ctx.ProviderName = config.ProviderName
		ctx.DefaultAssignee = config.DefaultAssignee
		ctx.DefaultLabels = config.DefaultLabels
		ctx.DefaultPriority = config.DefaultPriority
		ctx.ProjectType = config.ProjectType
		ctx.Complexity = config.Complexity
		ctx.Timeline = config.Timeline
		ctx.TeamSize = config.TeamSize
		ctx.WorkflowType = config.WorkflowType
		ctx.AutoAssignment = config.AutoAssignment
		ctx.AutoProgress = config.AutoProgress
		ctx.AIEnabled = config.AIEnabled
		
		for k, v := range config.CustomFields {
			ctx.CustomFields[k] = v
		}
	}

	cm.contexts[id] = ctx
	
	if err := cm.saveContexts(); err != nil {
		delete(cm.contexts, id)
		return nil, fmt.Errorf("failed to save context: %w", err)
	}

	cm.logger.Info("Context created", "id", id, "name", name)
	return ctx, nil
}

// SetActiveContext устанавливает активный контекст
func (cm *ContextManager) SetActiveContext(contextID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ctx, exists := cm.contexts[contextID]
	if !exists {
		return fmt.Errorf("context %s not found", contextID)
	}

	// Деактивируем предыдущий активный контекст
	if cm.activeID != "" {
		if prevCtx, exists := cm.contexts[cm.activeID]; exists {
			prevCtx.IsActive = false
		}
	}

	// Активируем новый контекст
	ctx.IsActive = true
	ctx.UpdatedAt = time.Now()
	cm.activeID = contextID

	if err := cm.saveContexts(); err != nil {
		return fmt.Errorf("failed to save context: %w", err)
	}

	cm.logger.Info("Active context changed", "id", contextID, "name", ctx.Name)
	return nil
}

// GetActiveContext возвращает активный контекст
func (cm *ContextManager) GetActiveContext() (*WorkingContext, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.activeID == "" {
		return nil, fmt.Errorf("no active context set")
	}

	ctx, exists := cm.contexts[cm.activeID]
	if !exists {
		return nil, fmt.Errorf("active context %s not found", cm.activeID)
	}

	return ctx, nil
}

// UpdateContext обновляет существующий контекст
func (cm *ContextManager) UpdateContext(contextID string, updates map[string]interface{}) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ctx, exists := cm.contexts[contextID]
	if !exists {
		return fmt.Errorf("context %s not found", contextID)
	}

	// Обновляем поля
	for key, value := range updates {
		switch key {
		case "name":
			if v, ok := value.(string); ok {
				ctx.Name = v
			}
		case "description":
			if v, ok := value.(string); ok {
				ctx.Description = v
			}
		case "board_id":
			if v, ok := value.(string); ok {
				ctx.BoardID = v
			}
		case "project_id":
			if v, ok := value.(string); ok {
				ctx.ProjectID = v
			}
		case "provider_name":
			if v, ok := value.(string); ok {
				ctx.ProviderName = v
			}
		case "default_assignee":
			if v, ok := value.(string); ok {
				ctx.DefaultAssignee = v
			}
		case "default_labels":
			if v, ok := value.([]string); ok {
				ctx.DefaultLabels = v
			}
		case "default_priority":
			if v, ok := value.(string); ok {
				ctx.DefaultPriority = providers.TaskPriority(v)
			}
		case "project_type":
			if v, ok := value.(string); ok {
				ctx.ProjectType = v
			}
		case "complexity":
			if v, ok := value.(string); ok {
				ctx.Complexity = v
			}
		case "timeline":
			if v, ok := value.(int); ok {
				ctx.Timeline = v
			}
		case "team_size":
			if v, ok := value.(int); ok {
				ctx.TeamSize = v
			}
		case "ai_enabled":
			if v, ok := value.(bool); ok {
				ctx.AIEnabled = v
			}
		default:
			// Пользовательское поле
			ctx.CustomFields[key] = value
		}
	}

	ctx.UpdatedAt = time.Now()

	if err := cm.saveContexts(); err != nil {
		return fmt.Errorf("failed to save context: %w", err)
	}

	cm.logger.Info("Context updated", "id", contextID)
	return nil
}

// ListContexts возвращает список всех контекстов
func (cm *ContextManager) ListContexts() []*WorkingContext {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	contexts := make([]*WorkingContext, 0, len(cm.contexts))
	for _, ctx := range cm.contexts {
		contexts = append(contexts, ctx)
	}

	return contexts
}

// DeleteContext удаляет контекст
func (cm *ContextManager) DeleteContext(contextID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ctx, exists := cm.contexts[contextID]
	if !exists {
		return fmt.Errorf("context %s not found", contextID)
	}

	// Если это активный контекст, деактивируем его
	if cm.activeID == contextID {
		cm.activeID = ""
	}

	delete(cm.contexts, contextID)

	if err := cm.saveContexts(); err != nil {
		return fmt.Errorf("failed to save contexts: %w", err)
	}

	cm.logger.Info("Context deleted", "id", contextID, "name", ctx.Name)
	return nil
}

// UpdateStats обновляет статистику контекста
func (cm *ContextManager) UpdateStats(contextID string, stats map[string]interface{}) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ctx, exists := cm.contexts[contextID]
	if !exists {
		return fmt.Errorf("context %s not found", contextID)
	}

	if ctx.Stats == nil {
		ctx.Stats = &ContextStats{}
	}

	// Обновляем статистику
	for key, value := range stats {
		switch key {
		case "tasks_created":
			if v, ok := value.(int); ok {
				ctx.Stats.TasksCreated += v
			}
		case "tasks_completed":
			if v, ok := value.(int); ok {
				ctx.Stats.TasksCompleted += v
			}
		case "plans_generated":
			if v, ok := value.(int); ok {
				ctx.Stats.PlansGenerated += v
			}
		case "time_spent":
			if v, ok := value.(int64); ok {
				ctx.Stats.TotalTimeSpent += v
			}
		}
	}

	ctx.Stats.LastActivity = time.Now()
	
	// Пересчитываем success rate
	if ctx.Stats.TasksCreated > 0 {
		ctx.Stats.SuccessRate = float64(ctx.Stats.TasksCompleted) / float64(ctx.Stats.TasksCreated)
	}

	ctx.UpdatedAt = time.Now()

	if err := cm.saveContexts(); err != nil {
		return fmt.Errorf("failed to save contexts: %w", err)
	}

	return nil
}

// GetContextByBoardID находит контекст по board_id
func (cm *ContextManager) GetContextByBoardID(boardID string) (*WorkingContext, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, ctx := range cm.contexts {
		if ctx.BoardID == boardID {
			return ctx, nil
		}
	}

	return nil, fmt.Errorf("context with board_id %s not found", boardID)
}

// SetMultiProjectContext устанавливает контекст для работы с несколькими проектами
func (cm *ContextManager) SetMultiProjectContext(contextIDs []string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Проверяем что все контексты существуют
	for _, id := range contextIDs {
		if _, exists := cm.contexts[id]; !exists {
			return fmt.Errorf("context %s not found", id)
		}
	}

	// Деактивируем все контексты
	for _, ctx := range cm.contexts {
		ctx.IsActive = false
	}

	// Активируем выбранные контексты
	for _, id := range contextIDs {
		cm.contexts[id].IsActive = true
		cm.contexts[id].UpdatedAt = time.Now()
	}

	// Устанавливаем первый как основной активный
	if len(contextIDs) > 0 {
		cm.activeID = contextIDs[0]
	}

	if err := cm.saveContexts(); err != nil {
		return fmt.Errorf("failed to save contexts: %w", err)
	}

	cm.logger.Info("Multi-project context set", "contexts", len(contextIDs))
	return nil
}

// GetActiveContexts возвращает все активные контексты
func (cm *ContextManager) GetActiveContexts() []*WorkingContext {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var activeContexts []*WorkingContext
	for _, ctx := range cm.contexts {
		if ctx.IsActive {
			activeContexts = append(activeContexts, ctx)
		}
	}

	return activeContexts
}

// saveContexts сохраняет контексты в файл
func (cm *ContextManager) saveContexts() error {
	// Создаем директорию если не существует
	dir := filepath.Dir(cm.persistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data := struct {
		Contexts map[string]*WorkingContext `json:"contexts"`
		ActiveID string                     `json:"active_id"`
	}{
		Contexts: cm.contexts,
		ActiveID: cm.activeID,
	}

	file, err := os.Create(cm.persistPath)
	if err != nil {
		return fmt.Errorf("failed to create context file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode contexts: %w", err)
	}

	return nil
}

// loadContexts загружает контексты из файла
func (cm *ContextManager) loadContexts() {
	if _, err := os.Stat(cm.persistPath); os.IsNotExist(err) {
		cm.logger.Debug("Context file does not exist, starting fresh")
		return
	}

	file, err := os.Open(cm.persistPath)
	if err != nil {
		cm.logger.Error("Failed to open context file", err)
		return
	}
	defer file.Close()

	var data struct {
		Contexts map[string]*WorkingContext `json:"contexts"`
		ActiveID string                     `json:"active_id"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		cm.logger.Error("Failed to decode context file", err)
		return
	}

	cm.contexts = data.Contexts
	cm.activeID = data.ActiveID

	// Инициализируем пустые карты если nil
	for _, ctx := range cm.contexts {
		if ctx.CustomFields == nil {
			ctx.CustomFields = make(map[string]interface{})
		}
		if ctx.Stats == nil {
			ctx.Stats = &ContextStats{}
		}
	}

	cm.logger.Info("Contexts loaded", "count", len(cm.contexts), "active", cm.activeID)
}

// ContextConfig конфигурация для создания контекста
type ContextConfig struct {
	BoardID         string                 `json:"board_id"`
	ProjectID       string                 `json:"project_id"`
	ProviderName    string                 `json:"provider_name"`
	DefaultAssignee string                 `json:"default_assignee"`
	DefaultLabels   []string               `json:"default_labels"`
	DefaultPriority providers.TaskPriority  `json:"default_priority"`
	ProjectType     string                 `json:"project_type"`
	Complexity      string                 `json:"complexity"`
	Timeline        int                    `json:"timeline"`
	TeamSize        int                    `json:"team_size"`
	WorkflowType    string                 `json:"workflow_type"`
	AutoAssignment  bool                   `json:"auto_assignment"`
	AutoProgress    bool                   `json:"auto_progress"`
	AIEnabled       bool                   `json:"ai_enabled"`
	CustomFields    map[string]interface{} `json:"custom_fields"`
}

// ValidateProvider проверяет доступность провайдера для контекста
func (cm *ContextManager) ValidateProvider(ctx *WorkingContext, providerRegistry interface{}) error {
	// TODO: Реализовать валидацию провайдера через registry
	// Проверить что провайдер существует и доступен
	// Проверить что board_id и project_id валидны
	
	if ctx.ProviderName == "" {
		return fmt.Errorf("provider name is required")
	}
	
	if ctx.BoardID == "" {
		return fmt.Errorf("board_id is required")
	}
	
	if ctx.ProjectID == "" {
		return fmt.Errorf("project_id is required")
	}

	cm.logger.Debug("Context validation passed", "provider", ctx.ProviderName, "board", ctx.BoardID)
	return nil
}

// ApplyContextToTask применяет настройки контекста к задаче
func (cm *ContextManager) ApplyContextToTask(task *providers.UniversalTask) error {
	ctx, err := cm.GetActiveContext()
	if err != nil {
		return fmt.Errorf("no active context: %w", err)
	}

	// Применяем настройки контекста
	if task.ProjectID == "" {
		task.ProjectID = ctx.ProjectID
	}
	
	if task.AssigneeID == "" {
		task.AssigneeID = ctx.DefaultAssignee
	}
	
	if task.Priority == "" {
		task.Priority = ctx.DefaultPriority
	}
	
	// Добавляем лейблы по умолчанию
	if len(ctx.DefaultLabels) > 0 {
		taskLabels := make(map[string]bool)
		for _, label := range task.Labels {
			taskLabels[label] = true
		}
		
		for _, defaultLabel := range ctx.DefaultLabels {
			if !taskLabels[defaultLabel] {
				task.Labels = append(task.Labels, defaultLabel)
			}
		}
	}

	// Обновляем статистику контекста
	cm.UpdateStats(ctx.ID, map[string]interface{}{
		"tasks_created": 1,
	})

	return nil
}