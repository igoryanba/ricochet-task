# Интеграция Ricochet Task с MCP

Этот модуль предоставляет интеграцию с протоколом MCP (Model Control Protocol) для взаимодействия с редакторами кода, такими как Cursor и VS Code.

## Обзор

MCP (Model Control Protocol) - это протокол, который позволяет редакторам кода взаимодействовать с внешними сервисами через специальный интерфейс. Ricochet Task использует MCP для предоставления возможностей оркестрации моделей непосредственно из редактора кода.

## Возможности

### Основные функции

- **Управление цепочками моделей** - создание, запуск и мониторинг цепочек моделей
- **Выбор моделей для ролей** - интерактивный выбор моделей для различных ролей
- **Визуализация прогресса** - отображение прогресса выполнения цепочек в редакторе
- **Управление чекпоинтами** - работа с промежуточными результатами

### Команды MCP

| Команда | Описание |
|---------|----------|
| `chain_progress` | Отображение прогресса выполнения цепочки |
| `models_setup` | Интерактивный выбор моделей для ролей |
| `select_model` | Выбор конкретной модели для роли |
| `chain_list` | Получение списка доступных цепочек |
| `chain_run` | Запуск цепочки моделей |
| `chain_create` | Создание новой цепочки |
| `checkpoint_list` | Получение списка доступных чекпоинтов |
| `checkpoint_get` | Получение содержимого чекпоинта |

## Настройка

### Конфигурация MCP

Для интеграции с редактором кода необходимо настроить файл конфигурации MCP:

```json
{
  "mcpServers": {
    "ricochet": {
      "command": "ricochet-task",
      "args": ["mcp"],
      "env": {
        "ANTHROPIC_API_KEY": "YOUR_ANTHROPIC_API_KEY_HERE",
        "OPENAI_API_KEY": "YOUR_OPENAI_KEY_HERE",
        "GOOGLE_API_KEY": "YOUR_GOOGLE_KEY_HERE",
        "MISTRAL_API_KEY": "YOUR_MISTRAL_KEY_HERE"
      }
    }
  }
}
```

#### Cursor

Для Cursor этот файл должен находиться по пути `~/.cursor/mcp.json`

#### VS Code

Для VS Code необходимо установить соответствующее расширение и настроить его через файл settings.json.

## Пример использования в редакторе

### Выбор моделей для ролей

```
Пользователь: Настрой модели для использования в Ricochet

Ассистент: Запускаю настройку моделей...

? Выберите основную модель:
❯ OpenAI GPT-4o
  Anthropic Claude 3 Opus
  Anthropic Claude 3 Sonnet
  OpenAI GPT-3.5 Turbo
  DeepSeek Coder

? Выберите модель для роли "Анализатор":
❯ Anthropic Claude 3 Opus
  OpenAI GPT-4o
  OpenAI GPT-3.5 Turbo
  Anthropic Claude 3 Sonnet
  DeepSeek Coder

// ... другие роли ...

✅ Настройка моделей завершена!
```

### Мониторинг выполнения цепочки

```
Пользователь: Покажи статус выполнения цепочки "Анализ документа"

Ассистент: Получаю информацию о выполнении цепочки "Анализ документа"...

[▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓----] 65% | Цепочка: Анализ документа

├── [▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓] 100% | Модель: OpenAI GPT-4 (Анализатор)
│   └── Задача #123: Анализ структуры документа ✅ (2.3с)
│   └── Задача #124: Выделение ключевых тем ✅ (3.5с)
│   └── Задача #125: Анализ связей между темами ✅ (2.8с)
│
├── [▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓----------------] 66% | Модель: Claude-3 (Суммаризатор)
│   └── Задача #126: Создание резюме по теме A ✅ (1.7с)
│   └── Задача #127: Создание резюме по теме B ✅ (1.9с)
│   └── Задача #128: Создание резюме по теме C ⏳ (в процессе)
│
└── [-------------------------------] 0% | Модель: DeepSeek (Интегратор)
    └── Задача #129: Объединение резюме 🔜 (ожидание)
    └── Задача #130: Формирование выводов 🔜 (ожидание)

Время выполнения: 5м 0с | Осталось: ~2м 30с
Использовано токенов: 4,500 вход / 2,300 выход
Примерная стоимость: $0.047
```

## Разработка и расширение

### Добавление новой MCP-команды

Для добавления новой MCP-команды необходимо:

1. Создать функцию-обработчик для команды:

```go
func HandleMyCommand(params json.RawMessage) (interface{}, error) {
    var p MyCommandParams
    if err := json.Unmarshal(params, &p); err != nil {
        return nil, fmt.Errorf("failed to parse params: %v", err)
    }
    
    // Логика обработки команды
    
    return MyCommandResponse{...}, nil
}
```

2. Зарегистрировать команду в MCP-сервере:

```go
func RegisterMyCommands(server *MCPServer) {
    server.RegisterCommand("my_command", HandleMyCommand)
}
```

3. Обновить функцию `InitMCPServer`:

```go
func InitMCPServer() *MCPServer {
    server := NewMCPServer()
    
    // Регистрация команд
    RegisterChainProgressCommand(server)
    RegisterModelCommands(server)
    RegisterMyCommands(server)
    
    return server
}
```

# MCP-команды проекта Ricochet Task

Этот документ описывает доступные MCP-команды для интеграции Ricochet Task с редакторами кода.

## Команды для мониторинга выполнения цепочек

### `chain_progress`

Получение информации о прогрессе выполнения цепочки.

**Параметры:**
```json
{
  "chain_id": "chain-123"
}
```

**Пример ответа:**
```json
{
  "chain_id": "chain-123",
  "chain_name": "Анализ документации",
  "status": "running",
  "progress": 0.35,
  "started_at": "2023-06-01T10:15:30Z",
  "estimated_end_time": "2023-06-01T10:25:30Z",
  "elapsed_time": "5m 23s",
  "remaining_time": "10m 0s",
  "model_progresses": [
    {
      "model_id": "model-1",
      "model_name": "GPT-4",
      "provider": "OpenAI",
      "role": "Анализатор",
      "progress": 0.85,
      "status": "running",
      "tasks_total": 3,
      "tasks_done": 2
    },
    {
      "model_id": "model-2",
      "model_name": "Claude-3",
      "provider": "Anthropic",
      "role": "Суммаризатор",
      "progress": 0.0,
      "status": "pending",
      "tasks_total": 3,
      "tasks_done": 0
    }
  ],
  "metrics": {
    "tokens_input": 5420,
    "tokens_output": 834,
    "total_cost": 0.123,
    "requests_count": 5,
    "errors_count": 0
  },
  "current_task_id": "task-123",
  "completed_tasks_ids": ["task-121", "task-122"],
  "progress_chart": "[████████--] 80%"
}
```

### `chain_monitor`

Запуск мониторинга цепочки в реальном времени.

**Параметры:**
```json
{
  "chain_id": "chain-123",
  "include_history": true,
  "refresh_rate": 1000
}
```

**Пример ответа:**
```json
{
  "chain_id": "chain-123",
  "chain_name": "Анализ документации",
  "status": "running",
  "live_view": "┌─────────────┐    ┌─────────────┐    ┌─────────────┐\n│  Анализатор │───>│ Суммаризатор│───>│  Интегратор │\n│   (GPT-4)   │    │  (Claude-3) │    │ (DeepSeek)  │\n│  [██████--] │    │  [----]     │    │  [----]     │\n└─────────────┘    └─────────────┘    └─────────────┘\n      65%                0%                 0%      ",
  "events": [
    {
      "id": "evt-1",
      "chain_id": "chain-123",
      "type": "start",
      "timestamp": "2023-06-01T10:15:30Z",
      "message": "Запуск цепочки",
      "progress": 0.0
    },
    {
      "id": "evt-2",
      "chain_id": "chain-123",
      "type": "step",
      "timestamp": "2023-06-01T10:17:30Z",
      "model_id": "model-1",
      "message": "Выполнение модели анализа",
      "progress": 0.35,
      "task_id": "task-1"
    }
  ],
  "update_time": "2023-06-01T10:20:53Z"
}
```

### `chain_monitor_stop`

Остановка мониторинга цепочки.

**Параметры:**
```json
{
  "chain_id": "chain-123"
}
```

**Пример ответа:**
```json
{
  "success": true,
  "message": "Мониторинг цепочки chain-123 остановлен"
}
```

### `chain_visualization`

Получение визуального представления цепочки.

**Параметры:**
```json
{
  "chain_id": "chain-123",
  "format": "unicode",
  "show_progress": true,
  "show_tasks": true,
  "show_metrics": true,
  "compact": false
}
```

**Пример ответа:**
```json
{
  "chain_id": "chain-123",
  "chain_name": "Анализ документации",
  "visualization": "Цепочка: Анализ документации (ID: chain-123)\nСтатус: выполняется | Прогресс: 28.0%\nПрошло: 15м 23с | Токены: 5420 | Стоимость: $0.123\n\n┌─────────────┐    ───>    ┌─────────────┐    ───>    ┌─────────────┐\n│  Анализа... │    ───>    │  Сумма...   │    ───>    │  Интег...   │\n│   (OpenAI)  │    ───>    │  (Anthr...) │    ───>    │   (Deep...) │\n│  [████████] │    ───>    │  [────────] │    ───>    │  [────────] │\n│    85.0%    │    ───>    │    0.0%     │    ───>    │    0.0%     │\n└─────────────┘    ───>    └─────────────┘    ───>    └─────────────┘\n\n\nЗадачи:\n⏳ Анализатор: 2/3 выполнено\n⏱️ Суммаризатор: 0/3 выполнено\n⏱️ Интегратор: 0/2 выполнено",
  "format": "unicode",
  "generated_at": "2023-06-01T10:20:53Z"
}
```

## Команды для интерактивного конструирования цепочек

### `chain_builder_init`

Инициализация сессии конструктора цепочек.

**Параметры:**
```json
{
  "chain_name": "Новая цепочка анализа",
  "chain_description": "Цепочка для анализа больших документов",
  "metadata": {
    "author": "user@example.com",
    "purpose": "document-analysis"
  },
  "template_id": "analyze-document"
}
```

**Пример ответа:**
```json
{
  "session_id": "session-1623456789",
  "status": "editing",
  "current_step": 0,
  "total_steps": 0,
  "message": "Сессия конструктора цепочек создана",
  "updated_at": "2023-06-01T10:20:53Z"
}
```

### `chain_builder_add_step`

Добавление шага в цепочку.

**Параметры:**
```json
{
  "session_id": "session-1623456789",
  "step_index": 0,
  "model_role": "analyzer",
  "model_id": "gpt-4",
  "provider": "openai",
  "description": "Анализ структуры документа",
  "prompt": "Проанализируйте структуру и основные темы документа. Выделите ключевые разделы и их взаимосвязи.",
  "parameters": {
    "temperature": 0.3,
    "max_tokens": 2000
  }
}
```

**Пример ответа:**
```json
{
  "session_id": "session-1623456789",
  "status": "editing",
  "current_step": 1,
  "total_steps": 1,
  "message": "Шаг 0 добавлен в цепочку",
  "updated_at": "2023-06-01T10:21:53Z"
}
```

### `chain_builder_edit_step`

Редактирование существующего шага.

**Параметры:**
```json
{
  "session_id": "session-1623456789",
  "step_index": 0,
  "model_role": "analyzer",
  "model_id": "gpt-4-turbo",
  "provider": "openai",
  "description": "Улучшенный анализ структуры документа",
  "prompt": "Проанализируйте структуру и основные темы документа. Выделите ключевые разделы, их взаимосвязи и важность.",
  "parameters": {
    "temperature": 0.4,
    "max_tokens": 3000
  }
}
```

**Пример ответа:**
```json
{
  "session_id": "session-1623456789",
  "status": "editing",
  "current_step": 1,
  "total_steps": 1,
  "message": "Шаг 0 обновлен",
  "updated_at": "2023-06-01T10:22:53Z"
}
```

### `chain_builder_remove_step`

Удаление шага из цепочки.

**Параметры:**
```json
{
  "session_id": "session-1623456789",
  "step_index": 0
}
```

**Пример ответа:**
```json
{
  "session_id": "session-1623456789",
  "status": "editing",
  "current_step": 0,
  "total_steps": 0,
  "message": "Шаг 0 удален",
  "updated_at": "2023-06-01T10:23:53Z"
}
```

### `chain_builder_get_session`

Получение информации о текущей сессии конструктора.

**Параметры:**
```json
{
  "session_id": "session-1623456789"
}
```

**Пример ответа:**
```json
{
  "id": "session-1623456789",
  "chain_name": "Новая цепочка анализа",
  "chain_description": "Цепочка для анализа больших документов",
  "steps": [
    {
      "index": 0,
      "model_role": "analyzer",
      "model_id": "gpt-4-turbo",
      "provider": "openai",
      "description": "Улучшенный анализ структуры документа",
      "prompt": "Проанализируйте структуру и основные темы документа. Выделите ключевые разделы, их взаимосвязи и важность.",
      "parameters": {
        "temperature": 0.4,
        "max_tokens": 3000
      },
      "is_completed": true
    }
  ],
  "current_step": 1,
  "status": "editing",
  "metadata": {
    "author": "user@example.com",
    "purpose": "document-analysis"
  },
  "created_at": "2023-06-01T10:20:53Z",
  "updated_at": "2023-06-01T10:22:53Z"
}
```

### `chain_builder_complete`

Завершение сессии конструктора и создание цепочки.

**Параметры:**
```json
{
  "session_id": "session-1623456789",
  "save": true
}
```

**Пример ответа (при save=true):**
```json
{
  "session_id": "session-1623456789",
  "chain_id": "chain-987654321",
  "chain_name": "Новая цепочка анализа",
  "status": "completed",
  "message": "Цепочка успешно создана",
  "updated_at": "2023-06-01T10:25:53Z"
}
```

**Пример ответа (при save=false):**
```json
{
  "session_id": "session-1623456789",
  "status": "canceled",
  "current_step": 1,
  "total_steps": 1,
  "message": "Создание цепочки отменено",
  "updated_at": "2023-06-01T10:25:53Z"
}
```

## Команды для управления моделями

### `models_setup`

Интерактивный выбор моделей для различных ролей.

**Параметры:**
```json
{
  "roles": ["analyzer", "summarizer", "integrator"]
}
```

**Пример ответа:**
```json
{
  "roles": [
    {
      "role_id": "analyzer",
      "display_name": "Анализатор",
      "description": "Модель для анализа структуры и содержания",
      "current_model": {
        "provider": "openai",
        "model_id": "gpt-4",
        "display_name": "GPT-4",
        "max_tokens": 8000,
        "description": "Мощная модель для сложного анализа",
        "capabilities": ["analysis", "understanding", "extraction"],
        "context_size": 8000,
        "cost": "$0.03/1K tokens"
      },
      "options": [
        {
          "provider": "openai",
          "model_id": "gpt-4",
          "display_name": "GPT-4",
          "max_tokens": 8000
        },
        {
          "provider": "anthropic",
          "model_id": "claude-3-opus",
          "display_name": "Claude 3 Opus",
          "max_tokens": 100000
        }
      ]
    }
  ]
}
```

### `select_model`

Выбор конкретной модели для роли.

**Параметры:**
```json
{
  "role_id": "analyzer",
  "provider": "anthropic",
  "model_id": "claude-3-opus",
  "custom_params": {
    "temperature": 0.3
  }
}
```

**Пример ответа:**
```json
{
  "role_id": "analyzer",
  "provider": "anthropic",
  "model_id": "claude-3-opus",
  "display_name": "Claude 3 Opus",
  "success": true,
  "message": "Модель успешно выбрана для роли"
}
```

## Примеры использования MCP-команд в редакторе

### Пример интерактивного создания цепочки в редакторе

1. Инициализация сессии конструктора:
```javascript
const initResponse = await mcp.send("chain_builder_init", {
  chain_name: "Анализ документации проекта",
  chain_description: "Цепочка для анализа документации и создания резюме",
  template_id: "analyze-document"
});
const sessionId = initResponse.data.session_id;
```

2. Получение информации о текущей сессии:
```javascript
const sessionInfo = await mcp.send("chain_builder_get_session", {
  session_id: sessionId
});
const steps = sessionInfo.data.steps;
```

3. Редактирование существующего шага:
```javascript
await mcp.send("chain_builder_edit_step", {
  session_id: sessionId,
  step_index: 0,
  model_role: "analyzer",
  model_id: "gpt-4-turbo",
  provider: "openai",
  description: "Улучшенный анализ структуры документа",
  prompt: "Проанализируйте структуру и основные темы документа. Выделите ключевые разделы, их взаимосвязи и важность.",
  parameters: {
    temperature: 0.4,
    max_tokens: 3000
  }
});
```

4. Добавление нового шага:
```javascript
await mcp.send("chain_builder_add_step", {
  session_id: sessionId,
  step_index: 1,
  model_role: "summarizer",
  model_id: "claude-3-opus",
  provider: "anthropic",
  description: "Суммаризация документа",
  prompt: "На основе анализа структуры, создайте краткое резюме документа, выделив ключевые идеи и выводы.",
  parameters: {
    temperature: 0.4,
    max_tokens: 4000
  }
});
```

5. Завершение создания цепочки:
```javascript
const completeResponse = await mcp.send("chain_builder_complete", {
  session_id: sessionId,
  save: true
});
const chainId = completeResponse.data.chain_id;
```

### Пример мониторинга выполнения цепочки в редакторе

1. Запуск мониторинга:
```javascript
const monitorResponse = await mcp.send("chain_monitor", {
  chain_id: chainId,
  include_history: true,
  refresh_rate: 1000
});
```

2. Получение визуализации:
```javascript
const vizResponse = await mcp.send("chain_visualization", {
  chain_id: chainId,
  format: "unicode", // или "mermaid" для интеграции с Mermaid диаграммами
  show_progress: true,
  show_tasks: true,
  show_metrics: true
});

// Отображение визуализации в пользовательском интерфейсе
displayVisualization(vizResponse.data.visualization);
```

3. Периодическое обновление прогресса:
```javascript
const progressInterval = setInterval(async () => {
  const progressResponse = await mcp.send("chain_progress", {
    chain_id: chainId
  });
  updateProgressUI(progressResponse.data);
  
  if (progressResponse.data.status === "completed" || 
      progressResponse.data.status === "error") {
    clearInterval(progressInterval);
  }
}, 2000);
```

4. Остановка мониторинга при завершении:
```javascript
await mcp.send("chain_monitor_stop", {
  chain_id: chainId
});
``` 