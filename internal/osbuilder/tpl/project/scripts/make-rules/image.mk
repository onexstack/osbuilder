# ==============================================================================
# Makefile helper functions for docker image
#

DOCKER := docker
DOCKER_SUPPORTED_API_VERSION ?= 1.32
DOCKERFILE_DIR=$(PROJ_ROOT_DIR)/build/docker
REGISTRY_PREFIX ?= {{.Metadata.Registry}}
# Set to 1 to use Dockerfile.local when building images
LOCAL_DOCKERFILE ?= 0
EXTRA_ARGS ?= --no-cache
_DOCKER_BUILD_EXTRA_ARGS :=

ifdef HTTP_PROXY
_DOCKER_BUILD_EXTRA_ARGS += --build-arg HTTP_PROXY=${HTTP_PROXY}
endif

ifneq ($(EXTRA_ARGS), )
_DOCKER_BUILD_EXTRA_ARGS += $(EXTRA_ARGS)
endif

# Determine images names by stripping out the dir names.
# Filter out directories without Go files, as these directories cannot be compiled to build a docker image.
IMAGES ?= $(filter-out tools, $(foreach dir, $(CMD_DIRS), $(notdir $(if $(wildcard $(dir)/*.go), $(dir),))))
ifeq (${IMAGES},)
  $(error Could not determine IMAGES, set PROJ_ROOT_DIR or run in source dir)
endif

.PHONY: image.verify
image.verify: ## Verify docker version.
	$(eval API_VERSION := $(shell $(DOCKER) version | grep -E 'API version: {1,6}[0-9]' | head -n1 | awk '{print $$3} END { if (NR==0) print 0}' ))
	$(eval PASS := $(shell echo "$(API_VERSION) > $(DOCKER_SUPPORTED_API_VERSION)" | bc))
	@if [ $(PASS) -ne 1 ]; then \
		$(DOCKER) -v ;\
		echo "Unsupported docker version. Docker API version should be greater than $(DOCKER_SUPPORTED_API_VERSION)"; \
		exit 1; \
	fi

.PHONY: image.daemon.verify
image.daemon.verify: ## Verify docker daemon version.
	$(eval PASS := $(shell $(DOCKER) version | grep -q -E 'Experimental: {1,5}true' && echo 1 || echo 0))
	@if [ $(PASS) -ne 1 ]; then \
		echo "Experimental features of Docker daemon is not enabled. Please add \"experimental\": true in '/etc/docker/daemon.json' and then restart Docker daemon."; \
		exit 1; \
	fi

.PHONY: image.build
image.build: image.verify go.build.verify $(addprefix image.build., $(addprefix $(IMAGE_PLAT)., $(IMAGES))) ## Build all docker images for the selected platform.

.PHONY: image.build.%
image.build.%: go.build.% ## 构建指定的 Docker 镜像
	$(eval IMAGE := $(word 2,$(subst ., ,$*)))
	$(eval IMAGE_TAG := $(subst +,-,$(VERSION)))
	@echo "===========> Building docker image $(IMAGE) $(IMAGE_TAG) for $(IMAGE_PLAT)"
	$(eval PLATFORM := $(word 1,$(subst ., ,$*)))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
	$(eval DOCKERFILE := Dockerfile)
ifeq ($(LOCAL_DOCKERFILE),1)
	$(eval DOCKERFILE := Dockerfile.local)
endif
	@docker build \
		--build-arg OS=$(OS) \
		--build-arg ARCH=$(ARCH) \
		--build-arg GOPROXY=$($(GO) env GOPROXY) \
		--file $(DOCKERFILE_DIR)/$(IMAGE)/$(DOCKERFILE) \
		--tag $(REGISTRY_PREFIX)/$(IMAGE):$(IMAGE_TAG) \
		$(PROJ_ROOT_DIR)

.PHONY: image.push
image.push: image.verify go.build.verify $(addprefix image.push., $(addprefix $(IMAGE_PLAT)., $(IMAGES))) ## Build and push all docker images to docker registry.

.PHONY: image.push.%
image.push.%: image.build.% ## Build and push specified docker image.
	@echo "===========> Pushing image $(IMAGE) $(IMAGE_TAG) to $(REGISTRY)"
	@$(DOCKER) push $(REGISTRY_PREFIX)/$(IMAGE):$(IMAGE_TAG)
