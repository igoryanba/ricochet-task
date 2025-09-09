package providers

import (
	"encoding/json"
	"time"
)

// UniversalTask represents a task/issue across all providers
type UniversalTask struct {
	// Core identifiers
	ID         string `json:"id"`
	ExternalID string `json:"externalId"` // Provider-specific ID
	Key        string `json:"key,omitempty"` // Human-readable key (e.g. PROJ-123)

	// Basic fields
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Status      TaskStatus `json:"status"`
	Priority    TaskPriority `json:"priority"`
	Type        TaskType `json:"type"`

	// Project and organization
	ProjectID   string `json:"projectId"`
	ProjectKey  string `json:"projectKey,omitempty"`
	BoardID     string `json:"boardId,omitempty"`
	ColumnID    string `json:"columnId,omitempty"`
	SprintID    string `json:"sprintId,omitempty"`

	// People
	AssigneeID  string `json:"assigneeId,omitempty"`
	ReporterID  string `json:"reporterId,omitempty"`
	CreatorID   string `json:"creatorId,omitempty"`

	// Hierarchy and relationships
	ParentID    string   `json:"parentId,omitempty"`
	SubtaskIDs  []string `json:"subtaskIds,omitempty"`
	EpicID      string   `json:"epicId,omitempty"`

	// Dependencies
	BlockedBy   []string `json:"blockedBy,omitempty"`
	Blocks      []string `json:"blocks,omitempty"`
	RelatedTo   []string `json:"relatedTo,omitempty"`
	DuplicateOf string   `json:"duplicateOf,omitempty"`

	// Metadata
	Labels        []string               `json:"labels,omitempty"`
	Tags          []string               `json:"tags,omitempty"`
	CustomFields  map[string]interface{} `json:"customFields,omitempty"`
	Attachments   []*Attachment          `json:"attachments,omitempty"`
	Comments      []*Comment             `json:"comments,omitempty"`

	// Time tracking
	EstimatedTime   *time.Duration `json:"estimatedTime,omitempty"`
	TimeSpent       *time.Duration `json:"timeSpent,omitempty"`
	RemainingTime   *time.Duration `json:"remainingTime,omitempty"`
	
	// Timestamps
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
	StartDate   *time.Time `json:"startDate,omitempty"`
	ResolvedAt  *time.Time `json:"resolvedAt,omitempty"`

	// Ricochet integration
	RicochetMetadata *RicochetTaskMetadata `json:"ricochetMetadata,omitempty"`

	// Provider-specific data
	ProviderData   map[string]interface{} `json:"providerData,omitempty"`
	ProviderName   string                 `json:"providerName"`
	ProviderConfig *ProviderConfig        `json:"-"` // Not serialized
}

// TaskStatus represents task status across providers
type TaskStatus struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Category    StatusCategory `json:"category"`
	Color       string         `json:"color,omitempty"`
	Order       int            `json:"order"`
	IsFinal     bool           `json:"isFinal"`
	Description string         `json:"description,omitempty"`
}

func (ts *TaskStatus) IsCategory(category StatusCategory) bool {
	return ts.Category == category
}

type StatusCategory string

const (
	StatusCategoryTodo        StatusCategory = "todo"
	StatusCategoryInProgress  StatusCategory = "in_progress"
	StatusCategoryDone        StatusCategory = "done"
	StatusCategoryBlocked     StatusCategory = "blocked"
	StatusCategoryCancelled   StatusCategory = "cancelled"
	StatusCategoryReview      StatusCategory = "review"
	StatusCategoryTesting     StatusCategory = "testing"
)

// TaskPriority represents task priority levels
type TaskPriority string

const (
	TaskPriorityLowest   TaskPriority = "lowest"
	TaskPriorityLow      TaskPriority = "low"
	TaskPriorityMedium   TaskPriority = "medium"
	TaskPriorityHigh     TaskPriority = "high"
	TaskPriorityHighest  TaskPriority = "highest"
	TaskPriorityCritical TaskPriority = "critical"
)

// TaskType represents different types of tasks
type TaskType string

const (
	TaskTypeTask        TaskType = "task"
	TaskTypeStory       TaskType = "story"
	TaskTypeBug         TaskType = "bug"
	TaskTypeEpic        TaskType = "epic"
	TaskTypeSubtask     TaskType = "subtask"
	TaskTypeFeature     TaskType = "feature"
	TaskTypeImprovement TaskType = "improvement"
	TaskTypeSpike       TaskType = "spike"
	TaskTypeResearch    TaskType = "research"
	TaskTypeChore       TaskType = "chore"
)

// UniversalBoard represents a board/project across providers
type UniversalBoard struct {
	// Core identifiers
	ID         string `json:"id"`
	ExternalID string `json:"externalId"`
	Key        string `json:"key,omitempty"`

	// Basic fields
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Type        BoardType `json:"type"`
	ProjectID   string    `json:"projectId"`

	// Configuration
	Columns     []*BoardColumn    `json:"columns"`
	Swimlanes   []*BoardSwimlane  `json:"swimlanes,omitempty"`
	Settings    *BoardSettings    `json:"settings,omitempty"`
	Automation  []*AutomationRule `json:"automation,omitempty"`

	// Permissions
	IsPrivate   bool     `json:"isPrivate"`
	Members     []string `json:"members,omitempty"`
	Admins      []string `json:"admins,omitempty"`

	// Metadata
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	CreatorID   string    `json:"creatorId,omitempty"`

	// Provider-specific
	ProviderData map[string]interface{} `json:"providerData,omitempty"`
	ProviderName string                 `json:"providerName"`
}

type BoardType string

const (
	BoardTypeKanban BoardType = "kanban"
	BoardTypeScrum  BoardType = "scrum"
	BoardTypeList   BoardType = "list"
	BoardTypeTable  BoardType = "table"
	BoardTypeCustom BoardType = "custom"
)

// BoardColumn represents a column in a board
type BoardColumn struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Status      TaskStatus  `json:"status"`
	Order       int         `json:"order"`
	WIPLimit    int         `json:"wipLimit,omitempty"`
	IsCollapsed bool        `json:"isCollapsed"`
	Color       string      `json:"color,omitempty"`
}

// BoardSwimlane represents a swimlane in a board
type BoardSwimlane struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Query       string `json:"query,omitempty"`
	Order       int    `json:"order"`
	IsCollapsed bool   `json:"isCollapsed"`
}

// BoardSettings contains board configuration
type BoardSettings struct {
	EstimationField     string `json:"estimationField,omitempty"`
	ShowEstimation      bool   `json:"showEstimation"`
	ShowSubtasks        bool   `json:"showSubtasks"`
	ShowCards           bool   `json:"showCards"`
	CardFields          []string `json:"cardFields,omitempty"`
	DefaultAssignee     string   `json:"defaultAssignee,omitempty"`
	AutoProgressEnabled bool     `json:"autoProgressEnabled"`
}

// AutomationRule represents board automation
type AutomationRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Trigger     AutomationTrigger      `json:"trigger"`
	Conditions  []AutomationCondition  `json:"conditions,omitempty"`
	Actions     []AutomationAction     `json:"actions"`
	IsEnabled   bool                   `json:"isEnabled"`
}

type AutomationTrigger struct {
	Type   TriggerType            `json:"type"`
	Config map[string]interface{} `json:"config,omitempty"`
}

type TriggerType string

const (
	TriggerTypeStatusChange  TriggerType = "status_change"
	TriggerTypeAssignment    TriggerType = "assignment"
	TriggerTypeCreation      TriggerType = "creation"
	TriggerTypeUpdate        TriggerType = "update"
	TriggerTypeScheduled     TriggerType = "scheduled"
	TriggerTypeWebhook       TriggerType = "webhook"
)

type AutomationCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type AutomationAction struct {
	Type   ActionType             `json:"type"`
	Config map[string]interface{} `json:"config"`
}

type ActionType string

const (
	ActionTypeUpdateField    ActionType = "update_field"
	ActionTypeAssign         ActionType = "assign"
	ActionTypeTransition     ActionType = "transition"
	ActionTypeCreateSubtask  ActionType = "create_subtask"
	ActionTypeAddComment     ActionType = "add_comment"
	ActionTypeNotify         ActionType = "notify"
	ActionTypeExecuteChain   ActionType = "execute_chain"
)

// RicochetTaskMetadata contains Ricochet-specific task data
type RicochetTaskMetadata struct {
	// Ricochet task system integration
	RicochetTaskID  string   `json:"ricochetTaskId,omitempty"`
	ChainID         string   `json:"chainId,omitempty"`
	CheckpointIDs   []string `json:"checkpointIds,omitempty"`

	// AI execution metadata
	AIExecutionState   AIExecutionState     `json:"aiExecutionState,omitempty"`
	LastAIExecution    *time.Time           `json:"lastAiExecution,omitempty"`
	AIExecutionHistory []*AIExecutionRecord `json:"aiExecutionHistory,omitempty"`

	// Quality control
	QualityGates    []QualityGateResult `json:"qualityGates,omitempty"`
	TestResults     *TestResults        `json:"testResults,omitempty"`
	CodeReviewData  *CodeReviewData     `json:"codeReviewData,omitempty"`

	// Synchronization
	LastSyncTime    time.Time      `json:"lastSyncTime"`
	SyncConflicts   []SyncConflict `json:"syncConflicts,omitempty"`
	SyncStatus      SyncStatus     `json:"syncStatus"`

	// Configuration
	AutoExecution   bool                   `json:"autoExecution"`
	ExecutionConfig map[string]interface{} `json:"executionConfig,omitempty"`
}

type AIExecutionState string

const (
	AIExecutionStateIdle       AIExecutionState = "idle"
	AIExecutionStatePending    AIExecutionState = "pending"
	AIExecutionStateRunning    AIExecutionState = "running"
	AIExecutionStateCompleted  AIExecutionState = "completed"
	AIExecutionStateFailed     AIExecutionState = "failed"
	AIExecutionStateCancelled  AIExecutionState = "cancelled"
)

type AIExecutionRecord struct {
	ID          string           `json:"id"`
	ChainName   string           `json:"chainName"`
	StartTime   time.Time        `json:"startTime"`
	EndTime     *time.Time       `json:"endTime,omitempty"`
	Status      AIExecutionState `json:"status"`
	Result      interface{}      `json:"result,omitempty"`
	Error       string           `json:"error,omitempty"`
	Logs        []string         `json:"logs,omitempty"`
	TokensUsed  int              `json:"tokensUsed,omitempty"`
	Cost        float64          `json:"cost,omitempty"`
}

type QualityGateResult struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"` // passed, failed, skipped
	Score       *float64  `json:"score,omitempty"`
	Details     string    `json:"details,omitempty"`
	CheckedAt   time.Time `json:"checkedAt"`
	IsBlocking  bool      `json:"isBlocking"`
}

type TestResults struct {
	TotalTests   int     `json:"totalTests"`
	PassedTests  int     `json:"passedTests"`
	FailedTests  int     `json:"failedTests"`
	SkippedTests int     `json:"skippedTests"`
	Coverage     float64 `json:"coverage,omitempty"`
	TestFiles    []string `json:"testFiles,omitempty"`
	ReportURL    string   `json:"reportUrl,omitempty"`
}

type CodeReviewData struct {
	ReviewerID    string    `json:"reviewerId,omitempty"`
	Status        string    `json:"status"` // pending, approved, changes_requested
	Comments      []string  `json:"comments,omitempty"`
	Score         *float64  `json:"score,omitempty"`
	ReviewedAt    *time.Time `json:"reviewedAt,omitempty"`
	PullRequestURL string   `json:"pullRequestUrl,omitempty"`
}

type SyncStatus string

const (
	SyncStatusSynced     SyncStatus = "synced"
	SyncStatusPending    SyncStatus = "pending"
	SyncStatusConflict   SyncStatus = "conflict"
	SyncStatusFailed     SyncStatus = "failed"
	SyncStatusDisabled   SyncStatus = "disabled"
)

// Supporting types for operations
type TaskUpdate struct {
	Title         *string                `json:"title,omitempty"`
	Description   *string                `json:"description,omitempty"`
	Status        *TaskStatus            `json:"status,omitempty"`
	Priority      *TaskPriority          `json:"priority,omitempty"`
	AssigneeID    *string                `json:"assigneeId,omitempty"`
	DueDate       *time.Time             `json:"dueDate,omitempty"`
	Labels        []string               `json:"labels,omitempty"`
	CustomFields  map[string]interface{} `json:"customFields,omitempty"`
	EstimatedTime *time.Duration         `json:"estimatedTime,omitempty"`
}

type TaskFilters struct {
	ProjectID    string       `json:"projectId,omitempty"`
	BoardID      string       `json:"boardId,omitempty"`
	AssigneeID   string       `json:"assigneeId,omitempty"`
	ReporterID   string       `json:"reporterId,omitempty"`
	Status       []string     `json:"status,omitempty"`
	Priority     []string     `json:"priority,omitempty"`
	Type         []string     `json:"type,omitempty"`
	Labels       []string     `json:"labels,omitempty"`
	CreatedAfter *time.Time   `json:"createdAfter,omitempty"`
	CreatedBefore *time.Time  `json:"createdBefore,omitempty"`
	UpdatedAfter *time.Time   `json:"updatedAfter,omitempty"`
	UpdatedBefore *time.Time  `json:"updatedBefore,omitempty"`
	DueDateAfter *time.Time   `json:"dueDateAfter,omitempty"`
	DueDateBefore *time.Time  `json:"dueDateBefore,omitempty"`
	Query        string       `json:"query,omitempty"`
	Limit        int          `json:"limit,omitempty"`
	Offset       int          `json:"offset,omitempty"`
}

type BoardUpdate struct {
	Name        *string        `json:"name,omitempty"`
	Description *string        `json:"description,omitempty"`
	Settings    *BoardSettings `json:"settings,omitempty"`
	IsPrivate   *bool          `json:"isPrivate,omitempty"`
}

// Supporting entities
type Attachment struct {
	ID          string    `json:"id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"contentType"`
	Size        int64     `json:"size"`
	URL         string    `json:"url"`
	UploadedBy  string    `json:"uploadedBy,omitempty"`
	UploadedAt  time.Time `json:"uploadedAt"`
}

type Comment struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	AuthorID  string    `json:"authorId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	IsEdited  bool      `json:"isEdited"`
	ParentID  string    `json:"parentId,omitempty"`
}

// Sync related types
type SyncEvent struct {
	ID           string                 `json:"id"`
	Type         SyncEventType          `json:"type"`
	Source       string                 `json:"source"` // provider name
	Target       string                 `json:"target,omitempty"`
	TaskID       string                 `json:"taskId,omitempty"`
	BoardID      string                 `json:"boardId,omitempty"`
	Changes      map[string]interface{} `json:"changes,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	ProcessedAt  *time.Time             `json:"processedAt,omitempty"`
	Error        string                 `json:"error,omitempty"`
}

type SyncEventType string

const (
	SyncEventTaskCreated  SyncEventType = "task_created"
	SyncEventTaskUpdated  SyncEventType = "task_updated"
	SyncEventTaskDeleted  SyncEventType = "task_deleted"
	SyncEventBoardCreated SyncEventType = "board_created"
	SyncEventBoardUpdated SyncEventType = "board_updated"
	SyncEventBoardDeleted SyncEventType = "board_deleted"
)

type SyncConflict struct {
	ID           string                 `json:"id"`
	TaskID       string                 `json:"taskId"`
	Field        string                 `json:"field"`
	SourceValue  interface{}            `json:"sourceValue"`
	TargetValue  interface{}            `json:"targetValue"`
	Source       string                 `json:"source"`
	Target       string                 `json:"target"`
	DetectedAt   time.Time              `json:"detectedAt"`
	ResolvedAt   *time.Time             `json:"resolvedAt,omitempty"`
	Resolution   *ConflictResolution    `json:"resolution,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

type ConflictResolution struct {
	Strategy      ConflictStrategy       `json:"strategy"`
	ResolvedValue interface{}            `json:"resolvedValue,omitempty"`
	ResolvedBy    string                 `json:"resolvedBy,omitempty"`
	Reason        string                 `json:"reason,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type ConflictStrategy string

const (
	ConflictResolveUseSource ConflictStrategy = "use_source"
	ConflictResolveUseTarget ConflictStrategy = "use_target"
	ConflictResolveMerge     ConflictStrategy = "merge"
	ConflictResolveManual    ConflictStrategy = "manual"
	ConflictResolveSkip      ConflictStrategy = "skip"
)

type ConflictFilters struct {
	TaskID     string     `json:"taskId,omitempty"`
	Source     string     `json:"source,omitempty"`
	Target     string     `json:"target,omitempty"`
	Field      string     `json:"field,omitempty"`
	Status     string     `json:"status,omitempty"` // pending, resolved
	DateAfter  *time.Time `json:"dateAfter,omitempty"`
	DateBefore *time.Time `json:"dateBefore,omitempty"`
}

// Event types for async notifications
type UniversalEvent struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Source    string                 `json:"source"`
	TaskID    string                 `json:"taskId,omitempty"`
	BoardID   string                 `json:"boardId,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

type EventType string

const (
	EventTypeTaskCreated     EventType = "task.created"
	EventTypeTaskUpdated     EventType = "task.updated"
	EventTypeTaskDeleted     EventType = "task.deleted"
	EventTypeTaskAssigned    EventType = "task.assigned"
	EventTypeTaskStatusChanged EventType = "task.status_changed"
	EventTypeBoardCreated    EventType = "board.created"
	EventTypeBoardUpdated    EventType = "board.updated"
	EventTypeBoardDeleted    EventType = "board.deleted"
	EventTypeCommentAdded    EventType = "comment.added"
	EventTypeAttachmentAdded EventType = "attachment.added"
)

// Helper methods for UniversalTask
func (t *UniversalTask) GetDisplayID() string {
	if t.Key != "" {
		return t.Key
	}
	if t.ExternalID != "" {
		return t.ExternalID
	}
	return t.ID
}

func (t *UniversalTask) IsCompleted() bool {
	return t.Status.IsFinal || t.Status.Category == StatusCategoryDone
}

func (t *UniversalTask) IsBlocked() bool {
	return t.Status.Category == StatusCategoryBlocked || len(t.BlockedBy) > 0
}

func (t *UniversalTask) HasSubtasks() bool {
	return len(t.SubtaskIDs) > 0
}

func (t *UniversalTask) IsOverdue() bool {
	return t.DueDate != nil && t.DueDate.Before(time.Now()) && !t.IsCompleted()
}

func (t *UniversalTask) GetAge() time.Duration {
	return time.Since(t.CreatedAt)
}

func (t *UniversalTask) HasLabel(label string) bool {
	for _, l := range t.Labels {
		if l == label {
			return true
		}
	}
	return false
}

// JSON marshaling helpers
func (t *UniversalTask) MarshalJSON() ([]byte, error) {
	type Alias UniversalTask
	return json.Marshal(&struct {
		*Alias
		EstimatedTimeSeconds *int64 `json:"estimatedTimeSeconds,omitempty"`
		TimeSpentSeconds     *int64 `json:"timeSpentSeconds,omitempty"`
		RemainingTimeSeconds *int64 `json:"remainingTimeSeconds,omitempty"`
	}{
		Alias:                (*Alias)(t),
		EstimatedTimeSeconds: durationToSecondsPtr(t.EstimatedTime),
		TimeSpentSeconds:     durationToSecondsPtr(t.TimeSpent),
		RemainingTimeSeconds: durationToSecondsPtr(t.RemainingTime),
	})
}

func (t *UniversalTask) UnmarshalJSON(data []byte) error {
	type Alias UniversalTask
	aux := &struct {
		*Alias
		EstimatedTimeSeconds *int64 `json:"estimatedTimeSeconds,omitempty"`
		TimeSpentSeconds     *int64 `json:"timeSpentSeconds,omitempty"`
		RemainingTimeSeconds *int64 `json:"remainingTimeSeconds,omitempty"`
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	t.EstimatedTime = secondsToDurationPtr(aux.EstimatedTimeSeconds)
	t.TimeSpent = secondsToDurationPtr(aux.TimeSpentSeconds)
	t.RemainingTime = secondsToDurationPtr(aux.RemainingTimeSeconds)

	return nil
}

func durationToSecondsPtr(d *time.Duration) *int64 {
	if d == nil {
		return nil
	}
	seconds := int64(d.Seconds())
	return &seconds
}

func secondsToDurationPtr(seconds *int64) *time.Duration {
	if seconds == nil {
		return nil
	}
	duration := time.Duration(*seconds) * time.Second
	return &duration
}

// Additional types needed for interfaces
type WorkflowRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	BoardID     string                 `json:"boardId"`
	Trigger     AutomationTrigger      `json:"trigger"`
	Conditions  []AutomationCondition  `json:"conditions,omitempty"`
	Actions     []AutomationAction     `json:"actions"`
	IsEnabled   bool                   `json:"isEnabled"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

type SearchQuery struct {
	Query     string            `json:"query"`
	Filters   *TaskFilters      `json:"filters,omitempty"`
	SortBy    string            `json:"sortBy,omitempty"`
	SortOrder string            `json:"sortOrder,omitempty"`
	Limit     int               `json:"limit,omitempty"`
	Offset    int               `json:"offset,omitempty"`
}

type SavedSearch struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Query       SearchQuery  `json:"query"`
	CreatedBy   string       `json:"createdBy"`
	CreatedAt   time.Time    `json:"createdAt"`
	IsShared    bool         `json:"isShared"`
}

type MetricsFilters struct {
	ProjectID    string     `json:"projectId,omitempty"`
	TeamID       string     `json:"teamId,omitempty"`
	AssigneeID   string     `json:"assigneeId,omitempty"`
	StartDate    *time.Time `json:"startDate,omitempty"`
	EndDate      *time.Time `json:"endDate,omitempty"`
	TaskTypes    []string   `json:"taskTypes,omitempty"`
	Priorities   []string   `json:"priorities,omitempty"`
}

type TaskMetrics struct {
	TotalTasks      int                    `json:"totalTasks"`
	CompletedTasks  int                    `json:"completedTasks"`
	InProgressTasks int                    `json:"inProgressTasks"`
	BlockedTasks    int                    `json:"blockedTasks"`
	OverdueTasks    int                    `json:"overdueTasks"`
	ByStatus        map[string]int         `json:"byStatus"`
	ByPriority      map[string]int         `json:"byPriority"`
	ByType          map[string]int         `json:"byType"`
	AvgCycleTime    *time.Duration         `json:"avgCycleTime,omitempty"`
	AvgLeadTime     *time.Duration         `json:"avgLeadTime,omitempty"`
	Throughput      float64                `json:"throughput"` // tasks per day
	Burndown        []BurndownPoint        `json:"burndown,omitempty"`
}

type BurndownPoint struct {
	Date      time.Time `json:"date"`
	Remaining int       `json:"remaining"`
	Completed int       `json:"completed"`
}

type TimeFrame string

const (
	TimeFrameDay     TimeFrame = "day"
	TimeFrameWeek    TimeFrame = "week"
	TimeFrameMonth   TimeFrame = "month"
	TimeFrameQuarter TimeFrame = "quarter"
	TimeFrameYear    TimeFrame = "year"
)

type ProductivityReport struct {
	TeamID        string                    `json:"teamId"`
	Timeframe     TimeFrame                 `json:"timeframe"`
	StartDate     time.Time                 `json:"startDate"`
	EndDate       time.Time                 `json:"endDate"`
	TeamMetrics   *TaskMetrics              `json:"teamMetrics"`
	MemberMetrics map[string]*TaskMetrics   `json:"memberMetrics"`
	Insights      []ProductivityInsight     `json:"insights"`
	Trends        *ProductivityTrends       `json:"trends"`
}

type ProductivityInsight struct {
	Type        string      `json:"type"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Severity    string      `json:"severity"` // info, warning, critical
	Value       interface{} `json:"value,omitempty"`
}

type ProductivityTrends struct {
	ThroughputTrend     float64 `json:"throughputTrend"`     // % change
	CycleTimeTrend      float64 `json:"cycleTimeTrend"`      // % change
	QualityTrend        float64 `json:"qualityTrend"`        // % change
	CollaborationTrend  float64 `json:"collaborationTrend"`  // % change
}

type ReportConfig struct {
	Type        ReportType             `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Filters     *MetricsFilters        `json:"filters,omitempty"`
	GroupBy     []string               `json:"groupBy,omitempty"`
	Metrics     []string               `json:"metrics"`
	Format      ReportFormat           `json:"format"`
	Schedule    *ReportSchedule        `json:"schedule,omitempty"`
	Recipients  []string               `json:"recipients,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
}

type ReportType string

const (
	ReportTypeTaskSummary     ReportType = "task_summary"
	ReportTypeProductivity    ReportType = "productivity"
	ReportTypeBurndown        ReportType = "burndown"
	ReportTypeVelocity        ReportType = "velocity"
	ReportTypeTimeTracking    ReportType = "time_tracking"
	ReportTypeCustom          ReportType = "custom"
)

type ReportFormat string

const (
	ReportFormatJSON ReportFormat = "json"
	ReportFormatCSV  ReportFormat = "csv"
	ReportFormatPDF  ReportFormat = "pdf"
	ReportFormatHTML ReportFormat = "html"
)

type ReportSchedule struct {
	Frequency ReportFrequency `json:"frequency"`
	DayOfWeek *int            `json:"dayOfWeek,omitempty"`   // 0-6, Sunday = 0
	DayOfMonth *int           `json:"dayOfMonth,omitempty"`  // 1-31
	Hour      int             `json:"hour"`                  // 0-23
	Timezone  string          `json:"timezone"`
}

type ReportFrequency string

const (
	ReportFrequencyDaily   ReportFrequency = "daily"
	ReportFrequencyWeekly  ReportFrequency = "weekly"
	ReportFrequencyMonthly ReportFrequency = "monthly"
)

type Report struct {
	ID          string                 `json:"id"`
	Config      *ReportConfig          `json:"config"`
	GeneratedAt time.Time              `json:"generatedAt"`
	Data        map[string]interface{} `json:"data"`
	URL         string                 `json:"url,omitempty"`
	FileSize    int64                  `json:"fileSize,omitempty"`
	ExpiresAt   *time.Time             `json:"expiresAt,omitempty"`
}

type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatXML  ExportFormat = "xml"
	ExportFormatExcel ExportFormat = "excel"
)

type ExportFilters struct {
	ProjectID     string     `json:"projectId,omitempty"`
	BoardID       string     `json:"boardId,omitempty"`
	Status        []string   `json:"status,omitempty"`
	Priority      []string   `json:"priority,omitempty"`
	AssigneeID    string     `json:"assigneeId,omitempty"`
	CreatedAfter  *time.Time `json:"createdAfter,omitempty"`
	CreatedBefore *time.Time `json:"createdBefore,omitempty"`
	IncludeFields []string   `json:"includeFields,omitempty"`
	Limit         int        `json:"limit,omitempty"`
}