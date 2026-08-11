  # Makefile — developer commands. Recipe lines must be indented with a TAB.

DIST ?= dist

.PHONY: generate
generate: ## Generate charts into $(DIST)/
	go run ./cmd/factory

.PHONY: clean
clean: ## Remove generated output
	rm -rf $(DIST) bin