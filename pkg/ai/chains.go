package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ProjectAnalysis represents the result of project analysis
type ProjectAnalysis struct {
	Description    string                 `json:"description"`
	Complexity     string                 `json:"complexity"`
	EstimatedHours int                    `json:"estimated_hours"`
	Technologies   []string               `json:"technologies"`
	Risks          []string               `json:"risks"`
	Dependencies   []string               `json:"dependencies"`
	Tasks          []TaskSuggestion       `json:"tasks"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// TaskSuggestion represents a suggested task from AI analysis
type TaskSuggestion struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"`
	Type        string   `json:"type"`
	Hours       int      `json:"hours"`
	Tags        []string `json:"tags"`
	Dependencies []string `json:"dependencies"`
}

// ProjectPlan represents a comprehensive project plan
type ProjectPlan struct {
	ID             string           `json:"id"`
	Description    string           `json:"description"`
	ProjectType    string           `json:"project_type"`
	Complexity     string           `json:"complexity"`
	TimelineDays   int              `json:"timeline_days"`
	Priority       string           `json:"priority"`
	Tasks          []TaskSuggestion `json:"tasks"`
	TotalHours     int              `json:"total_hours"`
	CreatedAt      time.Time        `json:"created_at"`
}

// AIChains provides AI-powered analysis and planning capabilities
type AIChains struct {
	hybridClient *HybridAIClient
	mockChains   *MockAIChains
	useMock      bool
	logger       Logger
}

// NewAIChains creates a new AI chains instance
func NewAIChains(gatewayURL, gatewayToken, userID string, userKeys *UserAPIKeys, logger Logger) *AIChains {
	chains := &AIChains{
		hybridClient: NewHybridAIClient(gatewayURL, gatewayToken, userID, userKeys, logger),
		mockChains:   NewMockAIChains(),
		logger:       logger,
	}
	
	// Check if any AI services are available
	validationResults := chains.hybridClient.ValidateUserKeys(context.Background())
	allDown := true
	
	// Check if we have any valid user keys OR gateway connection
	for _, err := range validationResults {
		if err == nil {
			allDown = false
			break
		}
	}
	
	// If no user keys work, try gateway (always assume gateway works unless proven otherwise)
	if allDown && gatewayURL != "" && gatewayToken != "" {
		allDown = false // Assume gateway is available
	}
	
	chains.useMock = allDown
	if allDown {
		logger.Warn("All AI services unavailable, using mock chains")
	} else {
		logger.Info("AI services available, using hybrid client")
	}
	
	return chains
}

// SetPrimaryProvider sets the primary AI provider for chains
func (c *AIChains) SetPrimaryProvider(provider string) {
	// This method is kept for compatibility but routing is now handled by HybridAIClient
	c.logger.Debug("Setting primary provider", "provider", provider)
}

// AnalyzeProject performs comprehensive project analysis using AI
func (c *AIChains) AnalyzeProject(description, projectType string) (*ProjectAnalysis, error) {
	if c.useMock {
		return c.mockChains.AnalyzeProject(description, projectType)
	}
	prompt := fmt.Sprintf(`Analyze the following project and provide a detailed breakdown:

Project Description: %s
Project Type: %s

Please provide a comprehensive analysis in the following JSON format:
{
  "description": "refined project description",
  "complexity": "simple|medium|complex",
  "estimated_hours": number,
  "technologies": ["tech1", "tech2"],
  "risks": ["risk1", "risk2"],
  "dependencies": ["dep1", "dep2"],
  "tasks": [
    {
      "title": "Task title",
      "description": "Detailed task description",
      "priority": "low|medium|high|critical",
      "type": "feature|bugfix|research|testing|deployment",
      "hours": number,
      "tags": ["tag1", "tag2"],
      "dependencies": ["other_task_titles"]
    }
  ]
}

Guidelines:
- Be realistic with time estimates
- Include proper task dependencies
- Consider testing and deployment tasks
- Use appropriate priorities
- Suggest relevant technologies`, description, projectType)

	request := &HybridChatRequest{
		Model:    "gpt-4", // Default model for analysis
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		Temperature: 0.7,
		MaxTokens:   4000,
		Strategy:    RouteUserKeyFirst,
	}

	response, err := c.hybridClient.Chat(context.Background(), request)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze project: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	content := response.Choices[0].Message.Content
	
	// Extract JSON from the response
	jsonContent := extractJSON(content)
	if jsonContent == "" {
		return nil, fmt.Errorf("failed to extract JSON from AI response")
	}

	var analysis ProjectAnalysis
	if err := json.Unmarshal([]byte(jsonContent), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &analysis, nil
}

// CreateProjectPlan creates a comprehensive project plan
func (c *AIChains) CreateProjectPlan(description, projectType, complexity string, timelineDays int, priority string) (*ProjectPlan, error) {
	if c.useMock {
		return c.mockChains.CreateProjectPlan(description, projectType, complexity, timelineDays, priority)
	}
	prompt := fmt.Sprintf(`Create a detailed project plan for the following requirements:

Description: %s
Type: %s
Complexity: %s
Timeline: %d days
Priority: %s

Create a comprehensive plan with the following JSON structure:
{
  "description": "refined description",
  "project_type": "%s",
  "complexity": "%s", 
  "timeline_days": %d,
  "priority": "%s",
  "tasks": [
    {
      "title": "Task title",
      "description": "Detailed description with acceptance criteria",
      "priority": "low|medium|high|critical",
      "type": "planning|design|development|testing|deployment|documentation",
      "hours": estimated_hours,
      "tags": ["relevant", "tags"],
      "dependencies": ["dependent_task_titles"]
    }
  ],
  "total_hours": total_estimated_hours
}

Requirements:
- Break down into 3-8 manageable tasks
- Include proper task sequencing and dependencies
- Realistic time estimates (consider complexity)
- Include planning, development, testing, and deployment phases
- Add relevant tags for task categorization
- Ensure total timeline fits within %d days`, 
		description, projectType, complexity, timelineDays, priority,
		projectType, complexity, timelineDays, priority, timelineDays)

	request := &HybridChatRequest{
		Model:    "gpt-4", // Default model for planning
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		Temperature: 0.5,
		MaxTokens:   3000,
		Strategy:    RouteUserKeyFirst,
	}

	response, err := c.hybridClient.Chat(context.Background(), request)
	if err != nil {
		return nil, fmt.Errorf("failed to create project plan: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	content := response.Choices[0].Message.Content
	
	// Extract JSON from the response
	jsonContent := extractJSON(content)
	if jsonContent == "" {
		return nil, fmt.Errorf("failed to extract JSON from AI response")
	}

	var plan ProjectPlan
	if err := json.Unmarshal([]byte(jsonContent), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Set metadata
	plan.ID = fmt.Sprintf("plan_%d", time.Now().Unix())
	plan.CreatedAt = time.Now()

	return &plan, nil
}

// ExecuteTask performs AI-powered task execution
func (c *AIChains) ExecuteTask(taskTitle, taskDescription, taskType string) (string, error) {
	if c.useMock {
		return c.mockChains.ExecuteTask(taskTitle, taskDescription, taskType)
	}
	var prompt string
	
	switch taskType {
	case "planning":
		prompt = fmt.Sprintf(`Create a detailed plan for the following task:

Task: %s
Description: %s

Provide a step-by-step execution plan including:
- Prerequisites and dependencies
- Detailed steps with time estimates
- Success criteria and validation points
- Potential risks and mitigation strategies
- Resources and tools needed`, taskTitle, taskDescription)

	case "development":
		prompt = fmt.Sprintf(`Provide development guidance for the following task:

Task: %s
Description: %s

Include:
- Technical approach and architecture considerations
- Key implementation steps
- Code structure recommendations
- Testing strategy
- Integration points
- Performance considerations`, taskTitle, taskDescription)

	case "testing":
		prompt = fmt.Sprintf(`Create a comprehensive testing strategy for:

Task: %s
Description: %s

Include:
- Test scenarios and cases
- Unit, integration, and end-to-end tests
- Test data requirements
- Acceptance criteria validation
- Performance and security testing
- Test automation recommendations`, taskTitle, taskDescription)

	default:
		prompt = fmt.Sprintf(`Provide detailed execution guidance for:

Task: %s
Description: %s
Type: %s

Include specific steps, best practices, and success criteria.`, taskTitle, taskDescription, taskType)
	}

	request := &HybridChatRequest{
		Model:    "gpt-4", // Default model for task execution
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		Temperature: 0.6,
		MaxTokens:   2000,
		Strategy:    RouteUserKeyFirst,
	}

	response, err := c.hybridClient.Chat(context.Background(), request)
	if err != nil {
		return "", fmt.Errorf("failed to execute task: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return response.Choices[0].Message.Content, nil
}

// GenerateProgressComment creates an AI-generated progress comment
func (c *AIChains) GenerateProgressComment(taskTitle, currentStatus, progressPercentage string, completedWork []string) (string, error) {
	if c.useMock {
		return c.mockChains.GenerateProgressComment(taskTitle, currentStatus, progressPercentage, completedWork)
	}
	prompt := fmt.Sprintf(`Generate a professional progress comment for the following task:

Task: %s
Current Status: %s
Progress: %s%%
Completed Work: %v

Create a concise but informative progress update that includes:
- Summary of completed work
- Current progress status
- Next steps (if applicable)
- Any blockers or notes

Keep it professional and under 200 words.`, taskTitle, currentStatus, progressPercentage, completedWork)

	request := &HybridChatRequest{
		Model:    "gpt-4", // Default model for progress comments
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		Temperature: 0.4,
		MaxTokens:   300,
		Strategy:    RouteUserKeyFirst,
	}

	response, err := c.hybridClient.Chat(context.Background(), request)
	if err != nil {
		return "", fmt.Errorf("failed to generate progress comment: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return response.Choices[0].Message.Content, nil
}

// AnalyzeCodebase performs codebase analysis for project planning
func (c *AIChains) AnalyzeCodebase(codeFiles []string, projectDescription string) (*ProjectAnalysis, error) {
	if c.useMock {
		return c.mockChains.AnalyzeCodebase(codeFiles, projectDescription)
	}
	// Limit the number of files to analyze to avoid token limits
	maxFiles := 10
	if len(codeFiles) > maxFiles {
		codeFiles = codeFiles[:maxFiles]
	}

	filesContent := strings.Join(codeFiles, "\n\n---\n\n")
	
	prompt := fmt.Sprintf(`Analyze the following codebase and project description to create implementation tasks:

Project Description: %s

Code Files:
%s

Based on the existing codebase structure and the project requirements, provide analysis in JSON format:
{
  "description": "analysis of current codebase and suggested implementation approach",
  "complexity": "simple|medium|complex",
  "estimated_hours": number,
  "technologies": ["detected technologies"],
  "risks": ["potential implementation risks"],
  "dependencies": ["external dependencies needed"],
  "tasks": [
    {
      "title": "Implementation task title",
      "description": "What needs to be implemented",
      "priority": "low|medium|high|critical", 
      "type": "feature|bugfix|refactor|testing",
      "hours": estimated_hours,
      "tags": ["relevant", "tags"],
      "dependencies": ["prerequisite tasks"]
    }
  ]
}

Consider the existing code patterns, architecture, and suggest realistic implementation tasks.`, projectDescription, filesContent)

	request := &HybridChatRequest{
		Model:    "gpt-4", // Default model for codebase analysis
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		Temperature: 0.6,
		MaxTokens:   4000,
		Strategy:    RouteUserKeyFirst,
	}

	response, err := c.hybridClient.Chat(context.Background(), request)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze codebase: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	content := response.Choices[0].Message.Content
	
	// Extract JSON from the response
	jsonContent := extractJSON(content)
	if jsonContent == "" {
		return nil, fmt.Errorf("failed to extract JSON from AI response")
	}

	var analysis ProjectAnalysis
	if err := json.Unmarshal([]byte(jsonContent), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &analysis, nil
}

// HealthCheck checks if AI services are available
func (c *AIChains) HealthCheck() map[string]error {
	if c.useMock {
		return map[string]error{"mock": nil}
	}
	return c.hybridClient.ValidateUserKeys(context.Background())
}

// GetAvailableModels returns available models for the user
func (c *AIChains) GetAvailableModels() *AvailableModels {
	if c.useMock {
		return &AvailableModels{
			Subscription: []ModelInfo{{Name: "mock-model", Provider: "mock", Available: true}},
			UserKeys:     []ModelInfo{},
		}
	}
	return c.hybridClient.GetAvailableModels()
}

// UpdateUserAPIKeys updates user API keys
func (c *AIChains) UpdateUserAPIKeys(userKeys *UserAPIKeys) {
	if !c.useMock {
		c.hybridClient.UpdateUserAPIKeys(userKeys)
		
		// Re-evaluate if we should use mock
		validationResults := c.hybridClient.ValidateUserKeys(context.Background())
		allDown := true
		for _, err := range validationResults {
			if err == nil {
				allDown = false
				break
			}
		}
		
		c.useMock = allDown
		c.logger.Info("Updated user API keys", "use_mock", c.useMock)
	}
}

// GetUsageStats returns usage statistics
func (c *AIChains) GetUsageStats() *UsageStats {
	if c.useMock {
		return &UsageStats{
			UserKeys:     make(map[string]KeyUsageStats),
			Subscription: SubscriptionUsageStats{},
		}
	}
	return c.hybridClient.GetUsageStats()
}

// Helper function to extract JSON from AI response
func extractJSON(content string) string {
	// Try to find JSON block in markdown code blocks
	jsonRegex := regexp.MustCompile("```(?:json)?\n?({[^`]+})\n?```")
	matches := jsonRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Try to find JSON object directly
	jsonRegex = regexp.MustCompile(`({[\s\S]*})`)
	matches = jsonRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Return original if no JSON pattern found
	return content
}