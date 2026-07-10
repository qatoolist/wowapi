package httpclient

import "testing"

func TestAllowlistHostMatch(t *testing.T) {
	al, err := newAllowlist([]string{"internal.example.com", "Other.Example.COM"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !al.allowsHost("internal.example.com") {
		t.Error("expected exact host match to be allowed")
	}
	if !al.allowsHost("other.example.com") {
		t.Error("expected host match to be case-insensitive")
	}
	if al.allowsHost("evil.example.com") {
		t.Error("unrelated host must not be allowed")
	}
	if al.allowsHost("sub.internal.example.com") {
		t.Error("allowlist must be exact-match, not a suffix match, by default")
	}
}

func TestAllowlistCIDRMatch(t *testing.T) {
	al, err := newAllowlist(nil, []string{"10.20.0.0/16", "fd00:abcd::/32"})
	if err != nil {
		t.Fatal(err)
	}
	if !al.allowsIP(mustParseIP(t, "10.20.1.1")) {
		t.Error("expected IP within allowlisted CIDR to be allowed")
	}
	if al.allowsIP(mustParseIP(t, "10.21.1.1")) {
		t.Error("IP outside allowlisted CIDR must not be allowed")
	}
	if !al.allowsIP(mustParseIP(t, "fd00:abcd::1")) {
		t.Error("expected IPv6 CIDR match to be allowed")
	}
}

func TestAllowlistLoopbackCIDRForTest(t *testing.T) {
	// This is the trick the httptest-based dial tests rely on: allowlisting
	// 127.0.0.1/32 (or 127.0.0.0/8) lets a loopback httptest server through
	// while every other loopback/private target stays blocked.
	al, err := newAllowlist(nil, []string{"127.0.0.1/32"})
	if err != nil {
		t.Fatal(err)
	}
	if !al.allowsIP(mustParseIP(t, "127.0.0.1")) {
		t.Error("expected the exact allowlisted loopback IP to be allowed")
	}
	if al.allowsIP(mustParseIP(t, "127.0.0.2")) {
		t.Error("a different loopback IP outside the /32 must stay blocked")
	}
}

func TestAllowlistInvalidCIDRErrors(t *testing.T) {
	if _, err := newAllowlist(nil, []string{"not-a-cidr"}); err == nil {
		t.Fatal("expected an error for an invalid CIDR entry")
	}
}

func TestAllowlistEmptyAllowsNothing(t *testing.T) {
	al, err := newAllowlist(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if al.allowsHost("anything.example.com") || al.allowsIP(mustParseIP(t, "127.0.0.1")) {
		t.Error("an empty allowlist must allow nothing")
	}
}
