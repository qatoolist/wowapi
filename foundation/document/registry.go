// Package document is wowapi's document / file framework: modules register
// document CLASSES (the policy envelope for a kind of file — allowed MIME types,
// a size ceiling, a default sensitivity, and an optional retention window); the
// service manages metadata rows, presigned upload sessions, immutable versioned
// file pointers, authorized presigned downloads, explicit access grants, and a
// retention sweep. Blob bytes never transit the API process — they flow client
// ↔ object store through short-lived presigned URLs (kernel/storage). Contract:
// blueprint 07 §4.
package document

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"time"

	"github.com/qatoolist/wowapi/v2/internal/sealer"

	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
)

// Sensitivity ranks how protected a document is. Ordered; the download scan-gate
// blocks pending scans at confidential and above.
type Sensitivity string

const (
	SensitivityPublic       Sensitivity = "public"
	SensitivityInternal     Sensitivity = "internal"
	SensitivityConfidential Sensitivity = "confidential"
	SensitivityRestricted   Sensitivity = "restricted"
)

var sensitivityRank = map[Sensitivity]int{
	SensitivityPublic: 0, SensitivityInternal: 1, SensitivityConfidential: 2, SensitivityRestricted: 3,
}

func (s Sensitivity) valid() bool { _, ok := sensitivityRank[s]; return ok }

// atLeast reports whether s is as protected as other.
func (s Sensitivity) atLeast(other Sensitivity) bool {
	return sensitivityRank[s] >= sensitivityRank[other]
}

// classKeyRE constrains class keys to module.name.
var classKeyRE = regexp.MustCompile(`^[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*$`)

// Class is a registered document class: the policy envelope a module declares
// for one kind of document.
type Class struct {
	Key                string
	Module             string
	DefaultSensitivity Sensitivity   // applied when Create omits sensitivity
	MaxBytes           int64         // 0 = no ceiling
	AllowedMIME        []string      // empty = any sniffed type accepted
	Retention          time.Duration // 0 = keep forever; else retention_until = created_at + Retention
}

func (c Class) allowsMIME(mime string) bool {
	return len(c.AllowedMIME) == 0 || slices.Contains(c.AllowedMIME, mime)
}

// Registry collects document classes during module registration.
type Registry struct {
	classes map[string]Class
	errs    []error
	sealed  bool
}

// NewRegistry returns an empty class registry.
func NewRegistry() *Registry { return &Registry{classes: map[string]Class{}} }

// Seal freezes the registry once boot validation completes: any later Register
// panics rather than silently adding a document class the boot gates never saw
// (closure review 2026-07-17, F-10).
// The sealer.Authority parameter restricts sealing to the framework's boot
// path: internal/sealer is unimportable outside the wowapi module, so a
// product module cannot prematurely seal a shared registry during Register.
func (r *Registry) Seal(sealer.Authority) { r.sealed = true }

// Register adds a document class. Malformed keys, a module-prefix mismatch, an
// invalid default sensitivity, or a duplicate are recorded and surfaced by Err().
func (r *Registry) Register(module string, c Class) {
	if r.sealed {
		panic("document: class registration after boot: the extension model is sealed")
	}
	if !classKeyRE.MatchString(c.Key) {
		r.errf("document class key must be module.name: %s", c.Key)
		return
	}
	if prefix := module + "."; len(c.Key) <= len(prefix) || c.Key[:len(prefix)] != prefix {
		r.errf("module %s may not register document class %s", module, c.Key)
		return
	}
	if c.DefaultSensitivity == "" {
		c.DefaultSensitivity = SensitivityInternal
	}
	if !c.DefaultSensitivity.valid() {
		r.errf("document class %s has invalid default sensitivity %q", c.Key, c.DefaultSensitivity)
		return
	}
	if c.MaxBytes < 0 {
		r.errf("document class %s has negative max_bytes", c.Key)
		return
	}
	if _, dup := r.classes[c.Key]; dup {
		r.errf("document class registered more than once: %s", c.Key)
		return
	}
	c.Module = module
	r.classes[c.Key] = c.clone()
}

// clone returns a deep copy of c: the registry must not share the AllowedMIME
// slice with callers in either direction — a retained registration value or a
// mutated Get result must never change which MIME types a validated class
// accepts (second closure audit 2026-07-17, F-10).
func (c Class) clone() Class {
	out := c
	if c.AllowedMIME != nil {
		out.AllowedMIME = append([]string(nil), c.AllowedMIME...)
	}
	return out
}

// Get returns the registered class (a deep copy — mutating its nested fields
// cannot alter the registry).
func (r *Registry) Get(key string) (Class, bool) {
	c, ok := r.classes[key]
	if !ok {
		return Class{}, false
	}
	return c.clone(), true
}

// Keys returns registered class keys, sorted.
func (r *Registry) Keys() []string {
	out := make([]string, 0, len(r.classes))
	for k := range r.classes {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func (r *Registry) errf(format string, args ...any) {
	r.errs = append(r.errs, kerr.E(kerr.KindInternal, "invalid_document_class", fmt.Sprintf(format, args...)))
}

// Err returns accumulated registration errors joined, or nil.
func (r *Registry) Err() error {
	if len(r.errs) == 0 {
		return nil
	}
	msgs := make([]string, len(r.errs))
	for i, e := range r.errs {
		msgs[i] = e.Error()
	}
	joined := msgs[0]
	for i := 1; i < len(msgs); i++ {
		joined += "; " + msgs[i]
	}
	return kerr.E(kerr.KindInternal, "document_class_registration_failed", "document class registration failed: "+joined)
}
