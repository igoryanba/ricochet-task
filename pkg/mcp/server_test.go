package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sirupsen/logrus"

	"github.com/grik-ai/ricochet-task/pkg/providers"
)

// createTestMCPServer creates a test MCP server
func createTestMCPServer() (*MCPServer, *MockProviderRegistry) {
	mockRegistry := new(MockProviderRegistry)
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	server := NewMCPServer(mockRegistry, logger)
	return server, mockRegistry
}

// TestNewMCPServer tests MCP server creation
func TestNewMCPServer(t *testing.T) {
	mockRegistry := new(MockProviderRegistry)
	logger := logrus.New()

	server := NewMCPServer(mockRegistry, logger)

	assert.NotNil(t, server)
	assert.Equal(t, mockRegistry, server.toolProvider.registry)
	assert.Equal(t, logger, server.logger)
	assert.NotNil(t, server.connections)
	assert.Equal(t, 0, len(server.connections))
}

// TestProcessMessage tests MCP message processing
func TestProcessMessage(t *testing.T) {
	server, _ := createTestMCPServer()

	t.Run("Initialize request", func(t *testing.T) {
		msg := &MCPMessage{
			JSONRPC: "2.0",
			ID:      "1",
			Method:  "initialize",
			Params:  map[string]interface{}{},
		}

		response := server.processMessage(msg)

		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Equal(t, "1", response.ID)
		assert.Nil(t, response.Error)
		assert.NotNil(t, response.Result)

		result := response.Result.(map[string]interface{})
		assert.Equal(t, "2024-11-05", result["protocolVersion"])
		assert.Contains(t, result, "capabilities")
		assert.Contains(t, result, "serverInfo")
	})

	t.Run("Tools list request", func(t *testing.T) {
		msg := &MCPMessage{
			JSONRPC: "2.0",
			ID:      "2",
			Method:  "tools/list",
			Params:  map[string]interface{}{},
		}

		response := server.processMessage(msg)

		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Equal(t, "2", response.ID)
		assert.Nil(t, response.Error)
		assert.NotNil(t, response.Result)

		result := response.Result.(map[string]interface{})
		tools := result["tools"].([]ToolInfo)
		assert.Len(t, tools, 9) // We have 9 tools
	})

	t.Run("Ping request", func(t *testing.T) {
		msg := &MCPMessage{
			JSONRPC: "2.0",
			ID:      "3",
			Method:  "ping",
			Params:  map[string]interface{}{},
		}

		response := server.processMessage(msg)

		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Equal(t, "3", response.ID)
		assert.Nil(t, response.Error)
		assert.Equal(t, "pong", response.Result)
	})

	t.Run("Unknown method", func(t *testing.T) {
		msg := &MCPMessage{
			JSONRPC: "2.0",
			ID:      "4",
			Method:  "unknown/method",
			Params:  map[string]interface{}{},
		}

		response := server.processMessage(msg)

		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Equal(t, "4", response.ID)
		assert.NotNil(t, response.Error)
		assert.Equal(t, -32601, response.Error.Code)
		assert.Contains(t, response.Error.Message, "Method not found")
	})
}

// TestHandleToolCall tests tool call handling
func TestHandleToolCall(t *testing.T) {
	server, mockRegistry := createTestMCPServer()

	t.Run("Successful tool call", func(t *testing.T) {
		mockProviders := map[string]*providers.ProviderInfo{
			"test-provider": {
				Name:         "test-provider",
				HealthStatus: providers.HealthStatusHealthy,
			},
		}

		mockRegistry.On("ListProviders").Return(mockProviders)

		msg := &MCPMessage{
			JSONRPC: "2.0",
			ID:      "1",
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "providers_list",
				"arguments": map[string]interface{}{
					"enabled_only":   false,
					"output_format": "table",
				},
			},
		}

		response := server.processMessage(msg)

		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Equal(t, "1", response.ID)
		assert.Nil(t, response.Error)
		assert.NotNil(t, response.Result)

		result := response.Result.(map[string]interface{})
		assert.Contains(t, result, "content")
		assert.Equal(t, false, result["isError"])

		mockRegistry.AssertExpectations(t)
	})

	t.Run("Tool call with missing tool name", func(t *testing.T) {
		msg := &MCPMessage{
			JSONRPC: "2.0",
			ID:      "2",
			Method:  "tools/call",
			Params: map[string]interface{}{
				"arguments": map[string]interface{}{},
			},
		}

		response := server.processMessage(msg)

		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Equal(t, "2", response.ID)
		assert.NotNil(t, response.Error)
		assert.Equal(t, -32602, response.Error.Code)
		assert.Contains(t, response.Error.Message, "tool name required")
	})

	t.Run("Tool call with invalid tool name", func(t *testing.T) {
		msg := &MCPMessage{
			JSONRPC: "2.0",
			ID:      "3",
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name":      "invalid_tool",
				"arguments": map[string]interface{}{},
			},
		}

		response := server.processMessage(msg)

		assert.NotNil(t, response)
		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Equal(t, "3", response.ID)
		assert.NotNil(t, response.Error)
		assert.Equal(t, -32603, response.Error.Code)
		assert.Contains(t, response.Error.Message, "Unknown tool")
	})
}

// TestHTTPEndpoints tests HTTP endpoints
func TestHTTPEndpoints(t *testing.T) {
	server, mockRegistry := createTestMCPServer()

	t.Run("Health endpoint", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/health", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		server.handleHealth(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "healthy", response["status"])
		assert.Contains(t, response, "timestamp")
		assert.Contains(t, response, "connections")
		assert.Contains(t, response, "tools")
	})

	t.Run("Tools list endpoint", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/tools", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		server.handleToolsList(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "tools")
		assert.Contains(t, response, "count")
		assert.Equal(t, float64(9), response["count"]) // 9 tools
	})

	t.Run("HTTP tools list endpoint", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/mcp/tools", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		server.handleHTTPToolsList(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "tools")
		assert.Contains(t, response, "count")
	})

	t.Run("HTTP tool call endpoint", func(t *testing.T) {
		mockProviders := map[string]*providers.ProviderInfo{
			"test-provider": {
				Name:         "test-provider",
				HealthStatus: providers.HealthStatusHealthy,
			},
		}

		mockRegistry.On("ListProviders").Return(mockProviders)

		requestBody := `{
			"tool": "providers_list",
			"arguments": {
				"enabled_only": false,
				"output_format": "table"
			}
		}`

		req, err := http.NewRequest("POST", "/mcp/tools/call", strings.NewReader(requestBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		server.handleHTTPToolCall(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "result")

		mockRegistry.AssertExpectations(t)
	})

	t.Run("HTTP tool call with invalid method", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/mcp/tools/call", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		server.handleHTTPToolCall(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("HTTP tool call with invalid JSON", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/mcp/tools/call", strings.NewReader("invalid json"))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		server.handleHTTPToolCall(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("HTTP tool call with missing tool name", func(t *testing.T) {
		requestBody := `{"arguments": {}}`

		req, err := http.NewRequest("POST", "/mcp/tools/call", strings.NewReader(requestBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		server.handleHTTPToolCall(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestWebSocketUpgrade tests WebSocket upgrade
func TestWebSocketUpgrade(t *testing.T) {
	server, _ := createTestMCPServer()

	// Create test server
	testServer := httptest.NewServer(http.HandlerFunc(server.handleWebSocket))
	defer testServer.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")

	t.Run("Successful WebSocket connection", func(t *testing.T) {
		// Connect to WebSocket
		dialer := websocket.Dialer{}
		conn, _, err := dialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Send initialize message
		initMsg := MCPMessage{
			JSONRPC: "2.0",
			ID:      "1",
			Method:  "initialize",
			Params:  map[string]interface{}{},
		}

		err = conn.WriteJSON(initMsg)
		require.NoError(t, err)

		// Read response
		var response MCPMessage
		err = conn.ReadJSON(&response)
		require.NoError(t, err)

		assert.Equal(t, "2.0", response.JSONRPC)
		assert.Equal(t, "1", response.ID)
		assert.Nil(t, response.Error)
		assert.NotNil(t, response.Result)
	})
}

// TestBroadcastToolListChanged tests tool list change broadcasting
func TestBroadcastToolListChanged(t *testing.T) {
	server, _ := createTestMCPServer()

	// Simulate some connections
	mockConn1 := &mockWebSocketConn{}
	mockConn2 := &mockWebSocketConn{}

	server.connections["conn1"] = mockConn1
	server.connections["conn2"] = mockConn2

	// Broadcast change
	server.BroadcastToolListChanged()

	// Verify that messages were sent to all connections
	assert.Equal(t, 1, mockConn1.writeCount)
	assert.Equal(t, 1, mockConn2.writeCount)
}

// TestGetConnectionCount tests connection counting
func TestGetConnectionCount(t *testing.T) {
	server, _ := createTestMCPServer()

	assert.Equal(t, 0, server.GetConnectionCount())

	// Add mock connections
	server.connections["conn1"] = &mockWebSocketConn{}
	server.connections["conn2"] = &mockWebSocketConn{}

	assert.Equal(t, 2, server.GetConnectionCount())
}

// TestGetToolCount tests tool counting
func TestGetToolCount(t *testing.T) {
	server, _ := createTestMCPServer()

	assert.Equal(t, 9, server.GetToolCount()) // We have 9 tools
}

// TestShutdown tests server shutdown
func TestShutdown(t *testing.T) {
	server, _ := createTestMCPServer()

	// Add mock connections
	mockConn1 := &mockWebSocketConn{}
	mockConn2 := &mockWebSocketConn{}

	server.connections["conn1"] = mockConn1
	server.connections["conn2"] = mockConn2

	err := server.Shutdown()
	assert.NoError(t, err)

	// Verify connections were closed
	assert.True(t, mockConn1.closed)
	assert.True(t, mockConn2.closed)
}

// TestLogRequest tests request logging
func TestLogRequest(t *testing.T) {
	server, _ := createTestMCPServer()

	// This test verifies that logging doesn't panic
	// In a real scenario, you might want to capture log output
	server.LogRequest("test/method", map[string]interface{}{"param": "value"})

	// No assertion needed - just verify no panic
}

// TestLogResponse tests response logging
func TestLogResponse(t *testing.T) {
	server, _ := createTestMCPServer()

	// Test successful response logging
	server.LogResponse("test/method", map[string]interface{}{"result": "success"}, nil)

	// Test error response logging
	server.LogResponse("test/method", nil, assert.AnError)

	// No assertion needed - just verify no panic
}

// TestConcurrentConnections tests concurrent WebSocket connections
func TestConcurrentConnections(t *testing.T) {
	server, mockRegistry := createTestMCPServer()

	// Setup mock
	mockProviders := map[string]*providers.ProviderInfo{
		"test-provider": {
			Name:         "test-provider",
			HealthStatus: providers.HealthStatusHealthy,
		},
	}
	mockRegistry.On("ListProviders").Return(mockProviders)

	// Create test server
	testServer := httptest.NewServer(http.HandlerFunc(server.handleWebSocket))
	defer testServer.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")

	concurrency := 3
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(connID int) {
			dialer := websocket.Dialer{}
			conn, _, err := dialer.Dial(wsURL, nil)
			if err != nil {
				errors <- err
				return
			}
			defer conn.Close()

			// Send tool call
			msg := MCPMessage{
				JSONRPC: "2.0",
				ID:      string(rune('1' + connID)),
				Method:  "tools/call",
				Params: map[string]interface{}{
					"name": "providers_list",
					"arguments": map[string]interface{}{
						"enabled_only": false,
					},
				},
			}

			err = conn.WriteJSON(msg)
			if err != nil {
				errors <- err
				return
			}

			var response MCPMessage
			err = conn.ReadJSON(&response)
			errors <- err
		}(i)
	}

	// Collect results
	for i := 0; i < concurrency; i++ {
		err := <-errors
		assert.NoError(t, err)
	}

	mockRegistry.AssertExpectations(t)
}

// MockWebSocketConn is a mock WebSocket connection for testing
type mockWebSocketConn struct {
	writeCount int
	closed     bool
}

func (m *mockWebSocketConn) WriteJSON(v interface{}) error {
	m.writeCount++
	return nil
}

func (m *mockWebSocketConn) ReadJSON(v interface{}) error {
	return nil
}

func (m *mockWebSocketConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockWebSocketConn) WriteMessage(messageType int, data []byte) error {
	m.writeCount++
	return nil
}

func (m *mockWebSocketConn) ReadMessage() (messageType int, p []byte, err error) {
	return websocket.TextMessage, []byte(`{"jsonrpc":"2.0","method":"ping"}`), nil
}

// BenchmarkMCPServer benchmarks MCP server operations
func BenchmarkMCPServer(b *testing.B) {
	server, mockRegistry := createTestMCPServer()

	mockProviders := map[string]*providers.ProviderInfo{
		"bench-provider": {
			Name:         "bench-provider",
			HealthStatus: providers.HealthStatusHealthy,
		},
	}
	mockRegistry.On("ListProviders").Return(mockProviders)

	b.Run("ProcessMessage", func(b *testing.B) {
		msg := &MCPMessage{
			JSONRPC: "2.0",
			ID:      "bench",
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "providers_list",
				"arguments": map[string]interface{}{
					"enabled_only": false,
				},
			},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			response := server.processMessage(msg)
			if response.Error != nil {
				b.Fatalf("Message processing failed: %v", response.Error)
			}
		}
	})

	b.Run("HTTPToolCall", func(b *testing.B) {
		requestBody := `{
			"tool": "providers_list",
			"arguments": {
				"enabled_only": false
			}
		}`

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("POST", "/mcp/tools/call", strings.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			server.handleHTTPToolCall(rr, req)

			if rr.Code != http.StatusOK {
				b.Fatalf("HTTP tool call failed: %d", rr.Code)
			}
		}
	})
}
