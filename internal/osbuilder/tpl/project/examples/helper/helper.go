package helper

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"google.golang.org/grpc/metadata"
	"k8s.io/utils/ptr"

	{{.Web.APIImportPath}}
)

// ExampleCreateUserRequest 创建一个示例的 CreateUserRequest 对象
// 返回一个指针类型的 CreateUserRequest，其中包含随机生成的用户名、固定密码、昵称、邮箱和随机生成的手机号
func ExampleCreateUserRequest() *{{.D.APIAlias}}.CreateUserRequest {
	return &{{.D.APIAlias}}.CreateUserRequest{
		Username: fmt.Sprintf("%d", time.Now().Unix()), // 随机生成一个单词作为用户名，并转换为小写
		Password: "onex(#)666",                         // 设置固定密码
		Nickname: ptr.To("孔令飞"),                        // 设置固    定昵称
		Email:    "colin404@foxmail.com",               // 设置固定邮箱地址
		Phone:    GeneratePhoneNumber(),                // 调用 GeneratePhoneNumber 随机生成一个手机号
	}
}

// GeneratePhoneNumber 随机生成一个符合中国大陆手机格式的号码
// 手机号码规则：以 1 开头，第二位为 3-9，接下来是 9 位随机数字，总共 11 位
func GeneratePhoneNumber() string {
	prefixes := []int{3, 4, 5, 6, 7, 8, 9} // 第二位的合法范围
	rand.Seed(time.Now().UnixNano())       // 设置随机数种子，确保每次生成的号码不同

	// 随机选择第二位数字
	secondDigit := prefixes[rand.Intn(len(prefixes))]

	// 拼接手机号码
	phone := fmt.Sprintf("1%d", secondDigit)
	for i := 0; i < 9; i++ {
		phone += fmt.Sprintf("%d", rand.Intn(10)) // 随机生成剩余的 9 位数字
	}

	return phone
}

// MustWithAdminToken 使用管理员 Token 创建带有授权信息的上下文
// 参数：
// - ctx: 上下文对象
// - client: {{.Web.GRPCServiceName}}Client 客户端，用于调用登录接口
// 返回：
// - 一个附加了管理员 Token 的上下文对象
func MustWithAdminToken(ctx context.Context, client {{.D.APIAlias}}.{{.Web.GRPCServiceName}}Client) context.Context {
	// 使用 root 用户登录
	loginResponse, err := client.Login(ctx, &{{.D.APIAlias}}.LoginRequest{
		Username: "root",         // 固定的管理员用户名
		Password: "miniblog1234", // 固定的管理员密码
	})
	if err != nil {
		log.Printf("Failed to login with root account: %v", err) // 打印登录失败的错误信息
		panic(err)                                               // 如果登录失败，直接终止程序
	}
	log.Printf("[Login          ] Success to login with root account") // 登录成功日志

	// 创建 metadata，用于传递 Token
	md := metadata.Pairs("Authorization", "Bearer "+loginResponse.Token)
	// 将 metadata 附加到上下文中
	ctx = metadata.NewOutgoingContext(ctx, md)
	return ctx
}
