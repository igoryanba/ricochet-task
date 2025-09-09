package workflow

import (
	"context"
	"time"
)

// WorkflowEngine управляет автоматическими workflow процессами
type WorkflowEngine struct {
	eventBus       *EventBus
	ruleEngine     *RuleEngine
	definitions    map[string]*WorkflowDefinition
	executionLog   []*WorkflowExecution
	notifications  *NotificationService
}

// WorkflowDefinition описывает workflow в декларативном виде
type WorkflowDefinition struct {
	Name        string                    `yaml:"name"`
	Description string                    `yaml:"description"`
	Version     string                    `yaml:"version"`
	Triggers    []TriggerDefinition       `yaml:"triggers"`
	Stages      map[string]*StageDefinition `yaml:"stages"`
	Variables   map[string]interface{}    `yaml:"variables"`
	Settings    WorkflowSettings          `yaml:"settings"`
}

// StageDefinition описывает этап workflow
type StageDefinition struct {
	Name           string                 `yaml:"name"`
	Dependencies   []string               `yaml:"dependencies"`
	AutoAssign     AutoAssignSettings     `yaml:"auto_assign"`
	Completion     CompletionSettings     `yaml:"completion"`
	Notifications  []NotificationRule     `yaml:"notifications"`
	Actions        []ActionDefinition     `yaml:"actions"`
	Conditions     []ConditionDefinition  `yaml:"conditions"`
	Timeout        time.Duration          `yaml:"timeout"`
}

// AutoAssignSettings настройки автоматического назначения задач
type AutoAssignSettings struct {
	Strategy    string            `yaml:"strategy"`    // skills_based, workload_balanced, round_robin
	Skills      []string          `yaml:"skills"`
	MaxLoad     int               `yaml:"max_load"`
	Preferences map[string]string `yaml:"preferences"`
}

// CompletionSettings настройки завершения этапа
type CompletionSettings struct {
	Trigger    string   `yaml:"trigger"`    // manual, auto, condition
	Conditions []string `yaml:"conditions"` // git_commits, tests_passed, review_approved
	Required   bool     `yaml:"required"`
}

// TriggerDefinition события, запускающие workflow
type TriggerDefinition struct {
	Type       string                 `yaml:"type"`        // ai_plan_created, task_status_changed, git_push, etc.
	Conditions map[string]interface{} `yaml:"conditions"`
	Actions    []string               `yaml:"actions"`
}

// ActionDefinition действия workflow
type ActionDefinition struct {
	Type       string                 `yaml:"type"`        // create_task, update_status, send_notification, etc.
	Parameters map[string]interface{} `yaml:"parameters"`
	Condition  string                 `yaml:"condition"`   // optional condition for execution
}

// ConditionDefinition условия для выполнения действий
type ConditionDefinition struct {
	Field    string      `yaml:"field"`
	Operator string      `yaml:"operator"` // eq, ne, gt, lt, contains, matches
	Value    interface{} `yaml:"value"`
}

// NotificationRule правила уведомлений
type NotificationRule struct {
	Event    string   `yaml:"event"`    // stage_start, stage_complete, task_assigned, etc.
	Channels []string `yaml:"channels"` // email, slack, teams, webhook
	Template string   `yaml:"template"`
	Users    []string `yaml:"users"`
}

// WorkflowSettings общие настройки workflow
type WorkflowSettings struct {
	MaxConcurrency   int           `yaml:"max_concurrency"`
	DefaultTimeout   time.Duration `yaml:"default_timeout"`
	RetryPolicy      RetryPolicy   `yaml:"retry_policy"`
	LogLevel         string        `yaml:"log_level"`
	AIEnabled        bool          `yaml:"ai_enabled"`
}

// RetryPolicy политика повторных попыток
type RetryPolicy struct {
	MaxRetries int           `yaml:"max_retries"`
	BackoffMin time.Duration `yaml:"backoff_min"`
	BackoffMax time.Duration `yaml:"backoff_max"`
}

// Event базовый интерфейс для всех событий
type Event interface {
	GetType() string
	GetTimestamp() time.Time
	GetSource() string
	GetData() map[string]interface{}
}

// WorkflowEvent события workflow
type WorkflowEvent struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	WorkflowID string                `json:"workflow_id"`
	StageID   string                 `json:"stage_id,omitempty"`
}

func (e *WorkflowEvent) GetType() string                 { return e.Type }
func (e *WorkflowEvent) GetTimestamp() time.Time        { return e.Timestamp }
func (e *WorkflowEvent) GetSource() string              { return e.Source }
func (e *WorkflowEvent) GetData() map[string]interface{} { return e.Data }

// TaskEvent события задач
type TaskEvent struct {
	Type       string                 `json:"type"`
	Timestamp  time.Time              `json:"timestamp"`
	Source     string                 `json:"source"`
	Data       map[string]interface{} `json:"data"`
	TaskID     string                 `json:"task_id"`
	OldStatus  string                 `json:"old_status,omitempty"`
	NewStatus  string                 `json:"new_status,omitempty"`
	Assignee   string                 `json:"assignee,omitempty"`
	ProjectID  string                 `json:"project_id"`
	ProviderID string                 `json:"provider_id"`
}

func (e *TaskEvent) GetType() string                 { return e.Type }
func (e *TaskEvent) GetTimestamp() time.Time        { return e.Timestamp }
func (e *TaskEvent) GetSource() string              { return e.Source }
func (e *TaskEvent) GetData() map[string]interface{} { return e.Data }

// GitEvent события Git
type GitEvent struct {
	Type       string                 `json:"type"`
	Timestamp  time.Time              `json:"timestamp"`
	Source     string                 `json:"source"`
	Data       map[string]interface{} `json:"data"`
	Repository string                 `json:"repository"`
	Branch     string                 `json:"branch"`
	CommitHash string                 `json:"commit_hash"`
	Author     string                 `json:"author"`
	TaskRefs   []string               `json:"task_refs"` // extracted task IDs from commit message
}

func (e *GitEvent) GetType() string                 { return e.Type }
func (e *GitEvent) GetTimestamp() time.Time        { return e.Timestamp }
func (e *GitEvent) GetSource() string              { return e.Source }
func (e *GitEvent) GetData() map[string]interface{} { return e.Data }

// WorkflowExecution исполнение workflow
type WorkflowExecution struct {
	ID           string                 `json:"id"`
	WorkflowName string                 `json:"workflow_name"`
	Status       ExecutionStatus        `json:"status"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      *time.Time             `json:"end_time,omitempty"`
	Context      map[string]interface{} `json:"context"`
	Stages       []*StageExecution      `json:"stages"`
	Events       []*WorkflowEvent       `json:"events"`
	Error        string                 `json:"error,omitempty"`
}

// StageExecution исполнение этапа
type StageExecution struct {
	ID         string                 `json:"id"`
	StageName  string                 `json:"stage_name"`
	Status     ExecutionStatus        `json:"status"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    *time.Time             `json:"end_time,omitempty"`
	Context    map[string]interface{} `json:"context"`
	Actions    []*ActionExecution     `json:"actions"`
	Error      string                 `json:"error,omitempty"`
}

// ActionExecution исполнение действия
type ActionExecution struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Status    ExecutionStatus        `json:"status"`
	StartTime time.Time              `json:"start_time"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Input     map[string]interface{} `json:"input"`
	Output    map[string]interface{} `json:"output"`
	Error     string                 `json:"error,omitempty"`
}

// ExecutionStatus статус исполнения
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusCompleted ExecutionStatus = "completed"
	StatusFailed    ExecutionStatus = "failed"
	StatusCancelled ExecutionStatus = "cancelled"
	StatusRetrying  ExecutionStatus = "retrying"
)

// EventHandler обработчик событий
type EventHandler interface {
	Handle(ctx context.Context, event Event) error
	CanHandle(eventType string) bool
}

// ActionExecutor исполнитель действий
type ActionExecutor interface {
	Execute(ctx context.Context, action *ActionDefinition, context map[string]interface{}) (map[string]interface{}, error)
	GetType() string
}

// ConditionEvaluator оценщик условий
type ConditionEvaluator interface {
	Evaluate(ctx context.Context, condition *ConditionDefinition, context map[string]interface{}) (bool, error)
}

// AssignmentStrategy стратегия назначения задач
type AssignmentStrategy interface {
	Assign(ctx context.Context, task TaskInfo, teamInfo TeamInfo) (string, error)
	GetName() string
}

// TaskInfo информация о задаче для назначения
type TaskInfo struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Skills      []string          `json:"skills"`
	Priority    string            `json:"priority"`
	Complexity  string            `json:"complexity"`
	Estimate    int               `json:"estimate"`
	Tags        []string          `json:"tags"`
	Context     map[string]interface{} `json:"context"`
}

// TeamInfo информация о команде
type TeamInfo struct {
	Members []TeamMember `json:"members"`
}

// TeamMember участник команды
type TeamMember struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Email        string            `json:"email"`
	Skills       []string          `json:"skills"`
	CurrentLoad  int               `json:"current_load"`
	MaxLoad      int               `json:"max_load"`
	Timezone     string            `json:"timezone"`
	Preferences  map[string]string `json:"preferences"`
	Performance  PerformanceMetrics `json:"performance"`
}

// PerformanceMetrics метрики производительности
type PerformanceMetrics struct {
	AverageVelocity  float64 `json:"average_velocity"`
	QualityScore     float64 `json:"quality_score"`
	CompletionRate   float64 `json:"completion_rate"`
	ResponseTime     int     `json:"response_time_hours"`
	LastActiveDate   time.Time `json:"last_active_date"`
}

// NotificationService сервис уведомлений
type NotificationService struct {
	channels map[string]NotificationChannel
}

// NotificationChannel канал уведомлений
type NotificationChannel interface {
	Send(ctx context.Context, notification *Notification) error
	GetType() string
}

// Notification уведомление
type Notification struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Priority  string                 `json:"priority"`
	Recipients []string              `json:"recipients"`
	Data      map[string]interface{} `json:"data"`
	Template  string                 `json:"template"`
	Timestamp time.Time              `json:"timestamp"`
}