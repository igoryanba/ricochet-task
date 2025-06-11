package chain

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/grik-ai/ricochet-task/internal/config"
	"github.com/grik-ai/ricochet-task/pkg/chain"
	"github.com/spf13/cobra"
)

// Команда chain
var ChainCmd = &cobra.Command{
	Use:   "chain",
	Short: "Управление цепочками моделей",
	Long:  `Команды для создания, редактирования и запуска цепочек моделей.`,
}

// Инициализация команд
func init() {
	ChainCmd.AddCommand(createCmd)
	ChainCmd.AddCommand(listCmd)
	ChainCmd.AddCommand(addModelCmd)
	ChainCmd.AddCommand(runCmd)
	ChainCmd.AddCommand(statusCmd)
	ChainCmd.AddCommand(deleteCmd)
}

// Команда chain create
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Создать новую цепочку моделей",
	Long:  `Создание новой цепочки моделей с указанным именем и описанием.`,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")

		if name == "" {
			fmt.Println("Ошибка: имя цепочки не указано")
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

		// Создание хранилища цепочек
		chainStore, err := chain.NewFileChainStore(cfg.ConfigDir)
		if err != nil {
			fmt.Printf("Ошибка при создании хранилища цепочек: %v\n", err)
			os.Exit(1)
		}

		// Создание новой цепочки
		newChain := chain.Chain{
			ID:          uuid.New().String(),
			Name:        name,
			Description: description,
			Models:      []chain.Model{},
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Добавление цепочки
		err = chainStore.Save(newChain)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка при сохранении цепочки: %v\n", err)
			return
		}

		fmt.Printf("Цепочка моделей '%s' успешно создана. ID: %s\n", name, newChain.ID)
	},
}

// Команда chain list
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Список цепочек моделей",
	Long:  `Отображение списка всех созданных цепочек моделей.`,
	Run: func(cmd *cobra.Command, args []string) {
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

		// Создание хранилища цепочек
		chainStore, err := chain.NewFileChainStore(cfg.ConfigDir)
		if err != nil {
			fmt.Printf("Ошибка при создании хранилища цепочек: %v\n", err)
			os.Exit(1)
		}

		// Получение списка цепочек
		chains, err := chainStore.List()
		if err != nil {
			fmt.Printf("Ошибка при получении списка цепочек: %v\n", err)
			os.Exit(1)
		}

		if len(chains) == 0 {
			fmt.Println("Цепочки моделей не найдены.")
			return
		}

		// Вывод списка цепочек
		fmt.Println("Список цепочек моделей:")
		fmt.Println("----------------------------------------------------")
		for _, c := range chains {
			fmt.Printf("ID: %s\n", c.ID)
			fmt.Printf("Имя: %s\n", c.Name)
			if c.Description != "" {
				fmt.Printf("Описание: %s\n", c.Description)
			}
			fmt.Printf("Создана: %s\n", c.CreatedAt.Format(time.RFC3339))
			fmt.Printf("Количество моделей: %d\n", len(c.Models))
			fmt.Println("----------------------------------------------------")
		}
	},
}

// Команда chain add-model
var addModelCmd = &cobra.Command{
	Use:   "add-model",
	Short: "Добавить модель в цепочку",
	Long:  `Добавление новой модели в существующую цепочку с указанной ролью и параметрами.`,
	Run: func(cmd *cobra.Command, args []string) {
		chainID, _ := cmd.Flags().GetString("chain")
		name, _ := cmd.Flags().GetString("name")
		modelType, _ := cmd.Flags().GetString("type")
		role, _ := cmd.Flags().GetString("role")
		prompt, _ := cmd.Flags().GetString("prompt")
		temperature, _ := cmd.Flags().GetFloat64("temperature")
		maxTokens, _ := cmd.Flags().GetInt("max-tokens")

		if chainID == "" {
			fmt.Println("Ошибка: ID цепочки не указан")
			os.Exit(1)
		}

		if name == "" {
			fmt.Println("Ошибка: название модели не указано")
			os.Exit(1)
		}

		if modelType == "" {
			fmt.Println("Ошибка: тип модели не указан")
			os.Exit(1)
		}

		if role == "" {
			fmt.Println("Ошибка: роль модели не указана")
			os.Exit(1)
		}

		// Проверка типа модели
		var modelTypeEnum chain.ModelType
		switch modelType {
		case "openai":
			modelTypeEnum = chain.ModelTypeOpenAI
		case "claude":
			modelTypeEnum = chain.ModelTypeClaude
		case "deepseek":
			modelTypeEnum = chain.ModelTypeDeepSeek
		case "grok":
			modelTypeEnum = chain.ModelTypeGrok
		default:
			fmt.Printf("Ошибка: неизвестный тип модели '%s'. Допустимые значения: openai, claude, deepseek, grok\n", modelType)
			os.Exit(1)
		}

		// Проверка роли модели
		var roleEnum chain.ModelRole
		switch role {
		case "analyzer":
			roleEnum = chain.ModelRoleAnalyzer
		case "summarizer":
			roleEnum = chain.ModelRoleSummarizer
		case "integrator":
			roleEnum = chain.ModelRoleIntegrator
		case "extractor":
			roleEnum = chain.ModelRoleExtractor
		case "organizer":
			roleEnum = chain.ModelRoleOrganizer
		case "evaluator":
			roleEnum = chain.ModelRoleEvaluator
		default:
			fmt.Printf("Ошибка: неизвестная роль модели '%s'. Допустимые значения: analyzer, summarizer, integrator, extractor, organizer, evaluator\n", role)
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

		// Создание хранилища цепочек
		chainStore, err := chain.NewFileChainStore(cfg.ConfigDir)
		if err != nil {
			fmt.Printf("Ошибка при создании хранилища цепочек: %v\n", err)
			os.Exit(1)
		}

		// Получение цепочки
		c, err := chainStore.Get(chainID)
		if err != nil {
			fmt.Printf("Ошибка при получении цепочки: %v\n", err)
			os.Exit(1)
		}

		// Создание новой модели
		model := chain.Model{
			ID:        uuid.New().String(),
			Name:      chain.ModelName(name),
			Type:      modelTypeEnum,
			Role:      roleEnum,
			MaxTokens: maxTokens,
			Prompt:    prompt,
			Order:     len(c.Models),
			Parameters: chain.Parameters{
				Temperature:      temperature,
				TopP:             0.9,
				FrequencyPenalty: 0.0,
				PresencePenalty:  0.0,
				Stop:             []string{},
			},
			Temperature: temperature,
		}

		// Добавление модели в цепочку
		c.Models = append(c.Models, model)
		c.UpdatedAt = time.Now()

		// Обновление цепочки
		err = chainStore.Save(c)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка при обновлении цепочки: %v\n", err)
			return
		}

		fmt.Printf("Модель '%s' успешно добавлена в цепочку '%s'.\n", name, c.Name)
	},
}

// Команда chain run
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Запустить цепочку моделей",
	Long:  `Запуск цепочки моделей с указанным входным текстом или файлом.`,
	Run: func(cmd *cobra.Command, args []string) {
		chainID, _ := cmd.Flags().GetString("chain")
		input, _ := cmd.Flags().GetString("input")
		inputFile, _ := cmd.Flags().GetString("input-file")

		if chainID == "" {
			fmt.Println("Ошибка: ID цепочки не указан")
			os.Exit(1)
		}

		if input == "" && inputFile == "" {
			fmt.Println("Ошибка: необходимо указать входной текст через --input или путь к файлу через --input-file")
			os.Exit(1)
		}

		// Если указан файл, читаем из него
		if inputFile != "" {
			data, err := os.ReadFile(inputFile)
			if err != nil {
				fmt.Printf("Ошибка при чтении файла: %v\n", err)
				os.Exit(1)
			}
			input = string(data)
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

		// Создание хранилища цепочек
		chainStore, err := chain.NewFileChainStore(cfg.ConfigDir)
		if err != nil {
			fmt.Printf("Ошибка при создании хранилища цепочек: %v\n", err)
			os.Exit(1)
		}

		// Получение цепочки
		c, err := chainStore.Get(chainID)
		if err != nil {
			fmt.Printf("Ошибка при получении цепочки: %v\n", err)
			os.Exit(1)
		}

		if len(c.Models) == 0 {
			fmt.Printf("Ошибка: цепочка '%s' не содержит моделей\n", c.Name)
			os.Exit(1)
		}

		// TODO: Реализовать запуск цепочки с использованием Ricochet Service
		// В данной реализации просто выводим информацию о запуске
		fmt.Printf("Запущена цепочка '%s' с %d моделями.\n", c.Name, len(c.Models))
		fmt.Println("ID запуска: " + uuid.New().String())
		fmt.Println("Статус: обработка")
		fmt.Println("Для проверки статуса используйте команду: ricochet chain status --chain " + chainID)
	},
}

// Команда chain status
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Проверить статус выполнения цепочки",
	Long:  `Проверка статуса выполнения цепочки моделей и просмотр промежуточных результатов.`,
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

		// Создание хранилища цепочек
		chainStore, err := chain.NewFileChainStore(cfg.ConfigDir)
		if err != nil {
			fmt.Printf("Ошибка при создании хранилища цепочек: %v\n", err)
			os.Exit(1)
		}

		// Получение цепочки
		c, err := chainStore.Get(chainID)
		if err != nil {
			fmt.Printf("Ошибка при получении цепочки: %v\n", err)
			os.Exit(1)
		}

		// TODO: Получение статуса выполнения цепочки из Ricochet Service
		// В данной реализации просто выводим информацию о цепочке
		fmt.Printf("Цепочка: %s (%s)\n", c.Name, c.ID)
		fmt.Println("Статус: не выполняется")
		fmt.Println("Модели в цепочке:")

		for i, model := range c.Models {
			fmt.Printf("%d. %s (%s, роль: %s)\n", i+1, model.Name, model.Type, model.Role)
		}
	},
}

// Команда chain delete
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Удалить цепочку моделей",
	Long:  `Удаление цепочки моделей по ID.`,
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

		// Создание хранилища цепочек
		chainStore, err := chain.NewFileChainStore(cfg.ConfigDir)
		if err != nil {
			fmt.Printf("Ошибка при создании хранилища цепочек: %v\n", err)
			os.Exit(1)
		}

		// Получение цепочки для подтверждения
		c, err := chainStore.Get(chainID)
		if err != nil {
			fmt.Printf("Ошибка при получении цепочки: %v\n", err)
			os.Exit(1)
		}

		// Удаление цепочки
		if err := chainStore.Delete(chainID); err != nil {
			fmt.Printf("Ошибка при удалении цепочки: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Цепочка моделей '%s' успешно удалена.\n", c.Name)
	},
}

// Инициализация флагов для команд
func init() {
	// Флаги для команды chain create
	createCmd.Flags().String("name", "", "Имя цепочки")
	createCmd.Flags().String("description", "", "Описание цепочки")
	createCmd.MarkFlagRequired("name")

	// Флаги для команды chain add-model
	addModelCmd.Flags().String("chain", "", "ID цепочки")
	addModelCmd.Flags().String("name", "", "Название модели")
	addModelCmd.Flags().String("type", "", "Тип модели (openai, claude, deepseek, grok)")
	addModelCmd.Flags().String("role", "", "Роль модели (analyzer, summarizer, integrator, extractor, organizer, evaluator)")
	addModelCmd.Flags().String("prompt", "", "Системный промпт для модели")
	addModelCmd.Flags().Float64("temperature", 0.7, "Температура (0.0-1.0)")
	addModelCmd.Flags().Int("max-tokens", 1000, "Максимальное количество токенов")
	addModelCmd.MarkFlagRequired("chain")
	addModelCmd.MarkFlagRequired("name")
	addModelCmd.MarkFlagRequired("type")
	addModelCmd.MarkFlagRequired("role")

	// Флаги для команды chain run
	runCmd.Flags().String("chain", "", "ID цепочки")
	runCmd.Flags().String("input", "", "Входной текст")
	runCmd.Flags().String("input-file", "", "Путь к входному файлу")
	runCmd.MarkFlagRequired("chain")

	// Флаги для команды chain status
	statusCmd.Flags().String("chain", "", "ID цепочки")
	statusCmd.MarkFlagRequired("chain")

	// Флаги для команды chain delete
	deleteCmd.Flags().String("chain", "", "ID цепочки")
	deleteCmd.MarkFlagRequired("chain")
}
