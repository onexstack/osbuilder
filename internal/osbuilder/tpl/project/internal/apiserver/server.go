package {{.Web.Name}}

import (
	"context"
	"log/slog"
    "time"

    {{- if eq .Web.StorageType "memory" }}
	"github.com/onexstack/onexstack/pkg/db"
	{{- end}}
	genericoptions "github.com/onexstack/onexstack/pkg/options"
	"github.com/onexstack/onexstack/pkg/server"
	"github.com/onexstack/onexstack/pkg/store/registry"
	"gorm.io/gorm"
    {{- if .Web.WithUser}}
	"github.com/onexstack/onexstack/pkg/authz"
	"github.com/onexstack/onexstack/pkg/store/where"
	"github.com/onexstack/onexstack/pkg/token"
	{{- end}}

	"{{.D.ModuleName}}/internal/{{.Web.Name}}/handler"
    {{- if .Web.WithPreloader}}
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/asyncstore"
	{{- end}}
	{{- range .Web.Clients }}
	"{{$.D.ModuleName}}/internal/{{$.Web.Name}}/pkg/clientset/typed/{{. | lowerkind}}"
	{{- end}}
    {{- if .Web.WithUser}}
	"{{.D.ModuleName}}/internal/pkg/contextx"
	"{{.D.ModuleName}}/internal/pkg/known"
	{{- if eq .Web.WebFramework "gin" }}
	mw "{{.D.ModuleName}}/internal/pkg/middleware/gin"
	{{- else if eq .Web.WebFramework "grpc"}}
	mw "{{.D.ModuleName}}/internal/pkg/middleware/grpc"
	{{- end}}
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/store"
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/model"
	{{- end}}
	{{- if .Web.WithOTel}}
    "{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/metrics"
    {{- end}}
)

const serviceName = "{{.Web.BinaryName}}"

// Dependencies collects all components that need initialization but are not directly used
// by the main server struct during runtime (e.g., sidecar processes, cache warmers).
type Dependencies struct{}

// Config contains application-related configurations.
type Config struct {
    {{- if .Web.WithUser}}
    JWTKey            string
    Expiration        time.Duration
    {{- end}}
	TLSOptions        *genericoptions.TLSOptions
	{{- if or (eq .Web.WebFramework "gin") (eq .Web.WebFramework "grpc-gateway")}}
	HTTPOptions       *genericoptions.HTTPOptions
	{{- end}}
	{{- if or (eq .Web.WebFramework "grpc") (eq .Web.WebFramework "grpc-gateway")}}
	GRPCOptions       *genericoptions.GRPCOptions
    {{- if .Web.WithOTel}}
	MetricsAddr string
    {{- end}}
	{{- end}}
	{{- if eq .Web.StorageType "mariadb" }}
	MySQLOptions      *genericoptions.MySQLOptions
	{{- end}}
	{{- if eq .Web.StorageType "postgresql" }}
	PostgreSQLOptions *genericoptions.PostgreSQLOptions
	{{- end}}
	{{- if eq .Web.StorageType "sqlite" }}
	SQLiteOptions *genericoptions.SQLiteOptions
	{{- end}}
	{{- if eq .Web.ServiceRegistry "polaris" }}
    PolarisOptions *genericoptions.PolarisOptions
	{{- end}}
	{{- range .Web.Clients }}
	{{. | kind}}Options *genericoptions.RestyOptions	
	{{- end}}
}

// Server represents the web server and its background workers.
type Server struct {
    cfg         *ServerConfig
    srv         server.Server
}

// ServerConfig contains the core dependencies and configurations of the server.
type ServerConfig struct {
	*Config
    Dependencies *Dependencies
    Handler      *handler.Handler
    {{- if .Web.WithUser}}
    Retriever mw.UserRetriever
    Authz     *authz.Authz 
	{{- end}}
}

// New creates and returns a new Server instance.
func (cfg *Config) New(ctx context.Context) (*Server, error) {
    // Create the core server instance using dependency injection.
    // This relies on the wire-generated NewServer function.
    s, err := NewServer(ctx, cfg)
    if err != nil {
        return nil, err
    }

    return s.Prepare(ctx)
}

// Prepare performs post-initialization tasks such as registering subscribers.
func (s *Server) Prepare(ctx context.Context) (*Server, error) {
    {{- if .Web.WithUser}}
	where.RegisterTenant("user_id", func(ctx context.Context) string {
	    return contextx.UserID(ctx)
	})

    // 初始化 token 包的签名密钥、认证 Key 及 Token 默认过期时间
    token.Init(cfg.JWTKey, token.WithIdentityKey(known.XUserID), token.WithExpiration(cfg.Expiration), token.WithCommonSkipPaths())
	{{- end}}

	{{- if .Web.WithOTel}}
	metrics.Init(serviceName)
	{{- end}}
    return s, nil
}

// Run starts the server and listens for termination signals.
// It gracefully shuts down the server upon receiving a termination signal from the context.
func (s *Server) Run(ctx context.Context) error {
	// Start the HTTP/gRPC server in a background goroutine.
	go s.srv.RunOrDie(ctx)

	{{- if eq .Web.ServiceRegistry "polaris" }}
	if err := s.cfg.PolarisOptions.Register(); err != nil {
		slog.Error("polaris register failed", "error", err)
		return err
	}
	{{- end}}

	// Block until the context is canceled (e.g., via SIGINT/SIGTERM).
	<-ctx.Done()

	slog.Info("shutting down server...")

    // Create a new context with a timeout to ensure graceful shutdown doesn't hang indefinitely.
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

	{{- if eq .Web.ServiceRegistry "polaris" }}
	// Deregister from Polaris first (stop heartbeats)
	if err := s.cfg.PolarisOptions.Deregister(); err != nil {
		slog.Error("failed to deregister Polaris service", "error", err)
	}
	{{- end}}

    // Trigger graceful shutdown for all components.
	s.srv.GracefulStop(shutdownCtx)

	slog.Info("server exited successfully")

	return nil
}

// NewDB creates and returns a *gorm.DB instance for database operations.
func (cfg *Config) NewDB() (*gorm.DB, error) {
	slog.Info("initializing database connection", "type", "{{.Web.StorageType}}")

	{{- if eq .Web.StorageType "mariadb" }}
	dbInstance, err := cfg.MySQLOptions.NewDB()
	{{- end}}
	{{- if eq .Web.StorageType "postgresql" }}
	dbInstance, err := cfg.PostgreSQLOptions.NewDB()
	{{- end}}
	{{- if eq .Web.StorageType "sqlite" }}
	dbInstance, err := cfg.SQLiteOptions.NewDB()
	{{- end}}
	{{- if eq .Web.StorageType "memory" }}
	// TODO: Retrieve the database path from configuration instead of hardcoding.
	dbInstance, err := db.NewInMemorySQLite("/tmp/{{.Web.BinaryName}}.db")
	{{- end}}
	if err != nil {
		slog.Error("failed to create database connection", "error", err)
		return nil, err
	}

	// Automatically migrate database schema
	if err := registry.Migrate(dbInstance); err != nil {
		slog.Error("failed to migrate database schema", "error", err)
		return nil, err
	}

	return dbInstance, nil
}

{{- if .Web.WithUser}}
// UserRetriever defines a user data retriever. It is used to get user information.
type UserRetriever struct {
    store store.IStore
}

// GetUser retrieves user information by user ID.
func (r *UserRetriever) GetUser(ctx context.Context, userID string) (*model.UserM, error) {
    return r.store.User().Get(ctx, where.F("user_id", userID))
}
{{- end}}

// ProvideDB provides a database instance based on the configuration.
func ProvideDB(cfg *Config) (*gorm.DB, error) {
	return cfg.NewDB()
}

{{- range .Web.Clients }}
// Provide{{. | kind}}Client creates and returns a {{. | lowerkind}} client instance using the provided configuration.
func Provide{{. | kind}}Client(cfg *Config) {{. | lowerkind}}.Interface {
    return {{. | lowerkind}}.NewForConfig(cfg.{{. | kind}}Options)
}
{{- end}}

{{- if .Web.WithPreloader}}
// ProvideAStore creates and returns an asynchronous store factory.
func ProvideAStore(ctx context.Context) asyncstore.Factory {
	return asyncstore.NewStore(ctx, 30*time.Minute)
}
{{- end}}

// NewDependencies initializes all components that need to be started but are not directly stored.
// This is typically used for side-effects or warming up caches.
func NewDependencies(ctx context.Context{{- if .Web.WithPreloader }}, _ asyncstore.Factory{{- end -}}) *Dependencies {
	{{- if .Web.WithPreloader}}
	// Simulate cache warmup or check.
	fakeItem, _ := asyncstore.S.Fake().Get("fixed-item-001")
	slog.DebugContext(ctx, "successfully retrieved fake cache data", "data", fakeItem.String())
	{{- end}}

	return &Dependencies{}
}

// NewWebServer creates and returns a new web server instance using the provided server configuration.
func NewWebServer(serverConfig *ServerConfig) (server.Server, error) {
	{{- if or (eq .Web.WebFramework "grpc") (eq .Web.WebFramework "grpc-gateway")}}
	{{- if eq .Web.ServiceRegistry "polaris" }}
    return serverConfig.NewPolarisServer()
	{{- else}}
    return serverConfig.NewGRPCServer()
	{{- end}}
	{{else if eq .Web.WebFramework "gin"}}
    return serverConfig.NewGinServer()
	{{- end -}}
}
