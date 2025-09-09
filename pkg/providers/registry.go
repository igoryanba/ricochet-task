package providers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ProviderRegistry manages all available task management providers
type ProviderRegistry struct {
	mu               sync.RWMutex
	providers        map[string]TaskProvider
	plugins          map[string]TaskManagerPlugin
	config           *MultiProviderConfig
	healthCheckers   map[string]*HealthChecker
	logger           *logrus.Logger
	defaultProvider  string
}

// PluginFactory is a function that creates a new plugin instance
type PluginFactory func() TaskManagerPlugin

var globalPluginFactories = make(map[string]PluginFactory)

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry(config *MultiProviderConfig, logger *logrus.Logger) *ProviderRegistry {
	if logger == nil {
		logger = logrus.New()
	}

	registry := &ProviderRegistry{
		providers:      make(map[string]TaskProvider),
		plugins:        make(map[string]TaskManagerPlugin),
		config:         config,
		healthCheckers: make(map[string]*HealthChecker),
		logger:         logger,
		defaultProvider: config.DefaultProvider,
	}

	return registry
}

// RegisterPluginFactory registers a plugin factory globally
func RegisterPluginFactory(providerType string, factory PluginFactory) {
	globalPluginFactories[providerType] = factory
}

// Initialize initializes all configured providers
func (r *ProviderRegistry) Initialize(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logger.Info("Initializing provider registry")

	// Initialize each configured provider
	for name, providerConfig := range r.config.Providers {
		if !providerConfig.Enabled {
			r.logger.Infof("Skipping disabled provider: %s", name)
			continue
		}

		if err := r.initializeProvider(ctx, name, providerConfig); err != nil {
			r.logger.Errorf("Failed to initialize provider %s: %v", name, err)
			return fmt.Errorf("failed to initialize provider %s: %w", name, err)
		}

		r.logger.Infof("Successfully initialized provider: %s", name)
	}

	// Start health checking
	r.startHealthChecking(ctx)

	r.logger.Infof("Provider registry initialized with %d providers", len(r.providers))
	return nil
}

// initializeProvider initializes a single provider
func (r *ProviderRegistry) initializeProvider(ctx context.Context, name string, config *ProviderConfig) error {
	// Get plugin factory
	factory, exists := globalPluginFactories[string(config.Type)]
	if !exists {
		return fmt.Errorf("no plugin factory registered for provider type: %s", config.Type)
	}

	// Create plugin instance
	plugin := factory()
	
	// Initialize plugin
	if err := plugin.Initialize(config); err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// Get provider interface
	provider := plugin.GetProvider()
	if provider == nil {
		return fmt.Errorf("plugin returned nil provider")
	}

	// Test provider health
	if err := provider.HealthCheck(ctx); err != nil {
		r.logger.Warnf("Provider %s failed initial health check: %v", name, err)
	}

	// Store provider and plugin
	r.providers[name] = provider
	r.plugins[name] = plugin

	// Create health checker
	r.healthCheckers[name] = NewHealthChecker(provider, r.config.HealthCheck, r.logger)

	return nil
}

// GetProvider returns a provider by name
func (r *ProviderRegistry) GetProvider(name string) (TaskProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", name)
	}

	return provider, nil
}

// GetDefaultProvider returns the default provider
func (r *ProviderRegistry) GetDefaultProvider() (TaskProvider, error) {
	if r.defaultProvider == "" {
		return nil, fmt.Errorf("no default provider configured")
	}

	return r.GetProvider(r.defaultProvider)
}

// ListProviders returns all available providers
func (r *ProviderRegistry) ListProviders() map[string]*ProviderInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info := make(map[string]*ProviderInfo)
	for name, provider := range r.providers {
		info[name] = provider.GetProviderInfo()
	}

	return info
}

// ListEnabledProviders returns only enabled providers
func (r *ProviderRegistry) ListEnabledProviders() map[string]*ProviderInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info := make(map[string]*ProviderInfo)
	for name, provider := range r.providers {
		config := r.config.Providers[name]
		if config != nil && config.Enabled {
			info[name] = provider.GetProviderInfo()
		}
	}

	return info
}

// HasCapability checks if any provider has a specific capability
func (r *ProviderRegistry) HasCapability(capability Capability) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, provider := range r.providers {
		info := provider.GetProviderInfo()
		for _, cap := range info.Capabilities {
			if cap == capability {
				return true
			}
		}
	}

	return false
}

// GetProvidersWithCapability returns providers that have a specific capability
func (r *ProviderRegistry) GetProvidersWithCapability(capability Capability) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var providers []string
	for name, provider := range r.providers {
		info := provider.GetProviderInfo()
		for _, cap := range info.Capabilities {
			if cap == capability {
				providers = append(providers, name)
				break
			}
		}
	}

	return providers
}

// EnableProvider enables a provider
func (r *ProviderRegistry) EnableProvider(ctx context.Context, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	config, exists := r.config.Providers[name]
	if !exists {
		return fmt.Errorf("provider config not found: %s", name)
	}

	if config.Enabled {
		return nil // Already enabled
	}

	// Enable in config
	config.Enabled = true

	// Initialize if not already initialized
	if _, exists := r.providers[name]; !exists {
		if err := r.initializeProvider(ctx, name, config); err != nil {
			config.Enabled = false // Rollback
			return fmt.Errorf("failed to initialize provider: %w", err)
		}
	}

	r.logger.Infof("Provider %s enabled", name)
	return nil
}

// DisableProvider disables a provider
func (r *ProviderRegistry) DisableProvider(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	config, exists := r.config.Providers[name]
	if !exists {
		return fmt.Errorf("provider config not found: %s", name)
	}

	if !config.Enabled {
		return nil // Already disabled
	}

	// Disable in config
	config.Enabled = false

	// Stop health checker
	if checker, exists := r.healthCheckers[name]; exists {
		checker.Stop()
		delete(r.healthCheckers, name)
	}

	// Clean up provider
	if plugin, exists := r.plugins[name]; exists {
		if err := plugin.Cleanup(); err != nil {
			r.logger.Errorf("Error cleaning up plugin %s: %v", name, err)
		}
		delete(r.plugins, name)
	}

	delete(r.providers, name)

	r.logger.Infof("Provider %s disabled", name)
	return nil
}

// AddProvider adds a new provider configuration
func (r *ProviderRegistry) AddProvider(ctx context.Context, name string, config *ProviderConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate config
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid provider config: %w", err)
	}

	// Check if already exists
	if _, exists := r.config.Providers[name]; exists {
		return fmt.Errorf("provider already exists: %s", name)
	}

	// Add to config
	r.config.Providers[name] = config

	// Initialize if enabled
	if config.Enabled {
		if err := r.initializeProvider(ctx, name, config); err != nil {
			delete(r.config.Providers, name) // Rollback
			return fmt.Errorf("failed to initialize provider: %w", err)
		}
	}

	r.logger.Infof("Provider %s added", name)
	return nil
}

// RemoveProvider removes a provider
func (r *ProviderRegistry) RemoveProvider(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Disable first
	if err := r.DisableProvider(name); err != nil {
		return err
	}

	// Remove from config
	delete(r.config.Providers, name)

	// Update default provider if this was it
	if r.defaultProvider == name {
		r.defaultProvider = ""
		// Set first available provider as default
		for providerName := range r.config.Providers {
			r.defaultProvider = providerName
			break
		}
	}

	r.logger.Infof("Provider %s removed", name)
	return nil
}

// SetDefaultProvider sets the default provider
func (r *ProviderRegistry) SetDefaultProvider(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if provider exists
	if _, exists := r.config.Providers[name]; !exists {
		return fmt.Errorf("provider not found: %s", name)
	}

	r.defaultProvider = name
	r.config.DefaultProvider = name

	r.logger.Infof("Default provider set to %s", name)
	return nil
}

// GetHealthStatus returns health status for all providers
func (r *ProviderRegistry) GetHealthStatus() map[string]ProviderHealthStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	status := make(map[string]ProviderHealthStatus)
	for name, provider := range r.providers {
		info := provider.GetProviderInfo()
		status[name] = info.HealthStatus
	}

	return status
}

// startHealthChecking starts health checking for all providers
func (r *ProviderRegistry) startHealthChecking(ctx context.Context) {
	for name, checker := range r.healthCheckers {
		go func(name string, checker *HealthChecker) {
			checker.Start(ctx)
		}(name, checker)
	}
}

// Shutdown gracefully shuts down all providers
func (r *ProviderRegistry) Shutdown(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logger.Info("Shutting down provider registry")

	var lastError error

	// Stop health checkers
	for name, checker := range r.healthCheckers {
		checker.Stop()
		r.logger.Debugf("Stopped health checker for %s", name)
	}

	// Cleanup all plugins
	for name, plugin := range r.plugins {
		if err := plugin.Cleanup(); err != nil {
			r.logger.Errorf("Error cleaning up plugin %s: %v", name, err)
			lastError = err
		} else {
			r.logger.Debugf("Cleaned up plugin %s", name)
		}
	}

	// Close all providers
	for name, provider := range r.providers {
		if err := provider.Close(); err != nil {
			r.logger.Errorf("Error closing provider %s: %v", name, err)
			lastError = err
		} else {
			r.logger.Debugf("Closed provider %s", name)
		}
	}

	r.logger.Info("Provider registry shutdown complete")
	return lastError
}

// HealthChecker manages health checking for a provider
type HealthChecker struct {
	provider TaskProvider
	interval time.Duration
	logger   *logrus.Logger
	stopCh   chan struct{}
	stopped  bool
	mu       sync.Mutex
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(provider TaskProvider, interval time.Duration, logger *logrus.Logger) *HealthChecker {
	return &HealthChecker{
		provider: provider,
		interval: interval,
		logger:   logger,
		stopCh:   make(chan struct{}),
	}
}

// Start starts the health checking
func (h *HealthChecker) Start(ctx context.Context) {
	h.mu.Lock()
	if h.stopped {
		h.mu.Unlock()
		return
	}
	h.mu.Unlock()

	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopCh:
			return
		case <-ticker.C:
			h.performHealthCheck(ctx)
		}
	}
}

// Stop stops the health checking
func (h *HealthChecker) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.stopped {
		close(h.stopCh)
		h.stopped = true
	}
}

// performHealthCheck performs a single health check
func (h *HealthChecker) performHealthCheck(ctx context.Context) {
	checkCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	providerInfo := h.provider.GetProviderInfo()
	oldStatus := providerInfo.HealthStatus

	err := h.provider.HealthCheck(checkCtx)
	
	var newStatus ProviderHealthStatus
	if err != nil {
		newStatus = HealthStatusUnhealthy
		h.logger.Warnf("Health check failed for provider %s: %v", providerInfo.Name, err)
	} else {
		newStatus = HealthStatusHealthy
	}

	// Update provider info (this would need to be implemented in the provider)
	if oldStatus != newStatus {
		h.logger.Infof("Provider %s health status changed from %s to %s", 
			providerInfo.Name, oldStatus, newStatus)
	}
}

// Utility functions for registry operations
func (r *ProviderRegistry) ExecuteOnProvider(name string, operation func(TaskProvider) error) error {
	provider, err := r.GetProvider(name)
	if err != nil {
		return err
	}
	
	return operation(provider)
}

func (r *ProviderRegistry) ExecuteOnAllProviders(operation func(string, TaskProvider) error) map[string]error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]error)
	for name, provider := range r.providers {
		results[name] = operation(name, provider)
	}

	return results
}

func (r *ProviderRegistry) ExecuteOnEnabledProviders(operation func(string, TaskProvider) error) map[string]error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]error)
	for name, provider := range r.providers {
		config := r.config.Providers[name]
		if config != nil && config.Enabled {
			results[name] = operation(name, provider)
		}
	}

	return results
}