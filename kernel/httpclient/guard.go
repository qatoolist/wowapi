package httpclient

import "net"

// cgnatV4 is RFC 6598 100.64.0.0/10, the carrier-grade-NAT shared address
// space. net.IP has no built-in predicate for it, and it is NOT covered by
// IsPrivate (that's RFC1918 + RFC4193 only) — yet it is exactly as
// SSRF-relevant as RFC1918: AWS, many container/K8s overlay networks, and
// other cloud infra route internal traffic through it.
var cgnatV4 = mustParseCIDR("100.64.0.0/10")

func mustParseCIDR(s string) *net.IPNet {
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		panic("httpclient: invalid built-in CIDR " + s + ": " + err.Error())
	}
	return n
}

// isBlockedIP reports whether ip must never be dialed by default: loopback,
// link-local (unicast incl. 169.254.169.254 cloud metadata, and multicast),
// unspecified ("any"), RFC1918/ULA private ranges, RFC6598 CGNAT shared space,
// and general multicast. IPv4-mapped IPv6 addresses (::ffff:a.b.c.d) are
// unwrapped to their IPv4 form first so the same address can't bypass the
// guard by representation alone.
//
// net.IP.IsPrivate covers BOTH RFC1918 (10/8, 172.16/12, 192.168/16) and
// RFC4193 fc00::/7 (IPv6 unique local addresses) per the Go stdlib docs, so
// there is no separate ULA check here.
//
// A nil or otherwise unparseable IP is blocked — fail closed, never open
// (SEC posture: an address we cannot classify is treated as unsafe).
func isBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if v4 := ip.To4(); v4 != nil {
		ip = v4
	}
	switch {
	case ip.IsLoopback():
		return true
	case ip.IsLinkLocalUnicast():
		return true
	case ip.IsLinkLocalMulticast():
		return true
	case ip.IsUnspecified():
		return true
	case ip.IsMulticast():
		return true
	case ip.IsPrivate():
		return true
	case cgnatV4.Contains(ip):
		return true
	default:
		return false
	}
}
