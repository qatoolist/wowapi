package fakes

import (
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestClockAdvance(t *testing.T) {
	start := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	c := NewClock(start)

	if got := c.Now(); !got.Equal(start) {
		t.Fatalf("Now() = %v, want %v", got, start)
	}
	c.Advance(90 * time.Minute)
	want := start.Add(90 * time.Minute)
	if got := c.Now(); !got.Equal(want) {
		t.Fatalf("after Advance, Now() = %v, want %v", got, want)
	}

	// satisfies the ambient clock interface
	var _ interface{ Now() time.Time } = c
}

func TestIDGenDeterministic(t *testing.T) {
	a := NewIDGen(0x01)
	b := NewIDGen(0x01)
	for i := 0; i < 100; i++ {
		if a.New() != b.New() {
			t.Fatalf("two IDGens with the same seed diverged at %d", i)
		}
	}
}

func TestIDGenReadableLayout(t *testing.T) {
	g := NewIDGen(0x01)
	got := g.New()
	if want := uuid.MustParse("01000000-0000-7000-8000-000000000001"); got != want {
		t.Fatalf("first id = %s, want %s", got, want)
	}
	if v := got.Version(); v != 7 {
		t.Errorf("version = %d, want 7", v)
	}
	if got.Variant() != uuid.RFC4122 {
		t.Errorf("variant = %v, want RFC4122", got.Variant())
	}
}

func TestIDGenUniqueAndConcurrent(t *testing.T) {
	g := NewIDGen(0x02)
	const n = 1000
	var (
		mu   sync.Mutex
		seen = make(map[uuid.UUID]struct{}, n)
		wg   sync.WaitGroup
	)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := g.New()
			mu.Lock()
			seen[id] = struct{}{}
			mu.Unlock()
		}()
	}
	wg.Wait()
	if len(seen) != n {
		t.Fatalf("got %d unique ids, want %d", len(seen), n)
	}
}
