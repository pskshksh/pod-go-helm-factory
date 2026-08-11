# Makefile — developer commands. Recipe lines must be indented with a TAB.
#
# Common overrides:  make deploy ENV=prod TAG=$(git rev-parse --short HEAD)
#                    make render ENV=staging

DIST  ?= dist
ENV   ?= staging
CHART ?= sampleapp
IMAGE ?= ghcr.io/pskshksh/pod-go-helm-factory/sampleapp
TAG   ?= dev

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*## ' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

# --- Build & test ---

.PHONY: build
build: ## Compile all packages
	go build ./...

.PHONY: test
test: ## Run the full test suite
	go test ./...

.PHONY: update
update: ## Regenerate golden snapshots and fragments
	go test ./charts -run TestGenerateSnapshots -update

.PHONY: fmt
fmt: ## Format all Go code
	gofmt -w .

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: check
check: ## Format check + vet + test (the local CI gate)
	@test -z "$$(gofmt -l .)" || { echo "unformatted:"; gofmt -l .; exit 1; }
	go vet ./...
	go test ./...

# --- Charts ---

.PHONY: generate
generate: ## Generate charts into $(DIST)/
	go run ./cmd/factory

.PHONY: lint
lint: generate ## Generate, then helm lint the chart ($(CHART))
	helm lint $(DIST)/$(CHART)

.PHONY: render
render: ## Render a deploy without a cluster (ENV=$(ENV))
	go run ./cmd/deploy --env $(ENV) --render

# --- App & deploy ---

.PHONY: run
run: ## Run the sample app locally on :8080
	go run ./cmd/sampleapp

.PHONY: deploy
deploy: ## Deploy an environment (ENV=$(ENV) TAG=$(TAG))
	go run ./cmd/deploy --env $(ENV) --tag $(TAG)

.PHONY: image
image: ## Build the sample app image ($(IMAGE):$(TAG))
	docker build -f cmd/sampleapp/Dockerfile -t $(IMAGE):$(TAG) .

.PHONY: clean
clean: ## Remove generated output
	rm -rf $(DIST) bin
