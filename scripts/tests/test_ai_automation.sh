#!/bin/bash
# test_ai_automation.sh - Тестирование AI автоматизации

echo "🧪 Тестирование AI автоматизации..."

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

# Тест 1: Определение типа проекта
echo "1. Тестирование определения типа проекта..."
if ./scripts/detect-project-type.sh . > /dev/null 2>&1; then
    log_result "Определение типа проекта" "success" "Тип проекта определен"
else
    log_result "Определение типа проекта" "error" "Ошибка определения типа проекта"
fi

# Проверка создания файла с информацией о проекте
if [ -f ".ricochet-project-info.json" ]; then
    log_result "Создание файла проекта" "success" "Файл .ricochet-project-info.json создан"
else
    log_result "Создание файла проекта" "error" "Файл .ricochet-project-info.json не создан"
fi

# Тест 2: Создание контекста из папки
echo "2. Тестирование создания контекста из папки..."
if ./scripts/create-context-from-folder.sh . > /dev/null 2>&1; then
    log_result "Создание контекста из папки" "success" "Контекст создан из папки"
else
    log_result "Создание контекста из папки" "error" "Ошибка создания контекста из папки"
fi

# Тест 3: Синхронизация с редактором
echo "3. Тестирование синхронизации с редактором..."
if ./scripts/sync-with-editor.sh . > /dev/null 2>&1; then
    log_result "Синхронизация с редактором" "success" "Синхронизация выполнена"
else
    log_result "Синхронизация с редактором" "error" "Ошибка синхронизации с редактором"
fi

# Тест 4: Анализ сложности кода
echo "4. Тестирование анализа сложности кода..."
if ./scripts/analyze-code-complexity.sh . quick > /dev/null 2>&1; then
    log_result "Анализ сложности кода" "success" "Анализ кода выполнен"
else
    log_result "Анализ сложности кода" "error" "Ошибка анализа кода"
fi

# Тест 5: Предложения по рефакторингу
echo "5. Тестирование предложений по рефакторингу..."
if ./scripts/suggest-refactoring.sh . quick > /dev/null 2>&1; then
    log_result "Предложения по рефакторингу" "success" "Предложения сгенерированы"
else
    log_result "Предложения по рефакторингу" "error" "Ошибка генерации предложений"
fi

# Тест 6: Умный мониторинг
echo "6. Тестирование умного мониторинга..."
if ./scripts/ai-smart-monitor.sh . quick > /dev/null 2>&1; then
    log_result "Умный мониторинг" "success" "Мониторинг выполнен"
else
    log_result "Умный мониторинг" "error" "Ошибка мониторинга"
fi

# Тест 7: Автоматические уведомления
echo "7. Тестирование автоматических уведомлений..."
if ./scripts/ai-auto-notifications.sh . quick > /dev/null 2>&1; then
    log_result "Автоматические уведомления" "success" "Уведомления обработаны"
else
    log_result "Автоматические уведомления" "error" "Ошибка обработки уведомлений"
fi

# Тест 8: Анализ проекта
echo "8. Тестирование анализа проекта..."
if ./scripts/ai-project-analyzer.sh . quick > /dev/null 2>&1; then
    log_result "Анализ проекта" "success" "Анализ проекта выполнен"
else
    log_result "Анализ проекта" "error" "Ошибка анализа проекта"
fi

# Тест 9: Анализ команды
echo "9. Тестирование анализа команды..."
if ./scripts/ai-team-manager.sh . analyze > /dev/null 2>&1; then
    log_result "Анализ команды" "success" "Анализ команды выполнен"
else
    log_result "Анализ команды" "error" "Ошибка анализа команды"
fi

# Тест 10: Анализ навыков
echo "10. Тестирование анализа навыков..."
if ./scripts/ai-skill-analyzer.sh . quick > /dev/null 2>&1; then
    log_result "Анализ навыков" "success" "Анализ навыков выполнен"
else
    log_result "Анализ навыков" "error" "Ошибка анализа навыков"
fi

echo "✅ Тестирование AI автоматизации завершено"
