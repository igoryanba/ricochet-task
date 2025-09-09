#!/bin/bash
# test_workflow_coordination.sh - Тестирование workflow координации

echo "🧪 Тестирование workflow координации..."

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

# Тест 1: Анализ workflow
echo "1. Тестирование анализа workflow..."
if ./scripts/ai-workflow-orchestrator.sh . analyze > /dev/null 2>&1; then
    log_result "Анализ workflow" "success" "Анализ выполнен"
else
    log_result "Анализ workflow" "error" "Ошибка анализа workflow"
fi

# Тест 2: Координация агентов
echo "2. Тестирование координации агентов..."
if ./scripts/ai-agent-coordinator.sh . assign > /dev/null 2>&1; then
    log_result "Координация агентов" "success" "Координация выполнена"
else
    log_result "Координация агентов" "error" "Ошибка координации агентов"
fi

# Тест 3: Синхронизация агентов
echo "3. Тестирование синхронизации агентов..."
if ./scripts/ai-agent-coordinator.sh . sync > /dev/null 2>&1; then
    log_result "Синхронизация агентов" "success" "Синхронизация выполнена"
else
    log_result "Синхронизация агентов" "error" "Ошибка синхронизации агентов"
fi

# Тест 4: Передача задач между агентами
echo "4. Тестирование передачи задач между агентами..."
if ./scripts/ai-agent-coordinator.sh . handoff coordinator feature-developer 3-45 > /dev/null 2>&1; then
    log_result "Передача задач" "success" "Передача выполнена"
else
    log_result "Передача задач" "error" "Ошибка передачи задач"
fi

# Тест 5: Ревью работы агентов
echo "5. Тестирование ревью работы агентов..."
if ./scripts/ai-agent-coordinator.sh . review > /dev/null 2>&1; then
    log_result "Ревью агентов" "success" "Ревью выполнено"
else
    log_result "Ревью агентов" "error" "Ошибка ревью агентов"
fi

# Тест 6: Выполнение workflow
echo "6. Тестирование выполнения workflow..."
if ./scripts/ai-workflow-orchestrator.sh . execute > /dev/null 2>&1; then
    log_result "Выполнение workflow" "success" "Выполнение завершено"
else
    log_result "Выполнение workflow" "error" "Ошибка выполнения workflow"
fi

# Тест 7: Мониторинг workflow
echo "7. Тестирование мониторинга workflow..."
if ./scripts/ai-workflow-orchestrator.sh . monitor > /dev/null 2>&1; then
    log_result "Мониторинг workflow" "success" "Мониторинг выполнен"
else
    log_result "Мониторинг workflow" "error" "Ошибка мониторинга workflow"
fi

# Тест 8: Разблокировка задач
echo "8. Тестирование разблокировки задач..."
if ./scripts/ai-workflow-orchestrator.sh . unblock > /dev/null 2>&1; then
    log_result "Разблокировка задач" "success" "Разблокировка выполнена"
else
    log_result "Разблокировка задач" "error" "Ошибка разблокировки задач"
fi

# Тест 9: Генерация отчета workflow
echo "9. Тестирование генерации отчета workflow..."
if ./scripts/ai-workflow-orchestrator.sh . report > /dev/null 2>&1; then
    log_result "Отчет workflow" "success" "Отчет сгенерирован"
else
    log_result "Отчет workflow" "error" "Ошибка генерации отчета workflow"
fi

# Тест 10: Генерация отчета координации
echo "10. Тестирование генерации отчета координации..."
if ./scripts/ai-agent-coordinator.sh . report > /dev/null 2>&1; then
    log_result "Отчет координации" "success" "Отчет сгенерирован"
else
    log_result "Отчет координации" "error" "Ошибка генерации отчета координации"
fi

echo "✅ Тестирование workflow координации завершено"
