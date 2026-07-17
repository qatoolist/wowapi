package auth_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/auth"
	"github.com/qatoolist/wowapi/v2/testkit"
)

var benchmarkClaims auth.Claims

// BenchmarkTokenVerify measures complete RS256 parsing, kid lookup, signature
// verification, and registered-claim validation for a valid bearer token.
func BenchmarkTokenVerify(b *testing.B) {
	issuer := testkit.NewTokenIssuer()
	verifier := auth.NewVerifier(issuer.KeySource(), auth.Config{
		Issuer: "wowapi-test", Audience: "wowapi",
	})
	token := issuer.Issue("idp|benchmark", uuid.New(), uuid.New())
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		claims, err := verifier.Verify(ctx, token)
		if err != nil {
			b.Fatalf("verify token: %v", err)
		}
		benchmarkClaims = claims
	}
}
