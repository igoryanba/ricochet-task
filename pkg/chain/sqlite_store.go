package chain

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteChainStore реализует Store на основе SQLite.
// Хранит цепочку как JSON-документ; этого достаточно для MVP.
// Схема: chains(id TEXT PRIMARY KEY, data TEXT, updated_at TIMESTAMP)

type SQLiteChainStore struct {
	db *sql.DB
}

// NewSQLiteChainStore открывает (или создаёт) БД по указанному пути.
func NewSQLiteChainStore(dbPath string) (*SQLiteChainStore, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// минимальный пул
	db.SetMaxOpenConns(1)

	schema := `CREATE TABLE IF NOT EXISTS chains (
        id TEXT PRIMARY KEY,
        data TEXT NOT NULL,
        updated_at TIMESTAMP NOT NULL
    );`
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &SQLiteChainStore{db: db}, nil
}

// Save реализует Store.Save
func (s *SQLiteChainStore) Save(chain Chain) error {
	if chain.ID == "" {
		return fmt.Errorf("chain ID empty")
	}
	chain.UpdatedAt = time.Now()
	if chain.CreatedAt.IsZero() {
		chain.CreatedAt = chain.UpdatedAt
	}
	blob, err := json.Marshal(chain)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO chains(id,data,updated_at) VALUES(?,?,?)
        ON CONFLICT(id) DO UPDATE SET data=excluded.data, updated_at=excluded.updated_at`, chain.ID, string(blob), chain.UpdatedAt)
	return err
}

// Get реализует Store.Get
func (s *SQLiteChainStore) Get(id string) (Chain, error) {
	var data string
	row := s.db.QueryRow("SELECT data FROM chains WHERE id=?", id)
	if err := row.Scan(&data); err != nil {
		if err == sql.ErrNoRows {
			return Chain{}, fmt.Errorf("chain not found: %s", id)
		}
		return Chain{}, err
	}
	var c Chain
	if err := json.Unmarshal([]byte(data), &c); err != nil {
		return Chain{}, err
	}
	return c, nil
}

// List реализует Store.List
func (s *SQLiteChainStore) List() ([]Chain, error) {
	rows, err := s.db.Query("SELECT data FROM chains")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Chain
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var c Chain
		if err := json.Unmarshal([]byte(data), &c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

// Delete реализует Store.Delete
func (s *SQLiteChainStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM chains WHERE id=?", id)
	return err
}

// Exists реализует Store.Exists
func (s *SQLiteChainStore) Exists(id string) bool {
	var exists int
	_ = s.db.QueryRow("SELECT 1 FROM chains WHERE id=?", id).Scan(&exists)
	return exists == 1
}
