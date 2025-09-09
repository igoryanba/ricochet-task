# 🔌 MCP сервер - Интеграция с редакторами кода

Model Context Protocol (MCP) сервер Ricochet Task предоставляет мощные инструменты для AI-ассистентов в VS Code, Cursor и других редакторах. Это позволяет автоматизировать управление задачами прямо из среды разработки.

## 🚀 Запуск MCP сервера

### Основные команды запуска

```bash
# Стандартный запуск на порту 3001
./ricochet-task mcp start

# Запуск с подробным выводом
./ricochet-task mcp start --verbose --port 3001

# Запуск на другом порту
./ricochet-task mcp start --port 8080

# Запуск на всех интерфейсах (для удаленного доступа)
./ricochet-task mcp start --host 0.0.0.0 --port 3001
```

### Проверка работы сервера

```bash
# Проверка статуса
curl -s http://localhost:3001/tools | jq '."tools" | length'

# Список всех доступных инструментов
curl -s http://localhost:3001/tools | jq '.tools[].name'
```

## 🛠️ Доступные MCP инструменты

### 1. Управление провайдерами (3 инструмента)

**`providers_list`** - Список всех провайдеров
```json
{
  "enabled_only": false,
  "output_format": "table"
}
```

**`provider_health`** - Проверка здоровья провайдеров
```json
{
  "provider_name": "gamesdrop-youtrack",
  "include_details": true
}
```

**`providers_add`** - Добавление нового провайдера
```json
{
  "name": "my-youtrack",
  "type": "youtrack",
  "base_url": "https://company.youtrack.cloud",
  "token": "your-token-here",
  "enable": true
}
```

### 2. Управление задачами (4 инструмента)

**`task_create_smart`** - Умное создание задач
```json
{
  "title": "Implement user authentication",
  "description": "Add OAuth2 login functionality",
  "task_type": "feature",
  "priority": "high",
  "assignee": "john.doe",
  "provider": "gamesdrop-youtrack"
}
```

**`task_list_unified`** - Список задач из всех провайдеров
```json
{
  "providers": ["all"],
  "status": "open",
  "priority": "high",
  "limit": 50,
  "output_format": "json"
}
```

**`task_update_universal`** - Универсальное обновление задач
```json
{
  "task_id": "PROJ-123",
  "status": "in_progress",
  "assignee": "jane.doe",
  "add_labels": ["bug", "critical"],
  "priority": "highest"
}
```

**`cross_provider_search`** - Поиск по всем провайдерам
```json
{
  "query": "authentication bug",
  "providers": ["all"],
  "include_content": true,
  "limit": 20
}
```

### 3. Контекстное управление (3 инструмента)

**`context_set_board`** - Установка рабочего контекста
```json
{
  "board_id": "123-456",
  "project_id": "MYPROJ",
  "provider": "gamesdrop-youtrack",
  "default_assignee": "developer.team",
  "default_labels": ["sprint-1", "backend"]
}
```

**`context_get_current`** - Получение текущего контекста
```json
{
  "include_board_info": true
}
```

**`context_list_boards`** - Список досок всех провайдеров
```json
{
  "provider": "gamesdrop-youtrack",
  "output_format": "table"
}
```

### 4. AI-планирование (3 инструмента)

**`ai_create_project_plan`** - AI создание плана проекта
```json
{
  "description": "Create REST API for user management with authentication, CRUD operations, and role-based access control",
  "project_type": "feature",
  "complexity": "medium",
  "timeline_days": 21,
  "team_size": 3,
  "auto_create_tasks": true,
  "priority": "high"
}
```

**`ai_execute_plan`** - Выполнение плана
```json
{
  "plan_id": "plan-uuid-here",
  "create_epic": true,
  "start_immediately": false,
  "board_context": "current"
}
```

**`ai_track_progress`** - Отслеживание прогресса
```json
{
  "task_ids": ["PROJ-123", "PROJ-124"],
  "update_statuses": true,
  "add_progress_comments": true,
  "generate_report": false
}
```

### 5. AI-анализ (2 инструмента)

**`ai_analyze_project`** - Анализ проекта
```json
{
  "project_id": "MYPROJ",
  "analysis_type": "full",
  "providers": ["all"],
  "timeframe_days": 30
}
```

**`ai_execute_task`** - AI выполнение задач
```json
{
  "task_id": "PROJ-123",
  "execution_mode": "implement",
  "auto_update_status": true,
  "create_subtasks": true
}
```

## 🔧 Интеграция с VS Code

### 1. Установка Claude Dev Extension

```bash
# Установите расширение Claude Dev в VS Code
code --install-extension anthropic.claude-dev
```

### 2. Настройка MCP в VS Code

Добавьте в settings.json VS Code:

```json
{
  "claude-dev.mcpServers": {
    "ricochet-task": {
      "command": "/path/to/ricochet-task",
      "args": ["mcp", "start", "--port", "3001"],
      "url": "http://localhost:3001"
    }
  }
}
```

### 3. Запуск и использование

```bash
# 1. Запустите MCP сервер
./ricochet-task mcp start --port 3001 --verbose

# 2. Откройте VS Code
code .

# 3. Используйте Claude для работы с задачами
# Пример: "@ricochet создай задачу 'Исправить баг с авторизацией' в проекте BACKEND"
```

## 🎯 Интеграция с Cursor

### Настройка в Cursor

1. Откройте Cursor Settings
2. Найдите раздел "MCP Servers"
3. Добавьте конфигурацию:

```json
{
  "mcpServers": {
    "ricochet-task": {
      "command": "./ricochet-task",
      "args": ["mcp", "start", "--port", "3001"]
    }
  }
}
```

## 📋 Практические примеры использования

### Сценарий 1: Создание задачи из кода

```bash
# В VS Code/Cursor с Claude:
# "Создай задачу для рефакторинга этой функции с приоритетом high"

# MCP инструмент вызовется автоматически:
task_create_smart({
  "title": "Refactor getUserData function",
  "description": "Optimize database queries and improve error handling",
  "task_type": "refactoring",
  "priority": "high",
  "assignee": "current-user"
})
```

### Сценарий 2: Анализ проекта

```bash
# "Проанализируй текущий проект и покажи все критические задачи"

# Вызовы MCP:
ai_analyze_project({
  "project_id": "CURRENT",
  "analysis_type": "blockers",
  "timeframe_days": 7
})

task_list_unified({
  "priority": "critical",
  "status": "open",
  "output_format": "summary"
})
```

### Сценарий 3: Автоматическое планирование

```bash
# "Создай план разработки новой функции логина с OAuth"

# MCP создаст план и задачи:
ai_create_project_plan({
  "description": "OAuth login functionality with Google and GitHub providers",
  "complexity": "medium",
  "auto_create_tasks": true,
  "timeline_days": 14
})
```

## 🚨 Диагностика проблем

### Проверка подключения

```bash
# Тест подключения к MCP серверу
curl -v http://localhost:3001/tools

# Проверка логов сервера
./ricochet-task mcp start --verbose --debug
```

### Частые ошибки

**Порт занят:**
```bash
lsof -i :3001
./ricochet-task mcp start --port 8080
```

**Провайдеры не инициализированы:**
```bash
./ricochet-task providers list
./ricochet-task providers health --verbose
```

**Отсутствуют API ключи:**
```bash
./ricochet-task key list
./ricochet-task key add --provider openai --key YOUR_KEY
```

## 🎉 Результат

После настройки MCP интеграции вы получаете:

✅ **15 специализированных инструментов** для AI-ассистентов
✅ **Прямую интеграцию с VS Code/Cursor**
✅ **Автоматическое управление задачами из кода**
✅ **AI-планирование и анализ проектов**
✅ **Унифицированную работу с множеством провайдеров**

Теперь ваш AI-ассистент может создавать задачи, анализировать проекты и автоматизировать workflow прямо из редактора кода! 🚀