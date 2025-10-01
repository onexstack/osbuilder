package {{.Web.Name}}

import (
	"context"
    {{- if .Web.WithUser}}
    "time"
    {{- end}}

	"github.com/onexstack/onexstack/pkg/log"
	genericoptions "github.com/onexstack/onexstack/pkg/options"
	"github.com/onexstack/onexstack/pkg/server"
	"github.com/onexstack/onexstack/pkg/store/registry"
    {{- if eq .Web.StorageType "memory" }}
	"gorm.io/driver/sqlite"
	{{- end}}
    {{- if eq .Web.StorageType "mariadb" }}
	{{- end}}
	"gorm.io/gorm"
    {{- if .Web.WithUser}}
	"github.com/onexstack/onexstack/pkg/authz"
	"github.com/onexstack/onexstack/pkg/store/where"
	"github.com/onexstack/onexstack/pkg/token"
	{{- end}}

	"{{.D.ModuleName}}/internal/{{.Web.Name}}/biz"
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/validation"
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
)

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
	{{- end}}
	{{- if eq .Web.StorageType "mariadb" }}
	MySQLOptions      *genericoptions.MySQLOptions
	{{- end}}
}

// Server represents the web server.
type Server struct {
	srv server.Server
}

// ServerConfig contains the core dependencies and configurations of the server.
type ServerConfig struct {
    cfg       *Config
    biz       biz.IBiz
    val       *validation.Validator
    {{- if .Web.WithUser}}
    retriever mw.UserRetriever                
    authz     *authz.Authz 
	{{- end}}
}

// NewServer initializes and returns a new Server instance.
func (cfg *Config) NewServer(ctx context.Context) (*Server, error) {
    {{- if .Web.WithUser}}
	where.RegisterTenant("userID", func(ctx context.Context) string {
	    return contextx.UserID(ctx)
	})

    // 初始化 token 包的签名密钥、认证 Key 及 Token 默认过期时间
    token.Init(cfg.JWTKey, known.XUserID, cfg.Expiration)

	{{- end}}
	// Create the core server instance.
	srv, err := InitializeWebServer(cfg)
	if err != nil {
		return nil, err
	}

	return &Server{srv: srv}, nil
}

// Run starts the server and listens for termination signals.
// It gracefully shuts down the server upon receiving a termination signal.
func (s *Server) Run(ctx context.Context) error {
	return server.Serve(ctx, s.srv)
}

// NewDB creates and returns a *gorm.DB instance for MySQL.
func (cfg *Config) NewDB() (*gorm.DB, error) {
	log.Infow("Initializing database connection", "type", "{{.Web.StorageType}}")
	{{- if eq .Web.StorageType "mariadb" }}
	db, err := cfg.MySQLOptions.NewDB()
	if err != nil {
		log.Errorw(err, "Failed to create database connection")
		return nil, err
	}
	{{- else}}
	// Configure the database using SQLite memory mode
	// ?cache=shared is used to set SQLite's cache mode to shared cache mode.
	// By default, each SQLite database connection has its own private cache. This mode is called private cache.
	// Using shared cache mode allows different connections to share the same in-memory database and cache.
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		log.Errorw(err, "Failed to create database connection")
		return nil, err
	}
	{{- end}}

	// Automatically migrate database schema
	if err := registry.Migrate(db); err != nil {
		log.Errorw(err, "Failed to migrate database schema")
		return nil, err
	}

	return db, nil
}

{{- if .Web.WithUser}}
// UserRetriever 定义一个用户数据获取器. 用来获取用户信息.
type UserRetriever struct {
    store store.IStore
}
                                                          
// GetUser 根据用户 ID 获取用户信息.                        
func (r *UserRetriever) GetUser(ctx context.Context, userID string) (*model.UserM, error) {
    return r.store.User().Get(ctx, where.F("userID", userID))
}
{{- end}}

// ProvideDB provides a database instance based on the configuration.
func ProvideDB(cfg *Config) (*gorm.DB, error) {
	return cfg.NewDB()
}

func NewWebServer(serverConfig *ServerConfig) (server.Server, error) {
	{{- if or (eq .Web.WebFramework "grpc") (eq .Web.WebFramework "grpc-gateway")}}
    return serverConfig.NewGRPCServer()
	{{else if eq .Web.WebFramework "gin"}}
    return serverConfig.NewGinServer()
	{{- end -}}
}
