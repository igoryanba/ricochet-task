package key

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Key представляет API-ключ
type Key struct {
	ID         string    `json:"id"`           // Уникальный идентификатор
	Provider   string    `json:"provider"`     // Провайдер (openai, anthropic, и т.д.)
	Value      string    `json:"value"`        // Значение ключа (зашифрованное)
	Name       string    `json:"name"`         // Пользовательское имя ключа
	CreatedAt  time.Time `json:"created_at"`   // Время создания
	LastUsedAt time.Time `json:"last_used_at"` // Время последнего использования
	Usage      KeyUsage  `json:"usage"`        // Статистика использования
	Shared     bool      `json:"shared"`       // Флаг общего доступа
	Metadata   Metadata  `json:"metadata"`     // Дополнительные метаданные
}

// KeyUsage статистика использования ключа
type KeyUsage struct {
	TotalRequests int       `json:"total_requests"` // Общее количество запросов
	TotalTokens   int       `json:"total_tokens"`   // Общее количество токенов
	LastRequest   time.Time `json:"last_request"`   // Время последнего запроса
	MonthlyTokens int       `json:"monthly_tokens"` // Количество токенов за текущий месяц
	DailyTokens   int       `json:"daily_tokens"`   // Количество токенов за текущий день
}

// Metadata дополнительные метаданные ключа
type Metadata struct {
	Quota      int               `json:"quota"`      // Квота токенов
	Expiration time.Time         `json:"expiration"` // Срок действия
	Custom     map[string]string `json:"custom"`     // Пользовательские метаданные
}

// Store интерфейс для хранилища ключей
type Store interface {
	// Save сохраняет ключ
	Save(key Key) error

	// Get возвращает ключ по ID
	Get(id string) (Key, error)

	// List возвращает список всех ключей
	List() ([]Key, error)

	// Delete удаляет ключ
	Delete(id string) error

	// Exists проверяет существование ключа
	Exists(id string) bool

	// GetByProvider возвращает список ключей для указанного провайдера
	GetByProvider(provider string) ([]Key, error)
}

// FileKeyStore реализация хранилища ключей в файле
type FileKeyStore struct {
	path string
}

// NewFileKeyStore создает новое хранилище ключей в файле
func NewFileKeyStore(configDir string) (*FileKeyStore, error) {
	path := filepath.Join(configDir, "keys.json")

	// Создаем директорию, если она не существует
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("не удалось создать директорию для ключей: %w", err)
	}

	// Создаем файл, если он не существует
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := saveKeys(path, []Key{}); err != nil {
			return nil, fmt.Errorf("не удалось создать файл ключей: %w", err)
		}
	}

	return &FileKeyStore{path: path}, nil
}

// Add добавляет новый ключ
func (s *FileKeyStore) Add(key Key) error {
	keys, err := s.List()
	if err != nil {
		return err
	}

	// Проверяем, существует ли ключ с таким ID
	for _, k := range keys {
		if k.ID == key.ID {
			return fmt.Errorf("ключ с ID %s уже существует", key.ID)
		}
	}

	keys = append(keys, key)
	return saveKeys(s.path, keys)
}

// Get возвращает ключ по ID
func (s *FileKeyStore) Get(id string) (Key, error) {
	keys, err := s.List()
	if err != nil {
		return Key{}, err
	}

	for _, key := range keys {
		if key.ID == id {
			return key, nil
		}
	}

	return Key{}, fmt.Errorf("ключ с ID %s не найден", id)
}

// List возвращает список всех ключей
func (s *FileKeyStore) List() ([]Key, error) {
	return loadKeys(s.path)
}

// Update обновляет существующий ключ
func (s *FileKeyStore) Update(key Key) error {
	keys, err := s.List()
	if err != nil {
		return err
	}

	for i, k := range keys {
		if k.ID == key.ID {
			keys[i] = key
			return saveKeys(s.path, keys)
		}
	}

	return fmt.Errorf("ключ с ID %s не найден", key.ID)
}

// Delete удаляет ключ по ID
func (s *FileKeyStore) Delete(id string) error {
	keys, err := s.List()
	if err != nil {
		return err
	}

	var newKeys []Key
	for _, k := range keys {
		if k.ID != id {
			newKeys = append(newKeys, k)
		}
	}

	return saveKeys(s.path, newKeys)
}

// loadKeys загружает ключи из файла
func loadKeys(path string) ([]Key, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл ключей: %w", err)
	}

	var keys []Key
	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, fmt.Errorf("не удалось распарсить файл ключей: %w", err)
	}

	return keys, nil
}

// saveKeys сохраняет ключи в файл
func saveKeys(path string, keys []Key) error {
	data, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return fmt.Errorf("не удалось сериализовать ключи: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("не удалось записать файл ключей: %w", err)
	}

	return nil
}
