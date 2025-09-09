package mcp

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// ChainInteractiveBuilderParams параметры для интерактивного конструктора цепочек
type ChainInteractiveBuilderParams struct {
	SessionID  string `json:"session_id,omitempty"`
	ChainID    string `json:"chain_id,omitempty"` // Для редактирования существующей цепочки
	ChainName  string `json:"chain_name,omitempty"`
	Format     string `json:"format,omitempty"`      // ui, mermaid, text
	EditorMode string `json:"editor_mode,omitempty"` // inline, panel, dialog
}

// ChainInteractiveBuilderResponse ответ интерактивного конструктора цепочек
type ChainInteractiveBuilderResponse struct {
	SessionID     string    `json:"session_id"`
	EditorContent string    `json:"editor_content"` // HTML/мермейд/текст редактора
	Format        string    `json:"format"`
	EditorMode    string    `json:"editor_mode"`
	ChainName     string    `json:"chain_name,omitempty"`
	ChainID       string    `json:"chain_id,omitempty"`
	ModelsCount   int       `json:"models_count"`
	Status        string    `json:"status"`
	Message       string    `json:"message,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ModelSelectionResponse ответ на запрос списка доступных моделей
type ModelSelectionResponse struct {
	Models []ModelOption `json:"models"`
}

// HandleChainInteractiveBuilder обрабатывает запрос на открытие интерактивного конструктора цепочек
func HandleChainInteractiveBuilder(params json.RawMessage) (interface{}, error) {
	var builderParams ChainInteractiveBuilderParams
	if err := json.Unmarshal(params, &builderParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для интерактивного конструктора: %v", err)
	}

	// Установить значения по умолчанию
	if builderParams.Format == "" {
		builderParams.Format = "ui"
	}
	if builderParams.EditorMode == "" {
		builderParams.EditorMode = "panel"
	}

	// Создаем новую сессию или загружаем существующую
	var sessionID string
	var chainName string
	var chainID string
	var status string
	var message string
	var modelsCount int

	now := time.Now()

	if builderParams.SessionID != "" {
		// Загружаем существующую сессию
		sessionID = builderParams.SessionID
		// TODO: загрузить данные сессии из хранилища
		chainName = "Сессия " + sessionID
		status = "active"
		message = "Сессия восстановлена"
		modelsCount = 0
	} else if builderParams.ChainID != "" {
		// Редактируем существующую цепочку
		chainID = builderParams.ChainID
		// TODO: загрузить данные цепочки из хранилища
		chainName = "Цепочка " + chainID
		sessionID = "session-" + generateUniqueID()
		status = "editing"
		message = "Редактирование существующей цепочки"
		modelsCount = 0
	} else {
		// Создаем новую сессию
		sessionID = "session-" + generateUniqueID()
		chainName = builderParams.ChainName
		if chainName == "" {
			chainName = "Новая цепочка"
		}
		status = "new"
		message = "Создана новая сессия для конструирования цепочки"
		modelsCount = 0
	}

	// Генерируем содержимое редактора в зависимости от формата
	var editorContent string
	switch builderParams.Format {
	case "ui":
		editorContent = generateUIEditor(sessionID, chainName, modelsCount)
	case "mermaid":
		editorContent = generateMermaidEditor(sessionID, chainName, modelsCount)
	case "text":
		editorContent = generateTextEditor(sessionID, chainName, modelsCount)
	default:
		return nil, fmt.Errorf("неподдерживаемый формат редактора: %s", builderParams.Format)
	}

	response := ChainInteractiveBuilderResponse{
		SessionID:     sessionID,
		EditorContent: editorContent,
		Format:        builderParams.Format,
		EditorMode:    builderParams.EditorMode,
		ChainName:     chainName,
		ChainID:       chainID,
		ModelsCount:   modelsCount,
		Status:        status,
		Message:       message,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	return response, nil
}

// HandleChainAddModel обрабатывает запрос на добавление модели в конструктор цепочек
func HandleChainAddModel(params json.RawMessage) (interface{}, error) {
	var addParams struct {
		SessionID string `json:"session_id"`
		Provider  string `json:"provider"`
		ModelID   string `json:"model_id"`
		Role      string `json:"role"`
		Position  int    `json:"position,omitempty"` // Если не указано, добавляем в конец
	}

	if err := json.Unmarshal(params, &addParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для добавления модели: %v", err)
	}

	if addParams.SessionID == "" {
		return nil, fmt.Errorf("session_id является обязательным параметром")
	}
	if addParams.Provider == "" || addParams.ModelID == "" {
		return nil, fmt.Errorf("provider и model_id являются обязательными параметрами")
	}

	// TODO: Добавить модель в сессию конструктора

	return map[string]interface{}{
		"success":     true,
		"session_id":  addParams.SessionID,
		"message":     fmt.Sprintf("Модель %s/%s добавлена в цепочку", addParams.Provider, addParams.ModelID),
		"model_added": true,
	}, nil
}

// HandleChainRemoveModel обрабатывает запрос на удаление модели из конструктора цепочек
func HandleChainRemoveModel(params json.RawMessage) (interface{}, error) {
	var removeParams struct {
		SessionID string `json:"session_id"`
		Position  int    `json:"position"`
	}

	if err := json.Unmarshal(params, &removeParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для удаления модели: %v", err)
	}

	if removeParams.SessionID == "" {
		return nil, fmt.Errorf("session_id является обязательным параметром")
	}
	if removeParams.Position < 0 {
		return nil, fmt.Errorf("position должен быть неотрицательным числом")
	}

	// TODO: Удалить модель из сессии конструктора

	return map[string]interface{}{
		"success":       true,
		"session_id":    removeParams.SessionID,
		"message":       fmt.Sprintf("Модель на позиции %d удалена из цепочки", removeParams.Position),
		"model_removed": true,
	}, nil
}

// HandleChainMoveModel обрабатывает запрос на перемещение модели в конструкторе цепочек
func HandleChainMoveModel(params json.RawMessage) (interface{}, error) {
	var moveParams struct {
		SessionID string `json:"session_id"`
		FromPos   int    `json:"from_position"`
		ToPos     int    `json:"to_position"`
	}

	if err := json.Unmarshal(params, &moveParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для перемещения модели: %v", err)
	}

	if moveParams.SessionID == "" {
		return nil, fmt.Errorf("session_id является обязательным параметром")
	}
	if moveParams.FromPos < 0 || moveParams.ToPos < 0 {
		return nil, fmt.Errorf("позиции должны быть неотрицательными числами")
	}

	// TODO: Переместить модель в сессии конструктора

	return map[string]interface{}{
		"success":     true,
		"session_id":  moveParams.SessionID,
		"message":     fmt.Sprintf("Модель перемещена с позиции %d на позицию %d", moveParams.FromPos, moveParams.ToPos),
		"model_moved": true,
	}, nil
}

// HandleChainGetAvailableModels обрабатывает запрос на получение списка доступных моделей для конструктора
func HandleChainGetAvailableModels(params json.RawMessage) (interface{}, error) {
	var modelParams struct {
		Role string `json:"role,omitempty"` // Если указано, вернуть модели для конкретной роли
	}

	if err := json.Unmarshal(params, &modelParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для получения моделей: %v", err)
	}

	// Получаем менеджер моделей
	mm, err := getModelManager()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить менеджер моделей: %v", err)
	}

	// Получаем все доступные роли
	mmRoles := mm.GetAvailableRoles()

	// Собираем уникальные модели (ключ provider:modelID)
	modelMap := make(map[string]ModelOption)

	for _, r := range mmRoles {
		// Фильтр по роли, если указан
		if modelParams.Role != "" && r.ID != modelParams.Role {
			continue
		}

		if r.Provider == "" || r.ModelID == "" {
			continue
		}

		key := fmt.Sprintf("%s:%s", r.Provider, r.ModelID)
		if _, ok := modelMap[key]; ok {
			continue
		}

		modelMap[key] = ModelOption{
			Provider:    r.Provider,
			ModelID:     r.ModelID,
			DisplayName: r.DisplayName,
			Description: r.Description,
			// Прочие поля оставить нулевыми, т.к. их нет в ModelRole
		}
	}

	// Преобразуем карту в срез
	models := make([]ModelOption, 0, len(modelMap))
	for _, m := range modelMap {
		models = append(models, m)
	}

	// Сортируем по провайдеру + DisplayName
	sort.Slice(models, func(i, j int) bool {
		if models[i].Provider == models[j].Provider {
			return models[i].DisplayName < models[j].DisplayName
		}
		return models[i].Provider < models[j].Provider
	})

	return ModelSelectionResponse{Models: models}, nil
}

// HandleChainSaveInteractive обрабатывает запрос на сохранение цепочки из интерактивного конструктора
func HandleChainSaveInteractive(params json.RawMessage) (interface{}, error) {
	var saveParams struct {
		SessionID string `json:"session_id"`
		ChainName string `json:"chain_name,omitempty"`
	}

	if err := json.Unmarshal(params, &saveParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для сохранения цепочки: %v", err)
	}

	if saveParams.SessionID == "" {
		return nil, fmt.Errorf("session_id является обязательным параметром")
	}

	// TODO: Сохранить цепочку из сессии конструктора
	chainID := "chain-" + generateUniqueID()

	return map[string]interface{}{
		"success":    true,
		"session_id": saveParams.SessionID,
		"chain_id":   chainID,
		"message":    "Цепочка успешно сохранена",
	}, nil
}

// RegisterChainInteractiveBuilderCommands регистрирует команды интерактивного конструктора цепочек
func RegisterChainInteractiveBuilderCommands(server *MCPServer) {
	server.RegisterCommand("chain_interactive_builder", HandleChainInteractiveBuilder)
	server.RegisterCommand("chain_add_model", HandleChainAddModel)
	server.RegisterCommand("chain_remove_model", HandleChainRemoveModel)
	server.RegisterCommand("chain_move_model", HandleChainMoveModel)
	server.RegisterCommand("chain_get_available_models", HandleChainGetAvailableModels)
	server.RegisterCommand("chain_save_interactive", HandleChainSaveInteractive)
}

// Вспомогательные функции

// generateUniqueID генерирует уникальный ID
func generateUniqueID() string {
	// Простая реализация для примера
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// generateUIEditor генерирует HTML-интерфейс для редактора
func generateUIEditor(sessionID, chainName string, modelsCount int) string {
	// Пример простой HTML-разметки для редактора
	return fmt.Sprintf(`
<div class="ricochet-chain-editor" data-session-id="%s">
  <div class="editor-header">
    <h3>%s</h3>
    <div class="editor-tools">
      <button class="add-model-btn">Добавить модель</button>
      <button class="save-chain-btn">Сохранить цепочку</button>
    </div>
  </div>
  <div class="models-container">
    %s
  </div>
</div>
`, sessionID, chainName, getModelsPlaceholder(modelsCount))
}

// generateMermaidEditor генерирует Mermaid-диаграмму для редактора
func generateMermaidEditor(sessionID, chainName string, modelsCount int) string {
	// Пример Mermaid-диаграммы для редактора
	return fmt.Sprintf(`graph LR
    title["%s"]
    %s
    style title fill:#f9f9f9,stroke:#333,stroke-width:1px
`, chainName, getMermaidModels(modelsCount))
}

// generateTextEditor генерирует текстовое представление для редактора
func generateTextEditor(sessionID, chainName string, modelsCount int) string {
	// Пример текстового представления для редактора
	return fmt.Sprintf(`Цепочка: %s
Сессия: %s
Модели: %d

%s
`, chainName, sessionID, modelsCount, getTextModels(modelsCount))
}

// getModelsPlaceholder возвращает заполнитель для моделей в UI-редакторе
func getModelsPlaceholder(count int) string {
	if count <= 0 {
		return `<div class="empty-state">Нет моделей. Нажмите "Добавить модель", чтобы начать.</div>`
	}

	placeholder := ""
	for i := 0; i < count; i++ {
		placeholder += fmt.Sprintf(`<div class="model-item" data-position="%d">
			<div class="model-header">Модель #%d</div>
			<div class="model-actions">
			  <button class="edit-model-btn">✎</button>
			  <button class="remove-model-btn">✕</button>
			</div>
		  </div>`, i, i+1)
	}
	return placeholder
}

// getMermaidModels возвращает представление моделей для Mermaid-диаграммы
func getMermaidModels(count int) string {
	if count <= 0 {
		return "    empty[\"Нет моделей\"]"
	}
	nodes := ""
	links := ""
	for i := 0; i < count; i++ {
		nodes += fmt.Sprintf("    model%d[\"Модель %d\"]\n", i, i+1)
		if i < count-1 {
			links += fmt.Sprintf("    model%d --> model%d\n", i, i+1)
		}
	}
	return nodes + links
}

// getTextModels возвращает текстовое представление моделей
func getTextModels(count int) string {
	if count <= 0 {
		return "Нет моделей в цепочке"
	}
	text := ""
	for i := 0; i < count; i++ {
		text += fmt.Sprintf("%d. Модель %d\n", i+1, i+1)
	}
	return text
}
