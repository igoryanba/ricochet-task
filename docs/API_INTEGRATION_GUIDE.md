# Руководство по API и интеграции Ricochet

## Оглавление

1. [Введение](#введение)
2. [Архитектура API](#архитектура-api)
3. [MCP (Model Chain Protocol)](#mcp-model-chain-protocol)
4. [Интеграция с внешними сервисами](#интеграция-с-внешними-сервисами)
5. [Интеграция с Task Master](#интеграция-с-task-master)
6. [Примеры использования API](#примеры-использования-api)

## Введение

Ricochet предоставляет мощный API для управления цепочками моделей искусственного интеллекта. Это руководство описывает архитектуру API, доступные эндпоинты и способы интеграции с другими сервисами.

## Архитектура API

Ricochet использует RESTful API с JSON в качестве формата обмена данными. API организован вокруг нескольких ключевых ресурсов:

- **Chains (Цепочки)**: Последовательности моделей для обработки данных
- **Models (Модели)**: Отдельные модели искусственного интеллекта
- **Checkpoints (Контрольные точки)**: Промежуточные результаты выполнения цепочек
- **Tasks (Задачи)**: Отдельные единицы работы, выполняемые в рамках цепочки

### Базовый URL

Все API-запросы выполняются по URL:

```
http://localhost:<port>/api/v1/
```

где `<port>` - порт, на котором запущен Ricochet API-сервер.

### Аутентификация

API использует аутентификацию по API-ключу. Ключ должен быть передан в заголовке `Authorization`:

```
Authorization: Bearer <your-api-key>
```

## MCP (Model Chain Protocol)

MCP - это протокол для взаимодействия с цепочками моделей. Он позволяет:

- Создавать и управлять цепочками моделей
- Запускать цепочки с различными входными данными
- Мониторить выполнение цепочек
- Получать результаты и визуализации

### Основные команды MCP

| Команда | Описание |
|---------|----------|
| `chain_create` | Создание новой цепочки моделей |
| `chain_builder_init` | Инициализация сессии построителя цепочек |
| `chain_builder_add_step` | Добавление шага в построитель цепочек |
| `chain_builder_edit_step` | Редактирование шага в построителе цепочек |
| `chain_builder_remove_step` | Удаление шага из построителя цепочек |
| `chain_builder_get_session` | Получение данных сессии построителя |
| `chain_builder_complete` | Завершение сессии построителя цепочек |
| `chain_pause` | Пауза выполнения цепочки |
| `chain_resume` | Возобновление выполнения цепочки |
| `chain_stop` | Остановка выполнения цепочки |
| `chain_step_control` | Управление переходами между шагами |
| `chain_results` | Получение результатов выполнения цепочки |
| `chain_run_result` | Получение результата конкретного запуска |
| `chain_monitor` | Мониторинг выполнения цепочки |
| `chain_visualization` | Получение визуализации цепочки |
| `chain_progress` | Получение прогресса выполнения цепочки |

### Пример запроса MCP

```json
{
  "command": "chain_create",
  "params": {
    "name": "Анализ документа",
    "description": "Цепочка для анализа текстовых документов",
    "steps": [
      {
        "role_id": "analyzer",
        "model_id": "gpt-4",
        "provider": "openai",
        "name": "Анализ структуры",
        "description": "Анализ структуры документа",
        "prompt": "Проанализируйте структуру документа и выделите ключевые разделы."
      },
      {
        "role_id": "summarizer",
        "model_id": "claude-3-opus",
        "provider": "anthropic",
        "name": "Суммаризация",
        "description": "Суммаризация документа",
        "prompt": "Создайте краткое резюме документа."
      }
    ]
  }
}
```

## Интеграция с внешними сервисами

Ricochet может интегрироваться с различными внешними сервисами:

### Интеграция с провайдерами моделей

Ricochet поддерживает различных провайдеров моделей искусственного интеллекта:

- OpenAI (GPT-4, GPT-3.5 и др.)
- Anthropic (Claude)
- DeepSeek (DeepSeek Coder)
- Mistral AI
- Xeni AI (Grok)

Для каждого провайдера необходимо указать API-ключ в конфигурации или через переменные окружения.

### Интеграция с редакторами кода

Ricochet предоставляет интеграцию с популярными редакторами кода:

- VS Code
- Cursor
- JetBrains IDEs

Подробнее об интеграции с редакторами см. в разделе [Руководство по интеграции с редакторами кода](./EDITOR_INTEGRATION_GUIDE.md).

## Интеграция с Task Master

Ricochet тесно интегрируется с Task Master для управления задачами и рабочими процессами.

### Основные команды интеграции

| Команда | Описание |
|---------|----------|
| `task_master_export` | Экспорт моделей в Task Master |
| `task_master_import` | Импорт моделей из Task Master |

### Пример интеграции с Task Master

```go
// Создание задачи в Task Master на основе цепочки Ricochet
chainID := "chain-123"
taskID, err := ricochetTaskConverter.CreateTaskFromChain(chainID)
if err != nil {
    log.Fatalf("Failed to create task: %v", err)
}
fmt.Printf("Created task: %s\n", taskID)

// Создание цепочки Ricochet на основе задачи Task Master
taskID := "task-456"
chainID, err := ricochetTaskConverter.CreateChainFromTask(taskID)
if err != nil {
    log.Fatalf("Failed to create chain: %v", err)
}
fmt.Printf("Created chain: %s\n", chainID)
```

## Примеры использования API

### Создание цепочки моделей

```bash
curl -X POST http://localhost:8080/api/v1/chains \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Анализ кода",
    "description": "Цепочка для анализа кода",
    "steps": [
      {
        "role_id": "analyzer",
        "model_id": "deepseek-coder",
        "provider": "deepseek",
        "name": "Анализ кода",
        "prompt": "Проанализируйте следующий код и выделите основные компоненты и их взаимодействие."
      },
      {
        "role_id": "reviewer",
        "model_id": "gpt-4",
        "provider": "openai",
        "name": "Код-ревью",
        "prompt": "Проведите код-ревью, отметьте проблемы и предложите улучшения."
      }
    ]
  }'
```

### Запуск цепочки с текстовыми данными

```bash
curl -X POST http://localhost:8080/api/v1/chains/chain-123/run \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Ваш текст для обработки",
    "options": {
      "max_parallel_chunks": 4,
      "max_tokens_per_chunk": 2000,
      "segmentation_method": "semantic",
      "save_checkpoints": true
    }
  }'
```

### Получение результатов выполнения

```bash
curl -X GET http://localhost:8080/api/v1/runs/run-456/results \
  -H "Authorization: Bearer your-api-key"
```

---

Данное руководство предоставляет основную информацию по API и интеграции Ricochet. Для получения дополнительной информации обратитесь к исходному коду и документации. 