package mcpserver

import (
	"sync"
	"time"
)

// Chain описывает простейшую цепочку моделей.
type Chain struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Steps       interface{} `json:"steps,omitempty"` // пока произвольный тип
	CreatedAt   time.Time   `json:"created_at"`
}

// ChainStore определяет базовые CRUD операции над цепочками.
type ChainStore interface {
	CreateChain(c Chain) (string, error)
	GetChain(id string) (Chain, bool)
	ListChains() []Chain
	DeleteChain(id string) bool
}

// inMemoryChainStore — простая потокобезопасная реализация ChainStore.
type inMemoryChainStore struct {
	mu     sync.RWMutex
	chains map[string]Chain
}

func newInMemoryStore() *inMemoryChainStore {
	return &inMemoryChainStore{
		chains: make(map[string]Chain),
	}
}

func (s *inMemoryChainStore) CreateChain(c Chain) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chains[c.ID] = c
	return c.ID, nil
}

func (s *inMemoryChainStore) GetChain(id string) (Chain, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.chains[id]
	return c, ok
}

func (s *inMemoryChainStore) ListChains() []Chain {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]Chain, 0, len(s.chains))
	for _, c := range s.chains {
		res = append(res, c)
	}
	return res
}

func (s *inMemoryChainStore) DeleteChain(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.chains[id]; ok {
		delete(s.chains, id)
		return true
	}
	return false
}
