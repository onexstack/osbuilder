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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
    {{- end}}

	{{- if .Web.WithOTel}}
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/pkg/metrics"
    {{- end}}
	"{{.D.ModuleName}}/internal/pkg/errno"
	mw "{{.D.ModuleName}}/internal/pkg/middleware/gin"
	"{{.D.ModuleName}}/internal/{{.Web.Name}}/handler"
)

// ginServer 定义一个使用 Gin 框架开发的 HTTP 服务器.
type ginServer struct {
	srv server.Server
}

// 确保 *ginServer 实现了 server.Server 接口.
var _ server.Server = (*ginServer)(nil)

func (c *ServerConfig) NewGinServer() (*ginServer, error) {
	// 创建 Gin 引擎
	engine := gin.New()

	// 注册全局中间件，用于恢复 panic、设置 HTTP 头、添加请求 ID 等
	engine.Use(
		gin.Recovery(), 
		mw.NoCache, 
		mw.Cors, 
		mw.Secure, 
		{{- if .Web.WithOTel}}
		otelgin.Middleware("{{.Web.BinaryName}}", otelgin.WithFilter(func(rq *http.Request) bool {
            // 返回 false 表示不创建 span（过滤掉）
            return rq.URL.Path != "/metrics"
        })),
		genericmw.Observability(),
		mw.Context(),
		{{- end}}
	)

	// 注册.R API 路由
	c.InstallRESTAPI(engine)

	httpsrv := server.NewHTTPServer(c.HTTPOptions, c.TLSOptions, engine)

	return &ginServer{srv: httpsrv}, nil
}

// 注册 API 路由。路由的路径和 HTTP 方法，严格遵循.R 规范.
func (c *ServerConfig) InstallRESTAPI(engine *gin.Engine) {
	// 注册业务无关的 API 接口
	InstallGenericAPI(engine)

	{{if .Web.WithUser}}
	// 认证和授权中间件
	authMiddlewares := []gin.HandlerFunc{mw.AuthnMiddleware(c.retriever), mw.AuthzMiddleware(c.authz)}

	// 创建核心业务处理器
	hdl := handler.NewHandler(c.biz, c.val, authMiddlewares...)
	{{- else}}
	hdl := handler.NewHandler(c.biz, c.val)
	{{- end}}

	{{- if .Web.WithHealthz}}
	// 注册健康检查接口
	engine.GET("/healthz", hdl.Healthz)
	{{- end}}

	{{- if .Web.WithUser}}
	// 注册用户登录和令牌刷新接口。这2个接口比较简单，所以没有 API 版本
	engine.POST("/login", hdl.Login)
	// 注意：认证中间件要在 hdl.RefreshToken 之前加载
	engine.PUT("/refresh-token", mw.AuthnMiddleware(c.retriever), hdl.RefreshToken)
	{{- end}}

	// 注册 {{.D.APIVersion}} 版本 API 路由分组
	{{.D.APIVersion}} := engine.Group("/{{.D.APIVersion}}")
	//注册资源路由
	hdl.InstallAll({{.D.APIVersion}})
}

// InstallGenericAPI 注册业务无关的路由，例如 pprof、404 处理等.
func InstallGenericAPI(engine *gin.Engine) {
	// 注册 pprof 路由
	pprof.Register(engine)

	{{if .Web.WithOTel}}
	_ = metrics.Initialize(context.Background(), "{{.Web.BinaryName}}")

    // 暴露 /metrics 端点
    _ = engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

    {{- end}}

	// 注册 404 路由处理
	engine.NoRoute(func(c *gin.Context) {
		core.WriteResponse(c, errno.ErrPageNotFound, nil)
	})
}

// RunOrDie 启动 Gin 服务器，出错则程序崩溃退出.
func (s *ginServer) RunOrDie() {
	s.srv.RunOrDie()
}

// GracefulStop 优雅停止服务器.
func (s *ginServer) GracefulStop(ctx context.Context) {
	s.srv.GracefulStop(ctx)
}
