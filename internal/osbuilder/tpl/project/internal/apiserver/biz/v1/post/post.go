package {{.Web.R.Last.SingularLower}}

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/onexstack/onexstack/pkg/core"
	"github.com/onexstack/onexstack/pkg/store/where"
	{{- if .Web.WithOTel}}
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    oteltrace "go.opentelemetry.io/otel/trace"
    {{- end}}
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"{{.D.ModuleName}}/internal/{{.Web.Name}}/model"
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/conversion"
	"{{.D.ModuleName}}/internal/pkg/known"
	"{{.D.ModuleName}}/internal/pkg/errno"
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/store"
	// "{{.D.ModuleName}}/internal/pkg/contextx"
	{{.Web.APIImportPath}}
    {{- if .Web.Clients }}
    "{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/clientset"
    {{- end}}
)

// {{.Web.R.SingularName}}Biz defines the interface for handling {{.Web.R.SingularLower}}-related business logic.
type {{.Web.R.SingularName}}Biz interface {
	// Create creates a new {{.Web.R.SingularLower}} based on the provided request parameters.
	Create(ctx context.Context, rq *{{.D.APIAlias}}.Create{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Create{{.Web.R.SingularName}}Response, error)

	// Update updates an existing {{.Web.R.SingularLower}} based on the provided request parameters.
	Update(ctx context.Context, rq *{{.D.APIAlias}}.Update{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Update{{.Web.R.SingularName}}Response, error)

	// Delete remove one {{.Web.R.SingularLower}} based on the provided request parameters.
	Delete(ctx context.Context, rq *{{.D.APIAlias}}.Delete{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Delete{{.Web.R.SingularName}}Response, error)

	// DeleteCollection deletes a collection of {{.Web.R.PluralLower}} that match the specified criteria or identifiers.
	DeleteCollection(ctx context.Context, rq *v1.Delete{{.Web.R.PluralName}}Request) (*v1.Delete{{.Web.R.PluralName}}Response, error)

	// Get retrieves the details of a specific {{.Web.R.SingularLower}} based on the provided request parameters.
	Get(ctx context.Context, rq *{{.D.APIAlias}}.Get{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Get{{.Web.R.SingularName}}Response, error)

	// List retrieves a list of {{.Web.R.PluralLower}} and their total count based on the provided request parameters.
	List(ctx context.Context, rq *{{.D.APIAlias}}.List{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.List{{.Web.R.SingularName}}Response, error)

	// {{.Web.R.SingularName}}Expansion defines additional methods for extended {{.Web.R.SingularLower}} operations, if needed.
	{{.Web.R.SingularName}}Expansion
}

// {{.Web.R.SingularName}}Expansion defines custom methods for extended {{.Web.R.SingularLower}} business operations.
type {{.Web.R.SingularName}}Expansion interface{}

// {{.Web.R.SingularLowerFirst}}Biz implements the {{.Web.R.SingularName}}Biz interface.
type {{.Web.R.SingularLowerFirst}}Biz struct {
	store store.IStore
	{{- if .Web.Clients }}
	clientset clientset.Interface
	{{- end}}
}

// Ensure {{.Web.R.SingularLowerFirst}}Biz implements {{.Web.R.SingularName}}Biz at compile time.
var _ {{.Web.R.SingularName}}Biz = (*{{.Web.R.SingularLowerFirst}}Biz)(nil)

// New creates and returns a new instance of {{.Web.R.SingularName}}Biz.
func New(store store.IStore{{- if .Web.Clients }}, clientset clientset.Interface{{- end -}}) *{{.Web.R.SingularLowerFirst}}Biz {
	return &{{.Web.R.SingularLowerFirst}}Biz{store: store{{- if .Web.Clients}}, clientset: clientset{{- end -}}}
}

// Create implements the Create method of the {{.Web.R.SingularName}}Biz.
func (b *{{.Web.R.SingularLowerFirst}}Biz) Create(ctx context.Context, rq *{{.D.APIAlias}}.Create{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Create{{.Web.R.SingularName}}Response, error) {
	{{- if .Web.WithOTel}}
    ctx, span := otel.Tracer("biz").Start(ctx, "{{.Web.R.SingularLowerFirst}}Biz.Create")
    defer span.End()

    // Follow the component.operation.phase pattern
    span.AddEvent("{{.Web.R.Last.SingularLower}}.creation.started")
    {{- end}}

	var {{.Web.R.Last.SingularLowerFirst}}M model.{{.Web.R.GORMModel}}
	_ = core.Copy(&{{.Web.R.Last.SingularLowerFirst}}M, rq)
	// TODO: Retrieve the UserID from the custom context and assign it as needed.
	// {{.Web.R.SingularLowerFirst}}M.UserID = contextx.UserID(ctx)
                                                                                
	slog.InfoContext(ctx, "creating {{.Web.R.SingularLower}} in database")

	if err := b.store.{{.Web.R.SingularName}}().Create(ctx, &{{.Web.R.Last.SingularLowerFirst}}M); err != nil {
    	{{- if .Web.WithOTel}}
		core.RecordSpanError(ctx, span, err)
    	{{- end}}
		slog.ErrorContext(ctx, "failed to create {{.Web.R.SingularLower}}", "error", err)
		return nil, errno.Err{{.Web.R.SingularName}}CreateFailed.WithMessage(err.Error())
	}

	{{- if .Web.WithOTel}}
	span.AddEvent("{{.Web.R.Last.SingularLower}}.creation.completed", oteltrace.WithAttributes(attribute.String("{{.Web.R.Last.SingularLower}}_id", {{.Web.R.Last.SingularLowerFirst}}M.{{.Web.R.Last.SingularName}}ID)))
    {{- end}}
	return &{{.D.APIAlias}}.Create{{.Web.R.SingularName}}Response{ {{.Web.R.Last.SingularName}}ID: {{.Web.R.Last.SingularLowerFirst}}M.{{.Web.R.Last.SingularName}}ID}, nil
}

// Update implements the Update method of the {{.Web.R.SingularName}}Biz.
func (b *{{.Web.R.SingularLowerFirst}}Biz) Update(ctx context.Context, rq *{{.D.APIAlias}}.Update{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Update{{.Web.R.SingularName}}Response, error) {
	whr := where.F("{{.Web.R.Last.SingularLower}}_id", rq.{{.Web.R.Last.SingularName}}ID)
	{{.Web.R.Last.SingularLowerFirst}}M, err := b.store.{{.Web.R.SingularName}}().Get(ctx, whr)
	if err != nil {
		return nil, errno.Err{{.Web.R.SingularName}}UpdateFailed.WithMessage(err.Error())
	}

    // TODO: Apply updates to {{.Web.R.Last.SingularLowerFirst}}M from rq.
    // Example: {{.Web.R.Last.SingularLowerFirst}}M.Status = rq.Status

	if err := b.store.{{.Web.R.SingularName}}().Update(ctx, {{.Web.R.Last.SingularLowerFirst}}M); err != nil {
		return nil, errno.Err{{.Web.R.SingularName}}UpdateFailed.WithMessage(err.Error())
	}

	return &{{.D.APIAlias}}.Update{{.Web.R.SingularName}}Response{}, nil
}

// Delete implements the Delete method of the {{.Web.R.SingularName}}Biz.
func (b *{{.Web.R.SingularLowerFirst}}Biz) Delete(ctx context.Context, rq *{{.D.APIAlias}}.Delete{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Delete{{.Web.R.SingularName}}Response, error) {
	whr := where.F("{{.Web.R.Last.SingularLower}}_id", rq.{{.Web.R.Last.SingularName}}ID)
	if err := b.store.{{.Web.R.SingularName}}().Delete(ctx, whr); err != nil {
		return nil, errno.Err{{.Web.R.SingularName}}DeleteFailed.WithMessage(err.Error())
	}

	return &{{.D.APIAlias}}.Delete{{.Web.R.SingularName}}Response{}, nil
}

// DeleteCollection implements the DeleteCollection method of the {{.Web.R.SingularName}}Biz.
func (b *{{.Web.R.SingularLowerFirst}}Biz) DeleteCollection(ctx context.Context, rq *v1.Delete{{.Web.R.PluralName}}Request) (*v1.Delete{{.Web.R.PluralName}}Response, error) {
    whr := where.F("{{.Web.R.Last.SingularLower}}_id", rq.{{.Web.R.Last.SingularName}}IDs)
    if err := b.store.{{.Web.R.SingularName}}().Delete(ctx, whr); err != nil {
        return nil, errno.Err{{.Web.R.SingularName}}DeleteFailed.WithMessage(err.Error())
    }

    return &v1.Delete{{.Web.R.PluralName}}Response{}, nil
}

// Get implements the Get method of the {{.Web.R.SingularName}}Biz.
func (b *{{.Web.R.SingularLowerFirst}}Biz) Get(ctx context.Context, rq *{{.D.APIAlias}}.Get{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Get{{.Web.R.SingularName}}Response, error) {
	{{- if .Web.WithOTel}}
    ctx, span := otel.Tracer("biz").Start(ctx, "{{.Web.R.SingularLowerFirst}}Biz.Get")
    defer span.End()

	span.SetAttributes(attribute.String("{{.Web.R.Last.SingularLowerFirst}}_id", rq.{{.Web.R.Last.SingularName}}ID))
    {{- end}}

	slog.InfoContext(ctx, "retrieving job from database", "job_id", rq.{{.Web.R.Last.SingularName}}ID)

	whr := where.F("{{.Web.R.Last.SingularLower}}_id", rq.{{.Web.R.Last.SingularName}}ID)
	{{.Web.R.Last.SingularLowerFirst}}M, err := b.store.{{.Web.R.SingularName}}().Get(ctx, whr)
	if err != nil {
		{{- if .Web.WithOTel}}
		core.RecordSpanError(ctx, span, err)
    	{{- end}}
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errno.Err{{.Web.R.SingularName}}NotFound 
        }                       
		slog.ErrorContext(ctx, "failed to retrive {{.Web.R.SingularLower}}", "error", err, "{{.Web.R.Last.SingularLower}}_id", rq.{{.Web.R.Last.SingularName}}ID)
		return nil, errno.Err{{.Web.R.SingularName}}GetFailed.WithMessage(err.Error())
	}

	return &{{.D.APIAlias}}.Get{{.Web.R.SingularName}}Response{ {{.Web.R.Last.SingularName}}: conversion.{{.Web.R.MapModelToAPIFunc}}({{.Web.R.Last.SingularLowerFirst}}M)}, nil
}

// List implements the List method of the {{.Web.R.SingularName}}Biz.
func (b *{{.Web.R.SingularLowerFirst}}Biz) List(ctx context.Context, rq *{{.D.APIAlias}}.List{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.List{{.Web.R.SingularName}}Response, error) {
	whr := where.P(int(rq.Offset), int(rq.Limit))
	count, {{.Web.R.Last.SingularLowerFirst}}List, err := b.store.{{.Web.R.SingularName}}().List(ctx, whr)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list {{.Web.R.PluralLower}}", "error", err)
		return nil, errno.Err{{.Web.R.SingularName}}ListFailed.WithMessage(err.Error())
	}

	// Concurrent processing for list items conversion/enrichment.
	var m sync.Map
	eg, ctx := errgroup.WithContext(ctx)

	// Set the maximum concurrency limit using the constant MaxConcurrency
	eg.SetLimit(known.MaxErrGroupConcurrency)

	// Use goroutines to improve API performance
	for _, {{.Web.R.Last.SingularLowerFirst}} := range {{.Web.R.Last.SingularLowerFirst}}List {
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				converted := conversion.{{.Web.R.MapModelToAPIFunc}}({{.Web.R.Last.SingularLowerFirst}})
				// TODO: Add complex enrichment logic here if needed.
				m.Store({{.Web.R.Last.SingularLowerFirst}}.ID, converted)

				return nil
			}
		})
	}

	if err := eg.Wait(); err != nil {
		slog.ErrorContext(ctx, "error during concurrent {{.Web.R.SingularLower}} processing", "error", err)
		return nil, errno.Err{{.Web.R.SingularName}}ListFailed.WithMessage(err.Error())
	}

	// Reassemble the result in the correct order.
	{{.Web.R.Last.PluralLowerFirst}} := make([]*{{.D.APIAlias}}.{{.Web.R.SingularName}}, 0, len({{.Web.R.Last.SingularLowerFirst}}List))
	for _, item := range {{.Web.R.Last.SingularLowerFirst}}List {
        if val, ok := m.Load(item.ID); ok {
            {{.Web.R.Last.PluralLowerFirst}} = append({{.Web.R.Last.PluralLowerFirst}}, val.(*{{.D.APIAlias}}.{{.Web.R.SingularName}}))
        }
	}

	return &{{.D.APIAlias}}.List{{.Web.R.SingularName}}Response{Total: count, {{.Web.R.Last.PluralName}}: {{.Web.R.Last.PluralLowerFirst}}}, nil
}
