package handler

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"

	{{.Web.APIImportPath}}
)

// Healthz handles service health check requests.
func (h *Handler) Healthz(c *gin.Context) {
    slog.InfoContext(c.Request.Context(), "health check requested", "status", "healthy")
    core.WriteResponse(c, v1.HealthzResponse{
        Status:    v1.ServiceStatus_Healthy,
        Timestamp: time.Now().Format(time.RFC3339),
    }, nil)
}
