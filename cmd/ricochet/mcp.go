package ricochet

import (
	"fmt"

	mcp "github.com/grik-ai/ricochet-task/pkg/mcpserver"
	"github.com/spf13/cobra"
)

// mcpCmd запускает локальный MCP-сервер (Model Chain Protocol)
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Запустить MCP сервер (Model Chain Protocol)",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		addr := fmt.Sprintf(":%d", port)
		fmt.Printf("📡 Запуск MCP сервера на %s\n", addr)
		return mcp.RunMCPServer(addr)
	},
}

func init() {
	// Флаг --port/-p для выбора порта
	mcpCmd.Flags().IntP("port", "p", 8090, "Порт для MCP сервера")
	// Регистрируем подкоманду в корневом CLI
	rootCmd.AddCommand(mcpCmd)
}
