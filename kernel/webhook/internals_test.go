package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/outbox"
)

// TestTruncate covers both branches of truncate: the short-circuit return when
// len <= max, and the actual cut when the string is longer than max.
func TestTruncate(t *testing.T) {
	if got := truncate("short"); got != "short" {
		t.Fatalf("short string should pass through unchanged, got %q", got)
	}
	long := strings.Repeat("x", 600)
	got := truncate(long)
	if len(got) != 500 {
		t.Fatalf("want truncated length 500, got %d", len(got))
	}
	if got != strings.Repeat("x", 500) {
		t.Fatalf("truncate produced unexpected content")
	}
	// Exactly at the boundary is not truncated.
	exact := strings.Repeat("y", 500)
	if truncate(exact) != exact {
		t.Fatal("string of length == max must not be truncated")
	}
	// A multibyte rune straddling the cut must not yield invalid UTF-8: the cut
	// backs up to the rune boundary rather than splitting the rune.
	multibyte := strings.Repeat("a", 499) + "世" // '世' is 3 bytes at offsets 499..501
	mb := truncate(multibyte)
	if !utf8.ValidString(mb) {
		t.Fatalf("truncate produced invalid UTF-8: %q", mb)
	}
	if mb != strings.Repeat("a", 499) {
		t.Fatalf("truncate should drop the straddling rune, got %q (len %d)", mb, len(mb))
	}
}

// TestBackoff covers every case of the exponential backoff schedule, including
// the default (attempt >= 5) arm.
func TestBackoff(t *testing.T) {
	cases := map[int]time.Duration{
		1: time.Second,
		2: 5 * time.Second,
		3: 30 * time.Second,
		4: 2 * time.Minute,
		5: 5 * time.Minute,
		9: 5 * time.Minute, // default arm
	}
	for attempt, want := range cases {
		if got := backoff(attempt); got != want {
			t.Fatalf("backoff(%d) = %v, want %v", attempt, got, want)
		}
	}
}

// TestDedupExtID covers both arms: the provider-supplied external id, and the
// synthetic sha256-of-body id used when the provider omits an id.
func TestDedupExtID(t *testing.T) {
	// Provider-supplied id takes precedence.
	in := InboundIn{ExternalEventID: "evt-42", RawBody: []byte(`{"a":1}`)}
	if got := dedupExtID(in); got != "evt-42" {
		t.Fatalf("want provider id evt-42, got %q", got)
	}

	// Id-less: synthesized from the raw body as "sha256:<hex>".
	body := []byte(`{"b":2}`)
	idless := InboundIn{RawBody: body}
	sum := sha256.Sum256(body)
	want := "sha256:" + hex.EncodeToString(sum[:])
	if got := dedupExtID(idless); got != want {
		t.Fatalf("synthetic id mismatch\n got  %s\n want %s", got, want)
	}

	// Same body → same synthetic id (stable/deterministic).
	if dedupExtID(idless) != dedupExtID(InboundIn{RawBody: []byte(`{"b":2}`)}) {
		t.Fatal("synthetic dedup id is not stable for identical bodies")
	}
	// Different body → different id.
	if dedupExtID(idless) == dedupExtID(InboundIn{RawBody: []byte(`{"b":3}`)}) {
		t.Fatal("synthetic dedup id collided for different bodies")
	}
}

// TestSignPayload proves signPayload computes HMAC-SHA256 over "<ts>.<body>".
func TestSignPayload(t *testing.T) {
	secret := "s3cr3t"
	ts := "1720000000"
	body := []byte(`{"k":"v"}`)

	got := signPayload(secret, ts, body)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "."))
	mac.Write(body)
	want := hex.EncodeToString(mac.Sum(nil))
	if got != want {
		t.Fatalf("signPayload mismatch\n got  %s\n want %s", got, want)
	}
	// Timestamp is authenticated: altering it changes the signature.
	if signPayload(secret, ts+"1", body) == got {
		t.Fatal("signature unchanged when timestamp altered — timestamp not covered")
	}
}

// TestMarshalOutboundBody proves the outbound envelope carries id/type/tenant_id/
// payload and round-trips as JSON.
func TestMarshalOutboundBody(t *testing.T) {
	tenantID := uuid.New()
	ev := outbox.Event{
		ID:      uuid.New(),
		Type:    "order.created",
		Payload: json.RawMessage(`{"order_id":"abc"}`),
	}
	raw, err := marshalOutboundBody(ev, tenantID)
	if err != nil {
		t.Fatalf("marshalOutboundBody: %v", err)
	}
	var env struct {
		ID       uuid.UUID       `json:"id"`
		Type     string          `json:"type"`
		TenantID uuid.UUID       `json:"tenant_id"`
		Payload  json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if env.ID != ev.ID || env.Type != ev.Type || env.TenantID != tenantID {
		t.Fatalf("envelope fields mismatch: %+v", env)
	}
	if string(env.Payload) != `{"order_id":"abc"}` {
		t.Fatalf("payload mismatch: %s", env.Payload)
	}
}
