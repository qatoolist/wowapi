package model_test

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/qatoolist/wowapi/kernel/model"
)

// ---------------------------------------------------------------------------
// Temporal.ActiveAt
// ---------------------------------------------------------------------------

func TestTemporalActiveAt(t *testing.T) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		temporal   model.Temporal
		at         time.Time
		wantActive bool
	}{
		{
			name:       "open window: before ValidFrom",
			temporal:   model.Temporal{ValidFrom: t1, ValidTo: nil},
			at:         t0,
			wantActive: false,
		},
		{
			name:       "open window: at == ValidFrom (inclusive lower bound)",
			temporal:   model.Temporal{ValidFrom: t1, ValidTo: nil},
			at:         t1,
			wantActive: true,
		},
		{
			name:       "open window: after ValidFrom, no ValidTo",
			temporal:   model.Temporal{ValidFrom: t1, ValidTo: nil},
			at:         t2,
			wantActive: true,
		},
		{
			name:       "closed window: strictly inside",
			temporal:   model.Temporal{ValidFrom: t0, ValidTo: &t2},
			at:         t1,
			wantActive: true,
		},
		{
			name:       "closed window: at == ValidFrom (inclusive)",
			temporal:   model.Temporal{ValidFrom: t0, ValidTo: &t2},
			at:         t0,
			wantActive: true,
		},
		{
			name:       "closed window: at == ValidTo (exclusive upper bound)",
			temporal:   model.Temporal{ValidFrom: t0, ValidTo: &t2},
			at:         t2,
			wantActive: false,
		},
		{
			name:       "closed window: after ValidTo",
			temporal:   model.Temporal{ValidFrom: t0, ValidTo: &t1},
			at:         t2,
			wantActive: false,
		},
		{
			name:       "closed window: before ValidFrom",
			temporal:   model.Temporal{ValidFrom: t1, ValidTo: &t2},
			at:         t0,
			wantActive: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.temporal.ActiveAt(tc.at)
			if got != tc.wantActive {
				t.Errorf("ActiveAt(%v) = %v, want %v", tc.at, got, tc.wantActive)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// UUIDv7 generator: distinct + time-ordered
// ---------------------------------------------------------------------------

func TestUUIDv7Distinct(t *testing.T) {
	gen := model.UUIDv7()

	a := gen.New()
	b := gen.New()

	if a == b {
		t.Fatalf("UUIDv7 produced duplicate IDs: %v", a)
	}

	// UUIDv7 ids are lexicographically time-ordered when compared as [16]byte.
	// The first generated id must sort before (or at worst equal to) the second,
	// but since they're distinct the strict less-than must hold.
	aBytes := [16]byte(a)
	bBytes := [16]byte(b)

	less := false
	for i := range aBytes {
		if aBytes[i] < bBytes[i] {
			less = true
			break
		}
		if aBytes[i] > bBytes[i] {
			break
		}
	}
	if !less {
		t.Errorf("UUIDv7 ids not time-ordered: first=%v second=%v", a, b)
	}
}

// ---------------------------------------------------------------------------
// Statused with a custom string type
// ---------------------------------------------------------------------------

type orderStatus string

const (
	statusPending   orderStatus = "pending"
	statusShipped   orderStatus = "shipped"
	statusDelivered orderStatus = "delivered"
)

func TestStatusedCustomType(t *testing.T) {
	s := model.Statused[orderStatus]{Status: statusShipped}
	if s.Status != statusShipped {
		t.Fatalf("expected %q, got %q", statusShipped, s.Status)
	}
	// Confirm the zero value is the zero of the custom type, not a model constant.
	var zero model.Statused[orderStatus]
	if zero.Status != "" {
		t.Fatalf("zero Statused[orderStatus].Status should be empty string, got %q", zero.Status)
	}
}

// ---------------------------------------------------------------------------
// Money: exact decimal arithmetic (0.1 + 0.2 == 0.3)
// ---------------------------------------------------------------------------

func TestMoneyExactDecimals(t *testing.T) {
	a := model.Money{Amount: decimal.NewFromFloat(0.1), Currency: "USD"}
	b := model.Money{Amount: decimal.NewFromFloat(0.2), Currency: "USD"}
	sum := a.Amount.Add(b.Amount)

	expected := decimal.NewFromFloat(0.3)
	if !sum.Equal(expected) {
		t.Fatalf("0.1 + 0.2 = %v, want %v (float imprecision leaked)", sum, expected)
	}
}
