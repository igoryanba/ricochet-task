#!/bin/bash
# Скрипт для создания контекста из папки проекта

PROJECT_PATH=${1:-.}
CONTEXT_NAME=${2:-"Auto Context"}

echo "🎯 Создание контекста из папки: $PROJECT_PATH"

# Переход в директорию проекта
cd "$PROJECT_PATH" || exit 1

# Проверка, что мы в директории с ricochet-task
if [ ! -f "../ricochet-task" ] && [ ! -f "./ricochet-task" ]; then
    echo "❌ Ricochet Task не найден. Запустите скрипт из директории ricochet-task или из подпапки проекта"
    exit 1
fi

# Определение пути к ricochet-task
if [ -f "./ricochet-task" ]; then
    RICOCHET_CMD="./ricochet-task"
else
    RICOCHET_CMD="../ricochet-task"
fi

# Анализ проекта
echo "🔍 Анализ проекта..."
if [ -f ".ricochet-project-info.json" ]; then
    echo "📄 Используем существующую информацию о проекте"
    PROJECT_TYPE=$(jq -r '.project_type' .ricochet-project-info.json)
    TEAM_SIZE=$(jq -r '.team_size' .ricochet-project-info.json)
    PRIORITY=$(jq -r '.priority' .ricochet-project-info.json)
    TIMELINE_DAYS=$(jq -r '.timeline_days' .ricochet-project-info.json)
else
    echo "🔍 Запуск анализа проекта..."
    if [ -f "./scripts/detect-project-type.sh" ]; then
        ./scripts/detect-project-type.sh
    elif [ -f "../scripts/detect-project-type.sh" ]; then
        ../scripts/detect-project-type.sh
    else
        echo "❌ Скрипт анализа проекта не найден"
        exit 1
    fi
    
    PROJECT_TYPE=$(jq -r '.project_type' .ricochet-project-info.json)
    TEAM_SIZE=$(jq -r '.team_size' .ricochet-project-info.json)
    PRIORITY=$(jq -r '.priority' .ricochet-project-info.json)
    TIMELINE_DAYS=$(jq -r '.timeline_days' .ricochet-project-info.json)
fi

# Получение информации о Git
GIT_URL=$(git config --get remote.origin.url 2>/dev/null || echo "")
PROJECT_NAME=$(basename "$(pwd)")

# Создание контекста
echo "📝 Создание контекста..."
CONTEXT_DESCRIPTION="Автоматически созданный контекст для проекта $PROJECT_NAME ($PROJECT_TYPE)"

# Создание контекста через CLI
echo "Создаем контекст: $CONTEXT_NAME"
$RICOCHET_CMD context create \
    --name "$CONTEXT_NAME" \
    --description "$CONTEXT_DESCRIPTION" \
    --project-id "0-1" \
    --board-id "2" \
    --provider "gamesdrop-youtrack" \
    --assignee "admin" \
    --type "$PROJECT_TYPE" \
    --team-size "$TEAM_SIZE" \
    --priority "$PRIORITY" \
    --timeline "$TIMELINE_DAYS"

if [ $? -eq 0 ]; then
    echo "✅ Контекст создан успешно"
    
    # Получение ID созданного контекста
    CONTEXT_ID=$($RICOCHET_CMD context list | grep "$CONTEXT_NAME" | head -1 | awk '{print $1}')
    
    if [ ! -z "$CONTEXT_ID" ]; then
        echo "🔄 Переключение на новый контекст: $CONTEXT_ID"
        $RICOCHET_CMD context switch "$CONTEXT_ID"
    fi
    
    # Создание конфигурационного файла
    cat > .ricochet-context.json << JSON
{
  "context_id": "$CONTEXT_ID",
  "context_name": "$CONTEXT_NAME",
  "project_name": "$PROJECT_NAME",
  "project_type": "$PROJECT_TYPE",
  "team_size": $TEAM_SIZE,
  "priority": "$PRIORITY",
  "timeline_days": $TIMELINE_DAYS,
  "git_url": "$GIT_URL",
  "created_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "project_path": "$(pwd)"
}
JSON

    echo "💾 Конфигурация контекста сохранена в .ricochet-context.json"

    # Показ информации о созданном контексте
    echo "📊 Информация о контексте:"
    $RICOCHET_CMD context current

    echo "🎉 Контекст успешно создан и активирован!"
    echo "Теперь AI может работать с этим проектом автоматически."
else
    echo "❌ Не удалось создать контекст"
    exit 1
fi
