package workflow

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// CommitPatterns анализатор паттернов в коммитах и именах веток
type CommitPatterns struct {
	taskIDPatterns    []*regexp.Regexp
	featurePatterns   []*regexp.Regexp
	fixPatterns       []*regexp.Regexp
	testPatterns      []*regexp.Regexp
	docPatterns       []*regexp.Regexp
	releasePatterns   []*regexp.Regexp
	branchPatterns    []*regexp.Regexp
	pathPatterns      []*regexp.Regexp
}

// NewCommitPatterns создает новый анализатор паттернов
func NewCommitPatterns() *CommitPatterns {
	cp := &CommitPatterns{}
	cp.initializePatterns()
	return cp
}

// initializePatterns инициализирует регулярные выражения
func (cp *CommitPatterns) initializePatterns() {
	// Паттерны для извлечения ID задач
	cp.taskIDPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)#(\d+)`),                           // #123
		regexp.MustCompile(`(?i)task[:\-\s]*(\d+)`),               // task: 123, task-123, task 123
		regexp.MustCompile(`(?i)issue[:\-\s]*(\d+)`),              // issue: 123
		regexp.MustCompile(`(?i)ticket[:\-\s]*(\d+)`),             // ticket: 123
		regexp.MustCompile(`(?i)bug[:\-\s]*(\d+)`),                // bug: 123
		regexp.MustCompile(`(?i)feature[:\-\s]*(\d+)`),            // feature: 123
		regexp.MustCompile(`(?i)\b([A-Z]+-\d+)\b`),                // PROJ-123
		regexp.MustCompile(`(?i)\b(WF-\d+)\b`),                    // WF-123 (workflow)
		regexp.MustCompile(`(?i)\b(TASK-\d+)\b`),                  // TASK-123
		regexp.MustCompile(`(?i)closes?\s*#(\d+)`),                // closes #123
		regexp.MustCompile(`(?i)fixes?\s*#(\d+)`),                 // fixes #123
		regexp.MustCompile(`(?i)resolves?\s*#(\d+)`),              // resolves #123
	}
	
	// Паттерны для определения типа коммита - фичи
	cp.featurePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^feat(\([^)]+\))?:`),              // feat: или feat(scope):
		regexp.MustCompile(`(?i)^feature(\([^)]+\))?:`),           // feature:
		regexp.MustCompile(`(?i)^add(\([^)]+\))?:`),               // add:
		regexp.MustCompile(`(?i)^implement(\([^)]+\))?:`),         // implement:
		regexp.MustCompile(`(?i)^create(\([^)]+\))?:`),            // create:
		regexp.MustCompile(`(?i)\b(new feature|добавить|создать|реализовать)\b`),
		regexp.MustCompile(`(?i)^✨`),                             // sparkles emoji
	}
	
	// Паттерны для определения типа коммита - исправления
	cp.fixPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^fix(\([^)]+\))?:`),               // fix: или fix(scope):
		regexp.MustCompile(`(?i)^bugfix(\([^)]+\))?:`),            // bugfix:
		regexp.MustCompile(`(?i)^hotfix(\([^)]+\))?:`),            // hotfix:
		regexp.MustCompile(`(?i)^patch(\([^)]+\))?:`),             // patch:
		regexp.MustCompile(`(?i)\b(bug fix|исправить|фикс|баг)\b`),
		regexp.MustCompile(`(?i)^🐛`),                             // bug emoji
	}
	
	// Паттерны для определения типа коммита - тесты
	cp.testPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^test(\([^)]+\))?:`),              // test:
		regexp.MustCompile(`(?i)^tests(\([^)]+\))?:`),             // tests:
		regexp.MustCompile(`(?i)\b(add test|unit test|integration test|тест)\b`),
		regexp.MustCompile(`(?i)^✅`),                             // check mark emoji
		regexp.MustCompile(`(?i)\bspec\b`),                        // spec files
	}
	
	// Паттерны для определения типа коммита - документация
	cp.docPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^docs?(\([^)]+\))?:`),             // doc: или docs:
		regexp.MustCompile(`(?i)^documentation(\([^)]+\))?:`),     // documentation:
		regexp.MustCompile(`(?i)\b(readme|documentation|документация)\b`),
		regexp.MustCompile(`(?i)^📝`),                             // memo emoji
		regexp.MustCompile(`(?i)\.(md|txt|rst)$`),                 // doc file extensions
	}
	
	// Паттерны для определения релиза
	cp.releasePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^release(\([^)]+\))?:`),           // release:
		regexp.MustCompile(`(?i)^version(\([^)]+\))?:`),           // version:
		regexp.MustCompile(`(?i)^v?\d+\.\d+\.\d+`),                // v1.2.3
		regexp.MustCompile(`(?i)\b(release|релиз|версия)\b`),
		regexp.MustCompile(`(?i)^🚀`),                             // rocket emoji
	}
	
	// Паттерны для извлечения задач из названий веток
	cp.branchPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^feature[/\-](\d+)`),              // feature/123
		regexp.MustCompile(`(?i)^fix[/\-](\d+)`),                  // fix/123
		regexp.MustCompile(`(?i)^bugfix[/\-](\d+)`),               // bugfix/123
		regexp.MustCompile(`(?i)^task[/\-](\d+)`),                 // task/123
		regexp.MustCompile(`(?i)^([A-Z]+-\d+)`),                   // PROJ-123
		regexp.MustCompile(`(?i)[/\-](\d+)[/\-]`),                 // любое/123/любое
		regexp.MustCompile(`(?i)^(\d+)[/\-]`),                     // 123/любое
	}
	
	// Паттерны для извлечения задач из путей файлов
	cp.pathPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)/task[_\-]?(\d+)/`),               // /task_123/
		regexp.MustCompile(`(?i)/([A-Z]+-\d+)/`),                  // /PROJ-123/
		regexp.MustCompile(`(?i)task[_\-]?(\d+)\.(go|js|py|java|cpp|cs)`), // task_123.go
		regexp.MustCompile(`(?i)([A-Z]+-\d+)\.(go|js|py|java|cpp|cs)`),    // PROJ-123.go
	}
}

// ExtractTaskIDs извлекает ID задач из сообщения коммита
func (cp *CommitPatterns) ExtractTaskIDs(message string) []string {
	var taskIDs []string
	seen := make(map[string]bool)
	
	for _, pattern := range cp.taskIDPatterns {
		matches := pattern.FindAllStringSubmatch(message, -1)
		for _, match := range matches {
			if len(match) > 1 {
				taskID := strings.ToUpper(match[1])
				if !seen[taskID] {
					taskIDs = append(taskIDs, taskID)
					seen[taskID] = true
				}
			}
		}
	}
	
	return taskIDs
}

// ExtractTaskIDsFromBranch извлекает ID задач из названия ветки
func (cp *CommitPatterns) ExtractTaskIDsFromBranch(branch string) []string {
	var taskIDs []string
	seen := make(map[string]bool)
	
	for _, pattern := range cp.branchPatterns {
		matches := pattern.FindAllStringSubmatch(branch, -1)
		for _, match := range matches {
			if len(match) > 1 {
				taskID := strings.ToUpper(match[1])
				if !seen[taskID] {
					taskIDs = append(taskIDs, taskID)
					seen[taskID] = true
				}
			}
		}
	}
	
	return taskIDs
}

// ExtractTaskIDsFromPath извлекает ID задач из пути файла
func (cp *CommitPatterns) ExtractTaskIDsFromPath(path string) []string {
	var taskIDs []string
	seen := make(map[string]bool)
	
	for _, pattern := range cp.pathPatterns {
		matches := pattern.FindAllStringSubmatch(path, -1)
		for _, match := range matches {
			if len(match) > 1 {
				taskID := strings.ToUpper(match[1])
				if !seen[taskID] {
					taskIDs = append(taskIDs, taskID)
					seen[taskID] = true
				}
			}
		}
	}
	
	return taskIDs
}

// IsFeatureCommit проверяет, является ли коммит фичей
func (cp *CommitPatterns) IsFeatureCommit(message string) bool {
	return cp.matchesAnyPattern(message, cp.featurePatterns)
}

// IsFixCommit проверяет, является ли коммит исправлением
func (cp *CommitPatterns) IsFixCommit(message string) bool {
	return cp.matchesAnyPattern(message, cp.fixPatterns)
}

// IsTestCommit проверяет, является ли коммит тестом
func (cp *CommitPatterns) IsTestCommit(message string) bool {
	return cp.matchesAnyPattern(message, cp.testPatterns)
}

// IsDocCommit проверяет, является ли коммит документацией
func (cp *CommitPatterns) IsDocCommit(message string) bool {
	return cp.matchesAnyPattern(message, cp.docPatterns)
}

// IsReleaseCommit проверяет, является ли коммит релизом
func (cp *CommitPatterns) IsReleaseCommit(message string) bool {
	return cp.matchesAnyPattern(message, cp.releasePatterns)
}

// GetCommitType определяет тип коммита
func (cp *CommitPatterns) GetCommitType(message string) string {
	if cp.IsFeatureCommit(message) {
		return "feature"
	}
	if cp.IsFixCommit(message) {
		return "fix"
	}
	if cp.IsTestCommit(message) {
		return "test"
	}
	if cp.IsDocCommit(message) {
		return "docs"
	}
	if cp.IsReleaseCommit(message) {
		return "release"
	}
	return "other"
}

// GetCommitComplexity оценивает сложность коммита по сообщению
func (cp *CommitPatterns) GetCommitComplexity(message string, filesChanged []string, linesChanged int) CommitComplexity {
	complexity := CommitComplexity{
		Type:         cp.GetCommitType(message),
		Score:        1.0,
		Factors:      make(map[string]float64),
		Description:  "Simple commit",
	}
	
	// Базовая оценка по типу
	switch complexity.Type {
	case "feature":
		complexity.Score = 3.0
		complexity.Description = "Feature implementation"
	case "fix":
		complexity.Score = 2.0
		complexity.Description = "Bug fix"
	case "refactor":
		complexity.Score = 2.5
		complexity.Description = "Code refactoring"
	}
	
	// Факторы сложности
	
	// Количество файлов
	fileCount := len(filesChanged)
	if fileCount > 10 {
		complexity.Factors["many_files"] = 2.0
		complexity.Score *= 2.0
	} else if fileCount > 5 {
		complexity.Factors["several_files"] = 1.5
		complexity.Score *= 1.5
	}
	
	// Количество строк
	if linesChanged > 500 {
		complexity.Factors["large_change"] = 2.0
		complexity.Score *= 2.0
	} else if linesChanged > 100 {
		complexity.Factors["medium_change"] = 1.3
		complexity.Score *= 1.3
	}
	
	// Типы файлов
	hasTestFiles := false
	hasConfigFiles := false
	hasCoreFiles := false
	
	for _, file := range filesChanged {
		lower := strings.ToLower(file)
		if strings.Contains(lower, "test") || strings.Contains(lower, "spec") {
			hasTestFiles = true
		}
		if strings.Contains(lower, "config") || strings.Contains(lower, ".yml") || strings.Contains(lower, ".yaml") || strings.Contains(lower, ".json") {
			hasConfigFiles = true
		}
		if strings.Contains(lower, "main") || strings.Contains(lower, "core") || strings.Contains(lower, "engine") {
			hasCoreFiles = true
		}
	}
	
	if hasCoreFiles {
		complexity.Factors["core_changes"] = 2.0
		complexity.Score *= 2.0
		complexity.Description += " (affects core components)"
	}
	
	if hasConfigFiles {
		complexity.Factors["config_changes"] = 1.2
		complexity.Score *= 1.2
	}
	
	if hasTestFiles {
		complexity.Factors["includes_tests"] = 0.8
		complexity.Score *= 0.8
		complexity.Description += " (includes tests)"
	}
	
	// Ключевые слова в сообщении
	message = strings.ToLower(message)
	if strings.Contains(message, "refactor") || strings.Contains(message, "restructure") {
		complexity.Factors["refactoring"] = 1.8
		complexity.Score *= 1.8
		complexity.Type = "refactor"
	}
	
	if strings.Contains(message, "breaking") || strings.Contains(message, "major") {
		complexity.Factors["breaking_change"] = 3.0
		complexity.Score *= 3.0
		complexity.Description += " (BREAKING CHANGE)"
	}
	
	if strings.Contains(message, "wip") || strings.Contains(message, "work in progress") {
		complexity.Factors["work_in_progress"] = 0.5
		complexity.Score *= 0.5
		complexity.Description += " (work in progress)"
	}
	
	// Нормализуем финальный score
	if complexity.Score > 10.0 {
		complexity.Score = 10.0
	}
	
	return complexity
}

// CommitComplexity сложность коммита
type CommitComplexity struct {
	Type        string             `json:"type"`
	Score       float64            `json:"score"`        // 1.0 - 10.0
	Factors     map[string]float64 `json:"factors"`      // Факторы, влияющие на сложность
	Description string             `json:"description"`  // Описание сложности
}

// EstimateWorkTime оценивает время работы по активности в Git
func (cp *CommitPatterns) EstimateWorkTime(commits []CommitInfo, timeWindow time.Duration) WorkTimeEstimate {
	if len(commits) == 0 {
		return WorkTimeEstimate{}
	}
	
	estimate := WorkTimeEstimate{
		TotalCommits:      len(commits),
		TimeWindow:        timeWindow,
		EstimatedHours:    0,
		Complexity:        "medium",
		BreakdownByType:   make(map[string]int),
		ProductivityScore: 1.0,
	}
	
	totalComplexity := 0.0
	
	for _, commit := range commits {
		commitType := cp.GetCommitType(commit.Message)
		estimate.BreakdownByType[commitType]++
		
		complexity := cp.GetCommitComplexity(commit.Message, commit.FilesChanged, 
			commit.LinesAdded + commit.LinesDeleted)
		totalComplexity += complexity.Score
	}
	
	// Базовая оценка: 30 минут на коммит средней сложности
	avgComplexity := totalComplexity / float64(len(commits))
	estimate.EstimatedHours = (float64(len(commits)) * 0.5 * avgComplexity)
	
	// Корректировка по типам коммитов
	if estimate.BreakdownByType["feature"] > estimate.BreakdownByType["fix"] {
		estimate.EstimatedHours *= 1.5 // Фичи требуют больше времени
		estimate.Complexity = "high"
	} else if estimate.BreakdownByType["fix"] > 0 {
		estimate.EstimatedHours *= 1.2 // Багфиксы тоже требуют времени
	}
	
	// Оценка продуктивности
	hoursInWindow := timeWindow.Hours()
	if hoursInWindow > 0 {
		estimate.ProductivityScore = estimate.EstimatedHours / hoursInWindow
		if estimate.ProductivityScore > 1.0 {
			estimate.ProductivityScore = 1.0
		}
	}
	
	return estimate
}

// WorkTimeEstimate оценка времени работы
type WorkTimeEstimate struct {
	TotalCommits      int                `json:"total_commits"`
	TimeWindow        time.Duration      `json:"time_window"`
	EstimatedHours    float64           `json:"estimated_hours"`
	Complexity        string            `json:"complexity"`       // low, medium, high
	BreakdownByType   map[string]int    `json:"breakdown_by_type"`
	ProductivityScore float64           `json:"productivity_score"` // 0.0 - 1.0
}

// matchesAnyPattern проверяет соответствие любому из паттернов
func (cp *CommitPatterns) matchesAnyPattern(text string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

// AddCustomPattern добавляет пользовательский паттерн
func (cp *CommitPatterns) AddCustomTaskPattern(pattern string) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}
	
	cp.taskIDPatterns = append(cp.taskIDPatterns, regex)
	return nil
}

// GetAllPatterns возвращает все паттерны для отладки
func (cp *CommitPatterns) GetAllPatterns() map[string][]string {
	return map[string][]string{
		"task_ids": cp.patternsToStrings(cp.taskIDPatterns),
		"features": cp.patternsToStrings(cp.featurePatterns),
		"fixes":    cp.patternsToStrings(cp.fixPatterns),
		"tests":    cp.patternsToStrings(cp.testPatterns),
		"docs":     cp.patternsToStrings(cp.docPatterns),
		"releases": cp.patternsToStrings(cp.releasePatterns),
		"branches": cp.patternsToStrings(cp.branchPatterns),
		"paths":    cp.patternsToStrings(cp.pathPatterns),
	}
}

func (cp *CommitPatterns) patternsToStrings(patterns []*regexp.Regexp) []string {
	var result []string
	for _, pattern := range patterns {
		result = append(result, pattern.String())
	}
	return result
}