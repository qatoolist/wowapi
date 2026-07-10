# Package Dependencies

> The tool has detected code relationships. It has not assumed whether the project follows layered architecture, clean architecture, hexagonal architecture, MVC, CQRS, or any other pattern unless explicitly configured.

## github.com/qatoolist/wowapi/adapters/auth/pgprincipal

- **Package name:** pgprincipal

**Imports:**

- `context`
- `errors`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`

**Imported by:**

_Not imported by any recorded package._

## github.com/qatoolist/wowapi/adapters/metrics/prometheus

- **Package name:** prometheus

**Imports:**

- `github.com/prometheus/client_golang/prometheus`
- `github.com/prometheus/client_golang/prometheus/promhttp`
- `github.com/qatoolist/wowapi/kernel/observability`
- `net/http`
- `sort`
- `strconv`
- `sync`
- `time`

**Imported by:**

_Not imported by any recorded package._

## github.com/qatoolist/wowapi/adapters/secrets/envprovider

- **Package name:** envprovider

**Imports:**

- `context`
- `fmt`
- `github.com/qatoolist/wowapi/kernel/secrets`
- `os`

**Imported by:**

- `github.com/qatoolist/wowapi/internal/cli`

## github.com/qatoolist/wowapi/adapters/tracing/otel

- **Package name:** otel

**Imports:**

- `context`
- `github.com/qatoolist/wowapi/kernel/observability`
- `go.opentelemetry.io/otel/attribute`
- `go.opentelemetry.io/otel/codes`
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp`
- `go.opentelemetry.io/otel/propagation`
- `go.opentelemetry.io/otel/sdk/trace`
- `go.opentelemetry.io/otel/trace`

**Imported by:**

_Not imported by any recorded package._

## github.com/qatoolist/wowapi/app

- **Package name:** app

**Imports:**

- `context`
- `errors`
- `fmt`
- `github.com/google/uuid`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/kernel/artifact`
- `github.com/qatoolist/wowapi/kernel/attachment`
- `github.com/qatoolist/wowapi/kernel/audit`
- `github.com/qatoolist/wowapi/kernel/authz`
- `github.com/qatoolist/wowapi/kernel/bulk`
- `github.com/qatoolist/wowapi/kernel/comment`
- `github.com/qatoolist/wowapi/kernel/config`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/document`
- `github.com/qatoolist/wowapi/kernel/httpx`
- `github.com/qatoolist/wowapi/kernel/integration`
- `github.com/qatoolist/wowapi/kernel/jobs`
- `github.com/qatoolist/wowapi/kernel/model`
- `github.com/qatoolist/wowapi/kernel/notify`
- `github.com/qatoolist/wowapi/kernel/outbox`
- `github.com/qatoolist/wowapi/kernel/resource`
- `github.com/qatoolist/wowapi/kernel/retention`
- `github.com/qatoolist/wowapi/kernel/rules`
- `github.com/qatoolist/wowapi/kernel/seeds`
- `github.com/qatoolist/wowapi/kernel/sequence`
- `github.com/qatoolist/wowapi/kernel/validation`
- `github.com/qatoolist/wowapi/kernel/webhook`
- `github.com/qatoolist/wowapi/kernel/workflow`
- `github.com/qatoolist/wowapi/module`
- `io/fs`
- `log/slog`
- `slices`
- `sort`
- `sync`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/testkit`

## github.com/qatoolist/wowapi/cmd/wowapi

- **Package name:** main

**Imports:**

- `github.com/qatoolist/wowapi/internal/cli`
- `os`

**Imported by:**

_Not imported by any recorded package._

## github.com/qatoolist/wowapi/internal/buildinfo

- **Package name:** buildinfo

**Imports:**

- `bufio`
- `os`
- `path/filepath`
- `runtime/debug`
- `strings`

**Imported by:**

- `github.com/qatoolist/wowapi/internal/cli`

## github.com/qatoolist/wowapi/internal/cli

- **Package name:** cli

**Imports:**

- `bytes`
- `context`
- `embed`
- `encoding/json`
- `errors`
- `flag`
- `fmt`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5/pgxpool`
- `github.com/qatoolist/wowapi/adapters/secrets/envprovider`
- `github.com/qatoolist/wowapi/internal/buildinfo`
- `github.com/qatoolist/wowapi/kernel/apikey`
- `github.com/qatoolist/wowapi/kernel/audit`
- `github.com/qatoolist/wowapi/kernel/config`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/jobs`
- `github.com/qatoolist/wowapi/kernel/model`
- `github.com/qatoolist/wowapi/kernel/outbox`
- `github.com/qatoolist/wowapi/kernel/seeds`
- `go/format`
- `io`
- `io/fs`
- `os`
- `os/exec`
- `path/filepath`
- `regexp`
- `sort`
- `strconv`
- `strings`
- `text/tabwriter`
- `text/template`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/cmd/wowapi`

## github.com/qatoolist/wowapi/internal/e2e

- **Package name:** e2e

**Imports:**

_No imports recorded._

**Imported by:**

_Not imported by any recorded package._

## github.com/qatoolist/wowapi/internal/testmodules/requests

- **Package name:** requests

**Imports:**

- `context`
- `embed`
- `errors`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/qatoolist/wowapi/kernel/authz`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/httpx`
- `github.com/qatoolist/wowapi/kernel/model`
- `github.com/qatoolist/wowapi/kernel/resource`
- `github.com/qatoolist/wowapi/kernel/validation`
- `github.com/qatoolist/wowapi/module`
- `io/fs`
- `net/http`

**Imported by:**

_Not imported by any recorded package._

## github.com/qatoolist/wowapi/internal/tools/benchbudget

- **Package name:** main

**Imports:**

- `bufio`
- `fmt`
- `os`
- `strconv`
- `strings`

**Imported by:**

_Not imported by any recorded package._

## github.com/qatoolist/wowapi/internal/tools/migrate

- **Package name:** main

**Imports:**

- `context`
- `fmt`
- `github.com/qatoolist/wowapi/kernel/config`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/migrations`
- `os`
- `time`

**Imported by:**

_Not imported by any recorded package._

## github.com/qatoolist/wowapi/kernel

- **Package name:** kernel

**Imports:**

- `context`
- `fmt`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5/pgxpool`
- `github.com/qatoolist/wowapi/kernel/artifact`
- `github.com/qatoolist/wowapi/kernel/attachment`
- `github.com/qatoolist/wowapi/kernel/audit`
- `github.com/qatoolist/wowapi/kernel/authz`
- `github.com/qatoolist/wowapi/kernel/bulk`
- `github.com/qatoolist/wowapi/kernel/comment`
- `github.com/qatoolist/wowapi/kernel/config`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/document`
- `github.com/qatoolist/wowapi/kernel/integration`
- `github.com/qatoolist/wowapi/kernel/model`
- `github.com/qatoolist/wowapi/kernel/notify`
- `github.com/qatoolist/wowapi/kernel/observability`
- `github.com/qatoolist/wowapi/kernel/outbox`
- `github.com/qatoolist/wowapi/kernel/policy`
- `github.com/qatoolist/wowapi/kernel/relationship`
- `github.com/qatoolist/wowapi/kernel/resource`
- `github.com/qatoolist/wowapi/kernel/retention`
- `github.com/qatoolist/wowapi/kernel/rules`
- `github.com/qatoolist/wowapi/kernel/secrets`
- `github.com/qatoolist/wowapi/kernel/sequence`
- `github.com/qatoolist/wowapi/kernel/storage`
- `github.com/qatoolist/wowapi/kernel/webhook`
- `github.com/qatoolist/wowapi/kernel/workflow`
- `log/slog`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/testkit`

## github.com/qatoolist/wowapi/kernel/apikey

- **Package name:** apikey

**Imports:**

- `context`
- `crypto/rand`
- `crypto/sha256`
- `crypto/subtle`
- `encoding/base64`
- `encoding/hex`
- `errors`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/qatoolist/wowapi/kernel/audit`
- `github.com/qatoolist/wowapi/kernel/authz`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/model`
- `net/http`
- `strings`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/internal/cli`

## github.com/qatoolist/wowapi/kernel/artifact

- **Package name:** artifact

**Imports:**

- `context`
- `crypto/sha256`
- `encoding/hex`
- `encoding/json`
- `errors`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/jackc/pgx/v5/pgconn`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/model`
- `sort`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/module`

## github.com/qatoolist/wowapi/kernel/attachment

- **Package name:** attachment

**Imports:**

- `context`
- `errors`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/model`
- `github.com/qatoolist/wowapi/kernel/outbox`
- `github.com/qatoolist/wowapi/kernel/resource`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/module`

## github.com/qatoolist/wowapi/kernel/audit

- **Package name:** audit

**Imports:**

- `context`
- `crypto/sha256`
- `encoding/binary`
- `encoding/hex`
- `encoding/json`
- `errors`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/jackc/pgx/v5/pgxpool`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/httpx`
- `github.com/qatoolist/wowapi/kernel/model`
- `hash`
- `strconv`
- `strings`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/internal/cli`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/kernel/apikey`
- `github.com/qatoolist/wowapi/module`

## github.com/qatoolist/wowapi/kernel/auth

- **Package name:** auth

**Imports:**

- `context`
- `crypto/ecdsa`
- `crypto/elliptic`
- `crypto/rsa`
- `encoding/base64`
- `encoding/json`
- `github.com/golang-jwt/jwt/v5`
- `github.com/google/uuid`
- `github.com/qatoolist/wowapi/kernel/authz`
- `github.com/qatoolist/wowapi/kernel/errors`
- `io`
- `math/big`
- `net`
- `net/http`
- `net/url`
- `strings`
- `sync`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/testkit`

## github.com/qatoolist/wowapi/kernel/authz

- **Package name:** authz

**Imports:**

- `context`
- `encoding/json`
- `errors`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/resource`
- `regexp`
- `slices`
- `sort`
- `sync`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/internal/testmodules/requests`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/kernel/apikey`
- `github.com/qatoolist/wowapi/kernel/auth`
- `github.com/qatoolist/wowapi/kernel/document`
- `github.com/qatoolist/wowapi/kernel/httpx`
- `github.com/qatoolist/wowapi/kernel/policy`
- `github.com/qatoolist/wowapi/kernel/relationship`
- `github.com/qatoolist/wowapi/kernel/workflow`
- `github.com/qatoolist/wowapi/module`
- `github.com/qatoolist/wowapi/testkit`

## github.com/qatoolist/wowapi/kernel/bulk

- **Package name:** bulk

**Imports:**

- `context`
- `encoding/json`
- `errors`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/model`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/module`

## github.com/qatoolist/wowapi/kernel/comment

- **Package name:** comment

**Imports:**

- `context`
- `errors`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/model`
- `github.com/qatoolist/wowapi/kernel/outbox`
- `github.com/qatoolist/wowapi/kernel/resource`
- `strings`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/module`

## github.com/qatoolist/wowapi/kernel/config

- **Package name:** config

**Imports:**

- `bytes`
- `context`
- `crypto/sha256`
- `encoding`
- `encoding/hex`
- `encoding/json`
- `errors`
- `fmt`
- `github.com/qatoolist/wowapi/kernel/secrets`
- `gopkg.in/yaml.v3`
- `log/slog`
- `os`
- `reflect`
- `regexp`
- `slices`
- `strconv`
- `strings`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/internal/cli`
- `github.com/qatoolist/wowapi/internal/tools/migrate`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/integration`
- `github.com/qatoolist/wowapi/kernel/logging`
- `github.com/qatoolist/wowapi/module`
- `github.com/qatoolist/wowapi/testkit`

## github.com/qatoolist/wowapi/kernel/database

- **Package name:** database

**Imports:**

- `context`
- `errors`
- `fmt`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/jackc/pgx/v5/pgconn`
- `github.com/jackc/pgx/v5/pgxpool`
- `github.com/jackc/pgx/v5/stdlib`
- `github.com/pressly/goose/v3`
- `github.com/qatoolist/wowapi/kernel/config`
- `github.com/qatoolist/wowapi/kernel/errors`
- `io/fs`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/adapters/auth/pgprincipal`
- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/internal/cli`
- `github.com/qatoolist/wowapi/internal/testmodules/requests`
- `github.com/qatoolist/wowapi/internal/tools/migrate`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/kernel/apikey`
- `github.com/qatoolist/wowapi/kernel/artifact`
- `github.com/qatoolist/wowapi/kernel/attachment`
- `github.com/qatoolist/wowapi/kernel/audit`
- `github.com/qatoolist/wowapi/kernel/authz`
- `github.com/qatoolist/wowapi/kernel/bulk`
- `github.com/qatoolist/wowapi/kernel/comment`
- `github.com/qatoolist/wowapi/kernel/document`
- `github.com/qatoolist/wowapi/kernel/httpx`
- `github.com/qatoolist/wowapi/kernel/integration`
- `github.com/qatoolist/wowapi/kernel/jobs`
- `github.com/qatoolist/wowapi/kernel/notify`
- `github.com/qatoolist/wowapi/kernel/outbox`
- `github.com/qatoolist/wowapi/kernel/relationship`
- `github.com/qatoolist/wowapi/kernel/resource`
- `github.com/qatoolist/wowapi/kernel/retention`
- `github.com/qatoolist/wowapi/kernel/rules`
- `github.com/qatoolist/wowapi/kernel/seeds`
- `github.com/qatoolist/wowapi/kernel/sequence`
- `github.com/qatoolist/wowapi/kernel/webhook`
- `github.com/qatoolist/wowapi/kernel/workflow`
- `github.com/qatoolist/wowapi/module`
- `github.com/qatoolist/wowapi/testkit`

## github.com/qatoolist/wowapi/kernel/document

- **Package name:** document

**Imports:**

- `context`
- `errors`
- `fmt`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/qatoolist/wowapi/kernel/authz`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/model`
- `github.com/qatoolist/wowapi/kernel/outbox`
- `github.com/qatoolist/wowapi/kernel/resource`
- `github.com/qatoolist/wowapi/kernel/storage`
- `net/http`
- `regexp`
- `slices`
- `sort`
- `strings`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/module`

## github.com/qatoolist/wowapi/kernel/errors

- **Package name:** errors

**Imports:**

- `errors`
- `fmt`
- `net/http`

**Imported by:**

- `github.com/qatoolist/wowapi/adapters/auth/pgprincipal`
- `github.com/qatoolist/wowapi/internal/testmodules/requests`
- `github.com/qatoolist/wowapi/kernel/apikey`
- `github.com/qatoolist/wowapi/kernel/artifact`
- `github.com/qatoolist/wowapi/kernel/attachment`
- `github.com/qatoolist/wowapi/kernel/audit`
- `github.com/qatoolist/wowapi/kernel/auth`
- `github.com/qatoolist/wowapi/kernel/authz`
- `github.com/qatoolist/wowapi/kernel/bulk`
- `github.com/qatoolist/wowapi/kernel/comment`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/document`
- `github.com/qatoolist/wowapi/kernel/filtering`
- `github.com/qatoolist/wowapi/kernel/httpx`
- `github.com/qatoolist/wowapi/kernel/integration`
- `github.com/qatoolist/wowapi/kernel/jobs`
- `github.com/qatoolist/wowapi/kernel/notify`
- `github.com/qatoolist/wowapi/kernel/outbox`
- `github.com/qatoolist/wowapi/kernel/pagination`
- `github.com/qatoolist/wowapi/kernel/policy`
- `github.com/qatoolist/wowapi/kernel/relationship`
- `github.com/qatoolist/wowapi/kernel/resource`
- `github.com/qatoolist/wowapi/kernel/retention`
- `github.com/qatoolist/wowapi/kernel/rules`
- `github.com/qatoolist/wowapi/kernel/seeds`
- `github.com/qatoolist/wowapi/kernel/sequence`
- `github.com/qatoolist/wowapi/kernel/storage`
- `github.com/qatoolist/wowapi/kernel/validation`
- `github.com/qatoolist/wowapi/kernel/webhook`
- `github.com/qatoolist/wowapi/kernel/workflow`

## github.com/qatoolist/wowapi/kernel/filtering

- **Package name:** filtering

**Imports:**

- `fmt`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/pagination`
- `sort`
- `strings`

**Imported by:**

- `github.com/qatoolist/wowapi/kernel/httpx`
- `github.com/qatoolist/wowapi/kernel/workflow`

## github.com/qatoolist/wowapi/kernel/httpx

- **Package name:** httpx

**Imports:**

- `context`
- `crypto/sha256`
- `encoding/hex`
- `encoding/json`
- `errors`
- `fmt`
- `github.com/google/uuid`
- `github.com/qatoolist/wowapi/kernel/authz`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/filtering`
- `github.com/qatoolist/wowapi/kernel/model`
- `github.com/qatoolist/wowapi/kernel/pagination`
- `github.com/qatoolist/wowapi/kernel/validation`
- `io`
- `log/slog`
- `math`
- `net`
- `net/http`
- `sort`
- `strconv`
- `strings`
- `sync`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/internal/testmodules/requests`
- `github.com/qatoolist/wowapi/kernel/audit`
- `github.com/qatoolist/wowapi/kernel/observability`
- `github.com/qatoolist/wowapi/module`

## github.com/qatoolist/wowapi/kernel/integration

- **Package name:** integration

**Imports:**

- `context`
- `encoding/json`
- `errors`
- `fmt`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/qatoolist/wowapi/kernel/config`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/model`
- `github.com/qatoolist/wowapi/kernel/secrets`
- `regexp`
- `sort`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/module`

## github.com/qatoolist/wowapi/kernel/jobs

- **Package name:** jobs

**Imports:**

- `context`
- `encoding/json`
- `errors`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/jackc/pgx/v5/pgxpool`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/model`
- `github.com/qatoolist/wowapi/kernel/observability`
- `log/slog`
- `strconv`
- `sync`
- `time`
- `unicode/utf8`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/internal/cli`
- `github.com/qatoolist/wowapi/module`

## github.com/qatoolist/wowapi/kernel/logging

- **Package name:** logging

**Imports:**

- `fmt`
- `github.com/qatoolist/wowapi/kernel/config`
- `io`
- `log/slog`
- `strings`

**Imported by:**

_Not imported by any recorded package._

## github.com/qatoolist/wowapi/kernel/model

- **Package name:** model

**Imports:**

- `github.com/google/uuid`
- `github.com/shopspring/decimal`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/internal/cli`
- `github.com/qatoolist/wowapi/internal/testmodules/requests`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/kernel/apikey`
- `github.com/qatoolist/wowapi/kernel/artifact`
- `github.com/qatoolist/wowapi/kernel/attachment`
- `github.com/qatoolist/wowapi/kernel/audit`
- `github.com/qatoolist/wowapi/kernel/bulk`
- `github.com/qatoolist/wowapi/kernel/comment`
- `github.com/qatoolist/wowapi/kernel/document`
- `github.com/qatoolist/wowapi/kernel/httpx`
- `github.com/qatoolist/wowapi/kernel/integration`
- `github.com/qatoolist/wowapi/kernel/jobs`
- `github.com/qatoolist/wowapi/kernel/notify`
- `github.com/qatoolist/wowapi/kernel/outbox`
- `github.com/qatoolist/wowapi/kernel/relationship`
- `github.com/qatoolist/wowapi/kernel/retention`
- `github.com/qatoolist/wowapi/kernel/rules`
- `github.com/qatoolist/wowapi/kernel/sequence`
- `github.com/qatoolist/wowapi/kernel/webhook`
- `github.com/qatoolist/wowapi/kernel/workflow`
- `github.com/qatoolist/wowapi/module`
- `github.com/qatoolist/wowapi/testkit/fakes`

## github.com/qatoolist/wowapi/kernel/notify

- **Package name:** notify

**Imports:**

- `context`
- `encoding/json`
- `errors`
- `fmt`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/model`
- `github.com/qatoolist/wowapi/kernel/observability`
- `github.com/qatoolist/wowapi/kernel/resource`
- `html/template`
- `regexp`
- `sort`
- `strings`
- `sync`
- `text/template`
- `text/template/parse`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/module`

## github.com/qatoolist/wowapi/kernel/observability

- **Package name:** observability

**Imports:**

- `context`
- `github.com/qatoolist/wowapi/kernel/httpx`
- `log/slog`
- `net/http`
- `strconv`
- `strings`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/adapters/metrics/prometheus`
- `github.com/qatoolist/wowapi/adapters/tracing/otel`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/kernel/jobs`
- `github.com/qatoolist/wowapi/kernel/notify`
- `github.com/qatoolist/wowapi/kernel/outbox`
- `github.com/qatoolist/wowapi/kernel/webhook`

## github.com/qatoolist/wowapi/kernel/outbox

- **Package name:** outbox

**Imports:**

- `context`
- `encoding/json`
- `errors`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/jackc/pgx/v5/pgxpool`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/model`
- `github.com/qatoolist/wowapi/kernel/observability`
- `github.com/qatoolist/wowapi/kernel/resource`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/internal/cli`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/kernel/attachment`
- `github.com/qatoolist/wowapi/kernel/comment`
- `github.com/qatoolist/wowapi/kernel/document`
- `github.com/qatoolist/wowapi/kernel/webhook`
- `github.com/qatoolist/wowapi/kernel/workflow`
- `github.com/qatoolist/wowapi/module`

## github.com/qatoolist/wowapi/kernel/pagination

- **Package name:** pagination

**Imports:**

- `bytes`
- `encoding/base64`
- `encoding/json`
- `fmt`
- `github.com/google/uuid`
- `github.com/qatoolist/wowapi/kernel/errors`
- `strconv`
- `strings`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/kernel/filtering`
- `github.com/qatoolist/wowapi/kernel/httpx`
- `github.com/qatoolist/wowapi/kernel/workflow`

## github.com/qatoolist/wowapi/kernel/policy

- **Package name:** policy

**Imports:**

- `encoding/json`
- `fmt`
- `github.com/qatoolist/wowapi/kernel/authz`
- `github.com/qatoolist/wowapi/kernel/errors`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/kernel`

## github.com/qatoolist/wowapi/kernel/relationship

- **Package name:** relationship

**Imports:**

- `context`
- `github.com/google/uuid`
- `github.com/qatoolist/wowapi/kernel/authz`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/model`
- `github.com/qatoolist/wowapi/kernel/resource`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/kernel`

## github.com/qatoolist/wowapi/kernel/resource

- **Package name:** resource

**Imports:**

- `context`
- `github.com/google/uuid`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `regexp`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/internal/testmodules/requests`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/kernel/attachment`
- `github.com/qatoolist/wowapi/kernel/authz`
- `github.com/qatoolist/wowapi/kernel/comment`
- `github.com/qatoolist/wowapi/kernel/document`
- `github.com/qatoolist/wowapi/kernel/notify`
- `github.com/qatoolist/wowapi/kernel/outbox`
- `github.com/qatoolist/wowapi/kernel/relationship`
- `github.com/qatoolist/wowapi/kernel/workflow`
- `github.com/qatoolist/wowapi/module`
- `github.com/qatoolist/wowapi/testkit`

## github.com/qatoolist/wowapi/kernel/retention

- **Package name:** retention

**Imports:**

- `context`
- `errors`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/jackc/pgx/v5/pgconn`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/model`
- `sort`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/module`

## github.com/qatoolist/wowapi/kernel/rules

- **Package name:** rules

**Imports:**

- `context`
- `encoding/json`
- `errors`
- `fmt`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/model`
- `regexp`
- `sort`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/module`

## github.com/qatoolist/wowapi/kernel/secrets

- **Package name:** secrets

**Imports:**

- `context`
- `fmt`
- `strings`

**Imported by:**

- `github.com/qatoolist/wowapi/adapters/secrets/envprovider`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/kernel/config`
- `github.com/qatoolist/wowapi/kernel/integration`

## github.com/qatoolist/wowapi/kernel/seeds

- **Package name:** seeds

**Imports:**

- `context`
- `fmt`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `gopkg.in/yaml.v3`
- `io/fs`
- `sort`
- `strings`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/internal/cli`
- `github.com/qatoolist/wowapi/testkit`

## github.com/qatoolist/wowapi/kernel/sequence

- **Package name:** sequence

**Imports:**

- `context`
- `errors`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/model`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/module`

## github.com/qatoolist/wowapi/kernel/storage

- **Package name:** storage

**Imports:**

- `context`
- `crypto/sha256`
- `encoding/hex`
- `github.com/qatoolist/wowapi/kernel/errors`
- `sync`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/kernel/document`

## github.com/qatoolist/wowapi/kernel/validation

- **Package name:** validation

**Imports:**

- `errors`
- `fmt`
- `github.com/go-playground/validator/v10`
- `github.com/qatoolist/wowapi/kernel/errors`
- `reflect`
- `strings`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/internal/testmodules/requests`
- `github.com/qatoolist/wowapi/kernel/httpx`
- `github.com/qatoolist/wowapi/module`

## github.com/qatoolist/wowapi/kernel/webhook

- **Package name:** webhook

**Imports:**

- `bytes`
- `context`
- `crypto/hmac`
- `crypto/sha256`
- `encoding/hex`
- `encoding/json`
- `errors`
- `fmt`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/model`
- `github.com/qatoolist/wowapi/kernel/observability`
- `github.com/qatoolist/wowapi/kernel/outbox`
- `net/http`
- `strings`
- `sync`
- `time`
- `unicode/utf8`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/module`

## github.com/qatoolist/wowapi/kernel/workflow

- **Package name:** workflow

**Imports:**

- `bytes`
- `context`
- `encoding/json`
- `fmt`
- `github.com/google/uuid`
- `github.com/qatoolist/wowapi/kernel/authz`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/errors`
- `github.com/qatoolist/wowapi/kernel/filtering`
- `github.com/qatoolist/wowapi/kernel/model`
- `github.com/qatoolist/wowapi/kernel/outbox`
- `github.com/qatoolist/wowapi/kernel/pagination`
- `github.com/qatoolist/wowapi/kernel/resource`
- `gopkg.in/yaml.v3`
- `sort`
- `strconv`
- `strings`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/module`
- `github.com/qatoolist/wowapi/testkit`

## github.com/qatoolist/wowapi/migrations

- **Package name:** migrations

**Imports:**

- `embed`
- `io/fs`

**Imported by:**

- `github.com/qatoolist/wowapi/internal/tools/migrate`
- `github.com/qatoolist/wowapi/testkit`

## github.com/qatoolist/wowapi/module

- **Package name:** module

**Imports:**

- `context`
- `github.com/qatoolist/wowapi/kernel/artifact`
- `github.com/qatoolist/wowapi/kernel/attachment`
- `github.com/qatoolist/wowapi/kernel/audit`
- `github.com/qatoolist/wowapi/kernel/authz`
- `github.com/qatoolist/wowapi/kernel/bulk`
- `github.com/qatoolist/wowapi/kernel/comment`
- `github.com/qatoolist/wowapi/kernel/config`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/document`
- `github.com/qatoolist/wowapi/kernel/httpx`
- `github.com/qatoolist/wowapi/kernel/integration`
- `github.com/qatoolist/wowapi/kernel/jobs`
- `github.com/qatoolist/wowapi/kernel/model`
- `github.com/qatoolist/wowapi/kernel/notify`
- `github.com/qatoolist/wowapi/kernel/outbox`
- `github.com/qatoolist/wowapi/kernel/resource`
- `github.com/qatoolist/wowapi/kernel/retention`
- `github.com/qatoolist/wowapi/kernel/rules`
- `github.com/qatoolist/wowapi/kernel/sequence`
- `github.com/qatoolist/wowapi/kernel/validation`
- `github.com/qatoolist/wowapi/kernel/webhook`
- `github.com/qatoolist/wowapi/kernel/workflow`
- `io/fs`
- `log/slog`
- `regexp`
- `time`

**Imported by:**

- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/internal/testmodules/requests`
- `github.com/qatoolist/wowapi/testkit`

## github.com/qatoolist/wowapi/testkit

- **Package name:** testkit

**Imports:**

- `context`
- `crypto/rand`
- `crypto/rsa`
- `crypto/sha256`
- `encoding/hex`
- `errors`
- `fmt`
- `github.com/golang-jwt/jwt/v5`
- `github.com/google/uuid`
- `github.com/jackc/pgx/v5`
- `github.com/jackc/pgx/v5/pgconn`
- `github.com/jackc/pgx/v5/pgxpool`
- `github.com/qatoolist/wowapi/app`
- `github.com/qatoolist/wowapi/kernel`
- `github.com/qatoolist/wowapi/kernel/auth`
- `github.com/qatoolist/wowapi/kernel/authz`
- `github.com/qatoolist/wowapi/kernel/config`
- `github.com/qatoolist/wowapi/kernel/database`
- `github.com/qatoolist/wowapi/kernel/resource`
- `github.com/qatoolist/wowapi/kernel/seeds`
- `github.com/qatoolist/wowapi/kernel/workflow`
- `github.com/qatoolist/wowapi/migrations`
- `github.com/qatoolist/wowapi/module`
- `io/fs`
- `log/slog`
- `os`
- `regexp`
- `sort`
- `strings`
- `sync`
- `testing`
- `time`

**Imported by:**

_Not imported by any recorded package._

## github.com/qatoolist/wowapi/testkit/fakes

- **Package name:** fakes

**Imports:**

- `encoding/binary`
- `github.com/google/uuid`
- `github.com/qatoolist/wowapi/kernel/model`
- `sync`
- `time`

**Imported by:**

_Not imported by any recorded package._

