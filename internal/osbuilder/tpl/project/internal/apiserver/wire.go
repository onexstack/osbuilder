//go:build wireinject
// +build wireinject

package {{.Web.Name}}

import (
	"github.com/google/wire"
    {{- if .Web.WithUser}}
    "github.com/onexstack/onexstack/pkg/authz"
    {{- end}}

	"{{.D.ModuleName}}/internal/{{.Web.Name}}/biz"
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/validation"
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/store"
    {{- if .Web.WithUser}}
    {{- if eq .Web.WebFramework "gin" }}
    mw "{{.D.ModuleName}}/internal/pkg/middleware/gin"
    {{- else if eq .Web.WebFramework "grpc"}}
    mw "{{.D.ModuleName}}/internal/pkg/middleware/grpc"
    {{- end}}
    {{- end}}
    {{- if .Web.Clients }}
    "{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/clientset"
    {{- end}}
)

// NewServer sets up and create the web server with all necessary dependencies.
func NewServer(*Config) (*Server, error) {
    wire.Build(
		NewWebServer,
        wire.Struct(new(ServerConfig), "*"), // * 表示注入全部字段
        wire.Struct(new(Server), "*"),
        wire.NewSet(store.ProviderSet, biz.ProviderSet),
        ProvideDB, // 提供数据库实例
        validation.ProviderSet,
        {{- if .Web.WithUser}}
        wire.NewSet(
            wire.Struct(new(UserRetriever), "*"),
            wire.Bind(new(mw.UserRetriever), new(*UserRetriever)),
        ),
        authz.ProviderSet,
        {{- end}}
        {{- if .Web.Clients }}
        {{- range .Web.Clients }}
        Provide{{. | kind}}Client,
        {{- end}}
        clientset.New,
        wire.Bind(new(clientset.Interface), new(*clientset.Clientset)),
        {{- end}}
    )
    return nil, nil
}
