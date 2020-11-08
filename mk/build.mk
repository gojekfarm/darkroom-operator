BUILD_DIR := ".out"

BUILD_INFO_GIT_TAG ?= $(shell git describe --tags 2>/dev/null || echo unknown)
BUILD_INFO_GIT_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
BUILD_INFO_BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" || echo unknown)
BUILD_INFO_VERSION ?= $(shell prefix=$$(echo $(BUILD_INFO_GIT_TAG) | cut -c 1); if [ "$${prefix}" = "v" ]; then echo $(BUILD_INFO_GIT_TAG) | cut -c 2- ; else echo $(BUILD_INFO_GIT_TAG) ; fi)

build_info_fields := \
	version=$(BUILD_INFO_VERSION) \
	gitTag=$(BUILD_INFO_GIT_TAG) \
	gitCommit=$(BUILD_INFO_GIT_COMMIT) \
	buildDate=$(BUILD_INFO_BUILD_DATE)
build_info_ld_flags := $(foreach entry,$(build_info_fields),-X github.com/gojekfarm/darkroom-operator/internal/version.$(entry))

LD_FLAGS := -ldflags="-s -w $(build_info_ld_flags)"
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GO_BUILD := GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go build $(LD_FLAGS)
GO_RUN := GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go run $(LD_FLAGS)
GOFLAGS :=

.PHONY: build
build: operator/manager/build ## Dev: Build all binaries

.PHONY: generate
generate: operator/generate operator/manifests ## Dev: Generate required code and manifests

.PHONY: clean
clean: clean/build ## Dev: Clean

.PHONY: clean/build
clean/build: ## Dev: Remove .out/ dir
	@rm -rf $(BUILD_DIR)


KUBEBUILDER := $(shell command -v /usr/local/kubebuilder/bin/kubebuilder 2> /dev/null)

# download kubebuilder if necessary
kubebuilder:
ifndef KUBEBUILDER
	@bin/install-kubebuilder
else
KUBEBUILDER := /usr/local/kubebuilder/bin/kubebuilder
endif
