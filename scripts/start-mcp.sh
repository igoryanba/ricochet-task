#!/bin/bash
# Скрипт для запуска MCP сервера Ricochet

echo "🚀 Запуск MCP сервера Ricochet..."

# Проверка, что мы в правильной директории
if [ ! -f "ricochet-task" ]; then
    echo "❌ Файл ricochet-task не найден. Запустите скрипт из директории ricochet-task"
    exit 1
fi

# Проверка, что MCP сервер не запущен
if curl -s http://localhost:8091/tools > /dev/null 2>&1; then
    echo "⚠️ MCP сервер уже запущен на порту 8091"
    echo "Остановите существующий сервер или используйте другой порт"
    exit 1
fi

# Запуск MCP сервера
echo "Запуск MCP сервера на порту 8091..."
./ricochet-task mcp start --port 8091 --verbose &

# Ожидание запуска
sleep 3

# Проверка, что сервер запустился
if curl -s http://localhost:8091/tools > /dev/null 2>&1; then
    echo "✅ MCP сервер успешно запущен"
    echo "Доступно инструментов: $(curl -s http://localhost:8091/tools | jq '.tools | length')"
    echo "URL: http://localhost:8091"
    echo "Для остановки используйте: pkill -f 'ricochet-task mcp'"
else
    echo "❌ Не удалось запустить MCP сервер"
    exit 1
fi
