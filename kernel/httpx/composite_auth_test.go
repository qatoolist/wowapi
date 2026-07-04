package httpx_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/authz"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/httpx"
)

type cAuth struct {
	actor authz.Actor
	err   error
}

func (f cAuth) Authenticate(*http.Request) (authz.Actor, error) { return f.actor, f.err }

func TestCompositeAuthenticator(t *testing.T) {
	req := httptest.NewRequest("GET", "/x", nil)
	unauth := kerr.E(kerr.KindUnauthenticated, "unauthenticated", "nope")
	good := authz.Actor{Kind: authz.ActorSystem, System: "apikey:svc", TenantID: uuid.New()}

	// First that succeeds wins, even after a decline.
	a, err := httpx.Composite(cAuth{err: unauth}, cAuth{actor: good}).Authenticate(req)
	if err != nil || a.System != "apikey:svc" {
		t.Fatalf("composite should return the first success, got actor=%+v err=%v", a, err)
	}

	// All decline → unauthenticated.
	if _, err := httpx.Composite(cAuth{err: unauth}, cAuth{err: unauth}).Authenticate(req); kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("all-decline should be unauthenticated, got %v", err)
	}

	// A hard (non-401) fault short-circuits and is not masked as a 401.
	hard := errors.New("key store unreachable")
	if _, err := httpx.Composite(cAuth{err: hard}, cAuth{actor: good}).Authenticate(req); err == nil || kerr.KindOf(err) == kerr.KindUnauthenticated {
		t.Fatalf("a hard fault must short-circuit (not a 401), got %v", err)
	}

	// No authenticators → fail closed.
	if _, err := httpx.Composite().Authenticate(req); kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("empty composite must fail closed, got %v", err)
	}
}
