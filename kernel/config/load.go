package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/qatoolist/wowapi/kernel/secrets"
)

// Layer identifies which precedence layer supplied a config value
// (blueprint 12 §3; surfaced by `wowapi config doctor`).
type Layer string

const (
	LayerDefault  Layer = "default"   // compiled default tag
	LayerBaseFile Layer = "base-file" // configs/base.yaml
	LayerEnvFile  Layer = "env-file"  // configs/<env>.yaml overlay
	LayerEnvVar   Layer = "env"       // PREFIX__SECTION__FIELD environment variable
	LayerFlag     Layer = "flag"      // local-only CLI flag
	LayerSecret   Layer = "secret"    // value resolved through the secret provider
)

// Provenance maps dotted config keys to the layer that supplied their value.
type Provenance map[string]Layer

// Options configures a Load call. Zero-value fields skip their layer.
type Options struct {
	// BaseFile is the committed product config file (configs/base.yaml).
	BaseFile string
	// EnvFile is the environment overlay (configs/<env>.yaml).
	EnvFile string
	// EnvPrefix enables the environment-variable layer:
	// "ACME__" maps ACME__DB__MAX_CONNS=32 onto db.max_conns. Empty = no env layer.
	EnvPrefix string
	// Environ supplies the environment ("KEY=VALUE" pairs); nil = os.Environ().
	Environ []string
	// Secrets resolves secretref:// values at boot. Required if any Secret
	// field is set; resolution failures fail the load.
	Secrets secrets.Provider
	// Flags holds local-tooling overrides by dotted key ("http.addr" → value).
	// The loader refuses to start when flags are set and environment=prod.
	Flags map[string]string
}

// Loaded is the full result of LoadDetailed.
type Loaded[T any] struct {
	Config      T
	Fingerprint Fingerprint
	Provenance  Provenance
	// Warnings carries non-fatal findings (e.g. unsafe knobs enabled in stage).
	Warnings []string
}

// Load computes the effective configuration exactly once, at boot:
// compiled defaults ← base file ← env overlay ← env vars ← flags, then
// secret resolution, then validation. It fails with ALL problems joined,
// never just the first (blueprint 12 §3–4).
func Load[T any](opts Options) (T, Fingerprint, error) {
	l, err := LoadDetailed[T](opts)
	return l.Config, l.Fingerprint, err
}

// LoadDetailed is Load plus per-key provenance and warnings, for
// `wowapi config doctor` and startup diagnostics.
func LoadDetailed[T any](opts Options) (Loaded[T], error) {
	var out Loaded[T]
	if reflect.TypeFor[T]().Kind() != reflect.Struct {
		return out, fmt.Errorf("config: Load target must be a struct, got %s", reflect.TypeFor[T]())
	}

	var errs []error
	prov := Provenance{}
	tree := map[string]any{}

	if opts.BaseFile != "" {
		m, err := parseYAMLFile(opts.BaseFile)
		if err != nil {
			errs = append(errs, err)
		} else {
			deepMerge(tree, m, LayerBaseFile, prov, "")
		}
	}
	if opts.EnvFile != "" {
		m, err := parseYAMLFile(opts.EnvFile)
		if err != nil {
			errs = append(errs, err)
		} else {
			deepMerge(tree, m, LayerEnvFile, prov, "")
		}
	}
	// Snapshot the file-layer environment before lower-trust layers apply:
	// prod checks must key off the committed value (SEC-5).
	fileEnv, _ := tree["environment"].(string)

	if opts.EnvPrefix != "" {
		environ := opts.Environ
		if environ == nil {
			environ = os.Environ()
		}
		applyEnviron(tree, opts.EnvPrefix, environ, prov)
	}
	if len(opts.Flags) > 0 {
		applyFlags(tree, opts.Flags, prov)
	}

	// Fail-closed environment (SEC-1/D-0010): deployed processes must say
	// which environment they are; nothing may default its way into `local`
	// validation rules. The key must live at the ROOT of the config document
	// (where config.Framework binds it — nesting or renaming it in a product
	// type is unsupported).
	//
	// Trust rules (SEC-5): the environment that gates production safety must
	// not be downgradable by a lower-trust layer. Flags may never set it; an
	// environment variable may supply it only when no config file does — a
	// mismatch with a committed file value is an error, not an override.
	finalEnv, _ := tree["environment"].(string)
	if _, ok := opts.Flags["environment"]; ok {
		errs = append(errs, errors.New(
			"environment: may not be set via CLI flags — commit it in a config file or set it via the platform environment variable"))
	}
	if fileEnv != "" && finalEnv != fileEnv {
		errs = append(errs, fmt.Errorf(
			"environment: %q from the %s layer conflicts with %q committed in config files — the environment gate is not overridable", finalEnv, prov["environment"], fileEnv))
	}
	env := Env(finalEnv)
	if fileEnv != "" {
		env = Env(fileEnv)
	}
	if finalEnv == "" {
		errs = append(errs, errors.New(
			"environment: required — set it explicitly in a config file or environment variable (fail-closed)"))
	}
	if env.IsProd() && len(opts.Flags) > 0 {
		errs = append(errs, errors.New(
			"flags: CLI flag overrides are for local tooling and are refused when environment=prod"))
	}

	b := &binder{env: env, prov: prov}
	var cfg T
	b.bindStruct(reflect.ValueOf(&cfg).Elem(), tree, "")
	b.reportLeftovers(tree, "")
	b.enforceUnsafe(reflect.ValueOf(&cfg).Elem(), "")
	errs = append(errs, b.errs...)

	// Resolve secret references once, at boot — never on hot paths.
	if len(b.secrets) > 0 {
		if opts.Secrets == nil {
			errs = append(errs, fmt.Errorf(
				"secrets: %d secretref value(s) present but no secret provider is configured", len(b.secrets)))
		} else {
			ctx := context.Background()
			for _, slot := range b.secrets {
				ref, err := secrets.ParseRef(slot.ptr.Ref())
				if err != nil {
					errs = append(errs, fmt.Errorf("%s: %w", slot.path, err))
					continue
				}
				val, err := opts.Secrets.Resolve(ctx, ref)
				if err != nil {
					errs = append(errs, fmt.Errorf("%s: resolving %s: %w", slot.path, ref, err))
					continue
				}
				*slot.ptr = NewSecret(ref.String(), val)
				prov[slot.path] = LayerSecret
			}
		}
	}

	// Range/cross-field/production-safety validation hook. Runs even when
	// bind errors exist so the boot failure lists everything at once.
	//
	// Composition contract (ARCH-10): a product Config embedding Framework
	// that defines its OWN Validate() shadows the promoted Framework.Validate
	// — Go method promotion, not a loader choice. Such a Validate MUST
	// delegate: `return errors.Join(c.Framework.Validate(), c.validateProduct())`.
	// The `wowapi init` scaffold generates exactly that shape.
	if v, ok := any(cfg).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			errs = append(errs, err)
		}
	} else if v, ok := any(&cfg).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			errs = append(errs, err)
		}
	}

	out.Provenance = prov
	out.Warnings = b.warnings
	if err := errors.Join(errs...); err != nil {
		return out, fmt.Errorf("config: invalid configuration:\n%w", err)
	}

	fp, err := FingerprintOf(cfg)
	if err != nil {
		return out, err
	}
	out.Config = cfg
	out.Fingerprint = fp
	return out, nil
}
