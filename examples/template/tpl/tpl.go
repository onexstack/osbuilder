package {{.SingularLower}}

//go:generate mockgen -destination mock_{{.SingularLower}}.go -package {{.SingularLower}} {{.ModuleName}}/internal/{{.ComponentName}}/biz/v1/{{.SingularLower}} {{.SingularName}}Biz

import (
	"context"

	"github.com/onexstack/onex/pkg/core"
	"github.com/onexstack/onexstack/pkg/store/where"

	"{{.ModuleName}}/internal/{{.ComponentName}}/model"
	"{{.ModuleName}}/internal/{{.ComponentName}}/pkg/conversion"
	"{{.ModuleName}}/internal/{{.ComponentName}}/store"
	{{- if .WithUser}}
	"{{.ModuleName}}/internal/pkg/contextx"
	{{- end}}
	{{.APIImportPath}}
)

const (
	// MaxErrGroupConcurrency defines the maximum concurrency level
	// for error group operations.
	MaxErrGroupConcurrency = 100
)

// {{.SingularName}}Biz defines the interface that contains methods for handling {{.SingularLower}} requests.
type {{.SingularName}}Biz interface {
	Create(ctx context.Context, rq *{{.APIAlias}}.Create{{.SingularName}}Request) (*{{.APIAlias}}.Create{{.SingularName}}Response, error)
	Update(ctx context.Context, rq *{{.APIAlias}}.Update{{.SingularName}}Request) (*{{.APIAlias}}.Update{{.SingularName}}Response, error)
	Delete(ctx context.Context, rq *{{.APIAlias}}.Delete{{.SingularName}}Request) (*{{.APIAlias}}.Delete{{.SingularName}}Response, error)
	Get(ctx context.Context, rq *{{.APIAlias}}.Get{{.SingularName}}Request) (*{{.APIAlias}}.Get{{.SingularName}}Response, error)
	List(ctx context.Context, rq *{{.APIAlias}}.List{{.SingularName}}Request) (*{{.APIAlias}}.List{{.SingularName}}Response, error)

	{{.SingularName}}Expansion
}

// {{.SingularName}}Expansion defines additional methods for {{.SingularLower}} operations.
type {{.SingularName}}Expansion interface{}

// {{.SingularLowerFirst}}Biz is the implementation of the {{.SingularName}}Biz interface.
type {{.SingularLowerFirst}}Biz struct {
	store store.IStore
}

// Ensure that *{{.SingularLowerFirst}}Biz implements the {{.SingularName}}Biz interface.
var _ {{.SingularName}}Biz = (*{{.SingularLowerFirst}}Biz)(nil)

// New creates and returns a new instance of *{{.SingularLowerFirst}}Biz.
func New(store store.IStore) *{{.SingularLowerFirst}}Biz {
	return &{{.SingularLowerFirst}}Biz{store: store}
}

// Create implements the Create method of the {{.SingularName}}Biz interface.
// It creates a new {{.SingularLower}} based on the provided *Create{{.SingularName}}Request.
func (b *{{.SingularLowerFirst}}Biz) Create(ctx context.Context, rq *{{.APIAlias}}.Create{{.SingularName}}Request) (*{{.APIAlias}}.Create{{.SingularName}}Response, error) {
	var {{.SingularLowerFirst}}M model.{{.GORMModel}}
	_ = core.Copy(&{{.SingularLowerFirst}}M, rq)
	{{- if .WithUser}}
	{{.SingularLowerFirst}}M.UserID = contextx.UserID(ctx)
	{{- end}}

	if err := b.store.{{.SingularName}}().Create(ctx, &{{.SingularLowerFirst}}M); err != nil {
		return nil, err
	}

	return &{{.APIAlias}}.Create{{.SingularName}}Response{ {{.SingularName}}ID: {{.SingularLowerFirst}}M.{{.SingularName}}ID}, nil
}

// Update implements the Update method of the {{.SingularName}}Biz interface.
// It updates an existing {{.SingularLower}} based on the provided *Update{{.SingularName}}Request.
func (b *{{.SingularLowerFirst}}Biz) Update(ctx context.Context, rq *{{.APIAlias}}.Update{{.SingularName}}Request) (*{{.APIAlias}}.Update{{.SingularName}}Response, error) {
	whr := where.T(ctx).F("{{.SingularLowerFirst}}ID", rq.Get{{.SingularName}}ID())
	{{.SingularLowerFirst}}M, err := b.store.{{.SingularName}}().Get(ctx, whr)
	if err != nil {
		return nil, err
	}

	// TODO: Implement additional business logic here.

	if err := b.store.{{.SingularName}}().Update(ctx, {{.SingularLowerFirst}}M); err != nil {
		return nil, err
	}

	return &{{.APIAlias}}.Update{{.SingularName}}Response{}, nil
}

// Delete implements the Delete method of the {{.SingularName}}Biz interface.
// It deletes one or more {{.PluralLower}} based on the provided *Delete{{.SingularName}}Request.
func (b *{{.SingularLowerFirst}}Biz) Delete(ctx context.Context, rq *{{.APIAlias}}.Delete{{.SingularName}}Request) (*{{.APIAlias}}.Delete{{.SingularName}}Response, error) {
	whr := where.T(ctx).F("{{.SingularLowerFirst}}ID", rq.Get{{.SingularName}}IDs())
	if err := b.store.{{.SingularName}}().Delete(ctx, whr); err != nil {
		return nil, err
	}

	return &{{.APIAlias}}.Delete{{.SingularName}}Response{}, nil
}

// Get implements the Get method of the {{.SingularName}}Biz interface.
// It retrieves a specific {{.SingularLower}} based on the provided *Get{{.SingularName}}Request.
func (b *{{.SingularLowerFirst}}Biz) Get(ctx context.Context, rq *{{.APIAlias}}.Get{{.SingularName}}Request) (*{{.APIAlias}}.Get{{.SingularName}}Response, error) {
	whr := where.T(ctx).F("{{.SingularLowerFirst}}ID", rq.Get{{.SingularName}}ID())
	{{.SingularLowerFirst}}M, err := b.store.{{.SingularName}}().Get(ctx, whr)
	if err != nil {
		return nil, err
	}

	return &{{.APIAlias}}.Get{{.SingularName}}Response{ {{.SingularName}}: conversion.{{.MapModelToAPIFunc}}({{.SingularLowerFirst}}M)}, nil
}

// List implements the List method of the {{.SingularName}}Biz interface.
// It retrieves a list of {{.PluralLower}} and their total count based on the provided *List{{.SingularName}}Request.
func (b *{{.SingularLowerFirst}}Biz) List(ctx context.Context, rq *{{.APIAlias}}.List{{.SingularName}}Request) (*{{.APIAlias}}.List{{.SingularName}}Response, error) {
	whr := where.T(ctx).P(int(rq.GetOffset()), int(rq.GetLimit()))
	count, {{.SingularLowerFirst}}List, err := b.store.{{.SingularName}}().List(ctx, whr)
	if err != nil {
		return nil, err
	}

	var m sync.Map
	eg, ctx := errgroup.WithContext(ctx)

	// Set the maximum concurrency limit using the constant MaxConcurrency
	eg.SetLimit(MaxErrGroupConcurrency)

	// Use goroutines to improve API performance
	for _, {{.SingularLowerFirst}} := range {{.SingularLowerFirst}}List {
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return nil
			default:
				converted := conversion.{{.MapModelToAPIFunc}}({{.SingularLowerFirst}})
				// TODO: Add additional processing logic and assign values to fields
				// that need updating, for example:
				// xxx := doSomething()
				// converted.XXX = xxx
				m.Store({{.SingularLowerFirst}}.ID, converted)

				return nil
			}
		})
	}

	if err := eg.Wait(); err != nil {
		log.W(ctx).Errorw(err, "Failed to wait all function calls returned")
		return nil, err
	}

	{{.PluralLowerFirst}} := make([]*{{.APIAlias}}.{{.SingularName}}, 0, len({{.SingularLowerFirst}}List))
	for _, item := range {{.SingularLowerFirst}}List {
		{{.SingularLowerFirst}}, _ := m.Load(item.ID)
		{{.PluralLowerFirst}} = append({{.PluralLowerFirst}}, {{.SingularLowerFirst}}.(*{{.APIAlias}}.{{.SingularName}}))
	}

	return &{{.APIAlias}}.List{{.SingularName}}Response{Total: count, {{.PluralName}}: {{.PluralLowerFirst}}}, nil
}
