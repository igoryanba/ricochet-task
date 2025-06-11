package model

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/grik-ai/ricochet-task/pkg/chain"
)

// Provider интерфейс для доступа к моделям
type Provider interface {
	// Execute выполняет запрос к модели
	Execute(ctx context.Context, model chain.Model, prompt string, options map[string]interface{}) (string, error)

	// EstimateTokens оценивает количество токенов в тексте
	EstimateTokens(text string) int

	// GetModel возвращает модель по имени
	GetModel(name chain.ModelName) (chain.ModelConfiguration, error)

	// GetAvailableModels возвращает список доступных моделей
	GetAvailableModels() []chain.ModelConfiguration

	// GetProviderType возвращает тип провайдера
	GetProviderType() chain.ModelType
}

// ProviderFactory фабрика для создания провайдеров
type ProviderFactory struct {
	providers map[chain.ModelType]Provider
}

// NewProviderFactory создает новую фабрику провайдеров
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{
		providers: make(map[chain.ModelType]Provider),
	}
}

// RegisterProvider регистрирует провайдера
func (f *ProviderFactory) RegisterProvider(provider Provider) {
	f.providers[provider.GetProviderType()] = provider
}

// GetProvider возвращает провайдера по типу
func (f *ProviderFactory) GetProvider(modelType chain.ModelType) (Provider, error) {
	provider, exists := f.providers[modelType]
	if !exists {
		return nil, fmt.Errorf("provider for model type '%s' not found", modelType)
	}
	return provider, nil
}

// GetProviderForModel возвращает провайдера для модели
func (f *ProviderFactory) GetProviderForModel(model chain.Model) (Provider, error) {
	return f.GetProvider(model.Type)
}

// BaseProvider базовая реализация провайдера
type BaseProvider struct {
	models     []chain.ModelConfiguration
	modelType  chain.ModelType
	apiKey     string
	apiBaseURL string
}

// NewBaseProvider создает базового провайдера
func NewBaseProvider(modelType chain.ModelType, apiKey string, apiBaseURL string) *BaseProvider {
	return &BaseProvider{
		models:     []chain.ModelConfiguration{},
		modelType:  modelType,
		apiKey:     apiKey,
		apiBaseURL: apiBaseURL,
	}
}

// GetProviderType возвращает тип провайдера
func (p *BaseProvider) GetProviderType() chain.ModelType {
	return p.modelType
}

// EstimateTokens оценивает количество токенов в тексте
// Это приблизительная оценка: 1 токен ~ 4 символа
func (p *BaseProvider) EstimateTokens(text string) int {
	return len(text) / 4
}

// GetModel возвращает модель по имени
func (p *BaseProvider) GetModel(name chain.ModelName) (chain.ModelConfiguration, error) {
	for _, model := range p.models {
		if model.Name == name {
			return model, nil
		}
	}
	return chain.ModelConfiguration{}, fmt.Errorf("model '%s' not found", name)
}

// GetAvailableModels возвращает список доступных моделей
func (p *BaseProvider) GetAvailableModels() []chain.ModelConfiguration {
	return p.models
}

// Execute выполняет запрос к модели
func (p *BaseProvider) Execute(ctx context.Context, model chain.Model, prompt string, options map[string]interface{}) (string, error) {
	return "", errors.New("not implemented in base provider")
}

// RegisterModels регистрирует модели
func (p *BaseProvider) RegisterModels(models []chain.ModelConfiguration) {
	p.models = models
}

// ValidateAPIKey проверяет API-ключ
func (p *BaseProvider) ValidateAPIKey() error {
	if p.apiKey == "" {
		return errors.New("API key is required")
	}
	return nil
}

// Errors
var (
	ErrAPIKeyRequired   = errors.New("API key is required")
	ErrModelNotFound    = errors.New("model not found")
	ErrProviderNotFound = errors.New("provider not found")
	ErrRequestFailed    = errors.New("request failed")
	ErrResponseParsing  = errors.New("failed to parse response")
)

// TokenEstimator оценщик токенов
type TokenEstimator struct {
	// Средняя длина токена в символах для разных языков
	avgTokenLengths map[string]float64
}

// NewTokenEstimator создает новый оценщик токенов
func NewTokenEstimator() *TokenEstimator {
	return &TokenEstimator{
		avgTokenLengths: map[string]float64{
			"en": 4.0, // Английский
			"ru": 5.0, // Русский
			"zh": 1.5, // Китайский
			"ja": 2.0, // Японский
			"ko": 2.5, // Корейский
			"de": 4.5, // Немецкий
			"fr": 4.5, // Французский
			"es": 4.2, // Испанский
			"it": 4.3, // Итальянский
			"pt": 4.2, // Португальский
			"nl": 4.5, // Голландский
			"pl": 4.8, // Польский
			"tr": 4.5, // Турецкий
			"ar": 3.5, // Арабский
			"hi": 4.0, // Хинди
		},
	}
}

// EstimateTokens оценивает количество токенов в тексте
func (e *TokenEstimator) EstimateTokens(text string, lang string) int {
	if lang == "" {
		lang = e.detectLanguage(text)
	}

	avgTokenLength, exists := e.avgTokenLengths[lang]
	if !exists {
		avgTokenLength = 4.0 // По умолчанию
	}

	// Базовая оценка по длине текста
	estimatedTokens := float64(len(text)) / avgTokenLength

	// Учитываем специальные символы, которые могут занимать больше токенов
	specialChars := strings.Count(text, "\n") + strings.Count(text, "\t") +
		strings.Count(text, "\"") + strings.Count(text, "'") +
		strings.Count(text, "(") + strings.Count(text, ")") +
		strings.Count(text, "[") + strings.Count(text, "]") +
		strings.Count(text, "{") + strings.Count(text, "}")

	// Добавляем дополнительные токены для специальных символов
	estimatedTokens += float64(specialChars) * 0.2

	return int(estimatedTokens)
}

// detectLanguage определяет язык текста по частотам символов
func (e *TokenEstimator) detectLanguage(text string) string {
	if len(text) < 10 {
		return "en" // По умолчанию для коротких текстов
	}

	// Упрощенное определение языка по частотам символов
	ruCount := 0
	enCount := 0
	otherCount := 0

	for _, r := range text {
		if r >= 'А' && r <= 'я' {
			ruCount++
		} else if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			enCount++
		} else if r > 128 {
			otherCount++
		}
	}

	if ruCount > enCount && ruCount > otherCount {
		return "ru"
	} else if otherCount > ruCount && otherCount > enCount {
		// Упрощенно считаем все неизвестные символы как другие языки
		return "zh" // По умолчанию для неизвестных языков
	}

	return "en"
}
