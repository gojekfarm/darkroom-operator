dev/create-cluster: kind
	@$(KIND_BIN) create cluster
	@operator-sdk olm install

dev/delete-cluster: kind
	@$(KIND_BIN) delete cluster

dev/load-controller: kind
	@$(KIND_BIN) load docker-image ${IMG}

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
