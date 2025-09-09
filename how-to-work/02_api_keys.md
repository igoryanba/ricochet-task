# 🔐 Управление API-ключами

Ricochet Task использует API-ключи для подключения к различным сервисам: OpenAI, Anthropic Claude, DeepSeek, а также к системам управления задачами как YouTrack и Jira.

## 🎯 Поддерживаемые провайдеры

### AI-модели:
- **OpenAI** - GPT-4, GPT-3.5-turbo
- **Anthropic** - Claude 3.5 Sonnet, Claude 3 Haiku  
- **DeepSeek** - DeepSeek Coder, DeepSeek Chat
- **Grok** - Grok-1, Grok-1.5

### Системы задач:
- **YouTrack** - JetBrains YouTrack
- **Jira** - Atlassian Jira
- **Notion** - Notion API
- **Linear** - Linear API

## 📝 Добавление API-ключей

### OpenAI (рекомендуется для начала)

```bash
# Добавление OpenAI ключа
./ricochet-task key add --provider openai --key "sk-proj-ваш-ключ-здесь"

# С лимитом токенов (необязательно)
./ricochet-task key add --provider openai --key "sk-proj-ваш-ключ" --limit 1000000

# С общим доступом (для команды)
./ricochet-task key add --provider openai --key "sk-proj-ваш-ключ" --shared
```

### Anthropic Claude

```bash
# Добавление Claude ключа
./ricochet-task key add --provider anthropic --key "sk-ant-ваш-ключ"

# Проверка работы
./ricochet-task key list | grep anthropic
```

### DeepSeek

```bash
# DeepSeek API ключ
./ricochet-task key add --provider deepseek --key "ваш-deepseek-ключ"
```

### YouTrack (система задач)

```bash
# YouTrack permanent token
./ricochet-task key add --provider youtrack --key "perm-ваш-токен"

# Примечание: YouTrack токены также настраиваются в ricochet.yaml
```

## 📋 Просмотр и управление ключами

### Список всех ключей

```bash
# Просмотр всех ключей (замаскированы для безопасности)
./ricochet-task key list
```

**Пример вывода:**
```
Список API-ключей:
----------------------------------------------------
ID: 1fef8e35-e6d2-41fe-a8fb-1b16c074e56b
Провайдер: openai
Ключ: sk-proj..._abc
Создан: 2025-09-06T08:50:34+05:00
Общий доступ: false
Использовано токенов: 0
Лимит использования: не ограничен
----------------------------------------------------
```

### Обновление существующих ключей

```bash
# Обновление ключа (найдется автоматически по провайдеру)
./ricochet-task key update --provider openai --key "новый-ключ"

# Обновление лимита
./ricochet-task key update --provider openai --limit 2000000
```

### Удаление ключей

```bash
# Удаление ключа по провайдеру
./ricochet-task key delete --provider openai

# Удаление по ID (если несколько ключей одного провайдера)
./ricochet-task key delete --id "1fef8e35-e6d2-41fe-a8fb-1b16c074e56b"
```

## 🔒 Безопасность ключей

### Шифрование и хранение

- Все ключи **зашифрованы** перед сохранением
- Хранятся в `~/.ricochet/keys/` (или `$RICOCHET_CONFIG_DIR`)
- При выводе автоматически **маскируются** (показывают только начало и конец)

### Общий доступ к ключам

```bash
# Включение общего доступа (для команд)
./ricochet-task key share --provider openai --enable

# Отключение общего доступа
./ricochet-task key share --provider openai --disable

# Просмотр настроек общего доступа
./ricochet-task key list | grep "Общий доступ"
```

### Лимиты использования

```bash
# Установка лимита токенов
./ricochet-task key update --provider openai --limit 500000

# Сброс счетчика использования
./ricochet-task key reset-usage --provider openai

# Просмотр использования
./ricochet-task key list | grep "Использовано токенов"
```

## 🔧 Получение API-ключей

### OpenAI

1. Перейдите на https://platform.openai.com/api-keys
2. Нажмите "Create new secret key"
3. Скопируйте ключ (начинается с `sk-proj-` или `sk-`)
4. Добавьте в Ricochet Task

### Anthropic Claude

1. Перейдите на https://console.anthropic.com/
2. В разделе "API Keys" создайте новый ключ
3. Ключ начинается с `sk-ant-`
4. Добавьте в систему

### DeepSeek

1. Зарегистрируйтесь на https://platform.deepseek.com/
2. Получите API ключ в разделе API
3. Добавьте в Ricochet Task

### YouTrack

1. В YouTrack перейдите в Settings → API Keys
2. Создайте "Permanent Token"
3. Скопируйте токен (начинается с `perm-`)
4. Добавьте как `--provider youtrack`

## ⚠️ Частые проблемы и решения

### Проблема: "Invalid API key"

```bash
# Проверьте формат ключа
./ricochet-task key list

# Удалите и добавьте заново
./ricochet-task key delete --provider openai
./ricochet-task key add --provider openai --key "правильный-ключ"
```

### Проблема: "Quota exceeded"

```bash
# Проверьте использование
./ricochet-task key list | grep "Использовано токенов"

# Увеличьте лимит или добавьте новый ключ
./ricochet-task key update --provider openai --limit 2000000
```

### Проблема: "Provider not supported"

```bash
# Список поддерживаемых провайдеров
./ricochet-task key add --help

# Убедитесь в правильности названия:
# openai, anthropic, deepseek, grok, youtrack, jira
```

## 🎯 Рекомендации по настройке

### Для индивидуального использования:

```bash
# Минимальный набор
./ricochet-task key add --provider openai --key "ваш-openai-ключ"
./ricochet-task key add --provider youtrack --key "ваш-youtrack-токен"
```

### Для команды разработчиков:

```bash
# Набор для AI-моделей
./ricochet-task key add --provider openai --key "ключ-1" --shared
./ricochet-task key add --provider anthropic --key "ключ-2" --shared
./ricochet-task key add --provider deepseek --key "ключ-3" --shared

# Системы задач
./ricochet-task key add --provider youtrack --key "команда-токен" --shared
```

### Для продакшена:

```bash
# С лимитами и мониторингом
./ricochet-task key add --provider openai --key "prod-ключ" --limit 10000000
./ricochet-task key add --provider anthropic --key "backup-ключ" --limit 5000000

# Регулярная проверка использования
./ricochet-task key list | grep "Использовано"
```

## ✅ Проверка настройки

После добавления ключей проверьте работу:

```bash
# 1. Список всех ключей
./ricochet-task key list

# 2. Проверка провайдеров задач
./ricochet-task providers list

# 3. Тестовый запрос (если настроен MCP)
./ricochet-task mcp start --port 3001 &
curl -s http://localhost:3001/tools | jq '.tools | length'
```

Если все команды работают без ошибок - ключи настроены правильно! 🎉

---

**Следующий шаг**: Переходите к [Работе с провайдерами](./03_providers.md) для настройки YouTrack, Jira и других систем.