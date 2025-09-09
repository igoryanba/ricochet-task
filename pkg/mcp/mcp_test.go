//go:build integration
// +build integration

package mcp_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grik-ai/ricochet-task/pkg/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMCPServerSetup тестирует настройку MCP-сервера
func TestMCPServerSetup(t *testing.T) {
	// Создаем новый сервер MCP
	server := mcp.NewMCPServer()
	assert.NotNil(t, server, "MCP Server should not be nil")

	// Регистрируем тестовую команду
	testHandler := func(params json.RawMessage) (interface{}, error) {
		return map[string]string{"status": "ok", "message": "Test command executed"}, nil
	}
	server.RegisterCommand("test_command", testHandler)
}

// TestMCPServerHandleRequest тестирует обработку запросов MCP-сервером
func TestMCPServerHandleRequest(t *testing.T) {
	// Создаем новый сервер MCP
	server := mcp.NewMCPServer()

	// Регистрируем тестовую команду
	testHandler := func(params json.RawMessage) (interface{}, error) {
		var p struct {
			Value string `json:"value"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, err
		}
		return map[string]string{"status": "ok", "message": "Received: " + p.Value}, nil
	}
	server.RegisterCommand("test_command", testHandler)

	// Создаем тестовый HTTP-запрос
	reqBody := mcp.MCPRequest{
		Command: "test_command",
		Params:  json.RawMessage(`{"value": "test value"}`),
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/mcp", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	// Создаем рекордер для записи ответа
	w := httptest.NewRecorder()

	// Вызываем обработчик
	server.HandleMCPRequest(w, req)

	// Проверяем ответ
	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response mcp.MCPResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "success", response.Status)
	assert.Equal(t, "test_command", response.Command)

	// Проверяем данные ответа
	respData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "ok", respData["status"])
	assert.Equal(t, "Received: test value", respData["message"])
}

// TestChainCreateCommand тестирует команду создания цепочки
func TestChainCreateCommand(t *testing.T) {
	// Создаем новый сервер MCP с зарегистрированными командами
	server := mcp.NewMCPServer()
	mcp.RegisterChainCreateCommands(server)

	// Создаем тестовые параметры для команды chain_create
	params := mcp.ChainCreateParams{
		Name:        "Test Chain",
		Description: "Chain for testing",
		Steps: []mcp.ChainStep{
			{
				RoleID:      "analyzer",
				ModelID:     "gpt-3.5-turbo",
				Provider:    "openai",
				Name:        "Text Analysis",
				Description: "Analyze text structure",
				Prompt:      "Analyze the following text and identify key components.",
			},
			{
				RoleID:      "summarizer",
				ModelID:     "gpt-3.5-turbo",
				Provider:    "openai",
				Name:        "Summarization",
				Description: "Summarize analysis results",
				Prompt:      "Create a concise summary based on the analysis.",
			},
		},
		Interactive: false,
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	// Создаем тестовый HTTP-запрос
	reqBody := mcp.MCPRequest{
		Command: "chain_create",
		Params:  paramsJSON,
	}
	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/mcp", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	// Создаем рекордер для записи ответа
	w := httptest.NewRecorder()

	// Вызываем обработчик
	server.HandleMCPRequest(w, req)

	// Проверяем ответ
	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response mcp.MCPResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "success", response.Status)
	assert.Equal(t, "chain_create", response.Command)

	// Проверяем данные ответа
	respData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, respData, "chain_id")
	assert.Contains(t, respData, "name")
	assert.Equal(t, "Test Chain", respData["name"])
}

// TestChainBuilderCommands тестирует команды построителя цепочек
func TestChainBuilderCommands(t *testing.T) {
	// Создаем новый сервер MCP с зарегистрированными командами
	server := mcp.NewMCPServer()
	mcp.RegisterChainBuilderCommands(server)

	// Шаг 1: Инициализация сессии построителя
	initParams := mcp.ChainBuilderInitParams{
		ChainName:        "Builder Test Chain",
		ChainDescription: "Chain created through builder test",
	}
	initParamsJSON, err := json.Marshal(initParams)
	require.NoError(t, err)

	// Создаем тестовый HTTP-запрос для инициализации
	initReqBody := mcp.MCPRequest{
		Command: "chain_builder_init",
		Params:  initParamsJSON,
	}
	initReqJSON, err := json.Marshal(initReqBody)
	require.NoError(t, err)

	initReq := httptest.NewRequest("POST", "/mcp", bytes.NewBuffer(initReqJSON))
	initReq.Header.Set("Content-Type", "application/json")

	// Создаем рекордер для записи ответа
	initW := httptest.NewRecorder()

	// Вызываем обработчик
	server.HandleMCPRequest(initW, initReq)

	// Проверяем ответ инициализации
	initResp := initW.Result()
	defer initResp.Body.Close()

	assert.Equal(t, http.StatusOK, initResp.StatusCode)

	var initResponse mcp.MCPResponse
	err = json.NewDecoder(initResp.Body).Decode(&initResponse)
	require.NoError(t, err)

	assert.Equal(t, "success", initResponse.Status)
	assert.Equal(t, "chain_builder_init", initResponse.Command)

	// Получаем ID сессии из ответа
	initData, ok := initResponse.Data.(map[string]interface{})
	require.True(t, ok)
	sessionID, ok := initData["session_id"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, sessionID)

	// Шаг 2: Добавление шага в построитель
	addStepParams := mcp.ChainBuilderStepParams{
		SessionID:   sessionID,
		StepIndex:   0,
		ModelRole:   "analyzer",
		ModelID:     "gpt-4",
		Provider:    "openai",
		Description: "Text Analysis",
		Prompt:      "Analyze the text structure and content.",
	}
	addStepParamsJSON, err := json.Marshal(addStepParams)
	require.NoError(t, err)

	// Создаем тестовый HTTP-запрос для добавления шага
	addStepReqBody := mcp.MCPRequest{
		Command: "chain_builder_add_step",
		Params:  addStepParamsJSON,
	}
	addStepReqJSON, err := json.Marshal(addStepReqBody)
	require.NoError(t, err)

	addStepReq := httptest.NewRequest("POST", "/mcp", bytes.NewBuffer(addStepReqJSON))
	addStepReq.Header.Set("Content-Type", "application/json")

	// Создаем рекордер для записи ответа
	addStepW := httptest.NewRecorder()

	// Вызываем обработчик
	server.HandleMCPRequest(addStepW, addStepReq)

	// Проверяем ответ добавления шага
	addStepResp := addStepW.Result()
	defer addStepResp.Body.Close()

	assert.Equal(t, http.StatusOK, addStepResp.StatusCode)

	var addStepResponse mcp.MCPResponse
	err = json.NewDecoder(addStepResp.Body).Decode(&addStepResponse)
	require.NoError(t, err)

	assert.Equal(t, "success", addStepResponse.Status)
	assert.Equal(t, "chain_builder_add_step", addStepResponse.Command)

	// Шаг 3: Получение сессии построителя
	getSessionParams := struct {
		SessionID string `json:"session_id"`
	}{
		SessionID: sessionID,
	}
	getSessionParamsJSON, err := json.Marshal(getSessionParams)
	require.NoError(t, err)

	// Создаем тестовый HTTP-запрос для получения сессии
	getSessionReqBody := mcp.MCPRequest{
		Command: "chain_builder_get_session",
		Params:  getSessionParamsJSON,
	}
	getSessionReqJSON, err := json.Marshal(getSessionReqBody)
	require.NoError(t, err)

	getSessionReq := httptest.NewRequest("POST", "/mcp", bytes.NewBuffer(getSessionReqJSON))
	getSessionReq.Header.Set("Content-Type", "application/json")

	// Создаем рекордер для записи ответа
	getSessionW := httptest.NewRecorder()

	// Вызываем обработчик
	server.HandleMCPRequest(getSessionW, getSessionReq)

	// Проверяем ответ получения сессии
	getSessionResp := getSessionW.Result()
	defer getSessionResp.Body.Close()

	assert.Equal(t, http.StatusOK, getSessionResp.StatusCode)

	var getSessionResponse mcp.MCPResponse
	err = json.NewDecoder(getSessionResp.Body).Decode(&getSessionResponse)
	require.NoError(t, err)

	assert.Equal(t, "success", getSessionResponse.Status)
	assert.Equal(t, "chain_builder_get_session", getSessionResponse.Command)

	// Проверяем данные сессии
	sessionData, ok := getSessionResponse.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, sessionID, sessionData["id"])
	assert.Equal(t, "Builder Test Chain", sessionData["chain_name"])

	// Проверяем, что шаг добавлен
	steps, ok := sessionData["steps"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 1, len(steps))

	step, ok := steps[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "analyzer", step["model_role"])
	assert.Equal(t, "gpt-4", step["model_id"])
}

// TestChainControlCommands тестирует команды управления цепочками
func TestChainControlCommands(t *testing.T) {
	// Создаем новый сервер MCP с зарегистрированными командами
	server := mcp.NewMCPServer()
	mcp.RegisterChainControlCommands(server)

	// Подготавливаем параметры для команды chain_pause
	pauseParams := mcp.ChainControlParams{
		ChainID: "test-chain-1",
		Reason:  "Pause for testing",
	}
	pauseParamsJSON, err := json.Marshal(pauseParams)
	require.NoError(t, err)

	// Создаем тестовый HTTP-запрос для паузы цепочки
	pauseReqBody := mcp.MCPRequest{
		Command: "chain_pause",
		Params:  pauseParamsJSON,
	}
	pauseReqJSON, err := json.Marshal(pauseReqBody)
	require.NoError(t, err)

	pauseReq := httptest.NewRequest("POST", "/mcp", bytes.NewBuffer(pauseReqJSON))
	pauseReq.Header.Set("Content-Type", "application/json")

	// Создаем рекордер для записи ответа
	pauseW := httptest.NewRecorder()

	// Вызываем обработчик
	server.HandleMCPRequest(pauseW, pauseReq)

	// Проверяем ответ паузы
	pauseResp := pauseW.Result()
	defer pauseResp.Body.Close()

	assert.Equal(t, http.StatusOK, pauseResp.StatusCode)

	var pauseResponse mcp.MCPResponse
	err = json.NewDecoder(pauseResp.Body).Decode(&pauseResponse)
	require.NoError(t, err)

	assert.Equal(t, "success", pauseResponse.Status)
	assert.Equal(t, "chain_pause", pauseResponse.Command)

	// Проверяем данные ответа
	pauseData, ok := pauseResponse.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test-chain-1", pauseData["chain_id"])
	assert.Equal(t, "pause", pauseData["action"])
	assert.Equal(t, "paused", pauseData["status"])
}

// TestModelCommands тестирует получение доступных моделей и автоселект
func TestModelCommands(t *testing.T) {
	server := mcp.NewMCPServer()
	// Регистрируем необходимые команды
	mcp.RegisterModelCommands(server)
	mcp.RegisterChainCreateCommands(server)
	mcp.RegisterChainBuilderCommands(server)
	mcp.RegisterChainInteractiveBuilderCommands(server)

	// Проверяем chain_get_available_models
	reqBody := mcp.MCPRequest{
		Command: "chain_get_available_models",
		Params:  json.RawMessage(`{}`),
	}
	reqJSON, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/mcp", bytes.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.HandleMCPRequest(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var mcpResp mcp.MCPResponse
	_ = json.NewDecoder(resp.Body).Decode(&mcpResp)
	assert.Equal(t, "success", mcpResp.Status)
	assert.Equal(t, "chain_get_available_models", mcpResp.Command)

	// Создаем простую цепочку, чтобы проверить auto_select_models
	chainParams := mcp.ChainCreateParams{
		Name: "AutoSelect Chain",
		Steps: []mcp.ChainStep{
			{RoleID: "analyzer", Name: "Шаг 1"},
			{RoleID: "summarizer", Name: "Шаг 2"},
		},
	}
	chainJSON, _ := json.Marshal(chainParams)
	createReqBody := mcp.MCPRequest{Command: "chain_create", Params: chainJSON}
	createReqJSON, _ := json.Marshal(createReqBody)
	createReq := httptest.NewRequest("POST", "/mcp", bytes.NewReader(createReqJSON))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	server.HandleMCPRequest(createW, createReq)
	var createResp mcp.MCPResponse
	_ = json.NewDecoder(createW.Result().Body).Decode(&createResp)
	chainID := createResp.Data.(map[string]interface{})["chain_id"].(string)

	// Вызываем auto_select_models
	autoParams := map[string]string{"chain_id": chainID}
	autoJSON, _ := json.Marshal(autoParams)
	autoReqBody := mcp.MCPRequest{Command: "auto_select_models", Params: autoJSON}
	autoReqJSON, _ := json.Marshal(autoReqBody)
	autoReq := httptest.NewRequest("POST", "/mcp", bytes.NewReader(autoReqJSON))
	autoReq.Header.Set("Content-Type", "application/json")
	autoW := httptest.NewRecorder()
	server.HandleMCPRequest(autoW, autoReq)

	autoResp := autoW.Result()
	assert.Equal(t, http.StatusOK, autoResp.StatusCode)

	var autoMCPResp mcp.MCPResponse
	_ = json.NewDecoder(autoResp.Body).Decode(&autoMCPResp)
	assert.Equal(t, "success", autoMCPResp.Status)
	assert.Equal(t, "auto_select_models", autoMCPResp.Command)
}
