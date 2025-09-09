package checkpoint

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/grik-ai/ricochet-task/pkg/checkpoint"
	"github.com/grik-ai/ricochet-task/internal/config"
	"github.com/spf13/cobra"
)

// Команда checkpoint
var CheckpointCmd = &cobra.Command{
	Use:   "checkpoint",
	Short: "Управление чекпоинтами цепочек",
	Long:  `Команды для просмотра, сохранения и удаления чекпоинтов цепочек моделей.`,
}

// Инициализация команд
func init() {
	CheckpointCmd.AddCommand(listCmd)
	CheckpointCmd.AddCommand(getCmd)
	CheckpointCmd.AddCommand(saveCmd)
	CheckpointCmd.AddCommand(deleteCmd)
}

// Команда checkpoint list
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Список чекпоинтов",
	Long:  `Отображение списка всех чекпоинтов для указанной цепочки.`,
	Run: func(cmd *cobra.Command, args []string) {
		chainID, _ := cmd.Flags().GetString("chain")

		if chainID == "" {
			fmt.Println("Ошибка: ID цепочки не указан")
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

		// Создание хранилища чекпоинтов
		checkpointStore, err := checkpoint.NewFileCheckpointStore(cfg.ConfigDir)
		if err != nil {
			fmt.Printf("Ошибка при создании хранилища чекпоинтов: %v\n", err)
			os.Exit(1)
		}

		// Получение списка чекпоинтов
		checkpoints, err := checkpointStore.List(chainID)
		if err != nil {
			fmt.Printf("Ошибка при получении списка чекпоинтов: %v\n", err)
			os.Exit(1)
		}

		if len(checkpoints) == 0 {
			fmt.Println("Чекпоинты не найдены.")
			return
		}

		// Вывод списка чекпоинтов
		fmt.Println("Список чекпоинтов:")
		fmt.Println("----------------------------------------------------")
		for _, c := range checkpoints {
			fmt.Printf("ID: %s\n", c.ID)
			fmt.Printf("Тип: %s\n", c.Type)
			if c.ModelID != "" {
				fmt.Printf("Модель: %s\n", c.ModelID)
			}
			fmt.Printf("Создан: %s\n", c.CreatedAt.Format(time.RFC3339))
			if c.ContentPath != "" {
				fmt.Printf("Путь к содержимому: %s\n", c.ContentPath)
			} else {
				contentPreview := c.Content
				if len(contentPreview) > 50 {
					contentPreview = contentPreview[:50] + "..."
				}
				fmt.Printf("Содержимое: %s\n", contentPreview)
			}
			fmt.Println("----------------------------------------------------")
		}
	},
}

// Команда checkpoint get
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Получить содержимое чекпоинта",
	Long:  `Получение полного содержимого чекпоинта по ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")
		outputFile, _ := cmd.Flags().GetString("output")

		if id == "" {
			fmt.Println("Ошибка: ID чекпоинта не указан")
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

		// Создание хранилища чекпоинтов
		checkpointStore, err := checkpoint.NewFileCheckpointStore(cfg.ConfigDir)
		if err != nil {
			fmt.Printf("Ошибка при создании хранилища чекпоинтов: %v\n", err)
			os.Exit(1)
		}

		// Получение чекпоинта
		cp, err := checkpointStore.Get(id)
		if err != nil {
			fmt.Printf("Ошибка при получении чекпоинта: %v\n", err)
			os.Exit(1)
		}

		// Вывод содержимого
		if outputFile != "" {
			// Сохранение в файл
			if err := os.WriteFile(outputFile, []byte(cp.Content), 0644); err != nil {
				fmt.Printf("Ошибка при сохранении содержимого в файл: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Содержимое чекпоинта сохранено в файл: %s\n", outputFile)
		} else {
			// Вывод в консоль
			fmt.Println("Содержимое чекпоинта:")
			fmt.Println("----------------------------------------------------")
			fmt.Println(cp.Content)
			fmt.Println("----------------------------------------------------")
		}
	},
}

// Команда checkpoint save
var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Сохранить чекпоинт",
	Long:  `Сохранение нового чекпоинта для цепочки моделей.`,
	Run: func(cmd *cobra.Command, args []string) {
		chainID, _ := cmd.Flags().GetString("chain")
		modelID, _ := cmd.Flags().GetString("model")
		cpType, _ := cmd.Flags().GetString("type")
		content, _ := cmd.Flags().GetString("content")
		inputFile, _ := cmd.Flags().GetString("input-file")

		if chainID == "" {
			fmt.Println("Ошибка: ID цепочки не указан")
			os.Exit(1)
		}

		if cpType == "" {
			fmt.Println("Ошибка: тип чекпоинта не указан")
			os.Exit(1)
		}

		// Проверка типа чекпоинта
		var checkpointType checkpoint.CheckpointType
		switch cpType {
		case "input":
			checkpointType = checkpoint.CheckpointTypeInput
		case "output":
			checkpointType = checkpoint.CheckpointTypeOutput
		case "segment":
			checkpointType = checkpoint.CheckpointTypeSegment
		case "complete":
			checkpointType = checkpoint.CheckpointTypeComplete
		default:
			fmt.Printf("Ошибка: неизвестный тип чекпоинта '%s'. Допустимые значения: input, output, segment, complete\n", cpType)
			os.Exit(1)
		}

		// Если указан файл, читаем из него
		if inputFile != "" {
			data, err := os.ReadFile(inputFile)
			if err != nil {
				fmt.Printf("Ошибка при чтении файла: %v\n", err)
				os.Exit(1)
			}
			content = string(data)
		}

		if content == "" {
			fmt.Println("Ошибка: содержимое чекпоинта не указано")
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

		// Создание хранилища чекпоинтов
		checkpointStore, err := checkpoint.NewFileCheckpointStore(cfg.ConfigDir)
		if err != nil {
			fmt.Printf("Ошибка при создании хранилища чекпоинтов: %v\n", err)
			os.Exit(1)
		}

		// Создание нового чекпоинта
		cp := checkpoint.Checkpoint{
			ID:        uuid.New().String(),
			ChainID:   chainID,
			ModelID:   modelID,
			Type:      checkpointType,
			Content:   content,
			CreatedAt: time.Now(),
			MetaData:  make(map[string]interface{}),
		}

		// Сохранение чекпоинта
		if err := checkpointStore.Save(cp); err != nil {
			fmt.Printf("Ошибка при сохранении чекпоинта: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Чекпоинт успешно сохранен. ID:", cp.ID)
	},
}

// Команда checkpoint delete
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Удалить чекпоинт",
	Long:  `Удаление чекпоинта по ID или всех чекпоинтов для цепочки.`,
	Run: func(cmd *cobra.Command, args []string) {
		id, _ := cmd.Flags().GetString("id")
		chainID, _ := cmd.Flags().GetString("chain")
		all, _ := cmd.Flags().GetBool("all")

		if id == "" && chainID == "" {
			fmt.Println("Ошибка: необходимо указать ID чекпоинта или ID цепочки")
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

		// Создание хранилища чекпоинтов
		checkpointStore, err := checkpoint.NewFileCheckpointStore(cfg.ConfigDir)
		if err != nil {
			fmt.Printf("Ошибка при создании хранилища чекпоинтов: %v\n", err)
			os.Exit(1)
		}

		// Удаление по ID чекпоинта
		if id != "" {
			if err := checkpointStore.Delete(id); err != nil {
				fmt.Printf("Ошибка при удалении чекпоинта: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Чекпоинт успешно удален.")
			return
		}

		// Удаление всех чекпоинтов для цепочки
		if chainID != "" && all {
			if err := checkpointStore.DeleteByChain(chainID); err != nil {
				fmt.Printf("Ошибка при удалении чекпоинтов: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Все чекпоинты для цепочки успешно удалены.")
			return
		}

		fmt.Println("Ошибка: для удаления всех чекпоинтов цепочки необходимо указать флаг --all")
	},
}

// Инициализация флагов для команд
func init() {
	// Флаги для команды checkpoint list
	listCmd.Flags().String("chain", "", "ID цепочки")
	listCmd.MarkFlagRequired("chain")

	// Флаги для команды checkpoint get
	getCmd.Flags().String("id", "", "ID чекпоинта")
	getCmd.Flags().String("output", "", "Путь для сохранения содержимого в файл")
	getCmd.MarkFlagRequired("id")

	// Флаги для команды checkpoint save
	saveCmd.Flags().String("chain", "", "ID цепочки")
	saveCmd.Flags().String("model", "", "ID модели (необязательно)")
	saveCmd.Flags().String("type", "", "Тип чекпоинта (input, output, segment, complete)")
	saveCmd.Flags().String("content", "", "Содержимое чекпоинта")
	saveCmd.Flags().String("input-file", "", "Путь к файлу с содержимым")
	saveCmd.MarkFlagRequired("chain")
	saveCmd.MarkFlagRequired("type")

	// Флаги для команды checkpoint delete
	deleteCmd.Flags().String("id", "", "ID чекпоинта")
	deleteCmd.Flags().String("chain", "", "ID цепочки")
	deleteCmd.Flags().Bool("all", false, "Удалить все чекпоинты цепочки")
}
