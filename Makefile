# wowapi — development & CI entrypoints. Human- and CI-usable (Goal 2 §Makefile).
# Container-first: `make up` starts local infra; `make shell` gives a toolbox
# with the repo mounted; every test target also runs inside that container.

COMPOSE := docker compose -f deployments/compose.yaml
GO      ?= go
PKGS    := ./...

.DEFAULT_GOAL := help

##@ General

.PHONY: help
help: ## List targets
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

.PHONY: setup
setup: tools ## One-time developer setup (tool install + go mod download)
	$(GO) mod download

.PHONY: tools
tools: ## Install host dev tools (golangci-lint; more per phase)
	@command -v golangci-lint >/dev/null 2>&1 || \
		$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

##@ Local infrastructure (containers)

.PHONY: up
up: ## Start postgres + minio + mailpit + tools runner
	$(COMPOSE) up -d --wait

.PHONY: down
down: ## Stop local infra (keep volumes)
	$(COMPOSE) down

.PHONY: reset
reset: ## Stop local infra and DELETE volumes
	$(COMPOSE) down -v

.PHONY: logs
logs: ## Tail infra logs
	$(COMPOSE) logs -f

.PHONY: shell
shell: ## Shell in the containerized toolbox (repo mounted at /src)
	$(COMPOSE) run --rm tools sh

.PHONY: db-shell
db-shell: ## psql into the local postgres
	$(COMPOSE) exec postgres psql -U wowapi -d wowapi

##@ Database

# Local default DSN matches deployments/compose.yaml; CI containers get
# DATABASE_URL injected by the compose tools service.
TEST_DSN ?= postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable

.PHONY: migrate
migrate: ## Apply kernel migrations to the local compose database
	DATABASE_URL="$${DATABASE_URL:-$(TEST_DSN)}" $(GO) run ./internal/tools/migrate

.PHONY: seed
seed:
	@echo "make seed: available in Phase 5 (seed loader) — see docs/implementation/phase-plan.md" >&2; exit 2

##@ Quality

.PHONY: fmt
fmt: ## gofmt all Go files
	$(GO) fmt $(PKGS)

.PHONY: lint
lint: ## golangci-lint (falls back to go vet)
	@if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run; else echo "golangci-lint not installed; running go vet"; $(GO) vet $(PKGS); fi

.PHONY: lint-boundaries
lint-boundaries: ## Import-law + vocabulary + Reveal() boundary lint
	sh scripts/lint_boundaries.sh

##@ Tests

.PHONY: test
test: test-unit ## All currently available test suites

.PHONY: test-unit
test-unit: ## Unit tests (no external services)
	$(GO) test $(PKGS)

.PHONY: test-race
test-race: ## Unit tests with the race detector
	$(GO) test -race $(PKGS)

.PHONY: test-integration
test-integration: ## Integration tests against real Postgres (needs `make up` or DATABASE_URL)
	DATABASE_URL="$${DATABASE_URL:-$(TEST_DSN)}" $(GO) test -run 'Integration' -count=1 ./testkit/... ./kernel/database/...

.PHONY: test-contract test-security
test-contract test-security:
	@echo "make $@: available from Phase 4/5 (needs authz + module SDK) — see docs/implementation/phase-plan.md" >&2; exit 2

.PHONY: bench
bench: ## Benchmarks (budget gates arrive in Phase 11)
	$(GO) test -bench=. -benchmem -run=^$$ $(PKGS)

.PHONY: coverage
coverage: ## Unit coverage report
	$(GO) test -coverprofile=coverage.out $(PKGS) && $(GO) tool cover -func=coverage.out | tail -1

##@ Generators & CLI (delivered in Phase 10)

.PHONY: gen new-module gen-crud openapi config-validate config-doctor
gen:
	@$(GO) run ./cmd/wowapi gen
new-module:
	@$(GO) run ./cmd/wowapi new-module $(name)
gen-crud:
	@$(GO) run ./cmd/wowapi gen crud --module $(module) --resource $(resource)
openapi:
	@$(GO) run ./cmd/wowapi openapi merge
config-validate:
	@$(GO) run ./cmd/wowapi config validate
config-doctor:
	@$(GO) run ./cmd/wowapi config doctor

##@ Graphify

.PHONY: graph-check
graph-check: ## Graphify freshness check
	sh scripts/graphify_refresh.sh check

.PHONY: graph-update
graph-update: ## Graphify incremental update
	sh scripts/graphify_refresh.sh update

##@ CI

.PHONY: build
build: ## Build all packages and the CLI
	$(GO) build $(PKGS)
	$(GO) build -o bin/wowapi ./cmd/wowapi

.PHONY: ci
ci: ## Full local CI: vet+lint, boundaries, unit, race, build
	$(GO) vet $(PKGS)
	$(MAKE) lint-boundaries
	$(MAKE) test-unit
	$(MAKE) test-race
	$(MAKE) build

.PHONY: ci-container
ci-container: ## Run `make ci` inside the toolbox container
	$(COMPOSE) run --rm tools make ci
