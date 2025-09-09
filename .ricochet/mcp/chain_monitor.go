package mcp

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ChainMonitorParams параметры для мониторинга цепочки
type ChainMonitorParams struct {
	ChainID        string `json:"chain_id"`
	IncludeHistory bool   `json:"include_history,omitempty"`
	RefreshRate    int    `json:"refresh_rate,omitempty"` // в миллисекундах
}

// ChainEvent представляет событие в процессе выполнения цепочки
type ChainEvent struct {
	ID        string    `json:"id"`
	ChainID   string    `json:"chain_id"`
	Type      string    `json:"type"` // start, step, complete, error
	Timestamp time.Time `json:"timestamp"`
	ModelID   string    `json:"model_id,omitempty"`
	Message   string    `json:"message,omitempty"`
	Progress  float64   `json:"progress,omitempty"`
	TaskID    string    `json:"task_id,omitempty"`
}

// ChainMonitorResponse ответ на запрос мониторинга
type ChainMonitorResponse struct {
	ChainID    string       `json:"chain_id"`
	ChainName  string       `json:"chain_name"`
	Status     string       `json:"status"`
	LiveView   string       `json:"live_view"` // ASCII визуализация цепочки
	Events     []ChainEvent `json:"events,omitempty"`
	UpdateTime time.Time    `json:"update_time"`
}

// ChainMonitorStreamResponse потоковый ответ на мониторинг
type ChainMonitorStreamResponse struct {
	ChainID       string     `json:"chain_id"`
	Status        string     `json:"status"`
	Event         ChainEvent `json:"event"`
	Visualization string     `json:"visualization,omitempty"`
	UpdateTime    time.Time  `json:"update_time"`
}

// chainMonitorSessions хранит активные сессии мониторинга
var chainMonitorSessions = struct {
	sessions map[string]chan ChainMonitorStreamResponse
	mutex    sync.RWMutex
}{
	sessions: make(map[string]chan ChainMonitorStreamResponse),
}

// HandleChainMonitor обрабатывает запрос на мониторинг цепочки
func HandleChainMonitor(params json.RawMessage) (interface{}, error) {
	var monitorParams ChainMonitorParams
	if err := json.Unmarshal(params, &monitorParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для мониторинга: %v", err)
	}

	if monitorParams.ChainID == "" {
		return nil, fmt.Errorf("chain_id является обязательным параметром")
	}

	// Установить значения по умолчанию
	if monitorParams.RefreshRate <= 0 {
		monitorParams.RefreshRate = 1000 // 1 секунда по умолчанию
	}

	// Получить информацию о цепочке и текущий статус выполнения
	chainInfo, err := getChainBasicInfo(monitorParams.ChainID)
	if err != nil {
		return nil, err
	}

	// Генерация ASCII-визуализации для цепочки
	liveView := generateChainVisualization(monitorParams.ChainID)

	// Получение истории событий, если требуется
	var events []ChainEvent
	if monitorParams.IncludeHistory {
		events = getChainEventHistory(monitorParams.ChainID)
	}

	// Создать ответ
	response := ChainMonitorResponse{
		ChainID:    monitorParams.ChainID,
		ChainName:  chainInfo.Name,
		Status:     getChainStatus(monitorParams.ChainID),
		LiveView:   liveView,
		Events:     events,
		UpdateTime: time.Now(),
	}

	// Если цепочка в процессе выполнения, создаём и регистрируем потоковую сессию
	if isChainRunning(monitorParams.ChainID) {
		go monitorChainExecution(monitorParams)
	}

	return response, nil
}

// HandleChainMonitorStop останавливает мониторинг цепочки
func HandleChainMonitorStop(params json.RawMessage) (interface{}, error) {
	var stopParams struct {
		ChainID string `json:"chain_id"`
	}

	if err := json.Unmarshal(params, &stopParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для остановки мониторинга: %v", err)
	}

	if stopParams.ChainID == "" {
		return nil, fmt.Errorf("chain_id является обязательным параметром")
	}

	// Остановить сессию мониторинга
	stopChainMonitoring(stopParams.ChainID)

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Мониторинг цепочки %s остановлен", stopParams.ChainID),
	}, nil
}

// RegisterChainMonitorCommands регистрирует команды мониторинга цепочек
func RegisterChainMonitorCommands(server *MCPServer) {
	server.RegisterCommand("chain_monitor", HandleChainMonitor)
	server.RegisterCommand("chain_monitor_stop", HandleChainMonitorStop)
}

// Вспомогательные функции

// getChainBasicInfo получает базовую информацию о цепочке
func getChainBasicInfo(chainID string) (struct{ Name string }, error) {
	// Используем функцию из mcp_utils.go
	info, err := getChainInfo(chainID)
	if err != nil {
		// Fallback
		return struct{ Name string }{Name: "Chain " + chainID}, nil
	}
	return struct{ Name string }{Name: info.Name}, nil
}

// getChainStatus получает текущий статус цепочки
func getChainStatus(chainID string) string {
	// TODO: Получить статус цепочки из оркестратора
	// Временная реализация
	return "running"
}

// isChainRunning проверяет, выполняется ли сейчас цепочка
func isChainRunning(chainID string) bool {
	// TODO: Проверить, выполняется ли цепочка
	// Временная реализация
	return true
}

// getChainEventHistory получает историю событий для цепочки
func getChainEventHistory(chainID string) []ChainEvent {
	// TODO: Получить историю событий из хранилища или лога
	// Временная реализация
	return []ChainEvent{
		{
			ID:        "evt-1",
			ChainID:   chainID,
			Type:      "start",
			Timestamp: time.Now().Add(-5 * time.Minute),
			Message:   "Запуск цепочки",
			Progress:  0.0,
		},
		{
			ID:        "evt-2",
			ChainID:   chainID,
			Type:      "step",
			Timestamp: time.Now().Add(-3 * time.Minute),
			ModelID:   "model-1",
			Message:   "Выполнение модели анализа",
			Progress:  0.35,
			TaskID:    "task-1",
		},
	}
}

// generateChainVisualization генерирует ASCII-визуализацию цепочки
func generateChainVisualization(chainID string) string {
	// TODO: Генерация ASCII-визуализации цепочки и ее текущего состояния
	// Временная реализация
	visualization := `
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Анализатор │───>│ Суммаризатор│───>│  Интегратор │
│   (GPT-4)   │    │  (Claude-3) │    │ (DeepSeek)  │
│  [██████--] │    │  [----]     │    │  [----]     │
└─────────────┘    └─────────────┘    └─────────────┘
      65%                0%                 0%      
`
	return visualization
}

// monitorChainExecution запускает мониторинг выполнения цепочки
func monitorChainExecution(params ChainMonitorParams) {
	chainID := params.ChainID
	refresh := time.Duration(params.RefreshRate) * time.Millisecond

	// Создаём канал для стрима и регистрируем его
	updateCh := make(chan ChainMonitorStreamResponse, 16)
	chainMonitorSessions.mutex.Lock()
	chainMonitorSessions.sessions[chainID] = updateCh
	chainMonitorSessions.mutex.Unlock()

	defer func() {
		// По завершении удаляем сессию
		chainMonitorSessions.mutex.Lock()
		delete(chainMonitorSessions.sessions, chainID)
		chainMonitorSessions.mutex.Unlock()
		close(updateCh)
	}()

	ticker := time.NewTicker(refresh)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status := getChainStatus(chainID)
			viz := generateChainVisualization(chainID)
			event := ChainEvent{
				ID:        fmt.Sprintf("evt-%d", time.Now().UnixNano()),
				ChainID:   chainID,
				Type:      "heartbeat",
				Timestamp: time.Now(),
				Message:   fmt.Sprintf("chain status: %s", status),
				Progress:  0, // TODO: заменить на реальный прогресс
			}
			resp := ChainMonitorStreamResponse{
				ChainID:       chainID,
				Status:        status,
				Event:         event,
				Visualization: viz,
				UpdateTime:    time.Now(),
			}
			// Не блокируем, если клиенты не читают
			select {
			case updateCh <- resp:
			default:
			}

			// Если цепочка завершилась – выходим
			if status == "completed" || status == "stopped" || status == "error" {
				return
			}
		}
	}
}

// stopChainMonitoring останавливает мониторинг цепочки
func stopChainMonitoring(chainID string) {
	chainMonitorSessions.mutex.Lock()
	defer chainMonitorSessions.mutex.Unlock()

	if ch, exists := chainMonitorSessions.sessions[chainID]; exists {
		close(ch)
		delete(chainMonitorSessions.sessions, chainID)
	}
}
