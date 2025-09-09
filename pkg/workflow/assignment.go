package workflow

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/ai"
)

// AssignmentEngine движок автоматического назначения задач
type AssignmentEngine struct {
	strategies map[string]AssignmentStrategy
	aiChains   *ai.AIChains
	logger     Logger
}

// NewAssignmentEngine создает новый движок назначения
func NewAssignmentEngine(aiChains *ai.AIChains, logger Logger) *AssignmentEngine {
	engine := &AssignmentEngine{
		strategies: make(map[string]AssignmentStrategy),
		aiChains:   aiChains,
		logger:     logger,
	}
	
	// Регистрируем стандартные стратегии
	engine.RegisterStrategy(&SkillsBasedStrategy{aiChains: aiChains, logger: logger})
	engine.RegisterStrategy(&WorkloadBalancedStrategy{logger: logger})
	engine.RegisterStrategy(&RoundRobinStrategy{logger: logger})
	engine.RegisterStrategy(&AIOptimizedStrategy{aiChains: aiChains, logger: logger})
	
	return engine
}

// RegisterStrategy регистрирует стратегию назначения
func (ae *AssignmentEngine) RegisterStrategy(strategy AssignmentStrategy) {
	ae.strategies[strategy.GetName()] = strategy
	ae.logger.Info("Registered assignment strategy", "name", strategy.GetName())
}

// AssignTask назначает задачу на основе выбранной стратегии
func (ae *AssignmentEngine) AssignTask(ctx context.Context, task TaskInfo, team TeamInfo, strategyName string) (string, error) {
	strategy, exists := ae.strategies[strategyName]
	if !exists {
		return "", fmt.Errorf("unknown assignment strategy: %s", strategyName)
	}
	
	assignee, err := strategy.Assign(ctx, task, team)
	if err != nil {
		ae.logger.Error("Task assignment failed", err, 
			"strategy", strategyName, 
			"task", task.ID)
		return "", err
	}
	
	ae.logger.Info("Task assigned successfully", 
		"task", task.ID, 
		"assignee", assignee, 
		"strategy", strategyName)
	
	return assignee, nil
}

// GetAvailableStrategies возвращает список доступных стратегий
func (ae *AssignmentEngine) GetAvailableStrategies() []string {
	var strategies []string
	for name := range ae.strategies {
		strategies = append(strategies, name)
	}
	sort.Strings(strategies)
	return strategies
}

// SkillsBasedStrategy стратегия назначения на основе навыков
type SkillsBasedStrategy struct {
	aiChains *ai.AIChains
	logger   Logger
}

func (s *SkillsBasedStrategy) GetName() string {
	return "skills_based"
}

func (s *SkillsBasedStrategy) Assign(ctx context.Context, task TaskInfo, team TeamInfo) (string, error) {
	if len(team.Members) == 0 {
		return "", fmt.Errorf("no team members available")
	}
	
	// Находим участников с подходящими навыками
	var candidates []TeamMember
	for _, member := range team.Members {
		if s.hasRequiredSkills(member, task.Skills) && s.isAvailable(member) {
			candidates = append(candidates, member)
		}
	}
	
	if len(candidates) == 0 {
		// Если никого с точными навыками нет, ищем похожих
		candidates = s.findSimilarSkills(team.Members, task.Skills)
	}
	
	if len(candidates) == 0 {
		return "", fmt.Errorf("no suitable candidates found for required skills: %v", task.Skills)
	}
	
	// Выбираем лучшего кандидата
	bestCandidate := s.selectBestCandidate(candidates, task)
	
	s.logger.Info("Skills-based assignment completed", 
		"task", task.ID, 
		"assignee", bestCandidate.ID,
		"skills_match", s.calculateSkillsMatch(bestCandidate, task.Skills))
	
	return bestCandidate.ID, nil
}

func (s *SkillsBasedStrategy) hasRequiredSkills(member TeamMember, requiredSkills []string) bool {
	memberSkills := make(map[string]bool)
	for _, skill := range member.Skills {
		memberSkills[strings.ToLower(skill)] = true
	}
	
	for _, required := range requiredSkills {
		if !memberSkills[strings.ToLower(required)] {
			return false
		}
	}
	
	return true
}

func (s *SkillsBasedStrategy) isAvailable(member TeamMember) bool {
	return member.CurrentLoad < member.MaxLoad
}

func (s *SkillsBasedStrategy) findSimilarSkills(members []TeamMember, requiredSkills []string) []TeamMember {
	var candidates []TeamMember
	
	for _, member := range members {
		if !s.isAvailable(member) {
			continue
		}
		
		match := s.calculateSkillsMatch(member, requiredSkills)
		if match > 0.3 { // Минимум 30% совпадения
			candidates = append(candidates, member)
		}
	}
	
	return candidates
}

func (s *SkillsBasedStrategy) calculateSkillsMatch(member TeamMember, requiredSkills []string) float64 {
	if len(requiredSkills) == 0 {
		return 1.0
	}
	
	memberSkills := make(map[string]bool)
	for _, skill := range member.Skills {
		memberSkills[strings.ToLower(skill)] = true
	}
	
	matches := 0
	for _, required := range requiredSkills {
		if memberSkills[strings.ToLower(required)] {
			matches++
		}
	}
	
	return float64(matches) / float64(len(requiredSkills))
}

func (s *SkillsBasedStrategy) selectBestCandidate(candidates []TeamMember, task TaskInfo) TeamMember {
	type candidateScore struct {
		member TeamMember
		score  float64
	}
	
	var scores []candidateScore
	
	for _, candidate := range candidates {
		score := s.calculateCandidateScore(candidate, task)
		scores = append(scores, candidateScore{
			member: candidate,
			score:  score,
		})
	}
	
	// Сортируем по убыванию скора
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	
	return scores[0].member
}

func (s *SkillsBasedStrategy) calculateCandidateScore(member TeamMember, task TaskInfo) float64 {
	skillsMatch := s.calculateSkillsMatch(member, task.Skills)
	loadFactor := 1.0 - (float64(member.CurrentLoad) / float64(member.MaxLoad))
	performanceFactor := member.Performance.QualityScore * member.Performance.CompletionRate
	
	// Взвешенная сумма факторов
	score := skillsMatch*0.5 + loadFactor*0.3 + performanceFactor*0.2
	
	return score
}

// WorkloadBalancedStrategy стратегия балансировки нагрузки
type WorkloadBalancedStrategy struct {
	logger Logger
}

func (s *WorkloadBalancedStrategy) GetName() string {
	return "workload_balanced"
}

func (s *WorkloadBalancedStrategy) Assign(ctx context.Context, task TaskInfo, team TeamInfo) (string, error) {
	if len(team.Members) == 0 {
		return "", fmt.Errorf("no team members available")
	}
	
	// Фильтруем доступных участников
	var available []TeamMember
	for _, member := range team.Members {
		if member.CurrentLoad < member.MaxLoad {
			available = append(available, member)
		}
	}
	
	if len(available) == 0 {
		return "", fmt.Errorf("all team members are at maximum capacity")
	}
	
	// Находим участника с минимальной нагрузкой
	bestMember := available[0]
	minLoadPercentage := float64(bestMember.CurrentLoad) / float64(bestMember.MaxLoad)
	
	for _, member := range available[1:] {
		loadPercentage := float64(member.CurrentLoad) / float64(member.MaxLoad)
		if loadPercentage < minLoadPercentage {
			bestMember = member
			minLoadPercentage = loadPercentage
		}
	}
	
	s.logger.Info("Workload-balanced assignment completed", 
		"task", task.ID, 
		"assignee", bestMember.ID,
		"load_percentage", int(minLoadPercentage*100))
	
	return bestMember.ID, nil
}

// RoundRobinStrategy стратегия круговой очереди
type RoundRobinStrategy struct {
	lastAssignedIndex int
	logger            Logger
}

func (s *RoundRobinStrategy) GetName() string {
	return "round_robin"
}

func (s *RoundRobinStrategy) Assign(ctx context.Context, task TaskInfo, team TeamInfo) (string, error) {
	if len(team.Members) == 0 {
		return "", fmt.Errorf("no team members available")
	}
	
	// Фильтруем доступных участников
	var available []TeamMember
	for _, member := range team.Members {
		if member.CurrentLoad < member.MaxLoad {
			available = append(available, member)
		}
	}
	
	if len(available) == 0 {
		return "", fmt.Errorf("all team members are at maximum capacity")
	}
	
	// Выбираем следующего в очереди
	s.lastAssignedIndex = (s.lastAssignedIndex + 1) % len(available)
	assignee := available[s.lastAssignedIndex]
	
	s.logger.Info("Round-robin assignment completed", 
		"task", task.ID, 
		"assignee", assignee.ID,
		"round_robin_index", s.lastAssignedIndex)
	
	return assignee.ID, nil
}

// AIOptimizedStrategy AI-оптимизированная стратегия назначения
type AIOptimizedStrategy struct {
	aiChains *ai.AIChains
	logger   Logger
}

func (s *AIOptimizedStrategy) GetName() string {
	return "ai_optimized"
}

func (s *AIOptimizedStrategy) Assign(ctx context.Context, task TaskInfo, team TeamInfo) (string, error) {
	if len(team.Members) == 0 {
		return "", fmt.Errorf("no team members available")
	}
	
	// Подготавливаем контекст для AI
	prompt := s.buildAssignmentPrompt(task, team)
	
	// Используем AI для анализа оптимального назначения
	suggestion, err := s.getAISuggestion(ctx, prompt)
	if err != nil {
		s.logger.Error("AI assignment suggestion failed, falling back to skills-based", err)
		// Fallback к skills-based стратегии
		fallback := &SkillsBasedStrategy{aiChains: s.aiChains, logger: s.logger}
		return fallback.Assign(ctx, task, team)
	}
	
	// Валидируем предложение AI
	if !s.validateAISuggestion(suggestion, team) {
		s.logger.Error("AI suggestion validation failed, using fallback", nil)
		fallback := &SkillsBasedStrategy{aiChains: s.aiChains, logger: s.logger}
		return fallback.Assign(ctx, task, team)
	}
	
	s.logger.Info("AI-optimized assignment completed", 
		"task", task.ID, 
		"assignee", suggestion,
		"source", "ai_recommendation")
	
	return suggestion, nil
}

func (s *AIOptimizedStrategy) buildAssignmentPrompt(task TaskInfo, team TeamInfo) string {
	prompt := fmt.Sprintf(`Analyze the following task and team to suggest the optimal assignment:

TASK DETAILS:
- Title: %s
- Description: %s
- Required Skills: %v
- Priority: %s
- Complexity: %s
- Estimated Hours: %d

TEAM MEMBERS:
`, task.Title, task.Description, task.Skills, task.Priority, task.Complexity, task.Estimate)
	
	for i, member := range team.Members {
		prompt += fmt.Sprintf(`%d. %s (%s)
   - Skills: %v
   - Current Load: %d/%d (%.1f%%)
   - Quality Score: %.2f
   - Completion Rate: %.2f
   - Avg Velocity: %.2f
   - Response Time: %dh

`, i+1, member.Name, member.ID, member.Skills, 
			member.CurrentLoad, member.MaxLoad, 
			float64(member.CurrentLoad)/float64(member.MaxLoad)*100,
			member.Performance.QualityScore,
			member.Performance.CompletionRate,
			member.Performance.AverageVelocity,
			member.Performance.ResponseTime)
	}
	
	prompt += `
Based on this information, recommend the BEST team member for this task considering:
1. Skills match and expertise
2. Current workload and availability
3. Historical performance and quality
4. Task complexity and urgency
5. Team collaboration patterns

Please respond with ONLY the member ID (e.g., "user123") of your recommendation.`
	
	return prompt
}

func (s *AIOptimizedStrategy) getAISuggestion(ctx context.Context, prompt string) (string, error) {
	// Используем AI chains для получения рекомендации
	response, err := s.aiChains.ExecuteTask("Assignment Analysis", prompt, "planning")
	if err != nil {
		return "", err
	}
	
	// Извлекаем ID пользователя из ответа
	suggestion := s.extractUserIDFromResponse(response)
	if suggestion == "" {
		return "", fmt.Errorf("failed to extract user ID from AI response")
	}
	
	return suggestion, nil
}

func (s *AIOptimizedStrategy) extractUserIDFromResponse(response string) string {
	// Простой парсер для извлечения ID пользователя
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(line, " ") {
			// Вероятно, это ID пользователя
			return line
		}
	}
	
	// Поиск паттернов вида "user123" или "member_456"
	words := strings.Fields(response)
	for _, word := range words {
		if strings.Contains(word, "user") || strings.Contains(word, "member") {
			return word
		}
	}
	
	return ""
}

func (s *AIOptimizedStrategy) validateAISuggestion(suggestion string, team TeamInfo) bool {
	// Проверяем, что предложенный пользователь существует и доступен
	for _, member := range team.Members {
		if member.ID == suggestion {
			return member.CurrentLoad < member.MaxLoad
		}
	}
	return false
}

// TeamAnalyzer анализирует состояние команды
type TeamAnalyzer struct {
	logger Logger
}

// NewTeamAnalyzer создает новый анализатор команды
func NewTeamAnalyzer(logger Logger) *TeamAnalyzer {
	return &TeamAnalyzer{logger: logger}
}

// AnalyzeTeamHealth анализирует здоровье команды
func (ta *TeamAnalyzer) AnalyzeTeamHealth(team TeamInfo) TeamHealthReport {
	report := TeamHealthReport{
		TotalMembers:    len(team.Members),
		ActiveMembers:   0,
		OverloadedMembers: 0,
		AverageLoad:     0,
		SkillsCoverage:  make(map[string]int),
		Recommendations: []string{},
	}
	
	totalLoad := 0
	totalCapacity := 0
	skillsMap := make(map[string]int)
	
	for _, member := range team.Members {
		totalLoad += member.CurrentLoad
		totalCapacity += member.MaxLoad
		
		if member.CurrentLoad > 0 {
			report.ActiveMembers++
		}
		
		if member.CurrentLoad >= member.MaxLoad {
			report.OverloadedMembers++
		}
		
		// Подсчитываем покрытие навыков
		for _, skill := range member.Skills {
			skillsMap[skill]++
		}
	}
	
	if totalCapacity > 0 {
		report.AverageLoad = float64(totalLoad) / float64(totalCapacity) * 100
	}
	
	report.SkillsCoverage = skillsMap
	
	// Генерируем рекомендации
	report.Recommendations = ta.generateRecommendations(report, team)
	
	return report
}

func (ta *TeamAnalyzer) generateRecommendations(report TeamHealthReport, team TeamInfo) []string {
	var recommendations []string
	
	// Проверка перегрузки
	if report.OverloadedMembers > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("⚠️ %d team members are overloaded. Consider redistributing tasks.", report.OverloadedMembers))
	}
	
	// Проверка общей нагрузки
	if report.AverageLoad > 90 {
		recommendations = append(recommendations, "🔥 Team is operating at very high capacity. Consider scaling up.")
	} else if report.AverageLoad < 30 {
		recommendations = append(recommendations, "💤 Team utilization is low. Consider taking on more work.")
	}
	
	// Анализ навыков
	criticalSkills := []string{"development", "testing", "devops"}
	for _, skill := range criticalSkills {
		if count, exists := report.SkillsCoverage[skill]; !exists || count < 2 {
			recommendations = append(recommendations, 
				fmt.Sprintf("⚠️ Limited coverage for critical skill: %s. Consider training or hiring.", skill))
		}
	}
	
	// Анализ производительности
	lowPerformers := 0
	for _, member := range team.Members {
		if member.Performance.CompletionRate < 0.7 || member.Performance.QualityScore < 0.7 {
			lowPerformers++
		}
	}
	
	if lowPerformers > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("📈 %d team members may benefit from additional support or training.", lowPerformers))
	}
	
	return recommendations
}

// TeamHealthReport отчет о здоровье команды
type TeamHealthReport struct {
	TotalMembers      int                `json:"total_members"`
	ActiveMembers     int                `json:"active_members"`
	OverloadedMembers int                `json:"overloaded_members"`
	AverageLoad       float64            `json:"average_load_percentage"`
	SkillsCoverage    map[string]int     `json:"skills_coverage"`
	Recommendations   []string           `json:"recommendations"`
	GeneratedAt       time.Time          `json:"generated_at"`
}

// AssignmentHistory история назначений для обучения
type AssignmentHistory struct {
	TaskID       string    `json:"task_id"`
	AssigneeID   string    `json:"assignee_id"`
	Strategy     string    `json:"strategy"`
	AssignedAt   time.Time `json:"assigned_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	Success      bool      `json:"success"`
	QualityScore float64   `json:"quality_score"`
	Duration     int       `json:"duration_hours"`
}

// LearningEngine движок машинного обучения для улучшения назначений
type LearningEngine struct {
	history []AssignmentHistory
	logger  Logger
}

// NewLearningEngine создает новый движок обучения
func NewLearningEngine(logger Logger) *LearningEngine {
	return &LearningEngine{
		history: make([]AssignmentHistory, 0),
		logger:  logger,
	}
}

// RecordAssignment записывает результат назначения
func (le *LearningEngine) RecordAssignment(assignment AssignmentHistory) {
	le.history = append(le.history, assignment)
	le.logger.Info("Assignment recorded for learning", 
		"task", assignment.TaskID, 
		"assignee", assignment.AssigneeID,
		"success", assignment.Success)
}

// GetStrategyEffectiveness возвращает эффективность стратегий
func (le *LearningEngine) GetStrategyEffectiveness() map[string]float64 {
	effectiveness := make(map[string]float64)
	strategyCounts := make(map[string]int)
	strategySuccesses := make(map[string]int)
	
	for _, record := range le.history {
		strategyCounts[record.Strategy]++
		if record.Success {
			strategySuccesses[record.Strategy]++
		}
	}
	
	for strategy, total := range strategyCounts {
		if total > 0 {
			effectiveness[strategy] = float64(strategySuccesses[strategy]) / float64(total)
		}
	}
	
	return effectiveness
}