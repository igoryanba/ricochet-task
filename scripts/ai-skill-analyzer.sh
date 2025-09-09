#!/bin/bash
# Анализатор навыков команды для AI

PROJECT_PATH=${1:-.}
ANALYSIS_TYPE=${2:-"full"}  # full, quick, gaps
OUTPUT_FORMAT=${3:-"ai"}  # ai, json, table

echo "🎯 AI Skill Analyzer - Анализ навыков команды для AI"
echo "📁 Проект: $PROJECT_PATH"
echo "🔍 Тип: $ANALYSIS_TYPE"
echo "📊 Формат: $OUTPUT_FORMAT"

# Функция для AI-дружественного вывода
ai_skill_output() {
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

# Функция анализа навыков по типам задач
analyze_task_skills() {
    echo "📊 Анализ навыков по типам задач..."
    
    # Получение статистики по типам задач и исполнителям
    local skill_data=$(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | awk '{print $NF, $3}' | sort | uniq -c | sort -nr)
    
    if [ -z "$skill_data" ]; then
        ai_skill_output "Нет данных о навыках команды" "warning" "Создай задачи с типами и исполнителями"
        return
    fi
    
    # Группировка по исполнителям
    local members=($(echo "$skill_data" | awk '{print $2}' | sort -u | grep -v "admin"))
    local skill_matrix=()
    
    for member in "${members[@]}"; do
        if [ ! -z "$member" ]; then
            local member_skills=$(echo "$skill_data" | grep "^[[:space:]]*[0-9]*[[:space:]]*$member " | sort -nr)
            local total_tasks=0
            local skill_types=()
            
            while IFS= read -r line; do
                local count=$(echo "$line" | awk '{print $1}')
                local task_type=$(echo "$line" | awk '{print $3}')
                total_tasks=$((total_tasks + count))
                skill_types+=("$task_type:$count")
            done <<< "$member_skills"
            
            skill_matrix+=("$member:$total_tasks:${skill_types[*]}")
        fi
    done
    
    # Вывод анализа навыков
    ai_skill_output "Навыки команды:" "info"
    for member_data in "${skill_matrix[@]}"; do
        local member=$(echo "$member_data" | cut -d: -f1)
        local total=$(echo "$member_data" | cut -d: -f2)
        local skills=$(echo "$member_data" | cut -d: -f3-)
        
        echo "   👤 $member (всего задач: $total):"
        
        # Анализ навыков
        local skill_array=($skills)
        for skill_data in "${skill_array[@]}"; do
            local skill_type=$(echo "$skill_data" | cut -d: -f1)
            local count=$(echo "$skill_data" | cut -d: -f2)
            local percentage=0
            if [ $total -gt 0 ]; then
                percentage=$((count * 100 / total))
            fi
            
            echo "      • $skill_type: $count задач ($percentage%)"
        done
    done
}

# Функция анализа пробелов в навыках
analyze_skill_gaps() {
    echo "🔍 Анализ пробелов в навыках..."
    
    # Получение всех типов задач
    local all_task_types=$(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | awk '{print $3}' | sort | uniq)
    local all_members=$(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | awk '{print $NF}' | sort | uniq | grep -v "admin")
    
    local gaps=()
    
    # Проверка каждого типа задач
    for task_type in $all_task_types; do
        if [ ! -z "$task_type" ]; then
            local has_expert=false
            local has_any=false
            
            # Проверка, есть ли эксперт по этому типу
            for member in $all_members; do
                if [ ! -z "$member" ]; then
                    local member_tasks=$(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | grep "$member" | grep "$task_type" | wc -l)
                    
                    if [ $member_tasks -gt 0 ]; then
                        has_any=true
                        if [ $member_tasks -gt 5 ]; then
                            has_expert=true
                        fi
                    fi
                fi
            done
            
            if [ "$has_any" = false ]; then
                gaps+=("$task_type: Нет исполнителей")
            elif [ "$has_expert" = false ]; then
                gaps+=("$task_type: Нет экспертов")
            fi
        fi
    done
    
    # Вывод пробелов
    if [ ${#gaps[@]} -gt 0 ]; then
        ai_skill_output "Обнаружены пробелы в навыках:" "warning"
        for gap in "${gaps[@]}"; do
            local skill_type=$(echo "$gap" | cut -d: -f1)
            local issue=$(echo "$gap" | cut -d: -f2)
            echo "   • $skill_type: $issue"
        done
        
        ai_skill_output "Рекомендации по развитию навыков" "info" "Создай задачи для обучения недостающим навыкам"
    else
        ai_skill_output "Пробелов в навыках не обнаружено" "success"
    fi
}

# Функция анализа специализации
analyze_specialization() {
    echo "🎯 Анализ специализации команды..."
    
    # Получение данных о навыках
    local skill_data=$(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | awk '{print $NF, $3}' | sort | uniq -c | sort -nr)
    
    if [ -z "$skill_data" ]; then
        ai_skill_output "Нет данных для анализа специализации" "warning"
        return
    fi
    
    # Анализ специализации по типам задач
    local task_types=($(echo "$skill_data" | awk '{print $3}' | sort -u | grep -v "admin"))
    local specialization_map=()
    
    for task_type in "${task_types[@]}"; do
        if [ ! -z "$task_type" ]; then
            local type_data=$(echo "$skill_data" | grep " $task_type$" | sort -nr)
            local total_tasks=0
            local top_member=""
            local top_count=0
            
            while IFS= read -r line; do
                local count=$(echo "$line" | awk '{print $1}')
                local member=$(echo "$line" | awk '{print $2}')
                total_tasks=$((total_tasks + count))
                
                if [ $count -gt $top_count ]; then
                    top_count=$count
                    top_member="$member"
                fi
            done <<< "$type_data"
            
            if [ $total_tasks -gt 0 ]; then
                local concentration=$((top_count * 100 / total_tasks))
                specialization_map+=("$task_type:$top_member:$concentration:$total_tasks")
            fi
        fi
    done
    
    # Вывод анализа специализации
    ai_skill_output "Специализация команды:" "info"
    for spec_data in "${specialization_map[@]}"; do
        local task_type=$(echo "$spec_data" | cut -d: -f1)
        local top_member=$(echo "$spec_data" | cut -d: -f2)
        local concentration=$(echo "$spec_data" | cut -d: -f3)
        local total_tasks=$(echo "$spec_data" | cut -d: -f4)
        
        echo "   🎯 $task_type:"
        echo "      👤 Эксперт: $top_member ($concentration% от $total_tasks задач)"
        
        if [ $concentration -gt 80 ]; then
            ai_skill_output "Высокая концентрация навыков в $task_type" "warning" "Рассмотри распределение знаний"
        elif [ $concentration -lt 30 ]; then
            ai_skill_output "Низкая концентрация навыков в $task_type" "warning" "Рассмотри специализацию"
        fi
    done
}

# Функция генерации рекомендаций по развитию
generate_development_recommendations() {
    echo "💡 Генерация рекомендаций по развитию..."
    
    local recommendations=()
    
    # Анализ пробелов в навыках
    local all_task_types=$(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | awk '{print $3}' | sort | uniq)
    local all_members=$(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | awk '{print $NF}' | sort | uniq | grep -v "admin")
    
    for task_type in $all_task_types; do
        if [ ! -z "$task_type" ]; then
            local has_any=false
            local member_count=0
            
            for member in $all_members; do
                if [ ! -z "$member" ]; then
                    local member_tasks=$(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | grep "$member" | grep "$task_type" | wc -l)
                    
                    if [ $member_tasks -gt 0 ]; then
                        has_any=true
                        member_count=$((member_count + 1))
                    fi
                fi
            done
            
            if [ "$has_any" = false ]; then
                recommendations+=("Создай задачи для обучения $task_type")
            elif [ $member_count -eq 1 ]; then
                recommendations+=("Рассмотри кросс-тренинг для $task_type")
            fi
        fi
    done
    
    # Вывод рекомендаций
    if [ ${#recommendations[@]} -gt 0 ]; then
        ai_skill_output "Рекомендации по развитию навыков:" "info"
        for rec in "${recommendations[@]}"; do
            echo "   💡 $rec"
        done
    else
        ai_skill_output "Команда имеет хорошее покрытие навыков" "success"
    fi
}

# Основная функция
main() {
    case "$ANALYSIS_TYPE" in
        "quick")
            analyze_task_skills
            ;;
        "gaps")
            analyze_skill_gaps
            ;;
        "full"|*)
            analyze_task_skills
            analyze_skill_gaps
            analyze_specialization
            generate_development_recommendations
            ;;
    esac
}

# Запуск
main
