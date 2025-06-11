/**
 * Модуль для визуализации чекпоинтов в MCP-интерфейсе
 */

/**
 * Создает панель просмотра чекпоинтов
 * @param {Object} options - Опции панели
 * @param {string} options.chainId - ID цепочки
 * @param {string} options.runId - ID выполнения (опционально)
 * @returns {Promise<string>} HTML-разметка панели
 */
async function createCheckpointViewerPanel(options) {
  const { chainId, runId } = options;
  
  try {
    // Получаем список чекпоинтов
    const response = await mcp.sendCommand('checkpoint_list', { 
      chain_id: chainId,
      run_id: runId
    });
    
    if (!response || !response.checkpoints || response.checkpoints.length === 0) {
      return `
        <div class="cp-panel">
          <div class="cp-header">
            <h2>Чекпоинты для цепочки: ${chainId}</h2>
          </div>
          <div class="cp-empty-state">
            <p>Чекпоинты не найдены для этой цепочки.</p>
          </div>
        </div>
      `;
    }
    
    // Создаем временную шкалу
    const timeline = createTimelineHTML(response.timeline);
    
    // Создаем список чекпоинтов
    const checkpointsList = createCheckpointsListHTML(response.checkpoints);
    
    // Возвращаем HTML-разметку
    return `
      <div class="cp-panel">
        <div class="cp-header">
          <h2>Чекпоинты для цепочки: ${chainId}</h2>
          ${runId ? `<div class="cp-subheader">Выполнение: ${runId}</div>` : ''}
        </div>
        
        <div class="cp-timeline-container">
          <h3>Временная шкала выполнения</h3>
          ${timeline}
        </div>
        
        <div class="cp-list-container">
          <h3>Список чекпоинтов</h3>
          ${checkpointsList}
        </div>
      </div>
    `;
  } catch (error) {
    console.error('Ошибка при получении чекпоинтов:', error);
    return `
      <div class="cp-panel cp-error">
        <h2>Ошибка при загрузке чекпоинтов</h2>
        <p>${error.message || 'Не удалось загрузить чекпоинты'}</p>
      </div>
    `;
  }
}

/**
 * Создает HTML-разметку временной шкалы
 * @param {Array} timeline - События временной шкалы
 * @returns {string} HTML-разметка
 */
function createTimelineHTML(timeline) {
  if (!timeline || timeline.length === 0) {
    return '<div class="cp-timeline-empty">Нет данных для отображения временной шкалы</div>';
  }
  
  // Сортируем события по времени
  const sortedEvents = [...timeline].sort((a, b) => 
    new Date(a.timestamp) - new Date(b.timestamp)
  );
  
  // Создаем элементы шкалы
  const timelineItems = sortedEvents.map((event, index) => {
    const date = new Date(event.timestamp);
    const formattedTime = date.toLocaleTimeString();
    const typeClass = `cp-event-${event.type.toLowerCase()}`;
    const progressPercentage = ((index + 1) / sortedEvents.length) * 100;
    
    let icon = '📄';
    switch (event.type.toLowerCase()) {
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
      <div class="cp-timeline-item ${typeClass}" style="left: ${progressPercentage}%;" 
           data-checkpoint-id="${event.id}" 
           onclick="showCheckpointDetails('${event.id}')">
        <div class="cp-timeline-icon">${icon}</div>
        <div class="cp-timeline-tooltip">
          <div>${event.type}</div>
          <div>${event.model_name || ''}</div>
          <div>${formattedTime}</div>
        </div>
      </div>
    `;
  });
  
  return `
    <div class="cp-timeline">
      <div class="cp-timeline-track"></div>
      ${timelineItems.join('')}
    </div>
  `;
}

/**
 * Создает HTML-разметку списка чекпоинтов
 * @param {Array} checkpoints - Список чекпоинтов
 * @returns {string} HTML-разметка
 */
function createCheckpointsListHTML(checkpoints) {
  if (!checkpoints || checkpoints.length === 0) {
    return '<div class="cp-list-empty">Нет доступных чекпоинтов</div>';
  }
  
  // Сортируем чекпоинты по времени создания (от новых к старым)
  const sortedCheckpoints = [...checkpoints].sort((a, b) => 
    new Date(b.created_at) - new Date(a.created_at)
  );
  
  // Создаем элементы списка
  const checkpointItems = sortedCheckpoints.map(cp => {
    const date = new Date(cp.created_at);
    const formattedDate = date.toLocaleString();
    const typeClass = `cp-type-${cp.type.toLowerCase()}`;
    const sizeFormatted = formatContentSize(cp.content_size);
    
    return `
      <div class="cp-list-item ${typeClass}" data-checkpoint-id="${cp.id}">
        <div class="cp-item-header" onclick="toggleCheckpointDetails('${cp.id}')">
          <div class="cp-item-type">${cp.type}</div>
          <div class="cp-item-model">${cp.model_id || 'Нет модели'}</div>
          <div class="cp-item-time">${formattedDate}</div>
          <div class="cp-item-size">${sizeFormatted}</div>
        </div>
        <div class="cp-item-actions">
          <button onclick="showCheckpointDetails('${cp.id}')" class="cp-btn cp-btn-view">Просмотр</button>
          <button onclick="deleteCheckpoint('${cp.id}')" class="cp-btn cp-btn-delete">Удалить</button>
        </div>
        <div id="cp-details-${cp.id}" class="cp-item-details" style="display: none;">
          <div class="cp-details-loading">Загрузка содержимого...</div>
        </div>
      </div>
    `;
  });
  
  return `
    <div class="cp-list">
      ${checkpointItems.join('')}
    </div>
  `;
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
 * Показывает детали чекпоинта
 * @param {string} checkpointId - ID чекпоинта
 */
async function showCheckpointDetails(checkpointId) {
  try {
    // Загружаем детали чекпоинта
    const checkpoint = await mcp.sendCommand('checkpoint_get', { 
      checkpoint_id: checkpointId 
    });
    
    // Отображаем детали в модальном окне
    const modal = document.createElement('div');
    modal.className = 'cp-modal';
    modal.innerHTML = `
      <div class="cp-modal-content">
        <div class="cp-modal-header">
          <h3>Чекпоинт: ${checkpoint.type}</h3>
          <button class="cp-modal-close" onclick="this.parentNode.parentNode.parentNode.remove()">×</button>
        </div>
        <div class="cp-modal-body">
          <div class="cp-modal-info">
            <div><strong>ID:</strong> ${checkpoint.id}</div>
            <div><strong>Тип:</strong> ${checkpoint.type}</div>
            <div><strong>Модель:</strong> ${checkpoint.model_id || 'Нет'}</div>
            <div><strong>Создан:</strong> ${new Date(checkpoint.created_at).toLocaleString()}</div>
            <div><strong>Размер:</strong> ${formatContentSize(checkpoint.content_size)}</div>
          </div>
          <div class="cp-modal-content-container">
            <h4>Содержимое:</h4>
            <pre class="cp-modal-content-pre">${escapeHtml(checkpoint.content)}</pre>
          </div>
          ${checkpoint.metadata ? `
            <div class="cp-modal-metadata">
              <h4>Метаданные:</h4>
              <pre>${escapeHtml(JSON.stringify(checkpoint.metadata, null, 2))}</pre>
            </div>
          ` : ''}
        </div>
        <div class="cp-modal-footer">
          <button class="cp-btn cp-btn-delete" onclick="deleteCheckpoint('${checkpoint.id}', true)">Удалить</button>
          <button class="cp-btn cp-btn-close" onclick="this.parentNode.parentNode.parentNode.remove()">Закрыть</button>
        </div>
      </div>
    `;
    
    document.body.appendChild(modal);
  } catch (error) {
    console.error('Ошибка при загрузке деталей чекпоинта:', error);
    alert(`Ошибка: ${error.message || 'Не удалось загрузить детали чекпоинта'}`);
  }
}

/**
 * Удаляет чекпоинт
 * @param {string} checkpointId - ID чекпоинта
 * @param {boolean} isModal - Флаг, указывающий, что вызов из модального окна
 */
async function deleteCheckpoint(checkpointId, isModal = false) {
  if (!confirm('Вы уверены, что хотите удалить этот чекпоинт?')) {
    return;
  }
  
  try {
    // Отправляем запрос на удаление
    await mcp.sendCommand('checkpoint_delete', { 
      checkpoint_id: checkpointId 
    });
    
    // Удаляем элемент из DOM
    const element = document.querySelector(`[data-checkpoint-id="${checkpointId}"]`);
    if (element) {
      element.remove();
    }
    
    // Если вызов из модального окна, закрываем его
    if (isModal) {
      const modal = document.querySelector('.cp-modal');
      if (modal) {
        modal.remove();
      }
    }
    
    // Обновляем список, если элемент был в нем
    const listContainer = document.querySelector('.cp-list-container');
    if (listContainer) {
      // Проверяем, есть ли еще элементы в списке
      const remainingItems = listContainer.querySelectorAll('.cp-list-item');
      if (remainingItems.length === 0) {
        listContainer.querySelector('.cp-list').innerHTML = 
          '<div class="cp-list-empty">Нет доступных чекпоинтов</div>';
      }
    }
    
    // Уведомляем пользователя
    alert('Чекпоинт успешно удален');
  } catch (error) {
    console.error('Ошибка при удалении чекпоинта:', error);
    alert(`Ошибка: ${error.message || 'Не удалось удалить чекпоинт'}`);
  }
}

/**
 * Переключает отображение деталей чекпоинта в списке
 * @param {string} checkpointId - ID чекпоинта
 */
async function toggleCheckpointDetails(checkpointId) {
  const detailsElement = document.getElementById(`cp-details-${checkpointId}`);
  
  if (!detailsElement) return;
  
  const isVisible = detailsElement.style.display !== 'none';
  
  if (isVisible) {
    detailsElement.style.display = 'none';
    return;
  }
  
  detailsElement.style.display = 'block';
  
  // Если содержимое еще не загружено
  if (detailsElement.querySelector('.cp-details-loading')) {
    try {
      // Загружаем детали чекпоинта
      const checkpoint = await mcp.sendCommand('checkpoint_get', { 
        checkpoint_id: checkpointId 
      });
      
      // Ограничиваем размер предпросмотра
      const previewContent = checkpoint.content.length > 500 
        ? checkpoint.content.substring(0, 500) + '...' 
        : checkpoint.content;
      
      // Обновляем содержимое
      detailsElement.innerHTML = `
        <div class="cp-preview">
          <pre>${escapeHtml(previewContent)}</pre>
          <button class="cp-btn cp-btn-view" onclick="showCheckpointDetails('${checkpoint.id}')">
            Полный просмотр
          </button>
        </div>
      `;
    } catch (error) {
      console.error('Ошибка при загрузке превью чекпоинта:', error);
      detailsElement.innerHTML = `
        <div class="cp-error">
          Ошибка загрузки: ${error.message || 'Не удалось загрузить данные чекпоинта'}
        </div>
      `;
    }
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
  createCheckpointViewerPanel,
  showCheckpointDetails,
  deleteCheckpoint,
  toggleCheckpointDetails
}; 