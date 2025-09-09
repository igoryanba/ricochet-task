package ai

import (
	"fmt"
	"time"
)

// MockAIChains provides mock AI functionality for testing when model services are not available
type MockAIChains struct{}

// NewMockAIChains creates a new mock AI chains instance
func NewMockAIChains() *MockAIChains {
	return &MockAIChains{}
}

// AnalyzeProject performs mock project analysis
func (m *MockAIChains) AnalyzeProject(description, projectType string) (*ProjectAnalysis, error) {
	// Determine complexity based on description keywords
	complexity := "medium"
	if containsAny(description, []string{"simple", "basic", "quick"}) {
		complexity = "simple"
	} else if containsAny(description, []string{"complex", "advanced", "enterprise", "scalable"}) {
		complexity = "complex"
	}

	// Generate realistic tasks based on project type and description
	var tasks []TaskSuggestion
	baseHours := 40

	switch projectType {
	case "feature", "mobile_app", "web_app":
		tasks = []TaskSuggestion{
			{
				Title:       "Requirements Analysis",
				Description: fmt.Sprintf("Analyze requirements for %s", description),
				Priority:    "high",
				Type:        "planning",
				Hours:       4,
				Tags:        []string{"planning", "requirements"},
			},
			{
				Title:       "Design & Architecture",
				Description: "Create technical design and architecture",
				Priority:    "high",
				Type:        "design",
				Hours:       8,
				Tags:        []string{"design", "architecture"},
				Dependencies: []string{"Requirements Analysis"},
			},
			{
				Title:       "Core Implementation",
				Description: fmt.Sprintf("Implement core functionality for %s", description),
				Priority:    "high",
				Type:        "development",
				Hours:       baseHours,
				Tags:        []string{"development", "core"},
				Dependencies: []string{"Design & Architecture"},
			},
			{
				Title:       "Testing & QA",
				Description: "Create and execute comprehensive tests",
				Priority:    "high",
				Type:        "testing",
				Hours:       8,
				Tags:        []string{"testing", "qa"},
				Dependencies: []string{"Core Implementation"},
			},
			{
				Title:       "Documentation",
				Description: "Create user and technical documentation",
				Priority:    "medium",
				Type:        "documentation",
				Hours:       4,
				Tags:        []string{"documentation"},
				Dependencies: []string{"Testing & QA"},
			},
		}
	case "bugfix":
		tasks = []TaskSuggestion{
			{
				Title:       "Bug Investigation",
				Description: fmt.Sprintf("Investigate and reproduce bug: %s", description),
				Priority:    "critical",
				Type:        "bugfix",
				Hours:       4,
				Tags:        []string{"investigation", "bugfix"},
			},
			{
				Title:       "Fix Implementation",
				Description: "Implement the bug fix",
				Priority:    "critical",
				Type:        "bugfix",
				Hours:       8,
				Tags:        []string{"bugfix", "implementation"},
				Dependencies: []string{"Bug Investigation"},
			},
			{
				Title:       "Regression Testing",
				Description: "Test fix and ensure no regressions",
				Priority:    "high",
				Type:        "testing",
				Hours:       4,
				Tags:        []string{"testing", "regression"},
				Dependencies: []string{"Fix Implementation"},
			},
		}
	}

	// Adjust hours based on complexity
	multiplier := 1.0
	switch complexity {
	case "simple":
		multiplier = 0.7
	case "complex":
		multiplier = 1.5
	}

	totalHours := 0
	for i := range tasks {
		tasks[i].Hours = int(float64(tasks[i].Hours) * multiplier)
		totalHours += tasks[i].Hours
	}

	return &ProjectAnalysis{
		Description:    fmt.Sprintf("Enhanced project: %s", description),
		Complexity:     complexity,
		EstimatedHours: totalHours,
		Technologies:   detectTechnologies(description),
		Risks:          generateRisks(complexity),
		Dependencies:   generateDependencies(projectType),
		Tasks:          tasks,
		Metadata: map[string]interface{}{
			"generated_at": time.Now().Unix(),
			"mock_version": "1.0",
		},
	}, nil
}

// CreateProjectPlan creates a mock comprehensive project plan
func (m *MockAIChains) CreateProjectPlan(description, projectType, complexity string, timelineDays int, priority string) (*ProjectPlan, error) {
	analysis, err := m.AnalyzeProject(description, projectType)
	if err != nil {
		return nil, err
	}

	return &ProjectPlan{
		ID:           fmt.Sprintf("plan_%d", time.Now().Unix()),
		Description:  analysis.Description,
		ProjectType:  projectType,
		Complexity:   complexity,
		TimelineDays: timelineDays,
		Priority:     priority,
		Tasks:        analysis.Tasks,
		TotalHours:   analysis.EstimatedHours,
		CreatedAt:    time.Now(),
	}, nil
}

// ExecuteTask performs mock AI-powered task execution
func (m *MockAIChains) ExecuteTask(taskTitle, taskDescription, taskType string) (string, error) {
	var result string

	switch taskType {
	case "planning":
		result = fmt.Sprintf(`## Implementation Plan for: %s

### Overview
%s

### Execution Steps
1. **Requirements Gathering**
   - Review existing documentation
   - Identify stakeholders and gather requirements
   - Create acceptance criteria

2. **Technical Planning**
   - Architecture review and design decisions
   - Identify required technologies and tools
   - Plan integration points

3. **Implementation Roadmap**
   - Break down into smaller, manageable tasks
   - Establish dependencies and critical path
   - Set milestones and checkpoints

4. **Risk Assessment**
   - Identify potential blockers
   - Plan mitigation strategies
   - Set up fallback options

### Success Criteria
- Clear requirements documented
- Technical approach validated
- Implementation roadmap approved
- Risk mitigation plans in place

### Next Steps
- Begin detailed implementation phase
- Set up development environment
- Create initial project structure`, taskTitle, taskDescription)

	case "development":
		result = fmt.Sprintf(`## Development Guide for: %s

### Technical Approach
%s

### Implementation Strategy
1. **Setup & Environment**
   - Configure development environment
   - Set up project structure and dependencies
   - Initialize version control and CI/CD

2. **Core Development**
   - Implement main functionality
   - Follow established coding standards
   - Write unit tests alongside code

3. **Integration**
   - Integrate with existing systems
   - Test integration points thoroughly
   - Handle error scenarios gracefully

4. **Performance Optimization**
   - Profile and optimize critical paths
   - Implement caching where appropriate
   - Ensure scalability requirements are met

### Code Quality Guidelines
- Follow established coding standards
- Maintain test coverage above 80%%
- Use meaningful variable and function names
- Document complex logic and APIs

### Testing Strategy
- Unit tests for all business logic
- Integration tests for external dependencies
- End-to-end tests for critical user journeys
- Performance tests for scalability validation`, taskTitle, taskDescription)

	case "testing":
		result = fmt.Sprintf(`## Testing Strategy for: %s

### Test Scope
%s

### Test Types
1. **Unit Tests**
   - Test individual functions and methods
   - Mock external dependencies
   - Aim for >90%% code coverage

2. **Integration Tests**
   - Test component interactions
   - Validate API contracts
   - Test database operations

3. **End-to-End Tests**
   - Test complete user workflows
   - Validate UI/UX functionality
   - Test critical business scenarios

4. **Performance Tests**
   - Load testing under normal conditions
   - Stress testing under peak load
   - Endurance testing for long-running operations

### Test Data Management
- Create realistic test datasets
- Implement data cleanup procedures
- Use test data factories for consistency

### Automation Strategy
- Integrate tests into CI/CD pipeline
- Automated test execution on code changes
- Regular test suite maintenance and updates

### Success Criteria
- All tests pass consistently
- Test coverage meets requirements
- Performance benchmarks achieved
- No critical defects in production`, taskTitle, taskDescription)

	default:
		result = fmt.Sprintf(`## Execution Plan for: %s

### Task Description
%s

### Approach
This task will be executed following industry best practices:

1. **Analysis Phase**
   - Understand requirements thoroughly
   - Identify potential challenges and solutions
   - Plan resource allocation

2. **Implementation Phase**
   - Execute planned approach
   - Monitor progress regularly
   - Adapt to changing requirements

3. **Validation Phase**
   - Test implemented solution
   - Gather feedback from stakeholders
   - Make necessary adjustments

4. **Delivery Phase**
   - Deploy to appropriate environment
   - Document solution and processes
   - Hand over to operations team

### Quality Assurance
- Regular code reviews
- Comprehensive testing
- Documentation updates
- Stakeholder approval

### Timeline
- Analysis: 1-2 days
- Implementation: 3-5 days
- Validation: 1-2 days
- Delivery: 1 day

### Deliverables
- Working solution meeting requirements
- Updated documentation
- Test coverage reports
- Deployment instructions`, taskTitle, taskDescription)
	}

	return result, nil
}

// AnalyzeCodebase performs mock codebase analysis
func (m *MockAIChains) AnalyzeCodebase(codeFiles []string, projectDescription string) (*ProjectAnalysis, error) {
	// Analyze file extensions to detect technologies
	technologies := []string{}
	hasGo := false
	hasJS := false
	hasPython := false

	for _, file := range codeFiles {
		if containsAny(file, []string{".go"}) {
			hasGo = true
		}
		if containsAny(file, []string{".js", ".ts", ".jsx", ".tsx"}) {
			hasJS = true
		}
		if containsAny(file, []string{".py"}) {
			hasPython = true
		}
	}

	if hasGo {
		technologies = append(technologies, "Go", "REST API")
	}
	if hasJS {
		technologies = append(technologies, "JavaScript", "Node.js", "React")
	}
	if hasPython {
		technologies = append(technologies, "Python", "FastAPI")
	}

	// Generate tasks based on codebase analysis
	tasks := []TaskSuggestion{
		{
			Title:       "Code Architecture Review",
			Description: "Review existing code architecture and patterns",
			Priority:    "high",
			Type:        "analysis",
			Hours:       6,
			Tags:        []string{"architecture", "review"},
		},
		{
			Title:       "Refactoring Implementation",
			Description: fmt.Sprintf("Implement changes for: %s", projectDescription),
			Priority:    "high",
			Type:        "development",
			Hours:       20,
			Tags:        []string{"refactoring", "implementation"},
			Dependencies: []string{"Code Architecture Review"},
		},
		{
			Title:       "Unit Test Coverage",
			Description: "Improve test coverage for modified components",
			Priority:    "medium",
			Type:        "testing",
			Hours:       8,
			Tags:        []string{"testing", "coverage"},
			Dependencies: []string{"Refactoring Implementation"},
		},
	}

	return &ProjectAnalysis{
		Description:    fmt.Sprintf("Codebase enhancement: %s", projectDescription),
		Complexity:     "medium",
		EstimatedHours: 34,
		Technologies:   technologies,
		Risks:          []string{"Breaking existing functionality", "Integration complexity"},
		Dependencies:   []string{"Existing codebase", "Test infrastructure"},
		Tasks:          tasks,
	}, nil
}

// GenerateProgressComment creates a mock progress comment
func (m *MockAIChains) GenerateProgressComment(taskTitle, currentStatus, progressPercentage string, completedWork []string) (string, error) {
	templates := []string{
		"Task '%s' is progressing well at %s%%. Current status: %s. Recent work includes: %v. Next steps involve finalizing implementation details.",
		"Good progress on '%s' - now at %s%% completion. Status updated to '%s'. Completed items: %v. Moving forward with testing phase.",
		"'%s' advancing as planned (%s%% complete). Current phase: %s. Recent achievements: %v. Preparing for next milestone.",
	}

	template := templates[len(taskTitle)%len(templates)]
	comment := fmt.Sprintf(template, taskTitle, progressPercentage, currentStatus, completedWork)

	return comment, nil
}

// Helper functions

func containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if len(text) >= len(keyword) {
			for i := 0; i <= len(text)-len(keyword); i++ {
				if text[i:i+len(keyword)] == keyword {
					return true
				}
			}
		}
	}
	return false
}

func detectTechnologies(description string) []string {
	technologies := []string{}
	
	if containsAny(description, []string{"api", "rest", "backend"}) {
		technologies = append(technologies, "REST API", "Backend")
	}
	if containsAny(description, []string{"frontend", "ui", "interface"}) {
		technologies = append(technologies, "Frontend", "UI/UX")
	}
	if containsAny(description, []string{"database", "data", "storage"}) {
		technologies = append(technologies, "Database", "Data Storage")
	}
	if containsAny(description, []string{"auth", "login", "security"}) {
		technologies = append(technologies, "Authentication", "Security")
	}
	if containsAny(description, []string{"notification", "message", "email"}) {
		technologies = append(technologies, "Notifications", "Messaging")
	}

	if len(technologies) == 0 {
		technologies = []string{"Web Development", "Software Engineering"}
	}

	return technologies
}

func generateRisks(complexity string) []string {
	baseRisks := []string{"Timeline constraints", "Resource availability"}
	
	switch complexity {
	case "simple":
		return append(baseRisks, "Limited scope creep")
	case "complex":
		return append(baseRisks, "Technical complexity", "Integration challenges", "Scalability concerns")
	default:
		return append(baseRisks, "Technical dependencies")
	}
}

func generateDependencies(projectType string) []string {
	switch projectType {
	case "feature":
		return []string{"Existing codebase", "API availability", "Database schema"}
	case "bugfix":
		return []string{"Bug reproduction", "Test environment", "Code access"}
	default:
		return []string{"Development environment", "Tool availability"}
	}
}