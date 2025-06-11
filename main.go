package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/grik-ai/ricochet-task/cmd/ricochet"
	"github.com/grik-ai/ricochet-task/pkg/api"
	"github.com/grik-ai/ricochet-task/pkg/chain"
	"github.com/grik-ai/ricochet-task/pkg/checkpoint"
	"github.com/grik-ai/ricochet-task/pkg/key"
	"github.com/grik-ai/ricochet-task/pkg/mcp"
	"github.com/grik-ai/ricochet-task/pkg/model"
	"github.com/grik-ai/ricochet-task/pkg/orchestrator"
	"github.com/grik-ai/ricochet-task/pkg/task"
)

// Config представляет конфигурацию приложения
type Config struct {
	ConfigDir     string
	DefaultChain  string
	WorkspacePath string
}

// Загрузка конфигурации
func loadConfig() (*Config, error) {
	// Пока упрощенная реализация
	return &Config{
		ConfigDir:     "",
		DefaultChain:  "default",
		WorkspacePath: "./",
	}, nil
}

// KeyStoreAdapter адаптер для приведения FileKeyStore к интерфейсу key.Store
type KeyStoreAdapter struct {
	*key.FileKeyStore
}

// Exists проверяет существование ключа (реализация недостающего метода)
func (a *KeyStoreAdapter) Exists(id string) bool {
	// Простая реализация: пытаемся получить ключ
	_, err := a.Get(id)
	return err == nil
}

// GetByProvider возвращает список ключей для указанного провайдера
func (a *KeyStoreAdapter) GetByProvider(provider string) ([]key.Key, error) {
	// Получаем все ключи
	keys, err := a.List()
	if err != nil {
		return nil, err
	}

	// Фильтруем по провайдеру
	var result []key.Key
	for _, k := range keys {
		if k.Provider == provider {
			result = append(result, k)
		}
	}

	return result, nil
}

// Save сохраняет ключ
func (a *KeyStoreAdapter) Save(k key.Key) error {
	// Проверяем, существует ли ключ
	if a.Exists(k.ID) {
		// Если существует, обновляем
		return a.Update(k)
	} else {
		// Если не существует, добавляем
		return a.Add(k)
	}
}

func main() {
	// Инициализируем конфигурацию
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка загрузки конфигурации: %v\n", err)
		os.Exit(1)
	}

	// Инициализируем хранилища
	configDir := cfg.ConfigDir
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка получения домашней директории: %v\n", err)
			os.Exit(1)
		}
		configDir = filepath.Join(homeDir, ".ricochet")
	}

	// Создаем директорию конфигурации, если она не существует
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка создания директории конфигурации: %v\n", err)
		os.Exit(1)
	}

	// Инициализируем хранилища
	fileKeyStore, err := key.NewFileKeyStore(configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка инициализации хранилища ключей: %v\n", err)
		os.Exit(1)
	}

	// Оборачиваем FileKeyStore в адаптер, реализующий интерфейс key.Store
	keyStore := &KeyStoreAdapter{FileKeyStore: fileKeyStore}

	chainStore, err := chain.NewFileChainStore(configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка инициализации хранилища цепочек: %v\n", err)
		os.Exit(1)
	}

	checkpointStore, err := checkpoint.NewFileCheckpointStore(configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка инициализации хранилища чекпоинтов: %v\n", err)
		os.Exit(1)
	}

	taskStore, err := task.NewFileTaskStore(configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка инициализации хранилища задач: %v\n", err)
		os.Exit(1)
	}

	// Инициализируем API-клиент (без аргументов)
	apiClient := api.NewClient()

	// Инициализируем менеджер задач
	taskManager := task.NewTaskManager(taskStore)

	// Инициализируем фабрику провайдеров моделей
	modelFactory := model.NewProviderFactory()

	// Получаем ключи API и регистрируем провайдеров
	keys, err := fileKeyStore.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка получения API-ключей: %v\n", err)
		os.Exit(1)
	}

	// Регистрируем доступные провайдеры
	for _, k := range keys {
		switch k.Provider {
		case "openai":
			modelFactory.RegisterProvider(model.NewOpenAIProvider(k.Value, ""))
		// Другие провайдеры будут добавлены позже
		default:
			fmt.Printf("Провайдер %s не поддерживается, ключ пропущен\n", k.Provider)
		}
	}

	// Инициализируем исполнитель задач
	taskExecutor := task.NewTaskExecutor(
		taskManager,
		&ModelProviderAdapter{Factory: modelFactory},
		task.DefaultExecutorConfig(),
	)

	// Инициализируем оркестратор
	orchestratorImpl := orchestrator.NewOrchestrator(
		apiClient,
		keyStore,
		chainStore,
		checkpointStore,
		taskManager,
		taskExecutor,
		modelFactory,
	)

	// Инициализируем интеграцию с MCP
	mcpIntegration := mcp.NewMCPIntegration("", cfg.DefaultChain)

	// Инициализируем обработчик MCP
	workspaceRoot := "./"
	if cfg.WorkspacePath != "" {
		workspaceRoot = cfg.WorkspacePath
	}

	// Храним обработчик в переменной, чтобы он использовался в CLI или API сервере
	_ = mcp.NewMCPHandler(
		orchestratorImpl,
		chainStore,
		taskManager,
		taskExecutor,
		modelFactory,
		mcpIntegration,
		workspaceRoot,
	)

	// Запускаем CLI
	if err := ricochet.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка выполнения команды: %v\n", err)
		os.Exit(1)
	}
}

// ModelProviderAdapter адаптер для использования фабрики провайдеров моделей с исполнителем задач
type ModelProviderAdapter struct {
	Factory *model.ProviderFactory
}

// Execute выполняет запрос к модели
func (a *ModelProviderAdapter) Execute(ctx context.Context, model chain.Model, prompt string, options map[string]interface{}) (string, error) {
	provider, err := a.Factory.GetProviderForModel(model)
	if err != nil {
		return "", fmt.Errorf("provider not found: %w", err)
	}
	return provider.Execute(ctx, model, prompt, options)
}

// EstimateTokens оценивает количество токенов в тексте
func (a *ModelProviderAdapter) EstimateTokens(text string) int {
	estimator := model.NewTokenEstimator()
	return estimator.EstimateTokens(text, "")
}

// GetModel возвращает модель по имени
func (a *ModelProviderAdapter) GetModel(name chain.ModelName) (chain.ModelConfiguration, error) {
	// Перебираем все типы моделей, которые мы можем поддерживать
	modelTypes := []chain.ModelType{
		chain.ModelTypeOpenAI,
		chain.ModelTypeClaude,
		chain.ModelTypeDeepSeek,
		chain.ModelTypeGrok,
		chain.ModelTypeLlama,
		chain.ModelTypeMistral,
	}

	// Ищем модель в каждом типе
	for _, modelType := range modelTypes {
		if provider, err := a.Factory.GetProvider(modelType); err == nil {
			if model, err := provider.GetModel(name); err == nil {
				return model, nil
			}
		}
	}

	return chain.ModelConfiguration{}, fmt.Errorf("model not found: %s", name)
}
