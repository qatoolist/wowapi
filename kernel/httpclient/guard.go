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

// nat64Prefix is the RFC 6052 "well-known prefix" 64:ff9b::/96 used by
// NAT64/DNS64 to synthesize an IPv6 address for an IPv4-only destination. The
// embedded IPv4 address occupies the last 4 bytes (bytes 12-15).
//
// Only the well-known /96 is covered here. RFC 6052 also defines
// Network-Specific Prefixes (/32,/40,/48,/56,/64) that embed the IPv4 octets
// at non-contiguous positions (skipping byte 8); those are site-local,
// operator-chosen, and not attacker-guessable in the generic case, so they
// are intentionally out of scope for this guard.
var nat64Prefix = mustParseCIDR("64:ff9b::/96")

// sixToFourPrefix is the RFC 3056 6to4 range 2002::/16. The embedded IPv4
// address occupies bytes 2-5 (immediately following the 2002 prefix word).
var sixToFourPrefix = mustParseCIDR("2002::/16")

// ffff0000Prefix is the ::ffff:0:0/96 "IPv4-translatable" range from RFC
// 4291 §2.5.5.1 (distinct from the ::ffff:a.b.c.d IPv4-mapped range that
// net.IP.To4 already unwraps). The embedded IPv4 address occupies the last
// 4 bytes (bytes 12-15).
//
// This can't be built with net.ParseCIDR("::ffff:0:0/96"): Go's IPv6 text
// parser recognizes the ::ffff:0:0 literal as looking like an IPv4-mapped
// address (::ffff:0.0.0) and silently collapses the parsed network down to
// the IPv4 0.0.0.0/0 — matching every IPv4 address, not the intended IPv6
// /96. Constructing the IPNet directly from raw bytes avoids that trap.
var ffff0000Prefix = &net.IPNet{
	IP:   net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff, 0, 0, 0, 0, 0, 0},
	Mask: net.CIDRMask(96, 128),
}

// embeddedV4 reports whether ip is one of the IPv6 transition/embedding
// forms that carry a literal IPv4 address inside an IPv6 address — NAT64
// (64:ff9b::/96), 6to4 (2002::/16), or the ::ffff:0:0/96 SIIT/alt-mapped
// range — and if so returns that embedded IPv4 address.
//
// These forms exist so IPv4-only destinations remain reachable from
// IPv6-only or DNS64/NAT64 networks; the IPv6 literal itself is not in any
// private/link-local range and would otherwise classify as ordinary global
// unicast, letting a blocked IPv4 target (e.g. 169.254.169.254 cloud
// metadata) through under an IPv6 disguise.
func embeddedV4(ip net.IP) (net.IP, bool) {
	ip16 := ip.To16()
	if ip16 == nil {
		return nil, false
	}
	switch {
	case nat64Prefix.Contains(ip16), ffff0000Prefix.Contains(ip16):
		return net.IPv4(ip16[12], ip16[13], ip16[14], ip16[15]), true
	case sixToFourPrefix.Contains(ip16):
		return net.IPv4(ip16[2], ip16[3], ip16[4], ip16[5]), true
	default:
		return nil, false
	}
}

// isBlockedIP reports whether ip must never be dialed by default: loopback,
// link-local (unicast incl. 169.254.169.254 cloud metadata, and multicast),
// unspecified ("any"), RFC1918/ULA private ranges, RFC6598 CGNAT shared space,
// and general multicast. IPv4-mapped IPv6 addresses (::ffff:a.b.c.d) are
// unwrapped to their IPv4 form first so the same address can't bypass the
// guard by representation alone.
//
// IPv6 transition/embedding forms — NAT64 (64:ff9b::/96), 6to4 (2002::/16),
// and the ::ffff:0:0/96 SIIT/alt-mapped range — are also unwrapped: the
// embedded IPv4 address is extracted and re-checked against every rule below
// (via a recursive call to isBlockedIP), so an attacker cannot smuggle a
// blocked v4 target (metadata, RFC1918, loopback, ...) past the guard by
// wrapping it in one of these IPv6 forms. Addresses outside all three
// embedding prefixes fall through to ordinary IPv6 classification below,
// unchanged from before this unwrapping was added.
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
	} else if embedded, ok := embeddedV4(ip); ok {
		return isBlockedIP(embedded)
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
