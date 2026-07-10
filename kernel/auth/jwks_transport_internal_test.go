package auth

import (
	"net/http"
	"testing"
)

// The default JWKS client must not honor an ambient HTTP(S)_PROXY: an
// attacker or misconfiguration setting the proxy env would otherwise route
// auth-critical key fetches through an unvalidated hop. Behavioral proxy tests
// can't prove this against a loopback httptest server (Go's httpproxy excludes
// loopback from proxying by default), so this asserts the transport policy
// directly — it fails if the Proxy=nil line in jwksTransport is removed.
func TestJWKSTransportDisablesProxy(t *testing.T) {
	tr := jwksTransport()
	if tr.Proxy != nil {
		t.Fatal("jwksTransport must set Proxy=nil so JWKS fetches never traverse an ambient proxy")
	}
	if tr == http.DefaultTransport {
		t.Fatal("jwksTransport must clone the default transport, not mutate the shared global")
	}
	// The shared global must stay untouched by our clone+override.
	if def, ok := http.DefaultTransport.(*http.Transport); ok && def.Proxy == nil {
		t.Fatal("http.DefaultTransport.Proxy was clobbered — jwksTransport must operate on a clone")
	}
}

// The default-client path (JWKSConfig.Client == nil) must wire the
// proxy-disabled transport, while an explicitly injected client is respected.
func TestNewJWKSKeySourceDefaultClientHasNoProxy(t *testing.T) {
	ks, err := NewJWKSKeySource(JWKSConfig{JWKSURI: "https://idp.example.test/jwks"})
	if err != nil {
		t.Fatalf("NewJWKSKeySource: %v", err)
	}
	src, ok := ks.(*jwksKeySource)
	if !ok {
		t.Fatalf("unexpected KeySource type %T", ks)
	}
	tr, ok := src.client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("default client transport is %T, want *http.Transport", src.client.Transport)
	}
	if tr.Proxy != nil {
		t.Fatal("default JWKS client transport must have Proxy=nil")
	}
}
