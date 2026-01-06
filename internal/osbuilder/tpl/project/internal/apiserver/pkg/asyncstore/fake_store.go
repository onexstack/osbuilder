package asyncstore

import (
	"context"
	"sync"

	"github.com/brianvoe/gofakeit/v7"

	v1 "github.com/onexstack/b-dms/pkg/api/apiserver/v1"
)

// FakeStore implements the storage mechanism for FakeData.
type FakeStore struct {
	// Use RWMutex to allow concurrent reads.
	mu    sync.RWMutex
	items map[string]*v1.FakeData
}

// NewFakeStore creates a new instance of FakeStore.
func NewFakeStore() *FakeStore {
	return &FakeStore{
		items: make(map[string]*v1.FakeData),
	}
}

// Sync simulates data synchronization by generating random fake data.
func (s *FakeStore) Sync(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Simulate generating 10 random items.
	newItems := make(map[string]*v1.FakeData)
	for i := 0; i < 10; i++ {
		// Generate a random UUID as the key
		id := gofakeit.UUID()

		newItems[id] = &v1.FakeData{
			ID:          id,
			Name:        gofakeit.AppName(),
			Category:    gofakeit.CarMaker(),
			Description: gofakeit.Phrase(),
			Status:      gofakeit.RandomString([]string{"active", "pending", "failed"}),
			Score:       gofakeit.Float64Range(0.0, 100.0),
		}
	}

	// Example: Add a fixed item for deterministic testing.
	fixedID := "fixed-item-001"
	newItems[fixedID] = &v1.FakeData{
		ID:          fixedID,
		Name:        "Fixed Test Item",
		Category:    "Testing",
		Description: "This item always exists.",
		Status:      "active",
		Score:       99.9,
	}

	// Atomically replace the entire map.
	s.items = newItems

	return nil
}

// Get retrieves an item by its ID.
// It returns the item pointer and a boolean indicating if the item was found.
func (s *FakeStore) Get(id string) (*v1.FakeData, bool) {
	s.mu.RLock() // Acquire read lock
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	return item, ok
}

// List retrieves all items.
func (s *FakeStore) List() []*v1.FakeData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*v1.FakeData, 0, len(s.items))
	for _, item := range s.items {
		list = append(list, item)
	}
	return list
}
