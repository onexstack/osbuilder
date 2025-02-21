package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/onexstack/onexstack/pkg/core"
)

// Create{{.Web.R.SingularName}} handles the creation of a new {{.Web.R.SingularLower}}.
func (h *Handler) Create{{.Web.R.SingularName}}(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.{{.Web.R.BusinessFactoryName}}().Create, h.val.ValidateCreate{{.Web.R.SingularName}}Request)
}

// Update{{.Web.R.SingularName}} handles updating an existing {{.Web.R.SingularLower}}'s details.
func (h *Handler) Update{{.Web.R.SingularName}}(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.{{.Web.R.BusinessFactoryName}}().Update, h.val.ValidateUpdate{{.Web.R.SingularName}}Request)
}

// Delete{{.Web.R.SingularName}} handles the deletion of one or more {{.Web.R.PluralLower}}.
func (h *Handler) Delete{{.Web.R.SingularName}}(c *gin.Context) {
	core.HandleJSONRequest(c, h.biz.{{.Web.R.BusinessFactoryName}}().Delete, h.val.ValidateDelete{{.Web.R.SingularName}}Request)
}

// Get{{.Web.R.SingularName}} retrieves information about a specific {{.Web.R.SingularLower}}.
func (h *Handler) Get{{.Web.R.SingularName}}(c *gin.Context) {
	core.HandleUriRequest(c, h.biz.{{.Web.R.BusinessFactoryName}}().Get, h.val.ValidateGet{{.Web.R.SingularName}}Request)
}

// List{{.Web.R.SingularName}} retrieves a list of {{.Web.R.PluralLower}} based on query parameters.
func (h *Handler) List{{.Web.R.SingularName}}(c *gin.Context) {
	core.HandleQueryRequest(c, h.biz.{{.Web.R.BusinessFactoryName}}().List, h.val.ValidateList{{.Web.R.SingularName}}Request)
}

func init() {
	Register(func(v1 *gin.RouterGroup, handler *Handler) {
		rg := v1.Group("/{{.Web.R.PluralLower}}", handler.mws...)
		rg.POST("", handler.Create{{.Web.R.SingularName}})
		rg.PUT(":{{.Web.R.SingularLowerFirst}}ID", handler.Update{{.Web.R.SingularName}})
		rg.DELETE("", handler.Delete{{.Web.R.SingularName}})
		rg.GET(":{{.Web.R.SingularLowerFirst}}ID", handler.Get{{.Web.R.SingularName}})
		rg.GET("", handler.List{{.Web.R.SingularName}})
	})
}
