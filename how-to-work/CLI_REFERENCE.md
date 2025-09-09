# 📖 CLI Reference - Полный справочник команд

Исчерпывающий справочник всех команд Ricochet Task с реальными примерами использования.

## 🎯 Основные команды

### Справка и информация

```bash
# Общая справка
./ricochet-task --help

# Справка по конкретной команде
./ricochet-task [command] --help

# Версия и информация
./ricochet-task --version
```

### Глобальные флаги

```bash
-c, --config string    # Путь к файлу конфигурации
-i, --interactive      # Интерактивный режим
-v, --verbose         # Подробный вывод
```

## 🔐 Команды key - Управление API-ключами

### Добавление ключей

```bash
# Базовое добавление
./ricochet-task key add --provider openai --key "sk-proj-ваш-ключ"

# С лимитом токенов
./ricochet-task key add --provider anthropic --key "sk-ant-ключ" --limit 1000000

# С общим доступом
./ricochet-task key add --provider deepseek --key "ключ" --shared

# Поддерживаемые провайдеры: openai, anthropic, deepseek, grok
```

### Просмотр и управление

```bash
# Список всех ключей
./ricochet-task key list

# Обновление существующего ключа
./ricochet-task key update --provider openai --key "новый-ключ"
./ricochet-task key update --provider openai --limit 2000000

# Удаление ключа
./ricochet-task key delete --provider openai
./ricochet-task key delete --id "uuid-ключа"
```

### Общий доступ

```bash
# Включить общий доступ
./ricochet-task key share --provider openai --enable

# Отключить общий доступ  
./ricochet-task key share --provider openai --disable
```

## 🔌 Команды providers - Управление провайдерами

### Просмотр провайдеров

```bash
# Список всех провайдеров
./ricochet-task providers list

# Только активные провайдеры
./ricochet-task providers list --enabled-only

# В формате JSON
./ricochet-task providers list --output json
```

### Добавление провайдеров

```bash
# YouTrack
./ricochet-task providers add my-youtrack \
  --type youtrack \
  --base-url "https://company.youtrack.cloud" \
  --token "perm-токен"

# Jira (когда будет поддержка)
./ricochet-task providers add company-jira \
  --type jira \
  --base-url "https://company.atlassian.net" \
  --token "jira-токен"

# Автоматическое включение после добавления
./ricochet-task providers add my-provider --enable
```

### Управление состоянием

```bash
# Включение/отключение
./ricochet-task providers enable gamesdrop-youtrack
./ricochet-task providers disable gamesdrop-youtrack

# Установка провайдера по умолчанию
./ricochet-task providers default gamesdrop-youtrack

# Удаление провайдера
./ricochet-task providers remove my-youtrack --force
```

### Мониторинг здоровья

```bash
# Проверка всех провайдеров
./ricochet-task providers health

# Конкретный провайдер
./ricochet-task providers health gamesdrop-youtrack

# С подробной информацией
./ricochet-task providers health gamesdrop-youtrack --verbose

# Непрерывный мониторинг
./ricochet-task providers health --watch --interval 30s
```

## 📋 Команды tasks - Управление задачами

### Создание задач

```bash
# Базовое создание
./ricochet-task tasks create \
  --title "Исправить баг авторизации" \
  --description "Описание проблемы"

# С полными параметрами
./ricochet-task tasks create \
  --title "Новая функция" \
  --description "Подробное описание" \
  --provider gamesdrop-youtrack \
  --type feature \
  --priority high \
  --assignee "john.doe" \
  --project "BACKEND"

# Типы задач: task, bug, feature, epic, story, subtask
# Приоритеты: lowest, low, medium, high, highest, critical
```

### Просмотр задач

```bash
# Все задачи из всех провайдеров
./ricochet-task tasks list --providers all

# Из конкретного провайдера
./ricochet-task tasks list --provider gamesdrop-youtrack

# С фильтрацией
./ricochet-task tasks list \
  --provider gamesdrop-youtrack \
  --status "Open" \
  --priority "High" \
  --assignee "me" \
  --limit 50

# В разных форматах
./ricochet-task tasks list --output table    # По умолчанию
./ricochet-task tasks list --output json
./ricochet-task tasks list --output summary
```

### Поиск задач

```bash
# Поиск по всем провайдерам
./ricochet-task tasks search "авторизация" --providers all

# В конкретном провайдере
./ricochet-task tasks search "баг" --provider gamesdrop-youtrack

# С лимитом результатов
./ricochet-task tasks search "security" --limit 100
```

### Управление задачами

```bash
# Получение информации о задаче
./ricochet-task tasks get PROJ-123 --provider gamesdrop-youtrack

# Обновление задачи
./ricochet-task tasks update PROJ-123 \
  --status "in_progress" \
  --assignee "jane.doe" \
  --priority "highest" \
  --provider gamesdrop-youtrack

# Удаление задачи (осторожно!)
./ricochet-task tasks delete PROJ-123 --provider gamesdrop-youtrack --force
```

## 🔗 Команды chain - Управление цепочками

### Создание цепочек

```bash
# Простое создание
./ricochet-task chain create \
  --name "test-chain" \
  --description "Тестовая цепочка"

# С моделями
./ricochet-task chain create \
  --name "analysis-chain" \
  --description "Цепочка для анализа кода"
```

### Просмотр цепочек

```bash
# Список всех цепочек
./ricochet-task chain list

# Подробная информация о цепочке
./ricochet-task chain get fde1701a-7890-4bf9-85b4-d20d4935ed5f

# Статус выполнения
./ricochet-task chain status fde1701a-7890-4bf9-85b4-d20d4935ed5f
```

### Управление цепочками

```bash
# Добавление модели в цепочку
./ricochet-task chain add-model \
  --chain fde1701a-7890-4bf9-85b4-d20d4935ed5f \
  --model "gpt-4" \
  --position 1

# Запуск цепочки
./ricochet-task chain run \
  --chain fde1701a-7890-4bf9-85b4-d20d4935ed5f \
  --input "Текст для обработки"

# Обновление цепочки
./ricochet-task chain update \
  --chain fde1701a-7890-4bf9-85b4-d20d4935ed5f \
  --name "новое-имя" \
  --description "новое описание"

# Удаление цепочки
./ricochet-task chain delete fde1701a-7890-4bf9-85b4-d20d4935ed5f --force
```

## 💾 Команды checkpoint - Управление чекпоинтами

### Создание и сохранение

```bash
# Сохранение чекпоинта
./ricochet-task checkpoint save \
  --chain fde1701a-7890-4bf9-85b4-d20d4935ed5f \
  --content '{"step": 1, "result": "processed"}' \
  --type input

# Из файла
./ricochet-task checkpoint save \
  --chain fde1701a-7890-4bf9-85b4-d20d4935ed5f \
  --input-file ./checkpoint-data.json \
  --type output

# Типы чекпоинтов: input, output, segment, complete
```

### Просмотр чекпоинтов

```bash
# Список чекпоинтов цепочки
./ricochet-task checkpoint list --chain fde1701a-7890-4bf9-85b4-d20d4935ed5f

# Получение содержимого чекпоинта
./ricochet-task checkpoint get 28ad8d9c-7874-4cae-9541-79010615294f

# Обновление чекпоинта
./ricochet-task checkpoint update \
  --id 28ad8d9c-7874-4cae-9541-79010615294f \
  --content '{"updated": "data"}'
```

### Удаление чекпоинтов

```bash
# Удаление конкретного чекпоинта
./ricochet-task checkpoint delete 28ad8d9c-7874-4cae-9541-79010615294f

# Очистка всех чекпоинтов цепочки
./ricochet-task checkpoint delete --chain fde1701a-7890-4bf9-85b4-d20d4935ed5f --all
```

## 🖥️ Команды mcp - MCP сервер

### Запуск сервера

```bash
# Стандартный запуск
./ricochet-task mcp start

# На конкретном порту
./ricochet-task mcp start --port 8080

# С подробным выводом
./ricochet-task mcp start --verbose --port 3001

# На всех интерфейсах
./ricochet-task mcp start --host 0.0.0.0 --port 3001

# С отладкой
./ricochet-task mcp start --debug --verbose
```

### Управление MCP

```bash
# Список доступных инструментов
./ricochet-task mcp tools

# Проверка конфигурации MCP
./ricochet-task mcp validate

# Остановка сервера (если запущен в фоне)
pkill -f "ricochet-task mcp"
```

## 🌍 Команды context - Управление контекстом

### Установка контекста

```bash
# Установка рабочей доски
./ricochet-task context set-board \
  --board-id "DEV-BOARD" \
  --project-id "BACKEND" \
  --provider gamesdrop-youtrack \
  --default-assignee "team-lead" \
  --default-labels "sprint-1,backend"
```

### Просмотр контекста

```bash
# Текущий контекст
./ricochet-task context get-current

# С подробной информацией
./ricochet-task context get-current --include-board-info

# Список всех досок
./ricochet-task context list-boards --provider gamesdrop-youtrack
```

## 📋 Команды board - Управление досками

### Интерактивная работа с досками

```bash
# Интерактивный выбор доски
./ricochet-task board

# Список досок
./ricochet-task board list --provider gamesdrop-youtrack

# Установка активной доски
./ricochet-task board set --board-id "MAIN-BOARD" --provider gamesdrop-youtrack
```

## ⚡ Команды workflow - Workflow Engine

### Создание workflow

```bash
# Создание из файла конфигурации
./ricochet-task workflow create --name "deploy-workflow" --config deploy.yaml

# Запуск workflow
./ricochet-task workflow run deploy-workflow --input '{"version": "1.2.3"}'

# Список workflow
./ricochet-task workflow list

# Статус выполнения
./ricochet-task workflow status workflow-id
```

## 🚀 Специальные команды

### Инициализация

```bash
# Интерактивная настройка (как Claude CLI)
./ricochet-task init

# Интерактивный режим для любой команды
./ricochet-task --interactive key add
./ricochet-task --interactive providers add
```

### HTTP сервер

```bash
# Запуск HTTP API сервера
./ricochet-task --http

# Проверка health
curl http://localhost:6004/health
```

### Примеры комплексных команд

```bash
# Создание полного workflow
./ricochet-task key add --provider openai --key "ключ" && \
./ricochet-task providers health && \
./ricochet-task chain create --name "prod-chain" && \
./ricochet-task mcp start --port 3001 &

# Batch операции
./ricochet-task tasks list --status "Open" | \
jq -r '.[] | .id' | \
xargs -I {} ./ricochet-task tasks update {} --priority "medium"
```

## ⚙️ Переменные окружения

```bash
# Конфигурация через переменные окружения
export RICOCHET_CONFIG_DIR="~/.ricochet"
export RICOCHET_DEFAULT_CHAIN="main-chain"
export RICOCHET_WORKSPACE_PATH="./"
export POSTGRES_DSN="postgres://user:pass@localhost/db"
export MINIO_ENDPOINT="localhost:9000"
export MINIO_ACCESS_KEY="minioadmin"
export MINIO_SECRET_KEY="password"
```

## 🆘 Отладка и диагностика

```bash
# Максимально подробный вывод
./ricochet-task --verbose command

# Отладочная информация
./ricochet-task --debug command

# Проверка конфигурации
./ricochet-task config validate

# Проверка подключений
./ricochet-task providers health --verbose
./ricochet-task key list
```

---

**Этот справочник покрывает все основные команды Ricochet Task. Для получения актуальной справки всегда используйте `--help`!** 📚