package workflow

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// GitProgressTracker отслеживает прогресс задач через Git активность
type GitProgressTracker struct {
	webhookHandler *GitWebhookHandler
	progressEngine *ProgressEngine
	ruleEngine     *RuleEngine
	logger         Logger
	patterns       *CommitPatterns
}

// NewGitProgressTracker создает новый tracker прогресса
func NewGitProgressTracker(ruleEngine *RuleEngine, logger Logger) *GitProgressTracker {
	tracker := &GitProgressTracker{
		ruleEngine: ruleEngine,
		logger:     logger,
		patterns:   NewCommitPatterns(),
	}
	
	tracker.webhookHandler = NewGitWebhookHandler(tracker, logger)
	tracker.progressEngine = NewProgressEngine(tracker, logger)
	
	return tracker
}

// GitProgressEvent событие от Git системы для прогресса
type GitProgressEvent struct {
	Type       string                 `json:"type"`
	Repository string                 `json:"repository"`
	Branch     string                 `json:"branch"`
	Commit     *CommitInfo            `json:"commit,omitempty"`
	PullRequest *PullRequestInfo      `json:"pull_request,omitempty"`
	Data       map[string]interface{} `json:"data"`
	Timestamp  time.Time              `json:"timestamp"`
}

func (ge *GitProgressEvent) GetType() string { return ge.Type }
func (ge *GitProgressEvent) GetData() map[string]interface{} { return ge.Data }

// CommitInfo информация о коммите
type CommitInfo struct {
	SHA       string    `json:"sha"`
	Message   string    `json:"message"`
	Author    string    `json:"author"`
	Timestamp time.Time `json:"timestamp"`
	FilesChanged []string `json:"files_changed"`
	LinesAdded   int     `json:"lines_added"`
	LinesDeleted int     `json:"lines_deleted"`
}

// PullRequestInfo информация о Pull Request
type PullRequestInfo struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	State       string    `json:"state"`
	Author      string    `json:"author"`
	SourceBranch string   `json:"source_branch"`
	TargetBranch string   `json:"target_branch"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TaskProgress прогресс задачи
type TaskProgress struct {
	TaskID          string           `json:"task_id"`
	CurrentStage    string           `json:"current_stage"`
	CompletedStages []string         `json:"completed_stages"`
	ProgressPercent float64          `json:"progress_percent"`
	LastActivity    time.Time        `json:"last_activity"`
	GitActivity     []GitActivity    `json:"git_activity"`
	Metrics         *ProgressMetrics `json:"metrics"`
}

// GitActivity активность в Git для задачи
type GitActivity struct {
	Type        string    `json:"type"`
	SHA         string    `json:"sha,omitempty"`
	Message     string    `json:"message"`
	Author      string    `json:"author"`
	Timestamp   time.Time `json:"timestamp"`
	FilesCount  int       `json:"files_count"`
	LinesAdded  int       `json:"lines_added"`
	LinesDeleted int      `json:"lines_deleted"`
}

// ProgressMetrics метрики прогресса
type ProgressMetrics struct {
	TotalCommits     int     `json:"total_commits"`
	TotalLinesAdded  int     `json:"total_lines_added"`
	TotalLinesDeleted int    `json:"total_lines_deleted"`
	FilesModified    int     `json:"files_modified"`
	AverageCommitSize float64 `json:"average_commit_size"`
	DevelopmentVelocity float64 `json:"development_velocity"`
	LastCommitDate   time.Time `json:"last_commit_date"`
}

// ProcessGitEvent обрабатывает Git событие
func (gpt *GitProgressTracker) ProcessGitEvent(ctx context.Context, event *GitProgressEvent) error {
	gpt.logger.Info("Processing Git event", 
		"type", event.Type, 
		"repository", event.Repository,
		"branch", event.Branch)
	
	// Извлекаем задачи из Git события
	tasks := gpt.extractTasksFromEvent(event)
	
	for _, taskID := range tasks {
		// Обновляем прогресс задачи
		if err := gpt.updateTaskProgress(ctx, taskID, event); err != nil {
			gpt.logger.Error("Failed to update task progress", err, "task", taskID)
			continue
		}
		
		// Проверяем автоматические переходы
		if err := gpt.checkAutoTransitions(ctx, taskID, event); err != nil {
			gpt.logger.Error("Failed to check auto transitions", err, "task", taskID)
		}
	}
	
	return nil
}

// extractTasksFromEvent извлекает ID задач из Git события
func (gpt *GitProgressTracker) extractTasksFromEvent(event *GitProgressEvent) []string {
	var tasks []string
	
	switch event.Type {
	case "push":
		if event.Commit != nil {
			tasks = gpt.patterns.ExtractTaskIDs(event.Commit.Message)
			
			// Дополнительно ищем в именах файлов
			for _, file := range event.Commit.FilesChanged {
				fileTasks := gpt.patterns.ExtractTaskIDsFromPath(file)
				tasks = append(tasks, fileTasks...)
			}
		}
		
	case "pull_request":
		if event.PullRequest != nil {
			// Ищем в заголовке PR
			prTasks := gpt.patterns.ExtractTaskIDs(event.PullRequest.Title)
			tasks = append(tasks, prTasks...)
			
			// Ищем в названии ветки
			branchTasks := gpt.patterns.ExtractTaskIDsFromBranch(event.PullRequest.SourceBranch)
			tasks = append(tasks, branchTasks...)
		}
	}
	
	// Убираем дубликаты
	return gpt.removeDuplicates(tasks)
}

// updateTaskProgress обновляет прогресс задачи
func (gpt *GitProgressTracker) updateTaskProgress(ctx context.Context, taskID string, event *GitProgressEvent) error {
	// Получаем текущий прогресс
	progress := gpt.progressEngine.GetTaskProgress(taskID)
	if progress == nil {
		progress = &TaskProgress{
			TaskID:          taskID,
			CurrentStage:    "development",
			CompletedStages: []string{},
			ProgressPercent: 0,
			LastActivity:    time.Now(),
			GitActivity:     []GitActivity{},
			Metrics:         &ProgressMetrics{},
		}
	}
	
	// Добавляем новую Git активность
	activity := gpt.createGitActivity(event)
	progress.GitActivity = append(progress.GitActivity, activity)
	progress.LastActivity = time.Now()
	
	// Обновляем метрики
	gpt.updateProgressMetrics(progress, activity)
	
	// Вычисляем новый процент прогресса
	newPercent := gpt.calculateProgressPercent(progress, event)
	if newPercent > progress.ProgressPercent {
		progress.ProgressPercent = newPercent
	}
	
	// Проверяем переходы между этапами
	newStage := gpt.determineStageFromActivity(progress, event)
	if newStage != "" && newStage != progress.CurrentStage {
		if !gpt.contains(progress.CompletedStages, progress.CurrentStage) {
			progress.CompletedStages = append(progress.CompletedStages, progress.CurrentStage)
		}
		progress.CurrentStage = newStage
		
		gpt.logger.Info("Task stage transition", 
			"task", taskID, 
			"new_stage", newStage,
			"progress", progress.ProgressPercent)
	}
	
	// Сохраняем прогресс
	return gpt.progressEngine.SaveTaskProgress(progress)
}

// createGitActivity создает запись Git активности
func (gpt *GitProgressTracker) createGitActivity(event *GitProgressEvent) GitActivity {
	activity := GitActivity{
		Type:      event.Type,
		Timestamp: time.Now(),
	}
	
	if event.Commit != nil {
		activity.SHA = event.Commit.SHA
		activity.Message = event.Commit.Message
		activity.Author = event.Commit.Author
		activity.FilesCount = len(event.Commit.FilesChanged)
		activity.LinesAdded = event.Commit.LinesAdded
		activity.LinesDeleted = event.Commit.LinesDeleted
	} else if event.PullRequest != nil {
		activity.Message = fmt.Sprintf("PR: %s", event.PullRequest.Title)
		activity.Author = event.PullRequest.Author
	}
	
	return activity
}

// updateProgressMetrics обновляет метрики прогресса
func (gpt *GitProgressTracker) updateProgressMetrics(progress *TaskProgress, activity GitActivity) {
	metrics := progress.Metrics
	
	if activity.Type == "push" {
		metrics.TotalCommits++
		metrics.TotalLinesAdded += activity.LinesAdded
		metrics.TotalLinesDeleted += activity.LinesDeleted
		metrics.FilesModified += activity.FilesCount
		metrics.LastCommitDate = activity.Timestamp
		
		// Обновляем средний размер коммита
		if metrics.TotalCommits > 0 {
			metrics.AverageCommitSize = float64(metrics.TotalLinesAdded+metrics.TotalLinesDeleted) / float64(metrics.TotalCommits)
		}
		
		// Вычисляем скорость разработки (строки кода в день)
		if len(progress.GitActivity) > 1 {
			firstActivity := progress.GitActivity[0]
			duration := activity.Timestamp.Sub(firstActivity.Timestamp)
			if duration.Hours() > 0 {
				totalLines := float64(metrics.TotalLinesAdded + metrics.TotalLinesDeleted)
				metrics.DevelopmentVelocity = totalLines / (duration.Hours() / 24)
			}
		}
	}
}

// calculateProgressPercent вычисляет процент прогресса
func (gpt *GitProgressTracker) calculateProgressPercent(progress *TaskProgress, event *GitProgressEvent) float64 {
	baseProgress := progress.ProgressPercent
	
	// Прогресс на основе активности
	switch event.Type {
	case "push":
		// Каждый коммит добавляет прогресс
		if event.Commit != nil {
			commitProgress := gpt.calculateCommitProgress(event.Commit)
			return gpt.minFloat64(baseProgress+commitProgress, 90.0) // Максимум 90% до закрытия
		}
		
	case "pull_request":
		if event.PullRequest != nil {
			switch event.PullRequest.State {
			case "opened":
				return gpt.maxFloat64(baseProgress, 70.0) // Минимум 70% при создании PR
			case "merged":
				return 100.0 // Завершено при мерже
			}
		}
	}
	
	return baseProgress
}

// calculateCommitProgress вычисляет прогресс от коммита
func (gpt *GitProgressTracker) calculateCommitProgress(commit *CommitInfo) float64 {
	// Базовый прогресс: 5% за коммит
	progress := 5.0
	
	// Бонус за размер коммита
	totalLines := commit.LinesAdded + commit.LinesDeleted
	if totalLines > 100 {
		progress += 10.0
	} else if totalLines > 50 {
		progress += 5.0
	}
	
	// Бонус за количество файлов
	if len(commit.FilesChanged) > 5 {
		progress += 5.0
	}
	
	// Анализ сообщения коммита
	message := strings.ToLower(commit.Message)
	if gpt.patterns.IsFeatureCommit(message) {
		progress += 10.0
	} else if gpt.patterns.IsFixCommit(message) {
		progress += 7.0
	} else if gpt.patterns.IsTestCommit(message) {
		progress += 3.0
	}
	
	return progress
}

// determineStageFromActivity определяет этап на основе активности
func (gpt *GitProgressTracker) determineStageFromActivity(progress *TaskProgress, event *GitProgressEvent) string {
	switch event.Type {
	case "push":
		if event.Commit != nil {
			message := strings.ToLower(event.Commit.Message)
			
			// Анализируем типы коммитов для определения этапа
			if gpt.patterns.IsTestCommit(message) {
				return "testing"
			} else if gpt.patterns.IsDocCommit(message) {
				return "documentation"
			} else if gpt.patterns.IsFeatureCommit(message) || gpt.patterns.IsFixCommit(message) {
				return "development"
			}
		}
		
	case "pull_request":
		if event.PullRequest != nil {
			switch event.PullRequest.State {
			case "opened":
				return "code_review"
			case "merged":
				return "completed"
			}
		}
	}
	
	return progress.CurrentStage
}

// checkAutoTransitions проверяет автоматические переходы workflow
func (gpt *GitProgressTracker) checkAutoTransitions(ctx context.Context, taskID string, event *GitProgressEvent) error {
	// Создаем событие для Rule Engine
	_ = &WorkflowGitEvent{
		TaskID:    taskID,
		GitEvent:  event,
		Timestamp: time.Now(),
	}
	
	// Здесь должна быть логика получения workflow для задачи
	// Пока используем заглушку
	// workflow := gpt.getWorkflowForTask(taskID)
	// if workflow != nil {
	// 	actions, err := gpt.ruleEngine.EvaluateTransitions(ctx, workflow, workflowEvent)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	
	// 	for _, action := range actions {
	// 		if _, err := gpt.ruleEngine.ExecuteAction(ctx, action, workflowEvent.GetData()); err != nil {
	// 			gpt.logger.Error("Failed to execute auto transition action", err)
	// 		}
	// 	}
	// }
	
	gpt.logger.Debug("Checked auto transitions", "task", taskID, "event_type", event.Type)
	return nil
}

// WorkflowGitEvent адаптер Git события для workflow
type WorkflowGitEvent struct {
	TaskID    string    `json:"task_id"`
	GitEvent  *GitProgressEvent `json:"git_event"`
	Timestamp time.Time `json:"timestamp"`
}

func (wge *WorkflowGitEvent) GetType() string { return "git_activity" }
func (wge *WorkflowGitEvent) GetData() map[string]interface{} {
	return map[string]interface{}{
		"task_id":    wge.TaskID,
		"git_event":  wge.GitEvent,
		"timestamp":  wge.Timestamp,
	}
}

// Утилиты

func (gpt *GitProgressTracker) removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}
	
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

func (gpt *GitProgressTracker) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (gpt *GitProgressTracker) minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func (gpt *GitProgressTracker) maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// GetTaskProgress возвращает прогресс задачи
func (gpt *GitProgressTracker) GetTaskProgress(taskID string) *TaskProgress {
	return gpt.progressEngine.GetTaskProgress(taskID)
}

// GetWebhookHandler возвращает обработчик webhooks
func (gpt *GitProgressTracker) GetWebhookHandler() *GitWebhookHandler {
	return gpt.webhookHandler
}

// StartWebhookServer запускает сервер для приема webhooks
func (gpt *GitProgressTracker) StartWebhookServer(port int) error {
	return gpt.webhookHandler.StartServer(port)
}