package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// ParsedFK is a foreign-key definition extracted from migration SQL.
type ParsedFK struct {
	SourceFile    string
	ChildTable    string
	Constraint    string
	ChildColumns  []string
	ParentTable   string
	ParentColumns []string
}

// IsComposite reports whether the FK includes tenant_id among its child columns.
func (f ParsedFK) IsComposite() bool {
	for _, c := range f.ChildColumns {
		if normalizeIdent(c) == "tenant_id" {
			return true
		}
	}
	return false
}

// ParseMigrationDir reads every *.sql file in dir and returns all FK
// definitions found in their +goose Up blocks.
func ParseMigrationDir(dir string) ([]ParsedFK, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("tenantfk: read migration dir %q: %w", dir, err)
	}
	var all []ParsedFK
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		fks, err := ParseMigrationFile(path)
		if err != nil {
			return nil, err
		}
		all = append(all, fks...)
	}
	return all, nil
}

// ParseMigrationFile reads a single migration SQL file and returns its FK
// definitions from the +goose Up block.
func ParseMigrationFile(path string) ([]ParsedFK, error) {
	// #nosec G304 -- paths are validated by the caller (migration-directory CLI args).
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("tenantfk: read %q: %w", path, err)
	}
	return ParseMigrationSQL(path, string(data)), nil
}

// ParseMigrationSQL extracts FK definitions from the +goose Up block of a
// migration SQL body. Down blocks are ignored because they exist only for
// reversibility drills and may legitimately re-create redundant constraints.
func ParseMigrationSQL(source, body string) []ParsedFK {
	up := extractUpBlock(body)
	clean := stripComments(up)
	tokens := tokenize(clean)
	return parseTokens(source, tokens)
}

// extractUpBlock returns the SQL between -- +goose Up and -- +goose Down.
func extractUpBlock(body string) string {
	low := strings.ToLower(body)
	upIdx := strings.Index(low, "-- +goose up")
	if upIdx == -1 {
		return body // no marker; scan whole file
	}
	upStart := upIdx + len("-- +goose up")
	upEnd := len(body)
	if downIdx := strings.Index(low[upStart:], "-- +goose down"); downIdx != -1 {
		upEnd = upStart + downIdx
	}
	return body[upStart:upEnd]
}

// stripComments removes SQL comments so tokenization is not confused.
func stripComments(s string) string {
	// Remove /* ... */ blocks.
	for {
		start := strings.Index(s, "/*")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], "*/")
		if end == -1 {
			s = s[:start]
			break
		}
		s = s[:start] + " " + s[start+end+2:]
	}
	// Remove -- to end of line, but not inside string literals.
	var out strings.Builder
	inQuote := false
	for i := range s {
		c := s[i]
		if c == '\'' {
			inQuote = !inQuote
		}
		if !inQuote && c == '-' && i+1 < len(s) && s[i+1] == '-' {
			for i < len(s) && s[i] != '\n' {
				i++
			}
			continue
		}
		out.WriteByte(c)
	}
	return out.String()
}

// token represents a lexical token from SQL source.
type token struct {
	typ string // ident, str, num, sym, other
	val string
}

// tokenize splits SQL into a simple token stream.
func tokenize(s string) []token {
	var toks []token
	runes := []rune(s)
	for i := 0; i < len(runes); {
		r := runes[i]
		switch {
		case unicode.IsSpace(r):
			i++
		case r == '\'':
			j := i + 1
			for j < len(runes) {
				if runes[j] == '\'' {
					if j+1 < len(runes) && runes[j+1] == '\'' {
						j += 2
						continue
					}
					break
				}
				j++
			}
			if j < len(runes) {
				j++
			}
			toks = append(toks, token{typ: "str", val: string(runes[i:j])})
			i = j
		case r == '"':
			j := i + 1
			for j < len(runes) && runes[j] != '"' {
				j++
			}
			if j < len(runes) {
				j++
			}
			toks = append(toks, token{typ: "ident", val: string(runes[i:j])})
			i = j
		case unicode.IsLetter(r) || r == '_':
			j := i
			for j < len(runes) && (unicode.IsLetter(runes[j]) || unicode.IsDigit(runes[j]) || runes[j] == '_') {
				j++
			}
			toks = append(toks, token{typ: "ident", val: string(runes[i:j])})
			i = j
		case unicode.IsDigit(r):
			j := i
			for j < len(runes) && unicode.IsDigit(runes[j]) {
				j++
			}
			toks = append(toks, token{typ: "num", val: string(runes[i:j])})
			i = j
		case r == '(' || r == ')' || r == ',' || r == ';' || r == '.':
			toks = append(toks, token{typ: "sym", val: string(r)})
			i++
		default:
			j := i
			for j < len(runes) && !unicode.IsSpace(runes[j]) && runes[j] != '\'' && runes[j] != '"' && !unicode.IsLetter(runes[j]) && !unicode.IsDigit(runes[j]) && runes[j] != '(' && runes[j] != ')' && runes[j] != ',' && runes[j] != ';' && runes[j] != '.' {
				j++
			}
			if j == i {
				j++
			}
			toks = append(toks, token{typ: "other", val: string(runes[i:j])})
			i = j
		}
	}
	return toks
}

func identEq(t token, s string) bool {
	return t.typ == "ident" && strings.EqualFold(t.val, s)
}

func normalizeIdent(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\"")
	return strings.ToLower(s)
}

func parseIdents(toks []token, start int) (idents []string, next int) {
	i := start
	if i >= len(toks) || toks[i].val != "(" {
		return nil, start
	}
	i++
	for i < len(toks) {
		if toks[i].val == ")" {
			return idents, i + 1
		}
		if toks[i].typ == "ident" {
			idents = append(idents, normalizeIdent(toks[i].val))
		}
		i++
	}
	return idents, i
}

func parseTokens(source string, toks []token) []ParsedFK {
	var fks []ParsedFK
	for i := 0; i < len(toks); {
		// ALTER TABLE ... ADD CONSTRAINT ... FOREIGN KEY (...) REFERENCES ... (...)
		if identEq(toks[i], "ALTER") && i+1 < len(toks) && identEq(toks[i+1], "TABLE") {
			i += 2
			if i < len(toks) && identEq(toks[i], "IF") {
				i += 4 // IF NOT EXISTS <table>
			}
			if i >= len(toks) {
				break
			}
			childTable := normalizeIdent(toks[i].val)
			i++
			for i < len(toks) && !identEq(toks[i], "ADD") {
				i++
			}
			if i >= len(toks) || !identEq(toks[i], "ADD") {
				continue
			}
			i++ // ADD
			if i < len(toks) && identEq(toks[i], "CONSTRAINT") {
				i += 2 // CONSTRAINT <name>
			}
			if i >= len(toks) || !identEq(toks[i], "FOREIGN") {
				continue
			}
			i += 2 // FOREIGN KEY
			if i >= len(toks) || !identEq(toks[i-1], "KEY") {
				continue
			}
			childCols, next := parseIdents(toks, i)
			if next >= len(toks) || !identEq(toks[next], "REFERENCES") {
				continue
			}
			i = next + 1
			if i < len(toks) && identEq(toks[i], "IF") {
				i += 4 // IF NOT EXISTS <table>
			}
			if i >= len(toks) {
				break
			}
			parentTable := normalizeIdent(toks[i].val)
			i++
			var parentCols []string
			if i < len(toks) && toks[i].val == "(" {
				parentCols, next = parseIdents(toks, i)
				i = next
			}
			fks = append(fks, ParsedFK{
				SourceFile:    source,
				ChildTable:    childTable,
				ParentTable:   parentTable,
				ChildColumns:  childCols,
				ParentColumns: parentCols,
			})
			continue
		}

		// CREATE TABLE ... (... <col> <type> [constraints] REFERENCES <table> [(<cols>)] ...)
		if identEq(toks[i], "CREATE") && i+1 < len(toks) && identEq(toks[i+1], "TABLE") {
			i += 2
			if i < len(toks) && identEq(toks[i], "IF") {
				i += 4 // IF NOT EXISTS <table>
			}
			if i >= len(toks) {
				break
			}
			childTable := normalizeIdent(toks[i].val)
			i++
			if i >= len(toks) || toks[i].val != "(" {
				continue
			}
			i++ // skip (
			for i < len(toks) {
				if toks[i].val == ")" {
					break
				}
				if toks[i].typ != "ident" {
					i++
					continue
				}
				colName := toks[i].val
				i++
				// Skip column type and any constraints until we hit , or ) or REFERENCES.
				for i < len(toks) && toks[i].val != "," && toks[i].val != ")" && !identEq(toks[i], "REFERENCES") {
					i++
				}
				if i < len(toks) && identEq(toks[i], "REFERENCES") {
					i++
					if i < len(toks) && identEq(toks[i], "IF") {
						i += 4
					}
					if i >= len(toks) {
						break
					}
					parentTable := normalizeIdent(toks[i].val)
					i++
					var parentCols []string
					if i < len(toks) && toks[i].val == "(" {
						var next int
						parentCols, next = parseIdents(toks, i)
						i = next
					}
					fks = append(fks, ParsedFK{
						SourceFile:    source,
						ChildTable:    childTable,
						ParentTable:   parentTable,
						ChildColumns:  []string{normalizeIdent(colName)},
						ParentColumns: parentCols,
					})
				}
				if i < len(toks) && toks[i].val == "," {
					i++
				}
			}
			continue
		}

		i++
	}

	// Second pass: fill in constraint names for ALTER TABLE ADD CONSTRAINT forms.
	// We do this by re-scanning and matching FKs to the nearest preceding CONSTRAINT name.
	for idx := range fks {
		if fks[idx].Constraint != "" {
			continue
		}
		// Find the ALTER TABLE statement this FK belongs to and extract the constraint name.
		fks[idx].Constraint = findConstraintName(toks, fks[idx])
	}

	return fks
}

func findConstraintName(toks []token, fk ParsedFK) string {
	// Heuristic: locate the FOREIGN KEY token whose child table and columns match,
	// then step back to the CONSTRAINT name if present.
	for i := 0; i < len(toks); i++ {
		if !identEq(toks[i], "ALTER") || i+1 >= len(toks) || !identEq(toks[i+1], "TABLE") {
			continue
		}
		j := i + 2
		if j < len(toks) && identEq(toks[j], "IF") {
			j += 4
		}
		if j >= len(toks) || normalizeIdent(toks[j].val) != fk.ChildTable {
			continue
		}
		j++
		for j < len(toks) && !identEq(toks[j], "ADD") {
			j++
		}
		if j >= len(toks) {
			continue
		}
		j++ // ADD
		var constraintName string
		if j < len(toks) && identEq(toks[j], "CONSTRAINT") {
			j++
			if j < len(toks) {
				constraintName = normalizeIdent(toks[j].val)
				j++
			}
		}
		if j+1 >= len(toks) || !identEq(toks[j], "FOREIGN") || !identEq(toks[j+1], "KEY") {
			continue
		}
		j += 2
		childCols, next := parseIdents(toks, j)
		if slicesEqual(childCols, fk.ChildColumns) {
			return constraintName
		}
		i = next
	}
	return ""
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Violation is a non-composite tenant FK detected in a migration.
type Violation struct {
	FK     ParsedFK
	Reason string
}

// CheckMigrations parses the migration files/directories in paths and returns
// violations: FKs whose child and parent tables are both tenant-scoped in db
// but whose FK columns are not composite on (tenant_id, ...).
func (s *Scanner) CheckMigrations(ctx context.Context, paths []string) ([]Violation, error) {
	tenantTables, err := s.TenantScopedTables(ctx)
	if err != nil {
		return nil, err
	}

	var all []ParsedFK
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("tenantfk: stat %q: %w", p, err)
		}
		var fks []ParsedFK
		if info.IsDir() {
			fks, err = ParseMigrationDir(p)
		} else {
			fks, err = ParseMigrationFile(p)
		}
		if err != nil {
			return nil, err
		}
		all = append(all, fks...)
	}

	var violations []Violation
	for _, fk := range all {
		if _, ok := tenantTables[fk.ChildTable]; !ok {
			continue // not a tenant-scoped child table
		}
		if _, ok := tenantTables[fk.ParentTable]; !ok {
			continue // parent is not tenant-scoped; tenant agreement does not apply
		}
		if fk.IsComposite() {
			continue // already composite on tenant_id
		}
		violations = append(violations, Violation{
			FK:     fk,
			Reason: fmt.Sprintf("%s.%s references %s.%s without tenant_id in the FK columns", fk.ChildTable, strings.Join(fk.ChildColumns, ","), fk.ParentTable, strings.Join(fk.ParentColumns, ",")),
		})
	}
	return violations, nil
}
