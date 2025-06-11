package ricochet_task

import (
	"fmt"
	"os"

	"github.com/grik-ai/ricochet-task/internal/config"
	"github.com/grik-ai/ricochet-task/pkg/chain"
	"github.com/grik-ai/ricochet-task/pkg/task"
	"github.com/grik-ai/ricochet-task/pkg/ui"
	"github.com/spf13/cobra"
)

func init() {
	// Команда для создания задачи на основе цепочки
	var createTaskCmd = &cobra.Command{
		Use:   "create-task",
		Short: "Создать задачу Ricochet Task на основе цепочки",
		Run: func(cmd *cobra.Command, args []string) {
			chainID, _ := cmd.Flags().GetString("chain-id")
			if chainID == "" {
				ui.PrintError("Необходимо указать ID цепочки")
				return
			}

			taskID, err := createTaskFromChain(chainID)
			if err != nil {
				ui.PrintError(fmt.Sprintf("Ошибка создания задачи: %s", err))
				return
			}

			ui.PrintSuccess(fmt.Sprintf("Задача успешно создана с ID: %s", taskID))
		},
	}
	createTaskCmd.Flags().String("chain-id", "", "ID цепочки Ricochet")
	createTaskCmd.MarkFlagRequired("chain-id")

	// Команда для создания цепочки на основе задачи
	var createChainCmd = &cobra.Command{
		Use:   "create-chain",
		Short: "Создать цепочку на основе задачи Ricochet Task",
		Run: func(cmd *cobra.Command, args []string) {
			taskID, _ := cmd.Flags().GetString("task-id")
			if taskID == "" {
				ui.PrintError("Необходимо указать ID задачи")
				return
			}

			chainID, err := createChainFromTask(taskID)
			if err != nil {
				ui.PrintError(fmt.Sprintf("Ошибка создания цепочки: %s", err))
				return
			}

			ui.PrintSuccess(fmt.Sprintf("Цепочка успешно создана с ID: %s", chainID))
		},
	}
	createChainCmd.Flags().String("task-id", "", "ID задачи Ricochet Task")
	createChainCmd.MarkFlagRequired("task-id")

	// Команда для синхронизации статуса задачи
	var syncStatusCmd = &cobra.Command{
		Use:   "sync-status",
		Short: "Синхронизировать статус задачи с прогрессом выполнения цепочки",
		Run: func(cmd *cobra.Command, args []string) {
			taskID, _ := cmd.Flags().GetString("task-id")
			chainID, _ := cmd.Flags().GetString("chain-id")

			if taskID == "" {
				ui.PrintError("Необходимо указать ID задачи")
				return
			}

			// Если chainID не указан, пытаемся найти его по задаче
			if chainID == "" {
				var err error
				chainID, err = getChainForTask(taskID)
				if err != nil {
					ui.PrintError(fmt.Sprintf("Ошибка получения цепочки для задачи: %s", err))
					return
				}
			}

			err := syncTaskStatus(taskID, chainID)
			if err != nil {
				ui.PrintError(fmt.Sprintf("Ошибка синхронизации статуса: %s", err))
				return
			}

			ui.PrintSuccess("Статус задачи успешно синхронизирован")
		},
	}
	syncStatusCmd.Flags().String("task-id", "", "ID задачи Ricochet Task")
	syncStatusCmd.Flags().String("chain-id", "", "ID цепочки Ricochet (опционально)")
	syncStatusCmd.MarkFlagRequired("task-id")

	// Команда для получения списка задач
	var listTasksCmd = &cobra.Command{
		Use:   "list-tasks",
		Short: "Получить список задач Ricochet Task",
		Run: func(cmd *cobra.Command, args []string) {
			status, _ := cmd.Flags().GetString("status")
			showSubtasks, _ := cmd.Flags().GetBool("subtasks")

			tasks, err := getRicochetTasks()
			if err != nil {
				ui.PrintError(fmt.Sprintf("Ошибка получения задач: %s", err))
				return
			}

			printTasks(tasks, status, showSubtasks)
		},
	}
	listTasksCmd.Flags().String("status", "", "Фильтр по статусу задач (pending, in-progress, done)")
	listTasksCmd.Flags().Bool("subtasks", false, "Показывать подзадачи")

	// Команда для получения информации о задаче
	var showTaskCmd = &cobra.Command{
		Use:   "show-task",
		Short: "Показать информацию о задаче Ricochet Task",
		Run: func(cmd *cobra.Command, args []string) {
			taskID, _ := cmd.Flags().GetString("task-id")
			if taskID == "" {
				ui.PrintError("Необходимо указать ID задачи")
				return
			}

			task, err := getRicochetTask(taskID)
			if err != nil {
				ui.PrintError(fmt.Sprintf("Ошибка получения задачи: %s", err))
				return
			}

			printTaskDetails(task)
		},
	}
	showTaskCmd.Flags().String("task-id", "", "ID задачи Ricochet Task")
	showTaskCmd.MarkFlagRequired("task-id")

	// Команда для обновления статуса задачи
	var updateStatusCmd = &cobra.Command{
		Use:   "update-status",
		Short: "Обновить статус задачи Ricochet Task",
		Run: func(cmd *cobra.Command, args []string) {
			taskID, _ := cmd.Flags().GetString("task-id")
			status, _ := cmd.Flags().GetString("status")

			if taskID == "" {
				ui.PrintError("Необходимо указать ID задачи")
				return
			}

			if status == "" {
				ui.PrintError("Необходимо указать статус задачи")
				return
			}

			// Проверяем корректность статуса
			if !isValidStatus(status) {
				ui.PrintError("Некорректный статус. Допустимые значения: pending, in-progress, done, deferred, blocked, review")
				return
			}

			err := updateTaskStatus(taskID, status)
			if err != nil {
				ui.PrintError(fmt.Sprintf("Ошибка обновления статуса: %s", err))
				return
			}

			ui.PrintSuccess(fmt.Sprintf("Статус задачи %s успешно обновлен на %s", taskID, status))
		},
	}
	updateStatusCmd.Flags().String("task-id", "", "ID задачи Ricochet Task")
	updateStatusCmd.Flags().String("status", "", "Новый статус задачи (pending, in-progress, done, deferred, blocked, review)")
	updateStatusCmd.MarkFlagRequired("task-id")
	updateStatusCmd.MarkFlagRequired("status")

	// Добавляем команды
	TaskCmd.AddCommand(createTaskCmd)
	TaskCmd.AddCommand(createChainCmd)
	TaskCmd.AddCommand(syncStatusCmd)
	TaskCmd.AddCommand(listTasksCmd)
	TaskCmd.AddCommand(showTaskCmd)
	TaskCmd.AddCommand(updateStatusCmd)
}

// isValidStatus проверяет корректность статуса
func isValidStatus(status string) bool {
	validStatuses := []string{
		task.RicochetTaskStatusPending,
		task.RicochetTaskStatusProgress,
		task.RicochetTaskStatusDone,
		task.RicochetTaskStatusDeferred,
		task.RicochetTaskStatusBlocked,
		task.RicochetTaskStatusReview,
	}

	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}

	return false
}

// createTaskFromChain создает задачу Ricochet Task на основе цепочки
func createTaskFromChain(chainID string) (string, error) {
	converter, err := getRicochetTaskConverter()
	if err != nil {
		return "", err
	}

	return converter.CreateTaskFromChain(chainID)
}

// createChainFromTask создает цепочку на основе задачи Ricochet Task
func createChainFromTask(taskID string) (string, error) {
	converter, err := getRicochetTaskConverter()
	if err != nil {
		return "", err
	}

	return converter.CreateChainFromTask(taskID)
}

// syncTaskStatus синхронизирует статус задачи с прогрессом выполнения цепочки
func syncTaskStatus(taskID, chainID string) error {
	converter, err := getRicochetTaskConverter()
	if err != nil {
		return err
	}

	return converter.SyncTaskStatus(taskID, chainID)
}

// getRicochetTasks возвращает список задач из Ricochet Task
func getRicochetTasks() ([]task.RicochetTaskTask, error) {
	converter, err := getRicochetTaskConverter()
	if err != nil {
		return nil, err
	}

	return converter.GetRicochetTaskTasks()
}

// getRicochetTask возвращает задачу по ID
func getRicochetTask(taskID string) (task.RicochetTaskTask, error) {
	tasks, err := getRicochetTasks()
	if err != nil {
		return task.RicochetTaskTask{}, err
	}

	// Ищем задачу с указанным ID
	for _, t := range tasks {
		if t.ID == taskID {
			return t, nil
		}

		// Проверяем подзадачи
		for _, st := range t.Subtasks {
			if st.ID == taskID {
				return st, nil
			}
		}
	}

	return task.RicochetTaskTask{}, fmt.Errorf("задача с ID %s не найдена", taskID)
}

// updateTaskStatus обновляет статус задачи
func updateTaskStatus(taskID, status string) error {
	// Получаем задачу
	taskObj, err := getRicochetTask(taskID)
	if err != nil {
		return err
	}

	// Обновляем статус
	taskObj.Status = status

	// Сохраняем изменения
	converter, err := getRicochetTaskConverter()
	if err != nil {
		return err
	}

	return converter.UpdateRicochetTaskTask(taskObj)
}

// getChainForTask возвращает ID цепочки для указанной задачи
func getChainForTask(taskID string) (string, error) {
	converter, err := getRicochetTaskConverter()
	if err != nil {
		return "", err
	}

	return converter.GetChainForTask(taskID)
}

// printTasks выводит список задач
func printTasks(tasks []task.RicochetTaskTask, statusFilter string, showSubtasks bool) {
	if len(tasks) == 0 {
		fmt.Println("Задачи не найдены")
		return
	}

	fmt.Println("\n=== Список задач Ricochet Task ===")
	fmt.Println()

	for _, t := range tasks {
		// Фильтрация по статусу
		if statusFilter != "" && t.Status != statusFilter {
			continue
		}

		// Эмодзи статуса
		statusEmoji := getStatusEmoji(t.Status)

		// Выводим информацию о задаче
		fmt.Printf("%s [%s] %s: %s\n", statusEmoji, t.ID, t.Priority, t.Title)

		// Выводим информацию о подзадачах, если необходимо
		if showSubtasks && len(t.Subtasks) > 0 {
			for _, st := range t.Subtasks {
				// Фильтрация по статусу
				if statusFilter != "" && st.Status != statusFilter {
					continue
				}

				subtaskStatusEmoji := getStatusEmoji(st.Status)
				fmt.Printf("  %s [%s] %s: %s\n", subtaskStatusEmoji, st.ID, st.Priority, st.Title)
			}
		}
	}
}

// printTaskDetails выводит детальную информацию о задаче
func printTaskDetails(t task.RicochetTaskTask) {
	fmt.Println("\n=== Информация о задаче ===")
	fmt.Printf("ID: %s\n", t.ID)
	fmt.Printf("Название: %s\n", t.Title)
	fmt.Printf("Статус: %s %s\n", getStatusEmoji(t.Status), t.Status)
	fmt.Printf("Приоритет: %s\n", t.Priority)
	fmt.Println("Описание:")
	fmt.Println(t.Description)

	if t.Details != "" {
		fmt.Println("\nДетали реализации:")
		fmt.Println(t.Details)
	}

	if t.TestStrategy != "" {
		fmt.Println("\nСтратегия тестирования:")
		fmt.Println(t.TestStrategy)
	}

	if len(t.Dependencies) > 0 {
		fmt.Println("\nЗависимости:")
		for _, dep := range t.Dependencies {
			fmt.Printf("- %s\n", dep)
		}
	}

	if len(t.Subtasks) > 0 {
		fmt.Println("\nПодзадачи:")
		for _, st := range t.Subtasks {
			statusEmoji := getStatusEmoji(st.Status)
			fmt.Printf("%s [%s] %s: %s\n", statusEmoji, st.ID, st.Priority, st.Title)
		}
	}

	if t.Metadata != nil {
		fmt.Println("\nМетаданные:")
		for k, v := range t.Metadata {
			fmt.Printf("- %s: %v\n", k, v)
		}
	}
}

// getStatusEmoji возвращает эмодзи для статуса
func getStatusEmoji(status string) string {
	switch status {
	case task.RicochetTaskStatusPending:
		return "⏱️"
	case task.RicochetTaskStatusProgress:
		return "🔄"
	case task.RicochetTaskStatusDone:
		return "✅"
	case task.RicochetTaskStatusDeferred:
		return "⏳"
	case task.RicochetTaskStatusBlocked:
		return "🚫"
	case task.RicochetTaskStatusReview:
		return "👀"
	default:
		return "❓"
	}
}

// getRicochetTaskConverter возвращает конвертер Ricochet Task
func getRicochetTaskConverter() (*task.DefaultRicochetTaskConverter, error) {
	// Получаем рабочую директорию
	workspacePath, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить рабочую директорию: %w", err)
	}

	// Получаем конфигурацию
	configPath, err := config.GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить путь к конфигурации: %w", err)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("не удалось загрузить конфигурацию: %w", err)
	}

	// Создаем хранилище цепочек
	chainStore, err := chain.NewFileChainStore(cfg.ConfigDir)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать хранилище цепочек: %w", err)
	}

	// Создаем хранилище задач
	taskStore, err := task.NewFileTaskStore(cfg.ConfigDir)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать хранилище задач: %w", err)
	}

	// Создаем менеджер задач
	taskManager := task.NewTaskManager(taskStore)

	// Создаем конвертер
	converter, err := task.NewRicochetTaskConverter(workspacePath, taskManager, chainStore)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать конвертер: %w", err)
	}

	return converter, nil
}
