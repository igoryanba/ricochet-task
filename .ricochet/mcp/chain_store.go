package mcp

import (
	"fmt"
	"sync"
)

// StoredStep представляет шаг цепочки, используемый в ин-мемори-хранилище.
// Структура включает только те поля, которые потребуются существующему коду
// (auto_select_models, updateStepModel и т. д.). При необходимости можно
// расширить, это не нарушит обратную совместимость.
//
// NB:  Поля именованы так же, как их ожидают другие части пакета.
//
//	Это позволяет обойтись без кастомных конвертаций/reflection.
type StoredStep struct {
	ID            string `json:"id"`
	Name          string `json:"name,omitempty"`
	Type          string `json:"type,omitempty"`
	RoleID        string `json:"role_id,omitempty"`
	ModelProvider string `json:"model_provider,omitempty"`
	ModelID       string `json:"model_id,omitempty"`
}

// Chain описывает сохранённую цепочку.
type Chain struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Steps       []StoredStep `json:"steps"`
}

// --- глобальное хранилище ---------------------------------------------------

var (
	chainsMu sync.RWMutex
	chains   = make(map[string]Chain)
)

// getChain возвращает цепочку по ID либо ошибку, если она не найдена.
func getChain(id string) (Chain, error) {
	chainsMu.RLock()
	defer chainsMu.RUnlock()
	if c, ok := chains[id]; ok {
		return c, nil
	}
	return Chain{}, fmt.Errorf("chain not found: %s", id)
}

// saveChain сохраняет (создаёт или обновляет) цепочку в хранилище.
func saveChain(c Chain) error {
	if c.ID == "" {
		return fmt.Errorf("chain ID is empty")
	}
	chainsMu.Lock()
	defer chainsMu.Unlock()
	chains[c.ID] = c
	return nil
}

// listChains возвращает копию списка всех цепочек. Может пригодиться позже.
func listChains() []Chain {
	chainsMu.RLock()
	defer chainsMu.RUnlock()
	out := make([]Chain, 0, len(chains))
	for _, c := range chains {
		out = append(out, c)
	}
	return out
}
