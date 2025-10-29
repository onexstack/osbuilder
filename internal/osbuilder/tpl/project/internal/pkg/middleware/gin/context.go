// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"

	"{{.D.ModuleName}}/internal/pkg/contextx"
)

// Context is a middleware that injects common prefix fields to gin.Context.
func Context() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从当前 span 中获取 traceID 并设置到 gin.Context
		traceID := trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String()

		// 将 traceID 存储到新的 context 中，并更新请求的 context
		ctx := contextx.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
