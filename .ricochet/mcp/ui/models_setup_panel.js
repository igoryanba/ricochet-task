/**
 * Панель для настройки моделей Ricochet
 * 
 * Предоставляет интерактивный интерфейс для выбора моделей по ролям,
 * настройки параметров моделей и сохранения конфигурации.
 */

// Список всех доступных ролей
const ALL_ROLES = [
    {
        id: 'main',
        displayName: 'Основная модель',
        description: 'Используется для основных задач генерации и обновления'
    },
    {
        id: 'research',
        displayName: 'Исследовательская модель',
        description: 'Используется для анализа данных и исследовательских задач'
    },
    {
        id: 'fallback',
        displayName: 'Резервная модель',
        description: 'Используется при недоступности основной модели'
    },
    {
        id: 'analyzer',
        displayName: 'Модель-анализатор',
        description: 'Анализирует и структурирует входные данные'
    },
    {
        id: 'summarizer',
        displayName: 'Модель-суммаризатор',
        description: 'Создает краткие резюме на основе анализа'
    },
    {
        id: 'integrator',
        displayName: 'Модель-интегратор',
        description: 'Объединяет результаты работы других моделей'
    },
    {
        id: 'extractor',
        displayName: 'Модель-экстрактор',
        description: 'Извлекает ключевую информацию из текста'
    },
    {
        id: 'critic',
        displayName: 'Модель-критик',
        description: 'Проверяет и критически оценивает выходные данные'
    },
    {
        id: 'refiner',
        displayName: 'Модель-улучшатель',
        description: 'Улучшает и дорабатывает результаты других моделей'
    },
    {
        id: 'creator',
        displayName: 'Модель-генератор',
        description: 'Создает новый контент на основе входных данных'
    }
];

// Роли, необходимые для работы Task Master
const TASK_MASTER_ROLES = ['main', 'research', 'fallback'];

// Роли, используемые в цепочках
const CHAIN_ROLES = ['analyzer', 'summarizer', 'integrator', 'extractor', 'critic', 'refiner', 'creator'];

// Возможности моделей по ролям
const ROLE_CAPABILITIES = {
    'main': ['text-generation', 'context-aware'],
    'research': ['text-generation', 'research', 'large-context'],
    'fallback': ['text-generation', 'fast-response'],
    'analyzer': ['classification', 'extraction', 'analysis'],
    'summarizer': ['summarization', 'extraction'],
    'integrator': ['text-generation', 'synthesis', 'large-context'],
    'extractor': ['extraction', 'classification'],
    'critic': ['analysis', 'evaluation'],
    'refiner': ['text-generation', 'editing'],
    'creator': ['text-generation', 'creative']
};

class ModelsSetupPanel {
    constructor() {
        this.roles = [];
        this.selectedRole = null;
        this.currentRoleIndex = 0;
        this.panel = null;
        this.initialized = false;
        this.callbacks = {};
        this.roleCategories = {};
    }

    /**
     * Инициализирует панель настройки моделей
     * @param {Object} options Опции инициализации
     * @param {Array} options.roles Список ролей для настройки, если не указан - все роли
     * @param {Function} options.onSave Callback, вызываемый при сохранении настроек
     * @param {Function} options.onCancel Callback, вызываемый при отмене настройки
     * @param {Function} options.onComplete Callback, вызываемый при завершении настройки
     */
    async initialize(options = {}) {
        this.callbacks = {
            onSave: options.onSave || function() {},
            onCancel: options.onCancel || function() {},
            onComplete: options.onComplete || function() {}
        };

        try {
            // Загружаем роли для настройки
            this.roles = options.roles || ALL_ROLES;
            
            // Если роли не указаны, загружаем все роли
            if (!options.roles || options.roles.length === 0) {
                // Получаем список ролей с сервера
                const response = await mcp.invoke("model_setup", {});
                if (response && response.status === 'success' && response.data && response.data.roles) {
                    this.roles = response.data.roles;
                } else {
                    this.roles = ALL_ROLES;
                }
            }
            
            // Определяем категории ролей
            this.roleCategories = {
                'basic': TASK_MASTER_ROLES,
                'chain': CHAIN_ROLES
            };
            
            // Создаем панель
            this.createPanel();
            this.addStyles();
            this.addEventHandlers();
            
            // Устанавливаем первую роль как активную
            this.currentRoleIndex = 0;
            this.updateRoleView();
            
            // Показываем панель
            this.showPanel();
            
            this.initialized = true;
        } catch (error) {
            console.error("Ошибка инициализации панели настройки моделей:", error);
            this.showError("Не удалось инициализировать панель настройки моделей: " + error.message);
        }
    }

    /**
     * Создает DOM-элементы панели
     */
    createPanel() {
        // Удаляем существующую панель, если есть
        if (this.panel) {
            document.body.removeChild(this.panel);
        }

        // Создаем основной контейнер
        this.panel = document.createElement("div");
        this.panel.className = "ricochet-models-setup-panel";
        this.panel.style.display = "none";

        // Добавляем заголовок
        const header = document.createElement("div");
        header.className = "panel-header";
        header.innerHTML = `
            <h2>Настройка моделей Ricochet</h2>
            <div class="panel-progress">
                <span class="progress-text">Шаг <span id="current-step">1</span> из <span id="total-steps">${this.roles.length}</span>: <span id="current-role-name">Основная модель</span></span>
                <div class="progress-bar">
                    <div class="progress-fill" style="width: ${100 / this.roles.length}%"></div>
                </div>
            </div>
            <button class="close-button" title="Закрыть">×</button>
        `;
        this.panel.appendChild(header);

        // Контейнер для ролей
        const content = document.createElement("div");
        content.className = "panel-content";
        content.innerHTML = `
            <div class="role-description">
                <h3 id="role-title">Основная модель</h3>
                <p id="role-description">Основная модель для генерации контента и обновлений</p>
            </div>
            <div class="model-selector">
                <div class="provider-tabs" id="provider-tabs"></div>
                <div class="models-grid" id="models-grid"></div>
            </div>
            <div class="model-details" id="model-details">
                <div class="model-info">
                    <h4 id="selected-model-name">Выберите модель</h4>
                    <p id="selected-model-description"></p>
                    <div class="model-capabilities" id="model-capabilities"></div>
                    <div class="model-params">
                        <div class="param-group">
                            <label for="temperature">Температура:</label>
                            <input type="range" id="temperature" min="0" max="1" step="0.1" value="0.7">
                            <span id="temperature-value">0.7</span>
                        </div>
                        <div class="param-group">
                            <label for="top-p">Top P:</label>
                            <input type="range" id="top-p" min="0" max="1" step="0.05" value="1.0">
                            <span id="top-p-value">1.0</span>
                        </div>
                    </div>
                </div>
            </div>
        `;
        this.panel.appendChild(content);

        // Кнопки
        const footer = document.createElement("div");
        footer.className = "panel-footer";
        footer.innerHTML = `
            <button id="prev-button" class="panel-button secondary" disabled>Назад</button>
            <button id="next-button" class="panel-button primary">Далее</button>
            <button id="cancel-button" class="panel-button secondary">Отмена</button>
        `;
        this.panel.appendChild(footer);

        // Добавляем панель на страницу
        document.body.appendChild(this.panel);

        // Добавляем стили
        this.addStyles();

        // Добавляем обработчики событий
        this.addEventHandlers();
    }

    /**
     * Добавляет стили для панели
     */
    addStyles() {
        // Проверяем, есть ли уже стили
        if (document.getElementById('ricochet-models-setup-styles')) {
            return;
        }

        const styleSheet = document.createElement("style");
        styleSheet.id = 'ricochet-models-setup-styles';
        styleSheet.textContent = `
            .ricochet-models-setup-panel {
                position: fixed;
                top: 50%;
                left: 50%;
                transform: translate(-50%, -50%);
                width: 800px;
                max-width: 90vw;
                max-height: 90vh;
                background-color: #1e1e1e;
                color: #e0e0e0;
                border-radius: 8px;
                box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
                z-index: 9999;
                display: flex;
                flex-direction: column;
                overflow: hidden;
                font-family: system-ui, -apple-system, sans-serif;
            }

            .panel-header {
                padding: 16px 20px;
                background-color: #2d2d2d;
                border-bottom: 1px solid #3e3e3e;
                display: flex;
                flex-direction: column;
                position: relative;
            }

            .panel-header h2 {
                margin: 0 0 12px 0;
                font-size: 18px;
                font-weight: 600;
            }

            .close-button {
                position: absolute;
                top: 12px;
                right: 16px;
                background: none;
                border: none;
                color: #a0a0a0;
                font-size: 24px;
                cursor: pointer;
                padding: 0;
                width: 24px;
                height: 24px;
                display: flex;
                align-items: center;
                justify-content: center;
                border-radius: 4px;
            }

            .close-button:hover {
                background-color: rgba(255, 255, 255, 0.1);
                color: #ffffff;
            }

            .panel-progress {
                display: flex;
                flex-direction: column;
                gap: 6px;
            }

            .progress-text {
                font-size: 14px;
                color: #cccccc;
            }

            .progress-bar {
                height: 4px;
                background-color: #3e3e3e;
                border-radius: 2px;
                overflow: hidden;
            }

            .progress-fill {
                height: 100%;
                background-color: #007acc;
                transition: width 0.3s ease;
            }

            .panel-content {
                padding: 20px;
                flex: 1;
                overflow-y: auto;
                display: flex;
                flex-direction: column;
                gap: 20px;
            }

            .role-description {
                margin-bottom: 8px;
            }

            .role-description h3 {
                margin: 0 0 8px 0;
                font-size: 16px;
                font-weight: 600;
            }

            .role-description p {
                margin: 0;
                font-size: 14px;
                color: #bbbbbb;
            }

            .model-selector {
                display: flex;
                flex-direction: column;
                gap: 12px;
            }

            .provider-tabs {
                display: flex;
                gap: 2px;
                overflow-x: auto;
                padding-bottom: 6px;
            }

            .provider-tab {
                padding: 8px 12px;
                background-color: #2d2d2d;
                border: 1px solid #3e3e3e;
                border-radius: 4px;
                font-size: 14px;
                cursor: pointer;
                white-space: nowrap;
            }

            .provider-tab.active {
                background-color: #3e3e3e;
                border-color: #5e5e5e;
                color: #ffffff;
            }

            .models-grid {
                display: grid;
                grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
                gap: 12px;
            }

            .model-card {
                background-color: #2a2a2a;
                border: 1px solid #3e3e3e;
                border-radius: 6px;
                padding: 12px;
                cursor: pointer;
                transition: all 0.2s ease;
                display: flex;
                flex-direction: column;
                gap: 8px;
            }

            .model-card:hover {
                background-color: #333333;
                border-color: #5e5e5e;
            }

            .model-card.selected {
                background-color: #0e639c;
                border-color: #1c90e0;
            }

            .model-card-title {
                font-weight: 600;
                font-size: 14px;
                margin: 0;
            }

            .model-card-context {
                font-size: 12px;
                color: #bbbbbb;
                margin: 0;
            }

            .model-card-cost {
                font-size: 12px;
                color: #999999;
                margin-top: auto;
            }

            .model-details {
                background-color: #2a2a2a;
                border: 1px solid #3e3e3e;
                border-radius: 6px;
                padding: 16px;
            }

            .model-details h4 {
                margin: 0 0 8px 0;
                font-size: 16px;
                font-weight: 600;
            }

            .model-details p {
                margin: 0 0 12px 0;
                font-size: 14px;
                color: #cccccc;
            }

            .model-capabilities {
                display: flex;
                flex-wrap: wrap;
                gap: 6px;
                margin-bottom: 16px;
            }

            .model-capability {
                font-size: 12px;
                padding: 4px 8px;
                background-color: #3e3e3e;
                border-radius: 4px;
                color: #cccccc;
            }

            .model-params {
                display: flex;
                flex-direction: column;
                gap: 12px;
            }

            .param-group {
                display: flex;
                align-items: center;
                gap: 8px;
            }

            .param-group label {
                font-size: 14px;
                min-width: 100px;
            }

            .param-group input[type="range"] {
                flex: 1;
            }

            .param-group span {
                font-size: 14px;
                min-width: 40px;
                text-align: right;
            }

            .panel-footer {
                padding: 16px 20px;
                background-color: #2d2d2d;
                border-top: 1px solid #3e3e3e;
                display: flex;
                justify-content: flex-end;
                gap: 12px;
            }

            .panel-button {
                padding: 8px 16px;
                border-radius: 4px;
                font-size: 14px;
                cursor: pointer;
                border: none;
                outline: none;
            }

            .panel-button.primary {
                background-color: #0e639c;
                color: #ffffff;
            }

            .panel-button.primary:hover {
                background-color: #1c90e0;
            }

            .panel-button.secondary {
                background-color: #3e3e3e;
                color: #cccccc;
            }

            .panel-button.secondary:hover {
                background-color: #4e4e4e;
            }

            .panel-button:disabled {
                background-color: #2a2a2a;
                color: #666666;
                cursor: not-allowed;
            }

            .error-message {
                background-color: #5a1d1d;
                color: #ff9999;
                padding: 8px 12px;
                border-radius: 4px;
                margin-bottom: 16px;
                font-size: 14px;
            }
        `;
        document.head.appendChild(styleSheet);
    }

    /**
     * Добавляет обработчики событий для элементов панели
     */
    addEventHandlers() {
        // Закрытие панели
        const closeButton = this.panel.querySelector('.close-button');
        closeButton.addEventListener('click', () => {
            this.hidePanel();
            this.callbacks.onCancel();
        });

        // Кнопка отмены
        const cancelButton = this.panel.querySelector('#cancel-button');
        cancelButton.addEventListener('click', () => {
            this.hidePanel();
            this.callbacks.onCancel();
        });

        // Кнопка "Назад"
        const prevButton = this.panel.querySelector('#prev-button');
        prevButton.addEventListener('click', () => {
            this.navigateToPreviousRole();
        });

        // Кнопка "Далее"
        const nextButton = this.panel.querySelector('#next-button');
        nextButton.addEventListener('click', () => {
            this.navigateToNextRole();
        });

        // Обработчики изменения параметров
        const temperatureSlider = this.panel.querySelector('#temperature');
        const temperatureValue = this.panel.querySelector('#temperature-value');
        temperatureSlider.addEventListener('input', () => {
            temperatureValue.textContent = temperatureSlider.value;
        });

        const topPSlider = this.panel.querySelector('#top-p');
        const topPValue = this.panel.querySelector('#top-p-value');
        topPSlider.addEventListener('input', () => {
            topPValue.textContent = topPSlider.value;
        });

        // Добавляем кнопки интеграции с Task Master
        this.addTaskMasterIntegrationButtons();
    }

    /**
     * Показывает панель
     */
    showPanel() {
        if (this.panel) {
            this.panel.style.display = 'flex';
        }
    }

    /**
     * Скрывает панель
     */
    hidePanel() {
        if (this.panel) {
            this.panel.style.display = 'none';
        }
    }

    /**
     * Обновляет отображение текущей роли и моделей
     */
    updateRoleView() {
        if (this.roles.length === 0 || this.currentRoleIndex < 0 || this.currentRoleIndex >= this.roles.length) {
            return;
        }
        
        const role = this.roles[this.currentRoleIndex];
        
        // Обновляем заголовок и описание роли
        const roleTitle = document.getElementById('role-title');
        const roleDescription = document.getElementById('role-description');
        
        roleTitle.textContent = role.displayName || role.id;
        roleDescription.textContent = role.description || `Настройка модели для роли ${role.id}`;
        
        // Обновляем шаги прогресса
        const progressSteps = document.getElementById('progress-steps');
        progressSteps.innerHTML = '';
        
        // Группируем роли по категориям
        const categoryElements = {
            'basic': document.createElement('div'),
            'chain': document.createElement('div')
        };
        
        categoryElements['basic'].className = 'progress-category';
        categoryElements['basic'].innerHTML = '<span class="category-title">Основные роли</span>';
        
        categoryElements['chain'].className = 'progress-category';
        categoryElements['chain'].innerHTML = '<span class="category-title">Роли для цепочек</span>';
        
        this.roles.forEach((r, index) => {
            const step = document.createElement('div');
            step.className = 'progress-step';
            if (index === this.currentRoleIndex) {
                step.classList.add('active');
            }
            step.dataset.index = index;
            step.textContent = r.displayName || r.id;
            
            step.addEventListener('click', () => {
                this.currentRoleIndex = parseInt(step.dataset.index);
                this.updateRoleView();
            });
            
            // Определяем категорию роли
            let category = 'basic';
            if (CHAIN_ROLES.includes(r.id)) {
                category = 'chain';
            }
            
            categoryElements[category].appendChild(step);
        });
        
        // Добавляем категории в прогресс
        for (const category in categoryElements) {
            if (categoryElements[category].childElementCount > 1) { // >1, потому что у нас есть заголовок
                progressSteps.appendChild(categoryElements[category]);
            }
        }
        
        // Показываем кнопки навигации
        const prevButton = document.getElementById('prev-role');
        const nextButton = document.getElementById('next-role');
        
        prevButton.disabled = this.currentRoleIndex === 0;
        nextButton.textContent = this.currentRoleIndex === this.roles.length - 1 ? 'Завершить' : 'Далее';
        
        // Очищаем текущие провайдеры и показываем первый провайдер
        const providerTabs = document.getElementById('provider-tabs');
        providerTabs.innerHTML = '';
        
        // Обновляем рекомендации моделей для текущей роли
        this.loadRecommendedModels(role.id);
        
        // Показываем модели первого провайдера
        if (role.options && role.options.length > 0) {
            // Группируем модели по провайдеру
            const modelsByProvider = {};
            role.options.forEach(model => {
                if (!modelsByProvider[model.provider]) {
                    modelsByProvider[model.provider] = [];
                }
                modelsByProvider[model.provider].push(model);
            });
            
            // Создаем вкладки для провайдеров
            Object.keys(modelsByProvider).sort().forEach((provider, index) => {
                const tab = document.createElement('div');
                tab.className = 'provider-tab';
                if (index === 0) {
                    tab.classList.add('active');
                }
                tab.dataset.provider = provider;
                tab.textContent = this.getProviderDisplayName(provider);
                
                tab.addEventListener('click', () => {
                    document.querySelectorAll('.provider-tab').forEach(t => t.classList.remove('active'));
                    tab.classList.add('active');
                    this.showModelsForProvider(provider);
                });
                
                providerTabs.appendChild(tab);
            });
            
            // Показываем модели первого провайдера
            const firstProvider = Object.keys(modelsByProvider).sort()[0];
            this.showModelsForProvider(firstProvider);
            
            // Если у роли уже есть выбранная модель, показываем её
            if (role.currentModel) {
                this.updateModelDetails(role.currentModel);
                
                // Находим и выделяем карточку выбранной модели
                const selectedProvider = role.currentModel.provider;
                if (selectedProvider) {
                    // Активируем вкладку провайдера
                    const providerTab = document.querySelector(`.provider-tab[data-provider="${selectedProvider}"]`);
                    if (providerTab) {
                        document.querySelectorAll('.provider-tab').forEach(t => t.classList.remove('active'));
                        providerTab.classList.add('active');
                        this.showModelsForProvider(selectedProvider);
                        
                        // Выделяем карточку модели
                        setTimeout(() => {
                            const modelCard = document.querySelector(`.model-card[data-model-id="${role.currentModel.modelId}"][data-provider="${selectedProvider}"]`);
                            if (modelCard) {
                                document.querySelectorAll('.model-card').forEach(card => card.classList.remove('selected'));
                                modelCard.classList.add('selected');
                            }
                        }, 100);
                    }
                }
            }
        } else {
            // Если нет доступных моделей
            document.getElementById('models-grid').innerHTML = '<div class="no-models">Нет доступных моделей для этой роли</div>';
            document.getElementById('model-details').innerHTML = '';
        }
    }

    /**
     * Показывает модели для выбранного провайдера
     * @param {string} provider Провайдер
     */
    showModelsForProvider(provider) {
        const modelsGrid = this.panel.querySelector('#models-grid');
        modelsGrid.innerHTML = '';

        // Фильтруем модели по провайдеру
        const models = this.selectedRole.Options.filter(model => model.Provider === provider);

        // Определяем выбранную модель
        const currentModel = this.selectedRole.CurrentModel;
        let selectedModel = null;

        // Создаем карточки моделей
        models.forEach(model => {
            const isSelected = currentModel && 
                               currentModel.Provider === model.Provider && 
                               currentModel.ModelID === model.ModelID;
            
            if (isSelected) {
                selectedModel = model;
            }

            const card = document.createElement('div');
            card.className = `model-card ${isSelected ? 'selected' : ''}`;
            card.innerHTML = `
                <p class="model-card-title">${model.DisplayName}</p>
                <p class="model-card-context">Контекст: ${this.formatTokens(model.ContextSize)}</p>
                <p class="model-card-cost">${model.Cost || ''}</p>
            `;
            card.dataset.provider = model.Provider;
            card.dataset.modelId = model.ModelID;
            card.addEventListener('click', () => {
                // Переключаем выбранную модель
                this.panel.querySelectorAll('.model-card').forEach(c => c.classList.remove('selected'));
                card.classList.add('selected');
                
                // Обновляем детали модели
                this.updateModelDetails(model);
                
                // Обновляем текущую модель для роли
                this.selectedRole.CurrentModel = model;
            });
            modelsGrid.appendChild(card);
        });

        // Если есть выбранная модель, показываем ее детали
        if (selectedModel) {
            this.updateModelDetails(selectedModel);
        } else if (models.length > 0) {
            // Если нет выбранной модели, выбираем первую
            this.panel.querySelector('.model-card').classList.add('selected');
            this.updateModelDetails(models[0]);
            this.selectedRole.CurrentModel = models[0];
        }
    }

    /**
     * Обновляет детали выбранной модели
     * @param {Object} model Модель
     */
    updateModelDetails(model) {
        const modelName = this.panel.querySelector('#selected-model-name');
        const modelDescription = this.panel.querySelector('#selected-model-description');
        const modelCapabilities = this.panel.querySelector('#model-capabilities');

        modelName.textContent = model.DisplayName;
        modelDescription.textContent = model.Description || '';

        // Обновляем возможности модели
        modelCapabilities.innerHTML = '';
        if (model.Capabilities && model.Capabilities.length > 0) {
            model.Capabilities.forEach(cap => {
                const capSpan = document.createElement('span');
                capSpan.className = 'model-capability';
                capSpan.textContent = cap;
                modelCapabilities.appendChild(capSpan);
            });
        }

        // Обновляем параметры модели по умолчанию
        // Здесь можно добавить логику для загрузки параметров из конфигурации
        const temperatureSlider = this.panel.querySelector('#temperature');
        const temperatureValue = this.panel.querySelector('#temperature-value');
        const topPSlider = this.panel.querySelector('#top-p');
        const topPValue = this.panel.querySelector('#top-p-value');

        // Устанавливаем значения по умолчанию или из сохраненной конфигурации
        temperatureSlider.value = 0.7;
        temperatureValue.textContent = temperatureSlider.value;
        
        topPSlider.value = 1.0;
        topPValue.textContent = topPSlider.value;
    }

    /**
     * Форматирует количество токенов
     * @param {number} tokens Количество токенов
     * @returns {string} Отформатированное количество токенов
     */
    formatTokens(tokens) {
        if (!tokens) return 'N/A';
        if (tokens >= 1000000) {
            return `${(tokens / 1000000).toFixed(1)}M`;
        } else if (tokens >= 1000) {
            return `${(tokens / 1000).toFixed(1)}K`;
        }
        return tokens.toString();
    }

    /**
     * Возвращает отображаемое имя провайдера
     * @param {string} provider Провайдер
     * @returns {string} Отображаемое имя
     */
    getProviderDisplayName(provider) {
        const displayNames = {
            'openai': 'OpenAI',
            'anthropic': 'Anthropic',
            'deepseek': 'DeepSeek',
            'mistral': 'Mistral AI',
            'grok': 'Grok AI',
            'llama': 'Llama (local)'
        };
        return displayNames[provider] || provider;
    }

    /**
     * Переходит к предыдущей роли
     */
    navigateToPreviousRole() {
        if (this.currentRoleIndex > 0) {
            // Сохраняем выбор для текущей роли
            this.saveCurrentSelection();
            
            // Переходим к предыдущей роли
            this.currentRoleIndex--;
            this.selectedRole = this.roles[this.currentRoleIndex];
            this.updateRoleView();
        }
    }

    /**
     * Переходит к следующей роли
     */
    async navigateToNextRole() {
        // Сохраняем выбор для текущей роли
        await this.saveCurrentSelection();
        
        if (this.currentRoleIndex < this.roles.length - 1) {
            // Переходим к следующей роли
            this.currentRoleIndex++;
            this.selectedRole = this.roles[this.currentRoleIndex];
            this.updateRoleView();
        } else {
            // Завершаем настройку
            this.completeSetup();
        }
    }

    /**
     * Сохраняет текущий выбор для текущей роли
     */
    async saveCurrentSelection() {
        if (!this.selectedRole || !this.selectedRole.CurrentModel) return;

        try {
            // Получаем параметры
            const temperatureSlider = this.panel.querySelector('#temperature');
            const topPSlider = this.panel.querySelector('#top-p');

            const customParams = {
                temperature: parseFloat(temperatureSlider.value),
                top_p: parseFloat(topPSlider.value)
            };

            // Вызываем MCP-команду для сохранения выбора
            await mcp.invoke("select_model", {
                role_id: this.selectedRole.RoleID,
                provider: this.selectedRole.CurrentModel.Provider,
                model_id: this.selectedRole.CurrentModel.ModelID,
                custom_params: customParams
            });

            // Вызываем callback
            this.callbacks.onSave(this.selectedRole.RoleID, this.selectedRole.CurrentModel);

        } catch (error) {
            console.error("Ошибка сохранения выбора модели:", error);
            this.showError("Не удалось сохранить выбор модели: " + error.message);
        }
    }

    /**
     * Завершает настройку моделей
     */
    completeSetup() {
        this.hidePanel();
        this.callbacks.onComplete(this.roles);
    }

    /**
     * Показывает сообщение об ошибке
     * @param {string} message Сообщение об ошибке
     */
    showError(message) {
        // Удаляем старое сообщение об ошибке, если есть
        const oldError = this.panel.querySelector('.error-message');
        if (oldError) {
            oldError.remove();
        }

        // Создаем новое сообщение об ошибке
        const errorDiv = document.createElement('div');
        errorDiv.className = 'error-message';
        errorDiv.textContent = message;

        // Добавляем сообщение в начало контента
        const content = this.panel.querySelector('.panel-content');
        content.insertBefore(errorDiv, content.firstChild);

        // Автоматически скрываем сообщение через 5 секунд
        setTimeout(() => {
            if (errorDiv.parentNode) {
                errorDiv.remove();
            }
        }, 5000);
    }

    /**
     * Загружает рекомендуемые модели для роли
     * @param {string} roleId - ID роли
     */
    async loadRecommendedModels(roleId) {
        try {
            const response = await mcp.invoke("recommend_models", {
                role_id: roleId
            });
            
            if (response && response.status === 'success' && response.data) {
                const data = response.data;
                
                // Если есть рекомендуемая модель, добавляем её в список и выделяем
                if (data.recommendedModel) {
                    const recommendedModel = data.recommendedModel;
                    
                    // Создаем элемент рекомендации
                    const recommendationEl = document.createElement('div');
                    recommendationEl.className = 'model-recommendation';
                    recommendationEl.innerHTML = `
                        <div class="recommendation-header">
                            <span class="recommendation-icon">💡</span>
                            <span class="recommendation-title">Рекомендуемая модель для роли</span>
                        </div>
                        <div class="recommendation-model">
                            <span class="model-name">${recommendedModel.displayName || recommendedModel.modelId}</span>
                            <span class="model-provider">${this.getProviderDisplayName(recommendedModel.provider)}</span>
                        </div>
                        <button class="use-recommended-btn">Использовать рекомендуемую</button>
                    `;
                    
                    // Добавляем обработчик для кнопки
                    recommendationEl.querySelector('.use-recommended-btn').addEventListener('click', () => {
                        this.selectRecommendedModel(recommendedModel);
                    });
                    
                    // Добавляем элемент на страницу
                    const roleDescription = document.getElementById('role-description');
                    roleDescription.parentNode.insertBefore(recommendationEl, roleDescription.nextSibling);
                }
            }
        } catch (error) {
            console.error('Ошибка при загрузке рекомендуемых моделей:', error);
        }
    }
    
    /**
     * Выбирает рекомендуемую модель
     * @param {Object} model - Модель для выбора
     */
    async selectRecommendedModel(model) {
        if (!model || !model.provider || !model.modelId) return;
        
        const role = this.roles[this.currentRoleIndex];
        
        try {
            const response = await mcp.invoke("select_model", {
                role_id: role.id,
                provider: model.provider,
                model_id: model.modelId
            });
            
            if (response && response.status === 'success') {
                // Обновляем текущую модель роли
                role.currentModel = model;
                
                // Активируем вкладку провайдера
                const providerTab = document.querySelector(`.provider-tab[data-provider="${model.provider}"]`);
                if (providerTab) {
                    document.querySelectorAll('.provider-tab').forEach(t => t.classList.remove('active'));
                    providerTab.classList.add('active');
                    this.showModelsForProvider(model.provider);
                    
                    // Выделяем карточку модели
                    setTimeout(() => {
                        const modelCard = document.querySelector(`.model-card[data-model-id="${model.modelId}"][data-provider="${model.provider}"]`);
                        if (modelCard) {
                            document.querySelectorAll('.model-card').forEach(card => card.classList.remove('selected'));
                            modelCard.classList.add('selected');
                            
                            // Обновляем детали модели
                            this.updateModelDetails(model);
                        }
                    }, 100);
                }
                
                // Показываем сообщение об успехе
                this.showSuccessMessage(`Модель ${model.displayName || model.modelId} выбрана для роли ${role.displayName || role.id}`);
            } else {
                this.showError(response.error || 'Не удалось выбрать модель');
            }
        } catch (error) {
            console.error('Ошибка при выборе модели:', error);
            this.showError('Ошибка при выборе модели: ' + error.message);
        }
    }
    
    /**
     * Показывает сообщение об успешном действии
     * @param {string} message - Текст сообщения
     */
    showSuccessMessage(message) {
        const errorElement = document.getElementById('error-message');
        errorElement.textContent = message;
        errorElement.classList.remove('error-message');
        errorElement.classList.add('success-message');
        
        // Скрываем сообщение через 3 секунды
        setTimeout(() => {
            errorElement.textContent = '';
            errorElement.classList.remove('success-message');
            errorElement.classList.add('error-message');
        }, 3000);
    }
    
    /**
     * Экспортирует настройки моделей в Task Master
     */
    async exportToTaskMaster() {
        try {
            const response = await mcp.invoke("taskmaster_export", {});
            
            if (response && response.status === 'success') {
                this.showSuccessMessage(`Настройки экспортированы в ${response.data.exportedPath}`);
            } else {
                this.showError(response.error || 'Не удалось экспортировать настройки');
            }
        } catch (error) {
            console.error('Ошибка при экспорте настроек:', error);
            this.showError('Ошибка при экспорте настроек: ' + error.message);
        }
    }
    
    /**
     * Импортирует настройки моделей из Task Master
     */
    async importFromTaskMaster() {
        try {
            const response = await mcp.invoke("taskmaster_import", {});
            
            if (response && response.status === 'success') {
                this.showSuccessMessage(`Импортировано ${Object.keys(response.data.models || {}).length} моделей`);
                
                // Перезагружаем панель для отображения новых настроек
                await this.initialize();
            } else {
                this.showError(response.error || 'Не удалось импортировать настройки');
            }
        } catch (error) {
            console.error('Ошибка при импорте настроек:', error);
            this.showError('Ошибка при импорте настроек: ' + error.message);
        }
    }
    
    // Добавляем кнопки для интеграции с Task Master в футер панели
    addTaskMasterIntegrationButtons() {
        const footerActions = this.panel.querySelector('.panel-footer');
        
        // Добавляем кнопки, если их еще нет
        if (!document.getElementById('export-taskmaster-btn')) {
            const exportButton = document.createElement('button');
            exportButton.id = 'export-taskmaster-btn';
            exportButton.className = 'panel-button secondary';
            exportButton.textContent = 'Экспорт в Task Master';
            exportButton.addEventListener('click', () => this.exportToTaskMaster());
            footerActions.insertBefore(exportButton, footerActions.firstChild);
        }
        
        if (!document.getElementById('import-taskmaster-btn')) {
            const importButton = document.createElement('button');
            importButton.id = 'import-taskmaster-btn';
            importButton.className = 'panel-button secondary';
            importButton.textContent = 'Импорт из Task Master';
            importButton.addEventListener('click', () => this.importFromTaskMaster());
            footerActions.insertBefore(importButton, footerActions.firstChild);
        }
    }
}

// Создаем глобальный экземпляр панели
if (typeof window !== 'undefined') {
    window.ModelsSetupPanel = new ModelsSetupPanel();
} 