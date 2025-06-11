package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

var (
	// Цветовые схемы
	InfoColor    = color.New(color.FgCyan).SprintFunc()
	SuccessColor = color.New(color.FgGreen).SprintFunc()
	WarningColor = color.New(color.FgYellow).SprintFunc()
	ErrorColor   = color.New(color.FgRed).SprintFunc()
	BoldColor    = color.New(color.Bold).SprintFunc()

	// Эмодзи для различных сообщений
	InfoEmoji    = "ℹ️ "
	SuccessEmoji = "✅ "
	WarningEmoji = "⚠️ "
	ErrorEmoji   = "❌ "
	BulletEmoji  = "• "
)

// ASCII-арт логотип Ricochet Task
const Logo = `
 ______  ___  _______  _______  _______  ___   _  _______  _______    _______  _______  _______  ___   _  
|    _ ||   ||       ||       ||       ||   | | ||       ||       |  |       ||   _   ||       ||   | | | 
|   | |||   ||       ||       ||   _   ||   |_| ||    ___||_     _|  |_     _||  |_|  ||  _____||   |_| | 
|   |_|||   ||       ||       ||  | |  ||      _||   |___   |   |      |   |  |       || |_____ |      _| 
|    __||   ||      _||      _||  |_|  ||     |_ |    ___|  |   |      |   |  |       ||_____  ||     |_  
|   |   |   ||     |_ |     |_ |       ||    _  ||   |___   |   |      |   |  |   _   | _____| ||    _  | 
|___|   |___||_______||_______||_______||___| |_||_______|  |___|      |___|  |__| |__||_______||___| |_| 
`

// PrintLogo выводит ASCII-арт логотип
func PrintLogo() {
	cyan := color.New(color.FgCyan, color.Bold)
	cyan.Print(Logo)
	fmt.Println()
}

// PrintWelcomeMessage выводит приветственное сообщение
func PrintWelcomeMessage() {
	cyan := color.New(color.FgCyan)
	fmt.Println()
	cyan.Println("Добро пожаловать в Ricochet Task - инструмент для управления задачами и цепочками моделей!")
	fmt.Println("Версия: 1.0.0")
	fmt.Println()
	fmt.Println(InfoColor("Используйте команды ниже для работы с приложением."))
	fmt.Println()
}

// DrawBox выводит текст в рамке
func DrawBox(title string, content string, boxWidth int) {
	// Рассчитываем ширину, если не указана
	if boxWidth < 10 {
		boxWidth = 50
	}

	// Верхняя граница с заголовком
	topBorder := "╔"
	titleWithSpaces := " " + title + " "
	sideLength := (boxWidth - len(titleWithSpaces)) / 2
	topBorder += strings.Repeat("═", sideLength) + titleWithSpaces + strings.Repeat("═", boxWidth-sideLength-len(titleWithSpaces))
	topBorder += "╗"

	// Разбиваем содержимое на строки
	lines := strings.Split(content, "\n")

	// Выводим верхнюю границу
	color.New(color.FgCyan).Println(topBorder)

	// Выводим содержимое
	for _, line := range lines {
		paddedLine := line
		if len(line) < boxWidth {
			paddedLine += strings.Repeat(" ", boxWidth-len(line))
		} else if len(line) > boxWidth {
			paddedLine = line[:boxWidth-3] + "..."
		}
		color.New(color.FgCyan).Print("║ ")
		fmt.Print(paddedLine)
		color.New(color.FgCyan).Println(" ║")
	}

	// Нижняя граница
	bottomBorder := "╚" + strings.Repeat("═", boxWidth) + "╝"
	color.New(color.FgCyan).Println(bottomBorder)
}

// CreateSpinner создает индикатор загрузки
func CreateSpinner(message string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	s.Color("cyan")
	return s
}

// ConfirmPrompt запрашивает подтверждение у пользователя
func ConfirmPrompt(message string) bool {
	response := false
	prompt := &survey.Confirm{
		Message: message,
	}
	survey.AskOne(prompt, &response)
	return response
}

// SelectPrompt запрашивает выбор из списка опций
func SelectPrompt(message string, options []string) (string, error) {
	var selected string
	prompt := &survey.Select{
		Message: message,
		Options: options,
	}
	err := survey.AskOne(prompt, &selected)
	return selected, err
}

// InputPrompt запрашивает текстовый ввод
func InputPrompt(message string) (string, error) {
	var input string
	prompt := &survey.Input{
		Message: message,
	}
	err := survey.AskOne(prompt, &input)
	return input, err
}

// PrintInfo выводит информационное сообщение
func PrintInfo(message string) {
	fmt.Println(InfoEmoji + InfoColor(message))
}

// PrintSuccess выводит сообщение об успехе
func PrintSuccess(message string) {
	fmt.Println(SuccessEmoji + SuccessColor(message))
}

// PrintWarning выводит предупреждение
func PrintWarning(message string) {
	fmt.Println(WarningEmoji + WarningColor(message))
}

// PrintError выводит сообщение об ошибке
func PrintError(message string) {
	fmt.Println(ErrorEmoji + ErrorColor(message))
}
