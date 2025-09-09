#!/bin/bash
# Мощный AI-анализатор проекта

PROJECT_PATH=${1:-.}
ANALYSIS_DEPTH=${2:-"deep"}  # quick, deep, full
OUTPUT_FORMAT=${3:-"ai"}  # ai, json, table

echo "🧠 AI Project Analyzer - Мощный анализ проекта для AI"
echo "📁 Проект: $PROJECT_PATH"
echo "🔍 Глубина: $ANALYSIS_DEPTH"
echo "📊 Формат: $OUTPUT_FORMAT"

# Функция для AI-дружественного вывода
ai_analyze() {
    local category="$1"
    local message="$2"
    local score="$3"  # 1-10
    local recommendation="$4"
    
    echo "📊 $category: $message"
    echo "   🎯 Оценка: $score/10"
    if [ ! -z "$recommendation" ]; then
        echo "   💡 Рекомендация: $recommendation"
    fi
    echo ""
}

# Анализ архитектуры проекта
analyze_architecture() {
    echo "🏗️ Анализ архитектуры проекта..."
    
    local score=5
    local issues=()
    local strengths=()
    
    # Анализ структуры папок
    local has_src=$(find "$PROJECT_PATH" -type d -name "src" | wc -l)
    local has_docs=$(find "$PROJECT_PATH" -type d -name "docs" -o -name "doc" | wc -l)
    local has_tests=$(find "$PROJECT_PATH" -type d -name "test" -o -name "tests" -o -name "__tests__" | wc -l)
    local has_config=$(find "$PROJECT_PATH" -name "*.json" -o -name "*.yaml" -o -name "*.yml" -o -name "*.toml" | wc -l)
    
    if [ "$has_src" -gt 0 ]; then
        strengths+=("Есть папка src")
        score=$((score + 1))
    else
        issues+=("Нет папки src")
    fi
    
    if [ "$has_docs" -gt 0 ]; then
        strengths+=("Есть документация")
        score=$((score + 1))
    else
        issues+=("Нет документации")
    fi
    
    if [ "$has_tests" -gt 0 ]; then
        strengths+=("Есть тесты")
        score=$((score + 1))
    else
        issues+=("Нет тестов")
    fi
    
    if [ "$has_config" -gt 0 ]; then
        strengths+=("Есть конфигурация")
        score=$((score + 1))
    else
        issues+=("Нет конфигурации")
    fi
    
    # Анализ файлов
    local total_files=$(find "$PROJECT_PATH" -type f \( -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.py" -o -name "*.rs" -o -name "*.java" \) | wc -l)
    local avg_file_size=$(find "$PROJECT_PATH" -type f \( -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.py" \) -exec wc -l {} + 2>/dev/null | tail -1 | awk '{print $1}' | awk '{print $1/'$total_files'}' 2>/dev/null || echo "0")
    
    if [ "$total_files" -gt 10 ]; then
        strengths+=("Достаточно файлов ($total_files)")
        score=$((score + 1))
    else
        issues+=("Мало файлов ($total_files)")
    fi
    
    # Вывод анализа
    local message="Архитектура проекта"
    if [ ${#strengths[@]} -gt 0 ]; then
        message+=" (сильные стороны: ${strengths[*]})"
    fi
    if [ ${#issues[@]} -gt 0 ]; then
        message+=" (проблемы: ${issues[*]})"
    fi
    
    local recommendation=""
    if [ $score -lt 6 ]; then
        recommendation="Рассмотри улучшение структуры проекта: добавь папки src, docs, tests"
    elif [ $score -gt 8 ]; then
        recommendation="Отличная архитектура! Продолжай в том же духе"
    fi
    
    ai_analyze "Архитектура" "$message" "$score" "$recommendation"
}

# Анализ качества кода
analyze_code_quality() {
    echo "🔍 Анализ качества кода..."
    
    local score=5
    local issues=()
    local strengths=()
    
    # Анализ TODO комментариев
    local todo_count=$(find "$PROJECT_PATH" -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.py" 2>/dev/null | xargs grep -c "TODO\|FIXME\|HACK" 2>/dev/null | awk '{sum+=$1} END {print sum+0}')
    local file_count=$(find "$PROJECT_PATH" -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.py" 2>/dev/null | wc -l)
    local todo_ratio=0
    if [ "$file_count" -gt 0 ]; then
        todo_ratio=$((todo_count * 100 / file_count))
    fi
    
    if [ "$todo_ratio" -lt 5 ]; then
        strengths+=("Мало TODO ($todo_ratio%)")
        score=$((score + 2))
    elif [ "$todo_ratio" -gt 20 ]; then
        issues+=("Много TODO ($todo_ratio%)")
        score=$((score - 2))
    fi
    
    # Анализ больших файлов
    local large_files=$(find "$PROJECT_PATH" -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.py" 2>/dev/null | xargs wc -l 2>/dev/null | awk '$1 > 500 {print $2}' | wc -l)
    if [ "$large_files" -eq 0 ]; then
        strengths+=("Нет больших файлов")
        score=$((score + 2))
    else
        issues+=("$large_files больших файлов")
        score=$((score - 1))
    fi
    
    # Анализ дублирования
    local duplicate_files=$(find "$PROJECT_PATH" -name "*.js" -o -name "*.ts" 2>/dev/null | head -10 | xargs -I {} sh -c 'echo "{}:$(sort "{}" | uniq -d | wc -l)"' 2>/dev/null | awk -F: '$2 > 10 {print $1}' | wc -l)
    if [ "$duplicate_files" -eq 0 ]; then
        strengths+=("Нет дублирования")
        score=$((score + 1))
    else
        issues+=("$duplicate_files файлов с дублированием")
        score=$((score - 1))
    fi
    
    # Вывод анализа
    local message="Качество кода"
    if [ ${#strengths[@]} -gt 0 ]; then
        message+=" (сильные стороны: ${strengths[*]})"
    fi
    if [ ${#issues[@]} -gt 0 ]; then
        message+=" (проблемы: ${issues[*]})"
    fi
    
    local recommendation=""
    if [ $score -lt 6 ]; then
        recommendation="Улучши качество кода: исправь TODO, разбей большие файлы, устрани дублирование"
    elif [ $score -gt 8 ]; then
        recommendation="Отличное качество кода! Продолжай поддерживать высокие стандарты"
    fi
    
    ai_analyze "Качество кода" "$message" "$score" "$recommendation"
}

# Анализ производительности команды
analyze_team_performance() {
    echo "👥 Анализ производительности команды..."
    
    local score=5
    
    # Получение статистики задач
    local total_tasks=$(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    local completed_tasks=$(./ricochet-task tasks list --status "completed" --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    local in_progress_tasks=$(./ricochet-task tasks list --status "in_progress" --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    local critical_tasks=$(./ricochet-task tasks list --priority "critical" --status "open" --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    
    # Расчет метрик
    local completion_rate=0
    if [ $total_tasks -gt 0 ]; then
        completion_rate=$((completed_tasks * 100 / total_tasks))
    fi
    
    local active_rate=0
    if [ $total_tasks -gt 0 ]; then
        active_rate=$((in_progress_tasks * 100 / total_tasks))
    fi
    
    local critical_rate=0
    if [ $total_tasks -gt 0 ]; then
        critical_rate=$((critical_tasks * 100 / total_tasks))
    fi
    
    # Оценка производительности
    if [ $completion_rate -gt 70 ]; then
        score=$((score + 2))
    elif [ $completion_rate -lt 30 ]; then
        score=$((score - 2))
    fi
    
    if [ $active_rate -gt 20 ] && [ $active_rate -lt 60 ]; then
        score=$((score + 1))
    elif [ $active_rate -gt 80 ]; then
        score=$((score - 1))
    fi
    
    if [ $critical_rate -lt 10 ]; then
        score=$((score + 1))
    elif [ $critical_rate -gt 30 ]; then
        score=$((score - 2))
    fi
    
    # Вывод анализа
    local message="Производительность команды (завершено: $completion_rate%, в работе: $active_rate%, критических: $critical_rate%)"
    
    local recommendation=""
    if [ $score -lt 6 ]; then
        recommendation="Улучши производительность: упрости задачи, увеличь команду, решай критические проблемы"
    elif [ $score -gt 8 ]; then
        recommendation="Отличная производительность! Команда работает эффективно"
    fi
    
    ai_analyze "Производительность команды" "$message" "$score" "$recommendation"
}

# Анализ рисков проекта
analyze_project_risks() {
    echo "⚠️ Анализ рисков проекта..."
    
    local score=8
    local risks=()
    
    # Анализ размера проекта
    local project_size=$(find "$PROJECT_PATH" -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.py" 2>/dev/null | wc -l)
    if [ "$project_size" -gt 200 ]; then
        risks+=("Большой проект ($project_size файлов)")
        score=$((score - 1))
    fi
    
    # Анализ зависимостей
    if [ -f "package.json" ]; then
        local dep_count=$(grep -A 20 '"dependencies"' package.json | grep -c '".*":' 2>/dev/null || echo "0")
        if [ "$dep_count" -gt 100 ]; then
            risks+=("Много зависимостей ($dep_count)")
            score=$((score - 1))
        fi
    fi
    
    # Анализ git статуса
    if [ -d ".git" ]; then
        local uncommitted=$(git status --porcelain 2>/dev/null | wc -l)
        if [ "$uncommitted" -gt 100 ]; then
            risks+=("Много несохраненных изменений ($uncommitted)")
            score=$((score - 1))
        fi
        
        local unpushed=$(git log --oneline origin/HEAD..HEAD 2>/dev/null | wc -l)
        if [ "$unpushed" -gt 50 ]; then
            risks+=("Много неотправленных коммитов ($unpushed)")
            score=$((score - 1))
        fi
    fi
    
    # Анализ критических задач
    local critical_tasks=$(./ricochet-task tasks list --priority "critical" --status "open" --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    if [ "$critical_tasks" -gt 5 ]; then
        risks+=("Много критических задач ($critical_tasks)")
        score=$((score - 2))
    fi
    
    # Вывод анализа
    local message="Риски проекта"
    if [ ${#risks[@]} -gt 0 ]; then
        message+=" (риски: ${risks[*]})"
    else
        message+=" (рисков не обнаружено)"
    fi
    
    local recommendation=""
    if [ $score -lt 6 ]; then
        recommendation="Высокие риски! Устрани критические проблемы, упрости проект"
    elif [ $score -gt 8 ]; then
        recommendation="Низкие риски! Проект в хорошем состоянии"
    fi
    
    ai_analyze "Риски проекта" "$message" "$score" "$recommendation"
}

# Генерация общих рекомендаций для AI
generate_ai_recommendations() {
    echo "💡 Генерация рекомендаций для AI..."
    
    local recommendations=()
    
    # Рекомендации на основе анализа
    local critical_tasks=$(./ricochet-task tasks list --priority "critical" --status "open" --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    if [ "$critical_tasks" -gt 0 ]; then
        recommendations+=("Сосредоточься на $critical_tasks критических задачах")
    fi
    
    local todo_count=$(find "$PROJECT_PATH" -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.py" 2>/dev/null | xargs grep -c "TODO\|FIXME" 2>/dev/null | awk '{sum+=$1} END {print sum+0}')
    if [ "$todo_count" -gt 10 ]; then
        recommendations+=("Обрати внимание на $todo_count TODO комментариев")
    fi
    
    local large_files=$(find "$PROJECT_PATH" -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.py" 2>/dev/null | xargs wc -l 2>/dev/null | awk '$1 > 500 {print $2}' | wc -l)
    if [ "$large_files" -gt 0 ]; then
        recommendations+=("Рассмотри рефакторинг $large_files больших файлов")
    fi
    
    local completion_rate=0
    local total_tasks=$(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    local completed_tasks=$(./ricochet-task tasks list --status "completed" --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    if [ $total_tasks -gt 0 ]; then
        completion_rate=$((completed_tasks * 100 / total_tasks))
    fi
    
    if [ $completion_rate -lt 30 ]; then
        recommendations+=("Улучши скорость завершения задач ($completion_rate%)")
    fi
    
    # Вывод рекомендаций
    if [ ${#recommendations[@]} -gt 0 ]; then
        echo "🎯 Рекомендации для AI:"
        for rec in "${recommendations[@]}"; do
            echo "   💡 $rec"
        done
    else
        echo "✅ Проект в отличном состоянии! Продолжай в том же духе"
    fi
}

# Основная функция анализа
main_analysis() {
    local start_time=$(date +%s)
    
    echo "🚀 Запуск AI Project Analyzer..."
    echo "=========================================="
    
    # Анализ архитектуры
    analyze_architecture
    
    # Анализ качества кода
    analyze_code_quality
    
    # Анализ производительности команды
    analyze_team_performance
    
    # Анализ рисков проекта
    analyze_project_risks
    
    echo "=========================================="
    
    # Генерация рекомендаций
    generate_ai_recommendations
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    echo ""
    echo "⏱️ Анализ завершен за ${duration}с"
}

# Запуск в зависимости от глубины
case "$ANALYSIS_DEPTH" in
    "quick")
        echo "⚡ Быстрый анализ"
        analyze_architecture
        analyze_code_quality
        ;;
    "deep")
        main_analysis
        ;;
    "full")
        echo "🔍 Полный анализ"
        main_analysis
        generate_ai_recommendations
        ;;
    *)
        main_analysis
        ;;
esac
