package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// The merged configuration tree is a plain map[string]any built fresh per
// Load call: nested mappings are map[string]any, scalars keep their YAML
// types, env-var/flag values stay strings (the binder converts per field).
// Because every map in the tree is freshly allocated here, the binder may
// consume it destructively for unknown-key detection.

func parseYAMLFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		// yaml errors embed node content in backticks; a malformed raw value
		// in a config file may itself be a secret, so scrub the payloads and
		// keep only positions (SEC-7). Redaction must survive operator error.
		return nil, fmt.Errorf("config: %s: %s", path, scrubYAMLError(err))
	}
	if m == nil {
		m = map[string]any{}
	}
	return m, nil
}

var yamlPayloadRE = regexp.MustCompile("`[^`]*`")

func scrubYAMLError(err error) string {
	return yamlPayloadRE.ReplaceAllString(err.Error(), "`…`")
}

// deepCopyTree returns a copy sharing no maps or slices with the input, so
// values captured from the parse tree (module namespaces) cannot alias state
// that outlives or crosses the load (SEC-10).
func deepCopyTree(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = deepCopyValue(v)
	}
	return out
}

func deepCopyValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		return deepCopyTree(x)
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = deepCopyValue(e)
		}
		return out
	}
	return v
}

// deepMerge copies src into dst (later layers win), recording per-leaf
// provenance. Mappings merge recursively; scalars and sequences replace.
func deepMerge(dst, src map[string]any, layer Layer, prov Provenance, path string) {
	for k, v := range src {
		p := joinPath(path, k)
		if vm, ok := v.(map[string]any); ok {
			dm, ok := dst[k].(map[string]any)
			if !ok {
				dm = map[string]any{}
				dst[k] = dm
			}
			deepMerge(dm, vm, layer, prov, p)
			continue
		}
		dst[k] = v
		prov[p] = layer
	}
}

// setPath writes a leaf value at a dotted/segmented path, creating
// intermediate mappings as needed. Non-mapping intermediates are replaced;
// the binder reports the resulting shape errors.
func setPath(dst map[string]any, segs []string, val any, layer Layer, prov Provenance) {
	m := dst
	for _, s := range segs[:len(segs)-1] {
		next, ok := m[s].(map[string]any)
		if !ok {
			next = map[string]any{}
			m[s] = next
		}
		m = next
	}
	m[segs[len(segs)-1]] = val
	prov[strings.Join(segs, ".")] = layer
}

// applyEnviron overlays PREFIX__SECTION__FIELD=value pairs onto the tree.
// Segments are lowercased to match conf keys; values stay strings.
func applyEnviron(dst map[string]any, prefix string, environ []string, prov Provenance) {
	for _, kv := range environ {
		k, v, ok := strings.Cut(kv, "=")
		if !ok || !strings.HasPrefix(k, prefix) {
			continue
		}
		rest := strings.TrimPrefix(k, prefix)
		if rest == "" {
			continue
		}
		setPath(dst, strings.Split(strings.ToLower(rest), "__"), v, LayerEnvVar, prov)
	}
}

// applyFlags overlays dotted-key CLI flag values ("http.addr" → value).
// Flags are local tooling only; the loader refuses them when environment=prod.
func applyFlags(dst map[string]any, flags map[string]string, prov Provenance) {
	for k, v := range flags {
		setPath(dst, strings.Split(k, "."), v, LayerFlag, prov)
	}
}
