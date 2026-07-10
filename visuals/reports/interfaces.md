# Interfaces

> The tool has detected code relationships. It has not assumed whether the project follows layered architecture, clean architecture, hexagonal architecture, MVC, CQRS, or any other pattern unless explicitly configured.

## Lookup

- **Package:** github.com/qatoolist/wowapi/internal/testmodules/requests
- **Location:** `internal/testmodules/requests/module.go:66`

**Methods:**

- `ByID func(ctx context.Context, id github.com/google/uuid.UUID) (*RequestDTO, error)`

**Implemented by:**

- `github.com/qatoolist/wowapi/internal/testmodules/requests.lookupImpl`

## KeySource

- **Package:** github.com/qatoolist/wowapi/kernel/auth
- **Location:** `kernel/auth/auth.go:43`

**Methods:**

- `Key func(ctx context.Context, kid string) (any, error)`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/auth.jwksKeySource`
- `github.com/qatoolist/wowapi/kernel/auth.staticKeySource`

## PrincipalStore

- **Package:** github.com/qatoolist/wowapi/kernel/auth
- **Location:** `kernel/auth/auth.go:163`

**Methods:**

- `UserIDBySubject func(ctx context.Context, subject string) (github.com/google/uuid.UUID, error)`
- `ValidateCapacity func(ctx context.Context, userID github.com/google/uuid.UUID, tenantID github.com/google/uuid.UUID, capacityID github.com/google/uuid.UUID) error`

**Implemented by:**

- `github.com/qatoolist/wowapi/adapters/auth/pgprincipal.Store`

## AuditSink

- **Package:** github.com/qatoolist/wowapi/kernel/authz
- **Location:** `kernel/authz/store.go:89`

**Methods:**

- `AuthzDenial func(ctx context.Context, a Actor, perm string, t Target, reason string)`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel.durableAudit`
- `github.com/qatoolist/wowapi/kernel.loggingAudit`

## Evaluator

- **Package:** github.com/qatoolist/wowapi/kernel/authz
- **Location:** `kernel/authz/authz.go:108`

**Methods:**

- `Evaluate func(ctx context.Context, db github.com/qatoolist/wowapi/kernel/database.TenantDB, a Actor, perm string, t Target) (Decision, error)`
- `Filter func(ctx context.Context, db github.com/qatoolist/wowapi/kernel/database.TenantDB, a Actor, perm string, rt string) (ListFilter, error)`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/authz.engine`

## PolicyEngine

- **Package:** github.com/qatoolist/wowapi/kernel/authz
- **Location:** `kernel/authz/evaluator.go:17`

**Methods:**

- `Matches func(conds []Condition, attrs map[string]any) (bool, error)`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/policy.Engine`

## RelationshipChecker

- **Package:** github.com/qatoolist/wowapi/kernel/authz
- **Location:** `kernel/authz/authz.go:119`

**Methods:**

- `Has func(ctx context.Context, db github.com/qatoolist/wowapi/kernel/database.TenantDB, subject Actor, relType string, obj github.com/qatoolist/wowapi/kernel/resource.Ref, at time.Time) (bool, error)`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/relationship.Checker`

## Store

- **Package:** github.com/qatoolist/wowapi/kernel/authz
- **Location:** `kernel/authz/store.go:66`

**Methods:**

- `ActiveAssignments func(ctx context.Context, db github.com/qatoolist/wowapi/kernel/database.TenantDB, a Actor, at time.Time) ([]Assignment, error)`
- `OrgAncestors func(ctx context.Context, db github.com/qatoolist/wowapi/kernel/database.TenantDB, orgID github.com/google/uuid.UUID) ([]github.com/google/uuid.UUID, error)`
- `OrgSubtree func(ctx context.Context, db github.com/qatoolist/wowapi/kernel/database.TenantDB, orgID github.com/google/uuid.UUID) ([]github.com/google/uuid.UUID, error)`
- `Policies func(ctx context.Context, db github.com/qatoolist/wowapi/kernel/database.TenantDB, a Actor, perm string, rt string) ([]Policy, error)`
- `ResourceOrg func(ctx context.Context, db github.com/qatoolist/wowapi/kernel/database.TenantDB, ref github.com/qatoolist/wowapi/kernel/resource.Ref) (github.com/google/uuid.UUID, error)`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/authz.CachingStore`
- `github.com/qatoolist/wowapi/kernel/authz.PgStore`

## ModuleView

- **Package:** github.com/qatoolist/wowapi/kernel/config
- **Location:** `kernel/config/moduleview.go:13`

**Methods:**

- `Decode func(out any) error`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/config.MapView`
- `github.com/qatoolist/wowapi/kernel/rules.Resolved`

## DB

- **Package:** github.com/qatoolist/wowapi/kernel/database
- **Location:** `kernel/database/database.go:42`

**Methods:**

- `platformSealed func()`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/database.platformTx`

## DBTX

- **Package:** github.com/qatoolist/wowapi/kernel/database
- **Location:** `kernel/database/database.go:25`

**Methods:**

- `Exec func(ctx context.Context, sql string, args ...any) (github.com/jackc/pgx/v5/pgconn.CommandTag, error)`
- `Query func(ctx context.Context, sql string, args ...any) (github.com/jackc/pgx/v5.Rows, error)`
- `QueryRow func(ctx context.Context, sql string, args ...any) github.com/jackc/pgx/v5.Row`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/database.platformTx`
- `github.com/qatoolist/wowapi/kernel/database.tenantTx`

## IdemStore

- **Package:** github.com/qatoolist/wowapi/kernel/database
- **Location:** `kernel/database/idempotency.go:27`

**Methods:**

- `Begin func(ctx context.Context, db TenantDB, actorScope string, key string, requestHash string, ttl time.Duration) (Replay, error)`
- `Complete func(ctx context.Context, db TenantDB, actorScope string, key string, status int, body []byte) error`
- `Discard func(ctx context.Context, db TenantDB, actorScope string, key string) error`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/database.PgIdemStore`

## TenantDB

- **Package:** github.com/qatoolist/wowapi/kernel/database
- **Location:** `kernel/database/database.go:35`

**Methods:**

- `tenantSealed func()`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/database.tenantTx`

## TxManager

- **Package:** github.com/qatoolist/wowapi/kernel/database
- **Location:** `kernel/database/txmanager.go:17`

**Methods:**

- `Platform func(ctx context.Context, fn func(ctx context.Context, db DB) error) error`
- `WithTenant func(ctx context.Context, fn func(ctx context.Context, db TenantDB) error) error`
- `WithTenantRO func(ctx context.Context, fn func(ctx context.Context, db TenantDB) error) error`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/database.Manager`

## Authenticator

- **Package:** github.com/qatoolist/wowapi/kernel/httpx
- **Location:** `kernel/httpx/authz_gate.go:24`

**Methods:**

- `Authenticate func(r *net/http.Request) (github.com/qatoolist/wowapi/kernel/authz.Actor, error)`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/apikey.Authenticator`
- `github.com/qatoolist/wowapi/kernel/auth.Authenticator`
- `github.com/qatoolist/wowapi/kernel/httpx.DenyAllAuthenticator`
- `github.com/qatoolist/wowapi/kernel/httpx.compositeAuthenticator`

## RateLimiter

- **Package:** github.com/qatoolist/wowapi/kernel/httpx
- **Location:** `kernel/httpx/ratelimit.go:31`

**Methods:**

- `Allow func(key string) (allowed bool, retryAfter time.Duration)`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/httpx.TokenBucket`

## Provider

- **Package:** github.com/qatoolist/wowapi/kernel/integration
- **Location:** `kernel/integration/integration.go:49`

**Methods:**

- `HealthCheck func(ctx context.Context, cfg Config) error`
- `Key func() string`
- `Kind func() string`

**Implemented by:**

_No implementers detected._

## Job

- **Package:** github.com/qatoolist/wowapi/kernel/jobs
- **Location:** `kernel/jobs/jobs.go:29`

**Methods:**

- `Kind func() string`

**Implemented by:**

_No implementers detected._

## IDGen

- **Package:** github.com/qatoolist/wowapi/kernel/model
- **Location:** `kernel/model/model.go:147`

**Methods:**

- `New func() github.com/google/uuid.UUID`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/model.uuidV7Gen`
- `github.com/qatoolist/wowapi/testkit/fakes.IDGen`

## ChannelSender

- **Package:** github.com/qatoolist/wowapi/kernel/notify
- **Location:** `kernel/notify/sender.go:13`

**Methods:**

- `Send func(ctx context.Context, d Delivery) (providerMessageID string, err error)`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/notify.FakeSender`
- `github.com/qatoolist/wowapi/kernel/notify.inAppSender`

## Metrics

- **Package:** github.com/qatoolist/wowapi/kernel/observability
- **Location:** `kernel/observability/metrics.go:20`

**Methods:**

- `IncCounter func(name string, value float64, labels map[string]string)`
- `ObserveRequest func(route string, method string, status int, dur time.Duration, respBytes int)`
- `SetGauge func(name string, value float64, labels map[string]string)`

**Implemented by:**

- `github.com/qatoolist/wowapi/adapters/metrics/prometheus.Prometheus`
- `github.com/qatoolist/wowapi/kernel/observability.noOp`

## Span

- **Package:** github.com/qatoolist/wowapi/kernel/observability
- **Location:** `kernel/observability/tracing.go:33`

**Methods:**

- `End func()`
- `RecordError func(err error)`
- `SetAttr func(key string, value string)`

**Implemented by:**

- `github.com/qatoolist/wowapi/adapters/tracing/otel.otelSpan`
- `github.com/qatoolist/wowapi/kernel/observability.noopSpan`

## Tracer

- **Package:** github.com/qatoolist/wowapi/kernel/observability
- **Location:** `kernel/observability/tracing.go:17`

**Methods:**

- `Extract func(ctx context.Context, carrier string) context.Context`
- `Inject func(ctx context.Context) string`
- `StartSpan func(ctx context.Context, name string) (context.Context, Span)`

**Implemented by:**

- `github.com/qatoolist/wowapi/adapters/tracing/otel.Tracer`
- `github.com/qatoolist/wowapi/kernel/observability.noopTracer`

## Writer

- **Package:** github.com/qatoolist/wowapi/kernel/outbox
- **Location:** `kernel/outbox/outbox.go:39`

**Methods:**

- `Write func(ctx context.Context, db github.com/qatoolist/wowapi/kernel/database.TenantDB, e Event) error`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/outbox.pgWriter`

## Registrar

- **Package:** github.com/qatoolist/wowapi/kernel/resource
- **Location:** `kernel/resource/resource.go:90`

**Methods:**

- `Upsert func(ctx context.Context, ref Ref, orgID *github.com/google/uuid.UUID, label string, status string) error`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/resource.boundRegistrar`

## Provider

- **Package:** github.com/qatoolist/wowapi/kernel/secrets
- **Location:** `kernel/secrets/secrets.go:48`

**Methods:**

- `Resolve func(ctx context.Context, ref Ref) (string, error)`

**Implemented by:**

- `github.com/qatoolist/wowapi/adapters/secrets/envprovider.Provider`

## SpineInvalidator

- **Package:** github.com/qatoolist/wowapi/kernel/seeds
- **Location:** `kernel/seeds/seeds.go:165`

**Methods:**

- `InvalidateAll func()`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/authz.CachingStore`

## Adapter

- **Package:** github.com/qatoolist/wowapi/kernel/storage
- **Location:** `kernel/storage/storage.go:37`

**Methods:**

- `Delete func(ctx context.Context, key string) error`
- `Peek func(ctx context.Context, key string, n int) ([]byte, error)`
- `PresignGet func(ctx context.Context, key string, ttl time.Duration) (PresignedURL, error)`
- `PresignPut func(ctx context.Context, key string, ttl time.Duration) (PresignedURL, error)`
- `Stat func(ctx context.Context, key string) (ObjectInfo, error)`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/storage.Memory`

## SecretResolver

- **Package:** github.com/qatoolist/wowapi/kernel/webhook
- **Location:** `kernel/webhook/webhook.go:110`

**Methods:**

- `Resolve func(ctx context.Context, ref string) (string, error)`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel.secretRefResolver`
- `github.com/qatoolist/wowapi/kernel/webhook.FakeSecretResolver`

## Sender

- **Package:** github.com/qatoolist/wowapi/kernel/webhook
- **Location:** `kernel/webhook/webhook.go:103`

**Methods:**

- `Post func(ctx context.Context, url string, body []byte, headers map[string]string) (statusCode int, err error)`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/webhook.FakeSender`
- `github.com/qatoolist/wowapi/kernel/webhook.HTTPSender`

## Verifier

- **Package:** github.com/qatoolist/wowapi/kernel/webhook
- **Location:** `kernel/webhook/webhook.go:98`

**Methods:**

- `Verify func(secret string, body []byte, headers map[string]string) error`

**Implemented by:**

- `github.com/qatoolist/wowapi/kernel/webhook.FakeVerifier`
- `github.com/qatoolist/wowapi/kernel/webhook.HMACVerifier`

## Context

- **Package:** github.com/qatoolist/wowapi/module
- **Location:** `module/module.go:67`

**Methods:**

- `Artifacts func() *github.com/qatoolist/wowapi/kernel/artifact.Pipeline`
- `Attachments func() *github.com/qatoolist/wowapi/kernel/attachment.Service`
- `Audit func() *github.com/qatoolist/wowapi/kernel/audit.Writer`
- `Authz func() github.com/qatoolist/wowapi/kernel/authz.Evaluator`
- `Bulk func() *github.com/qatoolist/wowapi/kernel/bulk.Service`
- `Comments func() *github.com/qatoolist/wowapi/kernel/comment.Service`
- `Config func() github.com/qatoolist/wowapi/kernel/config.ModuleView`
- `DocumentClasses func() *github.com/qatoolist/wowapi/kernel/document.Registry`
- `DocumentHooks func() *github.com/qatoolist/wowapi/kernel/document.Hooks`
- `Documents func() *github.com/qatoolist/wowapi/kernel/document.Service`
- `Events func() *github.com/qatoolist/wowapi/kernel/outbox.HandlerRegistry`
- `Health func(name string, check func(context.Context) error)`
- `IDGen func() github.com/qatoolist/wowapi/kernel/model.IDGen`
- `IntegrationProviders func() *github.com/qatoolist/wowapi/kernel/integration.Registry`
- `Integrations func() *github.com/qatoolist/wowapi/kernel/integration.Store`
- `Jobs func() *github.com/qatoolist/wowapi/kernel/jobs.Registry`
- `Logger func() *log/slog.Logger`
- `Migrations func(fsys io/fs.FS)`
- `Notify func() *github.com/qatoolist/wowapi/kernel/notify.Service`
- `NotifyTemplates func() *github.com/qatoolist/wowapi/kernel/notify.Registry`
- `OpenAPI func(fragment []byte)`
- `Outbox func() github.com/qatoolist/wowapi/kernel/outbox.Writer`
- `Permissions func() *github.com/qatoolist/wowapi/kernel/authz.Registry`
- `Port func(name string) (any, error)`
- `ProvidePort func(name string, impl any)`
- `RecurringJob func(name string, every time.Duration, fn func(ctx context.Context, db github.com/qatoolist/wowapi/kernel/database.TenantDB) error)`
- `Resources func() *github.com/qatoolist/wowapi/kernel/resource.Registry`
- `RetentionClasses func() *github.com/qatoolist/wowapi/kernel/retention.Registry`
- `Routes func() *github.com/qatoolist/wowapi/kernel/httpx.Router`
- `Rules func() *github.com/qatoolist/wowapi/kernel/rules.Registry`
- `RulesResolver func() *github.com/qatoolist/wowapi/kernel/rules.Resolver`
- `Seeds func(fsys io/fs.FS)`
- `Sequence func() *github.com/qatoolist/wowapi/kernel/sequence.Allocator`
- `Tx func() github.com/qatoolist/wowapi/kernel/database.TxManager`
- `Validator func() *github.com/qatoolist/wowapi/kernel/validation.Validator`
- `Webhooks func() *github.com/qatoolist/wowapi/kernel/webhook.Service`
- `WorkflowRuntime func() *github.com/qatoolist/wowapi/kernel/workflow.Runtime`
- `Workflows func() *github.com/qatoolist/wowapi/kernel/workflow.Registry`

**Implemented by:**

- `github.com/qatoolist/wowapi/app.moduleContext`

## Module

- **Package:** github.com/qatoolist/wowapi/module
- **Location:** `module/module.go:48`

**Methods:**

- `DependsOn func() []string`
- `Name func() string`
- `Register func(ctx Context) error`

**Implemented by:**

- `github.com/qatoolist/wowapi/internal/testmodules/requests.Module`

