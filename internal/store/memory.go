package store

import (
	"errors"
	"relay/internal/job"
	"sync"
)

// errors.New returns pointer to unexported (private) struct for unique memory address
var ErrAlreadyExists = errors.New("job already exists")
var ErrNotFound = errors.New("job not found")

// stores a new job by ID
type MemoryStore struct {
	mu   sync.RWMutex
	jobs map[string]job.Job
}

// constructor method for MemoryStore
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		jobs: make(map[string]job.Job),
	}
}

// store job if it doesn't exist
func (s *MemoryStore) Create(j job.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.jobs[j.ID]
	if ok {
		return ErrAlreadyExists
	}

	s.jobs[j.ID] = job.Job{
		ID:        j.ID,
		Type:      j.Type,
		Payload:   append([]byte(nil), j.Payload...),
		Status:    j.Status,
		CreatedAt: j.CreatedAt,
		UpdatedAt: j.UpdatedAt,
	}
	return nil
}

// return matching job when present
func (s *MemoryStore) Get(id string) (job.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.jobs[id]
	if !ok {
		return val, ErrNotFound
	} else {
		val.Payload = append([]byte(nil), val.Payload...)
		return val, nil
	}
}
