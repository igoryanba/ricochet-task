# Руководство по интеграции Ricochet с редакторами кода

## Оглавление

1. [Введение](#введение)
2. [Интеграция с VS Code](#интеграция-с-vs-code)
3. [Интеграция с Cursor](#интеграция-с-cursor)
4. [Интеграция с JetBrains IDEs](#интеграция-с-jetbrains-ides)
5. [Создание собственных расширений](#создание-собственных-расширений)
6. [Устранение неполадок](#устранение-неполадок)

## Введение

Ricochet предоставляет возможность интеграции с популярными редакторами кода, что позволяет использовать цепочки моделей ИИ непосредственно в процессе разработки. Это руководство описывает процесс настройки и использования Ricochet с различными редакторами кода.

## Интеграция с VS Code

### Установка расширения

1. Откройте VS Code
2. Перейдите в раздел Extensions (Ctrl+Shift+X)
3. Найдите "Ricochet Task" и нажмите "Install"

Альтернативно, расширение можно установить из VSIX-файла:

1. Скачайте VSIX-файл из репозитория Ricochet
2. В VS Code выберите View -> Command Palette
3. Введите "Extensions: Install from VSIX" и выберите скачанный файл

### Настройка расширения

После установки необходимо настроить расширение:

1. Откройте настройки VS Code (File -> Preferences -> Settings)
2. Найдите раздел "Ricochet Task"
3. Укажите следующие параметры:
   - **API Endpoint**: URL Ricochet API (по умолчанию `http://localhost:8080`)
   - **API Key**: Ваш API-ключ для доступа к Ricochet API
   - **Default Chain**: ID цепочки по умолчанию

Пример файла конфигурации `.vscode/settings.json`:

```json
{
  "ricochet.apiEndpoint": "http://localhost:8080",
  "ricochet.apiKey": "your-api-key",
  "ricochet.defaultChain": "chain-123",
  "ricochet.defaultMode": "text"
}
```

### Использование в VS Code

После настройки расширения вы можете использовать Ricochet в VS Code:

1. **Контекстное меню**: Правый клик на коде -> "Process with Ricochet"
2. **Командная палитра**: Ctrl+Shift+P -> "Ricochet: Process Selection"
3. **Горячие клавиши**: 
   - `Ctrl+Alt+R`: Обработать выделенный текст
   - `Ctrl+Alt+C`: Выбрать цепочку для обработки

### Настройка пользовательских команд

Вы можете добавить пользовательские команды для конкретных цепочек:

1. Откройте файл `.vscode/settings.json`
2. Добавьте секцию `ricochet.commands`:

```json
"ricochet.commands": [
  {
    "title": "Analyze Code",
    "command": "ricochet.process",
    "chainId": "code-analysis-chain",
    "chainName": "Code Analysis"
  },
  {
    "title": "Generate Tests",
    "command": "ricochet.process",
    "chainId": "test-generator-chain",
    "chainName": "Test Generator"
  }
]
```

## Интеграция с Cursor

### Установка и настройка

1. Убедитесь, что установлена последняя версия Cursor
2. Создайте файл конфигурации Ricochet для Cursor:
   - Создайте директорию `.cursor` в корне вашего проекта
   - Создайте файл `.cursor/mcp.json` со следующим содержимым:

```json
{
  "chains": [
    {
      "id": "code-analysis-chain",
      "name": "Code Analysis",
      "description": "Analyze code structure and patterns"
    },
    {
      "id": "test-generator-chain",
      "name": "Test Generator",
      "description": "Generate tests for selected code"
    }
  ],
  "defaultMode": "chain"
}
```

3. Установите переменные окружения в Cursor:
   - **RICOCHET_API_ENDPOINT**: URL Ricochet API
   - **RICOCHET_API_KEY**: Ваш API-ключ

### Использование в Cursor

После настройки вы можете использовать Ricochet в Cursor:

1. Выделите код, который хотите обработать
2. Вызовите меню MCP (Alt+M или Cmd+M на macOS)
3. Выберите "Run with Ricochet"
4. Выберите нужную цепочку из списка

## Интеграция с JetBrains IDEs

### Установка плагина

1. Откройте IDE JetBrains (IntelliJ IDEA, PyCharm, WebStorm и т.д.)
2. Перейдите в File -> Settings -> Plugins
3. Найдите "Ricochet Task" и нажмите "Install"

Альтернативно, плагин можно установить из ZIP-файла:

1. Скачайте ZIP-файл из репозитория Ricochet
2. В IDE выберите File -> Settings -> Plugins
3. Нажмите на значок шестеренки и выберите "Install Plugin from Disk"
4. Выберите скачанный ZIP-файл

### Настройка плагина

После установки необходимо настроить плагин:

1. Перейдите в File -> Settings -> Tools -> Ricochet Task
2. Укажите следующие параметры:
   - **API Endpoint**: URL Ricochet API
   - **API Key**: Ваш API-ключ
   - **Default Chain**: ID цепочки по умолчанию

### Использование в JetBrains IDEs

После настройки плагина вы можете использовать Ricochet в IDE:

1. **Контекстное меню**: Правый клик на коде -> "Ricochet" -> "Process Selection"
2. **Меню**: Tools -> Ricochet -> "Process Selection"
3. **Горячие клавиши**:
   - `Ctrl+Alt+R`: Обработать выделенный текст
   - `Ctrl+Alt+C`: Выбрать цепочку для обработки

## Создание собственных расширений

Если вы хотите создать собственное расширение для редактора, не имеющего официальной поддержки, вы можете использовать Ricochet API напрямую.

### Основные принципы интеграции

1. **Аутентификация**: Используйте API-ключ для аутентификации запросов к Ricochet API
2. **Выбор цепочки**: Предоставьте пользователю возможность выбрать цепочку моделей
3. **Обработка текста**: Отправляйте выделенный текст на обработку через API
4. **Отображение результатов**: Отображайте результаты обработки в удобном для пользователя виде

### Пример интеграции с использованием API

```javascript
// Пример интеграции на JavaScript
async function processWithRicochet(text, chainId) {
  const apiUrl = 'http://localhost:8080/api/v1/chains/' + chainId + '/run';
  const apiKey = 'your-api-key';
  
  const response = await fetch(apiUrl, {
    method: 'POST',
    headers: {
      'Authorization': 'Bearer ' + apiKey,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      text: text,
      options: {
        max_parallel_chunks: 4,
        max_tokens_per_chunk: 2000,
        segmentation_method: 'semantic',
        save_checkpoints: true
      }
    })
  });
  
  const data = await response.json();
  return data.runId; // ID запуска для последующего получения результатов
}

// Получение результатов
async function getRicochetResults(runId) {
  const apiUrl = 'http://localhost:8080/api/v1/runs/' + runId + '/results';
  const apiKey = 'your-api-key';
  
  const response = await fetch(apiUrl, {
    method: 'GET',
    headers: {
      'Authorization': 'Bearer ' + apiKey
    }
  });
  
  return await response.json();
}
```

## Устранение неполадок

### Общие проблемы

#### Ошибка аутентификации

**Проблема**: API-запросы возвращают ошибку аутентификации.

**Решение**:
- Убедитесь, что API-ключ указан правильно в настройках
- Проверьте срок действия API-ключа
- Убедитесь, что у ключа есть необходимые права доступа

#### Невозможно подключиться к API

**Проблема**: Не удается подключиться к Ricochet API.

**Решение**:
- Убедитесь, что Ricochet API сервер запущен
- Проверьте, что указан правильный URL API
- Проверьте настройки брандмауэра и сетевые подключения

#### Ошибки при обработке текста

**Проблема**: Возникают ошибки при обработке текста.

**Решение**:
- Проверьте, что выбранная цепочка моделей существует
- Убедитесь, что размер обрабатываемого текста не превышает допустимые лимиты
- Проверьте логи Ricochet API для получения подробной информации об ошибке

### Получение поддержки

Если вы столкнулись с проблемами при использовании Ricochet с редакторами кода, вы можете получить поддержку:

1. Создайте issue на GitHub: [https://github.com/grik-ai/ricochet-task/issues](https://github.com/grik-ai/ricochet-task/issues)
2. Обратитесь к документации: [https://grik-ai.github.io/ricochet-task/docs](https://grik-ai.github.io/ricochet-task/docs)
3. Напишите в сообщество Ricochet: [https://grik-ai.github.io/community](https://grik-ai.github.io/community)

---

Это руководство предоставляет основную информацию по интеграции Ricochet с редакторами кода. Для получения дополнительной информации обратитесь к исходному коду и документации отдельных расширений. 