package {{.Web.Name}}

import (
	"context"
    {{- if .Web.WithOTel}}                                            
    "log/slog" 
    "net/http"            
    "os"                                                
    {{- end}}

    {{- if .Web.WithUser}}
	{{- if or (eq .Web.WebFramework "grpc") (eq .Web.WebFramework "grpc-gateway")}}
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	{{- end}}
	{{- end}}
	"github.com/onexstack/onexstack/pkg/server"
	genericvalidation "github.com/onexstack/onexstack/pkg/validation"
	"google.golang.org/grpc"
    {{- if .Web.WithOTel}}                
    "github.com/gin-gonic/gin"                         
    genericmw "github.com/onexstack/onexstack/pkg/middleware/grpc"
	"github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
    {{- end}}

    {{- if .Web.WithOTel}}        
    "{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/metrics"
    {{- end}}
	{{.D.APIAlias}} "{{.D.ModuleName}}/pkg/api/{{.Web.Name}}/{{.D.APIVersion}}"
	mw "{{.D.ModuleName}}/internal/pkg/middleware/grpc"
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/handler"
)

// polarisServer defines a polaris gRPC server.
type polarisServer struct {
	srv server.Server
    // stop is the graceful shutdown function.
    stop func(context.Context)
}

// Ensure that *polarisServer implements the server.Server interface.
var _ server.Server = (*polarisServer)(nil)

// NewPolarisServer creates and initializes a polaris gRPC server.
func (c *ServerConfig) NewPolarisServer() (*polarisServer, error) {
    {{if .Web.WithOTel}}
    _ = metrics.Initialize(context.Background(), "{{.Web.BinaryName}}")
 
    // Start Gin in a separate goroutine (Prometheus metrics endpoint)
    go func() {
        r := gin.Default()
        r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(
            prometheus.DefaultGatherer,                            
            promhttp.HandlerOpts{
                EnableOpenMetrics: true, // 开启 OpenMetrics 支持
            },
		)))
        // You can change this port if needed (e.g. ":9090")
        slog.Info("Start metrics server on %s", c.MetricsAddr)
        if err := r.Run(c.MetricsAddr); err != nil && err != http.ErrServerClosed {
            slog.Error("Failed to start metrics server", "error", err)
            os.Exit(1)
        }
    }()
    {{- end}}

	// Configure gRPC server options, including interceptor chains.
	serverOptions := []grpc.ServerOption{
        {{- if .Web.WithOTel}}            
        // Note the order of interceptors!             
        grpc.StatsHandler(otelgrpc.NewServerHandler()),
        {{- end}}                  
		// Note the order of interceptors!
		grpc.ChainUnaryInterceptor(
			// Request ID interceptor.
			mw.RequestIDInterceptor(),
            {{- if .Web.WithUser}}
            // 认证拦截器                     
            selector.UnaryServerInterceptor(mw.AuthnInterceptor(c.Retriever), NewAuthnWhiteListMatcher()),
            // 授权拦截器
            selector.UnaryServerInterceptor(mw.AuthzInterceptor(c.Authz), NewAuthzWhiteListMatcher()),
            {{- end}}
			// Default value setting interceptor.
			mw.DefaulterInterceptor(),
			//.D validation interceptor.
			mw.ValidatorInterceptor(genericvalidation.NewValidator(c.val)),
            {{- if .Web.WithOTel}}                                                                                                                             
            genericmw.Observability(),
            mw.Context(),
            {{- end}}
		),
	}

	// Create the polaris gRPC server.
	polarissrv, err := server.NewPolarisServer(
		c.PolarisOptions,
		c.GRPCOptions,
		c.TLSOptions,
		serverOptions,
	 	func() (func(s grpc.ServiceRegistrar), string) {
        	return func(s grpc.ServiceRegistrar) { 
				{{.D.APIAlias}}.Register{{.Web.GRPCServiceName}}Server(s, handler.NewHandler(c.biz)) 
			}, v1.{{.Web.GRPCServiceName}}_ServiceDesc.ServiceName
        },
	)
	if err != nil {
		return nil, err
	}

	{{- if eq .Web.WebFramework "grpc"}}
	return &polarisServer{
		srv: polarissrv,
		stop: func(ctx context.Context) {
			polarissrv.GracefulStop(ctx)
		},
	}, nil
	{{- else}}

	// Start the gRPC server first, as the HTTP server depends on the gRPC server.
	go polarissrv.RunOrDie()

	httpsrv, err := server.NewGRPCGatewayServer(
		c.HTTPOptions,
		c.GRPCOptions,
		c.TLSOptions,
		func(mux *runtime.ServeMux, conn *grpc.ClientConn) error {
			return {{.D.APIAlias}}.Register{{.Web.GRPCServiceName}}Handler(context.Background(), mux, conn)
		},
	)
	if err != nil {
		return nil, err
	}

	return &polarisServer{
		srv: httpsrv,
		stop: func(ctx context.Context) {
			polaris.GracefulStop(ctx)
			httpsrv.GracefulStop(ctx)
		},
	}, nil
	{{- end}}
}

// RunOrDie starts the gRPC server or HTTP reverse proxy server and exits on errors.
func (s *polarisServer) RunOrDie() {
	s.srv.RunOrDie()
}

// GracefulStop gracefully stops the HTTP and gRPC servers.
func (s *polarisServer) GracefulStop(ctx context.Context) {
	s.stop(ctx)
}

{{- if .Web.WithUser}}
// NewAuthnWhiteListMatcher creates an authentication whitelist matcher.
func NewAuthnWhiteListMatcher() selector.Matcher {
	whitelist := map[string]struct{}{
		{{- if .Web.WithHealthz}}
		{{.D.APIAlias}}.{{.Web.GRPCServiceName}}_Healthz_FullMethodName:    {},
		{{- end}}
		{{.D.APIAlias}}.{{.Web.GRPCServiceName}}_CreateUser_FullMethodName: {},
		{{.D.APIAlias}}.{{.Web.GRPCServiceName}}_Login_FullMethodName:      {},
	}
	return selector.MatchFunc(func(ctx context.Context, call interceptors.CallMeta) bool {
		_, ok := whitelist[call.FullMethod()]
		return !ok
	})
}

// NewAuthzWhiteListMatcher creates an authorization whitelist matcher.
func NewAuthzWhiteListMatcher() selector.Matcher {
	whitelist := map[string]struct{}{
		{{- if .Web.WithHealthz}}
		{{.D.APIAlias}}.{{.Web.GRPCServiceName}}_Healthz_FullMethodName:    {},
		{{- end}}
		{{.D.APIAlias}}.{{.Web.GRPCServiceName}}_CreateUser_FullMethodName: {},
		{{.D.APIAlias}}.{{.Web.GRPCServiceName}}_Login_FullMethodName:      {},
	}
	return selector.MatchFunc(func(ctx context.Context, call interceptors.CallMeta) bool {
		_, ok := whitelist[call.FullMethod()]
		return !ok
	})
}
{{- end}}
