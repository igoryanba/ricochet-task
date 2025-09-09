package mcp

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// CommandHandler представляет обработчик MCP-команды
type CommandHandler func(json.RawMessage) (interface{}, error)

// MCPServer представляет сервер MCP
type MCPServer struct {
	commands map[string]CommandHandler
	mutex    sync.RWMutex
}

// NewMCPServer создает новый экземпляр MCP-сервера
func NewMCPServer() *MCPServer {
	return &MCPServer{
		commands: make(map[string]CommandHandler),
	}
}

// RegisterCommand регистрирует обработчик команды
func (s *MCPServer) RegisterCommand(commandName string, handler CommandHandler) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.commands[commandName] = handler
}

// RegisterTool – обертка для совместимости с интерфейсом Task Master, позволяющая
// регистрировать обработчики, возвращающие json.RawMessage. Внутри конвертируем
// их к CommandHandler.
func (s *MCPServer) RegisterTool(toolName string, handler func(json.RawMessage) (json.RawMessage, error)) {
	wrapped := func(params json.RawMessage) (interface{}, error) {
		return handler(params)
	}
	s.RegisterCommand(toolName, wrapped)
}

// MCPRequest представляет запрос к MCP-серверу
type MCPRequest struct {
	Command string          `json:"command"`
	Params  json.RawMessage `json:"params"`
}

// MCPResponse представляет ответ от MCP-сервера
type MCPResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Command string      `json:"command"`
}

// HandleMCPRequest обрабатывает MCP-запрос
func (s *MCPServer) HandleMCPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MCPRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		sendErrorResponse(w, "invalid_request", fmt.Sprintf("Failed to decode request: %v", err), req.Command)
		return
	}

	s.mutex.RLock()
	handler, exists := s.commands[req.Command]
	s.mutex.RUnlock()

	if !exists {
		sendErrorResponse(w, "unknown_command", fmt.Sprintf("Unknown command: %s", req.Command), req.Command)
		return
	}

	result, err := handler(req.Params)
	if err != nil {
		sendErrorResponse(w, "command_error", fmt.Sprintf("Command execution error: %v", err), req.Command)
		return
	}

	response := MCPResponse{
		Status:  "success",
		Data:    result,
		Command: req.Command,
	}

	sendJSONResponse(w, response)
}

// Start запускает HTTP-сервер для MCP
func (s *MCPServer) Start(address string) error {
	http.HandleFunc("/mcp", s.HandleMCPRequest)
	log.Printf("Starting MCP server on %s", address)
	return http.ListenAndServe(address, nil)
}

// sendErrorResponse отправляет ответ с ошибкой
func sendErrorResponse(w http.ResponseWriter, status string, errorMsg, command string) {
	response := MCPResponse{
		Status:  status,
		Error:   errorMsg,
		Command: command,
	}
	sendJSONResponse(w, response)
}

// sendJSONResponse отправляет JSON-ответ
func sendJSONResponse(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
	}
}

// InitMCPServer инициализирует MCP-сервер со всеми зарегистрированными командами
func InitMCPServer() *MCPServer {
	server := NewMCPServer()

	// Регистрация всех команд через единый метод
	InitializeAllMCPHandlers(server)

	return server
}

// RunMCPServer запускает MCP-сервер на указанном адресе
func RunMCPServer(address string) error {
	server := InitMCPServer()
	return server.Start(address)
}

// InitializeAllMCPHandlers регистрирует все доступные MCP-команды.
// На текущем этапе реализована как заглушка, поскольку каждое расширение
// регистрирует свои команды самостоятельно в тестах.
func InitializeAllMCPHandlers(server *MCPServer) {
	// Централизованная регистрация всех поддерживаемых MCP-команд.

	// 1. Прогресс, мониторинг и визуализация цепочек
	RegisterChainProgressCommand(server)
	RegisterChainMonitorCommands(server)
	RegisterChainVisualizationCommand(server)

	// 2. Управление жизненным циклом цепочек
	RegisterChainCreateCommands(server)
	RegisterChainBuilderCommands(server)
	RegisterChainInteractiveBuilderCommands(server)
	RegisterChainControlCommands(server)

	// 3. Чекпоинты
	RegisterCheckpointCommands(server)

	// 4. Работа с моделями
	RegisterModelCommands(server)

	// 5. Результаты выполнения цепочек
	RegisterChainResultsCommands(server)

	// 6. Интеграция с Task Master (инструменты/хендлеры)
	RegisterTaskMasterIntegrationCommands(server)
	RegisterTaskMasterTools(server)
}

// Пример использования:
/*
func main() {
	if err := RunMCPServer(":8080"); err != nil {
		log.Fatalf("Failed to start MCP server: %v", err)
	}
}
*/
