#!/bin/bash

onexstack_dir=/home/colin/workspace/golang/src/github.com/onexstack

function test() {
  eval "$1"
  cd ${onexstack_dir}/osbuilder/internal/osbuilder/tpl
  ./gen-statik.sh

  cd ${onexstack_dir}/osbuilder/cmd/osbuilder
  go build -v . || exit 1

  cd ${onexstack_dir}
  rm -rf osdemo; /tmp/osbuilder create project --config /tmp/osdemo.yaml ./osdemo

  cd ${onexstack_dir}/osdemo
  make protoc.apiserver
  go get cloud.google.com/go/compute@latest
  go get cloud.google.com/go/compute/metadata@latest
  go mod tidy
  go generate ./...
  if [ "$1" == "osdemo_9" -o "$1" == "osdemo_10" ];then
    make image LOCAL_DOCKERFILE=1
  else
    make build
  fi
}

function osdemo_1() {
  cat << EOF > /tmp/osdemo.yaml
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
  - binaryName: os-apiserver
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

function osdemo_2() {
  cat << EOF > /tmp/osdemo.yaml
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
  - binaryName: os-apiserver
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

function osdemo_2() {
  cat << EOF > /tmp/osdemo.yaml
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
  - binaryName: os-apiserver
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

function osdemo_3() {
  cat << EOF > /tmp/osdemo.yaml
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
  - binaryName: os-apiserver
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

function osdemo_4() {
  cat << EOF > /tmp/osdemo.yaml
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
  - binaryName: os-apiserver
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

function osdemo_5() {
  cat << EOF > /tmp/osdemo.yaml
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
  - binaryName: os-apiserver
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

function osdemo_2() {
  cat << EOF > /tmp/osdemo.yaml
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
  - binaryName: os-apiserver
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

function osdemo_6() {
  cat << EOF > /tmp/osdemo.yaml
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
  - binaryName: os-apiserver
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

function osdemo_7() {
  cat << EOF > /tmp/osdemo.yaml
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
  - binaryName: os-apiserver
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

function osdemo_8() {
  cat << EOF > /tmp/osdemo.yaml
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
  - binaryName: os-apiserver
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

function osdemo_9() {
  cat << EOF > /tmp/osdemo.yaml
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
  - binaryName: os-apiserver
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

function osdemo_10() {
  cat << EOF > /tmp/osdemo.yaml
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
  - binaryName: os-apiserver
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

test osdemo_1 && echo -e "\033[32m osdemo_1 success\033[0m" || echo -e "\033[31m osdemo_1 failed\033[0m"
test osdemo_2 && echo -e "\033[32m osdemo_2 success\033[0m" || echo -e "\033[31m osdemo_2 failed\033[0m"
test osdemo_3 && echo -e "\033[32m osdemo_3 success\033[0m" || echo -e "\033[31m osdemo_3 failed\033[0m"
test osdemo_4 && echo -e "\033[32m osdemo_4 success\033[0m" || echo -e "\033[31m osdemo_4 failed\033[0m"
test osdemo_5 && echo -e "\033[32m osdemo_5 success\033[0m" || echo -e "\033[31m osdemo_5 failed\033[0m"
test osdemo_6 && echo -e "\033[32m osdemo_6 success\033[0m" || echo -e "\033[31m osdemo_6 failed\033[0m"
test osdemo_7 && echo -e "\033[32m osdemo_7 success\033[0m" || echo -e "\033[31m osdemo_7 failed\033[0m"
test osdemo_8 && echo -e "\033[32m osdemo_8 success\033[0m" || echo -e "\033[31m osdemo_8 failed\033[0m"
test osdemo_9 && echo -e "\033[32m osdemo_9 success\033[0m" || echo -e "\033[31m osdemo_9 failed\033[0m"
test osdemo_10 && echo -e "\033[32m osdemo_10 success\033[0m" || echo -e "\033[31m osdemo_10 failed\033[0m"
