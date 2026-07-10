package config

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Privileged is the product config section that widens a module's
// kernel/privileged ownership beyond its own name-prefixed keys (backlog B10;
// evidence app/context.go — module.Context.Privileged() used to construct
// privileged.New with an always-empty privileged.Config, so a product could
// only reach a cross-namespace or kernel-owned relationship type / rule key by
// building its own privileged.Services outside the standard module.Context
// path). It maps a module name to the extra relationship types / rule keys
// that module is allowed to manage via mc.Privileged().
//
// SECURITY (fail closed): every entry must be a concrete, fully-spelled
// relationship-type or rule-key string. Wildcards, glob syntax, and empty
// entries are REJECTED at boot by Validate — see the doc comment there for
// the exact rule. There is no "allow everything" escape hatch; each grant
// must be enumerated explicitly, one string per key, reviewable in a diff.
//
// The zero value (nil map, or a module absent from it) changes NOTHING: that
// module keeps exactly today's prefix-only ownership.
type Privileged map[string]PrivilegedGrant

// PrivilegedGrant is one module's allow-list. Both fields feed directly into
// privileged.Config{AllowRelTypes, AllowRuleKeys} (kernel/privileged) — see
// app/context.go's moduleContext.Privileged().
type PrivilegedGrant struct {
	// AllowRelTypes lists relationship-type keys (relationship_types.key) this
	// module may Grant/Revoke beyond its own "<module>." prefix, e.g. a kernel
	// "core.owner_of" type a product module is sanctioned to grant.
	AllowRelTypes []string `conf:"allow_rel_types" json:"allow_rel_types" doc:"explicit relationship-type keys this module may manage beyond its own prefix; NO wildcards"`
	// AllowRuleKeys lists rule keys (rule_definitions.key) this module may
	// activate tenant versions of, beyond its own "<module>." prefix.
	AllowRuleKeys []string `conf:"allow_rule_keys" json:"allow_rule_keys" doc:"explicit rule keys this module may activate beyond its own prefix; NO wildcards"`
}

// globChars are the characters that make a string a glob/wildcard pattern
// rather than a concrete key. Any entry containing one of these is rejected —
// this is deliberately broader than just "*" so a product cannot smuggle in
// pattern-matching via "?" or bracket classes either.
const globChars = "*?[]"

// Validate enforces the explicit-enumeration rule (fail closed): every
// AllowRelTypes/AllowRuleKeys entry, for every module, must be a non-empty,
// whitespace-free, glob-free concrete string, and every module name (map key)
// must be non-empty. Like Framework.Validate, it collects ALL problems and
// joins them rather than stopping at the first.
func (p Privileged) Validate() error {
	if len(p) == 0 {
		return nil
	}
	var errs []error
	add := func(format string, args ...any) { errs = append(errs, fmt.Errorf(format, args...)) }

	// Deterministic order so repeated boots produce identical error text.
	modules := make([]string, 0, len(p))
	for m := range p {
		modules = append(modules, m)
	}
	sort.Strings(modules)

	for _, m := range modules {
		grant := p[m]
		if strings.TrimSpace(m) == "" {
			add("privileged: empty module name is not allowed")
		}
		for _, key := range grant.AllowRelTypes {
			if err := validPrivilegedKey(key); err != nil {
				add("privileged.%s.allow_rel_types: %w", safeModuleLabel(m), err)
			}
		}
		for _, key := range grant.AllowRuleKeys {
			if err := validPrivilegedKey(key); err != nil {
				add("privileged.%s.allow_rule_keys: %w", safeModuleLabel(m), err)
			}
		}
	}

	return errors.Join(errs...)
}

// safeModuleLabel keeps the module name visible in the joined error even when
// it is itself the empty string being rejected.
func safeModuleLabel(m string) string {
	if m == "" {
		return "\"\""
	}
	return "\"" + m + "\""
}

// validPrivilegedKey enforces one entry: non-empty (after trimming), no
// internal whitespace, and no glob/wildcard characters. Fail closed — an
// ambiguous or pattern-like entry is rejected rather than interpreted.
func validPrivilegedKey(key string) error {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return fmt.Errorf("%q: empty pattern is not allowed — every key must be a concrete relationship-type/rule-key string", key)
	}
	if trimmed != key {
		return fmt.Errorf("%q: leading/trailing whitespace is not allowed", key)
	}
	if strings.ContainsAny(key, globChars) {
		return fmt.Errorf("%q: wildcard/glob syntax is not allowed — enumerate the exact key", key)
	}
	return nil
}
