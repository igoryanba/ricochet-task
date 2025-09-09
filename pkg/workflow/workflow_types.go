package workflow

import (
	"context"
	"time"
)

// WorkflowInstance экземпляр выполняющегося workflow
type WorkflowInstance struct {
	ID           string                 `json:"id"`
	Definition   *WorkflowDefinition    `json:"definition"`
	Status       string                 `json:"status"` // created, running, completed, failed, paused
	CreatedAt    time.Time              `json:"created_at"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	CurrentStage string                 `json:"current_stage"`
	Context      map[string]interface{} `json:"context"`
	Tasks        map[string]*TaskInstance `json:"tasks"`
	Progress     float64                `json:"progress"` // 0.0 - 1.0
	Error        string                 `json:"error,omitempty"`
}

// TaskInstance экземпляр выполняющейся задачи
type TaskInstance struct {
	ID           string                 `json:"id"`
	Definition   *TaskDefinition        `json:"definition"`
	Status       string                 `json:"status"` // created, assigned, in_progress, completed, failed
	CreatedAt    time.Time              `json:"created_at"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	AssignedTo   string                 `json:"assigned_to,omitempty"`
	Context      map[string]interface{} `json:"context"`
	Progress     float64                `json:"progress"`
	Error        string                 `json:"error,omitempty"`
}

// TaskDefinition определение задачи
type TaskDefinition struct {
	Name              string        `yaml:"name" json:"name"`
	Type              string        `yaml:"type" json:"type"` // manual, automated, approval
	Description       string        `yaml:"description" json:"description"`
	AutoAssign        bool          `yaml:"auto_assign" json:"auto_assign"`
	Requirements      []string      `yaml:"requirements" json:"requirements"`
	EstimatedDuration time.Duration `yaml:"estimated_duration" json:"estimated_duration"`
	Priority          string        `yaml:"priority" json:"priority"`
	DependsOn         []string      `yaml:"depends_on" json:"depends_on"`
	Conditions        []ConditionDefinition   `yaml:"conditions" json:"conditions"`
}

// WorkflowStage стадия workflow
type WorkflowStage struct {
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description" json:"description"`
	Tasks       []*TaskDefinition `yaml:"tasks" json:"tasks"`
	DependsOn   []string          `yaml:"depends_on" json:"depends_on"`
	Parallel    bool              `yaml:"parallel" json:"parallel"`
	Conditions  []ConditionDefinition       `yaml:"conditions" json:"conditions"`
}

// EventBusConfig конфигурация Event Bus
type EventBusConfig struct {
	MaxQueueSize    int    `json:"max_queue_size"`
	WorkerPoolSize  int    `json:"worker_pool_size"`
	EnableMetrics   bool   `json:"enable_metrics"`
	RetryAttempts   int    `json:"retry_attempts"`
	RetryDelay      time.Duration `json:"retry_delay"`
}

// RuleEngineConfig конфигурация Rule Engine
type RuleEngineConfig struct {
	EnableCaching     bool          `json:"enable_caching"`
	CacheSize         int           `json:"cache_size"`
	CacheTTL          time.Duration `json:"cache_ttl"`
	MaxRuleDepth      int           `json:"max_rule_depth"`
	EnableMetrics     bool          `json:"enable_metrics"`
}

// ProgressTrackingConfig конфигурация отслеживания прогресса
type ProgressTrackingConfig struct {
	EnableGitIntegration bool          `json:"enable_git_integration"`
	EnableMetrics        bool          `json:"enable_metrics"`
	UpdateInterval       time.Duration `json:"update_interval"`
	GitConfig            *GitConfig    `json:"git_config,omitempty"`
}

// GitConfig конфигурация Git интеграции
type GitConfig struct {
	RepoPath       string `json:"repo_path"`
	DefaultBranch  string `json:"default_branch"`
	RemoteURL      string `json:"remote_url"`
	Username       string `json:"username"`
	Token          string `json:"token"`
}

// AutoAssignmentConfig конфигурация автоназначения
type AutoAssignmentConfig struct {
	EnableAI           bool          `json:"enable_ai"`
	MaxRetries         int           `json:"max_retries"`
	AssignmentTimeout  time.Duration `json:"assignment_timeout"`
	FallbackStrategy   string        `json:"fallback_strategy"` // round_robin, random, load_based
	EnableMetrics      bool          `json:"enable_metrics"`
}

// Простые заглушки для отсутствующих типов

// ProgressTracker заглушка для отслеживания прогресса
type ProgressTracker struct {
	config *ProgressTrackingConfig
	logger Logger
}

// NewProgressTracker создает новый трекер прогресса
func NewProgressTracker(config *ProgressTrackingConfig, logger Logger) (*ProgressTracker, error) {
	if config == nil {
		config = &ProgressTrackingConfig{
			EnableGitIntegration: false,
			EnableMetrics:        true,
			UpdateInterval:       time.Minute,
		}
	}

	return &ProgressTracker{
		config: config,
		logger: logger,
	}, nil
}

// HandleEvent обрабатывает события
func (pt *ProgressTracker) HandleEvent(ctx context.Context, event Event) error {
	pt.logger.Debug("Progress tracking event", "event_type", event.GetType())
	return nil
}

// AutoAssignment заглушка для автоназначения
type AutoAssignment struct {
	config   *AutoAssignmentConfig
	aiChains interface{} // AI chains interface
	eventBus *EventBus
	logger   Logger
}

// NewAutoAssignment создает новую систему автоназначения
func NewAutoAssignment(aiChains interface{}, eventBus *EventBus, config *AutoAssignmentConfig, logger Logger) *AutoAssignment {
	if config == nil {
		config = &AutoAssignmentConfig{
			EnableAI:          true,
			MaxRetries:        3,
			AssignmentTimeout: 5 * time.Minute,
			FallbackStrategy:  "round_robin",
			EnableMetrics:     true,
		}
	}

	return &AutoAssignment{
		config:   config,
		aiChains: aiChains,
		eventBus: eventBus,
		logger:   logger,
	}
}

// ProcessEvent обрабатывает события
func (aa *AutoAssignment) ProcessEvent(ctx context.Context, event Event) error {
	aa.logger.Debug("Auto assignment event", "event_type", event.GetType())
	return nil
}