package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// OpenAIDirectClient прямой клиент для OpenAI API
type OpenAIDirectClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	Logger     Logger
}

// AnthropicDirectClient прямой клиент для Anthropic API
type AnthropicDirectClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	Logger     Logger
}

// DeepSeekDirectClient прямой клиент для DeepSeek API
type DeepSeekDirectClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	Logger     Logger
}

// GrokDirectClient прямой клиент для Grok API
type GrokDirectClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	Logger     Logger
}

// NewOpenAIDirectClient создает новый OpenAI клиент
func NewOpenAIDirectClient(config *APIKeyConfig, logger Logger) DirectAIClient {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &OpenAIDirectClient{
		APIKey:     config.APIKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
		Logger:     logger,
	}
}

// NewAnthropicDirectClient создает новый Anthropic клиент
func NewAnthropicDirectClient(config *APIKeyConfig, logger Logger) DirectAIClient {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	return &AnthropicDirectClient{
		APIKey:     config.APIKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
		Logger:     logger,
	}
}

// NewDeepSeekDirectClient создает новый DeepSeek клиент
func NewDeepSeekDirectClient(config *APIKeyConfig, logger Logger) DirectAIClient {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}

	return &DeepSeekDirectClient{
		APIKey:     config.APIKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
		Logger:     logger,
	}
}

// NewGrokDirectClient создает новый Grok клиент
func NewGrokDirectClient(config *APIKeyConfig, logger Logger) DirectAIClient {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.x.ai/v1"
	}

	return &GrokDirectClient{
		APIKey:     config.APIKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
		Logger:     logger,
	}
}

// OpenAI Direct Client Implementation

func (c *OpenAIDirectClient) Chat(ctx context.Context, request *HybridChatRequest) (*HybridChatResponse, error) {
	url := fmt.Sprintf("%s/chat/completions", c.BaseURL)

	// Формируем запрос в формате OpenAI
	openaiRequest := map[string]interface{}{
		"model":    request.Model,
		"messages": request.Messages,
	}

	if request.Temperature > 0 {
		openaiRequest["temperature"] = request.Temperature
	}
	if request.MaxTokens > 0 {
		openaiRequest["max_tokens"] = request.MaxTokens
	}
	if request.Stream {
		openaiRequest["stream"] = request.Stream
	}

	reqBody, err := json.Marshal(openaiRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OpenAI request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenAI request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: %d", resp.StatusCode)
	}

	var openaiResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("failed to decode OpenAI response: %w", err)
	}

	// Преобразуем в наш формат
	return c.convertOpenAIResponse(openaiResp), nil
}

func (c *OpenAIDirectClient) GetModels() []string {
	return []string{"gpt-4", "gpt-4o", "gpt-3.5-turbo", "gpt-4-turbo"}
}

func (c *OpenAIDirectClient) ValidateKey() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/models", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("validation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid OpenAI API key")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OpenAI API validation error: %d", resp.StatusCode)
	}

	return nil
}

func (c *OpenAIDirectClient) convertOpenAIResponse(resp map[string]interface{}) *HybridChatResponse {
	response := &HybridChatResponse{
		Model:   getString(resp, "model"),
		Created: int64(getFloat64(resp, "created")),
	}

	if id, ok := resp["id"].(string); ok {
		response.ID = id
	}

	// Конвертируем choices
	if choices, ok := resp["choices"].([]interface{}); ok {
		for _, choice := range choices {
			if choiceMap, ok := choice.(map[string]interface{}); ok {
				if message, ok := choiceMap["message"].(map[string]interface{}); ok {
					response.Choices = append(response.Choices, Choice{
						Message: Message{
							Role:    getString(message, "role"),
							Content: getString(message, "content"),
						},
					})
				}
			}
		}
	}

	// Конвертируем usage
	if usage, ok := resp["usage"].(map[string]interface{}); ok {
		response.Usage = Usage{
			PromptTokens:     int(getFloat64(usage, "prompt_tokens")),
			CompletionTokens: int(getFloat64(usage, "completion_tokens")),
			TotalTokens:      int(getFloat64(usage, "total_tokens")),
		}
	}

	return response
}

// Anthropic Direct Client Implementation

func (c *AnthropicDirectClient) Chat(ctx context.Context, request *HybridChatRequest) (*HybridChatResponse, error) {
	url := fmt.Sprintf("%s/v1/messages", c.BaseURL)

	// Формируем запрос в формате Anthropic
	anthropicRequest := map[string]interface{}{
		"model":      request.Model,
		"messages":   request.Messages,
		"max_tokens": 4000, // Anthropic требует max_tokens
	}

	if request.MaxTokens > 0 {
		anthropicRequest["max_tokens"] = request.MaxTokens
	}
	if request.Temperature > 0 {
		anthropicRequest["temperature"] = request.Temperature
	}

	reqBody, err := json.Marshal(anthropicRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Anthropic request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create Anthropic request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Anthropic request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Anthropic API error: %d", resp.StatusCode)
	}

	var anthropicResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to decode Anthropic response: %w", err)
	}

	return c.convertAnthropicResponse(anthropicResp), nil
}

func (c *AnthropicDirectClient) GetModels() []string {
	return []string{"claude-3-5-sonnet-20241022", "claude-3-opus-20240229", "claude-3-sonnet-20240229", "claude-3-haiku-20240307"}
}

func (c *AnthropicDirectClient) ValidateKey() error {
	// Anthropic не имеет простого эндпоинта для валидации, используем тестовый запрос
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	testRequest := &HybridChatRequest{
		Model:     "claude-3-haiku-20240307",
		Messages:  []Message{{Role: "user", Content: "Hi"}},
		MaxTokens: 10,
	}

	_, err := c.Chat(ctx, testRequest)
	if err != nil && strings.Contains(err.Error(), "401") {
		return fmt.Errorf("invalid Anthropic API key")
	}

	return nil // Считаем ключ валидным если не получили 401
}

func (c *AnthropicDirectClient) convertAnthropicResponse(resp map[string]interface{}) *HybridChatResponse {
	response := &HybridChatResponse{
		Model:   getString(resp, "model"),
		Created: time.Now().Unix(),
	}

	if id, ok := resp["id"].(string); ok {
		response.ID = id
	}

	// Anthropic возвращает content как массив
	if content, ok := resp["content"].([]interface{}); ok {
		for _, item := range content {
			if textItem, ok := item.(map[string]interface{}); ok {
				if textItem["type"] == "text" {
					response.Choices = append(response.Choices, Choice{
						Message: Message{
							Role:    "assistant",
							Content: getString(textItem, "text"),
						},
					})
				}
			}
		}
	}

	// Anthropic usage
	if usage, ok := resp["usage"].(map[string]interface{}); ok {
		response.Usage = Usage{
			PromptTokens:     int(getFloat64(usage, "input_tokens")),
			CompletionTokens: int(getFloat64(usage, "output_tokens")),
			TotalTokens:      int(getFloat64(usage, "input_tokens")) + int(getFloat64(usage, "output_tokens")),
		}
	}

	return response
}

// DeepSeek Direct Client Implementation

func (c *DeepSeekDirectClient) Chat(ctx context.Context, request *HybridChatRequest) (*HybridChatResponse, error) {
	url := fmt.Sprintf("%s/chat/completions", c.BaseURL)

	// DeepSeek использует OpenAI-совместимый формат
	deepseekRequest := map[string]interface{}{
		"model":    request.Model,
		"messages": request.Messages,
	}

	if request.Temperature > 0 {
		deepseekRequest["temperature"] = request.Temperature
	}
	if request.MaxTokens > 0 {
		deepseekRequest["max_tokens"] = request.MaxTokens
	}
	if request.Stream {
		deepseekRequest["stream"] = request.Stream
	}

	reqBody, err := json.Marshal(deepseekRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DeepSeek request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create DeepSeek request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("DeepSeek request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DeepSeek API error: %d", resp.StatusCode)
	}

	var deepseekResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&deepseekResp); err != nil {
		return nil, fmt.Errorf("failed to decode DeepSeek response: %w", err)
	}

	return c.convertDeepSeekResponse(deepseekResp), nil
}

func (c *DeepSeekDirectClient) GetModels() []string {
	return []string{"deepseek-chat", "deepseek-coder", "deepseek-reasoner"}
}

func (c *DeepSeekDirectClient) ValidateKey() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/models", c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("validation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid DeepSeek API key")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("DeepSeek API validation error: %d", resp.StatusCode)
	}

	return nil
}

func (c *DeepSeekDirectClient) convertDeepSeekResponse(resp map[string]interface{}) *HybridChatResponse {
	// DeepSeek использует OpenAI-совместимый формат
	return (&OpenAIDirectClient{}).convertOpenAIResponse(resp)
}

// Grok Direct Client Implementation

func (c *GrokDirectClient) Chat(ctx context.Context, request *HybridChatRequest) (*HybridChatResponse, error) {
	url := fmt.Sprintf("%s/chat/completions", c.BaseURL)

	// Grok использует OpenAI-совместимый формат
	grokRequest := map[string]interface{}{
		"model":    request.Model,
		"messages": request.Messages,
	}

	if request.Temperature > 0 {
		grokRequest["temperature"] = request.Temperature
	}
	if request.MaxTokens > 0 {
		grokRequest["max_tokens"] = request.MaxTokens
	}
	if request.Stream {
		grokRequest["stream"] = request.Stream
	}

	reqBody, err := json.Marshal(grokRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Grok request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create Grok request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Grok request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Grok API error: %d", resp.StatusCode)
	}

	var grokResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&grokResp); err != nil {
		return nil, fmt.Errorf("failed to decode Grok response: %w", err)
	}

	return c.convertGrokResponse(grokResp), nil
}

func (c *GrokDirectClient) GetModels() []string {
	return []string{"grok-beta", "grok-vision-beta"}
}

func (c *GrokDirectClient) ValidateKey() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Grok может не иметь /models эндпоинта, используем тестовый запрос
	testRequest := &HybridChatRequest{
		Model:     "grok-beta",
		Messages:  []Message{{Role: "user", Content: "Hi"}},
		MaxTokens: 10,
	}

	_, err := c.Chat(ctx, testRequest)
	if err != nil && strings.Contains(err.Error(), "401") {
		return fmt.Errorf("invalid Grok API key")
	}

	return nil
}

func (c *GrokDirectClient) convertGrokResponse(resp map[string]interface{}) *HybridChatResponse {
	// Grok использует OpenAI-совместимый формат
	return (&OpenAIDirectClient{}).convertOpenAIResponse(resp)
}

// Вспомогательные функции

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getFloat64(data map[string]interface{}, key string) float64 {
	if val, ok := data[key].(float64); ok {
		return val
	}
	if val, ok := data[key].(int); ok {
		return float64(val)
	}
	return 0
}

// Общие типы для всех клиентов

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Choice struct {
	Message Message `json:"message"`
	Index   int     `json:"index"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}