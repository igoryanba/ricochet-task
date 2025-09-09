package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/models_manager"
)

// ModelSetupParams представляет параметры для команды настройки моделей
type ModelSetupParams struct {
	Roles []string `json:"roles,omitempty"` // Список ролей для настройки, если пусто - все роли
}

// ModelOption представляет опцию выбора модели
type ModelOption struct {
	Provider     string   `json:"provider"`
	ModelID      string   `json:"model_id"`
	DisplayName  string   `json:"display_name"`
	MaxTokens    int      `json:"max_tokens"`
	Description  string   `json:"description,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	ContextSize  int      `json:"context_size,omitempty"`
	Cost         string   `json:"cost,omitempty"`
}

// ModelRole представляет роль модели с опциями выбора
type ModelRole struct {
	RoleID       string        `json:"role_id"`
	DisplayName  string        `json:"display_name"`
	Description  string        `json:"description,omitempty"`
	CurrentModel *ModelOption  `json:"current_model,omitempty"`
	Options      []ModelOption `json:"options"`
}

// ModelSetupResponse представляет ответ команды настройки моделей
type ModelSetupResponse struct {
	Roles []ModelRole `json:"roles"`
}

// SelectModelParams представляет параметры для выбора модели для роли
type SelectModelParams struct {
	RoleID       string          `json:"role_id"`
	Provider     string          `json:"provider"`
	ModelID      string          `json:"model_id"`
	CustomParams json.RawMessage `json:"custom_params,omitempty"`
}

// SelectModelResponse представляет ответ на выбор модели
type SelectModelResponse struct {
	RoleID      string `json:"role_id"`
	Provider    string `json:"provider"`
	ModelID     string `json:"model_id"`
	DisplayName string `json:"display_name"`
	Success     bool   `json:"success"`
	Message     string `json:"message,omitempty"`
}

// ModelListParams представляет параметры для запроса списка моделей
type ModelListParams struct {
	IncludeAll   bool     `json:"include_all,omitempty"`   // Включать все модели, а не только используемые
	ByProvider   string   `json:"by_provider,omitempty"`   // Фильтр по провайдеру
	ByCapability string   `json:"by_capability,omitempty"` // Фильтр по возможностям
	Roles        []string `json:"roles,omitempty"`         // Список ролей для включения
}

// ModelListResponse представляет ответ на запрос списка моделей
type ModelListResponse struct {
	Models map[string]ModelOption `json:"models"` // Ключ - роль
}

// RoleCapabilityMapping связывает роли с необходимыми возможностями моделей
var RoleCapabilityMapping = map[string][]string{
	"main":       {"text-generation", "context-aware"},
	"research":   {"text-generation", "research", "large-context"},
	"fallback":   {"text-generation", "fast-response"},
	"analyzer":   {"classification", "extraction", "analysis"},
	"summarizer": {"summarization", "extraction"},
	"integrator": {"text-generation", "synthesis", "large-context"},
	"extractor":  {"extraction", "classification"},
	"critic":     {"analysis", "evaluation"},
	"refiner":    {"text-generation", "editing"},
	"creator":    {"text-generation", "creative"},
}

// TaskMasterModelExportParams параметры для экспорта моделей в Task Master
type TaskMasterModelExportParams struct {
	ExportPath string `json:"export_path,omitempty"` // Путь для экспорта настроек
}

// TaskMasterModelExportResponse ответ на экспорт моделей в Task Master
type TaskMasterModelExportResponse struct {
	Success      bool   `json:"success"`
	ExportedPath string `json:"exported_path,omitempty"`
	Message      string `json:"message,omitempty"`
}

// TaskMasterImportParams параметры для импорта моделей из Task Master
type TaskMasterImportParams struct {
	ImportPath string `json:"import_path"` // Путь к файлу конфигурации Task Master
}

// TaskMasterImportResponse ответ на импорт моделей из Task Master
type TaskMasterImportResponse struct {
	Success    bool                   `json:"success"`
	Models     map[string]ModelOption `json:"models,omitempty"`
	Message    string                 `json:"message,omitempty"`
	ImportedAt time.Time              `json:"imported_at"`
}

// RecommendModelsParams параметры для рекомендации моделей
type RecommendModelsParams struct {
	RoleID     string   `json:"role_id"`              // Роль для рекомендации
	Priorities []string `json:"priorities,omitempty"` // Приоритетные возможности
}

// RecommendModelsResponse ответ на рекомендацию моделей
type RecommendModelsResponse struct {
	Role             string        `json:"role"`
	RecommendedModel *ModelOption  `json:"recommended_model,omitempty"`
	Alternatives     []ModelOption `json:"alternatives,omitempty"`
	Message          string        `json:"message,omitempty"`
}

// getModelManager возвращает экземпляр менеджера моделей
func getModelManager() (*models_manager.ModelsManager, error) {
	// В реальной реализации здесь должно быть получение пути к конфигурации из настроек проекта
	return models_manager.New("")
}

// HandleModelSetup обрабатывает MCP-команду для настройки моделей
func HandleModelSetup(params json.RawMessage) (interface{}, error) {
	var p ModelSetupParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("failed to parse models_setup params: %v", err)
	}

	// Получаем менеджер моделей
	manager, err := getModelManager()
	if err != nil {
		return nil, fmt.Errorf("failed to get model manager: %v", err)
	}

	// Получаем список ролей и конвертируем в MCP-формат
	roles := convertRoles(manager.GetAvailableRoles())

	// Если указаны конкретные роли, фильтруем их
	if len(p.Roles) > 0 {
		filteredRoles := make([]ModelRole, 0)
		for _, role := range roles {
			for _, requestedRole := range p.Roles {
				if role.RoleID == requestedRole {
					filteredRoles = append(filteredRoles, role)
					break
				}
			}
		}
		roles = filteredRoles
	}

	response := ModelSetupResponse{
		Roles: roles,
	}

	return response, nil
}

// convertRoles конвертирует роли из формата менеджера моделей в формат MCP
func convertRoles(mmRoles []models_manager.ModelRole) []ModelRole {
	roles := make([]ModelRole, len(mmRoles))

	for i, mmRole := range mmRoles {
		// Создаём структуру текущей модели, если в роли задан провайдер и ID модели
		var currentModel *ModelOption
		if mmRole.Provider != "" && mmRole.ModelID != "" {
			cm := ModelOption{
				Provider:    mmRole.Provider,
				ModelID:     mmRole.ModelID,
				DisplayName: mmRole.DisplayName,
				Description: mmRole.Description,
				// MaxTokens/Capabilities/etc. недоступны в ModelRole – оставляем нулевые значения
			}
			currentModel = &cm
		}

		// У менеджера моделей сейчас нет альтернативных опций для роли – оставляем пустой срез
		options := []ModelOption{}

		roles[i] = ModelRole{
			RoleID:       mmRole.ID,
			DisplayName:  mmRole.DisplayName,
			Description:  mmRole.Description,
			CurrentModel: currentModel,
			Options:      options,
		}
	}

	return roles
}

// HandleSelectModel обрабатывает MCP-команду для выбора модели для роли
func HandleSelectModel(params json.RawMessage) (interface{}, error) {
	var p SelectModelParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("failed to parse select_model params: %v", err)
	}

	// Пользовательские параметры пока игнорируются, так как сохранение модели не реализовано.

	response := SelectModelResponse{
		RoleID:      p.RoleID,
		Provider:    p.Provider,
		ModelID:     p.ModelID,
		DisplayName: p.ModelID,
		Success:     true,
		Message:     fmt.Sprintf("(demo) Модель %s/%s выбрана для роли %s (изменения не сохранены)", p.Provider, p.ModelID, p.RoleID),
	}

	return response, nil
}

// HandleModelList обрабатывает MCP-команду для получения списка моделей
func HandleModelList(params json.RawMessage) (interface{}, error) {
	var p ModelListParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("failed to parse models_list params: %v", err)
	}

	// Получаем менеджер моделей
	manager, err := getModelManager()
	if err != nil {
		return nil, fmt.Errorf("failed to get model manager: %v", err)
	}

	// Получаем список ролей и конвертируем в MCP-формат
	roles := convertRoles(manager.GetAvailableRoles())

	// Формируем ответ
	response := ModelListResponse{
		Models: make(map[string]ModelOption),
	}

	// Добавляем модели из ролей
	for _, role := range roles {
		// Пропускаем роли, которые не запрошены, если есть список ролей
		if len(p.Roles) > 0 {
			found := false
			for _, requestedRole := range p.Roles {
				if role.RoleID == requestedRole {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Добавляем текущую модель роли
		if role.CurrentModel != nil {
			// Фильтруем по провайдеру, если указан
			if p.ByProvider != "" && role.CurrentModel.Provider != p.ByProvider {
				continue
			}

			// Фильтруем по возможностям, если указаны
			if p.ByCapability != "" {
				found := false
				for _, cap := range role.CurrentModel.Capabilities {
					if cap == p.ByCapability {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			response.Models[role.RoleID] = ModelOption{
				Provider:     role.CurrentModel.Provider,
				ModelID:      role.CurrentModel.ModelID,
				DisplayName:  role.CurrentModel.DisplayName,
				MaxTokens:    role.CurrentModel.MaxTokens,
				Description:  role.CurrentModel.Description,
				Capabilities: role.CurrentModel.Capabilities,
				ContextSize:  role.CurrentModel.ContextSize,
				Cost:         role.CurrentModel.Cost,
			}
		}
	}

	return response, nil
}

// HandleTaskMasterExport обработчик экспорта настроек моделей в Task Master
func HandleTaskMasterExport(params json.RawMessage) (interface{}, error) {
	var p TaskMasterModelExportParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unable to parse params: %v", err)
	}

	mm, err := getModelManager()
	if err != nil {
		return nil, err
	}

	// Если путь не указан, используем значение по умолчанию
	exportPath := p.ExportPath
	if exportPath == "" {
		// Определяем путь по умолчанию в домашней директории пользователя
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("unable to determine home directory: %v", err)
		}
		exportPath = filepath.Join(homeDir, ".taskmaster", "config.json")
	}

	// Получаем настройки моделей
	roles := mm.GetRoles()

	// Формируем структуру для экспорта в Task Master
	taskMasterConfig := map[string]interface{}{
		"models": map[string]interface{}{},
	}

	// Добавляем основные модели
	for _, role := range roles {
		if role.ModelID == "" || role.Provider == "" {
			continue
		}

		modelKey := fmt.Sprintf("%s-%s", role.Provider, role.ModelID)
		modelName := fmt.Sprintf("%s (%s)", role.DisplayName, role.Provider)

		taskMasterConfig["models"].(map[string]interface{})[role.ID] = map[string]interface{}{
			"id":        modelKey,
			"name":      modelName,
			"provider":  role.Provider,
			"modelType": role.ID,
			"params":    role.Parameters,
		}
	}

	// Создаем директорию для файла, если она не существует
	if err := os.MkdirAll(filepath.Dir(exportPath), 0755); err != nil {
		return nil, fmt.Errorf("unable to create directory: %v", err)
	}

	// Сохраняем конфигурацию в файл
	configJSON, err := json.MarshalIndent(taskMasterConfig, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("unable to marshal config: %v", err)
	}

	if err := os.WriteFile(exportPath, configJSON, 0644); err != nil {
		return nil, fmt.Errorf("unable to write config file: %v", err)
	}

	return TaskMasterModelExportResponse{
		Success:      true,
		ExportedPath: exportPath,
		Message:      fmt.Sprintf("Models successfully exported to Task Master config at %s", exportPath),
	}, nil
}

// HandleTaskMasterImport обработчик импорта настроек моделей из Task Master
func HandleTaskMasterImport(params json.RawMessage) (interface{}, error) {
	var p TaskMasterImportParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unable to parse params: %v", err)
	}

	// Проверяем, что путь к файлу указан
	importPath := p.ImportPath
	if importPath == "" {
		// Пробуем найти файл по умолчанию
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("import path not specified and unable to determine home directory: %v", err)
		}
		importPath = filepath.Join(homeDir, ".taskmaster", "config.json")
	}

	// Проверяем, что файл существует
	if _, err := os.Stat(importPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("import file not found: %s", importPath)
	}

	// Читаем файл конфигурации
	configData, err := os.ReadFile(importPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file: %v", err)
	}

	// Парсим JSON
	var taskMasterConfig map[string]interface{}
	if err := json.Unmarshal(configData, &taskMasterConfig); err != nil {
		return nil, fmt.Errorf("unable to parse config file: %v", err)
	}

	// Проверяем, что есть секция моделей
	modelsSection, ok := taskMasterConfig["models"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid config format: missing models section")
	}

	// Импортируем модели
	importedModels := make(map[string]ModelOption)

	for roleID, modelData := range modelsSection {
		modelConfig, ok := modelData.(map[string]interface{})
		if !ok {
			continue
		}

		provider, _ := modelConfig["provider"].(string)
		modelID, _ := modelConfig["id"].(string)
		if strings.Contains(modelID, "-") {
			// Если ID содержит провайдера, извлекаем только ID модели
			parts := strings.SplitN(modelID, "-", 2)
			if len(parts) == 2 && parts[0] == provider {
				modelID = parts[1]
			}
		}

		name, _ := modelConfig["name"].(string)
		_, _ = modelConfig["params"].(map[string]interface{})

		if provider == "" || modelID == "" {
			continue
		}

		// TODO: метод сохранения модели в менеджер пока не реализован –
		// просто добавляем в список импортированных моделей.

		importedModels[roleID] = ModelOption{
			Provider:    provider,
			ModelID:     modelID,
			DisplayName: name,
		}
	}

	return TaskMasterImportResponse{
		Success:    len(importedModels) > 0,
		Models:     importedModels,
		Message:    fmt.Sprintf("Successfully imported %d models from Task Master config", len(importedModels)),
		ImportedAt: time.Now(),
	}, nil
}

// HandleRecommendModels обработчик рекомендации моделей для роли
func HandleRecommendModels(params json.RawMessage) (interface{}, error) {
	var p RecommendModelsParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unable to parse params: %v", err)
	}

	if p.RoleID == "" {
		return nil, fmt.Errorf("role_id is required")
	}

	mm, err := getModelManager()
	if err != nil {
		return nil, err
	}

	// Определяем необходимые возможности для этой роли
	requiredCapabilities := RoleCapabilityMapping[p.RoleID]
	if len(requiredCapabilities) == 0 {
		// Если для роли не определены возможности, используем базовые
		requiredCapabilities = []string{"text-generation"}
	}

	// Добавляем приоритетные возможности
	if len(p.Priorities) > 0 {
		for _, priority := range p.Priorities {
			if !contains(requiredCapabilities, priority) {
				requiredCapabilities = append(requiredCapabilities, priority)
			}
		}
	}

	// Получаем все доступные модели через доступные роли
	roleModels := convertRoles(mm.GetAvailableRoles())

	var allModels []ModelOption
	for _, r := range roleModels {
		if r.CurrentModel != nil {
			allModels = append(allModels, *r.CurrentModel)
		}
	}

	// Оцениваем каждую модель по соответствию требуемым возможностям
	type scoredModel struct {
		model ModelOption
		score int
	}

	var scoredModels []scoredModel

	for _, model := range allModels {
		score := 0

		// Проверяем соответствие возможностям
		for _, capability := range requiredCapabilities {
			if contains(model.Capabilities, capability) {
				score++
			}
		}

		// Добавляем модель, если она соответствует хотя бы одной возможности
		if score > 0 {
			scoredModels = append(scoredModels, scoredModel{model, score})
		}
	}

	// Сортируем модели по оценке (от высшей к низшей)
	sort.Slice(scoredModels, func(i, j int) bool {
		return scoredModels[i].score > scoredModels[j].score
	})

	// Формируем ответ
	response := RecommendModelsResponse{
		Role: p.RoleID,
	}

	if len(scoredModels) > 0 {
		response.RecommendedModel = &scoredModels[0].model

		// Добавляем альтернативы (до 5 моделей)
		if len(scoredModels) > 1 {
			count := min(5, len(scoredModels)-1)
			for i := 1; i <= count; i++ {
				response.Alternatives = append(response.Alternatives, scoredModels[i].model)
			}
		}

		response.Message = fmt.Sprintf("Found %d suitable models for role %s", len(scoredModels), p.RoleID)
	} else {
		response.Message = fmt.Sprintf("No suitable models found for role %s", p.RoleID)
	}

	return response, nil
}

// Регистрируем новые команды
func RegisterTaskMasterIntegrationCommands(server *MCPServer) {
	server.RegisterCommand("taskmaster_export", HandleTaskMasterExport)
	server.RegisterCommand("taskmaster_import", HandleTaskMasterImport)
	server.RegisterCommand("recommend_models", HandleRecommendModels)
}

// Вспомогательные функции

// contains проверяет наличие строки в слайсе
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// min возвращает минимальное из двух целых чисел
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RegisterModelCommands регистрирует команды для работы с моделями в MCP-сервере
func RegisterModelCommands(server *MCPServer) {
	server.RegisterCommand("models_setup", HandleModelSetup)
	server.RegisterCommand("select_model", HandleSelectModel)
	server.RegisterCommand("models_list", HandleModelList)
}

// Обновляем InitMCPServer, чтобы включить наши новые команды:
/*
func InitMCPServer() *MCPServer {
	server := NewMCPServer()

	// Регистрация команд
	RegisterChainProgressCommand(server)
	RegisterModelCommands(server)

	return server
}
*/
