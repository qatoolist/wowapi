package i18n

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"sort"
	"strings"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"gopkg.in/yaml.v3"
)

// catalogFile is the on-disk shape of a product/framework YAML or JSON catalog
// file. locale is optional when the locale is implied by the path (a file under
// "<root>/<locale>/..." or named "<root>/<locale>.json"); when present it must
// agree with the path-derived locale, catching a mislabeled file.
type catalogFile struct {
	Locale   string            `yaml:"locale" json:"locale"`
	Messages map[string]string `yaml:"messages" json:"messages"`
}

// fsSource loads catalog files from an fs.FS. It backs three of the four
// first-class source kinds: the embedded framework defaults (KindFrameworkDefaults),
// product framework-override files, and product/module catalog files (both KindFS).
// It reads *.yaml/*.yml (when yaml is set) and *.json (when json is set) and
// derives each file's locale from its path:
//
//   - "<root>/<locale>/<name>.yaml"  -> locale = <locale> (directory layout)
//   - "<root>/<locale>.json"          -> locale = <locale> (flat layout)
//
// A file's optional "locale:" field, if set, must match the path-derived locale.
// Duplicate keys WITHIN a single file are a load error (typo defense); duplicates
// ACROSS files are the Loader's intra-layer concern.
type fsSource struct {
	fsys   fs.FS
	root   string
	kind   SourceKind
	label  string // provenance prefix; if empty, the file path is used
	yaml   bool
	json   bool
	labelP bool // when true, prefix origins with label (embedded case)
}

// NewFSSource returns a Source that loads YAML and JSON catalog files from fsys
// rooted at root (e.g. os.DirFS(productRoot) with root "locales"). formats picks
// which extensions are read; pass both "yaml" and "json" for the canonical
// mixed layout. It is used for product framework-override files and for
// product/module catalogs — the Loader's Layer.Policy decides which namespaces
// the loaded keys may occupy.
func NewFSSource(fsys fs.FS, root string, formats ...string) Source {
	s := &fsSource{fsys: fsys, root: root, kind: KindFS}
	if len(formats) == 0 {
		s.yaml, s.json = true, true
	}
	for _, f := range formats {
		switch strings.ToLower(f) {
		case "yaml", "yml":
			s.yaml = true
		case "json":
			s.json = true
		}
	}
	return s
}

func (s *fsSource) Kind() SourceKind { return s.kind }

func (s *fsSource) Load() ([]RawBundle, error) {
	// Accumulate per (locale,origin) so each file becomes one RawBundle with its
	// own provenance, preserving the intra-file duplicate check and giving the
	// Loader precise origins.
	var bundles []RawBundle
	var errs []string

	walkErr := fs.WalkDir(s.fsys, s.root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			// A missing root is not an error: a product may configure an fs source
			// before creating any files. Report other walk errors.
			if _, statErr := fs.Stat(s.fsys, s.root); statErr != nil {
				return fs.SkipAll
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(path.Ext(p))
		isYAML := ext == ".yaml" || ext == ".yml"
		isJSON := ext == ".json"
		if (isYAML && !s.yaml) || (isJSON && !s.json) || (!isYAML && !isJSON) {
			return nil
		}
		locale, lerr := s.localeFor(p)
		if lerr != "" {
			errs = append(errs, lerr)
			return nil
		}
		data, rerr := fs.ReadFile(s.fsys, p)
		if rerr != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", s.origin(p), rerr))
			return nil
		}
		cf, derr := decodeCatalogFile(data, isJSON)
		if derr != "" {
			errs = append(errs, fmt.Sprintf("%s: %s", s.origin(p), derr))
			return nil
		}
		if cf.Locale != "" && cf.Locale != locale {
			errs = append(errs, fmt.Sprintf("%s: locale field %q disagrees with path locale %q",
				s.origin(p), cf.Locale, locale))
			return nil
		}
		if len(cf.Messages) == 0 {
			return nil
		}
		bundles = append(bundles, RawBundle{Locale: locale, Messages: cf.Messages, Origin: s.origin(p)})
		return nil
	})
	if walkErr != nil {
		errs = append(errs, walkErr.Error())
	}
	if len(errs) > 0 {
		sort.Strings(errs)
		return nil, kerr.E(kerr.KindInternal, "invalid_i18n_source", strings.Join(errs, "; "))
	}
	// Deterministic order so error messages and merges are stable.
	sort.Slice(bundles, func(i, j int) bool { return bundles[i].Origin < bundles[j].Origin })
	return bundles, nil
}

// localeFor derives the locale from a file path relative to root. Returns an
// error string if the path shape is not recognized.
func (s *fsSource) localeFor(p string) (locale string, errStr string) {
	rel := strings.TrimPrefix(p, s.root+"/")
	rel = strings.TrimPrefix(rel, s.root) // handles root == "."
	rel = strings.TrimPrefix(rel, "/")
	segs := strings.Split(rel, "/")
	if len(segs) >= 2 {
		// "<locale>/<...>" directory layout.
		return segs[0], ""
	}
	// Flat "<locale>.<ext>" layout.
	base := segs[0]
	ext := path.Ext(base)
	name := strings.TrimSuffix(base, ext)
	if name == "" {
		return "", fmt.Sprintf("%s: cannot derive locale from path", s.origin(p))
	}
	return name, ""
}

func (s *fsSource) origin(p string) string {
	if s.labelP && s.label != "" {
		return s.label + " (" + p + ")"
	}
	return p
}

// decodeCatalogFile strict-decodes one catalog file. Unknown top-level keys are
// errors (typo defense). Duplicate keys within one YAML file are rejected by the
// yaml.v3 decoder; encoding/json, by contrast, silently keeps the last value for
// a repeated JSON object key (the stdlib gives no hook to detect it), so JSON
// files do NOT get intra-file duplicate detection — prefer YAML when authoring by
// hand, and rely on cross-file intra-layer duplicate detection (the Loader) plus
// `wowapi i18n validate` for the coverage/ownership guarantees.
func decodeCatalogFile(data []byte, isJSON bool) (catalogFile, string) {
	var cf catalogFile
	if isJSON {
		dec := json.NewDecoder(strings.NewReader(string(data)))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&cf); err != nil {
			// An empty (or whitespace-only) JSON file decodes to io.EOF; treat it as
			// an empty catalog, mirroring the YAML branch's EOF tolerance below, so a
			// placeholder file does not fail boot asymmetrically (review B1-corr #1).
			if errors.Is(err, io.EOF) {
				return catalogFile{}, ""
			}
			return cf, fmt.Sprintf("invalid JSON: %v", err)
		}
		return cf, ""
	}
	dec := yaml.NewDecoder(strings.NewReader(string(data)))
	dec.KnownFields(true)
	if err := dec.Decode(&cf); err != nil && err.Error() != "EOF" {
		return cf, fmt.Sprintf("invalid YAML: %v", err)
	}
	return cf, ""
}
