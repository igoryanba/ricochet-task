package orchestrator

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresRunStore хранилище метаданных запусков в PostgreSQL
type PostgresRunStore struct {
	db *sql.DB
}

// NewPostgresRunStore создает новое хранилище запусков в PostgreSQL
func NewPostgresRunStore(dsn string) (*PostgresRunStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &PostgresRunStore{db: db}

	// Создаем таблицы если они не существуют
	if err := store.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return store, nil
}

// createTables создает необходимые таблицы
func (s *PostgresRunStore) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS chain_runs (
		id VARCHAR(255) PRIMARY KEY,
		chain_id VARCHAR(255) NOT NULL,
		status VARCHAR(50) NOT NULL,
		start_time TIMESTAMP NOT NULL,
		end_time TIMESTAMP,
		progress FLOAT DEFAULT 0,
		current_model VARCHAR(255),
		total_tokens INTEGER DEFAULT 0,
		error_message TEXT,
		checkpoints JSONB DEFAULT '[]',
		extra_metadata JSONB DEFAULT '{}',
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_chain_runs_chain_id ON chain_runs(chain_id);
	CREATE INDEX IF NOT EXISTS idx_chain_runs_status ON chain_runs(status);
	CREATE INDEX IF NOT EXISTS idx_chain_runs_start_time ON chain_runs(start_time);
	CREATE INDEX IF NOT EXISTS idx_chain_runs_created_at ON chain_runs(created_at);
	`

	_, err := s.db.Exec(query)
	return err
}

// SaveRunMetadata сохраняет метаданные запуска
func (s *PostgresRunStore) SaveRunMetadata(metadata *RunMetadata) error {
	checkpointsJSON, err := json.Marshal(metadata.Checkpoints)
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoints: %w", err)
	}

	extraMetadataJSON, err := json.Marshal(metadata.ExtraMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal extra metadata: %w", err)
	}

	query := `
	INSERT INTO chain_runs (
		id, chain_id, status, start_time, end_time, progress, 
		current_model, total_tokens, error_message, checkpoints, 
		extra_metadata, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	ON CONFLICT (id) DO UPDATE SET
		status = $3,
		end_time = $5,
		progress = $6,
		current_model = $7,
		total_tokens = $8,
		error_message = $9,
		checkpoints = $10,
		extra_metadata = $11,
		updated_at = $12
	`

	_, err = s.db.Exec(query,
		metadata.ID,
		metadata.ChainID,
		string(metadata.Status),
		metadata.StartTime,
		nullTimeFromTime(metadata.EndTime),
		metadata.Progress,
		nullStringFromString(metadata.CurrentModel),
		metadata.TotalTokens,
		nullStringFromString(metadata.Error),
		checkpointsJSON,
		extraMetadataJSON,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to save run metadata: %w", err)
	}

	return nil
}

// GetRunMetadata возвращает метаданные запуска по ID
func (s *PostgresRunStore) GetRunMetadata(runID string) (*RunMetadata, error) {
	var metadata RunMetadata
	var checkpointsJSON, extraMetadataJSON []byte
	var endTime sql.NullTime
	var currentModel, errorMessage sql.NullString

	query := `
	SELECT id, chain_id, status, start_time, end_time, progress,
		   current_model, total_tokens, error_message, checkpoints, extra_metadata
	FROM chain_runs WHERE id = $1
	`

	row := s.db.QueryRow(query, runID)
	err := row.Scan(
		&metadata.ID,
		&metadata.ChainID,
		&metadata.Status,
		&metadata.StartTime,
		&endTime,
		&metadata.Progress,
		&currentModel,
		&metadata.TotalTokens,
		&errorMessage,
		&checkpointsJSON,
		&extraMetadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("run with ID '%s' not found", runID)
		}
		return nil, fmt.Errorf("failed to get run metadata: %w", err)
	}

	// Обрабатываем nullable поля
	if endTime.Valid {
		metadata.EndTime = endTime.Time
	}
	if currentModel.Valid {
		metadata.CurrentModel = currentModel.String
	}
	if errorMessage.Valid {
		metadata.Error = errorMessage.String
	}

	// Десериализуем JSON поля
	if err := json.Unmarshal(checkpointsJSON, &metadata.Checkpoints); err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoints: %w", err)
	}

	if err := json.Unmarshal(extraMetadataJSON, &metadata.ExtraMetadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal extra metadata: %w", err)
	}

	return &metadata, nil
}

// ListRunsForChain возвращает список запусков для цепочки
func (s *PostgresRunStore) ListRunsForChain(chainID string, limit int) ([]*RunMetadata, error) {
	query := `
	SELECT id, chain_id, status, start_time, end_time, progress,
		   current_model, total_tokens, error_message, checkpoints, extra_metadata
	FROM chain_runs 
	WHERE chain_id = $1 
	ORDER BY start_time DESC
	`

	args := []interface{}{chainID}
	if limit > 0 {
		query += " LIMIT $2"
		args = append(args, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list runs for chain: %w", err)
	}
	defer rows.Close()

	var runs []*RunMetadata
	for rows.Next() {
		var metadata RunMetadata
		var checkpointsJSON, extraMetadataJSON []byte
		var endTime sql.NullTime
		var currentModel, errorMessage sql.NullString

		err := rows.Scan(
			&metadata.ID,
			&metadata.ChainID,
			&metadata.Status,
			&metadata.StartTime,
			&endTime,
			&metadata.Progress,
			&currentModel,
			&metadata.TotalTokens,
			&errorMessage,
			&checkpointsJSON,
			&extraMetadataJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan run metadata: %w", err)
		}

		// Обрабатываем nullable поля
		if endTime.Valid {
			metadata.EndTime = endTime.Time
		}
		if currentModel.Valid {
			metadata.CurrentModel = currentModel.String
		}
		if errorMessage.Valid {
			metadata.Error = errorMessage.String
		}

		// Десериализуем JSON поля
		if err := json.Unmarshal(checkpointsJSON, &metadata.Checkpoints); err != nil {
			return nil, fmt.Errorf("failed to unmarshal checkpoints: %w", err)
		}

		if err := json.Unmarshal(extraMetadataJSON, &metadata.ExtraMetadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal extra metadata: %w", err)
		}

		runs = append(runs, &metadata)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return runs, nil
}

// ListAllRuns возвращает все запуски
func (s *PostgresRunStore) ListAllRuns(limit int) ([]*RunMetadata, error) {
	query := `
	SELECT id, chain_id, status, start_time, end_time, progress,
		   current_model, total_tokens, error_message, checkpoints, extra_metadata
	FROM chain_runs 
	ORDER BY start_time DESC
	`

	args := []interface{}{}
	if limit > 0 {
		query += " LIMIT $1"
		args = append(args, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list all runs: %w", err)
	}
	defer rows.Close()

	var runs []*RunMetadata
	for rows.Next() {
		var metadata RunMetadata
		var checkpointsJSON, extraMetadataJSON []byte
		var endTime sql.NullTime
		var currentModel, errorMessage sql.NullString

		err := rows.Scan(
			&metadata.ID,
			&metadata.ChainID,
			&metadata.Status,
			&metadata.StartTime,
			&endTime,
			&metadata.Progress,
			&currentModel,
			&metadata.TotalTokens,
			&errorMessage,
			&checkpointsJSON,
			&extraMetadataJSON,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan run metadata: %w", err)
		}

		// Обрабатываем nullable поля
		if endTime.Valid {
			metadata.EndTime = endTime.Time
		}
		if currentModel.Valid {
			metadata.CurrentModel = currentModel.String
		}
		if errorMessage.Valid {
			metadata.Error = errorMessage.String
		}

		// Десериализуем JSON поля
		if err := json.Unmarshal(checkpointsJSON, &metadata.Checkpoints); err != nil {
			return nil, fmt.Errorf("failed to unmarshal checkpoints: %w", err)
		}

		if err := json.Unmarshal(extraMetadataJSON, &metadata.ExtraMetadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal extra metadata: %w", err)
		}

		runs = append(runs, &metadata)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return runs, nil
}

// DeleteRunMetadata удаляет метаданные запуска
func (s *PostgresRunStore) DeleteRunMetadata(runID string) error {
	query := `DELETE FROM chain_runs WHERE id = $1`
	result, err := s.db.Exec(query, runID)
	if err != nil {
		return fmt.Errorf("failed to delete run metadata: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("run with ID '%s' not found", runID)
	}

	return nil
}

// GetRunStatistics возвращает статистику для цепочки
func (s *PostgresRunStore) GetRunStatistics(chainID string) (*RunStatistics, error) {
	query := `
	SELECT 
		COUNT(*) as total_runs,
		COUNT(CASE WHEN status = 'completed' THEN 1 END) as successful_runs,
		COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_runs,
		AVG(EXTRACT(EPOCH FROM (end_time - start_time)) * 1000) as avg_duration_ms,
		MIN(EXTRACT(EPOCH FROM (end_time - start_time)) * 1000) as min_duration_ms,
		MAX(EXTRACT(EPOCH FROM (end_time - start_time)) * 1000) as max_duration_ms,
		AVG(total_tokens) as avg_tokens,
		SUM(total_tokens) as total_tokens,
		MAX(start_time) as last_run_date,
		MAX(CASE WHEN status = 'completed' THEN start_time END) as last_successful_date
	FROM chain_runs 
	WHERE chain_id = $1 AND end_time IS NOT NULL
	`

	var stats RunStatistics
	var avgDuration, minDuration, maxDuration sql.NullFloat64
	var avgTokens sql.NullFloat64
	var lastRunDate, lastSuccessfulDate sql.NullTime

	row := s.db.QueryRow(query, chainID)
	err := row.Scan(
		&stats.TotalRuns,
		&stats.SuccessfulRuns,
		&stats.FailedRuns,
		&avgDuration,
		&minDuration,
		&maxDuration,
		&avgTokens,
		&stats.TotalTokensUsed,
		&lastRunDate,
		&lastSuccessfulDate,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get run statistics: %w", err)
	}

	// Обрабатываем nullable поля
	if avgDuration.Valid {
		stats.AverageDuration = int64(avgDuration.Float64)
	}
	if minDuration.Valid {
		stats.MinDuration = int64(minDuration.Float64)
	}
	if maxDuration.Valid {
		stats.MaxDuration = int64(maxDuration.Float64)
	}
	if avgTokens.Valid {
		stats.AverageTokensUsed = int(avgTokens.Float64)
	}

	// Вычисляем успешность
	if stats.TotalRuns > 0 {
		stats.SuccessRate = float64(stats.SuccessfulRuns) / float64(stats.TotalRuns) * 100
	}

	// Примерная стоимость ($0.02 за 1000 токенов)
	stats.EstimatedCost = float64(stats.TotalTokensUsed) * 0.02 / 1000

	// Форматируем даты
	if lastRunDate.Valid {
		stats.LastRunDate = lastRunDate.Time.Format(time.RFC3339)
	}
	if lastSuccessfulDate.Valid {
		stats.LastSuccessfulDate = lastSuccessfulDate.Time.Format(time.RFC3339)
	}

	return &stats, nil
}

// Close закрывает соединение с базой данных
func (s *PostgresRunStore) Close() error {
	return s.db.Close()
}

// Вспомогательные функции для работы с nullable типами
func nullTimeFromTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func nullStringFromString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}