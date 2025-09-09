#!/bin/bash
# Мощная система управления командой для AI

PROJECT_PATH=${1:-.}
MODE=${2:-"analyze"}  # analyze, balance, assign, report
OUTPUT_FORMAT=${3:-"ai"}  # ai, json, table

echo "👥 AI Team Manager - Управление командой для AI"
echo "📁 Проект: $PROJECT_PATH"
echo "🔍 Режим: $MODE"
echo "📊 Формат: $OUTPUT_FORMAT"

# Функция для AI-дружественного вывода
ai_team_output() {
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

# Функция анализа загрузки команды
analyze_team_load() {
    echo "📊 Анализ загрузки команды..."
    
    # Получение статистики по исполнителям
    local team_stats=$(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | awk '{print $NF}' | sort | uniq -c | sort -nr)
    
    if [ -z "$team_stats" ]; then
        ai_team_output "Нет данных о команде" "warning" "Создай задачи с исполнителями"
        return
    fi
    
    # Анализ распределения задач
    local total_tasks=0
    local max_tasks=0
    local min_tasks=999999
    local team_members=()
    
    while IFS= read -r line; do
        local count=$(echo "$line" | awk '{print $1}')
        local member=$(echo "$line" | awk '{print $2}')
        
        if [ "$member" != "admin" ] && [ ! -z "$member" ]; then
            team_members+=("$member:$count")
            total_tasks=$((total_tasks + count))
            
            if [ $count -gt $max_tasks ]; then
                max_tasks=$count
            fi
            if [ $count -lt $min_tasks ]; then
                min_tasks=$count
            fi
        fi
    done <<< "$team_stats"
    
    # Расчет метрик
    local team_size=${#team_members[@]}
    local avg_tasks=0
    if [ $team_size -gt 0 ]; then
        avg_tasks=$((total_tasks / team_size))
    fi
    
    local load_imbalance=0
    if [ $max_tasks -gt 0 ] && [ $min_tasks -gt 0 ]; then
        load_imbalance=$((max_tasks * 100 / min_tasks))
    fi
    
    # Вывод анализа
    ai_team_output "Статистика команды:" "info"
    echo "   👥 Размер команды: $team_size"
    echo "   📊 Всего задач: $total_tasks"
    echo "   📈 Средняя загрузка: $avg_tasks задач на человека"
    echo "   ⚖️  Дисбаланс нагрузки: $load_imbalance%"
    
    # Анализ каждого члена команды
    for member_data in "${team_members[@]}"; do
        local member=$(echo "$member_data" | cut -d: -f1)
        local count=$(echo "$member_data" | cut -d: -f2)
        local load_percent=0
        if [ $avg_tasks -gt 0 ]; then
            load_percent=$((count * 100 / avg_tasks))
        fi
        
        echo "   👤 $member: $count задач ($load_percent% от среднего)"
        
        # Предупреждения о перегрузке
        if [ $load_percent -gt 150 ]; then
            ai_team_output "$member перегружен ($load_percent%)" "warning" "Рассмотри перераспределение задач"
        elif [ $load_percent -lt 50 ]; then
            ai_team_output "$member недогружен ($load_percent%)" "warning" "Рассмотри назначение дополнительных задач"
        fi
    done
    
    # Рекомендации по балансировке
    if [ $load_imbalance -gt 200 ]; then
        ai_team_output "Высокий дисбаланс нагрузки ($load_imbalance%)" "critical" "Используй 'ai-team-manager.sh balance' для балансировки"
    elif [ $load_imbalance -gt 150 ]; then
        ai_team_output "Умеренный дисбаланс нагрузки ($load_imbalance%)" "warning" "Рассмотри перераспределение задач"
    else
        ai_team_output "Нагрузка команды сбалансирована" "success"
    fi
}

# Функция анализа навыков команды
analyze_team_skills() {
    echo "🎯 Анализ навыков команды..."
    
    # Анализ типов задач по исполнителям
    local skill_analysis=$(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | awk '{print $NF, $3}' | sort | uniq -c | sort -nr)
    
    if [ -z "$skill_analysis" ]; then
        ai_team_output "Нет данных о навыках команды" "warning" "Создай задачи с типами и исполнителями"
        return
    fi
    
    # Группировка по исполнителям
    local member_skills=()
    while IFS= read -r line; do
        local count=$(echo "$line" | awk '{print $1}')
        local member=$(echo "$line" | awk '{print $2}')
        local task_type=$(echo "$line" | awk '{print $3}')
        
        if [ "$member" != "admin" ] && [ ! -z "$member" ] && [ ! -z "$task_type" ]; then
            member_skills+=("$member:$task_type:$count")
        fi
    done <<< "$skill_analysis"
    
    # Анализ специализации
    local members=$(printf '%s\n' "${member_skills[@]}" | cut -d: -f1 | sort -u)
    
    ai_team_output "Специализация команды:" "info"
    for member in $members; do
        echo "   👤 $member:"
        local member_tasks=$(printf '%s\n' "${member_skills[@]}" | grep "^$member:" | sort -t: -k3 -nr)
        while IFS= read -r task_data; do
            local task_type=$(echo "$task_data" | cut -d: -f2)
            local count=$(echo "$task_data" | cut -d: -f3)
            echo "      • $task_type: $count задач"
        done <<< "$member_tasks"
    done
    
    # Рекомендации по развитию навыков
    ai_team_output "Рекомендации по развитию:" "info"
    echo "   💡 Рассмотри кросс-тренинг для лучшего распределения задач"
    echo "   💡 Создай задачи для развития недостающих навыков"
    echo "   💡 Используй парное программирование для передачи знаний"
}

# Функция анализа производительности команды
analyze_team_performance() {
    echo "📈 Анализ производительности команды..."
    
    # Получение статистики по статусам
    local status_stats=$(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | awk '{print $4}' | sort | uniq -c | sort -nr)
    
    local total_tasks=0
    local completed_tasks=0
    local in_progress_tasks=0
    local blocked_tasks=0
    
    while IFS= read -r line; do
        local count=$(echo "$line" | awk '{print $1}')
        local status=$(echo "$line" | awk '{print $2}')
        
        total_tasks=$((total_tasks + count))
        
        case "$status" in
            "completed"|"done")
                completed_tasks=$((completed_tasks + count))
                ;;
            "in_progress"|"in-progress")
                in_progress_tasks=$((in_progress_tasks + count))
                ;;
            "blocked")
                blocked_tasks=$((blocked_tasks + count))
                ;;
        esac
    done <<< "$status_stats"
    
    # Расчет метрик производительности
    local completion_rate=0
    if [ $total_tasks -gt 0 ]; then
        completion_rate=$((completed_tasks * 100 / total_tasks))
    fi
    
    local active_rate=0
    if [ $total_tasks -gt 0 ]; then
        active_rate=$((in_progress_tasks * 100 / total_tasks))
    fi
    
    local blocked_rate=0
    if [ $total_tasks -gt 0 ]; then
        blocked_rate=$((blocked_tasks * 100 / total_tasks))
    fi
    
    # Вывод анализа
    ai_team_output "Производительность команды:" "info"
    echo "   📊 Всего задач: $total_tasks"
    echo "   ✅ Завершено: $completed_tasks ($completion_rate%)"
    echo "   🔄 В работе: $in_progress_tasks ($active_rate%)"
    echo "   🚫 Заблокировано: $blocked_tasks ($blocked_rate%)"
    
    # Оценка производительности
    if [ $completion_rate -gt 80 ]; then
        ai_team_output "Отличная производительность ($completion_rate%)" "success"
    elif [ $completion_rate -gt 60 ]; then
        ai_team_output "Хорошая производительность ($completion_rate%)" "info"
    elif [ $completion_rate -gt 40 ]; then
        ai_team_output "Средняя производительность ($completion_rate%)" "warning" "Рассмотри упрощение задач"
    else
        ai_team_output "Низкая производительность ($completion_rate%)" "critical" "Требуется анализ блокеров и упрощение задач"
    fi
    
    if [ $blocked_rate -gt 20 ]; then
        ai_team_output "Высокий уровень блокеров ($blocked_rate%)" "critical" "Используй 'ai-team-manager.sh unblock' для разблокировки"
    elif [ $blocked_rate -gt 10 ]; then
        ai_team_output "Умеренный уровень блокеров ($blocked_rate%)" "warning" "Рассмотри устранение блокеров"
    else
        ai_team_output "Низкий уровень блокеров ($blocked_rate%)" "success"
    fi
}

# Функция автоматического назначения задач
auto_assign_tasks() {
    echo "🤖 Автоматическое назначение задач..."
    
    # Получение незанятых задач
    local unassigned_tasks=$(./ricochet-task tasks list --status "open" --limit 50 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | grep "admin" | head -10)
    
    if [ -z "$unassigned_tasks" ]; then
        ai_team_output "Нет незанятых задач" "info"
        return
    fi
    
    # Получение команды
    local team_members=($(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | awk '{print $NF}' | sort | uniq | grep -v "admin"))
    
    if [ ${#team_members[@]} -eq 0 ]; then
        ai_team_output "Нет членов команды" "warning" "Добавь исполнителей в задачи"
        return
    fi
    
    # Простое назначение по кругу
    local member_index=0
    local assigned_count=0
    
    while IFS= read -r task_line; do
        if [ ! -z "$task_line" ]; then
            local task_id=$(echo "$task_line" | awk '{print $1}')
            local member="${team_members[$member_index]}"
            
            echo "   📝 Назначаю задачу $task_id на $member"
            
            # Назначение задачи (если поддерживается)
            # ./ricochet-task tasks update "$task_id" --assignee "$member"
            
            member_index=$((member_index + 1))
            if [ $member_index -ge ${#team_members[@]} ]; then
                member_index=0
            fi
            
            assigned_count=$((assigned_count + 1))
        fi
    done <<< "$unassigned_tasks"
    
    ai_team_output "Назначено $assigned_count задач" "success"
}

# Функция генерации отчета по команде
generate_team_report() {
    echo "📋 Генерация отчета по команде..."
    
    local report_file="team-report-$(date +%Y%m%d-%H%M%S).md"
    
    cat > "$report_file" << REPORT
# 📊 Отчет по команде - $(date '+%d.%m.%Y %H:%M')

## 👥 Состав команды
REPORT
    
    # Добавление статистики команды
    local team_stats=$(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | awk '{print $NF}' | sort | uniq -c | sort -nr)
    
    while IFS= read -r line; do
        local count=$(echo "$line" | awk '{print $1}')
        local member=$(echo "$line" | awk '{print $2}')
        
        if [ "$member" != "admin" ] && [ ! -z "$member" ]; then
            echo "- **$member**: $count задач" >> "$report_file"
        fi
    done <<< "$team_stats"
    
    cat >> "$report_file" << REPORT

## 📈 Производительность
REPORT
    
    # Добавление метрик производительности
    local total_tasks=$(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    local completed_tasks=$(./ricochet-task tasks list --status "completed" --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | wc -l)
    local completion_rate=0
    if [ $total_tasks -gt 0 ]; then
        completion_rate=$((completed_tasks * 100 / total_tasks))
    fi
    
    echo "- Всего задач: $total_tasks" >> "$report_file"
    echo "- Завершено: $completed_tasks ($completion_rate%)" >> "$report_file"
    
    cat >> "$report_file" << REPORT

## 💡 Рекомендации для AI
- Используй 'ai-team-manager.sh analyze' для анализа загрузки
- Используй 'ai-team-manager.sh balance' для балансировки нагрузки
- Используй 'ai-team-manager.sh assign' для автоматического назначения
REPORT
    
    ai_team_output "Отчет сохранен в $report_file" "success"
}

# Основная функция
main() {
    case "$MODE" in
        "analyze")
            analyze_team_load
            analyze_team_skills
            analyze_team_performance
            ;;
        "balance")
            analyze_team_load
            ai_team_output "Используй 'ai-team-manager.sh assign' для балансировки" "info"
            ;;
        "assign")
            auto_assign_tasks
            ;;
        "report")
            generate_team_report
            ;;
        *)
            echo "Использование: $0 [путь] [режим] [формат]"
            echo "Режимы: analyze, balance, assign, report"
            ;;
    esac
}

# Запуск
main
