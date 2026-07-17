// Package audit is the durable, append-only, field-level audit trail (roadmap
// E1): a standardized record of who changed what — entity, field, before/after,
// actor, capacity, impersonator, request id — written INSIDE the business
// transaction so an audit row commits iff the change does. Append-only is
// enforced by the grants (app_rt has no UPDATE/DELETE on audit_logs); this
// package never offers a mutate path. Cryptographic tamper-evidence
// (hash-chaining, S6) layers on top of this table later.
package audit

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/model"
)

// Entry is a change to record. Action is required (e.g. "document.download",
// "receipt.void"); the rest are optional. For a field-level change set Field +
// OldValue + NewValue; for a whole-entity action leave Field empty. Values are
// passed through the Writer's redactor before persistence.
type Entry struct {
	Action         string
	EntityType     string
	EntityID       uuid.UUID // uuid.Nil → NULL
	Field          string
	OldValue       string
	NewValue       string
	Reason         string
	ActorKind      string    // user | system | webhook (optional)
	ImpersonatorID uuid.UUID // support impersonation (optional)
	Metadata       map[string]any
}

// Log is a persisted audit row returned by Query.
type Log struct {
	ID             uuid.UUID
	OccurredAt     time.Time
	ActorID        *uuid.UUID
	ActorKind      string
	ImpersonatorID *uuid.UUID
	RequestID      string
	Action         string
	EntityType     string
	EntityID       *uuid.UUID
	Field          string
	OldValue       string
	NewValue       string
	Reason         string
	Metadata       map[string]any
	TxID           string // database transaction id (forensic correlation; roadmap CA-11)
	HashVersion    int16  // 1 = historical 15-field scheme, 2 = widened scheme incl. metadata + tx_id
}

// Redactor may mutate an Entry before it is written — e.g. mask the values of
// known-sensitive fields so they never land in the audit table. It is the
// module's per-record redaction hook (blueprint 07 §1 "per-module redaction").
type Redactor func(*Entry)

// Writer appends and queries audit rows. It is stateless beyond its id generator
// and optional redactor.
type Writer struct {
	idgen  model.IDGen
	redact Redactor
}

// New builds a Writer. redact may be nil (no redaction).
func New(idgen model.IDGen, redact Redactor) *Writer {
	if idgen == nil {
		idgen = model.UUIDv7()
	}
	return &Writer{idgen: idgen, redact: redact}
}

// Record appends one audit row in db's transaction (so it commits with the
// business write) and extends the tenant's hash chain (S6). The acting actor id
// and request id are read from ctx; the caller supplies the semantic fields via
// e. Action is required. Concurrent audit writes for a tenant serialize on the
// audit_chain head row, so seq is gap-free and the chain is well-ordered.
func (w *Writer) Record(ctx context.Context, db database.TenantDB, e Entry) error {
	if e.Action == "" {
		return kerr.E(kerr.KindValidation, "invalid_audit", "audit action is required")
	}
	if w.redact != nil {
		w.redact(&e)
	}
	meta := e.Metadata
	if meta == nil {
		meta = map[string]any{}
	}
	metaBytes, err := canonicalizeMetadata(meta)
	if err != nil {
		return kerr.Wrapf(err, "audit.Record", "canonicalize metadata")
	}

	// Microsecond precision: Postgres timestamptz stores micros, so truncating
	// here keeps the hash input identical to what Verify later reads back.
	occurredAt := time.Now().UTC().Truncate(time.Microsecond)
	id := w.idgen.New()
	requestID := httpx.RequestIDFrom(ctx)
	actorStr := ""
	var actorArg any
	if aid, ok := database.ActorIDFrom(ctx); ok {
		actorStr, actorArg = aid.String(), aid
	}

	// Lock/read this tenant's chain head (genesis on first write). ON CONFLICT DO
	// UPDATE is a no-op that still row-locks and returns the current head.
	var seq int64
	var prevHash string
	if err := db.QueryRow(ctx,
		`INSERT INTO audit_chain (tenant_id) VALUES (app_tenant_id())
		 ON CONFLICT (tenant_id) DO UPDATE SET tenant_id = audit_chain.tenant_id
		 RETURNING next_seq, head_hash`).Scan(&seq, &prevHash); err != nil {
		return kerr.Wrapf(err, "audit.Record", "lock chain head")
	}

	// tx_id is part of the widened hash (hash_version=2). Read it from the
	// current transaction so the hash input is deterministic and reproducible.
	var txID string
	if err := db.QueryRow(ctx, `SELECT pg_current_xact_id()::text`).Scan(&txID); err != nil {
		return kerr.Wrapf(err, "audit.Record", "read tx_id")
	}

	rowHash := chainHash(hashVersion2, prevHash, seq, id, occurredAt, txID, metaBytes,
		actorStr, e.ActorKind, uuidStr(e.ImpersonatorID), requestID, e.Action,
		e.EntityType, uuidStr(e.EntityID), e.Field, e.OldValue, e.NewValue, e.Reason)

	if _, err := db.Exec(ctx,
		`INSERT INTO audit_logs
		    (id, tenant_id, occurred_at, actor_id, actor_kind, impersonator_id, request_id,
		     action, entity_type, entity_id, field, old_value, new_value, reason, metadata,
		     seq, row_hash, prev_hash, tx_id, hash_version)
		 VALUES ($1, app_tenant_id(), $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`,
		id, occurredAt, actorArg, nullStr(e.ActorKind), nullUUID(e.ImpersonatorID), nullStr(requestID),
		e.Action, nullStr(e.EntityType), nullUUID(e.EntityID), nullStr(e.Field), nullStr(e.OldValue),
		nullStr(e.NewValue), nullStr(e.Reason), metaBytes, seq, rowHash, prevHash, txID, hashVersion2); err != nil {
		return kerr.Wrapf(err, "audit.Record", "insert audit row")
	}

	if _, err := db.Exec(ctx,
		`UPDATE audit_chain SET next_seq = $1, head_hash = $2 WHERE tenant_id = app_tenant_id()`,
		seq+1, rowHash); err != nil {
		return kerr.Wrapf(err, "audit.Record", "advance chain head")
	}
	return nil
}

const (
	hashVersion1 int16 = 1 // historical 15-field scheme (metadata and tx_id excluded)
	hashVersion2 int16 = 2 // widened scheme: canonicalized metadata + tx_id included
)

// canonicalizeMetadata returns a deterministic JSON representation of m suitable
// for hashing. It relies on encoding/json's default map-key sorting and stable
// encoding of scalar values; the hash never uses the stored jsonb bytes directly
// because Postgres jsonb does not preserve formatting or key order. The returned
// bytes are also the canonical form stored in the metadata column.
func canonicalizeMetadata(m map[string]any) ([]byte, error) {
	if m == nil {
		m = map[string]any{}
	}
	return json.Marshal(m)
}

// chainHash computes row_hash = sha256(lp(prev) || lp(seq) || lp(id) || lp(occurredAt)
// || [v2: lp(txID) || lp(metadata)] || lp(fields…)). Each field is length-prefixed
// so no combination of values can collide with a different one.
//
// version selects the scheme: v1 is the historical 15-field scheme; v2 widens the
// input set to include canonicalized metadata and tx_id. Historical rows
// (hash_version=1) continue to verify under v1; new rows are written as v2.
func chainHash(version int16, prev string, seq int64, id uuid.UUID, occurredAt time.Time, txID string, metadata []byte, fields ...string) string {
	h := sha256.New()
	lp(h, prev)
	var seqb [8]byte
	binary.BigEndian.PutUint64(seqb[:], uint64(seq)) // #nosec G115 -- bijective int64→uint64 reinterpretation for hash-input encoding; seq is a monotonically advanced chain sequence (never negative)
	h.Write(seqb[:])
	lp(h, id.String())
	lp(h, occurredAt.UTC().Format(time.RFC3339Nano))
	if version >= hashVersion2 {
		lp(h, txID)
		lp(h, string(metadata))
	}
	for _, f := range fields {
		lp(h, f)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func lp(h hash.Hash, s string) {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], uint32(len(s))) // #nosec G115 -- audit field values are single Postgres row fields, bounded far below 4 GiB, so the length prefix cannot truncate
	_, _ = h.Write(b[:])
	_, _ = h.Write([]byte(s))
}

// VerifyResult reports a chain verification. OK is true when every row's hash
// recomputes and links to its predecessor with no seq gap.
type VerifyResult struct {
	OK        bool
	Count     int64  // rows checked
	HeadSeq   int64  // last seq seen
	BrokenSeq int64  // first seq where the chain broke (0 when OK)
	Reason    string // why it broke
}

// Verify walks the current tenant's audit chain in seq order, recomputing each
// row's hash and checking the prev-links and sequence continuity. It detects any
// mutation of a past row (its hash no longer matches) and any deletion (a seq
// gap). Read-only; safe in a read-only transaction.
func (w *Writer) Verify(ctx context.Context, db database.TenantDB) (VerifyResult, error) {
	rows, err := db.Query(ctx,
		`SELECT seq, id, occurred_at, actor_id, actor_kind, impersonator_id, request_id,
		        action, entity_type, entity_id, field, old_value, new_value, reason,
		        metadata, tx_id, hash_version, row_hash, prev_hash
		   FROM audit_logs WHERE tenant_id = app_tenant_id() ORDER BY seq`)
	if err != nil {
		return VerifyResult{}, kerr.Wrapf(err, "audit.Verify", "read chain")
	}
	defer rows.Close()

	var res VerifyResult
	expected := int64(1)
	prev := ""
	for rows.Next() {
		var (
			seq                                                         int64
			id                                                          uuid.UUID
			occurredAt                                                  time.Time
			actorID, impersonator, entityID                             *uuid.UUID
			action                                                      string
			actorKind, requestID, entityType, field, oldV, newV, reason *string
			metadata                                                    map[string]any
			txID                                                        *string
			hashVersion                                                 int16
			rowHash, prevHash                                           string
		)
		if err := rows.Scan(&seq, &id, &occurredAt, &actorID, &actorKind, &impersonator, &requestID,
			&action, &entityType, &entityID, &field, &oldV, &newV, &reason,
			&metadata, &txID, &hashVersion, &rowHash, &prevHash); err != nil {
			return VerifyResult{}, kerr.Wrapf(err, "audit.Verify", "scan row")
		}
		res.Count++
		res.HeadSeq = seq
		if seq != expected {
			res.BrokenSeq, res.Reason = seq, "sequence gap (a row was deleted)"
			return res, nil
		}
		if prevHash != prev {
			res.BrokenSeq, res.Reason = seq, "prev_hash does not link to the previous row"
			return res, nil
		}
		want, err := recomputeRowHash(prev, seq, id, occurredAt, hashVersion, deref(txID), metadata,
			uuidStr(ptrUUID(actorID)), deref(actorKind), uuidStr(ptrUUID(impersonator)),
			deref(requestID), action, deref(entityType), uuidStr(ptrUUID(entityID)),
			deref(field), deref(oldV), deref(newV), deref(reason))
		if err != nil {
			res.BrokenSeq, res.Reason = seq, fmt.Sprintf("row hash recompute failed (hash_version=%d): %v", hashVersion, err)
			return res, nil
		}
		if want != rowHash {
			res.BrokenSeq, res.Reason = seq, fmt.Sprintf("row_hash does not match (hash_version=%d)", hashVersion)
			return res, nil
		}
		prev = rowHash
		expected++
	}
	if err := rows.Err(); err != nil {
		return VerifyResult{}, kerr.Wrapf(err, "audit.Verify", "iterate chain")
	}
	res.OK = true
	return res, nil
}

// recomputeRowHash rebuilds the expected row_hash for a row under its recorded
// hash_version. v1 preserves the historical 15-field scheme; v2 includes the
// canonicalized metadata and tx_id. An unknown version fails closed.
func recomputeRowHash(prev string, seq int64, id uuid.UUID, occurredAt time.Time, hashVersion int16, txID string, metadata map[string]any, fields ...string) (string, error) {
	switch hashVersion {
	case hashVersion1:
		return chainHash(hashVersion1, prev, seq, id, occurredAt, "", nil, fields...), nil
	case hashVersion2:
		metaBytes, err := canonicalizeMetadata(metadata)
		if err != nil {
			return "", err
		}
		return chainHash(hashVersion2, prev, seq, id, occurredAt, txID, metaBytes, fields...), nil
	default:
		return "", fmt.Errorf("unknown hash_version %d", hashVersion)
	}
}

// Anchor returns the tenant's current chain head — the last seq and its hash.
// Exporting/publishing an anchor lets a later Verify prove no tampering occurred
// up to that seq. Returns (0, "") when the tenant has no audit rows yet.
func (w *Writer) Anchor(ctx context.Context, db database.TenantDB) (seq int64, headHash string, err error) {
	err = db.QueryRow(ctx,
		`SELECT next_seq - 1, head_hash FROM audit_chain WHERE tenant_id = app_tenant_id()`).
		Scan(&seq, &headHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, "", nil
	}
	if err != nil {
		return 0, "", kerr.Wrapf(err, "audit.Anchor", "read chain head")
	}
	return seq, headHash, nil
}

// CheckAnchor reports whether the tenant's LIVE chain still contains the row at
// anchored seq with anchored hash — the offline verifier's primitive against a
// previously exported anchor (roadmap CA-11). It closes Verify's blind spot: a
// tail truncation (drop the last k rows and rewind audit_chain.head_hash) leaves
// the remaining chain internally consistent, so Verify still returns OK; but the
// anchored (seq, hash) is gone, so CheckAnchor returns false. present=false means
// the tail was truncated or the anchored row rewritten after the anchor was taken.
// Read-only; safe in a read-only transaction.
func (w *Writer) CheckAnchor(ctx context.Context, db database.TenantDB, seq int64, headHash string) (present bool, err error) {
	if seq <= 0 { // the genesis/empty anchor covers nothing to check
		return true, nil
	}
	err = db.QueryRow(ctx,
		`SELECT EXISTS (
		   SELECT 1 FROM audit_logs
		    WHERE tenant_id = app_tenant_id() AND seq = $1 AND row_hash = $2)`,
		seq, headHash).Scan(&present)
	if err != nil {
		return false, kerr.Wrapf(err, "audit.CheckAnchor", "probe anchored row")
	}
	return present, nil
}

// ExportAnchors snapshots every tenant's current audit-chain head into the
// append-only audit_anchors table as immutable tamper-evidence (roadmap CA-11),
// returning the number of anchors written. Drive it from the leader-safe
// scheduler (a single replica claims each interval) so anchors are written once,
// not once per replica. It runs cross-tenant as app_platform (one INSERT..SELECT
// for all tenants, mirroring the idempotency sweep) on the platform pool.
//
// A tenant is anchored only when its chain head advanced past its last anchored
// seq, so re-running within an interval (or on a quiescent chain) writes nothing
// and the table stays bounded — one anchor per burst of audit activity.
func ExportAnchors(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	tag, err := pool.Exec(ctx,
		`INSERT INTO audit_anchors (tenant_id, anchor_seq, chain_head_hash, row_count)
		 SELECT c.tenant_id, c.next_seq - 1, c.head_hash, c.next_seq - 1
		   FROM audit_chain c
		  WHERE c.next_seq > 1
		    AND c.next_seq - 1 > COALESCE(
		        (SELECT max(a.anchor_seq) FROM audit_anchors a WHERE a.tenant_id = c.tenant_id), 0)`)
	if err != nil {
		return 0, kerr.Wrapf(err, "audit.ExportAnchors", "write anchors")
	}
	return int(tag.RowsAffected()), nil
}

// Filter narrows a Query. Zero-valued fields are ignored; Limit defaults to 100.
type Filter struct {
	EntityType string
	EntityID   uuid.UUID
	ActorID    uuid.UUID
	Action     string
	Limit      int
}

// Query returns audit rows matching the filter, newest first, in the caller's
// tenant transaction (RLS-scoped). All filter values are bound as parameters.
func (w *Writer) Query(ctx context.Context, db database.TenantDB, f Filter) ([]Log, error) {
	conds := []string{"true"}
	args := []any{}
	add := func(clause string, val any) {
		args = append(args, val)
		conds = append(conds, clause+" $"+strconv.Itoa(len(args)))
	}
	if f.EntityType != "" {
		add("entity_type =", f.EntityType)
	}
	if f.EntityID != uuid.Nil {
		add("entity_id =", f.EntityID)
	}
	if f.ActorID != uuid.Nil {
		add("actor_id =", f.ActorID)
	}
	if f.Action != "" {
		add("action =", f.Action)
	}
	limit := f.Limit
	if limit <= 0 {
		limit = 100
	}
	args = append(args, limit)
	sql := `SELECT id, occurred_at, actor_id, actor_kind, impersonator_id, request_id,
	               action, entity_type, entity_id, field, old_value, new_value, reason,
	               metadata, tx_id, hash_version
	          FROM audit_logs
	         WHERE ` + strings.Join(conds, " AND ") +
		// id (UUIDv7) is a creation-ordered tiebreaker so rows written in the same
		// transaction (identical occurred_at) still sort newest-first.
		` ORDER BY occurred_at DESC, id DESC LIMIT $` + strconv.Itoa(len(args))

	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, kerr.Wrapf(err, "audit.Query", "query audit logs")
	}
	defer rows.Close()
	var out []Log
	for rows.Next() {
		var l Log
		var actorKind, requestID, entityType, field, oldV, newV, reason, txID *string
		if err := rows.Scan(&l.ID, &l.OccurredAt, &l.ActorID, &actorKind, &l.ImpersonatorID,
			&requestID, &l.Action, &entityType, &l.EntityID, &field, &oldV, &newV, &reason,
			&l.Metadata, &txID, &l.HashVersion); err != nil {
			return nil, kerr.Wrapf(err, "audit.Query", "scan audit row")
		}
		l.TxID = deref(txID)
		l.ActorKind = deref(actorKind)
		l.RequestID = deref(requestID)
		l.EntityType = deref(entityType)
		l.Field = deref(field)
		l.OldValue = deref(oldV)
		l.NewValue = deref(newV)
		l.Reason = deref(reason)
		if l.Metadata == nil {
			l.Metadata = map[string]any{}
		}
		out = append(out, l)
	}
	if err := rows.Err(); err != nil {
		return nil, kerr.Wrapf(err, "audit.Query", "iterate audit logs")
	}
	return out, nil
}

// --- helpers: NULL-safe binding + scanning ---

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullUUID(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// uuidStr renders a uuid for hashing: "" for the nil uuid, canonical form else.
func uuidStr(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return id.String()
}

func ptrUUID(p *uuid.UUID) uuid.UUID {
	if p == nil {
		return uuid.Nil
	}
	return *p
}
