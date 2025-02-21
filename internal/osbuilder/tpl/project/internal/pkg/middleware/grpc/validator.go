package grpc

import (
	"context"

	"google.golang.org/grpc"
)

// RequestValidator 定义了用于自定义验证的接口.
type RequestValidator interface {
	Validate(ctx context.Context, rq any) error
}

// ValidatorInterceptor 是一个 gRPC 拦截器，用于对请求进行验证.
func ValidatorInterceptor(validator RequestValidator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, rq any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// 调用自定义验证方法
		if err := validator.Validate(ctx, rq); err != nil {
			// 注意这里不用返回 errno.ErrInvalidArgument 类型的错误信息，由 validator.Validate 返回.
			return nil, err // 返回验证错误
		}

		// 继续处理请求
		return handler(ctx, rq)
	}
}
