##@ Development

create-cluster: require/kind ## Create a kind cluster to spin up a dev cluster
	@$(KIND_BIN) create cluster
	@operator-sdk olm install --version 0.17.0

delete-cluster: require/kind ## Delete the kind cluster
	@$(KIND_BIN) delete cluster

load-controller: require/kind ## Load a local container image into the kind cluster
	@$(KIND_BIN) load docker-image ${IMG}
