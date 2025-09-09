# 🔌 Работа с провайдерами задач

Ricochet Task поддерживает множественные провайдеры систем управления задачами. Это позволяет работать с разными командами и проектами через единый интерфейс.

## 🎯 Поддерживаемые провайдеры

- **YouTrack** - JetBrains YouTrack (полная поддержка)
- **Jira** - Atlassian Jira (в разработке)
- **Notion** - Notion Database (планируется)
- **Linear** - Linear Issues (планируется)

## 📋 Просмотр провайдеров

### Список всех провайдеров

```bash
# Просмотр всех настроенных провайдеров
./ricochet-task providers list
```

**Пример вывода:**
```
NAME                 TYPE         STATUS     HEALTH          CAPABILITIES                  
----                 ----         ------     ------          ------------                  
gamesdrop-youtrack   youtrack     enabled    healthy         tasks, boards, real_time
```

### Проверка здоровья провайдеров

```bash
# Проверка всех провайдеров
./ricochet-task providers health

# Проверка конкретного провайдера
./ricochet-task providers health gamesdrop-youtrack --verbose

# Непрерывный мониторинг
./ricochet-task providers health --watch --interval 30s
```

## ⚙️ Конфигурация YouTrack

### Настройка через ricochet.yaml

Основная конфигурация YouTrack находится в файле `ricochet.yaml`:

```yaml
providers:
  gamesdrop-youtrack:
    name: gamesdrop-youtrack
    type: youtrack
    enabled: true
    baseUrl: https://gamesdrop.youtrack.cloud
    authType: bearer
    token: perm-YWRtaW4=.NTItMA==.75T2Un6ARYfePI3oP9ZoJAXzC8bZgs
    timeout: 60s
    settings:
      defaultProject: ""
      defaultBoard: ""
      autoCreateBoards: false
      useShortNames: true
      syncComments: true
      syncAttachments: true
      syncTimeTracking: true
      syncCustomFields: true
      customFieldMappings:
        story_points: Story Points
        sprint: Sprint
        epic: Epic
      workflowMappings:
        todo: Open
        in_progress: In Progress
        done: Fixed
        blocked: Blocked
    rateLimits:
      requestsPerSecond: 10
      burstSize: 50
    retryConfig:
      maxRetries: 3
      retryableErrors:
        - "429"
        - "500"
        - "502"
        - "503"
        - "504"

defaultProvider: gamesdrop-youtrack
```

### Добавление нового YouTrack провайдера

```bash
# Добавление через CLI
./ricochet-task providers add my-youtrack \
  --type youtrack \
  --base-url "https://company.youtrack.cloud" \
  --token "perm-ваш-токен-здесь"

# Проверка добавления
./ricochet-task providers list

# Включение провайдера
./ricochet-task providers enable my-youtrack
```

## 🔧 Управление провайдерами

### Включение/отключение провайдеров

```bash
# Отключение провайдера (не удаляет, только деактивирует)
./ricochet-task providers disable gamesdrop-youtrack

# Включение провайдера
./ricochet-task providers enable gamesdrop-youtrack

# Установка провайдера по умолчанию
./ricochet-task providers default gamesdrop-youtrack
```

### Удаление провайдеров

```bash
# Удаление провайдера (внимательно! удаляет всю конфигурацию)
./ricochet-task providers remove my-youtrack

# Подтверждение удаления обычно требуется
./ricochet-task providers remove my-youtrack --force
```

## 🎯 Работа с задачами через провайдеры

### Создание задач

```bash
# Создание задачи в конкретном провайдере
./ricochet-task tasks create \
  --title "Исправить баг авторизации" \
  --description "Пользователи не могут войти через OAuth" \
  --provider gamesdrop-youtrack \
  --type bug \
  --priority high

# Создание в провайдере по умолчанию
./ricochet-task tasks create \
  --title "Добавить новую функцию" \
  --type feature \
  --priority medium
```

### Просмотр задач

```bash
# Задачи из всех провайдеров
./ricochet-task tasks list --providers all

# Задачи из конкретного провайдера
./ricochet-task tasks list --provider gamesdrop-youtrack

# С фильтрацией
./ricochet-task tasks list \
  --provider gamesdrop-youtrack \
  --status "Open" \
  --priority "High" \
  --assignee "john.doe"
```

### Поиск задач

```bash
# Поиск по всем провайдерам
./ricochet-task tasks search "авторизация" --providers all

# Поиск в конкретном провайдере
./ricochet-task tasks search "баг" --provider gamesdrop-youtrack --limit 50
```

## 🌐 Мультипровайдерные операции

### Синхронизация между провайдерами

```bash
# Настройка синхронизации (в ricochet.yaml)
globalSync:
  enabled: true
  rules:
    - sourceProvider: "youtrack-dev"
      targetProvider: "jira-prod"
      syncType: "bidirectional"
      fieldMappings:
        title: summary
        description: description
        status: status
```

### Кросс-провайдерный поиск через MCP

```bash
# Запуск MCP сервера
./ricochet-task mcp start --port 3001

# В VS Code с Claude:
# "Найди все задачи по безопасности во всех системах за последний месяц"
```

**MCP автоматически использует инструмент:**
```json
{
  "name": "cross_provider_search",
  "parameters": {
    "query": "security OR безопасность",
    "providers": ["all"],
    "include_content": true,
    "limit": 100
  }
}
```

## 🏗️ Настройка дополнительных провайдеров

### Подготовка к добавлению Jira

```bash
# Когда поддержка Jira будет готова:
./ricochet-task providers add company-jira \
  --type jira \
  --base-url "https://company.atlassian.net" \
  --token "ваш-jira-токен"

# Настройка синхронизации YouTrack <-> Jira
./ricochet-task workflow create --name "youtrack-jira-sync"
```

### Подготовка Notion интеграции

```bash
# Будущая поддержка Notion:
./ricochet-task providers add team-notion \
  --type notion \
  --base-url "https://api.notion.com" \
  --token "secret_ваш-notion-токен"
```

## 🔍 Диагностика провайдеров

### Проверка подключения

```bash
# Детальная проверка здоровья
./ricochet-task providers health gamesdrop-youtrack --verbose
```

**Успешный вывод:**
```
[08:51:09] gamesdrop-youtrack: 🟢 HEALTHY
Capabilities: tasks, boards, real_time_sync, webhooks
Response time: 245ms
Last sync: 2025-09-06T08:50:15+05:00
```

### Отладка проблем подключения

```bash
# Проверка с подробным выводом
./ricochet-task --verbose providers health gamesdrop-youtrack

# Проверка конфигурации
cat ricochet.yaml | grep -A 20 "gamesdrop-youtrack"

# Проверка сетевого подключения
curl -I https://gamesdrop.youtrack.cloud/api/admin/projects
```

### Логи и мониторинг

```bash
# Запуск с детальными логами
./ricochet-task --verbose providers list

# Непрерывный мониторинг всех провайдеров
./ricochet-task providers health --watch --interval 60s
```

## ⚡ Оптимизация производительности

### Настройка rate limiting

В `ricochet.yaml`:

```yaml
rateLimits:
  requestsPerSecond: 10    # Не более 10 запросов в секунду
  burstSize: 50           # Пиковая нагрузка до 50 запросов
```

### Настройка retry логики

```yaml
retryConfig:
  maxRetries: 3
  retryableErrors:
    - "429"  # Rate limit exceeded
    - "500"  # Server error
    - "502"  # Bad gateway
    - "503"  # Service unavailable
    - "504"  # Gateway timeout
```

### Настройка таймаутов

```yaml
timeout: 60s              # Общий таймаут запросов
settings:
  connectionTimeout: 30s   # Таймаут подключения
  readTimeout: 45s         # Таймаут чтения ответа
```

## 🎉 Готовые workflow с провайдерами

### Автоматическое создание задач

```bash
# Через MCP в VS Code:
# "Создай задачу 'Code Review для PR #123' в YouTrack с высоким приоритетом"

# MCP использует:
{
  "tool": "task_create_smart",
  "parameters": {
    "title": "Code Review для PR #123",
    "priority": "high",
    "provider": "gamesdrop-youtrack",
    "task_type": "task"
  }
}
```

### Массовое обновление задач

```bash
# Через MCP:
# "Обнови статус всех задач типа 'bug' со статусом 'Open' на 'In Review'"

# MCP использует batch операции для эффективности
```

---

**Следующий шаг**: Переходите к [Созданию цепочек](./04_chains.md) для автоматизации обработки данных! 🚀