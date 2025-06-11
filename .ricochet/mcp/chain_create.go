package mcp

import (
	"encoding/json"
	"fmt"
	"time"
)

// ChainCreateParams представляет параметры для команды создания цепочки моделей
type ChainCreateParams struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Steps       []ChainStep     `json:"steps"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	Interactive bool            `json:"interactive,omitempty"`
}

// ChainStep представляет шаг в цепочке моделей
type ChainStep struct {
	RoleID      string          `json:"role_id"`
	ModelID     string          `json:"model_id,omitempty"`
	Provider    string          `json:"provider,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Prompt      string          `json:"prompt,omitempty"`
	InputFrom   []string        `json:"input_from,omitempty"`
	Params      json.RawMessage `json:"params,omitempty"`
}

// ChainCreateResponse представляет ответ команды создания цепочки моделей
type ChainCreateResponse struct {
	ChainID     string      `json:"chain_id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Steps       []ChainStep `json:"steps"`
	CreatedAt   time.Time   `json:"created_at"`
	Success     bool        `json:"success"`
	Message     string      `json:"message,omitempty"`
}

// InteractiveChainCreateResponse представляет ответ для интерактивного создания цепочки
type InteractiveChainCreateResponse struct {
	Status        string    `json:"status"`
	CurrentStep   int       `json:"current_step"`
	TotalSteps    int       `json:"total_steps"`
	CurrentPrompt string    `json:"current_prompt"`
	ChainID       string    `json:"chain_id,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
}

// HandleChainCreate обрабатывает MCP-команду для создания цепочки моделей
func HandleChainCreate(params json.RawMessage) (interface{}, error) {
	var p ChainCreateParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("failed to parse chain_create params: %v", err)
	}

	// В реальной реализации здесь будет создание цепочки моделей
	// и сохранение её в хранилище

	// Это демонстрационный пример, создающий фиктивную цепочку
	chainID := fmt.Sprintf("chain-%d", time.Now().Unix())
	createdAt := time.Now()

	// Если запрошено интерактивное создание, возвращаем промежуточный ответ
	if p.Interactive {
		return InteractiveChainCreateResponse{
			Status:        "in_progress",
			CurrentStep:   1,
			TotalSteps:    4,
			CurrentPrompt: "Выберите модель для первого шага цепочки",
			ChainID:       "",
			CreatedAt:     time.Time{},
		}, nil
	}

	// Валидация параметров
	if p.Name == "" {
		return nil, fmt.Errorf("chain name is required")
	}
	if len(p.Steps) == 0 {
		return nil, fmt.Errorf("at least one step is required")
	}

	// Проверка каждого шага на наличие роли и имени
	for i, step := range p.Steps {
		if step.RoleID == "" {
			return nil, fmt.Errorf("role_id is required for step %d", i+1)
		}
		if step.Name == "" {
			return nil, fmt.Errorf("name is required for step %d", i+1)
		}
	}

	response := ChainCreateResponse{
		ChainID:     chainID,
		Name:        p.Name,
		Description: p.Description,
		Steps:       p.Steps,
		CreatedAt:   createdAt,
		Success:     true,
		Message:     fmt.Sprintf("Chain '%s' created successfully", p.Name),
	}

	return response, nil
}

// HandleChainCreateStep обрабатывает шаг интерактивного создания цепочки
func HandleChainCreateStep(params json.RawMessage) (interface{}, error) {
	// В реальной реализации здесь будет обработка отдельного шага
	// интерактивного создания цепочки

	// Это демонстрационный пример
	var rawParams map[string]interface{}
	if err := json.Unmarshal(params, &rawParams); err != nil {
		return nil, fmt.Errorf("failed to parse chain_create_step params: %v", err)
	}

	// Получаем текущий шаг и общее количество шагов
	currentStepFloat, ok := rawParams["current_step"].(float64)
	if !ok {
		return nil, fmt.Errorf("current_step is required")
	}
	currentStep := int(currentStepFloat)

	// Симулируем прогресс создания цепочки
	var nextStep int
	var nextPrompt string
	var status string
	var chainID string
	var createdAt time.Time

	if currentStep < 4 {
		nextStep = currentStep + 1
		status = "in_progress"
		chainID = ""
		createdAt = time.Time{}

		switch nextStep {
		case 2:
			nextPrompt = "Опишите задачу для первого шага цепочки"
		case 3:
			nextPrompt = "Выберите модель для второго шага цепочки"
		case 4:
			nextPrompt = "Опишите задачу для второго шага цепочки"
		}
	} else {
		// Последний шаг - цепочка создана
		nextStep = 4
		status = "completed"
		nextPrompt = ""
		chainID = fmt.Sprintf("chain-%d", time.Now().Unix())
		createdAt = time.Now()
	}

	response := InteractiveChainCreateResponse{
		Status:        status,
		CurrentStep:   nextStep,
		TotalSteps:    4,
		CurrentPrompt: nextPrompt,
		ChainID:       chainID,
		CreatedAt:     createdAt,
	}

	return response, nil
}

// RegisterChainCreateCommands регистрирует команды для создания цепочек в MCP-сервере
func RegisterChainCreateCommands(server *MCPServer) {
	server.RegisterCommand("chain_create", HandleChainCreate)
	server.RegisterCommand("chain_create_step", HandleChainCreateStep)
}

// Обновляем InitMCPServer, чтобы включить наши новые команды:
/*
func InitMCPServer() *MCPServer {
	server := NewMCPServer()

	// Регистрация команд
	RegisterChainProgressCommand(server)
	RegisterModelCommands(server)
	RegisterChainCreateCommands(server)

	return server
}
*/
