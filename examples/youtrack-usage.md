# YouTrack Integration Usage Examples

## 🔐 Первоначальная настройка

### 1. Получение токена в YouTrack

1. Заходим в YouTrack: `https://your-company.youtrack.cloud`
2. Profile → Authentication → New token...
3. Создаем permanent token с scope:
   - `YouTrack` (полный доступ к задачам)
   - Или конкретные права: `Read Issues`, `Create Issues`, `Update Issues`

### 2. Добавление провайдера

```bash
# Добавляем YouTrack провайдер
ricochet providers add youtrack-prod \
  --type youtrack \
  --base-url https://your-company.youtrack.cloud \
  --token perm:your-permanent-token-here

# Проверяем что провайдер работает
ricochet providers health youtrack-prod
```

## 🎯 Основные операции

### Создание задач

```bash
# Простое создание
ricochet tasks create \
  --title "Fix authentication bug" \
  --provider youtrack-prod

# Создание с деталями
ricochet tasks create \
  --title "Implement OAuth integration" \
  --description "Add OAuth 2.0 support for external APIs" \
  --provider youtrack-prod \
  --project BACKEND \
  --type feature \
  --priority high \
  --assignee john.doe
```

### Поиск и фильтрация

```bash
# Список открытых задач
ricochet tasks list --provider youtrack-prod --status open

# Мои задачи
ricochet tasks list --provider youtrack-prod --assignee me

# Поиск по ключевым словам
ricochet tasks search "authentication" --provider youtrack-prod

# Сложный поиск
ricochet tasks search --query "assignee:me and priority:high" --provider youtrack-prod
```

### Обновление задач

```bash
# Изменение статуса
ricochet tasks update PROJ-123 \
  --status "In Progress" \
  --provider youtrack-prod

# Назначение исполнителя
ricochet tasks update PROJ-123 \
  --assignee jane.smith \
  --provider youtrack-prod
```

## 🤖 MCP интеграция в VS Code/Cursor

### Запуск MCP сервера

```bash
# Запуск MCP сервера для VS Code/Cursor
ricochet mcp start --port 3001

# Проверка доступных инструментов
ricochet mcp tools
```

### Доступные MCP инструменты

Когда MCP сервер запущен, AI ассистент в VS Code/Cursor получает доступ к:

1. **providers_list** - Список всех провайдеров
2. **task_create_smart** - Умное создание задач
3. **task_list_unified** - Список задач из всех провайдеров
4. **cross_provider_search** - Поиск по всем системам
5. **ai_analyze_project** - Анализ проекта с AI

### Пример работы в VS Code

```typescript
// AI ассистент может:

// 1. Создать задачу из комментария в коде
/* TODO: Optimize database queries */
// → Автоматически создается задача в YouTrack

// 2. Найти связанные задачи
// "Найди все задачи по authentication"
// → Ищет по всем провайдерам

// 3. Обновить статус после коммита
// git commit -m "Fix auth bug"
// → Автоматически обновляет статус связанной задачи
```

## 📊 Консольный вывод

### Успешные операции

```bash
$ ricochet tasks create --title "New feature" --provider youtrack-prod
✅ Task created successfully
ID: PROJ-456
Title: New feature
Provider: youtrack-prod
```

### Список задач (таблица)

```bash
$ ricochet tasks list --provider youtrack-prod --limit 5
ID             PROVIDER     TITLE                                    STATUS       PRIORITY  
--             --------     -----                                    ------       --------  
PROJ-123       youtrack-prod Fix authentication bug                   Open         High      
PROJ-124       youtrack-prod Implement OAuth integration              In Progress  Medium    
PROJ-125       youtrack-prod Update documentation                     Open         Low       
```

### Здоровье провайдеров

```bash
$ ricochet providers health
Provider Health Status:
========================
🟢 youtrack-prod: healthy
🟢 jira-dev: healthy
🔴 notion-docs: unhealthy
```

### Ошибки

```bash
$ ricochet tasks create --title "Test" --provider invalid-provider
❌ Error: Provider 'invalid-provider' not found

$ ricochet providers add test --type youtrack --base-url https://invalid.url --token invalid
❌ Error: Failed to add provider: YouTrack API error 401: Unauthorized
```

## 🔧 Отладка и логирование

### Подробные логи

```bash
# Запуск с отладкой
ricochet --debug providers health youtrack-prod

# MCP сервер с логированием
ricochet mcp start --debug --port 3001
```

### Конфигурация

```yaml
# ricochet.yaml
providers:
  youtrack-prod:
    type: "youtrack"
    enabled: true
    baseUrl: "https://company.youtrack.cloud"
    token: "${YOUTRACK_TOKEN}"
    settings:
      defaultProject: "BACKEND"
      autoCreateBoards: false
```

## 🚀 Продвинутые сценарии

### Синхронизация между провайдерами

```bash
# Синхронизация между YouTrack и Jira
ricochet tasks sync \
  --from youtrack-prod \
  --to jira-company \
  --project BACKEND
```

### Мульти-провайдерный поиск

```bash
# Поиск по всем провайдерам
ricochet tasks search "authentication" --providers all

# Поиск в конкретных провайдерах
ricochet tasks search "bug" --providers youtrack-prod,jira-dev
```

### Аналитика (будущее)

```bash
# AI анализ проекта
ricochet analytics project BACKEND --providers youtrack-prod

# Отчет по производительности команды
ricochet analytics velocity --timeframe week --providers all
```

## 🔐 Безопасность

### Переменные окружения

```bash
# Безопасное хранение токенов
export YOUTRACK_TOKEN="perm:your-token-here"
export JIRA_TOKEN="your-jira-token"

# Использование в конфигурации
ricochet providers add youtrack-prod \
  --type youtrack \
  --base-url https://company.youtrack.cloud \
  --token "$YOUTRACK_TOKEN"
```

### Проверка прав

```bash
# Проверка что пользователь может создавать задачи
ricochet providers validate youtrack-prod

# Проверка конкретного проекта
ricochet providers validate youtrack-prod --project BACKEND
```

Таким образом пользователь получает полный контроль над задачами через удобный интерфейс, а AI ассистенты могут автоматизировать рутинные операции!