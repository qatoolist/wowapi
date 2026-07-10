package httpclient

import "net"

// uniqueLocalV6 is RFC4193 fc00::/7 (IPv6 unique local addresses) — the IPv6
// analogue of RFC1918 private space. net.IP has no built-in predicate for it.
var _, uniqueLocalV6, _ = net.ParseCIDR("fc00::/7")

// isBlockedIP reports whether ip must never be dialed by default: loopback,
// link-local (unicast incl. 169.254.169.254 cloud metadata, and multicast),
// unspecified ("any"), RFC1918/ULA private ranges, and general multicast.
// IPv4-mapped IPv6 addresses (::ffff:a.b.c.d) are unwrapped to their IPv4 form
// first so the same address can't bypass the guard by representation alone.
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
		// covers RFC1918 (10/8, 172.16/12, 192.168/16) AND RFC4193 fc00::/7 —
		// net.IP.IsPrivate implements both per the Go stdlib docs, but the
		// explicit ULA check below stays as defense-in-depth against any
		// stdlib version-specific narrowing.
		return true
	case uniqueLocalV6.Contains(ip):
		return true
	default:
		return false
	}
}
