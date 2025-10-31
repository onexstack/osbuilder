
package main

import (
    "context"
    "log"
    "net"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
    grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    // 假设的 proto 生成文件
    pb "your-project/proto/user"
)

// 方案 1: 使用单独的 HTTP 端口暴露 metrics（最常用）
func setupGRPCWithSeparateMetricsPort() {
    // 1. 创建 gRPC 服务器并添加 Prometheus 中间件
    server := grpc.NewServer(
        grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
        grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
    )

    // 2. 注册你的服务
    userService := &UserService{}
    pb.RegisterUserServiceServer(server, userService)

    // 3. 初始化 Prometheus metrics
    grpc_prometheus.Register(server)

    // 4. 启动 gRPC 服务器
    go func() {
        lis, err := net.Listen("tcp", ":8080")
        if err != nil {
            log.Fatalf("Failed to listen: %v", err)
        }
        log.Println("gRPC server listening on :8080")
        if err := server.Serve(lis); err != nil {
            log.Fatalf("Failed to serve: %v", err)
        }
    }()

    // 5. 启动 HTTP 服务器暴露 metrics
    http.Handle("/metrics", promhttp.Handler())
    http.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    }))

    log.Println("Metrics server listening on :9090")
    log.Fatal(http.ListenAndServe(":9090", nil))
}

// 方案 2: 使用 Gin 框架暴露 metrics 和其他 HTTP 接口
func setupGRPCWithGinMetrics() {
    // 1. 创建 gRPC 服务器
    server := grpc.NewServer(
        grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
        grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
    )

    userService := &UserService{}
    pb.RegisterUserServiceServer(server, userService)
    grpc_prometheus.Register(server)

    // 2. 启动 gRPC 服务器
    go func() {
        lis, err := net.Listen("tcp", ":8080")
        if err != nil {
            log.Fatalf("Failed to listen: %v", err)
        }
        log.Println("gRPC server listening on :8080")
        if err := server.Serve(lis); err != nil {
            log.Fatalf("Failed to serve: %v", err)
        }
    }()

    // 3. 创建 Gin HTTP 服务器
    r := gin.New()
    r.Use(gin.Recovery())

    // 4. 暴露 metrics 接口
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))

    // 5. 添加健康检查和其他接口
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "healthy"})
    })

    r.GET("/ready", func(c *gin.Context) {
        // 可以添加就绪检查逻辑
        c.JSON(200, gin.H{"status": "ready"})
    })

    // 6. 添加调试接口
    r.GET("/debug/grpc", func(c *gin.Context) {
        // 返回 gRPC 服务信息
        c.JSON(200, gin.H{
            "grpc_address": ":8080",
            "services":     []string{"UserService"},
        })
    })

    log.Println("HTTP server listening on :9090")
    log.Fatal(r.Run(":9090"))
}

// 方案 3: 自定义 metrics 收集器
type CustomMetrics struct {
    requestsTotal   *prometheus.CounterVec
    requestDuration *prometheus.HistogramVec
    activeConns     prometheus.Gauge
}

func NewCustomMetrics() *CustomMetrics {
    return &CustomMetrics{
        requestsTotal: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "grpc_requests_total",
                Help: "Total number of gRPC requests",
            },
            []string{"method", "status"},
        ),
        requestDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "grpc_request_duration_seconds",
                Help:    "Duration of gRPC requests",
                Buckets: prometheus.DefBuckets,
            },
            []string{"method"},
        ),
        activeConns: prometheus.NewGauge(
            prometheus.GaugeOpts{
                Name: "grpc_active_connections",
                Help: "Number of active gRPC connections",
            },
        ),
    }
}

func (m *CustomMetrics) Register() {
    prometheus.MustRegister(m.requestsTotal)
    prometheus.MustRegister(m.requestDuration)
    prometheus.MustRegister(m.activeConns)
}

func (m *CustomMetrics) UnaryInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        m.activeConns.Inc()
        defer func() {
            m.activeConns.Dec()
            duration := time.Since(start).Seconds()
            m.requestDuration.WithLabelValues(info.FullMethod).Observe(duration)
        }()

        resp, err := handler(ctx, req)

        status := "success"
        if err != nil {
            status = "error"
        }
        m.requestsTotal.WithLabelValues(info.FullMethod, status).Inc()

        return resp, err
    }
}

func setupGRPCWithCustomMetrics() {
    // 1. 初始化自定义 metrics
    metrics := NewCustomMetrics()
    metrics.Register()

    // 2. 创建 gRPC 服务器，同时使用内置和自定义 metrics
    server := grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            metrics.UnaryInterceptor(),
            grpc_prometheus.UnaryServerInterceptor,
            recovery.UnaryServerInterceptor(),
        ),
    )

    userService := &UserService{}
    pb.RegisterUserServiceServer(server, userService)
    grpc_prometheus.Register(server)

    // 3. 启动 gRPC 服务器
    go func() {
        lis, err := net.Listen("tcp", ":8080")
        if err != nil {
            log.Fatalf("Failed to listen: %v", err)
        }
        log.Println("gRPC server listening on :8080")
        if err := server.Serve(lis); err != nil {
            log.Fatalf("Failed to serve: %v", err)
        }
    }()

    // 4. 启动 metrics 服务器
    http.Handle("/metrics", promhttp.Handler())
    log.Println("Metrics server listening on :9090")
    log.Fatal(http.ListenAndServe(":9090", nil))
}

// 方案 4: 完整的生产级配置
func setupProductionGRPCWithMetrics() {
    // 1. 创建自定义 registry（可选，用于隔离 metrics）
    registry := prometheus.NewRegistry()

    // 2. 创建自定义 metrics 收集器
    grpcMetrics := grpc_prometheus.NewServerMetrics()
    registry.MustRegister(grpcMetrics)

    // 3. 添加自定义业务 metrics
    businessMetrics := prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "business_operations_total",
            Help: "Total number of business operations",
        },
        []string{"operation", "result"},
    )
    registry.MustRegister(businessMetrics)

    // 4. 创建 gRPC 服务器
    server := grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            grpcMetrics.UnaryServerInterceptor(),
            recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(func(p interface{}) (err error) {
                log.Printf("Recovered from panic: %v", p)
                return status.Errorf(codes.Internal, "internal server error")
            })),
        ),
        grpc.ChainStreamInterceptor(
            grpcMetrics.StreamServerInterceptor(),
        ),
    )

    // 5. 注册服务
    userService := &UserServiceWithMetrics{businessMetrics: businessMetrics}
    pb.RegisterUserServiceServer(server, userService)
    grpcMetrics.InitializeMetrics(server)

    // 6. 启动 gRPC 服务器
    go func() {
        lis, err := net.Listen("tcp", ":8080")
        if err != nil {
            log.Fatalf("Failed to listen: %v", err)
        }
        log.Println("gRPC server listening on :8080")
        if err := server.Serve(lis); err != nil {
            log.Fatalf("Failed to serve: %v", err)
        }
    }()

    // 7. 创建 HTTP 服务器
    mux := http.NewServeMux()

    // Metrics 接口
    mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

    // 健康检查接口
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
    })

    // 就绪检查接口
    mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
        // 这里可以添加依赖检查逻辑
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ready"}`))
    })

    // 服务信息接口
    mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{
            "service": "user-service",
            "version": "1.0.0",
            "grpc_port": 8080,
            "http_port": 9090
        }`))
    })

    log.Println("HTTP server listening on :9090")
    log.Fatal(http.ListenAndServe(":9090", mux))
}

// 示例服务实现
type UserService struct {
    pb.UnimplementedUserServiceServer
}

func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    // 模拟业务逻辑
    return &pb.GetUserResponse{
        User: &pb.User{
            Id:   req.Id,
            Name: "John Doe",
            Email: "john@example.com",
        },
    }, nil
}

// 带业务 metrics 的服务实现
type UserServiceWithMetrics struct {
    pb.UnimplementedUserServiceServer
    businessMetrics *prometheus.CounterVec
}

func (s *UserServiceWithMetrics) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    // 记录业务指标
    defer func() {
        s.businessMetrics.WithLabelValues("get_user", "success").Inc()
    }()

    // 模拟业务逻辑
    if req.Id <= 0 {
        s.businessMetrics.WithLabelValues("get_user", "invalid_input").Inc()
        return nil, status.Error(codes.InvalidArgument, "invalid user id")
    }

    return &pb.GetUserResponse{
        User: &pb.User{
            Id:    req.Id,
            Name:  "John Doe",
            Email: "john@example.com",
        },
    }, nil
}

// main 函数示例
func main() {
    // 选择一种方案运行
    // setupGRPCWithSeparateMetricsPort()
    // setupGRPCWithGinMetrics()
    // setupGRPCWithCustomMetrics()
    setupProductionGRPCWithMetrics()
}

