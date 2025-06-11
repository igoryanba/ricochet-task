package orchestrator

import (
	"context"
	"errors"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/checkpoint"
)

// Статусы выполнения задачи
type RunStatus string

const (
	StatusPending    RunStatus = "pending"
	StatusRunning    RunStatus = "running"
	StatusCompleted  RunStatus = "completed"
	StatusFailed     RunStatus = "failed"
	StatusCancelled  RunStatus = "cancelled"
	StatusProcessing RunStatus = "processing" // Промежуточное состояние при обработке отдельной модели
)

// RunMetadata содержит метаданные о выполнении цепочки
type RunMetadata struct {
	ID            string                 `json:"id"`
	ChainID       string                 `json:"chain_id"`
	Status        RunStatus              `json:"status"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       time.Time              `json:"end_time,omitempty"`
	Progress      float64                `json:"progress"`
	CurrentModel  string                 `json:"current_model,omitempty"`
	TotalTokens   int                    `json:"total_tokens"`
	Error         string                 `json:"error,omitempty"`
	Checkpoints   []string               `json:"checkpoints"` // ID чекпоинтов
	ExtraMetadata map[string]interface{} `json:"extra_metadata,omitempty"`
}

// ProcessingOptions содержит опции для обработки цепочки
type ProcessingOptions struct {
	MaxParallelChunks  int    `json:"max_parallel_chunks"`
	MaxTokensPerChunk  int    `json:"max_tokens_per_chunk"`
	SegmentationMethod string `json:"segmentation_method"` // simple, semantic, recursive
	SaveCheckpoints    bool   `json:"save_checkpoints"`
	AutoRetry          bool   `json:"auto_retry"`
	RetryAttempts      int    `json:"retry_attempts"`
	RetryDelay         int    `json:"retry_delay"` // в секундах
}

// DefaultProcessingOptions возвращает настройки по умолчанию
func DefaultProcessingOptions() ProcessingOptions {
	return ProcessingOptions{
		MaxParallelChunks:  3,
		MaxTokensPerChunk:  2000,
		SegmentationMethod: "simple",
		SaveCheckpoints:    true,
		AutoRetry:          true,
		RetryAttempts:      3,
		RetryDelay:         5,
	}
}

// TaskInput представляет входные данные для задачи
type TaskInput struct {
	Text     string                 `json:"text"`
	Files    []string               `json:"files,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TaskOutput представляет выходные данные задачи
type TaskOutput struct {
	Text     string                 `json:"text"`
	Files    []string               `json:"files,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Orchestrator определяет основной интерфейс оркестратора
type Orchestrator interface {
	// RunChain запускает цепочку моделей с указанными входными данными
	RunChain(ctx context.Context, chainID string, input TaskInput, options ProcessingOptions) (string, error)

	// GetRunStatus возвращает статус выполнения
	GetRunStatus(runID string) (*RunMetadata, error)

	// CancelRun отменяет выполнение
	CancelRun(runID string) error

	// ListRuns возвращает список всех выполнений
	ListRuns() []*RunMetadata

	// GetRunResults возвращает результаты выполнения
	GetRunResults(runID string) (TaskOutput, error)

	// GetCheckpoint возвращает чекпоинт
	GetCheckpoint(checkpointID string) (checkpoint.Checkpoint, error)

	// ListCheckpoints возвращает список чекпоинтов для указанного выполнения
	ListCheckpoints(runID string) ([]checkpoint.Checkpoint, error)
}

// Errors
var (
	ErrChainNotFound = errors.New("chain not found")
	ErrRunNotFound   = errors.New("run not found")
	ErrRunCancelled  = errors.New("run cancelled")
	ErrInvalidInput  = errors.New("invalid input")
)
