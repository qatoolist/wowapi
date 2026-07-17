// i18n_cmd.go — `wowapi i18n validate` (B1). Mirrors `wowapi seed validate`
// (seed_cmd.go): it loads a product's locale catalogs through the SAME
// kernel/i18n loader the app uses at boot, so an authoring defect — missing
// locale coverage, a duplicate key within a layer, an unauthorized kernel.*
// write, or a placeholder-arity mismatch between a translation and the framework
// default — is caught in CI (exit 1) instead of at deploy time.
//
// It runs over the framework's embedded defaults PLUS the product's locales/
// tree (loaded as a sanctioned framework-override fs layer, so a product-local
// kernel.* override is checked, not rejected), which is exactly the layer stack
// the generated binaries build from the product's i18n config. It needs no
// database and no product config.
package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/qatoolist/wowapi/kernel/i18n"
)

func i18nUsage(w io.Writer) {
	fmt.Fprint(w, `usage: wowapi i18n validate [flags]

Subcommands:
  validate   load and validate a product's locale catalogs (no database needed)

Flags (validate):
  --dir            directory holding the product's locale catalogs (default "locales")
  --default-locale fallback locale every key must define (default "en")
  --supported      comma-separated supported locales (default: the default locale)
  --strict         treat a supported locale missing a key (that the default
                   locale defines, so it still falls back) as a failure, not a pass

Checks: intra-layer duplicate keys, kernel.* namespace ownership (a product may
retranslate framework strings but not invent new kernel.* keys), locale coverage
(every key present in the default locale; with --strict, in every supported
locale), and placeholder compatibility (a translation's percent-verb count must
match the framework/default template's, so a localized parameterised message
can't drop or add a parameter).

Scope: this command validates the framework defaults + the --dir file catalogs
(YAML/JSON). It does NOT compile or load your internal/i18n/catalogs Go bundles —
those are compiled product code and are validated at BOOT (a bad Go-bundle key
fails app.Boot), not by this CLI. Keep Go bundles small, or add a product-side
test that calls kernel/i18n.Validate with your compiled Go layer.
`)
}

func runI18n(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		i18nUsage(stderr)
		return 2
	}
	switch args[0] {
	case "validate":
		return runI18nValidate(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		i18nUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "wowapi i18n: unknown subcommand %q\n", args[0])
		i18nUsage(stderr)
		return 2
	}
}

func runI18nValidate(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("wowapi i18n validate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dir := fs.String("dir", "locales", "directory holding the product's locale catalogs")
	defLocale := fs.String("default-locale", i18n.DefaultLocale, "fallback locale every key must define")
	supported := fs.String("supported", "", "comma-separated supported locales")
	strict := fs.Bool("strict", false, "treat a fallback-covered supported-locale gap as a failure")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if info, err := os.Stat(*dir); err != nil || !info.IsDir() {
		fmt.Fprintf(stderr, "wowapi i18n validate: %s is not a directory\n", *dir)
		return 1
	}

	supportedLocales := []string{*defLocale}
	if *supported != "" {
		supportedLocales = nil
		for _, s := range strings.Split(*supported, ",") {
			if s = strings.TrimSpace(s); s != "" {
				supportedLocales = append(supportedLocales, s)
			}
		}
	}

	// Frame the product's locales/ tree as a sanctioned framework-override fs
	// layer on top of the embedded framework defaults — the same stack the
	// generated binaries load — so kernel.* overrides are validated (not rejected)
	// and framework keys count toward coverage.
	layers := []i18n.Layer{
		{
			Name:    "framework defaults",
			Policy:  i18n.Policy{OwnsFramework: true},
			Sources: []i18n.Source{i18n.FrameworkDefaultsSource()},
		},
		{
			Name:    "product catalogs",
			Policy:  i18n.Policy{OwnsFramework: true, FrameworkOverrideOnly: true},
			Sources: []i18n.Source{i18n.NewFSSource(os.DirFS(*dir), ".", "yaml", "json")},
		},
	}

	rep, err := i18n.Validate(i18n.ValidateOptions{
		DefaultLocale:    *defLocale,
		SupportedLocales: supportedLocales,
		StrictCoverage:   *strict,
	}, layers...)
	if err != nil {
		fmt.Fprintf(stderr, "wowapi i18n validate: %v\n", err)
		return 1
	}
	if !rep.OK() {
		fmt.Fprintf(stderr, "wowapi i18n validate: %d problem(s):\n", len(rep.Problems))
		for _, p := range rep.Problems {
			fmt.Fprintf(stderr, "  - %s\n", p)
		}
		return 1
	}
	fmt.Fprintf(stdout, "OK: i18n catalogs valid — %d keys across locales %s\n",
		rep.Keys, strings.Join(rep.Locales, ", "))
	return 0
}
