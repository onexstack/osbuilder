// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/miniblog. The professional
// version of this repository is https://github.com/onexstack/onex.

package grpc

import (
	"context"

	"google.golang.org/grpc"
)

// DefaulterInterceptor 是一个 gRPC 拦截器，用于对请求进行默认值设置.
func DefaulterInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, rq any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// 调用 Default() 方法（如果存在）
		if defaulter, ok := rq.(interface{ Default() }); ok {
			defaulter.Default()
		}

		// 继续处理请求
		return handler(ctx, rq)
	}
}
