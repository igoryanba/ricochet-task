package workflow

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestMCPIntegration тестирует основную MCP интеграцию
func TestMCPIntegration(t *testing.T) {
	logger := &MockLogger{}
	eventBus := NewEventBus(logger)
	
	// Создаем базовые компоненты
	config := &MCPConfig{
		MaxConcurrentOps: 5,
		Timeout:          10 * time.Second,
		RetryAttempts:    2,
		EnableAutoTools:  true,
		SecurityPolicy: &MCPSecurityPolicy{
			AllowedTools:     []string{"ai_analysis", "workflow_control"},
			RequireApproval:  false,
			SandboxMode:      true,
			TimeoutPerTool:   5 * time.Second,
		},
	}

	mcp := NewMCPIntegration(nil, nil, eventBus, config, logger)

	t.Run("RegisterTool", func(t *testing.T) {
		// Создаем тестовый инструмент
		tool := NewAIAnalysisTool(nil, logger)
		
		err := mcp.RegisterTool(tool)
		if err != nil {
			t.Fatalf("Failed to register tool: %v", err)
		}

		// Проверяем что инструмент зарегистрирован
		if len(mcp.tools) == 0 {
			t.Error("Tool not registered")
		}

		if mcp.tools["ai_analysis"] == nil {
			t.Error("AI analysis tool not found")
		}
	})

	t.Run("GetAvailableTools", func(t *testing.T) {
		tools := mcp.GetAvailableTools()
		
		if len(tools) == 0 {
			t.Error("No tools available")
		}

		// Проверяем что есть встроенные инструменты
		foundAITool := false
		for _, tool := range tools {
			if tool.Name == "ai_analysis" {
				foundAITool = true
				break
			}
		}

		if !foundAITool {
			t.Error("AI analysis tool not found in available tools")
		}
	})

	t.Run("ExecuteTool", func(t *testing.T) {
		input := &MCPToolInput{
			ToolName: "ai_analysis",
			Parameters: map[string]interface{}{
				"text":          "This is a test text for analysis",
				"analysis_type": "sentiment",
			},
			Context: &MCPExecutionContext{
				Environment: "test",
				Permissions: []string{"read", "execute"},
			},
			RequestID: "test-request-1",
			Timestamp: time.Now(),
		}

		output, err := mcp.ExecuteTool(context.Background(), input)
		if err != nil {
			t.Fatalf("Failed to execute tool: %v", err)
		}

		if output == nil {
			t.Fatal("Tool output is nil")
		}

		if output.RequestID != input.RequestID {
			t.Error("Request ID mismatch")
		}

		// Проверяем что есть результат
		if output.Result == nil && output.Success {
			t.Error("Expected result for successful execution")
		}
	})

	t.Run("SecurityPolicy", func(t *testing.T) {
		// Тестируем блокированный инструмент
		config.SecurityPolicy.BlockedTools = []string{"ai_analysis"}
		
		tool := NewAIAnalysisTool(nil, logger)
		_ = mcp.RegisterTool(tool)
	})
}

// TestAIAnalysisTool тестирует AI анализ инструмент
func TestAIAnalysisTool(t *testing.T) {
	logger := &MockLogger{}
	tool := NewAIAnalysisTool(nil, logger)

	t.Run("GetSchema", func(t *testing.T) {
		schema := tool.GetSchema()
		
		if schema.Name != "ai_analysis" {
			t.Error("Wrong tool name in schema")
		}

		if len(schema.Parameters) == 0 {
			t.Error("No parameters in schema")
		}

		// Проверяем обязательные параметры
		hasTextParam := false
		hasTypeParam := false
		
		for _, param := range schema.Parameters {
			if param.Name == "text" && param.Required {
				hasTextParam = true
			}
			if param.Name == "analysis_type" && param.Required {
				hasTypeParam = true
			}
		}

		if !hasTextParam {
			t.Error("Missing required text parameter")
		}

		if !hasTypeParam {
			t.Error("Missing required analysis_type parameter")
		}
	})

	t.Run("ValidateInput", func(t *testing.T) {
		// Валидный ввод
		validInput := &MCPToolInput{
			Parameters: map[string]interface{}{
				"text":          "Test text",
				"analysis_type": "sentiment",
			},
		}

		err := tool.ValidateInput(validInput)
		if err != nil {
			t.Errorf("Valid input rejected: %v", err)
		}

		// Невалидный ввод - отсутствует text
		invalidInput1 := &MCPToolInput{
			Parameters: map[string]interface{}{
				"analysis_type": "sentiment",
			},
		}

		err = tool.ValidateInput(invalidInput1)
		if err == nil {
			t.Error("Expected error for missing text parameter")
		}

		// Невалидный тип анализа
		invalidInput2 := &MCPToolInput{
			Parameters: map[string]interface{}{
				"text":          "Test text",
				"analysis_type": "invalid_type",
			},
		}

		err = tool.ValidateInput(invalidInput2)
		if err == nil {
			t.Error("Expected error for invalid analysis type")
		}
	})

	t.Run("Execute", func(t *testing.T) {
		input := &MCPToolInput{
			ToolName: "ai_analysis",
			Parameters: map[string]interface{}{
				"text":          "This is a great product!",
				"analysis_type": "sentiment",
			},
			RequestID: "test-analysis-1",
		}

		// Без AI chains - должен вернуть ошибку
		output, _ := tool.Execute(context.Background(), input)
		
		// Проверяем что получили ответ (даже если с ошибкой)
		if output == nil {
			t.Fatal("Output is nil")
		}

		if output.RequestID != input.RequestID {
			t.Error("Request ID not preserved")
		}
	})
}

// TestWorkflowControlTool тестирует инструмент управления workflow
func TestWorkflowControlTool(t *testing.T) {
	logger := &MockLogger{}
	tool := NewWorkflowControlTool(nil, logger)

	t.Run("GetSchema", func(t *testing.T) {
		schema := tool.GetSchema()
		
		if schema.Name != "workflow_control" {
			t.Error("Wrong tool name")
		}

		if len(schema.Capabilities) == 0 {
			t.Error("No capabilities defined")
		}
	})

	t.Run("ListWorkflows", func(t *testing.T) {
		input := &MCPToolInput{
			Parameters: map[string]interface{}{
				"action": "list",
			},
		}

		output, err := tool.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("Failed to list workflows: %v", err)
		}

		if !output.Success {
			t.Error("List workflows failed")
		}

		// Проверяем структуру результата
		result, ok := output.Result.([]map[string]interface{})
		if !ok {
			t.Error("Unexpected result format")
		}

		if len(result) == 0 {
			t.Error("No workflows returned")
		}
	})

	t.Run("GetStatus", func(t *testing.T) {
		input := &MCPToolInput{
			Parameters: map[string]interface{}{
				"action":      "status",
				"workflow_id": "test-workflow-1",
			},
		}

		output, err := tool.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("Failed to get status: %v", err)
		}

		if !output.Success {
			t.Error("Get status failed")
		}
	})

	t.Run("TransitionTask", func(t *testing.T) {
		input := &MCPToolInput{
			Parameters: map[string]interface{}{
				"action":    "transition",
				"task_id":   "task-123",
				"new_stage": "testing",
			},
		}

		output, err := tool.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("Failed to transition task: %v", err)
		}

		if !output.Success {
			t.Error("Task transition failed")
		}

		// Проверяем результат перехода
		result, ok := output.Result.(map[string]interface{})
		if !ok {
			t.Error("Unexpected result format")
		}

		if result["task_id"] != "task-123" {
			t.Error("Task ID not preserved")
		}

		if result["new_stage"] != "testing" {
			t.Error("New stage not set correctly")
		}
	})
}

// TestResourceManagementTool тестирует инструмент управления ресурсами
func TestResourceManagementTool(t *testing.T) {
	logger := &MockLogger{}
	resourceManager := NewMCPResourceManager(logger)
	tool := NewResourceManagementTool(resourceManager, logger)

	t.Run("CreateResource", func(t *testing.T) {
		input := &MCPToolInput{
			Parameters: map[string]interface{}{
				"action":       "create",
				"resource_uri": "test://resource1",
				"resource_data": map[string]interface{}{
					"title": "Test Resource",
					"data":  "Some test data",
				},
			},
		}

		output, err := tool.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("Failed to create resource: %v", err)
		}

		if !output.Success {
			t.Error("Resource creation failed")
		}
	})

	t.Run("GrantAndListAccess", func(t *testing.T) {
		// Сначала предоставляем доступ
		grantInput := &MCPToolInput{
			Parameters: map[string]interface{}{
				"action":       "grant_access",
				"resource_uri": "test://resource1",
				"user_id":      "user123",
			},
		}

		output, err := tool.Execute(context.Background(), grantInput)
		if err != nil {
			t.Fatalf("Failed to grant access: %v", err)
		}

		if !output.Success {
			t.Error("Grant access failed")
		}

		// Теперь проверяем список ресурсов для пользователя
		listInput := &MCPToolInput{
			Parameters: map[string]interface{}{
				"action":  "list",
				"user_id": "user123",
			},
		}

		listOutput, err := tool.Execute(context.Background(), listInput)
		if err != nil {
			t.Fatalf("Failed to list resources: %v", err)
		}

		if !listOutput.Success {
			t.Error("List resources failed")
		}

		// Проверяем что ресурс в списке
		result, ok := listOutput.Result.(map[string]interface{})
		if !ok {
			t.Error("Unexpected result format")
		}

		resources, ok := result["resources"].([]string)
		if !ok {
			t.Error("Resources not found in result")
		}

		if len(resources) == 0 {
			t.Error("No resources found for user")
		}

		found := false
		for _, uri := range resources {
			if uri == "test://resource1" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Created resource not found in user's list")
		}
	})

	t.Run("ReadResource", func(t *testing.T) {
		input := &MCPToolInput{
			Parameters: map[string]interface{}{
				"action":       "read",
				"resource_uri": "test://resource1",
				"user_id":      "user123",
			},
		}

		output, err := tool.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("Failed to read resource: %v", err)
		}

		if !output.Success {
			t.Error("Resource read failed")
		}

		// Проверяем что получили ресурс
		resource, ok := output.Result.(*MCPResource)
		if !ok {
			t.Error("Unexpected result format")
		}

		if resource.URI != "test://resource1" {
			t.Error("Wrong resource URI")
		}
	})
}

// TestCodeAnalysisTool тестирует инструмент анализа кода
func TestCodeAnalysisTool(t *testing.T) {
	logger := &MockLogger{}
	tool := NewCodeAnalysisTool(nil, logger)

	t.Run("ValidateInput", func(t *testing.T) {
		validInput := &MCPToolInput{
			Parameters: map[string]interface{}{
				"code":     "function hello() { return 'world'; }",
				"language": "javascript",
			},
		}

		err := tool.ValidateInput(validInput)
		if err != nil {
			t.Errorf("Valid input rejected: %v", err)
		}

		invalidInput := &MCPToolInput{
			Parameters: map[string]interface{}{
				"language": "javascript",
				// Отсутствует code
			},
		}

		err = tool.ValidateInput(invalidInput)
		if err == nil {
			t.Error("Expected error for missing code parameter")
		}
	})

	t.Run("Execute", func(t *testing.T) {
		input := &MCPToolInput{
			Parameters: map[string]interface{}{
				"code":          "function add(a, b) { return a + b; }",
				"language":      "javascript",
				"analysis_type": "quality",
			},
		}

		// Без AI chains будет ошибка, но структура ответа должна быть корректной
		output, _ := tool.Execute(context.Background(), input)
		
		if output == nil {
			t.Fatal("Output is nil")
		}

		// Проверяем что есть метаданные
		if output.Metadata == nil {
			t.Error("No metadata in output")
		}

		if output.Metadata["language"] != "javascript" {
			t.Error("Language not preserved in metadata")
		}
	})
}

// TestNotificationTool тестирует инструмент уведомлений
func TestNotificationTool(t *testing.T) {
	logger := &MockLogger{}
	tool := NewNotificationTool(nil, logger)

	t.Run("ValidateInput", func(t *testing.T) {
		validInput := &MCPToolInput{
			Parameters: map[string]interface{}{
				"title":      "Test Notification",
				"message":    "This is a test message",
				"recipients": []interface{}{"user1@example.com", "user2@example.com"},
			},
		}

		err := tool.ValidateInput(validInput)
		if err != nil {
			t.Errorf("Valid input rejected: %v", err)
		}

		invalidInput := &MCPToolInput{
			Parameters: map[string]interface{}{
				"message":    "Missing title",
				"recipients": []interface{}{"user1@example.com"},
			},
		}

		err = tool.ValidateInput(invalidInput)
		if err == nil {
			t.Error("Expected error for missing title")
		}
	})

	t.Run("Execute", func(t *testing.T) {
		input := &MCPToolInput{
			Parameters: map[string]interface{}{
				"title":      "Test Notification",
				"message":    "This is a test notification message",
				"recipients": []interface{}{"user1@example.com", "user2@example.com"},
				"priority":   "high",
			},
		}

		output, err := tool.Execute(context.Background(), input)
		if err != nil {
			t.Fatalf("Failed to execute notification tool: %v", err)
		}

		if !output.Success {
			t.Error("Notification tool execution failed")
		}

		// Проверяем результат
		result, ok := output.Result.(map[string]interface{})
		if !ok {
			t.Error("Unexpected result format")
		}

		if result["title"] != "Test Notification" {
			t.Error("Title not preserved")
		}

		if result["priority"] != "high" {
			t.Error("Priority not preserved")
		}

		recipients, ok := result["recipients"].([]string)
		if !ok {
			t.Error("Recipients format incorrect")
		}

		if len(recipients) != 2 {
			t.Error("Wrong number of recipients")
		}
	})
}

// BenchmarkMCPToolExecution бенчмарк выполнения MCP инструментов
func BenchmarkMCPToolExecution(b *testing.B) {
	logger := &MockLogger{}
	mcp := NewMCPIntegration(nil, nil, NewEventBus(logger), nil, logger)

	input := &MCPToolInput{
		ToolName: "ai_analysis",
		Parameters: map[string]interface{}{
			"text":          "Sample text for analysis",
			"analysis_type": "sentiment",
		},
		Context: &MCPExecutionContext{
			Environment: "benchmark",
			Permissions: []string{"execute"},
		},
		RequestID: "benchmark-request",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mcp.ExecuteTool(context.Background(), input)
	}
}

// BenchmarkResourceAccess бенчмарк доступа к ресурсам
func BenchmarkResourceAccess(b *testing.B) {
	logger := &MockLogger{}
	rm := NewMCPResourceManager(logger)

	// Создаем тестовые ресурсы
	for i := 0; i < 100; i++ {
		resource := &MCPResource{
			URI:  fmt.Sprintf("test://resource%d", i),
			Type: "data",
			Content: map[string]interface{}{
				"id":   i,
				"data": fmt.Sprintf("Resource %d data", i),
			},
		}
		rm.AddResource(resource.URI, resource)
		rm.GrantAccess(resource.URI, "user123")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rm.GetResource(fmt.Sprintf("test://resource%d", i%100), "user123")
	}
}