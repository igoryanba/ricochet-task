package mcp

import (
	"encoding/json"
	"fmt"
	"time"
)

// ChainProgressParams представляет параметры для команды отображения прогресса цепочки
type ChainProgressParams struct {
	ChainID string `json:"chain_id"`
}

// ChainProgressResponse представляет ответ команды отображения прогресса цепочки
type ChainProgressResponse struct {
	ChainID           string          `json:"chain_id"`
	ChainName         string          `json:"chain_name"`
	Status            string          `json:"status"`
	Progress          float64         `json:"progress"`
	StartedAt         time.Time       `json:"started_at"`
	EstimatedEndTime  time.Time       `json:"estimated_end_time,omitempty"`
	ElapsedTime       string          `json:"elapsed_time"`
	RemainingTime     string          `json:"remaining_time,omitempty"`
	ModelProgresses   []ModelProgress `json:"model_progresses"`
	Metrics           ChainMetrics    `json:"metrics"`
	CurrentTaskID     string          `json:"current_task_id,omitempty"`
	CompletedTasksIDs []string        `json:"completed_tasks_ids"`
	ErrorMessage      string          `json:"error_message,omitempty"`
	ProgressChart     string          `json:"progress_chart"`
}

// ModelProgress представляет прогресс выполнения для отдельной модели в цепочке
type ModelProgress struct {
	ModelID      string  `json:"model_id"`
	ModelName    string  `json:"model_name"`
	Provider     string  `json:"provider"`
	Role         string  `json:"role"`
	Progress     float64 `json:"progress"`
	Status       string  `json:"status"`
	TasksTotal   int     `json:"tasks_total"`
	TasksDone    int     `json:"tasks_done"`
	ErrorMessage string  `json:"error_message,omitempty"`
}

// ChainMetrics представляет метрики выполнения цепочки
type ChainMetrics struct {
	TokensInput   int     `json:"tokens_input"`
	TokensOutput  int     `json:"tokens_output"`
	TotalCost     float64 `json:"total_cost"`
	RequestsCount int     `json:"requests_count"`
	ErrorsCount   int     `json:"errors_count"`
}

// HandleChainProgress обрабатывает MCP-команду для отображения прогресса цепочки
func HandleChainProgress(params json.RawMessage) (interface{}, error) {
	var p ChainProgressParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("failed to parse chain_progress params: %v", err)
	}

	// В реальной реализации здесь будет получение данных о прогрессе цепочки
	// из хранилища или от сервиса оркестрации

	// Это демонстрационный пример, возвращающий фиктивные данные
	response := ChainProgressResponse{
		ChainID:       p.ChainID,
		ChainName:     "Анализ документа",
		Status:        "running",
		Progress:      0.65,
		StartedAt:     time.Now().Add(-time.Minute * 5),
		ElapsedTime:   "5m 0s",
		RemainingTime: "2m 30s",
		ModelProgresses: []ModelProgress{
			{
				ModelID:    "model-1",
				ModelName:  "GPT-4",
				Provider:   "openai",
				Role:       "analyzer",
				Progress:   1.0,
				Status:     "completed",
				TasksTotal: 3,
				TasksDone:  3,
			},
			{
				ModelID:    "model-2",
				ModelName:  "Claude-3",
				Provider:   "anthropic",
				Role:       "summarizer",
				Progress:   0.66,
				Status:     "running",
				TasksTotal: 3,
				TasksDone:  2,
			},
			{
				ModelID:    "model-3",
				ModelName:  "DeepSeek",
				Provider:   "deepseek",
				Role:       "integrator",
				Progress:   0.0,
				Status:     "pending",
				TasksTotal: 2,
				TasksDone:  0,
			},
		},
		Metrics: ChainMetrics{
			TokensInput:   4500,
			TokensOutput:  2300,
			TotalCost:     0.047,
			RequestsCount: 5,
			ErrorsCount:   0,
		},
		CurrentTaskID:     "task-125",
		CompletedTasksIDs: []string{"task-123", "task-124"},
		ProgressChart: `
[▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓----] 65% | Цепочка: Анализ документа

├── [▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓] 100% | Модель: OpenAI GPT-4 (Анализатор)
│   └── Задача #123: Анализ структуры документа ✅ (2.3с)
│   └── Задача #124: Выделение ключевых тем ✅ (3.5с)
│   └── Задача #125: Анализ связей между темами ✅ (2.8с)
│
├── [▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓----------------] 66% | Модель: Claude-3 (Суммаризатор)
│   └── Задача #126: Создание резюме по теме A ✅ (1.7с)
│   └── Задача #127: Создание резюме по теме B ✅ (1.9с)
│   └── Задача #128: Создание резюме по теме C ⏳ (в процессе)
│
└── [-------------------------------] 0% | Модель: DeepSeek (Интегратор)
    └── Задача #129: Объединение резюме 🔜 (ожидание)
    └── Задача #130: Формирование выводов 🔜 (ожидание)
`,
	}

	return response, nil
}

// RegisterChainProgressCommand регистрирует команду chain_progress в MCP-сервере
func RegisterChainProgressCommand(server *MCPServer) {
	server.RegisterCommand("chain_progress", HandleChainProgress)
}

// Пример использования в MCP-сервере:
/*
func InitMCPServer() *MCPServer {
	server := NewMCPServer()

	// Регистрация команд
	RegisterChainProgressCommand(server)

	return server
}
*/
