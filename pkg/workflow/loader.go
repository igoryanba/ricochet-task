package workflow

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// WorkflowLoader загружает workflow определения из YAML файлов
type WorkflowLoader struct {
	workflowsDir string
	logger       Logger
}

// NewWorkflowLoader создает новый загрузчик workflow
func NewWorkflowLoader(workflowsDir string, logger Logger) *WorkflowLoader {
	return &WorkflowLoader{
		workflowsDir: workflowsDir,
		logger:       logger,
	}
}

// LoadWorkflow загружает workflow из файла
func (wl *WorkflowLoader) LoadWorkflow(filename string) (*WorkflowDefinition, error) {
	fullPath := filepath.Join(wl.workflowsDir, filename)
	
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open workflow file %s: %w", fullPath, err)
	}
	defer file.Close()
	
	return wl.LoadWorkflowFromReader(file, filename)
}

// LoadWorkflowFromReader загружает workflow из Reader
func (wl *WorkflowLoader) LoadWorkflowFromReader(reader io.Reader, name string) (*WorkflowDefinition, error) {
	var workflow WorkflowDefinition
	
	decoder := yaml.NewDecoder(reader)
	if err := decoder.Decode(&workflow); err != nil {
		return nil, fmt.Errorf("failed to parse workflow YAML %s: %w", name, err)
	}
	
	// Валидация workflow
	if err := wl.validateWorkflow(&workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed for %s: %w", name, err)
	}
	
	// Постобработка workflow
	wl.postProcessWorkflow(&workflow)
	
	wl.logger.Info("Workflow loaded successfully", "name", workflow.Name, "version", workflow.Version)
	
	return &workflow, nil
}

// LoadAllWorkflows загружает все workflow из директории
func (wl *WorkflowLoader) LoadAllWorkflows() (map[string]*WorkflowDefinition, error) {
	workflows := make(map[string]*WorkflowDefinition)
	
	files, err := os.ReadDir(wl.workflowsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflows directory %s: %w", wl.workflowsDir, err)
	}
	
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yml") {
			continue
		}
		
		workflow, err := wl.LoadWorkflow(file.Name())
		if err != nil {
			wl.logger.Error("Failed to load workflow", err, "file", file.Name())
			continue
		}
		
		workflows[workflow.Name] = workflow
	}
	
	wl.logger.Info("Loaded workflows", "count", len(workflows))
	return workflows, nil
}

// validateWorkflow валидирует workflow определение
func (wl *WorkflowLoader) validateWorkflow(workflow *WorkflowDefinition) error {
	if workflow.Name == "" {
		return fmt.Errorf("workflow name is required")
	}
	
	if workflow.Version == "" {
		return fmt.Errorf("workflow version is required")
	}
	
	if len(workflow.Stages) == 0 {
		return fmt.Errorf("workflow must have at least one stage")
	}
	
	// Валидация этапов
	stageNames := make(map[string]bool)
	for stageName, stage := range workflow.Stages {
		if stage == nil {
			return fmt.Errorf("stage %s is nil", stageName)
		}
		
		if stage.Name == "" {
			stage.Name = stageName
		}
		
		stageNames[stageName] = true
		
		// Валидация зависимостей
		for _, dep := range stage.Dependencies {
			if !stageNames[dep] && !wl.stageExistsInWorkflow(dep, workflow) {
				return fmt.Errorf("stage %s depends on non-existent stage %s", stageName, dep)
			}
		}
		
		// Валидация действий
		for i, action := range stage.Actions {
			if action.Type == "" {
				return fmt.Errorf("action %d in stage %s has no type", i, stageName)
			}
		}
		
		// Валидация условий
		for i, condition := range stage.Conditions {
			if condition.Field == "" {
				return fmt.Errorf("condition %d in stage %s has no field", i, stageName)
			}
			if condition.Operator == "" {
				return fmt.Errorf("condition %d in stage %s has no operator", i, stageName)
			}
		}
	}
	
	// Валидация триггеров
	for i, trigger := range workflow.Triggers {
		if trigger.Type == "" {
			return fmt.Errorf("trigger %d has no type", i)
		}
	}
	
	// Проверка циклических зависимостей
	if err := wl.checkCyclicDependencies(workflow); err != nil {
		return fmt.Errorf("cyclic dependencies detected: %w", err)
	}
	
	return nil
}

// postProcessWorkflow выполняет постобработку workflow
func (wl *WorkflowLoader) postProcessWorkflow(workflow *WorkflowDefinition) {
	// Устанавливаем значения по умолчанию
	if workflow.Settings.DefaultTimeout == 0 {
		workflow.Settings.DefaultTimeout = 24 * 60 * 60 * 1000000000 // 24 часа в наносекундах
	}
	
	if workflow.Settings.MaxConcurrency == 0 {
		workflow.Settings.MaxConcurrency = 10
	}
	
	if workflow.Settings.LogLevel == "" {
		workflow.Settings.LogLevel = "info"
	}
	
	// Устанавливаем значения по умолчанию для этапов
	for _, stage := range workflow.Stages {
		if stage.Timeout == 0 {
			stage.Timeout = workflow.Settings.DefaultTimeout
		}
		
		if stage.AutoAssign.Strategy == "" {
			stage.AutoAssign.Strategy = "round_robin"
		}
		
		if stage.AutoAssign.MaxLoad == 0 {
			stage.AutoAssign.MaxLoad = 5
		}
		
		if stage.Completion.Trigger == "" {
			stage.Completion.Trigger = "manual"
		}
	}
	
	// Инициализация переменных по умолчанию
	if workflow.Variables == nil {
		workflow.Variables = make(map[string]interface{})
	}
}

// stageExistsInWorkflow проверяет существование этапа в workflow
func (wl *WorkflowLoader) stageExistsInWorkflow(stageName string, workflow *WorkflowDefinition) bool {
	_, exists := workflow.Stages[stageName]
	return exists
}

// checkCyclicDependencies проверяет циклические зависимости между этапами
func (wl *WorkflowLoader) checkCyclicDependencies(workflow *WorkflowDefinition) error {
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)
	
	for stageName := range workflow.Stages {
		if !visited[stageName] {
			if wl.hasCyclicDependency(stageName, workflow, visited, recursionStack) {
				return fmt.Errorf("cyclic dependency detected involving stage %s", stageName)
			}
		}
	}
	
	return nil
}

// hasCyclicDependency выполняет DFS для поиска циклических зависимостей
func (wl *WorkflowLoader) hasCyclicDependency(stageName string, workflow *WorkflowDefinition, visited, recursionStack map[string]bool) bool {
	visited[stageName] = true
	recursionStack[stageName] = true
	
	stage := workflow.Stages[stageName]
	if stage == nil {
		return false
	}
	
	for _, dep := range stage.Dependencies {
		if !visited[dep] {
			if wl.hasCyclicDependency(dep, workflow, visited, recursionStack) {
				return true
			}
		} else if recursionStack[dep] {
			return true
		}
	}
	
	recursionStack[stageName] = false
	return false
}

// SaveWorkflow сохраняет workflow в YAML файл
func (wl *WorkflowLoader) SaveWorkflow(workflow *WorkflowDefinition, filename string) error {
	fullPath := filepath.Join(wl.workflowsDir, filename)
	
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create workflow file %s: %w", fullPath, err)
	}
	defer file.Close()
	
	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	
	if err := encoder.Encode(workflow); err != nil {
		return fmt.Errorf("failed to encode workflow to YAML: %w", err)
	}
	
	wl.logger.Info("Workflow saved successfully", "name", workflow.Name, "file", fullPath)
	return nil
}

// CreateWorkflowsDirectory создает директорию для workflow, если её нет
func (wl *WorkflowLoader) CreateWorkflowsDirectory() error {
	if err := os.MkdirAll(wl.workflowsDir, 0755); err != nil {
		return fmt.Errorf("failed to create workflows directory %s: %w", wl.workflowsDir, err)
	}
	return nil
}

// GetWorkflowsDirectory возвращает путь к директории workflow
func (wl *WorkflowLoader) GetWorkflowsDirectory() string {
	return wl.workflowsDir
}

// ValidateWorkflowSyntax валидирует только синтаксис YAML без полной валидации
func (wl *WorkflowLoader) ValidateWorkflowSyntax(reader io.Reader) error {
	var workflow WorkflowDefinition
	
	decoder := yaml.NewDecoder(reader)
	if err := decoder.Decode(&workflow); err != nil {
		return fmt.Errorf("invalid YAML syntax: %w", err)
	}
	
	return nil
}

// ListWorkflowFiles возвращает список всех YAML файлов в директории workflow
func (wl *WorkflowLoader) ListWorkflowFiles() ([]string, error) {
	var workflowFiles []string
	
	files, err := os.ReadDir(wl.workflowsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflows directory: %w", err)
	}
	
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		if strings.HasSuffix(file.Name(), ".yaml") || strings.HasSuffix(file.Name(), ".yml") {
			workflowFiles = append(workflowFiles, file.Name())
		}
	}
	
	return workflowFiles, nil
}