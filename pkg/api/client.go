package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Provider определяет тип провайдера API
type Provider string

const (
	ProviderOpenAI   Provider = "openai"
	ProviderClaude   Provider = "claude"
	ProviderDeepSeek Provider = "deepseek"
	ProviderGrok     Provider = "grok"
	ProviderMistral  Provider = "mistral"
	ProviderLlama    Provider = "llama"
)

// ChatRole определяет роль в диалоге
type ChatRole string

const (
	ChatRoleSystem    ChatRole = "system"
	ChatRoleUser      ChatRole = "user"
	ChatRoleAssistant ChatRole = "assistant"
)

// ChatMessage представляет сообщение в диалоге
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest представляет запрос к API чата
type ChatRequest struct {
	Model            string        `json:"model"`
	Messages         []ChatMessage `json:"messages"`
	MaxTokens        int           `json:"max_tokens,omitempty"`
	Temperature      float64       `json:"temperature,omitempty"`
	TopP             float64       `json:"top_p,omitempty"`
	FrequencyPenalty float64       `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64       `json:"presence_penalty,omitempty"`
	Stop             []string      `json:"stop,omitempty"`
}

// ChatResponse представляет ответ от API чата
type ChatResponse struct {
	Message      ChatMessage `json:"message"`
	TokensUsed   int         `json:"tokens_used"`
	Model        string      `json:"model"`
	CreatedAt    time.Time   `json:"created_at"`
	FinishedAt   time.Time   `json:"finished_at"`
	TotalTokens  int         `json:"total_tokens"`
	PromptTokens int         `json:"prompt_tokens"`
}

// Client представляет клиент для работы с API
type Client struct {
	httpClient  *http.Client
	openAIKey   string
	claudeKey   string
	deepseekKey string
	mistralKey  string
	grokKey     string
	baseURLs    map[Provider]string
}

// NewClient создает новый клиент API
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: time.Second * 120,
		},
		baseURLs: map[Provider]string{
			ProviderOpenAI:   "https://api.openai.com/v1",
			ProviderClaude:   "https://api.claude.com/v1",
			ProviderDeepSeek: "https://api.deepseek.com/v1",
			ProviderGrok:     "https://api.grok.ai/v1",
			ProviderMistral:  "https://api.mistral.ai/v1",
			ProviderLlama:    "https://api.llama.ai/v1",
		},
	}
}

// SetAPIKey устанавливает API-ключ для указанного провайдера
func (c *Client) SetAPIKey(provider Provider, key string) {
	switch provider {
	case ProviderOpenAI:
		c.openAIKey = key
	case ProviderClaude:
		c.claudeKey = key
	case ProviderDeepSeek:
		c.deepseekKey = key
	case ProviderGrok:
		c.grokKey = key
	case ProviderMistral:
		c.mistralKey = key
	case ProviderLlama:
		c.grokKey = key // Assuming the same key for llama
	}
}

// SetBaseURL устанавливает базовый URL для указанного провайдера
func (c *Client) SetBaseURL(provider Provider, url string) {
	c.baseURLs[provider] = url
}

// GetAPIKey возвращает API-ключ для указанного провайдера
func (c *Client) GetAPIKey(provider Provider) string {
	switch provider {
	case ProviderOpenAI:
		return c.openAIKey
	case ProviderClaude:
		return c.claudeKey
	case ProviderDeepSeek:
		return c.deepseekKey
	case ProviderGrok:
		return c.grokKey
	case ProviderMistral:
		return c.mistralKey
	case ProviderLlama:
		return c.grokKey
	default:
		return ""
	}
}

// ChatService представляет сервис для работы с API чата
type ChatService struct {
	client *Client
}

// NewChatService создает новый сервис для работы с API чата
func NewChatService(client *Client) *ChatService {
	return &ChatService{
		client: client,
	}
}

// getProviderFromModel возвращает провайдера для указанной модели
func getProviderFromModel(model string) Provider {
	model = strings.ToLower(model)

	if strings.HasPrefix(model, "gpt") {
		return ProviderOpenAI
	} else if strings.HasPrefix(model, "claude") {
		return ProviderClaude
	} else if strings.HasPrefix(model, "deepseek") {
		return ProviderDeepSeek
	} else if strings.HasPrefix(model, "mistral") {
		return ProviderMistral
	} else if strings.HasPrefix(model, "grok") {
		return ProviderGrok
	} else if strings.HasPrefix(model, "llama") {
		return ProviderLlama
	}

	// По умолчанию предполагаем OpenAI
	return ProviderOpenAI
}

// SendMessage отправляет запрос к API чата
func (s *ChatService) SendMessage(req *ChatRequest) (*ChatResponse, error) {
	// Определяем провайдера на основе модели
	provider := getProviderFromModel(req.Model)

	// Получаем API-ключ для провайдера
	apiKey := s.client.GetAPIKey(provider)
	if apiKey == "" {
		return nil, fmt.Errorf("API key for provider %s is not set", provider)
	}

	// Отправляем запрос к соответствующему API
	switch provider {
	case ProviderOpenAI:
		return s.sendOpenAIRequest(req, apiKey)
	case ProviderClaude:
		return s.sendClaudeRequest(req, apiKey)
	case ProviderDeepSeek:
		// Пока используем временную заглушку
		return s.sendGenericRequest(req, apiKey, provider)
	case ProviderMistral:
		// Пока используем временную заглушку
		return s.sendGenericRequest(req, apiKey, provider)
	case ProviderGrok:
		// Пока используем временную заглушку
		return s.sendGenericRequest(req, apiKey, provider)
	case ProviderLlama:
		// Пока используем временную заглушку
		return s.sendGenericRequest(req, apiKey, provider)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// sendOpenAIRequest отправляет запрос к API OpenAI
func (s *ChatService) sendOpenAIRequest(req *ChatRequest, apiKey string) (*ChatResponse, error) {
	// Конвертируем запрос в формат OpenAI
	openAIReq := map[string]interface{}{
		"model":             req.Model,
		"messages":          req.Messages,
		"max_tokens":        req.MaxTokens,
		"temperature":       req.Temperature,
		"top_p":             req.TopP,
		"frequency_penalty": req.FrequencyPenalty,
		"presence_penalty":  req.PresencePenalty,
		"stop":              req.Stop,
	}

	// Сериализуем запрос
	reqBody, err := json.Marshal(openAIReq)
	if err != nil {
		return nil, err
	}

	// Создаем HTTP-запрос
	httpReq, err := http.NewRequest("POST", s.client.baseURLs[ProviderOpenAI]+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	// Устанавливаем заголовки
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	// Отправляем запрос
	resp, err := s.client.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Читаем ответ
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Проверяем статус-код
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: %s", string(respBody))
	}

	// Парсим ответ
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens      int `json:"total_tokens"`
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}

	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return nil, err
	}

	// Проверяем, что есть хотя бы один ответ
	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("OpenAI API returned empty response")
	}

	// Формируем ответ в универсальном формате
	return &ChatResponse{
		Message: ChatMessage{
			Role:    openAIResp.Choices[0].Message.Role,
			Content: openAIResp.Choices[0].Message.Content,
		},
		TokensUsed:   openAIResp.Usage.CompletionTokens,
		Model:        openAIResp.Model,
		CreatedAt:    time.Now(),
		FinishedAt:   time.Now(),
		TotalTokens:  openAIResp.Usage.TotalTokens,
		PromptTokens: openAIResp.Usage.PromptTokens,
	}, nil
}

// sendClaudeRequest отправляет запрос к API Claude
func (s *ChatService) sendClaudeRequest(req *ChatRequest, apiKey string) (*ChatResponse, error) {
	// Адаптируем формат сообщений для Claude
	var systemPrompt string
	var userMessages []map[string]string

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else {
			userMessages = append(userMessages, map[string]string{
				"role":    msg.Role,
				"content": msg.Content,
			})
		}
	}

	// Формируем запрос в формате Claude
	claudeReq := map[string]interface{}{
		"model":          req.Model,
		"messages":       userMessages,
		"max_tokens":     req.MaxTokens,
		"temperature":    req.Temperature,
		"top_p":          req.TopP,
		"stop_sequences": req.Stop,
	}

	// Если есть системный промпт, добавляем его
	if systemPrompt != "" {
		claudeReq["system"] = systemPrompt
	}

	// Сериализуем запрос
	reqBody, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, err
	}

	// Создаем HTTP-запрос
	httpReq, err := http.NewRequest("POST", s.client.baseURLs[ProviderClaude]+"/messages", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	// Устанавливаем заголовки
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Отправляем запрос
	resp, err := s.client.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Читаем ответ
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Проверяем статус-код
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Claude API error: %s", string(respBody))
	}

	// Парсим ответ
	var claudeResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Model string `json:"model"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return nil, err
	}

	// Проверяем, что есть хотя бы один ответ
	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("Claude API returned empty response")
	}

	// Извлекаем текст ответа
	var content string
	for _, c := range claudeResp.Content {
		if c.Type == "text" {
			content = c.Text
			break
		}
	}

	// Формируем ответ в универсальном формате
	return &ChatResponse{
		Message: ChatMessage{
			Role:    "assistant",
			Content: content,
		},
		TokensUsed:   claudeResp.Usage.OutputTokens,
		Model:        claudeResp.Model,
		CreatedAt:    time.Now(),
		FinishedAt:   time.Now(),
		TotalTokens:  claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		PromptTokens: claudeResp.Usage.InputTokens,
	}, nil
}

// sendGenericRequest отправляет запрос к API в общем формате (временная заглушка)
func (s *ChatService) sendGenericRequest(req *ChatRequest, _ string, provider Provider) (*ChatResponse, error) {
	return &ChatResponse{
		Message: ChatMessage{
			Role:    "assistant",
			Content: fmt.Sprintf("Поддержка API для провайдера %s будет реализована в ближайшее время.", provider),
		},
		TokensUsed:   0,
		Model:        req.Model,
		CreatedAt:    time.Now(),
		FinishedAt:   time.Now(),
		TotalTokens:  0,
		PromptTokens: 0,
	}, nil
}

// TODO: Реализовать методы для других провайдеров
// sendDeepseekRequest, sendMistralRequest, sendGrokRequest

// ModelService представляет сервис для работы с моделями
type ModelService struct {
	client *Client
}

// NewModelService создает новый сервис для работы с моделями
func NewModelService(client *Client) *ModelService {
	return &ModelService{
		client: client,
	}
}

// Model представляет информацию о модели
type Model struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Provider  Provider `json:"provider"`
	Context   int      `json:"context"`
	MaxTokens int      `json:"max_tokens"`
	Tags      []string `json:"tags"`
}

// ListModels возвращает список доступных моделей
func (s *ModelService) ListModels() ([]Model, error) {
	// Здесь можно реализовать запрос к API для получения списка моделей
	// или использовать предопределенный список

	// Пока возвращаем предопределенный список
	return []Model{
		{
			ID:        "gpt-3.5-turbo",
			Name:      "GPT-3.5 Turbo",
			Provider:  ProviderOpenAI,
			Context:   4096,
			MaxTokens: 4096,
			Tags:      []string{"cheap", "fast"},
		},
		{
			ID:        "gpt-4",
			Name:      "GPT-4",
			Provider:  ProviderOpenAI,
			Context:   8192,
			MaxTokens: 4096,
			Tags:      []string{"premium", "reasoning"},
		},
		{
			ID:        "claude-3-haiku",
			Name:      "Claude 3 Haiku",
			Provider:  ProviderClaude,
			Context:   200000,
			MaxTokens: 4096,
			Tags:      []string{"fast", "reasoning", "long-context"},
		},
		{
			ID:        "claude-3-sonnet",
			Name:      "Claude 3 Sonnet",
			Provider:  ProviderClaude,
			Context:   200000,
			MaxTokens: 4096,
			Tags:      []string{"premium", "reasoning", "long-context"},
		},
		{
			ID:        "claude-3-opus",
			Name:      "Claude 3 Opus",
			Provider:  ProviderClaude,
			Context:   200000,
			MaxTokens: 4096,
			Tags:      []string{"premium", "reasoning", "long-context", "expert"},
		},
	}, nil
}

// GetModel возвращает информацию о модели по ID
func (s *ModelService) GetModel(id string) (*Model, error) {
	models, err := s.ListModels()
	if err != nil {
		return nil, err
	}

	for _, model := range models {
		if model.ID == id {
			return &model, nil
		}
	}

	return nil, fmt.Errorf("model with ID '%s' not found", id)
}

// ModelSelector представляет интерактивный селектор моделей
type ModelSelector struct {
	client *Client
}

// NewModelSelector создает новый селектор моделей
func NewModelSelector(client *Client) *ModelSelector {
	return &ModelSelector{
		client: client,
	}
}

// SelectModel выбирает модель на основе роли и параметров
func (s *ModelSelector) SelectModel(role string, tags []string) (*Model, error) {
	// Получаем список доступных моделей
	modelService := NewModelService(s.client)
	models, err := modelService.ListModels()
	if err != nil {
		return nil, err
	}

	// Фильтруем модели по тегам
	var filteredModels []Model
	if len(tags) > 0 {
		for _, model := range models {
			// Проверяем, содержит ли модель все указанные теги
			containsAllTags := true
			for _, tag := range tags {
				found := false
				for _, modelTag := range model.Tags {
					if modelTag == tag {
						found = true
						break
					}
				}
				if !found {
					containsAllTags = false
					break
				}
			}
			if containsAllTags {
				filteredModels = append(filteredModels, model)
			}
		}
	} else {
		filteredModels = models
	}

	// Если нет подходящих моделей, возвращаем ошибку
	if len(filteredModels) == 0 {
		return nil, fmt.Errorf("no models found matching role '%s' and tags %v", role, tags)
	}

	// Пока просто возвращаем первую подходящую модель
	// TODO: Реализовать интерактивный выбор модели
	return &filteredModels[0], nil
}
