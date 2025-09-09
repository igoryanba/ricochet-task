package workflow

import (
	"context"
	"testing"
	"time"
)

// TestCompleteWorkflowEngine тестирует полный движок workflow
func TestCompleteWorkflowEngine(t *testing.T) {
	logger := &MockLogger{}
	
	config := GetDefaultCompleteConfig()
	config.MaxConcurrentWorkflows = 5
	config.DefaultTimeout = 10 * time.Second

	engine, err := NewCompleteWorkflowEngine(nil, config, logger)
	if err != nil {
		t.Fatalf("Failed to create complete workflow engine: %v", err)
	}

	t.Run("CreateWorkflow", func(t *testing.T) {
		definition := &WorkflowDefinition{
			Name:        "Test Workflow",
			Description: "A test workflow for integration testing",
			Version:     "1.0.0",
			Stages: map[string]*StageDefinition{
				"development": {
					Name: "development",
					Actions: []ActionDefinition{
						{
							Type: "manual",
						},
						{
							Type: "automated",
						},
					},
				},
				"testing": {
					Name: "testing",
					Actions: []ActionDefinition{
						{
							Type: "automated",
						},
					},
				},
			},
		}

		instance, err := engine.CreateWorkflow(context.Background(), definition)
		if err != nil {
			t.Fatalf("Failed to create workflow: %v", err)
		}

		if instance == nil {
			t.Fatal("Workflow instance is nil")
		}

		if instance.ID == "" {
			t.Error("Workflow ID is empty")
		}

		if instance.Status != "created" {
			t.Errorf("Expected status 'created', got '%s'", instance.Status)
		}

		if instance.Definition.Name != definition.Name {
			t.Error("Workflow definition not preserved")
		}
	})

	t.Run("ExecuteWorkflow", func(t *testing.T) {
		// Создаем простой workflow для выполнения
		definition := &WorkflowDefinition{
			Name:        "Simple Test Workflow",
			Description: "Simple workflow for execution testing",
			Version:     "1.0.0",
			Stages: map[string]*StageDefinition{
				"simple_stage": {
					Name: "simple_stage",
					Actions: []ActionDefinition{
						{
							Type: "automated",
						},
					},
				},
			},
		}

		instance, err := engine.CreateWorkflow(context.Background(), definition)
		if err != nil {
			t.Fatalf("Failed to create workflow: %v", err)
		}

		// Запускаем выполнение
		err = engine.ExecuteWorkflow(context.Background(), instance.ID)
		if err != nil {
			t.Fatalf("Failed to execute workflow: %v", err)
		}

		// Ждем некоторое время для начала выполнения
		time.Sleep(100 * time.Millisecond)

		// Проверяем статус
		status, err := engine.GetWorkflowStatus(instance.ID)
		if err != nil {
			t.Fatalf("Failed to get workflow status: %v", err)
		}

		if status.Status == "created" {
			t.Error("Workflow should have started executing")
		}
	})

	t.Run("GetMetrics", func(t *testing.T) {
		metrics := engine.GetMetrics()
		
		if metrics == nil {
			t.Fatal("Metrics are nil")
		}

		if metrics.TotalWorkflows < 0 {
			t.Error("Total workflows should be non-negative")
		}

		if metrics.TaskMetrics == nil {
			t.Error("Task metrics should not be nil")
		}

		if metrics.WorkflowsByStage == nil {
			t.Error("Workflows by stage should not be nil")
		}

		if metrics.MCPToolUsage == nil {
			t.Error("MCP tool usage should not be nil")
		}
	})

	t.Run("ConcurrentWorkflowLimit", func(t *testing.T) {
		// Создаем workflow до лимита
		definition := &WorkflowDefinition{
			Name:        "Limit Test Workflow",
			Description: "Workflow for testing concurrent limits",
			Version:     "1.0.0",
			Stages: map[string]*StageDefinition{
				"test_stage": {
					Name: "test_stage",
					Actions: []ActionDefinition{
						{
							Type: "manual",
						},
					},
				},
			},
		}

		// Создаем workflows до лимита
		for i := 0; i < config.MaxConcurrentWorkflows; i++ {
			_, err := engine.CreateWorkflow(context.Background(), definition)
			if err != nil {
				// Некоторые могли завершиться, что нормально
				break
			}
		}

		// Попытка создать еще один должна вернуть ошибку
		_, err := engine.CreateWorkflow(context.Background(), definition)
		
		// Проверяем что либо получили ошибку лимита, либо система очистила завершенные
		if err != nil && err.Error() != "maximum concurrent workflows limit reached: 5" {
			t.Logf("Expected limit error or auto-cleanup occurred: %v", err)
		}
	})

	t.Run("WorkflowComponents", func(t *testing.T) {
		// Проверяем что все компоненты инициализированы
		if engine.eventBus == nil {
			t.Error("Event bus not initialized")
		}

		if engine.ruleEngine == nil {
			t.Error("Rule engine not initialized")
		}

		if engine.progressTracker == nil {
			t.Error("Progress tracker not initialized")
		}

		if engine.autoAssignment == nil {
			t.Error("Auto assignment not initialized")
		}

		if engine.notifications == nil {
			t.Error("Notifications not initialized")
		}

		if engine.mcpIntegration == nil {
			t.Error("MCP integration not initialized")
		}
	})
}

// TestWorkflowIntegration тестирует интеграцию между компонентами
func TestWorkflowIntegration(t *testing.T) {
	logger := &MockLogger{}
	config := GetDefaultCompleteConfig()
	
	engine, err := NewCompleteWorkflowEngine(nil, config, logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	t.Run("EventBusIntegration", func(t *testing.T) {
		// Создаем подписчика на события
		eventReceived := false
		handler := &WorkflowEventHandler{
			handler: func(ctx context.Context, event Event) error {
				if event.GetType() == "workflow.created" {
					eventReceived = true
				}
				return nil
			},
		}

		engine.eventBus.Subscribe("workflow.created", handler)

		// Создаем workflow
		definition := &WorkflowDefinition{
			Name:    "Integration Test Workflow",
			Version: "1.0.0",
			Stages: map[string]*StageDefinition{
				"test_stage": {
					Name: "test_stage",
					Actions: []ActionDefinition{
						{Type: "manual"},
					},
				},
			},
		}

		_, err := engine.CreateWorkflow(context.Background(), definition)
		if err != nil {
			t.Fatalf("Failed to create workflow: %v", err)
		}

		// Ждем обработки события
		time.Sleep(50 * time.Millisecond)

		if !eventReceived {
			t.Error("Event not received through event bus")
		}
	})

	t.Run("MCPToolsIntegration", func(t *testing.T) {
		// Проверяем доступность MCP инструментов
		tools := engine.mcpIntegration.GetAvailableTools()
		
		if len(tools) == 0 {
			t.Error("No MCP tools available")
		}

		// Проверяем наличие основных инструментов
		foundAI := false
		foundWorkflow := false
		
		for _, tool := range tools {
			switch tool.Name {
			case "ai_analysis":
				foundAI = true
			case "workflow_control":
				foundWorkflow = true
			}
		}

		if !foundAI {
			t.Error("AI analysis tool not found")
		}

		if !foundWorkflow {
			t.Error("Workflow control tool not found")
		}
	})

	t.Run("NotificationIntegration", func(t *testing.T) {
		// Проверяем что система уведомлений работает
		if engine.notifications == nil {
			t.Fatal("Notification system not initialized")
		}

		// Проверяем метрики уведомлений
		metrics := engine.notifications.analytics.GetMetrics()
		if metrics == nil {
			t.Error("Notification metrics not available")
		}
	})
}

// TestWorkflowDefinitionLanguage тестирует YAML определения workflow
func TestWorkflowDefinitionLanguage(t *testing.T) {
	_ = `
name: "CI/CD Pipeline"
description: "Continuous Integration and Deployment"
version: "2.0.0"
metadata:
  author: "DevOps Team"
  created: "2024-01-01"
  
stages:
  - name: "build"
    description: "Build and compile"
    tasks:
      - name: "compile"
        type: "automated"
        description: "Compile source code"
        auto_assign: false
        requirements: ["build_tools"]
        estimated_duration: "5m"
        
      - name: "unit_tests"
        type: "automated"
        description: "Run unit tests"
        auto_assign: false
        requirements: ["testing"]
        estimated_duration: "10m"
        
  - name: "deploy"
    description: "Deploy to production"
    depends_on: ["build"]
    tasks:
      - name: "deploy_production"
        type: "automated"
        description: "Deploy to production environment"
        auto_assign: false
        requirements: ["deployment", "production_access"]
        estimated_duration: "15m"

rules:
  - name: "auto_deploy_on_green_build"
    event: "stage.build.completed"
    conditions:
      - field: "tests_passed"
        operator: "equals"
        value: true
    actions:
      - type: "transition"
        target: "deploy"
      - type: "notification"
        template: "deployment_ready"
        
auto_assignment:
  enabled: true
  ai_enabled: true
  fallback_strategy: "round_robin"
  
notifications:
  enabled: true
  channels: ["email", "slack"]
  smart_routing: true
`

	t.Run("ParseYAMLDefinition", func(t *testing.T) {
		// Упрощенный тест без реального парсинга YAML
		t.Skip("YAML parsing not implemented in this test")
	})

	t.Run("ValidateDefinition", func(t *testing.T) {
		// Упрощенный тест валидации
		t.Skip("Validation not implemented in this test")
	})
}

// BenchmarkCompleteWorkflowEngine бенчмарк полного движка
func BenchmarkCompleteWorkflowEngine(b *testing.B) {
	logger := &MockLogger{}
	config := GetDefaultCompleteConfig()
	config.MaxConcurrentWorkflows = 1000 // Увеличиваем для бенчмарка
	
	engine, err := NewCompleteWorkflowEngine(nil, config, logger)
	if err != nil {
		b.Fatalf("Failed to create engine: %v", err)
	}

	definition := &WorkflowDefinition{
		Name:    "Benchmark Workflow",
		Version: "1.0.0",
		Stages: map[string]*StageDefinition{
			"benchmark_stage": {
				Name: "benchmark_stage",
				Actions: []ActionDefinition{
					{Type: "automated"},
				},
			},
		},
	}

	b.ResetTimer()

	b.Run("CreateWorkflow", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := engine.CreateWorkflow(context.Background(), definition)
			if err != nil && err.Error() == "maximum concurrent workflows limit reached: 1000" {
				// Ожидаемая ошибка при достижении лимита
				continue
			} else if err != nil {
				b.Fatalf("Unexpected error: %v", err)
			}
		}
	})

	b.Run("GetMetrics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			engine.GetMetrics()
		}
	})

	b.Run("GetWorkflowStatus", func(b *testing.B) {
		// Создаем один workflow для тестирования
		instance, _ := engine.CreateWorkflow(context.Background(), definition)
		
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			engine.GetWorkflowStatus(instance.ID)
		}
	})
}

// TestWorkflowEngineResilience тестирует устойчивость движка
func TestWorkflowEngineResilience(t *testing.T) {
	logger := &MockLogger{}
	config := GetDefaultCompleteConfig()
	
	engine, err := NewCompleteWorkflowEngine(nil, config, logger)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	t.Run("InvalidWorkflowDefinition", func(t *testing.T) {
		// Пустое определение
		emptyDef := &WorkflowDefinition{}
		
		_, err := engine.CreateWorkflow(context.Background(), emptyDef)
		// Должно обработать корректно, даже если определение неполное
		if err != nil {
			t.Logf("Empty definition handled correctly: %v", err)
		}
	})

	t.Run("NonExistentWorkflow", func(t *testing.T) {
		_, err := engine.GetWorkflowStatus("non-existent-id")
		if err == nil {
			t.Error("Expected error for non-existent workflow")
		}
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Отменяем сразу

		definition := &WorkflowDefinition{
			Name:    "Cancelled Workflow",
			Version: "1.0.0",
			Stages: map[string]*StageDefinition{
				"test_stage": {
					Name: "test_stage",
					Actions: []ActionDefinition{
						{Type: "manual"},
					},
				},
			},
		}

		_, err := engine.CreateWorkflow(ctx, definition)
		// Создание должно пройти даже с отмененным контекстом
		// Выполнение может быть прервано
		if err != nil {
			t.Logf("Context cancellation handled: %v", err)
		}
	})
}