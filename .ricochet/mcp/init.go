//go:build ricochet_ignore
// +build ricochet_ignore

package mcp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
)

// InitializeAllMCPHandlers регистрирует все MCP-хендлеры в сервере
func InitializeAllMCPHandlers(server *MCPServer) {
	log.Println("Регистрация MCP-хендлеров...")

	// Регистрация хендлеров для мониторинга цепочек
	log.Println("- Регистрация хендлеров мониторинга цепочек")
	RegisterChainProgressCommand(server)
	RegisterChainMonitorCommands(server)
	RegisterChainVisualizationCommand(server)

	// Регистрация хендлеров для управления цепочками
	log.Println("- Регистрация хендлеров управления цепочками")
	RegisterChainCreateCommands(server)
	RegisterChainBuilderCommands(server)
	RegisterChainControlCommands(server)

	// Регистрация хендлеров для интерактивного конструктора цепочек
	log.Println("- Регистрация хендлеров интерактивного конструктора цепочек")
	RegisterChainInteractiveBuilderCommands(server)

	// Регистрация хендлеров для моделей
	log.Println("- Регистрация хендлеров управления моделями")
	RegisterModelCommands(server)

	// Регистрация хендлеров для чекпоинтов
	log.Println("- Регистрация хендлеров управления чекпоинтами")
	RegisterCheckpointCommands(server)

	// Регистрация хендлеров для результатов цепочек
	log.Println("- Регистрация хендлеров результатов цепочек")
	RegisterChainResultsCommands(server)

	// Регистрация команд для интеграции с Task Master
	RegisterTaskMasterIntegrationCommands(server)

	log.Println("Регистрация MCP-хендлеров завершена")
}

// InitMCPServer инициализирует и запускает MCP сервер
func InitMCPServer(port int) *Server {
	server := NewServer()

	// Регистрируем инструменты
	registerChainTools(server)
	registerCheckpointTools(server)
	registerModelTools(server)
	registerTaskMasterTools(server)

	// Запускаем HTTP сервер
	go func() {
		address := fmt.Sprintf(":%d", port)
		log.Printf("Запуск MCP сервера на %s\n", address)
		if err := http.ListenAndServe(address, server); err != nil {
			log.Fatalf("Ошибка запуска MCP сервера: %v", err)
		}
	}()

	return server
}

// ShutdownMCPServer останавливает MCP сервер
func ShutdownMCPServer(server *Server, ctx context.Context) error {
	if server == nil {
		return nil
	}

	log.Println("Остановка MCP сервера...")
	return nil
}

// GetWorkspacePath возвращает путь к рабочей директории
func GetWorkspacePath() (string, error) {
	workspacePath := os.Getenv("RICOCHET_WORKSPACE")
	if workspacePath == "" {
		var err error
		workspacePath, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("не удалось получить текущую директорию: %w", err)
		}
	}
	return workspacePath, nil
}
