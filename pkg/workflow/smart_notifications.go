package workflow

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/ai"
)

// SmartNotificationEngine интеллектная система уведомлений
type SmartNotificationEngine struct {
	channels        map[string]NotificationChannel
	rules           []*NotificationRule
	aiChains        *ai.AIChains
	logger          Logger
	subscribers     map[string][]*NotificationSubscriber
	templates       *NotificationTemplates
	analytics       *NotificationAnalytics
	rateLimiter     *NotificationRateLimiter
	contextAnalyzer *NotificationContextAnalyzer
	mutex           sync.RWMutex
}

// NotificationSubscriber подписчик на уведомления
type NotificationSubscriber struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Preferences *NotificationPrefs     `json:"preferences"`
	Filters     []*NotificationFilter  `json:"filters"`
	Schedule    *NotificationSchedule  `json:"schedule"`
	Context     map[string]interface{} `json:"context"`
}

// NotificationPrefs предпочтения пользователя
type NotificationPrefs struct {
	Channels        []string               `json:"channels"`          // email, slack, teams, sms, push
	Frequency       string                 `json:"frequency"`         // immediate, batched, daily, weekly
	Priority        string                 `json:"min_priority"`      // low, medium, high, critical
	QuietHours      *QuietHours           `json:"quiet_hours"`
	GroupSimilar    bool                  `json:"group_similar"`
	AIPersonalization bool                `json:"ai_personalization"`
	CustomSettings  map[string]interface{} `json:"custom_settings"`
}

// QuietHours тихие часы
type QuietHours struct {
	Enabled   bool   `json:"enabled"`
	StartTime string `json:"start_time"` // "22:00"
	EndTime   string `json:"end_time"`   // "08:00"
	Timezone  string `json:"timezone"`   // "UTC", "America/New_York", etc.
	Weekends  bool   `json:"include_weekends"`
}

// NotificationFilter фильтр уведомлений
type NotificationFilter struct {
	Type      string      `json:"type"`       // include, exclude
	Field     string      `json:"field"`      // task_id, project, author, etc.
	Operator  string      `json:"operator"`   // equals, contains, matches, in
	Value     interface{} `json:"value"`
	Priority  int         `json:"priority"`   // порядок применения фильтров
}

// NotificationSchedule расписание уведомлений
type NotificationSchedule struct {
	Enabled     bool     `json:"enabled"`
	DaysOfWeek  []int    `json:"days_of_week"`  // 0=Sunday, 1=Monday, etc.
	TimeWindows []string `json:"time_windows"`  // ["09:00-12:00", "14:00-18:00"]
	Timezone    string   `json:"timezone"`
}

// NotificationContextAnalyzer анализатор контекста
type NotificationContextAnalyzer struct {
	aiChains *ai.AIChains
	logger   Logger
}

// NotificationContext контекст для анализа
type NotificationContext struct {
	User           *NotificationSubscriber `json:"user"`
	Event          Event                   `json:"event"`
	RecentActivity []Event                 `json:"recent_activity"`
	TeamContext    map[string]interface{}  `json:"team_context"`
	ProjectContext map[string]interface{}  `json:"project_context"`
	TimeContext    *TimeContext           `json:"time_context"`
}

// TimeContext временной контекст
type TimeContext struct {
	CurrentTime    time.Time `json:"current_time"`
	IsBusinessHours bool     `json:"is_business_hours"`
	IsWeekend      bool     `json:"is_weekend"`
	UserTimezone   string   `json:"user_timezone"`
	Urgency        string   `json:"urgency"`        // low, medium, high, critical
}

// SmartNotification умное уведомление
type SmartNotification struct {
	*Notification
	Priority          string                 `json:"priority"`
	Urgency           string                 `json:"urgency"`
	PersonalizedContent *PersonalizedContent `json:"personalized_content"`
	OptimalChannels   []string               `json:"optimal_channels"`
	OptimalTiming     *OptimalTiming         `json:"optimal_timing"`
	Context           *NotificationContext   `json:"context"`
	AIAnalysis        *AINotificationAnalysis `json:"ai_analysis"`
}

// PersonalizedContent персонализированный контент
type PersonalizedContent struct {
	Subject     string            `json:"subject"`
	Body        string            `json:"body"`
	Summary     string            `json:"summary"`
	ActionItems []string          `json:"action_items"`
	Context     map[string]string `json:"context"`
}

// OptimalTiming оптимальное время доставки
type OptimalTiming struct {
	DeliverAt        time.Time `json:"deliver_at"`
	Reasoning        string    `json:"reasoning"`
	ConfidenceScore  float64   `json:"confidence_score"`
	AlternativeTimes []time.Time `json:"alternative_times"`
}

// AINotificationAnalysis AI анализ уведомления
type AINotificationAnalysis struct {
	Importance      float64            `json:"importance"`        // 0.0 - 1.0
	Relevance       float64            `json:"relevance"`         // 0.0 - 1.0
	ActionRequired  bool               `json:"action_required"`
	Sentiment       string             `json:"sentiment"`         // positive, negative, neutral
	Topics          []string           `json:"topics"`
	Recommendations []string           `json:"recommendations"`
	Insights        map[string]interface{} `json:"insights"`
}

// NewSmartNotificationEngine создает новый движок уведомлений
func NewSmartNotificationEngine(aiChains *ai.AIChains, logger Logger) *SmartNotificationEngine {
	engine := &SmartNotificationEngine{
		channels:        make(map[string]NotificationChannel),
		rules:           []*NotificationRule{},
		aiChains:        aiChains,
		logger:          logger,
		subscribers:     make(map[string][]*NotificationSubscriber),
		templates:       NewNotificationTemplates(),
		analytics:       NewNotificationAnalytics(logger),
		rateLimiter:     NewNotificationRateLimiter(logger),
		contextAnalyzer: NewNotificationContextAnalyzer(aiChains, logger),
	}
	
	// Регистрируем стандартные каналы
	engine.RegisterChannel(NewEmailChannel(logger))
	engine.RegisterChannel(NewSlackChannel(logger))
	engine.RegisterChannel(NewTeamsChannel(logger))
	engine.RegisterChannel(NewWebhookChannel(logger))
	
	return engine
}

// RegisterChannel регистрирует канал уведомлений
func (sne *SmartNotificationEngine) RegisterChannel(channel NotificationChannel) {
	sne.mutex.Lock()
	defer sne.mutex.Unlock()
	
	sne.channels[channel.GetType()] = channel
	sne.logger.Info("Registered notification channel", "type", channel.GetType())
}

// Subscribe подписывает пользователя на уведомления
func (sne *SmartNotificationEngine) Subscribe(ctx context.Context, subscriber *NotificationSubscriber) error {
	sne.mutex.Lock()
	defer sne.mutex.Unlock()
	
	// Валидация подписчика
	if err := sne.validateSubscriber(subscriber); err != nil {
		return fmt.Errorf("invalid subscriber: %w", err)
	}
	
	// Добавляем подписчика
	if _, exists := sne.subscribers[subscriber.UserID]; !exists {
		sne.subscribers[subscriber.UserID] = []*NotificationSubscriber{}
	}
	
	sne.subscribers[subscriber.UserID] = append(sne.subscribers[subscriber.UserID], subscriber)
	
	sne.logger.Info("User subscribed to notifications", 
		"user_id", subscriber.UserID, 
		"subscription_id", subscriber.ID)
	
	return nil
}

// ProcessEvent обрабатывает событие и отправляет уведомления
func (sne *SmartNotificationEngine) ProcessEvent(ctx context.Context, event Event) error {
	sne.logger.Debug("Processing event for notifications", "event_type", event.GetType())
	
	// Находим подходящие правила уведомлений
	matchingRules := sne.findMatchingRules(event)
	if len(matchingRules) == 0 {
		sne.logger.Debug("No matching notification rules", "event_type", event.GetType())
		return nil
	}
	
	// Находим подписчиков для каждого правила
	for _, rule := range matchingRules {
		subscribers := sne.findRelevantSubscribers(event, rule)
		
		for _, subscriber := range subscribers {
			// Создаем умное уведомление
			smartNotification, err := sne.createSmartNotification(ctx, event, subscriber, rule)
			if err != nil {
				sne.logger.Error("Failed to create smart notification", err,
					"user_id", subscriber.UserID, "event_type", event.GetType())
				continue
			}
			
			// Проверяем, нужно ли отправлять уведомление
			if !sne.shouldSendNotification(ctx, smartNotification) {
				sne.logger.Debug("Notification filtered out", 
					"user_id", subscriber.UserID, 
					"reason", "filtering_rules")
				continue
			}
			
			// Отправляем уведомление
			if err := sne.sendSmartNotification(ctx, smartNotification); err != nil {
				sne.logger.Error("Failed to send notification", err,
					"user_id", subscriber.UserID, "notification_id", smartNotification.ID)
			}
		}
	}
	
	return nil
}

// createSmartNotification создает умное уведомление
func (sne *SmartNotificationEngine) createSmartNotification(ctx context.Context, event Event, subscriber *NotificationSubscriber, rule *NotificationRule) (*SmartNotification, error) {
	// Анализируем контекст
	context := sne.contextAnalyzer.AnalyzeContext(ctx, event, subscriber)
	
	// Создаем базовое уведомление
	baseNotification := &Notification{
		ID:         fmt.Sprintf("notif-%d", time.Now().UnixNano()),
		Type:       rule.Event,
		Title:      sne.generateTitle(event, rule),
		Message:    sne.generateMessage(event, rule),
		Priority:   sne.determinePriority(event, context),
		Recipients: []string{subscriber.UserID},
		Data:       event.GetData(),
		Timestamp:  time.Now(),
	}
	
	// AI анализ важности и релевантности
	aiAnalysis, err := sne.performAIAnalysis(ctx, event, subscriber, context)
	if err != nil {
		sne.logger.Error("AI analysis failed", err)
		// Продолжаем без AI анализа
	}
	
	// Персонализируем контент через отдельный движок
	personalizedEngine := NewPersonalizedTemplateEngine(sne.templates, sne.aiChains, sne.logger)
	personalizedContent, err := personalizedEngine.PersonalizeContent(ctx, baseNotification, subscriber, context)
	if err != nil {
		sne.logger.Error("Content personalization failed", err)
		// Используем базовый контент
		personalizedContent = &PersonalizedContent{
			Subject: baseNotification.Title,
			Body:    baseNotification.Message,
			Summary: baseNotification.Message,
			Context: make(map[string]string),
		}
	}
	
	// Определяем оптимальные каналы
	optimalChannels := sne.determineOptimalChannels(subscriber, context, aiAnalysis)
	
	// Определяем оптимальное время доставки
	optimalTiming := sne.determineOptimalTiming(subscriber, context, aiAnalysis)
	
	smartNotification := &SmartNotification{
		Notification:        baseNotification,
		Priority:           context.TimeContext.Urgency,
		Urgency:            sne.calculateUrgency(event, context),
		PersonalizedContent: personalizedContent,
		OptimalChannels:    optimalChannels,
		OptimalTiming:      optimalTiming,
		Context:            context,
		AIAnalysis:         aiAnalysis,
	}
	
	return smartNotification, nil
}

// performAIAnalysis выполняет AI анализ уведомления
func (sne *SmartNotificationEngine) performAIAnalysis(ctx context.Context, event Event, subscriber *NotificationSubscriber, context *NotificationContext) (*AINotificationAnalysis, error) {
	if sne.aiChains == nil || !subscriber.Preferences.AIPersonalization {
		return nil, nil
	}
	
	prompt := sne.buildAIAnalysisPrompt(event, subscriber, context)
	
	response, err := sne.aiChains.ExecuteTask("Notification Analysis", prompt, "analysis")
	if err != nil {
		return nil, err
	}
	
	analysis := &AINotificationAnalysis{
		Importance:      sne.extractImportanceScore(response),
		Relevance:       sne.extractRelevanceScore(response),
		ActionRequired:  sne.extractActionRequired(response),
		Sentiment:       sne.extractSentiment(response),
		Topics:          sne.extractTopics(response),
		Recommendations: sne.extractRecommendations(response),
		Insights:        make(map[string]interface{}),
	}
	
	return analysis, nil
}

// buildAIAnalysisPrompt строит промпт для AI анализа
func (sne *SmartNotificationEngine) buildAIAnalysisPrompt(event Event, subscriber *NotificationSubscriber, context *NotificationContext) string {
	prompt := fmt.Sprintf(`Analyze this notification for user relevance and importance:

EVENT DETAILS:
- Type: %s
- Data: %v
- Time: %s

USER CONTEXT:
- User ID: %s
- Preferences: %v
- Recent Activity: %d events in last hour
- Current Time Context: %s

TEAM/PROJECT CONTEXT:
- Team Context: %v
- Project Context: %v

Please analyze:
1. Importance (0.0-1.0): How important is this event for the user?
2. Relevance (0.0-1.0): How relevant is this to the user's current work?
3. Action Required (true/false): Does this require immediate user action?
4. Sentiment (positive/negative/neutral): Overall sentiment of the event
5. Topics: Key topics/themes in this notification
6. Recommendations: Specific recommendations for the user

Provide analysis in a structured format.`,
		event.GetType(),
		event.GetData(),
		context.TimeContext.CurrentTime.Format(time.RFC3339),
		subscriber.UserID,
		subscriber.Preferences,
		len(context.RecentActivity),
		context.TimeContext.Urgency,
		context.TeamContext,
		context.ProjectContext)
	
	return prompt
}

// shouldSendNotification проверяет, стоит ли отправлять уведомление
func (sne *SmartNotificationEngine) shouldSendNotification(ctx context.Context, notification *SmartNotification) bool {
	// Проверяем наличие получателей
	if len(notification.Recipients) == 0 {
		return false
	}
	
	// Проверка rate limiting
	if !sne.rateLimiter.AllowNotification(notification.Recipients[0], notification.Type) {
		return false
	}
	
	// Проверка тихих часов
	if sne.isQuietHours(notification) {
		return false
	}
	
	// Проверка важности
	if notification.AIAnalysis != nil && notification.AIAnalysis.Importance < 0.3 {
		return false
	}
	
	// Проверка фильтров пользователя
	if !sne.passesUserFilters(notification) {
		return false
	}
	
	return true
}

// sendSmartNotification отправляет умное уведомление
func (sne *SmartNotificationEngine) sendSmartNotification(ctx context.Context, notification *SmartNotification) error {
	// Определяем время доставки
	deliveryTime := notification.OptimalTiming.DeliverAt
	if deliveryTime.After(time.Now()) {
		// Планируем отправку на потом
		return sne.scheduleNotification(ctx, notification, deliveryTime)
	}
	
	// Отправляем немедленно по оптимальным каналам
	var errors []string
	for _, channelType := range notification.OptimalChannels {
		channel, exists := sne.channels[channelType]
		if !exists {
			errors = append(errors, fmt.Sprintf("channel %s not found", channelType))
			continue
		}
		
		// Подготавливаем уведомление для канала
		channelNotification := sne.prepareForChannel(notification, channelType)
		
		// Отправляем
		if err := channel.Send(ctx, channelNotification); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", channelType, err))
		} else {
			sne.logger.Info("Notification sent successfully", 
				"channel", channelType, 
				"notification_id", notification.ID,
				"user_id", notification.Recipients[0])
		}
	}
	
	// Записываем аналитику
	sne.analytics.RecordNotification(notification, len(errors) == 0)
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to send via some channels: %s", strings.Join(errors, "; "))
	}
	
	return nil
}

// Утилиты

func (sne *SmartNotificationEngine) validateSubscriber(subscriber *NotificationSubscriber) error {
	if subscriber.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if subscriber.Preferences == nil {
		return fmt.Errorf("preferences are required")
	}
	return nil
}

func (sne *SmartNotificationEngine) findMatchingRules(event Event) []*NotificationRule {
	var matching []*NotificationRule
	for _, rule := range sne.rules {
		if rule.Event == event.GetType() || rule.Event == "*" {
			matching = append(matching, rule)
		}
	}
	return matching
}

func (sne *SmartNotificationEngine) findRelevantSubscribers(event Event, rule *NotificationRule) []*NotificationSubscriber {
	var relevant []*NotificationSubscriber
	
	sne.mutex.RLock()
	defer sne.mutex.RUnlock()
	
	for _, subscriberList := range sne.subscribers {
		for _, subscriber := range subscriberList {
			if sne.subscriberMatchesRule(subscriber, event, rule) {
				relevant = append(relevant, subscriber)
			}
		}
	}
	
	return relevant
}

func (sne *SmartNotificationEngine) subscriberMatchesRule(subscriber *NotificationSubscriber, event Event, rule *NotificationRule) bool {
	// Проверяем фильтры подписчика
	for _, filter := range subscriber.Filters {
		if !sne.applyFilter(filter, event) {
			return false
		}
	}
	return true
}

func (sne *SmartNotificationEngine) applyFilter(filter *NotificationFilter, event Event) bool {
	// Упрощенная логика фильтрации
	data := event.GetData()
	value, exists := data[filter.Field]
	if !exists {
		return filter.Type == "exclude"
	}
	
	switch filter.Operator {
	case "equals":
		match := value == filter.Value
		return (filter.Type == "include" && match) || (filter.Type == "exclude" && !match)
	case "contains":
		if str, ok := value.(string); ok {
			if filterStr, ok := filter.Value.(string); ok {
				match := strings.Contains(str, filterStr)
				return (filter.Type == "include" && match) || (filter.Type == "exclude" && !match)
			}
		}
	}
	
	return true
}

func (sne *SmartNotificationEngine) generateTitle(event Event, rule *NotificationRule) string {
	if rule.Template != "" {
		return sne.templates.RenderTemplate(rule.Template+"_title", event.GetData())
	}
	return fmt.Sprintf("Event: %s", event.GetType())
}

func (sne *SmartNotificationEngine) generateMessage(event Event, rule *NotificationRule) string {
	if rule.Template != "" {
		return sne.templates.RenderTemplate(rule.Template+"_body", event.GetData())
	}
	return fmt.Sprintf("Event %s occurred", event.GetType())
}

func (sne *SmartNotificationEngine) determinePriority(event Event, context *NotificationContext) string {
	if context.TimeContext.Urgency == "critical" {
		return "high"
	}
	return "medium"
}

func (sne *SmartNotificationEngine) calculateUrgency(event Event, context *NotificationContext) string {
	return context.TimeContext.Urgency
}

func (sne *SmartNotificationEngine) determineOptimalChannels(subscriber *NotificationSubscriber, context *NotificationContext, aiAnalysis *AINotificationAnalysis) []string {
	// Базовые каналы из предпочтений
	channels := subscriber.Preferences.Channels
	
	// AI может предложить изменения
	if aiAnalysis != nil && aiAnalysis.Importance > 0.8 {
		// Для важных уведомлений добавляем более срочные каналы
		if !contains(channels, "sms") && contains(subscriber.Preferences.Channels, "sms") {
			channels = append(channels, "sms")
		}
	}
	
	return channels
}

func (sne *SmartNotificationEngine) determineOptimalTiming(subscriber *NotificationSubscriber, context *NotificationContext, aiAnalysis *AINotificationAnalysis) *OptimalTiming {
	now := time.Now()
	
	// Если критично - немедленно
	if context.TimeContext.Urgency == "critical" {
		return &OptimalTiming{
			DeliverAt:       now,
			Reasoning:       "Critical urgency requires immediate delivery",
			ConfidenceScore: 0.9,
		}
	}
	
	// Если тихие часы - отложить
	if sne.isInQuietHours(subscriber, now) {
		nextWindow := sne.findNextActiveWindow(subscriber, now)
		return &OptimalTiming{
			DeliverAt:       nextWindow,
			Reasoning:       "Respecting user's quiet hours",
			ConfidenceScore: 0.8,
		}
	}
	
	// Обычная доставка
	return &OptimalTiming{
		DeliverAt:       now,
		Reasoning:       "Normal business hours, immediate delivery",
		ConfidenceScore: 0.7,
	}
}

// AI анализ helper functions

func (sne *SmartNotificationEngine) extractImportanceScore(response string) float64 {
	// Упрощенное извлечение скора
	if strings.Contains(strings.ToLower(response), "high importance") {
		return 0.8
	}
	if strings.Contains(strings.ToLower(response), "medium importance") {
		return 0.5
	}
	return 0.3
}

func (sne *SmartNotificationEngine) extractRelevanceScore(response string) float64 {
	if strings.Contains(strings.ToLower(response), "highly relevant") {
		return 0.9
	}
	if strings.Contains(strings.ToLower(response), "relevant") {
		return 0.6
	}
	return 0.3
}

func (sne *SmartNotificationEngine) extractActionRequired(response string) bool {
	return strings.Contains(strings.ToLower(response), "action required")
}

func (sne *SmartNotificationEngine) extractSentiment(response string) string {
	lower := strings.ToLower(response)
	if strings.Contains(lower, "positive") {
		return "positive"
	}
	if strings.Contains(lower, "negative") {
		return "negative"
	}
	return "neutral"
}

func (sne *SmartNotificationEngine) extractTopics(response string) []string {
	// Упрощенное извлечение тем
	return []string{"workflow", "task", "progress"}
}

func (sne *SmartNotificationEngine) extractRecommendations(response string) []string {
	// Упрощенное извлечение рекомендаций
	return []string{"review immediately", "check progress"}
}

// Вспомогательные функции

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (sne *SmartNotificationEngine) isQuietHours(notification *SmartNotification) bool {
	// Упрощенная проверка тихих часов
	return false
}

func (sne *SmartNotificationEngine) isInQuietHours(subscriber *NotificationSubscriber, now time.Time) bool {
	// Упрощенная проверка
	return false
}

func (sne *SmartNotificationEngine) findNextActiveWindow(subscriber *NotificationSubscriber, now time.Time) time.Time {
	// Упрощенно - через 8 часов
	return now.Add(8 * time.Hour)
}

func (sne *SmartNotificationEngine) passesUserFilters(notification *SmartNotification) bool {
	// Упрощенная проверка фильтров
	return true
}

func (sne *SmartNotificationEngine) scheduleNotification(ctx context.Context, notification *SmartNotification, deliveryTime time.Time) error {
	// Реализация планировщика уведомлений
	sne.logger.Info("Notification scheduled for later delivery", 
		"notification_id", notification.ID,
		"delivery_time", deliveryTime.Format(time.RFC3339))
	return nil
}

func (sne *SmartNotificationEngine) prepareForChannel(notification *SmartNotification, channelType string) *Notification {
	// Адаптируем уведомление для конкретного канала
	adapted := *notification.Notification
	
	if notification.PersonalizedContent != nil {
		adapted.Title = notification.PersonalizedContent.Subject
		adapted.Message = notification.PersonalizedContent.Body
	}
	
	return &adapted
}