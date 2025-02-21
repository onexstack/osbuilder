package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/onexstack/onexstack/pkg/errorsx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"{{.D.ModuleName}}/internal/pkg/contextx"
	"{{.D.ModuleName}}/internal/pkg/known"
)

// RequestIDInterceptor 是一个 gRPC 拦截器，用于设置请求 ID.
func RequestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var requestID string
		md, _ := metadata.FromIncomingContext(ctx)

		// 从请求中获取请求 ID
		if requestIDs := md[known.XRequestID]; len(requestIDs) > 0 {
			requestID = requestIDs[0]
		}

		// 如果没有请求 ID，则生成一个新的 UUID
		if requestID == "" {
			requestID = uuid.New().String()
			md.Append(known.XRequestID, requestID)
		}

		// 将元数据设置为新的 incoming context
		ctx = metadata.NewIncomingContext(ctx, md)

		// 将请求 ID 设置到响应的 Header Metadata 中
		// grpc.SetHeader 会在 gRPC 方法响应中添加元数据（Metadata），
		// 此处将包含请求 ID 的 Metadata 设置到 Header 中。
		// 注意：grpc.SetHeader 仅设置数据，它不会立即发送给客户端。
		// Header Metadata 会在 RPC 响应返回时一并发送。
		_ = grpc.SetHeader(ctx, md)

		// 将请求 ID 添加到 ctx 中
		//nolint: staticcheck
		ctx = contextx.WithRequestID(ctx, requestID)

		// 继续处理请求
		res, err := handler(ctx, req)
		// 错误处理，附加请求 ID
		if err != nil {
			return res, errorsx.FromError(err).WithRequestID(requestID)
		}

		return res, nil
	}
}
