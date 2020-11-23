operator/manager/run: operator/generate operator/manifests ## Run manager against the configured Kubernetes cluster in ~/.kube/config
	@$(GO_RUN) cmd/operator/main.go

apiserver/run: ## Run api-server against the configured Kubernetes cluster in ~/.kube/config
	@$(GO_RUN) cmd/api-server/main.go
