package httpclient

import (
	"net"
	"testing"
)

func mustParseIP(t *testing.T, s string) net.IP {
	t.Helper()
	ip := net.ParseIP(s)
	if ip == nil {
		t.Fatalf("net.ParseIP(%q) = nil", s)
	}
	return ip
}

// Table-driven blocked-dial classification tests: one case per address class
// named in the backlog acceptance criteria (B2). isBlockedIP is the pure
// unit-testable core that dialGuard.checkIPs calls per resolved address.
func TestIsBlockedIP(t *testing.T) {
	cases := []struct {
		name    string
		ip      string
		blocked bool
	}{
		// Loopback
		{"loopback v4", "127.0.0.1", true},
		{"loopback v4 other", "127.5.5.5", true},
		{"loopback v6", "::1", true},

		// Link-local unicast, including the cloud-metadata address.
		{"link-local v4", "169.254.1.1", true},
		{"cloud metadata", "169.254.169.254", true},
		{"link-local v6", "fe80::1", true},

		// Unspecified ("any") address.
		{"unspecified v4", "0.0.0.0", true},
		{"unspecified v6", "::", true},

		// RFC1918 private ranges.
		{"rfc1918 10/8", "10.0.0.1", true},
		{"rfc1918 172.16/12 low", "172.16.0.1", true},
		{"rfc1918 172.16/12 high", "172.31.255.254", true},
		{"rfc1918 192.168/16", "192.168.1.1", true},

		// IPv4-mapped IPv6 forms of private/loopback addresses must not bypass
		// the guard via representation trickery.
		{"v4-mapped loopback", "::ffff:127.0.0.1", true},
		{"v4-mapped rfc1918", "::ffff:10.1.2.3", true},

		// ULA (IPv6 unique local, RFC4193) fc00::/7.
		{"ULA fc00::/7 low", "fc00::1", true},
		{"ULA fc00::/7 high", "fdff::1", true},

		// Link-local multicast / multicast in general are not globally
		// routable either; treat as blocked (belt-and-braces).
		{"multicast v4", "224.0.0.1", true},
		{"multicast v6", "ff02::1", true},

		// Public / globally routable addresses must NOT be blocked.
		{"public v4 (google dns)", "8.8.8.8", false},
		{"public v4 (cloudflare dns)", "1.1.1.1", false},
		{"public v6", "2606:4700:4700::1111", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ip := mustParseIP(t, c.ip)
			got := isBlockedIP(ip)
			if got != c.blocked {
				t.Errorf("isBlockedIP(%s) = %v, want %v", c.ip, got, c.blocked)
			}
		})
	}
}

func TestIsBlockedIPNilIsBlocked(t *testing.T) {
	// A nil/unparseable IP must fail closed (blocked), never open.
	if !isBlockedIP(nil) {
		t.Fatal("isBlockedIP(nil) must be true (fail closed)")
	}
}
