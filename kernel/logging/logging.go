// Package logging provides process-wide structured logging for wowapi processes.
// It constructs an *slog.Logger from a kernel/config.Log and installs a
// defense-in-depth ReplaceAttr that catches accidental raw-string sensitive
// attributes. The primary redaction mechanism is structural — config.Secret
// implements slog.LogValuer — this handler is a second line of defense only.
//
// See docs/blueprint/12-configuration-and-deployment.md §7 for the startup
// config-fingerprint requirement satisfied by LogStartup.
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/qatoolist/wowapi/kernel/config"
)

// sensitiveKeySuffixes is the heuristic list of key suffixes (case-insensitive)
// whose attribute values are replaced with "[redacted]" by redactAttr. Both
// exact matches and suffix matches apply: "db_password" matches "password".
//
// This is defense-in-depth only — not the security boundary. Structural
// config.Secret redaction via slog.LogValuer is the primary mechanism.
var sensitiveKeySuffixes = []string{
	"password", "passwd", "secret", "token",
	"api_key", "apikey", "authorization",
	"credential", "credentials", "dsn", "private_key",
}

// isSensitiveKey reports whether the key (already lowercased) equals or ends
// with one of the sensitive suffixes.
func isSensitiveKey(lower string) bool {
	for _, s := range sensitiveKeySuffixes {
		if strings.HasSuffix(lower, s) {
			return true
		}
	}
	return false
}

// redactAttr is the slog handler ReplaceAttr function. It replaces the value
// of ANY attribute — regardless of kind (string, int, bool, duration, …) —
// whose key (case-insensitive) equals or ends with a sensitive suffix.
// Group attrs are passed through unchanged; groups are never prepended to the
// key being tested (only the immediate attr key is matched).
//
// This is defense-in-depth only — not the security boundary. Structural
// config.Secret redaction via slog.LogValuer is the primary mechanism; this
// handler catches accidental raw non-string values (e.g. numeric tokens,
// duration-typed secrets) that bypass the string-only check.
func redactAttr(_ []string, a slog.Attr) slog.Attr {
	if a.Value.Kind() == slog.KindGroup {
		return a
	}
	if isSensitiveKey(strings.ToLower(a.Key)) {
		return slog.String(a.Key, "[redacted]")
	}
	return a
}

// parseLevel maps the config level string to a slog.Level.
// Returns an error for unrecognized values — the config loader validates too;
// this is a second line of defense, not a silent fallback.
func parseLevel(level string) (slog.Level, error) {
	switch level {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("logging: level %q is not one of debug|info|warn|error", level)
	}
}

// New constructs an *slog.Logger writing to w, configured from cfg.
// Format "json" produces slog.NewJSONHandler; "text" produces slog.NewTextHandler.
// An unknown level or format is an error — both conditions are also caught by
// config.Framework.Validate, but New enforces them independently.
func New(w io.Writer, cfg config.Log) (*slog.Logger, error) {
	lvl, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	opts := &slog.HandlerOptions{
		Level:       lvl,
		ReplaceAttr: redactAttr,
	}
	var h slog.Handler
	switch cfg.Format {
	case "json":
		h = slog.NewJSONHandler(w, opts)
	case "text":
		h = slog.NewTextHandler(w, opts)
	default:
		return nil, fmt.Errorf("logging: format %q is not one of json|text", cfg.Format)
	}
	return slog.New(h), nil
}

// LogStartup emits a single Info record "starting" with the process name,
// deployment environment, and the full and short config fingerprints.
// This satisfies blueprint 12 §7: each process logs its config fingerprint at
// boot so drift between api/worker/migrate can be detected cheaply.
func LogStartup(l *slog.Logger, process string, env config.Env, fp config.Fingerprint) {
	l.Info("starting",
		"process", process,
		"environment", string(env),
		"config_fingerprint", fp.String(),
		"config_fingerprint_short", fp.Short(),
	)
}
