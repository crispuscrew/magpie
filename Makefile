.DEFAULT_GOAL := help
ENGINE := $(shell command -v podman 2>/dev/null || command -v docker 2>/dev/null)
IMAGE  ?= localhost/magpie-bench
# rootless podman behind VPN DNS-leak protection: DNS_FLAGS=--dns=1.1.1.1 (see tech/containers.md)
DNS_FLAGS ?=
ifneq (,$(findstring podman,$(ENGINE)))
ENGINE_FLAGS := --userns=keep-id
endif
# expands in recipes only, so `make help` works without an engine
ENGINE_REQ = $(if $(ENGINE),$(ENGINE),$(error no container engine found: install podman or docker))
RUN_RO = $(ENGINE_REQ) run --rm $(ENGINE_FLAGS) -v $(CURDIR):/work:ro,Z -w /work $(IMAGE)
RUN_RW = $(ENGINE_REQ) run --rm $(ENGINE_FLAGS) -v $(CURDIR):/work:Z -w /work $(IMAGE)

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) \
	  | awk 'BEGIN{FS=":.*?## "}{printf "  %-16s %s\n", $$1, $$2}'

build: ## Build the container image (tests, lint, task checks)
	$(ENGINE_REQ) build $(DNS_FLAGS) -t $(IMAGE) .

test: ## Harness + fixture tests, in the container
	$(RUN_RO) sh -c 'go test ./... && cd benchmarks/fixture && go test ./...'

lint: ## gofmt + go vet, in the container
	$(RUN_RO) sh -c 'test -z "$$(gofmt -l .)" && go vet ./... && cd benchmarks/fixture && go vet ./...'

adapters: ## Regenerate host adapters from AGENTS.md
	$(RUN_RW) go run ./cmd/adapters

adapters-check: ## Fail if adapters drifted from AGENTS.md
	$(RUN_RO) go run ./cmd/adapters --check

check: test lint adapters-check ## All gates (what CI runs)

# host-run: drives the host claude CLI (~/.claude auth) and ollama socket;
# correctness checks still execute inside the container
bench: build ## Run the benchmark; pass flags via ARGS="-backend ollama ..."
	go run ./cmd/bench -image $(IMAGE) -engine-flags "$(DNS_FLAGS)" $(ARGS)

.PHONY: help build test lint adapters adapters-check check bench
