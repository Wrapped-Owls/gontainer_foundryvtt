WORKSPACE_MODULES := libs/foundrykit libs/foundryacquire libs/foundrypatch libs/foundryruntime apps/foundryctl

.PHONY: all vet test fmt tidy tidy-tests test-modules workspace-modules clean nix-image docker-image

all: vet test

vet:
	@for d in $(WORKSPACE_MODULES); do \
	  (cd $$d && go vet ./...) || exit 1; \
	done

test:
	@for d in $(WORKSPACE_MODULES); do \
	  (cd $$d && go test ./...) || exit 1; \
	done

fmt:
	golines --base-formatter=gofumpt -w .

tidy:
	@for d in $(WORKSPACE_MODULES); do \
	  (cd $$d && go mod tidy) || exit 1; \
	done

# Build the Docker image using plain Docker (non-Nix alternative to nix-image).
docker-image:
	docker build -t foundryvtt-docker:dev .
