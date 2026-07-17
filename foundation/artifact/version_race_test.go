package artifact_test

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/foundation/artifact"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/testkit"
)

// TestIntegrationGenerateConcurrentVersionAllocation is DATA-05's concurrency
// bar for kernel/artifact.Generate (W02-E03-S001 T1+T5, AC-01/AC-05): N
// concurrent callers must be issued N unique, monotonic versions with zero
// unexpected conflicts. Before the locked-counter fix, Generate computed the
// next version via an inline MAX(version)+1 subselect, so overlapping
// transactions read the same maximum and collided on the (tenant,kind,version)
// unique index — the losing INSERT surfaced as a KindConflict "retry" error
// (fail-first evidence for this story). With the version_counters allocation
// the callers serialize on the counter row and every one commits a distinct
// version.
//
// The 24 callers are concurrent goroutines; the testkit runtime pool caps
// DB-side concurrency at 4 in-flight transactions, which is enough overlap to
// reproduce the MAX()+1 race and to measure counter-row lock wait
// (RISK-W02-E03-001) after the fix.
func TestIntegrationGenerateConcurrentVersionAllocation(t *testing.T) {
	const callers = 24

	h := testkit.NewDB(t)
	p := artifact.New(model.UUIDv7())
	ctx := actx(uuid.New())

	var (
		start    = make(chan struct{})
		wg       sync.WaitGroup
		mu       sync.Mutex
		versions []int
		errs     []error
		waits    []time.Duration
	)
	wg.Add(callers)
	for range callers {
		go func() {
			defer wg.Done()
			<-start
			began := time.Now()
			var v int
			err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
				a, e := p.Generate(ctx, db, artifact.Input{Kind: "receipt", Content: []byte("x")})
				v = a.Version
				return e
			})
			took := time.Since(began)
			mu.Lock()
			defer mu.Unlock()
			waits = append(waits, took)
			if err != nil {
				errs = append(errs, err)
				return
			}
			versions = append(versions, v)
		}()
	}
	close(start)
	wg.Wait()

	// Zero unexpected conflicts: every caller must succeed.
	for _, err := range errs {
		t.Errorf("concurrent Generate failed: %v", err)
	}
	if len(errs) > 0 {
		t.Fatalf("%d of %d concurrent callers failed — version allocation is not race-free", len(errs), callers)
	}

	// N callers → N unique monotonic versions: exactly the set 1..callers.
	sort.Ints(versions)
	if len(versions) != callers {
		t.Fatalf("got %d versions, want %d", len(versions), callers)
	}
	for i, v := range versions {
		if v != i+1 {
			t.Fatalf("versions not the contiguous set 1..%d: got %v", callers, versions)
		}
	}

	// Lock-wait measurement (RISK-W02-E03-001): the counter row serializes
	// allocation, so per-caller wall time IS the contention signal.
	var maxW, sum time.Duration
	for _, w := range waits {
		if w > maxW {
			maxW = w
		}
		sum += w
	}
	t.Logf("lock-wait under %d concurrent callers: max=%v avg=%v", callers, maxW, sum/time.Duration(len(waits)))
}
