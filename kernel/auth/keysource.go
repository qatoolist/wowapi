package auth

import (
	"context"

	"github.com/qatoolist/wowapi/v2/kernel/errors"
)

// staticKeySource is a KeySource backed by an in-memory map[kid]key. It is
// useful for tests and simple single-signer deployments where the verification
// keys are known at boot.
type staticKeySource struct {
	keys map[string]any
}

// NewStaticKeySource returns a KeySource that resolves kids from an in-memory
// map. The map is copied so later mutation by the caller cannot affect it.
func NewStaticKeySource(keys map[string]any) KeySource {
	cp := make(map[string]any, len(keys))
	for k, v := range keys {
		cp[k] = v
	}
	return &staticKeySource{keys: cp}
}

// Key returns the verification key for kid, or KindUnauthenticated when the kid
// is unknown. The kid is not echoed in the error.
func (s *staticKeySource) Key(_ context.Context, kid string) (any, error) {
	key, ok := s.keys[kid]
	if !ok {
		return nil, errors.E(errors.KindUnauthenticated, "unauthenticated",
			"unknown key id", errors.Op("auth.StaticKeySource.Key"))
	}
	return key, nil
}
