package mcpserver

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// RunStatus возможные статусы выполнения цепочки
const (
	RunStatusRunning   = "running"
	RunStatusCompleted = "completed"
)

// ChainRun описывает запуск цепочки
type ChainRun struct {
	ID          string    `json:"id"`
	ChainID     string    `json:"chain_id"`
	Status      string    `json:"status"`
	Progress    int       `json:"progress"` // 0-100
	Result      string    `json:"result,omitempty"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

// runStore управляет запусками
type runStore struct {
	mu   sync.RWMutex
	runs map[string]ChainRun
}

func newRunStore() *runStore {
	return &runStore{runs: make(map[string]ChainRun)}
}

func (s *runStore) createRun(chainID string) ChainRun {
	id := fmtID("run")
	r := ChainRun{
		ID:        id,
		ChainID:   chainID,
		Status:    RunStatusRunning,
		Progress:  0,
		StartedAt: time.Now(),
	}
	s.mu.Lock()
	s.runs[id] = r
	s.mu.Unlock()
	return r
}

func (s *runStore) updateRun(r ChainRun) {
	s.mu.Lock()
	s.runs[r.ID] = r
	s.mu.Unlock()
}

func (s *runStore) getRun(id string) (ChainRun, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.runs[id]
	return r, ok
}

// simulateRun прогоняет фиктивную работу и обновляет прогресс
func (s *runStore) simulateRun(id string) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		r, ok := s.getRun(id)
		if !ok {
			return // удалён
		}
		if r.Progress >= 100 {
			return
		}
		inc := rand.Intn(20) + 5 // 5-24%
		r.Progress += inc
		if r.Progress >= 100 {
			r.Progress = 100
			r.Status = RunStatusCompleted
			r.CompletedAt = time.Now()
			r.Result = "Simulated result text"
		}
		s.updateRun(r)
		if r.Status == RunStatusCompleted {
			return
		}
		// создать чекпоинт каждые 25 %
		if r.Progress/25 > (r.Progress-inc)/25 {
			cp := Checkpoint{
				ID:        fmtID("cp"),
				RunID:     r.ID,
				Progress:  r.Progress,
				Data:      fmt.Sprintf("Checkpoint at %d%%", r.Progress),
				CreatedAt: time.Now(),
			}
			globalCPStore.add(cp)
		}
	}
}

// fmtID helper
func fmtID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()/1e6)
}
