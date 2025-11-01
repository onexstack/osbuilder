#!/bin/bash

onexstack_dir=/home/colin/workspace/golang/src/github.com/onexstack

function test() {
  eval "$1"
  cd ${onexstack_dir}/osbuilder/internal/osbuilder/tpl
  ./gen-statik.sh

  cd ${onexstack_dir}/osbuilder/cmd/osbuilder
  go build -v . || exit 1

  cd ${onexstack_dir}
  rm -rf miniblog; /tmp/osbuilder create project --config /tmp/miniblog.yaml ./miniblog

  cd ${onexstack_dir}/miniblog
  make protoc.apiserver
  go get cloud.google.com/go/compute@latest
  go get cloud.google.com/go/compute/metadata@latest
  go mod tidy
  go generate ./...
  if [ "$1" == "miniblog_9" -o "$1" == "miniblog_10" ];then
    make image LOCAL_DOCKERFILE=1
  else
    make build
  fi
}

function miniblog_1() {
  cat << EOF > /tmp/miniblog.yaml
scaffold: osbuilder
version: v0.0.1
metadata:
  shortDescription: Please update the short description of the binary file.
  longMessage: Please update the detailed description of the binary file.
  # 当指定deploymentMethod为docker、kubernetes时，构建镜像的地址
  registry: docker.io
  # 选择二进制文件的部署形式。当前近支持systemd。未来会支持docker、kubernetes，会生产Dockerfile、Kubernetes YAML 等资源
  deploymentMethod: systemd
  # 是否使用结构化的 makefile。非结构化功能简单，结构化设计复杂，但扩展能力强
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
    # 未来会支持kratos、grpc-gateway、go-zero、kitex、hertz等
    webFramework: gin
    # Web Server 后端使用的存储类型。当前支持 memory、mysql
    # 未来会支持etcd、redis、sqlite、mongo、postgresql
    storageType: memory
    # 是否添加健康检查接口
    withHealthz: true
    # 是否添加用户默认，开启后，有完整的认证、鉴权流程
    withUser: true
    withPolaris: true
EOF
}

function miniblog_2() {
  cat << EOF > /tmp/miniblog.yaml
scaffold: osbuilder
version: v0.0.1
metadata:
  shortDescription: Please update the short description of the binary file.
  longMessage: Please update the detailed description of the binary file.
  # 当指定deploymentMethod为docker、kubernetes时，构建镜像的地址
  registry: docker.io
  # 选择二进制文件的部署形式。当前近支持systemd。未来会支持docker、kubernetes，会生产Dockerfile、Kubernetes YAML 等资源
  deploymentMethod: systemd
  # 是否使用结构化的 makefile。非结构化功能简单，结构化设计复杂，但扩展能力强
  makefileMode: structured
  # 项目创建者名字，用于生成版权信息
  author: 孔令飞
  # 项目创建者邮箱，用于生成版权信息
  email: colin404@foxmail.com
# osbuilder 支持多种应用类型。当前仅支持 Web 服务类型
# 未来会支持：异步任务 Job 类型、命令行工具类型、声明式API服务器类型
webServers:
  - binaryName: mb-apiserver
    # Web Server 使用的框架。当前支持 gin、grpc
    # 未来会支持kratos、grpc-gateway、go-zero、kitex、hertz等
    webFramework: gin
    # Web Server 后端使用的存储类型。当前支持 memory、mysql
    # 未来会支持etcd、redis、sqlite、mongo、postgresql
    storageType: memory
    # 是否添加健康检查接口
    withHealthz: true
    # 是否添加用户默认，开启后，有完整的认证、鉴权流程
    withUser: true
    withPolaris: true
EOF
}

function miniblog_2() {
  cat << EOF > /tmp/miniblog.yaml
scaffold: osbuilder
version: v0.0.1
metadata:
  shortDescription: Please update the short description of the binary file.
  longMessage: Please update the detailed description of the binary file.
  # 当指定deploymentMethod为docker、kubernetes时，构建镜像的地址
  registry: docker.io
  # 选择二进制文件的部署形式。当前近支持systemd。未来会支持docker、kubernetes，会生产Dockerfile、Kubernetes YAML 等资源
  deploymentMethod: systemd
  # 是否使用结构化的 makefile。非结构化功能简单，结构化设计复杂，但扩展能力强
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
    # 未来会支持kratos、grpc-gateway、go-zero、kitex、hertz等
    webFramework: gin
    # Web Server 后端使用的存储类型。当前支持 memory、mysql
    # 未来会支持etcd、redis、sqlite、mongo、postgresql
    storageType: memory
    # 是否添加健康检查接口
    withHealthz: false
    # 是否添加用户默认，开启后，有完整的认证、鉴权流程
    withUser: true
    withPolaris: true
EOF
}

function miniblog_3() {
  cat << EOF > /tmp/miniblog.yaml
scaffold: osbuilder
version: v0.0.1
metadata:
  shortDescription: Please update the short description of the binary file.
  longMessage: Please update the detailed description of the binary file.
  # 当指定deploymentMethod为docker、kubernetes时，构建镜像的地址
  registry: docker.io
  # 选择二进制文件的部署形式。当前近支持systemd。未来会支持docker、kubernetes，会生产Dockerfile、Kubernetes YAML 等资源
  deploymentMethod: systemd
  # 是否使用结构化的 makefile。非结构化功能简单，结构化设计复杂，但扩展能力强
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
    # 未来会支持kratos、grpc-gateway、go-zero、kitex、hertz等
    webFramework: gin
    # Web Server 后端使用的存储类型。当前支持 memory、mysql
    # 未来会支持etcd、redis、sqlite、mongo、postgresql
    storageType: memory
    # 是否添加健康检查接口
    withHealthz: true
    # 是否添加用户默认，开启后，有完整的认证、鉴权流程
    withUser: false
    withPolaris: true
EOF
}

function miniblog_4() {
  cat << EOF > /tmp/miniblog.yaml
scaffold: osbuilder
version: v0.0.1
metadata:
  shortDescription: Please update the short description of the binary file.
  longMessage: Please update the detailed description of the binary file.
  # 当指定deploymentMethod为docker、kubernetes时，构建镜像的地址
  registry: docker.io
  # 选择二进制文件的部署形式。当前近支持systemd。未来会支持docker、kubernetes，会生产Dockerfile、Kubernetes YAML 等资源
  deploymentMethod: systemd
  # 是否使用结构化的 makefile。非结构化功能简单，结构化设计复杂，但扩展能力强
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
    # 未来会支持kratos、grpc-gateway、go-zero、kitex、hertz等
    webFramework: gin
    # Web Server 后端使用的存储类型。当前支持 memory、mysql
    # 未来会支持etcd、redis、sqlite、mongo、postgresql
    storageType: memory
    # 是否添加健康检查接口
    withHealthz: true
    # 是否添加用户默认，开启后，有完整的认证、鉴权流程
    withUser: true
    withPolaris: false
EOF
}

function miniblog_5() {
  cat << EOF > /tmp/miniblog.yaml
scaffold: osbuilder
version: v0.0.1
metadata:
  shortDescription: Please update the short description of the binary file.
  longMessage: Please update the detailed description of the binary file.
  # 当指定deploymentMethod为docker、kubernetes时，构建镜像的地址
  registry: docker.io
  # 选择二进制文件的部署形式。当前近支持systemd。未来会支持docker、kubernetes，会生产Dockerfile、Kubernetes YAML 等资源
  deploymentMethod: systemd
  # 是否使用结构化的 makefile。非结构化功能简单，结构化设计复杂，但扩展能力强
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
    # 未来会支持kratos、grpc-gateway、go-zero、kitex、hertz等
    webFramework: grpc
    # Web Server 后端使用的存储类型。当前支持 memory、mysql
    # 未来会支持etcd、redis、sqlite、mongo、postgresql
    storageType: memory
    # 是否添加健康检查接口
    withHealthz: true
    # 是否添加用户默认，开启后，有完整的认证、鉴权流程
    withUser: true
    withPolaris: true
EOF
}

function miniblog_2() {
  cat << EOF > /tmp/miniblog.yaml
scaffold: osbuilder
version: v0.0.1
metadata:
  shortDescription: Please update the short description of the binary file.
  longMessage: Please update the detailed description of the binary file.
  # 当指定deploymentMethod为docker、kubernetes时，构建镜像的地址
  registry: docker.io
  # 选择二进制文件的部署形式。当前近支持systemd。未来会支持docker、kubernetes，会生产Dockerfile、Kubernetes YAML 等资源
  deploymentMethod: systemd
  # 是否使用结构化的 makefile。非结构化功能简单，结构化设计复杂，但扩展能力强
  makefileMode: structured
  # 项目创建者名字，用于生成版权信息
  author: 孔令飞
  # 项目创建者邮箱，用于生成版权信息
  email: colin404@foxmail.com
# osbuilder 支持多种应用类型。当前仅支持 Web 服务类型
# 未来会支持：异步任务 Job 类型、命令行工具类型、声明式API服务器类型
webServers:
  - binaryName: mb-apiserver
    # Web Server 使用的框架。当前支持 gin、grpc
    # 未来会支持kratos、grpc-gateway、go-zero、kitex、hertz等
    webFramework: grpc
    # Web Server 后端使用的存储类型。当前支持 memory、mysql
    # 未来会支持etcd、redis、sqlite、mongo、postgresql
    storageType: memory
    # 是否添加健康检查接口
    withHealthz: true
    # 是否添加用户默认，开启后，有完整的认证、鉴权流程
    withUser: true
    withPolaris: true
EOF
}

function miniblog_6() {
  cat << EOF > /tmp/miniblog.yaml
scaffold: osbuilder
version: v0.0.1
metadata:
  shortDescription: Please update the short description of the binary file.
  longMessage: Please update the detailed description of the binary file.
  # 当指定deploymentMethod为docker、kubernetes时，构建镜像的地址
  registry: docker.io
  # 选择二进制文件的部署形式。当前近支持systemd。未来会支持docker、kubernetes，会生产Dockerfile、Kubernetes YAML 等资源
  deploymentMethod: systemd
  # 是否使用结构化的 makefile。非结构化功能简单，结构化设计复杂，但扩展能力强
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
    # 未来会支持kratos、grpc-gateway、go-zero、kitex、hertz等
    webFramework: grpc
    # Web Server 后端使用的存储类型。当前支持 memory、mysql
    # 未来会支持etcd、redis、sqlite、mongo、postgresql
    storageType: memory
    # 是否添加健康检查接口
    withHealthz: false
    # 是否添加用户默认，开启后，有完整的认证、鉴权流程
    withUser: true
    withPolaris: true
EOF
}

function miniblog_7() {
  cat << EOF > /tmp/miniblog.yaml
scaffold: osbuilder
version: v0.0.1
metadata:
  shortDescription: Please update the short description of the binary file.
  longMessage: Please update the detailed description of the binary file.
  # 当指定deploymentMethod为docker、kubernetes时，构建镜像的地址
  registry: docker.io
  # 选择二进制文件的部署形式。当前近支持systemd。未来会支持docker、kubernetes，会生产Dockerfile、Kubernetes YAML 等资源
  deploymentMethod: systemd
  # 是否使用结构化的 makefile。非结构化功能简单，结构化设计复杂，但扩展能力强
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
    # 未来会支持kratos、grpc-gateway、go-zero、kitex、hertz等
    webFramework: grpc
    # Web Server 后端使用的存储类型。当前支持 memory、mysql
    # 未来会支持etcd、redis、sqlite、mongo、postgresql
    storageType: memory
    # 是否添加健康检查接口
    withHealthz: true
    # 是否添加用户默认，开启后，有完整的认证、鉴权流程
    withUser: false
    withPolaris: true
EOF
}

function miniblog_8() {
  cat << EOF > /tmp/miniblog.yaml
scaffold: osbuilder
version: v0.0.1
metadata:
  shortDescription: Please update the short description of the binary file.
  longMessage: Please update the detailed description of the binary file.
  # 当指定deploymentMethod为docker、kubernetes时，构建镜像的地址
  registry: docker.io
  # 选择二进制文件的部署形式。当前近支持systemd。未来会支持docker、kubernetes，会生产Dockerfile、Kubernetes YAML 等资源
  deploymentMethod: systemd
  # 是否使用结构化的 makefile。非结构化功能简单，结构化设计复杂，但扩展能力强
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
    # 未来会支持kratos、grpc-gateway、go-zero、kitex、hertz等
    webFramework: grpc
    # Web Server 后端使用的存储类型。当前支持 memory、mysql
    # 未来会支持etcd、redis、sqlite、mongo、postgresql
    storageType: memory
    # 是否添加健康检查接口
    withHealthz: true
    # 是否添加用户默认，开启后，有完整的认证、鉴权流程
    withUser: true
    withPolaris: false
EOF
}

function miniblog_9() {
  cat << EOF > /tmp/miniblog.yaml
scaffold: osbuilder
version: v0.0.12
metadata:
  shortDescription: Please update the short description of the binary file.
  longMessage: Please update the detailed description of the binary file.
  # 当指定deploymentMethod为docker、kubernetes时，构建镜像的地址
  registry: docker.io
  # 选择二进制文件的部署形式。当前近支持systemd。未来会支持docker、kubernetes，会生产Dockerfile、Kubernetes YAML 等资源
  deploymentMethod: kubernetes
  # 是否使用结构化的 makefile。非结构化功能简单，结构化设计复杂，但扩展能力强
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
    # 未来会支持kratos、grpc-gateway、go-zero、kitex、hertz等
    webFramework: grpc
    # Web Server 后端使用的存储类型。当前支持 memory、mysql
    # 未来会支持etcd、redis、sqlite、mongo、postgresql
    storageType: memory
    # 是否添加健康检查接口
    withHealthz: true
    # 是否添加用户默认，开启后，有完整的认证、鉴权流程
    withUser: false
    withPolaris: false
EOF
}

function miniblog_10() {
  cat << EOF > /tmp/miniblog.yaml
scaffold: osbuilder
version: v0.0.12
metadata:
  shortDescription: Please update the short description of the binary file.
  longMessage: Please update the detailed description of the binary file.
  # 当指定deploymentMethod为docker、kubernetes时，构建镜像的地址
  registry: docker.io
  # 选择二进制文件的部署形式。当前近支持systemd。未来会支持docker、kubernetes，会生产Dockerfile、Kubernetes YAML 等资源
  deploymentMethod: kubernetes
  # 是否使用结构化的 makefile。非结构化功能简单，结构化设计复杂，但扩展能力强
  makefileMode: structured
  # 项目创建者名字，用于生成版权信息
  author: 孔令飞
  # 项目创建者邮箱，用于生成版权信息
  email: colin404@foxmail.com
# osbuilder 支持多种应用类型。当前仅支持 Web 服务类型
# 未来会支持：异步任务 Job 类型、命令行工具类型、声明式API服务器类型
webServers:
  - binaryName: mb-apiserver
    # Web Server 使用的框架。当前支持 gin、grpc
    # 未来会支持kratos、grpc-gateway、go-zero、kitex、hertz等
    webFramework: grpc
    # Web Server 后端使用的存储类型。当前支持 memory、mysql
    # 未来会支持etcd、redis、sqlite、mongo、postgresql
    storageType: memory
    # 是否添加健康检查接口
    withHealthz: true
    # 是否添加用户默认，开启后，有完整的认证、鉴权流程
    withUser: false
    withPolaris: false
EOF
}

test miniblog_1 && echo -e "\033[32m miniblog_1 success\033[0m" || echo -e "\033[31m miniblog_1 failed\033[0m"
test miniblog_2 && echo -e "\033[32m miniblog_2 success\033[0m" || echo -e "\033[31m miniblog_2 failed\033[0m"
test miniblog_3 && echo -e "\033[32m miniblog_3 success\033[0m" || echo -e "\033[31m miniblog_3 failed\033[0m"
test miniblog_4 && echo -e "\033[32m miniblog_4 success\033[0m" || echo -e "\033[31m miniblog_4 failed\033[0m"
test miniblog_5 && echo -e "\033[32m miniblog_5 success\033[0m" || echo -e "\033[31m miniblog_5 failed\033[0m"
test miniblog_6 && echo -e "\033[32m miniblog_6 success\033[0m" || echo -e "\033[31m miniblog_6 failed\033[0m"
test miniblog_7 && echo -e "\033[32m miniblog_7 success\033[0m" || echo -e "\033[31m miniblog_7 failed\033[0m"
test miniblog_8 && echo -e "\033[32m miniblog_8 success\033[0m" || echo -e "\033[31m miniblog_8 failed\033[0m"
test miniblog_9 && echo -e "\033[32m miniblog_9 success\033[0m" || echo -e "\033[31m miniblog_9 failed\033[0m"
test miniblog_10 && echo -e "\033[32m miniblog_10 success\033[0m" || echo -e "\033[31m miniblog_10 failed\033[0m"
