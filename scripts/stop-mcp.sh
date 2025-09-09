#!/bin/bash
# Скрипт для остановки MCP сервера Ricochet

echo "🛑 Остановка MCP сервера Ricochet..."

# Поиск процесса MCP сервера
MCP_PID=$(ps aux | grep "ricochet-task mcp" | grep -v grep | awk '{print $2}')

if [ -z "$MCP_PID" ]; then
    echo "⚠️ MCP сервер не запущен"
    exit 0
fi

echo "Найден процесс MCP сервера с PID: $MCP_PID"

# Остановка процесса
kill $MCP_PID

# Ожидание остановки
sleep 2

# Проверка, что процесс остановлен
if ps -p $MCP_PID > /dev/null 2>&1; then
    echo "⚠️ Процесс не остановился, принудительная остановка..."
    kill -9 $MCP_PID
fi

# Проверка, что сервер недоступен
if ! curl -s http://localhost:8091/tools > /dev/null 2>&1; then
    echo "✅ MCP сервер успешно остановлен"
else
    echo "❌ Не удалось остановить MCP сервер"
    exit 1
fi
