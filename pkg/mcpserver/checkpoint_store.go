package mcpserver

import (
	"sync"
	"time"
)

// Checkpoint описывает сохранённое промежуточное состояние.
type Checkpoint struct {
	ID        string    `json:"id"`
	RunID     string    `json:"run_id"`
	Progress  int       `json:"progress"`
	Data      string    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
}

type checkpointStore struct {
	mu          sync.RWMutex
	checkpoints map[string]Checkpoint // id -> cp
	byRun       map[string][]string   // runID -> []cpID
}

func newCheckpointStore() *checkpointStore {
	return &checkpointStore{
		checkpoints: make(map[string]Checkpoint),
		byRun:       make(map[string][]string),
	}
}

func (s *checkpointStore) add(cp Checkpoint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkpoints[cp.ID] = cp
	s.byRun[cp.RunID] = append(s.byRun[cp.RunID], cp.ID)
}

func (s *checkpointStore) list(runID string) []Checkpoint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := s.byRun[runID]
	res := make([]Checkpoint, 0, len(ids))
	for _, id := range ids {
		if cp, ok := s.checkpoints[id]; ok {
			res = append(res, cp)
		}
	}
	return res
}

func (s *checkpointStore) get(id string) (Checkpoint, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp, ok := s.checkpoints[id]
	return cp, ok
}
