// Package httpclient builds SSRF-safe *http.Client instances for outbound
// calls to user-configurable destinations (webhook targets, integration
// callbacks, …). It is dial-time protection: loopback, link-local (incl. the
// 169.254.169.254 cloud-metadata address), RFC1918/ULA private ranges, and
// unspecified addresses are refused, with an explicit host/CIDR allowlist as
// the escape hatch for intentional internal targets.
//
// Protection is resolve-then-verify: the custom DialContext resolves the
// hostname itself and checks the RESOLVED IPs, never the pre-DNS hostname, so
// a DNS-rebinding attacker (a name that resolves to a public IP at
// allowlist-check time but a private IP at connect time) cannot bypass it —
// there is only one resolution here and it happens immediately before Dial.
// Because http.Client re-invokes the Transport (and therefore DialContext)
// for every redirect hop, each hop is independently re-verified for free.
//
// Contract: backlog B2 (docs/implementation/framework-engineering-backlog.md).
package httpclient

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// DefaultTimeout is the client-wide request ceiling applied when Config.Timeout
// is zero. Callers with a shorter operational SLA (e.g. webhook delivery) pass
// their own Timeout or further bound the call via the request context.
const DefaultTimeout = 10 * time.Second

// ErrBlockedAddress is wrapped into every error returned when a dial target
// (or a resolved address behind a hostname) falls into a blocked address
// class and is not covered by the allowlist. Callers can check
// errors.Is(err, ErrBlockedAddress) to distinguish this from ordinary network
// failures.
var ErrBlockedAddress = errors.New("httpclient: destination address is blocked by SSRF policy")

// Config controls the SSRF guard and the underlying transport. The zero value
// is safe and maximally restrictive: no allowlist entries, default timeout —
// i.e. every private/loopback/link-local/metadata/unspecified address is
// blocked and only public destinations are reachable.
type Config struct {
	// AllowedHosts is the exact-match (case-insensitive), no-wildcard hostname
	// allowlist. A request whose URL host is listed here bypasses the
	// resolved-IP check entirely for that host (the operator vouches for a
	// specific, intentional internal target).
	AllowedHosts []string
	// AllowedCIDRs is the allowlist for RESOLVED addresses, e.g. "10.20.0.0/16"
	// or a single host as "10.20.1.5/32". A resolved IP inside any of these
	// networks bypasses the blocked-address-class check.
	AllowedCIDRs []string
	// Timeout is the client-wide request ceiling. Zero uses DefaultTimeout.
	Timeout time.Duration
}

// New builds an SSRF-safe *http.Client per Config. It panics if an
// AllowedCIDRs entry fails to parse — this is caller-supplied configuration
// validated at construction (boot time in production wiring), not a runtime
// condition to recover from.
func New(cfg Config) *http.Client {
	g := newDialGuard(cfg)

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	// SSRF safety: http.DefaultTransport.Proxy is http.ProxyFromEnvironment,
	// which Clone() preserves. If left in place, an HTTP_PROXY/HTTPS_PROXY
	// env var would make the transport dial the PROXY's address instead of
	// the request's real destination — the proxy's IP passes the guard (it
	// may be entirely legitimate), but the attacker-controlled final URL is
	// then sent to the proxy in the request/CONNECT line, and the guarded
	// DialContext below never sees, let alone validates, the actual target.
	// Disabling the proxy is the only way to guarantee the dial-time IP
	// check always runs against the real destination. A deployment that
	// genuinely needs an explicit egress proxy for outbound calls is a
	// distinct future feature (the proxy's own address/allowlisting would
	// need first-class support here) — not a config knob to bolt on now.
	transport.Proxy = nil
	transport.DialContext = g.dialContext((&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext)

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// resolveFunc resolves host to its candidate IP addresses. Production uses
// net.DefaultResolver; tests inject a fake to simulate DNS-rebind scenarios
// deterministically without touching real DNS.
type resolveFunc func(ctx context.Context, host string) ([]net.IP, error)

// dialGuard is the resolve-then-verify core: given a host:port dial target,
// it resolves the host, checks every resolved address against the blocked
// classes (unless the host or address is allowlisted), and only then permits
// the wrapped dialer to run.
type dialGuard struct {
	allow     *allowlist
	resolveFn resolveFunc
}

func newDialGuard(cfg Config) *dialGuard {
	al, err := newAllowlist(cfg.AllowedHosts, cfg.AllowedCIDRs)
	if err != nil {
		panic(err)
	}
	return &dialGuard{
		allow:     al,
		resolveFn: defaultResolve,
	}
}

func defaultResolve(ctx context.Context, host string) ([]net.IP, error) {
	// If host is already a literal IP, skip the resolver — net.LookupIP would
	// otherwise just hand it back, but this avoids an unnecessary syscall and
	// keeps behavior identical for literal-IP dial targets.
	if ip := net.ParseIP(host); ip != nil {
		return []net.IP{ip}, nil
	}
	ipAddrs, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	return ipAddrs, nil
}

// checkResolvedIPs enforces the address-class policy for host's resolved
// addresses ips. An allowlisted host bypasses the check unconditionally.
// Otherwise EVERY address must be either non-blocked or individually
// allowlisted by CIDR — a single unsafe candidate among several resolved
// addresses fails the whole dial closed, since an attacker-controlled
// resolver could otherwise round-robin between a decoy and a real target.
// An empty resolution set is treated as blocked (fail closed).
func (g *dialGuard) checkResolvedIPs(host string, ips []net.IP) error {
	if g.allow.allowsHost(host) {
		return nil
	}
	if len(ips) == 0 {
		return fmt.Errorf("%w: host %q resolved to no addresses", ErrBlockedAddress, host)
	}
	for _, ip := range ips {
		if g.allow.allowsIP(ip) {
			continue
		}
		if isBlockedIP(ip) {
			return fmt.Errorf("%w: host %q resolved to %s", ErrBlockedAddress, host, ip)
		}
	}
	return nil
}

// baseDialFunc matches net.Dialer.DialContext's signature so dialContext can
// wrap either the real dialer or a test double.
type baseDialFunc func(ctx context.Context, network, addr string) (net.Conn, error)

// dialContext returns a DialContext function that resolves addr's host,
// verifies the resolved addresses against policy, and dials one of the
// VERIFIED addresses directly. This is THE enforcement point: net/http calls
// it fresh for the initial connection and for every redirect hop, so
// resolve-then-verify automatically re-runs per hop with no extra wiring.
//
// Critically, base is called with the verified IP (net.JoinHostPort(ip, port)),
// NEVER the hostname: if we dialed the hostname, net.Dialer would perform its
// own second DNS lookup, and an attacker-controlled resolver could return a
// safe IP for the verification lookup and a blocked IP for that second
// lookup — the DNS-rebinding / TOCTOU bypass this package exists to close.
// Dialing the exact IP we checked collapses the two resolutions into one.
// TLS SNI and the Host header are unaffected: http.Transport derives them from
// the request URL's host, not the dial address.
func (g *dialGuard) dialContext(base baseDialFunc) baseDialFunc {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("httpclient: malformed dial address %q: %w", addr, err)
		}
		ips, err := g.resolveFn(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("httpclient: resolve %q: %w", host, err)
		}
		if err := g.checkResolvedIPs(host, ips); err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			// Only reachable for an allowlisted host whose resolver returned no
			// addresses (checkResolvedIPs bypasses the empty check for those).
			// There is nothing safe to dial — fail closed rather than fall back
			// to dialing the hostname (which would re-resolve).
			return nil, fmt.Errorf("%w: host %q resolved to no addresses", ErrBlockedAddress, host)
		}
		// Dial the verified IPs in order (mirrors the stdlib dialer's multi-record
		// fallback); every ip already passed checkResolvedIPs, so each is safe.
		var lastErr error
		for _, ip := range ips {
			conn, derr := base(ctx, network, net.JoinHostPort(ip.String(), port))
			if derr == nil {
				return conn, nil
			}
			lastErr = derr
		}
		return nil, lastErr
	}
}
