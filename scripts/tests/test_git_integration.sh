#!/bin/bash
# test_git_integration.sh - Тестирование Git интеграции

echo "🧪 Тестирование Git интеграции..."

# Функция для логирования результатов
log_result() {
    local test_name="$1"
    local result="$2"
    local message="$3"
    
    if [ "$result" = "success" ]; then
        echo "   ✅ $test_name: $message"
    else
        echo "   ❌ $test_name: $message"
    fi
}

# Тест 1: Проверка Git репозитория
echo "1. Проверка Git репозитория..."
if [ -d ".git" ]; then
    log_result "Git репозиторий" "success" "Репозиторий найден"
else
    log_result "Git репозиторий" "warning" "Репозиторий не найден, инициализация..."
    if git init > /dev/null 2>&1; then
        log_result "Инициализация Git" "success" "Репозиторий инициализирован"
    else
        log_result "Инициализация Git" "error" "Ошибка инициализации репозитория"
    fi
fi

# Тест 2: Анализ Git репозитория
echo "2. Тестирование анализа Git репозитория..."
if ./scripts/ai-git-integration.sh . analyze > /dev/null 2>&1; then
    log_result "Анализ Git" "success" "Анализ репозитория выполнен"
else
    log_result "Анализ Git" "error" "Ошибка анализа репозитория"
fi

# Тест 3: Установка Git хуков
echo "3. Тестирование установки Git хуков..."
if ./scripts/ai-git-hooks.sh install . > /dev/null 2>&1; then
    log_result "Установка Git хуков" "success" "Хуки установлены"
else
    log_result "Установка Git хуков" "error" "Ошибка установки хуков"
fi

# Тест 4: Проверка установленных хуков
echo "4. Проверка установленных хуков..."
if [ -f ".git/hooks/pre-commit" ]; then
    log_result "Pre-commit хук" "success" "Хук установлен"
else
    log_result "Pre-commit хук" "error" "Хук не установлен"
fi

if [ -f ".git/hooks/post-commit" ]; then
    log_result "Post-commit хук" "success" "Хук установлен"
else
    log_result "Post-commit хук" "error" "Хук не установлен"
fi

if [ -f ".git/hooks/pre-push" ]; then
    log_result "Pre-push хук" "success" "Хук установлен"
else
    log_result "Pre-push хук" "error" "Хук не установлен"
fi

if [ -f ".git/hooks/post-merge" ]; then
    log_result "Post-merge хук" "success" "Хук установлен"
else
    log_result "Post-merge хук" "error" "Хук не установлен"
fi

# Тест 5: Проверка прав доступа хуков
echo "5. Проверка прав доступа хуков..."
if [ -x ".git/hooks/pre-commit" ]; then
    log_result "Права pre-commit" "success" "Хук исполняемый"
else
    log_result "Права pre-commit" "error" "Хук не исполняемый"
fi

# Тест 6: Создание задач из TODO
echo "6. Тестирование создания задач из TODO..."
if ./scripts/ai-git-integration.sh . create > /dev/null 2>&1; then
    log_result "Создание задач из TODO" "success" "Задачи созданы"
else
    log_result "Создание задач из TODO" "error" "Ошибка создания задач"
fi

# Тест 7: Синхронизация коммитов
echo "7. Тестирование синхронизации коммитов..."
if ./scripts/ai-git-integration.sh . sync > /dev/null 2>&1; then
    log_result "Синхронизация коммитов" "success" "Синхронизация выполнена"
else
    log_result "Синхронизация коммитов" "error" "Ошибка синхронизации"
fi

# Тест 8: Обновление статусов
echo "8. Тестирование обновления статусов..."
if ./scripts/ai-git-integration.sh . update > /dev/null 2>&1; then
    log_result "Обновление статусов" "success" "Статусы обновлены"
else
    log_result "Обновление статусов" "error" "Ошибка обновления статусов"
fi

# Тест 9: Генерация отчета
echo "9. Тестирование генерации отчета..."
if ./scripts/ai-git-integration.sh . report > /dev/null 2>&1; then
    log_result "Генерация отчета" "success" "Отчет сгенерирован"
else
    log_result "Генерация отчета" "error" "Ошибка генерации отчета"
fi

echo "✅ Тестирование Git интеграции завершено"
