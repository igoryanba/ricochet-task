package workflow

import (
	"context"
	"testing"
	"time"
)

// TestSmartNotificationEngine тестирует основной движок уведомлений
func TestSmartNotificationEngine(t *testing.T) {
	logger := &MockLogger{}
	engine := NewSmartNotificationEngine(nil, logger)
	
	t.Run("RegisterChannel", func(t *testing.T) {
		emailChannel := NewEmailChannel(logger)
		engine.RegisterChannel(emailChannel)
		
		if len(engine.channels) == 0 {
			t.Error("Channel registration failed")
		}
		
		if engine.channels["email"] == nil {
			t.Error("Email channel not registered properly")
		}
	})
	
	t.Run("Subscribe", func(t *testing.T) {
		subscriber := &NotificationSubscriber{
			ID:     "sub1",
			UserID: "user123",
			Preferences: &NotificationPrefs{
				Channels:  []string{"email", "slack"},
				Frequency: "immediate",
				Priority:  "medium",
				QuietHours: &QuietHours{
					Enabled:   false,
					StartTime: "22:00",
					EndTime:   "08:00",
					Timezone:  "UTC",
				},
				GroupSimilar:      false,
				AIPersonalization: true,
			},
			Filters:  []*NotificationFilter{},
			Schedule: &NotificationSchedule{Enabled: false},
			Context:  map[string]interface{}{"role": "developer"},
		}
		
		err := engine.Subscribe(context.Background(), subscriber)
		if err != nil {
			t.Fatalf("Failed to subscribe: %v", err)
		}
		
		if len(engine.subscribers["user123"]) != 1 {
			t.Error("Subscriber not added properly")
		}
	})
	
	t.Run("ProcessEvent", func(t *testing.T) {
		// Создаем тестовое событие
		event := &WorkflowEvent{
			Type:      "task_assigned",
			Timestamp: time.Now(),
			Source:    "workflow",
			Data: map[string]interface{}{
				"task_id":    "task123",
				"task_title": "Implement feature X",
				"assignee":   "user123",
				"priority":   "high",
			},
			WorkflowID: "wf1",
		}
		
		// Добавляем правило уведомления
		rule := &NotificationRule{
			Event:    "task_assigned",
			Channels: []string{"email"},
			Template: "task_assigned",
			Users:    []string{"user123"},
		}
		engine.rules = append(engine.rules, rule)
		
		err := engine.ProcessEvent(context.Background(), event)
		if err != nil {
			t.Fatalf("Failed to process event: %v", err)
		}
	})
}

// TestNotificationTemplates тестирует систему шаблонов
func TestNotificationTemplates(t *testing.T) {
	templates := NewNotificationTemplates()
	
	t.Run("RegisterTemplate", func(t *testing.T) {
		err := templates.RegisterTemplate("test_title", "Hello {{.name}}")
		if err != nil {
			t.Fatalf("Failed to register template: %v", err)
		}
		
		if len(templates.templates) == 0 {
			t.Error("Template not registered")
		}
	})
	
	t.Run("RenderTemplate", func(t *testing.T) {
		data := map[string]interface{}{
			"name": "John",
		}
		
		result := templates.RenderTemplate("test_title", data)
		expected := "Hello John"
		
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
	
	t.Run("DefaultTemplates", func(t *testing.T) {
		// Проверяем что загружены шаблоны по умолчанию
		availableTemplates := templates.GetAvailableTemplates()
		
		if len(availableTemplates) < 10 {
			t.Error("Default templates not loaded properly")
		}
		
		// Проверяем конкретные шаблоны
		expectedTemplates := []string{
			"task_created_title",
			"task_assigned_title", 
			"task_completed_title",
			"git_push_title",
		}
		
		for _, expected := range expectedTemplates {
			found := false
			for _, actual := range availableTemplates {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Default template %s not found", expected)
			}
		}
	})
	
	t.Run("TemplateHelpers", func(t *testing.T) {
		// Тестируем helper функции
		
		// formatTime
		timeStr := formatTime(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC))
		if timeStr != "2024-01-01 12:00:00" {
			t.Errorf("formatTime failed: %s", timeStr)
		}
		
		// pluralize
		plural1 := pluralize(1, "item", "items")
		if plural1 != "item" {
			t.Errorf("pluralize(1) failed: %s", plural1)
		}
		
		plural2 := pluralize(2, "item", "items")
		if plural2 != "items" {
			t.Errorf("pluralize(2) failed: %s", plural2)
		}
		
		// truncate
		truncated := truncate("This is a very long string that should be truncated", 20)
		if len(truncated) != 20 {
			t.Errorf("truncate failed: length %d, content: %s", len(truncated), truncated)
		}
	})
}

// TestNotificationChannels тестирует каналы уведомлений
func TestNotificationChannels(t *testing.T) {
	logger := &MockLogger{}
	
	t.Run("EmailChannel", func(t *testing.T) {
		emailChannel := NewEmailChannel(logger)
		
		if emailChannel.GetType() != "email" {
			t.Error("Email channel type incorrect")
		}
		
		notification := &Notification{
			ID:         "notif1",
			Type:       "test",
			Title:      "Test Email",
			Message:    "This is a test email notification",
			Priority:   "medium",
			Recipients: []string{"test@example.com"},
			Data:       map[string]interface{}{},
			Timestamp:  time.Now(),
		}
		
		// В тестовом режиме это должно пройти без ошибок
		err := emailChannel.Send(context.Background(), notification)
		if err != nil {
			t.Fatalf("Failed to send email: %v", err)
		}
	})
	
	t.Run("SlackChannel", func(t *testing.T) {
		slackChannel := NewSlackChannel(logger)
		
		if slackChannel.GetType() != "slack" {
			t.Error("Slack channel type incorrect")
		}
		
		notification := &Notification{
			ID:        "notif2",
			Type:      "test",
			Title:     "Test Slack",
			Message:   "This is a test Slack notification",
			Priority:  "high",
			Recipients: []string{"#general"},
			Data:      map[string]interface{}{},
			Timestamp: time.Now(),
		}
		
		err := slackChannel.Send(context.Background(), notification)
		if err != nil {
			t.Fatalf("Failed to send Slack message: %v", err)
		}
	})
	
	t.Run("WebhookChannel", func(t *testing.T) {
		webhookChannel := NewWebhookChannel(logger)
		
		if webhookChannel.GetType() != "webhook" {
			t.Error("Webhook channel type incorrect")
		}
		
		notification := &Notification{
			ID:        "notif3",
			Type:      "test",
			Title:     "Test Webhook",
			Message:   "This is a test webhook notification",
			Priority:  "low",
			Recipients: []string{"webhook-service"},
			Data:      map[string]interface{}{
				"webhook_url": "https://example.com/webhook",
			},
			Timestamp: time.Now(),
		}
		
		err := webhookChannel.Send(context.Background(), notification)
		if err != nil {
			t.Fatalf("Failed to send webhook: %v", err)
		}
	})
}

// TestNotificationAnalytics тестирует аналитику уведомлений
func TestNotificationAnalytics(t *testing.T) {
	logger := &MockLogger{}
	analytics := NewNotificationAnalytics(logger)
	
	t.Run("RecordNotification", func(t *testing.T) {
		smartNotification := &SmartNotification{
			Notification: &Notification{
				ID:         "notif1",
				Type:       "task_assigned",
				Title:      "Task Assigned",
				Message:    "You have a new task",
				Priority:   "medium",
				Recipients: []string{"user123"},
				Timestamp:  time.Now(),
			},
			Priority:        "medium",
			OptimalChannels: []string{"email", "slack"},
			AIAnalysis:      &AINotificationAnalysis{
				Importance: 0.7,
				Relevance:  0.8,
			},
		}
		
		analytics.RecordNotification(smartNotification, true)
		
		metrics := analytics.GetMetrics()
		if metrics.TotalSent != 2 { // 2 канала
			t.Errorf("Expected 2 sent notifications, got %d", metrics.TotalSent)
		}
		
		if metrics.TotalDelivered != 2 {
			t.Errorf("Expected 2 delivered notifications, got %d", metrics.TotalDelivered)
		}
	})
	
	t.Run("RecordEngagement", func(t *testing.T) {
		// Записываем открытие уведомления
		analytics.RecordOpen("notif1", "email")
		
		// Записываем клик
		analytics.RecordClick("notif1", "email")
		
		metrics := analytics.GetMetrics()
		
		// Проверяем что open rate обновился
		if emailOpenRate, exists := metrics.OpenRates["email"]; !exists || emailOpenRate == 0 {
			t.Error("Open rate not updated properly")
		}
		
		// Проверяем что click rate обновился  
		if emailClickRate, exists := metrics.ClickRates["email"]; !exists || emailClickRate == 0 {
			t.Error("Click rate not updated properly")
		}
	})
	
	t.Run("GenerateInsights", func(t *testing.T) {
		insights := analytics.GenerateInsights(context.Background())
		
		if insights == nil {
			t.Fatal("Insights generation failed")
		}
		
		if len(insights.BestChannels) == 0 {
			t.Error("No best channels identified")
		}
		
		if len(insights.Recommendations) == 0 {
			t.Error("No recommendations generated")
		}
		
		if insights.GeneratedAt.IsZero() {
			t.Error("Invalid generation timestamp")
		}
	})
}

// TestNotificationRateLimiter тестирует ограничитель частоты
func TestNotificationRateLimiter(t *testing.T) {
	logger := &MockLogger{}
	rateLimiter := NewNotificationRateLimiter(logger)
	
	t.Run("AllowNotification", func(t *testing.T) {
		// Первые уведомления должны проходить
		allowed := rateLimiter.AllowNotification("user123", "task_assigned")
		if !allowed {
			t.Error("First notification should be allowed")
		}
		
		// Проверяем что лимиты обновились
		userLimits := rateLimiter.GetUserLimits("user123")
		if userLimits.GlobalLimit.CurrentCount != 1 {
			t.Errorf("Expected count 1, got %d", userLimits.GlobalLimit.CurrentCount)
		}
	})
	
	t.Run("RateLimitExceeded", func(t *testing.T) {
		// Отправляем много уведомлений чтобы превысить лимит
		userID := "user456"
		
		allowedCount := 0
		for i := 0; i < 100; i++ {
			if rateLimiter.AllowNotification(userID, "test") {
				allowedCount++
			}
		}
		
		// Не все должны пройти из-за лимитов
		if allowedCount >= 100 {
			t.Error("Rate limiting not working properly")
		}
	})
	
	t.Run("QuietMode", func(t *testing.T) {
		userID := "user789"
		
		// Включаем тихий режим
		quietSettings := &QuietModeSettings{
			Enabled:       true,
			StartTime:     "00:00",
			EndTime:       "23:59",
			Timezone:      "UTC",
			AllowCritical: false,
			WeekdaysOnly:  false,
		}
		
		rateLimiter.SetQuietMode(userID, quietSettings)
		
		// Уведомление не должно пройти
		allowed := rateLimiter.AllowNotification(userID, "test")
		if allowed {
			t.Error("Notification should be blocked by quiet mode")
		}
	})
	
	t.Run("AdaptiveLimits", func(t *testing.T) {
		userID := "user101"
		
		// Обновляем показатели вовлеченности
		rateLimiter.UpdateUserEngagement(userID, 0.9, 0.8) // Высокая вовлеченность
		
		userLimits := rateLimiter.GetUserLimits(userID)
		
		// Множитель должен увеличиться для активного пользователя
		if userLimits.AdaptiveLimits.BaseMultiplier <= 1.0 {
			t.Error("Adaptive limits not adjusted for high engagement")
		}
	})
}

// TestNotificationContextAnalyzer тестирует анализатор контекста
func TestNotificationContextAnalyzer(t *testing.T) {
	logger := &MockLogger{}
	analyzer := NewNotificationContextAnalyzer(nil, logger)
	
	t.Run("AnalyzeContext", func(t *testing.T) {
		subscriber := &NotificationSubscriber{
			ID:     "sub1",
			UserID: "user123",
			Preferences: &NotificationPrefs{
				Channels: []string{"email"},
			},
			Context: map[string]interface{}{
				"role":     "developer",
				"timezone": "UTC",
			},
		}
		
		event := &WorkflowEvent{
			Type:      "task_assigned",
			Timestamp: time.Now(),
			Source:    "workflow",
			Data: map[string]interface{}{
				"task_id":    "task123",
				"project_id": "proj456",
				"priority":   "high",
			},
		}
		
		context := analyzer.AnalyzeContext(context.Background(), event, subscriber)
		
		if context == nil {
			t.Fatal("Context analysis failed")
		}
		
		if context.User != subscriber {
			t.Error("User not set in context")
		}
		
		if context.Event != event {
			t.Error("Event not set in context")
		}
		
		if context.TimeContext == nil {
			t.Error("Time context not analyzed")
		}
		
		if context.TeamContext == nil {
			t.Error("Team context not analyzed")
		}
		
		if context.ProjectContext == nil {
			t.Error("Project context not analyzed")
		}
	})
	
	t.Run("TimeContextAnalysis", func(t *testing.T) {
		subscriber := &NotificationSubscriber{
			ID:     "sub1",
			UserID: "user123",
			Context: map[string]interface{}{
				"timezone": "UTC",
			},
		}
		
		timeContext := analyzer.analyzeTimeContext(subscriber)
		
		if timeContext == nil {
			t.Fatal("Time context analysis failed")
		}
		
		if timeContext.UserTimezone != "UTC" {
			t.Errorf("Expected timezone UTC, got %s", timeContext.UserTimezone)
		}
		
		if timeContext.Urgency == "" {
			t.Error("Urgency not determined")
		}
	})
}

// TestPersonalizedTemplateEngine тестирует персонализированные шаблоны
func TestPersonalizedTemplateEngine(t *testing.T) {
	logger := &MockLogger{}
	templates := NewNotificationTemplates()
	engine := NewPersonalizedTemplateEngine(templates, nil, logger)
	
	t.Run("PersonalizeContent", func(t *testing.T) {
		notification := &Notification{
			ID:      "notif1",
			Type:    "task_assigned",
			Title:   "New Task Assignment",
			Message: "You have been assigned a new task",
		}
		
		subscriber := &NotificationSubscriber{
			UserID: "user123",
			Preferences: &NotificationPrefs{
				AIPersonalization: false, // Отключаем AI для простого теста
			},
		}
		
		notificationContext := &NotificationContext{
			User: subscriber,
			TimeContext: &TimeContext{
				Urgency: "medium",
			},
		}
		
		personalizedContent, err := engine.PersonalizeContent(
			context.Background(), 
			notification, 
			subscriber, 
			notificationContext,
		)
		
		if err != nil {
			t.Fatalf("Content personalization failed: %v", err)
		}
		
		if personalizedContent == nil {
			t.Fatal("Personalized content is nil")
		}
		
		if personalizedContent.Subject == "" {
			t.Error("Personalized subject is empty")
		}
		
		if personalizedContent.Body == "" {
			t.Error("Personalized body is empty")
		}
		
		if len(personalizedContent.ActionItems) == 0 {
			t.Error("No action items generated")
		}
	})
}

// BenchmarkNotificationProcessing бенчмарк обработки уведомлений
func BenchmarkNotificationProcessing(b *testing.B) {
	logger := &MockLogger{}
	engine := NewSmartNotificationEngine(nil, logger)
	
	// Добавляем канал
	engine.RegisterChannel(NewEmailChannel(logger))
	
	// Добавляем подписчика
	subscriber := &NotificationSubscriber{
		ID:     "sub1",
		UserID: "user123",
		Preferences: &NotificationPrefs{
			Channels:          []string{"email"},
			Frequency:         "immediate",
			AIPersonalization: false,
		},
		Context: map[string]interface{}{"role": "developer"},
	}
	engine.Subscribe(context.Background(), subscriber)
	
	// Добавляем правило
	rule := &NotificationRule{
		Event:    "test_event",
		Channels: []string{"email"},
		Users:    []string{"user123"},
	}
	engine.rules = append(engine.rules, rule)
	
	// Создаем тестовое событие
	event := &WorkflowEvent{
		Type:      "test_event",
		Timestamp: time.Now(),
		Source:    "test",
		Data:      map[string]interface{}{"test": "data"},
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		engine.ProcessEvent(context.Background(), event)
	}
}

// BenchmarkTemplateRendering бенчмарк рендеринга шаблонов
func BenchmarkTemplateRendering(b *testing.B) {
	templates := NewNotificationTemplates()
	
	data := map[string]interface{}{
		"title":       "Test Task",
		"assignee":    "John Doe",
		"priority":    "high",
		"due_date":    time.Now(),
		"description": "This is a test task description",
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		templates.RenderTemplate("task_assigned_body", data)
	}
}