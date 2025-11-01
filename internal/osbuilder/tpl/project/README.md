# Project {{.D.ProjectName}}

{{.D.ProjectName}} 是一个基于 Go 语言开发的现代化微服务应用，采用简洁架构设计，具有代码质量高、扩展能力强、符合 Go 编码及最佳实践等特点。

{{.D.ProjectName}} 具有以下特性：
- 软件架构：采用简洁架构设计，确保项目结构清晰、易维护；
- 高频 Go 包：使用了 Go 项目开发中常用的包，如 {{if hasGin .WebServers }}gin、{{- end}}{{- if hasGRPC .WebServers }}grpc、{{- end}}{{if hasOTel .WebServers }}otel、{{- end}}gorm、gin、uuid、cobra、viper、pflag、resty、govalidator、slog、protobuf、casbin、onexstack 等；
- 目录结构：遵循 [project-layout](https://github.com/golang-standards/project-layout) 规范，采用标准化的目录结构；
- 认证与授权：实现了基于 JWT 的认证和基于 Casbin 的授权功能；
- 错误处理：设计了独立的错误包及错误码管理机制；
- 构建与管理：使用高质量的 Makefile 对项目进行管理；
- 代码质量：通过 golangci-lint 工具对代码进行静态检查，确保代码质量；
- 测试覆盖：包含单元测试、性能测试、模糊测试和示例测试等多种测试案例；
- 丰富的 Web 功能：支持 Trace ID、优雅关停、中间件、跨域处理、异常恢复等功能；
- 多种数据交换格式：支持 JSON 和 Protobuf 数据格式的交换；
- 开发规范：遵循多种开发规范，包括代码规范、版本规范、接口规范、日志规范、错误规范以及提交规范等；
- API 设计：接口设计遵循 RESTful API 规范；
- 项目具有 Dockerfile，并且 Dockerfile 符合最佳实践；
{{- if hasGRPC .WebServers }}
- 支持 OpenTelemetry 全链路可观测能力：Tracing、Metrics、Logs；
{{- end}}
{{- if hasServiceRegistry .WebServers }}
- 服务注册/服务发现；
{{- end}}

## Getting Started

### Prerequisites

在开始之前，请确保您的开发环境中安装了以下工具：

**必需工具：**
- [Go](https://golang.org/dl/) 1.25.3 或更高版本
- [Git](https://git-scm.com/) 版本控制工具
{{- if hasGRPC .WebServers }}
- Protobuf 编译工具（可执行 `make deps` 自动安装）：
  - [Protocol Buffers](https://protobuf.dev/) 编译器 (protoc) - 用于 gRPC 代码生成
  - [protoc-gen-go](https://github.com/protocolbuffers/protobuf-go) - Go protobuf 插件
  - [protoc-gen-go-grpc](https://github.com/grpc/grpc-go) - gRPC Go 插件
{{- end }}
{{- if ne .Metadata.MakefileMode "none" }}
- [Make](https://www.gnu.org/software/make/) 构建工具
{{- end }}

**可选工具：**
{{- if or (eq .Metadata.DeploymentMethod "docker") (eq .Metadata.DeploymentMethod "kubernetes") }}
- [Docker](https://www.docker.com/) 容器化部署
{{- end }}
{{- if eq .Metadata.DeploymentMethod "kubernetes" }}
- [Kubernetes](https://kubernetes.io/) 云原生部署
{{- end }}
- [golangci-lint](https://golangci-lint.run/) 代码静态检查

**验证安装：**

```bash
$ go version  
go version go1.25.3 linux/amd64  

{{- if ne .Metadata.MakefileMode "none" }}  
$ make --version  
GNU Make 4.3  
{{- end }}  

{{- if hasGRPC .WebServers }}  
$ protoc --version  
libprotoc 3.21.12  
{{- end }}  
```

### Building

> 提示：项目配置文件配置项 `metadata.makefileMode` 不能为 `none`，如果为 `none` 需要自行构建。

在项目根目录下，执行以下命令构建项目：

**1. 安装依赖工具和包**

```bash
$ make deps  # 安装项目所需的开发工具  
$ go mod tidy # 下载 Go 模块依赖  
```

**2. 生成代码**

```bash
$ make protoc # generate gRPC code  
$ go get cloud.google.com/go/compute@latest cloud.google.com/go/compute/metadata@latest  
$ go mod tidy # tidy dependencies  
$ go generate ./... # run all go:generate directives  
```

**3. 构建应用**

```bash
$ make build # build all binary files locate in cmd/  
```

**构建结果：**

```bash
_output/platforms/  
├── linux/  
│   └── amd64/  
{{- range .WebServers }}  
│       └── {{.BinaryName}}  # {{.Name}} 服务二进制文件  
{{- end }}  
└── darwin/  
    └── amd64/  
{{- range .WebServers }}  
        └── {{.BinaryName}}  
{{- end }}  
```

### Running

启动服务有多种方式：

**1. 使用构建的二进制文件运行**

```bash
{{- range $i, $ws := .WebServers }}  
# 启动 {{$ws.Name}} 服务  
$ _output/platforms/linux/amd64/{{$ws.BinaryName}} --config configs/{{$ws.BinaryName}}.yaml  
{{- if eq $ws.WebFramework "gin" }}  
# 服务将在以下端口启动：  
# - HTTP API: http://localhost:5555
# - Health Check: http://localhost:5555/healthz  
# - Metrics: http://localhost:5555/metrics  
$ curl http://localhost:5555/healthz # 测试：打开另外一个终端，调用健康检查接口
{{- else if eq $ws.WebFramework "grpc" }}  
# 服务将在以下端口启动：  
# - gRPC API: localhost:6666
# - Health Check: localhost:6666 (gRPC Health Protocol)  
$ go run examples/client/health/main.go # 测试：打开另外一个终端，调用健康检查接口
{{- end }}  
{{- end }}  
```

**2. 使用 Docker 运行**

```bash
# 构建镜像  
$ make image
{{- range .WebServers }}
$ docker run --name {{.BinaryName}} -v configs/{{.BinaryName}}.yaml:/etc/{{.BinaryName}}.yaml -p {{if eq .WebFramework "gin" }}5555:5555{{- end}}{{if eq .WebFramework "grpc" }}6666:6666{{- end}} {{$.Project.Metadata.Image.RegistryPrefix}}/{{.BinaryName}}:latest -c /etc/{{.BinaryName}}.yaml
{{- end }}
```

**配置文件示例：**

{{- range .WebServers }}  

{{.BinaryName}} 配置文件 `configs/{{.BinaryName}}.yaml`：

```yaml
{{- if eq .WebFramework "gin" }}
addr: 0.0.0.0:5555 # 服务监听地址
timeout: 30s # 服务端超时
{{- end -}}
{{- if eq .WebFramework "grpc" }}
{{- if .WithOTel }}
metrics-addr: 0.0.0.0:29090
{{- end -}}
grpc:  
  port: 6666
  timeout: 30s
{{- end -}}
{{- if not .WithOTel }}
slog:
  level: info # debug, info, warn, error
  add-source: true
  format: json # console, json
  time-format: "2006-01-02 15:04:05"
  output: stdout
{{- end -}}
{{- if .WithOTel }}
otel:
  endpoint: 127.0.0.1:4327
  service-name: {{.BinaryName}}
  output-mode: otel
  level: debug
  add-source: true
  use-prometheus-endpoint: true
  slog: # 改配置项只有 output-mod 为 slog 时生效
    format: text
    time-format: "2006-01-02 15:04:05"
    output: stdout
{{- end -}}
{{- if eq .ServiceRegistry "polaris" }}  
polaris:
  addr: 127.0.0.1:8091
  timeout: 30s
  retry-count: 3
  provider:
    namespace: {{$.D.ProjectName}}
    service: {{.BinaryName}}
{{- end -}}  
{{- if or (eq .StorageType "mysql") (eq .StorageType "mariadb") }}
mysql:  
  addr: 127.0.0.1:3306
  username: onex
  password: "onex(#)666"
  database: onex
  max-connection-life-time: 10s
  max-idle-connections: 100
  max-open-connections: 100
{{- end}}
```
{{- end }}  

## Versioning

本项目遵循 [语义版本控制](https://semver.org/lang/zh-CN/) 规范。

## Authors

### 主要贡献者

- **{{.Metadata.Author}}** - *项目创建者和维护者* - [{{.Metadata.Email}}](mailto:{{.Metadata.Email}})
  - 项目架构设计
  - 核心功能开发
  - 技术方案制定

### 贡献者列表

感谢所有为本项目做出贡献的开发者们！

<!-- 这里会自动显示贡献者头像 -->
<a href="{{.D.ModuleName}}/graphs/contributors">
  <img src="https://contrib.rocks/image?repo={{.D.ModuleName}}" />
</a>

*贡献者列表由 [contrib.rocks](https://contrib.rocks) 生成*

## 附录

### 项目结构

```bash
{{.D.ProjectName}}/  
├── cmd/                     # 应用程序入口  
{{- range .WebServers }}  
│   └── {{.BinaryName}}/       # {{.Name}} 服务  
│       └── main.go          # 主函数  
{{- end }}  
├── internal/                # 私有应用程序代码  
{{- range .WebServers }}  
│   └── {{.Name}}/             # {{.Name}} 内部包  
│       ├── biz/             # 业务逻辑层  
│       ├── handler/         # {{.WebFramework}} 处理器  
│       ├── model/           # GORM 数据模型  
│       ├── pkg/             # 内部工具包  
│       └── store/           # 数据访问层  
{{- end }}  
├── pkg/                     # 公共库代码  
│   ├── api/                 # API 定义{{if hasGRPC .WebServers}} (protobuf){{end}}  
├── examples/                # 示例代码  
│   └── client/              # 客户端示例  
├── configs/                 # 配置文件  
├── docs/                    # 项目文档  
├── build/                   # 构建配置  
│   └── docker/              # Docker 文件  
{{- if ne .Metadata.MakefileMode "none" }}  
├── scripts/                 # 构建和部署脚本  
{{- end }}  
├── third_party/             # 第三方依赖  
{{- if ne .Metadata.MakefileMode "none" }}  
├── Makefile                 # 构建配置  
{{- end }}  
├── go.mod                   # Go 模块文件  
├── go.sum                   # Go 模块校验文件  
└── README.md                # 项目说明文档  
```

### 相关链接

- [项目文档](docs/)
- [问题追踪]({{.D.ModuleName}}/issues)
- [讨论区]({{.D.ModuleName}}/discussions)
- [项目看板]({{.D.ModuleName}}/projects)
- [发布页面]({{.D.ModuleName}}/releases)

### 支持

如果这个项目对您有帮助，请考虑给我们一个 ⭐️ 来支持项目发展！

[![Star History Chart](https://api.star-history.com/svg?repos={{.D.ModuleName}}&type=Date)](https://star-history.com/#{{.D.ModuleName}}&Date)
