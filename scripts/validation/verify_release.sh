#!/bin/sh
set -eu

if [ "$#" -ne 2 ]; then
  echo "usage: $0 <version> <source-sha>" >&2
  exit 2
fi

version=$1
source_sha=$2
release_dir=${WOWAPI_RELEASE_DIR:-}

case "$source_sha" in
  *[!0-9a-f]*|'') echo "source SHA must be a full lowercase 40-hex SHA" >&2; exit 2 ;;
esac
if [ "${#source_sha}" -ne 40 ]; then
  echo "source SHA must be a full lowercase 40-hex SHA" >&2
  exit 2
fi
if [ -z "$release_dir" ] || [ ! -d "$release_dir" ]; then
  echo "WOWAPI_RELEASE_DIR must name a clean directory containing only downloaded release subjects" >&2
  exit 2
fi

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

# Scratch/throwaway fixtures carry a deterministic JSON attestation receipt. Real release
# verification MUST first verify GitHub's signed attestation; the receipt is only the local,
# inspectable binding consumed by the byte verifier after that cryptographic check.
if [ "${WOWAPI_OFFLINE_VERIFY:-0}" != "1" ]; then
  command -v gh >/dev/null 2>&1 || { echo "gh is required for attestation verification" >&2; exit 1; }
  gh attestation verify "$release_dir/release-manifest.json" \
    --repo qatoolist/wowapi \
    --signer-workflow qatoolist/wowapi/.github/workflows/release.yml >/dev/null
fi

exec python3 "$script_dir/release_contract.py" verify-release \
  --release-dir "$release_dir" \
  --version "$version" \
  --source-sha "$source_sha" \
  --attestation release-manifest.attestation.json
