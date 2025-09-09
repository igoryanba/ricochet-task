package context

import (
	"fmt"
	"os"
	"strings"

	"github.com/grik-ai/ricochet-task/pkg/ai"
	"github.com/grik-ai/ricochet-task/pkg/context"
	"github.com/grik-ai/ricochet-task/pkg/providers"
	"github.com/spf13/cobra"
)

// ContextLogger реализует интерфейс context.Logger
type ContextLogger struct{}

func (l *ContextLogger) Info(msg string, args ...interface{}) {
	fmt.Printf("[INFO] %s %v\n", msg, args)
}

func (l *ContextLogger) Error(msg string, err error, args ...interface{}) {
	fmt.Printf("[ERROR] %s: %v %v\n", msg, err, args)
}

func (l *ContextLogger) Warn(msg string, args ...interface{}) {
	fmt.Printf("[WARN] %s %v\n", msg, args)
}

func (l *ContextLogger) Debug(msg string, args ...interface{}) {
	fmt.Printf("[DEBUG] %s %v\n", msg, args)
}

var ContextCmd = &cobra.Command{
	Use:   "context",
	Short: "Управление контекстами проектов",
	Long: `Команды для управления контекстами работы с проектами.
Позволяет создавать, переключаться и управлять контекстами для разных проектов и досок.`,
}

var (
	contextName        string
	contextDescription string
	boardID            string
	projectID          string
	providerName       string
	defaultAssignee    string
	defaultPriority    string
	projectType        string
	complexity         string
	timeline           int
	teamSize           int
	aiEnabled          bool
	autoAssignment     bool
)

func init() {
	// Глобальные флаги
	ContextCmd.PersistentFlags().StringVar(&contextName, "name", "", "Название контекста")
	ContextCmd.PersistentFlags().StringVar(&contextDescription, "description", "", "Описание контекста")
	ContextCmd.PersistentFlags().StringVar(&boardID, "board-id", "", "ID доски")
	ContextCmd.PersistentFlags().StringVar(&projectID, "project-id", "", "ID проекта")
	ContextCmd.PersistentFlags().StringVar(&providerName, "provider", "", "Имя провайдера")
	ContextCmd.PersistentFlags().StringVar(&defaultAssignee, "assignee", "", "Исполнитель по умолчанию")
	ContextCmd.PersistentFlags().StringVar(&defaultPriority, "priority", "medium", "Приоритет по умолчанию")
	ContextCmd.PersistentFlags().StringVar(&projectType, "type", "", "Тип проекта")
	ContextCmd.PersistentFlags().StringVar(&complexity, "complexity", "medium", "Сложность проекта")
	ContextCmd.PersistentFlags().IntVar(&timeline, "timeline", 14, "Временные рамки в днях")
	ContextCmd.PersistentFlags().IntVar(&teamSize, "team-size", 1, "Размер команды")
	ContextCmd.PersistentFlags().BoolVar(&aiEnabled, "ai", true, "Включить AI")
	ContextCmd.PersistentFlags().BoolVar(&autoAssignment, "auto-assign", false, "Автоназначение")

	// Подкоманды
	ContextCmd.AddCommand(listCmd)
	ContextCmd.AddCommand(createCmd)
	ContextCmd.AddCommand(switchCmd)
	ContextCmd.AddCommand(currentCmd)
	ContextCmd.AddCommand(updateCmd)
	ContextCmd.AddCommand(deleteCmd)
	ContextCmd.AddCommand(analyzeCmd)
	ContextCmd.AddCommand(boardsCmd)
	ContextCmd.AddCommand(multiCmd)
}

// listCmd - список контекстов
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Показать список всех контекстов",
	Run: func(cmd *cobra.Command, args []string) {
		log := &ContextLogger{}
		cm := context.NewContextManager("", log)

		contexts := cm.ListContexts()
		if len(contexts) == 0 {
			fmt.Println("📭 Нет созданных контекстов")
			fmt.Println("Создайте новый контекст: ricochet context create")
			return
		}

		fmt.Println("📋 Контексты проектов:")
		fmt.Println(strings.Repeat("=", 60))

		for _, ctx := range contexts {
			status := ""
			if ctx.IsActive {
				status = "🟢 АКТИВЕН"
			} else {
				status = "⚪ неактивен"
			}

			fmt.Printf("%s %s (%s)\n", status, ctx.Name, ctx.ID)
			fmt.Printf("   📝 %s\n", ctx.Description)
			fmt.Printf("   📊 Доска: %s | Проект: %s\n", ctx.BoardID, ctx.ProjectID)
			fmt.Printf("   🔧 Провайдер: %s | Тип: %s\n", ctx.ProviderName, ctx.ProjectType)
			
			if ctx.Stats != nil {
				fmt.Printf("   📈 Задач: %d создано, %d выполнено\n", 
					ctx.Stats.TasksCreated, ctx.Stats.TasksCompleted)
			}
			
			fmt.Printf("   🕐 Создан: %s | Обновлен: %s\n", 
				ctx.CreatedAt.Format("2006-01-02 15:04"), 
				ctx.UpdatedAt.Format("2006-01-02 15:04"))
			fmt.Println()
		}
	},
}

// createCmd - создание контекста
var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Создать новый контекст",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		log := &ContextLogger{}
		cm := context.NewContextManager("", log)

		// Определяем имя контекста
		name := contextName
		if len(args) > 0 {
			name = args[0]
		}
		if name == "" {
			fmt.Print("📝 Введите название контекста: ")
			fmt.Scanln(&name)
		}

		// Определяем описание
		description := contextDescription
		if description == "" {
			fmt.Print("📋 Введите описание (опционально): ")
			fmt.Scanln(&description)
		}

		// Интерактивный ввод параметров если не заданы
		if boardID == "" {
			fmt.Print("🎯 Введите ID доски: ")
			fmt.Scanln(&boardID)
		}

		if projectID == "" {
			fmt.Print("📁 Введите ID проекта: ")
			fmt.Scanln(&projectID)
		}

		if providerName == "" {
			fmt.Print("🔧 Введите имя провайдера: ")
			fmt.Scanln(&providerName)
		}

		// Создаем конфигурацию
		config := &context.ContextConfig{
			BoardID:         boardID,
			ProjectID:       projectID,
			ProviderName:    providerName,
			DefaultAssignee: defaultAssignee,
			DefaultLabels:   []string{"ricochet-managed"},
			DefaultPriority: providers.TaskPriority(defaultPriority),
			ProjectType:     projectType,
			Complexity:      complexity,
			Timeline:        timeline,
			TeamSize:        teamSize,
			WorkflowType:    "agile",
			AutoAssignment:  autoAssignment,
			AutoProgress:    true,
			AIEnabled:       aiEnabled,
			CustomFields:    make(map[string]interface{}),
		}

		// Создаем контекст
		ctx, err := cm.CreateContext(name, description, config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Ошибка создания контекста: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Контекст создан!\n")
		fmt.Printf("🆔 ID: %s\n", ctx.ID)
		fmt.Printf("📝 Название: %s\n", ctx.Name)
		fmt.Printf("📊 Доска: %s | Проект: %s\n", ctx.BoardID, ctx.ProjectID)
		fmt.Printf("🔧 Провайдер: %s\n", ctx.ProviderName)

		// Предлагаем сделать активным
		fmt.Print("🔄 Сделать этот контекст активным? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
			if err := cm.SetActiveContext(ctx.ID); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Ошибка активации: %v\n", err)
			} else {
				fmt.Printf("✅ Контекст '%s' активирован!\n", ctx.Name)
			}
		}
	},
}

// switchCmd - переключение контекста
var switchCmd = &cobra.Command{
	Use:   "switch [context-id-or-name]",
	Short: "Переключиться на другой контекст",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		log := &ContextLogger{}
		cm := context.NewContextManager("", log)

		contexts := cm.ListContexts()
		if len(contexts) == 0 {
			fmt.Println("📭 Нет доступных контекстов")
			return
		}

		var targetID string

		// Если передан аргумент, ищем контекст
		if len(args) > 0 {
			search := args[0]
			for _, ctx := range contexts {
				if ctx.ID == search || strings.EqualFold(ctx.Name, search) {
					targetID = ctx.ID
					break
				}
			}

			if targetID == "" {
				fmt.Printf("❌ Контекст '%s' не найден\n", search)
				os.Exit(1)
			}
		} else {
			// Интерактивный выбор
			fmt.Println("📋 Выберите контекст:")
			for i, ctx := range contexts {
				status := ""
				if ctx.IsActive {
					status = "🟢"
				} else {
					status = "⚪"
				}
				fmt.Printf("%d. %s %s - %s\n", i+1, status, ctx.Name, ctx.Description)
			}

			fmt.Print("Введите номер контекста: ")
			var choice int
			fmt.Scanln(&choice)

			if choice < 1 || choice > len(contexts) {
				fmt.Println("❌ Неверный выбор")
				os.Exit(1)
			}

			targetID = contexts[choice-1].ID
		}

		// Переключаемся на контекст
		if err := cm.SetActiveContext(targetID); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Ошибка переключения: %v\n", err)
			os.Exit(1)
		}

		// Получаем информацию об активном контексте
		ctx, _ := cm.GetActiveContext()
		fmt.Printf("✅ Переключено на контекст '%s'\n", ctx.Name)
		fmt.Printf("📊 Доска: %s | Проект: %s\n", ctx.BoardID, ctx.ProjectID)
	},
}

// currentCmd - текущий контекст
var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Показать текущий активный контекст",
	Run: func(cmd *cobra.Command, args []string) {
		log := &ContextLogger{}
		cm := context.NewContextManager("", log)

		ctx, err := cm.GetActiveContext()
		if err != nil {
			fmt.Println("📭 Нет активного контекста")
			fmt.Println("Создайте новый: ricochet context create")
			return
		}

		fmt.Println("🎯 Текущий рабочий контекст:")
		fmt.Println(strings.Repeat("=", 40))
		fmt.Printf("📝 Название: %s\n", ctx.Name)
		fmt.Printf("📋 Описание: %s\n", ctx.Description)
		fmt.Printf("🆔 ID: %s\n", ctx.ID)
		fmt.Println()

		fmt.Println("📊 Настройки доски:")
		fmt.Printf("  • Доска: %s\n", ctx.BoardID)
		fmt.Printf("  • Проект: %s\n", ctx.ProjectID)
		fmt.Printf("  • Провайдер: %s\n", ctx.ProviderName)
		fmt.Println()

		fmt.Println("⚙️ Настройки по умолчанию:")
		fmt.Printf("  • Исполнитель: %s\n", ctx.DefaultAssignee)
		fmt.Printf("  • Приоритет: %s\n", ctx.DefaultPriority)
		fmt.Printf("  • Метки: %v\n", ctx.DefaultLabels)
		fmt.Println()

		fmt.Println("🚀 Параметры проекта:")
		fmt.Printf("  • Тип: %s\n", ctx.ProjectType)
		fmt.Printf("  • Сложность: %s\n", ctx.Complexity)
		fmt.Printf("  • Временные рамки: %d дней\n", ctx.Timeline)
		fmt.Printf("  • Размер команды: %d\n", ctx.TeamSize)
		fmt.Printf("  • AI включен: %v\n", ctx.AIEnabled)
		fmt.Printf("  • Автоназначение: %v\n", ctx.AutoAssignment)
		fmt.Println()

		if ctx.Stats != nil {
			fmt.Println("📈 Статистика:")
			fmt.Printf("  • Задач создано: %d\n", ctx.Stats.TasksCreated)
			fmt.Printf("  • Задач выполнено: %d\n", ctx.Stats.TasksCompleted)
			fmt.Printf("  • Планов создано: %d\n", ctx.Stats.PlansGenerated)
			if ctx.Stats.TasksCreated > 0 {
				fmt.Printf("  • Успешность: %.1f%%\n", ctx.Stats.SuccessRate*100)
			}
			if !ctx.Stats.LastActivity.IsZero() {
				fmt.Printf("  • Последняя активность: %s\n", 
					ctx.Stats.LastActivity.Format("2006-01-02 15:04"))
			}
		}

		fmt.Println()
		fmt.Printf("🕐 Создан: %s\n", ctx.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("🔄 Обновлен: %s\n", ctx.UpdatedAt.Format("2006-01-02 15:04:05"))
	},
}

// updateCmd - обновление контекста
var updateCmd = &cobra.Command{
	Use:   "update [context-id]",
	Short: "Обновить настройки контекста",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		log := &ContextLogger{}
		cm := context.NewContextManager("", log)

		var targetID string

		// Определяем ID контекста для обновления
		if len(args) > 0 {
			targetID = args[0]
		} else {
			// Используем активный контекст
			ctx, err := cm.GetActiveContext()
			if err != nil {
				fmt.Println("❌ Нет активного контекста для обновления")
				os.Exit(1)
			}
			targetID = ctx.ID
		}

		// Собираем обновления
		updates := make(map[string]interface{})

		if cmd.Flags().Changed("name") {
			updates["name"] = contextName
		}
		if cmd.Flags().Changed("description") {
			updates["description"] = contextDescription
		}
		if cmd.Flags().Changed("board-id") {
			updates["board_id"] = boardID
		}
		if cmd.Flags().Changed("project-id") {
			updates["project_id"] = projectID
		}
		if cmd.Flags().Changed("provider") {
			updates["provider_name"] = providerName
		}
		if cmd.Flags().Changed("assignee") {
			updates["default_assignee"] = defaultAssignee
		}
		if cmd.Flags().Changed("priority") {
			updates["default_priority"] = defaultPriority
		}
		if cmd.Flags().Changed("type") {
			updates["project_type"] = projectType
		}
		if cmd.Flags().Changed("complexity") {
			updates["complexity"] = complexity
		}
		if cmd.Flags().Changed("timeline") {
			updates["timeline"] = timeline
		}
		if cmd.Flags().Changed("team-size") {
			updates["team_size"] = teamSize
		}
		if cmd.Flags().Changed("ai") {
			updates["ai_enabled"] = aiEnabled
		}

		if len(updates) == 0 {
			fmt.Println("❌ Нет изменений для обновления")
			fmt.Println("Используйте флаги для указания изменений, например:")
			fmt.Println("  ricochet context update --name \"Новое название\"")
			return
		}

		// Обновляем контекст
		if err := cm.UpdateContext(targetID, updates); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Ошибка обновления: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Контекст обновлен!\n")
		fmt.Printf("🔄 Обновлено полей: %d\n", len(updates))
	},
}

// deleteCmd - удаление контекста
var deleteCmd = &cobra.Command{
	Use:   "delete [context-id-or-name]",
	Short: "Удалить контекст",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log := &ContextLogger{}
		cm := context.NewContextManager("", log)

		search := args[0]
		contexts := cm.ListContexts()

		var targetID string
		var targetName string

		// Находим контекст
		for _, ctx := range contexts {
			if ctx.ID == search || strings.EqualFold(ctx.Name, search) {
				targetID = ctx.ID
				targetName = ctx.Name
				break
			}
		}

		if targetID == "" {
			fmt.Printf("❌ Контекст '%s' не найден\n", search)
			os.Exit(1)
		}

		// Подтверждение удаления
		fmt.Printf("⚠️  Вы уверены что хотите удалить контекст '%s'? (y/N): ", targetName)
		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("❌ Удаление отменено")
			return
		}

		// Удаляем контекст
		if err := cm.DeleteContext(targetID); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Ошибка удаления: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Контекст '%s' удален\n", targetName)
	},
}

// analyzeCmd - анализ проекта
var analyzeCmd = &cobra.Command{
	Use:   "analyze [description]",
	Short: "Анализ проекта для создания контекста",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		log := &ContextLogger{}
		
		// Получаем описание проекта
		description := ""
		if len(args) > 0 {
			description = args[0]
		} else {
			fmt.Print("📝 Введите описание проекта: ")
			fmt.Scanln(&description)
		}

		if description == "" {
			fmt.Println("❌ Описание проекта обязательно")
			os.Exit(1)
		}

		// Создаем анализатор
		aiChains := &ai.AIChains{} // TODO: Подключить реальные AI chains
		analyzer := context.NewProjectAnalyzer(aiChains, log)

		fmt.Println("🔍 Анализирую проект...")
		
		// Выполняем анализ
		analysis, err := analyzer.AnalyzeProject(description, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Ошибка анализа: %v\n", err)
			os.Exit(1)
		}

		// Выводим результаты анализа
		fmt.Println("\n📊 Результаты анализа:")
		fmt.Println(strings.Repeat("=", 50))
		fmt.Printf("📝 Название: %s\n", analysis.ProjectName)
		fmt.Printf("🔧 Тип проекта: %s\n", analysis.ProjectType)
		fmt.Printf("⚙️  Фреймворк: %s\n", analysis.Framework)
		fmt.Printf("📊 Сложность: %s\n", analysis.Complexity)
		fmt.Printf("⏱️  Оценка времени: %d часов\n", analysis.EstimatedHours)
		fmt.Printf("👥 Размер команды: %d\n", analysis.TeamSize)
		fmt.Printf("📅 Временные рамки: %d дней\n", analysis.Timeline)
		fmt.Printf("🎯 Уверенность: %.1f%%\n", analysis.Confidence*100)

		if len(analysis.RequiredSkills) > 0 {
			fmt.Printf("🎓 Необходимые навыки: %s\n", strings.Join(analysis.RequiredSkills, ", "))
		}

		if len(analysis.Risks) > 0 {
			fmt.Println("\n⚠️  Выявленные риски:")
			for _, risk := range analysis.Risks {
				fmt.Printf("  • %s (%s): %s\n", risk.Type, risk.Severity, risk.Description)
			}
		}

		if len(analysis.Recommendations) > 0 {
			fmt.Println("\n💡 Рекомендации:")
			for _, rec := range analysis.Recommendations {
				fmt.Printf("  • %s\n", rec)
			}
		}

		// Предлагаем создать контекст
		fmt.Print("\n🔄 Создать контекст на основе анализа? (y/N): ")
		var response string
		fmt.Scanln(&response)

		if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
			// Интерактивный ввод недостающих параметров
			fmt.Print("🎯 Введите ID доски: ")
			fmt.Scanln(&boardID)

			fmt.Print("📁 Введите ID проекта: ")
			fmt.Scanln(&projectID)

			fmt.Print("🔧 Введите имя провайдера: ")
			fmt.Scanln(&providerName)

			// Создаем контекст менеджер
			cm := context.NewContextManager("", log)

			// Создаем контекст на основе анализа
			ctx, err := cm.CreateContext(
				analysis.ProjectName,
				description,
				analysis.Context,
			)
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Ошибка создания контекста: %v\n", err)
				os.Exit(1)
			}

			// Обновляем контекст с данными доски
			updates := map[string]interface{}{
				"board_id":      boardID,
				"project_id":    projectID,
				"provider_name": providerName,
			}

			if err := cm.UpdateContext(ctx.ID, updates); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Ошибка обновления контекста: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✅ Контекст '%s' создан на основе анализа!\n", ctx.Name)
			fmt.Printf("🆔 ID: %s\n", ctx.ID)
		}
	},
}

// boardsCmd - управление досками в контексте
var boardsCmd = &cobra.Command{
	Use:   "boards",
	Short: "Управление досками в контексте",
}

func init() {
	boardsCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Список доступных досок",
		Run: func(cmd *cobra.Command, args []string) {
			log := &ContextLogger{}
			registry := context.NewBoardRegistry("", log)

			// TODO: Зарегистрировать провайдеры
			fmt.Println("📋 Доступные доски:")
			fmt.Println("(Функция в разработке - нужно подключить провайдеры)")
			
			_ = registry // временно, чтобы избежать ошибки компиляции
		},
	})

	boardsCmd.AddCommand(&cobra.Command{
		Use:   "sync",
		Short: "Синхронизировать доски с провайдерами",
		Run: func(cmd *cobra.Command, args []string) {
			log := &ContextLogger{}
			registry := context.NewBoardRegistry("", log)

			fmt.Println("🔄 Синхронизация досок...")
			if err := registry.SyncBoards(); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Ошибка синхронизации: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✅ Синхронизация завершена")
		},
	})
}

// multiCmd - мульти-проектные контексты
var multiCmd = &cobra.Command{
	Use:   "multi",
	Short: "Управление мульти-проектными контекстами",
}

func init() {
	multiCmd.AddCommand(&cobra.Command{
		Use:   "set [context-ids...]",
		Short: "Установить несколько активных контекстов",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			log := &ContextLogger{}
			cm := context.NewContextManager("", log)

			if err := cm.SetMultiProjectContext(args); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Ошибка установки мульти-контекста: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✅ Установлено %d активных контекстов\n", len(args))
			
			// Показываем активные контексты
			activeContexts := cm.GetActiveContexts()
			for _, ctx := range activeContexts {
				fmt.Printf("  🟢 %s - %s\n", ctx.Name, ctx.ProjectID)
			}
		},
	})

	multiCmd.AddCommand(&cobra.Command{
		Use:   "active",
		Short: "Показать все активные контексты",
		Run: func(cmd *cobra.Command, args []string) {
			log := &ContextLogger{}
			cm := context.NewContextManager("", log)

			activeContexts := cm.GetActiveContexts()
			if len(activeContexts) == 0 {
				fmt.Println("📭 Нет активных контекстов")
				return
			}

			fmt.Printf("🟢 Активные контексты (%d):\n", len(activeContexts))
			fmt.Println(strings.Repeat("=", 40))

			for i, ctx := range activeContexts {
				fmt.Printf("%d. %s\n", i+1, ctx.Name)
				fmt.Printf("   📊 Доска: %s | Проект: %s\n", ctx.BoardID, ctx.ProjectID)
				fmt.Printf("   🔧 Провайдер: %s\n", ctx.ProviderName)
				fmt.Println()
			}
		},
	})
}