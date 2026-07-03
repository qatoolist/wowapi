package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Schema renders a JSON Schema for a config struct from the same `conf`,
// `default`, `required`, `unsafe`, and `doc` tags the binder reads — the
// two can't drift because there is only one tag set. Feeds
// `wowapi config schema` and the product configs/config.schema.json check.
func Schema[T any]() ([]byte, error) {
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("config: Schema target must be a struct, got %s", t)
	}
	s := structSchema(t)
	s["$schema"] = "https://json-schema.org/draft/2020-12/schema"
	return json.MarshalIndent(s, "", "  ")
}

func structSchema(t reflect.Type) map[string]any {
	props := map[string]any{}
	var required []string
	collectProps(t, props, &required)
	s := map[string]any{
		"type":                 "object",
		"properties":           props,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

func collectProps(t reflect.Type, props map[string]any, required *[]string) {
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		tag := f.Tag.Get("conf")
		if tag == "-" {
			continue
		}
		if f.Anonymous && tag == "" && f.Type.Kind() == reflect.Struct {
			collectProps(f.Type, props, required)
			continue
		}
		key := tag
		if key == "" {
			key = strings.ToLower(f.Name)
		}
		fs := fieldSchema(f.Type)
		if doc := f.Tag.Get("doc"); doc != "" {
			fs["description"] = doc
		}
		if def, ok := f.Tag.Lookup("default"); ok {
			fs["default"] = def
		}
		if f.Tag.Get("unsafe") == "true" {
			fs["x-unsafe"] = true
		}
		switch {
		case f.Tag.Get("required") == "true":
			*required = append(*required, key)
		case key == "environment" && f.Type == reflect.TypeFor[Env]():
			// No `required` tag by design (fail-closed rule lives in the
			// loader, D-0010), but the schema must tell the same story
			// (review finding ARCH-9).
			*required = append(*required, key)
			fs["x-fail-closed"] = true
		}
		props[key] = fs
	}
}

func fieldSchema(t reflect.Type) map[string]any {
	switch t {
	case secretType:
		return map[string]any{"type": "string", "pattern": "^secretref://", "x-secret": true}
	case reflect.TypeFor[Env]():
		return map[string]any{"type": "string", "enum": []string{"local", "dev", "stage", "prod"}}
	case durationType:
		return map[string]any{"type": "string", "x-go-type": "time.Duration", "examples": []string{"5s", "1m"}}
	case namespacesType:
		return map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "object"}}
	case reflect.TypeFor[time.Time]():
		return map[string]any{"type": "string", "format": "date-time"}
	}
	switch t.Kind() {
	case reflect.Struct:
		return structSchema(t)
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.Slice:
		return map[string]any{"type": "array", "items": fieldSchema(t.Elem())}
	}
	return map[string]any{}
}
