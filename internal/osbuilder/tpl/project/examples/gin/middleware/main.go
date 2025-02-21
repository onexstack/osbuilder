// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/miniblog. The professional
// version of this repository is https://github.com/onexstack/onex.

package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 定义一个通用中间件：打印请求路径
func LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Request path: %s\n", c.Request.URL.Path)
		// 继续处理后续的中间件或路由
		c.Next()
	}
}

func main() {
	r := gin.Default()

	// 使用全局中间件：所有路由都会经过该中间件
	// r.Use(gin.Logger(), gin.Recovery()) 同时设置多个 Gin 中间件
	r.Use(LogMiddleware())

	// 定义普通路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Home"})
	})

	// 定义一个路由组，并为组添加中间件
	apiGroup := r.Group("/api", LogMiddleware())
	{
		apiGroup.GET("/hello", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Hello, API"})
		})
		apiGroup.GET("/world", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "World, API"})
		})
	}

	// 为单个路由添加中间件
	r.GET("/secure", LogMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "This is a secure route"})
	})

	// 启动HTTP服务
	r.Run(":8080") // 监听在8080端口
}
