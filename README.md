# osbuilder：OneX 技术栈项目开发脚手架

osbuilder 项目作为 OneX 技术栈的项目开发脚手架，可以自动生成符合 OneX 技术栈开发规范、软件架构、开发风格的高质量 Go 项目。

osbuilder 支持生成多种应用、多种框架的代码，osbuilder 支持以下核心参数：
- `-t, --type`：支持生成 Web 服务器、异步任务处理服务、命令行工具；
- `--web-framework`：Web 服务器支持 gin、grpc、grpc-gateway、kitex、hertz、kratos、go-zero 等 Web 框架；
- `--storage-type`：底层存储支持 Memory、MariaDB、SQLite、PostgreSQL、Etcd、MongoDB、Redis 等；
- `--deployment-mode`：支持 Systemd、Docker、Kubernetes 等部署模式。

通过不同的可选项， osbuilder 工具会自动生成对应的高质量源码。在自动生成 Web 服务项目类型后，osbuilder 工具还支持给项目添加新的REST资源实现代码。

通过使用 osbuilder 工具自动生成项目源码，可以极大的提高项目开发的效率，并保持项目开发风格的高度一致性，从而降低项目后期的阅读和维护成本。

[miniblog](https://github.com/onexstack/miniblog) 项目，就是用 osbuilder 工具自动生成的。

该工具目前作为 [云原生 AI 实战营](https://konglingfei.com) 知识星球的专有工具，可供星球内的星友免费使用、升级。工具后期也会不断迭代完善。

osbuilder 工具支持 AI 编程能力，该能力正在开发完善中。

## osbuilder 工具使用

本节来演示下如何新建一个 gRPC 服务器，并给该服务器新增 REST 资源。

### 1. 新建一个项目

执行以下命令创建一个新的项目：

```bash
$ mkdir -p cd $GOPATH/src/github.com/onexstack
$ cd $GOPATH/src/github.com/onexstack
$ osbuilder create project -b qa-apiserver --web-framework grpc --kinds job,cron_job --storage-type memory ./demoproj # 创建新 gRPC 服务器项目
$ cd demoproj
$ make protoc.apiserver # 编译 Protobuf 文件
$ go mod tidy # 获取确实的依赖包
$ go generate ./... # 生成依赖注入代码
$ make build BINS=qa-apiserver # 编译 qa-apiserver 组件
$ _output/platforms/linux/amd64/qa-apiserver # 运行 qa-apiserver 组件
```

上述命令会成功创建项目，并给新项目添加 `job`、`cronjob` 资源。

打开另外一个 Linux 终端，执行以下命令进行测试：

```bash
$ cd $GOPATH/src/github.com/onexstack/demoproj
$ go run examples/client/job/main.go # 执行客户端访问 Job 资源
2025/02/23 09:09:35 Creating new job...
2025/02/23 09:09:35 Job created successfully with ID: job-w6irkg
2025/02/23 09:09:35 Creating new job...
2025/02/23 09:09:35 Job created successfully with ID: job-die7iy
2025/02/23 09:09:35 Listing jobs...
2025/02/23 09:09:35 Found 2 jobs in total.
2025/02/23 09:09:35    {"jobID":"job-die7iy","createdAt":{"seconds":1740272975},"updatedAt":{"seconds":1740272975,"nanos":697602742}}
2025/02/23 09:09:35    {"jobID":"job-w6irkg","createdAt":{"seconds":1740272975},"updatedAt":{"seconds":1740272975,"nanos":695978417}}
2025/02/23 09:09:35 Deleting job with ID: job-w6irkg...
2025/02/23 09:09:35 Job with ID: job-w6irkg deleted successfully.
2025/02/23 09:09:35 Listing jobs...
2025/02/23 09:09:35 Found 1 jobs in total.
2025/02/23 09:09:35    {"jobID":"job-die7iy","createdAt":{"seconds":1740272975},"updatedAt":{"seconds":1740272975,"nanos":697602742}}
```

可以看到，客户端可以成功访问 API 接口，并进行资源的 CURD 操作。在新增 REST 资源时，也会给该资源自动生成 `<prefix>-xxxxxx` 格式的资源唯一 ID。

### 2. 给新项目添加 REST 资源

执行以下命令新建资源：

```bash
$ cd $GOPATH/src/github.com/onexstack/demoproj
$ osbuilder create api -b qa-apiserver --kinds task,worker # 给 qa-apiserver 组件新增 task、worker REST 资源
$ make protoc.apiserver # 编译 Protobuf 文件
$ make build BINS=qa-apiserver # 编译 qa-apiserver 组件
$ _output/platforms/linux/amd64/qa-apiserver # 重启 qa-apiserver 组件
```

上述命令会在当前项目中新增 2 个 REST 资源 `task`、`worker` 的 CURD 实现代码。

在一个新的Linux终端中，执行以下命令测试新资源的 CURD：

```bash
$ cd $GOPATH/src/github.com/onexstack/demoproj
$ go run examples/client/worker/main.go # 执行客户端访问 Worker 资源
2025/02/23 09:18:09 Creating new worker...
2025/02/23 09:18:09 Worker created successfully with ID: worker-w6irkg
2025/02/23 09:18:09 Creating new worker...
2025/02/23 09:18:09 Worker created successfully with ID: worker-die7iy
2025/02/23 09:18:09 Listing workers...
2025/02/23 09:18:09 Found 2 workers in total.
2025/02/23 09:18:09    {"workerID":"worker-die7iy","createdAt":{"seconds":1740273489},"updatedAt":{"seconds":1740273489,"nanos":685988836}}
2025/02/23 09:18:09    {"workerID":"worker-w6irkg","createdAt":{"seconds":1740273489},"updatedAt":{"seconds":1740273489,"nanos":684324225}}
2025/02/23 09:18:09 Deleting worker with ID: worker-w6irkg...
2025/02/23 09:18:09 Worker with ID: worker-w6irkg deleted successfully.
2025/02/23 09:18:09 Listing workers...
2025/02/23 09:18:09 Found 1 workers in total.
2025/02/23 09:18:09    {"workerID":"worker-die7iy","createdAt":{"seconds":1740273489},"updatedAt":{"seconds":1740273489,"nanos":685988836}}
```

可以看到客户端可以成功对 Work 资源进行 CURD 操作
