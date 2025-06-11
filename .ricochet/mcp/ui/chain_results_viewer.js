/**
 * Модуль для визуализации результатов цепочек в MCP-интерфейсе
 */

/**
 * Создает панель просмотра результатов цепочки
 * @param {Object} options - Опции панели
 * @param {string} options.chainId - ID цепочки
 * @param {string} options.runId - ID выполнения (опционально)
 * @param {boolean} options.includeStats - Включать ли статистику
 * @returns {Promise<string>} HTML-разметка панели
 */
async function createChainResultsPanel(options) {
  const { chainId, runId, includeStats = true } = options;
  
  try {
    // Получаем результаты цепочки
    const response = await mcp.sendCommand('chain_results', { 
      chain_id: chainId,
      run_id: runId,
      include_stats: includeStats
    });
    
    if (!response || !response.results || response.results.length === 0) {
      return `
        <div class="cr-panel">
          <div class="cr-header">
            <h2>Результаты для цепочки: ${chainId}</h2>
          </div>
          <div class="cr-empty-state">
            <p>Результаты не найдены для этой цепочки.</p>
          </div>
        </div>
      `;
    }
    
    // Создаем таблицу результатов
    const resultsTable = createResultsTableHTML(response.results);
    
    // Создаем статистику, если требуется
    const statsSection = includeStats && response.stats 
      ? createStatsHTML(response.stats) 
      : '';
    
    // Создаем секцию чекпоинтов, если указан runId
    const checkpointsSection = runId && response.checkpoints 
      ? createCheckpointsHTML(response.checkpoints)
      : '';
    
    // Возвращаем HTML-разметку
    return `
      <div class="cr-panel">
        <div class="cr-header">
          <h2>Результаты для цепочки: ${response.chain_name || chainId}</h2>
          ${runId ? `<div class="cr-subheader">Выполнение: ${runId}</div>` : ''}
        </div>
        
        ${statsSection}
        
        <div class="cr-results-container">
          <h3>История выполнений</h3>
          ${resultsTable}
        </div>
        
        ${checkpointsSection}
      </div>
    `;
  } catch (error) {
    console.error('Ошибка при получении результатов:', error);
    return `
      <div class="cr-panel cr-error">
        <h2>Ошибка при загрузке результатов</h2>
        <p>${error.message || 'Не удалось загрузить результаты цепочки'}</p>
      </div>
    `;
  }
}

/**
 * Создает HTML-разметку таблицы результатов
 * @param {Array} results - Результаты выполнения
 * @returns {string} HTML-разметка
 */
function createResultsTableHTML(results) {
  if (!results || results.length === 0) {
    return '<div class="cr-results-empty">Нет данных о выполнении цепочки</div>';
  }
  
  // Создаем строки таблицы
  const tableRows = results.map(result => {
    const startDate = new Date(result.started_at);
    const formattedStartDate = startDate.toLocaleString();
    
    const duration = result.duration_ms 
      ? formatDuration(result.duration_ms) 
      : 'Н/Д';
    
    const statusClass = getStatusClass(result.status);
    const statusText = getStatusText(result.status);
    
    return `
      <tr data-run-id="${result.run_id}" onclick="showRunDetails('${result.run_id}')">
        <td><div class="cr-run-id">${result.run_id}</div></td>
        <td><div class="cr-status ${statusClass}">${statusText}</div></td>
        <td>${formattedStartDate}</td>
        <td>${duration}</td>
        <td>${result.result_summary}</td>
        <td>
          <button class="cr-btn cr-btn-view" onclick="event.stopPropagation(); showRunDetails('${result.run_id}')">
            Детали
          </button>
        </td>
      </tr>
    `;
  });
  
  return `
    <div class="cr-table-container">
      <table class="cr-results-table">
        <thead>
          <tr>
            <th>ID запуска</th>
            <th>Статус</th>
            <th>Дата запуска</th>
            <th>Длительность</th>
            <th>Результат</th>
            <th>Действия</th>
          </tr>
        </thead>
        <tbody>
          ${tableRows.join('')}
        </tbody>
      </table>
    </div>
  `;
}

/**
 * Создает HTML-разметку секции статистики
 * @param {Object} stats - Статистика выполнения
 * @returns {string} HTML-разметка
 */
function createStatsHTML(stats) {
  if (!stats) {
    return '';
  }
  
  const lastRunDate = new Date(stats.last_run_date);
  const formattedLastRunDate = lastRunDate.toLocaleString();
  
  const lastSuccessDate = new Date(stats.last_successful_date);
  const formattedLastSuccessDate = lastSuccessDate.toLocaleString();
  
  return `
    <div class="cr-stats-container">
      <h3>Статистика выполнений</h3>
      <div class="cr-stats-grid">
        <div class="cr-stat-item">
          <div class="cr-stat-value">${stats.total_runs}</div>
          <div class="cr-stat-label">Всего запусков</div>
        </div>
        <div class="cr-stat-item">
          <div class="cr-stat-value">${stats.successful_runs}</div>
          <div class="cr-stat-label">Успешных</div>
        </div>
        <div class="cr-stat-item">
          <div class="cr-stat-value">${stats.failed_runs}</div>
          <div class="cr-stat-label">Ошибок</div>
        </div>
        <div class="cr-stat-item">
          <div class="cr-stat-value">${stats.success_rate.toFixed(1)}%</div>
          <div class="cr-stat-label">Успешность</div>
        </div>
        <div class="cr-stat-item">
          <div class="cr-stat-value">${formatDuration(stats.average_duration_ms)}</div>
          <div class="cr-stat-label">Средняя длительность</div>
        </div>
        <div class="cr-stat-item">
          <div class="cr-stat-value">${stats.average_tokens_used}</div>
          <div class="cr-stat-label">Средн. токенов</div>
        </div>
        <div class="cr-stat-item">
          <div class="cr-stat-value">${stats.total_tokens_used}</div>
          <div class="cr-stat-label">Всего токенов</div>
        </div>
        <div class="cr-stat-item">
          <div class="cr-stat-value">$${stats.estimated_cost.toFixed(2)}</div>
          <div class="cr-stat-label">Стоимость</div>
        </div>
      </div>
      <div class="cr-stats-dates">
        <div><strong>Последний запуск:</strong> ${formattedLastRunDate}</div>
        <div><strong>Последний успешный:</strong> ${formattedLastSuccessDate}</div>
      </div>
    </div>
  `;
}

/**
 * Создает HTML-разметку секции чекпоинтов
 * @param {Array} checkpoints - Список чекпоинтов
 * @returns {string} HTML-разметка
 */
function createCheckpointsHTML(checkpoints) {
  if (!checkpoints || checkpoints.length === 0) {
    return '';
  }
  
  const checkpointTimeline = createCheckpointTimelineHTML(checkpoints);
  
  return `
    <div class="cr-checkpoints-container">
      <h3>Чекпоинты выполнения</h3>
      ${checkpointTimeline}
      <div class="cr-view-all-checkpoints">
        <button class="cr-btn cr-btn-view-all" onclick="showAllCheckpoints('${checkpoints[0].chain_id}')">
          Просмотреть все чекпоинты
        </button>
      </div>
    </div>
  `;
}

/**
 * Создает HTML-разметку временной шкалы чекпоинтов
 * @param {Array} checkpoints - Список чекпоинтов
 * @returns {string} HTML-разметка
 */
function createCheckpointTimelineHTML(checkpoints) {
  if (!checkpoints || checkpoints.length === 0) {
    return '<div class="cr-checkpoints-empty">Нет данных о чекпоинтах</div>';
  }
  
  // Сортируем по времени создания
  const sortedCheckpoints = [...checkpoints].sort((a, b) => 
    new Date(a.created_at) - new Date(b.created_at)
  );
  
  // Создаем элементы на временной шкале
  const timelineItems = sortedCheckpoints.map((cp, index) => {
    const typeClass = `cr-cp-type-${cp.type.toLowerCase()}`;
    const progressPercentage = ((index + 1) / sortedCheckpoints.length) * 100;
    
    let icon = '📄';
    switch (cp.type.toLowerCase()) {
      case 'input':
        icon = '📥';
        break;
      case 'output':
        icon = '📤';
        break;
      case 'intermediate':
        icon = '🔄';
        break;
      case 'complete':
        icon = '✅';
        break;
      case 'error':
        icon = '❌';
        break;
      case 'segment':
        icon = '📑';
        break;
    }
    
    return `
      <div class="cr-timeline-item ${typeClass}" style="left: ${progressPercentage}%;" 
           data-checkpoint-id="${cp.id}" 
           onclick="showCheckpointDetails('${cp.id}')">
        <div class="cr-timeline-icon">${icon}</div>
        <div class="cr-timeline-tooltip">
          <div>${cp.type}</div>
          <div>${formatContentSize(cp.content_size)}</div>
        </div>
      </div>
    `;
  });
  
  return `
    <div class="cr-timeline">
      <div class="cr-timeline-track"></div>
      ${timelineItems.join('')}
    </div>
  `;
}

/**
 * Показывает детали запуска
 * @param {string} runId - ID запуска
 */
async function showRunDetails(runId) {
  try {
    // Получаем детальную информацию о запуске
    const response = await mcp.sendCommand('chain_run_result', { 
      run_id: runId 
    });
    
    // Создаем модальное окно с деталями
    const modal = document.createElement('div');
    modal.className = 'cr-modal';
    
    // Определяем содержимое результата
    let resultContent = '';
    if (response.result && response.result.text) {
      resultContent = `<pre class="cr-result-content">${escapeHtml(response.result.text)}</pre>`;
    } else if (response.result) {
      resultContent = `<pre class="cr-result-content">${escapeHtml(JSON.stringify(response.result, null, 2))}</pre>`;
    } else {
      resultContent = '<div class="cr-result-empty">Нет данных о результате</div>';
    }
    
    // Создаем список чекпоинтов
    let checkpointsContent = '';
    if (response.checkpoints && response.checkpoints.length > 0) {
      const checkpointsList = response.checkpoints.map(cp => {
        const typeClass = `cr-cp-type-${cp.type.toLowerCase()}`;
        return `
          <div class="cr-cp-item ${typeClass}">
            <div class="cr-cp-item-header">
              <div class="cr-cp-item-type">${cp.type}</div>
              <div class="cr-cp-item-model">${cp.model_id || 'Нет модели'}</div>
              <div class="cr-cp-item-size">${formatContentSize(cp.content_size)}</div>
            </div>
            <div class="cr-cp-item-actions">
              <button class="cr-btn cr-btn-view" onclick="showCheckpointDetails('${cp.id}')">Просмотр</button>
            </div>
          </div>
        `;
      }).join('');
      
      checkpointsContent = `
        <div class="cr-modal-checkpoints">
          <h4>Чекпоинты:</h4>
          <div class="cr-cp-list">
            ${checkpointsList}
          </div>
        </div>
      `;
    }
    
    modal.innerHTML = `
      <div class="cr-modal-content">
        <div class="cr-modal-header">
          <h3>Результат выполнения: ${runId}</h3>
          <button class="cr-modal-close" onclick="this.parentNode.parentNode.parentNode.remove()">×</button>
        </div>
        <div class="cr-modal-body">
          <div class="cr-modal-result">
            <h4>Результат:</h4>
            ${resultContent}
          </div>
          ${checkpointsContent}
        </div>
        <div class="cr-modal-footer">
          <button class="cr-btn cr-btn-close" onclick="this.parentNode.parentNode.parentNode.remove()">Закрыть</button>
        </div>
      </div>
    `;
    
    document.body.appendChild(modal);
  } catch (error) {
    console.error('Ошибка при загрузке деталей запуска:', error);
    alert(`Ошибка: ${error.message || 'Не удалось загрузить детали запуска'}`);
  }
}

/**
 * Показывает все чекпоинты цепочки
 * @param {string} chainId - ID цепочки
 */
function showAllCheckpoints(chainId) {
  // Открываем панель просмотра чекпоинтов
  window.checkpointViewer.createCheckpointViewerPanel({ chainId })
    .then(html => {
      const modal = document.createElement('div');
      modal.className = 'cr-modal';
      modal.innerHTML = `
        <div class="cr-modal-content cr-modal-content-wide">
          <div class="cr-modal-header">
            <h3>Все чекпоинты цепочки: ${chainId}</h3>
            <button class="cr-modal-close" onclick="this.parentNode.parentNode.parentNode.remove()">×</button>
          </div>
          <div class="cr-modal-body">
            ${html}
          </div>
          <div class="cr-modal-footer">
            <button class="cr-btn cr-btn-close" onclick="this.parentNode.parentNode.parentNode.remove()">Закрыть</button>
          </div>
        </div>
      `;
      
      document.body.appendChild(modal);
    })
    .catch(error => {
      console.error('Ошибка при загрузке чекпоинтов:', error);
      alert(`Ошибка: ${error.message || 'Не удалось загрузить чекпоинты'}`);
    });
}

/**
 * Форматирует длительность в миллисекундах в читаемый формат
 * @param {number} ms - Длительность в миллисекундах
 * @returns {string} Отформатированная длительность
 */
function formatDuration(ms) {
  if (ms < 1000) {
    return `${ms} мс`;
  } else if (ms < 60000) {
    return `${(ms / 1000).toFixed(1)} сек`;
  } else if (ms < 3600000) {
    const minutes = Math.floor(ms / 60000);
    const seconds = Math.floor((ms % 60000) / 1000);
    return `${minutes}м ${seconds}с`;
  } else {
    const hours = Math.floor(ms / 3600000);
    const minutes = Math.floor((ms % 3600000) / 60000);
    return `${hours}ч ${minutes}м`;
  }
}

/**
 * Форматирует размер содержимого
 * @param {number} size - Размер в байтах
 * @returns {string} Отформатированный размер
 */
function formatContentSize(size) {
  if (size < 1024) {
    return `${size} Б`;
  } else if (size < 1024 * 1024) {
    return `${(size / 1024).toFixed(1)} КБ`;
  } else {
    return `${(size / (1024 * 1024)).toFixed(1)} МБ`;
  }
}

/**
 * Возвращает CSS-класс для статуса
 * @param {string} status - Статус выполнения
 * @returns {string} CSS-класс
 */
function getStatusClass(status) {
  switch (status.toLowerCase()) {
    case 'completed':
    case 'done':
    case 'success':
      return 'cr-status-success';
    case 'failed':
    case 'error':
      return 'cr-status-error';
    case 'running':
    case 'in-progress':
      return 'cr-status-running';
    case 'pending':
    case 'waiting':
      return 'cr-status-pending';
    case 'cancelled':
      return 'cr-status-cancelled';
    default:
      return 'cr-status-unknown';
  }
}

/**
 * Возвращает текст для статуса
 * @param {string} status - Статус выполнения
 * @returns {string} Текст статуса
 */
function getStatusText(status) {
  switch (status.toLowerCase()) {
    case 'completed':
    case 'done':
    case 'success':
      return 'Успешно';
    case 'failed':
    case 'error':
      return 'Ошибка';
    case 'running':
    case 'in-progress':
      return 'Выполняется';
    case 'pending':
    case 'waiting':
      return 'Ожидание';
    case 'cancelled':
      return 'Отменено';
    default:
      return 'Неизвестно';
  }
}

/**
 * Экранирует HTML-специальные символы
 * @param {string} text - Исходный текст
 * @returns {string} Экранированный текст
 */
function escapeHtml(text) {
  if (!text) return '';
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

/**
 * Экспортируем функции для использования в MCP
 */
module.exports = {
  createChainResultsPanel,
  showRunDetails,
  showAllCheckpoints
}; 