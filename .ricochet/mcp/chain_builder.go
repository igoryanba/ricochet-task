package mcp

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/chain"
)

// ChainBuilderSession представляет сессию конструирования цепочки
type ChainBuilderSession struct {
	ID          string                 `json:"id"`
	ChainName   string                 `json:"chain_name"`
	ChainDesc   string                 `json:"chain_description"`
	Steps       []BuilderStep          `json:"steps"`
	CurrentStep int                    `json:"current_step"`
	Status      string                 `json:"status"` // "editing", "completed", "canceled"
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// BuilderStep представляет шаг в процессе конструирования цепочки
type BuilderStep struct {
	Index       int                    `json:"index"`
	ModelRole   string                 `json:"model_role"`
	ModelID     string                 `json:"model_id"`
	Provider    string                 `json:"provider"`
	Description string                 `json:"description"`
	Prompt      string                 `json:"prompt"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	IsCompleted bool                   `json:"is_completed"`
}

// ChainBuilderInitParams параметры для инициализации конструктора цепочек
type ChainBuilderInitParams struct {
	ChainName        string                 `json:"chain_name"`
	ChainDescription string                 `json:"chain_description,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	TemplateID       string                 `json:"template_id,omitempty"`
}

// ChainBuilderStepParams параметры для шага конструктора цепочек
type ChainBuilderStepParams struct {
	SessionID   string                 `json:"session_id"`
	StepIndex   int                    `json:"step_index"`
	ModelRole   string                 `json:"model_role"`
	ModelID     string                 `json:"model_id"`
	Provider    string                 `json:"provider"`
	Description string                 `json:"description"`
	Prompt      string                 `json:"prompt"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ChainBuilderResponse ответ конструктора цепочек
type ChainBuilderResponse struct {
	SessionID   string    `json:"session_id"`
	Status      string    `json:"status"`
	CurrentStep int       `json:"current_step"`
	TotalSteps  int       `json:"total_steps"`
	Message     string    `json:"message"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SessionCompleteParams параметры для завершения сессии конструктора
type SessionCompleteParams struct {
	SessionID string `json:"session_id"`
	Save      bool   `json:"save"` // Сохранить или отменить цепочку
}

// AutoSelectModelsParams параметры для автоматического выбора моделей в цепочке
type AutoSelectModelsParams struct {
	ChainID string `json:"chain_id"` // ID цепочки
}

// AutoSelectModelsResponse ответ на автоматический выбор моделей
type AutoSelectModelsResponse struct {
	ChainID string                 `json:"chain_id"`
	Steps   []AutoSelectedStepInfo `json:"steps"`
	Success bool                   `json:"success"`
	Message string                 `json:"message,omitempty"`
}

// AutoSelectedStepInfo информация о выбранной модели для шага
type AutoSelectedStepInfo struct {
	StepID           string `json:"step_id"`
	StepName         string `json:"step_name"`
	SelectedRole     string `json:"selected_role"`
	SelectedModel    string `json:"selected_model"`
	SelectedProvider string `json:"selected_provider"`
}

// activeSessions хранит активные сессии конструктора
var activeSessions = struct {
	sessions map[string]*ChainBuilderSession
	mutex    sync.RWMutex
}{
	sessions: make(map[string]*ChainBuilderSession),
}

// HandleChainBuilderInit обрабатывает запрос на инициализацию конструктора цепочек
func HandleChainBuilderInit(params json.RawMessage) (interface{}, error) {
	var initParams ChainBuilderInitParams
	if err := json.Unmarshal(params, &initParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для инициализации конструктора: %v", err)
	}

	if initParams.ChainName == "" {
		return nil, fmt.Errorf("chain_name является обязательным параметром")
	}

	// Создаем новую сессию
	sessionID := generateSessionID()
	now := time.Now()

	session := &ChainBuilderSession{
		ID:          sessionID,
		ChainName:   initParams.ChainName,
		ChainDesc:   initParams.ChainDescription,
		Steps:       make([]BuilderStep, 0),
		CurrentStep: 0,
		Status:      "editing",
		Metadata:    initParams.Metadata,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Если указан шаблон, загружаем его
	if initParams.TemplateID != "" {
		if err := loadTemplateIntoSession(session, initParams.TemplateID); err != nil {
			return nil, err
		}
	}

	// Сохраняем сессию в активные
	activeSessions.mutex.Lock()
	activeSessions.sessions[sessionID] = session
	activeSessions.mutex.Unlock()

	response := ChainBuilderResponse{
		SessionID:   sessionID,
		Status:      "editing",
		CurrentStep: 0,
		TotalSteps:  0,
		Message:     "Сессия конструктора цепочек создана",
		UpdatedAt:   now,
	}

	return response, nil
}

// HandleChainBuilderAddStep обрабатывает запрос на добавление шага в конструктор цепочек
func HandleChainBuilderAddStep(params json.RawMessage) (interface{}, error) {
	var stepParams ChainBuilderStepParams
	if err := json.Unmarshal(params, &stepParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для добавления шага: %v", err)
	}

	if stepParams.SessionID == "" {
		return nil, fmt.Errorf("session_id является обязательным параметром")
	}

	// Получаем сессию
	activeSessions.mutex.Lock()
	defer activeSessions.mutex.Unlock()

	session, exists := activeSessions.sessions[stepParams.SessionID]
	if !exists {
		return nil, fmt.Errorf("сессия с ID %s не найдена", stepParams.SessionID)
	}

	if session.Status != "editing" {
		return nil, fmt.Errorf("невозможно добавить шаг: сессия уже %s", session.Status)
	}

	// Валидируем индекс шага
	if stepParams.StepIndex < 0 {
		stepParams.StepIndex = len(session.Steps)
	} else if stepParams.StepIndex > len(session.Steps) {
		return nil, fmt.Errorf("индекс шага выходит за пределы существующих шагов")
	}

	// Создаем новый шаг
	step := BuilderStep{
		Index:       stepParams.StepIndex,
		ModelRole:   stepParams.ModelRole,
		ModelID:     stepParams.ModelID,
		Provider:    stepParams.Provider,
		Description: stepParams.Description,
		Prompt:      stepParams.Prompt,
		Parameters:  stepParams.Parameters,
		IsCompleted: true,
	}

	// Добавляем шаг в сессию
	if stepParams.StepIndex == len(session.Steps) {
		session.Steps = append(session.Steps, step)
	} else {
		// Вставляем шаг и перенумеровываем последующие
		newSteps := make([]BuilderStep, 0, len(session.Steps)+1)
		newSteps = append(newSteps, session.Steps[:stepParams.StepIndex]...)
		newSteps = append(newSteps, step)

		for i, s := range session.Steps[stepParams.StepIndex:] {
			s.Index = stepParams.StepIndex + i + 1
			newSteps = append(newSteps, s)
		}

		session.Steps = newSteps
	}

	// Обновляем сессию
	session.CurrentStep = stepParams.StepIndex + 1
	session.UpdatedAt = time.Now()

	response := ChainBuilderResponse{
		SessionID:   stepParams.SessionID,
		Status:      "editing",
		CurrentStep: session.CurrentStep,
		TotalSteps:  len(session.Steps),
		Message:     fmt.Sprintf("Шаг %d добавлен в цепочку", stepParams.StepIndex),
		UpdatedAt:   session.UpdatedAt,
	}

	return response, nil
}

// HandleChainBuilderEditStep обрабатывает запрос на редактирование шага
func HandleChainBuilderEditStep(params json.RawMessage) (interface{}, error) {
	var stepParams ChainBuilderStepParams
	if err := json.Unmarshal(params, &stepParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для редактирования шага: %v", err)
	}

	if stepParams.SessionID == "" {
		return nil, fmt.Errorf("session_id является обязательным параметром")
	}

	// Получаем сессию
	activeSessions.mutex.Lock()
	defer activeSessions.mutex.Unlock()

	session, exists := activeSessions.sessions[stepParams.SessionID]
	if !exists {
		return nil, fmt.Errorf("сессия с ID %s не найдена", stepParams.SessionID)
	}

	if session.Status != "editing" {
		return nil, fmt.Errorf("невозможно редактировать шаг: сессия уже %s", session.Status)
	}

	// Проверяем существование шага
	if stepParams.StepIndex < 0 || stepParams.StepIndex >= len(session.Steps) {
		return nil, fmt.Errorf("шаг с индексом %d не существует", stepParams.StepIndex)
	}

	// Обновляем шаг
	step := &session.Steps[stepParams.StepIndex]
	step.ModelRole = stepParams.ModelRole
	step.ModelID = stepParams.ModelID
	step.Provider = stepParams.Provider
	step.Description = stepParams.Description
	step.Prompt = stepParams.Prompt
	if stepParams.Parameters != nil {
		step.Parameters = stepParams.Parameters
	}

	// Обновляем сессию
	session.UpdatedAt = time.Now()

	response := ChainBuilderResponse{
		SessionID:   stepParams.SessionID,
		Status:      "editing",
		CurrentStep: session.CurrentStep,
		TotalSteps:  len(session.Steps),
		Message:     fmt.Sprintf("Шаг %d обновлен", stepParams.StepIndex),
		UpdatedAt:   session.UpdatedAt,
	}

	return response, nil
}

// HandleChainBuilderRemoveStep обрабатывает запрос на удаление шага
func HandleChainBuilderRemoveStep(params json.RawMessage) (interface{}, error) {
	var removeParams struct {
		SessionID string `json:"session_id"`
		StepIndex int    `json:"step_index"`
	}

	if err := json.Unmarshal(params, &removeParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для удаления шага: %v", err)
	}

	if removeParams.SessionID == "" {
		return nil, fmt.Errorf("session_id является обязательным параметром")
	}

	// Получаем сессию
	activeSessions.mutex.Lock()
	defer activeSessions.mutex.Unlock()

	session, exists := activeSessions.sessions[removeParams.SessionID]
	if !exists {
		return nil, fmt.Errorf("сессия с ID %s не найдена", removeParams.SessionID)
	}

	if session.Status != "editing" {
		return nil, fmt.Errorf("невозможно удалить шаг: сессия уже %s", session.Status)
	}

	// Проверяем существование шага
	if removeParams.StepIndex < 0 || removeParams.StepIndex >= len(session.Steps) {
		return nil, fmt.Errorf("шаг с индексом %d не существует", removeParams.StepIndex)
	}

	// Удаляем шаг
	newSteps := make([]BuilderStep, 0, len(session.Steps)-1)
	newSteps = append(newSteps, session.Steps[:removeParams.StepIndex]...)

	for i, step := range session.Steps[removeParams.StepIndex+1:] {
		step.Index = removeParams.StepIndex + i
		newSteps = append(newSteps, step)
	}

	session.Steps = newSteps

	// Обновляем текущий шаг и сессию
	if session.CurrentStep > removeParams.StepIndex {
		session.CurrentStep--
	}
	if session.CurrentStep >= len(session.Steps) {
		session.CurrentStep = len(session.Steps) - 1
		if session.CurrentStep < 0 {
			session.CurrentStep = 0
		}
	}

	session.UpdatedAt = time.Now()

	response := ChainBuilderResponse{
		SessionID:   removeParams.SessionID,
		Status:      "editing",
		CurrentStep: session.CurrentStep,
		TotalSteps:  len(session.Steps),
		Message:     fmt.Sprintf("Шаг %d удален", removeParams.StepIndex),
		UpdatedAt:   session.UpdatedAt,
	}

	return response, nil
}

// HandleChainBuilderGetSession обрабатывает запрос на получение данных сессии
func HandleChainBuilderGetSession(params json.RawMessage) (interface{}, error) {
	var getParams struct {
		SessionID string `json:"session_id"`
	}

	if err := json.Unmarshal(params, &getParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для получения сессии: %v", err)
	}

	if getParams.SessionID == "" {
		return nil, fmt.Errorf("session_id является обязательным параметром")
	}

	// Получаем сессию
	activeSessions.mutex.RLock()
	defer activeSessions.mutex.RUnlock()

	session, exists := activeSessions.sessions[getParams.SessionID]
	if !exists {
		return nil, fmt.Errorf("сессия с ID %s не найдена", getParams.SessionID)
	}

	return session, nil
}

// HandleChainBuilderComplete обрабатывает запрос на завершение конструирования цепочки
func HandleChainBuilderComplete(params json.RawMessage) (interface{}, error) {
	var completeParams SessionCompleteParams
	if err := json.Unmarshal(params, &completeParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для завершения сессии: %v", err)
	}

	if completeParams.SessionID == "" {
		return nil, fmt.Errorf("session_id является обязательным параметром")
	}

	// Получаем сессию
	activeSessions.mutex.Lock()
	defer activeSessions.mutex.Unlock()

	session, exists := activeSessions.sessions[completeParams.SessionID]
	if !exists {
		return nil, fmt.Errorf("сессия с ID %s не найдена", completeParams.SessionID)
	}

	if session.Status != "editing" {
		return nil, fmt.Errorf("невозможно завершить сессию: она уже %s", session.Status)
	}

	// Проверяем, что есть хотя бы один шаг
	if len(session.Steps) == 0 && completeParams.Save {
		return nil, fmt.Errorf("невозможно сохранить пустую цепочку")
	}

	// Обновляем статус сессии
	if completeParams.Save {
		session.Status = "completed"

		// Создаем цепочку и сохраняем ее
		chainID, err := createChainFromSession(session)
		if err != nil {
			return nil, fmt.Errorf("ошибка при создании цепочки: %v", err)
		}

		session.UpdatedAt = time.Now()

		response := struct {
			SessionID string    `json:"session_id"`
			ChainID   string    `json:"chain_id"`
			ChainName string    `json:"chain_name"`
			Status    string    `json:"status"`
			Message   string    `json:"message"`
			UpdatedAt time.Time `json:"updated_at"`
		}{
			SessionID: completeParams.SessionID,
			ChainID:   chainID,
			ChainName: session.ChainName,
			Status:    "completed",
			Message:   "Цепочка успешно создана",
			UpdatedAt: session.UpdatedAt,
		}

		return response, nil
	} else {
		session.Status = "canceled"
		session.UpdatedAt = time.Now()

		response := ChainBuilderResponse{
			SessionID:   completeParams.SessionID,
			Status:      "canceled",
			CurrentStep: session.CurrentStep,
			TotalSteps:  len(session.Steps),
			Message:     "Создание цепочки отменено",
			UpdatedAt:   session.UpdatedAt,
		}

		return response, nil
	}
}

// HandleAutoSelectModels обработчик автоматического выбора моделей для цепочки
func HandleAutoSelectModels(params json.RawMessage) (interface{}, error) {
	var p AutoSelectModelsParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unable to parse params: %v", err)
	}

	// Проверяем, что указан ID цепочки
	if p.ChainID == "" {
		return nil, fmt.Errorf("chain_id is required")
	}

	// Получаем цепочку из хранилища
	chain, err := getChain(p.ChainID)
	if err != nil {
		return nil, err
	}

	// Получаем менеджер моделей
	mm, err := getModelManager()
	if err != nil {
		return nil, err
	}

	// Информация о выбранных моделях
	selectedModels := make([]AutoSelectedStepInfo, 0, len(chain.Steps))

	// Для каждого шага цепочки выбираем подходящую модель
	for _, step := range chain.Steps {
		// Определяем роль на основе типа шага
		roleID := getRoleForStepType(step.Type)

		// Получаем модель для этой роли
		role := mm.GetModelForRole(roleID)

		// Если модель не найдена, пропускаем шаг
		if role.ModelID == "" || role.Provider == "" {
			continue
		}

		// Обновляем шаг с выбранной моделью
		err = updateStepModel(p.ChainID, step.ID, role.Provider, role.ModelID, roleID)
		if err != nil {
			// Если не удалось обновить шаг, продолжаем с другими
			continue
		}

		// Добавляем информацию о выбранной модели
		selectedModels = append(selectedModels, AutoSelectedStepInfo{
			StepID:           step.ID,
			StepName:         step.Name,
			SelectedRole:     roleID,
			SelectedModel:    role.ModelID,
			SelectedProvider: role.Provider,
		})
	}

	// Формируем ответ
	response := AutoSelectModelsResponse{
		ChainID: p.ChainID,
		Steps:   selectedModels,
		Success: len(selectedModels) > 0,
	}

	if len(selectedModels) > 0 {
		response.Message = fmt.Sprintf("Successfully selected models for %d steps", len(selectedModels))
	} else {
		response.Message = "No models were selected for the chain steps"
	}

	return response, nil
}

// getRoleForStepType определяет роль модели на основе типа шага
func getRoleForStepType(stepType string) string {
	switch stepType {
	case "analysis", "classify":
		return "analyzer"
	case "summarize":
		return "summarizer"
	case "extract":
		return "extractor"
	case "evaluate", "review":
		return "critic"
	case "refine", "improve":
		return "refiner"
	case "generate", "create":
		return "creator"
	case "integrate", "combine":
		return "integrator"
	default:
		// По умолчанию используем основную модель
		return "main"
	}
}

// updateStepModel обновляет модель шага
func updateStepModel(chainID, stepID, provider, modelID, roleID string) error {
	// Получаем цепочку из хранилища
	chain, err := getChain(chainID)
	if err != nil {
		return err
	}

	// Ищем шаг по ID
	var stepIndex = -1
	for i, step := range chain.Steps {
		if step.ID == stepID {
			stepIndex = i
			break
		}
	}

	if stepIndex == -1 {
		return fmt.Errorf("step not found: %s", stepID)
	}

	// Обновляем модель шага
	chain.Steps[stepIndex].ModelProvider = provider
	chain.Steps[stepIndex].ModelID = modelID
	chain.Steps[stepIndex].RoleID = roleID

	// Сохраняем обновленную цепочку
	return saveChain(chain)
}

// RegisterChainBuilderCommands регистрирует команды для работы с построителем цепочек
func RegisterChainBuilderCommands(server *MCPServer) {
	server.RegisterCommand("chain_builder_init", HandleChainBuilderInit)
	server.RegisterCommand("chain_builder_add_step", HandleChainBuilderAddStep)
	server.RegisterCommand("chain_builder_edit_step", HandleChainBuilderEditStep)
	server.RegisterCommand("chain_builder_remove_step", HandleChainBuilderRemoveStep)
	server.RegisterCommand("chain_builder_get_session", HandleChainBuilderGetSession)
	server.RegisterCommand("chain_builder_complete", HandleChainBuilderComplete)
	server.RegisterCommand("auto_select_models", HandleAutoSelectModels)
}

// Вспомогательные функции

// generateSessionID генерирует уникальный ID сессии
func generateSessionID() string {
	// Временная реализация - использовать UUID в реальном коде
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}

// loadTemplateIntoSession загружает шаблон в сессию
func loadTemplateIntoSession(session *ChainBuilderSession, templateID string) error {
	// TODO: Загрузить шаблон из хранилища шаблонов
	// Временная реализация с тестовыми данными
	switch templateID {
	case "analyze-document":
		session.Steps = []BuilderStep{
			{
				Index:       0,
				ModelRole:   "analyzer",
				ModelID:     "gpt-4",
				Provider:    "openai",
				Description: "Анализ структуры документа",
				Prompt:      "Проанализируйте структуру и основные темы документа. Выделите ключевые разделы и их взаимосвязи.",
				Parameters:  map[string]interface{}{"temperature": 0.3},
				IsCompleted: true,
			},
			{
				Index:       1,
				ModelRole:   "summarizer",
				ModelID:     "claude-3-opus",
				Provider:    "anthropic",
				Description: "Суммаризация документа",
				Prompt:      "На основе анализа структуры, создайте краткое резюме документа, выделив ключевые идеи и выводы.",
				Parameters:  map[string]interface{}{"temperature": 0.4},
				IsCompleted: true,
			},
		}
	case "code-review":
		session.Steps = []BuilderStep{
			{
				Index:       0,
				ModelRole:   "analyzer",
				ModelID:     "deepseek-coder",
				Provider:    "deepseek",
				Description: "Анализ кода",
				Prompt:      "Проанализируйте представленный код. Выделите основные компоненты, архитектурные решения и потенциальные проблемы.",
				Parameters:  map[string]interface{}{"temperature": 0.2},
				IsCompleted: true,
			},
			{
				Index:       1,
				ModelRole:   "reviewer",
				ModelID:     "gpt-4",
				Provider:    "openai",
				Description: "Код-ревью",
				Prompt:      "На основе анализа кода, проведите детальное код-ревью. Отметьте проблемы, предложите улучшения и оцените качество кода.",
				Parameters:  map[string]interface{}{"temperature": 0.3},
				IsCompleted: true,
			},
		}
	default:
		// Пустой шаблон
	}

	return nil
}

// createChainFromSession создает цепочку на основе сессии конструктора
func createChainFromSession(session *ChainBuilderSession) (string, error) {
	// TODO: Создать цепочку и сохранить ее в хранилище
	// Временная реализация

	// Создаем модели
	models := make([]chain.Model, 0, len(session.Steps))
	for i, step := range session.Steps {
		model := chain.Model{
			ID:        fmt.Sprintf("model-%d", i),
			Name:      chain.ModelName(step.ModelID),
			Type:      chain.ModelType(step.Provider),
			Role:      chain.ModelRole(step.ModelRole),
			Prompt:    step.Prompt,
			Order:     i,
			MaxTokens: 2000, // Значение по умолчанию
		}

		// Добавляем параметры, если они есть
		if val, ok := step.Parameters["temperature"]; ok {
			if temp, ok := val.(float64); ok {
				model.Temperature = temp
			}
		}

		// Другие параметры можно добавить здесь

		models = append(models, model)
	}

	// Создаем цепочку
	chainID := fmt.Sprintf("chain-%d", time.Now().UnixNano())

	// В реальной реализации здесь было бы сохранение цепочки в хранилище

	return chainID, nil
}
