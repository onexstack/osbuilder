package handler

import (
    {{- if .Web.WithOTel}}
    "log/slog"
    {{- end}}
	"context"

	{{.Web.APIImportPath}}
    {{- if .Web.WithOTel}}
    "go.opentelemetry.io/otel"
 
    "{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/metrics"
    {{- end}}
)

// Create{{.Web.R.SingularName}} handles the creation of a new {{.Web.R.SingularLower}}.
func (h *Handler) Create{{.Web.R.SingularName}}(ctx context.Context, rq *{{.D.APIAlias}}.Create{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Create{{.Web.R.SingularName}}Response, error) {
	{{- if .Web.WithOTel}}
    ctx, span := otel.Tracer("handler").Start(ctx, "Handler.Create{{.Web.R.SingularName}}")
    defer span.End()

    metrics.M.RecordResourceCreate(ctx, "{{.Web.R.SingularLower}}", span.SpanContext().TraceID().String())
    {{- end}}

    slog.InfoContext(ctx, "Processing {{.Web.R.SingularLower}} creation request", "layer", "handler")

	return h.biz.{{.Web.R.BusinessFactoryName}}().Create(ctx, rq)
}

// Update{{.Web.R.SingularName}} handles updating an existing {{.Web.R.SingularLower}}'s details.
func (h *Handler) Update{{.Web.R.SingularName}}(ctx context.Context, rq *{{.D.APIAlias}}.Update{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Update{{.Web.R.SingularName}}Response, error) {
	return h.biz.{{.Web.R.BusinessFactoryName}}().Update(ctx, rq)
}

// Delete{{.Web.R.SingularName}} handles the deletion of a single {{.Web.R.SingularLower}}.
func (h *Handler) Delete{{.Web.R.SingularName}}(ctx context.Context, rq *{{.D.APIAlias}}.Delete{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Delete{{.Web.R.SingularName}}Response, error) {
	return h.biz.{{.Web.R.BusinessFactoryName}}().Delete(ctx, rq)
}

// Delete{{.Web.R.PluralName}} deletes one or more {{.Web.R.PluralName}}.
func (h *Handler) Delete{{.Web.R.PluralName}}(ctx context.Context, rq *{{.D.APIAlias}}.Delete{{.Web.R.PluralName}}Request) (*{{.D.APIAlias}}.Delete{{.Web.R.PluralName}}Response, error) {
    return h.biz.{{.Web.R.BusinessFactoryName}}().DeleteCollection(ctx, rq)
}

// Get{{.Web.R.SingularName}} retrieves information about a specific {{.Web.R.SingularLower}}.
func (h *Handler) Get{{.Web.R.SingularName}}(ctx context.Context, rq *{{.D.APIAlias}}.Get{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.Get{{.Web.R.SingularName}}Response, error) {
	{{- if .Web.WithOTel}}
    ctx, span := otel.Tracer("handler").Start(ctx, "Handler.Get{{.Web.R.SingularName}}")
    defer span.End()

    metrics.M.RecordResourceGet(ctx, "{{.Web.R.SingularLower}}", span.SpanContext().TraceID().String())
    {{- end}}

    slog.InfoContext(ctx, "Processing {{.Web.R.SingularLower}} retrive request", "layer", "handler")

	return h.biz.{{.Web.R.BusinessFactoryName}}().Get(ctx, rq)
}

// List{{.Web.R.SingularName}} retrieves a list of {{.Web.R.PluralLower}} based on query parameters.
func (h *Handler) List{{.Web.R.SingularName}}(ctx context.Context, rq *{{.D.APIAlias}}.List{{.Web.R.SingularName}}Request) (*{{.D.APIAlias}}.List{{.Web.R.SingularName}}Response, error) {
	return h.biz.{{.Web.R.BusinessFactoryName}}().List(ctx, rq)
}
