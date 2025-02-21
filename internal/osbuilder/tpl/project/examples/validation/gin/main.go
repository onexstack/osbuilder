// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/miniblog. The professional
// version of this repository is https://github.com/onexstack/onex.

package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

func main() {
	r := gin.Default()

	r.POST("/login", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			// 返回校验错误
			errs := err.(validator.ValidationErrors)
			c.JSON(http.StatusBadRequest, gin.H{"error": errs.Error()})
			return
		}

		// 校验通过
		c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
	})

	r.Run()
}
