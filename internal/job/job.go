package job

import (
	"encoding/json"
	"time"
)

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
)

type Job struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Status    Status          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusQueued, StatusFailed, StatusSucceeded, StatusRunning:
		return true
	default:
		return false
	}
}

func (s Status) CanTransitionTo(next Status) bool {
	switch s {
	case StatusQueued:
		if next == StatusRunning {
			return true
		} else {
			return false
		}
	case StatusRunning:
		if next == StatusSucceeded || next == StatusFailed {
			return true
		} else {
			return false
		}
	default:
		return false
	}
}
