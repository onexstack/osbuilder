# osbuilder: onexstack 技术栈脚手架工具

**osbuilder：** onexstack 技术栈使用的 Go 项目开发脚手架。

## onexstack 技术栈介绍

onexstack 是一整套 Go 开发技术栈。该技术栈包括了以下内容：
- 学习社群（欢迎加入）：[云原生 AI 实战营](https://t.zsxq.com/5T0qC)
- 高质量的 Go 项目：[「云原生 AI 实战营」项目介绍](https://konglingfei.com/cloudai/project/cloudai.html)
- 高质量的课程：[「云原生 AI 实战营」体系课介绍](https://konglingfei.com/cloudai/catalog/cloudai.html)
- 一系列开发规范：[技术栈相关规范](https://konglingfei.com/onex/convention/rest.html)
- 一系列开发标准包/工具：[onexstack 标准化包](https://github.com/onexstack/onexstack)

onexstack 技术栈中，所有的 Web 服务器类型的项目都是使用 `osbuilder` 脚手架自动生成，例如：[miniblog](https://github.com/onexstack/miniblog)。

## osbuilder 工具介绍

### 安装

```bash
$ go install github.com/onexstack/osbuilder/cmd/osbuilder@latest
$ osbuilder version
```

## osbuilder 脚手架使用

osbuilder 脚手架可以用来生产一个新的项目，也能够基于已有的项目添加新的 REST 资源。


### 1. 生成新项目

```bash
$ mkdir -p $GOPATH//src/github.com/onexstack
$ cd $GOPATH//src/github.com/onexstack
$ cat << EOF > project.yaml
scaffold: osbuilder
version: v0.0.14
metadata:
  shortDescription: Please update the short description of the binary file.
  longMessage: Please update the detailed description of the binary file.
  # 选择二进制文件的部署形式。当前支持 systemd、docker。未来会支持 kubernetes。会生成 Dockerfile、Kubernetes YAML 等资源
  deploymentMethod: kubernetes
  image:
    # 当指定 deploymentMethod 为 docker、kubernetes 时，构建镜像的地址
    registry: docker.io
    # 指定 Dockerfile 的生成模式。可选的模式有：
    # - none：不生成 Dockerfile。需要自行实现 build/docker/<component_name>/Dockerfile 文件；
    # - runtime-only：仅包含运行时阶段（适合已有外部构建产物），适合本地调试；
    # - multi-stage：多阶段构建（builder + runtime）；
    # - combined：同时生成 multi-stage、runtime-only 2 种类型的 Dockerfile：
    #   - multi-stage：Dockerfile 名字为 Dockerfile
    #   - runtime-only：Dockerfile 名字为 Dockerfile.runtime-only
    dockerfileMode: combined
    # 是否采用 distroless 运行时镜像。如果不采用会使用 debian 基础镜像，否则使用 gcr.io/distroless/base-debian12:nonroot
    # - true：采用 gcr.io/distroless/base-debian12:nonroot 基础镜像。生产环境建议设置为 true；
    # - false：采用 debian:bookworm 基础镜像。测试环境建议设置为 fasle；
    distroless: true
  # 控制 Makefile 的生成方式。当前支持以下 3 种：
  # - none：不生成 makefile
  # - structured：生成单个 makefile
  # - unstructured：生成结构化的 makefile
  makefileMode: unstructured
  # 项目创建者名字，用于生成版权信息
  author: 孔令飞
  # 项目创建者邮箱，用于生成版权信息
  email: colin404@foxmail.com
# osbuilder 支持多种应用类型。当前仅支持 Web 服务类型
# 未来会支持：异步任务 Job 类型、命令行工具类型、声明式API服务器类型
webServers:
  - binaryName: mb-apiserver
    # Web Server 使用的框架。当前支持 gin、grpc
    # 未来会支持 kratos、grpc-gateway、go-zero、kitex、hertz 等
    webFramework: gin
    # 可选，当 webFramework 为 grpc 时有效，指定 grpc 服务的名字
    grpcServiceName: APIServer
    # Web Server 后端使用的存储类型。当前支持 memory、mysql
    # 未来会支持etcd、redis、sqlite、mongo、postgresql
    storageType: memory 
    # 是否添加健康检查接口
    withHealthz: true
    # 是否添加用户默认，开启后，有完整的认证、鉴权流程
    withUser: false
    # 是否生成注册/反注册到腾讯北极星服务中心的代码
    withPolaris: false
EOF
$ osbuilder create project --config project.yaml ./miniblog
...
🍺 Project creation succeeded miniblog
💻 Use the following command to start the project 👇:

$ cd /home/colin/workspace/golang/src/github.com/onexstack/miniblog # enter project directory
$ make deps # (Optional, executed when dependencies missing) Install tools required by project.
$ make protoc.apiserver # generate gRPC code
$ go mod tidy # tidy dependencies
$ go generate ./... # run all go:generate directives
$ make build BINS=mb-apiserver # build mb-apiserver
$ _output/platforms/linux/amd64/mb-apiserver # run the compiled server
$ go run examples/client/health/main.go # run health client to test the API

🤝 Thanks for using osbuilder.
👉 Visit https://t.zsxq.com/5T0qC to learn how to develop miniblog project.
```

执行上述命令后，可以根据提示，执行以下命令来部署并测试服务：
```bash
$ cd /home/colin/workspace/golang/src/github.com/onexstack/miniblog # enter project directory
$ make deps # (Optional, executed when dependencies missing) Install tools required by project.
$ make protoc.apiserver # generate gRPC code
$ go mod tidy # tidy dependencies
$ go generate ./... # run all go:generate directives
$ make build BINS=mb-apiserver # build mb-apiserver
$ _output/platforms/linux/amd64/mb-apiserver # run the compiled server
$  go run examples/client/health/main.go # run health client to test the API
{"timestamp":"2025-08-24 13:23:19"}
```

可以看到，整个项目的生成过程很丝滑，而且生成的项目跟 [miniblog](https://github.com/onexstack/miniblog) 保持高度一致。miniblog 项目有完整的开发体系课，想学习的可以加入 [云原生 AI 实战营](https://t.zsxq.com/5T0qC)。


> 提示：如果想生产带认证鉴权的项目实例，需要设置：webserver[0].withUser 为 `true`。

### 2. 基于已有项目添加新的 REST 资源

```bash
$ cd /home/colin/workspace/golang/src/github.com/onexstack/miniblog
# -b 选项指定给 mb-apiserver 资源添加新的 REST 资源：cron_job、job
$ osbuilder create api --kinds cron_job,job -b mb-apiserver 
```

上述命令会添加 2 个新的 REST 资源：CronJob、Job。接下来，你只需要添加核心业务逻辑即可。

执行完 `osbuilder` 命令之后，会提示如何进行编译。按提示编译并测试：
```bash
$ make protoc.apiserver 
$ make build BINS=mb-apiserver
$ _output/platforms/linux/amd64/mb-apiserver
# 提示：如果指定了 withUser: true，则需要给 grpc 客户端添加认证信息，否则会报：Unauthenticated 错误
$ go run examples/client/cronjob/main.go 
2025/08/24 13:34:35 Creating new cronjob...
2025/08/24 13:34:35 CronJob created successfully with ID: cronjob-zhwu4c
2025/08/24 13:34:35 Creating new cronjob...
2025/08/24 13:34:35 CronJob created successfully with ID: cronjob-gus02u
2025/08/24 13:34:35 Listing cronjobs...
2025/08/24 13:34:35 Found 2 cronjobs in total.
2025/08/24 13:34:35    {"cronJobID":"cronjob-gus02u","createdAt":{"seconds":1756013675},"updatedAt":{"seconds":1756013675,"nanos":57765906}}
2025/08/24 13:34:35    {"cronJobID":"cronjob-zhwu4c","createdAt":{"seconds":1756013675},"updatedAt":{"seconds":1756013675,"nanos":57131637}}
2025/08/24 13:34:35 Deleting cronjob with ID: cronjob-zhwu4c...
2025/08/24 13:34:35 CronJob with ID: cronjob-zhwu4c deleted successfully.
2025/08/24 13:34:35 Listing cronjobs...
2025/08/24 13:34:35 Found 1 cronjobs in total.
2025/08/24 13:34:35    {"cronJobID":"cronjob-gus02u","createdAt":{"seconds":1756013675},"updatedAt":{"seconds":1756013675,"nanos":57765906}}
```
