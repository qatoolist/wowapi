package config

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// ModuleView is the ONLY configuration surface a module receives (via
// module.Context.Config()). It exposes exactly the module's own
// `modules.<name>.*` namespace: there is deliberately no Get(key), no parent
// traversal, and no way to read framework or sibling-module configuration.
type ModuleView interface {
	// Decode strict-decodes the module's namespace into the module-owned
	// typed struct. Unknown keys in the namespace are an error (typo
	// defense); the module's own validation runs after decoding. Errors here
	// fail application boot.
	Decode(out any) error
}

// MapView is a ModuleView backed by an in-memory map. The loader produces
// these from the `modules.<name>` subtree; tests construct them directly.
type MapView map[string]any

// Namespaces is the raw `modules.*` subtree of a product configuration: one
// isolated MapView per module name. The binder captures the subtree opaquely
// (module keys are validated by each module's own strict Decode, not by the
// framework binder), and the app hands each module exactly its own view —
// there is no API to traverse from a view back to framework, product, or
// sibling configuration.
type Namespaces map[string]MapView

// Decode implements ModuleView with strict unknown-key rejection.
func (m MapView) Decode(out any) error {
	raw, err := json.Marshal(map[string]any(m))
	if err != nil {
		return fmt.Errorf("config: encode module namespace: %w", err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("config: module config: %w", err)
	}
	return nil
}
