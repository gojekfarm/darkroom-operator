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

LD_FLAGS := -ldflags="-s"
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
