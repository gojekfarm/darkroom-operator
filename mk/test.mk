GO_TEST := go test $(GOFLAGS) $(LD_FLAGS)

.PHONY: test-ci
test-ci: check test ## CI: Run tests for all modules and report coverage

.PHONY: test
test: ## Dev: Run tests for all modules
	@$(GO_TEST) -v ./... -covermode=count -coverprofile=covprofile

.PHONY: coverage
coverage: goveralls ## Dev: Send coverage report to coveralls
	@$(GOVERALLS) -coverprofile=covprofile -service=github

# find or download goveralls
goveralls:
ifeq (, $(shell which goveralls))
	@{ \
	set -e ;\
	GOVERALLS_TMP_DIR=$$(mktemp -d) ;\
	cd $$GOVERALLS_TMP_DIR ;\
	go mod init tmp ;\
	go get github.com/mattn/goveralls ;\
	rm -rf $$GOVERALLS_TMP_DIR ;\
	}
GOVERALLS=$(GOBIN)/goveralls
else
GOVERALLS=$(shell which goveralls)
endif
