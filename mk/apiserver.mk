APISERVER_EXECUTABLE="$(BUILD_DIR)/api-server"

apiserver/build: ## Build api-server binary
	@$(GO_BUILD) -o $(APISERVER_EXECUTABLE) cmd/api-server/main.go
