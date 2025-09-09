package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Logger interface for hybrid AI client
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
	Warn(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// HybridAIClient клиент для работы с AI через GRIK Gateway + пользовательские ключи
type HybridAIClient struct {
	// GRIK AI Gateway (для подписки)
	GatewayURL   string
	GatewayToken string
	
	// Direct API clients (для пользовательских ключей)
	DirectClients map[string]DirectAIClient
	
	// User context
	UserID       string
	UserAPIKeys  *UserAPIKeys
	
	HTTPClient   *http.Client
	Logger       Logger
}

// UserAPIKeys пользовательские API ключи
type UserAPIKeys struct {
	OpenAI     *APIKeyConfig `json:"openai,omitempty"`
	Anthropic  *APIKeyConfig `json:"anthropic,omitempty"`
	DeepSeek   *APIKeyConfig `json:"deepseek,omitempty"`
	Grok       *APIKeyConfig `json:"grok,omitempty"`
}

// APIKeyConfig конфигурация API ключа
type APIKeyConfig struct {
	APIKey      string    `json:"api_key"`
	BaseURL     string    `json:"base_url,omitempty"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	UsageCount  int       `json:"usage_count"`
	// Лимиты пользователя (опционально)
	RateLimit   *RateLimitConfig `json:"rate_limit,omitempty"`
}

// RateLimitConfig лимиты для пользовательских ключей
type RateLimitConfig struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	RequestsPerHour   int `json:"requests_per_hour"`
	RequestsPerDay    int `json:"requests_per_day"`
}

// DirectAIClient интерфейс для прямых API клиентов
type DirectAIClient interface {
	Chat(ctx context.Context, request *HybridChatRequest) (*HybridChatResponse, error)
	GetModels() []string
	ValidateKey() error
}

// AIRoutingStrategy стратегия выбора AI провайдера
type AIRoutingStrategy string

const (
	RouteUserKeyFirst   AIRoutingStrategy = "user_key_first"   // Сначала пользовательский ключ
	RouteSubscription   AIRoutingStrategy = "subscription"     // Только подписка
	RouteUserKeyOnly    AIRoutingStrategy = "user_key_only"    // Только пользовательские ключи
	RouteCostOptimized  AIRoutingStrategy = "cost_optimized"   // Оптимизация по стоимости
	RouteBalanced       AIRoutingStrategy = "balanced"         // Балансировка нагрузки
)

// HybridChatRequest запрос на чат для HybridAIClient
type HybridChatRequest struct {
	Model         string                 `json:"model"`
	Messages      []Message              `json:"messages"`
	Temperature   float64                `json:"temperature,omitempty"`
	MaxTokens     int                    `json:"max_tokens,omitempty"`
	Stream        bool                   `json:"stream,omitempty"`
	
	// Ricochet-специфичные параметры
	Strategy      AIRoutingStrategy      `json:"strategy,omitempty"`
	ForceProvider string                 `json:"force_provider,omitempty"`
	UserContext   map[string]interface{} `json:"user_context,omitempty"`
}

// HybridChatResponse ответ чата от HybridAIClient
type HybridChatResponse struct {
	ID        string    `json:"id"`
	Model     string    `json:"model"`
	Provider  string    `json:"provider"`  // "grik_subscription" или "user_openai"
	Created   int64     `json:"created"`
	Choices   []Choice  `json:"choices"`
	Usage     Usage     `json:"usage"`
	
	// Ricochet метаданные
	RoutedVia     string    `json:"routed_via"`      // Как был выбран провайдер
	Cost          *float64  `json:"cost,omitempty"`  // Стоимость запроса
	BilledTo      string    `json:"billed_to"`       // "user_key" или "subscription"
}

// NewHybridAIClient создает новый гибридный AI клиент
func NewHybridAIClient(gatewayURL, gatewayToken, userID string, userKeys *UserAPIKeys, logger Logger) *HybridAIClient {
	client := &HybridAIClient{
		GatewayURL:    gatewayURL,
		GatewayToken:  gatewayToken,
		UserID:        userID,
		UserAPIKeys:   userKeys,
		DirectClients: make(map[string]DirectAIClient),
		HTTPClient:    &http.Client{Timeout: 60 * time.Second},
		Logger:        logger,
	}

	// Инициализируем прямые клиенты для пользовательских ключей
	client.initializeDirectClients()

	return client
}

// initializeDirectClients инициализирует прямые клиенты для пользовательских ключей
func (c *HybridAIClient) initializeDirectClients() {
	if c.UserAPIKeys == nil {
		return
	}

	// OpenAI Direct Client
	if c.UserAPIKeys.OpenAI != nil && c.UserAPIKeys.OpenAI.Enabled {
		c.DirectClients["openai"] = NewOpenAIDirectClient(c.UserAPIKeys.OpenAI, c.Logger)
	}

	// Anthropic Direct Client
	if c.UserAPIKeys.Anthropic != nil && c.UserAPIKeys.Anthropic.Enabled {
		c.DirectClients["anthropic"] = NewAnthropicDirectClient(c.UserAPIKeys.Anthropic, c.Logger)
	}

	// DeepSeek Direct Client
	if c.UserAPIKeys.DeepSeek != nil && c.UserAPIKeys.DeepSeek.Enabled {
		c.DirectClients["deepseek"] = NewDeepSeekDirectClient(c.UserAPIKeys.DeepSeek, c.Logger)
	}

	// Grok Direct Client
	if c.UserAPIKeys.Grok != nil && c.UserAPIKeys.Grok.Enabled {
		c.DirectClients["grok"] = NewGrokDirectClient(c.UserAPIKeys.Grok, c.Logger)
	}
}

// Chat выполняет чат запрос с интеллектуальным роутингом
func (c *HybridAIClient) Chat(ctx context.Context, request *HybridChatRequest) (*HybridChatResponse, error) {
	// Определяем стратегию роутинга
	strategy := request.Strategy
	if strategy == "" {
		strategy = RouteUserKeyFirst // По умолчанию сначала пользовательские ключи
	}

	// Принудительный выбор провайдера
	if request.ForceProvider != "" {
		return c.routeToSpecificProvider(ctx, request, request.ForceProvider)
	}

	// Роутинг по стратегии
	switch strategy {
	case RouteUserKeyFirst:
		return c.routeUserKeyFirst(ctx, request)
	case RouteSubscription:
		return c.routeToSubscription(ctx, request)
	case RouteUserKeyOnly:
		return c.routeUserKeyOnly(ctx, request)
	case RouteCostOptimized:
		return c.routeCostOptimized(ctx, request)
	case RouteBalanced:
		return c.routeBalanced(ctx, request)
	default:
		return c.routeUserKeyFirst(ctx, request)
	}
}

// routeUserKeyFirst сначала пытается использовать пользовательские ключи
func (c *HybridAIClient) routeUserKeyFirst(ctx context.Context, request *HybridChatRequest) (*HybridChatResponse, error) {
	provider := c.getProviderFromModel(request.Model)
	
	// Пытаемся использовать пользовательский ключ
	if directClient, exists := c.DirectClients[provider]; exists {
		c.Logger.Debug("Routing to user API key", "provider", provider, "model", request.Model)
		
		response, err := directClient.Chat(ctx, request)
		if err == nil {
			response.Provider = fmt.Sprintf("user_%s", provider)
			response.RoutedVia = "user_key_first"
			response.BilledTo = "user_key"
			c.updateKeyUsage(provider)
			return response, nil
		}
		
		c.Logger.Warn("User key failed, falling back to subscription", "provider", provider, "error", err)
	}

	// Fallback к подписке GRIK AI
	return c.routeToSubscription(ctx, request)
}

// routeToSubscription маршрутизирует через GRIK AI подписку
func (c *HybridAIClient) routeToSubscription(ctx context.Context, request *HybridChatRequest) (*HybridChatResponse, error) {
	c.Logger.Debug("Routing to GRIK subscription", "model", request.Model)

	// Формируем запрос к GRIK Gateway
	url := fmt.Sprintf("%s/api/auth-models/%s/chat/completions", c.GatewayURL, c.getProviderFromModel(request.Model))
	
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Заголовки для GRIK Gateway
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.GatewayToken)
	req.Header.Set("X-User-ID", c.UserID)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var response HybridChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Добавляем метаданные
	response.Provider = "grik_subscription"
	response.RoutedVia = "subscription"
	response.BilledTo = "subscription"

	return &response, nil
}

// routeUserKeyOnly только пользовательские ключи
func (c *HybridAIClient) routeUserKeyOnly(ctx context.Context, request *HybridChatRequest) (*HybridChatResponse, error) {
	provider := c.getProviderFromModel(request.Model)
	
	directClient, exists := c.DirectClients[provider]
	if !exists {
		return nil, fmt.Errorf("user API key for %s not configured", provider)
	}

	response, err := directClient.Chat(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("user key request failed: %w", err)
	}

	response.Provider = fmt.Sprintf("user_%s", provider)
	response.RoutedVia = "user_key_only"
	response.BilledTo = "user_key"
	c.updateKeyUsage(provider)

	return response, nil
}

// routeCostOptimized оптимизация по стоимости
func (c *HybridAIClient) routeCostOptimized(ctx context.Context, request *HybridChatRequest) (*HybridChatResponse, error) {
	provider := c.getProviderFromModel(request.Model)
	
	// Простая логика: пользовательские ключи всегда дешевле для пользователя
	if _, exists := c.DirectClients[provider]; exists {
		c.Logger.Debug("Cost optimization: using user key", "provider", provider)
		return c.routeUserKeyOnly(ctx, request)
	}

	// Если нет пользовательского ключа, используем подписку
	return c.routeToSubscription(ctx, request)
}

// routeBalanced балансировка нагрузки
func (c *HybridAIClient) routeBalanced(ctx context.Context, request *HybridChatRequest) (*HybridChatResponse, error) {
	provider := c.getProviderFromModel(request.Model)
	
	// Простая балансировка на основе счетчика использования
	if keyConfig, exists := c.getUserKeyConfig(provider); exists {
		// Если пользовательский ключ использовался меньше, используем его
		if keyConfig.UsageCount%2 == 0 {
			return c.routeUserKeyOnly(ctx, request)
		}
	}

	// Иначе используем подписку
	return c.routeToSubscription(ctx, request)
}

// routeToSpecificProvider принудительный роутинг к конкретному провайдеру
func (c *HybridAIClient) routeToSpecificProvider(ctx context.Context, request *HybridChatRequest, forceProvider string) (*HybridChatResponse, error) {
	if forceProvider == "grik_subscription" {
		return c.routeToSubscription(ctx, request)
	}

	// Формат: "user_openai", "user_anthropic", etc.
	if len(forceProvider) > 5 && forceProvider[:5] == "user_" {
		provider := forceProvider[5:]
		if directClient, exists := c.DirectClients[provider]; exists {
			response, err := directClient.Chat(ctx, request)
			if err != nil {
				return nil, fmt.Errorf("forced provider %s failed: %w", forceProvider, err)
			}
			response.Provider = forceProvider
			response.RoutedVia = "forced"
			response.BilledTo = "user_key"
			c.updateKeyUsage(provider)
			return response, nil
		}
		return nil, fmt.Errorf("forced provider %s not available", forceProvider)
	}

	return nil, fmt.Errorf("unknown forced provider: %s", forceProvider)
}

// Вспомогательные методы

func (c *HybridAIClient) getProviderFromModel(model string) string {
	switch {
	case model == "gpt-4" || model == "gpt-4o" || model == "gpt-3.5-turbo":
		return "openai"
	case model == "claude-3-5-sonnet" || model == "claude-3-opus":
		return "anthropic"
	case model == "deepseek-chat" || model == "deepseek-reasoner":
		return "deepseek"
	case model == "grok-beta":
		return "grok"
	default:
		return "openai" // По умолчанию
	}
}

func (c *HybridAIClient) getUserKeyConfig(provider string) (*APIKeyConfig, bool) {
	if c.UserAPIKeys == nil {
		return nil, false
	}

	switch provider {
	case "openai":
		return c.UserAPIKeys.OpenAI, c.UserAPIKeys.OpenAI != nil
	case "anthropic":
		return c.UserAPIKeys.Anthropic, c.UserAPIKeys.Anthropic != nil
	case "deepseek":
		return c.UserAPIKeys.DeepSeek, c.UserAPIKeys.DeepSeek != nil
	case "grok":
		return c.UserAPIKeys.Grok, c.UserAPIKeys.Grok != nil
	default:
		return nil, false
	}
}

func (c *HybridAIClient) updateKeyUsage(provider string) {
	keyConfig, exists := c.getUserKeyConfig(provider)
	if exists {
		keyConfig.UsageCount++
		now := time.Now()
		keyConfig.LastUsedAt = &now
		c.Logger.Debug("Updated key usage", "provider", provider, "count", keyConfig.UsageCount)
	}
}

// GetAvailableModels возвращает доступные модели для пользователя
func (c *HybridAIClient) GetAvailableModels() *AvailableModels {
	models := &AvailableModels{
		Subscription: []ModelInfo{
			{Name: "gpt-4", Provider: "grik_subscription", Available: true},
			{Name: "gpt-4o", Provider: "grik_subscription", Available: true},
			{Name: "claude-3-5-sonnet", Provider: "grik_subscription", Available: true},
			{Name: "claude-3-opus", Provider: "grik_subscription", Available: true},
			{Name: "deepseek-chat", Provider: "grik_subscription", Available: true},
			{Name: "deepseek-reasoner", Provider: "grik_subscription", Available: true},
			{Name: "grok-beta", Provider: "grik_subscription", Available: true},
		},
		UserKeys: []ModelInfo{},
	}

	// Добавляем модели для пользовательских ключей
	for provider, client := range c.DirectClients {
		for _, model := range client.GetModels() {
			models.UserKeys = append(models.UserKeys, ModelInfo{
				Name:      model,
				Provider:  fmt.Sprintf("user_%s", provider),
				Available: true,
			})
		}
	}

	return models
}

// AvailableModels доступные модели
type AvailableModels struct {
	Subscription []ModelInfo `json:"subscription"`
	UserKeys     []ModelInfo `json:"user_keys"`
}

// ModelInfo информация о модели
type ModelInfo struct {
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	Available bool   `json:"available"`
}

// ValidateUserKeys проверяет валидность пользовательских ключей
func (c *HybridAIClient) ValidateUserKeys(ctx context.Context) map[string]error {
	results := make(map[string]error)

	for provider, client := range c.DirectClients {
		err := client.ValidateKey()
		results[provider] = err
		if err != nil {
			c.Logger.Error("User key validation failed", err, "provider", provider)
		} else {
			c.Logger.Debug("User key validation successful", "provider", provider)
		}
	}

	return results
}

// UpdateUserAPIKeys обновляет пользовательские API ключи
func (c *HybridAIClient) UpdateUserAPIKeys(newKeys *UserAPIKeys) {
	c.UserAPIKeys = newKeys
	c.DirectClients = make(map[string]DirectAIClient)
	c.initializeDirectClients()
	c.Logger.Info("User API keys updated", "user_id", c.UserID)
}

// GetUsageStats возвращает статистику использования
func (c *HybridAIClient) GetUsageStats() *UsageStats {
	stats := &UsageStats{
		UserKeys:     make(map[string]KeyUsageStats),
		Subscription: SubscriptionUsageStats{},
	}

	// Статистика пользовательских ключей
	if c.UserAPIKeys != nil {
		providers := map[string]*APIKeyConfig{
			"openai":    c.UserAPIKeys.OpenAI,
			"anthropic": c.UserAPIKeys.Anthropic,
			"deepseek":  c.UserAPIKeys.DeepSeek,
			"grok":      c.UserAPIKeys.Grok,
		}

		for provider, keyConfig := range providers {
			if keyConfig != nil {
				stats.UserKeys[provider] = KeyUsageStats{
					UsageCount: keyConfig.UsageCount,
					LastUsedAt: keyConfig.LastUsedAt,
					Enabled:    keyConfig.Enabled,
				}
			}
		}
	}

	return stats
}

// UsageStats статистика использования
type UsageStats struct {
	UserKeys     map[string]KeyUsageStats `json:"user_keys"`
	Subscription SubscriptionUsageStats   `json:"subscription"`
}

// KeyUsageStats статистика использования ключа
type KeyUsageStats struct {
	UsageCount int        `json:"usage_count"`
	LastUsedAt *time.Time `json:"last_used_at"`
	Enabled    bool       `json:"enabled"`
}

// SubscriptionUsageStats статистика подписки
type SubscriptionUsageStats struct {
	RequestsThisMonth int     `json:"requests_this_month"`
	CostThisMonth     float64 `json:"cost_this_month"`
	// TODO: Получать из GRIK системы
}