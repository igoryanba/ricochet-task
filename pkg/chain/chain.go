package chain

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// ModelType определяет тип модели (провайдера)
type ModelType string

const (
	ModelTypeOpenAI   ModelType = "openai"   // OpenAI (GPT-3.5, GPT-4)
	ModelTypeClaude   ModelType = "claude"   // Anthropic Claude
	ModelTypeDeepSeek ModelType = "deepseek" // DeepSeek
	ModelTypeGrok     ModelType = "grok"     // Grok
	ModelTypeLlama    ModelType = "llama"    // LLaMA (local)
	ModelTypeMistral  ModelType = "mistral"  // Mistral AI
)

// ModelRole определяет роль модели в цепочке обработки
type ModelRole string

const (
	ModelRoleAnalyzer   ModelRole = "analyzer"   // Анализирует текст
	ModelRoleSummarizer ModelRole = "summarizer" // Создает резюме
	ModelRoleIntegrator ModelRole = "integrator" // Интегрирует результаты
	ModelRoleExtractor  ModelRole = "extractor"  // Извлекает данные
	ModelRoleOrganizer  ModelRole = "organizer"  // Организует структуру
	ModelRoleEvaluator  ModelRole = "evaluator"  // Оценивает результаты
)

// ModelName определяет конкретную модель определенного провайдера
type ModelName string

// Поддерживаемые модели для каждого провайдера
const (
	// OpenAI
	ModelNameGPT35Turbo    ModelName = "gpt-3.5-turbo"
	ModelNameGPT35Turbo16k ModelName = "gpt-3.5-turbo-16k"
	ModelNameGPT4          ModelName = "gpt-4"
	ModelNameGPT4Turbo     ModelName = "gpt-4-turbo"
	ModelNameGPT4Vision    ModelName = "gpt-4-vision"
	ModelNameGPT432k       ModelName = "gpt-4-32k"

	// Claude
	ModelNameClaude2       ModelName = "claude-2"
	ModelNameClaude2_1     ModelName = "claude-2.1"
	ModelNameClaude3Haiku  ModelName = "claude-3-haiku"
	ModelNameClaude3Sonnet ModelName = "claude-3-sonnet"
	ModelNameClaude3Opus   ModelName = "claude-3-opus"

	// DeepSeek
	ModelNameDeepSeekCoder ModelName = "deepseek-coder"
	ModelNameDeepSeekChat  ModelName = "deepseek-chat"

	// Grok
	ModelNameGrok1 ModelName = "grok-1"

	// Local models
	ModelNameLlama2 ModelName = "llama-2"
	ModelNameLlama3 ModelName = "llama-3"

	// Mistral
	ModelNameMistralMedium ModelName = "mistral-medium"
	ModelNameMistralLarge  ModelName = "mistral-large"
	ModelNameMistralSmall  ModelName = "mistral-small"
)

// ModelConfiguration конфигурация конкретной модели
type ModelConfiguration struct {
	Name       ModelName `json:"name"`       // Имя модели
	Type       ModelType `json:"type"`       // Тип модели (провайдер)
	Context    int       `json:"context"`    // Размер контекста в токенах
	MaxTokens  int       `json:"max_tokens"` // Максимальное количество токенов ответа
	Version    string    `json:"version"`    // Версия модели
	Provider   string    `json:"provider"`   // Провайдер (API)
	Endpoint   string    `json:"endpoint"`   // URL эндпоинта API
	Deprecated bool      `json:"deprecated"` // Флаг устаревшей модели
	Tags       []string  `json:"tags"`       // Теги для фильтрации
}

// Model представляет модель в цепочке обработки
type Model struct {
	ID          string     `json:"id"`          // Уникальный идентификатор
	Name        ModelName  `json:"name"`        // Имя модели
	Type        ModelType  `json:"type"`        // Тип модели (провайдер)
	Role        ModelRole  `json:"role"`        // Роль модели в цепочке
	MaxTokens   int        `json:"max_tokens"`  // Максимальное количество токенов в ответе
	Prompt      string     `json:"prompt"`      // Системный промпт для модели
	Order       int        `json:"order"`       // Порядок модели в цепочке
	Parameters  Parameters `json:"parameters"`  // Параметры запросов к модели
	Temperature float64    `json:"temperature"` // Температура (креативность)
}

// Parameters настройки запросов к модели
type Parameters struct {
	Temperature      float64  `json:"temperature"`       // Температура (креативность)
	TopP             float64  `json:"top_p"`             // Top-P сэмплирование
	FrequencyPenalty float64  `json:"frequency_penalty"` // Штраф за повторение слов
	PresencePenalty  float64  `json:"presence_penalty"`  // Штраф за повторение тем
	Stop             []string `json:"stop"`              // Стоп-слова
}

// Chain представляет цепочку моделей для обработки
type Chain struct {
	ID          string    `json:"id"`          // Уникальный идентификатор
	Name        string    `json:"name"`        // Название цепочки
	Description string    `json:"description"` // Описание цепочки
	Models      []Model   `json:"models"`      // Модели в цепочке
	CreatedAt   time.Time `json:"created_at"`  // Время создания
	UpdatedAt   time.Time `json:"updated_at"`  // Время последнего обновления
	Tags        []string  `json:"tags"`        // Теги для фильтрации
	Metadata    Metadata  `json:"metadata"`    // Дополнительные метаданные
}

// Metadata дополнительные метаданные цепочки
type Metadata struct {
	Author      string                 `json:"author"`       // Автор цепочки
	Version     string                 `json:"version"`      // Версия цепочки
	UseCase     string                 `json:"use_case"`     // Сценарий использования
	InputFormat string                 `json:"input_format"` // Формат входных данных
	Custom      map[string]interface{} `json:"custom"`       // Пользовательские метаданные
}

// Store интерфейс для хранилища цепочек
type Store interface {
	// Save сохраняет цепочку
	Save(chain Chain) error

	// Get возвращает цепочку по ID
	Get(id string) (Chain, error)

	// List возвращает список всех цепочек
	List() ([]Chain, error)

	// Delete удаляет цепочку
	Delete(id string) error

	// Exists проверяет существование цепочки
	Exists(id string) bool
}

// FileChainStore реализация хранилища цепочек в файловой системе
type FileChainStore struct {
	path string
}

// NewFileChainStore создает новое хранилище цепочек в файловой системе
func NewFileChainStore(configDir string) (*FileChainStore, error) {
	path := filepath.Join(configDir, "chains.json")

	// Создаем директорию, если она не существует
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	// Создаем файл, если он не существует
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := saveChains(path, []Chain{}); err != nil {
			return nil, err
		}
	}

	return &FileChainStore{path: path}, nil
}

// Save сохраняет цепочку
func (s *FileChainStore) Save(chain Chain) error {
	chains, err := loadChains(s.path)
	if err != nil {
		return err
	}

	// Для новой цепочки генерируем ID и устанавливаем дату создания
	if chain.ID == "" {
		chain.ID = uuid.New().String()
		chain.CreatedAt = time.Now()
	}

	// Обновляем дату изменения
	chain.UpdatedAt = time.Now()

	// Сортируем модели по порядку
	// TODO: Implement sorting by Order

	// Обновляем или добавляем цепочку
	found := false
	for i, c := range chains {
		if c.ID == chain.ID {
			chains[i] = chain
			found = true
			break
		}
	}

	if !found {
		chains = append(chains, chain)
	}

	return saveChains(s.path, chains)
}

// Get возвращает цепочку по ID
func (s *FileChainStore) Get(id string) (Chain, error) {
	chains, err := loadChains(s.path)
	if err != nil {
		return Chain{}, err
	}

	for _, chain := range chains {
		if chain.ID == id {
			return chain, nil
		}
	}

	return Chain{}, fmt.Errorf("chain with ID '%s' not found", id)
}

// List возвращает список всех цепочек
func (s *FileChainStore) List() ([]Chain, error) {
	return loadChains(s.path)
}

// Delete удаляет цепочку
func (s *FileChainStore) Delete(id string) error {
	chains, err := loadChains(s.path)
	if err != nil {
		return err
	}

	var newChains []Chain
	found := false

	for _, chain := range chains {
		if chain.ID != id {
			newChains = append(newChains, chain)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("chain with ID '%s' not found", id)
	}

	return saveChains(s.path, newChains)
}

// Exists проверяет существование цепочки
func (s *FileChainStore) Exists(id string) bool {
	chains, err := loadChains(s.path)
	if err != nil {
		return false
	}

	for _, chain := range chains {
		if chain.ID == id {
			return true
		}
	}

	return false
}

// ModelRegistry представляет собой реестр доступных моделей
type ModelRegistry struct {
	Models []ModelConfiguration `json:"models"` // Список всех моделей
}

// GetModelsByType возвращает список моделей указанного типа
func (r *ModelRegistry) GetModelsByType(modelType ModelType) []ModelConfiguration {
	var models []ModelConfiguration
	for _, model := range r.Models {
		if model.Type == modelType {
			models = append(models, model)
		}
	}
	return models
}

// GetModelsByRole возвращает список моделей, подходящих для указанной роли
func (r *ModelRegistry) GetModelsByRole(role ModelRole) []ModelConfiguration {
	// TODO: Implement mapping between roles and suitable models
	// For now, just return all models
	return r.Models
}

// GetModelByName возвращает конфигурацию модели по имени
func (r *ModelRegistry) GetModelByName(name ModelName) (ModelConfiguration, error) {
	for _, model := range r.Models {
		if model.Name == name {
			return model, nil
		}
	}
	return ModelConfiguration{}, fmt.Errorf("model '%s' not found", name)
}

// NewModelRegistry создает новый реестр моделей
func NewModelRegistry() *ModelRegistry {
	return &ModelRegistry{
		Models: []ModelConfiguration{
			// OpenAI
			{
				Name:      ModelNameGPT35Turbo,
				Type:      ModelTypeOpenAI,
				Context:   4096,
				MaxTokens: 4096,
				Version:   "1106",
				Provider:  "openai",
				Tags:      []string{"cheap", "fast"},
			},
			{
				Name:      ModelNameGPT35Turbo16k,
				Type:      ModelTypeOpenAI,
				Context:   16384,
				MaxTokens: 4096,
				Version:   "1106",
				Provider:  "openai",
				Tags:      []string{"cheap", "long-context"},
			},
			{
				Name:      ModelNameGPT4,
				Type:      ModelTypeOpenAI,
				Context:   8192,
				MaxTokens: 4096,
				Version:   "0613",
				Provider:  "openai",
				Tags:      []string{"premium", "reasoning"},
			},
			{
				Name:      ModelNameGPT4Turbo,
				Type:      ModelTypeOpenAI,
				Context:   128000,
				MaxTokens: 4096,
				Version:   "1106-preview",
				Provider:  "openai",
				Tags:      []string{"premium", "reasoning", "long-context"},
			},
			// Claude
			{
				Name:      ModelNameClaude3Haiku,
				Type:      ModelTypeClaude,
				Context:   200000,
				MaxTokens: 4096,
				Version:   "claude-3-haiku-20240307",
				Provider:  "anthropic",
				Tags:      []string{"fast", "reasoning", "long-context"},
			},
			{
				Name:      ModelNameClaude3Sonnet,
				Type:      ModelTypeClaude,
				Context:   200000,
				MaxTokens: 4096,
				Version:   "claude-3-sonnet-20240229",
				Provider:  "anthropic",
				Tags:      []string{"premium", "reasoning", "long-context"},
			},
			{
				Name:      ModelNameClaude3Opus,
				Type:      ModelTypeClaude,
				Context:   200000,
				MaxTokens: 4096,
				Version:   "claude-3-opus-20240229",
				Provider:  "anthropic",
				Tags:      []string{"premium", "reasoning", "long-context", "expert"},
			},
			// Deepseek
			{
				Name:      ModelNameDeepSeekCoder,
				Type:      ModelTypeDeepSeek,
				Context:   32768,
				MaxTokens: 4096,
				Version:   "1.5",
				Provider:  "deepseek",
				Tags:      []string{"code", "reasoning", "long-context"},
			},
			// Mistral
			{
				Name:      ModelNameMistralMedium,
				Type:      ModelTypeMistral,
				Context:   32768,
				MaxTokens: 4096,
				Version:   "mistral-medium-latest",
				Provider:  "mistral",
				Tags:      []string{"reasoning", "long-context"},
			},
		},
	}
}

// loadChains загружает список цепочек из файла
func loadChains(path string) ([]Chain, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Chain{}, nil
		}
		return nil, err
	}

	var chains []Chain
	if err := json.Unmarshal(data, &chains); err != nil {
		return nil, err
	}

	return chains, nil
}

// saveChains сохраняет список цепочек в файл
func saveChains(path string, chains []Chain) error {
	data, err := json.MarshalIndent(chains, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}
