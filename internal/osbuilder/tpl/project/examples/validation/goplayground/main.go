// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/miniblog. The professional
// version of this repository is https://github.com/onexstack/onex.

package main

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type LoginRequest struct {
	Username string `validate:"required"`       // 必填字段
	Password string `validate:"required,min=6"` // 最小长度为6
	Email    string `validate:"required,email"` // 必填且必须是邮箱格式
}

func main() {
	validate := validator.New() // 创建验证器

	req := LoginRequest{
		Username: "user",
		Password: "12345",
		Email:    "invalid-email",
	}

	// 校验结构体
	err := validate.Struct(req)
	if err != nil {
		// 获取校验错误并打印
		for _, err := range err.(validator.ValidationErrors) {
			fmt.Printf("Field '%s' failed validation, rule '%s'\n", err.Field(), err.Tag())
		}
	} else {
		fmt.Println("Validation passed!")
	}
}
