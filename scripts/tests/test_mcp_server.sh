#!/bin/bash
# test_mcp_server.sh - Тестирование MCP сервера

echo "🧪 Тестирование MCP сервера..."

# Функция для логирования результатов
log_result() {
    local test_name="$1"
    local result="$2"
    local message="$3"
    
    if [ "$result" = "success" ]; then
        echo "   ✅ $test_name: $message"
    else
        echo "   ❌ $test_name: $message"
    fi
}

# Функция для проверки порта
check_port() {
    local port="$1"
    if netstat -an 2>/dev/null | grep -q ":$port "; then
        return 0
    else
        return 1
    fi
}

# Тест 1: Запуск MCP сервера
echo "1. Запуск MCP сервера..."
./scripts/start-mcp.sh
sleep 3

if check_port 8091; then
    log_result "Запуск MCP сервера" "success" "Сервер запущен на порту 8091"
else
    log_result "Запуск MCP сервера" "error" "Сервер не запущен"
    exit 1
fi

# Тест 2: Проверка статуса
echo "2. Проверка статуса MCP сервера..."
if ./scripts/status-mcp.sh > /dev/null 2>&1; then
    log_result "Проверка статуса" "success" "Статус получен успешно"
else
    log_result "Проверка статуса" "error" "Ошибка получения статуса"
fi

# Тест 3: Проверка здоровья
echo "3. Проверка здоровья MCP сервера..."
if curl -s http://localhost:8091/health > /dev/null 2>&1; then
    log_result "Проверка здоровья" "success" "Сервер отвечает на health check"
else
    log_result "Проверка здоровья" "error" "Сервер не отвечает на health check"
fi

# Тест 4: Проверка доступных инструментов
echo "4. Проверка доступных инструментов..."
if curl -s http://localhost:8091/tools > /dev/null 2>&1; then
    log_result "Проверка инструментов" "success" "Инструменты доступны"
else
    log_result "Проверка инструментов" "error" "Инструменты недоступны"
fi

# Тест 5: Проверка JSON ответа
echo "5. Проверка JSON ответа..."
if curl -s http://localhost:8091/tools | jq . > /dev/null 2>&1; then
    log_result "Проверка JSON" "success" "JSON ответ валиден"
else
    log_result "Проверка JSON" "error" "JSON ответ невалиден"
fi

# Тест 6: Остановка сервера
echo "6. Остановка MCP сервера..."
./scripts/stop-mcp.sh
sleep 2

if ! check_port 8091; then
    log_result "Остановка сервера" "success" "Сервер остановлен"
else
    log_result "Остановка сервера" "error" "Сервер не остановлен"
fi

echo "✅ Тестирование MCP сервера завершено"
