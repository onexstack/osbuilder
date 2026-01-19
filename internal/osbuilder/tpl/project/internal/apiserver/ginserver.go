package {{.Web.Name}}

import (
	"context"
	{{- if .Web.WithOTel}}
	"net/http"
    {{- end}}

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/onexstack/onexstack/pkg/core"
	"github.com/onexstack/onexstack/pkg/server"
	{{- if .Web.WithOTel}}
	genericmw "github.com/onexstack/onexstack/pkg/middleware/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
    {{- end}}

	"{{.D.ModuleName}}/internal/pkg/errno"
	mw "{{.D.ModuleName}}/internal/pkg/middleware/gin"
)

// ginServer implements the server.Server interface using the Gin framework.
type ginServer struct {
    srv server.Server
}

// Ensure *ginServer implements the server.Server interface.
var _ server.Server = (*ginServer)(nil)

// NewGinServer initializes and returns a new HTTP server based on Gin.
// It returns the server.Server interface to abstract implementation details.
func (c *ServerConfig) NewGinServer() (server.Server, error) {
	engine := gin.New()

	engine.Use(
		gin.Recovery(), 
		mw.NoCache, 
		mw.Cors(mw.DefaultCorsConfig()),
		mw.Secure, 
		{{- if .Web.WithOTel}}
		otelgin.Middleware(serviceName, otelgin.WithFilter(shouldRecordTelemetry)),
		genericmw.Observability(),
		mw.Context(),
		{{- end}}
	)

	c.InstallRESTAPI(engine)

	httpsrv := server.NewHTTPServer(c.HTTPOptions, c.TLSOptions, engine)

	return &ginServer{srv: httpsrv}, nil
}

// InstallRESTAPI registers all RESTful API routes to the engine.
func (c *ServerConfig) InstallRESTAPI(engine *gin.Engine) {
	InstallGenericAPI(engine)

	{{- if .Web.WithHealthz}}
	engine.GET("/healthz", c.Handler.Healthz)
	{{- end}}

	{{.D.APIVersion}} := engine.Group("/{{.D.APIVersion}}")

	{{- if .Web.WithUser}}
	// Register user login and token refresh interfaces.
	// These two interfaces are relatively simple, so there is no API version.
	engine.POST("/login", c.Handler.Login)
	// Note: The authentication middleware should be loaded before c.Handler.RefreshToken.
	engine.PUT("/refresh-token", mw.AuthnMiddleware(c.Retriever), c.Handler.RefreshToken)
	
	authMiddlewares := []gin.HandlerFunc{mw.AuthnMiddleware(c.Retriever), mw.AuthzMiddleware(c.Authz)}
	c.Handler.ApplyTo({{.D.APIVersion}}, authMiddlewares...)
	{{- else}}
	c.Handler.ApplyTo({{.D.APIVersion}})
	{{- end}}
}

// InstallGenericAPI registers non-business logic routes such as pprof, metrics, and 404 handlers.
func InstallGenericAPI(engine *gin.Engine) {
	pprof.Register(engine)

    // Expose /metrics endpoint for Prometheus scraping.
    engine.GET("/metrics", gin.WrapH(promhttp.HandlerFor(
        prometheus.DefaultGatherer,
        promhttp.HandlerOpts{
            EnableOpenMetrics: true,
        },
    )))

	// Register 404 handler.
	engine.NoRoute(func(c *gin.Context) {
		core.WriteResponse(c, errno.ErrPageNotFound, nil)
	})
}

// RunOrDie starts the Gin server and panics if startup fails.
func (s *ginServer) RunOrDie(ctx context.Context) {
	s.srv.RunOrDie(ctx)
}

// GracefulStop gracefully shuts down the server.
func (s *ginServer) GracefulStop(ctx context.Context) {
	s.srv.GracefulStop(ctx)
}

{{if .Web.WithOTel}}
// shouldRecordTelemetry filters out paths that shouldn't generate traces.
func shouldRecordTelemetry(r *http.Request) bool {
    return r.URL.Path != "/metrics"
}
{{- end}}
