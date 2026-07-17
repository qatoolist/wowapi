package filtering_test

import (
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/filtering"
)

func testAllow() filtering.Allowlist {
	return filtering.Allowlist{
		"status": {Col: "status", Ops: []filtering.Op{filtering.OpEq, filtering.OpIn}},
		"age":    {Col: "age", Ops: []filtering.Op{filtering.OpGt, filtering.OpGte, filtering.OpLt, filtering.OpLte}},
		"name":   {Col: "full_name", Ops: []filtering.Op{filtering.OpEq, filtering.OpLike}},
	}
}

// TestInjectionValueBecomesPlaceholder is THE security test: a hostile value
// must land in args as a $N placeholder and never appear in the SQL text.
func TestInjectionValueBecomesPlaceholder(t *testing.T) {
	payload := "active'; DROP TABLE users;--"
	set, err := filtering.Parse(map[string][]string{
		"status": {"eq:" + payload},
	}, testAllow())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	sql, args, next := set.SQL(1)

	if strings.Contains(sql, payload) {
		t.Fatalf("payload leaked into SQL: %q", sql)
	}
	if strings.Contains(sql, "DROP") {
		t.Fatalf("SQL fragment contains injected keyword: %q", sql)
	}
	if sql != "status = $1" {
		t.Errorf("SQL: got %q, want %q", sql, "status = $1")
	}
	if len(args) != 1 || args[0] != payload {
		t.Errorf("args: got %v, want [%q]", args, payload)
	}
	if next != 2 {
		t.Errorf("nextArg: got %d, want 2", next)
	}
}

func TestUnknownFieldRejected(t *testing.T) {
	_, err := filtering.Parse(map[string][]string{"secret": {"eq:x"}}, testAllow())
	assertValidation(t, err)
}

func TestDisallowedOpRejected(t *testing.T) {
	// "status" permits eq,in but not like.
	_, err := filtering.Parse(map[string][]string{"status": {"like:a%"}}, testAllow())
	assertValidation(t, err)
}

func TestUnknownOpRejected(t *testing.T) {
	_, err := filtering.Parse(map[string][]string{"status": {"bogus:x"}}, testAllow())
	assertValidation(t, err)
}

func TestMissingOpPrefixRejected(t *testing.T) {
	_, err := filtering.Parse(map[string][]string{"status": {"active"}}, testAllow())
	assertValidation(t, err)
}

func TestInMultipleValues(t *testing.T) {
	set, err := filtering.Parse(map[string][]string{
		"status": {"in:active,pending,closed"},
	}, testAllow())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	sql, args, next := set.SQL(1)
	if sql != "status IN ($1, $2, $3)" {
		t.Errorf("SQL: got %q", sql)
	}
	if len(args) != 3 || args[0] != "active" || args[1] != "pending" || args[2] != "closed" {
		t.Errorf("args: got %v", args)
	}
	if next != 4 {
		t.Errorf("nextArg: got %d, want 4", next)
	}
}

func TestMultipleConditionsPlaceholderNumbering(t *testing.T) {
	// Fields are AND-combined in sorted order: age, status.
	set, err := filtering.Parse(map[string][]string{
		"status": {"eq:active"},
		"age":    {"gte:18"},
	}, testAllow())
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	sql, args, next := set.SQL(5) // start numbering at $5
	want := "age >= $5 AND status = $6"
	if sql != want {
		t.Errorf("SQL: got %q, want %q", sql, want)
	}
	if len(args) != 2 || args[0] != "18" || args[1] != "active" {
		t.Errorf("args: got %v", args)
	}
	if next != 7 {
		t.Errorf("nextArg: got %d, want 7", next)
	}
}

func TestWhereClause(t *testing.T) {
	set, err := filtering.Parse(map[string][]string{"status": {"eq:active"}}, testAllow())
	if err != nil {
		t.Fatal(err)
	}
	sql, args, next := set.Where(1)
	if sql != "WHERE status = $1" {
		t.Errorf("Where: got %q", sql)
	}
	if len(args) != 1 || next != 2 {
		t.Errorf("args=%v next=%d", args, next)
	}
}

func TestEmptySet(t *testing.T) {
	var empty filtering.Set
	sql, args, next := empty.SQL(3)
	if sql != "" || args != nil || next != 3 {
		t.Errorf("empty SQL: got (%q, %v, %d)", sql, args, next)
	}
	wsql, wargs, wnext := empty.Where(3)
	if wsql != "" || wargs != nil || wnext != 3 {
		t.Errorf("empty Where: got (%q, %v, %d)", wsql, wargs, wnext)
	}
	if !empty.IsEmpty() {
		t.Error("expected IsEmpty")
	}
}

func TestParseEmptyInputEmptySet(t *testing.T) {
	set, err := filtering.Parse(nil, testAllow())
	if err != nil {
		t.Fatal(err)
	}
	if !set.IsEmpty() {
		t.Error("nil input should yield empty set")
	}
}

func assertValidation(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
	if k := errors.KindOf(err); k != errors.KindValidation {
		t.Errorf("expected KindValidation, got %v", k)
	}
}
