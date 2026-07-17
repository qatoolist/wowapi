package filtering_test

import (
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/filtering"
)

func sortAllow() filtering.SortAllowlist {
	return filtering.SortAllowlist{
		"created_at": {Col: "created_at"},
		"id":         {Col: "id"},
		"name":       {Col: "full_name"}, // client key differs from physical column
	}
}

func TestParseSortMultipleKeys(t *testing.T) {
	s, err := filtering.ParseSort("created_at:desc,id:asc", sortAllow())
	if err != nil {
		t.Fatalf("ParseSort: %v", err)
	}
	want := "ORDER BY created_at DESC, id ASC"
	if got := s.SQL(); got != want {
		t.Errorf("SQL: got %q, want %q", got, want)
	}
}

func TestParseSortDefaultDirAsc(t *testing.T) {
	s, err := filtering.ParseSort("name", sortAllow())
	if err != nil {
		t.Fatal(err)
	}
	if got := s.SQL(); got != "ORDER BY full_name ASC" {
		t.Errorf("SQL: got %q", got)
	}
}

func TestParseSortUnknownKey(t *testing.T) {
	_, err := filtering.ParseSort("password:asc", sortAllow())
	if err == nil {
		t.Fatal("expected error")
	}
	if k := errors.KindOf(err); k != errors.KindValidation {
		t.Errorf("expected KindValidation, got %v", k)
	}
}

func TestParseSortBadDirection(t *testing.T) {
	_, err := filtering.ParseSort("id:sideways", sortAllow())
	if err == nil {
		t.Fatal("expected error")
	}
	if k := errors.KindOf(err); k != errors.KindValidation {
		t.Errorf("expected KindValidation, got %v", k)
	}
}

func TestParseSortEmpty(t *testing.T) {
	s, err := filtering.ParseSort("", sortAllow())
	if err != nil {
		t.Fatal(err)
	}
	if !s.IsEmpty() {
		t.Error("expected empty sort")
	}
	if got := s.SQL(); got != "" {
		t.Errorf("empty sort SQL: got %q, want \"\"", got)
	}
}
