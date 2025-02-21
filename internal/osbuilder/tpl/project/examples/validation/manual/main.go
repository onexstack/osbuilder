// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/miniblog. The professional
// version of this repository is https://github.com/onexstack/onex.

package main

import (
	"errors"
	"fmt"
)

type LoginRequest struct {
	Username string
	Password string
}

func validate(req LoginRequest) error {
	if req.Username == "" {
		return errors.New("username is required")
	}
	if len(req.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}
	return nil
}

func main() {
	req := LoginRequest{Username: "user", Password: "12345"}
	if err := validate(req); err != nil {
		fmt.Println("Validation failed:", err)
		return
	}
	fmt.Println("Validation passed")
}
