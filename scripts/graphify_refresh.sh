#!/bin/sh
set -eu

mode="${1:-update}"

case "$mode" in
  update)
    graphify update .
    ;;
  extract)
    # Semantic extraction needs an LLM. Pin the backend EXPLICITLY (default kimi =
    # Moonshot) so it never falls back to Claude via graphify's "whichever API key
    # is set" default. Override with GRAPHIFY_BACKEND / GRAPHIFY_MODEL if needed.
    backend="${GRAPHIFY_BACKEND:-kimi}"
    if [ "$backend" = "kimi" ] && [ -z "${MOONSHOT_API_KEY:-}" ]; then
      echo "graphify extract: MOONSHOT_API_KEY is required for --backend kimi (Moonshot)" >&2
      exit 2
    fi
    if [ -n "${GRAPHIFY_MODEL:-}" ]; then
      graphify extract . --backend "$backend" --model "$GRAPHIFY_MODEL" \
        --out . --max-concurrency "${GRAPHIFY_MAX_CONCURRENCY:-2}"
    else
      graphify extract . --backend "$backend" \
        --out . --max-concurrency "${GRAPHIFY_MAX_CONCURRENCY:-2}"
    fi
    ;;
  cluster)
    graphify cluster-only .
    ;;
  check)
    graphify check-update .
    ;;
  *)
    echo "usage: $0 [update|extract|cluster|check]" >&2
    exit 2
    ;;
esac
