#!/bin/bash
# Скрипт для анализа сложности кода

FILE_PATH=${1:-.}
COMPLEXITY_THRESHOLD=${2:-8}

echo "🔍 Анализ сложности кода: $FILE_PATH"

# Функция анализа JavaScript/TypeScript
analyze_js_complexity() {
    local file="$1"
    local complexity=0
    
    # Подсчет условных операторов
    local if_count=$(grep -c "if\|else\|switch\|case" "$file" 2>/dev/null || echo "0")
    local loop_count=$(grep -c "for\|while\|do" "$file" 2>/dev/null || echo "0")
    local try_count=$(grep -c "try\|catch\|finally" "$file" 2>/dev/null || echo "0")
    
    # Подсчет функций
    local function_count=$(grep -c "function\|=>" "$file" 2>/dev/null || echo "0")
    
    # Подсчет параметров функций
    local max_params=0
    while IFS= read -r line; do
        local param_count=$(echo "$line" | grep -o "," | wc -l)
        param_count=$((param_count + 1))
        if [ $param_count -gt $max_params ]; then
            max_params=$param_count
        fi
    done < <(grep "function.*(" "$file" 2>/dev/null)
    
    # Расчет сложности
    complexity=$((if_count + loop_count + try_count + function_count))
    
    # Штраф за длинные функции
    local line_count=$(wc -l < "$file" 2>/dev/null || echo "0")
    if [ $line_count -gt 50 ]; then
        complexity=$((complexity + 2))
    fi
    
    # Штраф за много параметров
    if [ $max_params -gt 5 ]; then
        complexity=$((complexity + 2))
    fi
    
    echo "$complexity"
}

# Функция анализа Go кода
analyze_go_complexity() {
    local file="$1"
    local complexity=0
    
    # Подсчет условных операторов
    local if_count=$(grep -c "if\|else\|switch\|case" "$file" 2>/dev/null || echo "0")
    local loop_count=$(grep -c "for\|range" "$file" 2>/dev/null || echo "0")
    local defer_count=$(grep -c "defer" "$file" 2>/dev/null || echo "0")
    
    # Подсчет функций
    local function_count=$(grep -c "func " "$file" 2>/dev/null || echo "0")
    
    # Подсчет горутин
    local goroutine_count=$(grep -c "go " "$file" 2>/dev/null || echo "0")
    
    # Расчет сложности
    complexity=$((if_count + loop_count + defer_count + function_count + goroutine_count))
    
    # Штраф за длинные функции
    local line_count=$(wc -l < "$file" 2>/dev/null || echo "0")
    if [ $line_count -gt 100 ]; then
        complexity=$((complexity + 3))
    fi
    
    echo "$complexity"
}

# Функция анализа Python кода
analyze_python_complexity() {
    local file="$1"
    local complexity=0
    
    # Подсчет условных операторов
    local if_count=$(grep -c "if\|elif\|else" "$file" 2>/dev/null || echo "0")
    local loop_count=$(grep -c "for\|while" "$file" 2>/dev/null || echo "0")
    local try_count=$(grep -c "try\|except\|finally" "$file" 2>/dev/null || echo "0")
    
    # Подсчет функций и классов
    local function_count=$(grep -c "def " "$file" 2>/dev/null || echo "0")
    local class_count=$(grep -c "class " "$file" 2>/dev/null || echo "0")
    
    # Подсчет вложенности (отступы)
    local max_indent=0
    while IFS= read -r line; do
        local indent=$(echo "$line" | sed 's/[^ ].*//' | wc -c)
        indent=$((indent - 1))
        if [ $indent -gt $max_indent ]; then
            max_indent=$indent
        fi
    done < "$file" 2>/dev/null
    
    # Расчет сложности
    complexity=$((if_count + loop_count + try_count + function_count + class_count))
    
    # Штраф за высокую вложенность
    if [ $max_indent -gt 4 ]; then
        complexity=$((complexity + 3))
    fi
    
    echo "$complexity"
}

# Функция анализа TODO комментариев
analyze_todos() {
    local file="$1"
    local todos=()
    
    # Поиск TODO комментариев
    while IFS= read -r line; do
        local line_num=$(echo "$line" | cut -d: -f1)
        local content=$(echo "$line" | cut -d: -f2-)
        todos+=("$line_num:$content")
    done < <(grep -n "TODO\|FIXME\|HACK\|XXX" "$file" 2>/dev/null)
    
    printf '%s\n' "${todos[@]}"
}

# Основная функция анализа
analyze_file() {
    local file="$1"
    local file_ext="${file##*.}"
    local complexity=0
    local todos=()
    
    echo "📄 Анализ файла: $file"
    
    # Определение типа файла и анализ
    case "$file_ext" in
        "js"|"ts"|"jsx"|"tsx")
            complexity=$(analyze_js_complexity "$file")
            ;;
        "go")
            complexity=$(analyze_go_complexity "$file")
            ;;
        "py")
            complexity=$(analyze_python_complexity "$file")
            ;;
        *)
            echo "⚠️ Неподдерживаемый тип файла: $file_ext"
            return
            ;;
    esac
    
    # Анализ TODO комментариев
    todos=($(analyze_todos "$file"))
    
    # Вывод результатов
    echo "📊 Результаты анализа:"
    echo "  Сложность: $complexity"
    echo "  TODO комментариев: ${#todos[@]}"
    
    # Проверка порога сложности
    if [ $complexity -gt $COMPLEXITY_THRESHOLD ]; then
        echo "⚠️ Высокая сложность! Рекомендуется рефакторинг"
        
        # Создание задачи для рефакторинга
        local task_title="Refactor $(basename "$file")"
        local task_description="Файл имеет высокую сложность ($complexity). Рекомендуется разбить на более мелкие функции."
        
        echo "📝 Создание задачи для рефакторинга..."
        ./ricochet-task tasks create \
            --title "$task_title" \
            --description "$task_description" \
            --type "refactoring" \
            --priority "medium" \
            --labels "complexity,refactoring" \
            --project "0-1"
    fi
    
    # Создание задач для TODO комментариев
    if [ ${#todos[@]} -gt 0 ]; then
        echo "📝 Создание задач для TODO комментариев..."
        
        for todo in "${todos[@]}"; do
            local line_num=$(echo "$todo" | cut -d: -f1)
            local content=$(echo "$todo" | cut -d: -f2- | sed 's/^[[:space:]]*//')
            
            local task_title="TODO: $content"
            local task_description="Найден в файле $file на строке $line_num"
            
            ./ricochet-task tasks create \
                --title "$task_title" \
                --description "$task_description" \
                --type "task" \
                --priority "low" \
                --labels "todo,code-review" \
                --project "0-1"
        done
    fi
}

# Анализ файла или директории
if [ -f "$FILE_PATH" ]; then
    analyze_file "$FILE_PATH"
elif [ -d "$FILE_PATH" ]; then
    echo "📁 Анализ директории: $FILE_PATH"
    
    # Поиск файлов для анализа
    find "$FILE_PATH" -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.py" | while read -r file; do
        analyze_file "$file"
        echo "---"
    done
else
    echo "❌ Файл или директория не найдены: $FILE_PATH"
    exit 1
fi

echo "✅ Анализ завершен"
