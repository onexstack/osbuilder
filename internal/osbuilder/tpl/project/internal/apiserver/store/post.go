// nolint: dupl
package store

import (
	"context"

	genericstore "github.com/onexstack/onexstack/pkg/store"
	storelogger "github.com/onexstack/onexstack/pkg/logger/slog/store"
	"github.com/onexstack/onexstack/pkg/store/where"

	"{{.D.ModuleName}}/internal/{{.Web.Name}}/model"
)

// {{.Web.R.SingularName}}Store defines the interface for managing {{.Web.R.SingularLower}}-related data operations.
type {{.Web.R.SingularName}}Store interface {
	// Create inserts a new {{.Web.R.SingularName}} record into the store.
	Create(ctx context.Context, obj *model.{{.Web.R.GORMModel}}) error

	// Update modifies an existing {{.Web.R.SingularName}} record in the store based on the given model.
	Update(ctx context.Context, obj *model.{{.Web.R.GORMModel}}) error

	// Delete removes {{.Web.R.SingularName}} records that satisfy the given query options.
	Delete(ctx context.Context, opts *where.Options) error

	// Get retrieves a single {{.Web.R.SingularName}} record that satisfies the given query options.
	Get(ctx context.Context, opts *where.Options) (*model.{{.Web.R.GORMModel}}, error)

	// List retrieves a list of {{.Web.R.SingularName}} records and their total count based on the given query options.
	List(ctx context.Context, opts *where.Options) (int64, []*model.{{.Web.R.GORMModel}}, error)

	// {{.Web.R.SingularName}}Expansion is a placeholder for extension methods for {{.Web.R.PluralLower}},
	// to be implemented by additional interfaces if needed.
	{{.Web.R.SingularName}}Expansion
}

// {{.Web.R.SingularName}}Expansion is an empty interface provided for extending
// the {{.Web.R.SingularName}}Store interface.
// Developers can define {{.Web.R.SingularLower}}-specific additional methods
// in this interface for future expansion.
type {{.Web.R.SingularName}}Expansion interface{}

// {{.Web.R.SingularLowerFirst}}Store implements the {{.Web.R.SingularName}}Store interface and provides
// default implementations of the methods.
type {{.Web.R.SingularLowerFirst}}Store struct {
	*genericstore.Store[model.{{.Web.R.GORMModel}}]
}

// Ensure that {{.Web.R.SingularLowerFirst}}Store satisfies the {{.Web.R.SingularName}}Store interface at compile time.
var _ {{.Web.R.SingularName}}Store = (*{{.Web.R.SingularLowerFirst}}Store)(nil)

// new{{.Web.R.SingularName}}Store creates a new {{.Web.R.SingularLowerFirst}}Store instance with the provided
// datastore and logger.
func new{{.Web.R.SingularName}}Store(store *datastore) *{{.Web.R.SingularLowerFirst}}Store {
	return &{{.Web.R.SingularLowerFirst}}Store{
		Store: genericstore.NewStore[model.{{.Web.R.GORMModel}}](store, storelogger.NewLogger()),
	}
}
