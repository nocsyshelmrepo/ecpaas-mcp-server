
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
# output
OUTPUT_DIR := $(abspath $(ROOT_DIR)/_output)
OUTPUT_BIN_DIR := $(OUTPUT_DIR)/bin
OUTPUT_TOOLS_DIR := $(OUTPUT_DIR)/tools
#ARTIFACTS ?= ${OUTPUT_DIR}/_artifacts

dirs := $(OUTPUT_DIR) $(OUTPUT_BIN_DIR) $(OUTPUT_TOOLS_DIR)

$(foreach dir, $(dirs), \
  $(if $(shell [ -d $(dir) ] && echo 1 || echo 0),, \
    $(shell mkdir -p $(dir)) \
  ) \
)

export PATH := $(abspath $(OUTPUT_BIN_DIR)):$(abspath $(OUTPUT_TOOLS_DIR)):$(PATH)

#
# Binaries.
#
# Note: Need to use abspath so we can invoke these from subdirectories
GO_INSTALL := ./hack/go_install.sh

GORELEASER_VER := $(shell cat .github/workflows/releaser.yaml | grep [[:space:]]version | sed 's/.*version: //')
GORELEASER_BIN := goreleaser
GORELEASER := $(abspath $(OUTPUT_TOOLS_DIR)/$(GORELEASER_BIN)-$(GORELEASER_VER))
GORELEASER_PKG := github.com/goreleaser/goreleaser/v2

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[0-9A-Za-z_-]+:.*?##/ { printf "  \033[36m%-45s\033[0m %s\n", $$1, $$2 } /^\$$\([0-9A-Za-z_-]+\):.*?##/ { gsub("_","-", $$1); printf "  \033[36m%-45s\033[0m %s\n", tolower(substr($$1, 3, length($$1)-7)), $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

$(GORELEASER): # Build goreleaser into tools folder.
	@if [ ! -f $(GORELEASER) ]; then \
		CGO_ENABLED=0 GOBIN=$(OUTPUT_TOOLS_DIR) $(GO_INSTALL) $(GORELEASER_PKG) $(GORELEASER_BIN) $(GORELEASER_VER); \
	fi

.PHONY: verify-releaser
verify-releaser: $(GORELEASER) ## Verify goreleaser
	@$(GORELEASER) check

.PHONY: releaser 
releaser: $(GORELEASER) ## build releaser in dist. it will show in https://github.com/kubesphere/ks-mcp-server/releases
	@LDFLAGS=$(bash ./hack/version.sh) $(GORELEASER) release --clean --skip validate --skip publish

