package key

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/grik-ai/ricochet-task/pkg/key"
	"github.com/grik-ai/ricochet-task/internal/config"
	"github.com/spf13/cobra"
)

// Команда key
var KeyCmd = &cobra.Command{
	Use:   "key",
	Short: "Управление API-ключами",
	Long:  `Команды для добавления, удаления и просмотра API-ключей для различных провайдеров.`,
}

// Инициализация команд
func init() {
	KeyCmd.AddCommand(addCmd)
	KeyCmd.AddCommand(listCmd)
	KeyCmd.AddCommand(deleteCmd)
	KeyCmd.AddCommand(shareCmd)
}

// Команда key add
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Добавить API-ключ",
	Long:  `Добавление нового API-ключа для доступа к моделям.`,
	Run: func(cmd *cobra.Command, args []string) {
		provider, _ := cmd.Flags().GetString("provider")
		value, _ := cmd.Flags().GetString("key")
		shared, _ := cmd.Flags().GetBool("shared")

		if provider == "" {
			fmt.Println("Ошибка: провайдер не указан")
			os.Exit(1)
		}

		if value == "" {
			fmt.Println("Ошибка: ключ не указан")
			os.Exit(1)
		}

		// Нормализация провайдера
		provider = strings.ToLower(provider)
		allowedProviders := []string{"openai", "claude", "deepseek", "grok"}
		validProvider := false
		for _, p := range allowedProviders {
			if p == provider {
				validProvider = true
				break
			}
		}

		if !validProvider {
			fmt.Printf("Ошибка: неизвестный провайдер '%s'. Допустимые значения: %s\n",
				provider, strings.Join(allowedProviders, ", "))
			os.Exit(1)
		}

		// Загрузка конфигурации
		configPath, err := config.GetConfigPath()
		if err != nil {
			fmt.Printf("Ошибка при получении пути конфигурации: %v\n", err)
			os.Exit(1)
		}

		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("Ошибка при загрузке конфигурации: %v\n", err)
			os.Exit(1)
		}

		// Создание хранилища ключей
		keyStore, err := key.NewFileKeyStore(cfg.ConfigDir)
		if err != nil {
			fmt.Printf("Ошибка при создании хранилища ключей: %v\n", err)
			os.Exit(1)
		}

		// Создание нового ключа
		newKey := key.Key{
			ID:         uuid.New().String(),
			Provider:   provider,
			Value:      value,
			Name:       fmt.Sprintf("%s-key", provider),
			CreatedAt:  time.Now(),
			LastUsedAt: time.Now(),
			Shared:     shared,
			Usage: key.KeyUsage{
				TotalRequests: 0,
				TotalTokens:   0,
				LastRequest:   time.Now(),
				MonthlyTokens: 0,
				DailyTokens:   0,
			},
			Metadata: key.Metadata{
				Quota:      0,
				Expiration: time.Now().AddDate(1, 0, 0), // По умолчанию ключ действителен 1 год
				Custom:     make(map[string]string),
			},
		}

		// Добавление ключа
		if err := keyStore.Add(newKey); err != nil {
			fmt.Printf("Ошибка при добавлении ключа: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("API-ключ для провайдера '%s' успешно добавлен.\n", provider)
		if shared {
			fmt.Println("Ключ доступен для общего использования.")
		}
	},
}

// Команда key list
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Список API-ключей",
	Long:  `Отображение списка всех добавленных API-ключей.`,
	Run: func(cmd *cobra.Command, args []string) {
		provider, _ := cmd.Flags().GetString("provider")

		// Загрузка конфигурации
		configPath, err := config.GetConfigPath()
		if err != nil {
			fmt.Printf("Ошибка при получении пути конфигурации: %v\n", err)
			os.Exit(1)
		}

		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("Ошибка при загрузке конфигурации: %v\n", err)
			os.Exit(1)
		}

		// Создание хранилища ключей
		keyStore, err := key.NewFileKeyStore(cfg.ConfigDir)
		if err != nil {
			fmt.Printf("Ошибка при создании хранилища ключей: %v\n", err)
			os.Exit(1)
		}

		// Получение списка ключей
		keys, err := keyStore.List()
		if err != nil {
			fmt.Printf("Ошибка при получении списка ключей: %v\n", err)
			os.Exit(1)
		}

		if len(keys) == 0 {
			fmt.Println("API-ключи не найдены.")
			return
		}

		// Фильтрация по провайдеру, если указан
		var filteredKeys []key.Key
		if provider != "" {
			for _, k := range keys {
				if k.Provider == provider {
					filteredKeys = append(filteredKeys, k)
				}
			}
		} else {
			filteredKeys = keys
		}

		if len(filteredKeys) == 0 {
			fmt.Printf("API-ключи для провайдера '%s' не найдены.\n", provider)
			return
		}

		// Вывод списка ключей
		fmt.Println("Список API-ключей:")
		fmt.Println("----------------------------------------------------")
		for _, k := range filteredKeys {
			// Маскируем ключ для безопасности
			maskedKey := maskKey(k.Value)

			fmt.Printf("ID: %s\n", k.ID)
			fmt.Printf("Провайдер: %s\n", k.Provider)
			fmt.Printf("Ключ: %s\n", maskedKey)
			fmt.Printf("Создан: %s\n", k.CreatedAt.Format(time.RFC3339))
			fmt.Printf("Общий доступ: %v\n", k.Shared)
			fmt.Printf("Использовано токенов: %d\n", k.Usage.TotalTokens)
			if k.Metadata.Quota > 0 {
				fmt.Printf("Лимит использования: %d\n", k.Metadata.Quota)
			} else {
				fmt.Println("Лимит использования: не ограничен")
			}
			fmt.Println("----------------------------------------------------")
		}
	},
}

// Команда key delete
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Удалить API-ключ",
	Long:  `Удаление API-ключа по ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")

		if id == "" {
			fmt.Println("Ошибка: ID ключа не указан")
			os.Exit(1)
		}

		// Загрузка конфигурации
		configPath, err := config.GetConfigPath()
		if err != nil {
			fmt.Printf("Ошибка при получении пути конфигурации: %v\n", err)
			os.Exit(1)
		}

		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("Ошибка при загрузке конфигурации: %v\n", err)
			os.Exit(1)
		}

		// Создание хранилища ключей
		keyStore, err := key.NewFileKeyStore(cfg.ConfigDir)
		if err != nil {
			fmt.Printf("Ошибка при создании хранилища ключей: %v\n", err)
			os.Exit(1)
		}

		// Удаление ключа
		if err := keyStore.Delete(id); err != nil {
			fmt.Printf("Ошибка при удалении ключа: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("API-ключ успешно удален.")
	},
}

// Команда key share
var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Настроить общий доступ к API-ключу",
	Long:  `Включение или отключение общего доступа к API-ключу.`,
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")
		shared, _ := cmd.Flags().GetBool("shared")

		if id == "" {
			fmt.Println("Ошибка: ID ключа не указан")
			os.Exit(1)
		}

		// Загрузка конфигурации
		configPath, err := config.GetConfigPath()
		if err != nil {
			fmt.Printf("Ошибка при получении пути конфигурации: %v\n", err)
			os.Exit(1)
		}

		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("Ошибка при загрузке конфигурации: %v\n", err)
			os.Exit(1)
		}

		// Создание хранилища ключей
		keyStore, err := key.NewFileKeyStore(cfg.ConfigDir)
		if err != nil {
			fmt.Printf("Ошибка при создании хранилища ключей: %v\n", err)
			os.Exit(1)
		}

		// Получение ключа
		apiKey, err := keyStore.Get(id)
		if err != nil {
			fmt.Printf("Ошибка при получении ключа: %v\n", err)
			os.Exit(1)
		}

		// Обновление статуса общего доступа
		apiKey.Shared = shared
		apiKey.LastUsedAt = time.Now()

		// Сохранение изменений
		if err := keyStore.Update(apiKey); err != nil {
			fmt.Printf("Ошибка при обновлении ключа: %v\n", err)
			os.Exit(1)
		}

		if shared {
			fmt.Println("Общий доступ к API-ключу включен.")
		} else {
			fmt.Println("Общий доступ к API-ключу отключен.")
		}
	},
}

// Инициализация флагов для команд
func init() {
	// Флаги для команды key add
	addCmd.Flags().String("provider", "", "Провайдер API (openai, claude, deepseek, grok)")
	addCmd.Flags().String("key", "", "Значение API-ключа")
	addCmd.Flags().Bool("shared", false, "Включить общий доступ к ключу")
	addCmd.Flags().Int64("limit", 0, "Лимит использования (токены)")
	addCmd.MarkFlagRequired("provider")
	addCmd.MarkFlagRequired("key")

	// Флаги для команды key list
	listCmd.Flags().String("provider", "", "Фильтр по провайдеру")

	// Флаги для команды key delete
	deleteCmd.Flags().String("id", "", "ID ключа для удаления")
	deleteCmd.MarkFlagRequired("id")

	// Флаги для команды key share
	shareCmd.Flags().String("id", "", "ID ключа")
	shareCmd.Flags().Bool("shared", true, "Включить (true) или отключить (false) общий доступ")
	shareCmd.MarkFlagRequired("id")
}

// Вспомогательные функции

// maskKey маскирует API-ключ для безопасного отображения
func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}

	return key[:4] + "..." + key[len(key)-4:]
}
