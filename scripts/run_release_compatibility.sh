#!/bin/sh
set -eu

: "${WOWAPI_RELEASE_TAG:?WOWAPI_RELEASE_TAG must be the exact release tag}"

mode_or_predecessor=$(python3 scripts/validation/release_contract.py compatibility-policy \
  --policy ci/release-line.json --tag "$WOWAPI_RELEASE_TAG")

# These fixtures and supported-toolchain builds are release-line independent.
go test ./internal/cli -run 'OpenAPI' -count=1
go test ./internal/compat -run '^(TestCheckConfigSchemaCompatibility|TestGoAPIDiffGateFixtures|TestCandidateArchitectureSmoke)' -count=1
go test ./internal/compatcli -run '^TestRunConfig' -count=1
GOTOOLCHAIN=go1.26.0 go test -run '^$' ./...
GOTOOLCHAIN=go1.26.5 go test -run '^$' ./...
go test ./migrations -run '^TestIntegrationMigrationsReversible$' -count=1 -v

if [ "$mode_or_predecessor" = bootstrap ]; then
  echo "v1.2 production-line bootstrap: no predecessor exists"
  exit 0
fi

baseline=$(mktemp -d)
trap 'rm -rf "$baseline"' EXIT HUP INT TERM
git archive "$mode_or_predecessor" | tar -x -C "$baseline"
scripts/check_go_api_compat.sh "$baseline" .
(cd "$baseline" && go run ./cmd/wowapi config schema) > "$baseline/config-baseline.json"
go run ./cmd/wowapi config schema > "$baseline/config-current.json"
go run ./cmd/compatcheck config \
  --baseline "$baseline/config-baseline.json" --current "$baseline/config-current.json"
