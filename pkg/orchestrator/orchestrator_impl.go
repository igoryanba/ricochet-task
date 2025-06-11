package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/grik-ai/ricochet-task/pkg/api"
	"github.com/grik-ai/ricochet-task/pkg/chain"
	"github.com/grik-ai/ricochet-task/pkg/checkpoint"
	"github.com/grik-ai/ricochet-task/pkg/key"
	"github.com/grik-ai/ricochet-task/pkg/model"
	"github.com/grik-ai/ricochet-task/pkg/segmentation"
	"github.com/grik-ai/ricochet-task/pkg/task"
)

// DefaultOrchestrator реализует интерфейс Orchestrator
type DefaultOrchestrator struct {
	apiClient       *api.Client
	keyStore        key.Store
	chainStore      chain.Store
	checkpointStore checkpoint.Store
	taskManager     task.TaskManager
	taskExecutor    task.TaskExecutor
	modelFactory    *model.ProviderFactory
	runs            map[string]*RunMetadata
	mutex           sync.RWMutex
}

// NewOrchestrator создает новый оркестратор
func NewOrchestrator(
	apiClient *api.Client,
	keyStore key.Store,
	chainStore chain.Store,
	checkpointStore checkpoint.Store,
	taskManager task.TaskManager,
	taskExecutor task.TaskExecutor,
	modelFactory *model.ProviderFactory,
) *DefaultOrchestrator {
	return &DefaultOrchestrator{
		apiClient:       apiClient,
		keyStore:        keyStore,
		chainStore:      chainStore,
		checkpointStore: checkpointStore,
		taskManager:     taskManager,
		taskExecutor:    taskExecutor,
		modelFactory:    modelFactory,
		runs:            make(map[string]*RunMetadata),
	}
}

// RunChain запускает цепочку моделей с указанными входными данными
func (o *DefaultOrchestrator) RunChain(ctx context.Context, chainID string, input TaskInput, options ProcessingOptions) (string, error) {
	// Проверяем существование цепочки
	chainObj, err := o.chainStore.Get(chainID)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrChainNotFound, err)
	}

	// Проверяем валидность входных данных
	if err := validateInput(input); err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Создаем ID для запуска
	runID := uuid.New().String()

	// Создаем метаданные запуска
	runMetadata := &RunMetadata{
		ID:          runID,
		ChainID:     chainID,
		Status:      StatusPending,
		StartTime:   time.Now(),
		Progress:    0.0,
		Checkpoints: []string{},
	}

	// Сохраняем метаданные запуска
	o.mutex.Lock()
	o.runs[runID] = runMetadata
	o.mutex.Unlock()

	// Обновляем статус запуска
	runMetadata.Status = StatusRunning

	// Запускаем горутину для выполнения цепочки
	go func() {
		err := o.executeChain(ctx, chainObj, input, options, runID)
		o.mutex.Lock()
		if err != nil {
			runMetadata.Status = StatusFailed
			runMetadata.Error = err.Error()
		} else {
			runMetadata.Status = StatusCompleted
		}
		runMetadata.EndTime = time.Now()
		o.mutex.Unlock()
	}()

	return runID, nil
}

// GetRunStatus возвращает статус выполнения
func (o *DefaultOrchestrator) GetRunStatus(runID string) (*RunMetadata, error) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	metadata, exists := o.runs[runID]
	if !exists {
		return nil, ErrRunNotFound
	}

	return metadata, nil
}

// CancelRun отменяет выполнение
func (o *DefaultOrchestrator) CancelRun(runID string) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	metadata, exists := o.runs[runID]
	if !exists {
		return ErrRunNotFound
	}

	if metadata.Status != StatusRunning && metadata.Status != StatusProcessing {
		return fmt.Errorf("run is not in progress: %s", metadata.Status)
	}

	metadata.Status = StatusCancelled
	metadata.EndTime = time.Now()

	// Отменяем выполняемые задачи
	tasks, err := o.taskManager.ListTasks()
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	for _, t := range tasks {
		if t.RunID == runID && (t.Status == task.StatusRunning || t.Status == task.StatusReady) {
			if err := o.taskExecutor.CancelTask(t.ID); err != nil {
				// Продолжаем отменять другие задачи
				fmt.Printf("Failed to cancel task %s: %v\n", t.ID, err)
			}
		}
	}

	return nil
}

// ListRuns возвращает список всех выполнений
func (o *DefaultOrchestrator) ListRuns() []*RunMetadata {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	runs := make([]*RunMetadata, 0, len(o.runs))
	for _, run := range o.runs {
		runs = append(runs, run)
	}

	return runs
}

// GetRunResults возвращает результаты выполнения
func (o *DefaultOrchestrator) GetRunResults(runID string) (TaskOutput, error) {
	o.mutex.RLock()
	metadata, exists := o.runs[runID]
	o.mutex.RUnlock()

	if !exists {
		return TaskOutput{}, ErrRunNotFound
	}

	if metadata.Status != StatusCompleted {
		return TaskOutput{}, fmt.Errorf("run is not completed: %s", metadata.Status)
	}

	// Ищем задачи, относящиеся к этому запуску
	tasks, err := o.taskManager.ListTasks()
	if err != nil {
		return TaskOutput{}, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Находим последнюю завершенную задачу
	var lastTask *task.Task
	for i := range tasks {
		if tasks[i].RunID == runID && tasks[i].Status == task.StatusCompleted {
			if lastTask == nil || tasks[i].CompletedAt.After(*lastTask.CompletedAt) {
				lastTask = &tasks[i]
			}
		}
	}

	if lastTask == nil {
		return TaskOutput{}, fmt.Errorf("no completed tasks found for run %s", runID)
	}

	// Преобразуем выходные данные задачи в формат TaskOutput
	output := TaskOutput{
		Text:     lastTask.Output.Destination,
		Metadata: lastTask.Output.Metadata,
	}

	return output, nil
}

// GetCheckpoint возвращает чекпоинт
func (o *DefaultOrchestrator) GetCheckpoint(checkpointID string) (checkpoint.Checkpoint, error) {
	return o.checkpointStore.Get(checkpointID)
}

// ListCheckpoints возвращает список чекпоинтов для указанного выполнения
func (o *DefaultOrchestrator) ListCheckpoints(runID string) ([]checkpoint.Checkpoint, error) {
	o.mutex.RLock()
	metadata, exists := o.runs[runID]
	o.mutex.RUnlock()

	if !exists {
		return nil, ErrRunNotFound
	}

	checkpoints := make([]checkpoint.Checkpoint, 0, len(metadata.Checkpoints))
	for _, id := range metadata.Checkpoints {
		cp, err := o.checkpointStore.Get(id)
		if err == nil {
			checkpoints = append(checkpoints, cp)
		}
	}

	return checkpoints, nil
}

// executeChain выполняет цепочку моделей
func (o *DefaultOrchestrator) executeChain(ctx context.Context, chain chain.Chain, input TaskInput, options ProcessingOptions, runID string) error {
	// Получаем метаданные запуска
	o.mutex.RLock()
	metadata := o.runs[runID]
	o.mutex.RUnlock()

	// Проверяем, не отменено ли выполнение
	if metadata.Status == StatusCancelled {
		return ErrRunCancelled
	}

	// Подготавливаем входные данные
	inputText := input.Text
	if len(input.Files) > 0 {
		// TODO: Обработка файлов
	}

	// Определяем, нужна ли сегментация
	needsSegmentation := len(inputText) > options.MaxTokensPerChunk

	// Задачи для выполнения
	var taskIDs []string

	// Если нужна сегментация, создаем задачу сегментации
	if needsSegmentation {
		segmentationTaskID, err := o.createSegmentationTask(inputText, options, runID, chain.ID)
		if err != nil {
			return fmt.Errorf("failed to create segmentation task: %w", err)
		}
		taskIDs = append(taskIDs, segmentationTaskID)
	} else {
		// Создаем задачи для каждой модели в цепочке
		previousTaskID := ""
		for _, model := range chain.Models {
			taskID, err := o.createModelTask(inputText, model, runID, chain.ID, previousTaskID)
			if err != nil {
				return fmt.Errorf("failed to create model task: %w", err)
			}
			taskIDs = append(taskIDs, taskID)
			previousTaskID = taskID
		}
	}

	// Запускаем выполнение задач
	for _, taskID := range taskIDs {
		// Проверяем, не отменено ли выполнение
		o.mutex.RLock()
		if metadata.Status == StatusCancelled {
			o.mutex.RUnlock()
			return ErrRunCancelled
		}
		o.mutex.RUnlock()

		// Запускаем задачу
		if err := o.taskExecutor.ExecuteTask(ctx, taskID); err != nil {
			return fmt.Errorf("task execution failed: %w", err)
		}
	}

	return nil
}

// createSegmentationTask создает задачу сегментации
func (o *DefaultOrchestrator) createSegmentationTask(inputText string, options ProcessingOptions, runID string, chainID string) (string, error) {
	segTask := task.Task{
		Type:        task.TaskTypeSegmentation,
		Title:       "Сегментация текста",
		Description: "Разбиение текста на сегменты для обработки",
		Status:      task.StatusReady,
		Input: task.TaskInput{
			Type:      "text",
			Source:    inputText,
			ChunkSize: options.MaxTokensPerChunk,
			Metadata: map[string]interface{}{
				"method": options.SegmentationMethod,
			},
		},
		RunID:   runID,
		ChainID: chainID,
	}

	return o.taskManager.CreateTask(segTask)
}

// createModelTask создает задачу выполнения модели
func (o *DefaultOrchestrator) createModelTask(inputText string, model chain.Model, runID string, chainID string, dependsOn string) (string, error) {
	modelTask := task.Task{
		Type:        task.TaskTypeModelExecution,
		Title:       fmt.Sprintf("Выполнение модели %s", model.Name),
		Description: fmt.Sprintf("Обработка текста с использованием модели %s", model.Name),
		Status:      task.StatusReady,
		Model:       &model,
		Input: task.TaskInput{
			Type:   "text",
			Source: inputText,
		},
		RunID:   runID,
		ChainID: chainID,
	}

	// Добавляем зависимость, если указана
	if dependsOn != "" {
		modelTask.Dependencies = []string{dependsOn}
		modelTask.Status = task.StatusPending
	}

	return o.taskManager.CreateTask(modelTask)
}

// validateInput проверяет валидность входных данных
func validateInput(input TaskInput) error {
	if input.Text == "" && len(input.Files) == 0 {
		return errors.New("input text or files must be provided")
	}
	return nil
}

// processChain обрабатывает цепочку моделей
func (o *DefaultOrchestrator) processChain(
	ctx context.Context,
	c chain.Chain,
	input TaskInput,
	runMeta *RunMetadata,
	options ProcessingOptions,
) {
	// Обновляем статус
	o.mutex.Lock()
	runMeta.Status = StatusRunning
	o.mutex.Unlock()

	// Выбираем текущий вход (текст для обработки)
	currentInput := input.Text

	// Обрабатываем каждую модель в цепочке последовательно
	for i, model := range c.Models {
		// Проверка на отмену
		select {
		case <-ctx.Done():
			o.mutex.Lock()
			runMeta.Status = StatusCancelled
			runMeta.Error = "Operation cancelled"
			runMeta.EndTime = time.Now()
			o.mutex.Unlock()
			return
		default:
			// Продолжаем выполнение
		}

		// Обновляем метаданные
		o.mutex.Lock()
		runMeta.CurrentModel = string(model.Name)
		runMeta.Status = StatusProcessing
		// Обновляем прогресс на основе позиции в цепочке
		runMeta.Progress = float64(i) / float64(len(c.Models))
		o.mutex.Unlock()

		// Обрабатываем текст с помощью текущей модели
		result, err := o.processModelWithText(ctx, model, currentInput, runMeta, options)
		if err != nil {
			o.mutex.Lock()
			runMeta.Status = StatusFailed
			runMeta.Error = fmt.Sprintf("Error processing model '%s': %v", model.Name, err)
			runMeta.EndTime = time.Now()
			o.mutex.Unlock()
			return
		}

		// Используем результат как вход для следующей модели
		currentInput = result

		// Создаем чекпоинт с промежуточным результатом
		if options.SaveCheckpoints {
			checkpointID, err := o.createCheckpoint(runMeta.ID, model.ID, currentInput)
			if err != nil {
				// Логируем ошибку, но продолжаем выполнение
				fmt.Printf("Warning: failed to create checkpoint: %v\n", err)
			} else {
				o.mutex.Lock()
				runMeta.Checkpoints = append(runMeta.Checkpoints, checkpointID)
				o.mutex.Unlock()
			}
		}
	}

	// Создаем финальный чекпоинт с результатом
	finalCheckpointID, err := o.createFinalCheckpoint(runMeta.ID, currentInput)
	if err != nil {
		fmt.Printf("Warning: failed to create final checkpoint: %v\n", err)
	} else {
		o.mutex.Lock()
		runMeta.Checkpoints = append(runMeta.Checkpoints, finalCheckpointID)
		o.mutex.Unlock()
	}

	// Обновляем статус завершения
	o.mutex.Lock()
	runMeta.Status = StatusCompleted
	runMeta.Progress = 1.0
	runMeta.EndTime = time.Now()
	runMeta.CurrentModel = ""
	o.mutex.Unlock()
}

// processModelWithText обрабатывает текст с помощью модели
func (o *DefaultOrchestrator) processModelWithText(
	ctx context.Context,
	model chain.Model,
	text string,
	runMeta *RunMetadata,
	options ProcessingOptions,
) (string, error) {
	// Если текст небольшой, обрабатываем его напрямую
	estimatedTokens := int(float64(len(text)) * 0.25) // Примерная оценка: 1 токен ~ 4 символа
	if estimatedTokens <= options.MaxTokensPerChunk {
		return o.processSingleChunk(ctx, model, text, runMeta)
	}

	// Для больших текстов используем сегментацию
	return o.processLargeText(ctx, model, text, runMeta, options)
}

// processSingleChunk обрабатывает один небольшой фрагмент текста
func (o *DefaultOrchestrator) processSingleChunk(
	ctx context.Context,
	model chain.Model,
	text string,
	runMeta *RunMetadata,
) (string, error) {
	// Получаем API-ключ для типа модели
	apiKey, err := o.getAPIKey(model.Type)
	if err != nil {
		return "", err
	}

	// Создаем промпт на основе роли модели и системного промпта
	systemPrompt := model.Prompt
	if systemPrompt == "" {
		// Используем промпт по умолчанию для роли
		systemPrompt = getDefaultPromptForRole(model.Role)
	}

	// Создаем запрос к API
	req := &api.ChatRequest{
		Model:     string(model.Name),
		MaxTokens: model.MaxTokens,
		Messages: []api.ChatMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: text,
			},
		},
	}

	// Отправляем запрос к API
	chatService := api.NewChatService(o.apiClient)
	provider := getProviderFromModelType(model.Type)
	o.apiClient.SetAPIKey(provider, apiKey) // Устанавливаем API-ключ

	resp, err := chatService.SendMessage(req)
	if err != nil {
		return "", err
	}

	// Обновляем счетчик токенов
	o.mutex.Lock()
	runMeta.TotalTokens += estimateTokenCount(text) + estimateTokenCount(resp.Message.Content)
	o.mutex.Unlock()

	return resp.Message.Content, nil
}

// processLargeText обрабатывает большой текст с использованием сегментации
func (o *DefaultOrchestrator) processLargeText(
	ctx context.Context,
	model chain.Model,
	text string,
	runMeta *RunMetadata,
	options ProcessingOptions,
) (string, error) {
	// Создаем сегментер указанного типа
	segmenter, err := segmentation.NewSegmenter(options.SegmentationMethod)
	if err != nil {
		return "", err
	}

	// Разбиваем текст на сегменты
	segments, err := segmenter.Split(text, options.MaxTokensPerChunk)
	if err != nil {
		return "", err
	}

	// Если нет сегментов, возвращаем пустой результат
	if len(segments) == 0 {
		return "", nil
	}

	// Если только один сегмент, обрабатываем его напрямую
	if len(segments) == 1 {
		return o.processSingleChunk(ctx, model, segments[0].Content, runMeta)
	}

	// Создаем канал для результатов обработки сегментов
	type segmentResult struct {
		segment segmentation.SegmentInfo
		result  string
		err     error
	}

	resultChan := make(chan segmentResult, len(segments))

	// Ограничиваем количество параллельных обработок
	semaphore := make(chan struct{}, options.MaxParallelChunks)

	// Запускаем обработку каждого сегмента в отдельной горутине
	var wg sync.WaitGroup
	for _, segment := range segments {
		wg.Add(1)
		go func(seg segmentation.SegmentInfo) {
			defer wg.Done()

			// Ждем доступного слота в семафоре
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Обрабатываем сегмент
			result, err := o.processSingleChunk(ctx, model, seg.Content, runMeta)

			// Отправляем результат в канал
			resultChan <- segmentResult{
				segment: seg,
				result:  result,
				err:     err,
			}
		}(segment)
	}

	// Закрываем канал после завершения всех горутин
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Собираем результаты
	processedSegments := make([]segmentation.SegmentInfo, len(segments))
	for res := range resultChan {
		if res.err != nil {
			return "", res.err
		}

		// Создаем новый сегмент с результатом обработки
		processedSegment := res.segment
		processedSegment.Content = res.result

		// Сохраняем сегмент в массиве результатов
		processedSegments[processedSegment.Order] = processedSegment
	}

	// Объединяем результаты обработки сегментов
	finalResult, err := segmenter.Merge(processedSegments)
	if err != nil {
		return "", err
	}

	return finalResult, nil
}

// createCheckpoint создает чекпоинт с промежуточным результатом
func (o *DefaultOrchestrator) createCheckpoint(runID, modelID, content string) (string, error) {
	checkpointID := uuid.New().String()

	// Создаем чекпоинт
	checkpoint := checkpoint.Checkpoint{
		ID:        checkpointID,
		ChainID:   runID,
		ModelID:   modelID,
		Type:      checkpoint.CheckpointTypeIntermediate,
		Content:   content,
		CreatedAt: time.Now(),
		MetaData:  map[string]interface{}{},
	}

	// Сохраняем чекпоинт
	err := o.checkpointStore.Save(checkpoint)
	if err != nil {
		return "", err
	}

	return checkpointID, nil
}

// createFinalCheckpoint создает финальный чекпоинт с результатом
func (o *DefaultOrchestrator) createFinalCheckpoint(runID, content string) (string, error) {
	checkpointID := uuid.New().String()

	// Создаем чекпоинт
	checkpoint := checkpoint.Checkpoint{
		ID:        checkpointID,
		ChainID:   runID,
		ModelID:   "",
		Type:      checkpoint.CheckpointTypeOutput,
		Content:   content,
		CreatedAt: time.Now(),
		MetaData:  map[string]interface{}{},
	}

	// Сохраняем чекпоинт
	err := o.checkpointStore.Save(checkpoint)
	if err != nil {
		return "", err
	}

	return checkpointID, nil
}

// getAPIKey возвращает API-ключ для указанного типа модели
func (o *DefaultOrchestrator) getAPIKey(modelType chain.ModelType) (string, error) {
	// Получаем список ключей
	keys, err := o.keyStore.List()
	if err != nil {
		return "", err
	}

	// Преобразуем тип модели в провайдера
	var provider string
	switch modelType {
	case chain.ModelTypeOpenAI:
		provider = "openai"
	case chain.ModelTypeClaude:
		provider = "claude"
	case chain.ModelTypeDeepSeek:
		provider = "deepseek"
	case chain.ModelTypeGrok:
		provider = "grok"
	default:
		return "", fmt.Errorf("unknown model type: %s", modelType)
	}

	// Ищем подходящий ключ
	for _, k := range keys {
		if k.Provider == provider {
			// TODO: Реализовать более сложную логику выбора ключа
			// с учетом лимитов, использования и т.д.
			return k.Value, nil
		}
	}

	return "", fmt.Errorf("no API key found for provider: %s", provider)
}

// getProviderFromModelType возвращает тип провайдера для модели
func getProviderFromModelType(modelType chain.ModelType) api.Provider {
	switch modelType {
	case chain.ModelTypeOpenAI:
		return api.ProviderOpenAI
	case chain.ModelTypeClaude:
		return api.ProviderClaude
	case chain.ModelTypeDeepSeek:
		return api.ProviderDeepSeek
	case chain.ModelTypeGrok:
		return api.ProviderGrok
	case chain.ModelTypeMistral:
		return api.ProviderMistral
	default:
		return api.ProviderOpenAI
	}
}

// Вспомогательные функции

// getDefaultPromptForRole возвращает промпт по умолчанию для указанной роли
func getDefaultPromptForRole(role chain.ModelRole) string {
	switch role {
	case chain.ModelRoleAnalyzer:
		return "Вы эксперт по анализу информации. Ваша задача - внимательно изучить предоставленный текст и выделить ключевые темы, концепции, идеи и взаимосвязи. Структурируйте ваш анализ логически и систематически."
	case chain.ModelRoleSummarizer:
		return "Вы эксперт по созданию кратких и содержательных резюме. Ваша задача - обобщить предоставленную информацию, сохраняя все ключевые моменты и выводы. Ваше резюме должно быть информативным и лаконичным."
	case chain.ModelRoleIntegrator:
		return "Вы эксперт по интеграции информации. Ваша задача - объединить различные фрагменты информации в целостную и связную картину. Находите взаимосвязи, устраняйте противоречия и представляйте информацию в логической последовательности."
	case chain.ModelRoleExtractor:
		return "Вы эксперт по извлечению информации. Ваша задача - найти и извлечь конкретные данные, факты, цифры и цитаты из предоставленного текста. Представьте результаты в структурированном виде."
	case chain.ModelRoleOrganizer:
		return "Вы эксперт по организации информации. Ваша задача - структурировать предоставленные данные в логическую и понятную систему. Создайте категории, группы и иерархии для упрощения восприятия информации."
	case chain.ModelRoleEvaluator:
		return "Вы эксперт по оценке информации. Ваша задача - критически проанализировать предоставленный материал, оценить достоверность фактов, обоснованность аргументов и выявить возможные предубеждения или неточности."
	default:
		return "Вы помощник, который отвечает на вопросы пользователя точно и информативно. Анализируйте предоставленную информацию тщательно и старайтесь дать полный и подробный ответ."
	}
}

// estimateTokenCount оценивает количество токенов в тексте
func estimateTokenCount(text string) int {
	// Примерная оценка: 1 токен ~ 4 символа
	return int(float64(len(text)) * 0.25)
}
 