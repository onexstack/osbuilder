# ==============================================================================
# 工具相关的 Makefile
#

TOOLS ?= golangci-lint goimports protoc-plugins addlicense wire addlicense protolint

tools.verify: $(addprefix tools.verify., $(TOOLS))

tools.install: $(addprefix tools.install., $(TOOLS))

tools.install.%:
	@echo "===========> Installing $*"
	@$(MAKE) install.$*

tools.verify.%:
	@if ! which $* &>/dev/null; then $(MAKE) tools.install.$*; fi

install.golangci-lint:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.4.0
	@golangci-lint completion bash > $(HOME)/.golangci-lint.bash
	@if ! grep -q .golangci-lint.bash $(HOME)/.bashrc; then echo "source \$$HOME/.golangci-lint.bash" >> $(HOME)/.bashrc; fi

install.goimports:
	@$(GO) install golang.org/x/tools/cmd/goimports@latest

install.protoc-plugins:
	@$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@$(GO) install github.com/onexstack/protoc-gen-defaults@latest
	@$(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@$(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@$(GO) install github.com/favadi/protoc-go-inject-tag@latest

install.wire:
	@$(GO) install github.com/google/wire/cmd/wire@latest

install.addlicense:
	@$(GO) install github.com/onexstack/addlicense@v0.0.3

install.protolint:
	@$(GO) install github.com/yoheimuta/protolint/cmd/protolint@latest

# 伪目标（防止文件与目标名称冲突）
.PHONY: tools.verify tools.install tools.install.% tools.verify.% install.golangci-lint \
	install.goimports install.protoc-plugins install.wire install.addlicense install.protolint
