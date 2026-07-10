package httpclient

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// --- resolve-then-verify unit tests (no real network) ---

// fakeResolveFunc lets tests substitute the DNS-rebind-vulnerable step: what
// IPs a hostname resolves to at dial time. This is the hook the acceptance
// criteria calls for ("DNS rebind style test... via a custom resolver hook").
type fakeResolver struct {
	ips []net.IP
	err error
}

func (f fakeResolver) resolve(_ context.Context, _ string) ([]net.IP, error) {
	return f.ips, f.err
}

func TestGuardCheckResolvedBlocksPrivateIP(t *testing.T) {
	g := newDialGuard(Config{})
	err := g.checkResolvedIPs("internal-looking-name", []net.IP{net.ParseIP("10.1.2.3")})
	if err == nil {
		t.Fatal("expected a private resolved IP to be blocked")
	}
	if !errors.Is(err, ErrBlockedAddress) {
		t.Errorf("expected ErrBlockedAddress, got %v", err)
	}
}

func TestGuardCheckResolvedAllowsPublicIP(t *testing.T) {
	g := newDialGuard(Config{})
	if err := g.checkResolvedIPs("example.com", []net.IP{net.ParseIP("93.184.216.34")}); err != nil {
		t.Fatalf("expected public IP to be allowed, got %v", err)
	}
}

func TestGuardCheckResolvedNoAddressesBlocked(t *testing.T) {
	g := newDialGuard(Config{})
	if err := g.checkResolvedIPs("nowhere.example", nil); err == nil {
		t.Fatal("expected an empty resolution set to be blocked (fail closed)")
	}
}

func TestGuardAnyBlockedAmongManyBlocks(t *testing.T) {
	// If a hostname resolves to MULTIPLE addresses and even one is unsafe, the
	// whole dial must be refused — an attacker-controlled DNS response could
	// otherwise round-robin between a decoy public IP and an internal one.
	g := newDialGuard(Config{})
	err := g.checkResolvedIPs("multi.example", []net.IP{
		net.ParseIP("93.184.216.34"),   // public
		net.ParseIP("169.254.169.254"), // metadata — blocked
	})
	if err == nil {
		t.Fatal("expected the dial to be blocked when any resolved address is unsafe")
	}
}

func TestGuardAllowlistedHostBypassesIPCheck(t *testing.T) {
	g := newDialGuard(Config{AllowedHosts: []string{"internal.example"}})
	if err := g.checkResolvedIPs("internal.example", []net.IP{net.ParseIP("10.1.2.3")}); err != nil {
		t.Fatalf("expected an allowlisted host to bypass the IP check, got %v", err)
	}
	// A different, non-allowlisted host resolving to the same private IP must
	// still be blocked — the allowlist opts in a HOST, not the address itself.
	if err := g.checkResolvedIPs("other.example", []net.IP{net.ParseIP("10.1.2.3")}); err == nil {
		t.Fatal("expected a non-allowlisted host to stay blocked")
	}
}

func TestGuardAllowlistedCIDRBypassesIPCheck(t *testing.T) {
	g := newDialGuard(Config{AllowedCIDRs: []string{"127.0.0.1/32"}})
	if err := g.checkResolvedIPs("localhost", []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		t.Fatalf("expected the allowlisted CIDR to bypass the IP check, got %v", err)
	}
	if err := g.checkResolvedIPs("localhost", []net.IP{net.ParseIP("127.0.0.2")}); err == nil {
		t.Fatal("expected an IP outside the allowlisted /32 to stay blocked")
	}
}

// --- resolve-then-verify: DNS-rebind style test via a custom resolver hook ---

func TestDialContextBlocksHostnameResolvingToPrivateIP(t *testing.T) {
	g := newDialGuard(Config{})
	g.resolveFn = fakeResolver{ips: []net.IP{net.ParseIP("169.254.169.254")}}.resolve

	dial := g.dialContext(func(ctx context.Context, network, addr string) (net.Conn, error) {
		t.Fatal("the real dialer must never be invoked once resolution is blocked")
		return nil, nil
	})
	_, err := dial(context.Background(), "tcp", "rebind.attacker.example:80")
	if err == nil {
		t.Fatal("expected DialContext to block a hostname resolving to a metadata IP")
	}
	if !errors.Is(err, ErrBlockedAddress) {
		t.Errorf("expected ErrBlockedAddress, got %v", err)
	}
}

func TestDialContextAllowsHostnameResolvingToPublicIP(t *testing.T) {
	g := newDialGuard(Config{})
	g.resolveFn = fakeResolver{ips: []net.IP{net.ParseIP("93.184.216.34")}}.resolve

	called := false
	dial := g.dialContext(func(ctx context.Context, network, addr string) (net.Conn, error) {
		called = true
		return nil, errors.New("no real dial in this unit test")
	})
	_, _ = dial(context.Background(), "tcp", "example.com:80")
	if !called {
		t.Fatal("expected the underlying dialer to be invoked for a public resolution")
	}
}

// TestDialContextAllowlistedHostWithNoResolvedAddressesFailsClosed covers the
// one edge case checkResolvedIPs itself can't guard: an ALLOWLISTED host
// bypasses the "empty resolution" check there (an allowlisted host is always
// permitted regardless of what it resolves to), but dialContext must still
// refuse to dial when there is genuinely nothing to connect to — never fall
// back to dialing the bare hostname, which would let the dialer re-resolve.
func TestDialContextAllowlistedHostWithNoResolvedAddressesFailsClosed(t *testing.T) {
	g := newDialGuard(Config{AllowedHosts: []string{"internal.example"}})
	g.resolveFn = fakeResolver{ips: nil}.resolve

	dial := g.dialContext(func(ctx context.Context, network, addr string) (net.Conn, error) {
		t.Fatal("dialer must not be reached when there are no resolved addresses to dial")
		return nil, nil
	})
	_, err := dial(context.Background(), "tcp", "internal.example:80")
	if err == nil {
		t.Fatal("expected an error when an allowlisted host resolves to nothing")
	}
	if !errors.Is(err, ErrBlockedAddress) {
		t.Errorf("expected ErrBlockedAddress, got %v", err)
	}
}

// TestDialContextDialsVerifiedIPNotHostname is the DNS-rebinding regression
// guard: the dialer MUST be handed the exact IP that was verified, never the
// hostname. If it received the hostname, net.Dialer would run a second,
// independent DNS lookup — and an attacker-controlled resolver could answer
// that second lookup with a blocked IP after answering the verification lookup
// with a safe one (TOCTOU). This test fails if the guard ever regresses to
// dialing addr's hostname.
func TestDialContextDialsVerifiedIPNotHostname(t *testing.T) {
	g := newDialGuard(Config{})
	g.resolveFn = fakeResolver{ips: []net.IP{net.ParseIP("93.184.216.34")}}.resolve

	var dialedAddr string
	dial := g.dialContext(func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialedAddr = addr
		return nil, errors.New("no real dial in this unit test")
	})
	_, _ = dial(context.Background(), "tcp", "rebind.example:80")
	if dialedAddr != "93.184.216.34:80" {
		t.Fatalf("dialed %q, want the verified IP 93.184.216.34:80 — dialing the hostname would let the dialer re-resolve and defeat DNS-rebinding protection", dialedAddr)
	}
}

// TestDialContextTriesAllVerifiedIPs proves that when a host resolves to
// multiple verified addresses and the first dial fails, the dialer falls back
// to the next — mirroring the stdlib dialer's multi-record behavior, but only
// over addresses that already passed verification.
func TestDialContextTriesAllVerifiedIPs(t *testing.T) {
	g := newDialGuard(Config{})
	g.resolveFn = fakeResolver{ips: []net.IP{
		net.ParseIP("93.184.216.34"),
		net.ParseIP("1.1.1.1"),
	}}.resolve

	var dialed []string
	dial := g.dialContext(func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialed = append(dialed, addr)
		return nil, errors.New("dial fails so the loop advances")
	})
	_, _ = dial(context.Background(), "tcp", "multi.example:443")
	want := []string{"93.184.216.34:443", "1.1.1.1:443"}
	if len(dialed) != len(want) || dialed[0] != want[0] || dialed[1] != want[1] {
		t.Fatalf("dialed %v, want %v (both verified IPs tried in order)", dialed, want)
	}
}

// --- defaultResolve: the production resolveFunc, exercised directly ---

func TestDefaultResolveLiteralIPSkipsLookup(t *testing.T) {
	ips, err := defaultResolve(context.Background(), "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ips) != 1 || !ips[0].Equal(net.ParseIP("127.0.0.1")) {
		t.Fatalf("ips = %v, want [127.0.0.1]", ips)
	}
}

func TestDefaultResolveHostname(t *testing.T) {
	// localhost is guaranteed resolvable in any test environment (no external
	// network dependency) and always resolves to loopback addresses.
	ips, err := defaultResolve(context.Background(), "localhost")
	if err != nil {
		t.Fatalf("unexpected error resolving localhost: %v", err)
	}
	if len(ips) == 0 {
		t.Fatal("expected at least one resolved address for localhost")
	}
	for _, ip := range ips {
		if !ip.IsLoopback() {
			t.Errorf("resolved %s is not loopback", ip)
		}
	}
}

func TestDefaultResolveUnresolvableHostnameErrors(t *testing.T) {
	// A syntactically invalid DNS label guarantees a resolution failure
	// without depending on external network reachability.
	_, err := defaultResolve(context.Background(), "this..is..not..a..valid..hostname..")
	if err == nil {
		t.Fatal("expected an error resolving an invalid hostname")
	}
}

func TestDialContextSurfacesResolveError(t *testing.T) {
	g := newDialGuard(Config{})
	g.resolveFn = fakeResolver{err: errors.New("dns server unreachable")}.resolve

	dial := g.dialContext(func(ctx context.Context, network, addr string) (net.Conn, error) {
		t.Fatal("dialer must not be reached when resolution itself errors")
		return nil, nil
	})
	_, err := dial(context.Background(), "tcp", "example.com:80")
	if err == nil {
		t.Fatal("expected the resolution error to be surfaced")
	}
	if errors.Is(err, ErrBlockedAddress) {
		t.Error("a resolver failure is a different error class than a blocked address")
	}
}

func TestDialContextRejectsMalformedAddress(t *testing.T) {
	g := newDialGuard(Config{})
	dial := g.dialContext(func(ctx context.Context, network, addr string) (net.Conn, error) {
		t.Fatal("dialer must not be reached for a malformed address")
		return nil, nil
	})
	if _, err := dial(context.Background(), "tcp", "not-a-valid-host-port"); err == nil {
		t.Fatal("expected an error for a malformed host:port")
	}
}

// --- end-to-end: httptest loopback server, blocked by default, allowed when allowlisted ---

func TestClientBlocksLoopbackByDefault(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := New(Config{})
	resp, err := client.Get(srv.URL)
	if err == nil {
		resp.Body.Close()
		t.Fatal("expected the default-deny client to block a loopback httptest server")
	}
	if !errors.Is(err, ErrBlockedAddress) {
		t.Errorf("expected the error chain to contain ErrBlockedAddress, got %v", err)
	}
}

func TestClientAllowsLoopbackWhenAllowlisted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	host, _, err := net.SplitHostPort(strings.TrimPrefix(strings.TrimPrefix(srv.URL, "http://"), "https://"))
	if err != nil {
		t.Fatal(err)
	}

	client := New(Config{AllowedHosts: []string{host}})
	resp, err := client.Get(srv.URL)
	if err != nil {
		t.Fatalf("expected the allowlisted loopback target to succeed, got %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestClientPublicDeliveryUnaffectedWhenAllowlisted(t *testing.T) {
	// Simulates "public delivery unaffected" by allowlisting the httptest addr
	// (loopback stands in for a public target reachable at dial time) — proves
	// the guard does not interfere with a request to an allowed destination
	// beyond the address-class check itself.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()
	host, _, err := net.SplitHostPort(strings.TrimPrefix(srv.URL, "http://"))
	if err != nil {
		t.Fatal(err)
	}

	client := New(Config{AllowedHosts: []string{host}})
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

// --- redirect-to-internal blocked; each hop re-checked ---

// newClientWithGuard builds an *http.Client around a caller-provided
// dialGuard, mirroring what New does internally. Tests use this (instead of
// the public New) when they need to swap resolveFn after construction, since
// Config alone cannot express "resolve this hostname to that IP".
func newClientWithGuard(g *dialGuard) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = g.dialContext((&net.Dialer{}).DialContext)
	return &http.Client{Transport: transport}
}

func TestClientBlocksRedirectToLoopback(t *testing.T) {
	// httptest.NewServer always binds 127.0.0.1, so two real loopback servers
	// can't be distinguished by a host-string allowlist alone. Instead: the
	// redirector (a real, allowlisted loopback server) 302s to a HOSTNAME
	// (never a raw IP) that a fake resolver resolves to a blocked metadata
	// address. Because http.Client invokes DialContext fresh for each
	// redirect hop, hop 1 (redirector) succeeds and hop 2 (the fake hostname)
	// is independently re-verified and blocked — proving per-hop re-checking
	// without ever needing a second real listener.
	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://internal.rebind.test:80/secret", http.StatusFound)
	}))
	defer redirector.Close()

	redirectHost, _, err := net.SplitHostPort(strings.TrimPrefix(redirector.URL, "http://"))
	if err != nil {
		t.Fatal(err)
	}

	g := newDialGuard(Config{AllowedHosts: []string{redirectHost}})
	realResolve := g.resolveFn
	g.resolveFn = func(ctx context.Context, host string) ([]net.IP, error) {
		if host == "internal.rebind.test" {
			return []net.IP{net.ParseIP("169.254.169.254")}, nil
		}
		return realResolve(ctx, host)
	}

	client := newClientWithGuard(g)
	resp, err := client.Get(redirector.URL)
	if err == nil {
		resp.Body.Close()
		t.Fatal("expected the redirect hop to a blocked hostname to be refused")
	}
	if !errors.Is(err, ErrBlockedAddress) {
		t.Errorf("expected ErrBlockedAddress in the redirect-hop error chain, got %v", err)
	}
}

func TestClientAllowsRedirectWhenBothHopsAllowlisted(t *testing.T) {
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer internal.Close()

	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, internal.URL, http.StatusFound)
	}))
	defer redirector.Close()

	redirectHost, _, err := net.SplitHostPort(strings.TrimPrefix(redirector.URL, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	internalHost, _, err := net.SplitHostPort(strings.TrimPrefix(internal.URL, "http://"))
	if err != nil {
		t.Fatal(err)
	}

	client := New(Config{AllowedHosts: []string{redirectHost, internalHost}})
	resp, err := client.Get(redirector.URL)
	if err != nil {
		t.Fatalf("expected both allowlisted hops to succeed, got %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

// --- config validation / defaults ---

func TestNewAppliesDefaultTimeout(t *testing.T) {
	client := New(Config{})
	if client.Timeout <= 0 {
		t.Fatal("expected New to apply a non-zero default timeout")
	}
}

func TestNewHonorsCustomTimeout(t *testing.T) {
	client := New(Config{Timeout: 3 * time.Second})
	if client.Timeout != 3*time.Second {
		t.Fatalf("Timeout = %v, want 3s", client.Timeout)
	}
}

func TestNewInvalidAllowlistCIDRPanicsDescriptively(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected New to panic on an invalid allowlist CIDR")
		}
		msg := ""
		if e, ok := r.(error); ok {
			msg = e.Error()
		} else if s, ok := r.(string); ok {
			msg = s
		}
		if !strings.Contains(msg, "invalid allowlist CIDR") {
			t.Errorf("panic value = %v, want it to name the bad CIDR", r)
		}
	}()
	New(Config{AllowedCIDRs: []string{"not-a-cidr"}})
}
