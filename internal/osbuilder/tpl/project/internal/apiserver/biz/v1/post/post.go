package {{.Web.R.SingularLower}}

import (
	"sync"
	"context"

	"github.com/onexstack/onexstack/pkg/core"
	"golang.org/x/sync/errgroup"
	"k8s.io/klog/v2"
	"github.com/onexstack/onexstack/pkg/store/where"

	"{{.D.ModuleName}}/internal/{{.Web.Name}}/model"
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/conversion"
	"{{.D.ModuleName}}/internal/pkg/known"
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/store"
	// "{{.D.ModuleName}}/internal/pkg/contextx"
	{{.Web.APIImportPath}}
)

// {{.Web.R.SingularName}}Biz defines the interface that contains methods for handling {{.Web.R.SingularLower}} requests.
type {{.Web.R.SingularName}}Biz interface {
	// Create creates a new {{.Web.R.SingularLower}} based on the provided request parameters.
	Create(ctx context.Context, rq *{{.D.APIAlias}}.Create{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Create{{.Web.R.SingularName}}Response, error)

	// Update updates an existing {{.Web.R.SingularLower}} based on the provided request parameters.
	Update(ctx context.Context, rq *{{.D.APIAlias}}.Update{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Update{{.Web.R.SingularName}}Response, error)

	// Delete removes one or more {{.Web.R.PluralLower}} based on the provided request parameters.
	Delete(ctx context.Context, rq *{{.D.APIAlias}}.Delete{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Delete{{.Web.R.SingularName}}Response, error)

	// Get retrieves the details of a specific {{.Web.R.SingularLower}} based on the provided request parameters.
	Get(ctx context.Context, rq *{{.D.APIAlias}}.Get{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Get{{.Web.R.SingularName}}Response, error)

	// List retrieves a list of {{.Web.R.PluralLower}} and their total count based on the provided request parameters.
	List(ctx context.Context, rq *{{.D.APIAlias}}.List{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.List{{.Web.R.SingularName}}Response, error)

	// {{.Web.R.SingularName}}Expansion defines additional methods for extended {{.Web.R.SingularLower}} operations, if needed.
	{{.Web.R.SingularName}}Expansion
}

// {{.Web.R.SingularName}}Expansion defines additional methods for {{.Web.R.SingularLower}} operations.
type {{.Web.R.SingularName}}Expansion interface{}

// {{.Web.R.SingularLowerFirst}}Biz is the implementation of the {{.Web.R.SingularName}}Biz.
type {{.Web.R.SingularLowerFirst}}Biz struct {
	store store.IStore
}

// Ensure that *{{.Web.R.SingularLowerFirst}}Biz implements the {{.Web.R.SingularName}}Biz.
var _ {{.Web.R.SingularName}}Biz = (*{{.Web.R.SingularLowerFirst}}Biz)(nil)

// New creates and returns a new instance of *{{.Web.R.SingularLowerFirst}}Biz.
func New(store store.IStore) *{{.Web.R.SingularLowerFirst}}Biz {
	return &{{.Web.R.SingularLowerFirst}}Biz{store: store}
}

// Create implements the Create method of the {{.Web.R.SingularName}}Biz.
func (b *{{.Web.R.SingularLowerFirst}}Biz) Create(ctx context.Context, rq *{{.D.APIAlias}}.Create{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Create{{.Web.R.SingularName}}Response, error) {
	var {{.Web.R.SingularLowerFirst}}M model.{{.Web.R.GORMModel}}
	_ = core.Copy(&{{.Web.R.SingularLowerFirst}}M, rq)
	// TODO: Retrieve the UserID from the custom context and assign it as needed.
	// {{.Web.R.SingularLowerFirst}}M.UserID = contextx.UserID(ctx)

	if err := b.store.{{.Web.R.SingularName}}().Create(ctx, &{{.Web.R.SingularLowerFirst}}M); err != nil {
		return nil, err
	}

	return &{{.D.APIAlias}}.Create{{.Web.R.SingularName}}Response{ {{.Web.R.SingularName}}ID: {{.Web.R.SingularLowerFirst}}M.{{.Web.R.SingularName}}ID}, nil
}

// Update implements the Update method of the {{.Web.R.SingularName}}Biz.
func (b *{{.Web.R.SingularLowerFirst}}Biz) Update(ctx context.Context, rq *{{.D.APIAlias}}.Update{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Update{{.Web.R.SingularName}}Response, error) {
	whr := where.F("{{.Web.R.SingularLowerFirst}}ID", rq.Get{{.Web.R.SingularName}}ID())
	{{.Web.R.SingularLowerFirst}}M, err := b.store.{{.Web.R.SingularName}}().Get(ctx, whr)
	if err != nil {
		return nil, err
	}

	// TODO: Implement additional business logic here.

	if err := b.store.{{.Web.R.SingularName}}().Update(ctx, {{.Web.R.SingularLowerFirst}}M); err != nil {
		return nil, err
	}

	return &{{.D.APIAlias}}.Update{{.Web.R.SingularName}}Response{}, nil
}

// Delete implements the Delete method of the {{.Web.R.SingularName}}Biz.
func (b *{{.Web.R.SingularLowerFirst}}Biz) Delete(ctx context.Context, rq *{{.D.APIAlias}}.Delete{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Delete{{.Web.R.SingularName}}Response, error) {
	whr := where.F("{{.Web.R.SingularLowerFirst}}ID", rq.Get{{.Web.R.SingularName}}IDs())
	if err := b.store.{{.Web.R.SingularName}}().Delete(ctx, whr); err != nil {
		return nil, err
	}

	return &{{.D.APIAlias}}.Delete{{.Web.R.SingularName}}Response{}, nil
}

// Get implements the Get method of the {{.Web.R.SingularName}}Biz.
func (b *{{.Web.R.SingularLowerFirst}}Biz) Get(ctx context.Context, rq *{{.D.APIAlias}}.Get{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Get{{.Web.R.SingularName}}Response, error) {
	whr := where.F("{{.Web.R.SingularLowerFirst}}ID", rq.Get{{.Web.R.SingularName}}ID())
	{{.Web.R.SingularLowerFirst}}M, err := b.store.{{.Web.R.SingularName}}().Get(ctx, whr)
	if err != nil {
		return nil, err
	}

	return &{{.D.APIAlias}}.Get{{.Web.R.SingularName}}Response{ {{.Web.R.SingularName}}: conversion.{{.Web.R.MapModelToAPIFunc}}({{.Web.R.SingularLowerFirst}}M)}, nil
}

// List implements the List method of the {{.Web.R.SingularName}}Biz.
func (b *{{.Web.R.SingularLowerFirst}}Biz) List(ctx context.Context, rq *{{.D.APIAlias}}.List{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.List{{.Web.R.SingularName}}Response, error) {
	whr := where.P(int(rq.GetOffset()), int(rq.GetLimit()))
	count, {{.Web.R.SingularLowerFirst}}List, err := b.store.{{.Web.R.SingularName}}().List(ctx, whr)
	if err != nil {
		return nil, err
	}

	var m sync.Map
	eg, ctx := errgroup.WithContext(ctx)

	// Set the maximum concurrency limit using the constant MaxConcurrency
	eg.SetLimit(known.MaxErrGroupConcurrency)

	// Use goroutines to improve API performance
	for _, {{.Web.R.SingularLowerFirst}} := range {{.Web.R.SingularLowerFirst}}List {
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return nil
			default:
				converted := conversion.{{.Web.R.MapModelToAPIFunc}}({{.Web.R.SingularLowerFirst}})
				// TODO: Add additional processing logic and assign values to fields
				// that need updating, for example:
				// xxx := doSomething()
				// converted.XXX = xxx
				m.Store({{.Web.R.SingularLowerFirst}}.ID, converted)

				return nil
			}
		})
	}

	if err := eg.Wait(); err != nil {
		klog.FromContext(ctx).Error(err, "Failed to wait all function calls returned")
		return nil, err
	}

	{{.Web.R.PluralLowerFirst}} := make([]*{{.D.APIAlias}}.{{.Web.R.SingularName}}, 0, len({{.Web.R.SingularLowerFirst}}List))
	for _, item := range {{.Web.R.SingularLowerFirst}}List {
		{{.Web.R.SingularLowerFirst}}, _ := m.Load(item.ID)
		{{.Web.R.PluralLowerFirst}} = append({{.Web.R.PluralLowerFirst}}, {{.Web.R.SingularLowerFirst}}.(*{{.D.APIAlias}}.{{.Web.R.SingularName}}))
	}

	return &{{.D.APIAlias}}.List{{.Web.R.SingularName}}Response{Total: count, {{.Web.R.PluralName}}: {{.Web.R.PluralLowerFirst}}}, nil
}
