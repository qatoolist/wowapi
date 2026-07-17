package compat

import (
	"reflect"
	"testing"

	"github.com/qatoolist/wowapi/v2/app"
	"github.com/qatoolist/wowapi/v2/foundation/document"
)

// V2 CONTRACT fixture (third/fifth closure audits 2026-07-17):
// independently maintained derived projects write POSITIONAL composite
// literals against stable exported structs — adding, removing, reordering, or
// retyping ANY field breaks their build even when godoc would call the change
// "additive". This gate freezes the exact field sequence of the known
// positional-literal types. The literal-form compile proofs live in the
// owning packages' external tests (TestHookUnkeyedLiteralCompatibility,
// TestUploadEventUnkeyedLiteralCompatibility — go vet's composites check
// exempts same-path _test packages, so they can spell the literal out); this
// fixture guards the same contract from a separate consumer-side package.
// If it fails, the change is source-incompatible for stable-v1 consumers and
// needs either a major version or a compatibility-preserving delivery
// mechanism (context values or new types), like
// document.UploadDeliveryFromContext and app.SupervisedHook.
func TestV2ContractStructShapesAreFrozen(t *testing.T) {
	assertShape := func(name string, typ reflect.Type, want []struct{ name, typ string }) {
		t.Helper()
		if typ.NumField() != len(want) {
			t.Fatalf("%s has %d fields, want the frozen %d — positional literals in derived projects no longer compile",
				name, typ.NumField(), len(want))
		}
		for i, w := range want {
			f := typ.Field(i)
			if f.Name != w.name || f.Type.String() != w.typ {
				t.Fatalf("%s field %d = %s %s, want frozen %s %s (order and types are part of the contract)",
					name, i, f.Name, f.Type, w.name, w.typ)
			}
		}
	}

	assertShape("app.Hook", reflect.TypeOf(app.Hook{}), []struct{ name, typ string }{
		{"Name", "string"},
		{"Start", "func(context.Context) error"},
		{"Stop", "func(context.Context) error"},
	})
	assertShape("document.UploadEvent", reflect.TypeOf(document.UploadEvent{}), []struct{ name, typ string }{
		{"DocumentID", "string"},
		{"Class", "string"},
		{"VersionNo", "int"},
		{"StorageKey", "string"},
		{"MIME", "string"},
		{"SizeBytes", "int64"},
		{"Sensitivity", "document.Sensitivity"},
	})
}
