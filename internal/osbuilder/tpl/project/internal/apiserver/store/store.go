package store

import (
	"context"
	"sync"

	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/onexstack/onexstack/pkg/store/where"
)

// ProviderSet defines the dependency injection providers for the store layer.
// It binds the abstract Interface to the concrete implementation *store.
var ProviderSet = wire.NewSet(NewStore, wire.Bind(new(IStore), new(*store)))

var (
	once sync.Once
	// S is a global variable for convenient access to the initialized store
	// instance from other packages.
	S *store
)

// IStore defines the methods that the persistence layer must implement.
type IStore interface {
	// DB returns the underlying *gorm.DB instance, optionally applying filter conditions.
	DB(ctx context.Context, wheres ...where.Where) *gorm.DB
	// TX executes the given function within a database transaction.
	TX(ctx context.Context, fn func(ctx context.Context) error) error
    {{- if .Web.WithUser}}
    User() UserStore
    {{- end}}
}

// txKey is the context key for storing the transaction *gorm.DB instance.
type txKey struct{}

// store is the concrete implementation of the Interface.
type store struct {
	db *gorm.DB

	// Additional database instances can be added as needed.
	// Example: fake *gorm.DB
}

// Ensure store implements the Interface at compile time.
var _ IStore = (*store)(nil)

// NewStore creates and returns a new store instance.
func NewStore(db *gorm.DB) *store {
	// Initialize the singleton store instance only once.
	once.Do(func() { S = &store{db} })

	return S
}

// DB returns the database instance. If a transaction exists in the context,
// it returns the transactional DB; otherwise, it returns the core DB.
// Optional 'where' clauses can be applied to the returned DB instance.
func (s *store) DB(ctx context.Context, wheres ...where.Where) *gorm.DB {
	db := s.db
	// Retrieve transaction from context if it exists.
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		db = tx
	}

	for _, w := range wheres {
		db = w.Where(db)
	}

	return db
}

// FakeDB is used to demonstrate multiple database instances.
// It returns a nil gorm.DB, indicating a fake database.
func (s *store) FakeDB(ctx context.Context) *gorm.DB { return nil }

// TX executes the provided function fn within a database transaction.
// It injects the transaction handle into the context passed to fn.
func (s *store) TX(ctx context.Context, fn func(ctx context.Context) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Inject the transaction DB into the context.
		ctx := context.WithValue(ctx, txKey{}, tx)
		return fn(ctx)
	})
}

{{- if .Web.WithUser}}
// User returns an instance that implements the UserStore interface.
func (s *store) User() UserStore {
    return newUserStore(store)
}
{{- end}}
