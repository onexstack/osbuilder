package asyncstore

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/wire"
	retryutil "github.com/onexstack/onexstack/pkg/util/retry"
)

// ProviderSet is the Wire provider set for the asyncstore package.
// It binds the Factory interface to the concrete store implementation.
var ProviderSet = wire.NewSet(NewStore, wire.Bind(new(Factory), new(*store)))

// Factory defines the interface for accessing asynchronous data stores.
// It acts as a gateway to various specific store implementations.
type Factory interface {
	Fake() *FakeStore
}

// store is the concrete implementation of the Factory interface.
type store struct {
	fakeStore *FakeStore
}

var (
	// S is the global singleton instance of the store factory.
	// It allows global access to the initialized store from legacy packages.
	S *store
	// once ensures that the store initialization logic executes exactly once.
	once sync.Once
)

// NewStore initializes the store factory and starts necessary background synchronization tasks.
// It implements the singleton pattern, ensuring only one instance and one sync loop exist.
func NewStore(ctx context.Context, interval time.Duration) Factory {
	once.Do(func() {
		fs := NewFakeStore()

		// Start the background synchronization task.
		// We log the error instead of returning it to maintain the constructor signature,
		// but in a critical system, this might warrant a panic or signature change.
		if err := retryutil.RunImmediatelyThenPeriod(ctx, fs.Sync, interval); err != nil {
			slog.ErrorContext(ctx, "failed to start async store background sync", "store", "fake_store", "err", err)
		}

		// Initialize the global singleton.
		S = &store{fakeStore: fs}
	})

	return S
}

// Fake returns the underlying FakeStore instance.
func (s *store) Fake() *FakeStore {
	return s.fakeStore
}
