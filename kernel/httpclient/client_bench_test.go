package httpclient

import (
	"context"
	"errors"
	"net"
	"testing"
)

var errBenchmarkDial = errors.New("benchmark dial boundary")

// BenchmarkGuardedDial measures the exact resolve -> classify -> verified-IP
// dial path. Resolution and the terminal dial syscall are deterministic doubles;
// all SSRF classification and address rewriting are production code.
func BenchmarkGuardedDial(b *testing.B) {
	guard := newDialGuard(Config{})
	guard.resolveFn = func(context.Context, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("93.184.216.34")}, nil
	}
	dial := guard.dialContext(func(context.Context, string, string) (net.Conn, error) {
		return nil, errBenchmarkDial
	})
	ctx := context.Background()

	b.ReportAllocs()
	for b.Loop() {
		_, err := dial(ctx, "tcp", "example.com:443")
		if !errors.Is(err, errBenchmarkDial) {
			b.Fatalf("guarded dial error = %v", err)
		}
	}
}
