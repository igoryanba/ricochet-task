package providers

import (
	"time"
)

// ProviderConfig contains configuration for a specific provider instance
type ProviderConfig struct {
	// Basic identification
	Name        string       `json:"name" yaml:"name"`
	Type        ProviderType `json:"type" yaml:"type"`
	Enabled     bool         `json:"enabled" yaml:"enabled"`
	Description string       `json:"description,omitempty" yaml:"description,omitempty"`

	// Connection settings
	BaseURL     string `json:"baseUrl,omitempty" yaml:"baseUrl,omitempty"`
	APIVersion  string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Region      string `json:"region,omitempty" yaml:"region,omitempty"`

	// Authentication
	AuthType   AuthenticationType     `json:"authType" yaml:"authType"`
	APIKey     string                 `json:"apiKey,omitempty" yaml:"apiKey,omitempty"`
	Token      string                 `json:"token,omitempty" yaml:"token,omitempty"`
	Username   string                 `json:"username,omitempty" yaml:"username,omitempty"`
	Password   string                 `json:"password,omitempty" yaml:"password,omitempty"`
	AuthConfig map[string]interface{} `json:"authConfig,omitempty" yaml:"authConfig,omitempty"`

	// Provider-specific settings
	Settings map[string]interface{} `json:"settings,omitempty" yaml:"settings,omitempty"`

	// Performance tuning
	RateLimit   *RateLimitConfig `json:"rateLimit,omitempty" yaml:"rateLimit,omitempty"`
	Timeout     time.Duration    `json:"timeout" yaml:"timeout"`
	RetryConfig *RetryConfig     `json:"retryConfig,omitempty" yaml:"retryConfig,omitempty"`

	// Caching
	CacheConfig *CacheConfig `json:"cacheConfig,omitempty" yaml:"cacheConfig,omitempty"`

	// Synchronization
	SyncConfig *SyncConfig `json:"syncConfig,omitempty" yaml:"syncConfig,omitempty"`

	// Security
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty" yaml:"tlsConfig,omitempty"`

	// Monitoring
	MetricsConfig *MetricsConfig `json:"metricsConfig,omitempty" yaml:"metricsConfig,omitempty"`
}

// MultiProviderConfig contains configuration for multiple providers
type MultiProviderConfig struct {
	// Default provider selection
	DefaultProvider string                      `json:"defaultProvider" yaml:"defaultProvider"`
	Providers       map[string]*ProviderConfig `json:"providers" yaml:"providers"`

	// Cross-provider settings
	GlobalSync   *GlobalSyncConfig `json:"globalSync,omitempty" yaml:"globalSync,omitempty"`
	Routing      *RoutingConfig    `json:"routing,omitempty" yaml:"routing,omitempty"`

	// AI integration
	AIChains     *AIChainConfig    `json:"aiChains,omitempty" yaml:"aiChains,omitempty"`

	// Quality gates
	QualityGates *QualityGatesConfig `json:"qualityGates,omitempty" yaml:"qualityGates,omitempty"`

	// Global settings
	LogLevel     string        `json:"logLevel" yaml:"logLevel"`
	MetricsPort  int           `json:"metricsPort,omitempty" yaml:"metricsPort,omitempty"`
	HealthCheck  time.Duration `json:"healthCheck" yaml:"healthCheck"`
}

// RateLimitConfig defines rate limiting settings
type RateLimitConfig struct {
	RequestsPerSecond float64       `json:"requestsPerSecond" yaml:"requestsPerSecond"`
	RequestsPerMinute int           `json:"requestsPerMinute,omitempty" yaml:"requestsPerMinute,omitempty"`
	RequestsPerHour   int           `json:"requestsPerHour,omitempty" yaml:"requestsPerHour,omitempty"`
	RequestsPerDay    int           `json:"requestsPerDay,omitempty" yaml:"requestsPerDay,omitempty"`
	BurstSize         int           `json:"burstSize" yaml:"burstSize"`
	BackoffStrategy   string        `json:"backoffStrategy,omitempty" yaml:"backoffStrategy,omitempty"`
	RetryAfter        time.Duration `json:"retryAfter,omitempty" yaml:"retryAfter,omitempty"`
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxRetries     int           `json:"maxRetries" yaml:"maxRetries"`
	InitialDelay   time.Duration `json:"initialDelay" yaml:"initialDelay"`
	MaxDelay       time.Duration `json:"maxDelay" yaml:"maxDelay"`
	BackoffFactor  float64       `json:"backoffFactor" yaml:"backoffFactor"`
	Jitter         bool          `json:"jitter" yaml:"jitter"`
	RetryableErrors []string     `json:"retryableErrors,omitempty" yaml:"retryableErrors,omitempty"`
}

// CacheConfig defines caching behavior
type CacheConfig struct {
	Enabled       bool          `json:"enabled" yaml:"enabled"`
	TTL           time.Duration `json:"ttl" yaml:"ttl"`
	MaxSize       int           `json:"maxSize,omitempty" yaml:"maxSize,omitempty"`
	CleanupInterval time.Duration `json:"cleanupInterval,omitempty" yaml:"cleanupInterval,omitempty"`
	
	// Cache strategies
	TasksTTL      time.Duration `json:"tasksTtl,omitempty" yaml:"tasksTtl,omitempty"`
	BoardsTTL     time.Duration `json:"boardsTtl,omitempty" yaml:"boardsTtl,omitempty"`
	ProjectsTTL   time.Duration `json:"projectsTtl,omitempty" yaml:"projectsTtl,omitempty"`
	
	// External cache
	Redis         *RedisConfig  `json:"redis,omitempty" yaml:"redis,omitempty"`
}

// RedisConfig defines Redis cache configuration
type RedisConfig struct {
	Address    string        `json:"address" yaml:"address"`
	Password   string        `json:"password,omitempty" yaml:"password,omitempty"`
	DB         int           `json:"db" yaml:"db"`
	PoolSize   int           `json:"poolSize,omitempty" yaml:"poolSize,omitempty"`
	Timeout    time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// SyncConfig defines synchronization settings
type SyncConfig struct {
	Enabled       bool          `json:"enabled" yaml:"enabled"`
	Interval      time.Duration `json:"interval" yaml:"interval"`
	BatchSize     int           `json:"batchSize,omitempty" yaml:"batchSize,omitempty"`
	ConflictResolution ConflictStrategy `json:"conflictResolution" yaml:"conflictResolution"`
	
	// Real-time sync
	RealTimeSync  bool   `json:"realTimeSync,omitempty" yaml:"realTimeSync,omitempty"`
	WebhookURL    string `json:"webhookUrl,omitempty" yaml:"webhookUrl,omitempty"`
	WebhookSecret string `json:"webhookSecret,omitempty" yaml:"webhookSecret,omitempty"`
	
	// Sync filters
	SyncFilters   *SyncFilters `json:"syncFilters,omitempty" yaml:"syncFilters,omitempty"`
}

// SyncFilters defines what should be synchronized
type SyncFilters struct {
	IncludeProjects []string `json:"includeProjects,omitempty" yaml:"includeProjects,omitempty"`
	ExcludeProjects []string `json:"excludeProjects,omitempty" yaml:"excludeProjects,omitempty"`
	IncludeTypes    []string `json:"includeTypes,omitempty" yaml:"includeTypes,omitempty"`
	ExcludeTypes    []string `json:"excludeTypes,omitempty" yaml:"excludeTypes,omitempty"`
	IncludeStatuses []string `json:"includeStatuses,omitempty" yaml:"includeStatuses,omitempty"`
	ExcludeStatuses []string `json:"excludeStatuses,omitempty" yaml:"excludeStatuses,omitempty"`
}

// TLSConfig defines TLS/SSL settings
type TLSConfig struct {
	Enabled            bool   `json:"enabled" yaml:"enabled"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify,omitempty" yaml:"insecureSkipVerify,omitempty"`
	CertFile           string `json:"certFile,omitempty" yaml:"certFile,omitempty"`
	KeyFile            string `json:"keyFile,omitempty" yaml:"keyFile,omitempty"`
	CAFile             string `json:"caFile,omitempty" yaml:"caFile,omitempty"`
}

// MetricsConfig defines metrics collection settings
type MetricsConfig struct {
	Enabled         bool          `json:"enabled" yaml:"enabled"`
	CollectInterval time.Duration `json:"collectInterval,omitempty" yaml:"collectInterval,omitempty"`
	Endpoint        string        `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Labels          map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

// GlobalSyncConfig defines cross-provider synchronization
type GlobalSyncConfig struct {
	Enabled    bool        `json:"enabled" yaml:"enabled"`
	Rules      []SyncRule  `json:"rules" yaml:"rules"`
	Interval   time.Duration `json:"interval" yaml:"interval"`
	BatchSize  int         `json:"batchSize,omitempty" yaml:"batchSize,omitempty"`
	MaxRetries int         `json:"maxRetries,omitempty" yaml:"maxRetries,omitempty"`
}

// SyncRule defines how to sync between providers
type SyncRule struct {
	Name           string            `json:"name" yaml:"name"`
	SourceProvider string            `json:"sourceProvider" yaml:"sourceProvider"`
	TargetProvider string            `json:"targetProvider" yaml:"targetProvider"`
	SyncType       SyncType          `json:"syncType" yaml:"syncType"`
	Conditions     []SyncCondition   `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	FieldMapping   map[string]string `json:"fieldMapping,omitempty" yaml:"fieldMapping,omitempty"`
	Enabled        bool              `json:"enabled" yaml:"enabled"`
	Priority       int               `json:"priority,omitempty" yaml:"priority,omitempty"`
}

type SyncType string

const (
	SyncTypeBidirectional   SyncType = "bidirectional"
	SyncTypeSourceToTarget  SyncType = "source_to_target"
	SyncTypeTargetToSource  SyncType = "target_to_source"
)

// SyncCondition defines when sync should occur
type SyncCondition struct {
	Field    string      `json:"field" yaml:"field"`
	Operator string      `json:"operator" yaml:"operator"`
	Value    interface{} `json:"value" yaml:"value"`
}

// RoutingConfig defines how tasks are routed to providers
type RoutingConfig struct {
	Rules           []RoutingRule `json:"rules" yaml:"rules"`
	DefaultProvider string        `json:"defaultProvider" yaml:"defaultProvider"`
	Strategy        RoutingStrategy `json:"strategy" yaml:"strategy"`
}

type RoutingStrategy string

const (
	RoutingStrategyRules      RoutingStrategy = "rules"
	RoutingStrategyRoundRobin RoutingStrategy = "round_robin"
	RoutingStrategyLoadBased  RoutingStrategy = "load_based"
	RoutingStrategyAI         RoutingStrategy = "ai"
)

// RoutingRule defines how to route tasks
type RoutingRule struct {
	Name       string            `json:"name" yaml:"name"`
	Condition  RoutingCondition  `json:"condition" yaml:"condition"`
	Provider   string            `json:"provider" yaml:"provider"`
	Priority   int               `json:"priority" yaml:"priority"`
	Enabled    bool              `json:"enabled" yaml:"enabled"`
	Weight     float64           `json:"weight,omitempty" yaml:"weight,omitempty"`
}

// RoutingCondition defines when a routing rule applies
type RoutingCondition struct {
	ProjectID   string       `json:"projectId,omitempty" yaml:"projectId,omitempty"`
	TaskType    TaskType     `json:"taskType,omitempty" yaml:"taskType,omitempty"`
	Priority    TaskPriority `json:"priority,omitempty" yaml:"priority,omitempty"`
	Labels      []string     `json:"labels,omitempty" yaml:"labels,omitempty"`
	Assignee    string       `json:"assignee,omitempty" yaml:"assignee,omitempty"`
	CustomField string       `json:"customField,omitempty" yaml:"customField,omitempty"`
	Query       string       `json:"query,omitempty" yaml:"query,omitempty"`
}

// AIChainConfig defines AI chain configurations
type AIChainConfig struct {
	Enabled bool                    `json:"enabled" yaml:"enabled"`
	Chains  map[string]*ChainConfig `json:"chains" yaml:"chains"`
	
	// Global AI settings
	DefaultModels *AIModelConfig `json:"defaultModels,omitempty" yaml:"defaultModels,omitempty"`
	TokenLimits   *TokenLimits   `json:"tokenLimits,omitempty" yaml:"tokenLimits,omitempty"`
	CostLimits    *CostLimits    `json:"costLimits,omitempty" yaml:"costLimits,omitempty"`
}

// ChainConfig defines a specific AI chain configuration
type ChainConfig struct {
	Name        string         `json:"name" yaml:"name"`
	Description string         `json:"description,omitempty" yaml:"description,omitempty"`
	Enabled     bool           `json:"enabled" yaml:"enabled"`
	Models      *AIModelConfig `json:"models" yaml:"models"`
	Triggers    []ChainTrigger `json:"triggers,omitempty" yaml:"triggers,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Timeout     time.Duration  `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// AIModelConfig defines AI model selections
type AIModelConfig struct {
	ProjectAnalyzer  string `json:"projectAnalyzer,omitempty" yaml:"projectAnalyzer,omitempty"`
	CodeExecutor     string `json:"codeExecutor,omitempty" yaml:"codeExecutor,omitempty"`
	QualityController string `json:"qualityController,omitempty" yaml:"qualityController,omitempty"`
	TaskPlanner      string `json:"taskPlanner,omitempty" yaml:"taskPlanner,omitempty"`
	CodeReviewer     string `json:"codeReviewer,omitempty" yaml:"codeReviewer,omitempty"`
	TestGenerator    string `json:"testGenerator,omitempty" yaml:"testGenerator,omitempty"`
	DocumentGenerator string `json:"documentGenerator,omitempty" yaml:"documentGenerator,omitempty"`
}

// ChainTrigger defines when AI chains should execute
type ChainTrigger struct {
	Type       TriggerType            `json:"type" yaml:"type"`
	Condition  string                 `json:"condition,omitempty" yaml:"condition,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// TokenLimits defines AI token usage limits
type TokenLimits struct {
	PerRequest int `json:"perRequest,omitempty" yaml:"perRequest,omitempty"`
	PerHour    int `json:"perHour,omitempty" yaml:"perHour,omitempty"`
	PerDay     int `json:"perDay,omitempty" yaml:"perDay,omitempty"`
	PerMonth   int `json:"perMonth,omitempty" yaml:"perMonth,omitempty"`
}

// CostLimits defines AI cost limits
type CostLimits struct {
	PerRequest float64 `json:"perRequest,omitempty" yaml:"perRequest,omitempty"`
	PerHour    float64 `json:"perHour,omitempty" yaml:"perHour,omitempty"`
	PerDay     float64 `json:"perDay,omitempty" yaml:"perDay,omitempty"`
	PerMonth   float64 `json:"perMonth,omitempty" yaml:"perMonth,omitempty"`
	Currency   string  `json:"currency,omitempty" yaml:"currency,omitempty"`
}

// QualityGatesConfig defines quality gate configurations
type QualityGatesConfig struct {
	Enabled bool                        `json:"enabled" yaml:"enabled"`
	Gates   map[string]*QualityGateConfig `json:"gates" yaml:"gates"`
	
	// Global settings
	DefaultThreshold  float64 `json:"defaultThreshold" yaml:"defaultThreshold"`
	FailOnBlockingGate bool   `json:"failOnBlockingGate" yaml:"failOnBlockingGate"`
}

// QualityGateConfig defines a specific quality gate
type QualityGateConfig struct {
	Name        string  `json:"name" yaml:"name"`
	Description string  `json:"description,omitempty" yaml:"description,omitempty"`
	Enabled     bool    `json:"enabled" yaml:"enabled"`
	Blocking    bool    `json:"blocking" yaml:"blocking"`
	Threshold   float64 `json:"threshold,omitempty" yaml:"threshold,omitempty"`
	
	// Gate-specific settings
	CodeCoverage     *CodeCoverageGate     `json:"codeCoverage,omitempty" yaml:"codeCoverage,omitempty"`
	UnitTests        *UnitTestsGate        `json:"unitTests,omitempty" yaml:"unitTests,omitempty"`
	IntegrationTests *IntegrationTestsGate `json:"integrationTests,omitempty" yaml:"integrationTests,omitempty"`
	SecurityScan     *SecurityScanGate     `json:"securityScan,omitempty" yaml:"securityScan,omitempty"`
	CodeStyle        *CodeStyleGate        `json:"codeStyle,omitempty" yaml:"codeStyle,omitempty"`
	Performance      *PerformanceGate      `json:"performance,omitempty" yaml:"performance,omitempty"`
}

// Specific quality gate configurations
type CodeCoverageGate struct {
	MinCoverage    float64 `json:"minCoverage" yaml:"minCoverage"`
	IncludeBranches bool   `json:"includeBranches,omitempty" yaml:"includeBranches,omitempty"`
	ExcludePaths   []string `json:"excludePaths,omitempty" yaml:"excludePaths,omitempty"`
}

type UnitTestsGate struct {
	MinPassRate     float64 `json:"minPassRate" yaml:"minPassRate"`
	RequiredTests   []string `json:"requiredTests,omitempty" yaml:"requiredTests,omitempty"`
	TestPatterns    []string `json:"testPatterns,omitempty" yaml:"testPatterns,omitempty"`
}

type IntegrationTestsGate struct {
	MinPassRate     float64 `json:"minPassRate" yaml:"minPassRate"`
	RequiredSuites  []string `json:"requiredSuites,omitempty" yaml:"requiredSuites,omitempty"`
	Timeout         time.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

type SecurityScanGate struct {
	MaxVulnerabilities map[string]int `json:"maxVulnerabilities,omitempty" yaml:"maxVulnerabilities,omitempty"`
	BlockCritical      bool          `json:"blockCritical" yaml:"blockCritical"`
	ScannerConfig      map[string]interface{} `json:"scannerConfig,omitempty" yaml:"scannerConfig,omitempty"`
}

type CodeStyleGate struct {
	Linters      []string `json:"linters,omitempty" yaml:"linters,omitempty"`
	ConfigFile   string   `json:"configFile,omitempty" yaml:"configFile,omitempty"`
	MaxIssues    int      `json:"maxIssues,omitempty" yaml:"maxIssues,omitempty"`
	IgnorePatterns []string `json:"ignorePatterns,omitempty" yaml:"ignorePatterns,omitempty"`
}

type PerformanceGate struct {
	MaxResponseTime time.Duration `json:"maxResponseTime,omitempty" yaml:"maxResponseTime,omitempty"`
	MaxMemoryUsage  int64         `json:"maxMemoryUsage,omitempty" yaml:"maxMemoryUsage,omitempty"`
	MaxCPUUsage     float64       `json:"maxCpuUsage,omitempty" yaml:"maxCpuUsage,omitempty"`
	BenchmarkTests  []string      `json:"benchmarkTests,omitempty" yaml:"benchmarkTests,omitempty"`
}

// Helper methods for configuration validation
func (c *ProviderConfig) Validate() error {
	if c.Name == "" {
		return NewProviderError(ErrorTypeValidation, "provider name is required", nil)
	}
	
	if c.Type == "" {
		return NewProviderError(ErrorTypeValidation, "provider type is required", nil)
	}
	
	if c.AuthType == "" {
		return NewProviderError(ErrorTypeValidation, "authentication type is required", nil)
	}
	
	// Validate authentication fields based on auth type
	switch c.AuthType {
	case AuthTypeAPIKey:
		if c.APIKey == "" {
			return NewProviderError(ErrorTypeValidation, "API key is required for API key authentication", nil)
		}
	case AuthTypeBearer:
		if c.Token == "" {
			return NewProviderError(ErrorTypeValidation, "token is required for bearer authentication", nil)
		}
	case AuthTypeBasic:
		if c.Username == "" || c.Password == "" {
			return NewProviderError(ErrorTypeValidation, "username and password are required for basic authentication", nil)
		}
	case AuthTypeOAuth2:
		if c.AuthConfig == nil {
			return NewProviderError(ErrorTypeValidation, "auth config is required for OAuth2 authentication", nil)
		}
	}
	
	return nil
}

func (c *MultiProviderConfig) Validate() error {
	if len(c.Providers) == 0 {
		return NewProviderError(ErrorTypeValidation, "at least one provider must be configured", nil)
	}
	
	// Validate default provider exists
	if c.DefaultProvider != "" {
		if _, exists := c.Providers[c.DefaultProvider]; !exists {
			return NewProviderError(ErrorTypeValidation, "default provider does not exist in providers list", nil)
		}
	}
	
	// Validate each provider
	for name, provider := range c.Providers {
		if err := provider.Validate(); err != nil {
			return NewProviderError(ErrorTypeValidation, "invalid provider "+name, err)
		}
	}
	
	return nil
}

func (c *ProviderConfig) GetSetting(key string) interface{} {
	if c.Settings == nil {
		return nil
	}
	return c.Settings[key]
}

func (c *ProviderConfig) GetStringSetting(key string) string {
	if value := c.GetSetting(key); value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func (c *ProviderConfig) GetBoolSetting(key string) bool {
	if value := c.GetSetting(key); value != nil {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

func (c *ProviderConfig) GetIntSetting(key string) int {
	if value := c.GetSetting(key); value != nil {
		if i, ok := value.(int); ok {
			return i
		}
		if f, ok := value.(float64); ok {
			return int(f)
		}
	}
	return 0
}

// Default configurations
func DefaultProviderConfig() *ProviderConfig {
	return &ProviderConfig{
		Enabled: true,
		Timeout: 30 * time.Second,
		RateLimit: &RateLimitConfig{
			RequestsPerSecond: 10,
			BurstSize:         20,
		},
		RetryConfig: &RetryConfig{
			MaxRetries:    3,
			InitialDelay:  1 * time.Second,
			MaxDelay:      30 * time.Second,
			BackoffFactor: 2.0,
			Jitter:        true,
		},
		CacheConfig: &CacheConfig{
			Enabled: true,
			TTL:     5 * time.Minute,
			MaxSize: 1000,
		},
		SyncConfig: &SyncConfig{
			Enabled:           false,
			Interval:          30 * time.Second,
			ConflictResolution: ConflictResolveManual,
		},
	}
}

func DefaultMultiProviderConfig() *MultiProviderConfig {
	return &MultiProviderConfig{
		Providers:   make(map[string]*ProviderConfig),
		LogLevel:    "info",
		HealthCheck: 1 * time.Minute,
		GlobalSync: &GlobalSyncConfig{
			Enabled:   false,
			Interval:  5 * time.Minute,
			BatchSize: 100,
		},
		QualityGates: &QualityGatesConfig{
			Enabled:          true,
			DefaultThreshold: 0.8,
			FailOnBlockingGate: true,
			Gates: map[string]*QualityGateConfig{
				"unit_tests": {
					Name:     "Unit Tests",
					Enabled:  true,
					Blocking: true,
					UnitTests: &UnitTestsGate{
						MinPassRate: 0.95,
					},
				},
				"code_coverage": {
					Name:     "Code Coverage",
					Enabled:  true,
					Blocking: false,
					CodeCoverage: &CodeCoverageGate{
						MinCoverage: 0.8,
					},
				},
			},
		},
	}
}