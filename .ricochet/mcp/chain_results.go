package mcp

import (
	"encoding/json"
	"fmt"
	"time"
)

// ChainResultsParams параметры для получения результатов цепочки
type ChainResultsParams struct {
	ChainID      string `json:"chain_id"`                // ID цепочки
	RunID        string `json:"run_id,omitempty"`        // ID конкретного выполнения (опционально)
	IncludeStats bool   `json:"include_stats,omitempty"` // Включать статистику выполнения
	Limit        int    `json:"limit,omitempty"`         // Ограничение количества результатов
}

// ChainResultsResponse ответ на запрос результатов цепочки
type ChainResultsResponse struct {
	ChainID      string                  `json:"chain_id"`
	ChainName    string                  `json:"chain_name"`
	RunsCount    int                     `json:"runs_count"`
	LastRunID    string                  `json:"last_run_id,omitempty"`
	LastRunDate  time.Time               `json:"last_run_date,omitempty"`
	Results      []ChainRunResult        `json:"results"`
	Stats        *ChainRunStats          `json:"stats,omitempty"`
	Checkpoints  []CheckpointSummaryInfo `json:"checkpoints,omitempty"`
	LastRunError string                  `json:"last_run_error,omitempty"`
}

// ChainRunResult информация о результате выполнения цепочки
type ChainRunResult struct {
	RunID         string                 `json:"run_id"`
	Status        string                 `json:"status"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   time.Time              `json:"completed_at,omitempty"`
	Duration      int64                  `json:"duration_ms"` // Длительность в миллисекундах
	ResultSummary string                 `json:"result_summary"`
	HasError      bool                   `json:"has_error"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ChainRunStats статистика выполнения цепочки
type ChainRunStats struct {
	TotalRuns          int     `json:"total_runs"`
	SuccessfulRuns     int     `json:"successful_runs"`
	FailedRuns         int     `json:"failed_runs"`
	AverageDuration    int64   `json:"average_duration_ms"` // Средняя длительность в миллисекундах
	MinDuration        int64   `json:"min_duration_ms"`     // Минимальная длительность в миллисекундах
	MaxDuration        int64   `json:"max_duration_ms"`     // Максимальная длительность в миллисекундах
	AverageTokensUsed  int     `json:"average_tokens_used"`
	TotalTokensUsed    int     `json:"total_tokens_used"`
	EstimatedCost      float64 `json:"estimated_cost"` // Оценочная стоимость выполнения
	SuccessRate        float64 `json:"success_rate"`   // Процент успешных выполнений
	LastRunDate        string  `json:"last_run_date"`
	LastSuccessfulDate string  `json:"last_successful_date"`
}

// HandleChainResults обрабатывает запрос на получение результатов цепочки
func HandleChainResults(params json.RawMessage) (interface{}, error) {
	var resultsParams ChainResultsParams
	if err := json.Unmarshal(params, &resultsParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для получения результатов: %v", err)
	}

	if resultsParams.ChainID == "" {
		return nil, fmt.Errorf("chain_id является обязательным параметром")
	}

	// Установка значений по умолчанию
	if resultsParams.Limit <= 0 {
		resultsParams.Limit = 10
	}

	// Получаем информацию о цепочке
	chainInfo, err := getChainInfo(resultsParams.ChainID)
	if err != nil {
		return nil, err
	}

	// Получаем результаты выполнения цепочки
	results, err := getChainResults(resultsParams.ChainID, resultsParams.RunID, resultsParams.Limit)
	if err != nil {
		return nil, err
	}

	// Формируем ответ
	response := ChainResultsResponse{
		ChainID:   resultsParams.ChainID,
		ChainName: chainInfo.Name,
		RunsCount: len(results),
		Results:   results,
	}

	// Добавляем информацию о последнем запуске
	if len(results) > 0 {
		response.LastRunID = results[0].RunID
		response.LastRunDate = results[0].StartedAt

		if results[0].HasError {
			response.LastRunError = results[0].ErrorMessage
		}
	}

	// Добавляем статистику, если требуется
	if resultsParams.IncludeStats {
		stats, err := getChainRunStats(resultsParams.ChainID)
		if err != nil {
			// Логируем ошибку, но продолжаем выполнение
			fmt.Printf("Ошибка при получении статистики: %v\n", err)
		} else {
			response.Stats = stats
		}
	}

	// Если указан конкретный runID, получаем чекпоинты для него
	if resultsParams.RunID != "" {
		checkpoints, err := getCheckpointsForRun(resultsParams.RunID)
		if err != nil {
			// Логируем ошибку, но продолжаем выполнение
			fmt.Printf("Ошибка при получении чекпоинтов: %v\n", err)
		} else {
			response.Checkpoints = checkpoints
		}
	}

	return response, nil
}

// HandleChainRunResult обрабатывает запрос на получение результата конкретного запуска
func HandleChainRunResult(params json.RawMessage) (interface{}, error) {
	var runParams struct {
		RunID string `json:"run_id"`
	}

	if err := json.Unmarshal(params, &runParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для получения результата: %v", err)
	}

	if runParams.RunID == "" {
		return nil, fmt.Errorf("run_id является обязательным параметром")
	}

	// Получаем сервис оркестратора
	orchestrator, err := GetOrchestratorService()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить сервис оркестратора: %v", err)
	}

	// Получаем результат выполнения
	result, err := orchestrator.GetRunResults(runParams.RunID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении результата выполнения: %v", err)
	}

	// Получаем чекпоинты для этого выполнения
	checkpoints, err := orchestrator.ListCheckpoints(runParams.RunID)
	if err != nil {
		// Логируем ошибку, но продолжаем выполнение
		fmt.Printf("Ошибка при получении чекпоинтов: %v\n", err)
	}

	// Формируем ответ
	response := struct {
		RunID       string                  `json:"run_id"`
		Result      interface{}             `json:"result"`
		Checkpoints []CheckpointSummaryInfo `json:"checkpoints,omitempty"`
	}{
		RunID:  runParams.RunID,
		Result: result,
	}

	// Добавляем информацию о чекпоинтах
	if checkpoints != nil {
		response.Checkpoints = make([]CheckpointSummaryInfo, 0, len(checkpoints))
		for _, cp := range checkpoints {
			summary := CheckpointSummaryInfo{
				ID:          cp.ID,
				Type:        string(cp.Type),
				ModelID:     cp.ModelID,
				CreatedAt:   cp.CreatedAt,
				ContentSize: len(cp.Content),
				Metadata:    cp.MetaData,
			}
			response.Checkpoints = append(response.Checkpoints, summary)
		}
	}

	return response, nil
}

// getChainResults возвращает результаты выполнения цепочки
func getChainResults(chainID string, runID string, limit int) ([]ChainRunResult, error) {
	// TODO: Реализовать получение результатов из оркестратора
	// Временная реализация
	now := time.Now()
	results := []ChainRunResult{
		{
			RunID:         "run-1",
			Status:        "completed",
			StartedAt:     now.Add(-30 * time.Minute),
			CompletedAt:   now.Add(-29 * time.Minute),
			Duration:      60000, // 1 минута
			ResultSummary: "Успешно обработано 3 модели",
			HasError:      false,
			Metadata: map[string]interface{}{
				"tokens_used": 1500,
				"models_used": []string{"gpt-4", "claude-3", "deepseek-coder"},
			},
		},
		{
			RunID:         "run-2",
			Status:        "failed",
			StartedAt:     now.Add(-2 * time.Hour),
			CompletedAt:   now.Add(-1*time.Hour - 55*time.Minute),
			Duration:      300000, // 5 минут
			ResultSummary: "Ошибка при выполнении",
			HasError:      true,
			ErrorMessage:  "Превышен лимит токенов для модели gpt-4",
			Metadata: map[string]interface{}{
				"tokens_used": 8000,
				"models_used": []string{"gpt-4"},
			},
		},
	}

	return results, nil
}

// getChainRunStats возвращает статистику выполнения цепочки
func getChainRunStats(chainID string) (*ChainRunStats, error) {
	// TODO: Реализовать получение статистики из оркестратора
	// Временная реализация
	stats := &ChainRunStats{
		TotalRuns:          10,
		SuccessfulRuns:     8,
		FailedRuns:         2,
		AverageDuration:    120000, // 2 минуты
		MinDuration:        60000,  // 1 минута
		MaxDuration:        300000, // 5 минут
		AverageTokensUsed:  2000,
		TotalTokensUsed:    20000,
		EstimatedCost:      1.25,
		SuccessRate:        80.0,
		LastRunDate:        time.Now().Format(time.RFC3339),
		LastSuccessfulDate: time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
	}

	return stats, nil
}

// getCheckpointsForRun возвращает чекпоинты для указанного выполнения
func getCheckpointsForRun(runID string) ([]CheckpointSummaryInfo, error) {
	// Получаем сервис оркестратора
	orchestrator, err := GetOrchestratorService()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить сервис оркестратора: %v", err)
	}

	// Получаем чекпоинты
	checkpoints, err := orchestrator.ListCheckpoints(runID)
	if err != nil {
		return nil, err
	}

	// Преобразуем в формат ответа
	result := make([]CheckpointSummaryInfo, 0, len(checkpoints))
	for _, cp := range checkpoints {
		summary := CheckpointSummaryInfo{
			ID:          cp.ID,
			Type:        string(cp.Type),
			ModelID:     cp.ModelID,
			CreatedAt:   cp.CreatedAt,
			ContentSize: len(cp.Content),
			Metadata:    cp.MetaData,
		}
		result = append(result, summary)
	}

	return result, nil
}

// RegisterChainResultsCommands регистрирует команды для работы с результатами цепочек
func RegisterChainResultsCommands(server *MCPServer) {
	server.RegisterCommand("chain_results", HandleChainResults)
	server.RegisterCommand("chain_run_result", HandleChainRunResult)
}
