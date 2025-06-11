package mcp

import (
	"encoding/json"
	"fmt"
	"time"
)

// ChainControlParams параметры для управления выполнением цепочки
type ChainControlParams struct {
	ChainID string `json:"chain_id"`
	Reason  string `json:"reason,omitempty"`
}

// ChainControlResponse ответ на запрос управления цепочкой
type ChainControlResponse struct {
	ChainID      string    `json:"chain_id"`
	Action       string    `json:"action"`
	Status       string    `json:"status"`
	Message      string    `json:"message"`
	ModelID      string    `json:"model_id,omitempty"`
	CurrentStep  int       `json:"current_step,omitempty"`
	TotalSteps   int       `json:"total_steps,omitempty"`
	PreviousStep string    `json:"previous_step,omitempty"`
	NextStep     string    `json:"next_step,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// StepControlParams параметры для управления переходами между шагами
type StepControlParams struct {
	ChainID   string `json:"chain_id"`
	Direction string `json:"direction"`            // "next", "previous", "goto"
	StepIndex int    `json:"step_index,omitempty"` // для direction=goto
}

// HandleChainPause обрабатывает запрос на паузу цепочки
func HandleChainPause(params json.RawMessage) (interface{}, error) {
	var controlParams ChainControlParams
	if err := json.Unmarshal(params, &controlParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для паузы цепочки: %v", err)
	}

	if controlParams.ChainID == "" {
		return nil, fmt.Errorf("chain_id является обязательным параметром")
	}

	// Проверяем, выполняется ли цепочка
	isRunning, err := isChainInState(controlParams.ChainID, "running")
	if err != nil {
		return nil, err
	}
	if !isRunning {
		return nil, fmt.Errorf("цепочка не выполняется в данный момент")
	}

	// Приостанавливаем выполнение цепочки
	if err := pauseChainExecution(controlParams.ChainID, controlParams.Reason); err != nil {
		return nil, err
	}

	// Получаем информацию о текущем состоянии
	chainState, err := getChainExecutionState(controlParams.ChainID)
	if err != nil {
		return nil, err
	}

	response := ChainControlResponse{
		ChainID:      controlParams.ChainID,
		Action:       "pause",
		Status:       "paused",
		Message:      fmt.Sprintf("Выполнение цепочки приостановлено. Причина: %s", controlParams.Reason),
		ModelID:      chainState.CurrentModelID,
		CurrentStep:  chainState.CurrentStep,
		TotalSteps:   chainState.TotalSteps,
		PreviousStep: chainState.PreviousModelName,
		NextStep:     chainState.NextModelName,
		Timestamp:    time.Now(),
	}

	return response, nil
}

// HandleChainResume обрабатывает запрос на возобновление цепочки
func HandleChainResume(params json.RawMessage) (interface{}, error) {
	var controlParams ChainControlParams
	if err := json.Unmarshal(params, &controlParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для возобновления цепочки: %v", err)
	}

	if controlParams.ChainID == "" {
		return nil, fmt.Errorf("chain_id является обязательным параметром")
	}

	// Проверяем, приостановлена ли цепочка
	isPaused, err := isChainInState(controlParams.ChainID, "paused")
	if err != nil {
		return nil, err
	}
	if !isPaused {
		return nil, fmt.Errorf("цепочка не находится в состоянии паузы")
	}

	// Возобновляем выполнение цепочки
	if err := resumeChainExecution(controlParams.ChainID); err != nil {
		return nil, err
	}

	// Получаем информацию о текущем состоянии
	chainState, err := getChainExecutionState(controlParams.ChainID)
	if err != nil {
		return nil, err
	}

	response := ChainControlResponse{
		ChainID:      controlParams.ChainID,
		Action:       "resume",
		Status:       "running",
		Message:      "Выполнение цепочки возобновлено",
		ModelID:      chainState.CurrentModelID,
		CurrentStep:  chainState.CurrentStep,
		TotalSteps:   chainState.TotalSteps,
		PreviousStep: chainState.PreviousModelName,
		NextStep:     chainState.NextModelName,
		Timestamp:    time.Now(),
	}

	return response, nil
}

// HandleChainStop обрабатывает запрос на остановку цепочки
func HandleChainStop(params json.RawMessage) (interface{}, error) {
	var controlParams ChainControlParams
	if err := json.Unmarshal(params, &controlParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для остановки цепочки: %v", err)
	}

	if controlParams.ChainID == "" {
		return nil, fmt.Errorf("chain_id является обязательным параметром")
	}

	// Проверяем состояние цепочки
	isActive, err := isChainActive(controlParams.ChainID)
	if err != nil {
		return nil, err
	}
	if !isActive {
		return nil, fmt.Errorf("цепочка не активна или уже завершена")
	}

	// Останавливаем выполнение цепочки
	if err := stopChainExecution(controlParams.ChainID, controlParams.Reason); err != nil {
		return nil, err
	}

	response := ChainControlResponse{
		ChainID:   controlParams.ChainID,
		Action:    "stop",
		Status:    "stopped",
		Message:   fmt.Sprintf("Выполнение цепочки остановлено. Причина: %s", controlParams.Reason),
		Timestamp: time.Now(),
	}

	return response, nil
}

// HandleStepControl обрабатывает запрос на управление переходами между шагами
func HandleStepControl(params json.RawMessage) (interface{}, error) {
	var stepParams StepControlParams
	if err := json.Unmarshal(params, &stepParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для управления шагами: %v", err)
	}

	if stepParams.ChainID == "" {
		return nil, fmt.Errorf("chain_id является обязательным параметром")
	}

	if stepParams.Direction == "" {
		return nil, fmt.Errorf("direction является обязательным параметром")
	}

	// Проверяем, приостановлена ли цепочка
	isPaused, err := isChainInState(stepParams.ChainID, "paused")
	if err != nil {
		return nil, err
	}
	if !isPaused {
		return nil, fmt.Errorf("управление шагами доступно только для приостановленных цепочек")
	}

	// Получаем информацию о текущем состоянии
	chainState, err := getChainExecutionState(stepParams.ChainID)
	if err != nil {
		return nil, err
	}

	// Обрабатываем команду направления
	var targetStep int
	var actionMessage string

	switch stepParams.Direction {
	case "next":
		if chainState.CurrentStep >= chainState.TotalSteps-1 {
			return nil, fmt.Errorf("достигнут последний шаг цепочки")
		}
		targetStep = chainState.CurrentStep + 1
		actionMessage = "Переход к следующему шагу"
	case "previous":
		if chainState.CurrentStep <= 0 {
			return nil, fmt.Errorf("достигнут первый шаг цепочки")
		}
		targetStep = chainState.CurrentStep - 1
		actionMessage = "Переход к предыдущему шагу"
	case "goto":
		if stepParams.StepIndex < 0 || stepParams.StepIndex >= chainState.TotalSteps {
			return nil, fmt.Errorf("указанный индекс шага вне допустимого диапазона")
		}
		targetStep = stepParams.StepIndex
		actionMessage = fmt.Sprintf("Переход к шагу %d", targetStep)
	default:
		return nil, fmt.Errorf("неподдерживаемое направление: %s", stepParams.Direction)
	}

	// Выполняем переход к указанному шагу
	if err := moveToChainStep(stepParams.ChainID, targetStep); err != nil {
		return nil, err
	}

	// Получаем обновленное состояние
	updatedState, err := getChainExecutionState(stepParams.ChainID)
	if err != nil {
		return nil, err
	}

	response := ChainControlResponse{
		ChainID:      stepParams.ChainID,
		Action:       "step_control",
		Status:       "paused",
		Message:      actionMessage,
		ModelID:      updatedState.CurrentModelID,
		CurrentStep:  updatedState.CurrentStep,
		TotalSteps:   updatedState.TotalSteps,
		PreviousStep: updatedState.PreviousModelName,
		NextStep:     updatedState.NextModelName,
		Timestamp:    time.Now(),
	}

	return response, nil
}

// RegisterChainControlCommands регистрирует команды управления цепочками
func RegisterChainControlCommands(server *MCPServer) {
	server.RegisterCommand("chain_pause", HandleChainPause)
	server.RegisterCommand("chain_resume", HandleChainResume)
	server.RegisterCommand("chain_stop", HandleChainStop)
	server.RegisterCommand("chain_step_control", HandleStepControl)
}

// Вспомогательные функции

// ChainExecutionState представляет состояние выполнения цепочки
type ChainExecutionState struct {
	ChainID           string
	Status            string
	CurrentStep       int
	TotalSteps        int
	CurrentModelID    string
	CurrentModelName  string
	PreviousModelName string
	NextModelName     string
}

// isChainInState проверяет, находится ли цепочка в указанном состоянии
func isChainInState(chainID, state string) (bool, error) {
	// TODO: Реализовать проверку состояния цепочки через оркестратор
	// Временная реализация
	return true, nil
}

// isChainActive проверяет, активна ли цепочка
func isChainActive(chainID string) (bool, error) {
	// TODO: Реализовать проверку активности цепочки через оркестратор
	// Временная реализация
	return true, nil
}

// pauseChainExecution приостанавливает выполнение цепочки
func pauseChainExecution(chainID, reason string) error {
	// TODO: Реализовать паузу выполнения цепочки через оркестратор
	// Временная реализация
	return nil
}

// resumeChainExecution возобновляет выполнение цепочки
func resumeChainExecution(chainID string) error {
	// TODO: Реализовать возобновление выполнения цепочки через оркестратор
	// Временная реализация
	return nil
}

// stopChainExecution останавливает выполнение цепочки
func stopChainExecution(chainID, reason string) error {
	// TODO: Реализовать остановку выполнения цепочки через оркестратор
	// Временная реализация
	return nil
}

// moveToChainStep перемещает выполнение цепочки к указанному шагу
func moveToChainStep(chainID string, stepIndex int) error {
	// TODO: Реализовать перемещение к шагу через оркестратор
	// Временная реализация
	return nil
}

// getChainExecutionState получает текущее состояние выполнения цепочки
func getChainExecutionState(chainID string) (ChainExecutionState, error) {
	// TODO: Получить состояние выполнения цепочки из оркестратора
	// Временная реализация
	state := ChainExecutionState{
		ChainID:           chainID,
		Status:            "paused",
		CurrentStep:       1,
		TotalSteps:        3,
		CurrentModelID:    "model-2",
		CurrentModelName:  "Claude-3-Opus",
		PreviousModelName: "GPT-4",
		NextModelName:     "DeepSeek-Coder",
	}
	return state, nil
}
