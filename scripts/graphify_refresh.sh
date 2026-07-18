#!/bin/sh
set -eu

mode="${1:-update}"

case "$mode" in
  update)
    graphify update .
    ;;
  extract)
    # Semantic extraction needs an LLM. Pin Google Gemini as the repository
    # default so extraction never depends on Graphify's key-discovery order.
    # Override with GRAPHIFY_BACKEND / GRAPHIFY_MODEL for an intentional run.
    backend="${GRAPHIFY_BACKEND:-gemini}"
    if [ "$backend" = "gemini" ] && [ -z "${GEMINI_API_KEY:-${GOOGLE_API_KEY:-}}" ]; then
      echo "graphify extract: GEMINI_API_KEY or GOOGLE_API_KEY is required for --backend gemini" >&2
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
