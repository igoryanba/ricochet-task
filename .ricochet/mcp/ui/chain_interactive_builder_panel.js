/**
 * ChainInteractiveBuilderPanel - UI-компонент для интерактивного конструирования цепочек в редакторе
 * 
 * Этот компонент предоставляет интерфейс для создания и редактирования цепочек моделей,
 * включая drag-and-drop интерфейс, интерактивный выбор моделей и визуализацию цепочек.
 */

class ChainInteractiveBuilderPanel {
    /**
     * Создает новую панель интерактивного конструктора цепочек
     * @param {Object} options - Параметры инициализации
     * @param {String} options.containerId - ID контейнера для панели
     * @param {Object} options.mcp - Объект MCP-клиента
     * @param {Function} options.onClose - Коллбэк при закрытии панели
     * @param {Function} options.onChainCreated - Коллбэк при создании цепочки
     */
    constructor(options) {
        this.container = document.getElementById(options.containerId);
        this.mcp = options.mcp;
        this.onClose = options.onClose || (() => {});
        this.onChainCreated = options.onChainCreated || (() => {});
        
        this.sessionId = null;
        this.chainId = null;
        this.chainName = 'Новая цепочка';
        this.models = [];
        this.format = 'ui';
        this.editorMode = 'panel';
        
        this.initPanel();
        this.bindEvents();
    }
    
    /**
     * Инициализирует панель
     */
    initPanel() {
        if (!this.container) {
            console.error('Контейнер для интерактивного конструктора цепочек не найден');
            return;
        }
        
        this.container.innerHTML = `
            <div class="ricochet-panel chain-interactive-builder">
                <div class="ricochet-panel-header">
                    <h3>Интерактивный конструктор цепочек</h3>
                    <div class="ricochet-panel-actions">
                        <button id="chain-builder-view-toggle" title="Переключить режим просмотра" class="icon-button">🔄</button>
                        <button id="chain-builder-close" title="Закрыть" class="icon-button">✕</button>
                    </div>
                </div>
                
                <div class="ricochet-panel-content">
                    <div id="builder-control-panel" class="builder-control-panel">
                        <div class="form-group">
                            <label for="chain-name-input">Название цепочки:</label>
                            <input type="text" id="chain-name-input" placeholder="Введите название цепочки" value="${this.chainName}">
                        </div>
                        
                        <div class="form-group builder-actions">
                            <button id="add-model-btn" class="primary-button">
                                <i class="fas fa-plus"></i> Добавить модель
                            </button>
                            <button id="save-chain-btn" class="success-button">
                                <i class="fas fa-save"></i> Сохранить цепочку
                            </button>
                        </div>
                    </div>
                    
                    <div id="builder-workspace" class="builder-workspace">
                        <div id="models-container" class="models-container">
                            <div class="empty-state">
                                <p>Нет моделей в цепочке</p>
                                <p>Нажмите "Добавить модель" для начала работы</p>
                            </div>
                        </div>
                    </div>
                    
                    <div id="model-selection-panel" class="model-selection-panel" style="display: none;">
                        <div class="panel-header">
                            <h4>Выберите модель</h4>
                            <button id="close-model-selection" class="icon-button">✕</button>
                        </div>
                        
                        <div class="form-group">
                            <label for="model-role-input">Роль модели:</label>
                            <input type="text" id="model-role-input" placeholder="analyzer, summarizer, integrator...">
                        </div>
                        
                        <div id="models-list" class="models-list">
                            <div class="loading">Загрузка доступных моделей...</div>
                        </div>
                    </div>
                    
                    <div id="model-edit-panel" class="model-edit-panel" style="display: none;">
                        <div class="panel-header">
                            <h4>Редактирование модели</h4>
                            <button id="close-model-edit" class="icon-button">✕</button>
                        </div>
                        
                        <div class="form-group">
                            <label for="edit-model-role">Роль модели:</label>
                            <input type="text" id="edit-model-role" placeholder="analyzer, summarizer, integrator...">
                        </div>
                        
                        <div class="form-group">
                            <label for="edit-model-prompt">Промпт:</label>
                            <textarea id="edit-model-prompt" placeholder="Введите промпт для модели" rows="5"></textarea>
                        </div>
                        
                        <div class="form-group">
                            <label>Параметры:</label>
                            <div class="parameters-container">
                                <div class="parameter-row">
                                    <label for="edit-param-temperature">Температура:</label>
                                    <input type="range" id="edit-param-temperature" min="0" max="1" step="0.1" value="0.7">
                                    <span id="edit-param-temperature-value">0.7</span>
                                </div>
                                <div class="parameter-row">
                                    <label for="edit-param-max-tokens">Макс. токенов:</label>
                                    <input type="number" id="edit-param-max-tokens" min="1" max="8000" value="2000">
                                </div>
                            </div>
                        </div>
                        
                        <div class="form-actions">
                            <button id="save-model-edit" class="primary-button">Сохранить</button>
                            <button id="cancel-model-edit" class="secondary-button">Отмена</button>
                        </div>
                    </div>
                </div>
                
                <div id="builder-preview-panel" class="builder-preview-panel">
                    <div class="preview-header">
                        <h4>Предпросмотр цепочки</h4>
                        <div class="preview-controls">
                            <button id="refresh-preview" class="icon-button" title="Обновить">🔄</button>
                            <select id="preview-format">
                                <option value="ui">UI</option>
                                <option value="mermaid">Mermaid</option>
                                <option value="text">Текст</option>
                            </select>
                        </div>
                    </div>
                    <div id="preview-container" class="preview-container">
                        <div class="empty-preview">
                            Предпросмотр будет доступен после добавления моделей
                        </div>
                    </div>
                </div>
            </div>
        `;
    }
    
    /**
     * Привязывает обработчики событий
     */
    bindEvents() {
        // Кнопки управления панелью
        document.getElementById('chain-builder-close').addEventListener('click', () => this.close());
        document.getElementById('chain-builder-view-toggle').addEventListener('click', () => this.toggleViewMode());
        
        // Основные действия
        document.getElementById('add-model-btn').addEventListener('click', () => this.openModelSelection());
        document.getElementById('save-chain-btn').addEventListener('click', () => this.saveChain());
        document.getElementById('chain-name-input').addEventListener('input', (e) => {
            this.chainName = e.target.value;
        });
        
        // Выбор модели
        document.getElementById('close-model-selection').addEventListener('click', () => this.closeModelSelection());
        document.getElementById('model-role-input').addEventListener('input', (e) => {
            this.filterModelsByRole(e.target.value);
        });
        
        // Редактирование модели
        document.getElementById('close-model-edit').addEventListener('click', () => this.closeModelEdit());
        document.getElementById('save-model-edit').addEventListener('click', () => this.saveModelEdit());
        document.getElementById('cancel-model-edit').addEventListener('click', () => this.closeModelEdit());
        
        // Предпросмотр
        document.getElementById('refresh-preview').addEventListener('click', () => this.refreshPreview());
        document.getElementById('preview-format').addEventListener('change', (e) => {
            this.format = e.target.value;
            this.refreshPreview();
        });
        
        // Обновление значения температуры при изменении ползунка
        document.getElementById('edit-param-temperature').addEventListener('input', (e) => {
            document.getElementById('edit-param-temperature-value').textContent = e.target.value;
        });
        
        // Делаем модели перетаскиваемыми
        this.initDragAndDrop();
    }
    
    /**
     * Инициализирует интерфейс drag-and-drop для моделей
     */
    initDragAndDrop() {
        // Используем библиотеку Sortable.js или встроенное HTML5 Drag and Drop
        // Для простоты пример с HTML5 Drag and Drop
        const container = document.getElementById('models-container');
        
        // Устанавливаем обработчики событий для контейнера
        container.addEventListener('dragover', (e) => {
            e.preventDefault();
            const afterElement = this.getDragAfterElement(container, e.clientY);
            const draggable = document.querySelector('.dragging');
            if (afterElement == null) {
                container.appendChild(draggable);
            } else {
                container.insertBefore(draggable, afterElement);
            }
        });
        
        // После завершения перетаскивания обновляем порядок моделей
        container.addEventListener('dragend', () => {
            const modelElements = container.querySelectorAll('.model-item');
            this.reorderModels(Array.from(modelElements).map(el => parseInt(el.dataset.position)));
        });
    }
    
    /**
     * Возвращает элемент, после которого нужно вставить перетаскиваемый элемент
     */
    getDragAfterElement(container, y) {
        const draggableElements = [...container.querySelectorAll('.model-item:not(.dragging)')];
        
        return draggableElements.reduce((closest, child) => {
            const box = child.getBoundingClientRect();
            const offset = y - box.top - box.height / 2;
            
            if (offset < 0 && offset > closest.offset) {
                return { offset: offset, element: child };
            } else {
                return closest;
            }
        }, { offset: Number.NEGATIVE_INFINITY }).element;
    }
    
    /**
     * Переупорядочивает модели после перетаскивания
     * @param {Array} newOrder - Новый порядок позиций моделей
     */
    reorderModels(newOrder) {
        if (!this.sessionId) return;
        
        // Отправляем запрос на перемещение каждой модели на новую позицию
        // Для простоты в примере просто обновляем локальные данные
        const newModels = [];
        
        for (let i = 0; i < newOrder.length; i++) {
            const oldPos = newOrder[i];
            if (this.models[oldPos]) {
                const model = { ...this.models[oldPos] };
                newModels.push(model);
            }
        }
        
        this.models = newModels;
        this.renderModels();
        
        // Обновляем предпросмотр
        this.refreshPreview();
    }
    
    /**
     * Открывает панель выбора модели
     */
    async openModelSelection() {
        // Если сессия еще не создана, создаем ее
        if (!this.sessionId) {
            await this.createSession();
        }
        
        const modelSelectionPanel = document.getElementById('model-selection-panel');
        modelSelectionPanel.style.display = 'block';
        
        // Получаем список доступных моделей
        this.loadAvailableModels();
    }
    
    /**
     * Закрывает панель выбора модели
     */
    closeModelSelection() {
        document.getElementById('model-selection-panel').style.display = 'none';
        document.getElementById('model-role-input').value = '';
    }
    
    /**
     * Загружает список доступных моделей
     * @param {String} role - Опциональная роль для фильтрации моделей
     */
    async loadAvailableModels(role = '') {
        const modelsList = document.getElementById('models-list');
        modelsList.innerHTML = '<div class="loading">Загрузка доступных моделей...</div>';
        
        try {
            const response = await this.mcp.send('chain_get_available_models', {
                role: role
            });
            
            if (response.status === 'success' && response.data && response.data.models) {
                this.renderModelsList(response.data.models);
            } else {
                modelsList.innerHTML = '<div class="error">Ошибка при загрузке моделей</div>';
            }
        } catch (error) {
            console.error('Ошибка при загрузке моделей:', error);
            modelsList.innerHTML = '<div class="error">Ошибка при загрузке моделей</div>';
        }
    }
    
    /**
     * Отображает список доступных моделей
     * @param {Array} models - Массив доступных моделей
     */
    renderModelsList(models) {
        const modelsList = document.getElementById('models-list');
        modelsList.innerHTML = '';
        
        if (!models || models.length === 0) {
            modelsList.innerHTML = '<div class="empty-list">Нет доступных моделей</div>';
            return;
        }
        
        models.forEach(model => {
            const modelElement = document.createElement('div');
            modelElement.className = 'model-option';
            if (model.role) {
                modelElement.classList.add('recommended');
            }
            
            modelElement.innerHTML = `
                <div class="model-option-header">
                    <span class="model-name">${model.name}</span>
                    <span class="model-provider">${model.provider}</span>
                    ${model.role ? `<span class="model-role">Рекомендуется для: ${model.role}</span>` : ''}
                </div>
                <div class="model-details">${model.details || ''}</div>
                <button class="select-model-btn">Выбрать</button>
            `;
            
            modelElement.querySelector('.select-model-btn').addEventListener('click', () => {
                this.selectModel(model);
            });
            
            modelsList.appendChild(modelElement);
        });
    }
    
    /**
     * Фильтрует модели по роли
     * @param {String} role - Роль для фильтрации
     */
    filterModelsByRole(role) {
        if (role.trim() === '') {
            this.loadAvailableModels();
            return;
        }
        
        this.loadAvailableModels(role.trim());
    }
    
    /**
     * Выбирает модель и добавляет ее в цепочку
     * @param {Object} model - Выбранная модель
     */
    async selectModel(model) {
        if (!this.sessionId) return;
        
        const role = document.getElementById('model-role-input').value.trim() || model.role || 'default';
        
        try {
            const response = await this.mcp.send('chain_add_model', {
                session_id: this.sessionId,
                provider: model.provider,
                model_id: model.id,
                role: role,
                position: this.models.length
            });
            
            if (response.status === 'success') {
                // Добавляем модель в локальный массив
                this.models.push({
                    id: model.id,
                    provider: model.provider,
                    name: model.name,
                    role: role,
                    position: this.models.length,
                    parameters: {
                        temperature: 0.7,
                        max_tokens: 2000
                    },
                    prompt: ''
                });
                
                // Закрываем панель выбора
                this.closeModelSelection();
                
                // Отображаем модели
                this.renderModels();
                
                // Обновляем предпросмотр
                this.refreshPreview();
            }
        } catch (error) {
            console.error('Ошибка при добавлении модели:', error);
            alert('Не удалось добавить модель: ' + (error.message || 'Неизвестная ошибка'));
        }
    }
    
    /**
     * Отображает модели в рабочей области
     */
    renderModels() {
        const container = document.getElementById('models-container');
        
        if (!this.models || this.models.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <p>Нет моделей в цепочке</p>
                    <p>Нажмите "Добавить модель" для начала работы</p>
                </div>
            `;
            return;
        }
        
        container.innerHTML = '';
        
        this.models.forEach((model, index) => {
            const modelElement = document.createElement('div');
            modelElement.className = 'model-item';
            modelElement.draggable = true;
            modelElement.dataset.position = index;
            
            modelElement.innerHTML = `
                <div class="model-header">
                    <span class="model-role">${model.role}</span>
                    <span class="model-name">${model.name}</span>
                    <span class="model-provider">${model.provider}</span>
                </div>
                <div class="model-actions">
                    <button class="edit-model-btn" title="Редактировать">✎</button>
                    <button class="remove-model-btn" title="Удалить">✕</button>
                </div>
            `;
            
            // Добавляем обработчики событий для drag and drop
            modelElement.addEventListener('dragstart', () => {
                modelElement.classList.add('dragging');
            });
            
            modelElement.addEventListener('dragend', () => {
                modelElement.classList.remove('dragging');
            });
            
            // Обработчики для кнопок действий
            modelElement.querySelector('.edit-model-btn').addEventListener('click', () => {
                this.editModel(index);
            });
            
            modelElement.querySelector('.remove-model-btn').addEventListener('click', () => {
                this.removeModel(index);
            });
            
            container.appendChild(modelElement);
        });
    }
    
    /**
     * Открывает панель редактирования модели
     * @param {Number} index - Индекс модели для редактирования
     */
    editModel(index) {
        if (index < 0 || index >= this.models.length) return;
        
        const model = this.models[index];
        const editPanel = document.getElementById('model-edit-panel');
        
        // Заполняем форму данными модели
        document.getElementById('edit-model-role').value = model.role || '';
        document.getElementById('edit-model-prompt').value = model.prompt || '';
        
        // Устанавливаем параметры
        if (model.parameters) {
            if (model.parameters.temperature !== undefined) {
                const tempValue = model.parameters.temperature;
                document.getElementById('edit-param-temperature').value = tempValue;
                document.getElementById('edit-param-temperature-value').textContent = tempValue;
            }
            
            if (model.parameters.max_tokens !== undefined) {
                document.getElementById('edit-param-max-tokens').value = model.parameters.max_tokens;
            }
        }
        
        // Сохраняем индекс редактируемой модели
        editPanel.dataset.modelIndex = index;
        
        // Показываем панель редактирования
        editPanel.style.display = 'block';
    }
    
    /**
     * Закрывает панель редактирования модели
     */
    closeModelEdit() {
        document.getElementById('model-edit-panel').style.display = 'none';
        document.getElementById('model-edit-panel').removeAttribute('data-model-index');
    }
    
    /**
     * Сохраняет изменения модели
     */
    saveModelEdit() {
        const editPanel = document.getElementById('model-edit-panel');
        const index = parseInt(editPanel.dataset.modelIndex);
        
        if (isNaN(index) || index < 0 || index >= this.models.length) {
            this.closeModelEdit();
            return;
        }
        
        // Получаем данные из формы
        const role = document.getElementById('edit-model-role').value.trim();
        const prompt = document.getElementById('edit-model-prompt').value.trim();
        const temperature = parseFloat(document.getElementById('edit-param-temperature').value);
        const maxTokens = parseInt(document.getElementById('edit-param-max-tokens').value);
        
        // Обновляем модель
        this.models[index].role = role;
        this.models[index].prompt = prompt;
        this.models[index].parameters = {
            temperature: temperature,
            max_tokens: maxTokens
        };
        
        // Закрываем панель редактирования
        this.closeModelEdit();
        
        // Обновляем отображение
        this.renderModels();
        
        // Обновляем предпросмотр
        this.refreshPreview();
    }
    
    /**
     * Удаляет модель из цепочки
     * @param {Number} index - Индекс модели для удаления
     */
    async removeModel(index) {
        if (!this.sessionId || index < 0 || index >= this.models.length) return;
        
        if (!confirm(`Вы уверены, что хотите удалить модель "${this.models[index].name}" из цепочки?`)) {
            return;
        }
        
        try {
            const response = await this.mcp.send('chain_remove_model', {
                session_id: this.sessionId,
                position: index
            });
            
            if (response.status === 'success') {
                // Удаляем модель из локального массива
                this.models.splice(index, 1);
                
                // Обновляем позиции моделей
                this.models.forEach((model, i) => {
                    model.position = i;
                });
                
                // Отображаем модели
                this.renderModels();
                
                // Обновляем предпросмотр
                this.refreshPreview();
            }
        } catch (error) {
            console.error('Ошибка при удалении модели:', error);
            alert('Не удалось удалить модель: ' + (error.message || 'Неизвестная ошибка'));
        }
    }
    
    /**
     * Создает новую сессию конструктора
     */
    async createSession() {
        try {
            const response = await this.mcp.send('chain_interactive_builder', {
                chain_name: this.chainName,
                format: this.format,
                editor_mode: this.editorMode
            });
            
            if (response.status === 'success') {
                this.sessionId = response.data.session_id;
                this.chainName = response.data.chain_name;
                document.getElementById('chain-name-input').value = this.chainName;
                
                // Отображаем предпросмотр
                if (response.data.editor_content) {
                    document.getElementById('preview-container').innerHTML = response.data.editor_content;
                }
                
                return true;
            } else {
                console.error('Ошибка при создании сессии:', response.error);
                return false;
            }
        } catch (error) {
            console.error('Ошибка при создании сессии:', error);
            return false;
        }
    }
    
    /**
     * Сохраняет цепочку
     */
    async saveChain() {
        if (!this.sessionId) {
            await this.createSession();
        }
        
        if (this.models.length === 0) {
            alert('Добавьте хотя бы одну модель в цепочку перед сохранением');
            return;
        }
        
        try {
            const response = await this.mcp.send('chain_save_interactive', {
                session_id: this.sessionId,
                chain_name: this.chainName
            });
            
            if (response.status === 'success') {
                alert('Цепочка успешно сохранена!');
                
                // Вызываем коллбэк о создании цепочки
                if (response.data.chain_id) {
                    this.chainId = response.data.chain_id;
                    this.onChainCreated(response.data.chain_id);
                }
                
                // Закрываем панель
                this.close();
            } else {
                console.error('Ошибка при сохранении цепочки:', response.error);
                alert('Ошибка при сохранении цепочки: ' + (response.error || 'Неизвестная ошибка'));
            }
        } catch (error) {
            console.error('Ошибка при сохранении цепочки:', error);
            alert('Ошибка при сохранении цепочки: ' + (error.message || 'Неизвестная ошибка'));
        }
    }
    
    /**
     * Обновляет предпросмотр цепочки
     */
    async refreshPreview() {
        if (!this.sessionId) return;
        
        const previewContainer = document.getElementById('preview-container');
        previewContainer.innerHTML = '<div class="loading">Загрузка предпросмотра...</div>';
        
        try {
            // В реальной реализации здесь будет вызов MCP для получения актуального предпросмотра
            // Для примера генерируем предпросмотр на основе текущих данных
            
            let previewContent = '';
            
            switch (this.format) {
                case 'mermaid':
                    previewContent = this.generateMermaidPreview();
                    break;
                case 'text':
                    previewContent = this.generateTextPreview();
                    break;
                default:
                    previewContent = this.generateUIPreview();
                    break;
            }
            
            previewContainer.innerHTML = previewContent;
            
            // Если используется mermaid, инициализируем его
            if (this.format === 'mermaid' && window.mermaid) {
                window.mermaid.init(undefined, document.querySelectorAll('.mermaid'));
            }
        } catch (error) {
            console.error('Ошибка при обновлении предпросмотра:', error);
            previewContainer.innerHTML = '<div class="error">Ошибка при загрузке предпросмотра</div>';
        }
    }
    
    /**
     * Генерирует UI-предпросмотр цепочки
     * @returns {String} HTML-разметка предпросмотра
     */
    generateUIPreview() {
        if (this.models.length === 0) {
            return '<div class="empty-preview">Предпросмотр будет доступен после добавления моделей</div>';
        }
        
        let preview = `
            <div class="chain-preview ui-preview">
                <div class="chain-title">${this.chainName}</div>
                <div class="chain-models">
        `;
        
        this.models.forEach((model, index) => {
            preview += `
                <div class="chain-model">
                    <div class="model-box">
                        <div class="model-title">${model.role}</div>
                        <div class="model-details">${model.name}</div>
                    </div>
                    ${index < this.models.length - 1 ? '<div class="model-arrow">→</div>' : ''}
                </div>
            `;
        });
        
        preview += `
                </div>
            </div>
        `;
        
        return preview;
    }
    
    /**
     * Генерирует Mermaid-предпросмотр цепочки
     * @returns {String} HTML-разметка с Mermaid-диаграммой
     */
    generateMermaidPreview() {
        if (this.models.length === 0) {
            return '<div class="empty-preview">Предпросмотр будет доступен после добавления моделей</div>';
        }
        
        let mermaidCode = 'graph LR\n';
        mermaidCode += `    title["${this.chainName}"]\n`;
        
        // Добавляем узлы моделей
        this.models.forEach((model, index) => {
            const nodeId = `model${index}`;
            mermaidCode += `    ${nodeId}["${model.role}<br/>(${model.name})"]\n`;
        });
        
        // Добавляем связи между моделями
        for (let i = 0; i < this.models.length - 1; i++) {
            mermaidCode += `    model${i} --> model${i+1}\n`;
        }
        
        mermaidCode += '    style title fill:#f9f9f9,stroke:#333,stroke-width:1px\n';
        
        return `<div class="mermaid">${mermaidCode}</div>`;
    }
    
    /**
     * Генерирует текстовый предпросмотр цепочки
     * @returns {String} HTML-разметка с текстовым представлением
     */
    generateTextPreview() {
        if (this.models.length === 0) {
            return '<div class="empty-preview">Предпросмотр будет доступен после добавления моделей</div>';
        }
        
        let text = `Цепочка: ${this.chainName}\n`;
        text += `Количество моделей: ${this.models.length}\n\n`;
        
        text += 'Модели в цепочке:\n';
        this.models.forEach((model, index) => {
            text += `${index + 1}. ${model.role} (${model.name}, ${model.provider})\n`;
            if (model.prompt) {
                text += `   Промпт: ${model.prompt.substring(0, 50)}${model.prompt.length > 50 ? '...' : ''}\n`;
            }
            if (model.parameters) {
                text += `   Параметры: температура=${model.parameters.temperature}, макс. токенов=${model.parameters.max_tokens}\n`;
            }
            text += '\n';
        });
        
        return `<pre class="text-preview">${text}</pre>`;
    }
    
    /**
     * Переключает режим отображения редактора
     */
    toggleViewMode() {
        const previewPanel = document.getElementById('builder-preview-panel');
        const workspacePanel = document.getElementById('builder-workspace');
        
        if (previewPanel.classList.contains('expanded')) {
            // Возвращаемся к нормальному виду
            previewPanel.classList.remove('expanded');
            workspacePanel.style.display = 'block';
        } else {
            // Расширяем панель предпросмотра
            previewPanel.classList.add('expanded');
            workspacePanel.style.display = 'none';
        }
    }
    
    /**
     * Закрывает панель конструктора
     */
    close() {
        if (this.sessionId && this.models.length > 0 && !confirm('Вы уверены, что хотите закрыть конструктор? Все несохраненные изменения будут потеряны.')) {
            return;
        }
        
        this.onClose();
    }
} 