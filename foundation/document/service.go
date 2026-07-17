package document

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	kaudit "github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/storage"
)

// Permissions the download gate evaluates when an authz Evaluator is wired. The
// consumer registers them (kernel.document.read/write) at boot; the framework
// wiring does so by default.
const (
	PermRead = "kernel.document.read"
	// PermWrite is the authz permission for mutating a document (managing grants,
	// adding versions). "update" is the closed-set verb for write access; the
	// grant table's own access column still uses the literal "write".
	PermWrite = "kernel.document.update"
	// docResourceType is the authz Target resource type for a document.
	docResourceType = "kernel.document"
)

// Default presigned-URL lifetimes (blueprint 07 §4: upload window generous,
// download window tight).
const (
	defaultPutTTL = 15 * time.Minute
	defaultGetTTL = 60 * time.Second
)

// Service is the document framework. Metadata mutations run inside the caller's
// tenant transaction (so a document commits with its business write); scan-status
// and retention voiding run on a platform-privileged tenant-bound manager because
// document_versions is append-only to the module role.
type Service struct {
	registry *Registry
	store    storage.Adapter
	authz    authz.Evaluator // optional secondary gate (adds role/policy access)
	outbox   outbox.Writer
	hooks    *Hooks
	idgen    model.IDGen
	audit    *kaudit.Writer
	now      func() time.Time
	putTTL   time.Duration
	getTTL   time.Duration
}

// Option configures a Service.
type Option func(*Service)

// WithAudit wires a durable audit writer. When set, document operations emit
// audit rows inside the same transaction as the business change.
func WithAudit(aud *kaudit.Writer) Option {
	return func(s *Service) { s.audit = aud }
}

// New wires the service. registry/store/outbox/idgen are required; authz,
// hooks, and options are optional.
func New(reg *Registry, store storage.Adapter, ev authz.Evaluator, ob outbox.Writer, hooks *Hooks, idgen model.IDGen, opts ...Option) *Service {
	if reg == nil || store == nil || ob == nil || idgen == nil {
		panic("document.New: registry, store, outbox, and idgen are required")
	}
	s := &Service{
		registry: reg, store: store, authz: ev, outbox: ob, hooks: hooks,
		idgen: idgen, now: time.Now, putTTL: defaultPutTTL, getTTL: defaultGetTTL,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// --- inputs / outputs ---

// CreateInput describes a new document's metadata. Resource is an optional anchor.
type CreateInput struct {
	Class       string
	Resource    resource.Ref
	Title       string
	Sensitivity Sensitivity // "" → the class default
}

// UploadSession is what a client needs to PUT a version's bytes.
type UploadSession struct {
	SessionID  uuid.UUID
	DocumentID uuid.UUID
	VersionNo  int
	StorageKey string
	Upload     storage.PresignedURL
}

// ConfirmInput finalizes an uploaded version. Declared size/checksum/MIME are
// verified against the stored object before the immutable row is written.
type ConfirmInput struct {
	SessionID        uuid.UUID
	DocumentID       uuid.UUID
	VersionNo        int
	StorageKey       string
	DeclaredSize     int64
	DeclaredChecksum string
	DeclaredMIME     string
}

// DownloadInput selects a version to download; VersionNo 0 → latest active.
type DownloadInput struct {
	DocumentID uuid.UUID
	VersionNo  int
}

// Download is an authorized, time-boxed download of one version.
type Download struct {
	VersionNo int
	MIME      string
	URL       storage.PresignedURL
}

// GrantInput adds an explicit access grant beyond policy.
type GrantInput struct {
	DocumentID  uuid.UUID
	GranteeKind string // capacity | role | relationship
	GranteeRef  string
	Access      string // read | write
	ValidTo     *time.Time
}

// --- module-facing operations (run on the caller's app_rt tenant tx) ---

// Create writes a document metadata row and returns its id.
func (s *Service) Create(ctx context.Context, db database.TenantDB, in CreateInput) (uuid.UUID, error) {
	class, ok := s.registry.Get(in.Class)
	if !ok {
		return uuid.Nil, kerr.E(kerr.KindValidation, "unknown_document_class", "unregistered document class: "+in.Class)
	}
	if strings.TrimSpace(in.Title) == "" {
		return uuid.Nil, kerr.E(kerr.KindValidation, "document_invalid", "document title is required")
	}
	sens := in.Sensitivity
	if sens == "" {
		sens = class.DefaultSensitivity
	}
	if !sens.valid() {
		return uuid.Nil, kerr.E(kerr.KindValidation, "document_invalid", "invalid sensitivity: "+string(sens))
	}
	var retentionUntil any
	if class.Retention > 0 {
		retentionUntil = s.now().Add(class.Retention)
	}
	id := s.idgen.New()
	actor := actorFromCtx(ctx)
	_, err := db.Exec(ctx,
		`INSERT INTO documents
		    (id, tenant_id, document_class, resource_type, resource_id, title, sensitivity, retention_until, created_by)
		 VALUES ($1, app_tenant_id(), $2, $3, $4, $5, $6, $7, $8)`,
		id, in.Class, nullStr(in.Resource.Type), nullUUID(in.Resource.ID), in.Title, string(sens), retentionUntil, actor)
	if err != nil {
		return uuid.Nil, kerr.Wrapf(err, "document.Create", "insert document")
	}
	if err := s.emit(ctx, db, "document.created", docRef(id), map[string]any{"class": in.Class, "sensitivity": sens}); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// InitiateUpload is retained as a fail-closed compatibility entry point.
// Framework uploads must provide the checksum before a URL can be signed.
func (s *Service) InitiateUpload(_ context.Context, _ database.TenantDB, _ uuid.UUID) (UploadSession, error) {
	return UploadSession{}, kerr.E(kerr.KindValidation, "upload_checksum_required", "upload SHA-256 is required at initiation")
}

// InitiateUploadChecksum reserves the next version number and returns a PUT
// whose signed headers bind S3's canonical SHA-256 metadata to the upload.
func (s *Service) InitiateUploadChecksum(ctx context.Context, db database.TenantDB, docID uuid.UUID, checksumSHA256 string) (UploadSession, error) {
	rawChecksum, checksumErr := hex.DecodeString(checksumSHA256)
	if checksumErr != nil || len(rawChecksum) != sha256.Size || checksumSHA256 != strings.ToLower(checksumSHA256) {
		return UploadSession{}, kerr.E(kerr.KindValidation, "invalid_upload_checksum", "upload checksum must be lowercase-hex SHA-256")
	}
	uploader, ok := s.store.(storage.ChecksumUploader)
	if !ok {
		return UploadSession{}, kerr.E(kerr.KindInternal, "checksum_upload_unsupported", "storage adapter does not support checksum-enforcing uploads")
	}
	var status string
	err := db.QueryRow(ctx, `SELECT status FROM documents WHERE id = $1`, docID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return UploadSession{}, kerr.E(kerr.KindNotFound, "not_found", "document not found")
	}
	if err != nil {
		return UploadSession{}, kerr.Wrapf(err, "document.InitiateUpload", "load document")
	}
	if status != "active" {
		return UploadSession{}, kerr.E(kerr.KindConflict, "document_voided", "document is not active")
	}
	var next int
	scope := "document:" + docID.String()
	if err := db.QueryRow(ctx,
		`INSERT INTO version_counters (tenant_id, scope, value)
		 VALUES (app_tenant_id(), $1, 1)
		 ON CONFLICT (tenant_id, scope) DO UPDATE SET value = version_counters.value + 1
		 RETURNING value`, scope).Scan(&next); err != nil {
		return UploadSession{}, kerr.Wrapf(err, "document.InitiateUpload", "allocate version")
	}
	// The storage key uses a random suffix, NOT the version number: two concurrent
	// InitiateUpload calls both compute the same next version_no, and a
	// version-derived key would make them PUT to the same object and clobber each
	// other (ARCH-66). A random key gives each attempt its own blob; the loser of
	// the version_no race gets a KindConflict at confirm and simply orphans its
	// unreferenced blob (swept by a future storage GC).
	tenantID, _ := database.TenantIDFrom(ctx)
	key := tenantID.String() + "/" + docID.String() + "/" + s.idgen.New().String()
	url, err := uploader.PresignPutChecksum(ctx, key, checksumSHA256, s.putTTL)
	if err != nil {
		return UploadSession{}, kerr.Wrapf(err, "document.InitiateUpload", "presign put")
	}
	sessID := s.idgen.New()
	expires := s.now().Add(s.putTTL)
	_, err = db.Exec(ctx,
		`INSERT INTO document_upload_sessions
		    (id, tenant_id, document_id, version_no, storage_key, status, expires_at, checksum_sha256, created_at)
		 VALUES ($1, app_tenant_id(), $2, $3, $4, 'pending', $5, $6, now())`,
		sessID, docID, next, key, expires, checksumSHA256)
	if err != nil {
		return UploadSession{}, kerr.Wrapf(err, "document.InitiateUpload", "insert upload session")
	}
	return UploadSession{SessionID: sessID, DocumentID: docID, VersionNo: next, StorageKey: key, Upload: url}, nil
}

// ConfirmUpload verifies the uploaded bytes (size, checksum, MIME sniff, class
// limits), runs OnFileUpload hooks, and writes the immutable version row.
func (s *Service) ConfirmUpload(ctx context.Context, db database.TenantDB, in ConfirmInput) (uuid.UUID, error) {
	var (
		class string
		sens  string
	)
	err := db.QueryRow(ctx, `SELECT document_class, sensitivity FROM documents WHERE id = $1`, in.DocumentID).
		Scan(&class, &sens)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, kerr.E(kerr.KindNotFound, "not_found", "document not found")
	}
	if err != nil {
		return uuid.Nil, kerr.Wrapf(err, "document.ConfirmUpload", "load document")
	}
	cl, ok := s.registry.Get(class)
	if !ok {
		return uuid.Nil, kerr.E(kerr.KindInternal, "unknown_document_class", "document references unregistered class: "+class)
	}

	info, err := s.store.Stat(ctx, in.StorageKey)
	if kerr.KindOf(err) == kerr.KindNotFound {
		return uuid.Nil, kerr.E(kerr.KindValidation, "upload_missing", "no object was uploaded to the presigned key")
	}
	if err != nil {
		return uuid.Nil, kerr.Wrapf(err, "document.ConfirmUpload", "stat object")
	}
	if info.Size != in.DeclaredSize {
		return uuid.Nil, kerr.E(kerr.KindValidation, "size_mismatch", "uploaded size does not match declared size")
	}
	if !strings.EqualFold(info.Checksum, in.DeclaredChecksum) {
		return uuid.Nil, kerr.E(kerr.KindValidation, "checksum_mismatch", "uploaded checksum does not match declared checksum")
	}
	if cl.MaxBytes > 0 && info.Size > cl.MaxBytes {
		return uuid.Nil, kerr.E(kerr.KindValidation, "too_large", "uploaded object exceeds the class size limit")
	}

	// MIME: sniff the leading bytes and reconcile with the declared type.
	prefix, err := s.store.Peek(ctx, in.StorageKey, 512)
	if err != nil {
		return uuid.Nil, kerr.Wrapf(err, "document.ConfirmUpload", "peek object")
	}
	sniffed := http.DetectContentType(prefix)
	mime := in.DeclaredMIME
	if mime == "" {
		mime = sniffed
	}
	if mimeConflict(mime, sniffed) {
		return uuid.Nil, kerr.E(kerr.KindValidation, "mime_mismatch", "declared MIME type conflicts with the sniffed content type")
	}
	if !cl.allowsMIME(mime) {
		return uuid.Nil, kerr.E(kerr.KindValidation, "mime_not_allowed", "MIME type not permitted for this document class: "+mime)
	}

	if err := s.hooks.runUpload(ctx, UploadEvent{
		DocumentID: in.DocumentID.String(), Class: class, VersionNo: in.VersionNo,
		StorageKey: in.StorageKey, MIME: mime, SizeBytes: info.Size, Sensitivity: Sensitivity(sens),
	}); err != nil {
		return uuid.Nil, err
	}

	// CAS confirm the durable session, predicated on the COMPLETE reserved
	// identity — session id, pending status, document, version, storage key,
	// checksum — AND the validity window (adversarial review 2026-07-17, F-05:
	// the old predicate covered only id+pending+checksum, so document A's
	// session could attach its content to document B, and an expired-but-
	// unswept session remained confirmable). The RETURNING values are the
	// authoritative identity used for every subsequent effect.
	var confirmed struct {
		documentID uuid.UUID
		versionNo  int
		storageKey string
	}
	checkSum := strings.ToLower(in.DeclaredChecksum)
	err = db.QueryRow(ctx,
		`UPDATE document_upload_sessions
		    SET status = 'confirmed',
		        mime_type = $2,
		        size_bytes = $3
		  WHERE id = $1 AND status = 'pending' AND checksum_sha256 = $4
		    AND document_id = $5 AND version_no = $6 AND storage_key = $7
		    AND expires_at > now()
		 RETURNING document_id, version_no, storage_key`,
		in.SessionID, mime, info.Size, checkSum, in.DocumentID, in.VersionNo, in.StorageKey).
		Scan(&confirmed.documentID, &confirmed.versionNo, &confirmed.storageKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, kerr.E(kerr.KindConflict, "session_settled",
			"session already settled, expired, or does not match the reserved document/version/key")
	}
	if err != nil {
		return uuid.Nil, kerr.Wrapf(err, "document.ConfirmUpload", "confirm session")
	}

	verID := s.idgen.New()
	actor := actorFromCtx(ctx)
	_, err = db.Exec(ctx,
		`INSERT INTO document_versions
		    (id, tenant_id, document_id, version_no, storage_key, mime_type, size_bytes, checksum_sha256, scan_status, uploaded_by)
		 VALUES ($1, app_tenant_id(), $2, $3, $4, $5, $6, $7, 'pending', $8)`,
		verID, confirmed.documentID, confirmed.versionNo, confirmed.storageKey, mime, info.Size, checkSum, actor)
	if isUniqueViolation(err) {
		return uuid.Nil, kerr.E(kerr.KindConflict, "version_exists", "that version number is already confirmed")
	}
	if err != nil {
		return uuid.Nil, kerr.Wrapf(err, "document.ConfirmUpload", "insert version")
	}
	if _, err := db.Exec(ctx,
		`UPDATE documents SET version = $2, updated_at = now(), updated_by = $3 WHERE id = $1`,
		confirmed.documentID, confirmed.versionNo, actor); err != nil {
		return uuid.Nil, kerr.Wrapf(err, "document.ConfirmUpload", "bump document version")
	}
	if err := s.emit(ctx, db, "document.version_added", docRef(confirmed.documentID), map[string]any{"version_no": confirmed.versionNo, "mime": mime}); err != nil {
		return uuid.Nil, err
	}
	return verID, nil
}

// Download authorizes and returns a short-lived presigned GET for a version.
func (s *Service) Download(ctx context.Context, db database.TenantDB, actor authz.Actor, in DownloadInput) (Download, error) {
	var (
		sens   string
		status string
		owner  uuid.UUID
	)
	err := db.QueryRow(ctx, `SELECT sensitivity, status, created_by FROM documents WHERE id = $1`, in.DocumentID).
		Scan(&sens, &status, &owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return Download{}, kerr.E(kerr.KindNotFound, "not_found", "document not found")
	}
	if err != nil {
		return Download{}, kerr.Wrapf(err, "document.Download", "load document")
	}
	if status != "active" {
		return Download{}, kerr.E(kerr.KindNotFound, "not_found", "document is no longer available")
	}

	// Authorize BEFORE touching version rows so an unauthorized caller cannot
	// probe version existence/void state via differing errors (SEC-47).
	if err := s.authorizeRead(ctx, db, actor, in.DocumentID, owner); err != nil {
		return Download{}, err
	}

	// Resolve the version — always filter to active, in both the explicit and
	// latest paths, so a voided version is indistinguishable from a missing one.
	var (
		verNo      int
		storageKey string
		mime       string
		scan       string
	)
	q := `SELECT version_no, storage_key, mime_type, scan_status FROM document_versions WHERE document_id = $1 AND status = 'active'`
	args := []any{in.DocumentID}
	if in.VersionNo > 0 {
		q += ` AND version_no = $2`
		args = append(args, in.VersionNo)
	} else {
		q += ` ORDER BY version_no DESC LIMIT 1`
	}
	err = db.QueryRow(ctx, q, args...).Scan(&verNo, &storageKey, &mime, &scan)
	if errors.Is(err, pgx.ErrNoRows) {
		return Download{}, kerr.E(kerr.KindNotFound, "not_found", "no such document version")
	}
	if err != nil {
		return Download{}, kerr.Wrapf(err, "document.Download", "load version")
	}

	// Scan gate (blueprint 07 §4): infected never serves; a pending scan blocks
	// confidential-and-above.
	switch scan {
	case "infected":
		return Download{}, kerr.E(kerr.KindConflict, "scan_infected", "this file failed a malware scan")
	case "pending":
		if Sensitivity(sens).atLeast(SensitivityConfidential) {
			return Download{}, kerr.E(kerr.KindConflict, "scan_pending", "download blocked until the malware scan completes")
		}
	}

	if err := s.hooks.runAccess(ctx, AccessEvent{
		DocumentID: in.DocumentID.String(), VersionNo: verNo, Sensitivity: Sensitivity(sens), ActorID: actorID(actor),
	}); err != nil {
		return Download{}, err
	}

	url, err := s.store.PresignGet(ctx, storageKey, s.getTTL)
	if err != nil {
		return Download{}, kerr.Wrapf(err, "document.Download", "presign get")
	}
	// Download is a READ: it must run inside a read-only tenant tx (the natural
	// home for a "get download URL" handler). We deliberately do NOT emit an
	// outbox event here — that INSERT would fail a WithTenantRO tx (ARCH-65). The
	// durable `document.downloaded` audit row lands via the audit_logs writer (a
	// later phase); until then a download is recorded by the access-hook slot.
	return Download{VersionNo: verNo, MIME: mime, URL: url}, nil
}

// Grant records an explicit access grant. The caller must be able to write the
// document (owner, a write grant, or a policy write-allow).
func (s *Service) Grant(ctx context.Context, db database.TenantDB, in GrantInput) (uuid.UUID, error) {
	if !validGranteeKind(in.GranteeKind) {
		return uuid.Nil, kerr.E(kerr.KindValidation, "grant_invalid", "invalid grantee kind: "+in.GranteeKind)
	}
	if in.Access != "read" && in.Access != "write" {
		return uuid.Nil, kerr.E(kerr.KindValidation, "grant_invalid", "access must be read or write")
	}
	if strings.TrimSpace(in.GranteeRef) == "" {
		return uuid.Nil, kerr.E(kerr.KindValidation, "grant_invalid", "grantee ref is required")
	}
	var owner uuid.UUID
	err := db.QueryRow(ctx, `SELECT created_by FROM documents WHERE id = $1`, in.DocumentID).Scan(&owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, kerr.E(kerr.KindNotFound, "not_found", "document not found")
	}
	if err != nil {
		return uuid.Nil, kerr.Wrapf(err, "document.Grant", "load document")
	}
	if err := s.authorizeWrite(ctx, db, actorAuthz(ctx), in.DocumentID, owner); err != nil {
		return uuid.Nil, err
	}
	id := s.idgen.New()
	_, err = db.Exec(ctx,
		`INSERT INTO document_access_grants
		    (id, tenant_id, document_id, grantee_kind, grantee_ref, access, valid_to, created_by)
		 VALUES ($1, app_tenant_id(), $2, $3, $4, $5, $6, $7)`,
		id, in.DocumentID, in.GranteeKind, in.GranteeRef, in.Access, in.ValidTo, actorFromCtx(ctx))
	if err != nil {
		return uuid.Nil, kerr.Wrapf(err, "document.Grant", "insert grant")
	}
	if err := s.emit(ctx, db, "document.access_granted", docRef(in.DocumentID),
		map[string]any{"grantee_kind": in.GranteeKind, "access": in.Access}); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// Revoke closes a grant's validity window (soft revoke; grants are auditable).
// The caller must be able to write the grant's document — the same gate as Grant
// (SEC-43); the restrictive grant-ownership RLS policy is the DB-level backstop.
func (s *Service) Revoke(ctx context.Context, db database.TenantDB, grantID uuid.UUID) error {
	var (
		docID uuid.UUID
		owner uuid.UUID
	)
	err := db.QueryRow(ctx,
		`SELECT g.document_id, d.created_by
		   FROM document_access_grants g JOIN documents d ON d.id = g.document_id
		  WHERE g.id = $1`, grantID).Scan(&docID, &owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return kerr.E(kerr.KindNotFound, "not_found", "no grant with that id")
	}
	if err != nil {
		return kerr.Wrapf(err, "document.Revoke", "load grant")
	}
	if err := s.authorizeWrite(ctx, db, actorAuthz(ctx), docID, owner); err != nil {
		return err
	}
	ct, err := db.Exec(ctx,
		`UPDATE document_access_grants SET valid_to = now() WHERE id = $1 AND (valid_to IS NULL OR valid_to > now())`, grantID)
	if err != nil {
		return kerr.Wrapf(err, "document.Revoke", "revoke grant")
	}
	if ct.RowsAffected() == 0 {
		return kerr.E(kerr.KindNotFound, "not_found", "no active grant with that id")
	}
	return nil
}

// --- platform-privileged operations (run on a tenant-bound app_platform tx) ---

// UpdateScanStatus records the result of an async malware scan. Runs as
// app_platform (document_versions is append-only to the module role); a scan may
// only leave the pending state once.
func (s *Service) UpdateScanStatus(ctx context.Context, plat database.TxManager, tenantID, versionID uuid.UUID, result string) error {
	switch result {
	case "clean", "infected", "skipped":
	default:
		return kerr.E(kerr.KindValidation, "scan_invalid", "scan result must be clean, infected, or skipped")
	}
	return plat.WithTenant(database.WithTenantID(ctx, tenantID), func(ctx context.Context, db database.TenantDB) error {
		ct, err := db.Exec(ctx,
			`UPDATE document_versions SET scan_status = $2 WHERE id = $1 AND scan_status = 'pending'`, versionID, result)
		if err != nil {
			return kerr.Wrapf(err, "document.UpdateScanStatus", "update scan status")
		}
		if ct.RowsAffected() == 0 {
			return kerr.E(kerr.KindConflict, "scan_settled", "version is absent or its scan already settled")
		}
		return nil
	})
}

// SweepRetention voids every active version of every document whose retention has
// lapsed (legal-hold documents are skipped) and tombstones the rows, then deletes
// the blobs. Runs as app_platform, tenant-bound. The row voiding commits FIRST;
// blob deletion happens only AFTER the commit succeeds (SEC-48) — so a mid-sweep
// failure never leaves an active row pointing at a deleted blob. A blob-delete
// failure after commit merely orphans the blob (swept by a future storage GC).
// Idempotent: a re-run over an already-swept tenant voids nothing. Returns the
// number of versions voided.
func (s *Service) SweepRetention(ctx context.Context, plat database.TxManager, tenantID uuid.UUID, at time.Time) (int, error) {
	var toDelete []string
	err := plat.WithTenant(database.WithTenantID(ctx, tenantID), func(ctx context.Context, db database.TenantDB) error {
		toDelete = toDelete[:0] // reset if the tx retries
		// FOR UPDATE locks each candidate row and, under READ COMMITTED, re-checks
		// the WHERE against the latest committed tuple after acquiring the lock
		// (EvalPlanQual). So a legal hold applied and committed before we lock
		// excludes the document; a hold attempted after we lock blocks until this
		// sweep finishes — closing the check-then-void race (roadmap R6). ORDER BY
		// id gives a stable lock order across concurrent runs.
		rows, err := db.Query(ctx,
			`SELECT id FROM documents
			  WHERE status = 'active' AND legal_hold = false
			    AND retention_until IS NOT NULL AND retention_until <= $1
			  ORDER BY id
			  FOR UPDATE`, at)
		if err != nil {
			return kerr.Wrapf(err, "document.SweepRetention", "find expired")
		}
		var docIDs []uuid.UUID
		for rows.Next() {
			var id uuid.UUID
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return kerr.Wrapf(err, "document.SweepRetention", "scan expired")
			}
			docIDs = append(docIDs, id)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return kerr.Wrapf(err, "document.SweepRetention", "iterate expired")
		}

		for _, docID := range docIDs {
			vrows, err := db.Query(ctx,
				`SELECT id, storage_key FROM document_versions WHERE document_id = $1 AND status = 'active'`, docID)
			if err != nil {
				return kerr.Wrapf(err, "document.SweepRetention", "load versions")
			}
			type ver struct {
				id  uuid.UUID
				key string
			}
			var vers []ver
			for vrows.Next() {
				var v ver
				if err := vrows.Scan(&v.id, &v.key); err != nil {
					vrows.Close()
					return kerr.Wrapf(err, "document.SweepRetention", "scan version")
				}
				vers = append(vers, v)
			}
			vrows.Close()
			if err := vrows.Err(); err != nil {
				return kerr.Wrapf(err, "document.SweepRetention", "iterate versions")
			}
			for _, v := range vers {
				if _, err := db.Exec(ctx,
					`UPDATE document_versions SET status = 'voided', voided_at = $2 WHERE id = $1`, v.id, at); err != nil {
					return kerr.Wrapf(err, "document.SweepRetention", "void version")
				}
				toDelete = append(toDelete, v.key)
			}
			// Re-assert legal_hold = false in the write itself: defense in depth
			// behind the FOR UPDATE lock, so a hold can never be voided even under
			// a stricter isolation level or a future refactor of the lock above.
			if _, err := db.Exec(ctx,
				`UPDATE documents SET status = 'voided', updated_at = $2
				  WHERE id = $1 AND legal_hold = false`, docID, at); err != nil {
				return kerr.Wrapf(err, "document.SweepRetention", "void document")
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	// Post-commit: the tombstones are durable; now free the blobs. A failure here
	// leaves a voided row + an orphaned blob (safe) rather than the reverse.
	for _, key := range toDelete {
		if derr := s.store.Delete(ctx, key); derr != nil {
			return len(toDelete), kerr.Wrapf(derr, "document.SweepRetention", "delete blob post-commit")
		}
	}
	return len(toDelete), nil
}

// SweepUploadSessions garbage-collects pending upload sessions whose expiry has
// passed, deleting the orphaned storage object and marking the session expired.
// Runs as app_platform, tenant-bound. Returns the number of sessions swept.
func (s *Service) SweepUploadSessions(ctx context.Context, plat database.TxManager, tenantID uuid.UUID, before time.Time) (int, error) {
	var toDelete []string
	swept := 0
	err := plat.WithTenant(database.WithTenantID(ctx, tenantID), func(ctx context.Context, db database.TenantDB) error {
		toDelete = toDelete[:0]
		swept = 0
		rows, err := db.Query(ctx,
			`SELECT id, storage_key FROM document_upload_sessions
			  WHERE status = 'pending' AND expires_at <= $1
			  ORDER BY id FOR UPDATE`, before)
		if err != nil {
			return kerr.Wrapf(err, "document.SweepUploadSessions", "find expired sessions")
		}
		type sess struct {
			id  uuid.UUID
			key string
		}
		var sessions []sess
		for rows.Next() {
			var r sess
			if err := rows.Scan(&r.id, &r.key); err != nil {
				rows.Close()
				return kerr.Wrapf(err, "document.SweepUploadSessions", "scan session")
			}
			sessions = append(sessions, r)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return kerr.Wrapf(err, "document.SweepUploadSessions", "iterate sessions")
		}
		for _, r := range sessions {
			if _, err := db.Exec(ctx,
				`UPDATE document_upload_sessions SET status = 'expired' WHERE id = $1 AND status = 'pending'`,
				r.id); err != nil {
				return kerr.Wrapf(err, "document.SweepUploadSessions", "expire session")
			}
			toDelete = append(toDelete, r.key)
			swept++
			if s.audit != nil {
				if err := s.audit.Record(ctx, db, kaudit.Entry{
					Action:     "document.upload_session_expired",
					EntityType: "document_upload_session",
					EntityID:   r.id,
					Reason:     "upload session expired before confirmation",
				}); err != nil {
					return kerr.Wrapf(err, "document.SweepUploadSessions", "record audit")
				}
			}
		}
		return nil
	})
	if err != nil {
		return swept, err
	}
	for _, key := range toDelete {
		if derr := s.store.Delete(ctx, key); derr != nil {
			return swept, kerr.Wrapf(derr, "document.SweepUploadSessions", "delete blob post-commit")
		}
	}
	return swept, nil
}

// --- authorization ---

// authorizeRead allows a download when: an explicit deny policy does NOT block
// (deny is authoritative), AND one of — the actor owns the document, an authz
// role/policy allows read, or a valid explicit grant permits read.
func (s *Service) authorizeRead(ctx context.Context, db database.TenantDB, actor authz.Actor, docID, owner uuid.UUID) error {
	return s.authorize(ctx, db, actor, docID, owner, PermRead, "read")
}

// authorizeWrite is authorizeRead for the write permission/access.
func (s *Service) authorizeWrite(ctx context.Context, db database.TenantDB, actor authz.Actor, docID, owner uuid.UUID) error {
	return s.authorize(ctx, db, actor, docID, owner, PermWrite, "write")
}

func (s *Service) authorize(ctx context.Context, db database.TenantDB, actor authz.Actor, docID, owner uuid.UUID, perm, access string) error {
	// 1. Evaluator (if wired + permission registered). An explicit deny policy is
	//    authoritative and cannot be overridden by ownership or a grant.
	if s.authz != nil {
		d, err := s.authz.Evaluate(ctx, db, actor, perm, authz.Target{Scope: authz.ScopeResource, Resource: docRef(docID)})
		switch {
		case err == nil && d.Allowed:
			return nil
		case err == nil && strings.HasPrefix(d.Reason, "policy:"):
			return kerr.E(kerr.KindForbidden, "forbidden", "access denied by policy")
		case err != nil && kerr.KindOf(err) != kerr.KindInternal:
			// A real evaluation failure (not a mere unregistered-permission wiring
			// gap) must not be swallowed.
			return err
		}
	}
	// 2. Owner-implicit.
	if owner != uuid.Nil && (owner == actor.CapacityID || owner == actor.UserID) {
		return nil
	}
	// 3. Explicit capacity grant within its validity window.
	if actor.CapacityID != uuid.Nil {
		ok, err := s.hasCapacityGrant(ctx, db, docID, actor.CapacityID, access)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
	return kerr.E(kerr.KindForbidden, "forbidden", "not authorized for this document")
}

func (s *Service) hasCapacityGrant(ctx context.Context, db database.TenantDB, docID, capacityID uuid.UUID, access string) (bool, error) {
	// write access satisfies a read request; read does not satisfy write.
	accepted := "('write')"
	if access == "read" {
		accepted = "('read','write')"
	}
	var one int
	err := db.QueryRow(ctx,
		`SELECT 1 FROM document_access_grants
		  WHERE document_id = $1 AND grantee_kind = 'capacity' AND grantee_ref = $2
		    AND access IN `+accepted+`
		    AND valid_from <= now() AND (valid_to IS NULL OR valid_to > now())
		  LIMIT 1`, docID, capacityID.String()).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, kerr.Wrapf(err, "document.authorize", "check grant")
	}
	return one == 1, nil
}

// --- helpers ---

// emit writes an event into the same tenant tx as the business change, so the
// two commit or roll back together. Its error is propagated (never dropped): a
// document op that cannot record its event fails as a whole.
func (s *Service) emit(ctx context.Context, db database.TenantDB, typ string, ref resource.Ref, payload any) error {
	if err := s.outbox.Write(ctx, db, outbox.Event{Type: typ, Resource: ref, Payload: payload}); err != nil {
		return kerr.Wrapf(err, "document.emit", "write %s event", typ)
	}
	return nil
}

func actorFromCtx(ctx context.Context) uuid.UUID {
	if id, ok := database.ActorIDFrom(ctx); ok {
		return id
	}
	return uuid.Nil
}

// actorAuthz builds a minimal Actor from context for internal write checks where
// the caller did not pass one explicitly (Grant). CapacityID is the context actor.
func actorAuthz(ctx context.Context) authz.Actor {
	id := actorFromCtx(ctx)
	tid, _ := database.TenantIDFrom(ctx)
	return authz.Actor{CapacityID: id, UserID: id, TenantID: tid}
}

func actorID(a authz.Actor) string {
	if a.CapacityID != uuid.Nil {
		return a.CapacityID.String()
	}
	return a.UserID.String()
}

func docRef(id uuid.UUID) resource.Ref { return resource.Ref{Type: docResourceType, ID: id} }

func validGranteeKind(k string) bool {
	return k == "capacity" || k == "role" || k == "relationship"
}

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

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}

// mimeConflict reports whether a declared MIME type conflicts with the sniffed
// one. http.DetectContentType returns application/octet-stream for content it
// cannot classify; in that case we trust the declared type. When the sniff IS
// confident, the declared full type/subtype must agree with it (parameters like
// "; charset=utf-8" are ignored) — this blocks text/html declared as text/plain
// (ARCH-69) as well as text/html declared as image/png.
func mimeConflict(declared, sniffed string) bool {
	if declared == "" || sniffed == "application/octet-stream" {
		return false
	}
	return essence(declared) != essence(sniffed)
}

// essence returns the type/subtype with any parameters and surrounding space
// stripped ("text/html; charset=utf-8" → "text/html").
func essence(mime string) string {
	base, _, _ := strings.Cut(mime, ";")
	return strings.TrimSpace(base)
}
