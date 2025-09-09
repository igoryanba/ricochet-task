package youtrack

import (
	"fmt"

	"github.com/grik-ai/ricochet-task/pkg/providers"
)

// YouTrackPlugin implements the TaskManagerPlugin interface for YouTrack
type YouTrackPlugin struct {
	provider *YouTrackProvider
	config   *providers.ProviderConfig
}

// NewYouTrackPlugin creates a new YouTrack plugin instance
func NewYouTrackPlugin() providers.TaskManagerPlugin {
	return &YouTrackPlugin{}
}

// Name returns the plugin name
func (p *YouTrackPlugin) Name() string {
	return "youtrack"
}

// Version returns the plugin version
func (p *YouTrackPlugin) Version() string {
	return "1.0.0"
}

// Description returns the plugin description
func (p *YouTrackPlugin) Description() string {
	return "JetBrains YouTrack integration for ricochet-task providing comprehensive issue tracking, project management, and agile board capabilities"
}

// Initialize initializes the plugin with the provided configuration
func (p *YouTrackPlugin) Initialize(config *providers.ProviderConfig) error {
	if config == nil {
		return fmt.Errorf("configuration is required")
	}

	// Validate YouTrack-specific configuration
	if err := p.validateConfig(config); err != nil {
		return fmt.Errorf("invalid YouTrack configuration: %w", err)
	}

	// Create YouTrack provider
	provider, err := NewYouTrackProvider(config)
	if err != nil {
		return fmt.Errorf("failed to create YouTrack provider: %w", err)
	}

	p.provider = provider
	p.config = config

	return nil
}

// GetProvider returns the TaskProvider interface
func (p *YouTrackPlugin) GetProvider() providers.TaskProvider {
	return p.provider
}

// GetBoardProvider returns the BoardProvider interface if supported
func (p *YouTrackPlugin) GetBoardProvider() providers.BoardProvider {
	if p.provider == nil {
		return nil
	}
	return NewYouTrackBoardProvider(p.provider.client, p.config)
}

// GetSyncProvider returns the SyncProvider interface if supported
func (p *YouTrackPlugin) GetSyncProvider() providers.SyncProvider {
	// YouTrack supports webhooks for real-time sync, but we'll implement this later
	// For now, return nil to indicate it's not implemented
	return nil
}

// GetSearchProvider returns the SearchProvider interface if supported
func (p *YouTrackPlugin) GetSearchProvider() providers.SearchProvider {
	// YouTrack has powerful search capabilities, but we'll implement this later
	// For now, return nil to indicate it's not implemented
	return nil
}

// GetAnalyticsProvider returns the AnalyticsProvider interface if supported
func (p *YouTrackPlugin) GetAnalyticsProvider() providers.AnalyticsProvider {
	// YouTrack has reporting capabilities, but we'll implement this later
	// For now, return nil to indicate it's not implemented
	return nil
}

// Cleanup cleans up plugin resources
func (p *YouTrackPlugin) Cleanup() error {
	if p.provider != nil {
		return p.provider.Close()
	}
	return nil
}

// validateConfig validates YouTrack-specific configuration
func (p *YouTrackPlugin) validateConfig(config *providers.ProviderConfig) error {
	// Check provider type
	if config.Type != providers.ProviderTypeYouTrack {
		return fmt.Errorf("invalid provider type: expected %s, got %s", providers.ProviderTypeYouTrack, config.Type)
	}

	// Check required fields
	if config.BaseURL == "" {
		return fmt.Errorf("baseUrl is required for YouTrack provider")
	}

	// Check authentication
	switch config.AuthType {
	case providers.AuthTypeBearer:
		if config.Token == "" {
			return fmt.Errorf("token is required for bearer authentication")
		}
	case providers.AuthTypeAPIKey:
		if config.APIKey == "" {
			return fmt.Errorf("apiKey is required for API key authentication")
		}
	default:
		return fmt.Errorf("unsupported authentication type for YouTrack: %s", config.AuthType)
	}

	// Validate YouTrack-specific settings
	if config.Settings != nil {
		if err := p.validateYouTrackSettings(config.Settings); err != nil {
			return fmt.Errorf("invalid YouTrack settings: %w", err)
		}
	}

	return nil
}

// validateYouTrackSettings validates YouTrack-specific settings
func (p *YouTrackPlugin) validateYouTrackSettings(settings map[string]interface{}) error {
	// Validate default project if specified and not empty
	if defaultProject, exists := settings["defaultProject"]; exists {
		if _, ok := defaultProject.(string); ok {
			// Empty string is allowed - provider can work without default project
		} else {
			return fmt.Errorf("defaultProject must be a string")
		}
	}

	// Validate default board if specified
	if defaultBoard, exists := settings["defaultBoard"]; exists {
		if _, ok := defaultBoard.(string); ok {
			// Empty string is allowed - provider can work without default board
		} else {
			return fmt.Errorf("defaultBoard must be a string")
		}
	}

	// Validate auto create boards setting
	if autoCreateBoards, exists := settings["autoCreateBoards"]; exists {
		if _, ok := autoCreateBoards.(bool); !ok {
			return fmt.Errorf("autoCreateBoards must be a boolean")
		}
	}

	// Validate custom field mappings if specified
	if customFieldMappings, exists := settings["customFieldMappings"]; exists {
		if mappings, ok := customFieldMappings.(map[string]interface{}); ok {
			for key, value := range mappings {
				if key == "" {
					return fmt.Errorf("custom field mapping key cannot be empty")
				}
				if valueStr, ok := value.(string); !ok || valueStr == "" {
					return fmt.Errorf("custom field mapping value for key '%s' must be a non-empty string", key)
				}
			}
		} else {
			return fmt.Errorf("customFieldMappings must be a map")
		}
	}

	// Validate workflow mappings if specified
	if workflowMappings, exists := settings["workflowMappings"]; exists {
		if mappings, ok := workflowMappings.(map[string]interface{}); ok {
			for key, value := range mappings {
				if key == "" {
					return fmt.Errorf("workflow mapping key cannot be empty")
				}
				if valueStr, ok := value.(string); !ok || valueStr == "" {
					return fmt.Errorf("workflow mapping value for key '%s' must be a non-empty string", key)
				}
			}
		} else {
			return fmt.Errorf("workflowMappings must be a map")
		}
	}

	return nil
}

// GetDefaultConfig returns default configuration for YouTrack
func GetDefaultConfig() *providers.ProviderConfig {
	config := providers.DefaultProviderConfig()
	config.Type = providers.ProviderTypeYouTrack
	config.AuthType = providers.AuthTypeBearer
	
	// YouTrack-specific settings
	config.Settings = map[string]interface{}{
		"defaultProject":     "",
		"defaultBoard":       "",
		"autoCreateBoards":   false,
		"useShortNames":      true,
		"syncComments":       true,
		"syncAttachments":    true,
		"syncTimeTracking":   true,
		"syncCustomFields":   true,
		"customFieldMappings": map[string]interface{}{
			// Map universal field names to YouTrack field names
			"story_points": "Story Points",
			"sprint":       "Sprint",
			"epic":         "Epic",
		},
		"workflowMappings": map[string]interface{}{
			// Map universal status categories to YouTrack states
			"todo":        "Open",
			"in_progress": "In Progress", 
			"done":        "Fixed",
			"blocked":     "Blocked",
		},
	}

	// YouTrack-specific rate limits (YouTrack allows quite generous limits)
	config.RateLimit.RequestsPerSecond = 10
	config.RateLimit.BurstSize = 50

	// YouTrack-specific retry configuration
	config.RetryConfig.MaxRetries = 3
	config.RetryConfig.RetryableErrors = []string{
		"429", // Too Many Requests
		"500", // Internal Server Error
		"502", // Bad Gateway
		"503", // Service Unavailable
		"504", // Gateway Timeout
	}

	return config
}

// GetCapabilities returns the capabilities of the YouTrack provider
func GetCapabilities() []providers.Capability {
	return []providers.Capability{
		providers.CapabilityTasks,
		providers.CapabilityBoards,
		providers.CapabilityRealTimeSync,
		providers.CapabilityCustomFields,
		providers.CapabilityWorkflows,
		providers.CapabilityTimeTracking,
		providers.CapabilityHierarchicalTasks,
		providers.CapabilityReporting,
		providers.CapabilityAdvancedSearch,
		providers.CapabilityWebhooks,
	}
}

// GetSupportedFeatures returns the features supported by YouTrack
func GetSupportedFeatures() map[string]bool {
	return map[string]bool{
		"hierarchical_tasks":  true,
		"custom_fields":       true,
		"time_tracking":       true,
		"agile_boards":        true,
		"workflows":           true,
		"webhooks":            true,
		"search_queries":      true,
		"bulk_operations":     true,
		"comments":            true,
		"attachments":         true,
		"tags":                true,
		"issue_links":         true,
		"sprints":             true,
		"backlogs":            true,
		"burndown_charts":     true,
		"reports":             true,
		"user_management":     true,
		"project_management":  true,
		"version_management":  true,
		"component_management": true,
	}
}

// Plugin factory function for registration
func init() {
	// Register the plugin factory
	providers.RegisterPluginFactory(string(providers.ProviderTypeYouTrack), NewYouTrackPlugin)
}