package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/grik-ai/ricochet-task/pkg/chain"
)

// RicochetTaskTask представляет структуру задачи из Ricochet Task
type RicochetTaskTask struct {
	ID           string                 `json:"id"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	Status       string                 `json:"status"`
	Dependencies []string               `json:"dependencies"`
	Priority     string                 `json:"priority"`
	Details      string                 `json:"details"`
	TestStrategy string                 `json:"testStrategy"`
	Subtasks     []RicochetTaskTask     `json:"subtasks,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// RicochetTaskConfig представляет конфигурацию Ricochet Task
type RicochetTaskConfig struct {
	Main       ModelConfig            `json:"main"`
	Research   ModelConfig            `json:"research"`
	Fallback   ModelConfig            `json:"fallback"`
	ChainRoles map[string]ModelConfig `json:"chain_roles"`
}

// ModelConfig представляет конфигурацию модели в Ricochet Task
type ModelConfig struct {
	Provider    string                 `json:"provider"`
	ModelID     string                 `json:"model_id"`
	DisplayName string                 `json:"display_name"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// RicochetTaskStatus определяет возможные статусы задач в Ricochet Task
const (
	RicochetTaskStatusPending  = "pending"
	RicochetTaskStatusProgress = "in-progress"
	RicochetTaskStatusDone     = "done"
	RicochetTaskStatusDeferred = "deferred"
	RicochetTaskStatusBlocked  = "blocked"
	RicochetTaskStatusReview   = "review"
)

// TaskMasterPriority определяет возможные приоритеты задач в Task Master
const (
	TaskMasterPriorityHigh   = "high"
	TaskMasterPriorityMedium = "medium"
	TaskMasterPriorityLow    = "low"
)

// RicochetTaskPriority определяет возможные приоритеты задач в Ricochet Task
const (
	RicochetTaskPriorityHigh   = "high"
	RicochetTaskPriorityMedium = "medium"
	RicochetTaskPriorityLow    = "low"
)

// RicochetTaskConverter интерфейс для преобразования между Ricochet Task и Ricochet
type RicochetTaskConverter interface {
	// ConvertTaskToChain преобразует задачу Ricochet Task в цепочку Ricochet
	ConvertTaskToChain(task RicochetTaskTask) (chain.Chain, error)

	// ConvertChainToTask преобразует цепочку Ricochet в задачу Ricochet Task
	ConvertChainToTask(c chain.Chain) (RicochetTaskTask, error)

	// SyncTaskStatus синхронизирует статус задачи с прогрессом выполнения цепочки
	SyncTaskStatus(taskID string, chainID string) error

	// GetRicochetTaskTasks возвращает список задач из Ricochet Task
	GetRicochetTaskTasks() ([]RicochetTaskTask, error)

	// UpdateRicochetTaskTask обновляет задачу в Ricochet Task
	UpdateRicochetTaskTask(task RicochetTaskTask) error
}

// DefaultRicochetTaskConverter реализация RicochetTaskConverter
type DefaultRicochetTaskConverter struct {
	tasksPath     string
	configPath    string
	taskManager   TaskManager
	chainStore    chain.Store
	workspacePath string
}

// NewRicochetTaskConverter создает новый конвертер Ricochet Task
func NewRicochetTaskConverter(
	workspacePath string,
	taskManager TaskManager,
	chainStore chain.Store,
) (*DefaultRicochetTaskConverter, error) {
	tasksPath := filepath.Join(workspacePath, ".taskmaster", "tasks", "tasks.json")
	configPath := filepath.Join(workspacePath, ".taskmaster", "config.json")

	return &DefaultRicochetTaskConverter{
		tasksPath:     tasksPath,
		configPath:    configPath,
		taskManager:   taskManager,
		chainStore:    chainStore,
		workspacePath: workspacePath,
	}, nil
}

// GetRicochetTaskTasks возвращает список задач из Ricochet Task
func (c *DefaultRicochetTaskConverter) GetRicochetTaskTasks() ([]RicochetTaskTask, error) {
	// Проверяем, существует ли файл задач
	if _, err := os.Stat(c.tasksPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("файл задач Ricochet Task не найден: %w", err)
	}

	// Читаем файл задач
	data, err := os.ReadFile(c.tasksPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл задач: %w", err)
	}

	// Разбираем JSON
	var tasks []RicochetTaskTask
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("не удалось разобрать файл задач: %w", err)
	}

	return tasks, nil
}

// UpdateRicochetTaskTask обновляет задачу в Ricochet Task
func (c *DefaultRicochetTaskConverter) UpdateRicochetTaskTask(task RicochetTaskTask) error {
	// Получаем все задачи
	tasks, err := c.GetRicochetTaskTasks()
	if err != nil {
		return err
	}

	// Ищем задачу с указанным ID
	found := false
	for i, t := range tasks {
		if t.ID == task.ID {
			tasks[i] = task
			found = true
			break
		}

		// Проверяем подзадачи
		for j, st := range t.Subtasks {
			if st.ID == task.ID {
				tasks[i].Subtasks[j] = task
				found = true
				break
			}
		}

		if found {
			break
		}
	}

	if !found {
		return fmt.Errorf("задача с ID %s не найдена", task.ID)
	}

	// Сохраняем обновленные задачи
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("не удалось сериализовать задачи: %w", err)
	}

	// Создаем директорию, если она не существует
	tasksDir := filepath.Dir(c.tasksPath)
	if _, err := os.Stat(tasksDir); os.IsNotExist(err) {
		if err := os.MkdirAll(tasksDir, 0755); err != nil {
			return fmt.Errorf("не удалось создать директорию для задач: %w", err)
		}
	}

	// Записываем файл
	if err := os.WriteFile(c.tasksPath, data, 0644); err != nil {
		return fmt.Errorf("не удалось записать файл задач: %w", err)
	}

	return nil
}

// GetChainForTask возвращает ID цепочки для задачи
func (c *DefaultRicochetTaskConverter) GetChainForTask(taskID string) (string, error) {
	// Получаем задачи из Ricochet Task
	tasks, err := c.GetRicochetTaskTasks()
	if err != nil {
		return "", err
	}

	// Ищем задачу с указанным ID
	for _, task := range tasks {
		if task.ID == taskID {
			if task.Metadata != nil {
				if chainID, ok := task.Metadata["chain_id"].(string); ok && chainID != "" {
					return chainID, nil
				}
			}
		}

		// Проверяем подзадачи
		for _, subtask := range task.Subtasks {
			if subtask.ID == taskID {
				if subtask.Metadata != nil {
					if chainID, ok := subtask.Metadata["chain_id"].(string); ok && chainID != "" {
						return chainID, nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("цепочка для задачи с ID %s не найдена", taskID)
}

func (c *DefaultRicochetTaskConverter) ConvertTaskToChain(task RicochetTaskTask) (chain.Chain, error) {
	// Определяем ID цепочки (создаем новый, если нет в метаданных)
	chainID := ""
	if task.Metadata != nil {
		if id, ok := task.Metadata["chain_id"].(string); ok && id != "" {
			chainID = id
		}
	}
	if chainID == "" {
		chainID = uuid.New().String()
	}

	// Создаем модели для цепочки на основе подзадач
	var models []chain.Model

	// Если нет подзадач, создаем цепочку с одной моделью
	if len(task.Subtasks) == 0 {
		// Создаем модель на основе заголовка и описания задачи
		model := chain.Model{
			ID:    uuid.New().String(),
			Name:  chain.ModelNameGPT4,
			Type:  chain.ModelTypeOpenAI,
			Role:  chain.ModelRoleAnalyzer,
			Order: 0,
			Prompt: fmt.Sprintf(
				"Задача: %s\nОписание: %s\nДетали: %s",
				task.Title, task.Description, task.Details,
			),
			MaxTokens:   4096,
			Temperature: 0.7,
			Parameters: chain.Parameters{
				Temperature:      0.7,
				TopP:             1.0,
				FrequencyPenalty: 0.0,
				PresencePenalty:  0.0,
			},
		}
		models = append(models, model)
	} else {
		// Создаем модели на основе подзадач
		for i, subtask := range task.Subtasks {
			// Определяем тип модели
			modelName := chain.ModelNameGPT4
			modelType := chain.ModelTypeOpenAI
			modelRole := chain.ModelRoleAnalyzer

			// Если есть метаданные, извлекаем из них информацию о модели
			if subtask.Metadata != nil {
				if name, ok := subtask.Metadata["model_name"].(string); ok && name != "" {
					modelName = chain.ModelName(name)
				}
				if provider, ok := subtask.Metadata["model_provider"].(string); ok && provider != "" {
					modelType = chain.ModelType(provider)
				}
				if role, ok := subtask.Metadata["model_role"].(string); ok && role != "" {
					modelRole = chain.ModelRole(role)
				}
			}

			// Создаем модель
			model := chain.Model{
				ID:    uuid.New().String(),
				Name:  modelName,
				Type:  modelType,
				Role:  modelRole,
				Order: i,
				Prompt: fmt.Sprintf(
					"Подзадача: %s\nОписание: %s\nДетали: %s",
					subtask.Title, subtask.Description, subtask.Details,
				),
				MaxTokens:   4096,
				Temperature: 0.7,
				Parameters: chain.Parameters{
					Temperature:      0.7,
					TopP:             1.0,
					FrequencyPenalty: 0.0,
					PresencePenalty:  0.0,
				},
			}
			models = append(models, model)
		}
	}

	newChain := chain.Chain{
		ID:          chainID,
		Name:        task.Title,
		Description: task.Description,
		Models:      models,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tags:        task.Tags,
		Metadata: chain.Metadata{
			Author:      "RicochetTask",
			Version:     "1.0",
			UseCase:     "RicochetTaskIntegration",
			InputFormat: "text",
			Custom: map[string]interface{}{
				"task_id":     task.ID,
				"task_status": task.Status,
				"priority":    task.Priority,
			},
		},
	}

	return newChain, nil
}

// ConvertChainToTask преобразует цепочку Ricochet в задачу Ricochet Task
func (c *DefaultRicochetTaskConverter) ConvertChainToTask(ch chain.Chain) (RicochetTaskTask, error) {
	// Определяем ID задачи (если есть в метаданных)
	taskID := ""
	if ch.Metadata.Custom != nil {
		if id, ok := ch.Metadata.Custom["task_id"].(string); ok && id != "" {
			taskID = id
		}
	}
	if taskID == "" {
		// Если ID нет, генерируем новый
		taskID = fmt.Sprintf("%d", time.Now().Unix())
	}

	// Определяем статус задачи (если есть в метаданных)
	taskStatus := TMStatusPending
	if ch.Metadata.Custom != nil {
		if status, ok := ch.Metadata.Custom["task_status"].(string); ok && status != "" {
			taskStatus = status
		}
	}

	// Определяем приоритет задачи (если есть в метаданных)
	taskPriority := TMPriorityMedium
	if ch.Metadata.Custom != nil {
		if priority, ok := ch.Metadata.Custom["priority"].(string); ok && priority != "" {
			taskPriority = priority
		}
	}

	// Создаем подзадачи на основе моделей в цепочке
	var subtasks []RicochetTaskTask
	for i, model := range ch.Models {
		subtaskID := fmt.Sprintf("%s.%d", taskID, i+1)

		// Извлекаем информацию из промпта модели
		promptLines := strings.Split(model.Prompt, "\n")
		subtaskTitle := ""
		subtaskDesc := ""
		subtaskDetails := ""

		for _, line := range promptLines {
			if strings.HasPrefix(line, "Подзадача:") || strings.HasPrefix(line, "Задача:") {
				subtaskTitle = strings.TrimSpace(line[strings.Index(line, ":")+1:])
			} else if strings.HasPrefix(line, "Описание:") {
				subtaskDesc = strings.TrimSpace(line[strings.Index(line, ":")+1:])
			} else {
				subtaskDetails += line + "\n"
			}
		}

		// Если не удалось извлечь заголовок, создаем его на основе роли модели
		if subtaskTitle == "" {
			subtaskTitle = fmt.Sprintf("Шаг %d: %s", i+1, model.Role)
		}

		subtask := RicochetTaskTask{
			ID:          subtaskID,
			Title:       subtaskTitle,
			Description: subtaskDesc,
			Status:      TMStatusPending,
			Priority:    TMPriorityMedium,
			Details:     subtaskDetails,
			Metadata: map[string]interface{}{
				"model_id":       model.ID,
				"model_name":     string(model.Name),
				"model_provider": string(model.Type),
				"model_role":     string(model.Role),
				"chain_id":       ch.ID,
			},
		}

		subtasks = append(subtasks, subtask)
	}

	// Создаем основную задачу
	task := RicochetTaskTask{
		ID:          taskID,
		Title:       ch.Name,
		Description: ch.Description,
		Status:      taskStatus,
		Priority:    taskPriority,
		Details:     fmt.Sprintf("Цепочка Ricochet: %s\n\n%s", ch.ID, ch.Description),
		Subtasks:    subtasks,
		Tags:        ch.Tags,
		Metadata: map[string]interface{}{
			"chain_id":   ch.ID,
			"created_at": ch.CreatedAt,
			"updated_at": ch.UpdatedAt,
		},
	}

	return task, nil
}

// SyncTaskStatus синхронизирует статус задачи с прогрессом выполнения цепочки
func (c *DefaultRicochetTaskConverter) SyncTaskStatus(taskID string, chainID string) error {
	// Получаем задачу из Ricochet Task
	tasks, err := c.GetRicochetTaskTasks()
	if err != nil {
		return err
	}

	// Ищем задачу с указанным ID
	var task *RicochetTaskTask
	var taskIndex int
	var parentIndex int
	var isSubtask bool

	for i, t := range tasks {
		if t.ID == taskID {
			task = &tasks[i]
			taskIndex = i
			break
		}

		// Проверяем подзадачи
		for j, st := range t.Subtasks {
			if st.ID == taskID {
				task = &tasks[i].Subtasks[j]
				parentIndex = i
				isSubtask = true
				break
			}
		}

		if task != nil {
			break
		}
	}

	if task == nil {
		return fmt.Errorf("задача с ID %s не найдена", taskID)
	}

	// Получаем выполненные задачи из Ricochet
	ricochetTasks, err := c.taskManager.ListTasks()
	if err != nil {
		return err
	}

	// Фильтруем задачи по цепочке
	var chainTasks []Task
	for _, t := range ricochetTasks {
		if t.ChainID == chainID {
			chainTasks = append(chainTasks, t)
		}
	}

	// Определяем общий прогресс выполнения
	totalTasks := len(chainTasks)
	completedTasks := 0
	inProgressTasks := 0

	for _, t := range chainTasks {
		if t.Status == StatusCompleted {
			completedTasks++
		} else if t.Status == StatusRunning || t.Status == StatusReady {
			inProgressTasks++
		}
	}

	// Обновляем статус задачи на основе прогресса
	var newStatus string
	if totalTasks == 0 {
		newStatus = TMStatusPending
	} else if completedTasks == totalTasks {
		newStatus = TMStatusDone
	} else if completedTasks > 0 || inProgressTasks > 0 {
		newStatus = TMStatusProgress
	} else {
		newStatus = TMStatusPending
	}

	// Обновляем статус
	task.Status = newStatus

	// Сохраняем обновленную задачу
	if isSubtask {
		return c.UpdateRicochetTaskTask(tasks[parentIndex].Subtasks[taskIndex])
	} else {
		return c.UpdateRicochetTaskTask(tasks[taskIndex])
	}
}

// CreateChainFromTask создает цепочку на основе задачи
func (c *DefaultRicochetTaskConverter) CreateChainFromTask(taskID string) (string, error) {
	// Получаем задачу из Ricochet Task
	tasks, err := c.GetRicochetTaskTasks()
	if err != nil {
		return "", err
	}

	// Ищем задачу с указанным ID
	var task RicochetTaskTask
	found := false

	for _, t := range tasks {
		if t.ID == taskID {
			task = t
			found = true
			break
		}

		// Проверяем подзадачи
		for _, st := range t.Subtasks {
			if st.ID == taskID {
				task = st
				found = true
				break
			}
		}

		if found {
			break
		}
	}

	if !found {
		return "", fmt.Errorf("задача с ID %s не найдена", taskID)
	}

	// Преобразуем задачу в цепочку
	chain, err := c.ConvertTaskToChain(task)
	if err != nil {
		return "", fmt.Errorf("не удалось преобразовать задачу в цепочку: %w", err)
	}

	// Сохраняем цепочку
	if err := c.chainStore.Save(chain); err != nil {
		return "", fmt.Errorf("не удалось сохранить цепочку: %w", err)
	}

	// Обновляем метаданные задачи
	if task.Metadata == nil {
		task.Metadata = make(map[string]interface{})
	}
	task.Metadata["chain_id"] = chain.ID

	// Сохраняем задачу
	if err := c.UpdateRicochetTaskTask(task); err != nil {
		return "", fmt.Errorf("не удалось обновить задачу: %w", err)
	}

	return chain.ID, nil
}

func (c *DefaultRicochetTaskConverter) CreateTaskFromChain(chainID string) (string, error) {
	// Получаем цепочку
	ch, err := c.chainStore.Get(chainID)
	if err != nil {
		return "", fmt.Errorf("не удалось получить цепочку: %w", err)
	}

	// Преобразуем цепочку в задачу
	task, err := c.ConvertChainToTask(ch)
	if err != nil {
		return "", fmt.Errorf("не удалось преобразовать цепочку в задачу: %w", err)
	}

	// Получаем текущие задачи
	tasks, err := c.GetRicochetTaskTasks()
	if err != nil {
		// Если файл задач не существует, создаем новый массив
		if os.IsNotExist(err) {
			tasks = []RicochetTaskTask{}
		} else {
			return "", err
		}
	}

	// Проверяем, существует ли уже задача с такой цепочкой
	for _, t := range tasks {
		if t.Metadata != nil {
			if id, ok := t.Metadata["chain_id"].(string); ok && id == chainID {
				return t.ID, nil
			}
		}
	}

	// Добавляем новую задачу
	// Находим максимальный ID и увеличиваем его на 1
	maxID := 0
	for _, t := range tasks {
		// Преобразуем ID в число (если возможно)
		if id, err := strconv.Atoi(t.ID); err == nil && id > maxID {
			maxID = id
		}
	}
	task.ID = strconv.Itoa(maxID + 1)

	// Обновляем ID подзадач
	for i := range task.Subtasks {
		task.Subtasks[i].ID = fmt.Sprintf("%s.%d", task.ID, i+1)
	}

	tasks = append(tasks, task)

	// Сохраняем обновленные задачи
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return "", fmt.Errorf("не удалось сериализовать задачи: %w", err)
	}

	// Создаем директорию, если она не существует
	tasksDir := filepath.Dir(c.tasksPath)
	if _, err := os.Stat(tasksDir); os.IsNotExist(err) {
		if err := os.MkdirAll(tasksDir, 0755); err != nil {
			return "", fmt.Errorf("не удалось создать директорию для задач: %w", err)
		}
	}

	// Записываем файл
	if err := os.WriteFile(c.tasksPath, data, 0644); err != nil {
		return "", fmt.Errorf("не удалось записать файл задач: %w", err)
	}

	return task.ID, nil
}

// GetTaskForChain возвращает ID задачи для цепочки
func (c *DefaultRicochetTaskConverter) GetTaskForChain(chainID string) (string, error) {
	// Получаем задачи из Ricochet Task
	tasks, err := c.GetRicochetTaskTasks()
	if err != nil {
		return "", err
	}

	// Ищем задачу с указанной цепочкой
	for _, task := range tasks {
		if task.Metadata != nil {
			if id, ok := task.Metadata["chain_id"].(string); ok && id == chainID {
				return task.ID, nil
			}
		}

		// Проверяем подзадачи
		for _, subtask := range task.Subtasks {
			if subtask.Metadata != nil {
				if id, ok := subtask.Metadata["chain_id"].(string); ok && id == chainID {
					return subtask.ID, nil
				}
			}
		}
	}

	return "", fmt.Errorf("задача для цепочки с ID %s не найдена", chainID)
}
