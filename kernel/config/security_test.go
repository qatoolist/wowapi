package config

import (
	"strings"
	"testing"
)

// TestDefaultSecurityIsAPIProfile proves the compiled default is the API
// profile — selecting no security config at all must reproduce today's
// behavior byte-for-byte (backlog B7 risk: browser mode is new surface,
// the API-only default must stay untouched).
func TestDefaultSecurityIsAPIProfile(t *testing.T) {
	sec := DefaultSecurity()
	if sec.Profile != SecurityProfileAPI {
		t.Fatalf("DefaultSecurity().Profile = %q, want %q", sec.Profile, SecurityProfileAPI)
	}
	if err := sec.Validate(); err != nil {
		t.Fatalf("DefaultSecurity() must validate: %v", err)
	}
}

// TestDefaultsIncludesSecurity proves the framework-wide Defaults() wires the
// Security section (not left zero-valued, which would fail Validate on an
// empty Profile).
func TestDefaultsIncludesSecurity(t *testing.T) {
	f := Defaults()
	if f.Security.Profile != SecurityProfileAPI {
		t.Fatalf("Defaults().Security.Profile = %q, want %q", f.Security.Profile, SecurityProfileAPI)
	}
	if err := f.Validate(); err != nil {
		t.Fatalf("Defaults() must validate with Security wired: %v", err)
	}
}

// TestEnforceRouteContractsDefaultsOff pins the FBL-08 compat guarantee
// (RISK-W01-002, "profile-flag first"): route-contract enforcement ships OFF
// by default in both the compiled defaults and the section default, and a
// config enabling it still validates in every environment (it is a hardening
// knob, not an unsafe one — no prod gate).
func TestEnforceRouteContractsDefaultsOff(t *testing.T) {
	if DefaultSecurity().EnforceRouteContracts {
		t.Fatal("DefaultSecurity().EnforceRouteContracts must be false (compat: profile-flag first)")
	}
	if Defaults().Security.EnforceRouteContracts {
		t.Fatal("Defaults().Security.EnforceRouteContracts must be false (compat: profile-flag first)")
	}
	for _, env := range []Env{EnvLocal, EnvDev, EnvStage, EnvProd} {
		f := Defaults()
		f.Environment = env
		f.Security.EnforceRouteContracts = true
		if err := f.Validate(); err != nil {
			t.Errorf("env=%s: enabling enforce_route_contracts must validate, got: %v", env, err)
		}
	}
}

func TestSecurityProfileValidValues(t *testing.T) {
	cases := []struct {
		profile SecurityProfile
		valid   bool
	}{
		{SecurityProfileAPI, true},
		{SecurityProfileBrowser, true},
		{"", false},
		{"bogus", false},
	}
	for _, c := range cases {
		if got := c.profile.Valid(); got != c.valid {
			t.Errorf("SecurityProfile(%q).Valid() = %v, want %v", c.profile, got, c.valid)
		}
	}
}

func TestSecurityValidateRejectsUnknownProfile(t *testing.T) {
	sec := DefaultSecurity()
	sec.Profile = "bogus"
	err := sec.Validate()
	if err == nil {
		t.Fatal("unknown profile must fail validation")
	}
	if !strings.Contains(err.Error(), "profile") {
		t.Errorf("error should mention profile: %v", err)
	}
}

// TestSecurityValidateAPIProfileIgnoresBrowserSettings proves the API profile
// (the default) never requires CSRF/cookie settings — those fields are simply
// inert when the profile is API, matching "CSRF-free by contract".
func TestSecurityValidateAPIProfileIgnoresBrowserSettings(t *testing.T) {
	sec := Security{Profile: SecurityProfileAPI}
	if err := sec.Validate(); err != nil {
		t.Fatalf("API profile with zero-valued CSRF/Cookie must validate: %v", err)
	}
}

// TestSecurityValidateBrowserProfileRequiresCoherentSettings is the "config
// validate rejects incoherent combos" acceptance criterion: a browser profile
// with missing CSRF cookie/header names, or an invalid SameSite, must fail.
func TestSecurityValidateBrowserProfileRequiresCoherentSettings(t *testing.T) {
	tests := []struct {
		name string
		mut  func(*Security)
	}{
		{"empty CSRF cookie name", func(s *Security) { s.CSRF.CookieName = "" }},
		{"empty CSRF header name", func(s *Security) { s.CSRF.HeaderName = "" }},
		{"invalid SameSite", func(s *Security) { s.Cookie.SameSite = "bogus" }},
		{"SameSite=none without Secure", func(s *Security) {
			s.Cookie.SameSite = "none"
			s.Cookie.Secure = false
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sec := DefaultSecurity()
			sec.Profile = SecurityProfileBrowser
			sec.CSRF = CSRF{CookieName: "csrf_token", HeaderName: "X-CSRF-Token"}
			sec.Cookie = CookieDefaults{SameSite: "lax", Secure: true}
			tt.mut(&sec)
			if err := sec.Validate(); err == nil {
				t.Fatalf("%s: expected validation error, got none", tt.name)
			}
		})
	}
}

func TestSecurityValidateBrowserProfileWithGoodSettingsPasses(t *testing.T) {
	sec := Security{
		Profile: SecurityProfileBrowser,
		CSRF:    CSRF{CookieName: "csrf_token", HeaderName: "X-CSRF-Token"},
		Cookie:  CookieDefaults{SameSite: "lax", Secure: true},
	}
	if err := sec.Validate(); err != nil {
		t.Fatalf("well-formed browser profile must validate: %v", err)
	}
}

// TestSecurityValidateSameSiteIsCaseInsensitive proves an uppercase/mixed-
// case SameSite value (e.g. hand-edited YAML) is normalized before the
// strict|lax|none check, not rejected on casing alone.
func TestSecurityValidateSameSiteIsCaseInsensitive(t *testing.T) {
	sec := Security{
		Profile: SecurityProfileBrowser,
		CSRF:    CSRF{CookieName: "csrf_token", HeaderName: "X-CSRF-Token"},
		Cookie:  CookieDefaults{SameSite: "LAX", Secure: true},
	}
	if err := sec.Validate(); err != nil {
		t.Fatalf("SameSite=LAX (uppercase) must validate: %v", err)
	}
}

func TestSecurityValidateSameSiteNoneWithSecureIsAllowed(t *testing.T) {
	sec := Security{
		Profile: SecurityProfileBrowser,
		CSRF:    CSRF{CookieName: "csrf_token", HeaderName: "X-CSRF-Token"},
		Cookie:  CookieDefaults{SameSite: "none", Secure: true},
	}
	if err := sec.Validate(); err != nil {
		t.Fatalf("SameSite=none with Secure=true must validate: %v", err)
	}
}

// TestFrameworkValidateCatchesMisconfiguredBrowserProfile is the framework-level
// acceptance test: an incoherent browser profile embedded in a full Framework
// must be caught by Framework.Validate(), not just Security.Validate() in
// isolation — proving config.Framework actually wires the Security section
// into the boot-time gate (`wowapi config validate`).
func TestFrameworkValidateCatchesMisconfiguredBrowserProfile(t *testing.T) {
	f := Defaults()
	f.Security.Profile = SecurityProfileBrowser
	f.Security.CSRF.CookieName = "" // incoherent: browser profile, no CSRF cookie name
	err := f.Validate()
	if err == nil {
		t.Fatal("Framework.Validate() must reject a misconfigured browser security profile")
	}
	if !strings.Contains(err.Error(), "csrf") {
		t.Errorf("error should mention csrf: %v", err)
	}
}
