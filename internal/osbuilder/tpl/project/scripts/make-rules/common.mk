# ==============================================================================
# 定义全局 Makefile 变量方便后面引用

SHELL := /bin/bash

COMMON_SELF_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
# 项目根目录
PROJ_ROOT_DIR := $(strip $(abspath $(shell cd $(COMMON_SELF_DIR)/../../ && pwd -P)))
# 构建产物、临时文件存放目录
OUTPUT_DIR := $(PROJ_ROOT_DIR)/_output

# 定义包名
ROOT_PACKAGE={{.D.ModuleName}}

# Protobuf 文件存放路径
APIROOT=$(PROJ_ROOT_DIR)/pkg/api

# ==============================================================================
# 定义版本相关变量

## 指定应用使用的 version 包，会通过 `-ldflags -X` 向该包中指定的变量注入值
VERSION_PACKAGE=github.com/onexstack/onexstack/pkg/version

## Defaults when git is unavailable or fails
DEFAULT_VERSION ?= v0.0.0
DEFAULT_GIT_COMMIT ?= unknown
DEFAULT_GIT_TREE_STATE ?= dirty

## 定义 VERSION 语义化版本号
ifeq ($(origin VERSION), undefined)
VERSION := $(shell git describe --tags --always --match='v*' 2>/dev/null || echo $(DEFAULT_VERSION))
endif

## 检查代码仓库是否是 dirty（默认 dirty；git 失败时也用默认）
GIT_TREE_STATE := $(shell \
	if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then \
		if [ -z "$(git status --porcelain 2>/dev/null)" ]; then echo clean; else echo dirty; fi; \
	else \
		echo $(DEFAULT_GIT_TREE_STATE); \
	fi \
)

## 提交哈希（git 失败时取默认）
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo $(DEFAULT_GIT_COMMIT))

GO_LDFLAGS += \
	-X $(VERSION_PACKAGE).gitVersion=$(VERSION) \
	-X $(VERSION_PACKAGE).gitCommit=$(GIT_COMMIT) \
	-X $(VERSION_PACKAGE).gitTreeState=$(GIT_TREE_STATE) \
	-X $(VERSION_PACKAGE).buildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# 编译的操作系统可以是 linux/windows/darwin
PLATFORMS ?= darwin_amd64 windows_amd64 linux_amd64 linux_arm64

# 设置一个指定的操作系统
ifeq ($(origin PLATFORM), undefined)
	ifeq ($(origin GOOS), undefined)
		GOOS := $(shell go env GOOS)
	endif
	ifeq ($(origin GOARCH), undefined)
		GOARCH := $(shell go env GOARCH)
	endif
	PLATFORM := $(GOOS)_$(GOARCH)
	# 构建镜像时，使用 linux 作为默认的 OS
	IMAGE_PLAT := linux_$(GOARCH)
else
	GOOS := $(word 1, $(subst _, ,$(PLATFORM)))
	GOARCH := $(word 2, $(subst _, ,$(PLATFORM)))
	IMAGE_PLAT := $(PLATFORM)
endif

# 设置单元测试覆盖率阈值
ifeq ($(origin COVERAGE),undefined)
COVERAGE := 1
endif

# Makefile 设置
ifndef V
MAKEFLAGS += --no-print-directory
endif

# Linux 命令设置
FIND := find . ! -path './third_party/*' ! -path './vendor/*'
XARGS := xargs --no-run-if-empty
