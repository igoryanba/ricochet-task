package service

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
)

// Status представляет статус выполнения цепочки
type Status string

// Константы для статусов
const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

// RunMetadata содержит метаданные о запуске цепочки
type RunMetadata struct {
	ID            string                 `json:"id"`
	ChainID       string                 `json:"chain_id"`
	Status        Status                 `json:"status"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       time.Time              `json:"end_time,omitempty"`
	Progress      float64                `json:"progress"`
	CurrentModel  string                 `json:"current_model,omitempty"`
	TotalTokens   int                    `json:"total_tokens"`
	Error         string                 `json:"error,omitempty"`
	Checkpoints   []string               `json:"checkpoints"`
	ExtraMetadata map[string]interface{} `json:"extra_metadata,omitempty"`
}

// ChunkInfo содержит информацию о сегменте текста
type ChunkInfo struct {
	ID       string `json:"id"`
	Content  string `json:"content"`
	StartPos int    `json:"start_pos"`
	EndPos   int    `json:"end_pos"`
	Order    int    `json:"order"`
}

// RicochetService отвечает за оркестрацию выполнения цепочек моделей
type RicochetService struct {
	apiClient       *api.Client
	keyStore        key.Store
	chainStore      chain.Store
	checkpointStore checkpoint.Store
	runs            map[string]*RunMetadata
	chunker         Chunker
	mutex           sync.RWMutex
}

// Chunker отвечает за разбиение текста на сегменты
type Chunker interface {
	Split(text string, maxChunkSize int) ([]ChunkInfo, error)
	Merge(chunks []ChunkInfo) (string, error)
}

// SimpleChunker реализует простой алгоритм разбиения текста
type SimpleChunker struct{}

// Split разбивает текст на фрагменты
func (c *SimpleChunker) Split(text string, maxChunkSize int) ([]ChunkInfo, error) {
	if maxChunkSize <= 0 {
		return nil, errors.New("максимальный размер сегмента должен быть положительным")
	}

	var chunks []ChunkInfo
	textRunes := []rune(text)
	totalLength := len(textRunes)

	// Если текст меньше максимального размера, возвращаем его как один фрагмент
	if totalLength <= maxChunkSize {
		return []ChunkInfo{
			{
				ID:       uuid.New().String(),
				Content:  text,
				StartPos: 0,
				EndPos:   totalLength,
				Order:    0,
			},
		}, nil
	}

	// Разбиваем текст на фрагменты
	order := 0
	for i := 0; i < totalLength; i += maxChunkSize {
		end := i + maxChunkSize
		if end > totalLength {
			end = totalLength
		}

		chunk := ChunkInfo{
			ID:       uuid.New().String(),
			Content:  string(textRunes[i:end]),
			StartPos: i,
			EndPos:   end,
			Order:    order,
		}
		chunks = append(chunks, chunk)
		order++
	}

	return chunks, nil
}

// Merge объединяет фрагменты в единый текст
func (c *SimpleChunker) Merge(chunks []ChunkInfo) (string, error) {
	if len(chunks) == 0 {
		return "", nil
	}

	// Сортировка фрагментов по их порядку
	orderedChunks := make([]ChunkInfo, len(chunks))
	copy(orderedChunks, chunks)

	// Пузырьковая сортировка по полю Order
	for i := 0; i < len(orderedChunks)-1; i++ {
		for j := 0; j < len(orderedChunks)-i-1; j++ {
			if orderedChunks[j].Order > orderedChunks[j+1].Order {
				orderedChunks[j], orderedChunks[j+1] = orderedChunks[j+1], orderedChunks[j]
			}
		}
	}

	// Объединение фрагментов
	result := ""
	for _, chunk := range orderedChunks {
		result += chunk.Content
	}

	return result, nil
}

// NewRicochetService создает новый экземпляр RicochetService
func NewRicochetService(
	apiClient *api.Client,
	keyStore key.Store,
	chainStore chain.Store,
	checkpointStore checkpoint.Store,
) *RicochetService {
	return &RicochetService{
		apiClient:       apiClient,
		keyStore:        keyStore,
		chainStore:      chainStore,
		checkpointStore: checkpointStore,
		runs:            make(map[string]*RunMetadata),
		chunker:         &SimpleChunker{},
		mutex:           sync.RWMutex{},
	}
}

// RunChain запускает выполнение цепочки моделей
func (s *RicochetService) RunChain(ctx context.Context, chainID string, input string) (string, error) {
	// Получение цепочки
	c, err := s.chainStore.Get(chainID)
	if err != nil {
		return "", fmt.Errorf("ошибка при получении цепочки: %w", err)
	}

	if len(c.Models) == 0 {
		return "", errors.New("цепочка не содержит моделей")
	}

	// Создание метаданных о запуске
	runID := uuid.New().String()
	runMeta := &RunMetadata{
		ID:            runID,
		ChainID:       chainID,
		Status:        StatusPending,
		StartTime:     time.Now(),
		Progress:      0,
		TotalTokens:   0,
		Checkpoints:   []string{},
		ExtraMetadata: map[string]interface{}{},
	}

	// Сохранение метаданных
	s.mutex.Lock()
	s.runs[runID] = runMeta
	s.mutex.Unlock()

	// Сохранение входного чекпоинта
	inputCheckpoint := checkpoint.Checkpoint{
		ID:        uuid.New().String(),
		ChainID:   chainID,
		Type:      checkpoint.CheckpointTypeInput,
		Content:   input,
		CreatedAt: time.Now(),
		MetaData:  make(map[string]interface{}),
	}

	if err := s.checkpointStore.Save(inputCheckpoint); err != nil {
		return "", fmt.Errorf("ошибка при сохранении входного чекпоинта: %w", err)
	}

	runMeta.Checkpoints = append(runMeta.Checkpoints, inputCheckpoint.ID)

	// Обновление статуса
	runMeta.Status = StatusRunning

	// Запуск выполнения цепочки в горутине
	go func() {
		result, err := s.processChain(ctx, c, input, runMeta)

		s.mutex.Lock()
		defer s.mutex.Unlock()

		if err != nil {
			runMeta.Status = StatusFailed
			runMeta.Error = err.Error()
		} else {
			runMeta.Status = StatusCompleted
		}

		runMeta.EndTime = time.Now()
		runMeta.Progress = 100

		// Сохранение финального результата как чекпоинта
		if err == nil {
			finalCheckpoint := checkpoint.Checkpoint{
				ID:        uuid.New().String(),
				ChainID:   chainID,
				Type:      checkpoint.CheckpointTypeComplete,
				Content:   result,
				CreatedAt: time.Now(),
				MetaData:  make(map[string]interface{}),
			}

			if err := s.checkpointStore.Save(finalCheckpoint); err != nil {
				fmt.Printf("Ошибка при сохранении финального чекпоинта: %v\n", err)
			} else {
				runMeta.Checkpoints = append(runMeta.Checkpoints, finalCheckpoint.ID)
			}
		}
	}()

	return runID, nil
}

// processChain обрабатывает цепочку моделей
func (s *RicochetService) processChain(ctx context.Context, c chain.Chain, input string, runMeta *RunMetadata) (string, error) {
	currentText := input
	totalModels := len(c.Models)

	for i, model := range c.Models {
		// Проверка контекста на отмену
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			// Продолжаем выполнение
		}

		s.mutex.Lock()
		runMeta.CurrentModel = string(model.Name)
		runMeta.Progress = float64(i) / float64(totalModels) * 100
		s.mutex.Unlock()

		// Получение ключа API
		apiKey, err := s.getAPIKey(model.Type)
		if err != nil {
			return "", fmt.Errorf("ошибка при получении API-ключа для модели %s: %w", model.Name, err)
		}

		// Создаем чат-сервис и отправляем запрос
		chatService := api.NewChatService(s.apiClient)

		// Определяем провайдера
		provider := getProviderFromModelType(model.Type)
		s.apiClient.SetAPIKey(provider, apiKey)

		// Создаем запрос к модели
		chatRequest := &api.ChatRequest{
			Model: string(model.Name),
			Messages: []api.ChatMessage{
				{
					Role:    "system",
					Content: model.Prompt,
				},
				{
					Role:    "user",
					Content: currentText,
				},
			},
			MaxTokens:   model.MaxTokens,
			Temperature: model.Temperature,
		}

		// Отправляем запрос
		chatResponse, err := chatService.SendMessage(chatRequest)
		if err != nil {
			return "", fmt.Errorf("ошибка при вызове модели %s: %w", model.Name, err)
		}

		// Обрабатываем ответ
		response := chatResponse.Message.Content

		// Обновляем текущий текст
		currentText = response

		// Обновление статистики токенов
		s.mutex.Lock()
		runMeta.TotalTokens += len(currentText) / 4 // Грубая оценка количества токенов
		s.mutex.Unlock()

		// Сохранение чекпоинта
		modelCheckpoint := checkpoint.Checkpoint{
			ID:        uuid.New().String(),
			ChainID:   c.ID,
			ModelID:   model.ID,
			Type:      checkpoint.CheckpointTypeOutput,
			Content:   response,
			CreatedAt: time.Now(),
			MetaData: map[string]interface{}{
				"model_name":  string(model.Name),
				"model_type":  string(model.Type),
				"model_role":  string(model.Role),
				"temperature": model.Temperature,
				"max_tokens":  model.MaxTokens,
			},
		}

		if err := s.checkpointStore.Save(modelCheckpoint); err != nil {
			return "", fmt.Errorf("ошибка при сохранении чекпоинта модели: %w", err)
		}

		s.mutex.Lock()
		runMeta.Checkpoints = append(runMeta.Checkpoints, modelCheckpoint.ID)
		s.mutex.Unlock()
	}

	return currentText, nil
}

// getAPIKey получает подходящий API-ключ для указанного типа модели
func (s *RicochetService) getAPIKey(modelType chain.ModelType) (string, error) {
	// Получение списка ключей
	keys, err := s.keyStore.List()
	if err != nil {
		return "", fmt.Errorf("ошибка при получении списка ключей: %w", err)
	}

	// Фильтрация ключей по типу провайдера
	var providerName string
	switch modelType {
	case chain.ModelTypeOpenAI:
		providerName = "openai"
	case chain.ModelTypeClaude:
		providerName = "claude"
	case chain.ModelTypeDeepSeek:
		providerName = "deepseek"
	case chain.ModelTypeGrok:
		providerName = "grok"
	default:
		return "", fmt.Errorf("неизвестный тип модели: %v", modelType)
	}

	var validKeys []key.Key
	for _, k := range keys {
		if k.Provider == providerName {
			validKeys = append(validKeys, k)
		}
	}

	if len(validKeys) == 0 {
		return "", fmt.Errorf("не найдены ключи для провайдера %s", providerName)
	}

	// Пока просто возвращаем первый подходящий ключ
	return validKeys[0].Value, nil
}

// GetRunStatus возвращает статус выполнения цепочки
func (s *RicochetService) GetRunStatus(runID string) (*RunMetadata, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	runMeta, exists := s.runs[runID]
	if !exists {
		return nil, fmt.Errorf("запуск с ID %s не найден", runID)
	}

	return runMeta, nil
}

// CancelRun отменяет выполнение цепочки
func (s *RicochetService) CancelRun(runID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	runMeta, exists := s.runs[runID]
	if !exists {
		return fmt.Errorf("запуск с ID %s не найден", runID)
	}

	if runMeta.Status == StatusCompleted || runMeta.Status == StatusFailed || runMeta.Status == StatusCancelled {
		return fmt.Errorf("нельзя отменить завершенный запуск")
	}

	runMeta.Status = StatusCancelled
	runMeta.EndTime = time.Now()

	return nil
}

// ListRuns возвращает список всех запусков
func (s *RicochetService) ListRuns() []*RunMetadata {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	runs := make([]*RunMetadata, 0, len(s.runs))
	for _, run := range s.runs {
		runs = append(runs, run)
	}

	return runs
}

// GetRunResults возвращает результаты выполнения цепочки
func (s *RicochetService) GetRunResults(runID string) (string, error) {
	s.mutex.RLock()
	runMeta, exists := s.runs[runID]
	s.mutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("запуск с ID %s не найден", runID)
	}

	if runMeta.Status != StatusCompleted {
		return "", fmt.Errorf("запуск не завершен, текущий статус: %s", runMeta.Status)
	}

	if len(runMeta.Checkpoints) == 0 {
		return "", fmt.Errorf("чекпоинты не найдены для запуска %s", runID)
	}

	// Получаем последний чекпоинт
	lastCheckpointID := runMeta.Checkpoints[len(runMeta.Checkpoints)-1]

	checkpoint, err := s.checkpointStore.Get(lastCheckpointID)
	if err != nil {
		return "", fmt.Errorf("ошибка при получении финального чекпоинта: %w", err)
	}

	return checkpoint.Content, nil
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
