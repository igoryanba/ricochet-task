package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/providers"
)

// BoardRegistry реестр доступных досок и проектов
type BoardRegistry struct {
	mu           sync.RWMutex
	boards       map[string]*BoardInfo
	providers    map[string]providers.TaskProvider
	cachePath    string
	cacheTimeout time.Duration
	logger       Logger
}

// BoardInfo информация о доске
type BoardInfo struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	ProjectID    string                 `json:"project_id"`
	ProjectName  string                 `json:"project_name"`
	ProviderName string                 `json:"provider_name"`
	Type         string                 `json:"type"`         // agile, kanban, scrum
	Status       string                 `json:"status"`       // active, archived, private
	URL          string                 `json:"url"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	LastSyncAt   time.Time              `json:"last_sync_at"`
	
	// Статистика доски
	Stats        *BoardStats            `json:"stats"`
	
	// Настройки доски
	Settings     *BoardSettings         `json:"settings"`
	
	// Команда и участники
	Team         []TeamMember           `json:"team"`
	
	// Кастомные поля
	CustomFields map[string]interface{} `json:"custom_fields"`
}

// BoardStats статистика доски
type BoardStats struct {
	TotalTasks       int       `json:"total_tasks"`
	ActiveTasks      int       `json:"active_tasks"`
	CompletedTasks   int       `json:"completed_tasks"`
	OverdueTasks     int       `json:"overdue_tasks"`
	TeamMembers      int       `json:"team_members"`
	AverageTaskTime  float64   `json:"average_task_time"` // days
	Velocity         float64   `json:"velocity"`          // tasks per week
	LastActivity     time.Time `json:"last_activity"`
}

// BoardSettings настройки доски
type BoardSettings struct {
	AutoAssignment  bool     `json:"auto_assignment"`
	AutoProgress    bool     `json:"auto_progress"`
	RequireReview   bool     `json:"require_review"`
	AllowGuests     bool     `json:"allow_guests"`
	DefaultLabels   []string `json:"default_labels"`
	DefaultPriority string   `json:"default_priority"`
	WorkflowType    string   `json:"workflow_type"`
	SprintLength    int      `json:"sprint_length"` // days
}

// TeamMember участник команды
type TeamMember struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Role     string   `json:"role"`     // admin, member, viewer
	Skills   []string `json:"skills"`
	Workload int      `json:"workload"` // current task count
	Active   bool     `json:"active"`
}

// BoardCapabilities возможности доски
type BoardCapabilities struct {
	SupportsEpics      bool `json:"supports_epics"`
	SupportsSubtasks   bool `json:"supports_subtasks"`
	SupportsSpints     bool `json:"supports_sprints"`
	SupportsTimeTrack  bool `json:"supports_time_tracking"`
	SupportsAutomation bool `json:"supports_automation"`
	SupportsCustomFields bool `json:"supports_custom_fields"`
}

// NewBoardRegistry создает новый реестр досок
func NewBoardRegistry(cachePath string, logger Logger) *BoardRegistry {
	if cachePath == "" {
		homeDir, _ := os.UserHomeDir()
		cachePath = filepath.Join(homeDir, ".ricochet", "boards_cache.json")
	}

	registry := &BoardRegistry{
		boards:       make(map[string]*BoardInfo),
		providers:    make(map[string]providers.TaskProvider),
		cachePath:    cachePath,
		cacheTimeout: 30 * time.Minute, // кеш на 30 минут
		logger:       logger,
	}

	// Загружаем кешированные данные
	registry.loadCache()

	return registry
}

// RegisterProvider регистрирует провайдера в реестре
func (br *BoardRegistry) RegisterProvider(name string, provider providers.TaskProvider) {
	br.mu.Lock()
	defer br.mu.Unlock()

	br.providers[name] = provider
	br.logger.Info("Provider registered", "name", name)
}

// SyncBoards синхронизирует доски со всеми провайдерами
func (br *BoardRegistry) SyncBoards() error {
	br.mu.Lock()
	defer br.mu.Unlock()

	totalBoards := 0
	errors := []error{}

	for providerName, provider := range br.providers {
		br.logger.Debug("Syncing boards", "provider", providerName)
		
		boards, err := br.syncProviderBoards(providerName, provider)
		if err != nil {
			errors = append(errors, fmt.Errorf("provider %s: %w", providerName, err))
			continue
		}

		totalBoards += len(boards)
	}

	// Сохраняем обновленный кеш
	if err := br.saveCache(); err != nil {
		br.logger.Error("Failed to save board cache", err)
	}

	br.logger.Info("Board sync completed", "total_boards", totalBoards, "errors", len(errors))

	if len(errors) > 0 {
		return fmt.Errorf("sync completed with %d errors: %v", len(errors), errors)
	}

	return nil
}

// syncProviderBoards синхронизирует доски конкретного провайдера
func (br *BoardRegistry) syncProviderBoards(providerName string, provider providers.TaskProvider) ([]*BoardInfo, error) {
	// TODO: Пока реализуем заглушку, так как GetBoardProvider не реализован
	// в базовом интерфейсе TaskProvider
	br.logger.Debug("Board sync not yet fully implemented", "provider", providerName)
	
	// Создаем пример доски для демонстрации структуры
	now := time.Now()
	boardInfo := &BoardInfo{
		ID:           "DEFAULT-BOARD",
		Name:         fmt.Sprintf("%s Default Board", providerName),
		Description:  "Default board created by board registry",
		ProjectID:    "DEFAULT-PROJECT",
		ProjectName:  "Default Project",
		ProviderName: providerName,
		Type:         "kanban",
		Status:       "active",
		URL:          "",
		CreatedAt:    now,
		UpdatedAt:    now,
		LastSyncAt:   now,
		CustomFields: make(map[string]interface{}),
		Stats:        &BoardStats{},
		Settings:     &BoardSettings{
			DefaultPriority: "medium",
			WorkflowType:    "kanban",
			SprintLength:    14,
		},
		Team: []TeamMember{},
	}

	// Создаем уникальный ключ для доски
	boardKey := fmt.Sprintf("%s:%s:%s", providerName, boardInfo.ProjectID, boardInfo.ID)
	br.boards[boardKey] = boardInfo

	return []*BoardInfo{boardInfo}, nil
}

// getBoardStats получает статистику доски
func (br *BoardRegistry) getBoardStats(provider providers.TaskProvider, boardID string) (*BoardStats, error) {
	// TODO: Реализовать получение статистики через provider
	// Пока возвращаем заглушку
	return &BoardStats{
		TotalTasks:      0,
		ActiveTasks:     0,
		CompletedTasks:  0,
		OverdueTasks:    0,
		TeamMembers:     0,
		AverageTaskTime: 0.0,
		Velocity:        0.0,
		LastActivity:    time.Now(),
	}, nil
}

// getBoardSettings получает настройки доски
func (br *BoardRegistry) getBoardSettings(provider providers.TaskProvider, boardID string) (*BoardSettings, error) {
	// TODO: Реализовать получение настроек через provider
	// Пока возвращаем настройки по умолчанию
	return &BoardSettings{
		AutoAssignment:  false,
		AutoProgress:    false,
		RequireReview:   true,
		AllowGuests:     false,
		DefaultLabels:   []string{},
		DefaultPriority: "medium",
		WorkflowType:    "agile",
		SprintLength:    14,
	}, nil
}

// getBoardTeam получает команду доски
func (br *BoardRegistry) getBoardTeam(provider providers.TaskProvider, projectID string) ([]TeamMember, error) {
	// TODO: Реализовать получение команды через provider
	// Пока возвращаем пустую команду
	return []TeamMember{}, nil
}

// ListBoards возвращает список всех досок
func (br *BoardRegistry) ListBoards(filters *BoardFilters) ([]*BoardInfo, error) {
	br.mu.RLock()
	defer br.mu.RUnlock()

	// Проверяем нужна ли синхронизация
	if br.needsSync() {
		br.mu.RUnlock()
		if err := br.SyncBoards(); err != nil {
			br.logger.Error("Failed to sync boards", err)
		}
		br.mu.RLock()
	}

	var result []*BoardInfo

	for _, board := range br.boards {
		if filters != nil && !br.matchesFilters(board, filters) {
			continue
		}
		result = append(result, board)
	}

	return result, nil
}

// BoardFilters фильтры для поиска досок
type BoardFilters struct {
	ProviderName string   `json:"provider_name"`
	ProjectID    string   `json:"project_id"`
	Type         string   `json:"type"`
	Status       string   `json:"status"`
	ActiveOnly   bool     `json:"active_only"`
	WithTeam     bool     `json:"with_team"`
	Skills       []string `json:"skills"`
	Search       string   `json:"search"`
}

// matchesFilters проверяет соответствие доски фильтрам
func (br *BoardRegistry) matchesFilters(board *BoardInfo, filters *BoardFilters) bool {
	if filters.ProviderName != "" && board.ProviderName != filters.ProviderName {
		return false
	}

	if filters.ProjectID != "" && board.ProjectID != filters.ProjectID {
		return false
	}

	if filters.Type != "" && board.Type != filters.Type {
		return false
	}

	if filters.Status != "" && board.Status != filters.Status {
		return false
	}

	if filters.ActiveOnly && board.Status != "active" {
		return false
	}

	if filters.WithTeam && len(board.Team) == 0 {
		return false
	}

	if filters.Search != "" {
		searchLower := strings.ToLower(filters.Search)
		if !strings.Contains(strings.ToLower(board.Name), searchLower) &&
		   !strings.Contains(strings.ToLower(board.Description), searchLower) &&
		   !strings.Contains(strings.ToLower(board.ProjectName), searchLower) {
			return false
		}
	}

	if len(filters.Skills) > 0 {
		hasSkill := false
		for _, member := range board.Team {
			for _, memberSkill := range member.Skills {
				for _, requiredSkill := range filters.Skills {
					if strings.EqualFold(memberSkill, requiredSkill) {
						hasSkill = true
						break
					}
				}
				if hasSkill {
					break
				}
			}
			if hasSkill {
				break
			}
		}
		if !hasSkill {
			return false
		}
	}

	return true
}

// GetBoard возвращает информацию о конкретной доске
func (br *BoardRegistry) GetBoard(providerName, projectID, boardID string) (*BoardInfo, error) {
	br.mu.RLock()
	defer br.mu.RUnlock()

	boardKey := fmt.Sprintf("%s:%s:%s", providerName, projectID, boardID)
	board, exists := br.boards[boardKey]
	if !exists {
		return nil, fmt.Errorf("board not found: %s", boardKey)
	}

	return board, nil
}

// FindBestBoard находит наиболее подходящую доску для проекта
func (br *BoardRegistry) FindBestBoard(analysis *ProjectAnalysis) (*BoardInfo, error) {
	filters := &BoardFilters{
		ActiveOnly: true,
		Skills:     analysis.RequiredSkills,
	}

	// Если предложена конкретная доска, ищем её
	if analysis.SuggestedBoard != "" {
		filters.Search = analysis.SuggestedBoard
	}

	boards, err := br.ListBoards(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list boards: %w", err)
	}

	if len(boards) == 0 {
		return nil, fmt.Errorf("no suitable boards found")
	}

	// Ранжируем доски по соответствию
	bestBoard := br.rankBoards(boards, analysis)[0]
	return bestBoard, nil
}

// rankBoards ранжирует доски по соответствию проекту
func (br *BoardRegistry) rankBoards(boards []*BoardInfo, analysis *ProjectAnalysis) []*BoardInfo {
	type scoredBoard struct {
		board *BoardInfo
		score float64
	}

	var scored []scoredBoard

	for _, board := range boards {
		score := 0.0

		// Бонус за соответствие типа проекта
		if strings.Contains(strings.ToLower(board.Name), analysis.ProjectType) {
			score += 10.0
		}

		// Бонус за соответствие фреймворка
		if strings.Contains(strings.ToLower(board.Name), analysis.Framework) {
			score += 8.0
		}

		// Бонус за активность доски
		if board.Stats != nil {
			if board.Stats.LastActivity.After(time.Now().AddDate(0, 0, -7)) {
				score += 5.0
			}
			// Бонус за команду подходящего размера
			if board.Stats.TeamMembers >= analysis.TeamSize-1 && 
			   board.Stats.TeamMembers <= analysis.TeamSize+2 {
				score += 7.0
			}
		}

		// Бонус за совпадение навыков команды
		skillMatches := 0
		for _, member := range board.Team {
			for _, memberSkill := range member.Skills {
				for _, requiredSkill := range analysis.RequiredSkills {
					if strings.EqualFold(memberSkill, requiredSkill) {
						skillMatches++
					}
				}
			}
		}
		score += float64(skillMatches) * 2.0

		// Штраф за перегруженную доску
		if board.Stats != nil && board.Stats.ActiveTasks > 50 {
			score -= 3.0
		}

		scored = append(scored, scoredBoard{board: board, score: score})
	}

	// Сортируем по убыванию score
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[i].score < scored[j].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// Возвращаем отсортированный список досок
	result := make([]*BoardInfo, len(scored))
	for i, s := range scored {
		result[i] = s.board
	}

	return result
}

// needsSync проверяет нужна ли синхронизация
func (br *BoardRegistry) needsSync() bool {
	// Если нет досок, нужна синхронизация
	if len(br.boards) == 0 {
		return true
	}

	// Проверяем время последней синхронизации
	for _, board := range br.boards {
		if time.Since(board.LastSyncAt) > br.cacheTimeout {
			return true
		}
		// Достаточно проверить одну доску
		break
	}

	return false
}

// saveCache сохраняет кеш досок в файл
func (br *BoardRegistry) saveCache() error {
	dir := filepath.Dir(br.cachePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	file, err := os.Create(br.cachePath)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	data := struct {
		Boards      map[string]*BoardInfo `json:"boards"`
		LastSync    time.Time             `json:"last_sync"`
		CacheExpiry time.Time             `json:"cache_expiry"`
	}{
		Boards:      br.boards,
		LastSync:    time.Now(),
		CacheExpiry: time.Now().Add(br.cacheTimeout),
	}

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode cache: %w", err)
	}

	return nil
}

// loadCache загружает кеш досок из файла
func (br *BoardRegistry) loadCache() {
	if _, err := os.Stat(br.cachePath); os.IsNotExist(err) {
		br.logger.Debug("Board cache file does not exist")
		return
	}

	file, err := os.Open(br.cachePath)
	if err != nil {
		br.logger.Error("Failed to open board cache", err)
		return
	}
	defer file.Close()

	var data struct {
		Boards      map[string]*BoardInfo `json:"boards"`
		LastSync    time.Time             `json:"last_sync"`
		CacheExpiry time.Time             `json:"cache_expiry"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		br.logger.Error("Failed to decode board cache", err)
		return
	}

	// Проверяем не истек ли кеш
	if time.Now().After(data.CacheExpiry) {
		br.logger.Debug("Board cache expired")
		return
	}

	br.boards = data.Boards

	// Инициализируем пустые карты если nil
	for _, board := range br.boards {
		if board.CustomFields == nil {
			board.CustomFields = make(map[string]interface{})
		}
		if board.Stats == nil {
			board.Stats = &BoardStats{}
		}
		if board.Settings == nil {
			board.Settings = &BoardSettings{}
		}
		if board.Team == nil {
			board.Team = []TeamMember{}
		}
	}

	br.logger.Info("Board cache loaded", "boards", len(br.boards), "last_sync", data.LastSync)
}

// GetBoardCapabilities возвращает возможности доски
func (br *BoardRegistry) GetBoardCapabilities(providerName string) *BoardCapabilities {
	// Возможности по типу провайдера
	capabilities := map[string]*BoardCapabilities{
		"youtrack": {
			SupportsEpics:        true,
			SupportsSubtasks:     true,
			SupportsSpints:       true,
			SupportsTimeTrack:    true,
			SupportsAutomation:   true,
			SupportsCustomFields: true,
		},
		"jira": {
			SupportsEpics:        true,
			SupportsSubtasks:     true,
			SupportsSpints:       true,
			SupportsTimeTrack:    true,
			SupportsAutomation:   true,
			SupportsCustomFields: true,
		},
		"github": {
			SupportsEpics:        false,
			SupportsSubtasks:     false,
			SupportsSpints:       false,
			SupportsTimeTrack:    false,
			SupportsAutomation:   true,
			SupportsCustomFields: false,
		},
	}

	if cap, exists := capabilities[providerName]; exists {
		return cap
	}

	// Возможности по умолчанию
	return &BoardCapabilities{
		SupportsEpics:        false,
		SupportsSubtasks:     false,
		SupportsSpints:       false,
		SupportsTimeTrack:    false,
		SupportsAutomation:   false,
		SupportsCustomFields: false,
	}
}

// RefreshBoard обновляет информацию о конкретной доске
func (br *BoardRegistry) RefreshBoard(providerName, projectID, boardID string) error {
	br.mu.Lock()
	defer br.mu.Unlock()

	provider, exists := br.providers[providerName]
	if !exists {
		return fmt.Errorf("provider %s not found", providerName)
	}

	boards, err := br.syncProviderBoards(providerName, provider)
	if err != nil {
		return fmt.Errorf("failed to sync provider boards: %w", err)
	}

	// Ищем конкретную доску
	boardKey := fmt.Sprintf("%s:%s:%s", providerName, projectID, boardID)
	found := false
	for _, board := range boards {
		if fmt.Sprintf("%s:%s:%s", board.ProviderName, board.ProjectID, board.ID) == boardKey {
			br.boards[boardKey] = board
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("board %s not found after sync", boardKey)
	}

	// Сохраняем обновленный кеш
	if err := br.saveCache(); err != nil {
		br.logger.Error("Failed to save board cache after refresh", err)
	}

	br.logger.Info("Board refreshed", "board", boardKey)
	return nil
}