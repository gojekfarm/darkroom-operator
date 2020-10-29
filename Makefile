# ====== Variables =======
PROJECT_NAME := "darkroom-operator"
BUILD_DIR := "./.out"
APP_EXECUTABLE="$(BUILD_DIR)/$(PROJECT_NAME)"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

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

# ====== Help =======
.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECT_NAME)":"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## compile: Compile the code to a binary
compile:
	@$(GO_BUILD) -o $(APP_EXECUTABLE) cmd/operator/main.go

lint: golint
	@$(GOLINT) ./... | { grep -vwE "exported (var|function|method|type|const) \S+ should have comment" || true; }

format:
	go fmt ./...

vet:
	go vet ./...

## test: Run all the tests
test:
	go test ./... -covermode=count -coverprofile=covprofile

coverage: goveralls
	@$(GOVERALLS) -coverprofile=covprofile -service=github

test-ci: compile lint format vet test

# ========= Helpers ===========

# find or download goveralls
goveralls:
ifeq (, $(shell which goveralls))
	@{ \
	cd .. ;\
	go get github.com/mattn/goveralls ;\
	cd - ;\
	}
GOVERALLS=$(GOBIN)/goveralls
else
GOVERALLS=$(shell which goveralls)
endif

# find or download golint
golint:
ifeq (, $(shell which golint))
	@{ \
	cd .. ;\
	go get golang.org/x/lint/golint ;\
	cd - ;\
	}
GOLINT=$(GOBIN)/golint
else
GOLINT=$(shell which golint)
endif
