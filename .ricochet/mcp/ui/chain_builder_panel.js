/**
 * ChainBuilderPanel - UI-компонент для интерактивного конструирования цепочек в редакторе
 * 
 * Этот компонент предоставляет интерфейс для создания и редактирования цепочек моделей,
 * включая интерактивный выбор моделей, настройку параметров и управление шагами.
 */

class ChainBuilderPanel {
    /**
     * Создает новую панель конструктора цепочек
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
        this.templateId = null;
        this.steps = [];
        this.currentStepIndex = -1;
        
        this.initPanel();
    }
    
    /**
     * Инициализирует панель
     */
    initPanel() {
        if (!this.container) {
            console.error('Контейнер для конструктора цепочек не найден');
            return;
        }
        
        this.container.innerHTML = `
            <div class="ricochet-panel chain-builder-panel">
                <div class="ricochet-panel-header">
                    <h3>Конструктор цепочек</h3>
                    <div class="ricochet-panel-actions">
                        <button id="chain-builder-close" title="Закрыть">✕</button>
                    </div>
                </div>
                
                <div class="ricochet-panel-content">
                    <div id="builder-start-view" class="builder-view active">
                        <h4>Создание новой цепочки</h4>
                        
                        <div class="form-group">
                            <label for="chain-name">Название цепочки:</label>
                            <input type="text" id="chain-name" placeholder="Введите название цепочки">
                        </div>
                        
                        <div class="form-group">
                            <label for="chain-description">Описание:</label>
                            <textarea id="chain-description" placeholder="Введите описание цепочки" rows="3"></textarea>
                        </div>
                        
                        <div class="form-group">
                            <label for="chain-template">Шаблон:</label>
                            <select id="chain-template">
                                <option value="">Без шаблона</option>
                            </select>
                        </div>
                        
                        <div class="builder-actions">
                            <button id="builder-start-button" class="primary-button">Начать конструирование</button>
                        </div>
                    </div>
                    
                    <div id="builder-edit-view" class="builder-view">
                        <div class="builder-header">
                            <h4 id="builder-chain-name">Название цепочки</h4>
                            <span id="builder-step-counter">Шаг 0 из 0</span>
                        </div>
                        
                        <div class="builder-steps-container">
                            <div id="builder-steps-list" class="builder-steps-list"></div>
                            
                            <div class="builder-add-step">
                                <button id="builder-add-step-button">+ Добавить шаг</button>
                            </div>
                        </div>
                        
                        <div id="builder-step-editor" class="builder-step-editor">
                            <h5>Редактирование шага</h5>
                            
                            <div class="form-group">
                                <label for="step-role">Роль модели:</label>
                                <input type="text" id="step-role" placeholder="analyzer, summarizer, integrator...">
                            </div>
                            
                            <div class="form-group">
                                <label for="step-model">Модель:</label>
                                <select id="step-model"></select>
                            </div>
                            
                            <div class="form-group">
                                <label for="step-description">Описание шага:</label>
                                <input type="text" id="step-description" placeholder="Описание шага">
                            </div>
                            
                            <div class="form-group">
                                <label for="step-prompt">Промпт:</label>
                                <textarea id="step-prompt" placeholder="Введите промпт для модели" rows="5"></textarea>
                            </div>
                            
                            <div class="form-group">
                                <label>Параметры:</label>
                                <div class="parameters-container">
                                    <div class="parameter-row">
                                        <label for="param-temperature">Температура:</label>
                                        <input type="range" id="param-temperature" min="0" max="1" step="0.1" value="0.7">
                                        <span id="param-temperature-value">0.7</span>
                                    </div>
                                    <div class="parameter-row">
                                        <label for="param-max-tokens">Макс. токенов:</label>
                                        <input type="number" id="param-max-tokens" min="1" max="8000" value="2000">
                                    </div>
                                </div>
                            </div>
                            
                            <div class="builder-step-actions">
                                <button id="step-save-button" class="primary-button">Сохранить шаг</button>
                                <button id="step-cancel-button">Отмена</button>
                            </div>
                        </div>
                        
                        <div class="builder-actions">
                            <button id="builder-save-button" class="primary-button">Сохранить цепочку</button>
                            <button id="builder-cancel-button">Отменить</button>
                        </div>
                    </div>
                </div>
            </div>
        `;
        
        this.bindEvents();
        this.loadTemplates();
        this.loadModels();
    }
    
    /**
     * Привязывает обработчики событий
     */
    bindEvents() {
        // Кнопки управления панелью
        document.getElementById('chain-builder-close').addEventListener('click', () => this.close());
        
        // События начального экрана
        document.getElementById('builder-start-button').addEventListener('click', () => this.startBuilding());
        
        // События редактора цепочки
        document.getElementById('builder-add-step-button').addEventListener('click', () => this.addNewStep());
        document.getElementById('builder-save-button').addEventListener('click', () => this.saveChain(false));
        document.getElementById('builder-cancel-button').addEventListener('click', () => this.cancelBuilding(false));
        
        // События редактора шага
        document.getElementById('step-save-button').addEventListener('click', () => this.saveStep());
        document.getElementById('step-cancel-button').addEventListener('click', () => this.cancelEditStep());
        
        // Обновление значения температуры при изменении ползунка
        document.getElementById('param-temperature').addEventListener('input', (e) => {
            document.getElementById('param-temperature-value').textContent = e.target.value;
        });
        
        // Добавляем кнопки для тестирования и экспорта/импорта
        const actionsContainer = document.querySelector('.builder-actions');
        
        // Создаем дополнительные кнопки, если их еще нет
        if (!document.getElementById('builder-test-button')) {
            const testButton = document.createElement('button');
            testButton.id = 'builder-test-button';
            testButton.textContent = 'Тестировать';
            testButton.addEventListener('click', () => this.testChain());
            actionsContainer.appendChild(testButton);
        }
        
        if (!document.getElementById('builder-export-button')) {
            const exportButton = document.createElement('button');
            exportButton.id = 'builder-export-button';
            exportButton.textContent = 'Экспорт';
            exportButton.addEventListener('click', () => this.exportChain());
            actionsContainer.appendChild(exportButton);
        }
        
        if (!document.getElementById('builder-import-button')) {
            const importButton = document.createElement('button');
            importButton.id = 'builder-import-button';
            importButton.textContent = 'Импорт';
            importButton.addEventListener('click', () => this.importChain());
            actionsContainer.appendChild(importButton);
        }
    }
    
    /**
     * Загружает список доступных шаблонов цепочек
     */
    async loadTemplates() {
        try {
            const response = await this.mcp.send('chain_builder_templates', {});
            
            if (response.status === 'success') {
                const templates = response.data.templates || [];
                const templateSelect = document.getElementById('chain-template');
                
                // Очищаем текущие опции, оставляя только дефолтную
                while (templateSelect.options.length > 1) {
                    templateSelect.remove(1);
                }
                
                // Добавляем новые опции
                templates.forEach(template => {
                    const option = document.createElement('option');
                    option.value = template.id;
                    option.textContent = template.name;
                    templateSelect.appendChild(option);
                });
            } else {
                console.error('Ошибка при загрузке шаблонов:', response.error);
            }
        } catch (error) {
            console.error('Ошибка при загрузке шаблонов:', error);
        }
    }
    
    /**
     * Загружает список доступных моделей
     */
    async loadModels() {
        try {
            const response = await this.mcp.send('available_models', {});
            
            if (response.status === 'success') {
                const models = response.data.models || [];
                const modelSelect = document.getElementById('step-model');
                
                // Очищаем текущие опции
                modelSelect.innerHTML = '<option value="">Выберите модель</option>';
                
                // Группируем модели по провайдеру
                const modelsByProvider = {};
                models.forEach(model => {
                    if (!modelsByProvider[model.provider]) {
                        modelsByProvider[model.provider] = [];
                    }
                    modelsByProvider[model.provider].push(model);
                });
                
                // Добавляем новые опции
                Object.keys(modelsByProvider).sort().forEach(provider => {
                    const optgroup = document.createElement('optgroup');
                    optgroup.label = provider;
                    
                    modelsByProvider[provider].sort((a, b) => a.name.localeCompare(b.name)).forEach(model => {
                        const option = document.createElement('option');
                        option.value = `${model.provider}:${model.id}`;
                        option.textContent = model.name;
                        optgroup.appendChild(option);
                    });
                    
                    modelSelect.appendChild(optgroup);
                });
            } else {
                console.error('Ошибка при загрузке моделей:', response.error);
            }
        } catch (error) {
            console.error('Ошибка при загрузке моделей:', error);
        }
    }
    
    /**
     * Начинает процесс конструирования цепочки
     */
    async startBuilding() {
        const chainName = document.getElementById('chain-name').value.trim();
        if (!chainName) {
            alert('Пожалуйста, введите название цепочки');
            return;
        }
        
        const chainDescription = document.getElementById('chain-description').value.trim();
        const templateId = document.getElementById('chain-template').value;
        
        try {
            const response = await this.mcp.send('chain_builder_init', {
                chain_name: chainName,
                chain_description: chainDescription,
                template_id: templateId
            });
            
            if (response.status === 'success') {
                this.sessionId = response.data.session_id;
                this.templateId = templateId;
                
                // Переключаемся на экран редактирования
                document.getElementById('builder-start-view').classList.remove('active');
                document.getElementById('builder-edit-view').classList.add('active');
                
                // Обновляем заголовок
                document.getElementById('builder-chain-name').textContent = chainName;
                
                // Загружаем данные сессии
                await this.loadSessionData();
            } else {
                console.error('Ошибка при создании сессии конструктора:', response.error);
                alert('Ошибка при создании сессии конструктора: ' + response.error);
            }
        } catch (error) {
            console.error('Ошибка при создании сессии конструктора:', error);
            alert('Ошибка при создании сессии конструктора');
        }
    }
    
    /**
     * Загружает данные текущей сессии
     */
    async loadSessionData() {
        if (!this.sessionId) return;
        
        try {
            const response = await this.mcp.send('chain_builder_get_session', {
                session_id: this.sessionId
            });
            
            if (response.status === 'success') {
                this.steps = response.data.steps || [];
                
                // Обновляем счетчик шагов
                document.getElementById('builder-step-counter').textContent = `Шаг ${response.data.current_step} из ${this.steps.length}`;
                
                // Обновляем список шагов
                this.renderSteps();
                
                // Если есть шаги, выбираем первый
                if (this.steps.length > 0) {
                    this.editStep(0);
                }
            } else {
                console.error('Ошибка при загрузке данных сессии:', response.error);
            }
        } catch (error) {
            console.error('Ошибка при загрузке данных сессии:', error);
        }
    }
    
    /**
     * Обновляет отображение шагов цепочки
     */
    renderSteps() {
        const stepsContainer = document.getElementById('builder-steps-list');
        stepsContainer.innerHTML = '';
        
        if (this.steps.length === 0) {
            stepsContainer.innerHTML = '<div class="builder-empty-steps">Нет шагов. Нажмите "Добавить шаг" для создания.</div>';
            return;
        }
        
        // Создаем элементы для каждого шага
        this.steps.forEach((step, index) => {
            const stepElement = document.createElement('div');
            stepElement.className = 'builder-step-item';
            stepElement.dataset.index = index;
            
            if (index === this.currentStepIndex) {
                stepElement.classList.add('active');
            }
            
            // Номер и роль шага
            const stepHeader = document.createElement('div');
            stepHeader.className = 'builder-step-header';
            stepHeader.innerHTML = `
                <div class="builder-step-number">${index + 1}</div>
                <div class="builder-step-role">${step.model_role || 'Без роли'}</div>
            `;
            
            // Описание и модель
            const stepInfo = document.createElement('div');
            stepInfo.className = 'builder-step-info';
            stepInfo.innerHTML = `
                <div class="builder-step-description">${step.description || 'Без описания'}</div>
                <div class="builder-step-model">${step.provider}:${step.model_id}</div>
            `;
            
            // Кнопки управления
            const stepActions = document.createElement('div');
            stepActions.className = 'builder-step-actions';
            
            // Кнопка редактирования
            const editButton = document.createElement('button');
            editButton.className = 'builder-step-edit';
            editButton.innerHTML = '<i class="fas fa-edit"></i>';
            editButton.title = 'Редактировать шаг';
            editButton.addEventListener('click', () => this.editStep(index));
            
            // Кнопка удаления
            const deleteButton = document.createElement('button');
            deleteButton.className = 'builder-step-delete';
            deleteButton.innerHTML = '<i class="fas fa-trash"></i>';
            deleteButton.title = 'Удалить шаг';
            deleteButton.addEventListener('click', () => this.deleteStep(index));
            
            // Кнопки перемещения
            const moveUpButton = document.createElement('button');
            moveUpButton.className = 'builder-step-move-up';
            moveUpButton.innerHTML = '<i class="fas fa-arrow-up"></i>';
            moveUpButton.title = 'Переместить вверх';
            moveUpButton.disabled = index === 0;
            moveUpButton.addEventListener('click', () => this.moveStep(index, index - 1));
            
            const moveDownButton = document.createElement('button');
            moveDownButton.className = 'builder-step-move-down';
            moveDownButton.innerHTML = '<i class="fas fa-arrow-down"></i>';
            moveDownButton.title = 'Переместить вниз';
            moveDownButton.disabled = index === this.steps.length - 1;
            moveDownButton.addEventListener('click', () => this.moveStep(index, index + 1));
            
            // Добавляем кнопки в контейнер действий
            stepActions.appendChild(editButton);
            stepActions.appendChild(deleteButton);
            stepActions.appendChild(moveUpButton);
            stepActions.appendChild(moveDownButton);
            
            // Собираем все вместе
            stepElement.appendChild(stepHeader);
            stepElement.appendChild(stepInfo);
            stepElement.appendChild(stepActions);
            
            stepsContainer.appendChild(stepElement);
        });
    }
    
    /**
     * Перемещает шаг вверх или вниз в списке
     * @param {Number} fromIndex - Текущий индекс шага
     * @param {Number} toIndex - Новый индекс шага
     */
    async moveStep(fromIndex, toIndex) {
        if (fromIndex < 0 || fromIndex >= this.steps.length || 
            toIndex < 0 || toIndex >= this.steps.length) {
            return;
        }
        
        try {
            const response = await this.mcp.send('chain_builder_move_step', {
                session_id: this.sessionId,
                from_index: fromIndex,
                to_index: toIndex
            });
            
            if (response.status === 'success') {
                // Обновляем текущий индекс, если он был перемещен
                if (this.currentStepIndex === fromIndex) {
                    this.currentStepIndex = toIndex;
                } else if (this.currentStepIndex === toIndex) {
                    this.currentStepIndex = fromIndex;
                }
                
                // Обновляем данные сессии
                await this.loadSessionData();
            } else {
                console.error('Ошибка при перемещении шага:', response.error);
                alert('Ошибка при перемещении шага: ' + response.error);
            }
        } catch (error) {
            console.error('Ошибка при перемещении шага:', error);
            alert('Ошибка при перемещении шага');
        }
    }
    
    /**
     * Добавляет новый шаг в цепочку
     */
    addNewStep() {
        // Очищаем форму редактирования шага
        document.getElementById('step-role').value = '';
        document.getElementById('step-model').value = '';
        document.getElementById('step-description').value = '';
        document.getElementById('step-prompt').value = '';
        document.getElementById('param-temperature').value = '0.7';
        document.getElementById('param-temperature-value').textContent = '0.7';
        document.getElementById('param-max-tokens').value = '2000';
        
        // Устанавливаем индекс нового шага
        this.currentStepIndex = this.steps.length;
        
        // Показываем редактор шага
        document.getElementById('builder-step-editor').style.display = 'block';
    }
    
    /**
     * Редактирует существующий шаг
     * @param {Number} index - Индекс шага
     */
    editStep(index) {
        if (index < 0 || index >= this.steps.length) return;
        
        const step = this.steps[index];
        this.currentStepIndex = index;
        
        // Заполняем форму редактирования шага
        document.getElementById('step-role').value = step.model_role || '';
        document.getElementById('step-description').value = step.description || '';
        document.getElementById('step-prompt').value = step.prompt || '';
        
        // Устанавливаем значение модели
        const modelSelect = document.getElementById('step-model');
        const modelValue = `${step.provider}:${step.model_id}`;
        for (let i = 0; i < modelSelect.options.length; i++) {
            if (modelSelect.options[i].value === modelValue) {
                modelSelect.selectedIndex = i;
                break;
            }
        }
        
        // Устанавливаем параметры
        if (step.parameters) {
            if (step.parameters.temperature !== undefined) {
                const tempValue = step.parameters.temperature;
                document.getElementById('param-temperature').value = tempValue;
                document.getElementById('param-temperature-value').textContent = tempValue;
            }
            
            if (step.parameters.max_tokens !== undefined) {
                document.getElementById('param-max-tokens').value = step.parameters.max_tokens;
            }
        }
        
        // Показываем редактор шага
        document.getElementById('builder-step-editor').style.display = 'block';
        
        // Обновляем визуальное выделение текущего шага
        document.querySelectorAll('.builder-step-item').forEach(item => {
            item.classList.remove('active');
            if (parseInt(item.dataset.index) === index) {
                item.classList.add('active');
            }
        });
    }
    
    /**
     * Сохраняет изменения в шаге
     */
    async saveStep() {
        const role = document.getElementById('step-role').value.trim();
        if (!role) {
            alert('Пожалуйста, укажите роль модели');
            return;
        }
        
        const modelSelect = document.getElementById('step-model');
        if (modelSelect.value === '') {
            alert('Пожалуйста, выберите модель');
            return;
        }
        
        const [provider, modelId] = modelSelect.value.split(':');
        const description = document.getElementById('step-description').value.trim();
        const prompt = document.getElementById('step-prompt').value.trim();
        
        if (!prompt) {
            alert('Пожалуйста, введите промпт для модели');
            return;
        }
        
        // Собираем параметры
        const temperature = parseFloat(document.getElementById('param-temperature').value);
        const maxTokens = parseInt(document.getElementById('param-max-tokens').value);
        
        const parameters = {
            temperature: temperature,
            max_tokens: maxTokens
        };
        
        // Определяем, это новый шаг или редактирование существующего
        const isNewStep = this.currentStepIndex === this.steps.length;
        
        try {
            let response;
            
            if (isNewStep) {
                // Добавляем новый шаг
                response = await this.mcp.send('chain_builder_add_step', {
                    session_id: this.sessionId,
                    step_index: this.currentStepIndex,
                    model_role: role,
                    model_id: modelId,
                    provider: provider,
                    description: description,
                    prompt: prompt,
                    parameters: parameters
                });
            } else {
                // Редактируем существующий шаг
                response = await this.mcp.send('chain_builder_edit_step', {
                    session_id: this.sessionId,
                    step_index: this.currentStepIndex,
                    model_role: role,
                    model_id: modelId,
                    provider: provider,
                    description: description,
                    prompt: prompt,
                    parameters: parameters
                });
            }
            
            if (response.status === 'success') {
                // Обновляем данные сессии
                await this.loadSessionData();
                
                // Скрываем редактор шага
                document.getElementById('builder-step-editor').style.display = 'none';
            } else {
                console.error('Ошибка при сохранении шага:', response.error);
                alert('Ошибка при сохранении шага: ' + response.error);
            }
        } catch (error) {
            console.error('Ошибка при сохранении шага:', error);
            alert('Ошибка при сохранении шага');
        }
    }
    
    /**
     * Отменяет редактирование шага
     */
    cancelEditStep() {
        // Скрываем редактор шага
        document.getElementById('builder-step-editor').style.display = 'none';
        this.currentStepIndex = -1;
        
        // Снимаем выделение со всех шагов
        document.querySelectorAll('.builder-step-item').forEach(item => {
            item.classList.remove('active');
        });
    }
    
    /**
     * Удаляет шаг из цепочки
     * @param {Number} index - Индекс шага
     */
    async deleteStep(index) {
        if (index < 0 || index >= this.steps.length) return;
        
        if (!confirm(`Вы уверены, что хотите удалить шаг ${index + 1}?`)) {
            return;
        }
        
        try {
            const response = await this.mcp.send('chain_builder_remove_step', {
                session_id: this.sessionId,
                step_index: index
            });
            
            if (response.status === 'success') {
                // Обновляем данные сессии
                await this.loadSessionData();
                
                // Если удаляемый шаг был текущим, скрываем редактор
                if (this.currentStepIndex === index) {
                    document.getElementById('builder-step-editor').style.display = 'none';
                    this.currentStepIndex = -1;
                }
            } else {
                console.error('Ошибка при удалении шага:', response.error);
                alert('Ошибка при удалении шага: ' + response.error);
            }
        } catch (error) {
            console.error('Ошибка при удалении шага:', error);
            alert('Ошибка при удалении шага');
        }
    }
    
    /**
     * Сохраняет цепочку
     * @param {Boolean} closeAfterSave - Закрыть панель после сохранения
     */
    async saveChain(closeAfterSave = false) {
        if (!this.sessionId) return;
        
        if (this.steps.length === 0) {
            alert('Пожалуйста, добавьте хотя бы один шаг в цепочку');
            return;
        }
        
        try {
            const response = await this.mcp.send('chain_builder_complete', {
                session_id: this.sessionId,
                save: true
            });
            
            if (response.status === 'success') {
                alert('Цепочка успешно сохранена!');
                
                // Вызываем коллбэк о создании цепочки
                this.onChainCreated(response.data.chain_id);
                
                // Сбрасываем состояние
                this.sessionId = null;
                this.templateId = null;
                this.steps = [];
                this.currentStepIndex = -1;
                
                // Переключаемся на начальный экран, если не нужно закрывать
                if (!closeAfterSave) {
                    document.getElementById('builder-edit-view').classList.remove('active');
                    document.getElementById('builder-start-view').classList.add('active');
                    
                    // Очищаем форму
                    document.getElementById('chain-name').value = '';
                    document.getElementById('chain-description').value = '';
                    document.getElementById('chain-template').selectedIndex = 0;
                } else {
                    this.onClose();
                }
            } else {
                console.error('Ошибка при сохранении цепочки:', response.error);
                alert('Ошибка при сохранении цепочки: ' + response.error);
            }
        } catch (error) {
            console.error('Ошибка при сохранении цепочки:', error);
            alert('Ошибка при сохранении цепочки');
        }
    }
    
    /**
     * Отменяет создание цепочки
     * @param {Boolean} closeAfter - Закрыть панель после отмены
     */
    async cancelBuilding(closeAfter = false) {
        if (!this.sessionId) return;
        
        if (this.steps.length > 0 && !confirm('Вы уверены, что хотите отменить создание цепочки? Все несохраненные изменения будут потеряны.')) {
            return;
        }
        
        try {
            await this.mcp.send('chain_builder_complete', {
                session_id: this.sessionId,
                save: false
            });
            
            // Сбрасываем состояние
            this.sessionId = null;
            this.templateId = null;
            this.steps = [];
            this.currentStepIndex = -1;
            
            // Переключаемся на начальный экран, если не нужно закрывать
            if (!closeAfter) {
                document.getElementById('builder-edit-view').classList.remove('active');
                document.getElementById('builder-start-view').classList.add('active');
                
                // Очищаем форму
                document.getElementById('chain-name').value = '';
                document.getElementById('chain-description').value = '';
                document.getElementById('chain-template').selectedIndex = 0;
            } else {
                this.onClose();
            }
        } catch (error) {
            console.error('Ошибка при отмене создания цепочки:', error);
        }
    }
    
    /**
     * Закрывает панель конструктора
     */
    close() {
        if (this.sessionId) {
            // Если есть активная сессия, предлагаем сохранить изменения
            if (confirm('Сохранить текущую цепочку перед выходом?')) {
                this.saveChain(true);
            } else {
                this.cancelBuilding(true);
            }
        }
        
        this.onClose();
    }
    
    /**
     * Запускает тестирование цепочки
     */
    async testChain() {
        if (!this.sessionId || this.steps.length === 0) {
            alert('Невозможно протестировать цепочку без шагов');
            return;
        }
        
        try {
            const testText = prompt('Введите текст для тестирования цепочки:');
            if (!testText) return;
            
            // Скрываем редактор шага, если он был открыт
            document.getElementById('builder-step-editor').style.display = 'none';
            this.currentStepIndex = -1;
            
            const testDialog = document.createElement('div');
            testDialog.className = 'builder-test-dialog';
            testDialog.innerHTML = `
                <div class="builder-test-header">
                    <h3>Тестирование цепочки</h3>
                    <button class="builder-test-close"><i class="fas fa-times"></i></button>
                </div>
                <div class="builder-test-content">
                    <div class="builder-test-input">
                        <h4>Входной текст:</h4>
                        <div class="builder-test-input-text">${testText}</div>
                    </div>
                    <div class="builder-test-status">
                        Выполняется тестирование...
                        <div class="builder-test-progress"></div>
                    </div>
                    <div class="builder-test-results"></div>
                </div>
            `;
            
            document.body.appendChild(testDialog);
            
            // Добавляем обработчик для закрытия диалога
            const closeButton = testDialog.querySelector('.builder-test-close');
            closeButton.addEventListener('click', () => {
                testDialog.remove();
            });
            
            // Запускаем тестирование
            const response = await this.mcp.send('chain_builder_test', {
                session_id: this.sessionId,
                input_text: testText
            });
            
            if (response.status === 'success') {
                // Получаем идентификатор тестового запуска
                const testId = response.data.test_id;
                
                // Обновляем статус
                const statusElement = testDialog.querySelector('.builder-test-status');
                const progressElement = testDialog.querySelector('.builder-test-progress');
                const resultsElement = testDialog.querySelector('.builder-test-results');
                
                // Начинаем отслеживать прогресс
                const checkProgress = async () => {
                    try {
                        const progressResponse = await this.mcp.send('chain_test_status', {
                            test_id: testId
                        });
                        
                        if (progressResponse.status === 'success') {
                            const testData = progressResponse.data;
                            
                            // Обновляем прогресс
                            const progress = testData.progress || 0;
                            progressElement.style.width = `${progress}%`;
                            
                            // Проверяем, завершено ли тестирование
                            if (testData.status === 'completed') {
                                statusElement.innerHTML = 'Тестирование завершено';
                                
                                // Отображаем результаты
                                resultsElement.innerHTML = '';
                                
                                testData.results.forEach((result, idx) => {
                                    const resultElement = document.createElement('div');
                                    resultElement.className = 'builder-test-result-item';
                                    resultElement.innerHTML = `
                                        <h4>Шаг ${idx + 1}: ${this.steps[idx].model_role}</h4>
                                        <div class="builder-test-result-content">${result.output_text}</div>
                                    `;
                                    resultsElement.appendChild(resultElement);
                                });
                                
                                return;
                            } else if (testData.status === 'error') {
                                statusElement.innerHTML = `Ошибка при тестировании: ${testData.error}`;
                                return;
                            }
                            
                            // Продолжаем проверять прогресс
                            setTimeout(checkProgress, 1000);
                        } else {
                            statusElement.innerHTML = `Ошибка при получении статуса: ${progressResponse.error}`;
                        }
                    } catch (error) {
                        console.error('Ошибка при получении статуса тестирования:', error);
                        statusElement.innerHTML = 'Ошибка при получении статуса тестирования';
                    }
                };
                
                // Запускаем первую проверку
                setTimeout(checkProgress, 1000);
            } else {
                testDialog.querySelector('.builder-test-status').innerHTML = `Ошибка при запуске тестирования: ${response.error}`;
            }
        } catch (error) {
            console.error('Ошибка при тестировании цепочки:', error);
            alert('Ошибка при тестировании цепочки');
        }
    }
    
    /**
     * Экспортирует цепочку в файл
     */
    async exportChain() {
        if (!this.sessionId || this.steps.length === 0) {
            alert('Невозможно экспортировать пустую цепочку');
            return;
        }
        
        try {
            const response = await this.mcp.send('chain_builder_export', {
                session_id: this.sessionId
            });
            
            if (response.status === 'success') {
                // Создаем объект для экспорта
                const exportData = response.data;
                const exportJson = JSON.stringify(exportData, null, 2);
                
                // Создаем элемент для скачивания
                const fileName = `chain_${exportData.chain_name.replace(/\s+/g, '_').toLowerCase()}.json`;
                const blob = new Blob([exportJson], { type: 'application/json' });
                const url = URL.createObjectURL(blob);
                
                const a = document.createElement('a');
                a.href = url;
                a.download = fileName;
                a.style.display = 'none';
                
                document.body.appendChild(a);
                a.click();
                
                // Очищаем
                setTimeout(() => {
                    document.body.removeChild(a);
                    URL.revokeObjectURL(url);
                }, 100);
            } else {
                console.error('Ошибка при экспорте цепочки:', response.error);
                alert('Ошибка при экспорте цепочки: ' + response.error);
            }
        } catch (error) {
            console.error('Ошибка при экспорте цепочки:', error);
            alert('Ошибка при экспорте цепочки');
        }
    }
    
    /**
     * Импортирует цепочку из файла
     */
    importChain() {
        // Создаем элемент для загрузки файла
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = 'application/json';
        input.style.display = 'none';
        
        input.addEventListener('change', async (event) => {
            if (!event.target.files || !event.target.files[0]) return;
            
            const file = event.target.files[0];
            const reader = new FileReader();
            
            reader.onload = async (e) => {
                try {
                    const importData = JSON.parse(e.target.result);
                    
                    // Проверяем, что это файл цепочки
                    if (!importData.chain_name || !importData.steps || !Array.isArray(importData.steps)) {
                        alert('Некорректный формат файла цепочки');
                        return;
                    }
                    
                    // Запрашиваем подтверждение
                    if (this.sessionId && this.steps.length > 0) {
                        if (!confirm('У вас есть несохраненная цепочка. Импорт перезапишет текущие изменения. Продолжить?')) {
                            return;
                        }
                        
                        // Отменяем текущую сессию
                        await this.cancelBuilding(false);
                    }
                    
                    // Создаем новую сессию с импортированными данными
                    const response = await this.mcp.send('chain_builder_import', {
                        import_data: importData
                    });
                    
                    if (response.status === 'success') {
                        this.sessionId = response.data.session_id;
                        
                        // Загружаем данные сессии
                        await this.loadSessionData();
                        
                        // Переключаемся на редактирование
                        document.getElementById('builder-start-view').classList.remove('active');
                        document.getElementById('builder-edit-view').classList.add('active');
                    } else {
                        console.error('Ошибка при импорте цепочки:', response.error);
                        alert('Ошибка при импорте цепочки: ' + response.error);
                    }
                } catch (error) {
                    console.error('Ошибка при чтении файла:', error);
                    alert('Ошибка при чтении файла');
                }
            };
            
            reader.readAsText(file);
            
            // Очищаем инпут
            document.body.removeChild(input);
        });
        
        document.body.appendChild(input);
        input.click();
    }

    /**
     * Выполняет автоматический выбор моделей для всех шагов цепочки
     * на основе ролей и типов шагов
     */
    async autoSelectModels() {
        try {
            // Показываем индикатор загрузки
            this.showLoading('Выбор оптимальных моделей для шагов цепочки...');
            
            // Отправляем запрос на сервер
            const response = await this.mcp.send('auto_select_models', {
                chain_id: this.currentChain.id
            });
            
            // Проверяем ответ
            if (response && response.status === 'success' && response.data) {
                const data = response.data;
                
                if (data.success) {
                    // Обновляем цепочку
                    await this.loadChain(this.currentChain.id);
                    
                    // Показываем сообщение об успешном выборе моделей
                    this.showSuccessMessage(`Выбраны оптимальные модели для ${data.steps.length} шагов`);
                    
                    // Формируем детальный отчет
                    const detailsMessage = document.createElement('div');
                    detailsMessage.className = 'model-selection-details';
                    
                    // Заголовок
                    const header = document.createElement('h3');
                    header.textContent = 'Выбранные модели для шагов цепочки';
                    detailsMessage.appendChild(header);
                    
                    // Список выбранных моделей
                    const list = document.createElement('ul');
                    
                    data.steps.forEach(step => {
                        const item = document.createElement('li');
                        item.innerHTML = `
                            <strong>${step.step_name}</strong>: 
                            <span class="role-badge">${step.selected_role}</span>
                            <span class="model-info">${step.selected_model} (${step.selected_provider})</span>
                        `;
                        list.appendChild(item);
                    });
                    
                    detailsMessage.appendChild(list);
                    
                    // Добавляем детальный отчет
                    const container = document.querySelector('.chain-builder-container');
                    const existingDetails = container.querySelector('.model-selection-details');
                    
                    if (existingDetails) {
                        container.replaceChild(detailsMessage, existingDetails);
                    } else {
                        container.appendChild(detailsMessage);
                    }
                    
                    // Скрываем детальный отчет через 10 секунд
                    setTimeout(() => {
                        const details = container.querySelector('.model-selection-details');
                        if (details) {
                            details.classList.add('fading');
                            setTimeout(() => {
                                if (details.parentNode) {
                                    details.parentNode.removeChild(details);
                                }
                            }, 500);
                        }
                    }, 10000);
                } else {
                    this.showError(data.message || 'Не удалось выбрать модели для шагов');
                }
            } else {
                this.showError(response && response.error ? response.error : 'Ошибка при выборе моделей');
            }
        } catch (error) {
            console.error('Ошибка при автоматическом выборе моделей:', error);
            this.showError('Ошибка при автоматическом выборе моделей: ' + (error.message || error));
        } finally {
            // Скрываем индикатор загрузки
            this.hideLoading();
        }
    }

    /**
     * Показывает сообщение об успешном действии
     * @param {string} message - Текст сообщения
     */
    showSuccessMessage(message) {
        const errorElement = document.getElementById('chain-builder-error');
        if (!errorElement) return;
        
        errorElement.textContent = message;
        errorElement.classList.remove('error-message');
        errorElement.classList.add('success-message');
        
        // Скрываем сообщение через 3 секунды
        setTimeout(() => {
            if (errorElement) {
                errorElement.textContent = '';
                errorElement.classList.remove('success-message');
                errorElement.classList.add('error-message');
            }
        }, 3000);
    }

    /**
     * Показывает сообщение об ошибке
     * @param {string} message - Текст сообщения
     */
    showError(message) {
        const errorElement = document.getElementById('chain-builder-error');
        if (!errorElement) return;
        
        errorElement.textContent = message;
        errorElement.classList.remove('success-message');
        errorElement.classList.add('error-message');
        
        // Скрываем сообщение через 3 секунды
        setTimeout(() => {
            if (errorElement) {
                errorElement.textContent = '';
                errorElement.classList.remove('error-message');
                errorElement.classList.add('success-message');
            }
        }, 3000);
    }

    /**
     * Показывает индикатор загрузки
     * @param {string} message - Текст сообщения
     */
    showLoading(message) {
        const loader = document.getElementById('chain-builder-loader');
        if (!loader) return;
        
        loader.style.display = 'flex';
        const textElement = loader.querySelector('.loader-text');
        if (textElement) {
            textElement.textContent = message;
        }
    }

    /**
     * Скрывает индикатор загрузки
     */
    hideLoading() {
        const loader = document.getElementById('chain-builder-loader');
        if (!loader) return;
        
        loader.style.display = 'none';
    }
}

// Экспортируем класс
if (typeof module !== 'undefined') {
    module.exports = { ChainBuilderPanel };
}