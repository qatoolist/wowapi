package notify

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
)

// TestBackoffClamps exercises the two guard branches of backoff (attempt below 1
// and attempt above the schedule length) that the production retry path never
// reaches: with maxAttempts == 3, backoff is only ever called with 1 or 2.
func TestBackoffClamps(t *testing.T) {
	if got := backoff(0); got != backoffSchedule[0] {
		t.Errorf("backoff(0) = %v, want %v (clamped to first)", got, backoffSchedule[0])
	}
	if got := backoff(-7); got != backoffSchedule[0] {
		t.Errorf("backoff(-7) = %v, want %v (clamped to first)", got, backoffSchedule[0])
	}
	if got := backoff(1); got != backoffSchedule[0] {
		t.Errorf("backoff(1) = %v, want %v", got, backoffSchedule[0])
	}
	if got := backoff(2); got != backoffSchedule[1] {
		t.Errorf("backoff(2) = %v, want %v", got, backoffSchedule[1])
	}
	last := backoffSchedule[len(backoffSchedule)-1]
	if got := backoff(len(backoffSchedule)); got != last {
		t.Errorf("backoff(len) = %v, want %v", got, last)
	}
	if got := backoff(99); got != last {
		t.Errorf("backoff(99) = %v, want %v (clamped to last)", got, last)
	}
	// Monotonic non-decreasing invariant (transient outage must not burn all
	// attempts in seconds).
	for i := 1; i < len(backoffSchedule); i++ {
		if backoffSchedule[i] < backoffSchedule[i-1] {
			t.Fatalf("backoff schedule not monotonic at %d: %v < %v", i, backoffSchedule[i], backoffSchedule[i-1])
		}
	}
}

// TestActorFromCtx covers both branches: a context with an actor id returns it;
// a bare context yields uuid.Nil (the system/anonymous actor).
func TestActorFromCtx(t *testing.T) {
	if got := actorFromCtx(context.Background()); got != uuid.Nil {
		t.Errorf("actorFromCtx(empty) = %v, want uuid.Nil", got)
	}
	want := uuid.New()
	ctx := database.WithActorID(context.Background(), want)
	if got := actorFromCtx(ctx); got != want {
		t.Errorf("actorFromCtx(withActor) = %v, want %v", got, want)
	}
}

// TestLocaleFallbackChains covers every branch of localeFallback: the empty/en
// shortcut, a region-qualified locale (hi-IN → hi → en), and a bare non-en
// locale (fr → en).
func TestLocaleFallbackChains(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", []string{"en"}},
		{"en", []string{"en"}},
		{"hi-IN", []string{"hi-IN", "hi", "en"}},
		{"fr", []string{"fr", "en"}},
		{"pt-BR", []string{"pt-BR", "pt", "en"}},
	}
	for _, tc := range cases {
		got := localeFallback(tc.in)
		if len(got) != len(tc.want) {
			t.Fatalf("localeFallback(%q) = %v, want %v", tc.in, got, tc.want)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("localeFallback(%q) = %v, want %v", tc.in, got, tc.want)
			}
		}
	}
}

// TestExtractTemplateVarsAllNodeKinds drives extractTemplateVars through every
// parse-node branch of its walk: if/else, range, with, template invocation (with
// and without a pipe argument), and a chained field access on a parenthesized
// pipe (ChainNode → PipeNode). The if-without-else and template-without-pipe
// arms also exercise the nil guards in walk and walkPipe.
func TestExtractTemplateVarsAllNodeKinds(t *testing.T) {
	body := `{{if .Cond}}{{.A}}{{else}}{{.B}}{{end}}` +
		`{{if .E}}{{.F}}{{end}}` + // no else → walk(nil) ElseList branch
		`{{range .Items}}{{.C}}{{end}}` +
		`{{with .Ctx}}{{.D}}{{end}}` +
		`{{template "sub" .Data}}` + // TemplateNode with a pipe arg
		`{{template "bare"}}` + // TemplateNode with nil pipe → walkPipe(nil)
		`{{(.Foo).Bar}}` // ChainNode over a PipeNode

	vars, err := extractTemplateVars(body)
	if err != nil {
		t.Fatalf("extractTemplateVars: %v", err)
	}
	seen := map[string]bool{}
	for _, v := range vars {
		seen[v] = true
	}
	for _, want := range []string{"Cond", "A", "B", "E", "F", "Items", "C", "Ctx", "D", "Data", "Foo"} {
		if !seen[want] {
			t.Errorf("extractTemplateVars missing top-level var %q; got %v", want, vars)
		}
	}
}

// TestExtractTemplateVarsEmptyBody covers the nil-tree short-circuit: a
// comment-only body parses to an empty/absent root and yields no vars.
func TestExtractTemplateVarsEmptyBody(t *testing.T) {
	vars, err := extractTemplateVars(`{{/* just a comment */}}`)
	if err != nil {
		t.Fatalf("extractTemplateVars(comment): %v", err)
	}
	if len(vars) != 0 {
		t.Fatalf("comment-only body should reference no vars, got %v", vars)
	}
}

// TestSenderForBuiltIn confirms senderFor reports the built-in in-app sender as
// present and an unwired channel as absent (the CA-15 "no silent success" gate).
func TestSenderForBuiltIn(t *testing.T) {
	svc := New(NewRegistry(), stubIDGen{})
	if _, ok := svc.senderFor(ChannelInApp); !ok {
		t.Error("in-app sender must be registered by default")
	}
	if _, ok := svc.senderFor(ChannelSMS); ok {
		t.Error("sms sender must NOT be present without explicit registration")
	}
}

// stubIDGen is a minimal model.IDGen for constructing a Service in unit tests
// that never touch the database.
type stubIDGen struct{}

func (stubIDGen) New() uuid.UUID { return uuid.New() }
