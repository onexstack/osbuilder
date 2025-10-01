package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
	"github.com/onexstack/onexstack/pkg/core"

	{{.Web.APIImportPath}}
)

// Healthz 服务健康检查.
func (h *Handler) Healthz(c *gin.Context) {
	klog.FromContext(c.Request.Context()).Info("Healthz handler is called", "method", "Healthz", "status", "healthy")
	core.WriteResponse(c, {{.D.APIAlias}}.HealthzResponse{
		Status:    {{.D.APIAlias}}.ServiceStatus_Healthy,
		Timestamp: time.Now().Format(time.DateTime),
	}, nil)
}
