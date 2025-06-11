package task

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/chain"
	"github.com/grik-ai/ricochet-task/pkg/segmentation"
)

// ModelProvider интерфейс для доступа к моделям
type ModelProvider interface {
	// Execute выполняет запрос к модели
	Execute(ctx context.Context, model chain.Model, prompt string, options map[string]interface{}) (string, error)

	// EstimateTokens оценивает количество токенов в тексте
	EstimateTokens(text string) int

	// GetModel возвращает модель по имени
	GetModel(name chain.ModelName) (chain.ModelConfiguration, error)
}

// TaskExecutor интерфейс для выполнения задач
type TaskExecutor interface {
	// ExecuteTask выполняет задачу
	ExecuteTask(ctx context.Context, taskID string) error

	// CancelTask отменяет выполнение задачи
	CancelTask(taskID string) error

	// ExecuteBatch выполняет группу задач
	ExecuteBatch(ctx context.Context, taskIDs []string) error
}

// ExecutorConfig конфигурация исполнителя задач
type ExecutorConfig struct {
	MaxConcurrentTasks int           // Максимальное количество одновременно выполняемых задач
	TaskTimeout        time.Duration // Таймаут выполнения задачи
	RetryCount         int           // Количество попыток в случае ошибки
	RetryDelay         time.Duration // Задержка между попытками
}

// DefaultExecutorConfig возвращает конфигурацию по умолчанию
func DefaultExecutorConfig() ExecutorConfig {
	return ExecutorConfig{
		MaxConcurrentTasks: 5,
		TaskTimeout:        10 * time.Minute,
		RetryCount:         3,
		RetryDelay:         5 * time.Second,
	}
}

// DefaultTaskExecutor реализация TaskExecutor
type DefaultTaskExecutor struct {
	taskManager   TaskManager
	modelProvider ModelProvider
	config        ExecutorConfig
	runningTasks  map[string]context.CancelFunc
	mu            sync.Mutex
}

// NewTaskExecutor создает новый исполнитель задач
func NewTaskExecutor(
	taskManager TaskManager,
	modelProvider ModelProvider,
	config ExecutorConfig,
) *DefaultTaskExecutor {
	return &DefaultTaskExecutor{
		taskManager:   taskManager,
		modelProvider: modelProvider,
		config:        config,
		runningTasks:  make(map[string]context.CancelFunc),
	}
}

// ExecuteTask выполняет задачу
func (e *DefaultTaskExecutor) ExecuteTask(ctx context.Context, taskID string) error {
	// Получаем задачу
	task, err := e.taskManager.GetTask(taskID)
	if err != nil {
		return err
	}

	// Проверяем, готова ли задача к выполнению
	ready, err := e.taskManager.IsTaskReady(taskID)
	if err != nil {
		return err
	}

	if !ready && task.Status != StatusReady {
		return fmt.Errorf("task is not ready for execution: %s", task.Status)
	}

	// Создаем контекст с таймаутом для задачи
	taskCtx, cancel := context.WithTimeout(ctx, e.config.TaskTimeout)

	// Регистрируем задачу в списке выполняемых
	e.mu.Lock()
	e.runningTasks[taskID] = cancel
	e.mu.Unlock()

	// Очищаем после завершения
	defer func() {
		cancel()
		e.mu.Lock()
		delete(e.runningTasks, taskID)
		e.mu.Unlock()
	}()

	// Обновляем статус задачи
	if err := e.taskManager.UpdateTaskStatus(taskID, StatusRunning); err != nil {
		return err
	}

	// Запускаем задачу в зависимости от ее типа
	var executeErr error
	startTime := time.Now()

	switch task.Type {
	case TaskTypeModelExecution:
		executeErr = e.executeModelTask(taskCtx, task)
	case TaskTypeSegmentation:
		executeErr = e.executeSegmentationTask(taskCtx, task)
	case TaskTypeIntegration:
		executeErr = e.executeIntegrationTask(taskCtx, task)
	case TaskTypePreprocessing:
		executeErr = e.executePreprocessingTask(taskCtx, task)
	default:
		executeErr = fmt.Errorf("unsupported task type: %s", task.Type)
	}

	// Обновляем метрики задачи
	task.Metrics.DurationMs = time.Since(startTime).Milliseconds()

	// Обновляем статус задачи в зависимости от результата
	var finalStatus TaskStatus
	if executeErr != nil {
		finalStatus = StatusFailed
		// Сохраняем ошибку в метаданных
		if task.Metadata == nil {
			task.Metadata = make(map[string]interface{})
		}
		task.Metadata["error"] = executeErr.Error()
	} else {
		finalStatus = StatusCompleted
	}

	// Обновляем задачу
	if err := e.taskManager.UpdateTaskStatus(taskID, finalStatus); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return executeErr
}

// CancelTask отменяет выполнение задачи
func (e *DefaultTaskExecutor) CancelTask(taskID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	cancel, exists := e.runningTasks[taskID]
	if !exists {
		return fmt.Errorf("task %s is not running", taskID)
	}

	// Отменяем выполнение
	cancel()

	// Обновляем статус задачи
	if err := e.taskManager.UpdateTaskStatus(taskID, StatusCancelled); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return nil
}

// ExecuteBatch выполняет группу задач
func (e *DefaultTaskExecutor) ExecuteBatch(ctx context.Context, taskIDs []string) error {
	var wg sync.WaitGroup
	taskCh := make(chan string, len(taskIDs))
	errCh := make(chan error, len(taskIDs))

	// Запускаем обработчики задач
	for i := 0; i < e.config.MaxConcurrentTasks; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for taskID := range taskCh {
				if err := e.ExecuteTask(ctx, taskID); err != nil {
					errCh <- fmt.Errorf("failed to execute task %s: %w", taskID, err)
				}
			}
		}()
	}

	// Отправляем задачи на выполнение
	for _, taskID := range taskIDs {
		taskCh <- taskID
	}
	close(taskCh)

	// Ждем завершения всех обработчиков
	wg.Wait()
	close(errCh)

	// Собираем ошибки
	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("batch execution completed with %d errors", len(errs))
	}

	return nil
}

// executeModelTask выполняет задачу типа TaskTypeModelExecution
func (e *DefaultTaskExecutor) executeModelTask(ctx context.Context, task Task) error {
	if task.Model == nil {
		return errors.New("model is not specified for model execution task")
	}

	// Получаем входные данные
	var inputText string
	switch task.Input.Type {
	case "text":
		inputText = task.Input.Source
	case "file":
		// TODO: Реализовать чтение из файла
		return errors.New("file input not implemented yet")
	case "checkpoint":
		// TODO: Реализовать чтение из чекпоинта
		return errors.New("checkpoint input not implemented yet")
	default:
		return fmt.Errorf("unsupported input type: %s", task.Input.Type)
	}

	// Выполняем запрос к модели
	options := make(map[string]interface{})
	options["temperature"] = task.Model.Temperature
	options["max_tokens"] = task.Model.MaxTokens

	// Добавляем системный промпт, если он указан
	if task.Model.Prompt != "" {
		options["system_prompt"] = task.Model.Prompt
	}

	// Оцениваем количество входных токенов
	task.Metrics.TokensInput = e.modelProvider.EstimateTokens(inputText)

	// Выполняем запрос к модели
	output, err := e.modelProvider.Execute(ctx, *task.Model, inputText, options)
	if err != nil {
		return fmt.Errorf("model execution failed: %w", err)
	}

	// Оцениваем количество выходных токенов
	task.Metrics.TokensOutput = e.modelProvider.EstimateTokens(output)

	// Сохраняем результат
	task.Output.Type = "text"
	task.Output.Destination = output

	// TODO: Рассчитать стоимость выполнения запроса
	// task.Metrics.Cost = ...

	return nil
}

// executeSegmentationTask выполняет задачу типа TaskTypeSegmentation
func (e *DefaultTaskExecutor) executeSegmentationTask(_ context.Context, task Task) error {
	// Получаем входные данные
	var inputText string
	switch task.Input.Type {
	case "text":
		inputText = task.Input.Source
	case "file":
		// TODO: Реализовать чтение из файла
		return errors.New("file input not implemented yet")
	case "checkpoint":
		// TODO: Реализовать чтение из чекпоинта
		return errors.New("checkpoint input not implemented yet")
	default:
		return fmt.Errorf("unsupported input type: %s", task.Input.Type)
	}

	// Определяем размер чанка
	chunkSize := task.Input.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 2000 // По умолчанию
	}

	// Определяем метод сегментации
	method := segmentation.MethodSimple
	if methodStr, ok := task.Input.Metadata["method"].(string); ok {
		method = segmentation.SegmentationMethod(methodStr)
	}

	// Создаем опции сегментации
	options := segmentation.SegmentationOptions{
		ChunkSize:   chunkSize,
		MaxSegments: 0, // Без ограничений
		Method:      method,
	}

	// Сегментируем текст
	segments, err := segmentation.Segment(inputText, options)
	if err != nil {
		return fmt.Errorf("segmentation failed: %w", err)
	}

	// Создаем подзадачи для каждого сегмента
	for i, segment := range segments {
		// Создаем новую задачу для обработки сегмента
		segmentTask := Task{
			Type:         TaskTypeModelExecution, // По умолчанию используем модель
			Title:        fmt.Sprintf("%s - Segment %d", task.Title, i+1),
			Description:  fmt.Sprintf("Processing segment %d of %d from task %s", i+1, len(segments), task.ID),
			Dependencies: []string{task.ID},
			Model:        task.Model, // Используем ту же модель
			Input: TaskInput{
				Type:     "text",
				Source:   segment,
				Segment:  i,
				Metadata: task.Input.Metadata,
			},
			RunID:   task.RunID,
			ChainID: task.ChainID,
		}

		// Сохраняем подзадачу
		_, err := e.taskManager.CreateTask(segmentTask)
		if err != nil {
			return fmt.Errorf("failed to create segment task: %w", err)
		}
	}

	// Обновляем выходные данные задачи
	task.Output.Type = "metadata"
	if task.Output.Metadata == nil {
		task.Output.Metadata = make(map[string]interface{})
	}
	task.Output.Metadata["segments_count"] = len(segments)

	return nil
}

// executeIntegrationTask выполняет задачу типа TaskTypeIntegration
func (e *DefaultTaskExecutor) executeIntegrationTask(_ context.Context, task Task) error {
	// Получаем зависимые задачи
	dependencies, err := e.taskManager.GetTaskDependencies(task.ID)
	if err != nil {
		return fmt.Errorf("failed to get task dependencies: %w", err)
	}

	// Собираем результаты зависимых задач
	var results []string
	var totalTokens int

	for _, dep := range dependencies {
		if dep.Status != StatusCompleted {
			return fmt.Errorf("dependency task %s is not completed: %s", dep.ID, dep.Status)
		}

		// Для задач с текстовым выводом добавляем результат
		if dep.Output.Type == "text" {
			results = append(results, dep.Output.Destination)
			totalTokens += dep.Metrics.TokensOutput
		}
	}

	// Объединяем результаты
	combinedResult := ""
	for i, result := range results {
		combinedResult += fmt.Sprintf("=== Результат %d ===\n%s\n\n", i+1, result)
	}

	// Сохраняем результат
	task.Output.Type = "text"
	task.Output.Destination = combinedResult

	// Обновляем метрики
	task.Metrics.TokensInput = totalTokens
	task.Metrics.TokensOutput = e.modelProvider.EstimateTokens(combinedResult)

	return nil
}

// executePreprocessingTask выполняет задачу типа TaskTypePreprocessing
func (e *DefaultTaskExecutor) executePreprocessingTask(_ context.Context, _ Task) error {
	// TODO: Реализовать предобработку данных
	return errors.New("preprocessing task not implemented yet")
}
