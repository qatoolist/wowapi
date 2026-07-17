package module_test

import (
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/v2/module"
)

// TestValidName pins the module-name grammar (^[a-z][a-z0-9_]{0,63}$) that
// app.Validate enforces at registration.
func TestValidName(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"requests", true},
		{"a", true},
		{"assets_v2", true},
		{"mod_1", true},
		{strings.Repeat("a", 64), true},  // 1 + 63 = max length
		{strings.Repeat("a", 65), false}, // one over
		{"", false},                      // empty
		{"1mod", false},                  // must start with a letter
		{"_mod", false},                  // must start with a letter
		{"Mod", false},                   // no uppercase
		{"my-module", false},             // no hyphen
		{"my.module", false},             // no dot
		{"my module", false},             // no space
	}
	for _, c := range cases {
		if got := module.ValidName(c.name); got != c.want {
			t.Errorf("ValidName(%q) = %v, want %v", c.name, got, c.want)
		}
	}
}
