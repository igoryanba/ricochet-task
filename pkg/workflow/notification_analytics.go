package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// NotificationAnalytics аналитика уведомлений
type NotificationAnalytics struct {
	metrics    *NotificationMetrics
	history    []*NotificationRecord
	insights   *NotificationInsights
	logger     Logger
	mutex      sync.RWMutex
}

// NotificationMetrics метрики уведомлений
type NotificationMetrics struct {
	TotalSent      int64                 `json:"total_sent"`
	TotalDelivered int64                 `json:"total_delivered"`
	TotalFailed    int64                 `json:"total_failed"`
	ByChannel      map[string]int64      `json:"by_channel"`
	ByPriority     map[string]int64      `json:"by_priority"`
	ByType         map[string]int64      `json:"by_type"`
	DeliveryTimes  []time.Duration       `json:"delivery_times"`
	OpenRates      map[string]float64    `json:"open_rates"`
	ClickRates     map[string]float64    `json:"click_rates"`
	LastUpdated    time.Time             `json:"last_updated"`
}

// NotificationRecord запись о уведомлении
type NotificationRecord struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"`
	UserID          string                 `json:"user_id"`
	Channel         string                 `json:"channel"`
	Priority        string                 `json:"priority"`
	SentAt          time.Time              `json:"sent_at"`
	DeliveredAt     *time.Time             `json:"delivered_at,omitempty"`
	OpenedAt        *time.Time             `json:"opened_at,omitempty"`
	ClickedAt       *time.Time             `json:"clicked_at,omitempty"`
	Failed          bool                   `json:"failed"`
	FailureReason   string                 `json:"failure_reason,omitempty"`
	DeliveryTime    time.Duration          `json:"delivery_time"`
	PersonalizedAI  bool                   `json:"personalized_ai"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// NotificationInsights инсайты по уведомлениям
type NotificationInsights struct {
	BestChannels       map[string]float64    `json:"best_channels"`        // канал -> скор эффективности
	BestTimes          []int                 `json:"best_times"`           // часы дня для отправки
	UserPreferences    map[string]UserInsight `json:"user_preferences"`   // пользователь -> инсайты
	ContentEffectiveness map[string]float64  `json:"content_effectiveness"` // тип контента -> эффективность
	AIPersonalizationImpact float64          `json:"ai_personalization_impact"` // влияние AI персонализации
	TrendAnalysis      *TrendAnalysis        `json:"trend_analysis"`
	Recommendations    []string              `json:"recommendations"`
	GeneratedAt        time.Time             `json:"generated_at"`
}

// UserInsight инсайты по пользователю
type UserInsight struct {
	PreferredChannels []string               `json:"preferred_channels"`
	BestTimes         []int                  `json:"best_times"`
	ResponseRate      float64                `json:"response_rate"`
	EngagementScore   float64                `json:"engagement_score"`
	ContentPrefs      map[string]float64     `json:"content_preferences"`
	LastActive        time.Time              `json:"last_active"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// TrendAnalysis анализ трендов
type TrendAnalysis struct {
	DeliveryTrends    []DataPoint `json:"delivery_trends"`
	EngagementTrends  []DataPoint `json:"engagement_trends"`
	ChannelTrends     []DataPoint `json:"channel_trends"`
	WeeklyPatterns    []DataPoint `json:"weekly_patterns"`
	MonthlyPatterns   []DataPoint `json:"monthly_patterns"`
}

// DataPoint точка данных для трендов
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Label     string    `json:"label,omitempty"`
}

// NewNotificationAnalytics создает новую аналитику
func NewNotificationAnalytics(logger Logger) *NotificationAnalytics {
	return &NotificationAnalytics{
		metrics: &NotificationMetrics{
			ByChannel:  make(map[string]int64),
			ByPriority: make(map[string]int64),
			ByType:     make(map[string]int64),
			OpenRates:  make(map[string]float64),
			ClickRates: make(map[string]float64),
		},
		history: make([]*NotificationRecord, 0),
		insights: &NotificationInsights{
			BestChannels:         make(map[string]float64),
			UserPreferences:      make(map[string]UserInsight),
			ContentEffectiveness: make(map[string]float64),
		},
		logger: logger,
	}
}

// RecordNotification записывает отправленное уведомление
func (na *NotificationAnalytics) RecordNotification(notification *SmartNotification, success bool) {
	na.mutex.Lock()
	defer na.mutex.Unlock()
	
	record := &NotificationRecord{
		ID:             notification.ID,
		Type:           notification.Type,
		UserID:         notification.Recipients[0], // Упрощенно берем первого
		Priority:       notification.Priority,
		SentAt:         time.Now(),
		Failed:         !success,
		PersonalizedAI: notification.AIAnalysis != nil,
		Metadata:       make(map[string]interface{}),
	}
	
	// Записываем по каналам
	for _, channel := range notification.OptimalChannels {
		channelRecord := *record
		channelRecord.Channel = channel
		na.history = append(na.history, &channelRecord)
		
		// Обновляем метрики
		na.metrics.TotalSent++
		na.metrics.ByChannel[channel]++
		na.metrics.ByPriority[notification.Priority]++
		na.metrics.ByType[notification.Type]++
		
		if success {
			na.metrics.TotalDelivered++
		} else {
			na.metrics.TotalFailed++
		}
	}
	
	na.metrics.LastUpdated = time.Now()
	
	// Очищаем старые записи (оставляем последние 10000)
	if len(na.history) > 10000 {
		na.history = na.history[len(na.history)-10000:]
	}
	
	na.logger.Debug("Notification recorded", 
		"id", notification.ID, 
		"success", success,
		"channels", len(notification.OptimalChannels))
}

// RecordDelivery записывает доставку уведомления
func (na *NotificationAnalytics) RecordDelivery(notificationID, channel string, deliveryTime time.Duration) {
	na.mutex.Lock()
	defer na.mutex.Unlock()
	
	for _, record := range na.history {
		if record.ID == notificationID && record.Channel == channel {
			now := time.Now()
			record.DeliveredAt = &now
			record.DeliveryTime = deliveryTime
			na.metrics.DeliveryTimes = append(na.metrics.DeliveryTimes, deliveryTime)
			break
		}
	}
}

// RecordOpen записывает открытие уведомления
func (na *NotificationAnalytics) RecordOpen(notificationID, channel string) {
	na.mutex.Lock()
	defer na.mutex.Unlock()
	
	for _, record := range na.history {
		if record.ID == notificationID && record.Channel == channel {
			now := time.Now()
			record.OpenedAt = &now
			
			// Обновляем open rate
			na.updateOpenRate(channel)
			break
		}
	}
}

// RecordClick записывает клик по уведомлению
func (na *NotificationAnalytics) RecordClick(notificationID, channel string) {
	na.mutex.Lock()
	defer na.mutex.Unlock()
	
	for _, record := range na.history {
		if record.ID == notificationID && record.Channel == channel {
			now := time.Now()
			record.ClickedAt = &now
			
			// Обновляем click rate
			na.updateClickRate(channel)
			break
		}
	}
}

// GetMetrics возвращает текущие метрики
func (na *NotificationAnalytics) GetMetrics() *NotificationMetrics {
	na.mutex.RLock()
	defer na.mutex.RUnlock()
	
	// Создаем копию для безопасности
	metrics := *na.metrics
	metrics.ByChannel = make(map[string]int64)
	metrics.ByPriority = make(map[string]int64)
	metrics.ByType = make(map[string]int64)
	metrics.OpenRates = make(map[string]float64)
	metrics.ClickRates = make(map[string]float64)
	
	for k, v := range na.metrics.ByChannel {
		metrics.ByChannel[k] = v
	}
	for k, v := range na.metrics.ByPriority {
		metrics.ByPriority[k] = v
	}
	for k, v := range na.metrics.ByType {
		metrics.ByType[k] = v
	}
	for k, v := range na.metrics.OpenRates {
		metrics.OpenRates[k] = v
	}
	for k, v := range na.metrics.ClickRates {
		metrics.ClickRates[k] = v
	}
	
	return &metrics
}

// GenerateInsights генерирует инсайты на основе данных
func (na *NotificationAnalytics) GenerateInsights(ctx context.Context) *NotificationInsights {
	na.mutex.Lock()
	defer na.mutex.Unlock()
	
	insights := &NotificationInsights{
		BestChannels:         na.calculateBestChannels(),
		BestTimes:            na.calculateBestTimes(),
		UserPreferences:      na.calculateUserPreferences(),
		ContentEffectiveness: na.calculateContentEffectiveness(),
		AIPersonalizationImpact: na.calculateAIImpact(),
		TrendAnalysis:        na.calculateTrends(),
		Recommendations:      na.generateRecommendations(),
		GeneratedAt:          time.Now(),
	}
	
	na.insights = insights
	return insights
}

// calculateBestChannels вычисляет лучшие каналы
func (na *NotificationAnalytics) calculateBestChannels() map[string]float64 {
	bestChannels := make(map[string]float64)
	
	for channel := range na.metrics.ByChannel {
		openRate := na.metrics.OpenRates[channel]
		clickRate := na.metrics.ClickRates[channel]
		
		// Комбинированный скор эффективности
		effectiveness := (openRate * 0.6) + (clickRate * 0.4)
		bestChannels[channel] = effectiveness
	}
	
	return bestChannels
}

// calculateBestTimes вычисляет лучшие времена для отправки
func (na *NotificationAnalytics) calculateBestTimes() []int {
	hourCounts := make(map[int]int)
	hourEngagement := make(map[int]float64)
	
	for _, record := range na.history {
		hour := record.SentAt.Hour()
		hourCounts[hour]++
		
		if record.OpenedAt != nil {
			hourEngagement[hour] += 1.0
		}
		if record.ClickedAt != nil {
			hourEngagement[hour] += 0.5
		}
	}
	
	// Нормализуем по количеству отправок
	for hour, engagement := range hourEngagement {
		if hourCounts[hour] > 0 {
			hourEngagement[hour] = engagement / float64(hourCounts[hour])
		}
	}
	
	// Возвращаем топ-5 часов
	type hourScore struct {
		hour  int
		score float64
	}
	
	var scores []hourScore
	for hour, score := range hourEngagement {
		scores = append(scores, hourScore{hour, score})
	}
	
	// Сортируем по убыванию
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[i].score < scores[j].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}
	
	var bestTimes []int
	for i := 0; i < len(scores) && i < 5; i++ {
		bestTimes = append(bestTimes, scores[i].hour)
	}
	
	return bestTimes
}

// calculateUserPreferences вычисляет предпочтения пользователей
func (na *NotificationAnalytics) calculateUserPreferences() map[string]UserInsight {
	userInsights := make(map[string]UserInsight)
	
	userStats := make(map[string]map[string]interface{})
	
	for _, record := range na.history {
		if _, exists := userStats[record.UserID]; !exists {
			userStats[record.UserID] = map[string]interface{}{
				"channels":     make(map[string]int),
				"times":        make(map[int]int),
				"opens":        0,
				"clicks":       0,
				"total":        0,
				"last_active": record.SentAt,
			}
		}
		
		stats := userStats[record.UserID]
		
		// Каналы
		channels := stats["channels"].(map[string]int)
		channels[record.Channel]++
		
		// Времена
		times := stats["times"].(map[int]int)
		times[record.SentAt.Hour()]++
		
		// Активность
		stats["total"] = stats["total"].(int) + 1
		if record.OpenedAt != nil {
			stats["opens"] = stats["opens"].(int) + 1
		}
		if record.ClickedAt != nil {
			stats["clicks"] = stats["clicks"].(int) + 1
		}
		
		// Последняя активность
		if record.SentAt.After(stats["last_active"].(time.Time)) {
			stats["last_active"] = record.SentAt
		}
	}
	
	// Конвертируем в инсайты
	for userID, stats := range userStats {
		insight := UserInsight{
			PreferredChannels: na.getTopChannels(stats["channels"].(map[string]int)),
			BestTimes:         na.getTopTimes(stats["times"].(map[int]int)),
			LastActive:        stats["last_active"].(time.Time),
			ContentPrefs:      make(map[string]float64),
			Metadata:          make(map[string]interface{}),
		}
		
		total := stats["total"].(int)
		if total > 0 {
			insight.ResponseRate = float64(stats["opens"].(int)) / float64(total)
			insight.EngagementScore = (float64(stats["opens"].(int)) + float64(stats["clicks"].(int))*2) / float64(total)
		}
		
		userInsights[userID] = insight
	}
	
	return userInsights
}

// calculateContentEffectiveness вычисляет эффективность контента
func (na *NotificationAnalytics) calculateContentEffectiveness() map[string]float64 {
	contentStats := make(map[string]map[string]int)
	
	for _, record := range na.history {
		if _, exists := contentStats[record.Type]; !exists {
			contentStats[record.Type] = map[string]int{
				"total": 0,
				"opens": 0,
				"clicks": 0,
			}
		}
		
		stats := contentStats[record.Type]
		stats["total"]++
		if record.OpenedAt != nil {
			stats["opens"]++
		}
		if record.ClickedAt != nil {
			stats["clicks"]++
		}
	}
	
	effectiveness := make(map[string]float64)
	for contentType, stats := range contentStats {
		if stats["total"] > 0 {
			openRate := float64(stats["opens"]) / float64(stats["total"])
			clickRate := float64(stats["clicks"]) / float64(stats["total"])
			effectiveness[contentType] = (openRate * 0.7) + (clickRate * 0.3)
		}
	}
	
	return effectiveness
}

// calculateAIImpact вычисляет влияние AI персонализации
func (na *NotificationAnalytics) calculateAIImpact() float64 {
	aiStats := map[string]int{"total": 0, "opens": 0, "clicks": 0}
	nonAIStats := map[string]int{"total": 0, "opens": 0, "clicks": 0}
	
	for _, record := range na.history {
		var stats map[string]int
		if record.PersonalizedAI {
			stats = aiStats
		} else {
			stats = nonAIStats
		}
		
		stats["total"]++
		if record.OpenedAt != nil {
			stats["opens"]++
		}
		if record.ClickedAt != nil {
			stats["clicks"]++
		}
	}
	
	var aiEffectiveness, nonAIEffectiveness float64
	
	if aiStats["total"] > 0 {
		aiEffectiveness = (float64(aiStats["opens"]) + float64(aiStats["clicks"])*2) / float64(aiStats["total"])
	}
	
	if nonAIStats["total"] > 0 {
		nonAIEffectiveness = (float64(nonAIStats["opens"]) + float64(nonAIStats["clicks"])*2) / float64(nonAIStats["total"])
	}
	
	if nonAIEffectiveness > 0 {
		return (aiEffectiveness - nonAIEffectiveness) / nonAIEffectiveness
	}
	
	return 0
}

// calculateTrends вычисляет тренды
func (na *NotificationAnalytics) calculateTrends() *TrendAnalysis {
	now := time.Now()
	
	// Группируем данные по дням
	dailyStats := make(map[string]map[string]int)
	
	for _, record := range na.history {
		day := record.SentAt.Format("2006-01-02")
		
		if _, exists := dailyStats[day]; !exists {
			dailyStats[day] = map[string]int{
				"sent": 0, "opened": 0, "clicked": 0,
			}
		}
		
		dailyStats[day]["sent"]++
		if record.OpenedAt != nil {
			dailyStats[day]["opened"]++
		}
		if record.ClickedAt != nil {
			dailyStats[day]["clicked"]++
		}
	}
	
	// Создаем тренды за последние 30 дней
	var deliveryTrends, engagementTrends []DataPoint
	
	for i := 29; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		day := date.Format("2006-01-02")
		
		stats := dailyStats[day]
		if stats == nil {
			stats = map[string]int{"sent": 0, "opened": 0, "clicked": 0}
		}
		
		deliveryTrends = append(deliveryTrends, DataPoint{
			Timestamp: date,
			Value:     float64(stats["sent"]),
			Label:     "sent",
		})
		
		var engagementRate float64
		if stats["sent"] > 0 {
			engagementRate = float64(stats["opened"]+stats["clicked"]) / float64(stats["sent"])
		}
		
		engagementTrends = append(engagementTrends, DataPoint{
			Timestamp: date,
			Value:     engagementRate,
			Label:     "engagement",
		})
	}
	
	return &TrendAnalysis{
		DeliveryTrends:   deliveryTrends,
		EngagementTrends: engagementTrends,
		// Остальные тренды можно добавить по аналогии
	}
}

// generateRecommendations генерирует рекомендации
func (na *NotificationAnalytics) generateRecommendations() []string {
	var recommendations []string
	
	// Анализ open rates
	avgOpenRate := na.calculateAverageOpenRate()
	if avgOpenRate < 0.2 {
		recommendations = append(recommendations, "Consider improving notification content - open rate is below 20%")
	}
	
	// Анализ лучших каналов
	bestChannels := na.calculateBestChannels()
	if len(bestChannels) > 0 {
		var bestChannel string
		var bestScore float64
		for channel, score := range bestChannels {
			if score > bestScore {
				bestChannel = channel
				bestScore = score
			}
		}
		recommendations = append(recommendations, 
			fmt.Sprintf("Focus on '%s' channel - it shows highest engagement (%.1f%%)", bestChannel, bestScore*100))
	}
	
	// AI персонализация
	aiImpact := na.calculateAIImpact()
	if aiImpact > 0.1 {
		recommendations = append(recommendations, 
			fmt.Sprintf("AI personalization improves engagement by %.1f%% - consider enabling for more users", aiImpact*100))
	}
	
	// Время отправки
	bestTimes := na.calculateBestTimes()
	if len(bestTimes) > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("Optimal sending times are %v - adjust scheduling accordingly", bestTimes))
	}
	
	return recommendations
}

// Вспомогательные методы

func (na *NotificationAnalytics) updateOpenRate(channel string) {
	sent := na.metrics.ByChannel[channel]
	opened := int64(0)
	
	for _, record := range na.history {
		if record.Channel == channel && record.OpenedAt != nil {
			opened++
		}
	}
	
	if sent > 0 {
		na.metrics.OpenRates[channel] = float64(opened) / float64(sent)
	}
}

func (na *NotificationAnalytics) updateClickRate(channel string) {
	sent := na.metrics.ByChannel[channel]
	clicked := int64(0)
	
	for _, record := range na.history {
		if record.Channel == channel && record.ClickedAt != nil {
			clicked++
		}
	}
	
	if sent > 0 {
		na.metrics.ClickRates[channel] = float64(clicked) / float64(sent)
	}
}

func (na *NotificationAnalytics) calculateAverageOpenRate() float64 {
	if len(na.metrics.OpenRates) == 0 {
		return 0
	}
	
	var total float64
	for _, rate := range na.metrics.OpenRates {
		total += rate
	}
	
	return total / float64(len(na.metrics.OpenRates))
}

func (na *NotificationAnalytics) getTopChannels(channels map[string]int) []string {
	type channelCount struct {
		channel string
		count   int
	}
	
	var counts []channelCount
	for channel, count := range channels {
		counts = append(counts, channelCount{channel, count})
	}
	
	// Сортируем по убыванию
	for i := 0; i < len(counts)-1; i++ {
		for j := i + 1; j < len(counts); j++ {
			if counts[i].count < counts[j].count {
				counts[i], counts[j] = counts[j], counts[i]
			}
		}
	}
	
	var result []string
	for i := 0; i < len(counts) && i < 3; i++ {
		result = append(result, counts[i].channel)
	}
	
	return result
}

func (na *NotificationAnalytics) getTopTimes(times map[int]int) []int {
	type timeCount struct {
		hour  int
		count int
	}
	
	var counts []timeCount
	for hour, count := range times {
		counts = append(counts, timeCount{hour, count})
	}
	
	// Сортируем по убыванию
	for i := 0; i < len(counts)-1; i++ {
		for j := i + 1; j < len(counts); j++ {
			if counts[i].count < counts[j].count {
				counts[i], counts[j] = counts[j], counts[i]
			}
		}
	}
	
	var result []int
	for i := 0; i < len(counts) && i < 3; i++ {
		result = append(result, counts[i].hour)
	}
	
	return result
}