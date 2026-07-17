// Package migration provides the DATA-09 online-migration protocol tooling:
// manifest parsing/validation, bounded lock-timeout DDL execution, expand-phase
// helpers, a resumable backfill harness, validation-phase artifact generation,
// and canary/switch/contract phase gates.
package migration

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Classification values for a migration's operational posture.
type Classification string

const (
	Online      Classification = "online"
	Maintenance Classification = "maintenance"
)

// ManifestRequiredVersion is the first post-baseline migration number that
// must carry a validated +wowapi:manifest block. The clean 00001 baseline is
// exempt; every future incremental kernel migration is governed by the online
// migration protocol.
const ManifestRequiredVersion = 2

// Manifest is the parsed, machine-readable declaration for one migration.
type Manifest struct {
	SourceFile         string
	Classification     Classification
	RowsEstimate       int64
	BytesEstimate      int64
	LockTimeoutMs      int64
	StatementTimeoutMs int64
	NN1Compatible      bool
	BackfillOwner      string
	ValidationQuery    string
	RollbackPlan       string
}

// OnlineLockBudgetMs is the DATA-09 T2 default/maximum online DDL lock budget.
const OnlineLockBudgetMs = 2000

var (
	manifestStartRE = regexp.MustCompile(`^\s*--\s*\+wowapi:manifest\s*$`)
	manifestEndRE   = regexp.MustCompile(`^\s*--\s*\+wowapi:end\s*$`)
	manifestLineRE  = regexp.MustCompile(`^\s*--\s*([a-z_][a-z0-9_]*):\s*(.*?)\s*$`)
)

// ParseManifest extracts the manifest block from migration SQL source.
// It returns (nil, nil) when no +wowapi:manifest block is present. It returns
// an error if the block is malformed.
func ParseManifest(sourceFile, body string) (*Manifest, error) {
	lines := strings.Split(body, "\n")
	inBlock := false
	found := false
	m := &Manifest{SourceFile: sourceFile}
	seen := map[string]int{}
	for i, line := range lines {
		if manifestStartRE.MatchString(line) {
			if inBlock {
				return nil, fmt.Errorf("%s:%d: nested +wowapi:manifest block", sourceFile, i+1)
			}
			inBlock = true
			found = true
			continue
		}
		if manifestEndRE.MatchString(line) {
			if !inBlock {
				return nil, fmt.Errorf("%s:%d: +wowapi:end without opening block", sourceFile, i+1)
			}
			inBlock = false
			continue
		}
		if !inBlock {
			continue
		}
		matches := manifestLineRE.FindStringSubmatch(line)
		if matches == nil {
			if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "--") {
				continue
			}
			return nil, fmt.Errorf("%s:%d: manifest line must be '-- key: value'", sourceFile, i+1)
		}
		key, value := matches[1], matches[2]
		if _, ok := seen[key]; ok {
			return nil, fmt.Errorf("%s:%d: duplicate manifest key %q", sourceFile, i+1, key)
		}
		seen[key] = i + 1
		if err := assignManifestField(m, key, value); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", sourceFile, i+1, err)
		}
	}
	if inBlock {
		return nil, fmt.Errorf("%s: unclosed +wowapi:manifest block", sourceFile)
	}
	if !found {
		return nil, nil
	}
	return m, nil
}

func assignManifestField(m *Manifest, key, value string) error {
	switch key {
	case "classification":
		c := Classification(value)
		if c != Online && c != Maintenance {
			return fmt.Errorf("classification must be 'online' or 'maintenance', got %q", value)
		}
		m.Classification = c
	case "rows_estimate":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil || v < 0 {
			return fmt.Errorf("rows_estimate must be a non-negative integer, got %q", value)
		}
		m.RowsEstimate = v
	case "bytes_estimate":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil || v < 0 {
			return fmt.Errorf("bytes_estimate must be a non-negative integer, got %q", value)
		}
		m.BytesEstimate = v
	case "lock_timeout_ms":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil || v <= 0 {
			return fmt.Errorf("lock_timeout_ms must be a positive integer, got %q", value)
		}
		m.LockTimeoutMs = v
	case "statement_timeout_ms":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil || v <= 0 {
			return fmt.Errorf("statement_timeout_ms must be a positive integer, got %q", value)
		}
		m.StatementTimeoutMs = v
	case "nn1_compatible":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("nn1_compatible must be 'true' or 'false', got %q", value)
		}
		m.NN1Compatible = v
	case "backfill_owner":
		m.BackfillOwner = value
	case "validation_query":
		m.ValidationQuery = value
	case "rollback_plan":
		m.RollbackPlan = value
	default:
		return fmt.Errorf("unknown manifest key %q", key)
	}
	return nil
}

// Validate checks that m contains every required field and satisfies the
// cross-field rules (online lock budget, timeout ordering, etc.).
func (m *Manifest) Validate() error {
	if m.Classification != Online && m.Classification != Maintenance {
		return fieldError(m.SourceFile, "classification", "required; must be 'online' or 'maintenance'")
	}
	if m.LockTimeoutMs <= 0 {
		return fieldError(m.SourceFile, "lock_timeout_ms", "required positive integer")
	}
	if m.StatementTimeoutMs <= 0 {
		return fieldError(m.SourceFile, "statement_timeout_ms", "required positive integer")
	}
	if m.StatementTimeoutMs < m.LockTimeoutMs {
		return fieldError(m.SourceFile, "statement_timeout_ms", "must be >= lock_timeout_ms (%d ms)", m.LockTimeoutMs)
	}
	if m.Classification == Online && m.LockTimeoutMs > OnlineLockBudgetMs {
		return fieldError(m.SourceFile, "lock_timeout_ms", "online migrations must be <= %d ms", OnlineLockBudgetMs)
	}
	if m.BackfillOwner == "" {
		return fieldError(m.SourceFile, "backfill_owner", "required (use 'none' for no backfill)")
	}
	if m.ValidationQuery == "" {
		return fieldError(m.SourceFile, "validation_query", "required (use 'none' for self-evident catalog DDL)")
	}
	if m.RollbackPlan == "" {
		return fieldError(m.SourceFile, "rollback_plan", "required")
	}
	return nil
}

func fieldError(file, field, format string, args ...any) error {
	prefix := "manifest"
	if file != "" {
		prefix = file
	}
	return fmt.Errorf("%s: field %q: %s", prefix, field, fmt.Sprintf(format, args...))
}

// MigrationVersion extracts the leading NNNNN number from a goose-style
// migration filename.
func MigrationVersion(filename string) (int, error) {
	parts := strings.SplitN(filename, "_", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("migration filename %q is not NNNNN_name.sql", filename)
	}
	v, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("migration filename %q: %w", filename, err)
	}
	return v, nil
}

// MustParseManifest is a test helper that panics on parse error.
func MustParseManifest(sourceFile, body string) *Manifest {
	m, err := ParseManifest(sourceFile, body)
	if err != nil {
		panic(err)
	}
	return m
}
