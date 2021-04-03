CONTROLLER_EXECUTABLE="$(BUILD_DIR)/controller"

ifneq ($(BUILD_INFO_VERSION), unknown)
	CONTROLLER_VERSION ?= $(BUILD_INFO_VERSION)
    VERSION := $(CONTROLLER_VERSION)
else
	CONTROLLER_VERSION ?= 0.0.1
	VERSION := $(CONTROLLER_VERSION)
endif

BUNDLE_IMG ?= controller-bundle:$(CONTROLLER_VERSION)
IMG ?= darkroom-controller:$(CONTROLLER_VERSION)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

ENVTEST_ASSETS_DIR = $(shell pwd)/testbin
operator/manager/test: operator/generate fmt vet manifests ## Run manager tests
	@mkdir -p $(ENVTEST_ASSETS_DIR)
	@test -f $(ENVTEST_ASSETS_DIR)/setup-envtest.sh || curl -sSLo $(ENVTEST_ASSETS_DIR)/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.6.3/hack/setup-envtest.sh
	@source $(ENVTEST_ASSETS_DIR)/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -coverprofile cover.out

operator/manager/build: operator/generate ## Build manager binary
	@$(GO_BUILD) -o $(CONTROLLER_EXECUTABLE) cmd/operator/main.go

operator/install: operator/manifests kustomize ## Install CRDs into a cluster
	@$(KUSTOMIZE) build config/crd | kubectl apply -f -

operator/uninstall: operator/manifests kustomize ## Uninstall CRDs from a cluster
	@$(KUSTOMIZE) build config/crd | kubectl delete -f -

operator/deploy: operator/manifests kustomize ## Deploy controller in the configured Kubernetes cluster in ~/.kube/config
	@cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	@$(KUSTOMIZE) build config/default | kubectl apply -f -

operator/undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	@$(KUSTOMIZE) build config/default | kubectl delete -f -

operator/manifests: controller-gen ## Generate manifests e.g. CRD, RBAC etc.
	@$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

operator/generate: controller-gen ## Generate operator code
	@$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

operator/docker-build: test ## Build the docker image
	@docker build -f build/package/operator.Dockerfile -t ${IMG} .

operator/docker-push: ## Push the docker image
	@docker push ${IMG}

.PHONY: bundle
bundle: operator/manifests ## Generate bundle manifests and metadata, then validate generated files
	@ln -s pkg/api api && operator-sdk generate kustomize manifests -q && rm api
	@cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	@$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(CONTROLLER_VERSION) $(BUNDLE_METADATA_OPTS)
	@operator-sdk bundle validate ./bundle

.PHONY: bundle-build
bundle-build: ## Build the bundle image
	@docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: install
install: operator/install ## Install all requites to the cluster

CONTROLLER_GEN = $(shell pwd)/.bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1)

KUSTOMIZE = $(shell pwd)/.bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)
