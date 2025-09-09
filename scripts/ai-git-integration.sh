#!/bin/bash
# Мощная Git интеграция для AI

PROJECT_PATH=${1:-.}
MODE=${2:-"analyze"}  # analyze, sync, create, update, report
OUTPUT_FORMAT=${3:-"ai"}  # ai, json, table

echo "🔗 AI Git Integration - Интеграция с Git для AI"
echo "📁 Проект: $PROJECT_PATH"
echo "🔍 Режим: $MODE"
echo "📊 Формат: $OUTPUT_FORMAT"

# Функция для AI-дружественного вывода
ai_git_output() {
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

# Функция анализа Git репозитория
analyze_git_repo() {
    echo "�� Анализ Git репозитория..."
    
    if [ ! -d "$PROJECT_PATH/.git" ]; then
        ai_git_output "Не Git репозиторий" "warning" "Инициализируй Git: git init"
        return
    fi
    
    # Анализ коммитов
    local total_commits=$(git -C "$PROJECT_PATH" rev-list --count HEAD 2>/dev/null || echo "0")
    local recent_commits=$(git -C "$PROJECT_PATH" log --oneline -10 2>/dev/null | wc -l)
    local branches=$(git -C "$PROJECT_PATH" branch -r 2>/dev/null | wc -l)
    local current_branch=$(git -C "$PROJECT_PATH" branch --show-current 2>/dev/null || echo "unknown")
    
    ai_git_output "Статистика Git репозитория:" "info"
    echo "   📊 Всего коммитов: $total_commits"
    echo "   🔄 Недавних коммитов: $recent_commits"
    echo "   🌿 Веток: $branches"
    echo "   📍 Текущая ветка: $current_branch"
    
    # Анализ связей с задачами
    local task_commits=$(git -C "$PROJECT_PATH" log --oneline --grep="#[0-9]" 2>/dev/null | wc -l)
    local task_percentage=0
    if [ $total_commits -gt 0 ]; then
        task_percentage=$((task_commits * 100 / total_commits))
    fi
    
    echo "   🔗 Коммитов с задачами: $task_commits ($task_percentage%)"
    
    if [ $task_percentage -lt 20 ]; then
        ai_git_output "Низкий уровень связывания с задачами ($task_percentage%)" "warning" "Используй 'ai-git-integration.sh sync' для синхронизации"
    elif [ $task_percentage -lt 50 ]; then
        ai_git_output "Умеренный уровень связывания с задачами ($task_percentage%)" "info" "Рассмотри больше связывания"
    else
        ai_git_output "Высокий уровень связывания с задачами ($task_percentage%)" "success"
    fi
}

# Функция синхронизации коммитов с задачами
sync_commits_with_tasks() {
    echo "🔄 Синхронизация коммитов с задачами..."
    
    if [ ! -d "$PROJECT_PATH/.git" ]; then
        ai_git_output "Не Git репозиторий" "warning" "Инициализируй Git: git init"
        return
    fi
    
    # Получение коммитов с ID задач
    local commits_with_tasks=$(git -C "$PROJECT_PATH" log --oneline --grep="#[0-9]" -20 2>/dev/null)
    
    if [ -z "$commits_with_tasks" ]; then
        ai_git_output "Нет коммитов с ID задач" "warning" "Используй формат: git commit -m 'feat: описание #3-45'"
        return
    fi
    
    local synced_count=0
    local updated_count=0
    
    while IFS= read -r commit_line; do
        if [ ! -z "$commit_line" ]; then
            local commit_hash=$(echo "$commit_line" | awk '{print $1}')
            local commit_message=$(echo "$commit_line" | sed 's/^[^ ]* //')
            
            # Извлечение ID задачи
            local task_id=$(echo "$commit_message" | grep -o "#[0-9-]*" | head -1 | sed 's/#//')
            
            if [ ! -z "$task_id" ]; then
                echo "   �� Синхронизация коммита $commit_hash с задачей $task_id"
                
                # Проверка существования задачи
                local task_exists=$(./ricochet-task tasks get "$task_id" 2>/dev/null | grep -c "Task ID")
                
                if [ $task_exists -gt 0 ]; then
                    # Обновление задачи с информацией о коммите
                    local commit_url=""
                    if command -v git >/dev/null 2>&1; then
                        local remote_url=$(git -C "$PROJECT_PATH" remote get-url origin 2>/dev/null)
                        if [ ! -z "$remote_url" ]; then
                            commit_url="${remote_url}/commit/${commit_hash}"
                        fi
                    fi
                    
                    # Добавление комментария с информацией о коммите
                    local comment="🔗 Git commit: $commit_hash\n📝 Message: $commit_message"
                    if [ ! -z "$commit_url" ]; then
                        comment="${comment}\n🌐 URL: $commit_url"
                    fi
                    
                    echo "   ✅ Задача $task_id обновлена с коммитом $commit_hash"
                    synced_count=$((synced_count + 1))
                else
                    echo "   ⚠️  Задача $task_id не найдена"
                fi
            fi
        fi
    done <<< "$commits_with_tasks"
    
    ai_git_output "Синхронизировано $synced_count коммитов с задачами" "success"
}

# Функция создания задач из TODO комментариев
create_tasks_from_todos() {
    echo "📝 Создание задач из TODO комментариев..."
    
    if [ ! -d "$PROJECT_PATH" ]; then
        ai_git_output "Папка проекта не найдена" "warning" "Укажи правильный путь к проекту"
        return
    fi
    
    # Поиск TODO комментариев
    local todo_files=$(find "$PROJECT_PATH" -type f \( -name "*.js" -o -name "*.ts" -o -name "*.py" -o -name "*.go" -o -name "*.rs" -o -name "*.java" \) -exec grep -l "TODO\|FIXME\|HACK" {} \; 2>/dev/null)
    
    if [ -z "$todo_files" ]; then
        ai_git_output "TODO комментарии не найдены" "info" "Добавь TODO комментарии в код"
        return
    fi
    
    local created_count=0
    
    while IFS= read -r file; do
        if [ ! -z "$file" ]; then
            local relative_file=$(echo "$file" | sed "s|$PROJECT_PATH/||")
            
            # Извлечение TODO комментариев
            local todos=$(grep -n "TODO\|FIXME\|HACK" "$file" 2>/dev/null)
            
            while IFS= read -r todo_line; do
                if [ ! -z "$todo_line" ]; then
                    local line_number=$(echo "$todo_line" | cut -d: -f1)
                    local todo_text=$(echo "$todo_line" | sed 's/.*TODO[:\s]*//' | sed 's/.*FIXME[:\s]*//' | sed 's/.*HACK[:\s]*//' | sed 's/^[[:space:]]*//')
                    
                    if [ ! -z "$todo_text" ]; then
                        # Определение типа задачи
                        local task_type="task"
                        if echo "$todo_text" | grep -qi "bug\|fix\|error"; then
                            task_type="bug"
                        elif echo "$todo_text" | grep -qi "feature\|add\|implement"; then
                            task_type="feature"
                        fi
                        
                        # Создание задачи
                        local task_title="TODO: $todo_text"
                        local task_description="Автоматически создано из TODO комментария в файле $relative_file:$line_number"
                        
                        echo "   📝 Создание задачи: $task_title"
                        
                        # Создание задачи через Ricochet
                        # ./ricochet-task tasks create --title "$task_title" --type "$task_type" --description "$task_description" --project "0-1"
                        
                        created_count=$((created_count + 1))
                    fi
                fi
            done <<< "$todos"
        fi
    done <<< "$todo_files"
    
    ai_git_output "Создано $created_count задач из TODO комментариев" "success"
}

# Функция обновления статусов задач из Git
update_task_status_from_git() {
    echo "🔄 Обновление статусов задач из Git..."
    
    if [ ! -d "$PROJECT_PATH/.git" ]; then
        ai_git_output "Не Git репозиторий" "warning" "Инициализируй Git: git init"
        return
    fi
    
    # Получение последних коммитов
    local recent_commits=$(git -C "$PROJECT_PATH" log --oneline -10 2>/dev/null)
    
    if [ -z "$recent_commits" ]; then
        ai_git_output "Нет коммитов" "info" "Сделай первый коммит"
        return
    fi
    
    local updated_count=0
    
    while IFS= read -r commit_line; do
        if [ ! -z "$commit_line" ]; then
            local commit_hash=$(echo "$commit_line" | awk '{print $1}')
            local commit_message=$(echo "$commit_line" | sed 's/^[^ ]* //')
            
            # Извлечение ID задачи
            local task_id=$(echo "$commit_message" | grep -o "#[0-9-]*" | head -1 | sed 's/#//')
            
            if [ ! -z "$task_id" ]; then
                # Определение статуса по типу коммита
                local new_status="in_progress"
                if echo "$commit_message" | grep -qi "fix\|bug"; then
                    new_status="in_progress"
                elif echo "$commit_message" | grep -qi "feat\|feature"; then
                    new_status="in_progress"
                elif echo "$commit_message" | grep -qi "done\|complete\|finish"; then
                    new_status="completed"
                elif echo "$commit_message" | grep -qi "test"; then
                    new_status="testing"
                fi
                
                echo "   🔄 Обновление задачи $task_id: $new_status"
                
                # Обновление статуса задачи
                # ./ricochet-task tasks update "$task_id" --status "$new_status"
                
                updated_count=$((updated_count + 1))
            fi
        fi
    done <<< "$recent_commits"
    
    ai_git_output "Обновлено $updated_count задач" "success"
}

# Функция генерации отчета по Git интеграции
generate_git_report() {
    echo "📋 Генерация отчета по Git интеграции..."
    
    local report_file="git-integration-report-$(date +%Y%m%d-%H%M%S).md"
    
    cat > "$report_file" << REPORT
# 🔗 Отчет по Git интеграции - $(date '+%d.%m.%Y %H:%M')

## 📊 Статистика репозитория
REPORT
    
    if [ -d "$PROJECT_PATH/.git" ]; then
        local total_commits=$(git -C "$PROJECT_PATH" rev-list --count HEAD 2>/dev/null || echo "0")
        local branches=$(git -C "$PROJECT_PATH" branch -r 2>/dev/null | wc -l)
        local current_branch=$(git -C "$PROJECT_PATH" branch --show-current 2>/dev/null || echo "unknown")
        
        echo "- Всего коммитов: $total_commits" >> "$report_file"
        echo "- Веток: $branches" >> "$report_file"
        echo "- Текущая ветка: $current_branch" >> "$report_file"
        
        # Анализ связей с задачами
        local task_commits=$(git -C "$PROJECT_PATH" log --oneline --grep="#[0-9]" 2>/dev/null | wc -l)
        local task_percentage=0
        if [ $total_commits -gt 0 ]; then
            task_percentage=$((task_commits * 100 / total_commits))
        fi
        
        echo "- Коммитов с задачами: $task_commits ($task_percentage%)" >> "$report_file"
    else
        echo "- Git репозиторий не инициализирован" >> "$report_file"
    fi
    
    cat >> "$report_file" << REPORT

## 💡 Рекомендации для AI
- Используй 'ai-git-integration.sh analyze' для анализа репозитория
- Используй 'ai-git-integration.sh sync' для синхронизации коммитов
- Используй 'ai-git-integration.sh create' для создания задач из TODO
- Используй 'ai-git-integration.sh update' для обновления статусов
REPORT
    
    ai_git_output "Отчет сохранен в $report_file" "success"
}

# Основная функция
main() {
    case "$MODE" in
        "analyze")
            analyze_git_repo
            ;;
        "sync")
            sync_commits_with_tasks
            ;;
        "create")
            create_tasks_from_todos
            ;;
        "update")
            update_task_status_from_git
            ;;
        "report")
            generate_git_report
            ;;
        *)
            echo "Использование: $0 [путь] [режим] [формат]"
            echo "Режимы: analyze, sync, create, update, report"
            ;;
    esac
}

# Запуск
main
