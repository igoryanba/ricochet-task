package context

import (
	"strings"

	"github.com/grik-ai/ricochet-task/pkg/ai"
)

// ProjectAnalyzer анализирует проекты для определения оптимального контекста
type ProjectAnalyzer struct {
	aiChains *ai.AIChains
	logger   Logger
}

// ProjectAnalysis результат анализа проекта
type ProjectAnalysis struct {
	ProjectName     string                 `json:"project_name"`
	ProjectType     string                 `json:"project_type"`     // web, mobile, api, library, etc.
	Framework       string                 `json:"framework"`        // react, go, python, etc.
	Complexity      string                 `json:"complexity"`       // simple, medium, complex
	EstimatedHours  int                    `json:"estimated_hours"`
	RequiredSkills  []string               `json:"required_skills"`
	Dependencies    []string               `json:"dependencies"`
	Architecture    string                 `json:"architecture"`     // monolith, microservices, serverless
	TeamSize        int                    `json:"team_size"`
	Timeline        int                    `json:"timeline"`         // days
	Risks           []Risk                 `json:"risks"`
	Recommendations []string               `json:"recommendations"`
	SuggestedBoard  string                 `json:"suggested_board"`
	Context         *ContextConfig         `json:"context"`
	Confidence      float64                `json:"confidence"`       // 0.0 - 1.0
}

// Risk представляет риск проекта
type Risk struct {
	Type        string  `json:"type"`        // technical, resource, timeline
	Severity    string  `json:"severity"`    // low, medium, high, critical
	Description string  `json:"description"`
	Mitigation  string  `json:"mitigation"`
	Impact      float64 `json:"impact"`      // 0.0 - 1.0
}

// FilePattern паттерн для анализа файлов
type FilePattern struct {
	Extension   string
	Framework   string
	ProjectType string
	Weight      float64
}

// NewProjectAnalyzer создает новый анализатор проектов
func NewProjectAnalyzer(aiChains *ai.AIChains, logger Logger) *ProjectAnalyzer {
	return &ProjectAnalyzer{
		aiChains: aiChains,
		logger:   logger,
	}
}

// AnalyzeProject анализирует проект по описанию или исходному коду
func (pa *ProjectAnalyzer) AnalyzeProject(description string, codebasePath string) (*ProjectAnalysis, error) {
	analysis := &ProjectAnalysis{
		Risks:           make([]Risk, 0),
		Recommendations: make([]string, 0),
		RequiredSkills:  make([]string, 0),
		Dependencies:    make([]string, 0),
	}

	// Анализ описания проекта
	if description != "" {
		if err := pa.analyzeDescription(description, analysis); err != nil {
			pa.logger.Error("Failed to analyze description", err)
		}
	}

	// Анализ кодовой базы если путь предоставлен
	if codebasePath != "" {
		if err := pa.analyzeCodebase(codebasePath, analysis); err != nil {
			pa.logger.Error("Failed to analyze codebase", err)
		}
	}

	// AI-анализ для улучшения точности
	if pa.aiChains != nil {
		if err := pa.enhanceWithAI(description, analysis); err != nil {
			pa.logger.Error("Failed to enhance with AI", err)
		}
	}

	// Генерируем рекомендации по контексту
	pa.generateContextRecommendations(analysis)

	// Рассчитываем уверенность в анализе
	pa.calculateConfidence(analysis)

	pa.logger.Info("Project analysis completed", 
		"type", analysis.ProjectType,
		"complexity", analysis.Complexity,
		"confidence", analysis.Confidence)

	return analysis, nil
}

// analyzeDescription анализирует текстовое описание проекта
func (pa *ProjectAnalyzer) analyzeDescription(description string, analysis *ProjectAnalysis) error {
	desc := strings.ToLower(description)

	// Определение типа проекта по ключевым словам
	projectTypes := map[string][]string{
		"web":        {"website", "web app", "frontend", "backend", "api", "server"},
		"mobile":     {"mobile", "app", "android", "ios", "react native", "flutter"},
		"desktop":    {"desktop", "electron", "gui", "application"},
		"library":    {"library", "package", "module", "sdk", "framework"},
		"devtools":   {"cli", "tool", "utility", "automation", "build"},
		"data":       {"data", "analytics", "ml", "ai", "database", "etl"},
		"game":       {"game", "gaming", "unity", "unreal"},
		"blockchain": {"blockchain", "crypto", "web3", "smart contract"},
	}

	maxScore := 0.0
	for projectType, keywords := range projectTypes {
		score := 0.0
		for _, keyword := range keywords {
			if strings.Contains(desc, keyword) {
				score += 1.0
			}
		}
		if score > maxScore {
			maxScore = score
			analysis.ProjectType = projectType
		}
	}

	// Определение фреймворка
	frameworks := map[string][]string{
		"react":     {"react", "next.js", "nextjs"},
		"vue":       {"vue", "nuxt"},
		"angular":   {"angular"},
		"go":        {"go", "golang", "gin", "echo", "fiber"},
		"python":    {"python", "django", "flask", "fastapi"},
		"node":      {"node", "nodejs", "express"},
		"java":      {"java", "spring", "springboot"},
		"rust":      {"rust", "actix", "warp"},
		"php":       {"php", "laravel", "symfony"},
		"flutter":   {"flutter", "dart"},
		"swift":     {"swift", "ios"},
		"kotlin":    {"kotlin", "android"},
	}

	for framework, keywords := range frameworks {
		for _, keyword := range keywords {
			if strings.Contains(desc, keyword) {
				analysis.Framework = framework
				break
			}
		}
		if analysis.Framework != "" {
			break
		}
	}

	// Определение сложности по ключевым индикаторам
	complexityIndicators := []string{
		"microservices", "distributed", "scalable", "enterprise",
		"machine learning", "ai", "blockchain", "real-time",
		"high load", "multi-tenant", "integration",
	}

	complexityScore := 0
	for _, indicator := range complexityIndicators {
		if strings.Contains(desc, indicator) {
			complexityScore++
		}
	}

	switch {
	case complexityScore >= 3:
		analysis.Complexity = "complex"
		analysis.EstimatedHours = 400 + complexityScore*50
		analysis.TeamSize = 5 + complexityScore
		analysis.Timeline = 60 + complexityScore*10
	case complexityScore >= 1:
		analysis.Complexity = "medium"
		analysis.EstimatedHours = 100 + complexityScore*50
		analysis.TeamSize = 2 + complexityScore
		analysis.Timeline = 21 + complexityScore*7
	default:
		analysis.Complexity = "simple"
		analysis.EstimatedHours = 40
		analysis.TeamSize = 1
		analysis.Timeline = 7
	}

	// Извлечение названия проекта
	lines := strings.Split(description, "\n")
	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])
		if len(firstLine) > 0 && len(firstLine) < 100 {
			analysis.ProjectName = firstLine
		}
	}

	return nil
}

// analyzeCodebase анализирует существующую кодовую базу
func (pa *ProjectAnalyzer) analyzeCodebase(codebasePath string, analysis *ProjectAnalysis) error {
	patterns := []FilePattern{
		{".js", "javascript", "web", 1.0},
		{".jsx", "react", "web", 1.5},
		{".ts", "typescript", "web", 1.2},
		{".tsx", "react", "web", 1.5},
		{".vue", "vue", "web", 1.5},
		{".go", "go", "api", 1.3},
		{".py", "python", "api", 1.1},
		{".java", "java", "api", 1.2},
		{".rs", "rust", "api", 1.3},
		{".php", "php", "web", 1.0},
		{".swift", "swift", "mobile", 1.4},
		{".kt", "kotlin", "mobile", 1.4},
		{".dart", "flutter", "mobile", 1.4},
		{".cpp", "cpp", "desktop", 1.2},
		{".cs", "csharp", "desktop", 1.2},
	}

	// TODO: Реализовать обход файловой системы и подсчет файлов
	// Пока используем заглушку
	_ = patterns // Временно, чтобы избежать ошибки компиляции
	pa.logger.Debug("Codebase analysis placeholder", "path", codebasePath)

	// Заглушка для анализа файлов
	if strings.Contains(codebasePath, "react") {
		analysis.Framework = "react"
		analysis.ProjectType = "web"
	} else if strings.Contains(codebasePath, "go") {
		analysis.Framework = "go"
		analysis.ProjectType = "api"
	}

	return nil
}

// enhanceWithAI использует AI для улучшения анализа
func (pa *ProjectAnalyzer) enhanceWithAI(description string, analysis *ProjectAnalysis) error {
	if pa.aiChains == nil {
		return nil
	}

	// AI-анализ для оценки рисков
	risks, err := pa.analyzeRisks(description, analysis)
	if err != nil {
		pa.logger.Error("Failed to analyze risks with AI", err)
	} else {
		analysis.Risks = risks
	}

	// AI-рекомендации по архитектуре
	architecture, err := pa.suggestArchitecture(description, analysis)
	if err != nil {
		pa.logger.Error("Failed to suggest architecture with AI", err)
	} else {
		analysis.Architecture = architecture
	}

	// AI-оценка необходимых навыков
	skills, err := pa.identifyRequiredSkills(description, analysis)
	if err != nil {
		pa.logger.Error("Failed to identify required skills with AI", err)
	} else {
		analysis.RequiredSkills = skills
	}

	return nil
}

// analyzeRisks анализирует потенциальные риски проекта
func (pa *ProjectAnalyzer) analyzeRisks(description string, analysis *ProjectAnalysis) ([]Risk, error) {
	risks := []Risk{}

	// Технические риски
	if analysis.Complexity == "complex" {
		risks = append(risks, Risk{
			Type:        "technical",
			Severity:    "high",
			Description: "High complexity may lead to architecture challenges",
			Mitigation:  "Detailed technical design and prototyping phase",
			Impact:      0.7,
		})
	}

	// Ресурсные риски
	if analysis.TeamSize > 5 {
		risks = append(risks, Risk{
			Type:        "resource",
			Severity:    "medium",
			Description: "Large team coordination overhead",
			Mitigation:  "Clear communication protocols and team leads",
			Impact:      0.5,
		})
	}

	// Временные риски
	if analysis.Timeline > 90 {
		risks = append(risks, Risk{
			Type:        "timeline",
			Severity:    "high",
			Description: "Long timeline increases scope creep risk",
			Mitigation:  "Regular milestone reviews and scope control",
			Impact:      0.6,
		})
	}

	return risks, nil
}

// suggestArchitecture предлагает архитектуру на основе анализа
func (pa *ProjectAnalyzer) suggestArchitecture(description string, analysis *ProjectAnalysis) (string, error) {
	desc := strings.ToLower(description)

	// Микросервисы для сложных систем
	if analysis.Complexity == "complex" || strings.Contains(desc, "microservices") {
		return "microservices", nil
	}

	// Serverless для простых API
	if analysis.ProjectType == "api" && analysis.Complexity == "simple" {
		return "serverless", nil
	}

	// Монолит по умолчанию
	return "monolith", nil
}

// identifyRequiredSkills определяет необходимые навыки команды
func (pa *ProjectAnalyzer) identifyRequiredSkills(description string, analysis *ProjectAnalysis) ([]string, error) {
	skills := []string{}

	// Базовые навыки по фреймворку
	frameworkSkills := map[string][]string{
		"react":   {"javascript", "react", "css", "html"},
		"go":      {"go", "sql", "rest-api"},
		"python":  {"python", "sql", "rest-api"},
		"flutter": {"dart", "flutter", "mobile-ui"},
	}

	if frameSkills, exists := frameworkSkills[analysis.Framework]; exists {
		skills = append(skills, frameSkills...)
	}

	// Дополнительные навыки по типу проекта
	projectSkills := map[string][]string{
		"web":     {"frontend", "backend", "database"},
		"mobile":  {"mobile-development", "ui-ux"},
		"api":     {"backend", "database", "devops"},
		"library": {"software-architecture", "documentation"},
	}

	if projSkills, exists := projectSkills[analysis.ProjectType]; exists {
		skills = append(skills, projSkills...)
	}

	// Навыки для сложности
	if analysis.Complexity == "complex" {
		skills = append(skills, "architecture", "system-design", "performance-optimization")
	}

	return pa.removeDuplicates(skills), nil
}

// generateContextRecommendations генерирует рекомендации по контексту
func (pa *ProjectAnalyzer) generateContextRecommendations(analysis *ProjectAnalysis) {
	// Рекомендуемые настройки контекста
	analysis.Context = &ContextConfig{
		ProjectType:     analysis.ProjectType,
		Complexity:      analysis.Complexity,
		Timeline:        analysis.Timeline,
		TeamSize:        analysis.TeamSize,
		AIEnabled:       true,
		AutoAssignment:  analysis.TeamSize > 1,
		AutoProgress:    analysis.Complexity != "complex",
		CustomFields: map[string]interface{}{
			"framework":    analysis.Framework,
			"architecture": analysis.Architecture,
			"risks_count":  len(analysis.Risks),
		},
	}

	// Рекомендуемая доска на основе типа проекта
	boardRecommendations := map[string]string{
		"web":        "Web Development",
		"mobile":     "Mobile Development", 
		"api":        "Backend Development",
		"library":    "Library Development",
		"devtools":   "DevOps & Tools",
		"data":       "Data & Analytics",
		"game":       "Game Development",
		"blockchain": "Blockchain Development",
	}

	if board, exists := boardRecommendations[analysis.ProjectType]; exists {
		analysis.SuggestedBoard = board
	} else {
		analysis.SuggestedBoard = "General Development"
	}

	// Рекомендации по процессу
	if analysis.Complexity == "simple" {
		analysis.Recommendations = append(analysis.Recommendations, 
			"Используйте kanban workflow для простых задач",
			"Включите автоматическое обновление прогресса")
	} else if analysis.Complexity == "complex" {
		analysis.Recommendations = append(analysis.Recommendations,
			"Разбейте проект на несколько эпиков",
			"Используйте подробное планирование спринтов",
			"Настройте регулярные code review")
	}

	// Рекомендации по команде
	if analysis.TeamSize == 1 {
		analysis.Recommendations = append(analysis.Recommendations,
			"Включите AI-ассистированное выполнение задач")
	} else if analysis.TeamSize > 3 {
		analysis.Recommendations = append(analysis.Recommendations,
			"Настройте автоматическое назначение задач",
			"Используйте балансировку нагрузки команды")
	}
}

// calculateConfidence рассчитывает уверенность в анализе
func (pa *ProjectAnalyzer) calculateConfidence(analysis *ProjectAnalysis) {
	confidence := 0.5 // базовая уверенность

	// Увеличиваем уверенность если есть четкие индикаторы
	if analysis.Framework != "" {
		confidence += 0.2
	}
	
	if analysis.ProjectType != "" {
		confidence += 0.2
	}
	
	if len(analysis.RequiredSkills) > 0 {
		confidence += 0.1
	}
	
	// Уменьшаем уверенность для сложных проектов
	if analysis.Complexity == "complex" {
		confidence -= 0.1
	}

	// Нормализуем в диапазон 0.0-1.0
	if confidence > 1.0 {
		confidence = 1.0
	} else if confidence < 0.0 {
		confidence = 0.0
	}

	analysis.Confidence = confidence
}

// removeDuplicates удаляет дубликаты из слайса строк
func (pa *ProjectAnalyzer) removeDuplicates(items []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// SuggestWorkflowType предлагает тип workflow на основе анализа
func (pa *ProjectAnalyzer) SuggestWorkflowType(analysis *ProjectAnalysis) string {
	// Agile для большинства проектов
	if analysis.TeamSize > 2 && analysis.Timeline > 14 {
		return "agile"
	}
	
	// Kanban для простых проектов
	if analysis.Complexity == "simple" {
		return "kanban"
	}
	
	// Waterfall для четко определенных проектов
	if analysis.ProjectType == "library" {
		return "waterfall"
	}
	
	return "agile"
}

// EstimateTaskBreakdown оценивает разбивку проекта на задачи
func (pa *ProjectAnalyzer) EstimateTaskBreakdown(analysis *ProjectAnalysis) map[string]int {
	breakdown := make(map[string]int)
	
	totalHours := analysis.EstimatedHours
	
	// Распределение по типам задач в процентах
	distributions := map[string]map[string]float64{
		"simple": {
			"planning":       0.10,
			"development":    0.70,
			"testing":        0.15,
			"documentation":  0.05,
		},
		"medium": {
			"planning":       0.15,
			"development":    0.60,
			"testing":        0.20,
			"documentation":  0.05,
		},
		"complex": {
			"planning":       0.20,
			"design":         0.15,
			"development":    0.45,
			"testing":        0.15,
			"documentation":  0.05,
		},
	}
	
	if dist, exists := distributions[analysis.Complexity]; exists {
		for taskType, percentage := range dist {
			breakdown[taskType] = int(float64(totalHours) * percentage)
		}
	}
	
	return breakdown
}