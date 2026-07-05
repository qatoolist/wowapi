package config_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/kernel/config"
)

// ---------- binder branch coverage via Load ----------

// dashCfg exercises the conf:"-" skip in bindStructInto.
type dashCfg struct {
	config.Framework
	Skip string `conf:"-" json:"-"`
}

func TestLoadConfKeyDashSkipped(t *testing.T) {
	base := writeYAML(t, "environment: dev\n")
	f, _, err := config.Load[dashCfg](config.Options{BaseFile: base})
	if err != nil {
		t.Fatal(err)
	}
	if f.Skip != "" {
		t.Errorf("conf:\"-\" field must stay unbound, got %q", f.Skip)
	}
}

// unexportedFieldCfg carries an unexported field the binder and the unsafe
// enforcer must both skip. Loaded in prod so enforceUnsafe runs too.
type unexportedFieldCfg struct {
	config.Framework
	secret string //nolint:unused // present to exercise the unexported-field skip
	Public string `conf:"public" default:"x" json:"public"`
}

func TestLoadSkipsUnexportedFields(t *testing.T) {
	base := writeYAML(t, "environment: prod\n")
	f, _, err := config.Load[unexportedFieldCfg](config.Options{BaseFile: base})
	if err != nil {
		t.Fatalf("unexported field must be skipped cleanly: %v", err)
	}
	if f.Public != "x" {
		t.Errorf("public default = %q", f.Public)
	}
	if f.secret != "" {
		t.Errorf("unexported field must stay unbound, got %q", f.secret)
	}
}

func TestLoadStructFieldExpectsMapping(t *testing.T) {
	base := writeYAML(t, "environment: dev\nhttp: notamapping\n")
	_, _, err := config.Load[config.Framework](config.Options{BaseFile: base})
	mustContain(t, err, "http", "expected a mapping")
}

func TestLoadModuleNamespacesRawNotMapping(t *testing.T) {
	base := writeYAML(t, `
environment: dev
product:
  name: acme
modules: notamapping
`)
	_, _, err := config.Load[prodConfig](config.Options{BaseFile: base})
	mustContain(t, err, "modules", "mapping of module namespaces")
}

func TestLoadModuleNamespaceEntryNotMapping(t *testing.T) {
	base := writeYAML(t, `
environment: dev
product:
  name: acme
modules:
  requests: scalarvalue
`)
	_, _, err := config.Load[prodConfig](config.Options{BaseFile: base})
	mustContain(t, err, "modules.requests", "must be a mapping")
}

func TestLoadSecretFieldNonString(t *testing.T) {
	base := writeYAML(t, `
environment: dev
product:
  name: acme
  dsn: 123
`)
	_, _, err := config.Load[prodConfig](config.Options{BaseFile: base})
	mustContain(t, err, "product.dsn", "secretref")
}

// secReqCfg has a required Secret field that is left unset.
type secReqCfg struct {
	config.Framework
	Token config.Secret `conf:"token" required:"true" json:"token"`
}

func TestLoadRequiredSecretMissing(t *testing.T) {
	base := writeYAML(t, "environment: dev\n")
	_, _, err := config.Load[secReqCfg](config.Options{BaseFile: base})
	mustContain(t, err, "token", "required")
}

// ptrSecretCfg tries to use a *Secret, which the binder refuses.
type ptrSecretCfg struct {
	config.Framework
	Bad *config.Secret `conf:"bad" json:"bad,omitempty"`
}

func TestLoadPointerToSecretRejected(t *testing.T) {
	base := writeYAML(t, "environment: dev\n")
	_, _, err := config.Load[ptrSecretCfg](config.Options{BaseFile: base})
	mustContain(t, err, "bad", "by value")
}

// ptrDefCfg has a pointer scalar with an invalid compiled default.
type ptrDefCfg struct {
	config.Framework
	N *int `conf:"n" default:"notanint" json:"n,omitempty"`
}

func TestLoadPointerInvalidCompiledDefault(t *testing.T) {
	base := writeYAML(t, "environment: dev\n")
	_, _, err := config.Load[ptrDefCfg](config.Options{BaseFile: base})
	mustContain(t, err, "n", "invalid compiled default")
}

// ptrReqCfg has a required pointer scalar left unset.
type ptrReqCfg struct {
	config.Framework
	N *int `conf:"n" required:"true" json:"n,omitempty"`
}

func TestLoadPointerRequiredMissing(t *testing.T) {
	base := writeYAML(t, "environment: dev\n")
	_, _, err := config.Load[ptrReqCfg](config.Options{BaseFile: base})
	mustContain(t, err, "n", "required")
}

// ptrStructCfg has a pointer-to-struct fed a scalar.
type ptrSubx struct {
	Name string `conf:"name" json:"name"`
}
type ptrStructCfg struct {
	config.Framework
	Sub *ptrSubx `conf:"sub" json:"sub,omitempty"`
}

func TestLoadPointerStructExpectsMapping(t *testing.T) {
	base := writeYAML(t, "environment: dev\nsub: scalar\n")
	_, _, err := config.Load[ptrStructCfg](config.Options{BaseFile: base})
	mustContain(t, err, "sub", "expected a mapping")
}

// ptrScalarCfg has a pointer scalar fed an unconvertible value.
type ptrScalarCfg struct {
	config.Framework
	N *int `conf:"n" json:"n,omitempty"`
}

func TestLoadPointerScalarConversionError(t *testing.T) {
	base := writeYAML(t, "environment: dev\nn: notanumber\n")
	_, _, err := config.Load[ptrScalarCfg](config.Options{BaseFile: base})
	mustContain(t, err, "n", "valid integer")
}

func TestLoadPointerStructAppliesDefaults(t *testing.T) {
	// Present pointer-to-struct binds and fills nested defaults — covers the
	// happy pointer-struct branch.
	base := writeYAML(t, "environment: dev\nsub: {}\n")
	f, _, err := config.Load[ptrStructCfg](config.Options{BaseFile: base})
	if err != nil {
		t.Fatal(err)
	}
	if f.Sub == nil {
		t.Fatal("present sub should allocate the pointer")
	}
}

// ---------- enforceUnsafe pointer-to-struct recursion ----------

type ptrUnsafeDev struct {
	Echo bool `conf:"echo" unsafe:"true" json:"echo"`
}
type ptrUnsafeCfg struct {
	config.Framework
	Dev *ptrUnsafeDev `conf:"dev" json:"dev,omitempty"`
}

func TestLoadUnsafeInsidePointerStructRefusedInProd(t *testing.T) {
	base := writeYAML(t, `
environment: prod
dev:
  echo: true
`)
	_, _, err := config.Load[ptrUnsafeCfg](config.Options{BaseFile: base})
	mustContain(t, err, "dev.echo", "prod")
}

func TestLoadUnsafeInsidePointerStructWarnsInStage(t *testing.T) {
	base := writeYAML(t, `
environment: stage
dev:
  echo: true
`)
	got, err := config.LoadDetailed[ptrUnsafeCfg](config.Options{BaseFile: base})
	if err != nil {
		t.Fatalf("stage should warn, not fail: %v", err)
	}
	if !strings.Contains(strings.Join(got.Warnings, "\n"), "dev.echo") {
		t.Errorf("expected a stage warning for dev.echo, got %q", got.Warnings)
	}
}

// ---------- LoadDetailed: env-file parse error ----------

func TestLoadEnvFileParseErrorAccumulates(t *testing.T) {
	base := writeYAML(t, "environment: dev\n")
	missing := filepath.Join(t.TempDir(), "overlay-nope.yaml")
	_, _, err := config.Load[config.Framework](config.Options{BaseFile: base, EnvFile: missing})
	mustContain(t, err, "overlay-nope.yaml")
}

// ---------- pointer-receiver Validate hook (load.go &cfg branch) ----------

// ptrValidateCfg does NOT embed Framework, so its value method set has no
// Validate; only *ptrValidateCfg does. This forces LoadDetailed down the
// any(&cfg) validation branch.
type ptrValidateCfg struct {
	Environment config.Env `conf:"environment" json:"environment"`
	Field       int        `conf:"field" default:"1" json:"field"`
}

func (c *ptrValidateCfg) Validate() error {
	if c.Field == 99 {
		return fmt.Errorf("field: 99 is forbidden")
	}
	return nil
}

func TestLoadPointerReceiverValidateRuns(t *testing.T) {
	bad := writeYAML(t, "environment: dev\nfield: 99\n")
	_, _, err := config.Load[ptrValidateCfg](config.Options{BaseFile: bad})
	mustContain(t, err, "field", "forbidden")

	ok := writeYAML(t, "environment: dev\nfield: 5\n")
	if _, _, err := config.Load[ptrValidateCfg](config.Options{BaseFile: ok}); err != nil {
		t.Fatalf("valid pointer-receiver config should load: %v", err)
	}
}

// ---------- LoadDetailed: fingerprint failure on an unmarshalable config ----------

// unmarshalableCfg binds fine (Bad is conf:"-" so the binder skips it) and
// validates (no Validate hook), but json.Marshal fails on the channel field,
// driving the FingerprintOf error path at the tail of LoadDetailed.
type unmarshalableCfg struct {
	Environment config.Env `conf:"environment" json:"environment"`
	Bad         chan int   `conf:"-"`
}

func TestLoadFingerprintErrorSurfaces(t *testing.T) {
	base := writeYAML(t, "environment: dev\n")
	_, _, err := config.Load[unmarshalableCfg](config.Options{BaseFile: base})
	mustContain(t, err, "fingerprint")
}
