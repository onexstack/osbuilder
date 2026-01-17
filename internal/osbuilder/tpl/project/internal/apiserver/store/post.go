// nolint: dupl
package store

import (
	"context"

	storelogger "github.com/onexstack/onexstack/pkg/logger/slog/store"
	genericstore "github.com/onexstack/onexstack/pkg/store"
	"github.com/onexstack/onexstack/pkg/store/where"

	"{{.D.ModuleName}}/internal/{{.Web.Name}}/model"
)

// {{.Web.R.SingularName}}Store defines the interface for managing {{.Web.R.SingularLower}}-related persistent data.
type {{.Web.R.SingularName}}Store interface {
	// Create persists a new {{.Web.R.SingularLower}} record.
	Create(ctx context.Context, obj *model.{{.Web.R.GORMModel}}) error

	// Update modifies an existing {{.Web.R.SingularLower}} record.
	Update(ctx context.Context, obj *model.{{.Web.R.GORMModel}}) error

	// Delete removes {{.Web.R.SingularLower}} records matching the specified criteria.
	Delete(ctx context.Context, opts *where.Options) error

	// Get retrieves a single {{.Web.R.SingularLower}} record matching the specified criteria.
	Get(ctx context.Context, opts *where.Options) (*model.{{.Web.R.GORMModel}}, error)

	// List retrieves a list of {{.Web.R.SingularLower}} records and the total count matching the criteria.
	List(ctx context.Context, opts *where.Options) (int64, []*model.{{.Web.R.GORMModel}}, error)

	// {{.Web.R.SingularName}}Expansion defines custom methods for the {{.Web.R.SingularLower}} store outside the generic CRUD operations.
	{{.Web.R.SingularName}}Expansion
}

// {{.Web.R.SingularName}}Expansion is an extension interface for {{.Web.R.SingularName}}Store.
type {{.Web.R.SingularName}}Expansion interface{}

// {{.Web.R.SingularLower}}Store implements the {{.Web.R.SingularName}}Store interface using a generic store implementation.
type {{.Web.R.SingularLower}}Store struct {
	*genericstore.Store[model.{{.Web.R.GORMModel}}]
}

// Ensure {{.Web.R.SingularLower}}Store implements {{.Web.R.SingularName}}Store at compile time.
var _ {{.Web.R.SingularName}}Store = (*{{.Web.R.SingularLower}}Store)(nil)

// new{{.Web.R.SingularName}}Store returns a new instance of {{.Web.R.SingularName}}Store.
func new{{.Web.R.SingularName}}Store(s *store) *{{.Web.R.SingularLower}}Store {
	return &{{.Web.R.SingularLower}}Store{
		Store: genericstore.NewStore[model.{{.Web.R.GORMModel}}](s, storelogger.NewLogger()),
	}
}
