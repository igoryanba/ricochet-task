#!/bin/bash
# Система автоматических уведомлений для AI

PROJECT_PATH=${1:-.}
CHECK_INTERVAL=${2:-300}  # 5 минут по умолчанию
NOTIFICATION_LEVEL=${3:-"all"}  # all, critical, warnings

echo "🔔 AI Auto Notifications - Автоматические уведомления для AI"
echo "📁 Проект: $PROJECT_PATH"
echo "⏱️ Интервал: ${CHECK_INTERVAL}с"
echo "📊 Уровень: $NOTIFICATION_LEVEL"

# Файл для хранения состояния
STATE_FILE=".ai-notifications-state.json"

# Инициализация состояния
init_state() {
    if [ ! -f "$STATE_FILE" ]; then
        cat > "$STATE_FILE" << JSON
{
  "last_check": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "notifications_sent": 0,
  "critical_alerts": 0,
  "warning_alerts": 0,
  "last_tasks_count": 0,
  "last_critical_count": 0
}
JSON
    fi
}

# Функция для отправки уведомления AI
send_ai_notification() {
    local message="$1"
    local level="$2"
    local action="$3"
    local timestamp=$(date '+%H:%M:%S')
    
    echo "[$timestamp] 🔔 AI NOTIFICATION [$level]: $message"
    if [ ! -z "$action" ]; then
        echo "[$timestamp] 💡 AI ACTION: $action"
    fi
    
    # Логирование в файл
    echo "[$timestamp] [$level] $message" >> .ai-notifications.log
}

# Функция проверки критических изменений
check_critical_changes() {
    local current_critical=$(./ricochet-task tasks list --priority "critical" --status "open" --limit 10 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    local last_critical=$(jq -r '.last_critical_count' "$STATE_FILE")
    
    if [ "$current_critical" -gt "$last_critical" ]; then
        local new_critical=$((current_critical - last_critical))
        send_ai_notification "Появилось $new_critical новых критических задач (всего: $current_critical)" "CRITICAL" "Используй 'ricochet-task tasks list --priority critical' для просмотра"
        
        # Обновление состояния
        jq --arg count "$current_critical" '.last_critical_count = ($count | tonumber)' "$STATE_FILE" > "$STATE_FILE.tmp" && mv "$STATE_FILE.tmp" "$STATE_FILE"
        return 1
    fi
    
    return 0
}

# Функция проверки новых задач
check_new_tasks() {
    local current_tasks=$(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    local last_tasks=$(jq -r '.last_tasks_count' "$STATE_FILE")
    
    if [ "$current_tasks" -gt "$last_tasks" ]; then
        local new_tasks=$((current_tasks - last_tasks))
        send_ai_notification "Создано $new_tasks новых задач (всего: $current_tasks)" "INFO" "Используй 'ricochet-task tasks list --limit 10' для просмотра последних"
        
        # Обновление состояния
        jq --arg count "$current_tasks" '.last_tasks_count = ($count | tonumber)' "$STATE_FILE" > "$STATE_FILE.tmp" && mv "$STATE_FILE.tmp" "$STATE_FILE"
    fi
}

# Функция проверки блокеров
check_blockers() {
    local blockers=()
    
    # Проверка заблокированных задач
    local blocked_tasks=$(./ricochet-task tasks list --status "blocked" --limit 10 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    if [ "$blocked_tasks" -gt 0 ]; then
        blockers+=("$blocked_tasks заблокированных задач")
    fi
    
    # Проверка задач без исполнителя
    local unassigned_tasks=$(./ricochet-task tasks list --status "open" --limit 50 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | grep -v "admin" | wc -l)
    if [ "$unassigned_tasks" -gt 10 ]; then
        blockers+=("$unassigned_tasks задач без исполнителя")
    fi
    
    if [ ${#blockers[@]} -gt 0 ]; then
        local blocker_message="Обнаружены блокеры: ${blockers[*]}"
        send_ai_notification "$blocker_message" "WARNING" "Рассмотри перераспределение задач или устранение блокеров"
        return 1
    fi
    
    return 0
}

# Функция проверки проблем в коде
check_code_issues() {
    local issues=()
    
    # Поиск критических TODO
    local critical_todos=$(find "$PROJECT_PATH" -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.py" 2>/dev/null | xargs grep -l "FIXME\|HACK\|XXX" 2>/dev/null | wc -l)
    if [ "$critical_todos" -gt 0 ]; then
        issues+=("$critical_todos файлов с критическими TODO")
    fi
    
    # Поиск больших файлов
    local large_files=$(find "$PROJECT_PATH" -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.py" 2>/dev/null | xargs wc -l 2>/dev/null | awk '$1 > 1000 {print $2}' | wc -l)
    if [ "$large_files" -gt 0 ]; then
        issues+=("$large_files файлов больше 1000 строк")
    fi
    
    if [ ${#issues[@]} -gt 0 ]; then
        local issue_message="Проблемы в коде: ${issues[*]}"
        send_ai_notification "$issue_message" "WARNING" "Используй 'scripts/analyze-code-complexity.sh' для детального анализа"
        return 1
    fi
    
    return 0
}

# Функция проверки производительности
check_performance() {
    local total_tasks=$(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    local completed_tasks=$(./ricochet-task tasks list --status "completed" --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    
    if [ $total_tasks -gt 0 ]; then
        local completion_rate=$((completed_tasks * 100 / total_tasks))
        
        if [ $completion_rate -lt 20 ]; then
            send_ai_notification "Низкая скорость завершения задач ($completion_rate%)" "WARNING" "Рассмотри упрощение задач или увеличение команды"
            return 1
        elif [ $completion_rate -gt 90 ]; then
            send_ai_notification "Отличная скорость завершения ($completion_rate%)" "SUCCESS" "Команда работает эффективно"
        fi
    fi
    
    return 0
}

# Функция проверки git статуса
check_git_status() {
    if [ -d ".git" ]; then
        local uncommitted=$(git status --porcelain 2>/dev/null | wc -l)
        if [ "$uncommitted" -gt 50 ]; then
            send_ai_notification "Много несохраненных изменений ($uncommitted файлов)" "WARNING" "Рассмотри коммит изменений"
            return 1
        fi
        
        local unpushed=$(git log --oneline origin/HEAD..HEAD 2>/dev/null | wc -l)
        if [ "$unpushed" -gt 20 ]; then
            send_ai_notification "Много неотправленных коммитов ($unpushed)" "WARNING" "Рассмотри отправку изменений"
            return 1
        fi
    fi
    
    return 0
}

# Основная функция проверки
run_check() {
    local alerts=0
    
    echo "🔍 Выполнение проверки..."
    
    # Проверка критических изменений
    if ! check_critical_changes; then
        alerts=$((alerts + 1))
    fi
    
    # Проверка новых задач
    check_new_tasks
    
    # Проверка блокеров
    if ! check_blockers; then
        alerts=$((alerts + 1))
    fi
    
    # Проверка проблем в коде
    if ! check_code_issues; then
        alerts=$((alerts + 1))
    fi
    
    # Проверка производительности
    if ! check_performance; then
        alerts=$((alerts + 1))
    fi
    
    # Проверка git статуса
    if ! check_git_status; then
        alerts=$((alerts + 1))
    fi
    
    # Обновление времени последней проверки
    jq --arg time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" '.last_check = $time' "$STATE_FILE" > "$STATE_FILE.tmp" && mv "$STATE_FILE.tmp" "$STATE_FILE"
    
    if [ $alerts -eq 0 ]; then
        echo "✅ Все проверки пройдены успешно"
    else
        echo "⚠️ Обнаружено $alerts проблем"
    fi
    
    return $alerts
}

# Функция непрерывного мониторинга
continuous_monitor() {
    echo "🔄 Запуск непрерывного мониторинга..."
    echo "Для остановки нажмите Ctrl+C"
    
    while true; do
        run_check
        sleep "$CHECK_INTERVAL"
    done
}

# Инициализация
init_state

# Запуск в зависимости от параметров
case "$NOTIFICATION_LEVEL" in
    "critical")
        echo "🚨 Режим критических уведомлений"
        check_critical_changes
        check_blockers
        ;;
    "warnings")
        echo "⚠️ Режим предупреждений"
        run_check
        ;;
    "all"|*)
        if [ "$CHECK_INTERVAL" = "0" ]; then
            echo "🔍 Однократная проверка"
            run_check
        else
            continuous_monitor
        fi
        ;;
esac
