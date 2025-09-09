package workflow

import (
	"sync"
	"time"
)

// NotificationRateLimiter ограничитель частоты уведомлений
type NotificationRateLimiter struct {
	userLimits    map[string]*UserRateLimit
	globalLimits  *GlobalRateLimit
	logger        Logger
	mutex         sync.RWMutex
}

// UserRateLimit лимиты для пользователя
type UserRateLimit struct {
	UserID           string                      `json:"user_id"`
	GlobalLimit      *RateWindow                 `json:"global_limit"`      // общий лимит на все уведомления
	ByChannel        map[string]*RateWindow      `json:"by_channel"`        // лимиты по каналам
	ByType           map[string]*RateWindow      `json:"by_type"`           // лимиты по типам
	QuietMode        *QuietModeSettings          `json:"quiet_mode"`
	BurstProtection  *BurstProtectionSettings    `json:"burst_protection"`
	AdaptiveLimits   *AdaptiveLimitSettings      `json:"adaptive_limits"`
	LastReset        time.Time                   `json:"last_reset"`
}

// RateWindow окно ограничения частоты
type RateWindow struct {
	MaxCount     int           `json:"max_count"`      // максимальное количество
	Window       time.Duration `json:"window"`         // временное окно
	CurrentCount int           `json:"current_count"`  // текущее количество
	WindowStart  time.Time     `json:"window_start"`   // начало окна
	Violations   int           `json:"violations"`     // количество нарушений
}

// GlobalRateLimit глобальные лимиты системы
type GlobalRateLimit struct {
	MaxPerSecond    int                 `json:"max_per_second"`
	MaxPerMinute    int                 `json:"max_per_minute"`
	MaxPerHour      int                 `json:"max_per_hour"`
	CurrentSecond   *RateWindow         `json:"current_second"`
	CurrentMinute   *RateWindow         `json:"current_minute"`
	CurrentHour     *RateWindow         `json:"current_hour"`
	ChannelLimits   map[string]*RateWindow `json:"channel_limits"`
	EmergencyMode   bool                `json:"emergency_mode"`
	LastChecked     time.Time           `json:"last_checked"`
}

// QuietModeSettings настройки тихого режима
type QuietModeSettings struct {
	Enabled         bool      `json:"enabled"`
	StartTime       string    `json:"start_time"`    // "22:00"
	EndTime         string    `json:"end_time"`      // "08:00"
	Timezone        string    `json:"timezone"`
	AllowCritical   bool      `json:"allow_critical"`
	WeekdaysOnly    bool      `json:"weekdays_only"`
	CustomSchedule  []QuietPeriod `json:"custom_schedule"`
}

// QuietPeriod период тишины
type QuietPeriod struct {
	Name      string    `json:"name"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	Days      []int     `json:"days"`        // 0=Sunday, 1=Monday, etc.
}

// BurstProtectionSettings защита от всплесков
type BurstProtectionSettings struct {
	Enabled           bool          `json:"enabled"`
	BurstThreshold    int           `json:"burst_threshold"`    // количество уведомлений для определения всплеска
	BurstWindow       time.Duration `json:"burst_window"`       // окно для определения всплеска
	CooldownPeriod    time.Duration `json:"cooldown_period"`    // период охлаждения после всплеска
	MaxBurstsPerHour  int           `json:"max_bursts_per_hour"`
	LastBurstTime     time.Time     `json:"last_burst_time"`
	BurstCount        int           `json:"burst_count"`
}

// AdaptiveLimitSettings адаптивные лимиты
type AdaptiveLimitSettings struct {
	Enabled             bool    `json:"enabled"`
	BaseMultiplier      float64 `json:"base_multiplier"`      // базовый множитель (1.0 = стандартные лимиты)
	EngagementFactor    float64 `json:"engagement_factor"`    // фактор вовлеченности пользователя
	ResponseTimeFactor  float64 `json:"response_time_factor"` // фактор времени ответа
	LastAdjustment      time.Time `json:"last_adjustment"`
	AdjustmentHistory   []AdaptiveLimitAdjustment `json:"adjustment_history"`
}

// AdaptiveLimitAdjustment запись корректировки лимитов
type AdaptiveLimitAdjustment struct {
	Timestamp    time.Time `json:"timestamp"`
	OldMultiplier float64  `json:"old_multiplier"`
	NewMultiplier float64  `json:"new_multiplier"`
	Reason       string    `json:"reason"`
}

// NewNotificationRateLimiter создает новый ограничитель
func NewNotificationRateLimiter(logger Logger) *NotificationRateLimiter {
	return &NotificationRateLimiter{
		userLimits: make(map[string]*UserRateLimit),
		globalLimits: &GlobalRateLimit{
			MaxPerSecond:  100,
			MaxPerMinute:  1000,
			MaxPerHour:    10000,
			CurrentSecond: NewRateWindow(100, time.Second),
			CurrentMinute: NewRateWindow(1000, time.Minute),
			CurrentHour:   NewRateWindow(10000, time.Hour),
			ChannelLimits: make(map[string]*RateWindow),
			EmergencyMode: false,
		},
		logger: logger,
	}
}

// NewRateWindow создает новое окно ограничения
func NewRateWindow(maxCount int, window time.Duration) *RateWindow {
	return &RateWindow{
		MaxCount:     maxCount,
		Window:       window,
		CurrentCount: 0,
		WindowStart:  time.Now(),
		Violations:   0,
	}
}

// AllowNotification проверяет, можно ли отправить уведомление
func (nrl *NotificationRateLimiter) AllowNotification(userID, notificationType string) bool {
	nrl.mutex.Lock()
	defer nrl.mutex.Unlock()
	
	// Проверяем глобальные лимиты
	if !nrl.checkGlobalLimits() {
		nrl.logger.Warn("Global rate limit exceeded")
		return false
	}
	
	// Получаем или создаем лимиты пользователя
	userLimit := nrl.getUserLimit(userID)
	
	// Проверяем тихий режим
	if nrl.isInQuietMode(userLimit) {
		return false
	}
	
	// Проверяем защиту от всплесков
	if !nrl.checkBurstProtection(userLimit) {
		return false
	}
	
	// Проверяем пользовательские лимиты
	if !nrl.checkUserLimits(userLimit, notificationType) {
		return false
	}
	
	// Записываем отправку
	nrl.recordNotification(userLimit, notificationType)
	
	return true
}

// checkGlobalLimits проверяет глобальные лимиты
func (nrl *NotificationRateLimiter) checkGlobalLimits() bool {
	now := time.Now()
	
	// Обновляем окна если нужно
	nrl.updateGlobalWindows(now)
	
	// Проверяем лимиты
	if nrl.globalLimits.CurrentSecond.CurrentCount >= nrl.globalLimits.MaxPerSecond {
		nrl.globalLimits.CurrentSecond.Violations++
		return false
	}
	
	if nrl.globalLimits.CurrentMinute.CurrentCount >= nrl.globalLimits.MaxPerMinute {
		nrl.globalLimits.CurrentMinute.Violations++
		return false
	}
	
	if nrl.globalLimits.CurrentHour.CurrentCount >= nrl.globalLimits.MaxPerHour {
		nrl.globalLimits.CurrentHour.Violations++
		return false
	}
	
	// Увеличиваем счетчики
	nrl.globalLimits.CurrentSecond.CurrentCount++
	nrl.globalLimits.CurrentMinute.CurrentCount++
	nrl.globalLimits.CurrentHour.CurrentCount++
	
	return true
}

// updateGlobalWindows обновляет глобальные окна
func (nrl *NotificationRateLimiter) updateGlobalWindows(now time.Time) {
	// Секундное окно
	if now.Sub(nrl.globalLimits.CurrentSecond.WindowStart) >= time.Second {
		nrl.globalLimits.CurrentSecond.CurrentCount = 0
		nrl.globalLimits.CurrentSecond.WindowStart = now
	}
	
	// Минутное окно
	if now.Sub(nrl.globalLimits.CurrentMinute.WindowStart) >= time.Minute {
		nrl.globalLimits.CurrentMinute.CurrentCount = 0
		nrl.globalLimits.CurrentMinute.WindowStart = now
	}
	
	// Часовое окно
	if now.Sub(nrl.globalLimits.CurrentHour.WindowStart) >= time.Hour {
		nrl.globalLimits.CurrentHour.CurrentCount = 0
		nrl.globalLimits.CurrentHour.WindowStart = now
	}
}

// getUserLimit получает или создает лимиты пользователя
func (nrl *NotificationRateLimiter) getUserLimit(userID string) *UserRateLimit {
	if limit, exists := nrl.userLimits[userID]; exists {
		return limit
	}
	
	// Создаем лимиты по умолчанию
	limit := &UserRateLimit{
		UserID:      userID,
		GlobalLimit: NewRateWindow(50, time.Hour), // 50 уведомлений в час
		ByChannel:   make(map[string]*RateWindow),
		ByType:      make(map[string]*RateWindow),
		QuietMode: &QuietModeSettings{
			Enabled:       false,
			StartTime:     "22:00",
			EndTime:       "08:00",
			Timezone:      "UTC",
			AllowCritical: true,
			WeekdaysOnly:  false,
		},
		BurstProtection: &BurstProtectionSettings{
			Enabled:          true,
			BurstThreshold:   10,
			BurstWindow:      5 * time.Minute,
			CooldownPeriod:   30 * time.Minute,
			MaxBurstsPerHour: 3,
		},
		AdaptiveLimits: &AdaptiveLimitSettings{
			Enabled:            true,
			BaseMultiplier:     1.0,
			EngagementFactor:   1.0,
			ResponseTimeFactor: 1.0,
		},
		LastReset: time.Now(),
	}
	
	// Лимиты по каналам
	limit.ByChannel["email"] = NewRateWindow(20, time.Hour)
	limit.ByChannel["slack"] = NewRateWindow(30, time.Hour)
	limit.ByChannel["sms"] = NewRateWindow(5, time.Hour)
	limit.ByChannel["push"] = NewRateWindow(50, time.Hour)
	
	// Лимиты по типам
	limit.ByType["critical"] = NewRateWindow(10, time.Hour)
	limit.ByType["high"] = NewRateWindow(15, time.Hour)
	limit.ByType["medium"] = NewRateWindow(25, time.Hour)
	limit.ByType["low"] = NewRateWindow(10, time.Hour)
	
	nrl.userLimits[userID] = limit
	return limit
}

// isInQuietMode проверяет тихий режим
func (nrl *NotificationRateLimiter) isInQuietMode(userLimit *UserRateLimit) bool {
	if !userLimit.QuietMode.Enabled {
		return false
	}
	
	now := time.Now()
	
	// Проверяем выходные дни если нужно
	if userLimit.QuietMode.WeekdaysOnly {
		weekday := now.Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			return true
		}
	}
	
	// Проверяем временные окна
	// Упрощенная проверка для базового периода
	currentHour := now.Hour()
	
	// Парсим время (упрощенно)
	startHour := nrl.parseHour(userLimit.QuietMode.StartTime)
	endHour := nrl.parseHour(userLimit.QuietMode.EndTime)
	
	if startHour > endHour {
		// Ночной период (например, 22:00 - 08:00)
		return currentHour >= startHour || currentHour < endHour
	} else {
		// Дневной период
		return currentHour >= startHour && currentHour < endHour
	}
}

// parseHour парсит час из строки "HH:MM"
func (nrl *NotificationRateLimiter) parseHour(timeStr string) int {
	if len(timeStr) >= 2 {
		if hour := timeStr[:2]; len(hour) == 2 {
			if h := 0; h >= 0 && h <= 23 {
				// Упрощенный парсинг
				switch hour {
				case "00": return 0
				case "01": return 1
				case "02": return 2
				case "03": return 3
				case "04": return 4
				case "05": return 5
				case "06": return 6
				case "07": return 7
				case "08": return 8
				case "09": return 9
				case "10": return 10
				case "11": return 11
				case "12": return 12
				case "13": return 13
				case "14": return 14
				case "15": return 15
				case "16": return 16
				case "17": return 17
				case "18": return 18
				case "19": return 19
				case "20": return 20
				case "21": return 21
				case "22": return 22
				case "23": return 23
				}
			}
		}
	}
	return 0
}

// checkBurstProtection проверяет защиту от всплесков
func (nrl *NotificationRateLimiter) checkBurstProtection(userLimit *UserRateLimit) bool {
	if !userLimit.BurstProtection.Enabled {
		return true
	}
	
	now := time.Now()
	burst := userLimit.BurstProtection
	
	// Проверяем период охлаждения
	if now.Sub(burst.LastBurstTime) < burst.CooldownPeriod {
		nrl.logger.Debug("User in burst cooldown period", "user_id", userLimit.UserID)
		return false
	}
	
	// Проверяем количество всплесков в час
	// Упрощенная логика - в реальности нужна более сложная
	if burst.BurstCount >= burst.MaxBurstsPerHour {
		return false
	}
	
	return true
}

// checkUserLimits проверяет пользовательские лимиты
func (nrl *NotificationRateLimiter) checkUserLimits(userLimit *UserRateLimit, notificationType string) bool {
	now := time.Now()
	
	// Обновляем окна пользователя
	nrl.updateUserWindows(userLimit, now)
	
	// Применяем адаптивные лимиты
	multiplier := nrl.getAdaptiveMultiplier(userLimit)
	
	// Проверяем глобальный лимит пользователя
	adjustedLimit := int(float64(userLimit.GlobalLimit.MaxCount) * multiplier)
	if userLimit.GlobalLimit.CurrentCount >= adjustedLimit {
		userLimit.GlobalLimit.Violations++
		return false
	}
	
	// Проверяем лимиты по типу
	if typeLimit, exists := userLimit.ByType[notificationType]; exists {
		adjustedTypeLimit := int(float64(typeLimit.MaxCount) * multiplier)
		if typeLimit.CurrentCount >= adjustedTypeLimit {
			typeLimit.Violations++
			return false
		}
	}
	
	return true
}

// updateUserWindows обновляет окна пользователя
func (nrl *NotificationRateLimiter) updateUserWindows(userLimit *UserRateLimit, now time.Time) {
	// Глобальное окно
	if now.Sub(userLimit.GlobalLimit.WindowStart) >= userLimit.GlobalLimit.Window {
		userLimit.GlobalLimit.CurrentCount = 0
		userLimit.GlobalLimit.WindowStart = now
	}
	
	// Окна по каналам
	for _, window := range userLimit.ByChannel {
		if now.Sub(window.WindowStart) >= window.Window {
			window.CurrentCount = 0
			window.WindowStart = now
		}
	}
	
	// Окна по типам
	for _, window := range userLimit.ByType {
		if now.Sub(window.WindowStart) >= window.Window {
			window.CurrentCount = 0
			window.WindowStart = now
		}
	}
}

// getAdaptiveMultiplier вычисляет адаптивный множитель
func (nrl *NotificationRateLimiter) getAdaptiveMultiplier(userLimit *UserRateLimit) float64 {
	if !userLimit.AdaptiveLimits.Enabled {
		return 1.0
	}
	
	adaptive := userLimit.AdaptiveLimits
	
	// Базовый множитель
	multiplier := adaptive.BaseMultiplier
	
	// Корректировка на основе вовлеченности
	if adaptive.EngagementFactor > 1.2 {
		// Высокая вовлеченность - увеличиваем лимиты
		multiplier *= 1.2
	} else if adaptive.EngagementFactor < 0.5 {
		// Низкая вовлеченность - уменьшаем лимиты
		multiplier *= 0.8
	}
	
	// Корректировка на основе времени ответа
	if adaptive.ResponseTimeFactor > 1.0 {
		// Быстрый ответ - увеличиваем лимиты
		multiplier *= 1.1
	} else if adaptive.ResponseTimeFactor < 0.5 {
		// Медленный ответ - уменьшаем лимиты
		multiplier *= 0.9
	}
	
	// Ограничиваем диапазон
	if multiplier > 2.0 {
		multiplier = 2.0
	} else if multiplier < 0.1 {
		multiplier = 0.1
	}
	
	return multiplier
}

// recordNotification записывает отправку уведомления
func (nrl *NotificationRateLimiter) recordNotification(userLimit *UserRateLimit, notificationType string) {
	// Увеличиваем глобальный счетчик
	userLimit.GlobalLimit.CurrentCount++
	
	// Увеличиваем счетчик по типу
	if typeLimit, exists := userLimit.ByType[notificationType]; exists {
		typeLimit.CurrentCount++
	}
}

// UpdateUserEngagement обновляет показатели вовлеченности пользователя
func (nrl *NotificationRateLimiter) UpdateUserEngagement(userID string, engagementFactor, responseTimeFactor float64) {
	nrl.mutex.Lock()
	defer nrl.mutex.Unlock()
	
	userLimit := nrl.getUserLimit(userID)
	
	oldMultiplier := userLimit.AdaptiveLimits.BaseMultiplier
	
	userLimit.AdaptiveLimits.EngagementFactor = engagementFactor
	userLimit.AdaptiveLimits.ResponseTimeFactor = responseTimeFactor
	
	// Обновляем базовый множитель
	userLimit.AdaptiveLimits.BaseMultiplier = nrl.getAdaptiveMultiplier(userLimit)
	
	// Записываем корректировку
	adjustment := AdaptiveLimitAdjustment{
		Timestamp:     time.Now(),
		OldMultiplier: oldMultiplier,
		NewMultiplier: userLimit.AdaptiveLimits.BaseMultiplier,
		Reason:        "engagement_update",
	}
	
	userLimit.AdaptiveLimits.AdjustmentHistory = append(userLimit.AdaptiveLimits.AdjustmentHistory, adjustment)
	userLimit.AdaptiveLimits.LastAdjustment = time.Now()
	
	nrl.logger.Debug("Updated user engagement", 
		"user_id", userID, 
		"engagement", engagementFactor,
		"response_time", responseTimeFactor,
		"new_multiplier", userLimit.AdaptiveLimits.BaseMultiplier)
}

// SetQuietMode устанавливает тихий режим для пользователя
func (nrl *NotificationRateLimiter) SetQuietMode(userID string, settings *QuietModeSettings) {
	nrl.mutex.Lock()
	defer nrl.mutex.Unlock()
	
	userLimit := nrl.getUserLimit(userID)
	userLimit.QuietMode = settings
	
	nrl.logger.Info("Quiet mode updated", "user_id", userID, "enabled", settings.Enabled)
}

// GetUserLimits возвращает лимиты пользователя
func (nrl *NotificationRateLimiter) GetUserLimits(userID string) *UserRateLimit {
	nrl.mutex.RLock()
	defer nrl.mutex.RUnlock()
	
	return nrl.getUserLimit(userID)
}

// GetGlobalStats возвращает глобальную статистику
func (nrl *NotificationRateLimiter) GetGlobalStats() *GlobalRateLimit {
	nrl.mutex.RLock()
	defer nrl.mutex.RUnlock()
	
	// Создаем копию для безопасности
	stats := *nrl.globalLimits
	return &stats
}

// ResetUserLimits сбрасывает лимиты пользователя
func (nrl *NotificationRateLimiter) ResetUserLimits(userID string) {
	nrl.mutex.Lock()
	defer nrl.mutex.Unlock()
	
	if userLimit, exists := nrl.userLimits[userID]; exists {
		now := time.Now()
		
		userLimit.GlobalLimit.CurrentCount = 0
		userLimit.GlobalLimit.WindowStart = now
		userLimit.LastReset = now
		
		for _, window := range userLimit.ByChannel {
			window.CurrentCount = 0
			window.WindowStart = now
		}
		
		for _, window := range userLimit.ByType {
			window.CurrentCount = 0
			window.WindowStart = now
		}
		
		nrl.logger.Info("User limits reset", "user_id", userID)
	}
}

// EnableEmergencyMode включает аварийный режим
func (nrl *NotificationRateLimiter) EnableEmergencyMode() {
	nrl.mutex.Lock()
	defer nrl.mutex.Unlock()
	
	nrl.globalLimits.EmergencyMode = true
	
	// В аварийном режиме резко снижаем лимиты
	nrl.globalLimits.MaxPerSecond = 10
	nrl.globalLimits.MaxPerMinute = 100
	nrl.globalLimits.MaxPerHour = 1000
	
	nrl.logger.Warn("Emergency mode enabled - rate limits reduced")
}

// DisableEmergencyMode отключает аварийный режим
func (nrl *NotificationRateLimiter) DisableEmergencyMode() {
	nrl.mutex.Lock()
	defer nrl.mutex.Unlock()
	
	nrl.globalLimits.EmergencyMode = false
	
	// Восстанавливаем нормальные лимиты
	nrl.globalLimits.MaxPerSecond = 100
	nrl.globalLimits.MaxPerMinute = 1000
	nrl.globalLimits.MaxPerHour = 10000
	
	nrl.logger.Info("Emergency mode disabled - rate limits restored")
}