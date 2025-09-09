package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AIProvider represents different AI service providers
type AIProvider string

const (
	OpenAI    AIProvider = "openai"
	Anthropic AIProvider = "anthropic"
	DeepSeek  AIProvider = "deepseek"
	Grok      AIProvider = "grok"
)

// AIClient manages communication with AI model services
type AIClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Provider   AIProvider
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// NewAIClient creates a new AI client for the specified provider
func NewAIClient(provider AIProvider) *AIClient {
	var baseURL string
	
	switch provider {
	case OpenAI:
		baseURL = "http://localhost:6000"
	case Anthropic:
		baseURL = "http://localhost:6001"
	case DeepSeek:
		baseURL = "http://localhost:6002"
	case Grok:
		baseURL = "http://localhost:6003"
	default:
		baseURL = "http://localhost:6000" // Default to OpenAI
	}

	return &AIClient{
		BaseURL:  baseURL,
		Provider: provider,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Chat sends a chat completion request to the AI service
func (c *AIClient) Chat(request *ChatRequest) (*ChatResponse, error) {
	// Ensure model is set based on provider if not specified
	if request.Model == "" {
		request.Model = c.getDefaultModel()
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.BaseURL)
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !chatResp.Success && chatResp.Error != nil {
		return nil, fmt.Errorf("AI service error: %s", chatResp.Error.Message)
	}

	return &chatResp, nil
}

// getDefaultModel returns the default model for the provider
func (c *AIClient) getDefaultModel() string {
	switch c.Provider {
	case OpenAI:
		return "gpt-4o"
	case Anthropic:
		return "claude-3-5-sonnet-20241022"
	case DeepSeek:
		return "deepseek-chat"
	case Grok:
		return "grok-beta"
	default:
		return "gpt-4o"
	}
}

// HealthCheck checks if the AI service is available
func (c *AIClient) HealthCheck() error {
	url := fmt.Sprintf("%s/health", c.BaseURL)
	
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// AIManager manages multiple AI clients with fallback support
type AIManager struct {
	Clients       map[AIProvider]*AIClient
	PrimaryClient AIProvider
	Fallbacks     []AIProvider
}

// NewAIManager creates a new AI manager with multiple providers
func NewAIManager() *AIManager {
	clients := make(map[AIProvider]*AIClient)
	
	// Initialize all available clients
	for _, provider := range []AIProvider{OpenAI, Anthropic, DeepSeek, Grok} {
		clients[provider] = NewAIClient(provider)
	}

	return &AIManager{
		Clients:       clients,
		PrimaryClient: OpenAI, // Default primary
		Fallbacks:     []AIProvider{Anthropic, DeepSeek, Grok},
	}
}

// SetPrimaryProvider sets the primary AI provider
func (m *AIManager) SetPrimaryProvider(provider AIProvider) {
	m.PrimaryClient = provider
}

// Chat sends a chat request with automatic fallback support
func (m *AIManager) Chat(request *ChatRequest) (*ChatResponse, error) {
	// Try primary client first
	if client, exists := m.Clients[m.PrimaryClient]; exists {
		resp, err := client.Chat(request)
		if err == nil {
			return resp, nil
		}
		// Log the error but continue with fallbacks
		fmt.Printf("Primary provider %s failed: %v\n", m.PrimaryClient, err)
	}

	// Try fallback providers
	for _, provider := range m.Fallbacks {
		if provider == m.PrimaryClient {
			continue // Skip primary as we already tried it
		}
		
		if client, exists := m.Clients[provider]; exists {
			resp, err := client.Chat(request)
			if err == nil {
				fmt.Printf("Fallback provider %s succeeded\n", provider)
				return resp, nil
			}
			fmt.Printf("Fallback provider %s failed: %v\n", provider, err)
		}
	}

	return nil, fmt.Errorf("all AI providers failed")
}

// HealthCheckAll checks health of all AI services
func (m *AIManager) HealthCheckAll() map[AIProvider]error {
	results := make(map[AIProvider]error)
	
	for provider, client := range m.Clients {
		results[provider] = client.HealthCheck()
	}
	
	return results
}