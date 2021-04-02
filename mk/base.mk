PROJECT_NAME := "darkroom-operator"

ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

KIND_BIN = $(shell pwd)/.bin/kind
kind: ## Download kind locally if necessary.
	$(call go-get-tool,$(KIND_BIN),sigs.k8s.io/kind@v0.10.0)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))/..
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/.bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
