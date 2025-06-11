package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config представляет конфигурацию приложения
type Config struct {
	APIGateway string `json:"api_gateway"`
	ConfigDir  string `json:"config_dir"`
	LogLevel   string `json:"log_level"`
	APIKey     string `json:"api_key,omitempty"`
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	configDir := filepath.Join(homeDir, ".ricochet")

	return Config{
		APIGateway: "http://localhost:8080",
		ConfigDir:  configDir,
		LogLevel:   "info",
	}
}

// LoadConfig загружает конфигурацию из файла
func LoadConfig(path string) (Config, error) {
	config := DefaultConfig()

	// Если файл не существует, используем конфигурацию по умолчанию
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return config, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("не удалось прочитать файл конфигурации: %w", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("не удалось распарсить файл конфигурации: %w", err)
	}

	return config, nil
}

// SaveConfig сохраняет конфигурацию в файл
func SaveConfig(path string, config Config) error {
	// Создаем директорию, если она не существует
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию для конфигурации: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("не удалось сериализовать конфигурацию: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("не удалось записать файл конфигурации: %w", err)
	}

	return nil
}

// GetConfigPath возвращает путь к файлу конфигурации
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("не удалось определить домашнюю директорию: %w", err)
	}

	return filepath.Join(homeDir, ".ricochet", "config.json"), nil
}
