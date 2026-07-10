package httpclient

import (
	"fmt"
	"net"
	"strings"
)

// allowlist is the escape hatch for intentional internal targets (backlog
// B2): hosts and/or CIDRs that bypass the default-deny address-class checks.
// Both dimensions are exact/CIDR match only — no wildcards, no suffix
// matching — so an operator opts a specific target in deliberately rather
// than accidentally widening the hole.
type allowlist struct {
	hosts map[string]struct{} // lower-cased exact hostnames
	nets  []*net.IPNet
}

// newAllowlist builds an allowlist from raw config values. hosts are matched
// case-insensitively and exactly (a bare hostname, no port). cidrs must each
// parse as a valid CIDR (e.g. "10.0.0.0/8" or a /32 for a single host).
func newAllowlist(hosts []string, cidrs []string) (*allowlist, error) {
	al := &allowlist{hosts: make(map[string]struct{}, len(hosts))}
	for _, h := range hosts {
		al.hosts[strings.ToLower(strings.TrimSpace(h))] = struct{}{}
	}
	for _, c := range cidrs {
		_, n, err := net.ParseCIDR(strings.TrimSpace(c))
		if err != nil {
			return nil, fmt.Errorf("httpclient: invalid allowlist CIDR %q: %w", c, err)
		}
		al.nets = append(al.nets, n)
	}
	return al, nil
}

// allowsHost reports whether host (no port, as sent to a dialer) is
// explicitly allowlisted by exact, case-insensitive match.
func (al *allowlist) allowsHost(host string) bool {
	_, ok := al.hosts[strings.ToLower(strings.TrimSpace(host))]
	return ok
}

// allowsIP reports whether ip falls within any allowlisted CIDR.
func (al *allowlist) allowsIP(ip net.IP) bool {
	for _, n := range al.nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
