package asyncstore

import (
	"context"
	"sync"
	"time"

	"github.com/google/wire"
	retryutil "github.com/onexstack/onexstack/pkg/util/retry"
)

var ProviderSet = wire.NewSet(NewStore, wire.Bind(new(Factory), new(*store)))

// Factory defines the unified interface for retrieving various stores.
type Factory interface {
	Fake() *FakeStore
}

// store is the concrete implementation of the Factory interface.
type store struct {
	fakeStore *FakeStore
}

var (
	once sync.Once
	// S is a global variable for convenient access to the initialized datasyncstore
	// instance from other packages.
	S *store
)

// NewStore initializes and returns the Store Factory interface.
func NewStore(ctx context.Context, refreshInterval time.Duration) Factory {
	// Enable fake data async store
	fakeStore := NewFakeStore()
	_ = retryutil.RunImmediatelyThenPeriod(ctx, fakeStore.Sync, refreshInterval)

	// Initialize the singleton datasyncstore instance only once.
	once.Do(func() {
		S = &store{fakeStore: fakeStore}
	})

	return S
}

// Fake returns the FakeStore instance.
func (s *store) Fake() *FakeStore {
	return s.fakeStore
}
