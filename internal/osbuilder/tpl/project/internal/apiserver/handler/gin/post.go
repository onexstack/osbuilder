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

// Create{{.Web.R.SingularName}} handles the HTTP request to create a new {{.Web.R.SingularLower}}.
func (h *Handler) Create{{.Web.R.SingularName}}(c *gin.Context) {
	{{- if .Web.WithOTel}}
    ctx, span := otel.Tracer("handler").Start(c.Request.Context(), "Handler.Create{{.Web.R.SingularName}}")
    defer span.End()

	// Update the Gin request context so subsequent middleware/handlers use the traced context.
	c.Request = c.Request.WithContext(ctx)

	metrics.M.RecordResourceCreate(ctx, "{{.Web.R.SingularLower}}")
    {{- end}}

	slog.InfoContext(ctx, "processing {{.Web.R.SingularLower}} creation request")

	core.HandleJSONRequest(c, h.biz.{{.Web.R.BusinessFactoryName}}().Create, h.val.ValidateCreate{{.Web.R.SingularName}}Request)
}

// Update{{.Web.R.SingularName}} handles the HTTP request to update an existing {{.Web.R.SingularLower}}'s details.
func (h *Handler) Update{{.Web.R.SingularName}}(c *gin.Context) {
	core.HandleAllRequest(c, h.biz.{{.Web.R.BusinessFactoryName}}().Update, h.val.ValidateUpdate{{.Web.R.SingularName}}Request)
}

// Delete{{.Web.R.SingularName}} handles the HTTP request to delete a single {{.Web.R.SingularLower}} specified by URI parameters.
func (h *Handler) Delete{{.Web.R.SingularName}}(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.{{.Web.R.BusinessFactoryName}}().Delete, h.val.ValidateDelete{{.Web.R.SingularName}}Request)
}

// Delete{{.Web.R.PluralName}} handles the HTTP request to delete a collection of {{.Web.R.PluralLower}} specified in the body.
func (h *Handler) Delete{{.Web.R.PluralName}}(c *gin.Context) {
    core.HandleJSONRequest(c, h.biz.{{.Web.R.BusinessFactoryName}}().DeleteCollection, h.val.ValidateDelete{{.Web.R.PluralName}}Request)
}

// Get{{.Web.R.SingularName}} retrieves details of a specific {{.Web.R.SingularLower}} based on the request parameters.
func (h *Handler) Get{{.Web.R.SingularName}}(c *gin.Context) {
	{{- if .Web.WithOTel}}
    ctx, span := otel.Tracer("handler").Start( c.Request.Context(), "Handler.Get{{.Web.R.SingularName}}")
    defer span.End()

	c.Request = c.Request.WithContext(ctx)

	metrics.M.RecordResourceGet(ctx, "{{.Web.R.SingularLower}}")
    {{- end}}

	slog.InfoContext(ctx, "processing {{.Web.R.SingularLower}} retrieve request")

	core.HandleUriRequest(c, h.biz.{{.Web.R.BusinessFactoryName}}().Get, h.val.ValidateGet{{.Web.R.SingularName}}Request)
}

// List{{.Web.R.SingularName}} retrieves a list of {{.Web.R.PluralLower}} based on query parameters.
func (h *Handler) List{{.Web.R.SingularName}}(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.{{.Web.R.BusinessFactoryName}}().List, h.val.ValidateList{{.Web.R.SingularName}}Request)
}

func init() {
	Register(func(v1 *gin.RouterGroup, handler *Handler, mws ...gin.HandlerFunc) {
		{{- if ne .Web.R.ResourcePathPrefix "" }}
		rg := v1.Group("/{{.Web.R.ResourcePathPrefix}}/{{.Web.R.Last.PluralLower}}", mws...)
		{{- else}}
		rg := v1.Group("/{{.Web.R.Last.PluralLower}}", mws...)
		{{- end}}
		rg.POST("", handler.Create{{.Web.R.SingularName}})
		rg.PUT(":{{.Web.R.Last.SingularLowerFirst}}ID", handler.Update{{.Web.R.SingularName}})
		rg.DELETE(":{{.Web.R.Last.SingularLowerFirst}}ID", handler.Delete{{.Web.R.SingularName}})
		rg.DELETE("", handler.Delete{{.Web.R.PluralName}})
		rg.GET(":{{.Web.R.Last.SingularLowerFirst}}ID", handler.Get{{.Web.R.SingularName}})
		rg.GET("", handler.List{{.Web.R.SingularName}})
	})
}
