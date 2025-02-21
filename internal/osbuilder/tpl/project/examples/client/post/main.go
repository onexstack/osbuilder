package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	{{.Web.APIImportPath}}
)

const (
	serverAddress = "localhost:6666" // 服务地址
	defaultPage   = 1                // 分页默认页码
	defaultSize   = 20               // 分页默认大小
)

func main() {
	// 创建客户端连接
	conn, err := newConnection(serverAddress)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// 创建 gRPC 客户端
	cli := {{.D.APIAlias}}.New{{.Web.GRPCServiceName}}Client(conn)
	ctx := context.Background()

	// 执行操作
	id1 := create{{.Web.R.SingularName}}(ctx, cli)
	_ = create{{.Web.R.SingularName}}(ctx, cli)
	list{{.Web.R.SingularName}}(ctx, cli)
	delete{{.Web.R.SingularName}}(ctx, cli, id1)
	list{{.Web.R.SingularName}}(ctx, cli)
}

// newConnection 创建 gRPC 连接
func newConnection(target string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 使用不安全连接，生产环境应使用 TLS
		grpc.WithBlock(), // 阻塞直到连接建立
	}

	return grpc.DialContext(ctx, target, opts...)
}

// checkError 通用错误检查函数
func checkError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

// create{{.Web.R.SingularName}} 创建一个新的 {{.Web.R.SingularName}}
func create{{.Web.R.SingularName}}(ctx context.Context, cli {{.D.APIAlias}}.{{.Web.GRPCServiceName}}Client) string {
	log.Println("Creating new {{.Web.R.SingularLower}}...")

	req := &{{.D.APIAlias}}.Create{{.Web.R.SingularName}}Request{
		// 设置请求参数
	}

	resp, err := cli.Create{{.Web.R.SingularName}}(ctx, req)
	checkError(err, "Failed to create {{.Web.R.SingularLower}}")

	log.Printf("{{.Web.R.SingularName}} created successfully with ID: %s\n", resp.Get{{.Web.R.SingularName}}ID())
	return resp.Get{{.Web.R.SingularName}}ID()
}

// list{{.Web.R.SingularName}} 列出所有 {{.Web.R.SingularName}}
func list{{.Web.R.SingularName}}(ctx context.Context, cli {{.D.APIAlias}}.{{.Web.GRPCServiceName}}Client) {
	log.Println("Listing {{.Web.R.PluralLower}}...")

	req := &{{.D.APIAlias}}.List{{.Web.R.SingularName}}Request{
		Offset: defaultPage,
		Limit:  defaultSize,
	}

	resp, err := cli.List{{.Web.R.SingularName}}(ctx, req)
	checkError(err, "Failed to list {{.Web.R.PluralLower}}")

	log.Printf("Found %d {{.Web.R.PluralLower}} in total.", resp.GetTotal())
	for _, {{.Web.R.SingularLower}} := range resp.Get{{.Web.R.SingularName}}s() {
		objBytes, _ := json.Marshal({{.Web.R.SingularLower}})
		log.Println("  ", string(objBytes))
	}
}

// delete{{.Web.R.SingularName}} 删除一个指定的 {{.Web.R.SingularName}}
func delete{{.Web.R.SingularName}}(ctx context.Context, cli {{.D.APIAlias}}.{{.Web.GRPCServiceName}}Client,  {{.Web.R.SingularLowerFirst}}ID string) {
	log.Printf("Deleting {{.Web.R.SingularLower}} with ID: %s...",  {{.Web.R.SingularLowerFirst}}ID)

	req := &{{.D.APIAlias}}.Delete{{.Web.R.SingularName}}Request{
		{{.Web.R.SingularName}}IDs: []string{ {{.Web.R.SingularLowerFirst}}ID},
	}

	_, err := cli.Delete{{.Web.R.SingularName}}(ctx, req)
	checkError(err, fmt.Sprintf("Failed to delete {{.Web.R.SingularLower}} with ID: %s",  {{.Web.R.SingularLowerFirst}}ID))

	log.Printf("{{.Web.R.SingularName}} with ID: %s deleted successfully.",  {{.Web.R.SingularLowerFirst}}ID)
}
