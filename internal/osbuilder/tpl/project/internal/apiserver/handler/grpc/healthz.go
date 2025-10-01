package handler

import (
	"context"
	"time"

	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"k8s.io/klog/v2"

    {{.Web.APIImportPath}}

)

// Healthz 服务健康检查.
func (h *Handler) Healthz(ctx context.Context, rq *emptypb.Empty) (*{{.D.APIAlias}}.HealthzResponse, error) {
	klog.FromContext(ctx).Info("Healthz handler is called", "method", "Healthz", "status", "healthy")
	return &{{.D.APIAlias}}.HealthzResponse{
		Status:    {{.D.APIAlias}}.ServiceStatus_Healthy,
		Timestamp: time.Now().Format(time.DateTime),
	}, nil
}
