package providers

import (
	"context"
	"errors"
	"time"
)

// TaskProvider defines the core interface for task management operations
type TaskProvider interface {
	// Core task operations
	CreateTask(ctx context.Context, task *UniversalTask) (*UniversalTask, error)
	GetTask(ctx context.Context, id string) (*UniversalTask, error)
	UpdateTask(ctx context.Context, id string, updates *TaskUpdate) error
	DeleteTask(ctx context.Context, id string) error
	ListTasks(ctx context.Context, filters *TaskFilters) ([]*UniversalTask, error)

	// Status operations
	UpdateStatus(ctx context.Context, taskID string, status TaskStatus) error
	GetAvailableStatuses(ctx context.Context, projectID string) ([]TaskStatus, error)

	// Bulk operations
	BulkCreateTasks(ctx context.Context, tasks []*UniversalTask) ([]*UniversalTask, error)
	BulkUpdateTasks(ctx context.Context, updates map[string]*TaskUpdate) error

	// Provider metadata
	GetProviderInfo() *ProviderInfo
	HealthCheck(ctx context.Context) error

	// Cleanup
	Close() error
}

// BoardProvider defines interface for board/project management operations
type BoardProvider interface {
	// Board operations
	GetBoard(ctx context.Context, id string) (*UniversalBoard, error)
	ListBoards(ctx context.Context, projectID string) ([]*UniversalBoard, error)
	CreateBoard(ctx context.Context, board *UniversalBoard) (*UniversalBoard, error)
	UpdateBoard(ctx context.Context, id string, updates *BoardUpdate) error
	DeleteBoard(ctx context.Context, id string) error

	// Column operations
	GetBoardColumns(ctx context.Context, boardID string) ([]*BoardColumn, error)
	MoveBetweenColumns(ctx context.Context, taskID, fromColumn, toColumn string) error

	// Board automation
	GetWorkflowRules(ctx context.Context, boardID string) ([]*WorkflowRule, error)
	CreateWorkflowRule(ctx context.Context, rule *WorkflowRule) error
}

// SyncProvider defines interface for real-time synchronization
type SyncProvider interface {
	// Synchronization
	StartRealTimeSync(ctx context.Context, callback SyncCallback) error
	StopRealTimeSync(ctx context.Context) error

	// Event handling
	SubscribeToEvents(ctx context.Context, events []EventType, callback EventCallback) error
	UnsubscribeFromEvents(ctx context.Context, events []EventType) error

	// Conflict resolution
	ResolveConflict(ctx context.Context, conflict *SyncConflict) (*ConflictResolution, error)
	GetConflicts(ctx context.Context, filters *ConflictFilters) ([]*SyncConflict, error)
}

// SearchProvider defines interface for advanced search capabilities
type SearchProvider interface {
	// Search operations
	SearchTasks(ctx context.Context, query *SearchQuery) ([]*UniversalTask, error)
	SearchBoards(ctx context.Context, query *SearchQuery) ([]*UniversalBoard, error)
	
	// Advanced search
	GetSearchSuggestions(ctx context.Context, partial string) ([]string, error)
	SavedSearches(ctx context.Context) ([]*SavedSearch, error)
}

// AnalyticsProvider defines interface for analytics and reporting
type AnalyticsProvider interface {
	// Task analytics
	GetTaskMetrics(ctx context.Context, filters *MetricsFilters) (*TaskMetrics, error)
	GetTeamProductivity(ctx context.Context, teamID string, timeframe TimeFrame) (*ProductivityReport, error)
	
	// Custom reports
	GenerateReport(ctx context.Context, config *ReportConfig) (*Report, error)
	ExportData(ctx context.Context, format ExportFormat, filters *ExportFilters) ([]byte, error)
}

// TaskManagerPlugin defines the plugin interface for dynamic loading
type TaskManagerPlugin interface {
	// Plugin metadata
	Name() string
	Version() string
	Description() string

	// Plugin lifecycle
	Initialize(config *ProviderConfig) error
	GetProvider() TaskProvider
	Cleanup() error

	// Optional interfaces
	GetBoardProvider() BoardProvider
	GetSyncProvider() SyncProvider
	GetSearchProvider() SearchProvider
	GetAnalyticsProvider() AnalyticsProvider
}

// Callback types for async operations
type SyncCallback func(event *SyncEvent) error
type EventCallback func(event *UniversalEvent) error

// Capability enum for provider features
type Capability string

const (
	CapabilityTasks             Capability = "tasks"
	CapabilityBoards            Capability = "boards"
	CapabilityRealTimeSync      Capability = "real_time_sync"
	CapabilityCustomFields      Capability = "custom_fields"
	CapabilityWorkflows         Capability = "workflows"
	CapabilityTimeTracking      Capability = "time_tracking"
	CapabilityHierarchicalTasks Capability = "hierarchical_tasks"
	CapabilityReporting         Capability = "reporting"
	CapabilityAdvancedSearch    Capability = "advanced_search"
	CapabilityWebhooks          Capability = "webhooks"
	CapabilityAPI               Capability = "api"
	CapabilityDocuments         Capability = "documents"
	CapabilityTemplates         Capability = "templates"
)

// ProviderInfo contains metadata about a provider
type ProviderInfo struct {
	Name            string                 `json:"name"`
	Type            ProviderType           `json:"type"`
	Version         string                 `json:"version"`
	Description     string                 `json:"description,omitempty"`
	Enabled         bool                   `json:"enabled"`
	Capabilities    []Capability           `json:"capabilities"`
	SupportedFeatures map[string]bool      `json:"supportedFeatures"`
	APILimits       *APILimits             `json:"apiLimits,omitempty"`
	HealthStatus    ProviderHealthStatus   `json:"healthStatus"`
	LastHealthCheck time.Time              `json:"lastHealthCheck"`
}

func (pi *ProviderInfo) HasCapability(capability Capability) bool {
	for _, c := range pi.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}

type APILimits struct {
	RequestsPerMinute int `json:"requestsPerMinute"`
	RequestsPerHour   int `json:"requestsPerHour"`
	RequestsPerDay    int `json:"requestsPerDay"`
	BurstSize         int `json:"burstSize"`
}

type ProviderHealthStatus string

const (
	HealthStatusHealthy   ProviderHealthStatus = "healthy"
	HealthStatusDegraded  ProviderHealthStatus = "degraded"
	HealthStatusUnhealthy ProviderHealthStatus = "unhealthy"
	HealthStatusUnknown   ProviderHealthStatus = "unknown"
)

// Authentication types
type AuthenticationType string

const (
	AuthTypeAPIKey    AuthenticationType = "api_key"
	AuthTypeBearer    AuthenticationType = "bearer"
	AuthTypeOAuth2    AuthenticationType = "oauth2"
	AuthTypeBasic     AuthenticationType = "basic"
	AuthTypeCustom    AuthenticationType = "custom"
)

// Provider types
type ProviderType string

const (
	ProviderTypeYouTrack ProviderType = "youtrack"
	ProviderTypeJira     ProviderType = "jira"
	ProviderTypeNotion   ProviderType = "notion"
	ProviderTypeLinear   ProviderType = "linear"
	ProviderTypeGitHub   ProviderType = "github"
	ProviderTypeGitLab   ProviderType = "gitlab"
	ProviderTypeAsana    ProviderType = "asana"
	ProviderTypeMonday   ProviderType = "monday"
	ProviderTypeClickUp  ProviderType = "clickup"
	ProviderTypeTrello   ProviderType = "trello"
	ProviderTypeAzure    ProviderType = "azure_devops"
	ProviderTypeCustom   ProviderType = "custom"
)

// Error types
type ErrorType string

const (
	ErrorTypeValidation     ErrorType = "validation"
	ErrorTypeNotFound       ErrorType = "not_found"
	ErrorTypeUnauthorized   ErrorType = "unauthorized"
	ErrorTypeForbidden      ErrorType = "forbidden"
	ErrorTypeRateLimit      ErrorType = "rate_limit"
	ErrorTypeNetwork        ErrorType = "network"
	ErrorTypeInternal       ErrorType = "internal"
	ErrorTypeConfiguration ErrorType = "configuration"
)

// Common errors
var (
	ErrTaskNotFound       = NewProviderError(ErrorTypeNotFound, "task not found", nil)
	ErrBoardNotFound      = NewProviderError(ErrorTypeNotFound, "board not found", nil)
	ErrProjectNotFound    = NewProviderError(ErrorTypeNotFound, "project not found", nil)
	ErrUnauthorized       = NewProviderError(ErrorTypeUnauthorized, "unauthorized", nil)
	ErrForbidden         = NewProviderError(ErrorTypeForbidden, "forbidden", nil)
	ErrRateLimited       = NewProviderError(ErrorTypeRateLimit, "rate limited", nil)
	ErrInvalidConfig     = NewProviderError(ErrorTypeConfiguration, "invalid configuration", nil)
)

// ProviderError represents a provider-specific error
type ProviderError struct {
	Type    ErrorType              `json:"type"`
	Message string                 `json:"message"`
	Context map[string]interface{} `json:"context,omitempty"`
	Cause   error                  `json:"-"`
}

func (e *ProviderError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *ProviderError) Unwrap() error {
	return e.Cause
}

func NewProviderError(errorType ErrorType, message string, cause error) *ProviderError {
	return &ProviderError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// IsErrorType checks if an error is of a specific type
func IsErrorType(err error, errorType ErrorType) bool {
	var providerErr *ProviderError
	if errors.As(err, &providerErr) {
		return providerErr.Type == errorType
	}
	return false
}

// IsNotFoundError checks if an error is a "not found" error
func IsNotFoundError(err error) bool {
	return IsErrorType(err, ErrorTypeNotFound)
}

// IsUnauthorizedError checks if an error is an "unauthorized" error
func IsUnauthorizedError(err error) bool {
	return IsErrorType(err, ErrorTypeUnauthorized)
}

// IsRateLimitError checks if an error is a "rate limit" error
func IsRateLimitError(err error) bool {
	return IsErrorType(err, ErrorTypeRateLimit)
}

// NewValidationError creates a new validation error
func NewValidationError(message string, context map[string]interface{}) *ProviderError {
	return &ProviderError{
		Type:    ErrorTypeValidation,
		Message: message,
		Context: context,
	}
}