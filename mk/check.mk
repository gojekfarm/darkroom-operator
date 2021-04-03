##@ Utils

.PHONY: fmt
fmt: fmt/go ## Run format tools

.PHONY: fmt/go
fmt/go: ## Run go fmt
	@go fmt $(GOFLAGS) ./...

.PHONY: vet
vet: ## Run go vet
	@go vet $(GOFLAGS) ./...

.PHONY: lint
lint: require/golangci-lint ## Runs all linters
	@$(GOLANGCI_LINT) run --timeout=10m -v

.PHONY: imports
imports: require/goimports ## Runs goimports in order to organize imports
	@$(GOIMPORTS) -w -local github.com/gojekfarm/darkroom-operator -d `find . -type f -name '*.go'`

.PHONY: check
check: operator/generate operator/manifests fmt vet lint imports ## Run code checks (go fmt, go vet, ...)
	@git diff --quiet || test $$(git diff --name-only | grep -v -e 'go.mod$$' -e 'go.sum$$' | wc -l) -eq 0 || ( echo "The following changes (result of code generators and code checks) have been detected:" && git --no-pager diff && false ) # fail if Git working tree is dirty

# find or download golangci-lint
require/golangci-lint:
ifeq (, $(shell which golangci-lint))
	@{ \
	set -e ;\
	GOLANGCI_LINT_TMP_DIR=$$(mktemp -d) ;\
	cd $$GOLANGCI_LINT_TMP_DIR ;\
	go mod init tmp ;\
	go get github.com/golangci/golangci-lint/cmd/golangci-lint ;\
	rm -rf $$GOLANGCI_LINT_TMP_DIR ;\
	}
GOLANGCI_LINT=$(GOBIN)/golangci-lint
else
GOLANGCI_LINT=$(shell which golangci-lint)
endif

# find or download goimports
require/goimports:
ifeq (, $(shell which goimports))
	@{ \
	set -e ;\
	GOIMPORTS_TMP_DIR=$$(mktemp -d) ;\
	cd $$GOIMPORTS_TMP_DIR ;\
	go mod init tmp ;\
	go get golang.org/x/tools/cmd/goimports ;\
	rm -rf $$GOIMPORTS_TMP_DIR ;\
	}
GOIMPORTS=$(GOBIN)/goimports
else
GOIMPORTS=$(shell which goimports)
endif
