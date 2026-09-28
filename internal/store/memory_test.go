package store

import (
	"encoding/json"
	"errors"
	"reflect"
	"relay/internal/job"
	"strconv"
	"sync"
	"testing"
)

func TestMemoryStoreCreateAndGet(t *testing.T) {
	store := NewMemoryStore()

	want := job.Job{
		ID:      "job-1",
		Type:    "send_email",
		Payload: json.RawMessage(`{"recipient":"user@example.com"}`),
		Status:  job.StatusQueued,
	}

	if err := store.Create(want); err != nil {
		t.Fatalf("Create() returned unexpected error: %v", err)
	}

	got, err := store.Get(want.ID)
	if err != nil {
		t.Fatalf("Get() returned unexpected error: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Get() = %#v, want %#v", got, want)
	}
}

func TestMemoryStoreCreateDuplicate(t *testing.T) {
	store := NewMemoryStore()

	original := job.Job{
		ID:     "job-1",
		Type:   "original",
		Status: job.StatusQueued,
	}

	duplicate := job.Job{
		ID:     "job-1",
		Type:   "replacement",
		Status: job.StatusQueued,
	}

	if err := store.Create(original); err != nil {
		t.Fatalf("first Create() returned unexpected error: %v", err)
	}

	err := store.Create(duplicate)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("second Create() error = %v, want %v", err, ErrAlreadyExists)
	}

	got, err := store.Get(original.ID)
	if err != nil {
		t.Fatalf("Get() returned unexpected error: %v", err)
	}

	if got.Type != original.Type {
		t.Errorf("stored job type = %q, want original type %q", got.Type, original.Type)
	}
}

func TestMemoryStoreGetMissing(t *testing.T) {
	store := NewMemoryStore()

	newJob := job.Job{
		ID:     "job-1",
		Type:   "send-email",
		Status: job.StatusQueued,
	}

	got, err := store.Get(newJob.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() returned error: %v, want: %v", err, ErrNotFound)
	}

	if !reflect.DeepEqual(got, job.Job{}) {
		t.Errorf("Get() return %#v, want zero-value job", got)
	}
}

func TestMemoryStoreCopiesPayloadOnCreate(t *testing.T) {
	store := NewMemoryStore()

	payload := json.RawMessage(`{"message":"original"}`)

	newJob := job.Job{
		ID:      "job-1",
		Type:    "send-email",
		Payload: payload,
		Status:  job.StatusQueued,
	}

	if err := store.Create(newJob); err != nil {
		t.Fatalf("Create() returned unexpected error: %v", err)
	}

	// Mutate original payload slice
	payload[2] = 'X'

	got, err := store.Get(newJob.ID)
	if err != nil {
		t.Fatalf("Get() returned unexpected error: %v", err)
	}

	wantPayload := json.RawMessage(`{"message":"original"}`)
	if !reflect.DeepEqual(got.Payload, wantPayload) {
		t.Errorf("stored payload = %s, want %s", got.Payload, wantPayload)
	}
}

func TestMemoryStoreCopiesPayloadOnGet(t *testing.T) {
	store := NewMemoryStore()

	payload := json.RawMessage(`{"message":"original"}`)
	newJob := job.Job{
		ID:      "job-1",
		Type:    "send-email",
		Payload: payload,
		Status:  job.StatusQueued,
	}

	if err := store.Create(newJob); err != nil {
		t.Fatalf("Create() returned unexpected error: %v", err)
	}

	first, err := store.Get(newJob.ID)
	if err != nil {
		t.Fatalf("first Get() returned unexpected error: %v", err)
	}

	// Mutate the retrieved payload
	first.Payload[2] = 'X'

	second, err := store.Get(newJob.ID)
	if err != nil {
		t.Fatalf("second Get() returned unexpected error: %v", err)
	}

	wantPayload := json.RawMessage(`{"message":"original"}`)
	if !reflect.DeepEqual(second.Payload, wantPayload) {
		t.Errorf("stored payload = %s, want %s", second.Payload, wantPayload)
	}
}

func TestMemoryConcurrentAccess(t *testing.T) {
	store := NewMemoryStore()

	const workers = 50
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			newJob := job.Job{
				ID:      "job-" + strconv.Itoa(i),
				Type:    "send-email",
				Payload: json.RawMessage(`{"message":"hello"}`),
				Status:  job.StatusQueued,
			}

			if err := store.Create(newJob); err != nil {
				t.Errorf("Create() returned unexpected error: %v", err)
				return
			}

			got, err := store.Get(newJob.ID)
			if err != nil {
				t.Errorf("Get() returned unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(got, newJob) {
				t.Errorf("Get() = %#v, want %#v", got, newJob)
			}
		}(i)
	}
	wg.Wait()
}
