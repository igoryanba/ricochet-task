package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/chain"
)

const (
	defaultOpenAIAPIBaseURL = "https://api.openai.com/v1"
	defaultOpenAITimeout    = 90 * time.Second
)

// OpenAIProvider провайдер для моделей OpenAI
type OpenAIProvider struct {
	*BaseProvider
	client *http.Client
}

// OpenAIRequest запрос к API OpenAI
type OpenAIRequest struct {
	Model       string                   `json:"model"`
	Messages    []OpenAIMessage          `json:"messages"`
	Temperature float64                  `json:"temperature"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
	TopP        float64                  `json:"top_p,omitempty"`
	FreqPenalty float64                  `json:"frequency_penalty,omitempty"`
	PresPenalty float64                  `json:"presence_penalty,omitempty"`
	Stop        []string                 `json:"stop,omitempty"`
	Stream      bool                     `json:"stream,omitempty"`
	Tools       []map[string]interface{} `json:"tools,omitempty"`
}

// OpenAIMessage сообщение в формате OpenAI
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse ответ от API OpenAI
type OpenAIResponse struct {
	ID      string      `json:"id"`
	Object  string      `json:"object"`
	Created int64       `json:"created"`
	Model   string      `json:"model"`
	Choices []Choice    `json:"choices"`
	Usage   Usage       `json:"usage"`
	Error   OpenAIError `json:"error,omitempty"`
}

// Choice выбор модели
type Choice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// Usage информация об использовании токенов
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIError ошибка API OpenAI
type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    string `json:"code"`
}

// NewOpenAIProvider создает новый провайдер для OpenAI
func NewOpenAIProvider(apiKey string, apiBaseURL string) *OpenAIProvider {
	if apiBaseURL == "" {
		apiBaseURL = defaultOpenAIAPIBaseURL
	}

	provider := &OpenAIProvider{
		BaseProvider: NewBaseProvider(chain.ModelTypeOpenAI, apiKey, apiBaseURL),
		client: &http.Client{
			Timeout: defaultOpenAITimeout,
		},
	}

	// Регистрируем поддерживаемые модели
	provider.RegisterModels([]chain.ModelConfiguration{
		{
			Name:      chain.ModelNameGPT35Turbo,
			Type:      chain.ModelTypeOpenAI,
			Context:   16385,
			MaxTokens: 4096,
			Version:   "2023-09-01",
			Provider:  "OpenAI",
			Endpoint:  "/chat/completions",
		},
		{
			Name:      chain.ModelNameGPT35Turbo16k,
			Type:      chain.ModelTypeOpenAI,
			Context:   16385,
			MaxTokens: 8192,
			Version:   "2023-09-01",
			Provider:  "OpenAI",
			Endpoint:  "/chat/completions",
		},
		{
			Name:      chain.ModelNameGPT4,
			Type:      chain.ModelTypeOpenAI,
			Context:   8192,
			MaxTokens: 4096,
			Version:   "2023-09-01",
			Provider:  "OpenAI",
			Endpoint:  "/chat/completions",
		},
		{
			Name:      chain.ModelNameGPT4Turbo,
			Type:      chain.ModelTypeOpenAI,
			Context:   128000,
			MaxTokens: 4096,
			Version:   "2023-12-01",
			Provider:  "OpenAI",
			Endpoint:  "/chat/completions",
		},
		{
			Name:      chain.ModelNameGPT432k,
			Type:      chain.ModelTypeOpenAI,
			Context:   32768,
			MaxTokens: 4096,
			Version:   "2023-09-01",
			Provider:  "OpenAI",
			Endpoint:  "/chat/completions",
		},
	})

	return provider
}

// Execute выполняет запрос к модели OpenAI
func (p *OpenAIProvider) Execute(ctx context.Context, model chain.Model, prompt string, options map[string]interface{}) (string, error) {
	// Проверяем API-ключ
	if err := p.ValidateAPIKey(); err != nil {
		return "", err
	}

	// Получаем конфигурацию модели
	modelConfig, err := p.GetModel(model.Name)
	if err != nil {
		return "", err
	}

	// Создаем запрос
	messages := []OpenAIMessage{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// Добавляем системный промпт, если указан
	if systemPrompt, ok := options["system_prompt"].(string); ok && systemPrompt != "" {
		messages = append([]OpenAIMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
		}, messages...)
	}

	// Параметры запроса
	temperature := model.Temperature
	if temperature <= 0 {
		temperature = 0.7
	}

	maxTokens := model.MaxTokens
	if maxTokens <= 0 {
		maxTokens = modelConfig.MaxTokens / 2
	}

	// Формируем запрос
	request := OpenAIRequest{
		Model:       string(model.Name),
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	// Дополнительные параметры
	if topP, ok := options["top_p"].(float64); ok {
		request.TopP = topP
	}

	if freqPenalty, ok := options["frequency_penalty"].(float64); ok {
		request.FreqPenalty = freqPenalty
	}

	if presPenalty, ok := options["presence_penalty"].(float64); ok {
		request.PresPenalty = presPenalty
	}

	if stop, ok := options["stop"].([]string); ok {
		request.Stop = stop
	}

	// Кодируем запрос в JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Создаем HTTP-запрос
	endpoint := fmt.Sprintf("%s%s", p.apiBaseURL, modelConfig.Endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))

	// Выполняем запрос
	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Проверяем статус-код
	if resp.StatusCode != http.StatusOK {
		var errorResp OpenAIResponse
		if err := json.Unmarshal(responseBody, &errorResp); err == nil && errorResp.Error.Message != "" {
			return "", fmt.Errorf("API error: %s", errorResp.Error.Message)
		}
		return "", fmt.Errorf("API error: %s", resp.Status)
	}

	// Разбираем ответ
	var response OpenAIResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Проверяем наличие ответа
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from model")
	}

	return response.Choices[0].Message.Content, nil
}

// EstimateTokens переопределяет метод базового провайдера для лучшей оценки
func (p *OpenAIProvider) EstimateTokens(text string) int {
	estimator := NewTokenEstimator()
	return estimator.EstimateTokens(text, "")
}
