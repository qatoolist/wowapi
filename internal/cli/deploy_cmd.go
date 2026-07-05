// deploy_cmd.go — `wowapi deploy render` (Phase 10). Renders a deployment
// manifest (docker-compose or a plain env file) for the api/worker/migrate
// processes from a small set of flags. The DB DSN is emitted as a
// secretref://env/… reference (config.DB.DSN is a Secret), never an inlined
// value — the manifest is safe to commit.
package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/qatoolist/wowapi/kernel/config"
)

func deployUsage(w io.Writer) {
	fmt.Fprint(w, `usage: wowapi deploy render [flags]

Render a deployment manifest for the api/worker/migrate processes.

Flags:
  --format   compose | env   (default "compose")
  --name     deployment/service base name (default "app")
  --image    container image (default "app:latest")
  --env      target environment: local|dev|stage|prod (default "prod")
  --out      output file (default: stdout)
`)
}

type deployVars struct {
	Name  string
	Image string
	Env   string
}

func runDeploy(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || (args[0] != "render" && args[0] != "-h" && args[0] != "--help" && args[0] != "help") {
		deployUsage(stderr)
		return 2
	}
	if args[0] != "render" {
		deployUsage(stdout)
		return 0
	}
	fs := flag.NewFlagSet("wowapi deploy render", flag.ContinueOnError)
	fs.SetOutput(stderr)
	format := fs.String("format", "compose", "compose | env")
	name := fs.String("name", "app", "deployment base name")
	image := fs.String("image", "app:latest", "container image")
	env := fs.String("env", "prod", "target environment: local|dev|stage|prod")
	out := fs.String("out", "", "output file (default stdout)")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}

	// The rendered WOWAPI__ENVIRONMENT must be a value the config loader accepts
	// (Env.Valid), or the deployed process fails at startup. Reject an invalid
	// --env here rather than emitting a manifest that cannot boot.
	if !config.Env(*env).Valid() {
		fmt.Fprintf(stderr, "wowapi deploy render: invalid --env %q (want local|dev|stage|prod)\n", *env)
		return 2
	}

	var tmpl string
	switch *format {
	case "compose":
		tmpl = composeTemplate
	case "env":
		tmpl = envTemplate
	default:
		fmt.Fprintf(stderr, "wowapi deploy render: unknown --format %q (want compose|env)\n", *format)
		return 2
	}

	t, err := template.New("deploy").Parse(tmpl)
	if err != nil {
		fmt.Fprintf(stderr, "wowapi deploy render: %v\n", err)
		return 1
	}
	var sb strings.Builder
	if err := t.Execute(&sb, deployVars{Name: *name, Image: *image, Env: *env}); err != nil {
		fmt.Fprintf(stderr, "wowapi deploy render: %v\n", err)
		return 1
	}
	if *out == "" {
		if _, err := io.WriteString(stdout, sb.String()); err != nil {
			fmt.Fprintf(stderr, "wowapi deploy render: write: %v\n", err)
			return 1
		}
		return 0
	}
	if err := os.WriteFile(*out, []byte(sb.String()), 0o644); err != nil {
		fmt.Fprintf(stderr, "wowapi deploy render: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, *out)
	return 0
}

const composeTemplate = `# Rendered by ` + "`wowapi deploy render`" + ` — deployment manifest for {{.Name}} ({{.Env}}).
# Every DB DSN is a secretref://env/… reference (config.DB.* are Secrets): the real
# DSNs live in the environment, never inlined here. api/worker need BOTH the runtime
# DSN (non-privileged app_rt) and the platform DSN (dedicated app_platform login) —
# they fail closed at startup without db.platform_dsn. migrate needs the migrate DSN.
services:
  {{.Name}}-api:
    image: {{.Image}}
    command: ["/app/api"]
    environment:
      WOWAPI__ENVIRONMENT: {{.Env}}
      WOWAPI__DB__DSN: secretref://env/WOWAPI_DB_DSN
      WOWAPI__DB__PLATFORM_DSN: secretref://env/WOWAPI_PLATFORM_DSN
    ports: ["8080:8080"]
    restart: unless-stopped
  {{.Name}}-worker:
    image: {{.Image}}
    command: ["/app/worker"]
    environment:
      WOWAPI__ENVIRONMENT: {{.Env}}
      WOWAPI__DB__DSN: secretref://env/WOWAPI_DB_DSN
      WOWAPI__DB__PLATFORM_DSN: secretref://env/WOWAPI_PLATFORM_DSN
    restart: unless-stopped
  {{.Name}}-migrate:
    image: {{.Image}}
    command: ["/app/migrate"]
    environment:
      WOWAPI__ENVIRONMENT: {{.Env}}
      WOWAPI__DB__MIGRATE_DSN: secretref://env/WOWAPI_MIGRATE_DSN
    restart: "no"
`

const envTemplate = `# Rendered by ` + "`wowapi deploy render`" + ` — {{.Name}} deployment env ({{.Env}}).
# All three WOWAPI__DB__* values are secret REFERENCES: set the real DSNs in the
# named env vars, never inline them. api/worker need the runtime DSN (app_rt) AND
# the platform DSN (app_platform) — they fail closed without db.platform_dsn; the
# migrate job needs the migrate DSN (app_migrate).
WOWAPI__ENVIRONMENT={{.Env}}
WOWAPI__DB__DSN=secretref://env/WOWAPI_DB_DSN
WOWAPI__DB__PLATFORM_DSN=secretref://env/WOWAPI_PLATFORM_DSN
WOWAPI__DB__MIGRATE_DSN=secretref://env/WOWAPI_MIGRATE_DSN
WOWAPI_IMAGE={{.Image}}
`
