package handler

import (
	"time"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"

	{{.Web.APIImportPath}}
)

// Healthz 服务健康检查.
func (h *Handler) Healthz(c *gin.Context) {
	slog.InfoContext(c.Request.Context(), "Healthz handler is called", "method", "Healthz", "status", "healthy")
	core.WriteResponse(c, {{.D.APIAlias}}.HealthzResponse{
		Status:    {{.D.APIAlias}}.ServiceStatus_Healthy,
		Timestamp: time.Now().Format(time.DateTime),
	}, nil)
}
