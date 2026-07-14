package aggregate

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/testkit"
)

// The suite runs the aggregate write contract against a real Postgres: a
// fixture business table (aggtest_thing) stands in for a module's table, RLS
// and grants matching the module convention. Fault injection substitutes a
// decorated stage that performs the REAL write, verifies it is visible inside
// the transaction, and then fails — proving the later rollback undid rows
// that were genuinely there, not writes that never happened.

const fixtureType = "aggtest.thing"

var errInjected = errors.New("injected stage fault")

type fixture struct {
	h      *testkit.DBHandle
	tenant uuid.UUID
	userID uuid.UUID
	capID  uuid.UUID
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h).ID
	userID := testkit.CreateUser(t, h)
	capID := testkit.CreateCapacity(t, h, tenant, userID)
	testkit.CreateResourceType(t, h, fixtureType)

	ctx := context.Background()
	for _, stmt := range []string{
		`CREATE TABLE aggtest_thing (
		     id         uuid PRIMARY KEY,
		     tenant_id  uuid NOT NULL,
		     title      text NOT NULL,
		     created_by uuid NOT NULL
		 )`,
		`ALTER TABLE aggtest_thing ENABLE ROW LEVEL SECURITY`,
		`ALTER TABLE aggtest_thing FORCE ROW LEVEL SECURITY`,
		`CREATE POLICY aggtest_thing_tenant ON aggtest_thing
		     USING      (tenant_id = app_tenant_id())
		     WITH CHECK (tenant_id = app_tenant_id())`,
		`GRANT SELECT, INSERT, UPDATE ON aggtest_thing TO app_rt`,
	} {
		if _, err := h.Admin.Exec(ctx, stmt); err != nil {
			t.Fatalf("fixture DDL: %v\n%s", err, stmt)
		}
	}
	return &fixture{h: h, tenant: tenant, userID: userID, capID: capID}
}

// userCtx mirrors what the authz gate binds for an authenticated user:
// tenant, the full principal, and the audit actor (acting capacity).
func (f *fixture) userCtx() context.Context {
	ctx := testkit.TenantCtx(f.tenant)
	ctx = httpx.WithActor(ctx, authz.Actor{
		Kind: authz.ActorUser, TenantID: f.tenant, UserID: f.userID, CapacityID: f.capID,
	})
	return database.WithActorID(ctx, f.capID)
}

func (f *fixture) writer() *Writer {
	return New(f.h.TxM, resource.NewRegistrar(), audit.New(model.UUIDv7(), nil), outbox.NewWriter(model.UUIDv7()))
}

// input is a fully-populated aggregate write for the fixture table.
// applyErr, when non-nil, is returned by Apply AFTER its real insert —
// the stage-1 fault point.
func (f *fixture) input(id uuid.UUID, applyErr error, inTxSeen *bool) Write {
	return Write{
		Resource: resource.Ref{Type: fixtureType, ID: id},
		Label:    "Thing",
		Status:   "open",
		Audit:    audit.Entry{Action: "aggtest.thing.create"},
		Event:    outbox.Event{Type: "aggtest.thing.created", Payload: map[string]string{"title": "Thing"}},
		Apply: func(ctx context.Context, db database.TenantDB, actorID uuid.UUID) error {
			if _, err := db.Exec(ctx,
				`INSERT INTO aggtest_thing (id, tenant_id, title, created_by)
				 VALUES ($1, app_tenant_id(), $2, $3)`, id, "Thing", actorID); err != nil {
				return err
			}
			if inTxSeen != nil {
				*inTxSeen = rowVisible(ctx, db, `SELECT count(*) FROM aggtest_thing WHERE id = $1`, id)
			}
			return applyErr
		},
	}
}

// rowVisible reports whether the query counts exactly one row through db —
// used INSIDE the transaction to prove a stage's write really happened
// before the injected fault rolled it back.
func rowVisible(ctx context.Context, db database.TenantDB, q string, args ...any) bool {
	var n int
	if err := db.QueryRow(ctx, q, args...).Scan(&n); err != nil {
		return false
	}
	return n == 1
}

// counts reads the four stage tables through the admin pool (bypassing RLS),
// so a row leaked under any tenant is caught.
func (f *fixture) counts(t *testing.T, id uuid.UUID) (business, mirror, auditRows, events int) {
	t.Helper()
	ctx := context.Background()
	for _, c := range []struct {
		q    string
		dest *int
	}{
		{`SELECT count(*) FROM aggtest_thing WHERE id = $1`, &business},
		{`SELECT count(*) FROM resources WHERE id = $1`, &mirror},
		{`SELECT count(*) FROM audit_logs WHERE entity_id = $1`, &auditRows},
		{`SELECT count(*) FROM events_outbox WHERE resource_id = $1`, &events},
	} {
		if err := f.h.Admin.QueryRow(ctx, c.q, id).Scan(c.dest); err != nil {
			t.Fatalf("count %q: %v", c.q, err)
		}
	}
	return business, mirror, auditRows, events
}

// --- stage fault decorators: real write first, then the injected fault ---

type faultMirror struct {
	real   mirrorUpserter
	inTx   *bool
	broken bool
}

func (m *faultMirror) UpsertAs(ctx context.Context, db database.TenantDB, actorID uuid.UUID,
	ref resource.Ref, orgID *uuid.UUID, label, status string,
) error {
	if err := m.real.UpsertAs(ctx, db, actorID, ref, orgID, label, status); err != nil {
		return err
	}
	*m.inTx = rowVisible(ctx, db, `SELECT count(*) FROM resources WHERE id = $1`, ref.ID)
	if m.broken {
		return errInjected
	}
	return nil
}

type faultAudit struct {
	real   auditRecorder
	inTx   *bool
	entity uuid.UUID
	broken bool
}

func (a *faultAudit) Record(ctx context.Context, db database.TenantDB, e audit.Entry) error {
	if err := a.real.Record(ctx, db, e); err != nil {
		return err
	}
	*a.inTx = rowVisible(ctx, db, `SELECT count(*) FROM audit_logs WHERE entity_id = $1`, a.entity)
	if a.broken {
		return errInjected
	}
	return nil
}

type faultOutbox struct {
	real   outbox.Writer
	inTx   *bool
	res    uuid.UUID
	broken bool
}

func (o *faultOutbox) Write(ctx context.Context, db database.TenantDB, e outbox.Event) error {
	if err := o.real.Write(ctx, db, e); err != nil {
		return err
	}
	*o.inTx = rowVisible(ctx, db, `SELECT count(*) FROM events_outbox WHERE resource_id = $1`, o.res)
	if o.broken {
		return errInjected
	}
	return nil
}

// TestIntegrationAggregateWriteCommitsAllFourStages proves the contract's
// happy path: one call produces the business row, the resources mirror, the
// audit row, and the outbox event — all attributed to the real acting
// capacity, none to a placeholder (AC-W02-E04-S001-01/-02).
func TestIntegrationAggregateWriteCommitsAllFourStages(t *testing.T) {
	f := newFixture(t)
	id := uuid.New()

	if err := f.writer().Write(f.userCtx(), f.input(id, nil, nil)); err != nil {
		t.Fatalf("aggregate write: %v", err)
	}

	business, mirror, auditRows, events := f.counts(t, id)
	if business != 1 || mirror != 1 || auditRows != 1 || events != 1 {
		t.Fatalf("stage rows = business:%d mirror:%d audit:%d outbox:%d, want 1/1/1/1",
			business, mirror, auditRows, events)
	}

	ctx := context.Background()
	var businessBy, mirrorBy, auditBy uuid.UUID
	var actorKind, eventActor string
	if err := f.h.Admin.QueryRow(ctx, `SELECT created_by FROM aggtest_thing WHERE id = $1`, id).Scan(&businessBy); err != nil {
		t.Fatalf("read business row: %v", err)
	}
	if err := f.h.Admin.QueryRow(ctx, `SELECT created_by FROM resources WHERE id = $1`, id).Scan(&mirrorBy); err != nil {
		t.Fatalf("read mirror row: %v", err)
	}
	if err := f.h.Admin.QueryRow(ctx,
		`SELECT actor_id, actor_kind FROM audit_logs WHERE entity_id = $1`, id).Scan(&auditBy, &actorKind); err != nil {
		t.Fatalf("read audit row: %v", err)
	}
	if err := f.h.Admin.QueryRow(ctx,
		`SELECT actor::text FROM events_outbox WHERE resource_id = $1`, id).Scan(&eventActor); err != nil {
		t.Fatalf("read outbox row: %v", err)
	}
	if businessBy != f.capID || mirrorBy != f.capID || auditBy != f.capID {
		t.Fatalf("created_by attribution business:%s mirror:%s audit:%s, want %s everywhere",
			businessBy, mirrorBy, auditBy, f.capID)
	}
	if businessBy == uuid.Nil {
		t.Fatal("attribution must never be the nil placeholder")
	}
	if actorKind != string(authz.ActorUser) {
		t.Fatalf("audit actor_kind = %q, want %q", actorKind, authz.ActorUser)
	}
	if !strings.Contains(eventActor, f.capID.String()) {
		t.Fatalf("outbox actor descriptor %q does not name the acting capacity %s", eventActor, f.capID)
	}
}

// TestIntegrationAggregateWriteFaultInjection injects a fault at each of the
// 4 stages independently — business write, mirror upsert, audit row, outbox
// event — after the stage's REAL database write (verified visible inside the
// transaction), and proves the entire transaction rolls back every time:
// zero rows in all four tables afterwards (AC-W02-E04-S001-01).
func TestIntegrationAggregateWriteFaultInjection(t *testing.T) {
	f := newFixture(t)

	stages := []struct {
		name string
		run  func(w *Writer, id uuid.UUID, seen *bool) error
	}{
		{"stage1-business-write", func(w *Writer, id uuid.UUID, seen *bool) error {
			return w.Write(f.userCtx(), f.input(id, errInjected, seen))
		}},
		{"stage2-mirror-upsert", func(w *Writer, id uuid.UUID, seen *bool) error {
			w.mirror = &faultMirror{real: w.mirror, inTx: seen, broken: true}
			return w.Write(f.userCtx(), f.input(id, nil, nil))
		}},
		{"stage3-audit-row", func(w *Writer, id uuid.UUID, seen *bool) error {
			w.audit = &faultAudit{real: w.audit, inTx: seen, entity: id, broken: true}
			return w.Write(f.userCtx(), f.input(id, nil, nil))
		}},
		{"stage4-outbox-event", func(w *Writer, id uuid.UUID, seen *bool) error {
			w.outbox = &faultOutbox{real: w.outbox, inTx: seen, res: id, broken: true}
			return w.Write(f.userCtx(), f.input(id, nil, nil))
		}},
	}

	for _, st := range stages {
		t.Run(st.name, func(t *testing.T) {
			id := uuid.New()
			seen := false
			err := st.run(f.writer(), id, &seen)
			if !errors.Is(err, errInjected) {
				t.Fatalf("fault at %s: error = %v, want the injected fault", st.name, err)
			}
			if !seen {
				t.Fatalf("fault at %s: stage write was not visible inside the transaction — the fault fired before the real write, weakening the rollback proof", st.name)
			}
			business, mirror, auditRows, events := f.counts(t, id)
			if business != 0 || mirror != 0 || auditRows != 0 || events != 0 {
				t.Fatalf("fault at %s leaked rows: business:%d mirror:%d audit:%d outbox:%d, want full rollback (0/0/0/0)",
					st.name, business, mirror, auditRows, events)
			}
		})
	}
}

// TestIntegrationAggregateWriteUserWithoutActorFailsFast proves a
// user-initiated write with no resolvable actor is rejected before any
// database work — never silently attributed to a placeholder
// (AC-W02-E04-S001-02).
func TestIntegrationAggregateWriteUserWithoutActorFailsFast(t *testing.T) {
	f := newFixture(t)
	id := uuid.New()

	// A user principal with no bound audit actor, no capacity, no user id.
	ctx := httpx.WithActor(testkit.TenantCtx(f.tenant), authz.Actor{Kind: authz.ActorUser, TenantID: f.tenant})

	err := f.writer().Write(ctx, f.input(id, nil, nil))
	if err == nil {
		t.Fatal("user-initiated write without an actor must fail fast")
	}
	if kerr.KindOf(err) != kerr.KindUnauthenticated {
		t.Fatalf("error kind = %v, want KindUnauthenticated: %v", kerr.KindOf(err), err)
	}
	business, mirror, auditRows, events := f.counts(t, id)
	if business != 0 || mirror != 0 || auditRows != 0 || events != 0 {
		t.Fatalf("rejected write leaked rows: business:%d mirror:%d audit:%d outbox:%d", business, mirror, auditRows, events)
	}
}

// TestIntegrationAggregateWriteSystemActorPathsSucceed proves system-initiated
// paths are unaffected by the user-actor requirement (AC-W02-E04-S001-02):
// a bare tenant context (how job runners and the outbox relay call, binding
// no principal) and a named machine principal both succeed, attributed to
// their deterministic system-actor ids.
func TestIntegrationAggregateWriteSystemActorPathsSucceed(t *testing.T) {
	f := newFixture(t)

	cases := []struct {
		name string
		ctx  context.Context
		want uuid.UUID
	}{
		{"background-no-principal", testkit.TenantCtx(f.tenant), SystemActorID("")},
		{
			"named-system-principal",
			httpx.WithActor(testkit.TenantCtx(f.tenant),
				authz.Actor{Kind: authz.ActorSystem, TenantID: f.tenant, System: "outbox-relay"}),
			SystemActorID("outbox-relay"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := uuid.New()
			if err := f.writer().Write(tc.ctx, f.input(id, nil, nil)); err != nil {
				t.Fatalf("system-actor write must succeed: %v", err)
			}
			var createdBy uuid.UUID
			var actorKind string
			if err := f.h.Admin.QueryRow(context.Background(),
				`SELECT created_by FROM resources WHERE id = $1`, id).Scan(&createdBy); err != nil {
				t.Fatalf("read mirror: %v", err)
			}
			if err := f.h.Admin.QueryRow(context.Background(),
				`SELECT actor_kind FROM audit_logs WHERE entity_id = $1`, id).Scan(&actorKind); err != nil {
				t.Fatalf("read audit: %v", err)
			}
			if createdBy != tc.want {
				t.Fatalf("created_by = %s, want deterministic system actor %s", createdBy, tc.want)
			}
			if createdBy == uuid.Nil {
				t.Fatal("system attribution must not be the nil placeholder")
			}
			if actorKind != string(authz.ActorSystem) {
				t.Fatalf("audit actor_kind = %q, want %q", actorKind, authz.ActorSystem)
			}
		})
	}
}

// TestSystemActorIDDeterministic pins the system-actor derivation: stable per
// name, distinct across names, never the nil placeholder.
func TestSystemActorIDDeterministic(t *testing.T) {
	first := SystemActorID("outbox-relay")
	second := SystemActorID("outbox-relay")
	if first != second {
		t.Fatal("SystemActorID must be deterministic per name")
	}
	if SystemActorID("") != SystemActorID("system") {
		t.Fatal("empty name must alias the generic system actor")
	}
	if SystemActorID("a") == SystemActorID("b") {
		t.Fatal("distinct system names must derive distinct ids")
	}
	if SystemActorID("") == uuid.Nil || SystemActorID("x") == uuid.Nil {
		t.Fatal("derived system-actor ids must never be nil")
	}
}

// TestAggregateWriteValidatesInput pins the contract's own preconditions:
// a write with no Apply, no resource ref, no audit action, or no event type
// is rejected before any actor resolution or database work.
func TestAggregateWriteValidatesInput(t *testing.T) {
	w := &Writer{} // never reaches any dependency
	base := Write{
		Resource: resource.Ref{Type: fixtureType, ID: uuid.New()},
		Audit:    audit.Entry{Action: "aggtest.thing.create"},
		Event:    outbox.Event{Type: "aggtest.thing.created"},
		Apply:    func(context.Context, database.TenantDB, uuid.UUID) error { return nil },
	}
	cases := []struct {
		name   string
		mutate func(*Write)
	}{
		{"missing-apply", func(in *Write) { in.Apply = nil }},
		{"missing-resource", func(in *Write) { in.Resource = resource.Ref{} }},
		{"missing-audit-action", func(in *Write) { in.Audit.Action = "" }},
		{"missing-event-type", func(in *Write) { in.Event.Type = "" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := base
			tc.mutate(&in)
			if err := w.Write(context.Background(), in); err == nil {
				t.Fatal("invalid aggregate write must be rejected")
			}
		})
	}
}
