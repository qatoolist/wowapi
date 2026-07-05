# wowapi — development & CI entrypoints. Human- and CI-usable (Goal 2 §Makefile).
# Container-first: `make up` starts local infra; `make shell` gives a toolbox
# with the repo mounted; every test target also runs inside that container.

COMPOSE := docker compose -f deployments/compose.yaml
GO      ?= go
PKGS    := ./...
# Baseline for the "changed code only" lint gate (lint-new). golangci-lint lints
# only issues introduced since the merge-base with this ref, so the large
# pre-existing backlog does not block while all NEW code is fully linted.
LINT_BASE ?= origin/main
# Pinned golangci-lint version — the single source of truth for local `make tools`
# and CI (see .github/workflows/ci.yml GOLANGCI_VERSION). Pinned (not @latest) so
# the enforced full-tree `make lint` gate is deterministic: a new upstream release
# can't fail CI until this is bumped deliberately. Bump in lockstep with CI.
GOLANGCI_VERSION ?= v2.11.4
# actionlint is pinned for the same reason (lockstep with ci.yml ACTIONLINT_VERSION):
# the workflow lint must not fail non-deterministically on a new actionlint release.
ACTIONLINT_VERSION ?= v1.7.12

.DEFAULT_GOAL := help

##@ General

.PHONY: help
help: ## List targets
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

.PHONY: setup
setup: tools hooks ## One-time developer setup (tools + git hooks + go mod download)
	$(GO) mod download

.PHONY: tools
tools: ## Install host dev tools (pinned golangci-lint $(GOLANGCI_VERSION))
	@if ! golangci-lint version 2>/dev/null | grep -q "$(patsubst v%,%,$(GOLANGCI_VERSION))"; then \
		echo "installing golangci-lint $(GOLANGCI_VERSION)"; \
		$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION); \
	fi

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

.PHONY: product-dev
product-dev: ## Build a product ON the framework in a dev box: make product-dev DIR=/path/to/product
	@test -n "$(DIR)" || { echo "usage: make product-dev DIR=/path/to/product-dir"; exit 2; }
	scripts/product-dev.sh "$(DIR)"

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

##@ Operational drills (backup/restore & migration reversibility)

.PHONY: drill-reversibility
drill-reversibility: ## O2/B-4: up->down->up on a scratch DB, DIFF the schema snapshots (fails on asymmetric Down). Needs pg_dump+go+DB.
	DATABASE_URL="$${DATABASE_URL:-$(TEST_DSN)}" scripts/migration_reversibility_drill.sh

.PHONY: drill-restore
drill-restore: ## O5: logical pg_dump -> pg_restore round-trip into a scratch DB. Needs pg_dump/psql+DB.
	SRC_URL="$${DATABASE_URL:-$(TEST_DSN)}" scripts/backup_restore_drill.sh

.PHONY: drill-pitr
drill-pitr: ## O5/B-5: real PITR — base backup + WAL replay to a target time in a throwaway PG container. Needs docker.
	scripts/pitr_restore_drill.sh

.PHONY: drill-object-storage
drill-object-storage: ## O5/B-5: object-storage backup/restore round-trip against compose MinIO. Needs docker + `make up`.
	scripts/object_storage_restore_drill.sh

.PHONY: drills
drills: drill-reversibility drill-restore drill-pitr drill-object-storage ## Run every backup/restore & reversibility drill

.PHONY: smoke-reference
smoke-reference: ## B-7/CA-6: scaffold a product, run it behind the reference nginx (TLS), smoke-test the security headers THROUGH the proxy. Needs go + docker + openssl.
	scripts/smoke_reference_stack.sh

##@ Quality

.PHONY: fmt
fmt: ## Format all Go files (gofumpt + goimports, via golangci-lint)
	@if command -v golangci-lint >/dev/null 2>&1; then golangci-lint fmt; else echo "golangci-lint not installed; running gofmt"; $(GO) fmt $(PKGS); fi

.PHONY: fmt-check
fmt-check: ## Fail if any file needs gofumpt/goimports formatting
	@d=$$(golangci-lint fmt --diff 2>/dev/null); if [ -n "$$d" ]; then echo "$$d"; echo ">> run 'make fmt'"; exit 1; fi

.PHONY: vet
vet: ## go vet
	$(GO) vet $(PKGS)

.PHONY: lint
lint: ## Full golangci-lint across the whole tree (backlog B-1 closed 2026-07-05 — now 0; see docs/working/lint-backlog.md)
	@if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run; else echo "golangci-lint not installed; running go vet"; $(GO) vet $(PKGS); fi

.PHONY: lint-new
lint-new: ## ENFORCED gate: golangci-lint on CHANGED code only (new since LINT_BASE=$(LINT_BASE))
	golangci-lint run --new-from-merge-base=$(LINT_BASE) ./...

.PHONY: lint-boundaries
lint-boundaries: ## Import-law + vocabulary + Reveal() boundary lint
	sh scripts/lint_boundaries.sh

.PHONY: tidy
tidy: ## go mod tidy
	$(GO) mod tidy

.PHONY: tidy-check
tidy-check: ## Fail if go.mod/go.sum are not tidy
	@cp go.mod go.mod.ci.bak; cp go.sum go.sum.ci.bak; $(GO) mod tidy; \
		if ! diff -q go.mod go.mod.ci.bak >/dev/null 2>&1 || ! diff -q go.sum go.sum.ci.bak >/dev/null 2>&1; then \
			mv go.mod.ci.bak go.mod; mv go.sum.ci.bak go.sum; echo ">> go.mod/go.sum not tidy — run 'make tidy'"; exit 1; fi; \
		rm -f go.mod.ci.bak go.sum.ci.bak

.PHONY: check
check: fmt-check vet lint-new tidy-check test-unit ## Fast pre-flight before commit/push (fmt, vet, lint changed code, tidy, unit tests)

.PHONY: hooks
hooks: ## Install the versioned git hooks (pre-commit + pre-push)
	@git config core.hooksPath .githooks && chmod +x .githooks/pre-commit .githooks/pre-push 2>/dev/null; echo "git hooks installed (core.hooksPath=.githooks)"

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
	DATABASE_URL="$${DATABASE_URL:-$(TEST_DSN)}" WOWAPI_REQUIRE_DB=1 $(GO) test -run 'Integration' -count=1 $(PKGS)

.PHONY: test-contract
test-contract: ## Module contract + scratch external-consumer suite (needs DB)
	DATABASE_URL="$${DATABASE_URL:-$(TEST_DSN)}" $(GO) test -run 'Contract|ScratchConsumer' -count=1 ./testkit/...

# Security-critical test suite (criterion #18, #26).
# Covers:
#   RLS/privilege escalation (kernel/authz integration: NoSelf, OverGrant, Scope)
#   Deny-by-default authz (kernel/authz unit: Deny, Authz, Sensitive)
#   Secret redaction in logs/CLI/dumps (kernel/config, kernel/logging, internal/cli: Secret, Redact)
#   Unsafe-config prod rejection (kernel/config: Prod, Unsafe, Security)
#   DSN credential non-echoing (kernel/database: DSN)
#   Env-mismatch gate (internal/cli: EnvMismatch)
#
# Integration sub-tests (NoSelf, OverGrant, IntegrationScope) need DATABASE_URL.
SECURITY_TESTS := Authz|Deny|DSN|Escalat|EnvMismatch|Isolation|NoSelf|OverGrant|Privilege|Prod|RLS|Redact|Secret|Security|Sensitive|Unsafe

.PHONY: test-security
test-security: ## Security-critical tests: authz, RLS, secrets, redaction, unsafe-config
	DATABASE_URL="$${DATABASE_URL:-$(TEST_DSN)}" \
		$(GO) test -run '$(SECURITY_TESTS)' -count=1 $(PKGS)

# Adversarial fuzzing drill (roadmap S8). CI runs the seed corpus as ordinary
# tests; this target drives the go native fuzzer against the two highest-value
# untrusted-input parsers. FUZZTIME is overridable (e.g. FUZZTIME=5m for nightly).
FUZZTIME ?= 30s
test-fuzz: ## Fuzz the filter DSL parser and cursor decoder (FUZZTIME=30s default)
	$(GO) test ./kernel/filtering/ -run '^$$' -fuzz=FuzzFilterParse  -fuzztime=$(FUZZTIME)
	$(GO) test ./kernel/filtering/ -run '^$$' -fuzz=FuzzParseSort     -fuzztime=$(FUZZTIME)
	$(GO) test ./kernel/pagination/ -run '^$$' -fuzz=FuzzDecodeCursor -fuzztime=$(FUZZTIME)

# Hot-path benchmarks (criterion #17).
# bench:        run all package benchmarks; outputs raw go test -bench lines.
# bench-budget: pipe bench output through the budget gate tool and fail on violation.
BENCH_PKGS := \
	./kernel/authz/... \
	./kernel/policy/... \
	./kernel/httpx/... \
	./kernel/config/... \
	./kernel/filtering/... \
	./kernel/pagination/... \
	./kernel/audit/... \
	./kernel/sequence/...

.PHONY: bench
bench: ## Run hot-path benchmarks with allocation counts (DB-backed benches need `make up` or DATABASE_URL)
	DATABASE_URL="$${DATABASE_URL:-$(TEST_DSN)}" WOWAPI_REQUIRE_DB=1 \
		$(GO) test -bench=. -benchmem -run=^$$ $(BENCH_PKGS)

.PHONY: bench-budget
bench-budget: ## Enforce performance budgets (fails if any benchmark exceeds bench-budgets.txt; needs a real DB for the audit/sequence benches)
	DATABASE_URL="$${DATABASE_URL:-$(TEST_DSN)}" WOWAPI_REQUIRE_DB=1 \
		$(GO) test -bench=. -benchmem -run=^$$ $(BENCH_PKGS) \
		| $(GO) run ./internal/tools/benchbudget bench-budgets.txt

.PHONY: coverage
# Packages measured for coverage. Excludes process/tool mains and test-only
# fixtures that cannot be meaningfully unit-tested (decision 2026-07-05: the
# coverage floor applies to the aggregate of the remaining packages).
COVER_EXCLUDE := /cmd/wowapi|/internal/tools/migrate|/internal/testmodules|/module$$
COVER_PKGS = $(shell $(GO) list ./... | grep -vE '$(COVER_EXCLUDE)')

coverage: ## Unit coverage report — runs against the real DB (needs `make up` or DATABASE_URL)
	DATABASE_URL="$${DATABASE_URL:-$(TEST_DSN)}" WOWAPI_REQUIRE_DB=1 \
		$(GO) test -coverprofile=coverage.out $(COVER_PKGS) && $(GO) tool cover -func=coverage.out | tail -1

COVERAGE_FLOOR ?= 90.0
.PHONY: coverage-check
coverage-check: coverage ## Enforce the coverage floor (raise COVERAGE_FLOOR as coverage grows)
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@pct=$$($(GO) tool cover -func=coverage.out | awk '/^total:/ {gsub("%","",$$3); print $$3}'); \
	echo "total coverage: $$pct% (floor $(COVERAGE_FLOOR)%)"; \
	awk -v p="$$pct" -v f="$(COVERAGE_FLOOR)" 'BEGIN{ if (p+0 < f+0) { printf "coverage %.1f%% below floor %.1f%%\n", p, f; exit 1 } }'

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
ci: ## Full local CI: vet + boundary lint, unit, race, perf budgets, build (golangci-lint = make lint-new / hosted CI)
	$(GO) vet $(PKGS)
	$(MAKE) lint-boundaries
	$(MAKE) test-unit
	$(MAKE) test-race
	$(MAKE) bench-budget
	$(MAKE) build

.PHONY: ci-container
ci-container: ## Run `make ci` inside the toolbox container (authoritative gate: DB tests MUST run)
	$(COMPOSE) run --rm -e WOWAPI_REQUIRE_DB=1 tools make ci

##@ Security & Release (CI-enforced)

.PHONY: actionlint
actionlint: ## Lint GitHub Actions workflows (pinned actionlint $(ACTIONLINT_VERSION))
	@if ! actionlint --version 2>/dev/null | grep -q "$(patsubst v%,%,$(ACTIONLINT_VERSION))"; then \
		$(GO) install github.com/rhysd/actionlint/cmd/actionlint@$(ACTIONLINT_VERSION); \
	fi
	actionlint -color

# govulncheck / goreleaser install @latest DELIBERATELY (unlike the lint tools):
# govulncheck is a security scanner and must track the newest vuln database + checks
# (pinning it would freeze detection — the opposite of what we want); the authoritative
# release build is pinned in .github/workflows/release.yml (SHA-pinned goreleaser-action
# + version "~> v2"), so these local convenience targets need no pin.
.PHONY: govulncheck
govulncheck: ## Scan for known Go vulnerabilities reachable by our code (govulncheck@latest — tracks newest checks)
	@command -v govulncheck >/dev/null 2>&1 || $(GO) install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

.PHONY: goreleaser-check
goreleaser-check: ## Validate the GoReleaser release config (release itself is pinned in release.yml)
	@command -v goreleaser >/dev/null 2>&1 || $(GO) install github.com/goreleaser/goreleaser/v2@latest
	goreleaser check

.PHONY: release-snapshot
release-snapshot: ## Dry-run the full release build locally (no publish)
	@command -v goreleaser >/dev/null 2>&1 || $(GO) install github.com/goreleaser/goreleaser/v2@latest
	goreleaser release --snapshot --clean --skip=sign,sbom
