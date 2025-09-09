#!/bin/bash
# Скрипт для синхронизации контекста с редактором

echo "🔄 Синхронизация контекста с редактором..."

# Проверка, что мы в директории с ricochet-task
if [ ! -f "./ricochet-task" ]; then
    echo "❌ Ricochet Task не найден. Запустите скрипт из директории ricochet-task"
    exit 1
fi

# Проверка существования контекста
if [ ! -f ".ricochet-context.json" ]; then
    echo "❌ Контекст не найден. Сначала создайте контекст командой:"
    echo "   ./scripts/create-context-from-folder.sh"
    exit 1
fi

# Загрузка информации о контексте
CONTEXT_ID=$(jq -r '.context_id' .ricochet-context.json)
CONTEXT_NAME=$(jq -r '.context_name' .ricochet-context.json)
PROJECT_TYPE=$(jq -r '.project_type' .ricochet-context.json)

echo "📊 Текущий контекст: $CONTEXT_NAME ($CONTEXT_ID)"
echo "📁 Тип проекта: $PROJECT_TYPE"

# Переключение на контекст
echo "🔄 Переключение на контекст..."
./ricochet-task context switch "$CONTEXT_ID"

# Создание файла для отслеживания открытых файлов
echo "📝 Создание файла отслеживания..."
cat > .ricochet-editor-sync.json << JSON
{
  "context_id": "$CONTEXT_ID",
  "project_type": "$PROJECT_TYPE",
  "last_sync": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "tracked_files": [],
  "active_tasks": []
}
JSON

# Функция для отслеживания изменений файлов
track_file_changes() {
    echo "👀 Отслеживание изменений файлов..."
    
    # Поиск недавно измененных файлов
    local recent_files=$(find . -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.py" -o -name "*.rs" -o -name "*.java" | head -10)
    
    for file in $recent_files; do
        echo "📄 Отслеживаем файл: $file"
        
        # Анализ файла на предмет TODO комментариев
        local todo_count=$(grep -c "TODO\|FIXME\|HACK" "$file" 2>/dev/null || echo "0")
        if [ "$todo_count" -gt 0 ]; then
            echo "  ⚠️ Найдено $todo_count TODO комментариев"
        fi
        
        # Анализ сложности (простая эвристика)
        local line_count=$(wc -l < "$file" 2>/dev/null || echo "0")
        if [ "$line_count" -gt 100 ]; then
            echo "  📊 Файл большой ($line_count строк) - возможен рефакторинг"
        fi
    done
}

# Функция для показа связанных задач
show_related_tasks() {
    echo "📋 Связанные задачи:"
    
    # Получение списка задач
    local tasks=$(./ricochet-task tasks list --limit 5)
    echo "$tasks"
    
    # Анализ активных задач
    local active_tasks=$(./ricochet-task tasks list --status "in_progress" --limit 3)
    if [ ! -z "$active_tasks" ]; then
        echo "🔄 Активные задачи:"
        echo "$active_tasks"
    fi
}

# Функция для показа блокеров
show_blockers() {
    echo "🚨 Проверка блокеров..."
    
    # Поиск критических задач
    local critical_tasks=$(./ricochet-task tasks list --priority "critical" --status "open" --limit 5)
    if [ ! -z "$critical_tasks" ]; then
        echo "⚠️ Критические задачи:"
        echo "$critical_tasks"
    fi
    
    # Поиск просроченных задач
    local overdue_tasks=$(./ricochet-task tasks list --overdue --limit 5)
    if [ ! -z "$overdue_tasks" ]; then
        echo "⏰ Просроченные задачи:"
        echo "$overdue_tasks"
    fi
}

# Основная функция синхронизации
sync_with_editor() {
    echo "🔄 Синхронизация с редактором..."
    
    # Отслеживание изменений файлов
    track_file_changes
    
    # Показ связанных задач
    show_related_tasks
    
    # Проверка блокеров
    show_blockers
    
    # Обновление времени синхронизации
    jq --arg time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" '.last_sync = $time' .ricochet-editor-sync.json > .ricochet-editor-sync.tmp && mv .ricochet-editor-sync.tmp .ricochet-editor-sync.json
    
    echo "✅ Синхронизация завершена"
}

# Запуск синхронизации
sync_with_editor

echo "🎉 Синхронизация с редактором завершена!"
echo "Контекст готов для работы с AI."
