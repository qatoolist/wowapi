package mfa_test

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/kernel/mfa"
)

func TestFakeSender_RecordsDeliveries(t *testing.T) {
	f := &mfa.FakeSender{}
	if err := f.Send(context.Background(), "+15551234567", "your code is 123456"); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if f.Count() != 1 {
		t.Fatalf("Count() = %d, want 1", f.Count())
	}
	if f.Deliveries[0].Destination != "+15551234567" || f.Deliveries[0].Body != "your code is 123456" {
		t.Errorf("recorded delivery = %+v, unexpected", f.Deliveries[0])
	}
}

func TestFakeSender_ReturnsConfiguredError(t *testing.T) {
	f := &mfa.FakeSender{Err: errors.New("boom")}
	if err := f.Send(context.Background(), "+1", "body"); err == nil {
		t.Fatal("expected configured error")
	}
	if f.Count() != 0 {
		t.Fatalf("Count() = %d after failed send, want 0", f.Count())
	}
}

func TestFakeSender_Reset(t *testing.T) {
	f := &mfa.FakeSender{Err: errors.New("boom")}
	f.Reset()
	if err := f.Send(context.Background(), "+1", "body"); err != nil {
		t.Fatalf("Send after Reset: %v", err)
	}
	if f.Count() != 1 {
		t.Fatalf("Count() after post-reset send = %d, want 1", f.Count())
	}
}

func TestFakeSender_LastCode_ExtractsTrailingDigits(t *testing.T) {
	f := &mfa.FakeSender{}
	_ = f.Send(context.Background(), "+1", "Your verification code is 048213")
	if got := f.LastCode(); got != "048213" {
		t.Errorf("LastCode() = %q, want %q", got, "048213")
	}
}

func TestFakeSender_LastCode_EmptyBeforeAnySend(t *testing.T) {
	f := &mfa.FakeSender{}
	if got := f.LastCode(); got != "" {
		t.Errorf("LastCode() before any Send = %q, want empty", got)
	}
}

func TestFakeSender_LastCode_EmptyBodyYieldsEmptyCode(t *testing.T) {
	f := &mfa.FakeSender{}
	_ = f.Send(context.Background(), "+1", "")
	if got := f.LastCode(); got != "" {
		t.Errorf("LastCode() for empty body = %q, want empty", got)
	}
}

func TestLogSender_Sends_NoError(t *testing.T) {
	var buf strings.Builder
	log := slog.New(slog.NewTextHandler(&buf, nil))
	s := mfa.NewLogSender(log)
	if err := s.Send(context.Background(), "+15551234567", "your code is 999999"); err != nil {
		t.Fatalf("Send: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "your code is 999999") {
		t.Errorf("log output missing body: %q", out)
	}
	if strings.Contains(out, "+15551234567") {
		t.Errorf("log output must redact the destination, got: %q", out)
	}
	if !strings.Contains(out, "4567") {
		t.Errorf("log output should keep the last 4 chars of destination for correlation, got: %q", out)
	}
}

func TestLogSender_RedactsShortDestination(t *testing.T) {
	var buf strings.Builder
	log := slog.New(slog.NewTextHandler(&buf, nil))
	s := mfa.NewLogSender(log)
	if err := s.Send(context.Background(), "ab", "body"); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if !strings.Contains(buf.String(), "****") {
		t.Errorf("short destination should be fully redacted, got: %q", buf.String())
	}
}

// TestSender_InterfaceSatisfiedByAdapters is a compile-time-ish check (run at
// test time) that both adapters satisfy the Sender port.
func TestSender_InterfaceSatisfiedByAdapters(t *testing.T) {
	var _ mfa.Sender = &mfa.FakeSender{}
	_ = mfa.NewLogSender(slog.Default()) // return type is already mfa.Sender
}
