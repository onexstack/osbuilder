package ws

import (
	"context"
	"log/slog"
	"time"

    {{- if .Web.Clients }}
    "{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/clientset"
    {{- end}}
    "{{.D.ModuleName}}/internal/{{.Web.Name}}/store"
    {{.Web.APIImportPath}}
)

// WSBiz defines the interface for WebSocket-related business logic.
type WSBiz interface {
	// Ping processes the heartbeat request and measures the latency.
	Ping(ctx context.Context, rq *{{.D.APIAlias}}.PingRequest) (*{{.D.APIAlias}}.PingResponse, error)
}

// wsBiz implements the WSBiz interface.
type wsBiz struct {
	store     store.IStore
    {{- if .Web.Clients }}
    clientset clientset.Interface
    {{- end}}
}

// Ensure wsBiz implements WSBiz at compile time.
var _ WSBiz = (*wsBiz)(nil)

// New creates and returns a new instance of wsBiz.
func New(s store.IStore{{- if .Web.Clients }}, c clientset.Interface{{- end -}}) *wsBiz {
    return &wsBiz{store: s{{- if .Web.Clients}}, clientset: c{{- end -}}}
}

// Ping handles the heartbeat request.
func (b *wsBiz) Ping(ctx context.Context, rq *{{.D.APIAlias}}.PingRequest) (*{{.D.APIAlias}}.PingResponse, error) {
	start := time.Now()

	// Simulate processing logic delay.
	// In a real scenario, this might involve database checks or external calls.
	time.Sleep(time.Millisecond)

	elapsed := time.Since(start).Microseconds()

	slog.InfoContext(ctx, "ping processed", "sequence", rq.Sequence, "elapsed_us", elapsed)

	return &{{.D.APIAlias}}.PingResponse{Sequence: rq.Sequence, ProcessingTimeUS: elapsed}, nil
}
