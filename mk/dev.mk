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
