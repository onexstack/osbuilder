# syntax=docker/dockerfile:1.7

# 0) 构建参数（可在 CI 中覆盖）
ARG BUILDER_IMAGE=golang:1.25.0
ARG RUNTIME_IMAGE=debian:bookworm
ARG USER=noroot
ARG UID=65532
ARG GID=65532

# 1) Builder 阶段
FROM ${BUILDER_IMAGE} AS builder

# 可选：安装构建期工具（按需）
# RUN apt-get update && apt-get install -y --no-install-recommends upx && rm -rf /var/lib/apt/lists/*

# 工作目录
WORKDIR /workspace

RUN wget -O /tmp/tini https://github.com/krallin/tini/releases/download/v0.19.0/tini-static-amd64 \
 && chmod +x /tmp/tini

# 再复制源码
COPY . .

# Go 构建参数（如你的项目需要静态编译，设置 CGO_ENABLED=1）
ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64

# 构建（使用 make 产出可执行文件）
RUN make build BINS={{.Web.BinaryName}}

# 2) Runtime 阶段
FROM ${RUNTIME_IMAGE} AS runtime

# 应用目录
WORKDIR /app

# 安装运行期必要组件
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates tzdata tini wget curl telnet \
 && rm -rf /var/lib/apt/lists/*

# 安全：创建非 root 用户（若已存在则忽略错误）
# 使用 || true 避免重复创建导致的构建失败（例如基础镜像已有同名组/用户）
RUN groupadd -g ${GID} ${USER} 2>/dev/null || true \
 && useradd -D -u ${UID} -G ${USER} ${USER} 2>/dev/null || true

# 复制产物并设置属主，使用数字 UID:GID 更稳妥
COPY --from=builder /tmp/tini /usr/bin/tini
COPY --from=builder --chown=${UID}:${GID} /workspace/_output/platforms/linux/amd64/{{.Web.BinaryName}} /app/{{.Web.BinaryName}}

# 安全性：以非 root 用户运行
USER ${UID}:${GID}

# 以 tini 作为最小 init，正确处理信号与僵尸进程
ENTRYPOINT ["/usr/bin/tini", "--", "/app/{{.Web.BinaryName}}"]
# 如需默认参数（可在 docker run 或 K8s args 覆盖），可追加 CMD：
# CMD ["--port=9091"]
