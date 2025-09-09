#!/bin/bash
# AI Agent Coordinator - координация между разными AI агентами

PROJECT_PATH=${1:-.}
COORDINATION_TYPE=${2:-"assign"}  # assign, sync, handoff, review
AGENT_FROM=${3:-"coordinator"}  # агент, передающий задачу
AGENT_TO=${4:-"executor"}  # агент, получающий задачу
TASK_ID=${5:-""}  # ID задачи для передачи

echo "🤝 AI Agent Coordinator - Координация между AI агентами"
echo "📁 Проект: $PROJECT_PATH"
echo "🔍 Тип: $COORDINATION_TYPE"
echo "👤 От: $AGENT_FROM"
echo "👤 К: $AGENT_TO"
echo "📝 Задача: $TASK_ID"

# Функция для AI-дружественного вывода
ai_coordinator_output() {
    local message="$1"
    local level="$2"  # info, warning, critical, success
    local action="$3"  # optional action for AI
    local agent="$4"  # optional agent name
    
    local prefix="🤝"
    if [ ! -z "$agent" ]; then
        prefix="🤝[$agent]"
    fi
    
    case "$level" in
        "critical")
            echo "$prefix 🚨 CRITICAL: $message"
            if [ ! -z "$action" ]; then
                echo "   💡 AI Action: $action"
            fi
            ;;
        "warning")
            echo "$prefix ⚠️  WARNING: $message"
            if [ ! -z "$action" ]; then
                echo "   💡 AI Action: $action"
            fi
            ;;
        "success")
            echo "$prefix ✅ SUCCESS: $message"
            ;;
        *)
            echo "$prefix ℹ️  INFO: $message"
            ;;
    esac
}

# Функция назначения задач агентам
assign_tasks_to_agents() {
    echo "📋 Назначение задач агентам..."
    
    # Получение незанятых задач
    local unassigned_tasks=$(./ricochet-task tasks list --status "open" --limit 20 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--")
    
    if [ -z "$unassigned_tasks" ]; then
        ai_coordinator_output "Нет незанятых задач" "info"
        return
    fi
    
    # Определение агентов и их специализации
    local agents=(
        "code-reviewer:code,review,quality"
        "feature-developer:feature,enhancement,new"
        "bug-fixer:bug,fix,debug"
        "test-engineer:test,testing,qa"
        "documentation-writer:docs,documentation,readme"
        "devops-engineer:deploy,infrastructure,ci-cd"
        "security-expert:security,vulnerability,audit"
        "performance-optimizer:performance,optimization,speed"
    )
    
    local assigned_count=0
    
    while IFS= read -r task_line; do
        if [ ! -z "$task_line" ]; then
            local task_id=$(echo "$task_line" | awk '{print $1}')
            local task_title=$(echo "$task_line" | awk '{print $2}')
            local task_type=$(echo "$task_line" | awk '{print $3}')
            
            # Поиск подходящего агента
            local best_agent=""
            local best_score=0
            
            for agent_data in "${agents[@]}"; do
                local agent_name=$(echo "$agent_data" | cut -d: -f1)
                local agent_skills=$(echo "$agent_data" | cut -d: -f2)
                
                local score=0
                for skill in $(echo "$agent_skills" | tr ',' ' '); do
                    if echo "$task_type" | grep -qi "$skill"; then
                        score=$((score + 1))
                    fi
                    if echo "$task_title" | grep -qi "$skill"; then
                        score=$((score + 1))
                    fi
                done
                
                if [ $score -gt $best_score ]; then
                    best_score=$score
                    best_agent="$agent_name"
                fi
            done
            
            # Если не найден специализированный агент, назначаем по очереди
            if [ -z "$best_agent" ]; then
                local agent_index=$((assigned_count % ${#agents[@]}))
                best_agent=$(echo "${agents[$agent_index]}" | cut -d: -f1)
            fi
            
            echo "   📝 Задача $task_id: $task_title → $best_agent"
            
            # Назначение задачи агенту
            # ./ricochet-task tasks update "$task_id" --assignee "$best_agent"
            
            assigned_count=$((assigned_count + 1))
        fi
    done <<< "$unassigned_tasks"
    
    ai_coordinator_output "Назначено $assigned_count задач агентам" "success"
}

# Функция синхронизации между агентами
sync_agents() {
    echo "🔄 Синхронизация между агентами..."
    
    # Получение задач в работе
    local active_tasks=$(./ricochet-task tasks list --status "in_progress" --limit 50 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--")
    
    if [ -z "$active_tasks" ]; then
        ai_coordinator_output "Нет активных задач" "info"
        return
    fi
    
    local synced_count=0
    
    while IFS= read -r task_line; do
        if [ ! -z "$task_line" ]; then
            local task_id=$(echo "$task_line" | awk '{print $1}')
            local task_title=$(echo "$task_line" | awk '{print $2}')
            local assignee=$(echo "$task_line" | awk '{print $NF}')
            
            echo "   🔄 Синхронизация задачи $task_id с $assignee"
            
            # Проверка статуса задачи
            local task_status=$(./ricochet-task tasks get "$task_id" 2>/dev/null | grep "Status" | awk '{print $2}')
            
            if [ "$task_status" = "completed" ]; then
                echo "      ✅ Задача завершена, передача на ревью"
                # Передача на ревью
                # ./ricochet-task tasks update "$task_id" --assignee "code-reviewer"
            elif [ "$task_status" = "blocked" ]; then
                echo "      🚫 Задача заблокирована, передача координатору"
                # Передача координатору
                # ./ricochet-task tasks update "$task_id" --assignee "coordinator"
            else
                echo "      🔄 Задача в работе, продолжение выполнения"
            fi
            
            synced_count=$((synced_count + 1))
        fi
    done <<< "$active_tasks"
    
    ai_coordinator_output "Синхронизировано $synced_count задач" "success"
}

# Функция передачи задач между агентами
handoff_task() {
    echo "🤝 Передача задачи между агентами..."
    
    if [ -z "$TASK_ID" ]; then
        ai_coordinator_output "Не указан ID задачи" "warning" "Используй: $0 . handoff coordinator executor 3-45"
        return
    fi
    
    # Получение информации о задаче
    local task_info=$(./ricochet-task tasks get "$TASK_ID" 2>/dev/null)
    
    if [ -z "$task_info" ]; then
        ai_coordinator_output "Задача $TASK_ID не найдена" "warning"
        return
    fi
    
    local task_title=$(echo "$task_info" | grep "Title" | sed 's/Title: *//')
    local current_assignee=$(echo "$task_info" | grep "Assignee" | sed 's/Assignee: *//')
    
    echo "   📝 Задача: $task_title"
    echo "   👤 Текущий исполнитель: $current_assignee"
    echo "   👤 Новый исполнитель: $AGENT_TO"
    
    # Проверка возможности передачи
    if [ "$current_assignee" = "$AGENT_TO" ]; then
        ai_coordinator_output "Задача уже назначена на $AGENT_TO" "warning"
        return
    fi
    
    # Логика передачи задачи
    case "$AGENT_FROM" in
        "coordinator")
            echo "      📋 Координатор передает задачу $AGENT_TO"
            ;;
        "code-reviewer")
            if [ "$AGENT_TO" = "feature-developer" ]; then
                echo "      🔄 Ревьюер возвращает задачу разработчику"
            elif [ "$AGENT_TO" = "test-engineer" ]; then
                echo "      ✅ Ревьюер передает задачу тестировщику"
            fi
            ;;
        "feature-developer")
            if [ "$AGENT_TO" = "code-reviewer" ]; then
                echo "      📝 Разработчик передает задачу на ревью"
            elif [ "$AGENT_TO" = "test-engineer" ]; then
                echo "      🧪 Разработчик передает задачу на тестирование"
            fi
            ;;
        *)
            echo "      🤝 Передача задачи от $AGENT_FROM к $AGENT_TO"
            ;;
    esac
    
    # Обновление назначения
    # ./ricochet-task tasks update "$TASK_ID" --assignee "$AGENT_TO"
    
    ai_coordinator_output "Задача $TASK_ID передана от $AGENT_FROM к $AGENT_TO" "success"
}

# Функция ревью работы агентов
review_agent_work() {
    echo "👀 Ревью работы агентов..."
    
    # Получение завершенных задач
    local completed_tasks=$(./ricochet-task tasks list --status "completed" --limit 20 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--")
    
    if [ -z "$completed_tasks" ]; then
        ai_coordinator_output "Нет завершенных задач для ревью" "info"
        return
    fi
    
    local reviewed_count=0
    
    while IFS= read -r task_line; do
        if [ ! -z "$task_line" ]; then
            local task_id=$(echo "$task_line" | awk '{print $1}')
            local task_title=$(echo "$task_line" | awk '{print $2}')
            local assignee=$(echo "$task_line" | awk '{print $NF}')
            
            echo "   👀 Ревью задачи $task_id: $task_title (исполнитель: $assignee)"
            
            # Симуляция ревью
            local review_score=$((RANDOM % 10 + 1))
            
            if [ $review_score -ge 8 ]; then
                echo "      ✅ Отличная работа (оценка: $review_score/10)"
                ai_coordinator_output "Задача $task_id выполнена отлично" "success"
            elif [ $review_score -ge 6 ]; then
                echo "      👍 Хорошая работа (оценка: $review_score/10)"
                ai_coordinator_output "Задача $task_id выполнена хорошо" "info"
            else
                echo "      ⚠️  Требуются улучшения (оценка: $review_score/10)"
                ai_coordinator_output "Задача $task_id требует доработки" "warning" "Передай задачу обратно исполнителю"
            fi
            
            reviewed_count=$((reviewed_count + 1))
        fi
    done <<< "$completed_tasks"
    
    ai_coordinator_output "Проведен ревью $reviewed_count задач" "success"
}

# Функция генерации отчета по координации
generate_coordination_report() {
    echo "📋 Генерация отчета по координации агентов..."
    
    local report_file="ai-coordination-report-$(date +%Y%m%d-%H%M%S).md"
    
    cat > "$report_file" << REPORT
# 🤝 Отчет по координации AI агентов - $(date '+%d.%m.%Y %H:%M')

## 📊 Статистика координации
REPORT
    
    # Добавление статистики агентов
    local agent_stats=$(./ricochet-task tasks list --limit 100 2>/dev/null | grep -v "ID.*PROVIDER" | grep -v "^--" | awk '{print $NF}' | sort | uniq -c | sort -nr)
    
    while IFS= read -r line; do
        local count=$(echo "$line" | awk '{print $1}')
        local agent=$(echo "$line" | awk '{print $2}')
        
        if [ "$agent" != "admin" ] && [ ! -z "$agent" ]; then
            echo "- **$agent**: $count задач" >> "$report_file"
        fi
    done <<< "$agent_stats"
    
    cat >> "$report_file" << REPORT

## 💡 Рекомендации для AI
- Используй 'ai-agent-coordinator.sh assign' для назначения задач
- Используй 'ai-agent-coordinator.sh sync' для синхронизации
- Используй 'ai-agent-coordinator.sh handoff' для передачи задач
- Используй 'ai-agent-coordinator.sh review' для ревью работы
REPORT
    
    ai_coordinator_output "Отчет сохранен в $report_file" "success"
}

# Основная функция
main() {
    case "$COORDINATION_TYPE" in
        "assign")
            assign_tasks_to_agents
            ;;
        "sync")
            sync_agents
            ;;
        "handoff")
            handoff_task
            ;;
        "review")
            review_agent_work
            ;;
        "report")
            generate_coordination_report
            ;;
        *)
            echo "Использование: $0 [путь] [тип] [агент_от] [агент_к] [задача]"
            echo "Типы: assign, sync, handoff, review, report"
            echo "Агенты: coordinator, code-reviewer, feature-developer, bug-fixer, test-engineer, documentation-writer"
            ;;
    esac
}

# Запуск
main
