#!/bin/bash
# run_all_tests.sh - Запуск всех тестов

echo "🚀 Запуск всех тестов Ricochet AI системы..."
echo "================================================"

# Переменные
TEST_DIR="scripts/tests"
LOG_DIR="test_logs"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_FILE="$LOG_DIR/test_results_$TIMESTAMP.log"

# Создание директории для логов
mkdir -p "$LOG_DIR"

# Функция для запуска теста
run_test() {
    local test_name="$1"
    local test_script="$2"
    
    echo ""
    echo "🧪 Запуск теста: $test_name"
    echo "----------------------------------------"
    
    if [ -f "$test_script" ]; then
        if bash "$test_script" 2>&1 | tee -a "$LOG_FILE"; then
            echo "✅ $test_name: ПРОЙДЕН"
            return 0
        else
            echo "❌ $test_name: ПРОВАЛЕН"
            return 1
        fi
    else
        echo "❌ $test_name: СКРИПТ НЕ НАЙДЕН ($test_script)"
        return 1
    fi
}

# Счетчики результатов
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Запуск тестов
echo "Начало тестирования: $(date)"
echo "Лог файл: $LOG_FILE"
echo ""

# Тест 1: Базовое тестирование Ricochet
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if run_test "Базовое тестирование Ricochet" "$TEST_DIR/test_basic_ricochet.sh"; then
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

# Тест 2: MCP сервер
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if run_test "MCP сервер" "$TEST_DIR/test_mcp_server.sh"; then
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

# Тест 3: AI автоматизация
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if run_test "AI автоматизация" "$TEST_DIR/test_ai_automation.sh"; then
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

# Тест 4: Git интеграция
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if run_test "Git интеграция" "$TEST_DIR/test_git_integration.sh"; then
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

# Тест 5: Workflow координация
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if run_test "Workflow координация" "$TEST_DIR/test_workflow_coordination.sh"; then
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

# Итоговые результаты
echo ""
echo "================================================"
echo "📊 ИТОГОВЫЕ РЕЗУЛЬТАТЫ ТЕСТИРОВАНИЯ"
echo "================================================"
echo "Всего тестов: $TOTAL_TESTS"
echo "Пройдено: $PASSED_TESTS"
echo "Провалено: $FAILED_TESTS"
echo "Процент успеха: $((PASSED_TESTS * 100 / TOTAL_TESTS))%"
echo ""
echo "Завершение тестирования: $(date)"
echo "Лог файл: $LOG_FILE"

# Определение статуса
if [ $FAILED_TESTS -eq 0 ]; then
    echo "🎉 ВСЕ ТЕСТЫ ПРОЙДЕНЫ УСПЕШНО!"
    exit 0
else
    echo "⚠️  НЕКОТОРЫЕ ТЕСТЫ ПРОВАЛЕНЫ"
    exit 1
fi
