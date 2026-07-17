package mfa_test

import (
	"testing"
	"time"

	"github.com/qatoolist/wowapi/v2/foundation/mfa"
)

func TestChallengePolicy_Defaults(t *testing.T) {
	p := mfa.ChallengePolicy{}
	if p.TTLOrDefault() != mfa.DefaultChallengeTTL {
		t.Errorf("TTLOrDefault() = %v, want %v", p.TTLOrDefault(), mfa.DefaultChallengeTTL)
	}
	if p.MaxAttemptsOrDefault() != mfa.DefaultChallengeMaxAttempts {
		t.Errorf("MaxAttemptsOrDefault() = %d, want %d", p.MaxAttemptsOrDefault(), mfa.DefaultChallengeMaxAttempts)
	}
}

func TestChallengePolicy_ExpiresAt(t *testing.T) {
	p := mfa.ChallengePolicy{TTL: 5 * time.Minute}
	issued := time.Unix(1_700_000_000, 0).UTC()
	want := issued.Add(5 * time.Minute)
	if got := p.ExpiresAt(issued); !got.Equal(want) {
		t.Errorf("ExpiresAt = %v, want %v", got, want)
	}
}

func TestChallengeState_Expired(t *testing.T) {
	p := mfa.ChallengePolicy{TTL: time.Minute}
	issued := time.Unix(1_700_000_000, 0).UTC()
	st := mfa.ChallengeState{IssuedAt: issued}

	if p.Expired(st, issued) {
		t.Error("challenge reported expired at issuance instant")
	}
	if !p.Expired(st, issued.Add(2*time.Minute)) {
		t.Error("challenge not reported expired after TTL elapsed")
	}
	if p.Expired(st, issued.Add(30*time.Second)) {
		t.Error("challenge reported expired before TTL elapsed")
	}
}

func TestChallengeState_AttemptsExhausted(t *testing.T) {
	p := mfa.ChallengePolicy{MaxAttempts: 3}
	if p.AttemptsExhausted(mfa.ChallengeState{Attempts: 2}) {
		t.Error("2 attempts against a limit of 3 should not be exhausted")
	}
	if !p.AttemptsExhausted(mfa.ChallengeState{Attempts: 3}) {
		t.Error("3 attempts against a limit of 3 should be exhausted")
	}
	if !p.AttemptsExhausted(mfa.ChallengeState{Attempts: 4}) {
		t.Error("attempts exceeding the limit should be exhausted")
	}
}

func TestChallengeState_Consumed(t *testing.T) {
	st := mfa.ChallengeState{}
	if st.Consumed {
		t.Error("zero-value ChallengeState must not be pre-consumed")
	}
}

func TestChallengePolicy_Evaluate(t *testing.T) {
	p := mfa.ChallengePolicy{TTL: time.Minute, MaxAttempts: 3}
	issued := time.Unix(1_700_000_000, 0).UTC()

	cases := []struct {
		name string
		st   mfa.ChallengeState
		now  time.Time
		want mfa.ChallengeStatus
	}{
		{"fresh", mfa.ChallengeState{IssuedAt: issued}, issued, mfa.ChallengeOK},
		{"expired", mfa.ChallengeState{IssuedAt: issued}, issued.Add(2 * time.Minute), mfa.ChallengeExpired},
		{"exhausted", mfa.ChallengeState{IssuedAt: issued, Attempts: 3}, issued, mfa.ChallengeAttemptsExhausted},
		{"consumed", mfa.ChallengeState{IssuedAt: issued, Consumed: true}, issued, mfa.ChallengeConsumed},
		{
			"expired takes priority over exhausted",
			mfa.ChallengeState{IssuedAt: issued, Attempts: 5},
			issued.Add(2 * time.Minute),
			mfa.ChallengeExpired,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := p.Evaluate(c.st, c.now); got != c.want {
				t.Errorf("Evaluate(%+v, %v) = %v, want %v", c.st, c.now, got, c.want)
			}
		})
	}
}

func TestChallengeStatus_OK(t *testing.T) {
	if !mfa.ChallengeOK.OK() {
		t.Error("ChallengeOK.OK() = false, want true")
	}
	for _, s := range []mfa.ChallengeStatus{mfa.ChallengeExpired, mfa.ChallengeAttemptsExhausted, mfa.ChallengeConsumed} {
		if s.OK() {
			t.Errorf("%v.OK() = true, want false", s)
		}
	}
}

func TestChallengeStatus_String(t *testing.T) {
	cases := []struct {
		s    mfa.ChallengeStatus
		want string
	}{
		{mfa.ChallengeOK, "ok"},
		{mfa.ChallengeExpired, "expired"},
		{mfa.ChallengeAttemptsExhausted, "attempts_exhausted"},
		{mfa.ChallengeConsumed, "consumed"},
		{mfa.ChallengeStatus(99), "unknown"},
	}
	for _, c := range cases {
		if got := c.s.String(); got != c.want {
			t.Errorf("ChallengeStatus(%d).String() = %q, want %q", c.s, got, c.want)
		}
	}
}
