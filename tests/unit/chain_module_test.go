package unit

import (
	"os"
	"testing"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/chain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModelConfiguration тестирует функциональность структуры ModelConfiguration
func TestModelConfiguration(t *testing.T) {
	// Создаем тестовую конфигурацию модели
	config := chain.ModelConfiguration{
		Name:       chain.ModelName("gpt-4"),
		Type:       chain.ModelType("openai"),
		Context:    8192,
		MaxTokens:  1000,
		Version:    "latest",
		Provider:   "openai",
		Endpoint:   "https://api.openai.com/v1",
		Deprecated: false,
		Tags:       []string{"large", "expensive"},
	}

	// Проверяем поля конфигурации
	assert.Equal(t, chain.ModelName("gpt-4"), config.Name)
	assert.Equal(t, chain.ModelType("openai"), config.Type)
	assert.Equal(t, 8192, config.Context)
	assert.Equal(t, 1000, config.MaxTokens)
	assert.Equal(t, "latest", config.Version)
	assert.Equal(t, "openai", config.Provider)
	assert.Equal(t, "https://api.openai.com/v1", config.Endpoint)
	assert.False(t, config.Deprecated)
	assert.ElementsMatch(t, []string{"large", "expensive"}, config.Tags)
}

// TestModel тестирует функциональность структуры Model
func TestModel(t *testing.T) {
	// Создаем тестовую модель
	model := chain.Model{
		ID:          "model-1",
		Name:        chain.ModelName("gpt-4"),
		Type:        chain.ModelType("openai"),
		Role:        chain.ModelRole("analyzer"),
		MaxTokens:   1000,
		Prompt:      "Analyze the following text and extract key points.",
		Order:       0,
		Temperature: 0.7,
		Parameters: chain.Parameters{
			Temperature:      0.7,
			TopP:             0.9,
			FrequencyPenalty: 0.0,
			PresencePenalty:  0.0,
			Stop:             []string{"\n\n"},
		},
	}

	// Проверяем поля модели
	assert.Equal(t, "model-1", model.ID)
	assert.Equal(t, chain.ModelName("gpt-4"), model.Name)
	assert.Equal(t, chain.ModelType("openai"), model.Type)
	assert.Equal(t, chain.ModelRole("analyzer"), model.Role)
	assert.Equal(t, 1000, model.MaxTokens)
	assert.Equal(t, "Analyze the following text and extract key points.", model.Prompt)
	assert.Equal(t, 0, model.Order)
	assert.Equal(t, 0.7, model.Temperature)
	assert.Equal(t, 0.9, model.Parameters.TopP)
	assert.Equal(t, 0.0, model.Parameters.FrequencyPenalty)
	assert.Equal(t, 0.0, model.Parameters.PresencePenalty)
	assert.ElementsMatch(t, []string{"\n\n"}, model.Parameters.Stop)
}

// TestChain тестирует функциональность структуры Chain
func TestChain(t *testing.T) {
	// Создаем текущее время для тестирования
	now := time.Now()

	// Создаем тестовую цепочку
	testChain := chain.Chain{
		ID:          "test-chain-1",
		Name:        "Test Chain",
		Description: "Test chain for unit tests",
		Models: []chain.Model{
			{
				ID:        "model-1",
				Name:      chain.ModelName("gpt-3.5-turbo"),
				Type:      chain.ModelType("openai"),
				Role:      chain.ModelRole("analyzer"),
				MaxTokens: 1000,
				Prompt:    "Analyze the following text and extract key points.",
				Order:     0,
			},
			{
				ID:        "model-2",
				Name:      chain.ModelName("gpt-3.5-turbo"),
				Type:      chain.ModelType("openai"),
				Role:      chain.ModelRole("summarizer"),
				MaxTokens: 500,
				Prompt:    "Summarize the analysis in a concise manner.",
				Order:     1,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
		Tags:      []string{"test", "unit"},
		Metadata: chain.Metadata{
			Author:      "Unit Test",
			Version:     "1.0",
			UseCase:     "Testing",
			InputFormat: "text",
			Custom: map[string]interface{}{
				"testKey": "testValue",
			},
		},
	}

	// Проверяем поля цепочки
	assert.Equal(t, "test-chain-1", testChain.ID)
	assert.Equal(t, "Test Chain", testChain.Name)
	assert.Equal(t, "Test chain for unit tests", testChain.Description)
	assert.Equal(t, 2, len(testChain.Models))
	assert.Equal(t, now, testChain.CreatedAt)
	assert.Equal(t, now, testChain.UpdatedAt)
	assert.ElementsMatch(t, []string{"test", "unit"}, testChain.Tags)
	assert.Equal(t, "Unit Test", testChain.Metadata.Author)
	assert.Equal(t, "1.0", testChain.Metadata.Version)
	assert.Equal(t, "Testing", testChain.Metadata.UseCase)
	assert.Equal(t, "text", testChain.Metadata.InputFormat)
	assert.Equal(t, "testValue", testChain.Metadata.Custom["testKey"])

	// Проверяем порядок моделей
	assert.Equal(t, chain.ModelRole("analyzer"), testChain.Models[0].Role)
	assert.Equal(t, chain.ModelRole("summarizer"), testChain.Models[1].Role)
	assert.Equal(t, 0, testChain.Models[0].Order)
	assert.Equal(t, 1, testChain.Models[1].Order)
}

// TestFileChainStore тестирует функциональность хранилища цепочек на основе файлов
func TestFileChainStore(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir, err := os.MkdirTemp("", "chain-store-test-*")
	require.NoError(t, err)

	// Очищаем временную директорию после тестов
	defer os.RemoveAll(tempDir)

	// Создаем хранилище цепочек
	store, err := chain.NewFileChainStore(tempDir)
	require.NoError(t, err)

	// Создаем тестовую цепочку
	now := time.Now()
	testChain := chain.Chain{
		ID:          "test-chain-1",
		Name:        "Test Chain",
		Description: "Test chain for unit tests",
		Models: []chain.Model{
			{
				ID:        "model-1",
				Name:      chain.ModelName("gpt-3.5-turbo"),
				Type:      chain.ModelType("openai"),
				Role:      chain.ModelRole("analyzer"),
				MaxTokens: 1000,
				Prompt:    "Analyze the following text and extract key points.",
				Order:     0,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
		Tags:      []string{"test", "unit"},
	}

	// Сохраняем цепочку
	err = store.Save(testChain)
	require.NoError(t, err)

	// Проверяем существование цепочки
	exists := store.Exists(testChain.ID)
	assert.True(t, exists)

	// Получаем цепочку
	retrievedChain, err := store.Get(testChain.ID)
	require.NoError(t, err)

	// Проверяем, что цепочка сохранена и загружена корректно
	assert.Equal(t, testChain.ID, retrievedChain.ID)
	assert.Equal(t, testChain.Name, retrievedChain.Name)
	assert.Equal(t, testChain.Description, retrievedChain.Description)
	assert.Equal(t, len(testChain.Models), len(retrievedChain.Models))
	assert.Equal(t, testChain.Models[0].ID, retrievedChain.Models[0].ID)
	assert.Equal(t, testChain.Models[0].Name, retrievedChain.Models[0].Name)
	assert.Equal(t, testChain.Models[0].Type, retrievedChain.Models[0].Type)
	assert.Equal(t, testChain.Models[0].Role, retrievedChain.Models[0].Role)
	assert.ElementsMatch(t, testChain.Tags, retrievedChain.Tags)

	// Получаем список всех цепочек
	chains, err := store.List()
	require.NoError(t, err)
	assert.Equal(t, 1, len(chains))
	assert.Equal(t, testChain.ID, chains[0].ID)

	// Удаляем цепочку
	err = store.Delete(testChain.ID)
	require.NoError(t, err)

	// Проверяем, что цепочка удалена
	exists = store.Exists(testChain.ID)
	assert.False(t, exists)

	// Убеждаемся, что получение удаленной цепочки возвращает ошибку
	_, err = store.Get(testChain.ID)
	assert.Error(t, err)
}

// TestModelRegistry тестирует функциональность реестра моделей
func TestModelRegistry(t *testing.T) {
	// Создаем реестр моделей
	registry := chain.NewModelRegistry()

	// Проверяем, что реестр не пуст
	assert.Greater(t, len(registry.Models), 0)

	// Проверяем получение моделей по типу
	openaiModels := registry.GetModelsByType(chain.ModelType("openai"))
	assert.Greater(t, len(openaiModels), 0)
	for _, model := range openaiModels {
		assert.Equal(t, chain.ModelType("openai"), model.Type)
	}

	// Проверяем получение моделей по роли
	analyzerModels := registry.GetModelsByRole(chain.ModelRole("analyzer"))
	assert.Greater(t, len(analyzerModels), 0)

	// Проверяем получение конкретной модели по имени
	gpt4, err := registry.GetModelByName(chain.ModelName("gpt-4"))
	require.NoError(t, err)
	assert.Equal(t, chain.ModelName("gpt-4"), gpt4.Name)
	assert.Equal(t, chain.ModelType("openai"), gpt4.Type)

	// Проверяем получение несуществующей модели
	_, err = registry.GetModelByName(chain.ModelName("non-existent-model"))
	assert.Error(t, err)
}
