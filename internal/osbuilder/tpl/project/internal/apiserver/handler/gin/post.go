package handler

import (
	{{- if .Web.WithOTel}}
    "log/slog"
    {{- end}}
	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"
	{{- if .Web.WithOTel}}
    "go.opentelemetry.io/otel"

    "{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/metrics"
    {{- end}}
)

// Create{{.Web.R.SingularName}} handles the creation of a new {{.Web.R.SingularLower}}.
func (h *Handler) Create{{.Web.R.SingularName}}(c *gin.Context) {
	{{- if .Web.WithOTel}}
    ctx, span := otel.Tracer("handler").Start(c.Request.Context(), "Handler.Create{{.Web.R.SingularName}}")
    defer span.End()

	c.Request = c.Request.WithContext(ctx)

	metrics.M.RecordResourceCreate(c.Request.Context(), "{{.Web.R.SingularLower}}", span.SpanContext().TraceID().String())
    {{- end}}

    slog.InfoContext(ctx, "Processing {{.Web.R.SingularLower}} creation request", "layer", "handler")

	core.HandleJSONRequest(c, h.biz.{{.Web.R.BusinessFactoryName}}().Create, h.val.ValidateCreate{{.Web.R.SingularName}}Request)
}

// Update{{.Web.R.SingularName}} handles updating an existing {{.Web.R.SingularLower}}'s details.
func (h *Handler) Update{{.Web.R.SingularName}}(c *gin.Context) {
	core.HandleAllRequest(c, h.biz.{{.Web.R.BusinessFactoryName}}().Update, h.val.ValidateUpdate{{.Web.R.SingularName}}Request)
}

// Delete{{.Web.R.SingularName}} handles the deletion of a single {{.Web.R.SingularLower}} specified by a URI parameter.
func (h *Handler) Delete{{.Web.R.SingularName}}(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.{{.Web.R.BusinessFactoryName}}().Delete, h.val.ValidateDelete{{.Web.R.SingularName}}Request)
}

// Delete{{.Web.R.PluralName}} deletes one or more {{.Web.R.PluralName}} specified in the JSON request body.
func (h *Handler) Delete{{.Web.R.PluralName}}(c *gin.Context) {
    core.HandleJSONRequest(c, h.biz.{{.Web.R.BusinessFactoryName}}().DeleteCollection, h.val.ValidateDelete{{.Web.R.PluralName}}Request)
}

// Get{{.Web.R.SingularName}} retrieves information about a specific {{.Web.R.SingularLower}}.
func (h *Handler) Get{{.Web.R.SingularName}}(c *gin.Context) {
	{{- if .Web.WithOTel}}
    ctx, span := otel.Tracer("handler").Start( c.Request.Context(), "Handler.Get{{.Web.R.SingularName}}")
    defer span.End()

	c.Request = c.Request.WithContext(ctx)

	metrics.M.RecordResourceGet(c.Request.Context(), "{{.Web.R.SingularLower}}", span.SpanContext().TraceID().String())
    {{- end}}

    slog.InfoContext(ctx, "Processing {{.Web.R.SingularLower}} retrive request", "layer", "handler")

	core.HandleUriRequest(c, h.biz.{{.Web.R.BusinessFactoryName}}().Get, h.val.ValidateGet{{.Web.R.SingularName}}Request)
}

// List{{.Web.R.SingularName}} retrieves a list of {{.Web.R.PluralLower}} based on query parameters.
func (h *Handler) List{{.Web.R.SingularName}}(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.{{.Web.R.BusinessFactoryName}}().List, h.val.ValidateList{{.Web.R.SingularName}}Request)
}

func init() {
	Register(func(v1 *gin.RouterGroup, handler *Handler) {
		{{- if ne .Web.R.ResourcePathPrefix "" }}
		rg := v1.Group("/{{.Web.R.ResourcePathPrefix}}/{{.Web.R.Last.PluralLower}}", handler.mws...)
		{{- else}}
		rg := v1.Group("/{{.Web.R.Last.PluralLower}}", handler.mws...)
		{{- end}}
		rg.POST("", handler.Create{{.Web.R.SingularName}})
		rg.PUT(":{{.Web.R.Last.SingularLowerFirst}}ID", handler.Update{{.Web.R.SingularName}})
		rg.DELETE(":{{.Web.R.Last.SingularLowerFirst}}ID", handler.Delete{{.Web.R.SingularName}})
		rg.DELETE("", handler.Delete{{.Web.R.PluralName}})
		rg.GET(":{{.Web.R.Last.SingularLowerFirst}}ID", handler.Get{{.Web.R.SingularName}})
		rg.GET("", handler.List{{.Web.R.SingularName}})
	})
}
