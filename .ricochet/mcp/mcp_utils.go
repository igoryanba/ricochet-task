package mcp

import (
	"fmt"
	"sync"
	"time"

	"github.com/grik-ai/ricochet-task/pkg/chain"
	"github.com/grik-ai/ricochet-task/pkg/orchestrator"
)

// GlobalServices глобальные сервисы для MCP обработчиков
type GlobalServices struct {
	orchestrator orchestrator.Orchestrator
	chainStore   chain.Store
	mutex        sync.RWMutex
}

var globalServices = &GlobalServices{}

// SetOrchestratorService устанавливает глобальный сервис оркестратора
func SetOrchestratorService(orch orchestrator.Orchestrator) {
	globalServices.mutex.Lock()
	defer globalServices.mutex.Unlock()
	globalServices.orchestrator = orch
}

// SetChainStore устанавливает глобальное хранилище цепочек
func SetChainStore(store chain.Store) {
	globalServices.mutex.Lock()
	defer globalServices.mutex.Unlock()
	globalServices.chainStore = store
}

// GetOrchestratorService возвращает глобальный сервис оркестратора
func GetOrchestratorService() (orchestrator.Orchestrator, error) {
	globalServices.mutex.RLock()
	defer globalServices.mutex.RUnlock()
	
	if globalServices.orchestrator == nil {
		return nil, fmt.Errorf("orchestrator service not initialized")
	}
	
	return globalServices.orchestrator, nil
}

// GetChainStore возвращает глобальное хранилище цепочек
func GetChainStore() (chain.Store, error) {
	globalServices.mutex.RLock()
	defer globalServices.mutex.RUnlock()
	
	if globalServices.chainStore == nil {
		return nil, fmt.Errorf("chain store not initialized")
	}
	
	return globalServices.chainStore, nil
}

// ChainInfo содержит информацию о цепочке
type ChainInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ModelsCount int    `json:"models_count"`
	CreatedAt   string `json:"created_at"`
}

// getChainInfo возвращает информацию о цепочке
func getChainInfo(chainID string) (*ChainInfo, error) {
	chainStore, err := GetChainStore()
	if err != nil {
		return nil, err
	}

	chainObj, err := chainStore.Get(chainID)
	if err != nil {
		return nil, fmt.Errorf("цепочка не найдена: %v", err)
	}

	return &ChainInfo{
		ID:          chainObj.ID,
		Name:        chainObj.Name,
		Description: chainObj.Description,
		ModelsCount: len(chainObj.Models),
		CreatedAt:   chainObj.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// CheckpointSummaryInfo содержит краткую информацию о чекпоинте
type CheckpointSummaryInfo struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	ModelID     string                 `json:"model_id"`
	CreatedAt   time.Time              `json:"created_at"`
	ContentSize int                    `json:"content_size"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}