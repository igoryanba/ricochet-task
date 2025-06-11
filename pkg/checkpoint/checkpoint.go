package checkpoint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// CheckpointType определяет тип чекпоинта
type CheckpointType string

const (
	CheckpointTypeInput        CheckpointType = "input"        // Входные данные
	CheckpointTypeOutput       CheckpointType = "output"       // Выходные данные
	CheckpointTypeIntermediate CheckpointType = "intermediate" // Промежуточный результат
	CheckpointTypeError        CheckpointType = "error"        // Ошибка
	CheckpointTypeSegment      CheckpointType = "segment"      // Сегмент
	CheckpointTypeComplete     CheckpointType = "complete"     // Завершенный результат
)

// StorageType определяет тип хранилища для чекпоинтов
type StorageType string

const (
	StorageTypeLocal StorageType = "local" // Локальная файловая система
	StorageTypeMinio StorageType = "minio" // MinIO хранилище
)

// Checkpoint представляет чекпоинт цепочки моделей
type Checkpoint struct {
	ID          string                 `json:"id"`
	ChainID     string                 `json:"chain_id"`
	ModelID     string                 `json:"model_id,omitempty"` // Может быть пустым для входных/выходных данных
	Type        CheckpointType         `json:"type"`
	Content     string                 `json:"content"`
	ContentPath string                 `json:"content_path,omitempty"` // Путь к файлу с содержимым, если оно большое
	StorageType StorageType            `json:"storage_type"`           // Тип хранилища (local, minio)
	CreatedAt   time.Time              `json:"created_at"`
	MetaData    map[string]interface{} `json:"metadata"`
}

// Store определяет интерфейс для работы с хранилищем чекпоинтов
type Store interface {
	// Save сохраняет чекпоинт
	Save(checkpoint Checkpoint) error

	// Get возвращает чекпоинт по ID
	Get(id string) (Checkpoint, error)

	// List возвращает список чекпоинтов для указанной цепочки
	List(chainID string) ([]Checkpoint, error)

	// Delete удаляет чекпоинт
	Delete(id string) error

	// DeleteByChain удаляет все чекпоинты для указанной цепочки
	DeleteByChain(chainID string) error
}

// FileCheckpointStore реализует хранилище чекпоинтов в файловой системе
type FileCheckpointStore struct {
	metadataPath string
	contentPath  string
}

// NewFileCheckpointStore создает новое хранилище чекпоинтов в файловой системе
func NewFileCheckpointStore(configDir string) (*FileCheckpointStore, error) {
	// Путь к метаданным чекпоинтов
	metadataPath := filepath.Join(configDir, "checkpoints", "metadata.json")

	// Путь к содержимому чекпоинтов
	contentPath := filepath.Join(configDir, "checkpoints", "content")

	// Создаем директории, если они не существуют
	if err := os.MkdirAll(filepath.Dir(metadataPath), 0755); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(contentPath, 0755); err != nil {
		return nil, err
	}

	// Создаем файл метаданных, если он не существует
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		if err := saveCheckpointMetadata(metadataPath, []Checkpoint{}); err != nil {
			return nil, err
		}
	}

	return &FileCheckpointStore{
		metadataPath: metadataPath,
		contentPath:  contentPath,
	}, nil
}

// Save сохраняет чекпоинт
func (s *FileCheckpointStore) Save(checkpoint Checkpoint) error {
	// Загружаем текущие метаданные
	checkpoints, err := loadCheckpointMetadata(s.metadataPath)
	if err != nil {
		return err
	}

	// Для нового чекпоинта генерируем ID, если он не указан
	if checkpoint.ID == "" {
		checkpoint.ID = uuid.New().String()
	}

	// Сохраняем содержимое чекпоинта в отдельный файл, если оно большое
	contentSize := len(checkpoint.Content)
	if contentSize > 1024*10 { // Если содержимое больше 10KB
		contentFilePath := filepath.Join(s.contentPath, checkpoint.ID+".txt")
		err := ioutil.WriteFile(contentFilePath, []byte(checkpoint.Content), 0644)
		if err != nil {
			return err
		}

		// Устанавливаем путь к файлу содержимого
		checkpoint.ContentPath = contentFilePath
		// Очищаем содержимое в метаданных
		checkpoint.Content = ""
		// Устанавливаем тип хранилища
		checkpoint.StorageType = StorageTypeLocal
	} else {
		// Если содержимое небольшое, храним его прямо в метаданных
		checkpoint.ContentPath = ""
		checkpoint.StorageType = StorageTypeLocal
	}

	// Проверяем, существует ли уже чекпоинт с таким ID
	found := false
	for i, c := range checkpoints {
		if c.ID == checkpoint.ID {
			// Обновляем существующий чекпоинт
			checkpoints[i] = checkpoint
			found = true
			break
		}
	}

	// Если чекпоинт не найден, добавляем его
	if !found {
		checkpoints = append(checkpoints, checkpoint)
	}

	// Сохраняем обновленные метаданные
	return saveCheckpointMetadata(s.metadataPath, checkpoints)
}

// Get возвращает чекпоинт по ID
func (s *FileCheckpointStore) Get(id string) (Checkpoint, error) {
	// Загружаем метаданные
	checkpoints, err := loadCheckpointMetadata(s.metadataPath)
	if err != nil {
		return Checkpoint{}, err
	}

	// Ищем чекпоинт с указанным ID
	for _, c := range checkpoints {
		if c.ID == id {
			// Если содержимое хранится в отдельном файле, загружаем его
			if c.ContentPath != "" && c.Content == "" {
				content, err := ioutil.ReadFile(c.ContentPath)
				if err != nil {
					return Checkpoint{}, err
				}
				c.Content = string(content)
			}

			return c, nil
		}
	}

	return Checkpoint{}, fmt.Errorf("checkpoint with ID '%s' not found", id)
}

// List возвращает список чекпоинтов для указанной цепочки
func (s *FileCheckpointStore) List(chainID string) ([]Checkpoint, error) {
	// Загружаем метаданные
	checkpoints, err := loadCheckpointMetadata(s.metadataPath)
	if err != nil {
		return nil, err
	}

	// Фильтруем чекпоинты по ID цепочки
	var result []Checkpoint
	for _, c := range checkpoints {
		if c.ChainID == chainID {
			// Не загружаем содержимое для списка
			c.Content = ""
			result = append(result, c)
		}
	}

	return result, nil
}

// Delete удаляет чекпоинт
func (s *FileCheckpointStore) Delete(id string) error {
	// Загружаем метаданные
	checkpoints, err := loadCheckpointMetadata(s.metadataPath)
	if err != nil {
		return err
	}

	// Ищем чекпоинт с указанным ID
	found := false
	var contentPath string
	var filteredCheckpoints []Checkpoint
	for _, c := range checkpoints {
		if c.ID == id {
			found = true
			contentPath = c.ContentPath
		} else {
			filteredCheckpoints = append(filteredCheckpoints, c)
		}
	}

	if !found {
		return fmt.Errorf("checkpoint with ID '%s' not found", id)
	}

	// Если содержимое хранится в отдельном файле, удаляем его
	if contentPath != "" {
		err := os.Remove(contentPath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	// Сохраняем обновленные метаданные
	return saveCheckpointMetadata(s.metadataPath, filteredCheckpoints)
}

// DeleteByChain удаляет все чекпоинты для указанной цепочки
func (s *FileCheckpointStore) DeleteByChain(chainID string) error {
	// Загружаем метаданные
	checkpoints, err := loadCheckpointMetadata(s.metadataPath)
	if err != nil {
		return err
	}

	// Фильтруем чекпоинты по ID цепочки
	var filteredCheckpoints []Checkpoint
	for _, c := range checkpoints {
		if c.ChainID == chainID {
			// Если содержимое хранится в отдельном файле, удаляем его
			if c.ContentPath != "" {
				err := os.Remove(c.ContentPath)
				if err != nil && !os.IsNotExist(err) {
					return err
				}
			}
		} else {
			filteredCheckpoints = append(filteredCheckpoints, c)
		}
	}

	// Сохраняем обновленные метаданные
	return saveCheckpointMetadata(s.metadataPath, filteredCheckpoints)
}

// MinioCheckpointStore реализует хранилище чекпоинтов в MinIO
type MinioCheckpointStore struct {
	metadataPath string             // Путь к локальному файлу метаданных
	endpoint     string             // Адрес MinIO сервера
	accessKey    string             // Ключ доступа
	secretKey    string             // Секретный ключ
	bucketName   string             // Имя бакета
	useSSL       bool               // Использовать SSL
	client       minioClientWrapper // Обертка над клиентом MinIO
}

// minioClientWrapper определяет интерфейс для работы с MinIO
// Это позволяет легко мокать клиент для тестирования
type minioClientWrapper interface {
	PutObject(bucketName, objectName string, reader *os.File, contentType string) error
	GetObject(bucketName, objectName string) ([]byte, error)
	RemoveObject(bucketName, objectName string) error
}

// NewMinioCheckpointStore создает новое хранилище чекпоинтов в MinIO
func NewMinioCheckpointStore(
	configDir string,
	endpoint string,
	accessKey string,
	secretKey string,
	bucketName string,
	useSSL bool,
) (*MinioCheckpointStore, error) {
	// TODO: Реализовать создание клиента MinIO и проверку доступа к бакету

	// Пока просто возвращаем заглушку
	return &MinioCheckpointStore{
		metadataPath: filepath.Join(configDir, "checkpoints", "metadata.json"),
		endpoint:     endpoint,
		accessKey:    accessKey,
		secretKey:    secretKey,
		bucketName:   bucketName,
		useSSL:       useSSL,
		client:       nil, // TODO: Инициализировать клиент MinIO
	}, nil
}

// TODO: Реализовать методы MinioCheckpointStore (Save, Get, List, Delete, DeleteByChain)

// loadCheckpointMetadata загружает метаданные чекпоинтов из файла
func loadCheckpointMetadata(path string) ([]Checkpoint, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Checkpoint{}, nil
		}
		return nil, err
	}

	var checkpoints []Checkpoint
	err = json.Unmarshal(data, &checkpoints)
	if err != nil {
		return nil, err
	}

	return checkpoints, nil
}

// saveCheckpointMetadata сохраняет метаданные чекпоинтов в файл
func saveCheckpointMetadata(path string, checkpoints []Checkpoint) error {
	data, err := json.MarshalIndent(checkpoints, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}
