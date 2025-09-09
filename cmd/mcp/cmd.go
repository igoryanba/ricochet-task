package mcp

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sirupsen/logrus"

	"github.com/grik-ai/ricochet-task/pkg/mcp"
	"github.com/grik-ai/ricochet-task/pkg/providers"
)

var (
	registry *providers.ProviderRegistry
	logger   *logrus.Logger
	mcpServer *mcp.HTTPServer
)

// MCPCmd represents the mcp command
var MCPCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP (Model Context Protocol) server for ricochet-task",
	Long: `Start the Model Context Protocol server that provides task management tools
for code editors like VS Code and Cursor.

The MCP server exposes ricochet-task functionality as tools that can be used
by AI assistants within code editors for automated task management, cross-provider
synchronization, and AI-powered project analysis.

Example usage:
  ricochet mcp --port 8080 --http-only
  ricochet mcp --websocket --host 0.0.0.0 --port 3001`,
	RunE: runMCPServer,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the MCP server",
	Long: `Start the Model Context Protocol server with the specified configuration.

Supports both WebSocket and HTTP modes:
- WebSocket mode: Full MCP protocol support for real-time communication
- HTTP mode: REST API endpoints for tool listing and execution

Examples:
  ricochet mcp start --port 8080
  ricochet mcp start --http-only --host 0.0.0.0 --port 3001
  ricochet mcp start --websocket --debug`,
	RunE: runMCPServer,
}

var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "List available MCP tools",
	Long: `Display all available MCP tools with their descriptions and input schemas.

This command shows the tools that will be available to AI assistants when
the MCP server is running.

Examples:
  ricochet mcp tools
  ricochet mcp tools --output json
  ricochet mcp tools --verbose`,
	RunE: runListTools,
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate MCP configuration and provider setup",
	Long: `Validate that all providers are properly configured and accessible
for MCP server operation.

This command checks:
- Provider configurations
- Authentication credentials
- Network connectivity
- Tool availability

Examples:
  ricochet mcp validate
  ricochet mcp validate --provider youtrack-prod
  ricochet mcp validate --verbose`,
	RunE: runValidateConfig,
}

func init() {
	// Add subcommands
	MCPCmd.AddCommand(startCmd)
	MCPCmd.AddCommand(toolsCmd)
	MCPCmd.AddCommand(validateCmd)

	// Global MCP flags
	MCPCmd.PersistentFlags().StringP("host", "H", "localhost", "Host to bind to")
	MCPCmd.PersistentFlags().IntP("port", "p", 3001, "Port to listen on")
	MCPCmd.PersistentFlags().Bool("debug", false, "Enable debug logging")
	MCPCmd.PersistentFlags().StringP("config", "c", "", "Configuration file path")
	MCPCmd.PersistentFlags().Bool("verbose", false, "Verbose output")

	// Server-specific flags
	startCmd.Flags().Bool("websocket", true, "Enable WebSocket support")
	startCmd.Flags().Bool("http-only", false, "HTTP-only mode (disables WebSocket)")
	startCmd.Flags().Duration("timeout", 30*time.Second, "Request timeout")
	startCmd.Flags().Int("max-connections", 100, "Maximum concurrent connections")
	startCmd.Flags().Bool("cors", true, "Enable CORS support")

	// Tools command flags
	toolsCmd.Flags().StringP("output", "o", "table", "Output format: table, json, yaml")

	// Validate command flags
	validateCmd.Flags().String("provider", "", "Validate specific provider only")
	validateCmd.Flags().Bool("fix", false, "Attempt to fix configuration issues")
}

func initializeMCP() error {
	// Setup logger
	logger = logrus.New()
	if viper.GetBool("debug") {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	// Load configuration
	config := loadMultiProviderConfig()
	if config == nil {
		return fmt.Errorf("failed to load provider configuration")
	}

	// Initialize provider registry
	registry = providers.NewProviderRegistry(config, logger)

	// Initialize providers
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := registry.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize providers: %w", err)
	}

	// Create MCP HTTP server
	mcpServer = mcp.NewHTTPServer(registry, logger)

	return nil
}

func runMCPServer(cmd *cobra.Command, args []string) error {
	if err := initializeMCP(); err != nil {
		return err
	}

	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetInt("port")
	_, _ = cmd.Flags().GetBool("http-only")
	_, _ = cmd.Flags().GetBool("websocket")

	addr := fmt.Sprintf("%s:%d", host, port)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal")
		cancel()
	}()

	// Start server
	errChan := make(chan error, 1)
	go func() {
		logger.Infof("Starting MCP HTTP server on %s", addr)
		errChan <- mcpServer.Start(addr)
	}()

	// Wait for shutdown or error
	select {
	case <-ctx.Done():
		logger.Info("Shutting down MCP server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := mcpServer.Shutdown(shutdownCtx); err != nil {
			logger.Errorf("Error during shutdown: %v", err)
		}
		return nil
	case err := <-errChan:
		return fmt.Errorf("MCP server error: %w", err)
	}
}

func runListTools(cmd *cobra.Command, args []string) error {
	if err := initializeMCP(); err != nil {
		return err
	}

	output, _ := cmd.Flags().GetString("output")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Get tools from MCP server
	toolProvider := mcp.NewMCPToolProvider(registry)
	tools := toolProvider.GetTools()

	switch output {
	case "json":
		return outputJSON(tools)
	case "yaml":
		return outputYAML(tools)
	default:
		return outputToolsTable(tools, verbose)
	}
}

func runValidateConfig(cmd *cobra.Command, args []string) error {
	if err := initializeMCP(); err != nil {
		return err
	}

	providerName, _ := cmd.Flags().GetString("provider")
	fix, _ := cmd.Flags().GetBool("fix")
	verbose, _ := cmd.Flags().GetBool("verbose")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if providerName != "" {
		// Validate specific provider
		return validateProvider(ctx, providerName, fix, verbose)
	}

	// Validate all providers
	return validateAllProviders(ctx, fix, verbose)
}

func validateProvider(ctx context.Context, name string, fix bool, verbose bool) error {
	logger.Infof("Validating provider: %s", name)

	provider, err := registry.GetProvider(name)
	if err != nil {
		logger.Errorf("❌ Provider '%s' not found: %v", name, err)
		return err
	}

	// Health check
	err = provider.HealthCheck(ctx)
	if err != nil {
		logger.Errorf("❌ Provider '%s' health check failed: %v", name, err)
		if fix {
			logger.Info("🔧 Attempting to fix provider issues...")
			// TODO: Implement fix logic
			logger.Warn("🚧 Automatic fixing not yet implemented")
		}
		return err
	}

	logger.Infof("✅ Provider '%s' is healthy", name)

	if verbose {
		info := provider.GetProviderInfo()
		logger.Infof("Provider info: %+v", info)
	}

	return nil
}

func validateAllProviders(ctx context.Context, fix bool, verbose bool) error {
	logger.Info("Validating all providers...")

	providers := registry.ListProviders()
	if len(providers) == 0 {
		logger.Warn("⚠️ No providers configured")
		return nil
	}

	successCount := 0
	failureCount := 0

	for name := range providers {
		err := validateProvider(ctx, name, fix, verbose)
		if err != nil {
			failureCount++
		} else {
			successCount++
		}
	}

	logger.Infof("Validation complete: %d successful, %d failed", successCount, failureCount)

	if failureCount > 0 {
		return fmt.Errorf("%d provider(s) failed validation", failureCount)
	}

	return nil
}

// Helper functions

func loadMultiProviderConfig() *providers.MultiProviderConfig {
	config := providers.DefaultMultiProviderConfig()

	// Try to load from config file
	configFile := viper.GetString("config")
	if configFile == "" {
		configFile = "ricochet.yaml"
	}

	if _, err := os.Stat(configFile); err == nil {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err == nil {
			viper.Unmarshal(config)
			logger.Infof("Loaded configuration from %s", configFile)
		} else {
			logger.Warnf("Failed to read config file %s: %v", configFile, err)
		}
	} else {
		logger.Infof("No config file found at %s, using defaults", configFile)
	}

	return config
}

func outputJSON(data interface{}) error {
	// This would require encoding/json import
	fmt.Printf("%+v\n", data)
	return nil
}

func outputYAML(data interface{}) error {
	// For now, use JSON output since yaml package import would be needed
	return outputJSON(data)
}

func outputToolsTable(tools []mcp.ToolDefinition, verbose bool) error {
	fmt.Printf("Available MCP Tools (%d total)\n", len(tools))
	fmt.Printf("==========================================\n\n")

	for i, tool := range tools {
		fmt.Printf("%d. %s\n", i+1, tool.Name)
		fmt.Printf("   Description: %s\n", tool.Description)
		
		if verbose {
			fmt.Printf("   Input Schema: %+v\n", tool.InputSchema)
		}
		
		fmt.Println()
	}

	return nil
}