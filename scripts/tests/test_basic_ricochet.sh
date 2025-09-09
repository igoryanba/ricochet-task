#!/bin/bash
# test_basic_ricochet.sh - Тестирование базовых функций Ricochet

echo "🧪 Тестирование базовых функций Ricochet..."

# Переменные для тестирования
TEST_PROJECT_NAME="TestProject_$(date +%s)"
TEST_CONTEXT_NAME="test_context_$(date +%s)"
TEST_TASK_TITLE="Test Task $(date +%s)"

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

# Тест 1: Инициализация
echo "1. Тестирование инициализации..."
if ./ricochet-task init --name "$TEST_PROJECT_NAME" --provider "youtrack" > /dev/null 2>&1; then
    log_result "Инициализация" "success" "Проект создан успешно"
else
    log_result "Инициализация" "error" "Ошибка создания проекта"
    exit 1
fi

# Тест 2: Создание контекста
echo "2. Тестирование создания контекста..."
if ./ricochet-task context create --name "$TEST_CONTEXT_NAME" --project "0-1" --board "0-2" > /dev/null 2>&1; then
    log_result "Создание контекста" "success" "Контекст создан успешно"
else
    log_result "Создание контекста" "error" "Ошибка создания контекста"
fi

# Тест 3: Создание задачи
echo "3. Тестирование создания задачи..."
if ./ricochet-task tasks create --title "$TEST_TASK_TITLE" --type "task" --project "0-1" > /dev/null 2>&1; then
    log_result "Создание задачи" "success" "Задача создана успешно"
else
    log_result "Создание задачи" "error" "Ошибка создания задачи"
fi

# Тест 4: Получение списка задач
echo "4. Тестирование получения списка задач..."
if ./ricochet-task tasks list > /dev/null 2>&1; then
    log_result "Получение списка задач" "success" "Список задач получен успешно"
else
    log_result "Получение списка задач" "error" "Ошибка получения списка задач"
fi

# Тест 5: Проверка статуса
echo "5. Тестирование проверки статуса..."
if ./ricochet-task status > /dev/null 2>&1; then
    log_result "Проверка статуса" "success" "Статус получен успешно"
else
    log_result "Проверка статуса" "error" "Ошибка получения статуса"
fi

# Тест 6: Проверка конфигурации
echo "6. Тестирование проверки конфигурации..."
if ./ricochet-task config validate > /dev/null 2>&1; then
    log_result "Проверка конфигурации" "success" "Конфигурация валидна"
else
    log_result "Проверка конфигурации" "error" "Ошибка конфигурации"
fi

echo "✅ Базовое тестирование Ricochet завершено"
