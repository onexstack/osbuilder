# ==============================================================================
# Define global Makefile variables for easy reference later

COMMON_SELF_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
# Project root directory
PROJ_ROOT_DIR := $(strip $(abspath $(shell cd $(COMMON_SELF_DIR)/ && pwd -P)))
# Directory for build artifacts and temporary files
OUTPUT_DIR := $(PROJ_ROOT_DIR)/_output
ROOT_PACKAGE=github.com/onexstack/osbuilder

# ==============================================================================
# Define version-related variables

## Specify the version package used by the application. Values will be injected into variables in this package using `-ldflags -X`.
VERSION_PACKAGE=github.com/onexstack/onexstack/pkg/version

## Check if current directory is a Git repository
IS_GIT_REPO := $(shell git rev-parse --is-inside-work-tree 2>/dev/null)
ifeq ($(IS_GIT_REPO),)
    # Non-Git repository, use fallback values
    GIT_TREE_STATE := "not_a_git_repo"
    VERSION := "v0.0.0"
    GIT_COMMIT := "unknown"
else
    # Define the VERSION semantic version (if not already set)
    ifeq ($(origin VERSION), undefined)
        VERSION := $(shell git describe --tags --abbrev=0 --match='v*')
    endif

    # Check if the code repository is in a dirty state (default is dirty)
    GIT_TREE_STATE := "dirty"
    ifeq (, $(shell git status --porcelain 2>/dev/null))
        GIT_TREE_STATE := "clean"
    endif

    # Get current Git commit hash
    GIT_COMMIT := $(shell git rev-parse HEAD)
endif

GO_LDFLAGS += \
    -X $(VERSION_PACKAGE).gitVersion=$(VERSION) \
    -X $(VERSION_PACKAGE).gitCommit=$(GIT_COMMIT) \
    -X $(VERSION_PACKAGE).gitTreeState=$(GIT_TREE_STATE) \
    -X $(VERSION_PACKAGE).buildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ') \
    -X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn

ifeq ($(GOOS),windows)
    GO_OUT_EXT := .exe
endif

# 修复这一行：移除重复的 +=
GO_BUILD_FLAGS += -ldflags "$(GO_LDFLAGS)"
COMMANDS ?= $(filter-out %.md, $(wildcard $(PROJ_ROOT_DIR)/cmd/*))
BINS ?= $(foreach cmd,${COMMANDS},$(notdir $(cmd)))

# 编译的操作系统可以是 linux/windows/darwin
PLATFORMS ?= darwin_amd64 darwin_arm64 windows_amd64 windows_arm64 linux_amd64 linux_arm64

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

# ==============================================================================
# Define the default target as `all`
.DEFAULT_GOAL := all

# Define the Makefile `all` phony target. When `make` is executed, it will default to executing the `all` target.
.PHONY: all
all: tidy format build

# ==============================================================================
# Usage

define USAGE_OPTIONS

Options:
  BINS             The binary files to build. Defaults to all files in the `cmd` directory.
                   This option can be used with the following command: make build
                   Example: make build BINS="<BinaryName>"
  VERSION          Version information to embed into the binary.
  V                Set to 1 to enable verbose build output. Default is 0.
endef
export USAGE_OPTIONS

# ==============================================================================
# Define other required phony targets
#

.PHONY: build.multiarch                          
build.multiarch: $(foreach p,$(PLATFORMS),$(addprefix build., $(addprefix $(p)., $(BINS)))) ## Build all applications with all supported arch.

.PHONY: build
build: $(addprefix build., $(addprefix $(PLATFORM)., $(BINS))) ## Build all binaries for the selected platform.

build.%: ## 编译 Go 源码.
	$(eval COMMAND := $(word 2,$(subst ., ,$*)))
	$(eval PLATFORM := $(word 1,$(subst ., ,$*)))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	@echo "===========> Building binary $(COMMAND) $(VERSION) for $(OS) $(ARCH)"
	@mkdir -p $(OUTPUT_DIR)/platforms/$(OS)/$(ARCH)
	@CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build $(GO_BUILD_FLAGS) \
		-o $(OUTPUT_DIR)/platforms/$(OS)/$(ARCH)/$(COMMAND)$(GO_OUT_EXT) \
		$(ROOT_PACKAGE)/cmd/$(COMMAND)

.PHONY: format
format: ## Format go code
	@find . -type f -name '*.go' -not -path '*/tpl/*' -print0 | xargs -0 gofmt -s -w

.PHONY: add-copyright
add-copyright: ## Add copyright headers (skip third_party, vendor, _output).
	@addlicense -v -f $(PROJ_ROOT_DIR)/scripts/boilerplate.txt $(PROJ_ROOT_DIR) --skip-dirs=third_party,vendor,$(OUTPUT_DIR)

.PHONY: tidy
tidy: ## Sync dependencies and update go.mod/go.sum (go mod tidy).
	@go mod tidy

.PHONY: clean
clean: ## Remove build artifacts and temp files (_output/).
	@-rm -vrf $(OUTPUT_DIR)

.PHONY: release
release: build.multiarch ## Create and publish a GitHub release
	@./scripts/github-release.sh $(VERSION)

.PHONY: release.draft
release.draft: build.multiarch ## Create a draft GitHub release
	@./scripts/github-release.sh $(VERSION) --draft

help: Makefile  ## Show available targets and usage.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<TARGETS> <OPTIONS>\033[0m\n\n\033[35mTargets:\033[0m\n"} /^[0-9A-Za-z._-]+:.*?##/ { printf "  \033[36m%-45s\033[0m %s\n", $$1, $$2 } /^\$$\([0-9A-Za-z_-]+\):.*?##/ { gsub("_","-", $$1); printf "  \033[36m%-45s\033[0m %s\n", tolower(substr($$1, 3, length($$1)-7)), $$2 } /^##@/{ printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' Makefile #$(MAKEFILE_LIST)
	@echo "$$USAGE_OPTIONS"
