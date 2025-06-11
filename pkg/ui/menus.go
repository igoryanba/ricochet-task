package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
)

// MenuOption представляет опцию меню
type MenuOption struct {
	Name        string
	Description string
	Action      func() error
}

// ShowMainMenu отображает главное меню приложения
func ShowMainMenu() error {
	for {
		// Опции для главного меню
		options := []string{
			"⛓️  Управление цепочками моделей",
			"🔖 Управление чекпоинтами",
			"🔑 Управление API-ключами",
			"📋 Управление задачами (Task Master)",
			"❌ Выход",
		}

		// Заголовок
		PrintLogo()
		PrintWelcomeMessage()

		// Отображаем меню
		var choice string
		prompt := &survey.Select{
			Message: "Выберите действие:",
			Options: options,
		}
		if err := survey.AskOne(prompt, &choice); err != nil {
			return err
		}

		switch choice {
		case "⛓️  Управление цепочками моделей":
			if err := ShowChainManagementMenu(); err != nil {
				return err
			}
		case "🔖 Управление чекпоинтами":
			if err := ShowCheckpointManagementMenu(); err != nil {
				return err
			}
		case "🔑 Управление API-ключами":
			if err := ShowKeyManagementMenu(); err != nil {
				return err
			}
		case "📋 Управление задачами (Task Master)":
			if err := ShowRicochetTaskMenu(); err != nil {
				return err
			}
		case "❌ Выход":
			fmt.Println(InfoColor("Выход из приложения..."))
			return nil
		}
	}
}

// ShowKeyManagementMenu отображает меню управления API-ключами
func ShowKeyManagementMenu() error {
	options := []string{
		"➕ Добавить API-ключ",
		"📋 Просмотреть API-ключи",
		"🔄 Обновить API-ключ",
		"❌ Удалить API-ключ",
		"⬅️ Назад в главное меню",
	}

	DrawBox("Управление API-ключами", "Здесь вы можете добавлять, просматривать, обновлять и удалять API-ключи для различных моделей.", 60)

	var choice string
	prompt := &survey.Select{
		Message: "Выберите действие:",
		Options: options,
	}

	err := survey.AskOne(prompt, &choice)
	if err != nil {
		return err
	}

	switch choice {
	case "➕ Добавить API-ключ":
		return handleAddKey()
	case "📋 Просмотреть API-ключи":
		return handleListKeys()
	case "🔄 Обновить API-ключ":
		return handleUpdateKey()
	case "❌ Удалить API-ключ":
		return handleDeleteKey()
	case "⬅️ Назад в главное меню":
		return ShowMainMenu()
	}

	return nil
}

// ShowChainManagementMenu отображает меню управления цепочками моделей
func ShowChainManagementMenu() error {
	options := []string{
		"➕ Создать цепочку моделей",
		"📋 Просмотреть цепочки моделей",
		"🔄 Обновить цепочку моделей",
		"❌ Удалить цепочку моделей",
		"▶️ Запустить цепочку моделей",
		"⬅️ Назад в главное меню",
	}

	DrawBox("Управление цепочками моделей", "Здесь вы можете создавать, просматривать, обновлять, удалять и запускать цепочки моделей.", 60)

	var choice string
	prompt := &survey.Select{
		Message: "Выберите действие:",
		Options: options,
	}

	err := survey.AskOne(prompt, &choice)
	if err != nil {
		return err
	}

	switch choice {
	case "➕ Создать цепочку моделей":
		return handleCreateChain()
	case "📋 Просмотреть цепочки моделей":
		return handleListChains()
	case "🔄 Обновить цепочку моделей":
		return handleUpdateChain()
	case "❌ Удалить цепочку моделей":
		return handleDeleteChain()
	case "▶️ Запустить цепочку моделей":
		return handleRunChain()
	case "⬅️ Назад в главное меню":
		return ShowMainMenu()
	}

	return nil
}

// ShowCheckpointManagementMenu отображает меню управления чекпоинтами
func ShowCheckpointManagementMenu() error {
	options := []string{
		"➕ Создать чекпоинт",
		"📋 Просмотреть чекпоинты",
		"🔄 Обновить чекпоинт",
		"❌ Удалить чекпоинт",
		"⬅️ Назад в главное меню",
	}

	DrawBox("Управление чекпоинтами", "Здесь вы можете создавать, просматривать, обновлять и удалять чекпоинты для сохранения прогресса обработки.", 60)

	var choice string
	prompt := &survey.Select{
		Message: "Выберите действие:",
		Options: options,
	}

	err := survey.AskOne(prompt, &choice)
	if err != nil {
		return err
	}

	switch choice {
	case "➕ Создать чекпоинт":
		return handleCreateCheckpoint()
	case "📋 Просмотреть чекпоинты":
		return handleListCheckpoints()
	case "🔄 Обновить чекпоинт":
		return handleUpdateCheckpoint()
	case "❌ Удалить чекпоинт":
		return handleDeleteCheckpoint()
	case "⬅️ Назад в главное меню":
		return ShowMainMenu()
	}

	return nil
}

// HandleInitProject - публичная функция для инициализации проекта
func HandleInitProject() error {
	return handleInitProject()
}

// HandleAddKey - публичная функция для добавления API-ключа
func HandleAddKey() error {
	return handleAddKey()
}

// HandleListKeys - публичная функция для просмотра API-ключей
func HandleListKeys() error {
	return handleListKeys()
}

// HandleUpdateKey - публичная функция для обновления API-ключа
func HandleUpdateKey() error {
	return handleUpdateKey()
}

// HandleDeleteKey - публичная функция для удаления API-ключа
func HandleDeleteKey() error {
	return handleDeleteKey()
}

// HandleCreateChain - публичная функция для создания цепочки моделей
func HandleCreateChain() error {
	return handleCreateChain()
}

// HandleListChains - публичная функция для просмотра цепочек моделей
func HandleListChains() error {
	return handleListChains()
}

// HandleUpdateChain - публичная функция для обновления цепочки моделей
func HandleUpdateChain() error {
	return handleUpdateChain()
}

// HandleDeleteChain - публичная функция для удаления цепочки моделей
func HandleDeleteChain() error {
	return handleDeleteChain()
}

// HandleRunChain - публичная функция для запуска цепочки моделей
func HandleRunChain() error {
	return handleRunChain()
}

// HandleCreateCheckpoint - публичная функция для создания чекпоинта
func HandleCreateCheckpoint() error {
	return handleCreateCheckpoint()
}

// HandleListCheckpoints - публичная функция для просмотра чекпоинтов
func HandleListCheckpoints() error {
	return handleListCheckpoints()
}

// HandleUpdateCheckpoint - публичная функция для обновления чекпоинта
func HandleUpdateCheckpoint() error {
	return handleUpdateCheckpoint()
}

// HandleDeleteCheckpoint - публичная функция для удаления чекпоинта
func HandleDeleteCheckpoint() error {
	return handleDeleteCheckpoint()
}

// Заглушки для обработчиков меню
func handleInitProject() error {
	PrintInfo("Инициализация проекта...")
	// TODO: Реализовать инициализацию проекта
	s := CreateSpinner("Создание конфигурации...")
	s.Start()
	// Имитация работы
	for i := 0; i < 5; i++ {
		time.Sleep(300 * time.Millisecond)
	}
	s.Stop()
	PrintSuccess("Проект успешно инициализирован!")
	return ShowMainMenu()
}

func handleAddKey() error {
	PrintInfo("Добавление нового API-ключа")

	// Выбор провайдера
	providers := []string{"OpenAI", "Anthropic", "Google", "Cohere", "Другой..."}
	provider, err := SelectPrompt("Выберите провайдера API:", providers)
	if err != nil {
		return err
	}

	// Если выбран "Другой", запрашиваем имя провайдера
	if provider == "Другой..." {
		provider, err = InputPrompt("Введите имя провайдера:")
		if err != nil {
			return err
		}
	}

	// Запрашиваем API-ключ
	key, err := InputPrompt("Введите API-ключ для " + provider + ":")
	if err != nil {
		return err
	}

	// Запрашиваем описание
	description, err := InputPrompt("Введите описание (опционально):")
	if err != nil {
		return err
	}

	// Имитация сохранения
	s := CreateSpinner("Сохранение API-ключа...")
	s.Start()
	time.Sleep(1 * time.Second)
	s.Stop()

	// Отображаем информацию о сохраненном ключе
	var maskedKey string
	if len(key) > 4 {
		maskedKey = strings.Repeat("*", len(key)-4) + key[len(key)-4:]
	} else {
		maskedKey = key // Если ключ слишком короткий, не маскируем
	}

	keyInfo := fmt.Sprintf("Провайдер: %s\nОписание: %s\nКлюч: %s",
		provider,
		description,
		maskedKey)

	DrawBox("Добавлен новый API-ключ", keyInfo, 50)
	PrintSuccess("API-ключ успешно добавлен!")

	// Возвращаемся в меню управления ключами
	return ShowKeyManagementMenu()
}

func handleListKeys() error {
	PrintInfo("Просмотр API-ключей")
	// TODO: Реализовать отображение ключей
	DrawBox("Доступные API-ключи", "OpenAI: ****************ABCD\nAnthropic: **************1234\nGoogle: *****************WXYZ", 50)
	return ShowKeyManagementMenu()
}

func handleUpdateKey() error {
	PrintInfo("Обновление API-ключа")
	// TODO: Реализовать обновление ключа
	return ShowKeyManagementMenu()
}

func handleDeleteKey() error {
	PrintInfo("Удаление API-ключа")
	// TODO: Реализовать удаление ключа
	return ShowKeyManagementMenu()
}

func handleCreateChain() error {
	PrintInfo("Создание цепочки моделей")
	// TODO: Реализовать создание цепочки
	return ShowChainManagementMenu()
}

func handleListChains() error {
	PrintInfo("Просмотр цепочек моделей")
	// TODO: Реализовать просмотр цепочек
	return ShowChainManagementMenu()
}

func handleUpdateChain() error {
	PrintInfo("Обновление цепочки моделей")
	// TODO: Реализовать обновление цепочки
	return ShowChainManagementMenu()
}

func handleDeleteChain() error {
	PrintInfo("Удаление цепочки моделей")
	// TODO: Реализовать удаление цепочки
	return ShowChainManagementMenu()
}

func handleRunChain() error {
	PrintInfo("Запуск цепочки моделей")
	// TODO: Реализовать запуск цепочки
	return ShowChainManagementMenu()
}

func handleCreateCheckpoint() error {
	PrintInfo("Создание чекпоинта")
	// TODO: Реализовать создание чекпоинта
	return ShowCheckpointManagementMenu()
}

func handleListCheckpoints() error {
	PrintInfo("Просмотр чекпоинтов")
	// TODO: Реализовать просмотр чекпоинтов
	return ShowCheckpointManagementMenu()
}

func handleUpdateCheckpoint() error {
	PrintInfo("Обновление чекпоинта")
	// TODO: Реализовать обновление чекпоинта
	return ShowCheckpointManagementMenu()
}

func handleDeleteCheckpoint() error {
	PrintInfo("Удаление чекпоинта")
	// TODO: Реализовать удаление чекпоинта
	return ShowCheckpointManagementMenu()
}
