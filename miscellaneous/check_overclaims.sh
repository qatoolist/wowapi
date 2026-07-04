#!/usr/bin/env bash
# check_overclaims.sh — guard the "deferred-claimed-as-done" pattern in docs.
#
# Flags lines in the docs/CHANGELOG that assert completeness ("complete", "done",
# "fully", "closed") while the same line or its neighbours mention a deferral
# ("follow-up", "deferred", "future", "TODO", "not this pass", "pending"). Reviewers
# have repeatedly found "complete" claims sitting next to deferrals.
#
# When to run: during the review gate, before updating decisions/CHANGELOG.
# Read-only. Exit 0 always (advisory — human judges each hit).
set -euo pipefail
cd "$(dirname "$0")/.."

echo "== overclaim audit (complete-claim near a deferral) =="
# Scan the artifacts where an overclaim actually matters (decisions, CHANGELOG,
# evidence bundles). Exclude docs/working/ — those docs describe the
# deferred-claimed-as-done pattern by design, so they'd be all false positives.
targets="docs/implementation docs/operations CHANGELOG.md"
defer='follow-up|deferred|future orchestration|not this pass|TODO|pending|out of scope'
claim='complete|fully implemented|\bdone\b|closed|end-to-end'

# Lines that mention a deferral AND a completeness claim on the same line.
# Cross-line pairing is left to the reviewer's eyeball of proof bundles.
hits="$(grep -rniE --include='*.md' "($defer)" $targets 2>/dev/null \
        | grep -iE "($claim)" || true)"

if [ -z "$hits" ]; then
    echo "OK: no single-line complete-claim-next-to-deferral found."
    echo "(still eyeball proof bundles: a 'complete' section must not describe a follow-up.)"
else
    echo "Review these lines — a completeness claim shares a line with a deferral word:"
    echo "$hits" | sed 's/^/  /'
fi
