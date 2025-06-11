# Ricochet Task

<p align="center">
  <img src="https://raw.githubusercontent.com/grik-ai/ricochet-task/main/assets/logo.png" alt="Ricochet Task Logo" width="200"/>
</p>

<p align="center">
  <strong>Мощный CLI-инструмент для оркестрации языковых моделей и обработки больших объемов текста</strong>
</p>

<p align="center">
  <a href="#возможности">Возможности</a> •
  <a href="#установка">Установка</a> •
  <a href="#быстрый-старт">Быстрый старт</a> •
  <a href="#примеры-использования">Примеры использования</a> •
  <a href="#интеграция-с-редакторами">Интеграция с редакторами</a> •
  <a href="#документация">Документация</a>
</p>

## Возможности

Ricochet Task — инструмент для управления цепочками языковых моделей, который позволяет обрабатывать большие объемы текстовой информации, значительно превышающие ограничения контекстного окна отдельных моделей.

🔄 **Цепочки моделей** — Создание и управление цепочками специализированных моделей для обработки данных.

📊 **Ролевая специализация** — Назначение различных ролей моделям (анализатор, суммаризатор, интегратор) для эффективной обработки информации.

📋 **Чекпоинты** — Сохранение промежуточных результатов между этапами обработки.

🚀 **Интеграция с редакторами кода** — Прямая интеграция с редакторами через MCP (Model Control Protocol).

🔑 **Управление API-ключами** — Безопасное управление и совместное использование API-ключей для различных провайдеров.

🎯 **Интерактивный выбор моделей** — Гибкое назначение моделей для различных ролей в цепочках обработки.

## Установка

### Через Go

```bash
go install github.com/grik-ai/ricochet-task@latest
```

### Бинарные файлы

Скачайте бинарный файл для вашей платформы с [страницы релизов](https://github.com/grik-ai/ricochet-task/releases).

### Из исходного кода

```bash
git clone https://github.com/grik-ai/ricochet-task.git
cd ricochet-task
go build -o ricochet-task main.go
```

## Быстрый старт

### 1. Инициализация проекта

```bash
ricochet init
```

### 2. Добавление API-ключей

```bash
ricochet key add --provider openai --key "sk-your-key"
ricochet key add --provider anthropic --key "sk-ant-your-key"
```

### 3. Настройка моделей

```bash
ricochet models setup
```

### 4. Создание цепочки моделей

```bash
ricochet chain create --name "Анализ документа"
ricochet chain add-model --chain YOUR_CHAIN_ID --name "Анализатор" --type openai --role analyzer --prompt "Проанализируй этот текст и выдели ключевые темы"
ricochet chain add-model --chain YOUR_CHAIN_ID --name "Суммаризатор" --type anthropic --role summarizer --prompt "Создай краткое резюме по каждой теме"
ricochet chain add-model --chain YOUR_CHAIN_ID --name "Интегратор" --type deepseek --role integrator --prompt "Объедини резюме в целостный обзор"
```

### 5. Запуск цепочки

```bash
ricochet chain run --chain YOUR_CHAIN_ID --input-file document.txt
```

## Примеры использования

### Анализ больших документов

```bash
# Создание цепочки для анализа научной статьи
ricochet chain create --name "Анализ научной статьи"
ricochet chain add-model --chain YOUR_CHAIN_ID --name "Извлечение методологии" --type claude --role extractor
ricochet chain add-model --chain YOUR_CHAIN_ID --name "Анализ результатов" --type gpt4 --role analyzer
ricochet chain add-model --chain YOUR_CHAIN_ID --name "Генерация выводов" --type deepseek --role integrator

# Запуск цепочки
ricochet chain run --chain YOUR_CHAIN_ID --input-file paper.pdf
```

### Обработка кодовой базы

```bash
# Создание цепочки для анализа кода
ricochet chain create --name "Анализ кода"
ricochet chain add-model --chain YOUR_CHAIN_ID --name "Анализ структуры" --type deepseek --role analyzer
ricochet chain add-model --chain YOUR_CHAIN_ID --name "Поиск проблем" --type claude --role evaluator
ricochet chain add-model --chain YOUR_CHAIN_ID --name "Генерация рекомендаций" --type gpt4 --role integrator

# Запуск цепочки
ricochet chain run --chain YOUR_CHAIN_ID --input-file "src/**/*.go"
```

## Интеграция с редакторами

### Cursor

Добавьте следующую конфигурацию в файл `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "ricochet-task": {
      "command": "ricochet-task",
      "args": ["mcp"],
      "env": {
        "OPENAI_API_KEY": "YOUR_OPENAI_KEY",
        "ANTHROPIC_API_KEY": "YOUR_ANTHROPIC_KEY",
        "DEEPSEEK_API_KEY": "YOUR_DEEPSEEK_KEY"
      }
    }
  }
}
```

### VS Code

Создайте файл `.vscode/mcp.json` в корне вашего проекта:

```json
{
  "servers": {
    "ricochet-task": {
      "command": "ricochet-task",
      "args": ["mcp"],
      "env": {
        "OPENAI_API_KEY": "YOUR_OPENAI_KEY",
        "ANTHROPIC_API_KEY": "YOUR_ANTHROPIC_KEY",
        "DEEPSEEK_API_KEY": "YOUR_DEEPSEEK_KEY"
      },
      "type": "stdio"
    }
  }
}
```

## Интеграция с Task Master

Ricochet Task совместим с Task Master и может использовать его настройки моделей. Для импорта настроек используйте:

```bash
ricochet models import-from-taskmaster
```

## Команды CLI

### Управление цепочками

- `ricochet chain create` - создание новой цепочки
- `ricochet chain list` - список цепочек
- `ricochet chain add-model` - добавление модели в цепочку
- `ricochet chain run` - запуск цепочки
- `ricochet chain status` - проверка статуса выполнения цепочки

### Управление чекпоинтами

- `ricochet checkpoint list` - список чекпоинтов
- `ricochet checkpoint get` - получение содержимого чекпоинта
- `ricochet checkpoint save` - сохранение чекпоинта
- `ricochet checkpoint delete` - удаление чекпоинта

### Управление API-ключами

- `ricochet key add` - добавление API-ключа
- `ricochet key list` - список ключей
- `ricochet key delete` - удаление ключа
- `ricochet key share` - настройка общего доступа к ключу

### Управление моделями

- `ricochet models setup` - интерактивная настройка моделей
- `ricochet models list` - список настроенных моделей
- `ricochet models reset` - сброс настроек моделей

## Документация

Полная документация доступна в [Wiki проекта](https://github.com/grik-ai/ricochet-task/wiki).

## Лицензия

MIT License

## Благодарности

Проект вдохновлен идеями [Claude Task Master](https://github.com/eyaltoledano/claude-task-master) от [@eyaltoledano](https://twitter.com/eyaltoledano).
