---
description: Описание MCP интеграции для Ricochet
globs: **/*
alwaysApply: true
---
# MCP Интеграция с Ricochet

Ricochet предоставляет интеграцию с Model Control Protocol (MCP) для расширения возможностей ИИ-агентов и IDE.

## Обзор MCP

Model Control Protocol (MCP) — это стандарт для взаимодействия между ИИ-агентами и инструментами. Он позволяет:

- ИИ-агентам (например, ассистентам в Cursor) вызывать внешние инструменты
- Инструментам предоставлять функциональность ИИ-агентам
- Обеспечивать структурированный обмен данными между компонентами

Ricochet интегрируется с MCP, предоставляя набор инструментов для управления задачами проекта.

## Настройка MCP-сервера Ricochet

### Запуск MCP-сервера

```bash
ricochet mcp-server
```

По умолчанию сервер запускается на порту 3456. Вы можете изменить порт с помощью флага:

```bash
ricochet mcp-server --port 4567
```

### Конфигурация в IDE

Для использования Ricochet в Cursor добавьте следующую конфигурацию в `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "ricochet": {
      "command": "ricochet-mcp",
      "args": [],
      "env": {
        "ANTHROPIC_API_KEY": "${ANTHROPIC_API_KEY}",
        "OPENAI_API_KEY": "${OPENAI_API_KEY}",
        "GOOGLE_API_KEY": "${GOOGLE_API_KEY}",
        "MISTRAL_API_KEY": "${MISTRAL_API_KEY}",
        "AZURE_OPENAI_API_KEY": "${AZURE_OPENAI_API_KEY}",
        "OLLAMA_API_KEY": "${OLLAMA_API_KEY}"
      }
    }
  }
}
```

## Доступные MCP инструменты

Ricochet предоставляет следующие MCP инструменты:

### Управление задачами

- `get_tasks`: Получение списка всех задач
- `get_task`: Получение информации о конкретной задаче
- `next_task`: Определение следующей задачи для работы
- `add_task`: Добавление новой задачи
- `update_task`: Обновление существующей задачи
- `set_task_status`: Изменение статуса задачи

### Управление подзадачами

- `add_subtask`: Добавление новой подзадачи
- `update_subtask`: Обновление существующей подзадачи
- `clear_subtasks`: Удаление всех подзадач для указанной задачи
- `expand_task`: Разбивка задачи на подзадачи с использованием ИИ
- `expand_all`: Разбивка всех задач со статусом "pending" на подзадачи

### Управление цепочками моделей

- `chain_list`: Получение списка доступных цепочек
- `chain_run`: Запуск цепочки моделей
- `chain_create`: Создание новой цепочки
- `chain_progress`: Отображение прогресса выполнения цепочки
- `chain_stop`: Остановка выполнения цепочки
- `chain_result`: Получение результата выполнения цепочки

### Управление моделями

- `models_setup`: Интерактивный выбор моделей для ролей
- `select_model`: Выбор конкретной модели для роли
- `models_list`: Получение списка доступных моделей
- `roles_list`: Получение списка доступных ролей
- `provider_info`: Получение информации о провайдере моделей

### Управление зависимостями

- `add_dependency`: Добавление зависимости между задачами
- `remove_dependency`: Удаление существующей зависимости
- `validate_dependencies`: Проверка зависимостей на наличие циклов и других ошибок
- `fix_dependencies`: Автоматическое исправление проблем с зависимостями

### Управление чекпоинтами

- `checkpoint_list`: Получение списка доступных чекпоинтов
- `checkpoint_get`: Получение содержимого чекпоинта
- `checkpoint_create`: Создание нового чекпоинта
- `checkpoint_delete`: Удаление чекпоинта
- `checkpoint_compare`: Сравнение двух чекпоинтов

### Анализ и генерация

- `analyze_project_complexity`: Анализ сложности задач проекта
- `complexity_report`: Получение отчета о сложности задач
- `update`: Обновление нескольких будущих задач на основе изменений в реализации
- `generate_prd`: Генерация PRD на основе текущего состояния проекта
- `parse_prd`: Генерация задач на основе PRD

### Управление проектом

- `initialize_project`: Инициализация нового проекта Ricochet
- `generate`: Генерация файлов задач на основе tasks.json
- `move_task`: Перемещение задачи или подзадачи в иерархии
- `get_project_info`: Получение информации о проекте и статистики задач
- `get_config`: Получение текущей конфигурации Ricochet
- `set_config`: Изменение конфигурации Ricochet

## Примеры использования MCP инструментов

### Получение списка задач

```javascript
// MCP вызов
const result = await mcp.invoke("get_tasks");
console.log(result.tasks);
```

### Получение конкретной задачи

```javascript
// MCP вызов
const result = await mcp.invoke("get_task", {
  id: "5.2"
});
console.log(result.task);
```

### Добавление новой задачи

```javascript
// MCP вызов
const result = await mcp.invoke("add_task", {
  prompt: "Реализовать API для интеграции с внешними системами",
  research: true
});
console.log(result.task);
```

### Изменение статуса задачи

```javascript
// MCP вызов
const result = await mcp.invoke("set_task_status", {
  id: "3",
  status: "in-progress"
});
console.log(result.success);
```

### Отслеживание прогресса цепочки

```javascript
// MCP вызов
const result = await mcp.invoke("chain_progress", {
  chain_id: "chain-123"
});
console.log(result.progress);
console.log(result.progressChart);
```

### Интерактивный выбор моделей

```javascript
// MCP вызов для получения списка ролей и моделей
const setupResult = await mcp.invoke("models_setup", {
  roles: ["main", "research", "analyzer"]
});
console.log(setupResult.roles);

// MCP вызов для выбора модели для роли
const selectResult = await mcp.invoke("select_model", {
  role_id: "main",
  provider: "openai",
  model_id: "gpt-4o"
});
console.log(selectResult.success);
```

### Разбивка задачи на подзадачи

```javascript
// MCP вызов
const result = await mcp.invoke("expand_task", {
  id: "6",
  force: true,
  research: true
});
console.log(result.subtasks);
```

## Структура данных

### Структура задачи

```json
{
  "id": "1",
  "title": "Настройка базовой структуры проекта",
  "description": "Создание первоначальной структуры директорий и базовых файлов проекта",
  "status": "done",
  "priority": "high",
  "dependencies": [],
  "details": "Создать основную структуру директорий проекта...",
  "testStrategy": "Проверить, что структура директорий создана правильно...",
  "subtasks": []
}
```

### Структура подзадачи

```json
{
  "id": "5.1",
  "title": "Реализация основных методов TaskManager",
  "description": "Имплементация базовых операций с задачами",
  "status": "done",
  "details": "Реализовать методы: AddTask, UpdateTask, SetStatus, GetTask"
}
```

### Структура прогресса цепочки

```json
{
  "chain_id": "chain-123",
  "chain_name": "Анализ документа",
  "status": "running",
  "progress": 0.65,
  "started_at": "2024-08-25T15:30:00Z",
  "elapsed_time": "5m 0s",
  "remaining_time": "2m 30s",
  "model_progresses": [
    {
      "model_id": "model-1",
      "model_name": "GPT-4",
      "provider": "openai",
      "role": "analyzer",
      "progress": 1.0,
      "status": "completed",
      "tasks_total": 3,
      "tasks_done": 3
    },
    {
      "model_id": "model-2",
      "model_name": "Claude-3",
      "provider": "anthropic",
      "role": "summarizer",
      "progress": 0.66,
      "status": "running",
      "tasks_total": 3,
      "tasks_done": 2
    }
  ],
  "current_task_id": "task-125",
  "completed_tasks_ids": ["task-123", "task-124"],
  "progress_chart": "..."
}
```

### Структура модели и роли

```json
{
  "roles": [
    {
      "role_id": "main",
      "display_name": "Основная модель",
      "description": "Основная модель для генерации контента и обновлений",
      "current_model": {
        "provider": "openai",
        "model_id": "gpt-4o",
        "display_name": "OpenAI GPT-4o",
        "max_tokens": 16000,
        "description": "Мощная модель для генерации контента и анализа",
        "capabilities": ["code", "reasoning", "creative"],
        "context_size": 32000,
        "cost": "~$0.01/1K токенов"
      },
      "options": [
        { /* Другие доступные модели */ }
      ]
    }
  ]
}
```

## Обработка ошибок

MCP инструменты возвращают структурированные ответы с информацией об ошибках:

```json
{
  "success": false,
  "error": {
    "code": "task_not_found",
    "message": "Задача с ID '12.3' не найдена"
  }
}
```

Типичные коды ошибок:

- `task_not_found`: Задача не найдена
- `invalid_id_format`: Неверный формат ID
- `cyclic_dependency`: Циклическая зависимость
- `invalid_status`: Неверный статус
- `api_error`: Ошибка при взаимодействии с API провайдеров
- `permission_denied`: Недостаточно прав
- `internal_error`: Внутренняя ошибка сервера
- `chain_not_found`: Цепочка не найдена
- `model_not_available`: Модель недоступна
- `invalid_model`: Неверная модель
- `role_not_found`: Роль не найдена

## Советы по эффективному использованию

1. **Выбирайте правильный инструмент**: Используйте специализированные инструменты вместо общих. Например, `update_subtask` лучше, чем `update_task` для подзадач.

2. **Используйте флаг `research`**: Для инструментов, связанных с ИИ (`expand_task`, `add_task`, `update_task`), флаг `research` значительно улучшает качество результатов, но требует больше времени и API вызовов.

3. **Визуализация прогресса**: Используйте `chain_progress` для отображения текущего состояния цепочки. Параметр `progress_chart` содержит готовое ASCII-представление для отображения в консоли или IDE.

4. **Интерактивный выбор моделей**: Инструмент `models_setup` предоставляет полную информацию о доступных моделях и ролях, что позволяет создать интерактивный интерфейс для выбора.

5. **Добавляйте подробности**: Чем детальнее вы опишете задачу или изменение в аргументе `prompt`, тем лучше будет результат.

6. **Обрабатывайте ошибки**: Всегда проверяйте поле `success` в ответе и обрабатывайте ошибки соответствующим образом.

7. **Минимизируйте запросы**: Используйте кэширование результатов для уменьшения количества MCP вызовов.

## Отладка и мониторинг

Для отладки MCP интеграции:

1. Запустите MCP-сервер с флагом `--verbose`:
   ```bash
   ricochet mcp-server --verbose
   ```

2. Проверьте логи MCP-сервера, которые показывают входящие запросы и исходящие ответы.

3. Используйте инструмент `get_config` для проверки текущей конфигурации.

## Безопасность

- MCP-сервер Ricochet не требует аутентификации по умолчанию, поэтому используйте его только в доверенной среде.
- API ключи хранятся в зашифрованном виде в конфигурации Ricochet.
- Для повышения безопасности можно настроить аутентификацию в `config.json`. 