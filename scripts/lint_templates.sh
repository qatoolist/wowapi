#!/bin/sh
# Generated templates must consume ONLY the boot-validated runtime accessors
# (Runtime*), never Booted's informational mirror fields: a field can be
# reassigned after boot, so a template reading it would serve unvalidated
# state (third/fourth closure audits 2026-07-17, F-10). Aliasing the booted
# variable is forbidden outright — `b := booted` would put field reads out of
# the field check's reach.
#
# Usage: lint_templates.sh [templates-dir]   (exit 1 on any violation)
set -u
dir="${1:-internal/cli/templates}"
fail=0

bad=$(grep -rnE 'booted\.(Kernel|Seeds|I18n|Router|Migrations|Health|Events|Jobs|Recurring|OpenAPI)\b' "$dir" 2>/dev/null || true)
if [ -n "$bad" ]; then
  echo "TEMPLATE VIOLATION (generated template reads an informational Booted field; use the Runtime* accessor):"
  echo "$bad" | sed 's/^/  /'
  fail=1
fi

# Alias detection: any assignment whose right-hand side is exactly `booted`.
alias_bad=$(grep -rnE '(:=|=)[[:space:]]*booted[[:space:]]*(//.*)?$' "$dir" 2>/dev/null | grep -vE 'booted, err' || true)
if [ -n "$alias_bad" ]; then
  echo "TEMPLATE VIOLATION (aliasing the booted value defeats the accessor lint; use booted.Runtime*() directly):"
  echo "$alias_bad" | sed 's/^/  /'
  fail=1
fi

# Framework imports in generated Go must come from the one template-data
# source (`buildinfo.ModulePath`). A literal path here was the source of the
# root-/v2 transition drift and would let generators disagree again.
module_bad=$(find "$dir" -type f -name '*.go.tmpl' -exec grep -nH 'github\.com/qatoolist/wowapi' {} + 2>/dev/null || true)
if [ -n "$module_bad" ]; then
  echo "TEMPLATE VIOLATION (literal framework module path; use {{.FrameworkModule}}):"
  echo "$module_bad" | sed 's/^/  /'
  fail=1
fi

exit $fail
