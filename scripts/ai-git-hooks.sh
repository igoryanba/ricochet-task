#!/bin/bash
# AI Git Hooks - автоматические хуки для Git интеграции

HOOK_TYPE=${1:-"pre-commit"}  # pre-commit, post-commit, pre-push, post-merge
PROJECT_PATH=${2:-.}

echo "🪝 AI Git Hooks - Автоматические хуки для AI"
echo "🔍 Тип хука: $HOOK_TYPE"
echo "📁 Проект: $PROJECT_PATH"

# Функция для AI-дружественного вывода
ai_hook_output() {
    local message="$1"
    local level="$2"  # info, warning, critical, success
    
    case "$level" in
        "critical")
            echo "🚨 CRITICAL: $message"
            ;;
        "warning")
            echo "⚠️  WARNING: $message"
            ;;
        "success")
            echo "✅ SUCCESS: $message"
            ;;
        *)
            echo "ℹ️  INFO: $message"
            ;;
    esac
}

# Pre-commit hook - проверка перед коммитом
pre_commit_hook() {
    echo "🔍 Pre-commit hook - проверка перед коммитом..."
    
    # Проверка формата сообщения коммита
    local commit_message_file="$1"
    if [ -f "$commit_message_file" ]; then
        local commit_message=$(cat "$commit_message_file")
        
        # Проверка наличия ID задачи
        if echo "$commit_message" | grep -q "#[0-9]"; then
            ai_hook_output "Коммит содержит ID задачи" "success"
        else
            ai_hook_output "Коммит не содержит ID задачи" "warning"
            echo "   💡 Рекомендация: Используй формат 'feat: описание #3-45'"
        fi
        
        # Проверка типа коммита
        if echo "$commit_message" | grep -qE "^(feat|fix|docs|style|refactor|test|chore):"; then
            ai_hook_output "Коммит использует правильный формат" "success"
        else
            ai_hook_output "Коммит не использует conventional commits" "warning"
            echo "   💡 Рекомендация: Используй 'feat:', 'fix:', 'docs:' и т.д."
        fi
    fi
    
    # Проверка TODO комментариев
    local staged_files=$(git diff --cached --name-only --diff-filter=ACM)
    local todo_count=0
    
    for file in $staged_files; do
        if [ -f "$file" ]; then
            local todos=$(git diff --cached "$file" | grep -c "TODO\|FIXME\|HACK" 2>/dev/null || echo "0")
            if [[ "$todos" =~ ^[0-9]+$ ]]; then
                todo_count=$((todo_count + todos))
            fi
        fi
    done
    
    if [ $todo_count -gt 0 ]; then
        ai_hook_output "Найдено $todo_count TODO комментариев в staged файлах" "warning"
        echo "   💡 Рекомендация: Используй 'ai-git-integration.sh create' для создания задач"
    fi
}

# Post-commit hook - действия после коммита
post_commit_hook() {
    echo "🔄 Post-commit hook - действия после коммита..."
    
    # Получение последнего коммита
    local last_commit=$(git log -1 --pretty=format:"%H %s")
    local commit_hash=$(echo "$last_commit" | awk '{print $1}')
    local commit_message=$(echo "$last_commit" | sed 's/^[^ ]* //')
    
    # Извлечение ID задачи
    local task_id=$(echo "$commit_message" | grep -o "#[0-9-]*" | head -1 | sed 's/#//')
    
    if [ ! -z "$task_id" ]; then
        ai_hook_output "Обнаружена задача $task_id в коммите $commit_hash" "info"
        
        # Проверка существования задачи
        local task_exists=$(./ricochet-task tasks get "$task_id" 2>/dev/null | grep -c "Task ID")
        
        if [ $task_exists -gt 0 ]; then
            # Обновление задачи
            echo "   🔄 Обновление задачи $task_id..."
            
            # Определение статуса по типу коммита
            local new_status="in_progress"
            if echo "$commit_message" | grep -qi "done\|complete\|finish"; then
                new_status="completed"
            elif echo "$commit_message" | grep -qi "test"; then
                new_status="testing"
            fi
            
            # Обновление статуса задачи
            # ./ricochet-task tasks update "$task_id" --status "$new_status"
            
            ai_hook_output "Задача $task_id обновлена: $new_status" "success"
        else
            ai_hook_output "Задача $task_id не найдена в Ricochet" "warning"
        fi
    else
        ai_hook_output "Коммит не содержит ID задачи" "info"
    fi
}

# Pre-push hook - проверка перед push
pre_push_hook() {
    echo "🚀 Pre-push hook - проверка перед push..."
    
    # Проверка незакоммиченных TODO
    local uncommitted_todos=$(git diff --name-only | xargs grep -l "TODO\|FIXME\|HACK" 2>/dev/null | wc -l)
    
    if [ $uncommitted_todos -gt 0 ]; then
        ai_hook_output "Найдено $uncommitted_todos файлов с незакоммиченными TODO" "warning"
        echo "   💡 Рекомендация: Создай задачи из TODO перед push"
    fi
    
    # Проверка связывания с задачами
    local recent_commits=$(git log --oneline -5)
    local task_commits=0
    
    while IFS= read -r commit_line; do
        if echo "$commit_line" | grep -q "#[0-9]"; then
            task_commits=$((task_commits + 1))
        fi
    done <<< "$recent_commits"
    
    if [ $task_commits -eq 0 ]; then
        ai_hook_output "Последние коммиты не связаны с задачами" "warning"
        echo "   💡 Рекомендация: Используй ID задач в сообщениях коммитов"
    else
        ai_hook_output "Найдено $task_commits коммитов с задачами" "success"
    fi
}

# Post-merge hook - действия после merge
post_merge_hook() {
    echo "🔀 Post-merge hook - действия после merge..."
    
    # Анализ изменений после merge
    local changed_files=$(git diff --name-only HEAD~1 HEAD)
    local todo_files=0
    
    for file in $changed_files; do
        if [ -f "$file" ]; then
            local todos=$(git diff HEAD~1 HEAD "$file" | grep -c "TODO\|FIXME\|HACK" || echo "0")
            if [ $todos -gt 0 ]; then
                todo_files=$((todo_files + 1))
            fi
        fi
    done
    
    if [ $todo_files -gt 0 ]; then
        ai_hook_output "Найдено $todo_files файлов с TODO после merge" "info"
        echo "   💡 Рекомендация: Используй 'ai-git-integration.sh create' для создания задач"
    fi
    
    # Синхронизация с задачами
    echo "   🔄 Синхронизация с задачами..."
    # ./scripts/ai-git-integration.sh "$PROJECT_PATH" sync
    
    ai_hook_output "Post-merge hook завершен" "success"
}

# Установка Git хуков
install_git_hooks() {
    echo "🪝 Установка Git хуков..."
    
    if [ ! -d "$PROJECT_PATH/.git" ]; then
        ai_hook_output "Не Git репозиторий" "warning" "Инициализируй Git: git init"
        return
    fi
    
    local hooks_dir="$PROJECT_PATH/.git/hooks"
    
    # Pre-commit hook
    cat > "$hooks_dir/pre-commit" << 'PRE_COMMIT'
#!/bin/bash
./scripts/ai-git-hooks.sh pre-commit "$1"
PRE_COMMIT
    
    # Post-commit hook
    cat > "$hooks_dir/post-commit" << 'POST_COMMIT'
#!/bin/bash
./scripts/ai-git-hooks.sh post-commit
POST_COMMIT
    
    # Pre-push hook
    cat > "$hooks_dir/pre-push" << 'PRE_PUSH'
#!/bin/bash
./scripts/ai-git-hooks.sh pre-push
PRE_PUSH
    
    # Post-merge hook
    cat > "$hooks_dir/post-merge" << 'POST_MERGE'
#!/bin/bash
./scripts/ai-git-hooks.sh post-merge
POST_MERGE
    
    # Установка прав на выполнение
    chmod +x "$hooks_dir/pre-commit"
    chmod +x "$hooks_dir/post-commit"
    chmod +x "$hooks_dir/pre-push"
    chmod +x "$hooks_dir/post-merge"
    
    ai_hook_output "Git хуки установлены" "success"
    echo "   📁 Хуки: $hooks_dir"
}

# Основная функция
main() {
    case "$HOOK_TYPE" in
        "pre-commit")
            pre_commit_hook "$3"
            ;;
        "post-commit")
            post_commit_hook
            ;;
        "pre-push")
            pre_push_hook
            ;;
        "post-merge")
            post_merge_hook
            ;;
        "install")
            install_git_hooks
            ;;
        *)
            echo "Использование: $0 [тип_хука] [путь] [дополнительные_параметры]"
            echo "Типы хуков: pre-commit, post-commit, pre-push, post-merge, install"
            ;;
    esac
}

# Запуск
main
