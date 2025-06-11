package mcp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/checkpoint"
	"github.com/grik-ai/ricochet-task/pkg/service"
)

// CheckpointListParams параметры для получения списка чекпоинтов
type CheckpointListParams struct {
	ChainID string `json:"chain_id"` // ID цепочки
	RunID   string `json:"run_id"`   // ID выполнения (опционально)
}

// CheckpointGetParams параметры для получения конкретного чекпоинта
type CheckpointGetParams struct {
	CheckpointID string `json:"checkpoint_id"` // ID чекпоинта
}

// CheckpointDeleteParams параметры для удаления чекпоинта
type CheckpointDeleteParams struct {
	CheckpointID string `json:"checkpoint_id"` // ID чекпоинта
}

// CheckpointListResponse ответ на запрос списка чекпоинтов
type CheckpointListResponse struct {
	ChainID     string                    `json:"chain_id"`
	RunID       string                    `json:"run_id,omitempty"`
	Checkpoints []CheckpointSummaryInfo   `json:"checkpoints"`
	Timeline    []CheckpointTimelineEvent `json:"timeline"`
}

// CheckpointSummaryInfo краткая информация о чекпоинте
type CheckpointSummaryInfo struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	ModelID     string                 `json:"model_id,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	ContentSize int                    `json:"content_size"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CheckpointTimelineEvent событие на временной шкале чекпоинтов
type CheckpointTimelineEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	ModelName string    `json:"model_name,omitempty"`
	ModelRole string    `json:"model_role,omitempty"`
}

// CheckpointDetailsResponse ответ на запрос детальной информации о чекпоинте
type CheckpointDetailsResponse struct {
	ID          string                 `json:"id"`
	ChainID     string                 `json:"chain_id"`
	ModelID     string                 `json:"model_id,omitempty"`
	Type        string                 `json:"type"`
	Content     string                 `json:"content"`
	CreatedAt   time.Time              `json:"created_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ContentSize int                    `json:"content_size"`
}

// HandleCheckpointList обрабатывает запрос на получение списка чекпоинтов
func HandleCheckpointList(params json.RawMessage) (interface{}, error) {
	var listParams CheckpointListParams
	if err := json.Unmarshal(params, &listParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для получения списка чекпоинтов: %v", err)
	}

	if listParams.ChainID == "" {
		return nil, fmt.Errorf("chain_id является обязательным параметром")
	}

	// Получаем сервис оркестратора
	orchestrator, err := GetOrchestratorService()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить сервис оркестратора: %v", err)
	}

	var checkpoints []checkpoint.Checkpoint
	if listParams.RunID != "" {
		// Получаем чекпоинты для конкретного выполнения
		checkpoints, err = orchestrator.ListCheckpoints(listParams.RunID)
	} else {
		// Получаем все чекпоинты для цепочки
		store, err := getCheckpointStore()
		if err != nil {
			return nil, err
		}
		checkpoints, err = store.List(listParams.ChainID)
	}

	if err != nil {
		return nil, fmt.Errorf("ошибка при получении списка чекпоинтов: %v", err)
	}

	// Формируем ответ
	response := CheckpointListResponse{
		ChainID:     listParams.ChainID,
		RunID:       listParams.RunID,
		Checkpoints: make([]CheckpointSummaryInfo, 0, len(checkpoints)),
		Timeline:    make([]CheckpointTimelineEvent, 0, len(checkpoints)),
	}

	// Заполняем список чекпоинтов и временную шкалу
	for _, cp := range checkpoints {
		// Краткая информация о чекпоинте
		summary := CheckpointSummaryInfo{
			ID:          cp.ID,
			Type:        string(cp.Type),
			ModelID:     cp.ModelID,
			CreatedAt:   cp.CreatedAt,
			ContentSize: len(cp.Content),
			Metadata:    cp.MetaData,
		}
		response.Checkpoints = append(response.Checkpoints, summary)

		// Событие на временной шкале
		event := CheckpointTimelineEvent{
			ID:        cp.ID,
			Type:      string(cp.Type),
			Timestamp: cp.CreatedAt,
		}

		// Дополнительная информация о модели, если есть
		if cp.MetaData != nil {
			if modelName, ok := cp.MetaData["model_name"].(string); ok {
				event.ModelName = modelName
			}
			if modelRole, ok := cp.MetaData["model_role"].(string); ok {
				event.ModelRole = modelRole
			}
		}

		response.Timeline = append(response.Timeline, event)
	}

	return response, nil
}

// HandleCheckpointGet обрабатывает запрос на получение конкретного чекпоинта
func HandleCheckpointGet(params json.RawMessage) (interface{}, error) {
	var getParams CheckpointGetParams
	if err := json.Unmarshal(params, &getParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для получения чекпоинта: %v", err)
	}

	if getParams.CheckpointID == "" {
		return nil, fmt.Errorf("checkpoint_id является обязательным параметром")
	}

	// Получаем сервис оркестратора
	orchestrator, err := GetOrchestratorService()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить сервис оркестратора: %v", err)
	}

	// Получаем чекпоинт
	cp, err := orchestrator.GetCheckpoint(getParams.CheckpointID)
	if err != nil {
		// Пробуем получить из хранилища чекпоинтов напрямую
		store, err := getCheckpointStore()
		if err != nil {
			return nil, err
		}
		cp, err = store.Get(getParams.CheckpointID)
		if err != nil {
			return nil, fmt.Errorf("чекпоинт с ID '%s' не найден: %v", getParams.CheckpointID, err)
		}
	}

	// Формируем ответ
	response := CheckpointDetailsResponse{
		ID:          cp.ID,
		ChainID:     cp.ChainID,
		ModelID:     cp.ModelID,
		Type:        string(cp.Type),
		Content:     cp.Content,
		CreatedAt:   cp.CreatedAt,
		Metadata:    cp.MetaData,
		ContentSize: len(cp.Content),
	}

	return response, nil
}

// HandleCheckpointDelete обрабатывает запрос на удаление чекпоинта
func HandleCheckpointDelete(params json.RawMessage) (interface{}, error) {
	var deleteParams CheckpointDeleteParams
	if err := json.Unmarshal(params, &deleteParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для удаления чекпоинта: %v", err)
	}

	if deleteParams.CheckpointID == "" {
		return nil, fmt.Errorf("checkpoint_id является обязательным параметром")
	}

	// Получаем хранилище чекпоинтов
	store, err := getCheckpointStore()
	if err != nil {
		return nil, err
	}

	// Удаляем чекпоинт
	if err := store.Delete(deleteParams.CheckpointID); err != nil {
		return nil, fmt.Errorf("ошибка при удалении чекпоинта: %v", err)
	}

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Чекпоинт с ID '%s' успешно удален", deleteParams.CheckpointID),
	}, nil
}

// getCheckpointStore возвращает хранилище чекпоинтов
func getCheckpointStore() (checkpoint.Store, error) {
	// Получаем сервис Ricochet
	ricochetService, err := GetRicochetService()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить сервис Ricochet: %v", err)
	}

	// Получаем хранилище чекпоинтов
	store := ricochetService.GetCheckpointStore()
	if store == nil {
		return nil, fmt.Errorf("хранилище чекпоинтов не инициализировано")
	}

	return store, nil
}

// GetOrchestratorService возвращает сервис оркестратора
func GetOrchestratorService() (service.RicochetService, error) {
	// TODO: Реализовать получение сервиса оркестратора
	// Временная реализация
	return nil, fmt.Errorf("сервис оркестратора не реализован")
}

// GetRicochetService возвращает сервис Ricochet
func GetRicochetService() (service.RicochetService, error) {
	// TODO: Реализовать получение сервиса Ricochet
	// Временная реализация
	return nil, fmt.Errorf("сервис Ricochet не реализован")
}

// RegisterCheckpointCommands регистрирует команды для работы с чекпоинтами
func RegisterCheckpointCommands(server *MCPServer) {
	server.RegisterCommand("checkpoint_list", HandleCheckpointList)
	server.RegisterCommand("checkpoint_get", HandleCheckpointGet)
	server.RegisterCommand("checkpoint_delete", HandleCheckpointDelete)
}
