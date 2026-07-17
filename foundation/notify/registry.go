// Package notify is wowapi's notification framework: modules register template
// keys with an allowlisted variable set and required channels (Registry);
// Send writes a notifications row + one notification_deliveries row per
// resolved channel inside the caller's tenant business transaction (atomicity
// with the business write); and SendPending is the async worker step that
// claims queued deliveries, calls channel-specific senders, and advances
// delivery status — dead-lettering after maxAttempts. In-app deliveries are
// rows queried by a future /notifications API. Contract: blueprint 07 §5.
package notify

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"text/template/parse"

	"github.com/qatoolist/wowapi/v2/internal/sealer"

	htmltemplate "html/template"

	texttemplate "text/template"

	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
)

// Channel is a notification delivery channel.
type Channel string

const (
	ChannelInApp    Channel = "inapp"
	ChannelEmail    Channel = "email"
	ChannelSMS      Channel = "sms"
	ChannelWhatsApp Channel = "whatsapp"
	ChannelPush     Channel = "push"
)

// Importance ranks how critical a notification is.
type Importance string

const (
	ImportanceNormal    Importance = "normal"
	ImportanceImportant Importance = "important"
	// ImportanceLegal requires an audit trail on delivery (blueprint 07 §5).
	ImportanceLegal Importance = "legal"
)

// keyRE constrains template keys to module.area.name (same format as rules).
var keyRE = regexp.MustCompile(`^[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*$`)

// TemplateSpec is the static declaration a module makes for a template key.
// Vars lists every variable name the template body is permitted to reference;
// the body must not reference any name outside this set (ValidateBody enforces
// this at seed time). Channels lists channels the module expects templates for.
type TemplateSpec struct {
	Key      string
	Vars     []string // allowlisted variable names — {{.Name}} needs "Name" here
	Channels []string // expected channel names (informational; guides seeding)
}

func (s TemplateSpec) allowsVar(v string) bool {
	for _, allowed := range s.Vars {
		if allowed == v {
			return true
		}
	}
	return false
}

// Registry holds TemplateSpec declarations made by modules at boot. Keys must
// be module.area.name and a module may only register keys with its own prefix.
type Registry struct {
	specs  map[string]TemplateSpec
	errs   []error
	sealed bool
}

// NewRegistry returns an empty template registry.
func NewRegistry() *Registry { return &Registry{specs: map[string]TemplateSpec{}} }

// Seal freezes the registry once boot validation completes: any later Register
// panics rather than silently adding a template the boot gates never saw
// (closure review 2026-07-17, F-10).
// The sealer.Authority parameter restricts sealing to the framework's boot
// path: internal/sealer is unimportable outside the wowapi module, so a
// product module cannot prematurely seal a shared registry during Register.
func (r *Registry) Seal(sealer.Authority) { r.sealed = true }

// Register records a template key's spec. Errors (bad key, prefix mismatch,
// duplicate) accumulate and are returned by Err().
func (r *Registry) Register(module string, spec TemplateSpec) {
	if r.sealed {
		panic("notify: template registration after boot: the extension model is sealed")
	}
	if !keyRE.MatchString(spec.Key) {
		r.errf("notify template key must be module.area.name: %s", spec.Key)
		return
	}
	prefix := module + "."
	if !strings.HasPrefix(spec.Key, prefix) {
		r.errf("module %s may not register notify key %s", module, spec.Key)
		return
	}
	if _, dup := r.specs[spec.Key]; dup {
		r.errf("notify template key registered more than once: %s", spec.Key)
		return
	}
	r.specs[spec.Key] = spec.clone()
}

// clone returns a deep copy of s: the registry must not share the Vars and
// Channels slices with callers in either direction — a retained registration
// value or a mutated Get result must never change a validated template's
// variable allowlist (second closure audit 2026-07-17, F-10).
func (s TemplateSpec) clone() TemplateSpec {
	out := s
	if s.Vars != nil {
		out.Vars = append([]string(nil), s.Vars...)
	}
	if s.Channels != nil {
		out.Channels = append([]string(nil), s.Channels...)
	}
	return out
}

// Get returns the spec for a key (a deep copy — mutating its nested fields
// cannot alter the registry).
func (r *Registry) Get(key string) (TemplateSpec, bool) {
	s, ok := r.specs[key]
	if !ok {
		return TemplateSpec{}, false
	}
	return s.clone(), true
}

// Keys returns registered keys, sorted.
func (r *Registry) Keys() []string {
	out := make([]string, 0, len(r.specs))
	for k := range r.specs {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func (r *Registry) errf(format string, args ...any) {
	r.errs = append(r.errs, kerr.E(kerr.KindInternal, "invalid_notify_spec", fmt.Sprintf(format, args...)))
}

// Err returns accumulated registration errors joined, or nil.
func (r *Registry) Err() error {
	if len(r.errs) == 0 {
		return nil
	}
	msgs := make([]string, len(r.errs))
	for i, e := range r.errs {
		msgs[i] = e.Error()
	}
	joined := msgs[0]
	for i := 1; i < len(msgs); i++ {
		joined += "; " + msgs[i]
	}
	return kerr.E(kerr.KindInternal, "notify_registration_failed", "notify template registration failed: "+joined)
}

// ValidateBody parses the template body and verifies every top-level field
// reference ({{.VarName}}) is declared in spec.Vars. Call this at seed time
// when writing a template row to the database — blueprint 07 §5: "must fail at
// REGISTER/seed-validation time (not at send time)". Returns KindValidation on
// violation.
func ValidateBody(spec TemplateSpec, body string) error {
	vars, err := extractTemplateVars(body)
	if err != nil {
		return kerr.E(kerr.KindValidation, "template_parse_error",
			"notify: template body is not valid Go text/template: "+err.Error())
	}
	for _, v := range vars {
		if !spec.allowsVar(v) {
			return kerr.E(kerr.KindValidation, "template_var_not_in_allowlist",
				fmt.Sprintf("notify: template key %s references variable %q which is not in the allowlist", spec.Key, v))
		}
	}
	return nil
}

// extractTemplateVars parses a Go text/template body and returns the distinct
// top-level field names referenced as {{.FieldName}}. It walks the parse tree
// recursively; nested paths like {{.Foo.Bar}} contribute only the top-level
// name "Foo".
func extractTemplateVars(body string) ([]string, error) {
	t, err := texttemplate.New("t").Parse(body)
	if err != nil {
		return nil, err
	}
	if t.Tree == nil || t.Root == nil {
		return nil, nil
	}
	seen := map[string]bool{}
	var vars []string

	var walk func(parse.Node)
	walkPipe := func(p *parse.PipeNode) {
		if p == nil {
			return
		}
		for _, cmd := range p.Cmds {
			for _, arg := range cmd.Args {
				walk(arg)
			}
		}
	}

	walk = func(n parse.Node) {
		if n == nil {
			return
		}
		switch v := n.(type) {
		case *parse.FieldNode:
			if len(v.Ident) > 0 && !seen[v.Ident[0]] {
				seen[v.Ident[0]] = true
				vars = append(vars, v.Ident[0])
			}
		case *parse.ListNode:
			if v != nil {
				for _, child := range v.Nodes {
					walk(child)
				}
			}
		case *parse.ActionNode:
			walkPipe(v.Pipe)
		case *parse.IfNode:
			walkPipe(v.Pipe)
			walk(v.List)
			walk(v.ElseList)
		case *parse.RangeNode:
			walkPipe(v.Pipe)
			walk(v.List)
			walk(v.ElseList)
		case *parse.WithNode:
			walkPipe(v.Pipe)
			walk(v.List)
			walk(v.ElseList)
		case *parse.TemplateNode:
			walkPipe(v.Pipe)
		case *parse.ChainNode:
			walk(v.Node)
		case *parse.PipeNode:
			walkPipe(v)
		}
	}

	walk(t.Root)
	return vars, nil
}

// renderBody executes body with vars as the dot data context. It rejects any
// key in vars that is not in spec.Vars (KindValidation) and uses
// missingkey=error so a template referencing a var absent from vars also fails.
//
// SEC-51: the email channel renders through html/template, which contextually
// auto-escapes variable values, so a variable carrying markup ("<script>…")
// cannot inject active content into the recipient's mail client. Non-HTML
// channels (sms, whatsapp, push, inapp) render through text/template — their
// transports are plain text, so escaping would corrupt legitimate values.
func renderBody(spec TemplateSpec, channel Channel, body string, vars map[string]any) (string, error) {
	for k := range vars {
		if !spec.allowsVar(k) {
			return "", kerr.E(kerr.KindValidation, "template_var_not_in_allowlist",
				fmt.Sprintf("notify: variable %q is not allowlisted for key %s", k, spec.Key))
		}
	}
	var sb strings.Builder
	if channel == ChannelEmail {
		tmpl, err := htmltemplate.New("t").Option("missingkey=error").Parse(body)
		if err != nil {
			return "", kerr.E(kerr.KindInternal, "template_parse_error",
				"notify: template parse failed: "+err.Error())
		}
		if err := tmpl.Execute(&sb, vars); err != nil {
			return "", kerr.E(kerr.KindValidation, "template_render_error",
				"notify: template execution failed: "+err.Error())
		}
		return sb.String(), nil
	}
	tmpl, err := texttemplate.New("t").Option("missingkey=error").Parse(body)
	if err != nil {
		return "", kerr.E(kerr.KindInternal, "template_parse_error",
			"notify: template parse failed: "+err.Error())
	}
	if err := tmpl.Execute(&sb, vars); err != nil {
		return "", kerr.E(kerr.KindValidation, "template_render_error",
			"notify: template execution failed: "+err.Error())
	}
	return sb.String(), nil
}

// RenderBody is the exported version of renderBody for use by ChannelSender
// adapters (e.g. smtp, sms) that need to render a body fetched from the DB
// before dispatching. channel selects the escaping context (see renderBody):
// email → html/template (auto-escaped), everything else → text/template.
func RenderBody(spec TemplateSpec, channel Channel, body string, vars map[string]any) (string, error) {
	return renderBody(spec, channel, body, vars)
}
