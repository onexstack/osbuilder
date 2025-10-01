package types

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/duke-git/lancet/v2/strutil"

	"github.com/onexstack/osbuilder/internal/osbuilder/helper"
	"github.com/onexstack/osbuilder/internal/osbuilder/known"
)

// REST captures naming conventions and helper metadata for generating REST resources.
type REST struct {
	// Singular form of the kind (e.g., "CronJob").
	SingularName string
	// Plural form of the kind (e.g., "CronJobs").
	PluralName string
	// Singular name in lower format (e.g., "cronjob").
	SingularLower string
	// Plural name in lower format (e.g., "cronjobs").
	PluralLower string
	// Singular name in lowerCamel (first letter lower) (e.g., "cronJob").
	SingularLowerFirst string
	// Plural name in lowerCamel (first letter lower) (e.g., "cronJobs").
	PluralLowerFirst string

	// Name of the associated GORM model (e.g., "CronJobModel").
	GORMModel string
	// Function name to map the model to the API.
	MapModelToAPIFunc string
	// Function name to map the API to the model.
	MapAPIToModelFunc string
	// Name of the business layer factory.
	BusinessFactoryName string

	// Name of the generated Go file.
	FileName string
}

// WebServer describes a web server component to generate (HTTP/gRPC/etc).
type WebServer struct {
	// BinaryName is the CLI binary name (e.g., "mb-apiserver").
	BinaryName string `yaml:"binaryName"`
	// WebFramework selects the framework (e.g., gin, grpc).
	WebFramework string `yaml:"webFramework"`
	// GRPCServiceName is the gRPC service name; default: UpperFirst(component name).
	GRPCServiceName string `yaml:"grpcServiceName,omitempty"`
	// StorageType selects backing storage (e.g., memory, mysql).
	StorageType string `yaml:"storageType"`
	// Feature flags
	WithHealthz bool `yaml:"withHealthz,omitempty"`
	WithUser    bool `yaml:"withUser,omitempty"`

	// Computed/derived fields (not serialized).
	Proj              *Project `yaml:"-"`
	Name              string   `yaml:"-"`
	EnvironmentPrefix string   `yaml:"-"`
	// APIImportPath is like: v1 "module/pkg/api/apiserver/v1"
	APIImportPath string `yaml:"-"`
	R             *REST  `yaml:"-"`
}

// Complete populates derived fields and sensible defaults.
func (ws *WebServer) Complete(proj *Project) *WebServer {
	ws.Proj = proj
	ws.Name = helper.GetComponentName(ws.BinaryName)

	// Default gRPC service name to UpperFirst(ComponentName), e.g., "Apiserver".
	if strings.TrimSpace(ws.GRPCServiceName) == "" {
		ws.GRPCServiceName = strutil.UpperFirst(ws.Name)
	}

	// Environment variable prefix: PROJECT_COMPONENT (uppercased).
	ws.EnvironmentPrefix = fmt.Sprintf("%s_%s",
		strings.ToUpper(proj.D.ProjectName),
		strings.ToUpper(ws.Name),
	)

	// Import alias path for API package.
	ws.APIImportPath = fmt.Sprintf(`%s "%s/pkg/api/%s/%s"`,
		ws.Proj.D.APIAlias,
		ws.Proj.D.ModuleName,
		ws.Name,
		ws.Proj.D.APIVersion,
	)

	return ws
}

// Base returns the component base directory: internal/<component>.
func (ws *WebServer) Base() string {
	return filepath.Join("internal", ws.Name)
}

// Model returns the model directory for the component.
func (ws *WebServer) Model() string {
	return filepath.Join(ws.Base(), "model")
}

// Pkg returns the pkg directory for the component.
func (ws *WebServer) Pkg() string {
	return filepath.Join(ws.Base(), "pkg")
}

// Handler returns the handler directory for the component.
func (ws *WebServer) Handler() string {
	return filepath.Join(ws.Base(), "handler")
}

// Biz returns the business logic directory for the component.
func (ws *WebServer) Biz() string {
	return filepath.Join(ws.Base(), "biz")
}

// Store returns the data store directory for the component.
func (ws *WebServer) Store() string {
	return filepath.Join(ws.Base(), "store")
}

// RESTBiz returns the path to a REST biz implementation for a singular resource.
func (ws *WebServer) RESTBiz(singularLower string) string {
	return filepath.Join(ws.Biz(), ws.Proj.D.APIVersion, singularLower, singularLower+".go")
}

// RESTStore returns the path to a REST store implementation for a singular resource.
func (ws *WebServer) RESTStore(singularLower string) string {
	return filepath.Join(ws.Store(), singularLower+".go")
}

// API returns the API directory for the component: pkg/api/<component>/<version>.
func (ws *WebServer) API() string {
	return filepath.Join("pkg/api", ws.Name, ws.Proj.D.APIVersion)
}

// SetREST attaches REST metadata for later template rendering.
func (ws *WebServer) SetREST(meta *REST) *WebServer {
	ws.R = meta
	return ws
}

// Pairs returns a map of destination relative paths to template paths.
// It drives file generation for this component.
func (ws *WebServer) Pairs() map[string]string {
	// Local shortcuts to reduce repetition.
	apiDir := filepath.Join("pkg/api", ws.Name, ws.Proj.D.APIVersion)
	internalPkg := ws.Proj.InternalPkg()
	baseDir := ws.Base()
	handlerDir := ws.Handler()
	storeDir := ws.Store()
	bizDir := ws.Biz()

	pairs := map[string]string{}
	add := func(dst, tpl string) {
		pairs[dst] = tpl
	}

	// Common command and component scaffolding.
	add(filepath.Join("cmd", ws.BinaryName, "app/options/options.go"), "/project/cmd/mb-apiserver/app/options/options.go")
	add(filepath.Join("cmd", ws.BinaryName, "app/server.go"), "/project/cmd/mb-apiserver/app/server.go")
	add(filepath.Join("cmd", ws.BinaryName, "main.go"), "/project/cmd/mb-apiserver/main.go")

	// Core internal packages reused across frameworks.
	add(filepath.Join(storeDir, "doc.go"), "/project/internal/apiserver/store/doc.go")
	add(filepath.Join(storeDir, "store.go"), "/project/internal/apiserver/store/store.go")
	add(filepath.Join(storeDir, "README.md"), "/project/internal/apiserver/store/README.md")

	add(filepath.Join(bizDir, "biz.go"), "/project/internal/apiserver/biz/biz.go")
	add(filepath.Join(bizDir, "doc.go"), "/project/internal/apiserver/biz/doc.go")
	add(filepath.Join(bizDir, "README.md"), "/project/internal/apiserver/biz/README.md")

	add(filepath.Join(ws.Pkg(), "validation/validation.go"), "/project/internal/apiserver/pkg/validation/validation.go")

	add(filepath.Join(internalPkg, "contextx/contextx.go"), "/project/internal/pkg/contextx/contextx.go")
	add(filepath.Join(internalPkg, "contextx/doc.go"), "/project/internal/pkg/contextx/doc.go")
	add(filepath.Join(internalPkg, "known/doc.go"), "/project/internal/pkg/known/doc.go")
	add(filepath.Join(internalPkg, "known/known.go"), "/project/internal/pkg/known/known.go")

	add(filepath.Join(internalPkg, "rid/doc.go"), "/project/internal/pkg/rid/doc.go")
	add(filepath.Join(internalPkg, "rid/example_test.go"), "/project/internal/pkg/rid/example_test.go")
	add(filepath.Join(internalPkg, "rid/rid.go"), "/project/internal/pkg/rid/rid.go")
	add(filepath.Join(internalPkg, "rid/rid_test.go"), "/project/internal/pkg/rid/rid_test.go")
	add(filepath.Join(internalPkg, "rid/salt.go"), "/project/internal/pkg/rid/salt.go")

	add(filepath.Join(internalPkg, "errno/doc.go"), "/project/internal/pkg/errno/doc.go")
	add(filepath.Join(internalPkg, "errno/code.go"), "/project/internal/pkg/errno/code.go")
	add(filepath.Join(internalPkg, "errno/post.go"), "/project/internal/pkg/errno/post.go")
	add(filepath.Join(internalPkg, "errno/user.go"), "/project/internal/pkg/errno/user.go")

	add(filepath.Join(baseDir, "server.go"), "/project/internal/apiserver/server.go")
	add(filepath.Join(baseDir, "wire.go"), "/project/internal/apiserver/wire.go")
	add(filepath.Join(baseDir, "wire_gen.go"), "/project/internal/apiserver/wire_gen.go")

	// Default proto for examples.
	add(filepath.Join(apiDir, "example.proto"), "/project/pkg/api/apiserver/v1/example.proto")

	// Optional 'user' feature.
	if ws.WithUser {
		add(filepath.Join(apiDir, "user.proto"), "/project/pkg/api/apiserver/v1/user.proto")
		add(filepath.Join(internalPkg, "known/role.go"), "/project/internal/pkg/known/role.go")

		// Model
		add(filepath.Join(ws.Model(), "user.gen.go"), "/project/internal/apiserver/model/user.gen.go")
		add(filepath.Join(ws.Model(), "hook_user.go"), "/project/internal/apiserver/model/hook_user.go")

		// Handler + middlewares by framework
		switch ws.WebFramework {
		case known.WebFrameworkGin:
			add(filepath.Join(handlerDir, "user.go"), "/project/internal/apiserver/handler/gin/user.go")
			add(filepath.Join(internalPkg, "middleware/gin/authn.go"), "/project/internal/pkg/middleware/gin/authn.go")
			add(filepath.Join(internalPkg, "middleware/gin/authz.go"), "/project/internal/pkg/middleware/gin/authz.go")
		case known.WebFrameworkGRPC:
			add(filepath.Join(handlerDir, "user.go"), "/project/internal/apiserver/handler/grpc/user.go")
			add(filepath.Join(internalPkg, "middleware/grpc/authn.go"), "/project/internal/pkg/middleware/grpc/authn.go")
			add(filepath.Join(internalPkg, "middleware/grpc/authz.go"), "/project/internal/pkg/middleware/grpc/authz.go")
			add(filepath.Join("examples/client/user/main.go"), "/project/examples/client/user/main.go")
			add(filepath.Join("examples/helper/helper.go"), "/project/examples/helper/helper.go")
			add(filepath.Join("examples/helper/README.md"), "/project/examples/helper/README.md")
		}

		// Conversion/validation
		add(filepath.Join(ws.Pkg(), "conversion/user.go"), "/project/internal/apiserver/pkg/conversion/user.go")
		add(filepath.Join(ws.Pkg(), "validation/user.go"), "/project/internal/apiserver/pkg/validation/user.go")

		// Biz + store
		add(ws.RESTBiz("user"), "/project/internal/apiserver/biz/v1/user/user.go")
		add(ws.RESTStore("user"), "/project/internal/apiserver/store/user.go")
	}

	// Optional healthz endpoints.
	if ws.WithHealthz {
		add(filepath.Join(apiDir, "healthz.proto"), "/project/pkg/api/apiserver/v1/healthz.proto")

		switch ws.WebFramework {
		case known.WebFrameworkGin:
			add(filepath.Join(handlerDir, "healthz.go"), "/project/internal/apiserver/handler/gin/healthz.go")
		case known.WebFrameworkGRPC:
			add(filepath.Join(handlerDir, "healthz.go"), "/project/internal/apiserver/handler/grpc/healthz.go")
			add(filepath.Join("examples/client/health/main.go"), "/project/examples/client/health/main.go")
		}
	}

	// Framework-specific scaffolding.
	switch ws.WebFramework {
	case known.WebFrameworkGin:
		add(filepath.Join(internalPkg, "middleware/gin/header.go"), "/project/internal/pkg/middleware/gin/header.go")
		add(filepath.Join(internalPkg, "middleware/gin/requestid.go"), "/project/internal/pkg/middleware/gin/requestid.go")
		add(filepath.Join(baseDir, "httpserver.go"), "/project/internal/apiserver/ginserver.go")
		add(filepath.Join(handlerDir, "handler.go"), "/project/internal/apiserver/handler/gin/handler.go")

	case known.WebFrameworkGRPC:
		// grpc middlewares
		add(filepath.Join(internalPkg, "middleware/grpc/requestid.go"), "/project/internal/pkg/middleware/grpc/requestid.go")
		add(filepath.Join(internalPkg, "middleware/grpc/doc.go"), "/project/internal/pkg/middleware/grpc/doc.go")
		add(filepath.Join(internalPkg, "middleware/grpc/defaulter.go"), "/project/internal/pkg/middleware/grpc/defaulter.go")
		add(filepath.Join(internalPkg, "middleware/grpc/validator.go"), "/project/internal/pkg/middleware/grpc/validator.go")

		// apiserver proto and servers
		add(filepath.Join(apiDir, ws.Name+".proto"), "/project/pkg/api/apiserver/v1/apiserver.proto")
		add(filepath.Join(baseDir, "grpcserver.go"), "/project/internal/apiserver/grpcserver.go")
		add(filepath.Join(handlerDir, "handler.go"), "/project/internal/apiserver/handler/grpc/handler.go")

	case known.WebFrameworkGRPCGateway:
		// TODO: add grpc-gateway templates if needed

	case known.WebFrameworkKratos:
		// TODO: add kratos templates if needed

	default:
		// Fallback to gRPC server scaffolding.
		add(filepath.Join(apiDir, ws.Name+".proto"), "/project/pkg/api/apiserver/v1/apiserver.proto")
		add(filepath.Join(baseDir, "grpcserver.go"), "/project/internal/apiserver/grpcserver.go")
		add(filepath.Join(handlerDir, "handler.go"), "/project/internal/apiserver/handler/grpc/handler.go")
	}

	// Ensure api dir exists in VCS.
	add(filepath.Join("api/.keep"), "/keep.tpl")

	return pairs
}

// TemplateData is the rendering context for templates (project + one component).
type TemplateData struct {
	*Project
	Web *WebServer
}
