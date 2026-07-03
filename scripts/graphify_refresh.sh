#!/bin/sh
set -eu

mode="${1:-update}"

case "$mode" in
  update)
    graphify update .
    ;;
  extract)
    graphify extract . --out . --max-concurrency "${GRAPHIFY_MAX_CONCURRENCY:-2}"
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
