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
ALLURE_RESULTS_DIR ?= allure-results
ALLURE_REPORT_DIR  ?= allure-report

.DEFAULT_GOAL := help

##@ General

.PHONY: help
help: ## List targets
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

.PHONY: setup
setup: tools hooks ## One-time developer setup (tools + git hooks + go mod download)
	$(GO) mod download

.PHONY: tools
tools: ## Install host dev tools (pinned golangci-lint $(GOLANGCI_VERSION) + golurectl from go.mod)
	@if ! golangci-lint version 2>/dev/null | grep -q "$(patsubst v%,%,$(GOLANGCI_VERSION))"; then \
		echo "installing golangci-lint $(GOLANGCI_VERSION)"; \
		$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION); \
	fi
	@$(GO) tool golurectl version >/dev/null

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
	@echo "make seed: available in Phase 5 (seed loader) — see docs/GOALS-TRACKER.md" >&2; exit 2

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

.PHONY: lint-constructors
lint-constructors: ## Reject infrastructure constructors outside composition packages (AR-06)
	$(GO) run ./internal/tools/constructorlint/cmd/constructorlint ./...

.PHONY: lint-boundaries
lint-boundaries: lint-constructors ## Import, constructor, vocabulary, and Reveal() boundary lint
	sh scripts/lint_boundaries.sh

.PHONY: tenantfk-gate
tenantfk-gate: ## DATA-01: fail if any post-cleanup migration adds a non-composite tenant FK
	DATABASE_URL="$${DATABASE_URL:-$(TEST_DSN)}" $(GO) run ./internal/tools/tenantfk gate --since=36 --migrations=migrations

.PHONY: lint-lifecycle
lint-lifecycle: ## Static provider/lifecycle manifest lint (backlog B9; kernel/lifecycle)
	$(GO) run ./cmd/wowapi lint lifecycle

.PHONY: docs-check
docs-check: ## Compile normative doc examples and verify generated references/future-state labels (AR-05)
	$(GO) run ./internal/tools/docexamples -root .

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

.PHONY: secret-scan
secret-scan: ## CI-parity gitleaks scan of the FULL branch range (merge-base(origin/main)^..HEAD) — covers commits the pre-push hook's incremental scan never saw
	@command -v gitleaks >/dev/null 2>&1 || { echo ">> gitleaks not found — brew install gitleaks"; exit 1; }
	@base=$$(git merge-base origin/main HEAD 2>/dev/null); \
		if [ -z "$$base" ]; then echo ">> cannot resolve merge-base with origin/main"; exit 1; fi; \
		echo ">> gitleaks $$base^..HEAD (CI range semantics)"; \
		gitleaks detect --redact --exit-code=1 --log-opts="--no-merges --first-parent $$base^..HEAD"

.PHONY: hooks
hooks: ## Install the versioned git hooks (pre-commit + pre-push)
	@git config core.hooksPath .githooks && chmod +x .githooks/pre-commit .githooks/pre-push 2>/dev/null; echo "git hooks installed (core.hooksPath=.githooks)"

##@ Tests

.PHONY: ensure-infra
ensure-infra: ## Ensure infra is healthy
	@$(COMPOSE) ps --format '{{.Name}} {{.Status}}' | grep -q "Up.*(healthy)" || $(COMPOSE) up -d --wait

.PHONY: test
test: test-unit ## All currently available test suites

.PHONY: test-full
test-full: ensure-infra test-unit test-integration test-contract golden-consumer test-security ## Run all test suites: unit, integration, contract, golden-consumer, security

.PHONY: test-full-html
test-full-html: ensure-infra test-unit test-integration test-contract golden-consumer test-security ## Run all test suites and generate HTML report
	@$(GO) test -json -count=1 ./... | python3 scripts/test_to_html.py > full_test_report.html
	@echo "Running full test suite..."
	@DATABASE_URL="$${DATABASE_URL:-$(TEST_DSN)}" WOWAPI_REQUIRE_DB=1 WOWAPI_REQUIRE_S3=1 \
		$(GO) test ./... -json -count=1 | python3 scripts/test_to_html.py > full_test_report.html
	@echo "Report generated at full_test_report.html"

.PHONY: test-allure
test-allure: ensure-infra ## Run the full suite and generate Allure results + HTML via the tools container
	@set -o pipefail; \
		status=0; \
		DATABASE_URL="$${DATABASE_URL:-$(TEST_DSN)}" \
		WOWAPI_REQUIRE_DB=1 WOWAPI_REQUIRE_S3=1 S3_TEST_ENDPOINT=localhost:9000 \
		$(GO) test ./... -json -count=1 \
			| $(COMPOSE) run --rm -T \
				-e ALLURE_RESULTS_DIR="$(ALLURE_RESULTS_DIR)" \
				-e ALLURE_REPORT_DIR="$(ALLURE_REPORT_DIR)" \
				tools sh -c '\
					for dir in "$$ALLURE_RESULTS_DIR" "$$ALLURE_REPORT_DIR"; do \
						case "$$dir" in ""|"/"|"."|"..") echo "refusing unsafe Allure artifact path: $$dir" >&2; exit 2;; esac; \
					done; \
					rm -rf -- "$$ALLURE_RESULTS_DIR" "$$ALLURE_REPORT_DIR"; \
					golurectl -l -e -s -a -o "$$ALLURE_RESULTS_DIR"' \
			|| status=$$?; \
		$(COMPOSE) run --rm -T \
			-e TEST_STATUS="$$status" \
			-e ALLURE_RESULTS_DIR="$(ALLURE_RESULTS_DIR)" \
			-e ALLURE_REPORT_DIR="$(ALLURE_REPORT_DIR)" \
			tools sh -c '\
				if [ "$$TEST_STATUS" -ne 0 ]; then \
					printf "{\"uuid\":\"00000000-0000-4000-8000-000000000001\",\"name\":\"go test ./...\",\"fullName\":\"wowapi test suite\",\"status\":\"broken\",\"stage\":\"finished\",\"statusDetails\":{\"message\":\"go test exited with status %s\"},\"labels\":[{\"name\":\"suite\",\"value\":\"wowapi\"}]}\n" \
						"$$TEST_STATUS" > "$$ALLURE_RESULTS_DIR/wowapi-suite-result.json"; \
				fi; \
				allure generate --clean -o "$$ALLURE_REPORT_DIR" "$$ALLURE_RESULTS_DIR"' || exit $$?; \
		printf "Allure results: %s\nAllure HTML report: %s/index.html\n" \
			"$(ALLURE_RESULTS_DIR)" "$(ALLURE_REPORT_DIR)"; \
		exit $$status

.PHONY: test-unit
test-unit: ensure-infra ## Unit tests (no external services)
	$(GO) test $(PKGS)

.PHONY: test-race
test-race: ## Unit tests with the race detector
	$(GO) test -race $(PKGS)

.PHONY: test-integration
test-integration: ensure-infra ## Integration tests against real Postgres (needs `make up` or DATABASE_URL)
	DATABASE_URL="$${DATABASE_URL:-$(TEST_DSN)}" WOWAPI_REQUIRE_DB=1 $(GO) test -run 'Integration' -count=1 $(PKGS)

.PHONY: test-contract
test-contract: ensure-infra ## Module contract + scratch external-consumer suite (needs DB)
	DATABASE_URL="$${DATABASE_URL:-$(TEST_DSN)}" $(GO) test -run 'Contract|ScratchConsumer' -count=1 ./testkit/...

.PHONY: golden-consumer
golden-consumer: ensure-infra ## Installed-CLI eight-subsystem consumer + real infra + N-1/N replay + RLS census
	DATABASE_URL="$${DATABASE_URL:-$(TEST_DSN)}" WOWAPI_REQUIRE_DB=1 WOWAPI_REQUIRE_S3=1 \
		$(GO) test ./internal/cli ./testkit \
		-run '^(TestGoldenConsumerInstalledBinaryTwoModules|TestGoldenConsumerRealInfrastructure|TestGoldenConsumerUpgradeReplay|TestGoldenConsumerFailingFixture|TestIntegrationRLSCensusComplete)$$' \
		-count=1 -v

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
test-security: ensure-infra ## Security-critical tests: authz, RLS, secrets, redaction, unsafe-config
	DATABASE_URL="$${DATABASE_URL:-$(TEST_DSN)}" \
		$(GO) test -run '$(SECURITY_TESTS)' -count=1 $(PKGS)

# Coverage-guided fuzzing. Both profiles use a persistent GOCACHE so Go's
# generated corpus ($$GOCACHE/fuzz) can be restored by CI on the next run.
FUZZTIME ?= 10s
FUZZ_CACHE ?= .fuzzcache/go-build
FUZZ_OUTPUT ?= fuzz-report

.PHONY: test-fuzz test-fuzz-pr test-fuzz-scheduled
test-fuzz: test-fuzz-pr ## Run the short PR coverage-guided fuzz profile

test-fuzz-pr: ## PR fuzz profile (FUZZTIME=10s per target)
	$(GO) run ./internal/tools/fuzzproof -profile pr -fuzztime $(FUZZTIME) -cache $(FUZZ_CACHE) -output $(FUZZ_OUTPUT)

test-fuzz-scheduled: ## Longer scheduled fuzz profile (FUZZTIME=1m per target)
	$(GO) run ./internal/tools/fuzzproof -profile scheduled -fuzztime $(FUZZTIME) -cache $(FUZZ_CACHE) -output $(FUZZ_OUTPUT)

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
	./kernel/sequence/... \
	./kernel/database/... \
	./kernel/jobs/... \
	./kernel/outbox/... \
	./kernel/workflow/... \
	./kernel/auth/... \
	./kernel/mfa/... \
	./kernel/httpclient/...

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

# Initial enforced baseline measured by the required real-DB gate on
# 2026-07-15: 84.4%. Keep a small deterministic-run margin and ratchet upward
# as coverage grows; a drop below 84.0 remains release-blocking.
COVERAGE_FLOOR ?= 84.0
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

NEO4J_URI ?= bolt://localhost:7687
NEO4J_USER ?= neo4j
NEO4J_PASSWORD ?= wowapi-local-only
NEO4J_DATABASE ?= neo4j
GRAPHIFY_CYPHER_PATH ?= graphify-out/cypher.txt

GRAPHIFY_VIZ_NODE_LIMIT = 40000

.PHONY: graph-check
graph-check: ## Graphify freshness check
	sh scripts/graphify_refresh.sh check

.PHONY: graph-update
graph-update: ## Graphify incremental update (code-only, no LLM)
	sh scripts/graphify_refresh.sh update

.PHONY: graph-build
graph-build: ## Full semantic graph build — Kimi/Moonshot ONLY, never Claude (needs MOONSHOT_API_KEY). Prefer this over in-session /graphify for the AI extraction.
	@test -n "$$MOONSHOT_API_KEY" || { echo "graph-build: MOONSHOT_API_KEY not set — required for Kimi/Moonshot semantic extraction (backend is pinned to kimi)"; exit 2; }
	GRAPHIFY_BACKEND=kimi sh scripts/graphify_refresh.sh extract

.PHONY: graph-neo4j
graph-neo4j: ## Export the current graph and load it into the local Neo4j container (needs `make up`)
	@test -n "$(NEO4J_URI)" || { echo "graph-neo4j: NEO4J_URI is required"; exit 2; }
	@test -n "$(NEO4J_USER)" || { echo "graph-neo4j: NEO4J_USER is required"; exit 2; }
	@test -n "$(NEO4J_PASSWORD)" || { echo "graph-neo4j: NEO4J_PASSWORD is required"; exit 2; }
	@test -n "$(NEO4J_DATABASE)" || { echo "graph-neo4j: NEO4J_DATABASE is required"; exit 2; }
	graphify export neo4j
	@test -f "$(GRAPHIFY_CYPHER_PATH)" || { echo "graph-neo4j: $(GRAPHIFY_CYPHER_PATH) not found after export"; exit 1; }
	$(COMPOSE) exec -T neo4j cypher-shell --non-interactive -a "$(NEO4J_URI)" -u "$(NEO4J_USER)" -p "$(NEO4J_PASSWORD)" -d "$(NEO4J_DATABASE)" < "$(GRAPHIFY_CYPHER_PATH)"

##@ CI

.PHONY: build
build: ## Build all packages and the CLI
	$(GO) build $(PKGS)
	$(GO) build -o bin/wowapi ./cmd/wowapi

.PHONY: ci
ci: ## Full local CI: vet + boundary lint + lifecycle lint, unit, race, perf budgets, build (golangci-lint = make lint-new / hosted CI)
	$(GO) vet $(PKGS)
	$(MAKE) lint-boundaries
	$(MAKE) lint-lifecycle
	$(MAKE) test-unit
	$(MAKE) test-race
	$(MAKE) bench-budget
	$(MAKE) build

.PHONY: ci-container
ci-container: ## Run `make ci` inside the toolbox container (authoritative gate: DB + S3 tests MUST run)
	$(COMPOSE) run --rm -e WOWAPI_REQUIRE_DB=1 -e WOWAPI_REQUIRE_S3=1 -e S3_TEST_ENDPOINT=minio:9000 tools make ci

# Hosted CI fans the container gate out into three parallel legs (same required-env
# posture as ci-container, so nothing can silently skip; vet/boundaries/build are
# proven by the host `unit` job at the same pinned Go version and are not repeated
# in-container). `make ci-container` remains the serial local equivalent.
# CI_TOOLS_ENV lets the workflow inject cacheable GOCACHE/GOMODCACHE paths.
CI_TOOLS_ENV ?=

INTEGRATION_RACE_PKGS := \
	./adapters/storage/s3 \
	./internal/e2e \
	./internal/tools/tenantfk \
	./kernel/database \
	./kernel/migration \
	./kernel/outbox \
	./testkit

.PHONY: check-test-skips check-required-test-prerequisites check-race-fixture test-race-integration
check-test-skips: ## Reject unapproved t.Skip sites and exercise negative/positive fixtures
	miscellaneous/check_test_skips.sh
	miscellaneous/check_test_skip_fixtures.sh

check-required-test-prerequisites: ## Prove required DB/S3 dependencies fail closed
	miscellaneous/check_required_test_prerequisites.sh

check-race-fixture: ## Prove the Go race detector catches the seeded negative fixture
	miscellaneous/check_race_detector.sh

test-race-integration: ## Race detector over DB/S3-backed integration packages
	$(GO) test -race -count=1 $(INTEGRATION_RACE_PKGS)

.PHONY: ci-container-test ci-container-race ci-container-bench
ci-container-test: ## Parallel gate leg 1: fail-closed prerequisites + DB/S3 tests + fuzz seed corpus
	$(COMPOSE) run --rm -e WOWAPI_REQUIRE_DB=1 -e WOWAPI_REQUIRE_S3=1 -e S3_TEST_ENDPOINT=minio:9000 $(CI_TOOLS_ENV) tools \
		sh -c 'make check-required-test-prerequisites && go test ./... && go test ./kernel/filtering/ ./kernel/pagination/ -run "^Fuzz" -count=1'

ci-container-race: ## Parallel gate leg 2: DB/S3 integration race detector + seeded negative fixture
	$(COMPOSE) run --rm -e WOWAPI_REQUIRE_DB=1 -e WOWAPI_REQUIRE_S3=1 -e S3_TEST_ENDPOINT=minio:9000 $(CI_TOOLS_ENV) tools \
		sh -c 'make check-race-fixture && make test-race-integration'

ci-container-bench: ## Parallel gate leg 3: performance budgets + lifecycle lint (DB required)
	$(COMPOSE) run --rm -e WOWAPI_REQUIRE_DB=1 -e WOWAPI_REQUIRE_S3=1 -e S3_TEST_ENDPOINT=minio:9000 $(CI_TOOLS_ENV) tools \
		sh -c 'make bench-budget && make lint-lifecycle'

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
