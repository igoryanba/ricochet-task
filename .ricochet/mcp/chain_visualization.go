package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// VisualizationParams параметры для визуализации цепочки
type VisualizationParams struct {
	ChainID      string `json:"chain_id"`
	Format       string `json:"format,omitempty"` // ascii, unicode, mermaid
	ShowProgress bool   `json:"show_progress,omitempty"`
	ShowTasks    bool   `json:"show_tasks,omitempty"`
	ShowMetrics  bool   `json:"show_metrics,omitempty"`
	Compact      bool   `json:"compact,omitempty"`
}

// VisualizationResponse ответ с визуализацией цепочки
type VisualizationResponse struct {
	ChainID       string    `json:"chain_id"`
	ChainName     string    `json:"chain_name"`
	Visualization string    `json:"visualization"`
	Format        string    `json:"format"`
	GeneratedAt   time.Time `json:"generated_at"`
}

// ModelVisualizationData данные для визуализации модели
type ModelVisualizationData struct {
	ID         string
	Name       string
	Provider   string
	Role       string
	Progress   float64
	Status     string
	TasksTotal int
	TasksDone  int
	TokensUsed int
	CostValue  float64
	ErrorMsg   string
}

// ChainVisualizationData данные для визуализации цепочки
type ChainVisualizationData struct {
	ID            string
	Name          string
	Models        []ModelVisualizationData
	TotalProgress float64
	Status        string
	StartedAt     time.Time
	ElapsedTime   string
	TotalTokens   int
	TotalCost     float64
}

// HandleChainVisualization обрабатывает запрос на визуализацию цепочки
func HandleChainVisualization(params json.RawMessage) (interface{}, error) {
	var vizParams VisualizationParams
	if err := json.Unmarshal(params, &vizParams); err != nil {
		return nil, fmt.Errorf("неверные параметры для визуализации: %v", err)
	}

	if vizParams.ChainID == "" {
		return nil, fmt.Errorf("chain_id является обязательным параметром")
	}

	// Установить значения по умолчанию
	if vizParams.Format == "" {
		vizParams.Format = "unicode"
	}

	// Получить данные цепочки
	chainData, err := getChainVisualizationData(vizParams.ChainID)
	if err != nil {
		return nil, err
	}

	// Сгенерировать визуализацию в соответствии с форматом
	var visualization string
	switch vizParams.Format {
	case "ascii":
		visualization = generateASCIIVisualization(chainData, vizParams)
	case "unicode":
		visualization = generateUnicodeVisualization(chainData, vizParams)
	case "mermaid":
		visualization = generateMermaidVisualization(chainData, vizParams)
	default:
		return nil, fmt.Errorf("неподдерживаемый формат визуализации: %s", vizParams.Format)
	}

	response := VisualizationResponse{
		ChainID:       vizParams.ChainID,
		ChainName:     chainData.Name,
		Visualization: visualization,
		Format:        vizParams.Format,
		GeneratedAt:   time.Now(),
	}

	return response, nil
}

// RegisterChainVisualizationCommand регистрирует команду визуализации цепочки
func RegisterChainVisualizationCommand(server *MCPServer) {
	server.RegisterCommand("chain_visualization", HandleChainVisualization)
}

// Вспомогательные функции

// getChainVisualizationData получает данные цепочки для визуализации
func getChainVisualizationData(chainID string) (ChainVisualizationData, error) {
	// TODO: Получить данные о цепочке и ее текущем статусе из оркестратора
	// Временная реализация с тестовыми данными
	models := []ModelVisualizationData{
		{
			ID:         "model-1",
			Name:       "GPT-4",
			Provider:   "OpenAI",
			Role:       "Анализатор",
			Progress:   0.85,
			Status:     "running",
			TasksTotal: 3,
			TasksDone:  2,
			TokensUsed: 5420,
			CostValue:  0.123,
		},
		{
			ID:         "model-2",
			Name:       "Claude-3",
			Provider:   "Anthropic",
			Role:       "Суммаризатор",
			Progress:   0.0,
			Status:     "pending",
			TasksTotal: 3,
			TasksDone:  0,
			TokensUsed: 0,
			CostValue:  0.0,
		},
		{
			ID:         "model-3",
			Name:       "DeepSeek",
			Provider:   "DeepSeek",
			Role:       "Интегратор",
			Progress:   0.0,
			Status:     "pending",
			TasksTotal: 2,
			TasksDone:  0,
			TokensUsed: 0,
			CostValue:  0.0,
		},
	}

	chainData := ChainVisualizationData{
		ID:            chainID,
		Name:          "Анализ документации",
		Models:        models,
		TotalProgress: 0.28,
		Status:        "running",
		StartedAt:     time.Now().Add(-15 * time.Minute),
		ElapsedTime:   "15м 23с",
		TotalTokens:   5420,
		TotalCost:     0.123,
	}

	return chainData, nil
}

// generateASCIIVisualization генерирует ASCII-визуализацию цепочки
func generateASCIIVisualization(data ChainVisualizationData, params VisualizationParams) string {
	var result strings.Builder

	// Заголовок
	result.WriteString(fmt.Sprintf("Chain: %s (ID: %s)\n", data.Name, data.ID))
	result.WriteString(fmt.Sprintf("Status: %s | Progress: %.1f%%\n", data.Status, data.TotalProgress*100))

	if params.ShowMetrics {
		result.WriteString(fmt.Sprintf("Elapsed: %s | Tokens: %d | Cost: $%.3f\n",
			data.ElapsedTime, data.TotalTokens, data.TotalCost))
	}

	result.WriteString("\n")

	// Модели
	for i := range data.Models {
		// Создаем строку модели
		modelBox := fmt.Sprintf("+---------------+\n")
		modelBox += fmt.Sprintf("| %-13s |\n", data.Models[i].Role)
		modelBox += fmt.Sprintf("| (%s)       |\n", data.Models[i].Provider)

		// Прогресс-бар
		if params.ShowProgress {
			progressStr := ""
			progressLength := 13
			filledChars := int(data.Models[i].Progress * float64(progressLength))

			for j := 0; j < progressLength; j++ {
				if j < filledChars {
					progressStr += "#"
				} else {
					progressStr += "-"
				}
			}

			modelBox += fmt.Sprintf("| [%s] |\n", progressStr)
			modelBox += fmt.Sprintf("| %5.1f%%        |\n", data.Models[i].Progress*100)
		}

		modelBox += fmt.Sprintf("+---------------+")

		// Для всех моделей, кроме последней, добавляем стрелку
		if i < len(data.Models)-1 {
			modelBox += fmt.Sprintf("---> ")
		}

		result.WriteString(modelBox)
	}

	// Детали задач
	if params.ShowTasks {
		result.WriteString("\n\nTasks:\n")
		for _, model := range data.Models {
			result.WriteString(fmt.Sprintf("%s: %d/%d completed\n",
				model.Role, model.TasksDone, model.TasksTotal))
		}
	}

	return result.String()
}

// generateUnicodeVisualization генерирует Unicode-визуализацию цепочки
func generateUnicodeVisualization(data ChainVisualizationData, params VisualizationParams) string {
	var result strings.Builder

	// Заголовок
	result.WriteString(fmt.Sprintf("Цепочка: %s (ID: %s)\n", data.Name, data.ID))
	result.WriteString(fmt.Sprintf("Статус: %s | Прогресс: %.1f%%\n", translateStatus(data.Status), data.TotalProgress*100))

	if params.ShowMetrics {
		result.WriteString(fmt.Sprintf("Прошло: %s | Токены: %d | Стоимость: $%.3f\n",
			data.ElapsedTime, data.TotalTokens, data.TotalCost))
	}

	result.WriteString("\n")

	// Формируем строки для моделей
	topLine := "┌─────────────┐"
	modelLine := "│  %s │"
	providerLine := "│   (%s)   │"
	progressLine := "│  [%s]     │"
	percentLine := "│    %.1f%%     │"
	bottomLine := "└─────────────┘"
	connector := "    ───>    "

	for i := range data.Models {
		if i > 0 {
			topLine += connector + "┌─────────────┐"
			modelLine += connector + "│  %s │"
			providerLine += connector + "│   (%s)   │"
			progressLine += connector + "│  [%s]     │"
			percentLine += connector + "│    %.1f%%     │"
			bottomLine += connector + "└─────────────┘"
		}
	}

	// Выводим верхнюю линию
	result.WriteString(topLine + "\n")

	// Выводим строку с ролями моделей
	roleArgs := make([]interface{}, len(data.Models))
	for i := range data.Models {
		roleArgs[i] = fitText(data.Models[i].Role, 9)
	}
	result.WriteString(fmt.Sprintf(modelLine+"\n", roleArgs...))

	// Выводим строку с провайдерами
	providerArgs := make([]interface{}, len(data.Models))
	for i := range data.Models {
		providerArgs[i] = fitText(data.Models[i].Provider, 7)
	}
	result.WriteString(fmt.Sprintf(providerLine+"\n", providerArgs...))

	// Выводим прогресс-бары, если нужно
	if params.ShowProgress {
		progressArgs := make([]interface{}, len(data.Models))
		for i := range data.Models {
			progressBar := ""
			progressLength := 8
			filledChars := int(data.Models[i].Progress * float64(progressLength))

			for j := 0; j < progressLength; j++ {
				if j < filledChars {
					progressBar += "█"
				} else {
					progressBar += "─"
				}
			}
			progressArgs[i] = progressBar
		}
		result.WriteString(fmt.Sprintf(progressLine+"\n", progressArgs...))

		// Выводим проценты
		percentArgs := make([]interface{}, len(data.Models))
		for i := range data.Models {
			percentArgs[i] = data.Models[i].Progress * 100
		}
		result.WriteString(fmt.Sprintf(percentLine+"\n", percentArgs...))
	}

	// Выводим нижнюю линию
	result.WriteString(bottomLine + "\n")

	// Детали задач
	if params.ShowTasks {
		result.WriteString("\n\nЗадачи:\n")
		for _, model := range data.Models {
			statusIcon := "⏳"
			if model.Status == "completed" {
				statusIcon = "✅"
			} else if model.Status == "error" {
				statusIcon = "❌"
			} else if model.Status == "pending" {
				statusIcon = "⏱️"
			}

			result.WriteString(fmt.Sprintf("%s %s: %d/%d выполнено\n",
				statusIcon, model.Role, model.TasksDone, model.TasksTotal))
		}
	}

	return result.String()
}

// generateMermaidVisualization генерирует Mermaid-визуализацию цепочки
func generateMermaidVisualization(data ChainVisualizationData, params VisualizationParams) string {
	var result strings.Builder

	// Начало диаграммы
	result.WriteString("```mermaid\nflowchart LR\n")

	// Узлы для моделей
	for i := range data.Models {
		nodeId := fmt.Sprintf("model%d", i+1)

		// Определяем стиль узла на основе статуса
		style := ""
		switch data.Models[i].Status {
		case "running":
			style = "fill:#b3e0ff,stroke:#4da6ff"
		case "completed":
			style = "fill:#c6ffb3,stroke:#7acc29"
		case "error":
			style = "fill:#ffb3b3,stroke:#ff4d4d"
		case "pending":
			style = "fill:#f2f2f2,stroke:#bfbfbf"
		}

		// Формируем содержимое узла
		content := fmt.Sprintf("%s<br/>(%s)", data.Models[i].Role, data.Models[i].Provider)
		if params.ShowProgress {
			content += fmt.Sprintf("<br/>%.1f%%", data.Models[i].Progress*100)
		}

		// Добавляем узел
		result.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", nodeId, content))

		// Добавляем стиль
		result.WriteString(fmt.Sprintf("    style %s %s\n", nodeId, style))
	}

	// Связи между моделями
	for i := 0; i < len(data.Models)-1; i++ {
		result.WriteString(fmt.Sprintf("    model%d --> model%d\n", i+1, i+2))
	}

	// Заголовок диаграммы
	result.WriteString(fmt.Sprintf("    title[\"%s - %.1f%% выполнено\"]\n", data.Name, data.TotalProgress*100))
	result.WriteString("    style title fill:none,stroke:none\n")

	// Конец диаграммы
	result.WriteString("```")

	return result.String()
}

// Вспомогательные функции

// fitText подгоняет текст под заданную длину
func fitText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	if maxLength <= 3 {
		return text[:maxLength]
	}
	return text[:maxLength-3] + "..."
}

// translateStatus переводит статус на русский язык
func translateStatus(status string) string {
	switch status {
	case "running":
		return "выполняется"
	case "completed":
		return "завершено"
	case "error":
		return "ошибка"
	case "pending":
		return "ожидание"
	default:
		return status
	}
}
