package config_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/secrets"
)

// ---------- helpers ----------

func writeYAML(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "cfg.yaml")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

type fakeProvider struct {
	values map[string]string
	errs   map[string]error
}

func (f fakeProvider) Resolve(_ context.Context, ref secrets.Ref) (string, error) {
	if err, ok := f.errs[ref.String()]; ok {
		return "", err
	}
	v, ok := f.values[ref.String()]
	if !ok {
		return "", fmt.Errorf("secret %s not found", ref)
	}
	return v, nil
}

// prodConfig mirrors the product-composition shape from blueprint 12 §2:
// an embedded config.Framework, a product section, and module namespaces.
type prodConfig struct {
	config.Framework
	Product prodSection       `conf:"product" json:"product"`
	Modules config.Namespaces `conf:"modules" json:"modules"`
}

type prodSection struct {
	Name string        `conf:"name" required:"true" doc:"service name" json:"name"`
	DSN  config.Secret `conf:"dsn" json:"dsn"`
	Dev  devSection    `conf:"dev" json:"dev"`
	Tags []string      `conf:"tags" json:"tags"`
	TTL  time.Duration `conf:"ttl" default:"1m" json:"ttl"`
}

type devSection struct {
	FakeIssuer bool `conf:"fake_issuer" unsafe:"true" json:"fake_issuer"`
	SQLEcho    bool `conf:"sql_echo" unsafe:"true" json:"sql_echo"`
}

func mustContain(t *testing.T, err error, subs ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", subs)
	}
	for _, s := range subs {
		if !strings.Contains(err.Error(), s) {
			t.Errorf("error does not mention %q:\n%v", s, err)
		}
	}
}

// ---------- precedence ----------

func TestLoadPrecedence(t *testing.T) {
	base := writeYAML(t, `
environment: dev
http:
  addr: ":1111"
log:
  level: warn
`)
	overlay := writeYAML(t, `
http:
  addr: ":2222"
`)
	got, err := config.LoadDetailed[config.Framework](config.Options{
		BaseFile:  base,
		EnvFile:   overlay,
		EnvPrefix: "WOWAPI__",
		Environ:   []string{"WOWAPI__HTTP__ADDR=:3333", "UNRELATED=x"},
		Flags:     map[string]string{"http.addr": ":4444"},
	})
	if err != nil {
		t.Fatal(err)
	}
	f := got.Config
	if f.HTTP.Addr != ":4444" {
		t.Errorf("flag should win: addr = %q", f.HTTP.Addr)
	}
	if f.Log.Level != "warn" {
		t.Errorf("base value lost: log.level = %q", f.Log.Level)
	}
	if f.HTTP.ReadHeaderTimeout != 5*time.Second {
		t.Errorf("default lost: read_header_timeout = %v", f.HTTP.ReadHeaderTimeout)
	}
	if f.Environment != config.EnvDev {
		t.Errorf("environment = %q", f.Environment)
	}
	wantProv := map[string]config.Layer{
		"http.addr":                config.LayerFlag,
		"log.level":                config.LayerBaseFile,
		"http.read_header_timeout": config.LayerDefault,
		"environment":              config.LayerBaseFile,
	}
	for k, want := range wantProv {
		if got.Provenance[k] != want {
			t.Errorf("provenance[%s] = %q, want %q", k, got.Provenance[k], want)
		}
	}
}

func TestLoadEnvFileBeatsBase(t *testing.T) {
	base := writeYAML(t, "environment: dev\nlog:\n  level: warn\n")
	overlay := writeYAML(t, "log:\n  level: error\n")
	f, _, err := config.Load[config.Framework](config.Options{BaseFile: base, EnvFile: overlay})
	if err != nil {
		t.Fatal(err)
	}
	if f.Log.Level != "error" {
		t.Errorf("overlay should beat base: log.level = %q", f.Log.Level)
	}
}

// ---------- fail-closed environment (SEC-1 / D-0010) ----------

func TestLoadEnvironmentFailClosed(t *testing.T) {
	base := writeYAML(t, "log:\n  level: info\n")
	_, _, err := config.Load[config.Framework](config.Options{BaseFile: base})
	mustContain(t, err, "environment")
}

func TestLoadEnvironmentFromEnvVarSatisfiesFailClosed(t *testing.T) {
	f, _, err := config.Load[config.Framework](config.Options{
		EnvPrefix: "WOWAPI__",
		Environ:   []string{"WOWAPI__ENVIRONMENT=stage"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if f.Environment != config.EnvStage {
		t.Errorf("environment = %q", f.Environment)
	}
}

func TestLoadUnknownEnvironmentRejected(t *testing.T) {
	base := writeYAML(t, "environment: production\n")
	_, _, err := config.Load[config.Framework](config.Options{BaseFile: base})
	mustContain(t, err, "production")
}

// ---------- error accumulation ----------

func TestLoadErrorsAccumulate(t *testing.T) {
	base := writeYAML(t, `
htp:
  addr: ":1111"
log:
  level: loud
http:
  read_header_timeout: nonsense
`)
	_, _, err := config.Load[config.Framework](config.Options{BaseFile: base})
	// all four problems reported at once (the bad duration is reported by
	// path only — raw values never echo into diagnostics, SEC-8)
	mustContain(t, err, "environment", "htp", "loud", "http.read_header_timeout")
	if strings.Contains(err.Error(), "nonsense") {
		t.Errorf("conversion error echoed the raw value: %v", err)
	}
}

func TestLoadUnknownKeyViaEnvVarTypo(t *testing.T) {
	_, _, err := config.Load[config.Framework](config.Options{
		EnvPrefix: "WOWAPI__",
		Environ:   []string{"WOWAPI__ENVIRONMENT=dev", "WOWAPI__HTP__ADDR=:9"},
	})
	mustContain(t, err, "htp.addr")
}

func TestLoadSchemaVersionBounds(t *testing.T) {
	base := writeYAML(t, "environment: dev\nschema_version: 99\n")
	_, _, err := config.Load[config.Framework](config.Options{BaseFile: base})
	mustContain(t, err, "schema_version")
}

func TestLoadMissingBaseFile(t *testing.T) {
	_, _, err := config.Load[config.Framework](config.Options{
		BaseFile: filepath.Join(t.TempDir(), "nope.yaml"),
	})
	mustContain(t, err, "nope.yaml")
}

func TestLoadRequiredFieldMissing(t *testing.T) {
	base := writeYAML(t, "environment: dev\n")
	_, _, err := config.Load[prodConfig](config.Options{BaseFile: base})
	mustContain(t, err, "product.name")
}

// ---------- production safety ----------

func TestLoadProdSafetyFloor(t *testing.T) {
	base := writeYAML(t, `
environment: prod
log:
  level: debug
  format: text
`)
	_, _, err := config.Load[config.Framework](config.Options{BaseFile: base})
	mustContain(t, err, "log.format", "log.level")
}

func TestLoadFlagsRefusedInProd(t *testing.T) {
	base := writeYAML(t, "environment: prod\n")
	_, _, err := config.Load[config.Framework](config.Options{
		BaseFile: base,
		Flags:    map[string]string{"log.level": "info"},
	})
	mustContain(t, err, "flag")
}

func TestLoadUnsafeKnobMatrix(t *testing.T) {
	for _, knob := range []string{"fake_issuer", "sql_echo"} {
		t.Run(knob+"/prod", func(t *testing.T) {
			base := writeYAML(t, fmt.Sprintf(`
environment: prod
product:
  name: acme
  dev:
    %s: true
`, knob))
			_, _, err := config.Load[prodConfig](config.Options{BaseFile: base})
			mustContain(t, err, "product.dev."+knob, "prod")
		})
		t.Run(knob+"/stage-warns", func(t *testing.T) {
			base := writeYAML(t, fmt.Sprintf(`
environment: stage
product:
  name: acme
  dev:
    %s: true
`, knob))
			got, err := config.LoadDetailed[prodConfig](config.Options{BaseFile: base})
			if err != nil {
				t.Fatal(err)
			}
			joined := strings.Join(got.Warnings, "\n")
			if !strings.Contains(joined, "product.dev."+knob) {
				t.Errorf("stage should warn about %s, warnings: %q", knob, got.Warnings)
			}
		})
		t.Run(knob+"/dev-ok", func(t *testing.T) {
			base := writeYAML(t, fmt.Sprintf(`
environment: dev
product:
  name: acme
  dev:
    %s: true
`, knob))
			got, err := config.LoadDetailed[prodConfig](config.Options{BaseFile: base})
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Warnings) != 0 {
				t.Errorf("dev should not warn: %q", got.Warnings)
			}
		})
	}
}

// unsafeDefaultCfg reproduces SEC-3: a knob whose unsafe value is its
// compiled default must still be refused in prod.
type unsafeDefaultCfg struct {
	config.Framework
	Dev struct {
		Echo bool `conf:"echo" unsafe:"true" default:"true" json:"echo"`
	} `conf:"dev" json:"dev"`
}

// unsafeStructCfg reproduces SEC-4: unsafe markers on non-scalar fields.
type unsafeStructCfg struct {
	config.Framework
	Dev unsafeIssuer `conf:"dev" unsafe:"true" json:"dev"`
}

type unsafeIssuer struct {
	FakeIssuer string `conf:"fake_issuer" json:"fake_issuer"`
}

func TestLoadUnsafeDefaultRefusedInProd(t *testing.T) {
	base := writeYAML(t, "environment: prod\n")
	_, _, err := config.Load[unsafeDefaultCfg](config.Options{BaseFile: base})
	mustContain(t, err, "dev.echo", "prod")
}

func TestLoadUnsafeStructKnobRefusedInProd(t *testing.T) {
	base := writeYAML(t, `
environment: prod
dev:
  fake_issuer: http://localhost:9999
`)
	_, _, err := config.Load[unsafeStructCfg](config.Options{BaseFile: base})
	mustContain(t, err, "dev", "prod")
}

func TestLoadUnsafeStructKnobUnsetIsFineInProd(t *testing.T) {
	base := writeYAML(t, "environment: prod\n")
	_, _, err := config.Load[unsafeStructCfg](config.Options{BaseFile: base})
	if err != nil {
		t.Fatalf("zero unsafe struct must not trip the prod gate: %v", err)
	}
}

// ---------- environment trust rules (SEC-5) ----------

func TestLoadEnvironmentNotDowngradableByEnvVar(t *testing.T) {
	base := writeYAML(t, "environment: prod\n")
	_, _, err := config.Load[config.Framework](config.Options{
		BaseFile:  base,
		EnvPrefix: "WOWAPI__",
		Environ:   []string{"WOWAPI__ENVIRONMENT=local"},
	})
	mustContain(t, err, "environment", "prod", "local")
}

func TestLoadEnvironmentNeverFromFlags(t *testing.T) {
	base := writeYAML(t, "environment: dev\n")
	_, _, err := config.Load[config.Framework](config.Options{
		BaseFile: base,
		Flags:    map[string]string{"environment": "local"},
	})
	mustContain(t, err, "environment", "flag")
}

func TestLoadFlagDowngradeStillRefusedInProd(t *testing.T) {
	// The flag both tries to lower the environment AND should itself be
	// refused because the committed environment is prod.
	base := writeYAML(t, "environment: prod\n")
	_, _, err := config.Load[config.Framework](config.Options{
		BaseFile: base,
		Flags:    map[string]string{"environment": "local"},
	})
	mustContain(t, err, "flag", "prod")
}

// ---------- pointer fields (ARCH-6) ----------

type ptrCfg struct {
	config.Framework
	Opt *int    `conf:"opt" json:"opt,omitempty"`
	Sub *ptrSub `conf:"sub" json:"sub,omitempty"`
}

type ptrSub struct {
	Name string `conf:"name" default:"fallback" json:"name"`
}

func TestLoadPointerFields(t *testing.T) {
	base := writeYAML(t, `
environment: dev
opt: 7
sub:
  name: set
`)
	f, _, err := config.Load[ptrCfg](config.Options{BaseFile: base})
	if err != nil {
		t.Fatal(err)
	}
	if f.Opt == nil || *f.Opt != 7 {
		t.Errorf("Opt = %v", f.Opt)
	}
	if f.Sub == nil || f.Sub.Name != "set" {
		t.Errorf("Sub = %+v", f.Sub)
	}

	empty := writeYAML(t, "environment: dev\n")
	f, _, err = config.Load[ptrCfg](config.Options{BaseFile: empty})
	if err != nil {
		t.Fatal(err)
	}
	if f.Opt != nil || f.Sub != nil {
		t.Errorf("absent pointers must stay nil: opt=%v sub=%v", f.Opt, f.Sub)
	}
}

// ---------- binder robustness (ARCH-15) ----------

type collideCfg struct {
	config.Framework
	Log string `conf:"log" json:"log_shadow"`
}

func TestLoadEmbeddedKeyCollision(t *testing.T) {
	base := writeYAML(t, "environment: dev\n")
	_, _, err := config.Load[collideCfg](config.Options{BaseFile: base})
	mustContain(t, err, "log", "more than one field")
}

func TestLoadNonStructTarget(t *testing.T) {
	_, _, err := config.Load[int](config.Options{})
	mustContain(t, err, "struct")
}

// ---------- module namespaces are file-layer only (ARCH-8) ----------

func TestLoadModuleNamespaceViaEnvVarRejected(t *testing.T) {
	base := writeYAML(t, `
environment: dev
product:
  name: acme
`)
	_, _, err := config.Load[prodConfig](config.Options{
		BaseFile:  base,
		EnvPrefix: "WOWAPI__",
		Environ:   []string{"WOWAPI__MODULES__REQUESTS__SLA_HOURS=4"},
	})
	mustContain(t, err, "modules.requests.sla_hours", "config files")
}

// ---------- diagnostics never echo values (SEC-7/SEC-8) ----------

func TestLoadConversionErrorsDoNotEchoValues(t *testing.T) {
	base := writeYAML(t, `
environment: dev
http:
  max_body_bytes: hunter2-not-a-number
`)
	_, _, err := config.Load[config.Framework](config.Options{BaseFile: base})
	mustContain(t, err, "http.max_body_bytes")
	if strings.Contains(err.Error(), "hunter2") {
		t.Errorf("conversion error echoed the raw value: %v", err)
	}
}

func TestLoadYAMLParseErrorScrubbed(t *testing.T) {
	base := writeYAML(t, "environment: !!int hunter2-raw-value\n")
	_, _, err := config.Load[config.Framework](config.Options{BaseFile: base})
	if err == nil {
		t.Fatal("expected a parse error")
	}
	if strings.Contains(err.Error(), "hunter2") {
		t.Errorf("yaml error echoed file content: %v", err)
	}
}

// ---------- secrets ----------

func TestLoadSecretResolution(t *testing.T) {
	base := writeYAML(t, `
environment: dev
product:
  name: acme
  dsn: secretref://env/TEST_DSN
`)
	got, err := config.LoadDetailed[prodConfig](config.Options{
		BaseFile: base,
		Secrets:  fakeProvider{values: map[string]string{"secretref://env/TEST_DSN": "postgres://u:p@h/db"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Config.Product.DSN.Reveal() != "postgres://u:p@h/db" {
		t.Errorf("secret not resolved: %q", got.Config.Product.DSN.Reveal())
	}
	if got.Provenance["product.dsn"] != config.LayerSecret {
		t.Errorf("provenance[product.dsn] = %q", got.Provenance["product.dsn"])
	}
}

func TestLoadRawSecretValueRejected(t *testing.T) {
	base := writeYAML(t, `
environment: dev
product:
  name: acme
  dsn: postgres://u:hunter2@h/db
`)
	_, _, err := config.Load[prodConfig](config.Options{BaseFile: base})
	mustContain(t, err, "secretref")
	if strings.Contains(err.Error(), "hunter2") {
		t.Errorf("error leaked raw secret material: %v", err)
	}
}

func TestLoadSecretRefWithoutProvider(t *testing.T) {
	base := writeYAML(t, `
environment: dev
product:
  name: acme
  dsn: secretref://env/TEST_DSN
`)
	_, _, err := config.Load[prodConfig](config.Options{BaseFile: base})
	mustContain(t, err, "secret", "provider")
}

func TestLoadSecretProviderErrorAccumulates(t *testing.T) {
	base := writeYAML(t, `
environment: dev
product:
  name: acme
  dsn: secretref://env/MISSING
`)
	_, _, err := config.Load[prodConfig](config.Options{
		BaseFile: base,
		Secrets:  fakeProvider{},
	})
	mustContain(t, err, "secretref://env/MISSING")
}

func TestLoadedConfigNeverPrintsSecrets(t *testing.T) {
	base := writeYAML(t, `
environment: dev
product:
  name: acme
  dsn: secretref://env/TEST_DSN
`)
	const raw = "postgres://u:sup3rsecret@h/db"
	got, err := config.LoadDetailed[prodConfig](config.Options{
		BaseFile: base,
		Secrets:  fakeProvider{values: map[string]string{"secretref://env/TEST_DSN": raw}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for name, rendered := range map[string]string{
		"fmt %v":  fmt.Sprintf("%v", got.Config),
		"fmt %+v": fmt.Sprintf("%+v", got.Config),
		"fmt %#v": fmt.Sprintf("%#v", got.Config),
	} {
		if strings.Contains(rendered, "sup3rsecret") {
			t.Errorf("%s leaked the secret: %s", name, rendered)
		}
	}
	js, err := json.Marshal(got.Config)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(js), "sup3rsecret") {
		t.Errorf("JSON leaked the secret: %s", js)
	}
	if !strings.Contains(string(js), "[redacted:secretref://env/TEST_DSN]") {
		t.Errorf("JSON should carry the redaction marker: %s", js)
	}
}

// ---------- module namespaces ----------

func TestLoadModuleNamespaces(t *testing.T) {
	base := writeYAML(t, `
environment: dev
product:
  name: acme
modules:
  requests:
    sla_hours: 4
  assets:
    bucket: media
`)
	got, err := config.LoadDetailed[prodConfig](config.Options{BaseFile: base})
	if err != nil {
		t.Fatal(err)
	}
	var reqCfg struct {
		SLAHours int `json:"sla_hours"`
	}
	if err := got.Config.Modules["requests"].Decode(&reqCfg); err != nil {
		t.Fatal(err)
	}
	if reqCfg.SLAHours != 4 {
		t.Errorf("sla_hours = %d", reqCfg.SLAHours)
	}
	// Isolation: a module's view contains ONLY its own keys — nothing from the
	// framework, product, or sibling namespaces is reachable through it.
	var catchAll map[string]any
	if err := got.Config.Modules["requests"].Decode(&catchAll); err != nil {
		t.Fatal(err)
	}
	if len(catchAll) != 1 {
		t.Errorf("requests view should hold exactly its own keys, got %v", catchAll)
	}
	for _, forbidden := range []string{"bucket", "environment", "http", "name"} {
		if _, ok := catchAll[forbidden]; ok {
			t.Errorf("module view leaked foreign key %q", forbidden)
		}
	}
	if _, ok := got.Config.Modules["missing"]; ok {
		t.Error("unregistered namespace should be absent")
	}
}

// ---------- defaults drift guard (D-0012) ----------

func TestLoadReproducesDefaults(t *testing.T) {
	base := writeYAML(t, "environment: local\n")
	f, _, err := config.Load[config.Framework](config.Options{BaseFile: base})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(f, config.Defaults()) {
		t.Errorf("tag defaults drifted from Defaults():\n got %+v\nwant %+v", f, config.Defaults())
	}
}

// ---------- fingerprint ----------

func TestFingerprintStableAndSensitive(t *testing.T) {
	base := writeYAML(t, "environment: dev\n")
	_, fp1, err := config.Load[config.Framework](config.Options{BaseFile: base})
	if err != nil {
		t.Fatal(err)
	}
	_, fp2, err := config.Load[config.Framework](config.Options{BaseFile: base})
	if err != nil {
		t.Fatal(err)
	}
	if fp1 != fp2 {
		t.Error("same input must produce the same fingerprint")
	}
	other := writeYAML(t, "environment: dev\nlog:\n  level: warn\n")
	_, fp3, err := config.Load[config.Framework](config.Options{BaseFile: other})
	if err != nil {
		t.Fatal(err)
	}
	if fp1 == fp3 {
		t.Error("different effective config must change the fingerprint")
	}
	if len(fp1.String()) != 64 {
		t.Errorf("String() should be full hex: %q", fp1.String())
	}
	if len(fp1.Short()) != 12 {
		t.Errorf("Short() should be 12 hex chars: %q", fp1.Short())
	}
}

// ---------- schema ----------

func TestSchemaFromTags(t *testing.T) {
	js, err := config.Schema[prodConfig]()
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(js, &schema); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}
	s := string(js)
	for _, want := range []string{
		`"environment"`, `"http"`, `"product"`, `"service name"`,
		`"additionalProperties": false`,
		`"x-fail-closed": true`, // schema tells the same story as the loader (ARCH-9)
	} {
		if !strings.Contains(s, want) {
			t.Errorf("schema missing %s", want)
		}
	}
	req, _ := schema["required"].([]any)
	found := false
	for _, r := range req {
		if r == "environment" {
			found = true
		}
	}
	if !found {
		t.Errorf("schema must list environment as required (fail-closed), got %v", req)
	}
}

// ---------- scalar conversion ----------

func TestLoadScalarConversions(t *testing.T) {
	base := writeYAML(t, `
environment: dev
http:
  max_body_bytes: 2048
  request_timeout: 45s
product:
  name: acme
  tags: [a, b]
`)
	f, _, err := config.Load[prodConfig](config.Options{
		BaseFile:  base,
		EnvPrefix: "WOWAPI__",
		Environ:   []string{"WOWAPI__PRODUCT__TTL=90s", "WOWAPI__HTTP__READ_HEADER_TIMEOUT=7s"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if f.HTTP.MaxBodyBytes != 2048 {
		t.Errorf("max_body_bytes = %d", f.HTTP.MaxBodyBytes)
	}
	if f.HTTP.RequestTimeout != 45*time.Second {
		t.Errorf("request_timeout = %v", f.HTTP.RequestTimeout)
	}
	if f.HTTP.ReadHeaderTimeout != 7*time.Second {
		t.Errorf("env var duration = %v", f.HTTP.ReadHeaderTimeout)
	}
	if f.Product.TTL != 90*time.Second {
		t.Errorf("ttl = %v", f.Product.TTL)
	}
	if len(f.Product.Tags) != 2 || f.Product.Tags[0] != "a" {
		t.Errorf("tags = %v", f.Product.Tags)
	}
}

func TestLoadTypeMismatchReported(t *testing.T) {
	base := writeYAML(t, `
environment: dev
http:
  max_body_bytes: lots
`)
	_, _, err := config.Load[config.Framework](config.Options{BaseFile: base})
	mustContain(t, err, "http.max_body_bytes")
}
