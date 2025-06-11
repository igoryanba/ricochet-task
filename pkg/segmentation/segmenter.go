package segmentation

import (
	"math"
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

// Константы для типов сегментации
const (
	SegmentationSimple    = "simple"    // По размеру
	SegmentationSemantic  = "semantic"  // По смысловым блокам
	SegmentationRecursive = "recursive" // Рекурсивное разбиение
)

// SegmentInfo содержит информацию о сегменте текста
type SegmentInfo struct {
	ID          string            `json:"id"`          // Уникальный идентификатор сегмента
	Content     string            `json:"content"`     // Содержимое сегмента
	StartPos    int               `json:"start_pos"`   // Начальная позиция в исходном тексте
	EndPos      int               `json:"end_pos"`     // Конечная позиция в исходном тексте
	Order       int               `json:"order"`       // Порядковый номер сегмента
	TokenCount  int               `json:"token_count"` // Примерное количество токенов
	Metadata    map[string]string `json:"metadata"`    // Метаданные сегмента
	Overlapping bool              `json:"overlapping"` // Флаг перекрытия с другими сегментами
}

// Segmenter определяет интерфейс для сегментации текста
type Segmenter interface {
	// Split разбивает текст на сегменты
	Split(text string, maxTokensPerSegment int) ([]SegmentInfo, error)

	// Merge объединяет сегменты в один текст
	Merge(segments []SegmentInfo) (string, error)

	// GetType возвращает тип сегментера
	GetType() string
}

// NewSegmenter создает новый сегментер указанного типа
func NewSegmenter(segmentationType string) (Segmenter, error) {
	switch segmentationType {
	case SegmentationSimple:
		return &SimpleSegmenter{}, nil
	case SegmentationSemantic:
		return &SemanticSegmenter{}, nil
	case SegmentationRecursive:
		return &RecursiveSegmenter{}, nil
	default:
		return &SimpleSegmenter{}, nil // По умолчанию используем простую сегментацию
	}
}

// SimpleSegmenter реализует простую сегментацию по размеру
type SimpleSegmenter struct{}

// Split разбивает текст на сегменты по размеру
func (s *SimpleSegmenter) Split(text string, maxTokensPerSegment int) ([]SegmentInfo, error) {
	if text == "" {
		return []SegmentInfo{}, nil
	}

	// Примерная оценка: 1 токен ~ 4 символа
	maxCharsPerSegment := maxTokensPerSegment * 4

	// Если текст меньше максимального размера, возвращаем его целиком
	textLength := len(text)
	if textLength <= maxCharsPerSegment {
		return []SegmentInfo{
			{
				ID:         uuid.New().String(),
				Content:    text,
				StartPos:   0,
				EndPos:     textLength,
				Order:      0,
				TokenCount: estimateTokenCount(text),
				Metadata:   map[string]string{},
			},
		}, nil
	}

	// Разбиваем текст на сегменты
	var segments []SegmentInfo

	// Находим границы предложений
	sentenceBoundaries := findSentenceBoundaries(text)

	// Начальная позиция текущего сегмента
	startPos := 0
	segmentOrder := 0

	for startPos < textLength {
		// Определяем конечную позицию сегмента
		endPos := startPos + maxCharsPerSegment
		if endPos > textLength {
			endPos = textLength
		} else {
			// Ищем ближайшую границу предложения к endPos
			endPos = findNearestSentenceBoundary(sentenceBoundaries, endPos)
		}

		// Создаем сегмент
		segment := SegmentInfo{
			ID:         uuid.New().String(),
			Content:    text[startPos:endPos],
			StartPos:   startPos,
			EndPos:     endPos,
			Order:      segmentOrder,
			TokenCount: estimateTokenCount(text[startPos:endPos]),
			Metadata:   map[string]string{},
		}

		segments = append(segments, segment)

		// Переходим к следующему сегменту
		startPos = endPos
		segmentOrder++
	}

	return segments, nil
}

// Merge объединяет сегменты в один текст
func (s *SimpleSegmenter) Merge(segments []SegmentInfo) (string, error) {
	if len(segments) == 0 {
		return "", nil
	}

	// Сортируем сегменты по порядку
	sortedSegments := make([]SegmentInfo, len(segments))
	copy(sortedSegments, segments)

	// Сортировка пузырьком для простоты
	for i := 0; i < len(sortedSegments)-1; i++ {
		for j := 0; j < len(sortedSegments)-i-1; j++ {
			if sortedSegments[j].Order > sortedSegments[j+1].Order {
				sortedSegments[j], sortedSegments[j+1] = sortedSegments[j+1], sortedSegments[j]
			}
		}
	}

	// Объединяем сегменты
	var result strings.Builder
	for _, segment := range sortedSegments {
		result.WriteString(segment.Content)
	}

	return result.String(), nil
}

// GetType возвращает тип сегментера
func (s *SimpleSegmenter) GetType() string {
	return SegmentationSimple
}

// SemanticSegmenter реализует сегментацию по смысловым блокам
type SemanticSegmenter struct{}

// Split разбивает текст на сегменты по смысловым блокам
func (s *SemanticSegmenter) Split(text string, maxTokensPerSegment int) ([]SegmentInfo, error) {
	if text == "" {
		return []SegmentInfo{}, nil
	}

	// Примерная оценка: 1 токен ~ 4 символа
	maxCharsPerSegment := maxTokensPerSegment * 4

	// Если текст меньше максимального размера, возвращаем его целиком
	textLength := len(text)
	if textLength <= maxCharsPerSegment {
		return []SegmentInfo{
			{
				ID:         uuid.New().String(),
				Content:    text,
				StartPos:   0,
				EndPos:     textLength,
				Order:      0,
				TokenCount: estimateTokenCount(text),
				Metadata:   map[string]string{},
			},
		}, nil
	}

	// Находим параграфы и заголовки
	paragraphs := findParagraphs(text)

	// Группируем параграфы в сегменты
	var segments []SegmentInfo
	var currentSegment strings.Builder
	currentTokenCount := 0
	startPos := 0
	segmentOrder := 0

	for _, p := range paragraphs {
		paragraphTokens := estimateTokenCount(p.Content)

		// Если параграф сам по себе больше максимального размера,
		// разбиваем его на предложения
		if paragraphTokens > maxTokensPerSegment {
			// Если текущий сегмент не пуст, сохраняем его
			if currentSegment.Len() > 0 {
				segment := SegmentInfo{
					ID:         uuid.New().String(),
					Content:    currentSegment.String(),
					StartPos:   startPos,
					EndPos:     p.StartPos,
					Order:      segmentOrder,
					TokenCount: currentTokenCount,
					Metadata:   map[string]string{},
				}
				segments = append(segments, segment)
				segmentOrder++
			}

			// Разбиваем большой параграф на сегменты
			paragraphSegments, err := s.splitParagraph(p.Content, p.StartPos, maxTokensPerSegment, segmentOrder)
			if err != nil {
				return nil, err
			}

			segments = append(segments, paragraphSegments...)
			segmentOrder += len(paragraphSegments)
			startPos = p.EndPos
			currentSegment.Reset()
			currentTokenCount = 0
		} else if currentTokenCount+paragraphTokens > maxTokensPerSegment {
			// Если текущий сегмент + параграф превышают максимальный размер,
			// сохраняем текущий сегмент и начинаем новый
			segment := SegmentInfo{
				ID:         uuid.New().String(),
				Content:    currentSegment.String(),
				StartPos:   startPos,
				EndPos:     p.StartPos,
				Order:      segmentOrder,
				TokenCount: currentTokenCount,
				Metadata:   map[string]string{},
			}
			segments = append(segments, segment)
			segmentOrder++

			// Начинаем новый сегмент с текущего параграфа
			currentSegment.Reset()
			currentSegment.WriteString(p.Content)
			currentTokenCount = paragraphTokens
			startPos = p.StartPos
		} else {
			// Добавляем параграф к текущему сегменту
			currentSegment.WriteString(p.Content)
			currentTokenCount += paragraphTokens
		}
	}

	// Добавляем последний сегмент, если он не пустой
	if currentSegment.Len() > 0 {
		segment := SegmentInfo{
			ID:         uuid.New().String(),
			Content:    currentSegment.String(),
			StartPos:   startPos,
			EndPos:     textLength,
			Order:      segmentOrder,
			TokenCount: currentTokenCount,
			Metadata:   map[string]string{},
		}
		segments = append(segments, segment)
	}

	return segments, nil
}

// splitParagraph разбивает большой параграф на сегменты
func (s *SemanticSegmenter) splitParagraph(paragraph string, startOffset int, maxTokensPerSegment int, startOrder int) ([]SegmentInfo, error) {
	// Используем SimpleSegmenter для разбиения параграфа на части
	simpleSegmenter := &SimpleSegmenter{}
	segments, err := simpleSegmenter.Split(paragraph, maxTokensPerSegment)
	if err != nil {
		return nil, err
	}

	// Корректируем позиции и порядок
	for i := range segments {
		segments[i].StartPos += startOffset
		segments[i].EndPos += startOffset
		segments[i].Order += startOrder
	}

	return segments, nil
}

// Merge объединяет сегменты в один текст
func (s *SemanticSegmenter) Merge(segments []SegmentInfo) (string, error) {
	// Используем тот же алгоритм, что и в SimpleSegmenter
	simpleSegmenter := &SimpleSegmenter{}
	return simpleSegmenter.Merge(segments)
}

// GetType возвращает тип сегментера
func (s *SemanticSegmenter) GetType() string {
	return SegmentationSemantic
}

// RecursiveSegmenter реализует рекурсивную сегментацию
type RecursiveSegmenter struct{}

// Split разбивает текст на сегменты рекурсивно
func (r *RecursiveSegmenter) Split(text string, maxTokensPerSegment int) ([]SegmentInfo, error) {
	if text == "" {
		return []SegmentInfo{}, nil
	}

	// Примерная оценка: 1 токен ~ 4 символа
	maxCharsPerSegment := maxTokensPerSegment * 4

	// Если текст меньше максимального размера, возвращаем его целиком
	textLength := len(text)
	if textLength <= maxCharsPerSegment {
		return []SegmentInfo{
			{
				ID:         uuid.New().String(),
				Content:    text,
				StartPos:   0,
				EndPos:     textLength,
				Order:      0,
				TokenCount: estimateTokenCount(text),
				Metadata:   map[string]string{},
			},
		}, nil
	}

	// Ищем разделы верхнего уровня (главы, разделы)
	sections := findSections(text)

	// Если нет явных разделов, используем семантическую сегментацию
	if len(sections) <= 1 {
		semanticSegmenter := &SemanticSegmenter{}
		return semanticSegmenter.Split(text, maxTokensPerSegment)
	}

	// Рекурсивно разбиваем каждый раздел
	var allSegments []SegmentInfo
	segmentOrder := 0

	for _, section := range sections {
		// Если раздел меньше максимального размера, сохраняем его целиком
		sectionTokens := estimateTokenCount(section.Content)
		if sectionTokens <= maxTokensPerSegment {
			segment := SegmentInfo{
				ID:         uuid.New().String(),
				Content:    section.Content,
				StartPos:   section.StartPos,
				EndPos:     section.EndPos,
				Order:      segmentOrder,
				TokenCount: sectionTokens,
				Metadata: map[string]string{
					"section": section.Title,
				},
			}
			allSegments = append(allSegments, segment)
			segmentOrder++
		} else {
			// Рекурсивно разбиваем раздел
			sectionSegments, err := r.Split(section.Content, maxTokensPerSegment)
			if err != nil {
				return nil, err
			}

			// Корректируем позиции и порядок
			for i := range sectionSegments {
				sectionSegments[i].StartPos += section.StartPos
				sectionSegments[i].EndPos += section.StartPos
				sectionSegments[i].Order += segmentOrder
				if sectionSegments[i].Metadata == nil {
					sectionSegments[i].Metadata = map[string]string{}
				}
				sectionSegments[i].Metadata["section"] = section.Title
			}

			allSegments = append(allSegments, sectionSegments...)
			segmentOrder += len(sectionSegments)
		}
	}

	return allSegments, nil
}

// Merge объединяет сегменты в один текст
func (r *RecursiveSegmenter) Merge(segments []SegmentInfo) (string, error) {
	// Используем тот же алгоритм, что и в SimpleSegmenter
	simpleSegmenter := &SimpleSegmenter{}
	return simpleSegmenter.Merge(segments)
}

// GetType возвращает тип сегментера
func (r *RecursiveSegmenter) GetType() string {
	return SegmentationRecursive
}

// Section представляет раздел текста
type Section struct {
	Title    string
	Content  string
	StartPos int
	EndPos   int
}

// Paragraph представляет параграф текста
type Paragraph struct {
	Content  string
	StartPos int
	EndPos   int
	IsHeader bool
}

// findSections находит разделы в тексте
func findSections(text string) []Section {
	// Регулярное выражение для поиска заголовков
	// Ищем строки вида "# Заголовок" или "## Подзаголовок"
	headerRegex := regexp.MustCompile(`(?m)^(#{1,3})\s+(.+)$`)

	matches := headerRegex.FindAllStringSubmatchIndex(text, -1)

	if len(matches) == 0 {
		// Нет заголовков, весь текст как один раздел
		return []Section{
			{
				Title:    "",
				Content:  text,
				StartPos: 0,
				EndPos:   len(text),
			},
		}
	}

	var sections []Section

	// Обрабатываем каждый заголовок
	for i, match := range matches {
		// Извлекаем информацию о заголовке
		headerEnd := match[1]
		headerLevel := len(text[match[2]:match[3]]) // Количество символов #
		headerTitle := text[match[4]:match[5]]

		// Определяем начало содержимого раздела
		contentStart := headerEnd

		// Определяем конец содержимого раздела
		contentEnd := len(text)
		if i < len(matches)-1 {
			// Если это не последний заголовок, конец раздела - начало следующего заголовка
			nextHeaderLevel := len(text[matches[i+1][2]:matches[i+1][3]])

			// Конец раздела - это начало следующего заголовка того же или более высокого уровня
			if nextHeaderLevel <= headerLevel {
				contentEnd = matches[i+1][0]
			} else {
				// Иначе ищем следующий заголовок того же или более высокого уровня
				for j := i + 1; j < len(matches); j++ {
					nextLevel := len(text[matches[j][2]:matches[j][3]])
					if nextLevel <= headerLevel {
						contentEnd = matches[j][0]
						break
					}
				}
			}
		}

		// Создаем раздел
		section := Section{
			Title:    headerTitle,
			Content:  text[contentStart:contentEnd],
			StartPos: contentStart,
			EndPos:   contentEnd,
		}

		sections = append(sections, section)
	}

	return sections
}

// findParagraphs находит параграфы в тексте
func findParagraphs(text string) []Paragraph {
	// Разбиваем текст на строки
	lines := strings.Split(text, "\n")

	var paragraphs []Paragraph
	var currentParagraph strings.Builder

	startPos := 0
	isHeader := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Проверяем, является ли строка заголовком
		if strings.HasPrefix(trimmedLine, "#") {
			// Если есть текущий параграф, сохраняем его
			if currentParagraph.Len() > 0 {
				paragraphContent := currentParagraph.String()
				paragraphs = append(paragraphs, Paragraph{
					Content:  paragraphContent,
					StartPos: startPos,
					EndPos:   startPos + len(paragraphContent),
					IsHeader: isHeader,
				})
				currentParagraph.Reset()
			}

			// Начинаем новый параграф (заголовок)
			currentParagraph.WriteString(line + "\n")
			startPos = len(strings.Join(lines[:len(paragraphs)], "\n"))
			isHeader = true
		} else if trimmedLine == "" {
			// Пустая строка - конец параграфа
			if currentParagraph.Len() > 0 {
				paragraphContent := currentParagraph.String()
				paragraphs = append(paragraphs, Paragraph{
					Content:  paragraphContent,
					StartPos: startPos,
					EndPos:   startPos + len(paragraphContent),
					IsHeader: isHeader,
				})
				currentParagraph.Reset()
				isHeader = false
			}

			// Обновляем начальную позицию следующего параграфа
			startPos = len(strings.Join(lines[:len(paragraphs)], "\n")) + 1
		} else {
			// Добавляем строку к текущему параграфу
			if currentParagraph.Len() > 0 {
				currentParagraph.WriteString("\n")
			}
			currentParagraph.WriteString(line)
		}
	}

	// Добавляем последний параграф, если он не пустой
	if currentParagraph.Len() > 0 {
		paragraphContent := currentParagraph.String()
		paragraphs = append(paragraphs, Paragraph{
			Content:  paragraphContent,
			StartPos: startPos,
			EndPos:   startPos + len(paragraphContent),
			IsHeader: isHeader,
		})
	}

	return paragraphs
}

// findSentenceBoundaries находит границы предложений в тексте
func findSentenceBoundaries(text string) []int {
	// Простой алгоритм для поиска границ предложений
	// Ищем точки, за которыми следует пробел или новая строка
	var boundaries []int

	// Добавляем начало текста как границу
	boundaries = append(boundaries, 0)

	for i := 0; i < len(text)-1; i++ {
		if (text[i] == '.' || text[i] == '!' || text[i] == '?') &&
			(i+1 < len(text) && (unicode.IsSpace(rune(text[i+1])) || text[i+1] == '\n')) {
			boundaries = append(boundaries, i+1)
		}
	}

	// Добавляем конец текста как границу
	boundaries = append(boundaries, len(text))

	return boundaries
}

// findNearestSentenceBoundary находит ближайшую границу предложения к указанной позиции
func findNearestSentenceBoundary(boundaries []int, position int) int {
	// Если позиция выходит за пределы текста, возвращаем последнюю границу
	if position >= boundaries[len(boundaries)-1] {
		return boundaries[len(boundaries)-1]
	}

	// Ищем ближайшую границу
	minDist := math.MaxInt32
	nearestBoundary := position

	for _, boundary := range boundaries {
		dist := abs(boundary - position)
		if dist < minDist {
			minDist = dist
			nearestBoundary = boundary
		}
	}

	return nearestBoundary
}

// abs возвращает абсолютное значение числа
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// estimateTokenCount оценивает количество токенов в тексте
func estimateTokenCount(text string) int {
	// Примерная оценка: 1 токен ~ 4 символа
	return int(float64(len(text)) * 0.25)
}
