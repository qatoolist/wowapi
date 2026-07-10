# Type Summary

> The tool has detected code relationships. It has not assumed whether the project follows layered architecture, clean architecture, hexagonal architecture, MVC, CQRS, or any other pattern unless explicitly configured.

## Structs

| Type | Package | Fields | Location |
| --- | --- | --: | --- |
| Store | github.com/qatoolist/wowapi/adapters/auth/pgprincipal | 2 | `adapters/auth/pgprincipal/pgprincipal.go:26` |
| Prometheus | github.com/qatoolist/wowapi/adapters/metrics/prometheus | 6 | `adapters/metrics/prometheus/prometheus.go:30` |
| Provider | github.com/qatoolist/wowapi/adapters/secrets/envprovider | 1 | `adapters/secrets/envprovider/envprovider.go:19` |
| Tracer | github.com/qatoolist/wowapi/adapters/tracing/otel | 3 | `adapters/tracing/otel/otel.go:23` |
| otelSpan | github.com/qatoolist/wowapi/adapters/tracing/otel | 1 | `adapters/tracing/otel/otel.go:99` |
| APIConfig | github.com/qatoolist/wowapi/app | 5 | `app/views.go:37` |
| App | github.com/qatoolist/wowapi/app | 1 | `app/app.go:21` |
| Booted | github.com/qatoolist/wowapi/app | 9 | `app/boot.go:27` |
| Hook | github.com/qatoolist/wowapi/app | 3 | `app/run.go:21` |
| MigrateConfig | github.com/qatoolist/wowapi/app | 3 | `app/views.go:58` |
| MigrateDB | github.com/qatoolist/wowapi/app | 2 | `app/views.go:28` |
| RecurringJob | github.com/qatoolist/wowapi/app | 3 | `app/boot.go:43` |
| RuntimeDB | github.com/qatoolist/wowapi/app | 2 | `app/views.go:21` |
| WorkerConfig | github.com/qatoolist/wowapi/app | 4 | `app/views.go:48` |
| WorkerConfigOpts | github.com/qatoolist/wowapi/app | 12 | `app/worker.go:14` |
| bootOpts | github.com/qatoolist/wowapi/app | 1 | `app/boot.go:52` |
| bootState | github.com/qatoolist/wowapi/app | 6 | `app/context.go:47` |
| moduleContext | github.com/qatoolist/wowapi/app | 33 | `app/context.go:68` |
| moduleDeps | github.com/qatoolist/wowapi/app | 30 | `app/context.go:106` |
| GoMod | github.com/qatoolist/wowapi/internal/buildinfo | 3 | `internal/buildinfo/buildinfo.go:35` |
| cfgFlags | github.com/qatoolist/wowapi/internal/cli | 4 | `internal/cli/config_cmd.go:49` |
| crudData | github.com/qatoolist/wowapi/internal/cli | 7 | `internal/cli/gen_cmd.go:35` |
| deployVars | github.com/qatoolist/wowapi/internal/cli | 3 | `internal/cli/deploy_cmd.go:33` |
| fieldDef | github.com/qatoolist/wowapi/internal/cli | 4 | `internal/cli/gen_cmd.go:27` |
| initData | github.com/qatoolist/wowapi/internal/cli | 5 | `internal/cli/init_cmd.go:30` |
| newModuleData | github.com/qatoolist/wowapi/internal/cli | 2 | `internal/cli/new_module_cmd.go:25` |
| Config | github.com/qatoolist/wowapi/internal/testmodules/requests | 1 | `internal/testmodules/requests/module.go:14` |
| CreateRequest | github.com/qatoolist/wowapi/internal/testmodules/requests | 1 | `internal/testmodules/requests/dto.go:9` |
| Handlers | github.com/qatoolist/wowapi/internal/testmodules/requests | 4 | `internal/testmodules/requests/handlers.go:28` |
| Module | github.com/qatoolist/wowapi/internal/testmodules/requests | 0 | `internal/testmodules/requests/module.go:21` |
| RequestDTO | github.com/qatoolist/wowapi/internal/testmodules/requests | 3 | `internal/testmodules/requests/dto.go:15` |
| lookupImpl | github.com/qatoolist/wowapi/internal/testmodules/requests | 1 | `internal/testmodules/requests/module.go:71` |
| budget | github.com/qatoolist/wowapi/internal/tools/benchbudget | 2 | `internal/tools/benchbudget/main.go:80` |
| result | github.com/qatoolist/wowapi/internal/tools/benchbudget | 2 | `internal/tools/benchbudget/main.go:86` |
| Deps | github.com/qatoolist/wowapi/kernel | 10 | `kernel/kernel.go:129` |
| Kernel | github.com/qatoolist/wowapi/kernel | 32 | `kernel/kernel.go:52` |
| durableAudit | github.com/qatoolist/wowapi/kernel | 3 | `kernel/kernel.go:349` |
| loggingAudit | github.com/qatoolist/wowapi/kernel | 1 | `kernel/kernel.go:326` |
| secretRefResolver | github.com/qatoolist/wowapi/kernel | 1 | `kernel/kernel.go:311` |
| Authenticator | github.com/qatoolist/wowapi/kernel/apikey | 2 | `kernel/apikey/apikey.go:256` |
| KeyInfo | github.com/qatoolist/wowapi/kernel/apikey | 7 | `kernel/apikey/apikey.go:222` |
| Principal | github.com/qatoolist/wowapi/kernel/apikey | 4 | `kernel/apikey/apikey.go:63` |
| Store | github.com/qatoolist/wowapi/kernel/apikey | 3 | `kernel/apikey/apikey.go:35` |
| Artifact | github.com/qatoolist/wowapi/kernel/artifact | 8 | `kernel/artifact/artifact.go:52` |
| Input | github.com/qatoolist/wowapi/kernel/artifact | 6 | `kernel/artifact/artifact.go:42` |
| Pipeline | github.com/qatoolist/wowapi/kernel/artifact | 1 | `kernel/artifact/artifact.go:28` |
| TemplateVersion | github.com/qatoolist/wowapi/kernel/artifact | 2 | `kernel/artifact/templates.go:12` |
| Templates | github.com/qatoolist/wowapi/kernel/artifact | 1 | `kernel/artifact/templates.go:20` |
| AttachIn | github.com/qatoolist/wowapi/kernel/attachment | 4 | `kernel/attachment/attachment.go:32` |
| Attachment | github.com/qatoolist/wowapi/kernel/attachment | 9 | `kernel/attachment/attachment.go:19` |
| Service | github.com/qatoolist/wowapi/kernel/attachment | 2 | `kernel/attachment/attachment.go:40` |
| Entry | github.com/qatoolist/wowapi/kernel/audit | 10 | `kernel/audit/audit.go:36` |
| Filter | github.com/qatoolist/wowapi/kernel/audit | 5 | `kernel/audit/audit.go:314` |
| Log | github.com/qatoolist/wowapi/kernel/audit | 14 | `kernel/audit/audit.go:50` |
| VerifyResult | github.com/qatoolist/wowapi/kernel/audit | 5 | `kernel/audit/audit.go:183` |
| Writer | github.com/qatoolist/wowapi/kernel/audit | 2 | `kernel/audit/audit.go:74` |
| Authenticator | github.com/qatoolist/wowapi/kernel/auth | 2 | `kernel/auth/auth.go:213` |
| Claims | github.com/qatoolist/wowapi/kernel/auth | 5 | `kernel/auth/auth.go:58` |
| Config | github.com/qatoolist/wowapi/kernel/auth | 3 | `kernel/auth/auth.go:49` |
| JWKSConfig | github.com/qatoolist/wowapi/kernel/auth | 5 | `kernel/auth/jwks.go:48` |
| Verifier | github.com/qatoolist/wowapi/kernel/auth | 4 | `kernel/auth/auth.go:70` |
| jwk | github.com/qatoolist/wowapi/kernel/auth | 8 | `kernel/auth/jwks.go:213` |
| jwksKeySource | github.com/qatoolist/wowapi/kernel/auth | 10 | `kernel/auth/jwks.go:102` |
| staticKeySource | github.com/qatoolist/wowapi/kernel/auth | 1 | `kernel/auth/keysource.go:12` |
| Actor | github.com/qatoolist/wowapi/kernel/authz | 9 | `kernel/authz/authz.go:36` |
| Assignment | github.com/qatoolist/wowapi/kernel/authz | 6 | `kernel/authz/store.go:16` |
| CachingStore | github.com/qatoolist/wowapi/kernel/authz | 5 | `kernel/authz/caching.go:29` |
| Condition | github.com/qatoolist/wowapi/kernel/authz | 3 | `kernel/authz/store.go:44` |
| Decision | github.com/qatoolist/wowapi/kernel/authz | 4 | `kernel/authz/authz.go:79` |
| ListFilter | github.com/qatoolist/wowapi/kernel/authz | 3 | `kernel/authz/authz.go:94` |
| Options | github.com/qatoolist/wowapi/kernel/authz | 6 | `kernel/authz/evaluator.go:34` |
| Permission | github.com/qatoolist/wowapi/kernel/authz | 4 | `kernel/authz/registry.go:24` |
| PgStore | github.com/qatoolist/wowapi/kernel/authz | 0 | `kernel/authz/store_pg.go:23` |
| Policy | github.com/qatoolist/wowapi/kernel/authz | 5 | `kernel/authz/store.go:51` |
| Registry | github.com/qatoolist/wowapi/kernel/authz | 2 | `kernel/authz/registry.go:38` |
| Target | github.com/qatoolist/wowapi/kernel/authz | 3 | `kernel/authz/authz.go:70` |
| cachedAssignments | github.com/qatoolist/wowapi/kernel/authz | 2 | `kernel/authz/caching.go:38` |
| engine | github.com/qatoolist/wowapi/kernel/authz | 6 | `kernel/authz/evaluator.go:23` |
| Progress | github.com/qatoolist/wowapi/kernel/bulk | 5 | `kernel/bulk/bulk.go:32` |
| Service | github.com/qatoolist/wowapi/kernel/bulk | 1 | `kernel/bulk/bulk.go:41` |
| claimedItem | github.com/qatoolist/wowapi/kernel/bulk | 2 | `kernel/bulk/bulk.go:118` |
| Comment | github.com/qatoolist/wowapi/kernel/comment | 11 | `kernel/comment/comment.go:20` |
| CreateIn | github.com/qatoolist/wowapi/kernel/comment | 4 | `kernel/comment/comment.go:35` |
| Service | github.com/qatoolist/wowapi/kernel/comment | 2 | `kernel/comment/comment.go:43` |
| DB | github.com/qatoolist/wowapi/kernel/config | 4 | `kernel/config/config.go:81` |
| Framework | github.com/qatoolist/wowapi/kernel/config | 6 | `kernel/config/config.go:54` |
| HTTP | github.com/qatoolist/wowapi/kernel/config | 6 | `kernel/config/config.go:98` |
| Loaded | github.com/qatoolist/wowapi/kernel/config | 4 | `kernel/config/load.go:49` |
| Log | github.com/qatoolist/wowapi/kernel/config | 2 | `kernel/config/config.go:121` |
| Options | github.com/qatoolist/wowapi/kernel/config | 6 | `kernel/config/load.go:30` |
| RateLimit | github.com/qatoolist/wowapi/kernel/config | 3 | `kernel/config/config.go:114` |
| Secret | github.com/qatoolist/wowapi/kernel/config | 2 | `kernel/config/secret.go:17` |
| SharedSection | github.com/qatoolist/wowapi/kernel/config | 3 | `kernel/config/shared.go:14` |
| Telemetry | github.com/qatoolist/wowapi/kernel/config | 1 | `kernel/config/config.go:73` |
| binder | github.com/qatoolist/wowapi/kernel/config | 5 | `kernel/config/bind.go:17` |
| secretSlot | github.com/qatoolist/wowapi/kernel/config | 2 | `kernel/config/bind.go:27` |
| Manager | github.com/qatoolist/wowapi/kernel/database | 4 | `kernel/database/txmanager.go:29` |
| MigrateResult | github.com/qatoolist/wowapi/kernel/database | 2 | `kernel/database/migrate.go:17` |
| PgIdemStore | github.com/qatoolist/wowapi/kernel/database | 1 | `kernel/database/idempotency.go:43` |
| Replay | github.com/qatoolist/wowapi/kernel/database | 4 | `kernel/database/idempotency.go:16` |
| actorIDKey | github.com/qatoolist/wowapi/kernel/database | 0 | `kernel/database/context.go:16` |
| platformTx | github.com/qatoolist/wowapi/kernel/database | 1 | `kernel/database/txmanager.go:175` |
| tenantIDKey | github.com/qatoolist/wowapi/kernel/database | 0 | `kernel/database/context.go:15` |
| tenantTx | github.com/qatoolist/wowapi/kernel/database | 1 | `kernel/database/txmanager.go:159` |
| AccessEvent | github.com/qatoolist/wowapi/kernel/document | 4 | `kernel/document/hooks.go:23` |
| Class | github.com/qatoolist/wowapi/kernel/document | 6 | `kernel/document/registry.go:48` |
| ConfirmInput | github.com/qatoolist/wowapi/kernel/document | 6 | `kernel/document/service.go:90` |
| CreateInput | github.com/qatoolist/wowapi/kernel/document | 4 | `kernel/document/service.go:73` |
| Download | github.com/qatoolist/wowapi/kernel/document | 3 | `kernel/document/service.go:106` |
| DownloadInput | github.com/qatoolist/wowapi/kernel/document | 2 | `kernel/document/service.go:100` |
| GrantInput | github.com/qatoolist/wowapi/kernel/document | 5 | `kernel/document/service.go:113` |
| Hooks | github.com/qatoolist/wowapi/kernel/document | 2 | `kernel/document/hooks.go:37` |
| Registry | github.com/qatoolist/wowapi/kernel/document | 2 | `kernel/document/registry.go:62` |
| Service | github.com/qatoolist/wowapi/kernel/document | 9 | `kernel/document/service.go:46` |
| UploadEvent | github.com/qatoolist/wowapi/kernel/document | 7 | `kernel/document/hooks.go:10` |
| UploadSession | github.com/qatoolist/wowapi/kernel/document | 4 | `kernel/document/service.go:81` |
| Error | github.com/qatoolist/wowapi/kernel/errors | 6 | `kernel/errors/errors.go:90` |
| FieldError | github.com/qatoolist/wowapi/kernel/errors | 3 | `kernel/errors/errors.go:44` |
| kindInfo | github.com/qatoolist/wowapi/kernel/errors | 2 | `kernel/errors/errors.go:50` |
| Condition | github.com/qatoolist/wowapi/kernel/filtering | 4 | `kernel/filtering/filtering.go:73` |
| FieldSpec | github.com/qatoolist/wowapi/kernel/filtering | 2 | `kernel/filtering/filtering.go:61` |
| Set | github.com/qatoolist/wowapi/kernel/filtering | 1 | `kernel/filtering/filtering.go:82` |
| Sort | github.com/qatoolist/wowapi/kernel/filtering | 1 | `kernel/filtering/sort.go:29` |
| SortSpec | github.com/qatoolist/wowapi/kernel/filtering | 1 | `kernel/filtering/sort.go:15` |
| Term | github.com/qatoolist/wowapi/kernel/filtering | 2 | `kernel/filtering/sort.go:119` |
| sortKey | github.com/qatoolist/wowapi/kernel/filtering | 2 | `kernel/filtering/sort.go:23` |
| APIResponse | github.com/qatoolist/wowapi/kernel/httpx | 2 | `kernel/httpx/response.go:19` |
| AuditMeta | github.com/qatoolist/wowapi/kernel/httpx | 4 | `kernel/httpx/response.go:32` |
| CORSPolicy | github.com/qatoolist/wowapi/kernel/httpx | 6 | `kernel/httpx/edge.go:80` |
| DenyAllAuthenticator | github.com/qatoolist/wowapi/kernel/httpx | 0 | `kernel/httpx/authz_gate.go:32` |
| Health | github.com/qatoolist/wowapi/kernel/httpx | 3 | `kernel/httpx/health.go:21` |
| IdempotencyConfig | github.com/qatoolist/wowapi/kernel/httpx | 3 | `kernel/httpx/idempotency.go:44` |
| Meta | github.com/qatoolist/wowapi/kernel/httpx | 2 | `kernel/httpx/response.go:25` |
| ProblemError | github.com/qatoolist/wowapi/kernel/httpx | 8 | `kernel/httpx/errors.go:14` |
| Route | github.com/qatoolist/wowapi/kernel/httpx | 4 | `kernel/httpx/router.go:49` |
| RouteMeta | github.com/qatoolist/wowapi/kernel/httpx | 5 | `kernel/httpx/router.go:21` |
| TokenBucket | github.com/qatoolist/wowapi/kernel/httpx | 7 | `kernel/httpx/ratelimit.go:158` |
| actorKey | github.com/qatoolist/wowapi/kernel/httpx | 0 | `kernel/httpx/context.go:15` |
| compositeAuthenticator | github.com/qatoolist/wowapi/kernel/httpx | 1 | `kernel/httpx/authz_gate.go:51` |
| rateLimitCfg | github.com/qatoolist/wowapi/kernel/httpx | 1 | `kernel/httpx/ratelimit.go:38` |
| requestIDKey | github.com/qatoolist/wowapi/kernel/httpx | 0 | `kernel/httpx/context.go:14` |
| secureHeadersConfig | github.com/qatoolist/wowapi/kernel/httpx | 2 | `kernel/httpx/edge.go:23` |
| tokenBucket | github.com/qatoolist/wowapi/kernel/httpx | 3 | `kernel/httpx/ratelimit.go:149` |
| trackedWriter | github.com/qatoolist/wowapi/kernel/httpx | 2 | `kernel/httpx/middleware.go:79` |
| Config | github.com/qatoolist/wowapi/kernel/integration | 5 | `kernel/integration/integration.go:34` |
| Registry | github.com/qatoolist/wowapi/kernel/integration | 2 | `kernel/integration/integration.go:58` |
| Store | github.com/qatoolist/wowapi/kernel/integration | 3 | `kernel/integration/store.go:22` |
| UpsertIn | github.com/qatoolist/wowapi/kernel/integration | 5 | `kernel/integration/store.go:37` |
| DeadJob | github.com/qatoolist/wowapi/kernel/jobs | 5 | `kernel/jobs/runner.go:141` |
| DeadJobEntry | github.com/qatoolist/wowapi/kernel/jobs | 8 | `kernel/jobs/dlq.go:24` |
| Registry | github.com/qatoolist/wowapi/kernel/jobs | 2 | `kernel/jobs/registry.go:18` |
| RetryPolicy | github.com/qatoolist/wowapi/kernel/jobs | 2 | `kernel/jobs/jobs.go:62` |
| Scheduler | github.com/qatoolist/wowapi/kernel/jobs | 4 | `kernel/jobs/scheduler.go:23` |
| claimedJob | github.com/qatoolist/wowapi/kernel/jobs | 7 | `kernel/jobs/runner.go:292` |
| enqueueConfig | github.com/qatoolist/wowapi/kernel/jobs | 3 | `kernel/jobs/runner.go:23` |
| entry | github.com/qatoolist/wowapi/kernel/jobs | 2 | `kernel/jobs/registry.go:8` |
| task | github.com/qatoolist/wowapi/kernel/jobs | 3 | `kernel/jobs/scheduler.go:32` |
| ActorRef | github.com/qatoolist/wowapi/kernel/model | 4 | `kernel/model/model.go:110` |
| Auditable | github.com/qatoolist/wowapi/kernel/model | 4 | `kernel/model/model.go:39` |
| BaseFields | github.com/qatoolist/wowapi/kernel/model | 1 | `kernel/model/model.go:27` |
| CreatedOnly | github.com/qatoolist/wowapi/kernel/model | 2 | `kernel/model/model.go:48` |
| ExternalRef | github.com/qatoolist/wowapi/kernel/model | 2 | `kernel/model/model.go:136` |
| Money | github.com/qatoolist/wowapi/kernel/model | 2 | `kernel/model/model.go:119` |
| ResourceRef | github.com/qatoolist/wowapi/kernel/model | 2 | `kernel/model/model.go:93` |
| Statused | github.com/qatoolist/wowapi/kernel/model | 1 | `kernel/model/model.go:83` |
| Temporal | github.com/qatoolist/wowapi/kernel/model | 2 | `kernel/model/model.go:62` |
| TenantScoped | github.com/qatoolist/wowapi/kernel/model | 1 | `kernel/model/model.go:33` |
| TimeRange | github.com/qatoolist/wowapi/kernel/model | 2 | `kernel/model/model.go:125` |
| Versioned | github.com/qatoolist/wowapi/kernel/model | 1 | `kernel/model/model.go:55` |
| uuidV7Gen | github.com/qatoolist/wowapi/kernel/model | 0 | `kernel/model/model.go:152` |
| ChannelDest | github.com/qatoolist/wowapi/kernel/notify | 2 | `kernel/notify/service.go:57` |
| Delivery | github.com/qatoolist/wowapi/kernel/notify | 9 | `kernel/notify/service.go:97` |
| DeliveryReceipt | github.com/qatoolist/wowapi/kernel/notify | 9 | `kernel/notify/service.go:360` |
| FakeSender | github.com/qatoolist/wowapi/kernel/notify | 3 | `kernel/notify/sender.go:23` |
| Message | github.com/qatoolist/wowapi/kernel/notify | 7 | `kernel/notify/service.go:63` |
| Notification | github.com/qatoolist/wowapi/kernel/notify | 11 | `kernel/notify/service.go:79` |
| Registry | github.com/qatoolist/wowapi/kernel/notify | 2 | `kernel/notify/registry.go:68` |
| Service | github.com/qatoolist/wowapi/kernel/notify | 5 | `kernel/notify/service.go:113` |
| TemplateSpec | github.com/qatoolist/wowapi/kernel/notify | 3 | `kernel/notify/registry.go:51` |
| inAppSender | github.com/qatoolist/wowapi/kernel/notify | 0 | `kernel/notify/sender.go:60` |
| noOp | github.com/qatoolist/wowapi/kernel/observability | 0 | `kernel/observability/metrics.go:42` |
| noopSpan | github.com/qatoolist/wowapi/kernel/observability | 0 | `kernel/observability/tracing.go:54` |
| noopTracer | github.com/qatoolist/wowapi/kernel/observability | 0 | `kernel/observability/tracing.go:46` |
| statusWriter | github.com/qatoolist/wowapi/kernel/observability | 3 | `kernel/observability/middleware.go:66` |
| DeadEventEntry | github.com/qatoolist/wowapi/kernel/outbox | 8 | `kernel/outbox/dlq.go:25` |
| DispatchedEvent | github.com/qatoolist/wowapi/kernel/outbox | 7 | `kernel/outbox/outbox.go:124` |
| Event | github.com/qatoolist/wowapi/kernel/outbox | 7 | `kernel/outbox/outbox.go:25` |
| HandlerRegistry | github.com/qatoolist/wowapi/kernel/outbox | 3 | `kernel/outbox/outbox.go:142` |
| Relay | github.com/qatoolist/wowapi/kernel/outbox | 5 | `kernel/outbox/relay.go:32` |
| pgWriter | github.com/qatoolist/wowapi/kernel/outbox | 2 | `kernel/outbox/outbox.go:67` |
| row | github.com/qatoolist/wowapi/kernel/outbox | 10 | `kernel/outbox/relay.go:68` |
| subscription | github.com/qatoolist/wowapi/kernel/outbox | 3 | `kernel/outbox/outbox.go:135` |
| Cursor | github.com/qatoolist/wowapi/kernel/pagination | 2 | `kernel/pagination/cursor.go:38` |
| CursorPage | github.com/qatoolist/wowapi/kernel/pagination | 3 | `kernel/pagination/pagination.go:37` |
| Defaults | github.com/qatoolist/wowapi/kernel/pagination | 2 | `kernel/pagination/pagination.go:48` |
| PageResponse | github.com/qatoolist/wowapi/kernel/pagination | 4 | `kernel/pagination/pagination.go:27` |
| Request | github.com/qatoolist/wowapi/kernel/pagination | 2 | `kernel/pagination/pagination.go:55` |
| Engine | github.com/qatoolist/wowapi/kernel/policy | 0 | `kernel/policy/policy.go:21` |
| Checker | github.com/qatoolist/wowapi/kernel/relationship | 0 | `kernel/relationship/relationship.go:35` |
| PgRegistrar | github.com/qatoolist/wowapi/kernel/resource | 0 | `kernel/resource/registrar_pg.go:17` |
| Ref | github.com/qatoolist/wowapi/kernel/resource | 2 | `kernel/resource/resource.go:21` |
| Registry | github.com/qatoolist/wowapi/kernel/resource | 2 | `kernel/resource/resource.go:45` |
| TypeSpec | github.com/qatoolist/wowapi/kernel/resource | 2 | `kernel/resource/resource.go:37` |
| boundRegistrar | github.com/qatoolist/wowapi/kernel/resource | 1 | `kernel/resource/registrar_pg.go:30` |
| DSR | github.com/qatoolist/wowapi/kernel/retention | 1 | `kernel/retention/dsr.go:20` |
| Engine | github.com/qatoolist/wowapi/kernel/retention | 2 | `kernel/retention/engine.go:90` |
| Hold | github.com/qatoolist/wowapi/kernel/retention | 4 | `kernel/retention/retention.go:37` |
| Holds | github.com/qatoolist/wowapi/kernel/retention | 1 | `kernel/retention/retention.go:24` |
| RecordClass | github.com/qatoolist/wowapi/kernel/retention | 5 | `kernel/retention/engine.go:39` |
| Registry | github.com/qatoolist/wowapi/kernel/retention | 2 | `kernel/retention/engine.go:48` |
| Request | github.com/qatoolist/wowapi/kernel/retention | 5 | `kernel/retention/dsr.go:41` |
| Point | github.com/qatoolist/wowapi/kernel/rules | 7 | `kernel/rules/rules.go:36` |
| Proposal | github.com/qatoolist/wowapi/kernel/rules | 5 | `kernel/rules/store.go:20` |
| Registry | github.com/qatoolist/wowapi/kernel/rules | 2 | `kernel/rules/rules.go:58` |
| Resolved | github.com/qatoolist/wowapi/kernel/rules | 5 | `kernel/rules/resolver.go:15` |
| Resolver | github.com/qatoolist/wowapi/kernel/rules | 2 | `kernel/rules/resolver.go:38` |
| Store | github.com/qatoolist/wowapi/kernel/rules | 2 | `kernel/rules/store.go:30` |
| Ref | github.com/qatoolist/wowapi/kernel/secrets | 2 | `kernel/secrets/secrets.go:20` |
| Bundle | github.com/qatoolist/wowapi/kernel/seeds | 4 | `kernel/seeds/seeds.go:23` |
| PermissionSeed | github.com/qatoolist/wowapi/kernel/seeds | 4 | `kernel/seeds/seeds.go:31` |
| RelationshipTypeSeed | github.com/qatoolist/wowapi/kernel/seeds | 5 | `kernel/seeds/seeds.go:53` |
| ResourceTypeSeed | github.com/qatoolist/wowapi/kernel/seeds | 2 | `kernel/seeds/seeds.go:47` |
| RoleSeed | github.com/qatoolist/wowapi/kernel/seeds | 3 | `kernel/seeds/seeds.go:40` |
| Allocation | github.com/qatoolist/wowapi/kernel/sequence | 2 | `kernel/sequence/sequence.go:41` |
| Allocator | github.com/qatoolist/wowapi/kernel/sequence | 1 | `kernel/sequence/sequence.go:28` |
| Memory | github.com/qatoolist/wowapi/kernel/storage | 3 | `kernel/storage/memory.go:14` |
| ObjectInfo | github.com/qatoolist/wowapi/kernel/storage | 2 | `kernel/storage/storage.go:25` |
| PresignedURL | github.com/qatoolist/wowapi/kernel/storage | 3 | `kernel/storage/storage.go:17` |
| Validator | github.com/qatoolist/wowapi/kernel/validation | 1 | `kernel/validation/validation.go:43` |
| Endpoint | github.com/qatoolist/wowapi/kernel/webhook | 9 | `kernel/webhook/webhook.go:53` |
| Event | github.com/qatoolist/wowapi/kernel/webhook | 13 | `kernel/webhook/webhook.go:66` |
| FakeSecretResolver | github.com/qatoolist/wowapi/kernel/webhook | 1 | `kernel/webhook/sender.go:70` |
| FakeSender | github.com/qatoolist/wowapi/kernel/webhook | 3 | `kernel/webhook/sender.go:42` |
| FakeVerifier | github.com/qatoolist/wowapi/kernel/webhook | 1 | `kernel/webhook/verifier.go:63` |
| HMACVerifier | github.com/qatoolist/wowapi/kernel/webhook | 1 | `kernel/webhook/verifier.go:23` |
| HTTPSender | github.com/qatoolist/wowapi/kernel/webhook | 1 | `kernel/webhook/sender.go:13` |
| InboundIn | github.com/qatoolist/wowapi/kernel/webhook | 7 | `kernel/webhook/webhook.go:84` |
| SentCall | github.com/qatoolist/wowapi/kernel/webhook | 3 | `kernel/webhook/sender.go:53` |
| breakerRegistry | github.com/qatoolist/wowapi/kernel/webhook | 3 | `kernel/webhook/breaker.go:86` |
| breakerState | github.com/qatoolist/wowapi/kernel/webhook | 4 | `kernel/webhook/breaker.go:17` |
| insertEventParams | github.com/qatoolist/wowapi/kernel/webhook | 8 | `kernel/webhook/service.go:458` |
| Assignee | github.com/qatoolist/wowapi/kernel/workflow | 2 | `kernel/workflow/registry.go:50` |
| AssigneeSpec | github.com/qatoolist/wowapi/kernel/workflow | 6 | `kernel/workflow/definition.go:111` |
| AutoInput | github.com/qatoolist/wowapi/kernel/workflow | 4 | `kernel/workflow/registry.go:14` |
| Branch | github.com/qatoolist/wowapi/kernel/workflow | 2 | `kernel/workflow/definition.go:154` |
| Condition | github.com/qatoolist/wowapi/kernel/workflow | 2 | `kernel/workflow/definition.go:160` |
| Decision | github.com/qatoolist/wowapi/kernel/workflow | 3 | `kernel/workflow/runtime.go:40` |
| Definition | github.com/qatoolist/wowapi/kernel/workflow | 5 | `kernel/workflow/definition.go:70` |
| Electorate | github.com/qatoolist/wowapi/kernel/workflow | 3 | `kernel/workflow/definition.go:166` |
| Fraction | github.com/qatoolist/wowapi/kernel/workflow | 2 | `kernel/workflow/definition.go:173` |
| Instance | github.com/qatoolist/wowapi/kernel/workflow | 7 | `kernel/workflow/runtime.go:47` |
| Policy | github.com/qatoolist/wowapi/kernel/workflow | 2 | `kernel/workflow/definition.go:121` |
| Registry | github.com/qatoolist/wowapi/kernel/workflow | 5 | `kernel/workflow/registry.go:64` |
| ResolveInput | github.com/qatoolist/wowapi/kernel/workflow | 4 | `kernel/workflow/registry.go:27` |
| Runtime | github.com/qatoolist/wowapi/kernel/workflow | 6 | `kernel/workflow/runtime.go:75` |
| SLA | github.com/qatoolist/wowapi/kernel/workflow | 3 | `kernel/workflow/definition.go:127` |
| Step | github.com/qatoolist/wowapi/kernel/workflow | 15 | `kernel/workflow/definition.go:80` |
| Task | github.com/qatoolist/wowapi/kernel/workflow | 11 | `kernel/workflow/runtime.go:58` |
| Transition | github.com/qatoolist/wowapi/kernel/workflow | 4 | `kernel/workflow/definition.go:134` |
| DBHandle | github.com/qatoolist/wowapi/testkit | 6 | `testkit/db.go:47` |
| TenantHandle | github.com/qatoolist/wowapi/testkit | 1 | `testkit/fixtures.go:20` |
| TokenIssuer | github.com/qatoolist/wowapi/testkit | 2 | `testkit/auth.go:27` |
| WorkflowSim | github.com/qatoolist/wowapi/testkit | 5 | `testkit/workflowsim.go:26` |
| discard | github.com/qatoolist/wowapi/testkit | 0 | `testkit/contract.go:187` |
| tokenConfig | github.com/qatoolist/wowapi/testkit | 5 | `testkit/auth.go:56` |
| Clock | github.com/qatoolist/wowapi/testkit/fakes | 2 | `testkit/fakes/clock.go:15` |
| IDGen | github.com/qatoolist/wowapi/testkit/fakes | 3 | `testkit/fakes/idgen.go:17` |

## Named Types

| Type | Package | Location |
| --- | --- | --- |
| BootOption | github.com/qatoolist/wowapi/app | `app/boot.go:50` |
| workerErr | github.com/qatoolist/wowapi/app | `app/worker.go:150` |
| StoreOption | github.com/qatoolist/wowapi/kernel/apikey | `kernel/apikey/apikey.go:42` |
| Redactor | github.com/qatoolist/wowapi/kernel/audit | `kernel/audit/audit.go:70` |
| ActorKind | github.com/qatoolist/wowapi/kernel/authz | `kernel/authz/authz.go:25` |
| PolicyEffect | github.com/qatoolist/wowapi/kernel/authz | `kernel/authz/store.go:36` |
| ScopeKind | github.com/qatoolist/wowapi/kernel/authz | `kernel/authz/authz.go:60` |
| ItemFunc | github.com/qatoolist/wowapi/kernel/bulk | `kernel/bulk/bulk.go:29` |
| Env | github.com/qatoolist/wowapi/kernel/config | `kernel/config/config.go:24` |
| Fingerprint | github.com/qatoolist/wowapi/kernel/config | `kernel/config/fingerprint.go:18` |
| Layer | github.com/qatoolist/wowapi/kernel/config | `kernel/config/load.go:15` |
| MapView | github.com/qatoolist/wowapi/kernel/config | `kernel/config/moduleview.go:23` |
| Namespaces | github.com/qatoolist/wowapi/kernel/config | `kernel/config/moduleview.go:31` |
| Provenance | github.com/qatoolist/wowapi/kernel/config | `kernel/config/load.go:27` |
| ManagerOption | github.com/qatoolist/wowapi/kernel/database | `kernel/database/txmanager.go:37` |
| Option | github.com/qatoolist/wowapi/kernel/database | `kernel/database/database.go:48` |
| AccessHook | github.com/qatoolist/wowapi/kernel/document | `kernel/document/hooks.go:33` |
| Sensitivity | github.com/qatoolist/wowapi/kernel/document | `kernel/document/registry.go:23` |
| UploadHook | github.com/qatoolist/wowapi/kernel/document | `kernel/document/hooks.go:32` |
| Kind | github.com/qatoolist/wowapi/kernel/errors | `kernel/errors/errors.go:20` |
| opString | github.com/qatoolist/wowapi/kernel/errors | `kernel/errors/errors.go:125` |
| Allowlist | github.com/qatoolist/wowapi/kernel/filtering | `kernel/filtering/filtering.go:68` |
| Dir | github.com/qatoolist/wowapi/kernel/filtering | `kernel/filtering/sort.go:6` |
| Op | github.com/qatoolist/wowapi/kernel/filtering | `kernel/filtering/filtering.go:32` |
| SortAllowlist | github.com/qatoolist/wowapi/kernel/filtering | `kernel/filtering/sort.go:20` |
| HealthCheck | github.com/qatoolist/wowapi/kernel/httpx | `kernel/httpx/health.go:18` |
| Middleware | github.com/qatoolist/wowapi/kernel/httpx | `kernel/httpx/middleware.go:13` |
| Operation | github.com/qatoolist/wowapi/kernel/httpx | `kernel/httpx/idempotency.go:41` |
| RateLimitOption | github.com/qatoolist/wowapi/kernel/httpx | `kernel/httpx/ratelimit.go:36` |
| ScopeExtractor | github.com/qatoolist/wowapi/kernel/httpx | `kernel/httpx/router.go:16` |
| SecureHeadersOption | github.com/qatoolist/wowapi/kernel/httpx | `kernel/httpx/edge.go:29` |
| BackoffPolicy | github.com/qatoolist/wowapi/kernel/jobs | `kernel/jobs/jobs.go:55` |
| Opt | github.com/qatoolist/wowapi/kernel/jobs | `kernel/jobs/runner.go:30` |
| RunnerOpt | github.com/qatoolist/wowapi/kernel/jobs | `kernel/jobs/runner.go:172` |
| Worker | github.com/qatoolist/wowapi/kernel/jobs | `kernel/jobs/jobs.go:49` |
| ActorKind | github.com/qatoolist/wowapi/kernel/model | `kernel/model/model.go:99` |
| Metadata | github.com/qatoolist/wowapi/kernel/model | `kernel/model/model.go:133` |
| Channel | github.com/qatoolist/wowapi/kernel/notify | `kernel/notify/registry.go:24` |
| Importance | github.com/qatoolist/wowapi/kernel/notify | `kernel/notify/registry.go:35` |
| Option | github.com/qatoolist/wowapi/kernel/notify | `kernel/notify/service.go:122` |
| Handler | github.com/qatoolist/wowapi/kernel/outbox | `kernel/outbox/outbox.go:120` |
| RelayOption | github.com/qatoolist/wowapi/kernel/outbox | `kernel/outbox/relay.go:41` |
| WriterOption | github.com/qatoolist/wowapi/kernel/outbox | `kernel/outbox/outbox.go:44` |
| DisposeFunc | github.com/qatoolist/wowapi/kernel/retention | `kernel/retention/engine.go:26` |
| EraseFunc | github.com/qatoolist/wowapi/kernel/retention | `kernel/retention/engine.go:34` |
| ExportFunc | github.com/qatoolist/wowapi/kernel/retention | `kernel/retention/engine.go:30` |
| Kind | github.com/qatoolist/wowapi/kernel/retention | `kernel/retention/dsr.go:33` |
| OrgAncestry | github.com/qatoolist/wowapi/kernel/rules | `kernel/rules/resolver.go:34` |
| ScopeKind | github.com/qatoolist/wowapi/kernel/rules | `kernel/rules/rules.go:24` |
| InboundHandler | github.com/qatoolist/wowapi/kernel/webhook | `kernel/webhook/webhook.go:116` |
| Option | github.com/qatoolist/wowapi/kernel/webhook | `kernel/webhook/webhook.go:135` |
| AssigneeResolver | github.com/qatoolist/wowapi/kernel/workflow | `kernel/workflow/registry.go:57` |
| AutoAction | github.com/qatoolist/wowapi/kernel/workflow | `kernel/workflow/registry.go:24` |
| DecisionType | github.com/qatoolist/wowapi/kernel/workflow | `kernel/workflow/runtime.go:27` |
| ResolvedKind | github.com/qatoolist/wowapi/kernel/workflow | `kernel/workflow/registry.go:36` |
| StepType | github.com/qatoolist/wowapi/kernel/workflow | `kernel/workflow/definition.go:30` |
| RowFactory | github.com/qatoolist/wowapi/testkit | `testkit/asserts.go:20` |
| TokenOption | github.com/qatoolist/wowapi/testkit | `testkit/auth.go:66` |

