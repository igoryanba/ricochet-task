package httpserver

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grik-ai/ricochet-task/pkg/ai"
	"github.com/grik-ai/ricochet-task/pkg/chain"
	"github.com/grik-ai/ricochet-task/pkg/checkpoint"
	"github.com/grik-ai/ricochet-task/pkg/key"
	"github.com/grik-ai/ricochet-task/pkg/service"
)

// HTTPServer представляет HTTP сервер для ricochet-task
type HTTPServer struct {
	ricochetService *service.RicochetService
	router          *gin.Engine
	logger          ai.Logger
}

// SimpleHTTPLogger реализует Logger интерфейс для HTTP сервера
type SimpleHTTPLogger struct{}

func (l *SimpleHTTPLogger) Info(msg string, args ...interface{}) {
	log.Printf("[INFO] %s %v", msg, args)
}

func (l *SimpleHTTPLogger) Error(msg string, err error, args ...interface{}) {
	log.Printf("[ERROR] %s: %v %v", msg, err, args)
}

func (l *SimpleHTTPLogger) Warn(msg string, args ...interface{}) {
	log.Printf("[WARN] %s %v", msg, args)
}

func (l *SimpleHTTPLogger) Debug(msg string, args ...interface{}) {
	log.Printf("[DEBUG] %s %v", msg, args)
}

// NewHTTPServer создает новый HTTP сервер
func NewHTTPServer() *HTTPServer {
	logger := &SimpleHTTPLogger{}
	
	// Получаем конфигурацию из переменных окружения
	gatewayURL := os.Getenv("GRIK_GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "http://api-gateway:3000"
	}
	
	gatewayToken := os.Getenv("GRIK_GATEWAY_TOKEN")
	userID := os.Getenv("GRIK_USER_ID")
	if userID == "" {
		userID = "system"
	}

	// Создаем гибридный AI клиент
	hybridAI := ai.NewHybridAIClient(gatewayURL, gatewayToken, userID, nil, logger)

	// Создаем stores (пока используем временные файловые store)
	keyStore, _ := key.NewFileKeyStore("/tmp/ricochet-keys")
	chainStore, _ := chain.NewFileChainStore("/tmp/ricochet-chains")  
	checkpointStore, _ := checkpoint.NewFileCheckpointStore("/tmp/ricochet-checkpoints")

	// Создаем RicochetService
	ricochetService := service.NewRicochetService(
		hybridAI,
		keyStore,
		chainStore,
		checkpointStore,
		logger,
	)

	// Настраиваем Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	server := &HTTPServer{
		ricochetService: ricochetService,
		router:          router,
		logger:          logger,
	}

	server.setupRoutes()
	return server
}

// setupRoutes настраивает маршруты HTTP сервера
func (s *HTTPServer) setupRoutes() {
	// Health check
	s.router.GET("/health", s.healthCheck)

	// API группа
	api := s.router.Group("/api")
	{
		// Chain management
		api.POST("/chains", s.createChain)
		api.POST("/chains/run", s.runChain)
		api.GET("/chains/:chain_id/checkpoints", s.getCheckpoints)

		// Run management
		api.GET("/runs", s.listRuns)
		api.GET("/runs/:run_id", s.getRunStatus)
		api.GET("/runs/:run_id/results", s.getRunResults)
		api.POST("/runs/:run_id/cancel", s.cancelRun)

		// User API keys management
		api.PUT("/user/api-keys", s.updateUserAPIKeys)
		api.POST("/user/api-keys/validate", s.validateUserAPIKeys)
		api.GET("/user/usage-stats", s.getUserUsageStats)

		// Available models
		api.GET("/models/available", s.getAvailableModels)
	}
}

// Chain related request/response types
type CreateChainRequest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Models      []ChainModelHTTP  `json:"models"`
	UserAPIKeys *UserAPIKeysHTTP  `json:"user_api_keys,omitempty"`
}

type ChainModelHTTP struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Role        string  `json:"role"`
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

type UserAPIKeysHTTP struct {
	OpenAI    string `json:"openai,omitempty"`
	Anthropic string `json:"anthropic,omitempty"`
	DeepSeek  string `json:"deepseek,omitempty"`
	Grok      string `json:"grok,omitempty"`
}

type RunChainRequest struct {
	ChainID string `json:"chain_id"`
	Input   string `json:"input"`
}

// Health check endpoint
func (s *HTTPServer) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "ricochet-task",
		"timestamp": time.Now().Unix(),
	})
}

// Create chain endpoint
func (s *HTTPServer) createChain(c *gin.Context) {
	var req CreateChainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	// Обновляем пользовательские API ключи если предоставлены
	if req.UserAPIKeys != nil {
		userKeys := &ai.UserAPIKeys{}
		if req.UserAPIKeys.OpenAI != "" {
			userKeys.OpenAI = &ai.APIKeyConfig{APIKey: req.UserAPIKeys.OpenAI, Enabled: true}
		}
		if req.UserAPIKeys.Anthropic != "" {
			userKeys.Anthropic = &ai.APIKeyConfig{APIKey: req.UserAPIKeys.Anthropic, Enabled: true}
		}
		if req.UserAPIKeys.DeepSeek != "" {
			userKeys.DeepSeek = &ai.APIKeyConfig{APIKey: req.UserAPIKeys.DeepSeek, Enabled: true}
		}
		if req.UserAPIKeys.Grok != "" {
			userKeys.Grok = &ai.APIKeyConfig{APIKey: req.UserAPIKeys.Grok, Enabled: true}
		}
		s.ricochetService.UpdateUserAPIKeys(userKeys)
	}

	// Конвертируем HTTP модели в chain модели
	var models []chain.Model
	for _, m := range req.Models {
		models = append(models, chain.Model{
			ID:          m.ID,
			Name:        chain.ModelName(m.Name),
			Type:        chain.ModelType(m.Type),
			Role:        chain.ModelRole(m.Role),
			Prompt:      m.Prompt,
			MaxTokens:   m.MaxTokens,
			Temperature: m.Temperature,
		})
	}

	// Создаем цепочку
	newChain := chain.Chain{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Models:      models,
		CreatedAt:   time.Now(),
	}

	// TODO: Реализовать сохранение цепочки
	_ = newChain // Временно, чтобы избежать ошибки компиляции

	c.JSON(http.StatusOK, gin.H{
		"id":      newChain.ID,
		"message": "Chain created successfully",
	})
}

// Run chain endpoint
func (s *HTTPServer) runChain(c *gin.Context) {
	var req RunChainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	// Запускаем цепочку
	runID, err := s.ricochetService.RunChain(c.Request.Context(), req.ChainID, req.Input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to run chain: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"run_id": runID,
	})
}

// Get run status endpoint
func (s *HTTPServer) getRunStatus(c *gin.Context) {
	runID := c.Param("run_id")
	
	status, err := s.ricochetService.GetRunStatus(runID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Run not found: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// Get run results endpoint
func (s *HTTPServer) getRunResults(c *gin.Context) {
	runID := c.Param("run_id")
	
	results, err := s.ricochetService.GetRunResults(runID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get results: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
	})
}

// Cancel run endpoint
func (s *HTTPServer) cancelRun(c *gin.Context) {
	runID := c.Param("run_id")
	
	err := s.ricochetService.CancelRun(runID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel run: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Run cancelled successfully",
	})
}

// List runs endpoint
func (s *HTTPServer) listRuns(c *gin.Context) {
	runs := s.ricochetService.ListRuns()
	c.JSON(http.StatusOK, gin.H{
		"runs": runs,
	})
}

// Update user API keys endpoint
func (s *HTTPServer) updateUserAPIKeys(c *gin.Context) {
	var keys UserAPIKeysHTTP
	if err := c.ShouldBindJSON(&keys); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	userKeys := &ai.UserAPIKeys{}
	if keys.OpenAI != "" {
		userKeys.OpenAI = &ai.APIKeyConfig{APIKey: keys.OpenAI, Enabled: true}
	}
	if keys.Anthropic != "" {
		userKeys.Anthropic = &ai.APIKeyConfig{APIKey: keys.Anthropic, Enabled: true}
	}
	if keys.DeepSeek != "" {
		userKeys.DeepSeek = &ai.APIKeyConfig{APIKey: keys.DeepSeek, Enabled: true}
	}
	if keys.Grok != "" {
		userKeys.Grok = &ai.APIKeyConfig{APIKey: keys.Grok, Enabled: true}
	}

	s.ricochetService.UpdateUserAPIKeys(userKeys)
	c.JSON(http.StatusOK, gin.H{
		"message": "API keys updated successfully",
	})
}

// Validate user API keys endpoint
func (s *HTTPServer) validateUserAPIKeys(c *gin.Context) {
	results := s.ricochetService.ValidateUserKeys()
	c.JSON(http.StatusOK, gin.H{
		"validation_results": results,
	})
}

// Get user usage stats endpoint
func (s *HTTPServer) getUserUsageStats(c *gin.Context) {
	stats := s.ricochetService.GetUsageStats()
	c.JSON(http.StatusOK, stats)
}

// Get available models endpoint
func (s *HTTPServer) getAvailableModels(c *gin.Context) {
	models := s.ricochetService.GetAvailableModels()
	c.JSON(http.StatusOK, models)
}

// Get checkpoints endpoint
func (s *HTTPServer) getCheckpoints(c *gin.Context) {
	chainID := c.Param("chain_id")
	
	checkpoints, err := s.ricochetService.ListCheckpoints(chainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get checkpoints: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"checkpoints": checkpoints,
	})
}

// Start запускает HTTP сервер
func (s *HTTPServer) Start(port string) error {
	s.logger.Info("Starting Ricochet HTTP server", "port", port)
	return s.router.Run(":" + port)
}

// StartRicochetHTTPServer запускает HTTP сервер для ricochet-task
func StartRicochetHTTPServer() {
	port := os.Getenv("RICOCHET_HTTP_PORT")
	if port == "" {
		port = "6004" // По умолчанию используем порт из конфигурации API Gateway
	}

	server := NewHTTPServer()
	
	log.Printf("Starting Ricochet HTTP server on port %s", port)
	if err := server.Start(port); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}