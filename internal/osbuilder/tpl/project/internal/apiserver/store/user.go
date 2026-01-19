// nolint: dupl
package store

import (
	"context"

	storelogger "github.com/onexstack/onexstack/pkg/logger/slog/store"
	genericstore "github.com/onexstack/onexstack/pkg/store"
	"github.com/onexstack/onexstack/pkg/store/where"

	"{{.D.ModuleName}}/internal/{{.Web.Name}}/model"
)

// UserStore defines the interface for managing user-related persistent data.
type UserStore interface {
    // Create persists a new user record.
    Create(ctx context.Context, obj *model.UserM) error

    // Update modifies an existing user record.
    Update(ctx context.Context, obj *model.UserM) error

    // Delete removes user records matching the specified criteria.
    Delete(ctx context.Context, opts *where.Options) error

    // Get retrieves a single user record matching the specified criteria.
    Get(ctx context.Context, opts *where.Options) (*model.UserM, error)

    // List retrieves a list of user records and the total count matching the criteria.
    List(ctx context.Context, opts *where.Options) (int64, []*model.UserM, error)

    // UserExpansion defines custom methods for the User store outside the generic CRUD operations.
    UserExpansion
}

// UserExpansion is an extension interface for UserStore.
type UserExpansion interface{}

// userStore implements the UserStore interface using a generic store implementation.
type userStore struct {
    *genericstore.Store[model.UserM]
}

// Ensure userStore implements UserStore at compile time.
var _ UserStore = (*userStore)(nil)

// newUserStore returns a new instance of UserStore.
func newUserStore(s *store) *userStore {
    return &userStore{
        Store: genericstore.NewStore[model.UserM](s, storelogger.NewLogger()),
    }
}
