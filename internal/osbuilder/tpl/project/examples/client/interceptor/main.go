// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/miniblog. The professional
// version of this repository is https://github.com/onexstack/onex.

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
	"google.golang.org/grpc/metadata"

	apiv1 "github.com/onexstack/miniblog/pkg/api/apiserver/v1"
)

var (
	// 定义命令行参数
	addr  = flag.String("addr", "localhost:6666", "The grpc server address to connect to.") // gRPC 服务的地址
	limit = flag.Int64("limit", 10, "Limit to list users.")                                 // 限制列出用户的数量
)

func main() {
	flag.Parse() // 解析命令行参数

	// 建立与 gRPC 服务器的连接
	conn, err := grpc.NewClient(*addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(unaryClientInterceptor()), // 添加 UnaryClientInterceptor
	)
	if err != nil {
		log.Fatalf("Failed to connect to grpc server: %v", err) // 如果连接失败，记录错误并退出程序
	}
	defer conn.Close() // 确保在函数结束时关闭连接，避免资源泄漏

	// 创建 MiniBlog 客户端
	client := apiv1.NewMiniBlogClient(conn) // 使用连接创建一个 MiniBlog 的 gRPC 客户端实例

	// 设置上下文，带有 3 秒的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // 在函数结束时取消上下文，释放资源

	// 创建一个 Metadata 用于传递请求元数据
	md := metadata.Pairs("custom-header", "value123")
	ctx = metadata.NewOutgoingContext(ctx, md)

	// 调用 MiniBlog 的 Healthz 方法，检查服务健康状况
	var header metadata.MD                                      // 用于存储返回的 Header 元数据
	resp, err := client.Healthz(ctx, nil, grpc.Header(&header)) // 发起 gRPC 请求
	if err != nil {
		log.Fatalf("Failed to call healthz: %v", err) // 如果调用失败，记录错误并退出程序
	}

	for key, val := range header {
		fmt.Printf("Response Header (key: %s, value: %s)\n", key, val)
	}

	// 将返回的响应数据转换为 JSON 格式
	jso.D, _ := json.Marshal(resp) // 使用 json.Marshal 将响应对象序列化为 JSON 格式
	fmt.Println(string(jso.D))     // 输出 JSON 数据到终端
}

func unaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// 打印请求方法和请求参数
		log.Printf("[UnaryClientInterceptor] Invoking method: %s", method)

		// 添加自定义元数据
		md := metadata.Pairs("interceptor-header", "interceptor-value")
		ctx = metadata.NewOutgoingContext(ctx, md)

		// 调用实际的 gRPC 方法
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			log.Printf("[UnaryClientInterceptor] Method: %s, Error: %v", method, err)
			return err
		}

		return nil
	}
}
