package job

import "testing"

func TestStatusValid(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"queued is valid", StatusQueued, true},
		{"running is valid", StatusRunning, true},
		{"succeeded is valid", StatusSucceeded, true},
		{"failed is valid", StatusFailed, true},
		{"unknown is invalid", Status("unknown"), false},
		{"empty is invalid", Status(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.Valid()

			if got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusCanTransitionTo(t *testing.T) {
	tests := []struct {
		name string
		from Status
		to   Status
		want bool
	}{
		{"Queued to Running", StatusQueued, StatusRunning, true},
		{"Running to Succeed", StatusRunning, StatusSucceeded, true},
		{"Running to Failed", StatusRunning, StatusFailed, true},
		{"self: queue -> queue", StatusQueued, StatusQueued, false},
		{"skipped", StatusQueued, StatusSucceeded, false},
		{"self: running -> running", StatusRunning, StatusRunning, false},
		{"terminal: succeeded -> queued", StatusSucceeded, StatusQueued, false},
		{"terminal: failed -> running", StatusFailed, StatusRunning, false},
		{"invalid source", Status("invalid"), StatusRunning, false},
		{"invalid destination", StatusQueued, Status("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.from.CanTransitionTo(tt.to)

			if got != tt.want {
				t.Errorf("CanTransitionTo() = %v, want %v", got, tt.want)
			}
		})
	}
}
