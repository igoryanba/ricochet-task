#!/bin/bash
# Скрипт для автоматического определения типа проекта

PROJECT_PATH=${1:-.}

echo "🔍 Анализ проекта в директории: $PROJECT_PATH"

# Переход в директорию проекта
cd "$PROJECT_PATH" || exit 1

# Функция определения типа проекта
detect_project_type() {
    local project_type="unknown"
    local team_size=1
    local priority="medium"
    local timeline_days=14
    
    # Node.js проект
    if [ -f "package.json" ]; then
        project_type="nodejs"
        team_size=$(jq -r '.team_size // 2' package.json 2>/dev/null || echo "2")
        priority=$(jq -r '.priority // "medium"' package.json 2>/dev/null || echo "medium")
        echo "📦 Обнаружен Node.js проект"
    fi
    
    # Go проект
    if [ -f "go.mod" ]; then
        project_type="golang"
        team_size=2
        priority="high"
        echo "🐹 Обнаружен Go проект"
    fi
    
    # Python проект
    if [ -f "requirements.txt" ] || [ -f "pyproject.toml" ] || [ -f "setup.py" ]; then
        project_type="python"
        team_size=1
        priority="medium"
        echo "🐍 Обнаружен Python проект"
    fi
    
    # Rust проект
    if [ -f "Cargo.toml" ]; then
        project_type="rust"
        team_size=1
        priority="high"
        echo "🦀 Обнаружен Rust проект"
    fi
    
    # Java проект
    if [ -f "pom.xml" ] || [ -f "build.gradle" ]; then
        project_type="java"
        team_size=3
        priority="medium"
        echo "☕ Обнаружен Java проект"
    fi
    
    # React/Next.js проект
    if [ -f "package.json" ] && grep -q "react\|next" package.json; then
        project_type="react"
        team_size=2
        priority="medium"
        echo "⚛️ Обнаружен React/Next.js проект"
    fi
    
    # Vue.js проект
    if [ -f "package.json" ] && grep -q "vue" package.json; then
        project_type="vue"
        team_size=2
        priority="medium"
        echo "💚 Обнаружен Vue.js проект"
    fi
    
    # Angular проект
    if [ -f "package.json" ] && grep -q "angular" package.json; then
        project_type="angular"
        team_size=3
        priority="medium"
        echo "🅰️ Обнаружен Angular проект"
    fi
    
    # Docker проект
    if [ -f "Dockerfile" ] || [ -f "docker-compose.yml" ]; then
        project_type="docker"
        team_size=2
        priority="high"
        echo "🐳 Обнаружен Docker проект"
    fi
    
    # Git репозиторий
    if [ -d ".git" ]; then
        local git_url=$(git config --get remote.origin.url 2>/dev/null || echo "")
        echo "📁 Git репозиторий: $git_url"
    fi
    
    # Вывод результата
    echo "📊 Результат анализа:"
    echo "  Тип проекта: $project_type"
    echo "  Размер команды: $team_size"
    echo "  Приоритет: $priority"
    echo "  Временные рамки: $timeline_days дней"
    
    # Сохранение в JSON для использования другими скриптами
    cat > .ricochet-project-info.json << JSON
{
  "project_type": "$project_type",
  "team_size": $team_size,
  "priority": "$priority",
  "timeline_days": $timeline_days,
  "detected_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "project_path": "$(pwd)"
}
JSON
    
    echo "💾 Информация сохранена в .ricochet-project-info.json"
}

# Запуск анализа
detect_project_type
