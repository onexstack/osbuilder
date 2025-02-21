// Copyright 2024 孔令飞 <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/miniblog. The professional
// version of this repository is https://github.com/onexstack/onex.

package main

import (
	"fmt"

	"github.com/onexstack/onexstack/pkg/errorsx"

	"github.com/onexstack/miniblog/internal/pkg/errno"
)

func main() {
	// 创建了一个 ErrorX 错误，表示数据库连接失败。
	// Code: 500，表明是服务器内部错误。
	// Reason: "InternalError.DBConnection"，表示错误的具体分类。
	// Message: "Something went wrong: DB connection failed"，表示该错误的具体信息。
	errx := errorsx.New(500, "InternalError.DBConnection", "Something went wrong: %s", "DB connection failed")

	// fmt.Println 会调用 errx 的 Error 方法，输出：
	// error: code = 500 reason = InternalError.DBConnection message = Something went wrong: DB connection failed metadata = map[]
	fmt.Println(errx)

	// 给错误添加元数据，增强错误的上下文信息，便于调试和追踪。
	errx.WithMetadata(map[string]string{
		"user_id":    "12345",   // 添加用户 ID 信息
		"request_id": "abc-def", // 添加请求 ID 信息
	})

	// 继续向错误中添加元数据，这次使用了 KV 方法，它是一种更加简洁的方式，用 key-value 的模式逐一设置元数据。
	// 这里添加 trace_id 信息，用于关联分布式链路信息。
	errx.KV("trace_id", "xyz-789")

	// 使用 WithMessage 方法更新错误的 Message 字段。
	// 更新后的 Message 是：Updated message: retry failed。
	// Note: 更新消息字段并不会影响 Code、Reason 和 Metadata，它只是说明错误的上下文发生了变化。
	errx.WithMessage("Updated message: %s", "retry failed")

	// 再次打印 errx，此时的内容已经发生了变化：
	// error: code = 500 reason = InternalError.DBConnection message = Updated message: retry failed metadata = map[request_id:abc-def trace_id:xyz-789 user_id:12345]
	// 元数据也会被一并输出。
	fmt.Println(errx)

	// 调用 doSomething 函数，生成一个错误，并打印它，这里返回一个更新过 Message 字段的预定义错误 errno.ErrUsernameInvalid。
	someerr := doSomething()
	// 打印错误。
	// error: code = 400 reason = InvalidArgument.UsernameInvalid message = Username is too short metadata = map[]
	fmt.Println(someerr)

	// 调用预定义错误 errno.ErrUsernameInvalid 的 Is 方法，判断 someerr 是否属于该类型错误。
	// Is 方法会比较 Code 和 Reason 字段（不会比较 Message 字段），如果两者一致，则返回 true。
	// 因为 doSomething 返回的错误正是 errno.ErrUsernameInvalid 的实例，因此这里输出 true。
	fmt.Println(errno.ErrUsernameInvalid.Is(someerr))

	// 调用另外一个预定义错误 errno.ErrPasswordInvalid 的 Is 方法，比较 someerr 是否属于该错误。
	// 因为 Reason 和 Code 不匹配（someerr 是 username 错误，而不是 password 错误），因此返回 false。
	fmt.Println(errno.ErrPasswordInvalid.Is(someerr))
}

// 定义一个函数 doSomething，返回一个错误
func doSomething() error {
	// 这里返回了一个已经定义的错误类型 errno.ErrUsernameInvalid，但动态地设置了 Message 字段为 "Username is too short"。
	// 重点是：虽然错误的 Message 不同，但错误的 Code 和 Reason 是一致的，这方便使用 Is 方法进行类型判断而不受具体内容影响。
	return errno.ErrUsernameInvalid.WithMessage("Username is too short")
}
