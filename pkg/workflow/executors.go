package workflow

import (
	"context"
	"fmt"
	"time"
)

// TaskActionExecutor исполнитель действий с задачами
type TaskActionExecutor struct {
	providerRegistry interface{} // TODO: заменить на реальный тип
}

func (e *TaskActionExecutor) GetType() string {
	return "task"
}

func (e *TaskActionExecutor) Execute(ctx context.Context, action *ActionDefinition, context map[string]interface{}) (map[string]interface{}, error) {
	subAction := e.getStringParam(action.Parameters, "action", "")
	
	switch subAction {
	case "create":
		return e.createTask(ctx, action, context)
	case "update":
		return e.updateTask(ctx, action, context)
	case "assign":
		return e.assignTask(ctx, action, context)
	case "close":
		return e.closeTask(ctx, action, context)
	default:
		return nil, fmt.Errorf("unsupported task action: %s", subAction)
	}
}

func (e *TaskActionExecutor) createTask(ctx context.Context, action *ActionDefinition, context map[string]interface{}) (map[string]interface{}, error) {
	title := e.getStringParam(action.Parameters, "title", "")
	description := e.getStringParam(action.Parameters, "description", "")
	priority := e.getStringParam(action.Parameters, "priority", "medium")
	assignee := e.getStringParam(action.Parameters, "assignee", "")
	
	if title == "" {
		return nil, fmt.Errorf("task title is required")
	}
	
	// Заглушка для создания задачи
	// В реальной реализации здесь был бы вызов провайдера
	taskID := fmt.Sprintf("TASK-%d", time.Now().Unix())
	
	result := map[string]interface{}{
		"task_id":     taskID,
		"title":       title,
		"description": description,
		"priority":    priority,
		"assignee":    assignee,
		"status":      "created",
		"created_at":  time.Now(),
	}
	
	return result, nil
}

func (e *TaskActionExecutor) updateTask(ctx context.Context, action *ActionDefinition, context map[string]interface{}) (map[string]interface{}, error) {
	taskID := e.getStringParam(action.Parameters, "task_id", "")
	newStatus := e.getStringParam(action.Parameters, "status", "")
	comment := e.getStringParam(action.Parameters, "comment", "")
	
	if taskID == "" {
		return nil, fmt.Errorf("task_id is required for update")
	}
	
	// Заглушка для обновления задачи
	result := map[string]interface{}{
		"task_id":    taskID,
		"status":     newStatus,
		"comment":    comment,
		"updated_at": time.Now(),
		"action":     "updated",
	}
	
	return result, nil
}

func (e *TaskActionExecutor) assignTask(ctx context.Context, action *ActionDefinition, context map[string]interface{}) (map[string]interface{}, error) {
	taskID := e.getStringParam(action.Parameters, "task_id", "")
	assignee := e.getStringParam(action.Parameters, "assignee", "")
	
	if taskID == "" || assignee == "" {
		return nil, fmt.Errorf("task_id and assignee are required for assignment")
	}
	
	// Заглушка для назначения задачи
	result := map[string]interface{}{
		"task_id":     taskID,
		"assignee":    assignee,
		"assigned_at": time.Now(),
		"action":      "assigned",
	}
	
	return result, nil
}

func (e *TaskActionExecutor) closeTask(ctx context.Context, action *ActionDefinition, context map[string]interface{}) (map[string]interface{}, error) {
	taskID := e.getStringParam(action.Parameters, "task_id", "")
	resolution := e.getStringParam(action.Parameters, "resolution", "completed")
	
	if taskID == "" {
		return nil, fmt.Errorf("task_id is required for closing")
	}
	
	// Заглушка для закрытия задачи
	result := map[string]interface{}{
		"task_id":    taskID,
		"status":     "closed",
		"resolution": resolution,
		"closed_at":  time.Now(),
		"action":     "closed",
	}
	
	return result, nil
}

func (e *TaskActionExecutor) getStringParam(params map[string]interface{}, key, defaultValue string) string {
	if value, exists := params[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// NotificationActionExecutor исполнитель уведомлений
type NotificationActionExecutor struct {
	notificationService *NotificationService
}

func (e *NotificationActionExecutor) GetType() string {
	return "notification"
}

func (e *NotificationActionExecutor) Execute(ctx context.Context, action *ActionDefinition, context map[string]interface{}) (map[string]interface{}, error) {
	notificationType := e.getStringParam(action.Parameters, "type", "info")
	title := e.getStringParam(action.Parameters, "title", "Workflow Notification")
	message := e.getStringParam(action.Parameters, "message", "")
	recipients := e.getSliceParam(action.Parameters, "recipients", []string{})
	channel := e.getStringParam(action.Parameters, "channel", "default")
	
	if message == "" {
		return nil, fmt.Errorf("notification message is required")
	}
	
	notification := &Notification{
		ID:         fmt.Sprintf("notif-%d", time.Now().Unix()),
		Type:       notificationType,
		Title:      title,
		Message:    message,
		Recipients: recipients,
		Timestamp:  time.Now(),
		Data:       action.Parameters,
	}
	
	// Заглушка для отправки уведомления
	// В реальной реализации здесь был бы вызов notification service
	
	result := map[string]interface{}{
		"notification_id": notification.ID,
		"type":            notificationType,
		"title":           title,
		"message":         message,
		"recipients":      recipients,
		"channel":         channel,
		"sent_at":         time.Now(),
		"status":          "sent",
	}
	
	return result, nil
}

func (e *NotificationActionExecutor) getStringParam(params map[string]interface{}, key, defaultValue string) string {
	if value, exists := params[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

func (e *NotificationActionExecutor) getSliceParam(params map[string]interface{}, key string, defaultValue []string) []string {
	if value, exists := params[key]; exists {
		if slice, ok := value.([]interface{}); ok {
			result := make([]string, len(slice))
			for i, v := range slice {
				result[i] = fmt.Sprintf("%v", v)
			}
			return result
		}
		if slice, ok := value.([]string); ok {
			return slice
		}
	}
	return defaultValue
}

// StatusActionExecutor исполнитель изменения статусов
type StatusActionExecutor struct{}

func (e *StatusActionExecutor) GetType() string {
	return "status"
}

func (e *StatusActionExecutor) Execute(ctx context.Context, action *ActionDefinition, context map[string]interface{}) (map[string]interface{}, error) {
	taskID := e.getStringParam(action.Parameters, "task_id", "")
	newStatus := e.getStringParam(action.Parameters, "new_status", "")
	reason := e.getStringParam(action.Parameters, "reason", "Automatic workflow transition")
	
	if taskID == "" {
		return nil, fmt.Errorf("task_id is required for status change")
	}
	
	if newStatus == "" {
		return nil, fmt.Errorf("new_status is required")
	}
	
	// Заглушка для изменения статуса
	result := map[string]interface{}{
		"task_id":        taskID,
		"old_status":     e.getStringParam(action.Parameters, "old_status", "unknown"),
		"new_status":     newStatus,
		"reason":         reason,
		"changed_at":     time.Now(),
		"action":         "status_changed",
	}
	
	return result, nil
}

func (e *StatusActionExecutor) getStringParam(params map[string]interface{}, key, defaultValue string) string {
	if value, exists := params[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GitActionExecutor исполнитель Git действий
type GitActionExecutor struct{}

func (e *GitActionExecutor) GetType() string {
	return "git"
}

func (e *GitActionExecutor) Execute(ctx context.Context, action *ActionDefinition, context map[string]interface{}) (map[string]interface{}, error) {
	gitAction := e.getStringParam(action.Parameters, "action", "")
	_ = e.getStringParam(action.Parameters, "repository", "")
	
	switch gitAction {
	case "create_branch":
		return e.createBranch(ctx, action, context)
	case "create_pr":
		return e.createPullRequest(ctx, action, context)
	case "merge":
		return e.mergeBranch(ctx, action, context)
	default:
		return nil, fmt.Errorf("unsupported git action: %s", gitAction)
	}
}

func (e *GitActionExecutor) createBranch(ctx context.Context, action *ActionDefinition, context map[string]interface{}) (map[string]interface{}, error) {
	branchName := e.getStringParam(action.Parameters, "branch_name", "")
	repository := e.getStringParam(action.Parameters, "repository", "")
	
	if branchName == "" {
		return nil, fmt.Errorf("branch_name is required")
	}
	
	// Заглушка для создания ветки
	result := map[string]interface{}{
		"repository":  repository,
		"branch_name": branchName,
		"created_at":  time.Now(),
		"action":      "branch_created",
	}
	
	return result, nil
}

func (e *GitActionExecutor) createPullRequest(ctx context.Context, action *ActionDefinition, context map[string]interface{}) (map[string]interface{}, error) {
	title := e.getStringParam(action.Parameters, "title", "")
	sourceBranch := e.getStringParam(action.Parameters, "source_branch", "")
	targetBranch := e.getStringParam(action.Parameters, "target_branch", "main")
	
	if title == "" || sourceBranch == "" {
		return nil, fmt.Errorf("title and source_branch are required for PR creation")
	}
	
	// Заглушка для создания PR
	result := map[string]interface{}{
		"pr_id":         fmt.Sprintf("PR-%d", time.Now().Unix()),
		"title":         title,
		"source_branch": sourceBranch,
		"target_branch": targetBranch,
		"created_at":    time.Now(),
		"action":        "pr_created",
	}
	
	return result, nil
}

func (e *GitActionExecutor) mergeBranch(ctx context.Context, action *ActionDefinition, context map[string]interface{}) (map[string]interface{}, error) {
	sourceBranch := e.getStringParam(action.Parameters, "source_branch", "")
	targetBranch := e.getStringParam(action.Parameters, "target_branch", "main")
	
	if sourceBranch == "" {
		return nil, fmt.Errorf("source_branch is required for merge")
	}
	
	// Заглушка для merge
	result := map[string]interface{}{
		"source_branch": sourceBranch,
		"target_branch": targetBranch,
		"merged_at":     time.Now(),
		"action":        "branch_merged",
	}
	
	return result, nil
}

func (e *GitActionExecutor) getStringParam(params map[string]interface{}, key, defaultValue string) string {
	if value, exists := params[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// EmailActionExecutor исполнитель email уведомлений
type EmailActionExecutor struct{}

func (e *EmailActionExecutor) GetType() string {
	return "email"
}

func (e *EmailActionExecutor) Execute(ctx context.Context, action *ActionDefinition, context map[string]interface{}) (map[string]interface{}, error) {
	to := e.getSliceParam(action.Parameters, "to", []string{})
	subject := e.getStringParam(action.Parameters, "subject", "Workflow Notification")
	body := e.getStringParam(action.Parameters, "body", "")
	template := e.getStringParam(action.Parameters, "template", "")
	
	if len(to) == 0 {
		return nil, fmt.Errorf("email recipients are required")
	}
	
	if body == "" && template == "" {
		return nil, fmt.Errorf("either body or template is required")
	}
	
	// Заглушка для отправки email
	result := map[string]interface{}{
		"email_id":  fmt.Sprintf("email-%d", time.Now().Unix()),
		"to":        to,
		"subject":   subject,
		"body":      body,
		"template":  template,
		"sent_at":   time.Now(),
		"status":    "sent",
		"action":    "email_sent",
	}
	
	return result, nil
}

func (e *EmailActionExecutor) getStringParam(params map[string]interface{}, key, defaultValue string) string {
	if value, exists := params[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

func (e *EmailActionExecutor) getSliceParam(params map[string]interface{}, key string, defaultValue []string) []string {
	if value, exists := params[key]; exists {
		if slice, ok := value.([]interface{}); ok {
			result := make([]string, len(slice))
			for i, v := range slice {
				result[i] = fmt.Sprintf("%v", v)
			}
			return result
		}
		if slice, ok := value.([]string); ok {
			return slice
		}
	}
	return defaultValue
}