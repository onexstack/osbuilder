//go:build wireinject
// +build wireinject

package {{.Web.Name}}

import (
    "context"

	"github.com/google/wire"
    {{- if .Web.WithUser}}
    "github.com/onexstack/onexstack/pkg/authz"
    {{- end}}

	"{{.D.ModuleName}}/internal/{{.Web.Name}}/biz"
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/handler"
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/validation"
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/store"
    {{- if .Web.Clients }}
    "{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/clientset"
    {{- end}}
    {{- if .Web.WithUser}}
    {{- if eq .Web.WebFramework "gin" }}
    mw "{{.D.ModuleName}}/internal/pkg/middleware/gin"
    {{- else if eq .Web.WebFramework "grpc"}}
    mw "{{.D.ModuleName}}/internal/pkg/middleware/grpc"
    {{- end}}
    {{- end}}
)

// infrastructureSet groups all infrastructure-related providers.
// This keeps the main wire.Build call clean.
var infrastructureSet = wire.NewSet(
    ProvideDB,
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
    {{- if .Web.WithPreloader }}
    ProvideAStore,
    {{- end}}
    clientset.New,
    wire.Bind(new(clientset.Interface), new(*clientset.Clientset)),
    {{- end}}
)

// NewServer initializes and creates the web server with all necessary dependencies using Wire.
func NewServer(context.Context, *Config) (*Server, error) {
    wire.Build(
        // Server infrastructure
        NewWebServer,
        NewDependencies,
        wire.Struct(new(ServerConfig), "*"), // Inject all fields
        wire.Struct(new(Server), "*"),

        // Domain layers
        store.ProviderSet,
        biz.ProviderSet,
        validation.ProviderSet,
        handler.NewHandler,

        // Infrastructure dependencies
        infrastructureSet,
    )
    return nil, nil
}
