package config

import "reflect"

// Clone returns a deep copy of the framework config: every nested slice, map,
// and pointer in the value graph is copied, so a captured copy shares no
// mutable storage with the receiver (fifth closure audit 2026-07-17 — boot
// captures the kernel dependency view with an isolated config). A JSON round
// trip cannot do this job: secret fields REDACT on marshal by design.
func (f Framework) Clone() Framework {
	out, ok := deepCopyReflect(reflect.ValueOf(f)).Interface().(Framework)
	if !ok {
		// Unreachable: deepCopyReflect preserves the input type.
		panic("config: Clone produced a non-Framework value")
	}
	return out
}

// deepCopyReflect recursively copies slices, maps, pointers, and struct
// fields. Structs start from a WHOLE-VALUE copy before exported fields are
// deep-copied over it: reflection cannot Set unexported fields, and skipping
// them would silently ZERO value types like config.Secret (whose ref/value
// are unexported strings — immutable, so the value copy is exactly right).
// Scalars and strings copy by value.
func deepCopyReflect(v reflect.Value) reflect.Value {
	// An if-chain rather than a reflect.Kind switch: the exhaustive linter
	// demands every Kind in a tagged switch, and staticcheck QF1002 rejects
	// the equality-only untagged form.
	k := v.Kind()
	if k == reflect.Slice {
		if v.IsNil() {
			return v
		}
		out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		for i := 0; i < v.Len(); i++ {
			out.Index(i).Set(deepCopyReflect(v.Index(i)))
		}
		return out
	}
	if k == reflect.Map {
		if v.IsNil() {
			return v
		}
		out := reflect.MakeMapWithSize(v.Type(), v.Len())
		iter := v.MapRange()
		for iter.Next() {
			out.SetMapIndex(deepCopyReflect(iter.Key()), deepCopyReflect(iter.Value()))
		}
		return out
	}
	if k == reflect.Ptr {
		if v.IsNil() {
			return v
		}
		out := reflect.New(v.Type().Elem())
		out.Elem().Set(deepCopyReflect(v.Elem()))
		return out
	}
	if k == reflect.Struct {
		out := reflect.New(v.Type()).Elem()
		// Whole-value copy FIRST so unexported fields (config.Secret's
		// ref/value) survive; then deep-copy every settable exported field.
		out.Set(v)
		for i := 0; i < v.NumField(); i++ {
			if out.Field(i).CanSet() {
				out.Field(i).Set(deepCopyReflect(v.Field(i)))
			}
		}
		return out
	}
	return v
}
