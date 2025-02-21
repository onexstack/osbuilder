package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"k8s.io/utils/ptr"

	"{{.D.ModuleName}}/examples/helper"
	"{{.D.ModuleName}}/internal/pkg/known"
	{{.Web.APIImportPath}}
)

var (
	addr  = flag.String("addr", "localhost:6666", "The grpc server address to connect to.")
	limit = flag.Int64("limit", 10, "Limit to list users.")
)

func main() {
	flag.Parse()

	// 建立与 gRPC 服务器的连接
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to grpc server: %v", err)
	}
	defer conn.Close() // 确保连接在函数结束时关闭

	client := {{.D.APIAlias}}.New{{.Web.GRPCServiceName}}Client(conn) // 创建 {{.Web.GRPCServiceName}} 客户端

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	_ = uuid.New().String()

	createUserRequest := helper.ExampleCreateUserRequest()
	createUserResponse, err := client.CreateUser(ctx, createUserRequest)
	if err != nil {
		log.Fatalf("Failed to create user: %v, username: %s", err, createUserRequest.Username)
	}
	log.Printf("[CreateUser     ] Success to create user, userID: %s", createUserResponse.UserID)

	loginResponse, err := client.Login(ctx, &{{.D.APIAlias}}.LoginRequest{
		Username: createUserRequest.Username,
		Password: createUserRequest.Password,
	})
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}
	if loginResponse.Token == "" {
		log.Printf("Failed to validate token string: received an empty toke")
		return
	}
	log.Printf("[Login          ] Success to login")

	// 创建 metadata，用于传递 Token
	md := metadata.Pairs("Authorization", "Bearer "+loginResponse.Token, known.XUserID, createUserResponse.UserID)
	// 将 metadata 附加到上下文中
	ctx = metadata.NewOutgoingContext(ctx, md)

	defer func() {
		_, _ = client.DeleteUser(ctx, &{{.D.APIAlias}}.DeleteUserRequest{UserID: createUserResponse.UserID})
	}()

	refreshTokenResponse, err := client.RefreshToken(ctx, &{{.D.APIAlias}}.RefreshTokenRequest{})
	if err != nil {
		log.Printf("Failed to refresh token: %v", err)
		return
	}
	if refreshTokenResponse.Token == "" {
		log.Printf("Token cannot be empty")
		return
	}
	log.Printf("[RefreshToken   ] Success to refresh token")

	// 请求 UpdateUser 接口
	_, err = client.UpdateUser(ctx, &{{.D.APIAlias}}.UpdateUserRequest{
		UserID:   createUserResponse.UserID,
		Nickname: ptr.To("令飞孔"),
	})
	if err != nil {
		log.Printf("Failed to update user: %v", err)
		return
	}
	log.Printf("[UpdateUser     ] Success to update user: %v", createUserResponse.UserID)

	// 请求 ChangePassword 接口
	newPassword := "onex(#)888"
	_, err = client.ChangePassword(ctx, &{{.D.APIAlias}}.ChangePasswordRequest{
		UserID:      createUserResponse.UserID,
		OldPassword: createUserRequest.Password,
		NewPassword: newPassword,
	})
	if err != nil {
		log.Printf("Failed to change password: %v", err)
		return
	}
	log.Printf("[ChangePassword ] Success to change password")

	loginResponse, err = client.Login(ctx, &{{.D.APIAlias}}.LoginRequest{
		Username: createUserRequest.Username,
		Password: newPassword,
	})
	if err != nil {
		log.Printf("Failed to login with new password: %v", err)
		return
	}
	log.Printf("[Login          ] Success to login with new password")
	// 创建 metadata，用于传递 Token
	md = metadata.Pairs("Authorization", "Bearer "+loginResponse.Token, known.XUserID, createUserResponse.UserID)
	// 将 metadata 附加到上下文中
	ctx = metadata.NewOutgoingContext(ctx, md)

	getUserResponse, err := client.GetUser(ctx, &{{.D.APIAlias}}.GetUserRequest{UserID: createUserResponse.UserID})
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		return
	}
	if getUserResponse.User.UserID != createUserResponse.UserID || getUserResponse.User.Username != createUserRequest.Username {
		log.Printf("Failed to get user: Username or UserID does not match")
		return
	}
	log.Printf("[GetUser        ] Success to get user: %v", createUserResponse.UserID)

	_, err = client.ListUser(ctx, &{{.D.APIAlias}}.ListUserRequest{Offset: 0, Limit: *limit})
	if err != nil {
	    log.Printf("Failed to list user: %v", err)
	}
    log.Printf("[ListUser       ] Success to list users")

	// 请求 DeleteUser 接口
	_, err = client.DeleteUser(ctx, &{{.D.APIAlias}}.DeleteUserRequest{UserID: createUserResponse.UserID})
	if err != nil {
		log.Printf("Failed to delete user: %v", err)
		return
	}
	log.Printf("[DeleteUser     ] Success to delete user: %v", createUserResponse.UserID)

	log.Printf("[All            ] Success to test all user api")
}

// 随机生成一个符合中国大陆手机格式的号码
func generatePhoneNumber() string {
	// 手机号码规则：以1开头，第二位为3-9，接下来9位随机数字
	prefixes := []int{3, 4, 5, 6, 7, 8, 9} // 第二位的合法范围
	rand.Seed(time.Now().UnixNano())       // 设置随机数种子

	// 随机选择第二位
	secondDigit := prefixes[rand.Intn(len(prefixes))]

	// 随机生成后9位数字
	phone := fmt.Sprintf("1%d", secondDigit)
	for i := 0; i < 9; i++ {
		phone += fmt.Sprintf("%d", rand.Intn(10))
	}

	return phone
}
