package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/grik-ai/ricochet-task/pkg/providers"
)

// HTTPServer provides HTTP interface for MCP tools
type HTTPServer struct {
	toolProvider *MCPToolProvider
	logger       *logrus.Logger
	server       *http.Server
}

// NewHTTPServer creates a new HTTP server for MCP tools
func NewHTTPServer(registry *providers.ProviderRegistry, logger *logrus.Logger) *HTTPServer {
	if logger == nil {
		logger = logrus.New()
	}

	return &HTTPServer{
		toolProvider: NewMCPToolProvider(registry),
		logger:       logger,
	}
}

// ToolListResponse represents the response for listing tools
type ToolListResponse struct {
	Tools []ToolDefinition `json:"tools"`
}

// ToolExecuteRequest represents a request to execute a tool
type ToolExecuteRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolExecuteResponse represents the response from tool execution
type ToolExecuteResponse struct {
	Content []map[string]interface{} `json:"content"`
	IsError bool                     `json:"isError"`
	Error   *string                  `json:"error,omitempty"`
}

// Start starts the HTTP server
func (s *HTTPServer) Start(addr string) error {
	mux := http.NewServeMux()
	
	// CORS middleware
	corsHandler := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next(w, r)
		}
	}

	// Routes
	mux.HandleFunc("/health", corsHandler(s.handleHealth))
	mux.HandleFunc("/tools", corsHandler(s.handleTools))
	mux.HandleFunc("/tools/execute", corsHandler(s.handleToolExecute))

	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	s.logger.Infof("Starting MCP HTTP server on %s", addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// handleHealth handles health check requests
func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":  "healthy",
		"service": "ricochet-task-mcp",
		"version": "1.0.0",
		"time":    time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleTools handles tool listing requests
func (s *HTTPServer) handleTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tools := s.toolProvider.GetTools()
	response := ToolListResponse{Tools: tools}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Errorf("Failed to encode tools response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	s.logger.Infof("Listed %d tools", len(tools))
}

// handleToolExecute handles tool execution requests
func (s *HTTPServer) handleToolExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ToolExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Errorf("Failed to decode request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.logger.Infof("Executing tool: %s", req.Name)

	result, err := s.toolProvider.ExecuteTool(ctx, req.Name, req.Arguments)
	if err != nil {
		s.logger.Errorf("Tool execution failed: %v", err)
		response := ToolExecuteResponse{
			IsError: true,
			Error:   stringPtr(fmt.Sprintf("Tool execution failed: %v", err)),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ToolExecuteResponse{
		Content: result.Content,
		IsError: result.Error != nil,
		Error:   result.Error,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Errorf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	s.logger.Infof("Tool %s executed successfully", req.Name)
}

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}