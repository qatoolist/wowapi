package workflow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

const definitionDigestBytes = sha256.Size

var workflowDefinitionNamespace = uuid.MustParse("35c6f1ee-36ab-5f36-9b5f-52e8216fbf7a")

func definitionRowID(key string, version int) uuid.UUID {
	return uuid.NewSHA1(workflowDefinitionNamespace, []byte(fmt.Sprintf("%s\x00%d", key, version)))
}

// canonicalDefinitionV1 is the sole workflow-definition persistence format.
// It rejects definitions that do not satisfy the closed graph/shape contract,
// normalizes semantically empty collections, and relies on encoding/json's
// deterministic string-key ordering. Registry synchronization additionally
// requires that the exact snapshot passed the registry's external-reference
// validation for its current generation.
func canonicalDefinitionV1(def Definition) ([]byte, error) {
	// Persistence validation must cover the definition's complete structural
	// contract. External callback existence is checked by Registry.Err; mark the
	// referenced names present here so this function does not invent a second
	// registry and cannot disagree about which callbacks are installed.
	autos := make(map[string]bool)
	resolvers := make(map[string]bool)
	for _, step := range def.Steps {
		if step.Action != "" {
			autos[step.Action] = true
		}
		for _, assignee := range step.Assignees {
			if assignee.Resolver != "" {
				resolvers[assignee.Resolver] = true
			}
		}
	}
	if err := def.Validate(autos, resolvers); err != nil {
		return nil, kerr.E(kerr.KindValidation, "workflow_definition_not_canonicalizable",
			"workflow definition is not valid for persistence: "+err.Error())
	}

	normalized := def.clone()
	for name, step := range normalized.Steps {
		if len(step.Assignees) == 0 {
			step.Assignees = nil
		}
		if len(step.Branches) == 0 {
			step.Branches = nil
		} else {
			for i := range step.Branches {
				if step.Branches[i].When != nil {
					value, err := canonicalConditionScalar(step.Branches[i].When.Equals)
					if err != nil {
						return nil, kerr.E(kerr.KindValidation, "workflow_definition_not_canonicalizable",
							"workflow condition cannot be represented canonically: "+err.Error())
					}
					step.Branches[i].When.Equals = value
				}
			}
		}
		normalized.Steps[name] = step
	}
	b, err := json.Marshal(normalized)
	if err != nil {
		return nil, kerr.E(kerr.KindValidation, "workflow_definition_not_canonicalizable",
			"workflow definition cannot be represented as canonical JSON: "+err.Error())
	}
	return b, nil
}

func canonicalConditionScalar(value any) (any, error) {
	switch value.(type) {
	case string, bool:
		return value, nil
	}
	b, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	n := strings.TrimSpace(string(b))
	if n == "" || n == "null" || n[0] == '"' || n[0] == '{' || n[0] == '[' {
		return nil, fmt.Errorf("unsupported scalar %T", value)
	}
	canonical, err := canonicalJSONNumber(n)
	if err != nil {
		return nil, err
	}
	return json.Number(canonical), nil
}

// canonicalJSONNumber normalizes a valid JSON number without converting it to
// float64. The coefficient/exponent form keeps arbitrary source precision and
// also avoids allocating enormous zero-filled strings for inputs such as
// 1e1000000. Equivalent spellings therefore have one stable representation.
func canonicalJSONNumber(number string) (string, error) {
	n := number
	negative := strings.HasPrefix(n, "-")
	if negative {
		n = n[1:]
	}

	mantissa, exponentText := n, "0"
	if i := strings.IndexAny(n, "eE"); i >= 0 {
		mantissa, exponentText = n[:i], n[i+1:]
	}
	exponent := new(big.Int)
	if _, ok := exponent.SetString(exponentText, 10); !ok {
		return "", fmt.Errorf("invalid JSON number %q", number)
	}

	integer, fraction := mantissa, ""
	if i := strings.IndexByte(mantissa, '.'); i >= 0 {
		integer, fraction = mantissa[:i], mantissa[i+1:]
	}
	digits := integer + fraction
	for _, digit := range digits {
		if digit < '0' || digit > '9' {
			return "", fmt.Errorf("invalid JSON number %q", number)
		}
	}
	digits = strings.TrimLeft(digits, "0")
	if digits == "" {
		return "0", nil
	}

	trailing := len(digits) - len(strings.TrimRight(digits, "0"))
	digits = strings.TrimRight(digits, "0")
	exponent.Sub(exponent, big.NewInt(int64(len(fraction))))
	exponent.Add(exponent, big.NewInt(int64(trailing)))
	if negative {
		digits = "-" + digits
	}
	if exponent.Sign() == 0 {
		return digits, nil
	}
	return digits + "e" + exponent.String(), nil
}

func definitionDigestV1(def Definition) (string, error) {
	b, err := canonicalDefinitionV1(def)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

type definitionSnapshot struct {
	Definition Definition
	Canonical  []byte
	Digest     string
}

type persistedDefinition struct {
	ID        uuid.UUID
	Key       string
	Version   int
	AppliesTo string
	Raw       []byte
	Digest    string
	Status    string
}

func validDefinitionDigest(digest string) bool {
	if len(digest) != hex.EncodedLen(definitionDigestBytes) {
		return false
	}
	decoded, err := hex.DecodeString(digest)
	return err == nil && len(decoded) == definitionDigestBytes && hex.EncodeToString(decoded) == digest
}

// verifyDefinition is the one execution-time identity gate shared by starts,
// task/instance mutations, and SLA processing. The persisted row is never an
// execution fallback: it must prove byte-meaning equivalence to the exact
// currently validated registered definition.
func (rt *Runtime) verifyDefinition(row persistedDefinition) (Definition, error) {
	registered, validated, found := rt.registry.resolveValidated(row.Key, row.Version)
	if !validated {
		return Definition{}, kerr.E(kerr.KindInternal, "workflow_registry_unvalidated",
			"workflow registry was invalidated during definition verification")
	}
	if !found {
		return Definition{}, kerr.E(kerr.KindInternal, "workflow_definition_unregistered",
			fmt.Sprintf("persisted workflow definition %s v%d is not registered", row.Key, row.Version))
	}
	if row.ID != definitionRowID(row.Key, row.Version) || row.Status != "active" || row.AppliesTo != registered.AppliesTo {
		return Definition{}, definitionIdentityError(row, "persisted scalar identity differs from the registered definition")
	}
	if !validDefinitionDigest(row.Digest) {
		return Definition{}, definitionIdentityError(row, "definition digest is not lowercase SHA-256")
	}
	persisted, err := rt.parseAndValidateDefinition(row.Raw)
	if err != nil {
		return Definition{}, definitionIdentityError(row, "persisted JSON is invalid: "+err.Error())
	}
	persistedCanonical, err := canonicalDefinitionV1(persisted)
	if err != nil {
		return Definition{}, definitionIdentityError(row, "persisted JSON cannot be canonicalized: "+err.Error())
	}
	persistedSum := sha256.Sum256(persistedCanonical)
	if hex.EncodeToString(persistedSum[:]) != row.Digest {
		return Definition{}, definitionIdentityError(row, "persisted JSON does not match its stored digest")
	}
	registeredDigest, err := definitionDigestV1(registered)
	if err != nil {
		return Definition{}, kerr.Wrapf(err, "workflow.verifyDefinition", "canonicalize registered %s v%d", row.Key, row.Version)
	}
	if registeredDigest != row.Digest {
		return Definition{}, definitionIdentityError(row, "persisted digest does not match the registered definition")
	}
	return registered, nil
}

func definitionIdentityError(row persistedDefinition, detail string) error {
	return kerr.E(kerr.KindConflict, "workflow_definition_identity_mismatch",
		fmt.Sprintf("workflow definition %s v%d failed identity verification: %s", row.Key, row.Version, detail))
}

func scanPersistedDefinition(row pgx.Row) (persistedDefinition, error) {
	var def persistedDefinition
	err := row.Scan(&def.ID, &def.Key, &def.Version, &def.AppliesTo, &def.Raw, &def.Digest, &def.Status)
	return def, err
}

func (rt *Runtime) loadVerifiedDefinitionByID(ctx context.Context, db database.DBTX, id uuid.UUID) (Definition, error) {
	row, err := scanPersistedDefinition(db.QueryRow(ctx, `SELECT id, key, version, applies_to, definition, definition_digest, status
        FROM workflow_definitions WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Definition{}, kerr.E(kerr.KindNotFound, "workflow_definition_not_found",
				"workflow definition row not found: "+id.String())
		}
		return Definition{}, kerr.Wrapf(err, "workflow.loadDefinition", "load definition row %s", id)
	}
	return rt.verifyDefinition(row)
}

func (rt *Runtime) loadVerifiedDefinitionByIdentity(ctx context.Context, db database.DBTX, registered Definition) (uuid.UUID, Definition, error) {
	row, err := scanPersistedDefinition(db.QueryRow(ctx, `SELECT id, key, version, applies_to, definition, definition_digest, status
        FROM workflow_definitions WHERE key = $1 AND version = $2`, registered.Key, registered.Version))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, Definition{}, kerr.E(kerr.KindNotFound, "workflow_definition_not_found",
				fmt.Sprintf("workflow definition %s v%d is not synchronized", registered.Key, registered.Version))
		}
		return uuid.Nil, Definition{}, kerr.Wrapf(err, "workflow.loadDefinition", "load definition %s v%d", registered.Key, registered.Version)
	}
	verified, err := rt.verifyDefinition(row)
	if err != nil {
		return uuid.Nil, Definition{}, err
	}
	return row.ID, verified, nil
}

func makeDefinitionSnapshot(def Definition) (definitionSnapshot, error) {
	canonical, err := canonicalDefinitionV1(def)
	if err != nil {
		return definitionSnapshot{}, err
	}
	sum := sha256.Sum256(canonical)
	return definitionSnapshot{
		Definition: def.clone(),
		Canonical:  canonical,
		Digest:     hex.EncodeToString(sum[:]),
	}, nil
}

// SyncDefinitions atomically materializes one current, validated registry
// snapshot into the global workflow_definitions catalog. Existing identities
// are immutable: a conflict is accepted only when every persisted identity
// field and canonical byte meaning is identical.
//
// db must be a platform-privileged *pgxpool.Pool or an existing pgx.Tx. The
// pool form owns one transaction for the complete definition set; the tx form
// composes into a caller-owned transaction.
func SyncDefinitions(ctx context.Context, db database.DBTX, reg *Registry) error {
	if reg == nil {
		return kerr.E(kerr.KindInternal, "workflow_registry_required", "workflow definition sync requires a registry")
	}
	defs, generation, err := reg.validatedSnapshot()
	if err != nil {
		return err
	}
	snapshots := make([]definitionSnapshot, 0, len(defs))
	for _, def := range defs {
		snapshot, err := makeDefinitionSnapshot(def)
		if err != nil {
			return kerr.Wrapf(err, "workflow.SyncDefinitions", "canonicalize %s v%d", def.Key, def.Version)
		}
		snapshots = append(snapshots, snapshot)
	}
	sort.Slice(snapshots, func(i, j int) bool {
		a, b := snapshots[i].Definition, snapshots[j].Definition
		if a.Key != b.Key {
			return a.Key < b.Key
		}
		return a.Version < b.Version
	})

	tx, own, err := beginDefinitionSyncTx(ctx, db)
	if err != nil {
		return err
	}
	if own {
		defer func() { _ = tx.Rollback(ctx) }()
	}
	for _, snapshot := range snapshots {
		if err := syncDefinition(ctx, tx, snapshot); err != nil {
			return err
		}
	}
	if !reg.snapshotStillValidated(generation) {
		return kerr.E(kerr.KindInternal, "workflow_registry_changed_during_sync",
			"workflow registry changed or lost validation while its definitions were synchronizing")
	}
	if own {
		if err := tx.Commit(ctx); err != nil {
			return kerr.Wrapf(err, "workflow.SyncDefinitions", "commit")
		}
	}
	return nil
}

func beginDefinitionSyncTx(ctx context.Context, db database.DBTX) (pgx.Tx, bool, error) {
	switch d := db.(type) {
	case pgx.Tx:
		return d, false, nil
	case *pgxpool.Pool:
		tx, err := d.Begin(ctx)
		if err != nil {
			return nil, false, kerr.Wrapf(err, "workflow.SyncDefinitions", "begin transaction")
		}
		return tx, true, nil
	default:
		return nil, false, fmt.Errorf("workflow.SyncDefinitions: db must be *pgxpool.Pool or pgx.Tx, got %T", db)
	}
}

func syncDefinition(ctx context.Context, tx pgx.Tx, snapshot definitionSnapshot) error {
	def := snapshot.Definition
	newID := definitionRowID(def.Key, def.Version)
	if _, err := tx.Exec(ctx, `INSERT INTO workflow_definitions
            (id, key, version, applies_to, definition, definition_digest, status, created_by)
        VALUES ($1, $2, $3, $4, $5, $6, 'active', $7)
        ON CONFLICT (key, version) DO NOTHING`,
		newID, def.Key, def.Version, def.AppliesTo, snapshot.Canonical, snapshot.Digest, uuid.Nil); err != nil {
		return kerr.Wrapf(err, "workflow.SyncDefinitions", "insert %s v%d", def.Key, def.Version)
	}

	var (
		id        uuid.UUID
		key       string
		version   int
		appliesTo string
		raw       []byte
		digest    string
		status    string
	)
	if err := tx.QueryRow(ctx, `SELECT id, key, version, applies_to, definition, definition_digest, status
        FROM workflow_definitions WHERE key = $1 AND version = $2 FOR UPDATE`, def.Key, def.Version).
		Scan(&id, &key, &version, &appliesTo, &raw, &digest, &status); err != nil {
		return kerr.Wrapf(err, "workflow.SyncDefinitions", "load %s v%d", def.Key, def.Version)
	}
	persisted, err := ParseDefinition(raw)
	if err != nil {
		return definitionConflict(def, "persisted definition is malformed: "+err.Error())
	}
	persistedCanonical, err := canonicalDefinitionV1(persisted)
	if err != nil {
		return definitionConflict(def, "persisted definition is invalid: "+err.Error())
	}
	if id != newID || key != def.Key || version != def.Version || appliesTo != def.AppliesTo ||
		status != "active" || digest != snapshot.Digest || !bytes.Equal(persistedCanonical, snapshot.Canonical) {
		return definitionConflict(def, "persisted identity differs from the registered canonical definition")
	}
	return nil
}

func definitionConflict(def Definition, detail string) error {
	return kerr.E(kerr.KindConflict, "workflow_definition_conflict",
		fmt.Sprintf("workflow definition %s v%d is immutable: %s", def.Key, def.Version, detail))
}
