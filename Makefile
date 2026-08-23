WORKSPACE_MODULES := libs/foundrykit libs/fourcery libs/foundrypatch libs/foundryruntime apps/foundrymanager apps/taverncord apps/foundryctl

IMAGE_TAG ?= foundryvtt-docker:dev

.PHONY: all vet test race fmt tidy image nix-hash

all: vet test

vet:
	@for d in $(WORKSPACE_MODULES); do \
	  (cd $$d && go vet ./...) || exit 1; \
	done

test:
	@for d in $(WORKSPACE_MODULES); do \
	  (cd $$d && go test ./...) || exit 1; \
	done

race:
	@for d in $(WORKSPACE_MODULES); do \
	  (cd $$d && go test -race ./...) || exit 1; \
	done

fmt:
	golines --base-formatter=gofumpt -w .

tidy:
	@for d in $(WORKSPACE_MODULES); do \
	  (cd $$d && go mod tidy) || exit 1; \
	done

image:
	docker build -f Containerfile -t $(IMAGE_TAG) .

# Recompute vendorHash after go.mod / go.sum changes.
nix-hash:
	nix run .#update-vendor-hash
