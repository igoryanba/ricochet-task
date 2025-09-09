package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/grik-ai/ricochet-task/pkg/chain"
	"github.com/grik-ai/ricochet-task/pkg/task"
	"github.com/grik-ai/ricochet-task/internal/config"
)

// ShowRicochetTaskMenu отображает меню для управления Task Master
func ShowRicochetTaskMenu() error {
	options := []string{
		"Просмотреть список задач",
		"Просмотреть задачу по ID",
		"Обновить статус задачи",
		"Создать задачу из цепочки",
		"Создать цепочку из задачи",
		"Синхронизировать статус",
		"Назад",
	}

	var choice string
	prompt := &survey.Select{
		Message: "Выберите действие:",
		Options: options,
	}
	if err := survey.AskOne(prompt, &choice); err != nil {
		return err
	}

	switch choice {
	case "Просмотреть список задач":
		return HandleListTasks()
	case "Просмотреть задачу по ID":
		return HandleShowTask()
	case "Обновить статус задачи":
		return HandleUpdateTaskStatus()
	case "Создать задачу из цепочки":
		return HandleCreateTaskFromChain()
	case "Создать цепочку из задачи":
		return HandleCreateChainFromTask()
	case "Синхронизировать статус":
		return HandleSyncTaskStatus()
	case "Назад":
		return nil
	}

	return nil
}

// HandleListTasks отображает список задач
func HandleListTasks() error {
	// Запрашиваем фильтр по статусу
	var statusFilter string
	statusPrompt := &survey.Select{
		Message: "Фильтр по статусу:",
		Options: []string{"Все", "pending", "in-progress", "done", "deferred", "blocked", "review"},
		Default: "Все",
	}
	if err := survey.AskOne(statusPrompt, &statusFilter); err != nil {
		return err
	}

	// Запрашиваем, показывать ли подзадачи
	var showSubtasks bool
	subtasksPrompt := &survey.Confirm{
		Message: "Показывать подзадачи?",
		Default: false,
	}
	if err := survey.AskOne(subtasksPrompt, &showSubtasks); err != nil {
		return err
	}

	// Получаем конвертер
	converter, err := getRicochetTaskConverter()
	if err != nil {
		PrintError(fmt.Sprintf("Ошибка инициализации: %s", err))
		return err
	}

	// Получаем задачи
	tasks, err := converter.GetRicochetTaskTasks()
	if err != nil {
		PrintError(fmt.Sprintf("Ошибка получения задач: %s", err))
		return err
	}

	// Выводим задачи
	PrintSuccess("Список задач:")
	fmt.Println()

	for _, t := range tasks {
		// Фильтр по статусу
		if statusFilter != "Все" && t.Status != statusFilter {
			continue
		}

		// Эмодзи статуса
		statusEmoji := getStatusEmoji(t.Status)

		// Выводим информацию о задаче
		fmt.Printf("%s [%s] %s: %s\n", statusEmoji, t.ID, t.Priority, t.Title)

		// Выводим информацию о подзадачах, если необходимо
		if showSubtasks && len(t.Subtasks) > 0 {
			for _, st := range t.Subtasks {
				// Фильтр по статусу
				if statusFilter != "Все" && st.Status != statusFilter {
					continue
				}

				subtaskStatusEmoji := getStatusEmoji(st.Status)
				fmt.Printf("  %s [%s] %s: %s\n", subtaskStatusEmoji, st.ID, st.Priority, st.Title)
			}
		}
	}

	fmt.Println()
	return nil
}

// HandleShowTask отображает информацию о задаче
func HandleShowTask() error {
	// Запрашиваем ID задачи
	var taskID string
	taskIDPrompt := &survey.Input{
		Message: "Введите ID задачи:",
	}
	if err := survey.AskOne(taskIDPrompt, &taskID, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Получаем конвертер
	converter, err := getRicochetTaskConverter()
	if err != nil {
		PrintError(fmt.Sprintf("Ошибка инициализации: %s", err))
		return err
	}

	// Получаем задачи
	tasks, err := converter.GetRicochetTaskTasks()
	if err != nil {
		PrintError(fmt.Sprintf("Ошибка получения задач: %s", err))
		return err
	}

	// Ищем задачу с указанным ID
	var foundTask task.RicochetTaskTask
	found := false

	for _, t := range tasks {
		if t.ID == taskID {
			foundTask = t
			found = true
			break
		}

		// Проверяем подзадачи
		for _, st := range t.Subtasks {
			if st.ID == taskID {
				foundTask = st
				found = true
				break
			}
		}

		if found {
			break
		}
	}

	if !found {
		PrintError(fmt.Sprintf("Задача с ID %s не найдена", taskID))
		return nil
	}

	// Выводим информацию о задаче
	PrintSuccess(fmt.Sprintf("Информация о задаче %s:", taskID))
	fmt.Println()
	fmt.Printf("ID: %s\n", foundTask.ID)
	fmt.Printf("Название: %s\n", foundTask.Title)
	fmt.Printf("Статус: %s %s\n", getStatusEmoji(foundTask.Status), foundTask.Status)
	fmt.Printf("Приоритет: %s\n", foundTask.Priority)
	fmt.Println("Описание:")
	fmt.Println(foundTask.Description)

	if foundTask.Details != "" {
		fmt.Println("\nДетали реализации:")
		fmt.Println(foundTask.Details)
	}

	if foundTask.TestStrategy != "" {
		fmt.Println("\nСтратегия тестирования:")
		fmt.Println(foundTask.TestStrategy)
	}

	if len(foundTask.Dependencies) > 0 {
		fmt.Println("\nЗависимости:")
		for _, dep := range foundTask.Dependencies {
			fmt.Printf("- %s\n", dep)
		}
	}

	if len(foundTask.Subtasks) > 0 {
		fmt.Println("\nПодзадачи:")
		for _, st := range foundTask.Subtasks {
			statusEmoji := getStatusEmoji(st.Status)
			fmt.Printf("%s [%s] %s: %s\n", statusEmoji, st.ID, st.Priority, st.Title)
		}
	}

	if foundTask.Metadata != nil {
		fmt.Println("\nМетаданные:")
		for k, v := range foundTask.Metadata {
			fmt.Printf("- %s: %v\n", k, v)
		}
	}

	fmt.Println()
	return nil
}

// HandleUpdateTaskStatus обновляет статус задачи
func HandleUpdateTaskStatus() error {
	// Запрашиваем ID задачи
	var taskID string
	taskIDPrompt := &survey.Input{
		Message: "Введите ID задачи:",
	}
	if err := survey.AskOne(taskIDPrompt, &taskID, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	// Запрашиваем новый статус
	var status string
	statusPrompt := &survey.Select{
		Message: "Выберите новый статус:",
		Options: []string{"pending", "in-progress", "done", "deferred", "blocked", "review"},
	}
	if err := survey.AskOne(statusPrompt, &status); err != nil {
		return err
	}

	// Получаем конвертер
	converter, err := getRicochetTaskConverter()
	if err != nil {
		PrintError(fmt.Sprintf("Ошибка инициализации: %s", err))
		return err
	}

	// Получаем задачи
	tasks, err := converter.GetRicochetTaskTasks()
	if err != nil {
		PrintError(fmt.Sprintf("Ошибка получения задач: %s", err))
		return err
	}

	// Ищем задачу с указанным ID
	var foundTask task.RicochetTaskTask
	found := false

	for _, t := range tasks {
		if t.ID == taskID {
			foundTask = t
			found = true
			break
		}

		// Проверяем подзадачи
		for _, st := range t.Subtasks {
			if st.ID == taskID {
				foundTask = st
				found = true
				break
			}
		}

		if found {
			break
		}
	}

	if !found {
		PrintError(fmt.Sprintf("Задача с ID %s не найдена", taskID))
		return nil
	}

	// Обновляем статус
	foundTask.Status = status

	// Сохраняем задачу
	err = converter.UpdateRicochetTaskTask(foundTask)
	if err != nil {
		PrintError(fmt.Sprintf("Ошибка обновления задачи: %s", err))
		return err
	}

	PrintSuccess(fmt.Sprintf("Статус задачи %s успешно обновлен на %s", taskID, status))
	return nil
}

// HandleCreateTaskFromChain создает задачу на основе цепочки
func HandleCreateTaskFromChain() error {
	// Получаем список цепочек
	chainStore, err := getChainStore()
	if err != nil {
		PrintError(fmt.Sprintf("Ошибка инициализации хранилища цепочек: %s", err))
		return err
	}

	chains, err := chainStore.List()
	if err != nil {
		PrintError(fmt.Sprintf("Ошибка получения списка цепочек: %s", err))
		return err
	}

	if len(chains) == 0 {
		PrintError("Нет доступных цепочек")
		return nil
	}

	// Создаем список цепочек для выбора
	chainOptions := make([]string, len(chains))
	for i, ch := range chains {
		chainOptions[i] = fmt.Sprintf("%s: %s", ch.ID, ch.Name)
	}

	// Запрашиваем выбор цепочки
	var chainChoice string
	chainPrompt := &survey.Select{
		Message: "Выберите цепочку:",
		Options: chainOptions,
	}
	if err := survey.AskOne(chainPrompt, &chainChoice); err != nil {
		return err
	}

	// Извлекаем ID цепочки
	chainID := strings.Split(chainChoice, ":")[0]
	chainID = strings.TrimSpace(chainID)

	// Получаем конвертер
	converter, err := getRicochetTaskConverter()
	if err != nil {
		PrintError(fmt.Sprintf("Ошибка инициализации: %s", err))
		return err
	}

	// Создаем задачу
	taskID, err := converter.CreateTaskFromChain(chainID)
	if err != nil {
		PrintError(fmt.Sprintf("Ошибка создания задачи: %s", err))
		return err
	}

	PrintSuccess(fmt.Sprintf("Задача успешно создана с ID: %s", taskID))
	return nil
}

// HandleCreateChainFromTask создает цепочку на основе задачи
func HandleCreateChainFromTask() error {
	// Получаем конвертер
	converter, err := getRicochetTaskConverter()
	if err != nil {
		PrintError(fmt.Sprintf("Ошибка инициализации: %s", err))
		return err
	}

	// Получаем задачи
	tasks, err := converter.GetRicochetTaskTasks()
	if err != nil {
		PrintError(fmt.Sprintf("Ошибка получения задач: %s", err))
		return err
	}

	if len(tasks) == 0 {
		PrintError("Нет доступных задач")
		return nil
	}

	// Создаем список задач для выбора
	taskOptions := make([]string, 0)
	for _, t := range tasks {
		taskOptions = append(taskOptions, fmt.Sprintf("%s: %s", t.ID, t.Title))

		// Добавляем подзадачи
		for _, st := range t.Subtasks {
			taskOptions = append(taskOptions, fmt.Sprintf("%s: %s", st.ID, st.Title))
		}
	}

	// Запрашиваем выбор задачи
	var taskChoice string
	taskPrompt := &survey.Select{
		Message: "Выберите задачу:",
		Options: taskOptions,
	}
	if err := survey.AskOne(taskPrompt, &taskChoice); err != nil {
		return err
	}

	// Извлекаем ID задачи
	taskID := strings.Split(taskChoice, ":")[0]
	taskID = strings.TrimSpace(taskID)

	// Создаем цепочку
	chainID, err := converter.CreateChainFromTask(taskID)
	if err != nil {
		PrintError(fmt.Sprintf("Ошибка создания цепочки: %s", err))
		return err
	}

	PrintSuccess(fmt.Sprintf("Цепочка успешно создана с ID: %s", chainID))
	return nil
}

// HandleSyncTaskStatus синхронизирует статус задачи с прогрессом выполнения цепочки
func HandleSyncTaskStatus() error {
	// Получаем конвертер
	converter, err := getRicochetTaskConverter()
	if err != nil {
		PrintError(fmt.Sprintf("Ошибка инициализации: %s", err))
		return err
	}

	// Получаем задачи
	tasks, err := converter.GetRicochetTaskTasks()
	if err != nil {
		PrintError(fmt.Sprintf("Ошибка получения задач: %s", err))
		return err
	}

	if len(tasks) == 0 {
		PrintError("Нет доступных задач")
		return nil
	}

	// Создаем список задач для выбора
	taskOptions := make([]string, 0)
	for _, t := range tasks {
		taskOptions = append(taskOptions, fmt.Sprintf("%s: %s", t.ID, t.Title))

		// Добавляем подзадачи
		for _, st := range t.Subtasks {
			taskOptions = append(taskOptions, fmt.Sprintf("%s: %s", st.ID, st.Title))
		}
	}

	// Запрашиваем выбор задачи
	var taskChoice string
	taskPrompt := &survey.Select{
		Message: "Выберите задачу:",
		Options: taskOptions,
	}
	if err := survey.AskOne(taskPrompt, &taskChoice); err != nil {
		return err
	}

	// Извлекаем ID задачи
	taskID := strings.Split(taskChoice, ":")[0]
	taskID = strings.TrimSpace(taskID)

	// Получаем ID цепочки для задачи
	chainID, err := converter.GetChainForTask(taskID)
	if err != nil {
		// Если цепочка не найдена, запрашиваем ее ID
		var chainIDInput string
		chainIDPrompt := &survey.Input{
			Message: "Введите ID цепочки:",
		}
		if err := survey.AskOne(chainIDPrompt, &chainIDInput, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
		chainID = chainIDInput
	}

	// Синхронизируем статус
	err = converter.SyncTaskStatus(taskID, chainID)
	if err != nil {
		PrintError(fmt.Sprintf("Ошибка синхронизации статуса: %s", err))
		return err
	}

	PrintSuccess("Статус задачи успешно синхронизирован")
	return nil
}

// getStatusEmoji возвращает эмодзи для статуса
func getStatusEmoji(status string) string {
	switch status {
	case "pending":
		return "⏱️"
	case "in-progress":
		return "🔄"
	case "done":
		return "✅"
	case "deferred":
		return "⏳"
	case "blocked":
		return "🚫"
	case "review":
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
	return task.NewRicochetTaskConverter(workspacePath, taskManager, chainStore)
}

// getChainStore возвращает хранилище цепочек
func getChainStore() (chain.Store, error) {
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
	return chain.NewFileChainStore(cfg.ConfigDir)
}
