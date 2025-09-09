package mcp

// Пакет mcp выступает публичным фасадом для MCP функционала.
// Вся основная реализация находится в под-директории `.ricochet/mcp`.
// Данный файл переэкспортирует ключевые типы и функции, чтобы
// существующий код (а также интеграционные тесты) мог работать,
// импортируя `github.com/grik-ai/ricochet-task/pkg/mcp`.

import (
	internal "github.com/grik-ai/ricochet-task/.ricochet/mcp"
)

// ------------------------------------------------------------------
// Типы
// ------------------------------------------------------------------

type MCPServer = internal.MCPServer

type MCPRequest = internal.MCPRequest

type MCPResponse = internal.MCPResponse

type CommandHandler = internal.CommandHandler

type ChainCreateParams = internal.ChainCreateParams
type ChainStep = internal.ChainStep

type ChainBuilderInitParams = internal.ChainBuilderInitParams
type ChainBuilderStepParams = internal.ChainBuilderStepParams
type ChainBuilderResponse = internal.ChainBuilderResponse

type SessionCompleteParams = internal.SessionCompleteParams
type AutoSelectModelsParams = internal.AutoSelectModelsParams
type AutoSelectModelsResponse = internal.AutoSelectModelsResponse
type AutoSelectedStepInfo = internal.AutoSelectedStepInfo

// ------------------------------------------------------------------
// Конструкторы
// ------------------------------------------------------------------

func NewMCPServer() *MCPServer { return internal.NewMCPServer() }

// ------------------------------------------------------------------
// Переэкспорт функций регистрации команд
// ------------------------------------------------------------------

func RegisterChainCreateCommands(s *MCPServer)  { internal.RegisterChainCreateCommands(s) }
func RegisterChainBuilderCommands(s *MCPServer) { internal.RegisterChainBuilderCommands(s) }
func RegisterChainInteractiveBuilderCommands(s *MCPServer) {
	internal.RegisterChainInteractiveBuilderCommands(s)
}
func RegisterChainControlCommands(s *MCPServer) { internal.RegisterChainControlCommands(s) }
func RegisterModelCommands(s *MCPServer)        { internal.RegisterModelCommands(s) }
func RegisterChainProgressCommand(s *MCPServer) { internal.RegisterChainProgressCommand(s) }
