package segmentation

// SegmentationMethod определяет метод сегментации
type SegmentationMethod string

const (
	MethodSimple    SegmentationMethod = "simple"    // Простая сегментация по размеру
	MethodSemantic  SegmentationMethod = "semantic"  // Сегментация по смысловым блокам
	MethodRecursive SegmentationMethod = "recursive" // Рекурсивная сегментация
)

// SegmentationOptions опции для сегментации
type SegmentationOptions struct {
	ChunkSize   int                // Максимальный размер чанка в токенах
	MaxSegments int                // Максимальное количество сегментов (0 - без ограничений)
	Method      SegmentationMethod // Метод сегментации
	Overlap     int                // Размер перекрытия в токенах (для MethodSimple и MethodSemantic)
	MinSegSize  int                // Минимальный размер сегмента в токенах
}

// DefaultSegmentationOptions возвращает опции по умолчанию
func DefaultSegmentationOptions() SegmentationOptions {
	return SegmentationOptions{
		ChunkSize:   2000,
		MaxSegments: 0,
		Method:      MethodSimple,
		Overlap:     200,
		MinSegSize:  100,
	}
}

// Segment сегментирует текст и возвращает список сегментов
func Segment(text string, options SegmentationOptions) ([]string, error) {
	// Создаем сегментер
	segmenter, err := NewSegmenter(string(options.Method))
	if err != nil {
		return nil, err
	}

	// Если размер чанка не указан, используем значение по умолчанию
	if options.ChunkSize <= 0 {
		options.ChunkSize = 2000
	}

	// Сегментируем текст
	segments, err := segmenter.Split(text, options.ChunkSize)
	if err != nil {
		return nil, err
	}

	// Ограничиваем количество сегментов, если нужно
	if options.MaxSegments > 0 && len(segments) > options.MaxSegments {
		segments = segments[:options.MaxSegments]
	}

	// Преобразуем SegmentInfo в строки
	result := make([]string, len(segments))
	for i, segment := range segments {
		result[i] = segment.Content
	}

	return result, nil
}
