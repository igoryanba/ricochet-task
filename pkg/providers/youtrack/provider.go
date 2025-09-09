package youtrack

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/grik-ai/ricochet-task/pkg/providers"
)

// YouTrackProvider implements TaskProvider interface for YouTrack
type YouTrackProvider struct {
	client     *YouTrackClient
	config     *providers.ProviderConfig
	translator *YouTrackTranslator
	logger     *logrus.Entry
}

// NewYouTrackProvider creates a new YouTrack provider
func NewYouTrackProvider(config *providers.ProviderConfig) (*YouTrackProvider, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	client, err := NewYouTrackClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	logger := logrus.WithFields(logrus.Fields{
		"provider": "youtrack",
		"instance": config.Name,
	})

	return &YouTrackProvider{
		client:     client,
		config:     config,
		translator: NewYouTrackTranslator(),
		logger:     logger,
	}, nil
}

// CreateTask creates a new task in YouTrack
func (p *YouTrackProvider) CreateTask(ctx context.Context, task *providers.UniversalTask) (*providers.UniversalTask, error) {
	p.logger.WithField("task_title", task.Title).Debug("Creating task in YouTrack")

	// Validate task
	if err := p.validateTask(task); err != nil {
		return nil, fmt.Errorf("task validation failed: %w", err)
	}

	// Convert to YouTrack format
	ytIssue := p.translator.UniversalToYouTrack(task)

	// Create in YouTrack
	createdIssue, err := p.client.CreateIssue(ctx, ytIssue)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue in YouTrack: %w", err)
	}

	// Convert back to universal format
	universalTask := p.translator.YouTrackToUniversal(createdIssue)

	// Add Ricochet metadata
	universalTask.RicochetMetadata = &providers.RicochetTaskMetadata{
		LastSyncTime: time.Now(),
		SyncStatus:   providers.SyncStatusSynced,
	}

	// Set provider information
	universalTask.ProviderName = p.config.Name
	universalTask.ProviderConfig = p.config

	p.logger.WithFields(logrus.Fields{
		"task_id": universalTask.ID,
		"youtrack_id": universalTask.ExternalID,
	}).Info("Task created successfully in YouTrack")

	return universalTask, nil
}

// GetTask retrieves a task from YouTrack
func (p *YouTrackProvider) GetTask(ctx context.Context, id string) (*providers.UniversalTask, error) {
	p.logger.WithField("task_id", id).Debug("Getting task from YouTrack")

	ytIssue, err := p.client.GetIssue(ctx, id)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, providers.ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to get issue from YouTrack: %w", err)
	}

	universalTask := p.translator.YouTrackToUniversal(ytIssue)
	universalTask.ProviderName = p.config.Name
	universalTask.ProviderConfig = p.config

	return universalTask, nil
}

// UpdateTask updates a task in YouTrack
func (p *YouTrackProvider) UpdateTask(ctx context.Context, id string, updates *providers.TaskUpdate) error {
	p.logger.WithField("task_id", id).Debug("Updating task in YouTrack")

	// Convert updates to YouTrack format
	ytUpdates := p.translator.UniversalUpdatesToYouTrack(updates)

	err := p.client.UpdateIssue(ctx, id, ytUpdates)
	if err != nil {
		if IsNotFoundError(err) {
			return providers.ErrTaskNotFound
		}
		return fmt.Errorf("failed to update issue in YouTrack: %w", err)
	}

	p.logger.WithField("task_id", id).Info("Task updated successfully in YouTrack")
	return nil
}

// DeleteTask deletes a task from YouTrack
func (p *YouTrackProvider) DeleteTask(ctx context.Context, id string) error {
	p.logger.WithField("task_id", id).Debug("Deleting task from YouTrack")

	err := p.client.DeleteIssue(ctx, id)
	if err != nil {
		if IsNotFoundError(err) {
			return providers.ErrTaskNotFound
		}
		return fmt.Errorf("failed to delete issue from YouTrack: %w", err)
	}

	p.logger.WithField("task_id", id).Info("Task deleted successfully from YouTrack")
	return nil
}

// ListTasks lists tasks from YouTrack with filters
func (p *YouTrackProvider) ListTasks(ctx context.Context, filters *providers.TaskFilters) ([]*providers.UniversalTask, error) {
	p.logger.WithField("filters", filters).Debug("Listing tasks from YouTrack")

	// Convert filters to YouTrack format
	ytFilters := p.translator.UniversalFiltersToYouTrack(filters)

	ytIssues, err := p.client.ListIssues(ctx, ytFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to list issues from YouTrack: %w", err)
	}

	// Convert all issues to universal format
	universalTasks := make([]*providers.UniversalTask, len(ytIssues))
	for i, issue := range ytIssues {
		universalTasks[i] = p.translator.YouTrackToUniversal(issue)
		universalTasks[i].ProviderName = p.config.Name
		universalTasks[i].ProviderConfig = p.config
	}

	p.logger.WithField("count", len(universalTasks)).Info("Tasks listed successfully from YouTrack")
	return universalTasks, nil
}

// UpdateStatus updates the status of a task
func (p *YouTrackProvider) UpdateStatus(ctx context.Context, taskID string, status providers.TaskStatus) error {
	p.logger.WithFields(logrus.Fields{
		"task_id": taskID,
		"status": status.Name,
	}).Debug("Updating task status in YouTrack")

	// Convert status to YouTrack format
	ytStatus := p.translator.UniversalStatusToYouTrack(status)

	err := p.client.UpdateIssueStatus(ctx, taskID, ytStatus)
	if err != nil {
		if IsNotFoundError(err) {
			return providers.ErrTaskNotFound
		}
		return fmt.Errorf("failed to update status in YouTrack: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"task_id": taskID,
		"status": status.Name,
	}).Info("Task status updated successfully in YouTrack")

	return nil
}

// GetAvailableStatuses returns available statuses for a project
func (p *YouTrackProvider) GetAvailableStatuses(ctx context.Context, projectID string) ([]providers.TaskStatus, error) {
	p.logger.WithField("project_id", projectID).Debug("Getting available statuses from YouTrack")

	ytStatuses, err := p.client.GetProjectStatuses(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get statuses from YouTrack: %w", err)
	}

	// Convert to universal format
	universalStatuses := make([]providers.TaskStatus, len(ytStatuses))
	for i, status := range ytStatuses {
		universalStatuses[i] = p.translator.YouTrackStatusToUniversal(status)
	}

	return universalStatuses, nil
}

// BulkCreateTasks creates multiple tasks in YouTrack
func (p *YouTrackProvider) BulkCreateTasks(ctx context.Context, tasks []*providers.UniversalTask) ([]*providers.UniversalTask, error) {
	p.logger.WithField("count", len(tasks)).Debug("Bulk creating tasks in YouTrack")

	if len(tasks) == 0 {
		return []*providers.UniversalTask{}, nil
	}

	// Convert to YouTrack format
	ytIssues := make([]*YouTrackIssue, len(tasks))
	for i, task := range tasks {
		if err := p.validateTask(task); err != nil {
			return nil, fmt.Errorf("task %d validation failed: %w", i, err)
		}
		ytIssues[i] = p.translator.UniversalToYouTrack(task)
	}

	// Create in YouTrack (batch operation if supported)
	createdIssues, err := p.client.BulkCreateIssues(ctx, ytIssues)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk create issues in YouTrack: %w", err)
	}

	// Convert back to universal format
	universalTasks := make([]*providers.UniversalTask, len(createdIssues))
	for i, issue := range createdIssues {
		universalTasks[i] = p.translator.YouTrackToUniversal(issue)
		universalTasks[i].RicochetMetadata = &providers.RicochetTaskMetadata{
			LastSyncTime: time.Now(),
			SyncStatus:   providers.SyncStatusSynced,
		}
		universalTasks[i].ProviderName = p.config.Name
		universalTasks[i].ProviderConfig = p.config
	}

	p.logger.WithField("count", len(universalTasks)).Info("Tasks bulk created successfully in YouTrack")
	return universalTasks, nil
}

// BulkUpdateTasks updates multiple tasks in YouTrack
func (p *YouTrackProvider) BulkUpdateTasks(ctx context.Context, updates map[string]*providers.TaskUpdate) error {
	p.logger.WithField("count", len(updates)).Debug("Bulk updating tasks in YouTrack")

	if len(updates) == 0 {
		return nil
	}

	// Convert updates to YouTrack format
	ytUpdates := make(map[string]*YouTrackIssueUpdate)
	for id, update := range updates {
		ytUpdates[id] = p.translator.UniversalUpdatesToYouTrack(update)
	}

	err := p.client.BulkUpdateIssues(ctx, ytUpdates)
	if err != nil {
		return fmt.Errorf("failed to bulk update issues in YouTrack: %w", err)
	}

	p.logger.WithField("count", len(updates)).Info("Tasks bulk updated successfully in YouTrack")
	return nil
}

// GetProviderInfo returns information about this provider
func (p *YouTrackProvider) GetProviderInfo() *providers.ProviderInfo {
	return &providers.ProviderInfo{
		Name:        "YouTrack",
		Version:     "1.0.0",
		Description: "JetBrains YouTrack integration for ricochet-task",
		Capabilities: []providers.Capability{
			providers.CapabilityTasks,
			providers.CapabilityBoards,
			providers.CapabilityRealTimeSync,
			providers.CapabilityCustomFields,
			providers.CapabilityWorkflows,
			providers.CapabilityTimeTracking,
			providers.CapabilityHierarchicalTasks,
			providers.CapabilityReporting,
			providers.CapabilityAdvancedSearch,
			providers.CapabilityWebhooks,
		},
		SupportedFeatures: map[string]bool{
			"hierarchical_tasks": true,
			"custom_fields":      true,
			"time_tracking":      true,
			"agile_boards":       true,
			"workflows":          true,
			"webhooks":           true,
			"search_queries":     true,
			"bulk_operations":    true,
		},
		APILimits: &providers.APILimits{
			RequestsPerMinute: 600,
			RequestsPerHour:   36000,
			BurstSize:         100,
		},
		HealthStatus:    providers.HealthStatusHealthy,
		LastHealthCheck: time.Now(),
	}
}

// HealthCheck performs a health check on the YouTrack connection
func (p *YouTrackProvider) HealthCheck(ctx context.Context) error {
	p.logger.Debug("Performing YouTrack health check")

	// Simple health check - get server info
	err := p.client.HealthCheck(ctx)
	if err != nil {
		p.logger.WithError(err).Warn("YouTrack health check failed")
		return fmt.Errorf("YouTrack health check failed: %w", err)
	}

	p.logger.Debug("YouTrack health check passed")
	return nil
}

// Close closes the provider and cleans up resources
func (p *YouTrackProvider) Close() error {
	p.logger.Info("Closing YouTrack provider")

	if p.client != nil {
		return p.client.Close()
	}

	return nil
}

// validateTask validates a universal task before creating/updating
func (p *YouTrackProvider) validateTask(task *providers.UniversalTask) error {
	if task == nil {
		return providers.NewProviderError(providers.ErrorTypeValidation, "task cannot be nil", nil)
	}

	if task.Title == "" {
		return providers.NewProviderError(providers.ErrorTypeValidation, "task title is required", nil)
	}

	if task.ProjectID == "" {
		return providers.NewProviderError(providers.ErrorTypeValidation, "project ID is required", nil)
	}

	// Validate YouTrack-specific constraints
	if len(task.Title) > 255 {
		return providers.NewProviderError(providers.ErrorTypeValidation, "task title too long (max 255 characters)", nil)
	}

	return nil
}

// Helper methods for specific YouTrack operations
func (p *YouTrackProvider) GetIssueByKey(ctx context.Context, key string) (*providers.UniversalTask, error) {
	ytIssue, err := p.client.GetIssueByKey(ctx, key)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, providers.ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to get issue by key from YouTrack: %w", err)
	}

	universalTask := p.translator.YouTrackToUniversal(ytIssue)
	universalTask.ProviderName = p.config.Name
	universalTask.ProviderConfig = p.config

	return universalTask, nil
}

func (p *YouTrackProvider) AddComment(ctx context.Context, taskID string, comment string) error {
	err := p.client.AddComment(ctx, taskID, &YouTrackComment{
		Text: comment,
	})
	if err != nil {
		if IsNotFoundError(err) {
			return providers.ErrTaskNotFound
		}
		return fmt.Errorf("failed to add comment in YouTrack: %w", err)
	}

	return nil
}

func (p *YouTrackProvider) GetComments(ctx context.Context, taskID string) ([]*providers.Comment, error) {
	ytComments, err := p.client.GetComments(ctx, taskID)
	if err != nil {
		if IsNotFoundError(err) {
			return nil, providers.ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to get comments from YouTrack: %w", err)
	}

	comments := make([]*providers.Comment, len(ytComments))
	for i, ytComment := range ytComments {
		comments[i] = p.translator.YouTrackCommentToUniversal(ytComment)
	}

	return comments, nil
}

// IsNotFoundError checks if an error is a "not found" error from YouTrack
func IsNotFoundError(err error) bool {
	if ytErr, ok := err.(*YouTrackError); ok {
		return ytErr.StatusCode == 404
	}
	return false
}

// IsRateLimitError checks if an error is a rate limit error from YouTrack
func IsRateLimitError(err error) bool {
	if ytErr, ok := err.(*YouTrackError); ok {
		return ytErr.StatusCode == 429
	}
	return false
}

// IsUnauthorizedError checks if an error is an unauthorized error from YouTrack
func IsUnauthorizedError(err error) bool {
	if ytErr, ok := err.(*YouTrackError); ok {
		return ytErr.StatusCode == 401 || ytErr.StatusCode == 403
	}
	return false
}

// SearchTasks searches for tasks with a query string
func (p *YouTrackProvider) SearchTasks(ctx context.Context, query string, filters *providers.TaskFilters) ([]*providers.UniversalTask, error) {
	if query == "" {
		return nil, providers.NewValidationError("search query cannot be empty", nil)
	}

	// For YouTrack, we'll combine the query with filters
	ytFilters := &YouTrackIssueFilters{
		Query: query,
	}

	if filters != nil {
		if filters.ProjectID != "" {
			ytFilters.ProjectID = filters.ProjectID
		}
		if len(filters.Status) > 0 {
			ytFilters.State = filters.Status[0] // Take first status for simplicity
		}
		if filters.AssigneeID != "" {
			ytFilters.Assignee = filters.AssigneeID
		}
		if filters.Limit > 0 {
			ytFilters.Top = filters.Limit
		}
		if filters.Offset > 0 {
			ytFilters.Skip = filters.Offset
		}
	}

	issues, err := p.client.ListIssues(ctx, ytFilters)
	if err != nil {
		return nil, err
	}

	tasks := make([]*providers.UniversalTask, len(issues))
	for i, issue := range issues {
		task, convertErr := p.convertToUniversalTask(issue)
		if convertErr != nil {
			return nil, convertErr
		}
		tasks[i] = task
	}

	return tasks, nil
}

// convertToUniversalTask converts a YouTrack issue to a Universal task
func (p *YouTrackProvider) convertToUniversalTask(issue *YouTrackIssue) (*providers.UniversalTask, error) {
	return p.translator.YouTrackToUniversal(issue), nil
}