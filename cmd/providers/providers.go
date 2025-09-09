package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/grik-ai/ricochet-task/pkg/providers"
	"github.com/grik-ai/ricochet-task/pkg/providers/youtrack"
)

var (
	registry *providers.ProviderRegistry
	logger   *logrus.Logger
)

// ProvidersCmd represents the providers command
var ProvidersCmd = &cobra.Command{
	Use:   "providers",
	Short: "Manage task management providers",
	Long: `Manage task management providers including YouTrack, Jira, Notion, and others.
	
Providers allow ricochet-task to integrate with various task management systems,
enabling unified operations across multiple platforms.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initializeProviders()
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured providers",
	Long: `Display all configured providers with their status, capabilities, and health information.
	
Examples:
  ricochet providers list
  ricochet providers list --enabled-only
  ricochet providers list --output json`,
	RunE: runListProviders,
}

var addCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new provider",
	Long: `Add a new task management provider with the specified configuration.
	
Examples:
  ricochet providers add youtrack-prod --type youtrack --config config.yaml
  ricochet providers add jira-company --type jira --base-url https://company.atlassian.net
  ricochet providers add notion-workspace --type notion --token $NOTION_TOKEN`,
	Args: cobra.ExactArgs(1),
	RunE: runAddProvider,
}

var removeCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a provider",
	Long: `Remove a task management provider and clean up its resources.
	
Examples:
  ricochet providers remove youtrack-dev
  ricochet providers remove old-jira --force`,
	Args: cobra.ExactArgs(1),
	RunE: runRemoveProvider,
}

var enableCmd = &cobra.Command{
	Use:   "enable [name]",
	Short: "Enable a provider",
	Long: `Enable a task management provider that was previously disabled.
	
Examples:
  ricochet providers enable youtrack-prod
  ricochet providers enable jira-company`,
	Args: cobra.ExactArgs(1),
	RunE: runEnableProvider,
}

var disableCmd = &cobra.Command{
	Use:   "disable [name]",
	Short: "Disable a provider",
	Long: `Disable a task management provider without removing its configuration.
	
Examples:
  ricochet providers disable youtrack-dev
  ricochet providers disable old-jira`,
	Args: cobra.ExactArgs(1),
	RunE: runDisableProvider,
}

var healthCmd = &cobra.Command{
	Use:   "health [name]",
	Short: "Check provider health",
	Long: `Check the health status of one or all providers.
	
Examples:
  ricochet providers health
  ricochet providers health youtrack-prod
  ricochet providers health --watch`,
	RunE: runHealthCheck,
}

var defaultCmd = &cobra.Command{
	Use:   "default [name]",
	Short: "Set default provider",
	Long: `Set the default provider for task operations.
	
Examples:
  ricochet providers default youtrack-prod
  ricochet providers default --show`,
	RunE: runSetDefault,
}

func init() {
	// Add subcommands
	ProvidersCmd.AddCommand(listCmd)
	ProvidersCmd.AddCommand(addCmd)
	ProvidersCmd.AddCommand(removeCmd)
	ProvidersCmd.AddCommand(enableCmd)
	ProvidersCmd.AddCommand(disableCmd)
	ProvidersCmd.AddCommand(healthCmd)
	ProvidersCmd.AddCommand(defaultCmd)

	// List command flags
	listCmd.Flags().Bool("enabled-only", false, "Show only enabled providers")
	listCmd.Flags().StringP("output", "o", "table", "Output format: table, json, yaml")

	// Add command flags
	addCmd.Flags().StringP("type", "t", "", "Provider type (youtrack, jira, notion, etc.)")
	addCmd.Flags().StringP("config", "c", "", "Configuration file path")
	addCmd.Flags().String("base-url", "", "Base URL for the provider")
	addCmd.Flags().String("token", "", "Authentication token")
	addCmd.Flags().String("api-key", "", "API key")
	addCmd.Flags().String("username", "", "Username for basic auth")
	addCmd.Flags().String("password", "", "Password for basic auth")
	addCmd.Flags().Bool("enable", true, "Enable the provider after adding")
	addCmd.MarkFlagRequired("type")

	// Remove command flags
	removeCmd.Flags().Bool("force", false, "Force removal without confirmation")

	// Health command flags
	healthCmd.Flags().Bool("watch", false, "Watch health status continuously")
	healthCmd.Flags().Duration("interval", 30*time.Second, "Watch interval")

	// Default command flags
	defaultCmd.Flags().Bool("show", false, "Show current default provider")
}

func initializeProviders() {
	logger = logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration
	config := loadMultiProviderConfig()

	// Initialize registry
	registry = providers.NewProviderRegistry(config, logger)

	// Initialize providers
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := registry.Initialize(ctx); err != nil {
		logger.Fatalf("Failed to initialize providers: %v", err)
	}
}

func runListProviders(cmd *cobra.Command, args []string) error {
	enabledOnly, _ := cmd.Flags().GetBool("enabled-only")
	output, _ := cmd.Flags().GetString("output")

	var providerInfos map[string]*providers.ProviderInfo
	if enabledOnly {
		providerInfos = registry.ListEnabledProviders()
	} else {
		providerInfos = registry.ListProviders()
	}

	switch output {
	case "json":
		return outputJSON(providerInfos)
	case "yaml":
		return outputYAML(providerInfos)
	default:
		return outputTable(providerInfos)
	}
}

func runAddProvider(cmd *cobra.Command, args []string) error {
	name := args[0]
	providerType, _ := cmd.Flags().GetString("type")
	configFile, _ := cmd.Flags().GetString("config")
	baseURL, _ := cmd.Flags().GetString("base-url")
	token, _ := cmd.Flags().GetString("token")
	apiKey, _ := cmd.Flags().GetString("api-key")
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")
	enable, _ := cmd.Flags().GetBool("enable")

	var config *providers.ProviderConfig

	if configFile != "" {
		// Load from config file
		var err error
		config, err = loadProviderConfigFromFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}
	} else {
		// Create config from flags
		config = createProviderConfigFromFlags(name, providerType, baseURL, token, apiKey, username, password, enable)
	}

	// Add provider
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := registry.AddProvider(ctx, name, config); err != nil {
		return fmt.Errorf("failed to add provider: %w", err)
	}

	fmt.Printf("✅ Provider '%s' added successfully\n", name)

	if enable {
		fmt.Printf("✅ Provider '%s' enabled\n", name)
	}

	return nil
}

func runRemoveProvider(cmd *cobra.Command, args []string) error {
	name := args[0]
	force, _ := cmd.Flags().GetBool("force")

	if !force {
		fmt.Printf("Are you sure you want to remove provider '%s'? (y/N): ", name)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Operation cancelled")
			return nil
		}
	}

	if err := registry.RemoveProvider(name); err != nil {
		return fmt.Errorf("failed to remove provider: %w", err)
	}

	fmt.Printf("✅ Provider '%s' removed successfully\n", name)
	return nil
}

func runEnableProvider(cmd *cobra.Command, args []string) error {
	name := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := registry.EnableProvider(ctx, name); err != nil {
		return fmt.Errorf("failed to enable provider: %w", err)
	}

	fmt.Printf("✅ Provider '%s' enabled successfully\n", name)
	return nil
}

func runDisableProvider(cmd *cobra.Command, args []string) error {
	name := args[0]

	if err := registry.DisableProvider(name); err != nil {
		return fmt.Errorf("failed to disable provider: %w", err)
	}

	fmt.Printf("✅ Provider '%s' disabled successfully\n", name)
	return nil
}

func runHealthCheck(cmd *cobra.Command, args []string) error {
	watch, _ := cmd.Flags().GetBool("watch")
	interval, _ := cmd.Flags().GetDuration("interval")

	if len(args) > 0 {
		// Check specific provider
		return checkProviderHealth(args[0], watch, interval)
	}

	// Check all providers
	return checkAllProvidersHealth(watch, interval)
}

func runSetDefault(cmd *cobra.Command, args []string) error {
	show, _ := cmd.Flags().GetBool("show")

	if show {
		provider, err := registry.GetDefaultProvider()
		if err != nil {
			fmt.Println("No default provider set")
			return nil
		}
		
		info := provider.GetProviderInfo()
		fmt.Printf("Default provider: %s\n", info.Name)
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("provider name is required")
	}

	name := args[0]
	if err := registry.SetDefaultProvider(name); err != nil {
		return fmt.Errorf("failed to set default provider: %w", err)
	}

	fmt.Printf("✅ Default provider set to '%s'\n", name)
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
		}
	}

	return config
}

func loadProviderConfigFromFile(filename string) (*providers.ProviderConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config providers.ProviderConfig
	if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
		err = yaml.Unmarshal(data, &config)
	} else {
		err = json.Unmarshal(data, &config)
	}

	return &config, err
}

func createProviderConfigFromFlags(name, providerType, baseURL, token, apiKey, username, password string, enable bool) *providers.ProviderConfig {
	var config *providers.ProviderConfig

	switch providers.ProviderType(providerType) {
	case providers.ProviderTypeYouTrack:
		config = youtrack.GetDefaultConfig()
	default:
		config = providers.DefaultProviderConfig()
		config.Type = providers.ProviderType(providerType)
	}

	config.Name = name
	config.Enabled = enable

	if baseURL != "" {
		config.BaseURL = baseURL
	}

	// Set authentication
	if token != "" {
		config.AuthType = providers.AuthTypeBearer
		config.Token = token
	} else if apiKey != "" {
		config.AuthType = providers.AuthTypeAPIKey
		config.APIKey = apiKey
	} else if username != "" && password != "" {
		config.AuthType = providers.AuthTypeBasic
		config.Username = username
		config.Password = password
	}

	return config
}

func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func outputYAML(data interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(data)
}

func outputTable(providerInfos map[string]*providers.ProviderInfo) error {
	fmt.Printf("%-20s %-12s %-10s %-15s %-30s\n", "NAME", "TYPE", "STATUS", "HEALTH", "CAPABILITIES")
	fmt.Printf("%-20s %-12s %-10s %-15s %-30s\n", "----", "----", "------", "------", "------------")

	for name, info := range providerInfos {
		capabilities := strings.Join(getCapabilityNames(info.Capabilities), ", ")
		if len(capabilities) > 25 {
			capabilities = capabilities[:25] + "..."
		}

		fmt.Printf("%-20s %-12s %-10s %-15s %-30s\n",
			name,
			string(getProviderType(info.Name)),
			"enabled", // We'd need to track this from registry
			string(info.HealthStatus),
			capabilities,
		)
	}

	return nil
}

func getCapabilityNames(capabilities []providers.Capability) []string {
	names := make([]string, len(capabilities))
	for i, cap := range capabilities {
		names[i] = string(cap)
	}
	return names
}

func getProviderType(name string) providers.ProviderType {
	// This is a simplified mapping - in practice you'd get this from the registry
	if strings.Contains(strings.ToLower(name), "youtrack") {
		return providers.ProviderTypeYouTrack
	}
	if strings.Contains(strings.ToLower(name), "jira") {
		return providers.ProviderTypeJira
	}
	if strings.Contains(strings.ToLower(name), "notion") {
		return providers.ProviderTypeNotion
	}
	return providers.ProviderTypeCustom
}

func checkProviderHealth(name string, watch bool, interval time.Duration) error {
	for {
		provider, err := registry.GetProvider(name)
		if err != nil {
			return fmt.Errorf("provider not found: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err = provider.HealthCheck(ctx)
		cancel()

		status := "🟢 HEALTHY"
		if err != nil {
			status = fmt.Sprintf("🔴 UNHEALTHY: %v", err)
		}

		fmt.Printf("[%s] %s: %s\n", time.Now().Format("15:04:05"), name, status)

		if !watch {
			break
		}

		time.Sleep(interval)
	}

	return nil
}

func checkAllProvidersHealth(watch bool, interval time.Duration) error {
	for {
		healthStatus := registry.GetHealthStatus()

		fmt.Printf("\n[%s] Provider Health Status:\n", time.Now().Format("15:04:05"))
		fmt.Printf("%-20s %-15s\n", "PROVIDER", "STATUS")
		fmt.Printf("%-20s %-15s\n", "--------", "------")

		for name, status := range healthStatus {
			emoji := "🟢"
			if status != providers.HealthStatusHealthy {
				emoji = "🔴"
			}

			fmt.Printf("%-20s %s %-15s\n", name, emoji, string(status))
		}

		if !watch {
			break
		}

		time.Sleep(interval)
	}

	return nil
}