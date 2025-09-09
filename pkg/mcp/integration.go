package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/grik-ai/ricochet-task/pkg/chain"
	"github.com/grik-ai/ricochet-task/pkg/task"
)

// MCPIntegration интеграция с Multi-platform Component Protocol (MCP)
type MCPIntegration struct {
	configPath   string
	defaultChain string
}

// NewMCPIntegration создает новую интеграцию с MCP
func NewMCPIntegration(configPath, defaultChain string) *MCPIntegration {
	if configPath == "" {
		configPath = defaultConfigPath()
	}
	return &MCPIntegration{
		configPath:   configPath,
		defaultChain: defaultChain,
	}
}

// MCPConfig конфигурация MCP
type MCPConfig struct {
	EnabledEditors []string          `json:"enabled_editors"`
	DefaultChain   string            `json:"default_chain"`
	EditorConfig   map[string]string `json:"editor_config"`
	Providers      []string          `json:"providers"`
}

// LoadConfig загружает конфигурацию MCP
func (m *MCPIntegration) LoadConfig() (*MCPConfig, error) {
	// Проверяем существование файла конфигурации
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// Создаем конфигурацию по умолчанию
		defaultConfig := &MCPConfig{
			EnabledEditors: []string{"vscode", "cursor"},
			DefaultChain:   m.defaultChain,
			EditorConfig:   map[string]string{},
			Providers:      []string{"openai", "anthropic"},
		}
		return defaultConfig, m.SaveConfig(defaultConfig)
	}

	// Читаем файл конфигурации
	data, err := ioutil.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Разбираем JSON
	var config MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig сохраняет конфигурацию MCP
func (m *MCPIntegration) SaveConfig(config *MCPConfig) error {
	// Создаем директорию, если она не существует
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Сериализуем конфигурацию в JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// Записываем файл
	if err := ioutil.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ExecuteTaskWithMCP выполняет задачу с использованием MCP
func (m *MCPIntegration) ExecuteTaskWithMCP(ctx context.Context, taskID string, executor task.TaskExecutor) error {
	return executor.ExecuteTask(ctx, taskID)
}

// ProcessTextWithMCP обрабатывает текст с использованием MCP
func (m *MCPIntegration) ProcessTextWithMCP(ctx context.Context,
	text string,
	chainObj chain.Chain,
	taskManager task.TaskManager,
	executor task.TaskExecutor) (string, error) {

	// Создаем задачу для каждой модели в цепочке
	var lastTaskID string
	var previousTaskID string

	for _, model := range chainObj.Models {
		// Создаем задачу обработки текста
		taskObj := task.Task{
			Type:        task.TaskTypeModelExecution,
			Title:       fmt.Sprintf("MCP: %s", model.Name),
			Description: fmt.Sprintf("Обработка текста через MCP с использованием модели %s", model.Name),
			Status:      task.StatusReady,
			Model:       &model,
			Input: task.TaskInput{
				Type:   "text",
				Source: text,
			},
			ChainID: chainObj.ID,
		}

		// Если есть предыдущая задача, устанавливаем зависимость
		if previousTaskID != "" {
			taskObj.Dependencies = []string{previousTaskID}
			taskObj.Status = task.StatusPending
		}

		// Сохраняем задачу
		taskID, err := taskManager.CreateTask(taskObj)
		if err != nil {
			return "", fmt.Errorf("failed to create task: %w", err)
		}

		previousTaskID = taskID
		lastTaskID = taskID
	}

	// Если нет задач, возвращаем ошибку
	if lastTaskID == "" {
		return "", errors.New("no tasks created")
	}

	// Выполняем последнюю задачу (которая зависит от всех предыдущих)
	if err := executor.ExecuteTask(ctx, lastTaskID); err != nil {
		return "", fmt.Errorf("task execution failed: %w", err)
	}

	// Получаем результат выполнения задачи
	taskObj, err := taskManager.GetTask(lastTaskID)
	if err != nil {
		return "", fmt.Errorf("failed to get task: %w", err)
	}

	if taskObj.Status != task.StatusCompleted {
		return "", fmt.Errorf("task not completed: %s", taskObj.Status)
	}

	return taskObj.Output.Destination, nil
}

// GenerateEditorConfig генерирует конфигурацию для редактора
func (m *MCPIntegration) GenerateEditorConfig(editor string, chainStore chain.Store) (string, error) {
	config, err := m.LoadConfig()
	if err != nil {
		return "", err
	}

	// Проверяем, поддерживается ли редактор
	supported := false
	for _, e := range config.EnabledEditors {
		if e == editor {
			supported = true
			break
		}
	}

	if !supported {
		return "", fmt.Errorf("editor %s is not supported", editor)
	}

	// Получаем список доступных цепочек
	chains, err := chainStore.List()
	if err != nil {
		return "", fmt.Errorf("failed to list chains: %w", err)
	}

	// Формируем конфигурацию для редактора
	switch editor {
	case "vscode":
		return generateVSCodeConfig(chains, config)
	case "cursor":
		return generateCursorConfig(chains, config)
	default:
		return "", fmt.Errorf("unsupported editor: %s", editor)
	}
}

// generateVSCodeConfig генерирует конфигурацию для VS Code
func generateVSCodeConfig(chains []chain.Chain, _ *MCPConfig) (string, error) {
	// Создаем JSON-структуру для конфигурации
	type CommandConfig struct {
		Title     string `json:"title"`
		Command   string `json:"command"`
		Category  string `json:"category"`
		ChainID   string `json:"chainId"`
		ChainName string `json:"chainName"`
	}

	type VSCodeConfig struct {
		Commands    []CommandConfig `json:"commands"`
		DefaultMode string          `json:"defaultMode"`
	}

	// Создаем команды для каждой цепочки
	commands := make([]CommandConfig, 0, len(chains))
	for _, c := range chains {
		commands = append(commands, CommandConfig{
			Title:     fmt.Sprintf("Run %s", c.Name),
			Command:   fmt.Sprintf("ricochet.run.%s", sanitizeID(c.ID)),
			Category:  "Ricochet",
			ChainID:   c.ID,
			ChainName: c.Name,
		})
	}

	// Формируем итоговую конфигурацию
	vsConfig := VSCodeConfig{
		Commands:    commands,
		DefaultMode: "interactive",
	}

	// Сериализуем в JSON
	data, err := json.MarshalIndent(vsConfig, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize config: %w", err)
	}

	return string(data), nil
}

// generateCursorConfig генерирует конфигурацию для Cursor
func generateCursorConfig(chains []chain.Chain, _ *MCPConfig) (string, error) {
	// Создаем JSON-структуру для конфигурации
	type ChainConfig struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	type CursorConfig struct {
		Chains      []ChainConfig `json:"chains"`
		DefaultMode string        `json:"defaultMode"`
	}

	// Создаем конфигурацию для каждой цепочки
	chainConfigs := make([]ChainConfig, 0, len(chains))
	for _, c := range chains {
		chainConfigs = append(chainConfigs, ChainConfig{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
		})
	}

	// Формируем итоговую конфигурацию
	cursorConfig := CursorConfig{
		Chains:      chainConfigs,
		DefaultMode: "interactive",
	}

	// Сериализуем в JSON
	data, err := json.MarshalIndent(cursorConfig, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize config: %w", err)
	}

	return string(data), nil
}

// defaultConfigPath возвращает путь к конфигурации по умолчанию
func defaultConfigPath() string {
	// Получаем путь к домашней директории пользователя
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Если не удалось получить, используем текущую директорию
		return ".ricochet/mcp_config.json"
	}
	return filepath.Join(homeDir, ".ricochet", "mcp_config.json")
}

// sanitizeID санитизирует ID для использования в командах
func sanitizeID(id string) string {
	// Заменяем специальные символы на подчеркивания
	return strings.ReplaceAll(id, "-", "_")
}
