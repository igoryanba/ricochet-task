package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ProgressEngine движок отслеживания прогресса задач
type ProgressEngine struct {
	tracker      *GitProgressTracker
	logger       Logger
	storage      ProgressStorage
	cache        map[string]*TaskProgress
	cacheMutex   sync.RWMutex
	subscribers  map[string][]ProgressSubscriber
	subsMutex    sync.RWMutex
}

// ProgressStorage интерфейс для хранения прогресса
type ProgressStorage interface {
	SaveProgress(taskID string, progress *TaskProgress) error
	LoadProgress(taskID string) (*TaskProgress, error)
	LoadAllProgress() (map[string]*TaskProgress, error)
	DeleteProgress(taskID string) error
}

// ProgressSubscriber интерфейс для подписчиков на изменения прогресса
type ProgressSubscriber interface {
	OnProgressUpdate(taskID string, progress *TaskProgress, event *ProgressEvent)
}

// ProgressEvent событие изменения прогресса
type ProgressEvent struct {
	Type        string                 `json:"type"`         // stage_change, progress_update, activity_added
	TaskID      string                 `json:"task_id"`
	OldValue    interface{}            `json:"old_value"`
	NewValue    interface{}            `json:"new_value"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NewProgressEngine создает новый движок прогресса
func NewProgressEngine(tracker *GitProgressTracker, logger Logger) *ProgressEngine {
	storage := NewFileProgressStorage("./data/progress", logger)
	
	engine := &ProgressEngine{
		tracker:     tracker,
		logger:      logger,
		storage:     storage,
		cache:       make(map[string]*TaskProgress),
		subscribers: make(map[string][]ProgressSubscriber),
	}
	
	// Загружаем прогресс из хранилища
	if err := engine.loadProgressFromStorage(); err != nil {
		logger.Error("Failed to load progress from storage", err)
	}
	
	return engine
}

// GetTaskProgress возвращает прогресс задачи
func (pe *ProgressEngine) GetTaskProgress(taskID string) *TaskProgress {
	pe.cacheMutex.RLock()
	defer pe.cacheMutex.RUnlock()
	
	if progress, exists := pe.cache[taskID]; exists {
		return progress
	}
	
	// Пытаемся загрузить из хранилища
	progress, err := pe.storage.LoadProgress(taskID)
	if err != nil {
		pe.logger.Debug("Task progress not found", "task", taskID)
		return nil
	}
	
	pe.cacheMutex.RUnlock()
	pe.cacheMutex.Lock()
	pe.cache[taskID] = progress
	pe.cacheMutex.Unlock()
	pe.cacheMutex.RLock()
	
	return progress
}

// SaveTaskProgress сохраняет прогресс задачи
func (pe *ProgressEngine) SaveTaskProgress(progress *TaskProgress) error {
	pe.cacheMutex.Lock()
	oldProgress := pe.cache[progress.TaskID]
	pe.cache[progress.TaskID] = progress
	pe.cacheMutex.Unlock()
	
	// Сохраняем в хранилище
	if err := pe.storage.SaveProgress(progress.TaskID, progress); err != nil {
		return fmt.Errorf("failed to save progress: %w", err)
	}
	
	// Уведомляем подписчиков
	pe.notifyProgressUpdate(oldProgress, progress)
	
	pe.logger.Debug("Task progress saved", 
		"task", progress.TaskID, 
		"stage", progress.CurrentStage,
		"progress", progress.ProgressPercent)
	
	return nil
}

// DeleteTaskProgress удаляет прогресс задачи
func (pe *ProgressEngine) DeleteTaskProgress(taskID string) error {
	pe.cacheMutex.Lock()
	delete(pe.cache, taskID)
	pe.cacheMutex.Unlock()
	
	return pe.storage.DeleteProgress(taskID)
}

// GetAllTaskProgress возвращает прогресс всех задач
func (pe *ProgressEngine) GetAllTaskProgress() map[string]*TaskProgress {
	pe.cacheMutex.RLock()
	defer pe.cacheMutex.RUnlock()
	
	result := make(map[string]*TaskProgress)
	for taskID, progress := range pe.cache {
		result[taskID] = progress
	}
	
	return result
}

// GetTasksByStage возвращает задачи в определенном этапе
func (pe *ProgressEngine) GetTasksByStage(stage string) []*TaskProgress {
	pe.cacheMutex.RLock()
	defer pe.cacheMutex.RUnlock()
	
	var tasks []*TaskProgress
	for _, progress := range pe.cache {
		if progress.CurrentStage == stage {
			tasks = append(tasks, progress)
		}
	}
	
	return tasks
}

// GetStageStatistics возвращает статистику по этапам
func (pe *ProgressEngine) GetStageStatistics() map[string]StageStats {
	pe.cacheMutex.RLock()
	defer pe.cacheMutex.RUnlock()
	
	stats := make(map[string]StageStats)
	
	for _, progress := range pe.cache {
		stage := progress.CurrentStage
		if _, exists := stats[stage]; !exists {
			stats[stage] = StageStats{
				Stage: stage,
				Count: 0,
				AverageProgress: 0,
				TotalTasks: 0,
			}
		}
		
		stageStats := stats[stage]
		stageStats.Count++
		stageStats.TotalTasks++
		stageStats.AverageProgress = (stageStats.AverageProgress*(float64(stageStats.Count-1)) + progress.ProgressPercent) / float64(stageStats.Count)
		stats[stage] = stageStats
	}
	
	return stats
}

// StageStats статистика по этапу
type StageStats struct {
	Stage           string  `json:"stage"`
	Count           int     `json:"count"`
	AverageProgress float64 `json:"average_progress"`
	TotalTasks      int     `json:"total_tasks"`
}

// GetTasksNeedingAttention возвращает задачи, требующие внимания
func (pe *ProgressEngine) GetTasksNeedingAttention() []*TaskProgress {
	pe.cacheMutex.RLock()
	defer pe.cacheMutex.RUnlock()
	
	var needingAttention []*TaskProgress
	now := time.Now()
	
	for _, progress := range pe.cache {
		// Проверяем различные условия
		
		// Долго нет активности
		if now.Sub(progress.LastActivity) > 48*time.Hour {
			needingAttention = append(needingAttention, progress)
			continue
		}
		
		// Застряли на одном этапе
		if progress.ProgressPercent < 10 && len(progress.GitActivity) > 5 {
			needingAttention = append(needingAttention, progress)
			continue
		}
		
		// Много коммитов, но мало прогресса
		if len(progress.GitActivity) > 10 && progress.ProgressPercent < 30 {
			needingAttention = append(needingAttention, progress)
			continue
		}
	}
	
	return needingAttention
}

// Subscribe подписывается на изменения прогресса задачи
func (pe *ProgressEngine) Subscribe(taskID string, subscriber ProgressSubscriber) {
	pe.subsMutex.Lock()
	defer pe.subsMutex.Unlock()
	
	if _, exists := pe.subscribers[taskID]; !exists {
		pe.subscribers[taskID] = []ProgressSubscriber{}
	}
	
	pe.subscribers[taskID] = append(pe.subscribers[taskID], subscriber)
}

// Unsubscribe отписывается от изменений прогресса задачи
func (pe *ProgressEngine) Unsubscribe(taskID string, subscriber ProgressSubscriber) {
	pe.subsMutex.Lock()
	defer pe.subsMutex.Unlock()
	
	if subs, exists := pe.subscribers[taskID]; exists {
		for i, sub := range subs {
			if sub == subscriber {
				pe.subscribers[taskID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}
}

// notifyProgressUpdate уведомляет подписчиков об изменении прогресса
func (pe *ProgressEngine) notifyProgressUpdate(oldProgress, newProgress *TaskProgress) {
	pe.subsMutex.RLock()
	subscribers := pe.subscribers[newProgress.TaskID]
	pe.subsMutex.RUnlock()
	
	if len(subscribers) == 0 {
		return
	}
	
	// Определяем тип события
	var events []*ProgressEvent
	
	if oldProgress == nil {
		// Новая задача
		events = append(events, &ProgressEvent{
			Type:      "task_created",
			TaskID:    newProgress.TaskID,
			NewValue:  newProgress,
			Timestamp: time.Now(),
		})
	} else {
		// Изменение этапа
		if oldProgress.CurrentStage != newProgress.CurrentStage {
			events = append(events, &ProgressEvent{
				Type:      "stage_change",
				TaskID:    newProgress.TaskID,
				OldValue:  oldProgress.CurrentStage,
				NewValue:  newProgress.CurrentStage,
				Timestamp: time.Now(),
			})
		}
		
		// Изменение прогресса
		if oldProgress.ProgressPercent != newProgress.ProgressPercent {
			events = append(events, &ProgressEvent{
				Type:      "progress_update",
				TaskID:    newProgress.TaskID,
				OldValue:  oldProgress.ProgressPercent,
				NewValue:  newProgress.ProgressPercent,
				Timestamp: time.Now(),
			})
		}
		
		// Новая активность
		if len(newProgress.GitActivity) > len(oldProgress.GitActivity) {
			events = append(events, &ProgressEvent{
				Type:      "activity_added",
				TaskID:    newProgress.TaskID,
				NewValue:  newProgress.GitActivity[len(newProgress.GitActivity)-1],
				Timestamp: time.Now(),
			})
		}
	}
	
	// Уведомляем подписчиков
	for _, event := range events {
		for _, subscriber := range subscribers {
			go func(sub ProgressSubscriber, evt *ProgressEvent) {
				defer func() {
					if r := recover(); r != nil {
						pe.logger.Error("Panic in progress subscriber", fmt.Errorf("panic: %v", r))
					}
				}()
				sub.OnProgressUpdate(newProgress.TaskID, newProgress, evt)
			}(subscriber, event)
		}
	}
}

// loadProgressFromStorage загружает прогресс из хранилища
func (pe *ProgressEngine) loadProgressFromStorage() error {
	allProgress, err := pe.storage.LoadAllProgress()
	if err != nil {
		return err
	}
	
	pe.cacheMutex.Lock()
	defer pe.cacheMutex.Unlock()
	
	for taskID, progress := range allProgress {
		pe.cache[taskID] = progress
	}
	
	pe.logger.Info("Loaded progress from storage", "count", len(allProgress))
	return nil
}

// GenerateProgressReport генерирует отчет о прогрессе
func (pe *ProgressEngine) GenerateProgressReport() *ProgressReport {
	pe.cacheMutex.RLock()
	defer pe.cacheMutex.RUnlock()
	
	report := &ProgressReport{
		GeneratedAt:    time.Now(),
		TotalTasks:     len(pe.cache),
		StageStats:     pe.GetStageStatistics(),
		Summary:        make(map[string]interface{}),
	}
	
	// Вычисляем общую статистику
	totalProgress := 0.0
	completedTasks := 0
	activeTasks := 0
	stalledTasks := 0
	
	now := time.Now()
	
	for _, progress := range pe.cache {
		totalProgress += progress.ProgressPercent
		
		if progress.ProgressPercent >= 100 {
			completedTasks++
		} else if now.Sub(progress.LastActivity) > 48*time.Hour {
			stalledTasks++
		} else {
			activeTasks++
		}
	}
	
	if report.TotalTasks > 0 {
		report.AverageProgress = totalProgress / float64(report.TotalTasks)
	}
	
	report.Summary["completed_tasks"] = completedTasks
	report.Summary["active_tasks"] = activeTasks
	report.Summary["stalled_tasks"] = stalledTasks
	report.Summary["completion_rate"] = float64(completedTasks) / float64(report.TotalTasks) * 100
	
	return report
}

// ProgressReport отчет о прогрессе
type ProgressReport struct {
	GeneratedAt     time.Time              `json:"generated_at"`
	TotalTasks      int                    `json:"total_tasks"`
	AverageProgress float64                `json:"average_progress"`
	StageStats      map[string]StageStats  `json:"stage_stats"`
	Summary         map[string]interface{} `json:"summary"`
}

// FileProgressStorage файловое хранилище прогресса
type FileProgressStorage struct {
	dataDir string
	logger  Logger
	mutex   sync.RWMutex
}

// NewFileProgressStorage создает новое файловое хранилище
func NewFileProgressStorage(dataDir string, logger Logger) *FileProgressStorage {
	storage := &FileProgressStorage{
		dataDir: dataDir,
		logger:  logger,
	}
	
	// Создаем директорию, если не существует
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		logger.Error("Failed to create progress data directory", err, "dir", dataDir)
	}
	
	return storage
}

// SaveProgress сохраняет прогресс в файл
func (fps *FileProgressStorage) SaveProgress(taskID string, progress *TaskProgress) error {
	fps.mutex.Lock()
	defer fps.mutex.Unlock()
	
	filename := filepath.Join(fps.dataDir, fmt.Sprintf("%s.json", taskID))
	
	data, err := json.MarshalIndent(progress, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal progress: %w", err)
	}
	
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write progress file: %w", err)
	}
	
	return nil
}

// LoadProgress загружает прогресс из файла
func (fps *FileProgressStorage) LoadProgress(taskID string) (*TaskProgress, error) {
	fps.mutex.RLock()
	defer fps.mutex.RUnlock()
	
	filename := filepath.Join(fps.dataDir, fmt.Sprintf("%s.json", taskID))
	
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("progress not found for task %s", taskID)
		}
		return nil, fmt.Errorf("failed to read progress file: %w", err)
	}
	
	var progress TaskProgress
	if err := json.Unmarshal(data, &progress); err != nil {
		return nil, fmt.Errorf("failed to unmarshal progress: %w", err)
	}
	
	return &progress, nil
}

// LoadAllProgress загружает весь прогресс
func (fps *FileProgressStorage) LoadAllProgress() (map[string]*TaskProgress, error) {
	fps.mutex.RLock()
	defer fps.mutex.RUnlock()
	
	result := make(map[string]*TaskProgress)
	
	files, err := os.ReadDir(fps.dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, fmt.Errorf("failed to read progress directory: %w", err)
	}
	
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		
		taskID := strings.TrimSuffix(file.Name(), ".json")
		progress, err := fps.LoadProgress(taskID)
		if err != nil {
			fps.logger.Error("Failed to load progress file", err, "file", file.Name())
			continue
		}
		
		result[taskID] = progress
	}
	
	return result, nil
}

// DeleteProgress удаляет прогресс
func (fps *FileProgressStorage) DeleteProgress(taskID string) error {
	fps.mutex.Lock()
	defer fps.mutex.Unlock()
	
	filename := filepath.Join(fps.dataDir, fmt.Sprintf("%s.json", taskID))
	
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete progress file: %w", err)
	}
	
	return nil
}