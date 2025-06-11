/**
 * ChainMonitorPanel - UI-компонент для мониторинга цепочек в редакторе
 * 
 * Этот компонент представляет панель для отображения статуса и визуализации цепочек.
 * Поддерживает обновление в реальном времени, отображение прогресса и интерактивное
 * управление выполнением цепочки.
 */

class ChainMonitorPanel {
    /**
     * Создает новую панель мониторинга цепочек
     * @param {Object} options - Параметры инициализации
     * @param {String} options.containerId - ID контейнера для панели
     * @param {Object} options.mcp - Объект MCP-клиента
     * @param {Function} options.onClose - Коллбэк при закрытии панели
     */
    constructor(options) {
        this.container = document.getElementById(options.containerId);
        this.mcp = options.mcp;
        this.onClose = options.onClose || (() => {});
        
        this.refreshInterval = null;
        this.currentChainId = null;
        this.refreshRate = 2000; // Интервал обновления в мс
        
        this.initPanel();
    }
    
    /**
     * Инициализирует панель
     */
    initPanel() {
        if (!this.container) {
            console.error('Контейнер для панели мониторинга не найден');
            return;
        }
        
        this.container.innerHTML = `
            <div class="ricochet-panel">
                <div class="ricochet-panel-header">
                    <h3>Мониторинг цепочки</h3>
                    <div class="ricochet-panel-actions">
                        <button id="chain-monitor-refresh" title="Обновить">↻</button>
                        <button id="chain-monitor-close" title="Закрыть">✕</button>
                    </div>
                </div>
                
                <div class="ricochet-panel-toolbar">
                    <select id="chain-monitor-chain-select">
                        <option value="">Выберите цепочку...</option>
                    </select>
                    
                    <div class="ricochet-panel-controls">
                        <button id="chain-monitor-start" title="Запустить" disabled>▶</button>
                        <button id="chain-monitor-pause" title="Пауза" disabled>⏸</button>
                        <button id="chain-monitor-stop" title="Остановить" disabled>⏹</button>
                    </div>
                </div>
                
                <div class="ricochet-panel-content">
                    <div id="chain-monitor-status" class="ricochet-status-bar">
                        <span class="chain-name">Не выбрана</span>
                        <span class="chain-status">-</span>
                        <span class="chain-progress">0%</span>
                    </div>
                    
                    <div id="chain-monitor-visualization" class="ricochet-visualization">
                        <div class="placeholder">Выберите цепочку для отображения</div>
                    </div>
                    
                    <div id="chain-monitor-details" class="ricochet-details">
                        <div class="ricochet-tabs">
                            <button class="tab-button active" data-tab="metrics">Метрики</button>
                            <button class="tab-button" data-tab="events">События</button>
                            <button class="tab-button" data-tab="logs">Логи</button>
                        </div>
                        
                        <div class="ricochet-tab-content active" id="tab-metrics">
                            <table class="metrics-table">
                                <tbody>
                                    <tr>
                                        <td>Время выполнения:</td>
                                        <td id="metric-elapsed">-</td>
                                    </tr>
                                    <tr>
                                        <td>Токены (вход/выход):</td>
                                        <td id="metric-tokens">-</td>
                                    </tr>
                                    <tr>
                                        <td>Стоимость:</td>
                                        <td id="metric-cost">-</td>
                                    </tr>
                                    <tr>
                                        <td>Запросы:</td>
                                        <td id="metric-requests">-</td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                        
                        <div class="ricochet-tab-content" id="tab-events">
                            <div id="events-list" class="events-list"></div>
                        </div>
                        
                        <div class="ricochet-tab-content" id="tab-logs">
                            <pre id="chain-logs" class="chain-logs"></pre>
                        </div>
                    </div>
                </div>
            </div>
        `;
        
        this.bindEvents();
        this.loadChains();
    }
    
    /**
     * Привязывает обработчики событий
     */
    bindEvents() {
        // Кнопки управления панелью
        document.getElementById('chain-monitor-refresh').addEventListener('click', () => this.refresh());
        document.getElementById('chain-monitor-close').addEventListener('click', () => this.close());
        
        // Выбор цепочки
        document.getElementById('chain-monitor-chain-select').addEventListener('change', (e) => {
            this.setChain(e.target.value);
        });
        
        // Кнопки управления цепочкой
        document.getElementById('chain-monitor-start').addEventListener('click', () => this.startChain());
        document.getElementById('chain-monitor-pause').addEventListener('click', () => this.pauseChain());
        document.getElementById('chain-monitor-stop').addEventListener('click', () => this.stopChain());
        
        // Табы
        document.querySelectorAll('.tab-button').forEach(button => {
            button.addEventListener('click', (e) => {
                // Убираем активный класс у всех табов
                document.querySelectorAll('.tab-button').forEach(btn => btn.classList.remove('active'));
                document.querySelectorAll('.ricochet-tab-content').forEach(content => content.classList.remove('active'));
                
                // Активируем выбранный таб
                e.target.classList.add('active');
                document.getElementById('tab-' + e.target.dataset.tab).classList.add('active');
            });
        });
    }
    
    /**
     * Загружает список доступных цепочек
     */
    async loadChains() {
        try {
            const response = await this.mcp.send('chain_list', {});
            if (response.status === 'success' && response.data.chains) {
                const select = document.getElementById('chain-monitor-chain-select');
                
                // Очищаем текущие опции, кроме первой
                while (select.options.length > 1) {
                    select.remove(1);
                }
                
                // Добавляем новые опции
                response.data.chains.forEach(chain => {
                    const option = document.createElement('option');
                    option.value = chain.id;
                    option.textContent = chain.name;
                    select.appendChild(option);
                });
            }
        } catch (error) {
            console.error('Ошибка при загрузке списка цепочек:', error);
        }
    }
    
    /**
     * Выбирает цепочку для мониторинга
     * @param {String} chainId - ID цепочки
     */
    async setChain(chainId) {
        // Останавливаем текущий мониторинг
        this.stopMonitoring();
        
        if (!chainId) {
            this.currentChainId = null;
            this.updateUI({});
            return;
        }
        
        this.currentChainId = chainId;
        
        try {
            // Получаем информацию о цепочке
            const response = await this.mcp.send('chain_progress', {
                chain_id: chainId
            });
            
            if (response.status === 'success') {
                this.updateUI(response.data);
                this.startMonitoring();
            } else {
                console.error('Ошибка при получении информации о цепочке:', response.error);
            }
        } catch (error) {
            console.error('Ошибка при получении информации о цепочке:', error);
        }
    }
    
    /**
     * Запускает мониторинг цепочки
     */
    startMonitoring() {
        if (!this.currentChainId) return;
        
        // Запускаем интервал обновления
        this.refreshInterval = setInterval(() => this.refresh(), this.refreshRate);
        
        // Запускаем мониторинг событий
        this.mcp.send('chain_monitor', {
            chain_id: this.currentChainId,
            include_history: true,
            refresh_rate: this.refreshRate
        }).then(response => {
            if (response.status === 'success') {
                // Отображаем историю событий
                if (response.data.events) {
                    this.updateEvents(response.data.events);
                }
            }
        }).catch(error => {
            console.error('Ошибка при запуске мониторинга:', error);
        });
    }
    
    /**
     * Останавливает мониторинг цепочки
     */
    stopMonitoring() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
        
        if (this.currentChainId) {
            this.mcp.send('chain_monitor_stop', {
                chain_id: this.currentChainId
            }).catch(error => {
                console.error('Ошибка при остановке мониторинга:', error);
            });
        }
    }
    
    /**
     * Обновляет информацию о цепочке
     */
    async refresh() {
        if (!this.currentChainId) return;
        
        try {
            // Получаем обновленную информацию о цепочке
            const response = await this.mcp.send('chain_progress', {
                chain_id: this.currentChainId
            });
            
            if (response.status === 'success') {
                this.updateUI(response.data);
            }
            
            // Обновляем визуализацию
            const vizResponse = await this.mcp.send('chain_visualization', {
                chain_id: this.currentChainId,
                format: 'unicode',
                show_progress: true,
                show_tasks: true,
                show_metrics: true
            });
            
            if (vizResponse.status === 'success') {
                document.getElementById('chain-monitor-visualization').innerHTML = `
                    <pre class="chain-visualization-content">${vizResponse.data.visualization}</pre>
                `;
            }
        } catch (error) {
            console.error('Ошибка при обновлении информации о цепочке:', error);
        }
    }
    
    /**
     * Обновляет UI на основе данных о цепочке
     * @param {Object} data - Данные о цепочке
     */
    updateUI(data) {
        if (!data || !data.chain_id) {
            // Сбрасываем UI в исходное состояние
            document.getElementById('chain-monitor-status').innerHTML = `
                <span class="chain-name">Не выбрана</span>
                <span class="chain-status">-</span>
                <span class="chain-progress">0%</span>
            `;
            
            document.getElementById('chain-monitor-visualization').innerHTML = `
                <div class="placeholder">Выберите цепочку для отображения</div>
            `;
            
            document.getElementById('metric-elapsed').textContent = '-';
            document.getElementById('metric-tokens').textContent = '-';
            document.getElementById('metric-cost').textContent = '-';
            document.getElementById('metric-requests').textContent = '-';
            
            document.getElementById('chain-monitor-start').disabled = true;
            document.getElementById('chain-monitor-pause').disabled = true;
            document.getElementById('chain-monitor-stop').disabled = true;
            
            return;
        }
        
        // Обновляем статус
        document.getElementById('chain-monitor-status').innerHTML = `
            <span class="chain-name">${data.chain_name}</span>
            <span class="chain-status ${data.status}">${this.translateStatus(data.status)}</span>
            <span class="chain-progress">${Math.round(data.progress * 100)}%</span>
        `;
        
        // Обновляем метрики
        document.getElementById('metric-elapsed').textContent = data.elapsed_time || '-';
        
        if (data.metrics) {
            const tokens = `${data.metrics.tokens_input || 0}/${data.metrics.tokens_output || 0}`;
            document.getElementById('metric-tokens').textContent = tokens;
            document.getElementById('metric-cost').textContent = `$${data.metrics.total_cost?.toFixed(4) || '0.0000'}`;
            document.getElementById('metric-requests').textContent = data.metrics.requests_count || '0';
        }
        
        // Обновляем состояние кнопок
        const isRunning = data.status === 'running';
        const isPaused = data.status === 'paused';
        const isStopped = data.status === 'stopped' || data.status === 'completed' || data.status === 'error';
        
        document.getElementById('chain-monitor-start').disabled = isRunning || isStopped;
        document.getElementById('chain-monitor-pause').disabled = !isRunning;
        document.getElementById('chain-monitor-stop').disabled = isStopped;
    }
    
    /**
     * Обновляет список событий
     * @param {Array} events - Список событий
     */
    updateEvents(events) {
        if (!events || !events.length) return;
        
        const eventsList = document.getElementById('events-list');
        eventsList.innerHTML = '';
        
        events.forEach(event => {
            const time = new Date(event.timestamp).toLocaleTimeString();
            const typeClass = `event-type-${event.type}`;
            
            const eventElement = document.createElement('div');
            eventElement.className = `event-item ${typeClass}`;
            eventElement.innerHTML = `
                <div class="event-time">${time}</div>
                <div class="event-icon">${this.getEventIcon(event.type)}</div>
                <div class="event-message">${event.message}</div>
            `;
            
            eventsList.appendChild(eventElement);
        });
    }
    
    /**
     * Запускает выполнение цепочки
     */
    async startChain() {
        if (!this.currentChainId) return;
        
        try {
            const response = await this.mcp.send('chain_resume', {
                chain_id: this.currentChainId
            });
            
            if (response.status === 'success') {
                this.refresh();
            } else {
                console.error('Ошибка при запуске цепочки:', response.error);
            }
        } catch (error) {
            console.error('Ошибка при запуске цепочки:', error);
        }
    }
    
    /**
     * Приостанавливает выполнение цепочки
     */
    async pauseChain() {
        if (!this.currentChainId) return;
        
        try {
            const response = await this.mcp.send('chain_pause', {
                chain_id: this.currentChainId,
                reason: 'Приостановлено пользователем'
            });
            
            if (response.status === 'success') {
                this.refresh();
            } else {
                console.error('Ошибка при приостановке цепочки:', response.error);
            }
        } catch (error) {
            console.error('Ошибка при приостановке цепочки:', error);
        }
    }
    
    /**
     * Останавливает выполнение цепочки
     */
    async stopChain() {
        if (!this.currentChainId) return;
        
        try {
            const response = await this.mcp.send('chain_stop', {
                chain_id: this.currentChainId,
                reason: 'Остановлено пользователем'
            });
            
            if (response.status === 'success') {
                this.refresh();
            } else {
                console.error('Ошибка при остановке цепочки:', response.error);
            }
        } catch (error) {
            console.error('Ошибка при остановке цепочки:', error);
        }
    }
    
    /**
     * Закрывает панель мониторинга
     */
    close() {
        this.stopMonitoring();
        this.onClose();
    }
    
    /**
     * Переводит статус на русский язык
     * @param {String} status - Статус на английском
     * @returns {String} Статус на русском
     */
    translateStatus(status) {
        const translations = {
            'running': 'Выполняется',
            'paused': 'Приостановлено',
            'stopped': 'Остановлено',
            'completed': 'Завершено',
            'error': 'Ошибка',
            'pending': 'Ожидание'
        };
        
        return translations[status] || status;
    }
    
    /**
     * Возвращает иконку для типа события
     * @param {String} type - Тип события
     * @returns {String} HTML-код иконки
     */
    getEventIcon(type) {
        switch (type) {
            case 'start':
                return '▶️';
            case 'step':
                return '⏭️';
            case 'complete':
                return '✅';
            case 'error':
                return '❌';
            case 'pause':
                return '⏸️';
            case 'resume':
                return '⏯️';
            case 'stop':
                return '⏹️';
            default:
                return 'ℹ️';
        }
    }
}

// Экспортируем класс
if (typeof module !== 'undefined') {
    module.exports = { ChainMonitorPanel };
} 