#!/bin/bash
# Скрипт для проверки статуса MCP сервера Ricochet

echo "🔍 Проверка статуса MCP сервера Ricochet..."

# Проверка, что сервер доступен
if curl -s http://localhost:8091/tools > /dev/null 2>&1; then
    echo "✅ MCP сервер запущен и доступен"
    
    # Получение информации о сервере
    TOOLS_COUNT=$(curl -s http://localhost:8091/tools | jq '.tools | length' 2>/dev/null || echo "N/A")
    echo "📊 Доступно инструментов: $TOOLS_COUNT"
    
    # Получение списка инструментов
    echo "🛠️ Доступные инструменты:"
    curl -s http://localhost:8091/tools | jq -r '.tools[].name' 2>/dev/null | head -10
    
    if [ "$TOOLS_COUNT" != "N/A" ] && [ "$TOOLS_COUNT" -gt 10 ]; then
        echo "... и еще $((TOOLS_COUNT - 10)) инструментов"
    fi
    
    echo "🌐 URL: http://localhost:8091"
    
    # Проверка процесса
    MCP_PID=$(ps aux | grep "ricochet-task mcp" | grep -v grep | awk '{print $2}')
    if [ ! -z "$MCP_PID" ]; then
        echo "🔄 PID процесса: $MCP_PID"
    fi
    
else
    echo "❌ MCP сервер не запущен или недоступен"
    echo "Запустите сервер командой: ./scripts/start-mcp.sh"
    exit 1
fi
