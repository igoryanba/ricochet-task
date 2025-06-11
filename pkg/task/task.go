package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/grik-ai/ricochet-task/pkg/chain"
)

// TaskStatus определяет статус задачи
type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"   // Задача создана, но не может быть запущена из-за невыполненных зависимостей
	StatusReady     TaskStatus = "ready"     // Задача готова к выполнению, все зависимости удовлетворены
	StatusRunning   TaskStatus = "running"   // Задача в процессе выполнения
	StatusCompleted TaskStatus = "completed" // Задача успешно завершена
	StatusFailed    TaskStatus = "failed"    // Задача завершилась с ошибкой
	StatusPaused    TaskStatus = "paused"    // Выполнение задачи временно приостановлено
	StatusCancelled TaskStatus = "cancelled" // Задача отменена пользователем
)

// Constants for TaskMaster statuses
const (
	TMStatusPending  = "pending"
	TMStatusProgress = "in-progress"
	TMStatusDone     = "done"
	TMStatusDeferred = "deferred"
	TMStatusBlocked  = "blocked"
	TMStatusReview   = "review"
)

// TaskMasterPriority определяет возможные приоритеты задач в Task Master
const (
	TMPriorityHigh   = "high"
	TMPriorityMedium = "medium"
	TMPriorityLow    = "low"
)

// TaskType определяет тип задачи
type TaskType string

const (
	TaskTypeModelExecution TaskType = "model_execution" // Выполнение запроса к модели
	TaskTypeSegmentation   TaskType = "segmentation"    // Сегментация текста
	TaskTypeIntegration    TaskType = "integration"     // Интеграция результатов
	TaskTypePreprocessing  TaskType = "preprocessing"   // Предобработка данных
)

// TaskInput представляет входные данные для задачи
type TaskInput struct {
	Type      string                 `json:"type"`       // Тип входных данных (text, file, checkpoint)
	Source    string                 `json:"source"`     // Источник данных (путь к файлу, ID чекпоинта)
	Segment   int                    `json:"segment"`    // Номер сегмента (если применимо)
	Metadata  map[string]interface{} `json:"metadata"`   // Метаданные
	ChunkSize int                    `json:"chunk_size"` // Размер чанка для сегментации
}

// TaskOutput представляет выходные данные задачи
type TaskOutput struct {
	Type        string                 `json:"type"`        // Тип выходных данных (text, file, checkpoint)
	Destination string                 `json:"destination"` // Назначение данных (путь к файлу, ID чекпоинта)
	Metadata    map[string]interface{} `json:"metadata"`    // Метаданные
}

// TaskMetrics содержит метрики выполнения задачи
type TaskMetrics struct {
	TokensInput  int     `json:"tokens_input"`  // Количество входных токенов
	TokensOutput int     `json:"tokens_output"` // Количество выходных токенов
	DurationMs   int64   `json:"duration_ms"`   // Длительность выполнения в миллисекундах
	Cost         float64 `json:"cost"`          // Стоимость выполнения в долларах
}

// Task представляет собой задачу в системе Ricochet
type Task struct {
	ID           string                 `json:"id"`           // Уникальный идентификатор
	Type         TaskType               `json:"type"`         // Тип задачи
	Title        string                 `json:"title"`        // Название задачи
	Description  string                 `json:"description"`  // Описание задачи
	Status       TaskStatus             `json:"status"`       // Статус задачи
	Dependencies []string               `json:"dependencies"` // ID зависимых задач
	Model        *chain.Model           `json:"model"`        // Модель для выполнения (для TaskTypeModelExecution)
	Input        TaskInput              `json:"input"`        // Входные данные
	Output       TaskOutput             `json:"output"`       // Выходные данные
	Metrics      TaskMetrics            `json:"metrics"`      // Метрики выполнения
	CreatedAt    time.Time              `json:"created_at"`   // Время создания
	StartedAt    *time.Time             `json:"started_at"`   // Время начала выполнения
	CompletedAt  *time.Time             `json:"completed_at"` // Время завершения
	RunID        string                 `json:"run_id"`       // ID запуска цепочки
	ChainID      string                 `json:"chain_id"`     // ID цепочки
	Metadata     map[string]interface{} `json:"metadata"`     // Дополнительные метаданные
}

// Errors
var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrInvalidTaskID     = errors.New("invalid task ID")
	ErrInvalidTaskType   = errors.New("invalid task type")
	ErrInvalidTaskInput  = errors.New("invalid task input")
	ErrTaskAlreadyExists = errors.New("task already exists")
)

// TaskStore интерфейс для хранилища задач
type TaskStore interface {
	// Save сохраняет задачу
	Save(task Task) error

	// Get возвращает задачу по ID
	Get(id string) (Task, error)

	// List возвращает список всех задач
	List() ([]Task, error)

	// ListByRunID возвращает список задач для указанного запуска
	ListByRunID(runID string) ([]Task, error)

	// ListByChainID возвращает список задач для указанной цепочки
	ListByChainID(chainID string) ([]Task, error)

	// ListByStatus возвращает список задач с указанным статусом
	ListByStatus(status TaskStatus) ([]Task, error)

	// Delete удаляет задачу
	Delete(id string) error

	// Exists проверяет существование задачи
	Exists(id string) bool
}

// FileTaskStore реализация хранилища задач в файловой системе
type FileTaskStore struct {
	path string
}

// NewFileTaskStore создает новое хранилище задач в файловой системе
func NewFileTaskStore(configDir string) (*FileTaskStore, error) {
	path := filepath.Join(configDir, "tasks.json")

	// Создаем директорию, если она не существует
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	// Создаем файл, если он не существует
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := saveTasks(path, []Task{}); err != nil {
			return nil, err
		}
	}

	return &FileTaskStore{path: path}, nil
}

// Save сохраняет задачу
func (s *FileTaskStore) Save(task Task) error {
	tasks, err := loadTasks(s.path)
	if err != nil {
		return err
	}

	// Для новой задачи генерируем ID и устанавливаем дату создания
	if task.ID == "" {
		task.ID = uuid.New().String()
		task.CreatedAt = time.Now()
	}

	// Обновляем или добавляем задачу
	found := false
	for i, t := range tasks {
		if t.ID == task.ID {
			tasks[i] = task
			found = true
			break
		}
	}

	if !found {
		tasks = append(tasks, task)
	}

	return saveTasks(s.path, tasks)
}

// Get возвращает задачу по ID
func (s *FileTaskStore) Get(id string) (Task, error) {
	tasks, err := loadTasks(s.path)
	if err != nil {
		return Task{}, err
	}

	for _, task := range tasks {
		if task.ID == id {
			return task, nil
		}
	}

	return Task{}, fmt.Errorf("task with ID '%s' not found", id)
}

// List возвращает список всех задач
func (s *FileTaskStore) List() ([]Task, error) {
	return loadTasks(s.path)
}

// ListByRunID возвращает список задач для указанного запуска
func (s *FileTaskStore) ListByRunID(runID string) ([]Task, error) {
	tasks, err := loadTasks(s.path)
	if err != nil {
		return nil, err
	}

	var result []Task
	for _, task := range tasks {
		if task.RunID == runID {
			result = append(result, task)
		}
	}

	return result, nil
}

// ListByChainID возвращает список задач для указанной цепочки
func (s *FileTaskStore) ListByChainID(chainID string) ([]Task, error) {
	tasks, err := loadTasks(s.path)
	if err != nil {
		return nil, err
	}

	var result []Task
	for _, task := range tasks {
		if task.ChainID == chainID {
			result = append(result, task)
		}
	}

	return result, nil
}

// ListByStatus возвращает список задач с указанным статусом
func (s *FileTaskStore) ListByStatus(status TaskStatus) ([]Task, error) {
	tasks, err := loadTasks(s.path)
	if err != nil {
		return nil, err
	}

	var result []Task
	for _, task := range tasks {
		if task.Status == status {
			result = append(result, task)
		}
	}

	return result, nil
}

// Delete удаляет задачу
func (s *FileTaskStore) Delete(id string) error {
	tasks, err := loadTasks(s.path)
	if err != nil {
		return err
	}

	var newTasks []Task
	found := false

	for _, task := range tasks {
		if task.ID != id {
			newTasks = append(newTasks, task)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("task with ID '%s' not found", id)
	}

	return saveTasks(s.path, newTasks)
}

// Exists проверяет существование задачи
func (s *FileTaskStore) Exists(id string) bool {
	tasks, err := loadTasks(s.path)
	if err != nil {
		return false
	}

	for _, task := range tasks {
		if task.ID == id {
			return true
		}
	}

	return false
}

// TaskManager интерфейс для управления задачами
type TaskManager interface {
	// CreateTask создает новую задачу
	CreateTask(task Task) (string, error)

	// UpdateTaskStatus обновляет статус задачи
	UpdateTaskStatus(taskID string, status TaskStatus) error

	// GetTask возвращает задачу по ID
	GetTask(taskID string) (Task, error)

	// ListTasks возвращает список всех задач
	ListTasks() ([]Task, error)

	// DeleteTask удаляет задачу
	DeleteTask(taskID string) error

	// GetTaskDependencies возвращает список зависимых задач
	GetTaskDependencies(taskID string) ([]Task, error)

	// GetDependentTasks возвращает список задач, зависящих от указанной
	GetDependentTasks(taskID string) ([]Task, error)

	// IsTaskReady проверяет, готова ли задача к выполнению
	IsTaskReady(taskID string) (bool, error)
}

// DefaultTaskManager реализация TaskManager
type DefaultTaskManager struct {
	taskStore TaskStore
}

// NewTaskManager создает новый менеджер задач
func NewTaskManager(taskStore TaskStore) *DefaultTaskManager {
	return &DefaultTaskManager{
		taskStore: taskStore,
	}
}

// CreateTask создает новую задачу
func (m *DefaultTaskManager) CreateTask(task Task) (string, error) {
	// Валидация задачи
	if task.Type == "" {
		return "", ErrInvalidTaskType
	}

	// Для новой задачи устанавливаем начальный статус
	if task.Status == "" {
		// Если есть зависимости, устанавливаем статус "pending"
		if len(task.Dependencies) > 0 {
			task.Status = StatusPending
		} else {
			task.Status = StatusReady
		}
	}

	// Сохраняем задачу
	if err := m.taskStore.Save(task); err != nil {
		return "", err
	}

	return task.ID, nil
}

// UpdateTaskStatus обновляет статус задачи
func (m *DefaultTaskManager) UpdateTaskStatus(taskID string, status TaskStatus) error {
	task, err := m.taskStore.Get(taskID)
	if err != nil {
		return err
	}

	// Обновляем статус и соответствующие временные метки
	task.Status = status

	// Обновляем временные метки в зависимости от статуса
	now := time.Now()
	switch status {
	case StatusRunning:
		task.StartedAt = &now
	case StatusCompleted, StatusFailed, StatusCancelled:
		task.CompletedAt = &now
	}

	// Сохраняем изменения
	return m.taskStore.Save(task)
}

// GetTask возвращает задачу по ID
func (m *DefaultTaskManager) GetTask(taskID string) (Task, error) {
	return m.taskStore.Get(taskID)
}

// ListTasks возвращает список всех задач
func (m *DefaultTaskManager) ListTasks() ([]Task, error) {
	return m.taskStore.List()
}

// DeleteTask удаляет задачу
func (m *DefaultTaskManager) DeleteTask(taskID string) error {
	// Проверяем, есть ли задачи, зависящие от удаляемой
	dependentTasks, err := m.GetDependentTasks(taskID)
	if err != nil {
		return err
	}

	// Если есть зависимые задачи, возвращаем ошибку
	if len(dependentTasks) > 0 {
		return fmt.Errorf("cannot delete task: there are %d dependent tasks", len(dependentTasks))
	}

	return m.taskStore.Delete(taskID)
}

// GetTaskDependencies возвращает список зависимых задач
func (m *DefaultTaskManager) GetTaskDependencies(taskID string) ([]Task, error) {
	task, err := m.taskStore.Get(taskID)
	if err != nil {
		return nil, err
	}

	var dependencies []Task
	for _, depID := range task.Dependencies {
		dep, err := m.taskStore.Get(depID)
		if err != nil {
			return nil, fmt.Errorf("dependency task %s not found: %w", depID, err)
		}
		dependencies = append(dependencies, dep)
	}

	return dependencies, nil
}

// GetDependentTasks возвращает список задач, зависящих от указанной
func (m *DefaultTaskManager) GetDependentTasks(taskID string) ([]Task, error) {
	tasks, err := m.taskStore.List()
	if err != nil {
		return nil, err
	}

	var dependentTasks []Task
	for _, task := range tasks {
		for _, depID := range task.Dependencies {
			if depID == taskID {
				dependentTasks = append(dependentTasks, task)
				break
			}
		}
	}

	return dependentTasks, nil
}

// IsTaskReady проверяет, готова ли задача к выполнению
func (m *DefaultTaskManager) IsTaskReady(taskID string) (bool, error) {
	task, err := m.taskStore.Get(taskID)
	if err != nil {
		return false, err
	}

	// Задача должна быть в статусе "pending"
	if task.Status != StatusPending {
		return false, nil
	}

	// Проверяем, все ли зависимости выполнены
	for _, depID := range task.Dependencies {
		dep, err := m.taskStore.Get(depID)
		if err != nil {
			return false, fmt.Errorf("dependency task %s not found: %w", depID, err)
		}

		// Зависимость должна быть выполнена
		if dep.Status != StatusCompleted {
			return false, nil
		}
	}

	return true, nil
}

// Helper functions
func loadTasks(path string) ([]Task, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Task{}, nil
		}
		return nil, err
	}

	var tasks []Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func saveTasks(path string, tasks []Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}
