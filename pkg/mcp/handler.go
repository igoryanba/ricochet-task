package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/grik-ai/ricochet-task/pkg/chain"
	"github.com/grik-ai/ricochet-task/pkg/model"
	"github.com/grik-ai/ricochet-task/pkg/orchestrator"
	"github.com/grik-ai/ricochet-task/pkg/task"
)

// MCPHandler обработчик команд MCP
type MCPHandler struct {
	orchestrator      orchestrator.Orchestrator
	chainStore        chain.Store
	taskManager       task.TaskManager
	taskExecutor      task.TaskExecutor
	modelFactory      *model.ProviderFactory
	mcpIntegration    *MCPIntegration
	workspaceRootPath string
}

// NewMCPHandler создает новый обработчик команд MCP
func NewMCPHandler(
	orchestrator orchestrator.Orchestrator,
	chainStore chain.Store,
	taskManager task.TaskManager,
	taskExecutor task.TaskExecutor,
	modelFactory *model.ProviderFactory,
	mcpIntegration *MCPIntegration,
	workspaceRootPath string,
) *MCPHandler {
	return &MCPHandler{
		orchestrator:      orchestrator,
		chainStore:        chainStore,
		taskManager:       taskManager,
		taskExecutor:      taskExecutor,
		modelFactory:      modelFactory,
		mcpIntegration:    mcpIntegration,
		workspaceRootPath: workspaceRootPath,
	}
}

// MCPCommandRequest запрос команды MCP
type MCPCommandRequest struct {
	Command    string                 `json:"command"`
	Text       string                 `json:"text"`
	ChainID    string                 `json:"chainId"`
	Files      []string               `json:"files"`
	Parameters map[string]interface{} `json:"parameters"`
}

// MCPCommandResponse ответ на команду MCP
type MCPCommandResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Result  string `json:"result"`
	TaskID  string `json:"taskId"`
	RunID   string `json:"runId"`
}

// HandleCommand обрабатывает команду MCP
func (h *MCPHandler) HandleCommand(requestBody []byte) ([]byte, error) {
	// Разбираем запрос
	var request MCPCommandRequest
	if err := json.Unmarshal(requestBody, &request); err != nil {
		return createErrorResponse("Failed to parse request")
	}

	// Обрабатываем команду
	switch request.Command {
	case "process_text":
		return h.handleProcessText(request)
	case "process_file":
		return h.handleProcessFile(request)
	case "list_chains":
		return h.handleListChains()
	case "get_chain":
		return h.handleGetChain(request)
	case "get_config":
		return h.handleGetConfig()
	default:
		return createErrorResponse(fmt.Sprintf("Unknown command: %s", request.Command))
	}
}

// handleProcessText обрабатывает команду process_text
func (h *MCPHandler) handleProcessText(request MCPCommandRequest) ([]byte, error) {
	// Проверяем параметры
	if request.Text == "" {
		return createErrorResponse("Text is required")
	}

	// Получаем указанную цепочку
	chainObj, err := h.chainStore.Get(request.ChainID)
	if err != nil {
		return createErrorResponse(fmt.Sprintf("Chain not found: %s", err))
	}

	// Обрабатываем текст
	ctx := context.Background()
	result, err := h.mcpIntegration.ProcessTextWithMCP(ctx, request.Text, chainObj, h.taskManager, h.taskExecutor)
	if err != nil {
		return createErrorResponse(fmt.Sprintf("Failed to process text: %s", err))
	}

	// Возвращаем результат
	response := MCPCommandResponse{
		Success: true,
		Message: "Text processed successfully",
		Result:  result,
	}

	return json.Marshal(response)
}

// handleProcessFile обрабатывает команду process_file
func (h *MCPHandler) handleProcessFile(request MCPCommandRequest) ([]byte, error) {
	// Проверяем параметры
	if len(request.Files) == 0 {
		return createErrorResponse("At least one file is required")
	}

	// Получаем указанную цепочку
	chainObj, err := h.chainStore.Get(request.ChainID)
	if err != nil {
		return createErrorResponse(fmt.Sprintf("Chain not found: %s", err))
	}

	// Обрабатываем каждый файл
	var resultTexts []string
	var lastTaskID string
	ctx := context.Background()

	for _, filePath := range request.Files {
		// Проверяем путь к файлу
		if !isPathSafe(filePath, h.workspaceRootPath) {
			return createErrorResponse(fmt.Sprintf("Invalid file path: %s", filePath))
		}

		// Читаем содержимое файла
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return createErrorResponse(fmt.Sprintf("Failed to read file: %s", err))
		}

		// Создаем задачу обработки файла
		taskObj := task.Task{
			Type:        task.TaskTypeModelExecution,
			Title:       fmt.Sprintf("MCP: Process file %s", filepath.Base(filePath)),
			Description: fmt.Sprintf("Обработка файла через MCP: %s", filePath),
			Status:      task.StatusReady,
			Model:       &chainObj.Models[0], // Используем первую модель в цепочке
			Input: task.TaskInput{
				Type:   "text",
				Source: string(content),
				Metadata: map[string]interface{}{
					"file_path": filePath,
				},
			},
			ChainID: chainObj.ID,
		}

		// Сохраняем задачу
		taskID, err := h.taskManager.CreateTask(taskObj)
		if err != nil {
			return createErrorResponse(fmt.Sprintf("Failed to create task: %s", err))
		}

		// Выполняем задачу
		if err := h.taskExecutor.ExecuteTask(ctx, taskID); err != nil {
			return createErrorResponse(fmt.Sprintf("Task execution failed: %s", err))
		}

		// Получаем результат выполнения задачи
		taskObj, err = h.taskManager.GetTask(taskID)
		if err != nil {
			return createErrorResponse(fmt.Sprintf("Failed to get task: %s", err))
		}

		if taskObj.Status != task.StatusCompleted {
			return createErrorResponse(fmt.Sprintf("Task not completed: %s", taskObj.Status))
		}

		resultTexts = append(resultTexts, taskObj.Output.Destination)
		lastTaskID = taskID
	}

	// Возвращаем результат
	response := MCPCommandResponse{
		Success: true,
		Message: "Files processed successfully",
		Result:  strings.Join(resultTexts, "\n\n"),
		TaskID:  lastTaskID,
	}

	return json.Marshal(response)
}

// handleListChains обрабатывает команду list_chains
func (h *MCPHandler) handleListChains() ([]byte, error) {
	chains, err := h.chainStore.List()
	if err != nil {
		return createErrorResponse(fmt.Sprintf("Failed to list chains: %s", err))
	}

	// Преобразуем цепочки в формат для ответа
	type ChainInfo struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	chainInfos := make([]ChainInfo, 0, len(chains))
	for _, c := range chains {
		chainInfos = append(chainInfos, ChainInfo{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
		})
	}

	// Формируем ответ
	response := struct {
		Success bool        `json:"success"`
		Message string      `json:"message"`
		Chains  []ChainInfo `json:"chains"`
	}{
		Success: true,
		Message: "Chains retrieved successfully",
		Chains:  chainInfos,
	}

	return json.Marshal(response)
}

// handleGetChain обрабатывает команду get_chain
func (h *MCPHandler) handleGetChain(request MCPCommandRequest) ([]byte, error) {
	if request.ChainID == "" {
		return createErrorResponse("Chain ID is required")
	}

	chainObj, err := h.chainStore.Get(request.ChainID)
	if err != nil {
		return createErrorResponse(fmt.Sprintf("Chain not found: %s", err))
	}

	// Формируем ответ
	response := struct {
		Success bool        `json:"success"`
		Message string      `json:"message"`
		Chain   chain.Chain `json:"chain"`
	}{
		Success: true,
		Message: "Chain retrieved successfully",
		Chain:   chainObj,
	}

	return json.Marshal(response)
}

// handleGetConfig обрабатывает команду get_config
func (h *MCPHandler) handleGetConfig() ([]byte, error) {
	config, err := h.mcpIntegration.LoadConfig()
	if err != nil {
		return createErrorResponse(fmt.Sprintf("Failed to load config: %s", err))
	}

	// Формируем ответ
	response := struct {
		Success bool       `json:"success"`
		Message string     `json:"message"`
		Config  *MCPConfig `json:"config"`
	}{
		Success: true,
		Message: "Config retrieved successfully",
		Config:  config,
	}

	return json.Marshal(response)
}

// createErrorResponse создает ответ с ошибкой
func createErrorResponse(message string) ([]byte, error) {
	response := MCPCommandResponse{
		Success: false,
		Message: message,
	}
	return json.Marshal(response)
}

// isPathSafe проверяет, безопасен ли путь
func isPathSafe(path, rootPath string) bool {
	// Получаем абсолютный путь
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	// Проверяем, находится ли путь внутри рабочей директории
	return strings.HasPrefix(absPath, rootPath)
}
