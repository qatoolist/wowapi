package config

import (
	"encoding"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

// binder is wowapi's single audited struct binder (blueprint 12 §2, D-0012).
// It walks a target struct guided by `conf` tags, consuming keys from the
// merged tree; whatever remains unconsumed afterwards is an unknown key.
// Range/cross-field checks are NOT its job — they live in Validate() hooks.
type binder struct {
	env      Env
	prov     Provenance
	errs     []error
	warnings []string
	secrets  []secretSlot
}

// secretSlot remembers a bound Secret field so the loader can resolve its
// reference through the secrets.Provider after binding.
type secretSlot struct {
	path string
	ptr  *Secret
}

func (b *binder) errf(format string, args ...any) {
	b.errs = append(b.errs, fmt.Errorf(format, args...))
}

func joinPath(base, key string) string {
	if base == "" {
		return key
	}
	return base + "." + key
}

var (
	durationType   = reflect.TypeFor[time.Duration]()
	secretType     = reflect.TypeFor[Secret]()
	namespacesType = reflect.TypeFor[Namespaces]()
)

// bindStruct consumes tree keys into v (a settable struct). Anonymous
// embedded structs without a conf tag flatten into the parent namespace —
// that is how a product Config embeds config.Framework at the root.
func (b *binder) bindStruct(v reflect.Value, tree map[string]any, path string) {
	b.bindStructInto(v, tree, path, map[string]bool{})
}

// claimed tracks conf keys already bound at this level so a product field
// cannot silently shadow (or be shadowed by) an embedded framework field
// claiming the same key (review finding ARCH-15).
func (b *binder) bindStructInto(v reflect.Value, tree map[string]any, path string, claimed map[string]bool) {
	t := v.Type()
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		tag := f.Tag.Get("conf")
		if tag == "-" {
			continue
		}
		if f.Anonymous && tag == "" && v.Field(i).Kind() == reflect.Struct {
			b.bindStructInto(v.Field(i), tree, path, claimed)
			continue
		}
		key := tag
		if key == "" {
			key = strings.ToLower(f.Name)
		}
		if claimed[key] {
			b.errf("%s: conf key is claimed by more than one field at the same level (embedded struct collision)", joinPath(path, key))
			continue
		}
		claimed[key] = true
		b.bindField(v.Field(i), f, tree, path, key)
	}
}

func (b *binder) bindField(fv reflect.Value, f reflect.StructField, tree map[string]any, parent, key string) {
	path := joinPath(parent, key)
	raw, present := tree[key]
	if present {
		delete(tree, key)
	}
	if raw == nil {
		present = false // `key:` with no value = unset
	}

	switch {
	case fv.Type() == namespacesType:
		// Module namespaces are file-layer only for now: env-var/flag values
		// arrive as strings and would fail each module's typed Decode with a
		// confusing per-module error at boot — reject them here with a clear
		// one instead (review finding ARCH-8; lifted when module config
		// decoding learns string coercion).
		nsPrefix := path + "."
		for k, layer := range b.prov {
			if (layer == LayerEnvVar || layer == LayerFlag) && strings.HasPrefix(k, nsPrefix) {
				b.errf("%s: set via %s — module namespace values must come from config files", k, layer)
			}
		}
		if !present {
			return
		}
		m, ok := raw.(map[string]any)
		if !ok {
			b.errf("%s: expected a mapping of module namespaces", path)
			return
		}
		ns := Namespaces{}
		for name, sub := range m {
			subM, ok := sub.(map[string]any)
			if !ok {
				b.errf("%s.%s: module namespace must be a mapping", path, name)
				continue
			}
			// Deep-copy so the returned config aliases nothing from the parse
			// tree and namespaces stay independent (review finding SEC-10).
			ns[name] = MapView(deepCopyTree(subM))
		}
		fv.Set(reflect.ValueOf(ns))
		return

	case fv.Type() == secretType:
		if !present {
			if f.Tag.Get("required") == "true" {
				b.errf("%s: required", path)
			}
			return
		}
		s, ok := raw.(string)
		if !ok {
			b.errf("%s: secret fields take a secretref://<provider>/<path> string", path)
			return
		}
		var sec Secret
		if err := sec.UnmarshalText([]byte(s)); err != nil {
			b.errf("%s: %v", path, err)
			return
		}
		fv.Set(reflect.ValueOf(sec))
		b.secrets = append(b.secrets, secretSlot{path: path, ptr: fv.Addr().Interface().(*Secret)})
		return

	case fv.Kind() == reflect.Struct && fv.Type() != durationType:
		sub := map[string]any{}
		if present {
			m, ok := raw.(map[string]any)
			if !ok {
				b.errf("%s: expected a mapping", path)
				return
			}
			sub = m
		}
		b.bindStruct(fv, sub, path)
		b.reportLeftovers(sub, path)
		return

	case fv.Kind() == reflect.Pointer:
		// Optional values: nil when absent, allocated when any layer (or a
		// default tag) supplies one — gives products tri-state fields
		// (*bool unset≠false) and optional sub-structs (review finding ARCH-6).
		elemT := fv.Type().Elem()
		if elemT == secretType || elemT == namespacesType {
			b.errf("%s: use %s by value, not a pointer", path, elemT)
			return
		}
		if !present {
			if def, ok := f.Tag.Lookup("default"); ok {
				elem := reflect.New(elemT)
				if err := convertValue(elem.Elem(), def); err != nil {
					b.errf("%s: invalid compiled default %q: %v", path, def, err)
					return
				}
				fv.Set(elem)
				b.prov[path] = LayerDefault
				return
			}
			if f.Tag.Get("required") == "true" {
				b.errf("%s: required", path)
			}
			return
		}
		elem := reflect.New(elemT)
		if elemT.Kind() == reflect.Struct && elemT != durationType {
			m, ok := raw.(map[string]any)
			if !ok {
				b.errf("%s: expected a mapping", path)
				return
			}
			b.bindStruct(elem.Elem(), m, path)
			b.reportLeftovers(m, path)
		} else if err := convertValue(elem.Elem(), raw); err != nil {
			b.errf("%s: %v", path, err)
			return
		}
		fv.Set(elem)
		return
	}

	// Scalars and slices.
	if !present {
		if def, ok := f.Tag.Lookup("default"); ok {
			if err := convertValue(fv, def); err != nil {
				b.errf("%s: invalid compiled default %q: %v", path, def, err)
				return
			}
			b.prov[path] = LayerDefault
			return
		}
		if f.Tag.Get("required") == "true" {
			b.errf("%s: required", path)
		}
		return
	}
	if err := convertValue(fv, raw); err != nil {
		b.errf("%s: %v", path, err)
		return
	}
}

// enforceUnsafe applies the dev-only-knob rule to the FINAL bound values,
// after every layer and every default has been applied. Enforcing here — not
// inline during binding — closes two fail-open holes: a knob whose unsafe
// value is its compiled default, and unsafe tags on non-scalar (struct,
// Secret, slice, pointer) fields (review findings SEC-3/SEC-4). A knob is
// "set" when its bound value is non-zero; prod refuses to boot, stage warns
// loudly (blueprint 12 §4, D-0015).
func (b *binder) enforceUnsafe(v reflect.Value, path string) {
	if !b.env.IsProd() && b.env != EnvStage {
		return
	}
	t := v.Type()
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		tag := f.Tag.Get("conf")
		if tag == "-" {
			continue
		}
		fv := v.Field(i)
		if f.Anonymous && tag == "" && fv.Kind() == reflect.Struct {
			b.enforceUnsafe(fv, path)
			continue
		}
		key := tag
		if key == "" {
			key = strings.ToLower(f.Name)
		}
		p := joinPath(path, key)
		if f.Tag.Get("unsafe") == "true" && !fv.IsZero() {
			if b.env.IsProd() {
				b.errf("%s: dev-only (unsafe) knob is set — refused when environment=prod", p)
			} else {
				b.warnings = append(b.warnings, fmt.Sprintf("%s: dev-only (unsafe) knob is enabled in stage", p))
			}
			continue
		}
		switch {
		case fv.Type() == secretType || fv.Type() == namespacesType || fv.Type() == durationType:
		case fv.Kind() == reflect.Struct:
			b.enforceUnsafe(fv, p)
		case fv.Kind() == reflect.Pointer && !fv.IsNil() && fv.Elem().Kind() == reflect.Struct && fv.Type().Elem() != durationType:
			b.enforceUnsafe(fv.Elem(), p)
		}
	}
}

// reportLeftovers turns every key still present in tree into an unknown-key
// error with its full dotted path (typo defense, blueprint 12 §4).
func (b *binder) reportLeftovers(tree map[string]any, path string) {
	keys := make([]string, 0, len(tree))
	for k := range tree {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		p := joinPath(path, k)
		if sub, ok := tree[k].(map[string]any); ok && len(sub) > 0 {
			b.reportLeftovers(sub, p)
			continue
		}
		b.errf("%s: unknown key", p)
	}
}

// convertValue converts a tree value (typed YAML scalar, or raw string from
// an env var, flag, or default tag) into the target field. Strict: no silent
// truncation, durations only as strings ("5s"), no scalar↔mapping coercion.
// Error messages never echo the raw value — a mis-placed secret in a
// non-Secret field must not leak through boot diagnostics (SEC-8).
func convertValue(v reflect.Value, raw any) error {
	if _, isMap := raw.(map[string]any); isMap {
		return fmt.Errorf("expected a scalar value, got a mapping")
	}
	if v.CanAddr() && v.Type() != secretType {
		if tu, ok := v.Addr().Interface().(encoding.TextUnmarshaler); ok {
			if s, ok := raw.(string); ok {
				return tu.UnmarshalText([]byte(s))
			}
		}
	}
	if v.Type() == durationType {
		s, ok := raw.(string)
		if !ok {
			return fmt.Errorf("want a duration string like \"5s\", got %T", raw)
		}
		d, err := time.ParseDuration(s)
		if err != nil {
			return fmt.Errorf("value is not a valid duration (want e.g. \"5s\", \"1m30s\")")
		}
		v.SetInt(int64(d))
		return nil
	}

	switch v.Kind() {
	case reflect.String:
		if s, ok := raw.(string); ok {
			v.SetString(s)
			return nil
		}
		v.SetString(fmt.Sprint(raw))
		return nil
	case reflect.Bool:
		switch x := raw.(type) {
		case bool:
			v.SetBool(x)
			return nil
		case string:
			p, err := strconv.ParseBool(x)
			if err != nil {
				return fmt.Errorf("value is not a valid bool")
			}
			v.SetBool(p)
			return nil
		}
		return fmt.Errorf("cannot use %T value as bool", raw)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := toInt64(raw)
		if err != nil {
			return err
		}
		if v.OverflowInt(i) {
			return fmt.Errorf("%d overflows %s", i, v.Type())
		}
		v.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := toInt64(raw)
		if err != nil {
			return err
		}
		if i < 0 || v.OverflowUint(uint64(i)) {
			return fmt.Errorf("%d out of range for %s", i, v.Type())
		}
		v.SetUint(uint64(i))
		return nil
	case reflect.Float32, reflect.Float64:
		switch x := raw.(type) {
		case float64:
			v.SetFloat(x)
			return nil
		case int:
			v.SetFloat(float64(x))
			return nil
		case int64:
			v.SetFloat(float64(x))
			return nil
		case string:
			f, err := strconv.ParseFloat(x, 64)
			if err != nil {
				return fmt.Errorf("value is not a valid number")
			}
			v.SetFloat(f)
			return nil
		}
		return fmt.Errorf("cannot use %T value as number", raw)
	case reflect.Slice:
		var items []any
		switch x := raw.(type) {
		case []any:
			items = x
		case string:
			if x != "" {
				for part := range strings.SplitSeq(x, ",") {
					items = append(items, strings.TrimSpace(part))
				}
			}
		default:
			return fmt.Errorf("expected a list")
		}
		out := reflect.MakeSlice(v.Type(), len(items), len(items))
		for i, it := range items {
			if err := convertValue(out.Index(i), it); err != nil {
				return fmt.Errorf("[%d]: %w", i, err)
			}
		}
		v.Set(out)
		return nil
	}
	return fmt.Errorf("unsupported config field type %s", v.Type())
}

func toInt64(raw any) (int64, error) {
	switch x := raw.(type) {
	case int:
		return int64(x), nil
	case int64:
		return x, nil
	case uint64:
		if x > 1<<63-1 {
			return 0, fmt.Errorf("%d overflows int64", x)
		}
		return int64(x), nil
	case float64:
		i := int64(x)
		if float64(i) != x {
			return 0, fmt.Errorf("value is not a whole number")
		}
		return i, nil
	case string:
		i, err := strconv.ParseInt(x, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("value is not a valid integer")
		}
		return i, nil
	}
	return 0, fmt.Errorf("cannot use %T value as integer", raw)
}
