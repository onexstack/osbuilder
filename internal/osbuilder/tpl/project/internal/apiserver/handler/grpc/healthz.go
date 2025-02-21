package handler

import (
	"context"
	"time"

	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"github.com/onexstack/onexstack/pkg/log"

    {{.Web.APIImportPath}}

)

// Healthz 服务健康检查.
func (h *Handler) Healthz(ctx context.Context, rq *emptypb.Empty) (*{{.D.APIAlias}}.HealthzResponse, error) {
	log.W(ctx).Infow("Healthz handler is called", "method", "Healthz", "status", "healthy")
	return &{{.D.APIAlias}}.HealthzResponse{
		Status:    {{.D.APIAlias}}.ServiceStatus_Healthy,
		Timestamp: time.Now().Format(time.DateTime),
	}, nil
}
