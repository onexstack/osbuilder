package ws

import (
	"context"
	"time"

    {{- if .Web.Clients }}
    "{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/clientset"
    {{- end}}
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/store"
	{{.Web.APIImportPath}}
)

type WSBiz interface {
	Ping(ctx context.Context, rq *{{.D.APIAlias}}.PingRequest) (*{{.D.APIAlias}}.PingResponse, error)
}

// wsBiz is the implementation of the WSBiz.
type wsBiz struct {
	store     store.IStore
    {{- if .Web.Clients }}
	clientset clientset.Interface
    {{- end}}
}

// Ensure that *wsBiz implements the WSBiz.
var _ WSBiz = (*wsBiz)(nil)

// New creates and returns a new instance of *wsBiz.
func New(store store.IStore{{- if .Web.Clients }}, clientset clientset.Interface{{- end -}}) *wsBiz {
	return &wsBiz{store: store{{- if .Web.Clients}}, clientset: clientset{{- end -}}}
}

// Create implements the Create method of the WSBiz.
func (b *wsBiz) Ping(ctx context.Context, rq *{{.D.APIAlias}}.PingRequest) (*{{.D.APIAlias}}.PingResponse, error) {
	startTime := time.Now()

	// 模拟处理逻辑
	time.Sleep(time.Millisecond) // 模拟处理延迟

	processingTime := time.Since(startTime).Microseconds()

	return &{{.D.APIAlias}}.PingResponse{Sequence: rq.Sequence, ProcessingTimeUS: processingTime}, nil
}
