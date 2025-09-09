#!/bin/bash
# Скрипт для предложений по рефакторингу

FILE_PATH=${1:-.}
PRIORITY=${2:-"medium"}

echo "🔧 Анализ кода для предложений по рефакторингу: $FILE_PATH"

# Функция анализа длинных функций
analyze_long_functions() {
    local file="$1"
    local file_ext="${file##*.}"
    local suggestions=()
    
    case "$file_ext" in
        "js"|"ts"|"jsx"|"tsx")
            # Анализ JavaScript/TypeScript функций
            while IFS= read -r line; do
                local line_num=$(echo "$line" | cut -d: -f1)
                local content=$(echo "$line" | cut -d: -f2-)
                
                # Поиск длинных функций
                if echo "$content" | grep -q "function\|=>"; then
                    # Подсчет строк в функции (простая эвристика)
                    local func_start=$line_num
                    local func_end=$((func_start + 50)) # Предполагаем максимум 50 строк
                    
                    local func_lines=$(sed -n "${func_start},${func_end}p" "$file" | wc -l)
                    
                    if [ $func_lines -gt 30 ]; then
                        local func_name=$(echo "$content" | grep -o "function [a-zA-Z_][a-zA-Z0-9_]*" | head -1)
                        suggestions+=("$line_num:Длинная функция $func_name ($func_lines строк)")
                    fi
                fi
            done < <(grep -n "function\|=>" "$file" 2>/dev/null)
            ;;
        "go")
            # Анализ Go функций
            while IFS= read -r line; do
                local line_num=$(echo "$line" | cut -d: -f1)
                local content=$(echo "$line" | cut -d: -f2-)
                
                if echo "$content" | grep -q "func "; then
                    local func_start=$line_num
                    local func_end=$((func_start + 100)) # Go функции могут быть длиннее
                    
                    local func_lines=$(sed -n "${func_start},${func_end}p" "$file" | wc -l)
                    
                    if [ $func_lines -gt 50 ]; then
                        local func_name=$(echo "$content" | grep -o "func [a-zA-Z_][a-zA-Z0-9_]*" | head -1)
                        suggestions+=("$line_num:Длинная функция $func_name ($func_lines строк)")
                    fi
                fi
            done < <(grep -n "func " "$file" 2>/dev/null)
            ;;
        "py")
            # Анализ Python функций
            while IFS= read -r line; do
                local line_num=$(echo "$line" | cut -d: -f1)
                local content=$(echo "$line" | cut -d: -f2-)
                
                if echo "$content" | grep -q "def "; then
                    local func_start=$line_num
                    local func_end=$((func_start + 50))
                    
                    local func_lines=$(sed -n "${func_start},${func_end}p" "$file" | wc -l)
                    
                    if [ $func_lines -gt 30 ]; then
                        local func_name=$(echo "$content" | grep -o "def [a-zA-Z_][a-zA-Z0-9_]*" | head -1)
                        suggestions+=("$line_num:Длинная функция $func_name ($func_lines строк)")
                    fi
                fi
            done < <(grep -n "def " "$file" 2>/dev/null)
            ;;
    esac
    
    printf '%s\n' "${suggestions[@]}"
}

# Функция анализа дублирования кода
analyze_duplication() {
    local file="$1"
    local suggestions=()
    
    # Поиск повторяющихся блоков кода (простая эвристика)
    local lines=$(wc -l < "$file" 2>/dev/null || echo "0")
    
    if [ $lines -gt 100 ]; then
        # Поиск повторяющихся строк
        local duplicates=$(sort "$file" | uniq -d | wc -l)
        
        if [ $duplicates -gt 5 ]; then
            suggestions+=("0:Возможно дублирование кода ($duplicates повторяющихся строк)")
        fi
    fi
    
    printf '%s\n' "${suggestions[@]}"
}

# Функция анализа сложных условий
analyze_complex_conditions() {
    local file="$1"
    local suggestions=()
    
    # Поиск сложных if условий
    while IFS= read -r line; do
        local line_num=$(echo "$line" | cut -d: -f1)
        local content=$(echo "$line" | cut -d: -f2-)
        
        # Подсчет операторов в условии
        local and_count=$(echo "$content" | grep -o "&&" | wc -l)
        local or_count=$(echo "$content" | grep -o "||" | wc -l)
        local total_ops=$((and_count + or_count))
        
        if [ $total_ops -gt 3 ]; then
            suggestions+=("$line_num:Сложное условие ($total_ops операторов)")
        fi
    done < <(grep -n "if.*(" "$file" 2>/dev/null)
    
    printf '%s\n' "${suggestions[@]}"
}

# Функция анализа больших классов/модулей
analyze_large_modules() {
    local file="$1"
    local file_ext="${file##*.}"
    local suggestions=()
    
    local line_count=$(wc -l < "$file" 2>/dev/null || echo "0")
    
    case "$file_ext" in
        "js"|"ts"|"jsx"|"tsx")
            if [ $line_count -gt 200 ]; then
                suggestions+=("0:Большой модуль ($line_count строк) - рассмотрите разделение")
            fi
            ;;
        "go")
            if [ $line_count -gt 500 ]; then
                suggestions+=("0:Большой файл ($line_count строк) - рассмотрите разделение на пакеты")
            fi
            ;;
        "py")
            if [ $line_count -gt 300 ]; then
                suggestions+=("0:Большой модуль ($line_count строк) - рассмотрите разделение")
            fi
            ;;
    esac
    
    printf '%s\n' "${suggestions[@]}"
}

# Основная функция анализа
analyze_file() {
    local file="$1"
    local file_ext="${file##*.}"
    
    echo "📄 Анализ файла для рефакторинга: $file"
    
    # Анализ длинных функций
    echo "🔍 Поиск длинных функций..."
    local long_functions=($(analyze_long_functions "$file"))
    
    # Анализ дублирования
    echo "🔍 Поиск дублирования кода..."
    local duplications=($(analyze_duplication "$file"))
    
    # Анализ сложных условий
    echo "🔍 Поиск сложных условий..."
    local complex_conditions=($(analyze_complex_conditions "$file"))
    
    # Анализ больших модулей
    echo "🔍 Анализ размера модуля..."
    local large_modules=($(analyze_large_modules "$file"))
    
    # Создание задач для рефакторинга
    local task_count=0
    
    # Задачи для длинных функций
    for suggestion in "${long_functions[@]}"; do
        local line_num=$(echo "$suggestion" | cut -d: -f1)
        local description=$(echo "$suggestion" | cut -d: -f2-)
        
        local task_title="Refactor: $description"
        local task_description="Найден в файле $file на строке $line_num. Рекомендуется разбить на более мелкие функции."
        
        echo "📝 Создание задачи: $task_title"
        ./ricochet-task tasks create \
            --title "$task_title" \
            --description "$task_description" \
            --type "refactoring" \
            --priority "$PRIORITY" \
            --labels "refactoring,code-quality" \
            --project "0-1"
        
        task_count=$((task_count + 1))
    done
    
    # Задачи для дублирования
    for suggestion in "${duplications[@]}"; do
        local line_num=$(echo "$suggestion" | cut -d: -f1)
        local description=$(echo "$suggestion" | cut -d: -f2-)
        
        local task_title="Refactor: $description"
        local task_description="Найден в файле $file. Рекомендуется вынести дублирующийся код в отдельные функции."
        
        echo "📝 Создание задачи: $task_title"
        ./ricochet-task tasks create \
            --title "$task_title" \
            --description "$task_description" \
            --type "refactoring" \
            --priority "$PRIORITY" \
            --labels "refactoring,duplication" \
            --project "0-1"
        
        task_count=$((task_count + 1))
    done
    
    # Задачи для сложных условий
    for suggestion in "${complex_conditions[@]}"; do
        local line_num=$(echo "$suggestion" | cut -d: -f1)
        local description=$(echo "$suggestion" | cut -d: -f2-)
        
        local task_title="Refactor: $description"
        local task_description="Найден в файле $file на строке $line_num. Рекомендуется упростить условие."
        
        echo "📝 Создание задачи: $task_title"
        ./ricochet-task tasks create \
            --title "$task_title" \
            --description "$task_description" \
            --type "refactoring" \
            --priority "$PRIORITY" \
            --labels "refactoring,complexity" \
            --project "0-1"
        
        task_count=$((task_count + 1))
    done
    
    # Задачи для больших модулей
    for suggestion in "${large_modules[@]}"; do
        local line_num=$(echo "$suggestion" | cut -d: -f1)
        local description=$(echo "$suggestion" | cut -d: -f2-)
        
        local task_title="Refactor: $description"
        local task_description="Найден в файле $file. Рекомендуется разделить на более мелкие модули."
        
        echo "📝 Создание задачи: $task_title"
        ./ricochet-task tasks create \
            --title "$task_title" \
            --description "$task_description" \
            --type "refactoring" \
            --priority "$PRIORITY" \
            --labels "refactoring,architecture" \
            --project "0-1"
        
        task_count=$((task_count + 1))
    done
    
    echo "📊 Создано задач для рефакторинга: $task_count"
}

# Анализ файла или директории
if [ -f "$FILE_PATH" ]; then
    analyze_file "$FILE_PATH"
elif [ -d "$FILE_PATH" ]; then
    echo "📁 Анализ директории для рефакторинга: $FILE_PATH"
    
    # Поиск файлов для анализа
    find "$FILE_PATH" -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.py" | while read -r file; do
        analyze_file "$file"
        echo "---"
    done
else
    echo "❌ Файл или директория не найдены: $FILE_PATH"
    exit 1
fi

echo "✅ Анализ для рефакторинга завершен"
