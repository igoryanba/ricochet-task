package checkpoint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOCheckpointStore реализация Store для MinIO
type MinIOCheckpointStore struct {
	client     *minio.Client
	bucketName string
}

// MinIOConfig конфигурация для MinIO
type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

// NewMinIOCheckpointStore создает новое хранилище чекпоинтов в MinIO
func NewMinIOCheckpointStore(config MinIOConfig) (*MinIOCheckpointStore, error) {
	// Инициализируем MinIO клиент
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	ctx := context.Background()

	// Проверяем существование bucket и создаем если нужно
	exists, err := client.BucketExists(ctx, config.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, config.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &MinIOCheckpointStore{
		client:     client,
		bucketName: config.BucketName,
	}, nil
}

// Save сохраняет чекпоинт
func (s *MinIOCheckpointStore) Save(checkpoint Checkpoint) error {
	ctx := context.Background()

	// Обновляем тип хранилища
	checkpoint.StorageType = StorageTypeMinio

	// Для больших чекпоинтов сохраняем содержимое отдельно
	if len(checkpoint.Content) > 1024*1024 { // > 1MB
		contentPath := fmt.Sprintf("content/%s/%s.txt", checkpoint.ChainID, checkpoint.ID)
		
		contentReader := strings.NewReader(checkpoint.Content)
		_, err := s.client.PutObject(ctx, s.bucketName, contentPath, contentReader, 
			int64(len(checkpoint.Content)), minio.PutObjectOptions{
				ContentType: "text/plain",
			})
		if err != nil {
			return fmt.Errorf("failed to save checkpoint content: %w", err)
		}

		// Устанавливаем путь к содержимому и очищаем content для метаданных
		checkpoint.ContentPath = contentPath
		checkpoint.Content = "" // Содержимое хранится отдельно
	}

	// Сохраняем метаданные чекпоинта
	metadataJSON, err := json.Marshal(checkpoint)
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	metadataPath := fmt.Sprintf("metadata/%s/%s.json", checkpoint.ChainID, checkpoint.ID)
	metadataReader := bytes.NewReader(metadataJSON)

	_, err = s.client.PutObject(ctx, s.bucketName, metadataPath, metadataReader,
		int64(len(metadataJSON)), minio.PutObjectOptions{
			ContentType: "application/json",
		})
	if err != nil {
		return fmt.Errorf("failed to save checkpoint metadata: %w", err)
	}

	return nil
}

// Get возвращает чекпоинт по ID
func (s *MinIOCheckpointStore) Get(id string) (Checkpoint, error) {
	ctx := context.Background()

	// Сначала найдем метаданные чекпоинта
	// Поскольку мы не знаем chainID, ищем по всем цепочкам
	metadataPath, err := s.findCheckpointMetadata(ctx, id)
	if err != nil {
		return Checkpoint{}, err
	}

	// Загружаем метаданные
	obj, err := s.client.GetObject(ctx, s.bucketName, metadataPath, minio.GetObjectOptions{})
	if err != nil {
		return Checkpoint{}, fmt.Errorf("failed to get checkpoint metadata: %w", err)
	}
	defer obj.Close()

	metadataBytes, err := io.ReadAll(obj)
	if err != nil {
		return Checkpoint{}, fmt.Errorf("failed to read checkpoint metadata: %w", err)
	}

	var checkpoint Checkpoint
	if err := json.Unmarshal(metadataBytes, &checkpoint); err != nil {
		return Checkpoint{}, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	// Если содержимое хранится отдельно, загружаем его
	if checkpoint.ContentPath != "" {
		contentObj, err := s.client.GetObject(ctx, s.bucketName, checkpoint.ContentPath, minio.GetObjectOptions{})
		if err != nil {
			return Checkpoint{}, fmt.Errorf("failed to get checkpoint content: %w", err)
		}
		defer contentObj.Close()

		contentBytes, err := io.ReadAll(contentObj)
		if err != nil {
			return Checkpoint{}, fmt.Errorf("failed to read checkpoint content: %w", err)
		}

		checkpoint.Content = string(contentBytes)
	}

	return checkpoint, nil
}

// List возвращает список чекпоинтов для цепочки
func (s *MinIOCheckpointStore) List(chainID string) ([]Checkpoint, error) {
	ctx := context.Background()

	prefix := fmt.Sprintf("metadata/%s/", chainID)
	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	var checkpoints []Checkpoint
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", object.Err)
		}

		// Загружаем метаданные каждого чекпоинта
		obj, err := s.client.GetObject(ctx, s.bucketName, object.Key, minio.GetObjectOptions{})
		if err != nil {
			continue // Пропускаем поврежденные файлы
		}

		metadataBytes, err := io.ReadAll(obj)
		obj.Close()
		if err != nil {
			continue
		}

		var checkpoint Checkpoint
		if err := json.Unmarshal(metadataBytes, &checkpoint); err != nil {
			continue
		}

		// Не загружаем содержимое для списка (только метаданные)
		checkpoints = append(checkpoints, checkpoint)
	}

	return checkpoints, nil
}

// Delete удаляет чекпоинт
func (s *MinIOCheckpointStore) Delete(id string) error {
	ctx := context.Background()

	// Сначала найдем чекпоинт
	checkpoint, err := s.Get(id)
	if err != nil {
		return err
	}

	// Удаляем содержимое если оно хранится отдельно
	if checkpoint.ContentPath != "" {
		err = s.client.RemoveObject(ctx, s.bucketName, checkpoint.ContentPath, minio.RemoveObjectOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete checkpoint content: %w", err)
		}
	}

	// Удаляем метаданные
	metadataPath := fmt.Sprintf("metadata/%s/%s.json", checkpoint.ChainID, checkpoint.ID)
	err = s.client.RemoveObject(ctx, s.bucketName, metadataPath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete checkpoint metadata: %w", err)
	}

	return nil
}

// findCheckpointMetadata ищет метаданные чекпоинта по ID
func (s *MinIOCheckpointStore) findCheckpointMetadata(ctx context.Context, id string) (string, error) {
	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    "metadata/",
		Recursive: true,
	})

	targetFile := fmt.Sprintf("%s.json", id)
	for object := range objectCh {
		if object.Err != nil {
			return "", fmt.Errorf("error listing objects: %w", object.Err)
		}

		if strings.HasSuffix(object.Key, targetFile) {
			return object.Key, nil
		}
	}

	return "", fmt.Errorf("checkpoint with ID '%s' not found", id)
}

// DeleteByChain удаляет все чекпоинты для указанной цепочки
func (s *MinIOCheckpointStore) DeleteByChain(chainID string) error {
	ctx := context.Background()

	// Получаем список чекпоинтов для цепочки
	checkpoints, err := s.List(chainID)
	if err != nil {
		return fmt.Errorf("failed to list checkpoints for chain %s: %w", chainID, err)
	}

	// Удаляем каждый чекпоинт
	for _, checkpoint := range checkpoints {
		if err := s.Delete(checkpoint.ID); err != nil {
			return fmt.Errorf("failed to delete checkpoint %s: %w", checkpoint.ID, err)
		}
	}

	// Удаляем все объекты в папках цепочки
	prefixes := []string{
		fmt.Sprintf("metadata/%s/", chainID),
		fmt.Sprintf("content/%s/", chainID),
	}

	for _, prefix := range prefixes {
		objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: true,
		})

		for object := range objectCh {
			if object.Err != nil {
				return fmt.Errorf("error listing objects with prefix %s: %w", prefix, object.Err)
			}

			if err := s.client.RemoveObject(ctx, s.bucketName, object.Key, minio.RemoveObjectOptions{}); err != nil {
				return fmt.Errorf("failed to remove object %s: %w", object.Key, err)
			}
		}
	}

	return nil
}

// ListAll возвращает все чекпоинты
func (s *MinIOCheckpointStore) ListAll() ([]Checkpoint, error) {
	ctx := context.Background()

	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    "metadata/",
		Recursive: true,
	})

	var checkpoints []Checkpoint
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", object.Err)
		}

		// Загружаем метаданные каждого чекпоинта
		obj, err := s.client.GetObject(ctx, s.bucketName, object.Key, minio.GetObjectOptions{})
		if err != nil {
			continue // Пропускаем поврежденные файлы
		}

		metadataBytes, err := io.ReadAll(obj)
		obj.Close()
		if err != nil {
			continue
		}

		var checkpoint Checkpoint
		if err := json.Unmarshal(metadataBytes, &checkpoint); err != nil {
			continue
		}

		checkpoints = append(checkpoints, checkpoint)
	}

	return checkpoints, nil
}