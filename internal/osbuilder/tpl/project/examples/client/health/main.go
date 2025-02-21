package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	{{.Web.APIImportPath}}
)

var (
	// 定义命令行参数
	addr  = flag.String("addr", "localhost:6666", "The grpc server address to connect to.") // gRPC 服务的地址
	limit = flag.Int64("limit", 10, "Limit to list users.")                                 // 限制列出用户的数量
)

func main() {
	flag.Parse() // 解析命令行参数

	// 建立与 gRPC 服务器的连接
	// grpc.NewClient 用于建立客户端与 gRPC 服务端的连接
	// grpc.WithTransportCredentials(insecure.NewCredentials()) 表示使用不安全的传输（即不使用 TLS）
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to grpc server: %v", err) // 如果连接失败，记录错误并退出程序
	}
	defer conn.Close() // 确保在函数结束时关闭连接，避免资源泄漏

	// 创建 {{.Web.GRPCServiceName}} 客户端
	client := {{.D.APIAlias}}.New{{.Web.GRPCServiceName}}Client(conn) // 使用连接创建一个 {{.Web.GRPCServiceName}} 的 gRPC 客户端实例

	// 设置上下文，带有 3 秒的超时时间
	// context.WithTimeout 用于设置调用的超时时间，防止请求无限等待
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // 在函数结束时取消上下文，释放资源

	// 调用 {{.Web.GRPCServiceName}} 的 Healthz 方法，检查服务健康状况
	resp, err := client.Healthz(ctx, nil) // 发起 gRPC 请求，Healthz 是一个简单的健康检查方法
	if err != nil {
		log.Fatalf("Failed to call healthz: %v", err) // 如果调用失败，记录错误并退出程序
	}

	// 将返回的响应数据转换为 JSON 格式
	jsonData, _ := json.Marshal(resp) // 使用 json.Marshal 将响应对象序列化为 JSON 格式
	fmt.Println(string(jsonData))     // 输出 JSON 数据到终端
}
