package models_manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/grik-ai/ricochet-task/pkg/chain"
)

// ModelConfig представляет конфигурацию модели для определенной роли
type ModelConfig struct {
	Provider    string                 `json:"provider"`
	ModelID     string                 `json:"model_id"`
	DisplayName string                 `json:"display_name"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ModelsConfig представляет полную конфигурацию моделей
type ModelsConfig struct {
	Main       ModelConfig            `json:"main"`
	Research   ModelConfig            `json:"research"`
	Fallback   ModelConfig            `json:"fallback"`
	ChainRoles map[string]ModelConfig `json:"chain_roles"`
}

// ModelRole определяет роль модели в системе
type ModelRole struct {
	ID          string                 // Уникальный идентификатор роли
	DisplayName string                 // Отображаемое имя роли
	Description string                 // Описание роли
	Provider    string                 // Провайдер модели
	ModelID     string                 // ID модели
	Parameters  map[string]interface{} // Параметры модели
}

// ModelOption представляет опцию выбора модели
type ModelOption struct {
	Provider     string   `json:"provider"`
	ModelID      string   `json:"model_id"`
	DisplayName  string   `json:"display_name"`
	MaxTokens    int      `json:"max_tokens"`
	Description  string   `json:"description,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	ContextSize  int      `json:"context_size,omitempty"`
	Cost         string   `json:"cost,omitempty"`
}

// ChainStep представляет шаг в цепочке моделей
type ChainStep struct {
	RoleID      string          // Идентификатор роли модели
	ModelID     string          // Идентификатор модели
	Provider    string          // Провайдер модели
	Name        string          // Имя шага
	Description string          // Описание шага
	Prompt      string          // Промпт для модели
	InputFrom   []string        // Источники входных данных
	Params      json.RawMessage // Параметры шага
}

// ModelsManager менеджер моделей
type ModelsManager struct {
	configPath string
	config     ModelsConfig
	mutex      sync.RWMutex
	registry   ModelRegistry
}

// ModelRegistry реестр доступных моделей
type ModelRegistry struct {
	Models map[string][]ModelOption // ключ - провайдер
}

// Константы для стандартных ролей моделей
const (
	RoleMain     = "main"
	RoleResearch = "research"
	RoleFallback = "fallback"

	// Специализированные роли для цепочек
	RoleAnalyzer   = "analyzer"
	RoleSummarizer = "summarizer"
	RoleIntegrator = "integrator"
	RoleExtractor  = "extractor"
	RoleCritic     = "critic"
	RoleRefiner    = "refiner"
	RoleCreator    = "creator"
)

// New создает новый менеджер моделей
func New(configPath string) (*ModelsManager, error) {
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %v", err)
		}
		configPath = filepath.Join(home, ".ricochet", "models.json")
	}

	manager := &ModelsManager{
		configPath: configPath,
		registry: ModelRegistry{
			Models: make(map[string][]ModelOption),
		},
	}

	// Инициализация реестра моделей
	manager.initRegistry()

	// Загрузка конфигурации
	err := manager.LoadConfig()
	if err != nil {
		// Если файл не существует, создаем стандартную конфигурацию
		if os.IsNotExist(err) {
			manager.initDefaultConfig()
			err = manager.SaveConfig()
		}
	}

	return manager, err
}

// initRegistry инициализирует реестр доступных моделей
func (m *ModelsManager) initRegistry() {
	// OpenAI модели
	m.registry.Models["openai"] = []ModelOption{
		{
			Provider:     "openai",
			ModelID:      "gpt-4o",
			DisplayName:  "OpenAI GPT-4o",
			MaxTokens:    16000,
			Description:  "Мощная модель для генерации контента и анализа",
			Capabilities: []string{"code", "reasoning", "creative"},
			ContextSize:  32000,
			Cost:         "~$0.01/1K токенов",
		},
		{
			Provider:     "openai",
			ModelID:      "gpt-4-turbo",
			DisplayName:  "OpenAI GPT-4 Turbo",
			MaxTokens:    16000,
			Description:  "Продвинутая модель с хорошим соотношением цена/качество",
			Capabilities: []string{"code", "reasoning", "creative"},
			ContextSize:  128000,
			Cost:         "~$0.01/1K токенов",
		},
		{
			Provider:     "openai",
			ModelID:      "gpt-3.5-turbo-16k",
			DisplayName:  "OpenAI GPT-3.5 Turbo",
			MaxTokens:    8192,
			Description:  "Быстрая и экономичная модель",
			Capabilities: []string{"basic", "code", "creative"},
			ContextSize:  16000,
			Cost:         "~$0.0005/1K токенов",
		},
	}

	// Anthropic модели
	m.registry.Models["anthropic"] = []ModelOption{
		{
			Provider:     "anthropic",
			ModelID:      "claude-3-opus-20240229",
			DisplayName:  "Anthropic Claude 3 Opus",
			MaxTokens:    24000,
			Description:  "Продвинутая модель для сложных задач анализа и понимания",
			Capabilities: []string{"reasoning", "research", "creative", "code"},
			ContextSize:  200000,
			Cost:         "~$0.015/1K токенов",
		},
		{
			Provider:     "anthropic",
			ModelID:      "claude-3-sonnet-20240229",
			DisplayName:  "Anthropic Claude 3 Sonnet",
			MaxTokens:    18000,
			Description:  "Сбалансированная модель с хорошим соотношением цена/качество",
			Capabilities: []string{"reasoning", "creative", "code"},
			ContextSize:  180000,
			Cost:         "~$0.003/1K токенов",
		},
		{
			Provider:     "anthropic",
			ModelID:      "claude-3-haiku-20240307",
			DisplayName:  "Anthropic Claude 3 Haiku",
			MaxTokens:    12000,
			Description:  "Быстрая и экономичная модель от Anthropic",
			Capabilities: []string{"basic", "creative", "code"},
			ContextSize:  180000,
			Cost:         "~$0.0002/1K токенов",
		},
	}

	// DeepSeek модели
	m.registry.Models["deepseek"] = []ModelOption{
		{
			Provider:     "deepseek",
			ModelID:      "deepseek-coder",
			DisplayName:  "DeepSeek Coder",
			MaxTokens:    16000,
			Description:  "Специализированная модель для работы с кодом",
			Capabilities: []string{"code", "technical"},
			ContextSize:  32000,
			Cost:         "~$0.0008/1K токенов",
		},
		{
			Provider:     "deepseek",
			ModelID:      "deepseek-chat",
			DisplayName:  "DeepSeek Chat",
			MaxTokens:    12000,
			Description:  "Модель для общих задач чата и генерации контента",
			Capabilities: []string{"chat", "creative"},
			ContextSize:  24000,
			Cost:         "~$0.0006/1K токенов",
		},
	}

	// Mistral модели
	m.registry.Models["mistral"] = []ModelOption{
		{
			Provider:     "mistral",
			ModelID:      "mistral-large",
			DisplayName:  "Mistral Large",
			MaxTokens:    16000,
			Description:  "Крупная модель Mistral AI с высокой производительностью",
			Capabilities: []string{"reasoning", "creative", "code"},
			ContextSize:  32000,
			Cost:         "~$0.0006/1K токенов",
		},
		{
			Provider:     "mistral",
			ModelID:      "mistral-medium",
			DisplayName:  "Mistral Medium",
			MaxTokens:    12000,
			Description:  "Сбалансированная модель от Mistral AI",
			Capabilities: []string{"reasoning", "creative"},
			ContextSize:  24000,
			Cost:         "~$0.0003/1K токенов",
		},
		{
			Provider:     "mistral",
			ModelID:      "mistral-small",
			DisplayName:  "Mistral Small",
			MaxTokens:    8000,
			Description:  "Компактная и быстрая модель от Mistral AI",
			Capabilities: []string{"basic", "creative"},
			ContextSize:  16000,
			Cost:         "~$0.0001/1K токенов",
		},
	}
}

// initDefaultConfig инициализирует стандартную конфигурацию
func (m *ModelsManager) initDefaultConfig() {
	m.config = ModelsConfig{
		Main: ModelConfig{
			Provider:    "openai",
			ModelID:     "gpt-4o",
			DisplayName: "OpenAI GPT-4o",
			Parameters: map[string]interface{}{
				"temperature": 0.7,
				"top_p":       1.0,
			},
		},
		Research: ModelConfig{
			Provider:    "anthropic",
			ModelID:     "claude-3-opus-20240229",
			DisplayName: "Anthropic Claude 3 Opus",
			Parameters: map[string]interface{}{
				"temperature": 0.5,
				"top_p":       0.9,
			},
		},
		Fallback: ModelConfig{
			Provider:    "openai",
			ModelID:     "gpt-3.5-turbo-16k",
			DisplayName: "OpenAI GPT-3.5 Turbo",
			Parameters: map[string]interface{}{
				"temperature": 0.7,
				"top_p":       1.0,
			},
		},
		ChainRoles: map[string]ModelConfig{
			"analyzer": {
				Provider:    "anthropic",
				ModelID:     "claude-3-opus-20240229",
				DisplayName: "Anthropic Claude 3 Opus",
				Parameters: map[string]interface{}{
					"temperature": 0.3,
					"top_p":       0.9,
				},
			},
			"summarizer": {
				Provider:    "openai",
				ModelID:     "gpt-4o",
				DisplayName: "OpenAI GPT-4o",
				Parameters: map[string]interface{}{
					"temperature": 0.4,
					"top_p":       0.95,
				},
			},
			"integrator": {
				Provider:    "deepseek",
				ModelID:     "deepseek-coder",
				DisplayName: "DeepSeek Coder",
				Parameters: map[string]interface{}{
					"temperature": 0.5,
					"top_p":       0.9,
				},
			},
			"extractor": {
				Provider:    "openai",
				ModelID:     "gpt-4-turbo",
				DisplayName: "OpenAI GPT-4 Turbo",
				Parameters: map[string]interface{}{
					"temperature": 0.2,
					"top_p":       0.8,
				},
			},
			"organizer": {
				Provider:    "anthropic",
				ModelID:     "claude-3-sonnet-20240229",
				DisplayName: "Anthropic Claude 3 Sonnet",
				Parameters: map[string]interface{}{
					"temperature": 0.6,
					"top_p":       0.9,
				},
			},
			"evaluator": {
				Provider:    "mistral",
				ModelID:     "mistral-large",
				DisplayName: "Mistral Large",
				Parameters: map[string]interface{}{
					"temperature": 0.3,
					"top_p":       0.85,
				},
			},
		},
	}
}

// LoadConfig загружает конфигурацию из файла
func (m *ModelsManager) LoadConfig() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	data, err := ioutil.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &m.config)
}

// SaveConfig сохраняет конфигурацию в файл
func (m *ModelsManager) SaveConfig() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}

	// Создаем директории, если не существуют
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(m.configPath, data, 0644)
}

// GetModelForRole возвращает лучшую модель для указанной роли
// Если роль не назначена ни одной модели, выбирает модель на основе
// заданных предпочтений для типов задач
func (m *ModelsManager) GetModelForRole(roleID string) ModelRole {
	// Сначала проверяем, есть ли настроенная модель для этой роли
	for _, role := range m.GetRoles() {
		if role.ID == roleID {
			return role
		}
	}

	// Если модель для роли не найдена, используем стратегию выбора по умолчанию
	switch roleID {
	case RoleAnalyzer, RoleExtractor:
		// Для анализа и извлечения данных лучше использовать исследовательскую модель
		return m.GetRoleByID(RoleResearch)
	case RoleIntegrator, RoleCritic, RoleRefiner:
		// Для интеграции, критики и улучшений используем основную модель
		return m.GetRoleByID(RoleMain)
	case RoleSummarizer:
		// Для суммаризации подойдет более быстрая модель
		fallback := m.GetRoleByID(RoleFallback)
		if fallback.ModelID != "" {
			return fallback
		}
		return m.GetRoleByID(RoleMain)
	case RoleCreator:
		// Для создания контента нужна самая мощная модель
		main := m.GetRoleByID(RoleMain)
		if main.ModelID != "" {
			return main
		}
		return m.GetRoleByID(RoleResearch)
	default:
		// По умолчанию используем основную модель
		return m.GetRoleByID(RoleMain)
	}
}

// GetRoleByID возвращает конфигурацию роли по ID
func (m *ModelsManager) GetRoleByID(roleID string) ModelRole {
	for _, role := range m.GetRoles() {
		if role.ID == roleID {
			return role
		}
	}

	// Если роль не найдена, возвращаем пустую роль
	return ModelRole{
		ID:          roleID,
		DisplayName: roleID,
		Description: "Автоматически назначенная роль",
	}
}

// AutoAssignModelsToChain автоматически назначает модели для шагов цепочки
// на основе ролей, указанных в шагах
func (m *ModelsManager) AutoAssignModelsToChain(chainSteps []ChainStep) []ChainStep {
	result := make([]ChainStep, len(chainSteps))
	copy(result, chainSteps)

	for i, step := range result {
		// Если модель не указана явно, назначаем на основе роли
		if step.ModelID == "" && step.Provider == "" && step.RoleID != "" {
			role := m.GetModelForRole(step.RoleID)
			result[i].ModelID = role.ModelID
			result[i].Provider = role.Provider

			// Если параметры не указаны, используем параметры роли
			if result[i].Params == nil && role.Parameters != nil {
				paramsJSON, _ := json.Marshal(role.Parameters)
				result[i].Params = paramsJSON
			}
		}
	}

	return result
}

// GetModelConfigForChain возвращает конфигурацию модели для цепочки по роли
func (m *ModelsManager) GetModelConfigForChain(roleID string) chain.Model {
	config := m.GetModelForRole(roleID)

	// Преобразуем ModelConfig в chain.Model
	model := chain.Model{
		ID:         fmt.Sprintf("%s-%s", config.Provider, config.ModelID),
		Name:       chain.ModelName(config.ModelID),
		Type:       chain.ModelType(config.Provider),
		Role:       chain.ModelRole(roleID),
		Parameters: makeParameters(config.Parameters),
	}

	// Дополнительные параметры
	if temp, ok := config.Parameters["temperature"].(float64); ok {
		model.Temperature = temp
	}

	if maxTokens, ok := config.Parameters["max_tokens"].(float64); ok {
		model.MaxTokens = int(maxTokens)
	} else {
		// Значения по умолчанию
		switch config.Provider {
		case "openai":
			model.MaxTokens = 4096
		case "anthropic":
			model.MaxTokens = 4096
		case "deepseek":
			model.MaxTokens = 2048
		default:
			model.MaxTokens = 2048
		}
	}

	return model
}

// makeParameters преобразует map[string]interface{} в chain.Parameters
func makeParameters(params map[string]interface{}) chain.Parameters {
	result := chain.Parameters{
		Temperature:      0.7, // Значение по умолчанию
		TopP:             1.0, // Значение по умолчанию
		FrequencyPenalty: 0.0, // Значение по умолчанию
		PresencePenalty:  0.0, // Значение по умолчанию
		Stop:             []string{},
	}

	if temp, ok := params["temperature"].(float64); ok {
		result.Temperature = temp
	}

	if topP, ok := params["top_p"].(float64); ok {
		result.TopP = topP
	}

	if freqPenalty, ok := params["frequency_penalty"].(float64); ok {
		result.FrequencyPenalty = freqPenalty
	}

	if presPenalty, ok := params["presence_penalty"].(float64); ok {
		result.PresencePenalty = presPenalty
	}

	if stop, ok := params["stop"].([]string); ok {
		result.Stop = stop
	}

	return result
}

// DefaultRoles возвращает список ролей по умолчанию
func DefaultRoles() []ModelRole {
	return []ModelRole{
		{
			ID:          RoleMain,
			DisplayName: "Основная модель",
			Description: "Используется для основных задач генерации и обновления",
		},
		{
			ID:          RoleResearch,
			DisplayName: "Исследовательская модель",
			Description: "Используется для анализа данных и исследовательских задач",
		},
		{
			ID:          RoleFallback,
			DisplayName: "Резервная модель",
			Description: "Используется при недоступности основной модели",
		},
		{
			ID:          RoleAnalyzer,
			DisplayName: "Модель-анализатор",
			Description: "Анализирует и структурирует входные данные",
		},
		{
			ID:          RoleSummarizer,
			DisplayName: "Модель-суммаризатор",
			Description: "Создает краткие резюме на основе анализа",
		},
		{
			ID:          RoleIntegrator,
			DisplayName: "Модель-интегратор",
			Description: "Объединяет результаты работы других моделей",
		},
		{
			ID:          RoleExtractor,
			DisplayName: "Модель-экстрактор",
			Description: "Извлекает ключевую информацию из текста",
		},
		{
			ID:          RoleCritic,
			DisplayName: "Модель-критик",
			Description: "Проверяет и критически оценивает выходные данные",
		},
		{
			ID:          RoleRefiner,
			DisplayName: "Модель-улучшатель",
			Description: "Улучшает и дорабатывает результаты других моделей",
		},
		{
			ID:          RoleCreator,
			DisplayName: "Модель-генератор",
			Description: "Создает новый контент на основе входных данных",
		},
	}
}

// GetRoleDescription возвращает описание для указанной роли
func GetRoleDescription(roleID string) string {
	for _, role := range DefaultRoles() {
		if role.ID == roleID {
			return role.Description
		}
	}
	return "Роль не определена"
}

// GetRoleDisplayName возвращает отображаемое имя для указанной роли
func GetRoleDisplayName(roleID string) string {
	for _, role := range DefaultRoles() {
		if role.ID == roleID {
			return role.DisplayName
		}
	}
	return roleID
}

// GetRoles возвращает список всех доступных ролей моделей
func (m *ModelsManager) GetRoles() []ModelRole {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Получаем список всех моделей для формирования опций
	var allModels []ModelOption
	for _, providerModels := range m.registry.Models {
		allModels = append(allModels, providerModels...)
	}

	// Формируем стандартные роли
	roles := []ModelRole{
		{
			ID:          RoleMain,
			DisplayName: "Основная модель",
			Description: "Основная модель для генерации контента и обновлений",
			Provider:    m.config.Main.Provider,
			ModelID:     m.config.Main.ModelID,
			Parameters:  m.config.Main.Parameters,
		},
		{
			ID:          RoleResearch,
			DisplayName: "Исследовательская модель",
			Description: "Модель для анализа данных и исследования",
			Provider:    m.config.Research.Provider,
			ModelID:     m.config.Research.ModelID,
			Parameters:  m.config.Research.Parameters,
		},
		{
			ID:          RoleFallback,
			DisplayName: "Резервная модель",
			Description: "Модель, используемая при недоступности основной",
			Provider:    m.config.Fallback.Provider,
			ModelID:     m.config.Fallback.ModelID,
			Parameters:  m.config.Fallback.Parameters,
		},
	}

	// Добавляем специализированные роли для цепочек
	for _, role := range DefaultRoles()[3:] { // Пропускаем первые три стандартные роли
		roleID := role.ID
		config, exists := m.config.ChainRoles[roleID]

		if !exists {
			// Используем информацию из DefaultRoles
			roles = append(roles, role)
			continue
		}

		// Используем настроенную конфигурацию
		roles = append(roles, ModelRole{
			ID:          roleID,
			DisplayName: role.DisplayName,
			Description: role.Description,
			Provider:    config.Provider,
			ModelID:     config.ModelID,
			Parameters:  config.Parameters,
		})
	}

	return roles
}
