------- GOAL ACCOMPLISHED - DO NOT REWORK --------

Your goal is to design a **production-ready, reusable, domain-agnostic enterprise backend framework / platform kernel in Go/Golang**.

This is **not** a housing society management application yet.

The housing society / apartment society / co-operative society management product is only the **first intended use case** and should be treated strictly as a **reference domain**, not as the framework foundation.

The framework must be reusable for different multi-tenant, multi-actor, workflow-heavy, compliance-ready enterprise applications such as:

* Housing society management
* Apartment / gated community management
* Facility management
* Clubs and associations
* Schools or institutions
* NGOs
* Vendor management systems
* Case management systems
* Membership-based SaaS products
* Compliance-heavy internal enterprise tools
* Approval-heavy enterprise operations
* Document-heavy regulated applications

The goal is to design the **core reusable framework first**: the common structures, helpers, utilities, middleware, module contracts, base primitives, reusable services, scaffolding, extension points, security controls, performance guardrails, developer tooling, test utilities, and operational foundation.

The framework code should be reusable as-is for another product without needing to remove housing-society-specific concepts.

Do **not** design society-specific product tables or modules inside the core framework.

The core framework must not contain hard-coded concepts such as:

* building
* wing
* flat
* member
* associate member
* nominal member
* society committee
* chairman
* secretary
* treasurer
* maintenance bill
* defaulter
* AGM
* society notice
* parking allocation
* water / STP / WTP
* visitor gate entry
* conveyance
* redevelopment
* elections under society bye-laws

Those should be created later as **domain modules** using the framework’s generic extension system.

The framework itself should contain only generic platform primitives such as:

* tenant
* organization
* party / person / legal entity
* user
* actor
* acting capacity
* role
* permission
* policy
* relationship
* resource
* assignment
* workflow
* rule
* document
* comment
* attachment
* audit log
* event
* background job
* notification
* webhook
* integration provider
* file metadata
* idempotency key
* module registration
* common API helpers
* common database helpers
* common testkit

The final output should be **coding-ready** for building the reusable framework core, not just an architecture explanation.

Avoid generic enterprise theory.
Avoid society-specific implementation inside the framework core.
Avoid unnecessary abstractions.
Use practical Go patterns.
Prefer composition over inheritance.
Use explicit constructor injection / composition root instead of hidden service locators.
Use interfaces at boundaries, not everywhere.
Use generics only where they reduce real duplication.
Use code generation/templates only where they improve developer experience without hiding business logic.

---

# 1. Core Framework Objective

Design a reusable backend framework in Go that provides the foundation for building enterprise SaaS applications.

The framework must support:

* Multi-tenancy
* Organization hierarchy
* Identity and authentication
* Actor model
* Acting capacity
* Role and permission system
* Relationship-based access
* Policy-based restrictions
* Record-level access control
* Workflow and approval engine
* Rule/configuration engine
* Document and file management
* Comments and attachments
* Audit logging
* Event/outbox system
* Background jobs
* Notifications
* Webhooks
* Integration adapters
* API conventions
* Error handling
* Validation
* Observability
* Security
* Performance optimization
* Practical parallelization
* Testing utilities
* Module/plugin system
* Developer scaffolding
* Migration and seed management
* Developer experience tooling

The framework should be optimized for:

* Security by default
* Performance by default
* Maintainability
* Reusability
* Extensibility
* Scalability
* Observability
* Testability
* Low-boilerplate development
* Safe module boundaries
* Long-term enterprise adoption
* Operational simplicity
* Compliance readiness

---

# 2. Recommended Architecture Style

Recommend the architecture for this reusable framework.

Compare only where useful:

* Modular monolith
* Clean architecture
* Hexagonal architecture
* Layered architecture
* Plugin/module architecture
* Event-driven architecture
* Microservices later

Give a clear recommendation for the first version.

Explain:

* Why the framework should start as a modular monolith.
* What belongs in the framework core.
* What belongs in domain modules.
* What belongs in shared utilities.
* What belongs in adapters.
* How modules should plug into the framework.
* How domain logic is prevented from leaking into the framework.
* How future modules can be extracted into services if needed.
* How to avoid big-ball-of-mud architecture.
* How to avoid premature microservices.
* How to avoid over-engineering.

Provide:

* High-level architecture diagram in text/ASCII form.
* Framework layer diagram.
* Module registration diagram.
* Request lifecycle diagram.
* Background job/event lifecycle diagram.

---

# 3. Framework Layering

Design the framework in clear layers.

Use this separation:

## 3.1 Platform Kernel

The Platform Kernel must be reusable and domain-neutral.

It may include:

* tenant context
* tenant resolver
* auth middleware
* authorization evaluator
* policy evaluator
* relationship framework
* resource registry
* workflow runtime
* rule registry
* audit logger
* outbox/event system
* job runner
* notification dispatcher
* document/file service
* webhook service
* integration provider registry
* API response/error helpers
* validation helpers
* pagination/filtering/sorting helpers
* logging/tracing/metrics
* configuration
* database helpers
* transaction manager
* RLS helpers
* migration helpers
* seed loader
* testkit

It must not know about society management, schools, clubs, facilities, or any domain-specific business rules.

## 3.2 Domain Extension Layer

The Domain Extension Layer should allow product modules to define:

* resource types
* relationship types
* actor types
* role templates
* permission catalog
* policy conditions
* workflow definitions
* rule points
* document classes
* notification templates
* event schemas
* validation rules
* lifecycle states
* module APIs
* module migrations
* module seeds
* module jobs
* module webhooks
* module reports

## 3.3 Product Domain Modules

Product domain modules implement business-specific behavior.

Examples:

* society module
* school module
* club module
* facility module
* vendor management module
* billing module
* membership module
* work-order module
* request/approval module
* case management module

Explain how these layers communicate.

---

# 4. Generic Domain Model

Design a domain-agnostic model.

Define each concept clearly:

* Tenant
* Organization
* Workspace / Project / Account, if useful
* Party
* Person
* Legal entity
* User
* Actor
* Acting capacity
* Role
* Permission
* Policy
* Policy condition
* Relationship
* Relationship type
* Resource
* Resource type
* Resource instance
* Assignment
* Scope
* Workflow definition
* Workflow instance
* Workflow task
* Rule definition
* Rule version
* Feature flag
* Document
* Document version
* Document access grant
* Comment
* Attachment
* Event
* Job
* Notification
* Audit log
* Integration provider
* Webhook endpoint
* Webhook event
* Idempotency key

For each concept, provide:

* Meaning
* What it is not
* Key fields
* Important rules
* Example
* Whether it belongs in core framework or domain module

Important:

Do not assume the domain has buildings, flats, residents, members, committees, or society billing.

Use generic terms.

Then briefly explain how a future society module can define building, unit, owner, tenant, member, notice, complaint, bill, committee role, etc. using the generic framework without modifying the core.

---

# 5. Multi-Tenancy Framework

Design generic SaaS multi-tenancy.

Cover:

* Tenant as isolation boundary
* Organization as business entity inside tenant
* Parent/child tenant support
* Multiple organizations under one tenant
* Cross-tenant access grants
* Tenant-aware authentication
* Tenant-aware authorization
* Tenant-aware documents
* Tenant-aware audit logs
* Tenant-aware background jobs
* Tenant-aware rules
* Tenant-aware workflows
* Tenant-aware events
* Tenant-aware integrations
* Tenant-aware configuration
* Shared database with `tenant_id`
* PostgreSQL RLS strategy
* `SET LOCAL app.tenant_id` per transaction
* Future dedicated schema/database escape hatch
* Tenant data migration strategy
* Tenant isolation testing strategy
* Tenant-aware rate limiting
* Tenant-aware file storage
* Tenant-aware observability

This must be generic and must not assume that a housing society is always the tenant.

---

# 6. Actor, Relationship, Role, Permission, and Policy Framework

Design a reusable authorization and access-control framework.

Cover:

* RBAC
* ABAC
* ReBAC
* Policy conditions
* Actor assignment
* Acting capacity
* Delegation
* Temporary access
* Record-level access
* Scope hierarchy
* Tenant scope
* Organization scope
* Resource scope
* Record scope
* Permission naming convention
* Role naming convention
* Policy evaluation flow
* Deny-by-default model
* Permission-denial audit
* Emergency access
* Break-glass access
* Admin impersonation
* System actors
* Webhook actors
* Background worker actors

Use generic examples:

* Organization admin manages users.
* Manager approves a request.
* User can view resources related to them.
* External auditor gets read-only access for a limited period.
* Vendor can update only assigned work items.
* System worker can process background jobs.
* Webhook actor can only ingest verified callbacks.
* A user with multiple capacities must choose which capacity they act under.

Then briefly show how a society module could later define owner, tenant, secretary, treasurer, committee member, auditor, or vendor using the same framework.

Provide:

* Permission naming convention.
* Role naming convention.
* Scope model.
* Acting capacity model.
* Actor assignment model.
* Policy evaluation algorithm.
* Example permission matrix using generic actors.
* Go interface sketches for the authorization evaluator.

---

# 7. Generic Workflow Engine

Design a reusable workflow and approval engine.

It should support:

* Workflow definitions
* Workflow instances
* Workflow tasks
* Step types
* Assignee rules
* Actor-based assignment
* Role-based assignment
* Relationship-based assignment
* Resource-based assignment
* Sequential approval
* Parallel approval
* Voting/quorum-style approval as a generic capability
* Auto steps
* Task steps
* Human approval steps
* Rejection paths
* Delegation
* Escalation
* SLA / due date
* Reminders
* Comments
* Attachments
* Audit trail
* Emergency override
* Ratification
* Events on workflow transitions
* Workflow versioning
* Workflow template seeding
* Tenant-specific workflow overrides
* Workflow state validation
* Workflow test runner

Do not design society-specific workflows in the core.

Use generic examples:

* Access request approval
* Document approval
* Payment approval
* Vendor onboarding approval
* Policy change approval
* Data correction approval
* Resource lifecycle approval

Then briefly show how a society module could define membership approval, notice approval, bill approval, or complaint escalation as workflow configuration.

Recommend whether the framework should initially use:

* a small custom Postgres-backed workflow engine,
* Temporal,
* Camunda,
* Zeebe,
* Conductor,
* or another approach.

Give a practical recommendation and trade-offs.

---

# 8. Generic Rule / Configuration Engine

Design a reusable rule/configuration framework.

It should support:

* Rule definitions
* Rule points
* Rule versions
* Effective dates
* Tenant overrides
* Organization overrides
* Platform defaults
* Template bundles
* Approval before activation
* Historical evaluation
* Typed validation schemas
* Audit trail
* Fallback resolution
* Feature flags
* Configuration bundles
* Safe rollout
* Rule deactivation
* Rule supersession
* Rule change workflow

Do not hard-code housing society bye-law values.

Use generic examples:

* approval threshold
* late fee
* document retention
* SLA duration
* access restriction
* notification preference
* workflow escalation rule
* feature enablement
* data retention period
* upload size limits
* external integration settings

Then explain how a society module can later add rule points such as AGM notice period, maintenance billing frequency, defaulter threshold, or parking eligibility.

Provide:

* Rule naming convention.
* Rule schema strategy.
* Rule resolution algorithm.
* Rule versioning model.
* Example rule records.
* Go interface sketches for rule resolver and rule registry.

---

# 9. Generic Data Architecture

Recommend the database and persistence model.

Design only the **framework foundation tables**, not society product tables.

Include tables/entities such as:

* tenants
* organizations
* parties
* persons
* legal_entities
* party_contacts
* users
* user_tenant_access
* resource_types
* resources
* relationship_types
* relationships
* roles
* permissions
* role_permissions
* actor_assignments
* policies
* policy_conditions
* rule_definitions
* rule_versions
* feature_flags
* workflow_definitions
* workflow_instances
* workflow_tasks
* documents
* document_versions
* document_access_grants
* comments
* attachments
* notifications
* notification_templates
* notification_deliveries
* audit_logs
* events_outbox
* jobs
* job_runs
* idempotency_keys
* integration_providers
* webhook_endpoints
* webhook_events

For each table, provide:

* Purpose
* Key columns
* Primary key
* Foreign keys
* Unique constraints
* Important indexes
* Tenant-scoped or global
* Temporal or not
* Append-only or not
* Soft-delete / valid-to / hard-delete strategy
* RLS considerations
* Audit considerations

Provide:

* Text ERD.
* Important invariants.
* Recommended UUID strategy.
* Audit columns.
* Optimistic locking strategy.
* Temporal validity strategy.
* Append-only strategy.
* RLS strategy.
* Migration order.
* Notes on where JSONB is appropriate.
* Notes on where JSONB should be avoided.

Explain how a domain module can add its own domain-specific tables while still using framework services.

---

# 10. Base Model Primitives

Design reusable model primitives.

Do not create a bloated universal `BaseModel`.

Since this is Go, prefer composition and embedded structs.

Design reusable embedded structs such as:

* `BaseFields`
* `TenantScoped`
* `Auditable`
* `Versioned`
* `Temporal`
* `SoftDeletable`
* `Statused`
* `AppendOnly`
* `ResourceRef`
* `ActorRef`
* `Money`
* `TimeRange`
* `Ownership` only as generic ownership of a resource if appropriate, not housing-specific ownership
* `Metadata`
* `ExternalRef`

For each primitive, provide:

* Purpose
* Fields
* When to embed
* When not to embed
* Database convention
* Go struct example
* Anti-pattern to avoid

Explain:

* Which structs should be embedded in domain models.
* Which should exist only as database conventions.
* Which should not be embedded blindly in every model.
* How to avoid a bloated universal base model.
* How to keep models explicit but not repetitive.

---

# 11. Base DTO and API Response Primitives

Design reusable DTO and response patterns.

Include:

* Standard success response
* Standard error response
* Field validation error response
* Pagination response
* Cursor response
* Bulk operation response
* Long-running operation response
* File upload response
* Webhook response
* Audit metadata response

Provide Go struct examples:

* `APIResponse[T]`
* `PageResponse[T]`
* `CursorPage[T]`
* `ProblemError`
* `FieldError`
* `OperationResponse`
* `UploadSessionResponse`
* `WebhookAck`
* `AuditMeta`

Explain how handlers should use them.

---

# 12. Base Handler / Controller Pattern

Design reusable HTTP handler patterns.

Do not create a giant generic controller that hides all business logic.

Design small reusable helpers such as:

* `DecodeJSON`
* `Validate`
* `WriteJSON`
* `WriteError`
* `RequirePermission`
* `ResolveTenant`
* `ResolveActingCapacity`
* `WithTenantTx`
* `WithIdempotency`
* `ParsePagination`
* `ParseFilters`
* `ParseSort`
* `ParseResourceID`
* `BindAndValidate`
* `AuditAction`
* `HandleFileUpload`
* `HandleWebhook`

Cover:

* Handler dependencies
* Request decoding
* Validation
* Tenant extraction
* Acting capacity extraction
* Authorization check
* Transaction wrapper
* Service call
* Response writing
* Error mapping
* Audit logging
* Idempotency handling

Provide a recommended handler pattern.

Show how a module handler can stay small by using these helpers.

---

# 13. Repository and Persistence Helpers

Design reusable repository patterns.

Cover:

* Base repository dependencies
* Tenant-aware query execution
* Transaction manager
* Unit of work
* RLS `SET LOCAL` helper
* Optimistic locking helper
* Idempotency helper
* Outbox write helper
* Audit write helper
* Pagination query helper
* Filtering/sorting allowlist helper
* Soft-delete helper
* Temporal validity query helper
* Append-only insert helper
* Batch insert helper
* Query timeout helper

Provide interface examples:

* `TxManager`
* `UnitOfWork`
* `TenantDB`
* `Repository`
* `OutboxWriter`
* `AuditWriter`
* `IdempotencyStore`
* `Paginator`
* `FilterBuilder`
* `SortBuilder`

Explain how repositories should avoid:

* leaking SQL outside the module,
* cross-module joins,
* bypassing tenant scope,
* bypassing RLS,
* updating append-only tables,
* skipping optimistic locking,
* building unsafe dynamic SQL.

---

# 14. Base Service / Use Case Pattern

Design reusable application service conventions.

Cover:

* Command objects
* Query objects
* Result objects
* Use case interfaces
* Validation boundary
* Authorization boundary
* Transaction boundary
* Domain event emission
* Audit event recording
* Workflow triggering
* Rule resolution
* Error wrapping
* Idempotency
* Context cancellation
* Dependency injection

Provide naming conventions:

* `CreateXCommand`
* `UpdateXCommand`
* `ListXQuery`
* `XResult`
* `XService`
* `XRepository`
* `XPort`

Explain what belongs in:

* handler,
* application service,
* domain model,
* repository,
* platform service,
* event handler,
* background job.

---

# 15. Generic CRUD Scaffolding

Design optional CRUD scaffolding for simple resource modules.

It should support:

* create
* read
* list
* update
* deactivate
* restore, if applicable
* audit
* optimistic locking
* tenant scoping
* permission enforcement
* pagination
* filtering
* sorting
* validation
* OpenAPI generation

Clearly explain:

* Which modules can use generic CRUD.
* Which modules should not use generic CRUD.
* Why financial, workflow, audit, security, and complex domain modules need explicit logic.
* How to avoid building a weak low-code framework too early.

Provide a practical recommendation:

* Use scaffolding/code generation for repetitive CRUD.
* Keep complex domain workflows explicit.

---

# 16. Module Starter Template

Design a module template that engineers can copy for every new domain module.

Use this structure or improve it:

```text
/internal/modules/example/
  module.go
  domain/
    model.go
    errors.go
    events.go
    validation.go
  app/
    commands.go
    queries.go
    service.go
    ports.go
  api/
    routes.go
    handlers.go
    dto.go
    mapper.go
  store/
    queries.sql
    repository.go
    sqlc/
  seeds/
    permissions.yaml
    roles.yaml
    resource_types.yaml
    relationship_types.yaml
    workflows.yaml
    rules.yaml
    document_classes.yaml
    notification_templates.yaml
  migrations/
  tests/
```

For each file/folder, explain:

* purpose,
* what belongs there,
* what must not belong there.

---

# 17. Module Registration and Bootstrapping

Design a common module registration contract.

A module should be able to register:

* module name
* dependencies
* routes
* permissions
* roles
* resource types
* relationship types
* rule points
* workflow definitions
* event types
* event handlers
* jobs
* notification templates
* document classes
* migrations
* seed data
* health checks
* OpenAPI fragments

Provide Go interface examples, such as:

```go
type Module interface {
    Name() string
    Register(ctx ModuleContext) error
}

type ModuleContext interface {
    Router() Router
    DB() DB
    TxManager() TxManager
    Config() Config
    Logger() Logger
    Authz() AuthorizationService
    Audit() AuditLogger
    Outbox() Outbox
    Jobs() JobRegistry
    Workflows() WorkflowRegistry
    Rules() RuleRegistry
    Documents() DocumentService
    Notifications() NotificationService
}
```

Also provide:

* module lifecycle,
* dependency injection pattern,
* module dependency declaration,
* startup validation,
* seed loading strategy,
* migration discovery strategy,
* route registration strategy,
* shutdown strategy,
* health registration strategy.

---

# 18. Dependency Injection / IoC / Lifecycle Management

Design a practical IoC model suitable for Go.

Do not copy heavy Java/Spring-style dependency injection.

Prefer:

* explicit constructor injection,
* composition root,
* module registration,
* interface-based dependencies,
* manual DI or code-generated DI,
* optional Google Wire if useful,
* no hidden global service locator,
* no reflection-heavy runtime container unless strongly justified.

Cover:

* Application bootstrap.
* Composition root.
* Module lifecycle.
* Dependency registration.
* Dependency resolution.
* Startup validation.
* Health registration.
* Graceful shutdown.
* Background worker lifecycle.
* Database connection lifecycle.
* Transaction manager lifecycle.
* Event handler registration.
* Route registration.
* Seed registration.
* Migration registration.
* How to test modules with fake dependencies.
* How to avoid circular dependencies.
* How to avoid service locator anti-pattern.
* Startup order.
* Shutdown order.

Provide:

* Suggested `App` struct.
* Suggested `Kernel` struct.
* Suggested module registration flow.
* Suggested dependency graph rules.
* Example bootstrap pseudocode.

---

# 19. Common Error and Validation Framework

Design reusable error and validation framework.

Cover:

* Domain errors
* Validation errors
* Authorization errors
* Authentication errors
* Tenant isolation errors
* Not found errors
* Conflict errors
* Optimistic locking errors
* Idempotency errors
* Workflow state errors
* Rule validation errors
* External service errors
* Infrastructure errors
* Panic recovery

Provide:

* standard error codes,
* Go error wrapping pattern,
* HTTP mapping,
* logging strategy,
* safe user-facing messages,
* developer-facing internal messages,
* validation helper pattern,
* field error format,
* RFC 7807 / RFC 9457 style problem response,
* examples.

---

# 20. Security as a Framework Primitive

Security must be built into the framework core.

Cover:

* Secure defaults
* Authentication middleware
* OIDC/JWT support
* Session/token handling
* MFA readiness
* Tenant resolution middleware
* Authorization middleware
* Acting capacity enforcement
* Record-level access checks
* Policy evaluation
* Route permission metadata
* Rate limiting
* Input validation
* Request size limits
* Secure headers
* CORS strategy
* CSRF considerations where applicable
* SQL injection prevention
* RLS enforcement
* File upload validation
* Malware scanning hook for uploaded files
* Webhook signature verification
* Secrets management
* Encryption in transit
* Encryption at rest
* Sensitive data masking
* Audit logging
* Permission-denial audit
* Admin impersonation audit
* Break-glass access
* Token/session revocation hooks
* OWASP API Top 10 mitigations

The framework should make unsafe behavior difficult.

Examples:

* A handler should not be able to access tenant-scoped data without tenant context.
* A repository should not be able to run tenant-scoped queries without RLS context.
* A route should not be registered without permission metadata, unless explicitly marked public.
* A sensitive action should automatically create an audit trail.
* File upload should always pass through framework validation.
* Webhooks should always pass through signature verification and idempotency checks.

Provide framework-level interfaces and middleware for this.

---

# 21. Performance and Optimization Requirements

Design performance-conscious framework primitives.

Cover:

* Request lifecycle performance
* Minimal middleware overhead
* Efficient JSON encoding/decoding
* Connection pooling
* Query timeout enforcement
* Context cancellation
* Avoiding N+1 queries
* Pagination helpers
* Keyset pagination support
* Filtering/sorting allowlists
* Batch operations
* Efficient file upload/download using object storage
* Caching hooks
* In-memory cache where useful
* External cache adapter later
* Prepared queries / sqlc / pgx usage
* Avoiding reflection-heavy runtime patterns
* Memory allocation discipline
* pprof readiness
* Benchmarking strategy
* Performance budgets for common request paths
* Slow query detection
* Avoiding unnecessary goroutines
* Avoiding unnecessary abstractions on hot paths

Provide guidance on:

* What should be optimized early.
* What should not be prematurely optimized.
* Where abstraction overhead is acceptable.
* Where direct explicit code is better.
* Which framework components are latency-sensitive.
* Which framework components can run asynchronously.

The framework should help developers write fast code by default.

---

# 22. Practical Parallelization and Concurrency

Design concurrency support only where it is useful, not as unnecessary complexity.

Cover:

* Background job workers
* Outbox relay workers
* Notification fan-out
* Bulk import processing
* Bulk export/report generation
* File processing
* Webhook processing
* Workflow SLA/escalation jobs
* Parallel validation where safe
* Worker pools
* Bounded concurrency
* Backpressure
* Retry strategy
* Dead-letter queue
* Idempotency
* Context cancellation
* Graceful shutdown
* Avoiding unbounded goroutines
* Avoiding shared mutable state
* Avoiding race conditions
* Avoiding DB connection exhaustion

Provide framework primitives such as:

* `JobRunner`
* `WorkerPool`
* `OutboxRelay`
* `RetryPolicy`
* `BackoffPolicy`
* `DeadLetterStore`
* `AsyncTask`
* `BulkOperation`
* `ProgressTracker`

Explain where parallelization should not be used because it creates overhead or consistency risk.

---

# 23. Code Optimization and Maintainability

Design code conventions that prevent the framework from becoming messy.

Cover:

* Small packages
* Clear dependency direction
* Avoiding god services
* Avoiding generic interfaces everywhere
* Avoiding unnecessary abstractions
* Avoiding circular imports
* Avoiding hidden global state
* Avoiding excessive reflection
* Avoiding massive DTOs
* Avoiding direct SQL outside repositories
* Avoiding tenant bypass
* Avoiding audit bypass
* Avoiding duplicated validation
* Avoiding duplicated response/error handling
* Avoiding copied pagination logic
* Avoiding copied transaction logic
* Avoiding copied authz logic

Use Go-friendly patterns:

* Composition over inheritance
* Embedded structs for common fields
* Small interfaces at consumer side
* Generics only where they reduce real duplication
* Code generation for repetitive scaffolding
* Explicit domain code for complex workflows
* Interfaces for boundaries, not for every struct

Provide examples of:

* Good reusable helper.
* Bad over-abstracted helper.
* Good generic CRUD usage.
* Bad generic CRUD usage.
* Good module boundary.
* Bad module boundary.

---

# 24. Framework Hook and Interceptor System

Design a lightweight hook/interceptor system.

The framework should allow modules to attach behavior without modifying core code.

Support hooks such as:

* Before request
* After request
* Before command execution
* After command success
* After command failure
* Before workflow transition
* After workflow transition
* Before rule activation
* After rule activation
* Before document access
* After document access
* Before file upload
* After file upload
* Before audit write
* After event publish
* Before background job
* After background job

Explain:

* Which hooks are core.
* Which hooks are dangerous.
* How to keep hooks observable.
* How to prevent hooks from hiding business logic.
* How to test hooks.
* How to avoid ordering chaos.
* How to make hook failure behavior explicit.
* How hooks should be registered by modules.

---

# 25. Document, File, Comment, and Attachment Framework

Design this as a generic framework service.

Cover:

* Document metadata
* Document versions
* File storage abstraction
* Object storage adapter
* Presigned upload/download
* MIME/type validation
* Size limits
* Malware scanning hook
* Checksum/hash
* Retention policy
* Access grants
* Sensitive document handling
* Watermarking hook
* Download audit
* Comments
* Attachments
* Attachment linking to any resource
* Versioning
* Deletion/voiding strategy

Keep it domain-neutral.

Do not create society-specific document types in core.

---

# 26. Notification Framework

Design a reusable notification framework.

Cover:

* Notification templates
* Template variables
* Email
* SMS
* WhatsApp adapter
* Push adapter
* In-app notifications
* Notification preferences
* Tenant-specific templates
* Localization
* Delivery tracking
* Retries
* Dead-letter handling
* Audit for legal/important notifications
* Event-driven dispatch
* Synchronous vs asynchronous sending

Keep it generic.

---

# 27. Webhook and Integration Framework

Design reusable webhook and integration support.

Cover:

* Integration provider registry
* Provider credentials
* Webhook endpoint registration
* Signature verification
* Idempotency
* Replay protection
* Webhook event storage
* Retry strategy
* Dead-letter handling
* Outbound webhooks
* Circuit breaker
* Timeout
* Rate limit
* Provider adapter interface
* Audit logging
* Secrets handling

Use generic examples:

* payment provider callback
* external identity provider
* document verification provider
* messaging provider
* IoT/device provider

Do not bind it to any specific domain.

---

# 28. Event, Outbox, and Background Job Framework

Design reusable async infrastructure.

Cover:

* Event types
* Event schema/version
* Outbox table
* Atomic write with business transaction
* Outbox relay
* Event handlers
* Idempotent consumers
* Retry
* Dead-letter queue
* Scheduled jobs
* Job runs
* Job status
* Job locking
* Tenant-aware jobs
* Context cancellation
* Graceful shutdown
* Bulk operations
* Progress tracking
* Monitoring

Provide:

* Event naming convention.
* Event envelope format.
* Job interface.
* Job registry interface.
* Outbox interface.
* Retry policy interface.

---

# 29. REST API Framework

Design generic API conventions.

Cover:

* Versioning
* Tenant path/header strategy
* Resource naming
* Standard response format
* Standard error format
* Pagination
* Filtering
* Sorting
* Search
* Bulk operations
* Long-running operations
* File upload/download
* Webhooks
* OpenAPI generation
* Internal APIs
* External APIs
* Admin APIs
* Health APIs
* Idempotency keys
* Optimistic concurrency
* ETags or version headers
* Request/response DTO conventions
* DTO/domain mapping conventions

Provide generic endpoint groups:

* tenants
* organizations
* users
* parties/persons
* resources
* relationships
* roles/permissions
* assignments
* rules
* workflows
* documents
* comments
* attachments
* notifications
* audit logs
* integrations
* webhooks
* jobs

Do not include society-specific endpoints.

---

# 30. Observability and Operations Framework

Design reusable operational foundation.

Cover:

* Structured logs
* Request IDs
* Tenant IDs
* Actor IDs
* Acting capacity IDs
* Trace IDs
* Metrics
* Tracing
* Health checks
* Readiness checks
* Job monitoring
* Outbox monitoring
* Workflow monitoring
* Notification delivery monitoring
* Webhook monitoring
* Error monitoring
* Audit export
* Dashboards
* Alerting
* CI/CD
* Docker
* Deployment
* Database migrations
* Backup and restore
* Disaster recovery
* Graceful shutdown
* Config management
* Secret loading
* Environment management

Keep deployment recommendation practical for a small team.

---

# 31. Testing Framework and Testkit

Design reusable testing utilities and strategy.

Cover:

* Unit testing
* Integration testing
* Testcontainers
* RLS tenant isolation tests
* Authorization tests
* Workflow tests
* Rule resolution tests
* Audit immutability tests
* Outbox tests
* Idempotency tests
* Webhook replay tests
* Module contract tests
* Test fixtures
* Seed test data
* Fake notification provider
* Fake object storage
* Fake integration provider
* Fake webhook verifier
* Fake clock
* Fake UUID generator
* Performance tests
* Race tests
* Security tests

Provide suggested `/internal/testkit` structure.

Include helpers such as:

* create test tenant
* create test organization
* create test user
* create test party/person
* create actor assignment
* issue test auth token
* assert RLS isolation
* assert authorization allowed/denied
* assert audit log created
* assert outbox event created
* assert workflow transition
* assert rule resolution
* assert idempotency behavior

---

# 32. Code Generation / Templates / CLI

Recommend where code generation or templates should be used.

Consider generating:

* module boilerplate
* CRUD handlers
* DTO mappers
* SQLC query wrappers
* OpenAPI stubs
* permission seed files
* workflow seed files
* rule seed files
* mock interfaces
* test fixtures
* migration files
* API route skeleton

Explain:

* what should be generated,
* what should remain handwritten,
* how to avoid generator lock-in,
* how to keep generated code reviewable,
* how to keep generated code optional.

Recommend developer commands such as:

* `make new-module name=requests`
* `make migrate-create name=create_requests`
* `make seed-validate`
* `make openapi-generate`
* `make test-integration`
* `make lint-boundaries`
* `make gen`
* `make test-race`
* `make bench`

---

# 33. Go Project Structure

Provide a reusable Go project structure.

It should not be named around society management.

Use neutral names.

Include:

```text
/cmd/api
/cmd/worker
/cmd/migrate
/internal/kernel
/internal/platform
/internal/modules
/internal/adapters
/internal/shared
/internal/testkit
/pkg
/migrations
/api/openapi
/configs
/deployments
/scripts
/docs
/tools
```

For each folder, explain:

* Purpose
* What belongs there
* What must not belong there
* Which packages can import it
* Which packages it must not import

Also provide a detailed package map, such as:

```text
/internal/kernel/model
/internal/kernel/tenant
/internal/kernel/auth
/internal/kernel/authz
/internal/kernel/policy
/internal/kernel/resource
/internal/kernel/relationship
/internal/kernel/workflow
/internal/kernel/rules
/internal/kernel/audit
/internal/kernel/outbox
/internal/kernel/jobs
/internal/kernel/document
/internal/kernel/notify
/internal/kernel/webhook
/internal/kernel/integration
/internal/kernel/httpx
/internal/kernel/errors
/internal/kernel/validation
/internal/kernel/pagination
/internal/kernel/filtering
/internal/kernel/database
/internal/kernel/config
/internal/kernel/logging
/internal/kernel/observability
/internal/kernel/testkit
```

For each package, specify:

* Responsibility
* Exported interfaces
* Helper structs
* What it must not import
* Whether domain modules can import it

---

# 34. Domain Module SDK / Extension Contract

Design how a new domain module plugs into the framework.

A module should be able to declare:

* Module name
* Owned tables
* Routes
* Permissions
* Roles
* Resource types
* Relationship types
* Rule points
* Workflow definitions
* Event types
* Event handlers
* Notification templates
* Document classes
* Background jobs
* API handlers
* Service interfaces
* Repositories
* Migrations
* Seeds
* Tests
* OpenAPI fragments

Provide a sample generic module registration interface in Go.

Do not use society as the only example. Use a neutral example like `requests`, `assets`, or `approvals`.

Then add a short note showing that a future `society` module can register building, unit, owner, tenant, notice, complaint, bill, committee, etc. using the same extension contract.

---

# 35. Core vs Domain Module Boundary

Create a strict boundary table.

Core framework should include:

* tenant management
* organization model
* identity
* authorization
* policy evaluation
* relationship framework
* resource registry
* workflow runtime
* rule runtime
* document framework
* comments/attachments framework
* audit
* outbox
* jobs
* notifications
* webhooks
* integrations
* API helpers
* database helpers
* validation helpers
* testkit
* module SDK

Domain modules should include:

* housing-society-specific building/unit/member logic
* school-specific student/class logic
* facility-specific asset/work-order logic
* club-specific membership package logic
* domain billing formulas
* domain workflows
* domain reports
* domain validations
* domain-specific resource types
* domain-specific relationship types
* domain-specific roles
* domain-specific rules

Explain how to prevent accidental leakage.

---

# 36. Non-Functional Requirements Matrix

Create an NFR matrix for the framework.

Include:

* Security
* Performance
* Scalability
* Reliability
* Maintainability
* Extensibility
* Observability
* Testability
* Developer experience
* Portability
* Compliance readiness
* Operational simplicity

For each, provide:

* Requirement
* Design decision
* Framework component responsible
* Acceptance test
* Risk if ignored

---

# 37. Framework Acceptance Criteria

Define how we know the framework is successful.

Include measurable acceptance criteria:

* A new module can register routes, permissions, rules, workflows, events, jobs, and seeds without modifying framework core.
* Tenant-scoped repositories cannot run without tenant context.
* Routes cannot be exposed without permission metadata unless marked public.
* Audit logging works for sensitive actions.
* Permission denials are audited.
* RLS tenant isolation tests pass.
* A simple CRUD module can be generated and tested quickly.
* Workflow definitions can be added from seed/config.
* Rule versions can be approved and resolved historically.
* Outbox events are written atomically with business changes.
* Background jobs are idempotent and tenant-aware.
* Standard API errors are consistent.
* Standard pagination works across modules.
* OpenAPI docs can be generated.
* Testkit can create tenants, users, actors, permissions, and resources.
* Core framework has no society-specific concepts.
* Domain modules cannot import each other directly except through declared ports.
* Framework can be reused for another product without removing society-specific code.
* Performance budget for basic authenticated request path is defined.
* Security guardrails are enforced by middleware, route metadata, and tests.
* The framework avoids unnecessary abstraction overhead on hot paths.

---

# 38. Phase 0 Framework-Only Backlog

Create a Phase 0 backlog only for the reusable framework.

Do not create a society product backlog.

Include epics such as:

* Project skeleton
* Config and logging
* Database setup
* Migration runner
* Tenant context and RLS
* Identity/auth foundation
* Party/person foundation
* Resource registry
* Relationship framework
* Authorization framework
* Rule framework
* Workflow framework
* Audit framework
* Outbox/event framework
* Background jobs
* Document/file framework
* Comment/attachment framework
* Notification framework
* Webhook framework
* Integration framework
* API helpers
* Error/validation helpers
* Pagination/filtering/sorting helpers
* Base model primitives
* Base DTO primitives
* Repository helpers
* Transaction/RLS helpers
* Module SDK
* IoC/bootstrap
* Testkit
* CLI/codegen/templates
* CI/CD basics
* Performance benchmarks
* Security tests

For each story, provide:

* Description
* Acceptance criteria
* Dependencies
* Test coverage
* Risk

---

# 39. Reference Domain Boundary Check

At the end, perform a boundary check.

List anything that would be housing-society-specific and confirm it is **not** part of the core framework.

Examples that must remain outside core:

* building
* wing
* flat
* owner as housing flat owner
* member under society bye-laws
* associate member
* nominal member
* committee
* chairman
* secretary
* treasurer
* maintenance bill
* defaulter
* AGM
* society notice
* parking allocation
* water/STP/WTP
* visitor gate entry
* conveyance
* redevelopment
* election

Then explain how the future society module can register these using the generic framework.

---

# 40. Coding-Ready Deliverables

Produce coding-ready outputs.

The final answer must include:

1. Executive recommendation.
2. Core framework principles.
3. Architecture diagrams.
4. Framework glossary.
5. Platform core ERD.
6. PostgreSQL DDL skeleton for generic framework tables only.
7. Base model primitives.
8. Base API/DTO primitives.
9. Base handler/helper pattern.
10. Repository and transaction helpers.
11. Service/use-case conventions.
12. Module starter template.
13. Module registration interface.
14. IoC/bootstrap structure.
15. Common security middleware.
16. Common performance guardrails.
17. Concurrency/job primitives.
18. Hook/interceptor system.
19. Authorization evaluator interface.
20. Rule resolver interface.
21. Workflow runtime interface.
22. Audit logger interface.
23. Outbox interface.
24. Document service interface.
25. Notification service interface.
26. Webhook service interface.
27. Standard API response/error structs.
28. Seed data model.
29. Migration strategy.
30. Testkit structure.
31. Code generation/template recommendation.
32. Package map.
33. NFR matrix.
34. Framework acceptance criteria.
35. Phase 0 framework-only backlog.
36. Boundary check confirming no society-specific logic is in the core.
37. Final recommendation with the first 10 files/folders engineers should create.

The result should be practical, opinionated, and coding-ready for building the reusable framework core.

Do not produce a society management product design.

Do not produce society-specific database tables.

Do not produce society-specific workflows.

Do not produce society-specific backlog.

Only show society management as a brief example of how a future domain module can use the framework extension points.


# Additional Requirement: Practical Architecture and Design Patterns

The framework design must intentionally use applicable software architecture, enterprise integration, security, persistence, concurrency, and Go design patterns wherever they improve practicality, maintainability, safety, performance, and developer experience.

Do not list patterns academically.

For every pattern used, explain:

* Why it is useful in this framework.
* Where it should be used.
* Where it should not be used.
* What problem it solves.
* What Go implementation approach is recommended.
* What anti-pattern to avoid.

Cover applicable patterns from these categories.

---

## 1. Architectural Patterns

Evaluate and use where appropriate:

* Modular monolith
* Clean architecture
* Hexagonal architecture / ports and adapters
* Onion architecture
* Layered architecture
* Vertical slice architecture
* Plugin/module architecture
* Event-driven architecture
* CQRS where useful, but not everywhere
* Mediator pattern, if useful
* Shared kernel
* Anti-corruption layer
* Backend-for-frontend only if justified
* Strangler pattern for future extraction
* Microservices only as a future extraction option

Explain which patterns form the core framework architecture and which should be delayed.

---

## 2. Domain and Modeling Patterns

Use practical domain modeling patterns such as:

* Entity
* Value Object
* Aggregate
* Domain Service
* Application Service
* Repository
* Domain Event
* Specification pattern
* Policy pattern
* State machine pattern
* Temporal modeling
* Resource registry pattern
* Relationship graph pattern
* Actor-capacity pattern
* Metadata/extension field pattern using JSONB only where appropriate

Explain how these patterns apply to generic framework concepts like tenant, organization, party, actor, relationship, resource, workflow, rule, document, and audit.

Avoid forcing DDD terminology where it does not help.

---

## 3. Persistence and Transaction Patterns

Use applicable persistence patterns:

* Repository pattern
* Unit of Work
* Transaction script where simpler
* Transaction manager
* Optimistic locking
* Idempotency key pattern
* Outbox pattern
* Inbox pattern for idempotent consumers
* Append-only log
* Event log
* Audit log
* Soft delete / status lifecycle
* Temporal validity using `valid_from` / `valid_to`
* Row-Level Security pattern
* Read model / materialized view where useful
* Pagination pattern, preferably keyset pagination
* Migration expand-contract pattern

Explain where each pattern belongs and how it should be implemented in Go/PostgreSQL.

---

## 4. API and Integration Patterns

Use applicable API and integration patterns:

* REST resource pattern
* DTO pattern
* API response envelope
* Problem details error response
* Command/query DTO separation
* Webhook receiver pattern
* Webhook signature verification
* Retry with backoff
* Circuit breaker
* Bulk operation as async job
* Long-running operation pattern
* Presigned upload/download pattern
* Provider adapter pattern
* Anti-corruption layer for external integrations
* API versioning
* Backward compatibility pattern

Explain how the framework should make these patterns reusable.

---

## 5. Security Patterns

Use applicable security patterns:

* Secure by default
* Deny by default
* Defense in depth
* Least privilege
* Tenant isolation
* Record-level authorization
* Policy-based authorization
* Acting capacity selection
* Break-glass access
* Admin impersonation with full audit
* Token revocation hook
* Permission metadata on routes
* Sensitive action audit
* Secure file upload pipeline
* Secret provider abstraction
* Rate limiting
* Replay protection for webhooks
* Input validation at boundary
* Output encoding / safe error messages

Explain how these patterns are enforced by framework middleware, helpers, tests, and conventions.

---

## 6. Concurrency and Async Patterns

Use applicable concurrency and background processing patterns:

* Worker pool
* Bounded concurrency
* Backpressure
* Retry policy
* Dead-letter queue
* Scheduled job
* Outbox relay
* Idempotent consumer
* Bulk operation progress tracking
* Context cancellation
* Graceful shutdown
* Fan-out/fan-in where useful
* Single-flight where useful
* Locking / advisory locks where useful
* Avoiding unbounded goroutines

Explain where parallelization is practical and where it creates unnecessary overhead or consistency risk.

---

## 7. Go-Specific Implementation Patterns

Use Go-friendly patterns:

* Composition over inheritance
* Embedded structs for common fields
* Small interfaces at consumer boundaries
* Constructor injection
* Explicit dependencies
* Composition root
* Functional options only where helpful
* Generics only for real reusable primitives
* Code generation for repetitive scaffolding
* Interface segregation
* Context propagation
* Error wrapping
* Table-driven tests
* Test fixtures
* Fake clock and fake ID generator
* No global mutable state
* No reflection-heavy dependency container unless strongly justified

Explain how to implement these patterns idiomatically in Go.

---

## 8. Framework Developer Experience Patterns

Use patterns that improve ease of development:

* Module starter template
* Module registration contract
* Seed registration
* Migration discovery
* Permission catalog registration
* Workflow definition registration
* Rule point registration
* Event handler registration
* Job registration
* Testkit fixtures
* Common handler helpers
* Common repository helpers
* Common response/error helpers
* Code generation templates
* Makefile/CLI commands
* Boundary linting
* OpenAPI generation

Explain how a developer can create a new module with minimal boilerplate while still keeping domain logic explicit.

---

## 9. Anti-Patterns to Avoid

Explicitly identify anti-patterns that must be avoided:

* God service
* Fat controller
* Anemic domain model where harmful
* Over-abstracted generic repository
* One universal BaseModel with everything
* Service locator
* Hidden global dependencies
* Circular module imports
* Cross-module SQL joins
* Bypassing tenant context
* Bypassing audit logging
* Route without permission metadata
* Business logic in middleware
* Workflow logic hard-coded in handlers
* Rule constants hard-coded in business code
* Unbounded goroutines
* Reflection-heavy runtime magic
* Premature microservices
* Premature event sourcing
* Premature CQRS everywhere
* Generic low-code framework before real modules exist

For each major anti-pattern, explain the safer framework alternative.

---

## 10. Pattern Decision Matrix

Provide a practical decision matrix.

For each important pattern, include:

* Use in core framework?
* Use in domain modules?
* Use later only?
* Avoid?
* Reason
* Go implementation recommendation

The matrix should help engineers decide when to use a pattern and when not to.

---

## 11. Final Pattern Recommendation

End this section with a clear recommended pattern stack for the framework.

Example format:

* Core architecture:
* Module boundary:
* Data access:
* Transaction:
* Authorization:
* Workflow:
* Rules:
* Events:
* Audit:
* Background jobs:
* Integrations:
* API:
* Testing:
* Developer scaffolding:

Keep recommendations practical and implementation-oriented.
