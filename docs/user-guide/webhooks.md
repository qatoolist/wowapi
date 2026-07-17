# Webhooks

`foundation/webhook` is wowapi's webhook subsystem: inbound signature verification + replay protection +
async processing, and outbound signed HTTP delivery with per-endpoint circuit breakers. This page covers
the outbound delivery path's SSRF protection (backlog B2); for the inbound verify/dedup/process pipeline
and the `Service` API, read the package doc in `foundation/webhook/webhook.go`.

## Outbound SSRF protection

Outbound webhook **destinations are user-configurable URLs** — a tenant registers their own
`webhook_endpoints.url`. Without protection, a malicious or careless registration could point delivery
at `http://169.254.169.254/latest/meta-data/` (cloud instance metadata), `http://localhost:6379`
(an internal Redis), or any other address the deployment can reach but the public internet can't — a
live [SSRF](https://owasp.org/www-community/attacks/Server_Side_Request_Forgery) vector.

`kernel.New` wires `webhook.HTTPSender` — the production `Sender` — over **`kernel/httpclient`**, an
SSRF-safe `*http.Client`, **by default**. No opt-in is required; a scaffolded product is protected the
moment it registers an outbound endpoint.

### What's blocked by default

`kernel/httpclient` blocks dial targets in any of these address classes:

| Class | Examples |
|---|---|
| Loopback | `127.0.0.0/8`, `::1` |
| Link-local (unicast + multicast) | `169.254.0.0/16` (incl. the `169.254.169.254` cloud-metadata address), `fe80::/10` |
| Unspecified ("any") | `0.0.0.0`, `::` |
| RFC1918 private | `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16` |
| IPv6 unique local (ULA, RFC4193) | `fc00::/7` |
| RFC6598 carrier-grade NAT (CGNAT) | `100.64.0.0/10` — used internally by AWS, many container/K8s overlays, other cloud infra |
| Multicast | `224.0.0.0/4`, `ff00::/8` |

A blocked dial fails with an error wrapping `httpclient.ErrBlockedAddress`; the outbound delivery pipeline
records it as a normal delivery failure (`delivery_status = failed`, backing off per the usual retry
schedule, eventually `dead` at `MaxAttempts` in `finalizeOutboundDelivery`) — an SSRF attempt is never silently retried forever, and
it never crashes the dispatch loop for other endpoints.

### Resolve-then-verify (DNS-rebinding safe)

The guard does **not** just string-match the hostname before DNS resolution — that would be bypassed by
[DNS rebinding](https://en.wikipedia.org/wiki/DNS_rebinding) (a name that resolves to a public IP when
checked, then a private IP when actually dialed). Instead, `kernel/httpclient` installs a custom
`DialContext` that resolves the hostname **itself**, checks the **resolved IP(s)**, and only then opens
the connection — resolution and verification happen back-to-back with no window for a second, different
answer to slip in.

Because `net/http` invokes `DialContext` again for **every redirect hop**, each hop is independently
re-verified for free: an endpoint that redirects from an allowed public host to an internal one is
blocked on the second hop even though the first succeeded.

### The allowlist escape hatch

Some deployments have a legitimate reason to deliver to an internal target — a tenant's own relay behind
the VPN, a staging receiver, etc. Configure it explicitly rather than disabling protection wholesale:

```yaml
# configs/prod.yaml (or wherever your product's overlay lives)
webhook:
  outbound:
    allowed_hosts:
      - "relay.internal.acme-tenant.example"
    allowed_cidrs:
      - "10.20.0.0/16"   # a specific internal subnet, not the whole private range
```

- `allowed_hosts` is an **exact-match, case-insensitive** hostname list — no wildcards, no subdomain
  matching. A request to that literal host bypasses the resolved-address check entirely.
- `allowed_cidrs` allowlists **resolved addresses** (e.g. a single host as a `/32`, or a subnet). Any
  other hostname that happens to resolve into that range is also allowed — scope it as tightly as the
  real target requires.
- Both lists are opt-in per entry: allowlisting one host/CIDR does not widen the check for anything else.

### The blanket opt-out (local/dev only)

`webhook.outbound.ssrf_protection_disabled: true` removes the guard entirely, falling back to a bare
`net/http.Client`. This exists for local development against a hand-rolled test receiver on `localhost`
where an allowlist entry would be needless ceremony. The field is tagged `unsafe:"true"` — the
framework's standard dev-only-knob gate (`kernel/config/bind.go` `enforceUnsafe`) **refuses it when
`environment: prod`** and **warns loudly when `environment: stage`**, so a product cannot accidentally
ship it live, and staging deployments get an explicit signal instead of silent exposure.

### For product/module authors

If your module builds its own outbound HTTP sender (rather than using `kernel.Webhooks()`), reuse
`kernel/httpclient` directly instead of a bare `http.Client`:

```go
import "github.com/qatoolist/wowapi/kernel/httpclient"

client := httpclient.New(httpclient.Config{
    AllowedHosts: []string{"partner-api.example.com"},
})
```

The zero value `httpclient.Config{}` is safe and maximally restrictive — no allowlist, default timeout.

### Testing your own SSRF-sensitive code

`kernel/httpclient`'s own test suite (`kernel/httpclient/client_test.go`) is a good template: an
`httptest.Server` is always loopback, so a **default-deny client blocking it proves the guard works with
no external network** — allowlist the server's own `127.0.0.1:<port>` host to prove the escape hatch,
and to prove real delivery is unaffected.

## See also

- [Configuration](configuration.md) — the full `webhook.outbound.*` key reference.
- [Building & extending modules](modules.md) — `Webhooks()` on `module.Context`.
- `foundation/webhook/` — the `Service` API (inbound verify/dedup/process, outbound
  dispatch/retry, circuit breaker). Import `foundation/webhook` directly.
- `kernel/httpclient/client.go` — the SSRF guard implementation.
