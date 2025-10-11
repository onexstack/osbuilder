# syntax=docker/dockerfile:1.7

# 构建参数（可在 CI 中覆盖）
ARG RUNTIME_IMAGE=debian:bookworm
ARG USER=noroot
ARG UID=1001
ARG GID=1001

FROM ${RUNTIME_IMAGE} AS runtime

ARG USER
ARG UID
ARG GID
ARG OS
ARG ARCH

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
COPY --chown=${UID}:${GID} _output/platforms/${OS}/${ARCH}/{{.Web.BinaryName}} /app/{{.Web.BinaryName}}

# 安全性：以非 root 用户运行
USER ${UID}:${GID}

# 以 tini 作为最小 init，正确处理信号与僵尸进程
ENTRYPOINT ["/usr/bin/tini", "--", "/app/{{.Web.BinaryName}}"]
# 如需默认参数（可在 docker run 或 K8s args 覆盖），可追加 CMD：
# CMD ["--port=9091"]
