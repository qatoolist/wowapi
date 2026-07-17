package config_test

import (
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/config"
)

func TestRecordAllowlistChangeEmitsRecordWhenHostsChange(t *testing.T) {
	before := config.WebhookOutbound{AllowedHosts: []string{"old.example.com"}}
	after := config.WebhookOutbound{AllowedHosts: []string{"new.example.com"}}

	var got *config.AllowlistChange
	config.RecordAllowlistChange(before, after, func(c config.AllowlistChange) {
		got = &c
	})
	if got == nil {
		t.Fatal("expected an allowlist change record, got nil")
	}
	if got.Action != "webhook.outbound.allowlist_changed" {
		t.Fatalf("action = %q, want %q", got.Action, "webhook.outbound.allowlist_changed")
	}
	if len(got.OldHosts) != 1 || got.OldHosts[0] != "old.example.com" {
		t.Fatalf("old_hosts = %v, want [old.example.com]", got.OldHosts)
	}
	if len(got.NewHosts) != 1 || got.NewHosts[0] != "new.example.com" {
		t.Fatalf("new_hosts = %v, want [new.example.com]", got.NewHosts)
	}
}

func TestRecordAllowlistChangeEmitsRecordWhenCIDRsChange(t *testing.T) {
	before := config.WebhookOutbound{AllowedCIDRs: []string{"10.0.0.0/8"}}
	after := config.WebhookOutbound{AllowedCIDRs: []string{"10.0.0.0/16"}}

	var got *config.AllowlistChange
	config.RecordAllowlistChange(before, after, func(c config.AllowlistChange) {
		got = &c
	})
	if got == nil {
		t.Fatal("expected an allowlist change record, got nil")
	}
	if len(got.OldCIDRs) != 1 || got.OldCIDRs[0] != "10.0.0.0/8" {
		t.Fatalf("old_cidrs = %v, want [10.0.0.0/8]", got.OldCIDRs)
	}
	if len(got.NewCIDRs) != 1 || got.NewCIDRs[0] != "10.0.0.0/16" {
		t.Fatalf("new_cidrs = %v, want [10.0.0.0/16]", got.NewCIDRs)
	}
}

func TestRecordAllowlistChangeNoOpWhenUnchanged(t *testing.T) {
	before := config.WebhookOutbound{
		AllowedHosts: []string{"host.example.com"},
		AllowedCIDRs: []string{"10.0.0.0/8"},
	}
	after := before

	called := false
	config.RecordAllowlistChange(before, after, func(_ config.AllowlistChange) {
		called = true
	})
	if called {
		t.Fatal("expected no record when allowlist is unchanged")
	}
}

func TestRecordAllowlistChangeNoOpWithNilRecorder(t *testing.T) {
	before := config.WebhookOutbound{AllowedHosts: []string{"old.example.com"}}
	after := config.WebhookOutbound{AllowedHosts: []string{"new.example.com"}}
	// nil recorder must not panic.
	config.RecordAllowlistChange(before, after, nil)
}
