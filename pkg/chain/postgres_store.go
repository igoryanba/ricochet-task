package chain

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresChainStore реализация Store для PostgreSQL
type PostgresChainStore struct {
	db *sql.DB
}

// NewPostgresChainStore создает новое хранилище цепочек в PostgreSQL
func NewPostgresChainStore(dsn string) (*PostgresChainStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &PostgresChainStore{db: db}

	// Создаем таблицы если они не существуют
	if err := store.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return store, nil
}

// createTables создает необходимые таблицы
func (s *PostgresChainStore) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS chains (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		models JSONB,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_chains_name ON chains(name);
	CREATE INDEX IF NOT EXISTS idx_chains_created_at ON chains(created_at);
	`

	_, err := s.db.Exec(query)
	return err
}

// Save сохраняет цепочку
func (s *PostgresChainStore) Save(chain Chain) error {
	modelsJSON, err := json.Marshal(chain.Models)
	if err != nil {
		return fmt.Errorf("failed to marshal models: %w", err)
	}

	query := `
	INSERT INTO chains (id, name, description, models, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (id) DO UPDATE SET
		name = $2,
		description = $3,
		models = $4,
		updated_at = $6
	`

	_, err = s.db.Exec(query, 
		chain.ID, 
		chain.Name, 
		chain.Description, 
		modelsJSON, 
		chain.CreatedAt, 
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to save chain: %w", err)
	}

	return nil
}

// Get возвращает цепочку по ID
func (s *PostgresChainStore) Get(id string) (Chain, error) {
	var chain Chain
	var modelsJSON []byte

	query := `SELECT id, name, description, models, created_at FROM chains WHERE id = $1`
	row := s.db.QueryRow(query, id)

	err := row.Scan(&chain.ID, &chain.Name, &chain.Description, &modelsJSON, &chain.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Chain{}, fmt.Errorf("chain with ID '%s' not found", id)
		}
		return Chain{}, fmt.Errorf("failed to get chain: %w", err)
	}

	if err := json.Unmarshal(modelsJSON, &chain.Models); err != nil {
		return Chain{}, fmt.Errorf("failed to unmarshal models: %w", err)
	}

	return chain, nil
}

// List возвращает список всех цепочек
func (s *PostgresChainStore) List() ([]Chain, error) {
	query := `SELECT id, name, description, models, created_at FROM chains ORDER BY created_at DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list chains: %w", err)
	}
	defer rows.Close()

	var chains []Chain
	for rows.Next() {
		var chain Chain
		var modelsJSON []byte

		err := rows.Scan(&chain.ID, &chain.Name, &chain.Description, &modelsJSON, &chain.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chain: %w", err)
		}

		if err := json.Unmarshal(modelsJSON, &chain.Models); err != nil {
			return nil, fmt.Errorf("failed to unmarshal models: %w", err)
		}

		chains = append(chains, chain)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return chains, nil
}

// Delete удаляет цепочку
func (s *PostgresChainStore) Delete(id string) error {
	query := `DELETE FROM chains WHERE id = $1`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete chain: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("chain with ID '%s' not found", id)
	}

	return nil
}

// Exists проверяет существование цепочки
func (s *PostgresChainStore) Exists(id string) bool {
	query := `SELECT EXISTS(SELECT 1 FROM chains WHERE id = $1)`
	var exists bool
	err := s.db.QueryRow(query, id).Scan(&exists)
	return err == nil && exists
}

// Close закрывает соединение с базой данных
func (s *PostgresChainStore) Close() error {
	return s.db.Close()
}