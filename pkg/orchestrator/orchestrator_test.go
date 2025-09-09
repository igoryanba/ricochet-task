//go:build integration
// +build integration

package orchestrator_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/api"
	"github.com/grik-ai/ricochet-task/pkg/chain"
	"github.com/grik-ai/ricochet-task/pkg/checkpoint"
	"github.com/grik-ai/ricochet-task/pkg/orchestrator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestEnvironment настраивает тестовое окружение
func setupTestEnvironment(t *testing.T) (*api.Client, chain.Store, checkpoint.Store) {
	t.Helper()

	// Создаем временные директории для тестов
	tempDir, err := os.MkdirTemp("", "ricochet-integration-test-*")
	require.NoError(t, err)

	// Добавляем cleanup для удаления временных директорий после тестов
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	// Создаем необходимые поддиректории
	chainsDir := tempDir + "/chains"
	checkpointsDir := tempDir + "/checkpoints"
	require.NoError(t, os.Mkdir(chainsDir, 0755))
	require.NoError(t, os.Mkdir(checkpointsDir, 0755))

	// Инициализируем API-клиент с тестовыми ключами
	apiClient := api.NewClient()
	apiClient.SetAPIKey(api.ProviderOpenAI, os.Getenv("TEST_OPENAI_API_KEY"))
	apiClient.SetAPIKey("anthropic", os.Getenv("TEST_ANTHROPIC_API_KEY"))

	// Инициализируем хранилища
	chainStore, err := chain.NewFileChainStore(chainsDir)
	require.NoError(t, err)

	checkpointStore, err := checkpoint.NewFileCheckpointStore(checkpointsDir)
	require.NoError(t, err)

	return apiClient, chainStore, checkpointStore
}

// createTestChain создает тестовую цепочку моделей
func createTestChain() chain.Chain {
	now := time.Now()
	return chain.Chain{
		ID:          "test-chain-1",
		Name:        "Test Chain",
		Description: "Test chain for integration tests",
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
		Tags:      []string{"test", "integration"},
		Metadata: chain.Metadata{
			Author:      "Integration Test",
			Version:     "1.0",
			UseCase:     "Testing",
			InputFormat: "text",
		},
	}
}

// TestChainCreationAndRetrieval тестирует создание и получение цепочки
func TestChainCreationAndRetrieval(t *testing.T) {
	// Настраиваем тестовое окружение
	_, chainStore, _ := setupTestEnvironment(t)

	// Создаем тестовую цепочку
	testChain := createTestChain()

	// Сохраняем цепочку
	err := chainStore.Save(testChain)
	require.NoError(t, err)

	// Проверяем, что цепочка существует
	exists := chainStore.Exists(testChain.ID)
	assert.True(t, exists)

	// Получаем цепочку и проверяем ее поля
	retrievedChain, err := chainStore.Get(testChain.ID)
	require.NoError(t, err)

	assert.Equal(t, testChain.ID, retrievedChain.ID)
	assert.Equal(t, testChain.Name, retrievedChain.Name)
	assert.Equal(t, testChain.Description, retrievedChain.Description)
	assert.Equal(t, len(testChain.Models), len(retrievedChain.Models))
	assert.Equal(t, testChain.Models[0].Name, retrievedChain.Models[0].Name)
	assert.Equal(t, testChain.Models[1].Role, retrievedChain.Models[1].Role)
}

// TestChainExecution тестирует выполнение цепочки моделей
func TestChainExecution(t *testing.T) {
	// Пропускаем тест, если не указаны API-ключи
	if os.Getenv("TEST_OPENAI_API_KEY") == "" {
		t.Skip("Skipping test because TEST_OPENAI_API_KEY is not set")
	}

	// Настраиваем тестовое окружение
	apiClient, chainStore, checkpointStore := setupTestEnvironment(t)
	_ = apiClient // временно не используем

	// Создаем тестовую цепочку
	testChain := createTestChain()

	// Сохраняем цепочку
	err := chainStore.Save(testChain)
	require.NoError(t, err)

	// Создаем оркестратор
	orch := &orchestrator.DefaultOrchestrator{
		// Упрощенная инициализация для тестов
	}

	// Задаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Готовим входные данные
	input := orchestrator.TaskInput{
		Text: "This is a test input for the Ricochet chain execution. The integration test is checking if the chain can process text properly.",
	}

	// Выполняем цепочку
	runID, err := orch.RunChain(ctx, testChain.ID, input, orchestrator.DefaultProcessingOptions())
	require.NoError(t, err)
	assert.NotEmpty(t, runID)

	// TODO: дождаться завершения выполнения в mock-режиме и проверить статус

	// Ждем завершения выполнения (в реальном тесте нужно использовать polling)
	// В данном примере просто проверим, что контрольные точки созданы
	checkpoints, err := checkpointStore.List(runID)
	require.NoError(t, err)
	assert.True(t, len(checkpoints) > 0, "No checkpoints were created")
}

// TestChainBuilder тестирует построитель цепочек
func TestChainBuilder(t *testing.T) {
	// Настраиваем тестовое окружение
	_, chainStore, _ := setupTestEnvironment(t)

	// Создаем сессию построителя цепочек
	sessionID := "test-session-" + time.Now().Format("20060102150405")
	builderSession := &ChainBuilderSession{
		ID:          sessionID,
		ChainName:   "Test Builder Chain",
		ChainDesc:   "Chain created through the builder",
		Status:      "editing",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		CurrentStep: 0,
		Steps:       []BuilderStep{},
	}

	// Добавляем шаги в сессию
	builderSession.Steps = append(builderSession.Steps, BuilderStep{
		Index:       0,
		ModelRole:   "analyzer",
		ModelID:     "gpt-4",
		Provider:    "openai",
		Description: "Анализ структуры",
		Prompt:      "Проанализируйте структуру документа",
		IsCompleted: true,
	})

	builderSession.Steps = append(builderSession.Steps, BuilderStep{
		Index:       1,
		ModelRole:   "summarizer",
		ModelID:     "claude-3-opus",
		Provider:    "anthropic",
		Description: "Суммаризация",
		Prompt:      "Создайте краткое резюме документа",
		IsCompleted: true,
	})

	// Завершаем сессию и создаем цепочку
	chainID, err := createChainFromSession(builderSession)
	require.NoError(t, err)
	assert.NotEmpty(t, chainID)

	// Проверяем, что цепочка создана и сохранена
	retrievedChain, err := chainStore.Get(chainID)
	require.NoError(t, err)

	assert.Equal(t, builderSession.ChainName, retrievedChain.Name)
	assert.Equal(t, 2, len(retrievedChain.Models))
	assert.Equal(t, chain.ModelRole("analyzer"), retrievedChain.Models[0].Role)
	assert.Equal(t, chain.ModelRole("summarizer"), retrievedChain.Models[1].Role)
}

// Для тестов нам нужны типы из mcp пакета
type ChainBuilderSession struct {
	ID          string        `json:"id"`
	ChainName   string        `json:"chain_name"`
	ChainDesc   string        `json:"chain_description"`
	Steps       []BuilderStep `json:"steps"`
	CurrentStep int           `json:"current_step"`
	Status      string        `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type BuilderStep struct {
	Index       int                    `json:"index"`
	ModelRole   string                 `json:"model_role"`
	ModelID     string                 `json:"model_id"`
	Provider    string                 `json:"provider"`
	Description string                 `json:"description"`
	Prompt      string                 `json:"prompt"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	IsCompleted bool                   `json:"is_completed"`
}

// Упрощенная реализация функции createChainFromSession для тестов
func createChainFromSession(session *ChainBuilderSession) (string, error) {
	// В реальной реализации здесь создается и сохраняется цепочка
	// Для теста возвращаем фиктивный ID
	return "chain-" + time.Now().Format("20060102150405"), nil
}
