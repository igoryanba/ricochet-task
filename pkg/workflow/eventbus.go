package workflow

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// EventBus центральная система событий для workflow
type EventBus struct {
	handlers    map[string][]EventHandler
	subscribers map[string][]chan Event
	middleware  []EventMiddleware
	mu          sync.RWMutex
	logger      Logger
	metrics     *EventMetrics
}

// EventMiddleware промежуточное ПО для обработки событий
type EventMiddleware interface {
	Process(ctx context.Context, event Event, next func(Event) error) error
}

// Logger интерфейс логирования
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, err error, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// EventMetrics метрики событий
type EventMetrics struct {
	EventsProcessed   int64     `json:"events_processed"`
	EventsFailed      int64     `json:"events_failed"`
	AverageLatency    float64   `json:"average_latency_ms"`
	LastEventTime     time.Time `json:"last_event_time"`
	HandlerLatencies  map[string]float64 `json:"handler_latencies"`
	mu                sync.RWMutex
}

// NewEventBus создает новый Event Bus
func NewEventBus(logger Logger) *EventBus {
	return &EventBus{
		handlers:    make(map[string][]EventHandler),
		subscribers: make(map[string][]chan Event),
		middleware:  []EventMiddleware{},
		logger:      logger,
		metrics:     &EventMetrics{
			HandlerLatencies: make(map[string]float64),
		},
	}
}

// Subscribe подписывается на события определенного типа
func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	if eb.handlers[eventType] == nil {
		eb.handlers[eventType] = []EventHandler{}
	}
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
	
	eb.logger.Info("Event handler subscribed", "eventType", eventType, "handler", fmt.Sprintf("%T", handler))
}

// SubscribeChannel подписывается на события через канал
func (eb *EventBus) SubscribeChannel(eventType string, ch chan Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	if eb.subscribers[eventType] == nil {
		eb.subscribers[eventType] = []chan Event{}
	}
	eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
	
	eb.logger.Info("Channel subscribed", "eventType", eventType)
}

// Publish публикует событие
func (eb *EventBus) Publish(ctx context.Context, event Event) error {
	startTime := time.Now()
	
	eb.logger.Debug("Publishing event", "type", event.GetType(), "source", event.GetSource())
	
	// Применяем middleware
	err := eb.processWithMiddleware(ctx, event, func(e Event) error {
		return eb.publishToHandlers(ctx, e)
	})
	
	// Обновляем метрики
	eb.updateMetrics(startTime, err)
	
	if err != nil {
		eb.logger.Error("Failed to publish event", err, "type", event.GetType())
		return err
	}
	
	eb.logger.Debug("Event published successfully", "type", event.GetType())
	return nil
}

// PublishAsync публикует событие асинхронно
func (eb *EventBus) PublishAsync(ctx context.Context, event Event) {
	go func() {
		if err := eb.Publish(ctx, event); err != nil {
			eb.logger.Error("Async event publication failed", err, "type", event.GetType())
		}
	}()
}

// AddMiddleware добавляет middleware
func (eb *EventBus) AddMiddleware(middleware EventMiddleware) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.middleware = append(eb.middleware, middleware)
}

// GetMetrics возвращает метрики
func (eb *EventBus) GetMetrics() *EventMetrics {
	eb.metrics.mu.RLock()
	defer eb.metrics.mu.RUnlock()
	
	// Создаем копию для безопасности
	return &EventMetrics{
		EventsProcessed:  eb.metrics.EventsProcessed,
		EventsFailed:     eb.metrics.EventsFailed,
		AverageLatency:   eb.metrics.AverageLatency,
		LastEventTime:    eb.metrics.LastEventTime,
		HandlerLatencies: eb.copyLatencies(),
	}
}

// Close закрывает Event Bus
func (eb *EventBus) Close() error {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	
	// Закрываем все каналы
	for eventType, channels := range eb.subscribers {
		for _, ch := range channels {
			close(ch)
		}
		eb.logger.Info("Closed channels for event type", "eventType", eventType)
	}
	
	eb.logger.Info("Event Bus closed")
	return nil
}

// Внутренние методы

func (eb *EventBus) processWithMiddleware(ctx context.Context, event Event, handler func(Event) error) error {
	if len(eb.middleware) == 0 {
		return handler(event)
	}
	
	// Создаем цепочку middleware
	var processNext func(int, Event) error
	processNext = func(index int, e Event) error {
		if index >= len(eb.middleware) {
			return handler(e)
		}
		
		middleware := eb.middleware[index]
		return middleware.Process(ctx, e, func(nextEvent Event) error {
			return processNext(index+1, nextEvent)
		})
	}
	
	return processNext(0, event)
}

func (eb *EventBus) publishToHandlers(ctx context.Context, event Event) error {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	
	eventType := event.GetType()
	
	// Отправляем обработчикам
	if handlers, exists := eb.handlers[eventType]; exists {
		for _, handler := range handlers {
			if handler.CanHandle(eventType) {
				handlerStart := time.Now()
				err := handler.Handle(ctx, event)
				handlerDuration := time.Since(handlerStart)
				
				// Обновляем метрики для конкретного обработчика
				eb.updateHandlerMetrics(fmt.Sprintf("%T", handler), handlerDuration)
				
				if err != nil {
					eb.logger.Error("Handler failed", err, 
						"handler", fmt.Sprintf("%T", handler),
						"eventType", eventType)
					// Продолжаем обработку остальными обработчиками
				}
			}
		}
	}
	
	// Отправляем в каналы
	if channels, exists := eb.subscribers[eventType]; exists {
		for _, ch := range channels {
			select {
			case ch <- event:
			case <-ctx.Done():
				return ctx.Err()
			default:
				eb.logger.Error("Channel buffer full, skipping event", nil, "eventType", eventType)
			}
		}
	}
	
	return nil
}

func (eb *EventBus) updateMetrics(startTime time.Time, err error) {
	eb.metrics.mu.Lock()
	defer eb.metrics.mu.Unlock()
	
	duration := time.Since(startTime)
	
	if err != nil {
		eb.metrics.EventsFailed++
	} else {
		eb.metrics.EventsProcessed++
	}
	
	// Обновляем среднюю задержку (простая скользящая средняя)
	totalEvents := eb.metrics.EventsProcessed + eb.metrics.EventsFailed
	if totalEvents > 0 {
		eb.metrics.AverageLatency = (eb.metrics.AverageLatency*float64(totalEvents-1) + 
			float64(duration.Nanoseconds())/1000000) / float64(totalEvents)
	}
	
	eb.metrics.LastEventTime = time.Now()
}

func (eb *EventBus) updateHandlerMetrics(handlerName string, duration time.Duration) {
	eb.metrics.mu.Lock()
	defer eb.metrics.mu.Unlock()
	
	eb.metrics.HandlerLatencies[handlerName] = float64(duration.Nanoseconds()) / 1000000
}

func (eb *EventBus) copyLatencies() map[string]float64 {
	copy := make(map[string]float64)
	for k, v := range eb.metrics.HandlerLatencies {
		copy[k] = v
	}
	return copy
}

// Реализация базовых middleware

// LoggingMiddleware логирует все события
type LoggingMiddleware struct {
	logger Logger
}

func NewLoggingMiddleware(logger Logger) *LoggingMiddleware {
	return &LoggingMiddleware{logger: logger}
}

func (m *LoggingMiddleware) Process(ctx context.Context, event Event, next func(Event) error) error {
	m.logger.Info("Processing event",
		"type", event.GetType(),
		"source", event.GetSource(),
		"timestamp", event.GetTimestamp())
	
	err := next(event)
	
	if err != nil {
		m.logger.Error("Event processing failed", err, "type", event.GetType())
	}
	
	return err
}

// RateLimitMiddleware ограничивает скорость обработки событий
type RateLimitMiddleware struct {
	maxEvents int
	window    time.Duration
	events    []time.Time
	mu        sync.Mutex
}

func NewRateLimitMiddleware(maxEvents int, window time.Duration) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		maxEvents: maxEvents,
		window:    window,
		events:    make([]time.Time, 0),
	}
}

func (m *RateLimitMiddleware) Process(ctx context.Context, event Event, next func(Event) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	now := time.Now()
	
	// Очищаем старые события
	cutoff := now.Add(-m.window)
	validEvents := make([]time.Time, 0)
	for _, t := range m.events {
		if t.After(cutoff) {
			validEvents = append(validEvents, t)
		}
	}
	m.events = validEvents
	
	// Проверяем лимит
	if len(m.events) >= m.maxEvents {
		return fmt.Errorf("rate limit exceeded: %d events in %v window", m.maxEvents, m.window)
	}
	
	// Добавляем текущее событие
	m.events = append(m.events, now)
	
	return next(event)
}

// DeduplicationMiddleware предотвращает дублирование событий
type DeduplicationMiddleware struct {
	seen   map[string]time.Time
	ttl    time.Duration
	mu     sync.RWMutex
}

func NewDeduplicationMiddleware(ttl time.Duration) *DeduplicationMiddleware {
	return &DeduplicationMiddleware{
		seen: make(map[string]time.Time),
		ttl:  ttl,
	}
}

func (m *DeduplicationMiddleware) Process(ctx context.Context, event Event, next func(Event) error) error {
	// Создаем ключ на основе типа события и данных
	key := m.createEventKey(event)
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	now := time.Now()
	
	// Очищаем старые записи
	for k, t := range m.seen {
		if now.Sub(t) > m.ttl {
			delete(m.seen, k)
		}
	}
	
	// Проверяем, видели ли мы это событие недавно
	if lastSeen, exists := m.seen[key]; exists {
		if now.Sub(lastSeen) < m.ttl {
			// Дублирующееся событие, пропускаем
			return nil
		}
	}
	
	// Запоминаем событие
	m.seen[key] = now
	
	return next(event)
}

func (m *DeduplicationMiddleware) createEventKey(event Event) string {
	// Простая реализация - можно улучшить
	return fmt.Sprintf("%s:%s:%d", 
		event.GetType(), 
		event.GetSource(), 
		event.GetTimestamp().Unix())
}

// SimpleLogger простая реализация Logger
type SimpleLogger struct{}

func (l *SimpleLogger) Info(msg string, fields ...interface{}) {
	log.Printf("[INFO] %s %v", msg, fields)
}

func (l *SimpleLogger) Error(msg string, err error, fields ...interface{}) {
	log.Printf("[ERROR] %s: %v %v", msg, err, fields)
}

func (l *SimpleLogger) Debug(msg string, fields ...interface{}) {
	log.Printf("[DEBUG] %s %v", msg, fields)
}