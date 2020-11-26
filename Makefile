.PHONY: help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z0-9_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

include mk/base.mk
include mk/dev.mk
include mk/build.mk
include mk/check.mk
include mk/operator.mk
include mk/apiserver.mk
include mk/run.mk
include mk/test.mk
