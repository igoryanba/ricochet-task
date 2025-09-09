#!/bin/bash
# Мощная система интеллектуальных уведомлений для AI

PROJECT_PATH=${1:-.}
MODE=${2:-"full"}  # full, quick, critical
OUTPUT_FORMAT=${3:-"ai"}  # ai, json, table

echo "🧠 AI Smart Monitor - Интеллектуальный мониторинг для AI"
echo "📁 Проект: $PROJECT_PATH"
echo "🔍 Режим: $MODE"
echo "�� Формат: $OUTPUT_FORMAT"

# Функция для AI-дружественного вывода
ai_output() {
    local message="$1"
    local level="$2"  # info, warning, critical, success
    local action="$3"  # optional action for AI
    
    case "$level" in
        "critical")
            echo "🚨 CRITICAL: $message"
            if [ ! -z "$action" ]; then
                echo "   💡 AI Action: $action"
            fi
            ;;
        "warning")
            echo "⚠️  WARNING: $message"
            if [ ! -z "$action" ]; then
                echo "   💡 AI Action: $action"
            fi
            ;;
        "success")
            echo "✅ SUCCESS: $message"
            ;;
        *)
            echo "ℹ️  INFO: $message"
            ;;
    esac
}

# Функция анализа блокеров
analyze_blockers() {
    local blockers=()
    local critical_count=0
    local warning_count=0
    
    echo "🔍 Анализ блокеров..."
    
    # Проверка критических задач
    local critical_tasks=$(./ricochet-task tasks list --priority "critical" --status "open" --limit 10 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    if [ "$critical_tasks" -gt 0 ]; then
        blockers+=("$critical_tasks критических задач требуют внимания")
        critical_count=$((critical_count + critical_tasks))
    fi
    
    # Проверка просроченных задач (если поддерживается)
    local overdue_tasks=$(./ricochet-task tasks list --status "open" --limit 50 2>/dev/null | grep -i "overdue\|просроч" | wc -l)
    if [ "$overdue_tasks" -gt 0 ]; then
        blockers+=("$overdue_tasks просроченных задач")
        warning_count=$((warning_count + overdue_tasks))
    fi
    
    # Проверка заблокированных задач
    local blocked_tasks=$(./ricochet-task tasks list --status "blocked" --limit 10 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    if [ "$blocked_tasks" -gt 0 ]; then
        blockers+=("$blocked_tasks заблокированных задач")
        warning_count=$((warning_count + blocked_tasks))
    fi
    
    # Проверка задач без исполнителя
    local unassigned_tasks=$(./ricochet-task tasks list --status "open" --limit 50 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | grep -v "admin" | wc -l)
    if [ "$unassigned_tasks" -gt 5 ]; then
        blockers+=("$unassigned_tasks задач без исполнителя")
        warning_count=$((warning_count + 1))
    fi
    
    # Вывод результатов для AI
    if [ ${#blockers[@]} -gt 0 ]; then
        ai_output "Обнаружены блокеры:" "warning"
        for blocker in "${blockers[@]}"; do
            echo "   • $blocker"
        done
        
        # Предложения для AI
        if [ $critical_count -gt 0 ]; then
            ai_output "Критические задачи требуют немедленного внимания" "critical" "Используй 'ricochet-task tasks list --priority critical' для просмотра"
        fi
        
        if [ $warning_count -gt 0 ]; then
            ai_output "Есть задачи, требующие внимания" "warning" "Рассмотри возможность перераспределения нагрузки"
        fi
    else
        ai_output "Блокеров не обнаружено" "success"
    fi
    
    return $critical_count
}

# Функция анализа производительности команды
analyze_team_performance() {
    echo "👥 Анализ производительности команды..."
    
    # Получение статистики задач
    local total_tasks=$(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    local completed_tasks=$(./ricochet-task tasks list --status "completed" --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    local in_progress_tasks=$(./ricochet-task tasks list --status "in_progress" --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    
    # Расчет метрик
    local completion_rate=0
    if [ $total_tasks -gt 0 ]; then
        completion_rate=$((completed_tasks * 100 / total_tasks))
    fi
    
    local active_rate=0
    if [ $total_tasks -gt 0 ]; then
        active_rate=$((in_progress_tasks * 100 / total_tasks))
    fi
    
    # Анализ для AI
    ai_output "Статистика команды:" "info"
    echo "   📊 Всего задач: $total_tasks"
    echo "   ✅ Завершено: $completed_tasks ($completion_rate%)"
    echo "   🔄 В работе: $in_progress_tasks ($active_rate%)"
    
    # Предупреждения для AI
    if [ $completion_rate -lt 30 ]; then
        ai_output "Низкая скорость завершения задач ($completion_rate%)" "warning" "Рассмотри упрощение задач или увеличение команды"
    elif [ $completion_rate -gt 80 ]; then
        ai_output "Высокая скорость завершения ($completion_rate%)" "success"
    fi
    
    if [ $active_rate -lt 10 ]; then
        ai_output "Мало активных задач ($active_rate%)" "warning" "Возможно, команда перегружена или задачи слишком сложные"
    elif [ $active_rate -gt 50 ]; then
        ai_output "Много активных задач ($active_rate%)" "warning" "Возможна перегрузка команды"
    fi
}

# Функция анализа кода на предмет проблем
analyze_code_issues() {
    echo "🔍 Анализ кода на предмет проблем..."
    
    local issues=()
    local critical_issues=0
    
    # Поиск критических TODO
    local critical_todos=$(find "$PROJECT_PATH" -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.py" 2>/dev/null | xargs grep -l "FIXME\|HACK\|XXX" 2>/dev/null | wc -l)
    if [ "$critical_todos" -gt 0 ]; then
        issues+=("$critical_todos файлов с критическими TODO")
        critical_issues=$((critical_issues + critical_todos))
    fi
    
    # Поиск больших файлов
    local large_files=$(find "$PROJECT_PATH" -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.py" 2>/dev/null | xargs wc -l 2>/dev/null | awk '$1 > 500 {print $2}' | wc -l)
    if [ "$large_files" -gt 0 ]; then
        issues+=("$large_files файлов больше 500 строк")
    fi
    
    # Поиск дублирования
    local duplicate_files=$(find "$PROJECT_PATH" -name "*.js" -o -name "*.ts" 2>/dev/null | head -10 | xargs -I {} sh -c 'echo "{}:$(sort "{}" | uniq -d | wc -l)"' 2>/dev/null | awk -F: '$2 > 10 {print $1}' | wc -l)
    if [ "$duplicate_files" -gt 0 ]; then
        issues+=("$duplicate_files файлов с возможным дублированием")
    fi
    
    # Вывод для AI
    if [ ${#issues[@]} -gt 0 ]; then
        ai_output "Обнаружены проблемы в коде:" "warning"
        for issue in "${issues[@]}"; do
            echo "   • $issue"
        done
        
        if [ $critical_issues -gt 0 ]; then
            ai_output "Критические проблемы в коде требуют внимания" "critical" "Используй 'scripts/analyze-code-complexity.sh' для детального анализа"
        fi
    else
        ai_output "Проблем в коде не обнаружено" "success"
    fi
}

# Функция анализа рисков проекта
analyze_project_risks() {
    echo "⚠️ Анализ рисков проекта..."
    
    local risks=()
    
    # Проверка размера проекта
    local project_size=$(find "$PROJECT_PATH" -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.py" 2>/dev/null | wc -l)
    if [ "$project_size" -gt 100 ]; then
        risks+=("Большой проект ($project_size файлов) - сложность управления")
    fi
    
    # Проверка зависимостей
    if [ -f "package.json" ]; then
        local dependencies=$(grep -c '"dependencies"' package.json 2>/dev/null || echo "0")
        if [ "$dependencies" -gt 0 ]; then
            local dep_count=$(grep -A 20 '"dependencies"' package.json | grep -c '".*":' 2>/dev/null || echo "0")
            if [ "$dep_count" -gt 50 ]; then
                risks+=("Много зависимостей ($dep_count) - риск конфликтов")
            fi
        fi
    fi
    
    # Проверка git статуса
    if [ -d ".git" ]; then
        local uncommitted=$(git status --porcelain 2>/dev/null | wc -l)
        if [ "$uncommitted" -gt 20 ]; then
            risks+=("Много несохраненных изменений ($uncommitted файлов)")
        fi
        
        local unpushed=$(git log --oneline origin/HEAD..HEAD 2>/dev/null | wc -l)
        if [ "$unpushed" -gt 10 ]; then
            risks+=("Много неотправленных коммитов ($unpushed)")
        fi
    fi
    
    # Вывод для AI
    if [ ${#risks[@]} -gt 0 ]; then
        ai_output "Обнаружены риски проекта:" "warning"
        for risk in "${risks[@]}"; do
            echo "   • $risk"
        done
    else
        ai_output "Рисков проекта не обнаружено" "success"
    fi
}

# Функция генерации рекомендаций для AI
generate_ai_recommendations() {
    echo "💡 Генерация рекомендаций для AI..."
    
    local recommendations=()
    
    # Рекомендации на основе анализа
    local critical_tasks=$(./ricochet-task tasks list --priority "critical" --status "open" --limit 5 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    if [ "$critical_tasks" -gt 0 ]; then
        recommendations+=("Сосредоточься на $critical_tasks критических задачах")
    fi
    
    local todo_count=$(find "$PROJECT_PATH" -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.py" 2>/dev/null | xargs grep -c "TODO\|FIXME" 2>/dev/null | awk '{sum+=$1} END {print sum+0}')
    if [ "$todo_count" -gt 10 ]; then
        recommendations+=("Обрати внимание на $todo_count TODO комментариев в коде")
    fi
    
    local large_files=$(find "$PROJECT_PATH" -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.py" 2>/dev/null | xargs wc -l 2>/dev/null | awk '$1 > 500 {print $2}' | wc -l)
    if [ "$large_files" -gt 0 ]; then
        recommendations+=("Рассмотри рефакторинг $large_files больших файлов")
    fi
    
    # Вывод рекомендаций
    if [ ${#recommendations[@]} -gt 0 ]; then
        ai_output "Рекомендации для AI:" "info"
        for rec in "${recommendations[@]}"; do
            echo "   💡 $rec"
        done
    else
        ai_output "Проект в хорошем состоянии" "success"
    fi
}

# Основная функция мониторинга
main_monitor() {
    local start_time=$(date +%s)
    
    echo "🚀 Запуск AI Smart Monitor..."
    echo "=========================================="
    
    # Анализ блокеров
    analyze_blockers
    local blocker_status=$?
    
    echo ""
    
    # Анализ производительности команды
    analyze_team_performance
    
    echo ""
    
    # Анализ проблем в коде
    analyze_code_issues
    
    echo ""
    
    # Анализ рисков проекта
    analyze_project_risks
    
    echo ""
    
    # Генерация рекомендаций
    generate_ai_recommendations
    
    echo ""
    echo "=========================================="
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    ai_output "Мониторинг завершен за ${duration}с" "success"
    
    # Возврат статуса для AI
    if [ $blocker_status -gt 0 ]; then
        return 1  # Есть критические проблемы
    else
        return 0  # Все в порядке
    fi
}

# Запуск в зависимости от режима
case "$MODE" in
    "quick")
        echo "⚡ Быстрый режим мониторинга"
        analyze_blockers
        ;;
    "critical")
        echo "🚨 Критический режим мониторинга"
        analyze_blockers
        analyze_code_issues
        ;;
    "full"|*)
        main_monitor
        ;;
esac

exit $?
