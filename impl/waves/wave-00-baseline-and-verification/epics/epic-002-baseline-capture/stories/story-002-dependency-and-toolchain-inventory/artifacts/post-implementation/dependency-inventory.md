---
id: ART-W00-E02-S002-001
type: artifact
title: Dependency inventory — go.mod cross-check against REVIEW §L/§M
lifecycle_stage: post-implementation
parent_story: W00-E02-S002
producing_task: W00-E02-S002-T001
status: produced
created_at: 2026-07-13
updated_at: 2026-07-13
commit_sha: 0a31186cada5c275a588c74081cf977adf346e61
---

# Dependency inventory — W00-E02-S002-T001

Captured 2026-07-13 at commit `0a31186cada5c275a588c74081cf977adf346e61` (branch `main`),
environment: macOS 26.5.2 (Darwin 25.5.0), arm64, `go version go1.26.5 darwin/arm64`.
Register cross-checked: `docs/implementation/fable5-final-architecture-review-2026-07-11.md`
("REVIEW") §L (approved, lines 285–287) and §M (rejected, lines 289–294).

## Raw command output (evidence EV-W00-E02-S002-001)

Stored under `../../evidence/logs/`:

| Command | Output file | Size |
|---|---|---|
| `go list -m all` | `go-list-m-all.txt` | 340 lines (main module + 339 modules in the build list) |
| `go mod graph` | `go-mod-graph.txt` | 715 edges |
| `go list -m -json all` | `go-list-m-json-all.txt` | 2430 lines |
| `go mod why -m <pkg>` (targeted) | `go-mod-why.txt` | provenance checks quoted below |

## Module counts at this commit

- `go.mod` top `require` block: **13 direct** require lines (go.mod:8–20).
- `go.mod` second `require` block: **47 indirect** require lines (go.mod:24–70).
- Full build list (`go list -m all`): 339 modules besides the main module. The build list is
  larger than go.mod's 60 require lines because Go's module graph includes the full go.mod
  requirements of dependencies (e.g. `minio-go/v7`'s own test/lint tooling requirements), most of
  which are never needed to build any wowapi package — see the `go mod why -m` provenance notes
  below where this distinction matters.

## Direct-dependency disposition table (all 13 require lines — zero unaddressed)

Disposition vocabulary: `approved` (REVIEW §L original-10), `newly-approved` (§L reuse-work
approvals), `undocumented drift` (absent from §L — escalation required).

| # | go.mod line | Module | Version | Disposition | §L match |
|---|---|---|---|---|---|
| 1 | 8 | `github.com/go-playground/validator/v10` | v10.30.3 | **approved** | "validator/v10" |
| 2 | 9 | `github.com/golang-jwt/jwt/v5` | v5.3.1 | **approved** | "jwt/v5" (`WithValidMethods` alg-confusion mitigation noted in §L) |
| 3 | 10 | `github.com/google/uuid` | v1.6.0 | **approved** | "uuid" |
| 4 | 11 | `github.com/jackc/pgx/v5` | v5.10.0 | **approved** | "pgx/v5" |
| 5 | 12 | `github.com/minio/minio-go/v7` | v7.2.1 | **approved** | "minio-go/v7" |
| 6 | 13 | `github.com/pressly/goose/v3` | v3.27.2 | **approved** | "goose/v3" |
| 7 | 14 | `github.com/prometheus/client_golang` | v1.23.2 | **approved** | "prometheus/client_golang" |
| 8 | 15 | `github.com/shopspring/decimal` | v1.4.0 | **approved** | "shopspring/decimal" |
| 9 | 16 | `go.opentelemetry.io/otel` | v1.44.0 | **approved** | "otel×4" (1 of 4) |
| 10 | 17 | `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp` | v1.44.0 | **approved** | "otel×4" (2 of 4) |
| 11 | 18 | `go.opentelemetry.io/otel/sdk` | v1.44.0 | **approved** | "otel×4" (3 of 4) |
| 12 | 19 | `go.opentelemetry.io/otel/trace` | v1.44.0 | **approved** | "otel×4" (4 of 4) |
| 13 | 20 | `gopkg.in/yaml.v3` | v3.0.1 | **approved** | "yaml.v3" (§L watch item — see below) |

**Result: 13/13 direct dependencies dispositioned `approved`. Zero `undocumented drift`. Zero
unaddressed entries.**

## "10 vs 13" reconciliation (plan.md unresolved question — RESOLVED)

REVIEW §L's "All 10 current direct deps" reconciles **exactly** with go.mod's 13 require lines:
§L counts the four `go.opentelemetry.io/otel*` require lines (go.mod:16–19, all v1.44.0) as one
logical dependency, written "otel×4". 13 require lines − 4 otel lines + 1 logical otel entry =
10 logical dependencies, matching §L's list one-for-one. **No direct dependency has been added or
removed since REVIEW was authored.**

## §L "new approvals for reuse work" — presence/absence (explicit, all three)

| Package | §L status | Found at commit `0a31186` | Evidence |
|---|---|---|---|
| `github.com/cenkalti/backoff/v5` | approved (MIT, "already transitive") | **PRESENT — indirect**, v5.0.3 (go.mod:25) | `go mod why -m`: needed via `adapters/tracing/otel` → `otlptracehttp/internal/retry` (`go-mod-why.txt`) |
| `github.com/hashicorp/golang-lru/v2` | approved (MPL-2.0) for future adoption | **ABSENT from go.mod** (neither direct nor indirect). Appears in the unpruned module graph only via `minio-go/v7@v7.2.1`'s own go.mod (`go-mod-graph.txt`); `go mod why -m` reports "main module does not need module" | `go-mod-why.txt`, `go-mod-graph.txt` |
| `github.com/sony/gobreaker` | approved (MIT, P2) for future adoption | **ABSENT** — not in go.mod and not anywhere in the 339-module build list (`grep gobreaker go-list-m-all.txt` → no match) | `go-list-m-all.txt` |

Absence of golang-lru/v2 and gobreaker is **expected**, consistent with §L's "new approvals for
reuse work" describing packages approved for future adoption (FBL-04 et al.), not already present.

## Additional confirmations

- `github.com/sethvargo/go-retry v0.3.0` — **PRESENT, indirect** (go.mod:52), pulled in by
  `pressly/goose/v3` (`go mod why -m`: `kernel/database` → `goose/v3` → `go-retry`). Not imported
  by any wowapi package (present-and-unused at application level). This re-confirms REVIEW's
  Stage-7 adjudication at current HEAD by fresh measurement.
- `yaml.v3` watch item — `gopkg.in/yaml.v3 v3.0.1` remains the direct dependency (go.mod:20);
  the community fork `go.yaml.in/yaml/v3 v3.0.4` is **present, indirect** (go.mod:60), needed via
  `adapters/storage/s3` → `minio-go/v7` (`go-mod-why.txt`). Consistent with §L's "community fork
  already indirect — monitor, no action now." Monitor-only; no action taken.

## REVIEW §M rejected-register check — all four entries explicitly addressed

| §M rejected entry | Result at commit `0a31186` | Detail |
|---|---|---|
| viper / envconfig (config libs) | **ABSENT from go.mod, as required** | `github.com/spf13/viper` is NOT in go.mod (direct or indirect) and `go mod why -m` reports "main module does not need module"; it appears only in the unpruned module graph via `minio-go/v7@v7.2.1`'s own go.mod requirements (`go-mod-graph.txt`), never imported/built. No `envconfig` module (any path) appears anywhere in the 339-module build list. (`github.com/go-viper/mapstructure/v2` in the graph is an unrelated mapstructure fork, also not needed by the main module.) |
| New message bus (NATS/Kafka client) | **ABSENT, as required** | No `nats-io`, `segmentio/kafka`, `confluent`, `sarama`, or `twmb/franz` module anywhere in `go-list-m-all.txt`. |
| Any password-hashing lib | **ABSENT, as required** | No `bcrypt`/`argon2`/`scrypt`/`passlib` third-party module in the build list. `golang.org/x/crypto v0.53.0` (indirect, go.mod:61) is needed only for `x/crypto/sha3` via `validator/v10` (`go-mod-why.txt`) — not password hashing. |
| Custom crypto | **No drift observed in dependency terms** | No third-party crypto module beyond vetted `golang.org/x/crypto` (sha3, above) in the build list. (Code-level "all crypto is stdlib" verification is REVIEW's own claim; this inventory checks the module graph only.) |

**Result: zero rejected dependencies have entered go.mod. Zero drift requiring escalation.**

## Overall conclusion

Zero unexplained drift between go.mod at commit `0a31186cada5c275a588c74081cf977adf346e61` and
REVIEW §L/§M. All 13 direct require lines map onto §L's approved 10 logical dependencies; the
three reuse-work approvals are in the exact expected state (backoff present-indirect, the other
two absent); no §M rejected dependency is present; the yaml.v3 watch item is unchanged.
