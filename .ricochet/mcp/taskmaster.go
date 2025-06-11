package mcp

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/grik-ai/ricochet-task/internal/config"
	"github.com/grik-ai/ricochet-task/pkg/chain"
	"github.com/grik-ai/ricochet-task/pkg/task"
)

// TaskMasterToolProvider предоставляет инструменты для работы с Task Master
type TaskMasterToolProvider struct {
	workspacePath string
	converter     *task.DefaultRicochetTaskConverter
}

// NewTaskMasterToolProvider создает нового провайдера инструментов Task Master
func NewTaskMasterToolProvider(workspacePath string) (*TaskMasterToolProvider, error) {
	// Получаем конфигурацию
	configPath, err := config.GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить путь к конфигурации: %w", err)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось загрузить конфигурацию: %w", err)
	}

	// Создаем хранилище цепочек
	chainStore, err := chain.NewFileChainStore(cfg.ConfigDir)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать хранилище цепочек: %w", err)
	}

	// Создаем хранилище задач
	taskStore, err := task.NewFileTaskStore(cfg.ConfigDir)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать хранилище задач: %w", err)
	}

	// Создаем менеджер задач
	taskManager := task.NewTaskManager(taskStore)

	// Создаем конвертер
	converter, err := task.NewRicochetTaskConverter(workspacePath, taskManager, chainStore)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать конвертер Task Master: %w", err)
	}

	return &TaskMasterToolProvider{
		workspacePath: workspacePath,
		converter:     converter,
	}, nil
}

// ListTasksRequest запрос на получение списка задач
type ListTasksRequest struct {
	Status        string `json:"status"`        // Фильтр по статусу
	WithSubtasks  bool   `json:"with_subtasks"` // Включать подзадачи
	WorkspacePath string `json:"workspace_path"`
}

// ListTasksResponse ответ на запрос списка задач
type ListTasksResponse struct {
	Tasks []task.RicochetTaskTask `json:"tasks"`
	Error string                  `json:"error,omitempty"`
}

// GetTaskRequest запрос на получение информации о задаче
type GetTaskRequest struct {
	ID            string `json:"id"` // ID задачи
	WorkspacePath string `json:"workspace_path"`
}

// GetTaskResponse ответ на запрос информации о задаче
type GetTaskResponse struct {
	Task  task.RicochetTaskTask `json:"task"`
	Error string                `json:"error,omitempty"`
}

// UpdateTaskStatusRequest запрос на обновление статуса задачи
type UpdateTaskStatusRequest struct {
	ID            string `json:"id"`     // ID задачи
	Status        string `json:"status"` // Новый статус
	WorkspacePath string `json:"workspace_path"`
}

// UpdateTaskStatusResponse ответ на запрос обновления статуса задачи
type UpdateTaskStatusResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// SyncTaskStatusRequest запрос на синхронизацию статуса задачи
type SyncTaskStatusRequest struct {
	TaskID        string `json:"task_id"`  // ID задачи
	ChainID       string `json:"chain_id"` // ID цепочки
	WorkspacePath string `json:"workspace_path"`
}

// SyncTaskStatusResponse ответ на запрос синхронизации статуса задачи
type SyncTaskStatusResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// CreateTaskFromChainRequest запрос на создание задачи из цепочки
type CreateTaskFromChainRequest struct {
	ChainID       string `json:"chain_id"` // ID цепочки
	WorkspacePath string `json:"workspace_path"`
}

// CreateTaskFromChainResponse ответ на запрос создания задачи
type CreateTaskFromChainResponse struct {
	TaskID string `json:"task_id,omitempty"`
	Error  string `json:"error,omitempty"`
}

// CreateChainFromTaskRequest запрос на создание цепочки из задачи
type CreateChainFromTaskRequest struct {
	TaskID        string `json:"task_id"` // ID задачи
	WorkspacePath string `json:"workspace_path"`
}

// CreateChainFromTaskResponse ответ на запрос создания цепочки
type CreateChainFromTaskResponse struct {
	ChainID string `json:"chain_id,omitempty"`
	Error   string `json:"error,omitempty"`
}

// RegisterTaskMasterTools регистрирует инструменты Task Master в MCP
func RegisterTaskMasterTools(server *Server) {
	// Инструмент для получения списка задач
	server.RegisterTool("get_tasks", func(params json.RawMessage) (json.RawMessage, error) {
		var request ListTasksRequest
		if err := json.Unmarshal(params, &request); err != nil {
			return json.Marshal(ListTasksResponse{Error: fmt.Sprintf("ошибка разбора параметров: %s", err)})
		}

		workspacePath := request.WorkspacePath
		if workspacePath == "" {
			var err error
			workspacePath, err = os.Getwd()
			if err != nil {
				return json.Marshal(ListTasksResponse{Error: fmt.Sprintf("не удалось получить рабочую директорию: %s", err)})
			}
		}

		provider, err := NewTaskMasterToolProvider(workspacePath)
		if err != nil {
			return json.Marshal(ListTasksResponse{Error: fmt.Sprintf("ошибка создания провайдера: %s", err)})
		}

		tasks, err := provider.converter.GetRicochetTaskTasks()
		if err != nil {
			return json.Marshal(ListTasksResponse{Error: fmt.Sprintf("ошибка получения задач: %s", err)})
		}

		// Фильтруем задачи по статусу, если нужно
		if request.Status != "" {
			var filteredTasks []task.RicochetTaskTask
			for _, t := range tasks {
				if t.Status == request.Status {
					filteredTasks = append(filteredTasks, t)
				}
			}
			tasks = filteredTasks
		}

		// Удаляем подзадачи, если не нужно
		if !request.WithSubtasks {
			for i := range tasks {
				tasks[i].Subtasks = nil
			}
		}

		return json.Marshal(ListTasksResponse{Tasks: tasks})
	})

	// Инструмент для получения информации о задаче
	server.RegisterTool("get_task", func(params json.RawMessage) (json.RawMessage, error) {
		var request GetTaskRequest
		if err := json.Unmarshal(params, &request); err != nil {
			return json.Marshal(GetTaskResponse{Error: fmt.Sprintf("ошибка разбора параметров: %s", err)})
		}

		if request.ID == "" {
			return json.Marshal(GetTaskResponse{Error: "не указан ID задачи"})
		}

		workspacePath := request.WorkspacePath
		if workspacePath == "" {
			var err error
			workspacePath, err = os.Getwd()
			if err != nil {
				return json.Marshal(GetTaskResponse{Error: fmt.Sprintf("не удалось получить рабочую директорию: %s", err)})
			}
		}

		provider, err := NewTaskMasterToolProvider(workspacePath)
		if err != nil {
			return json.Marshal(GetTaskResponse{Error: fmt.Sprintf("ошибка создания провайдера: %s", err)})
		}

		tasks, err := provider.converter.GetRicochetTaskTasks()
		if err != nil {
			return json.Marshal(GetTaskResponse{Error: fmt.Sprintf("ошибка получения задач: %s", err)})
		}

		// Ищем задачу с указанным ID
		var foundTask task.RicochetTaskTask
		found := false

		for _, t := range tasks {
			if t.ID == request.ID {
				foundTask = t
				found = true
				break
			}

			// Проверяем подзадачи
			for _, st := range t.Subtasks {
				if st.ID == request.ID {
					foundTask = st
					found = true
					break
				}
			}

			if found {
				break
			}
		}

		if !found {
			return json.Marshal(GetTaskResponse{Error: fmt.Sprintf("задача с ID %s не найдена", request.ID)})
		}

		return json.Marshal(GetTaskResponse{Task: foundTask})
	})

	// Инструмент для обновления статуса задачи
	server.RegisterTool("set_task_status", func(params json.RawMessage) (json.RawMessage, error) {
		var request UpdateTaskStatusRequest
		if err := json.Unmarshal(params, &request); err != nil {
			return json.Marshal(UpdateTaskStatusResponse{Success: false, Error: fmt.Sprintf("ошибка разбора параметров: %s", err)})
		}

		if request.ID == "" {
			return json.Marshal(UpdateTaskStatusResponse{Success: false, Error: "не указан ID задачи"})
		}

		if request.Status == "" {
			return json.Marshal(UpdateTaskStatusResponse{Success: false, Error: "не указан статус задачи"})
		}

		workspacePath := request.WorkspacePath
		if workspacePath == "" {
			var err error
			workspacePath, err = os.Getwd()
			if err != nil {
				return json.Marshal(UpdateTaskStatusResponse{Success: false, Error: fmt.Sprintf("не удалось получить рабочую директорию: %s", err)})
			}
		}

		provider, err := NewTaskMasterToolProvider(workspacePath)
		if err != nil {
			return json.Marshal(UpdateTaskStatusResponse{Success: false, Error: fmt.Sprintf("ошибка создания провайдера: %s", err)})
		}

		tasks, err := provider.converter.GetRicochetTaskTasks()
		if err != nil {
			return json.Marshal(UpdateTaskStatusResponse{Success: false, Error: fmt.Sprintf("ошибка получения задач: %s", err)})
		}

		// Ищем задачу с указанным ID
		var taskToUpdate task.RicochetTaskTask
		found := false

		for _, t := range tasks {
			if t.ID == request.ID {
				taskToUpdate = t
				found = true
				break
			}

			// Проверяем подзадачи
			for _, st := range t.Subtasks {
				if st.ID == request.ID {
					taskToUpdate = st
					found = true
					break
				}
			}

			if found {
				break
			}
		}

		if !found {
			return json.Marshal(UpdateTaskStatusResponse{Success: false, Error: fmt.Sprintf("задача с ID %s не найдена", request.ID)})
		}

		// Обновляем статус
		taskToUpdate.Status = request.Status

		// Сохраняем задачу
		err = provider.converter.UpdateRicochetTaskTask(taskToUpdate)
		if err != nil {
			return json.Marshal(UpdateTaskStatusResponse{Success: false, Error: fmt.Sprintf("ошибка обновления задачи: %s", err)})
		}

		return json.Marshal(UpdateTaskStatusResponse{Success: true})
	})

	// Инструмент для синхронизации статуса задачи
	server.RegisterTool("sync_task_status", func(params json.RawMessage) (json.RawMessage, error) {
		var request SyncTaskStatusRequest
		if err := json.Unmarshal(params, &request); err != nil {
			return json.Marshal(SyncTaskStatusResponse{Success: false, Error: fmt.Sprintf("ошибка разбора параметров: %s", err)})
		}

		if request.TaskID == "" {
			return json.Marshal(SyncTaskStatusResponse{Success: false, Error: "не указан ID задачи"})
		}

		workspacePath := request.WorkspacePath
		if workspacePath == "" {
			var err error
			workspacePath, err = os.Getwd()
			if err != nil {
				return json.Marshal(SyncTaskStatusResponse{Success: false, Error: fmt.Sprintf("не удалось получить рабочую директорию: %s", err)})
			}
		}

		provider, err := NewTaskMasterToolProvider(workspacePath)
		if err != nil {
			return json.Marshal(SyncTaskStatusResponse{Success: false, Error: fmt.Sprintf("ошибка создания провайдера: %s", err)})
		}

		// Если chainID не указан, пытаемся найти его по задаче
		chainID := request.ChainID
		if chainID == "" {
			chainID, err = provider.converter.GetChainForTask(request.TaskID)
			if err != nil {
				return json.Marshal(SyncTaskStatusResponse{Success: false, Error: fmt.Sprintf("ошибка получения цепочки для задачи: %s", err)})
			}
		}

		// Синхронизируем статус
		err = provider.converter.SyncTaskStatus(request.TaskID, chainID)
		if err != nil {
			return json.Marshal(SyncTaskStatusResponse{Success: false, Error: fmt.Sprintf("ошибка синхронизации статуса: %s", err)})
		}

		return json.Marshal(SyncTaskStatusResponse{Success: true})
	})

	// Инструмент для создания задачи из цепочки
	server.RegisterTool("create_task_from_chain", func(params json.RawMessage) (json.RawMessage, error) {
		var request CreateTaskFromChainRequest
		if err := json.Unmarshal(params, &request); err != nil {
			return json.Marshal(CreateTaskFromChainResponse{Error: fmt.Sprintf("ошибка разбора параметров: %s", err)})
		}

		if request.ChainID == "" {
			return json.Marshal(CreateTaskFromChainResponse{Error: "не указан ID цепочки"})
		}

		workspacePath := request.WorkspacePath
		if workspacePath == "" {
			var err error
			workspacePath, err = os.Getwd()
			if err != nil {
				return json.Marshal(CreateTaskFromChainResponse{Error: fmt.Sprintf("не удалось получить рабочую директорию: %s", err)})
			}
		}

		provider, err := NewTaskMasterToolProvider(workspacePath)
		if err != nil {
			return json.Marshal(CreateTaskFromChainResponse{Error: fmt.Sprintf("ошибка создания провайдера: %s", err)})
		}

		// Создаем задачу из цепочки
		taskID, err := provider.converter.CreateTaskFromChain(request.ChainID)
		if err != nil {
			return json.Marshal(CreateTaskFromChainResponse{Error: fmt.Sprintf("ошибка создания задачи: %s", err)})
		}

		return json.Marshal(CreateTaskFromChainResponse{TaskID: taskID})
	})

	// Инструмент для создания цепочки из задачи
	server.RegisterTool("create_chain_from_task", func(params json.RawMessage) (json.RawMessage, error) {
		var request CreateChainFromTaskRequest
		if err := json.Unmarshal(params, &request); err != nil {
			return json.Marshal(CreateChainFromTaskResponse{Error: fmt.Sprintf("ошибка разбора параметров: %s", err)})
		}

		if request.TaskID == "" {
			return json.Marshal(CreateChainFromTaskResponse{Error: "не указан ID задачи"})
		}

		workspacePath := request.WorkspacePath
		if workspacePath == "" {
			var err error
			workspacePath, err = os.Getwd()
			if err != nil {
				return json.Marshal(CreateChainFromTaskResponse{Error: fmt.Sprintf("не удалось получить рабочую директорию: %s", err)})
			}
		}

		provider, err := NewTaskMasterToolProvider(workspacePath)
		if err != nil {
			return json.Marshal(CreateChainFromTaskResponse{Error: fmt.Sprintf("ошибка создания провайдера: %s", err)})
		}

		// Создаем цепочку из задачи
		chainID, err := provider.converter.CreateChainFromTask(request.TaskID)
		if err != nil {
			return json.Marshal(CreateChainFromTaskResponse{Error: fmt.Sprintf("ошибка создания цепочки: %s", err)})
		}

		return json.Marshal(CreateChainFromTaskResponse{ChainID: chainID})
	})
}
