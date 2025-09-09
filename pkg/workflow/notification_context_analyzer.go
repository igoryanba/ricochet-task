package workflow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/ai"
)

// NewNotificationContextAnalyzer создает новый анализатор контекста
func NewNotificationContextAnalyzer(aiChains *ai.AIChains, logger Logger) *NotificationContextAnalyzer {
	return &NotificationContextAnalyzer{
		aiChains: aiChains,
		logger:   logger,
	}
}

// AnalyzeContext анализирует контекст для уведомления
func (nca *NotificationContextAnalyzer) AnalyzeContext(ctx context.Context, event Event, subscriber *NotificationSubscriber) *NotificationContext {
	context := &NotificationContext{
		User:           subscriber,
		Event:          event,
		RecentActivity: nca.getRecentActivity(subscriber.UserID, event),
		TeamContext:    nca.analyzeTeamContext(ctx, event, subscriber),
		ProjectContext: nca.analyzeProjectContext(ctx, event, subscriber),
		TimeContext:    nca.analyzeTimeContext(subscriber),
	}
	
	return context
}

// getRecentActivity получает недавнюю активность пользователя
func (nca *NotificationContextAnalyzer) getRecentActivity(userID string, currentEvent Event) []Event {
	// В реальной реализации здесь был бы запрос к хранилищу событий
	// Пока возвращаем пустой массив
	return []Event{}
}

// analyzeTeamContext анализирует командный контекст
func (nca *NotificationContextAnalyzer) analyzeTeamContext(ctx context.Context, event Event, subscriber *NotificationSubscriber) map[string]interface{} {
	teamContext := make(map[string]interface{})
	
	// Базовая информация о команде
	teamContext["user_role"] = nca.getUserRole(subscriber)
	teamContext["team_size"] = nca.getTeamSize(subscriber)
	teamContext["active_projects"] = nca.getActiveProjects(subscriber)
	
	// Анализ событий команды
	if nca.isTeamEvent(event) {
		teamContext["is_team_event"] = true
		teamContext["affected_members"] = nca.getAffectedMembers(event)
		teamContext["team_impact"] = nca.assessTeamImpact(event)
	}
	
	// Загруженность команды
	teamContext["team_workload"] = nca.assessTeamWorkload(subscriber)
	teamContext["availability"] = nca.checkTeamAvailability(subscriber)
	
	return teamContext
}

// analyzeProjectContext анализирует проектный контекст
func (nca *NotificationContextAnalyzer) analyzeProjectContext(ctx context.Context, event Event, subscriber *NotificationSubscriber) map[string]interface{} {
	projectContext := make(map[string]interface{})
	
	// Получаем проект из события
	projectID := nca.extractProjectID(event)
	if projectID != "" {
		projectContext["project_id"] = projectID
		projectContext["project_name"] = nca.getProjectName(projectID)
		projectContext["project_priority"] = nca.getProjectPriority(projectID)
		projectContext["project_deadline"] = nca.getProjectDeadline(projectID)
		projectContext["user_involvement"] = nca.getUserInvolvement(subscriber.UserID, projectID)
	}
	
	// Контекст задачи если есть
	taskID := nca.extractTaskID(event)
	if taskID != "" {
		projectContext["task_id"] = taskID
		projectContext["task_priority"] = nca.getTaskPriority(taskID)
		projectContext["task_assignee"] = nca.getTaskAssignee(taskID)
		projectContext["is_user_assignee"] = nca.isUserAssignee(subscriber.UserID, taskID)
		projectContext["task_dependencies"] = nca.getTaskDependencies(taskID)
	}
	
	return projectContext
}

// analyzeTimeContext анализирует временной контекст
func (nca *NotificationContextAnalyzer) analyzeTimeContext(subscriber *NotificationSubscriber) *TimeContext {
	now := time.Now()
	
	// Определяем временную зону пользователя
	userTimezone := nca.getUserTimezone(subscriber)
	
	// Конвертируем время в зону пользователя
	userTime := nca.convertToUserTime(now, userTimezone)
	
	return &TimeContext{
		CurrentTime:     userTime,
		IsBusinessHours: nca.isBusinessHours(userTime),
		IsWeekend:       nca.isWeekend(userTime),
		UserTimezone:    userTimezone,
		Urgency:         nca.determineUrgency(userTime, subscriber),
	}
}

// Helper methods

func (nca *NotificationContextAnalyzer) getUserRole(subscriber *NotificationSubscriber) string {
	if role, exists := subscriber.Context["role"].(string); exists {
		return role
	}
	return "member"
}

func (nca *NotificationContextAnalyzer) getTeamSize(subscriber *NotificationSubscriber) int {
	// В реальной реализации запрос к базе данных
	return 5 // Заглушка
}

func (nca *NotificationContextAnalyzer) getActiveProjects(subscriber *NotificationSubscriber) []string {
	// В реальной реализации запрос к базе данных
	return []string{"project1", "project2"} // Заглушка
}

func (nca *NotificationContextAnalyzer) isTeamEvent(event Event) bool {
	eventType := event.GetType()
	teamEventTypes := []string{
		"team_assignment",
		"team_meeting",
		"team_milestone",
		"team_alert",
	}
	
	for _, teamType := range teamEventTypes {
		if eventType == teamType {
			return true
		}
	}
	
	return false
}

func (nca *NotificationContextAnalyzer) getAffectedMembers(event Event) []string {
	data := event.GetData()
	if members, ok := data["affected_members"].([]string); ok {
		return members
	}
	return []string{}
}

func (nca *NotificationContextAnalyzer) assessTeamImpact(event Event) string {
	// Анализируем влияние события на команду
	eventType := event.GetType()
	
	switch eventType {
	case "critical_alert", "system_down":
		return "high"
	case "task_completed", "milestone_reached":
		return "medium"
	default:
		return "low"
	}
}

func (nca *NotificationContextAnalyzer) assessTeamWorkload(subscriber *NotificationSubscriber) string {
	// Упрощенная оценка загруженности команды
	// В реальности нужен анализ текущих задач и метрик
	return "medium"
}

func (nca *NotificationContextAnalyzer) checkTeamAvailability(subscriber *NotificationSubscriber) map[string]interface{} {
	return map[string]interface{}{
		"available_members": 3,
		"busy_members":      2,
		"offline_members":   0,
	}
}

func (nca *NotificationContextAnalyzer) extractProjectID(event Event) string {
	data := event.GetData()
	if projectID, ok := data["project_id"].(string); ok {
		return projectID
	}
	return ""
}

func (nca *NotificationContextAnalyzer) extractTaskID(event Event) string {
	data := event.GetData()
	if taskID, ok := data["task_id"].(string); ok {
		return taskID
	}
	return ""
}

func (nca *NotificationContextAnalyzer) getProjectName(projectID string) string {
	// В реальной реализации запрос к базе данных
	return "Sample Project"
}

func (nca *NotificationContextAnalyzer) getProjectPriority(projectID string) string {
	// В реальной реализации запрос к базе данных
	return "high"
}

func (nca *NotificationContextAnalyzer) getProjectDeadline(projectID string) *time.Time {
	// В реальной реализации запрос к базе данных
	deadline := time.Now().AddDate(0, 1, 0) // Через месяц
	return &deadline
}

func (nca *NotificationContextAnalyzer) getUserInvolvement(userID, projectID string) string {
	// Анализируем степень вовлеченности пользователя в проект
	return "active" // high, medium, low, inactive
}

func (nca *NotificationContextAnalyzer) getTaskPriority(taskID string) string {
	// В реальной реализации запрос к базе данных
	return "medium"
}

func (nca *NotificationContextAnalyzer) getTaskAssignee(taskID string) string {
	// В реальной реализации запрос к базе данных
	return "user123"
}

func (nca *NotificationContextAnalyzer) isUserAssignee(userID, taskID string) bool {
	assignee := nca.getTaskAssignee(taskID)
	return assignee == userID
}

func (nca *NotificationContextAnalyzer) getTaskDependencies(taskID string) []string {
	// В реальной реализации запрос к базе данных
	return []string{"task456", "task789"}
}

func (nca *NotificationContextAnalyzer) getUserTimezone(subscriber *NotificationSubscriber) string {
	if tz, exists := subscriber.Context["timezone"].(string); exists {
		return tz
	}
	return "UTC"
}

func (nca *NotificationContextAnalyzer) convertToUserTime(t time.Time, timezone string) time.Time {
	// Упрощенная конвертация времени
	// В реальной реализации нужно использовать библиотеку time zones
	return t
}

func (nca *NotificationContextAnalyzer) isBusinessHours(t time.Time) bool {
	hour := t.Hour()
	weekday := t.Weekday()
	
	// Рабочие часы: 9-18, пн-пт
	isWeekday := weekday >= time.Monday && weekday <= time.Friday
	isBusinessHour := hour >= 9 && hour < 18
	
	return isWeekday && isBusinessHour
}

func (nca *NotificationContextAnalyzer) isWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

func (nca *NotificationContextAnalyzer) determineUrgency(t time.Time, subscriber *NotificationSubscriber) string {
	// Определяем срочность на основе времени и контекста
	
	if !nca.isBusinessHours(t) {
		// Вне рабочих часов - только критичные уведомления срочные
		return "low"
	}
	
	if nca.isWeekend(t) {
		// Выходные - низкая срочность
		return "low"
	}
	
	hour := t.Hour()
	
	// Утренние часы - высокая продуктивность
	if hour >= 9 && hour < 12 {
		return "high"
	}
	
	// После обеда - средняя продуктивность
	if hour >= 13 && hour < 17 {
		return "medium"
	}
	
	// Конец дня - низкая срочность
	return "low"
}

// AI-powered контекстный анализ

// AnalyzeEventImportance анализирует важность события с помощью AI
func (nca *NotificationContextAnalyzer) AnalyzeEventImportance(ctx context.Context, event Event, subscriber *NotificationSubscriber) (float64, error) {
	if nca.aiChains == nil {
		return 0.5, nil // Средняя важность по умолчанию
	}
	
	prompt := nca.buildImportanceAnalysisPrompt(event, subscriber)
	
	response, err := nca.aiChains.ExecuteTask("Importance Analysis", prompt, "analysis")
	if err != nil {
		return 0.5, err
	}
	
	// Извлекаем скор важности из ответа
	importance := nca.extractImportanceScore(response)
	
	return importance, nil
}

// AnalyzeUserContext анализирует пользовательский контекст с помощью AI
func (nca *NotificationContextAnalyzer) AnalyzeUserContext(ctx context.Context, subscriber *NotificationSubscriber, recentActivity []Event) (map[string]interface{}, error) {
	if nca.aiChains == nil {
		return make(map[string]interface{}), nil
	}
	
	prompt := nca.buildUserContextPrompt(subscriber, recentActivity)
	
	response, err := nca.aiChains.ExecuteTask("User Context Analysis", prompt, "analysis")
	if err != nil {
		return nil, err
	}
	
	// Парсим контекст из ответа
	userContext := nca.parseUserContextResponse(response)
	
	return userContext, nil
}

// PredictOptimalDeliveryTime предсказывает оптимальное время доставки
func (nca *NotificationContextAnalyzer) PredictOptimalDeliveryTime(ctx context.Context, subscriber *NotificationSubscriber, event Event) (time.Time, error) {
	if nca.aiChains == nil {
		return time.Now(), nil // Немедленная доставка по умолчанию
	}
	
	prompt := nca.buildTimingPredictionPrompt(subscriber, event)
	
	response, err := nca.aiChains.ExecuteTask("Delivery Timing Prediction", prompt, "prediction")
	if err != nil {
		return time.Now(), err
	}
	
	// Извлекаем оптимальное время из ответа
	optimalTime := nca.parseOptimalTime(response)
	
	return optimalTime, nil
}

// Helper methods for AI analysis

func (nca *NotificationContextAnalyzer) buildImportanceAnalysisPrompt(event Event, subscriber *NotificationSubscriber) string {
	return fmt.Sprintf(`Analyze the importance of this notification for the user:

EVENT:
Type: %s
Data: %v

USER:
ID: %s
Role: %s
Preferences: %v

Consider:
1. Event relevance to user's role and responsibilities
2. Potential impact on user's work
3. Urgency level
4. User's typical engagement patterns

Rate importance from 0.0 (not important) to 1.0 (critical).`,
		event.GetType(),
		event.GetData(),
		subscriber.UserID,
		nca.getUserRole(subscriber),
		subscriber.Preferences)
}

func (nca *NotificationContextAnalyzer) buildUserContextPrompt(subscriber *NotificationSubscriber, recentActivity []Event) string {
	return fmt.Sprintf(`Analyze user context for notification personalization:

USER:
ID: %s
Role: %s
Preferences: %v
Context: %v

RECENT ACTIVITY (%d events):
%s

Provide insights about:
1. User's current focus areas
2. Workload level
3. Preferred communication style
4. Optimal notification frequency
5. Key interests and priorities`,
		subscriber.UserID,
		nca.getUserRole(subscriber),
		subscriber.Preferences,
		subscriber.Context,
		len(recentActivity),
		nca.formatRecentActivity(recentActivity))
}

func (nca *NotificationContextAnalyzer) buildTimingPredictionPrompt(subscriber *NotificationSubscriber, event Event) string {
	now := time.Now()
	
	return fmt.Sprintf(`Predict optimal delivery time for this notification:

CURRENT TIME: %s
EVENT: %s (Priority: %s)
USER TIMEZONE: %s
USER PREFERENCES: %v

Consider:
1. User's typical active hours
2. Event urgency and type
3. Optimal engagement times
4. Work schedule patterns

Suggest delivery time (format: YYYY-MM-DD HH:MM) and reasoning.`,
		now.Format(time.RFC3339),
		event.GetType(),
		nca.getEventPriority(event),
		nca.getUserTimezone(subscriber),
		subscriber.Preferences)
}

func (nca *NotificationContextAnalyzer) extractImportanceScore(response string) float64 {
	// Упрощенное извлечение скора из AI ответа
	response = strings.ToLower(response)
	
	if strings.Contains(response, "critical") || strings.Contains(response, "1.0") {
		return 1.0
	} else if strings.Contains(response, "high") || strings.Contains(response, "0.8") {
		return 0.8
	} else if strings.Contains(response, "medium") || strings.Contains(response, "0.5") {
		return 0.5
	} else if strings.Contains(response, "low") || strings.Contains(response, "0.2") {
		return 0.2
	}
	
	return 0.5 // Средняя важность по умолчанию
}

func (nca *NotificationContextAnalyzer) parseUserContextResponse(response string) map[string]interface{} {
	context := make(map[string]interface{})
	
	// Упрощенный парсинг AI ответа
	lines := strings.Split(response, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.Contains(strings.ToLower(line), "workload") {
			if strings.Contains(strings.ToLower(line), "high") {
				context["workload_level"] = "high"
			} else if strings.Contains(strings.ToLower(line), "low") {
				context["workload_level"] = "low"
			} else {
				context["workload_level"] = "medium"
			}
		}
		
		if strings.Contains(strings.ToLower(line), "focus") {
			context["current_focus"] = line
		}
		
		if strings.Contains(strings.ToLower(line), "style") {
			context["communication_style"] = line
		}
	}
	
	return context
}

func (nca *NotificationContextAnalyzer) parseOptimalTime(response string) time.Time {
	// Упрощенный парсинг времени из AI ответа
	// В реальной реализации нужен более сложный парсер
	
	now := time.Now()
	
	// Ищем паттерны отложенной доставки
	response = strings.ToLower(response)
	
	if strings.Contains(response, "immediate") || strings.Contains(response, "now") {
		return now
	}
	
	if strings.Contains(response, "hour") {
		return now.Add(1 * time.Hour)
	}
	
	if strings.Contains(response, "morning") {
		// Доставляем утром следующего дня
		nextMorning := now.Add(24 * time.Hour)
		return time.Date(nextMorning.Year(), nextMorning.Month(), nextMorning.Day(), 9, 0, 0, 0, nextMorning.Location())
	}
	
	if strings.Contains(response, "next day") {
		return now.Add(24 * time.Hour)
	}
	
	return now // По умолчанию немедленно
}

func (nca *NotificationContextAnalyzer) formatRecentActivity(events []Event) string {
	if len(events) == 0 {
		return "No recent activity"
	}
	
	var activity []string
	for i, event := range events {
		if i >= 5 { // Ограничиваем до 5 последних событий
			break
		}
		activity = append(activity, fmt.Sprintf("- %s", event.GetType()))
	}
	
	return strings.Join(activity, "\n")
}

func (nca *NotificationContextAnalyzer) getEventPriority(event Event) string {
	data := event.GetData()
	if priority, ok := data["priority"].(string); ok {
		return priority
	}
	return "medium"
}

