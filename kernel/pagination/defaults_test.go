package pagination

import (
	"testing"

	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
)

// F-08 regression (adversarial-framework-review-2026-07-17): the parsed limit
// is documented as positive and clamped to [1, MaxPerPage]; zero or negative
// Defaults.PerPage previously flowed through Parse unchanged (Limit=0 / -5,
// err=nil), letting a misconfigured caller run SQL with LIMIT 0 or a negative
// limit. Invalid configuration must fail loudly, never pass silently.
func TestParseDefaultsValidation(t *testing.T) {
	cases := []struct {
		name      string
		perPage   string
		def       Defaults
		wantLimit int
		wantErr   bool
		wantKind  kerr.Kind
	}{
		{name: "zero default, empty request", perPage: "", def: Defaults{PerPage: 0, MaxPerPage: 100}, wantErr: true, wantKind: kerr.KindInternal},
		{name: "negative default, empty request", perPage: "", def: Defaults{PerPage: -5, MaxPerPage: 100}, wantErr: true, wantKind: kerr.KindInternal},
		{name: "zero default, explicit zero request", perPage: "0", def: Defaults{PerPage: 0, MaxPerPage: 100}, wantErr: true, wantKind: kerr.KindInternal},
		{name: "negative default only matters when used", perPage: "7", def: Defaults{PerPage: -5, MaxPerPage: 100}, wantLimit: 7},
		{name: "valid minimum default", perPage: "", def: Defaults{PerPage: 1, MaxPerPage: 100}, wantLimit: 1},
		{name: "default above max clamps", perPage: "", def: Defaults{PerPage: 250, MaxPerPage: 100}, wantLimit: 100},
		{name: "request at max", perPage: "100", def: Defaults{PerPage: 25, MaxPerPage: 100}, wantLimit: 100},
		{name: "request above max clamps", perPage: "101", def: Defaults{PerPage: 25, MaxPerPage: 100}, wantLimit: 100},
		{name: "no upper bound", perPage: "1000", def: Defaults{PerPage: 25}, wantLimit: 1000},
		{name: "negative request stays client error", perPage: "-1", def: Defaults{PerPage: 25, MaxPerPage: 100}, wantErr: true, wantKind: kerr.KindValidation},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := Parse(tc.perPage, "", tc.def)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Parse(%q, %+v) = %+v, want error", tc.perPage, tc.def, req)
				}
				if kerr.KindOf(err) != tc.wantKind {
					t.Fatalf("Parse(%q, %+v) error kind = %v, want %v (err=%v)", tc.perPage, tc.def, kerr.KindOf(err), tc.wantKind, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q, %+v) unexpected error: %v", tc.perPage, tc.def, err)
			}
			if req.Limit != tc.wantLimit {
				t.Fatalf("Parse(%q, %+v).Limit = %d, want %d", tc.perPage, tc.def, req.Limit, tc.wantLimit)
			}
			if req.Limit < 1 {
				t.Fatalf("documented invariant violated: Limit %d < 1", req.Limit)
			}
		})
	}
}
