#!/usr/bin/env bash
# Devbox entrypoint: make the framework CLI + helpers available, print a
# cheatsheet, then hand over an interactive shell in /workspace. The framework
# checkout at /wowapi is READ-ONLY; nothing here writes to it.
set -euo pipefail

GOBIN_DIR="$(go env GOPATH)/bin"
export PATH="/wowapi/scripts/devbox:${GOBIN_DIR}:${PATH}"
# Persist for any shell the developer spawns later (incl. login shells, which
# would otherwise re-source /etc/profile and drop these entries).
printf 'export PATH="/wowapi/scripts/devbox:%s:$PATH"\n' "$GOBIN_DIR" > /etc/profile.d/wowapi.sh

# Install the wowapi CLI from the read-only checkout, stamped with a valid dev
# version so `wowapi init` emits a parseable `require ... v0.0.0-dev` (wow-link
# then points that at the local source via a replace directive).
echo "devbox: installing the wowapi CLI from /wowapi ..."
( cd /wowapi && go install -ldflags "-X github.com/qatoolist/wowapi/internal/buildinfo.version=v0.0.0-dev" ./cmd/wowapi )

# If a command was passed (e.g. `docker compose run devbox -lc '...'`), run it
# after the setup above instead of the interactive shell — handy for scripting.
if [ "$#" -gt 0 ]; then
  exec bash "$@"
fi

cat <<'CHEAT'

  ┌────────────────────────────────────────────────────────────────────────┐
  │  wowapi product-dev box — build a real product on the framework         │
  └────────────────────────────────────────────────────────────────────────┘
  Workspace : /workspace  (bind-mounted to your host dir — files persist)
  Framework : /wowapi     (read-only checkout; product links to it locally)
  Services  : postgres:5432 · minio:9000 · mailpit:1025   (product DB already set up)

  First-run flow:
    wowapi init --module github.com/qatoolist/wowproduct --name wowproduct
    wow-link                                  # link the product to /wowapi (replace)
    wowapi config validate --dir configs --env local
    go run ./cmd/migrate up                   # migrate product schema (runs as MIGRATE_URL)
    wowapi new-module --name tasks
    wowapi gen crud --module tasks --resource task
    go run ./cmd/migrate up
    go run ./cmd/api                          # serves :8080 (runs as app_rt — RLS enforced)

  From your Mac:  curl localhost:8080/healthz     → 200
  CLI help:       wowapi help        Re-link after go.mod edits:  wow-link

CHEAT

exec bash
