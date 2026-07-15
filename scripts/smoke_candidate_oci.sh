#!/bin/sh
set -eu

if [ "$#" -ne 3 ]; then
    echo "usage: smoke_candidate_oci.sh <candidate-oci.tar> <tag> <sha256:digest>" >&2
    exit 2
fi

archive=$1
tag=$2
digest=$3
if [ ! -f "$archive" ]; then
    echo "candidate OCI smoke: archive not found: $archive" >&2
    exit 2
fi
case "$tag" in
    ""|*[!A-Za-z0-9._-]*)
        echo "candidate OCI smoke: malformed tag: $tag" >&2
        exit 2
        ;;
esac
case "$digest" in
    sha256:*) digest_hex=${digest#sha256:} ;;
    *)
        echo "candidate OCI smoke: malformed sha256 digest" >&2
        exit 2
        ;;
esac
case "$digest_hex" in
    *[!0-9a-f]*|"")
        echo "candidate OCI smoke: malformed sha256 digest" >&2
        exit 2
        ;;
esac
if [ "${#digest_hex}" -ne 64 ]; then
    echo "candidate OCI smoke: malformed sha256 digest" >&2
    exit 2
fi
for tool in curl docker oras tar; do
    command -v "$tool" >/dev/null 2>&1 || {
        echo "candidate OCI smoke: required tool not found: $tool" >&2
        exit 1
    }
done

work=$(mktemp -d "${TMPDIR:-/tmp}/wowapi-candidate-oci.XXXXXX")
registry_name="wowapi-candidate-registry-$$"
cleanup() {
    docker rm -f "$registry_name" >/dev/null 2>&1 || true
    rm -rf "$work"
}
trap cleanup EXIT HUP INT TERM

layout=$work/layout
mkdir -p "$layout"
tar -xf "$archive" -C "$layout"
docker run --detach --rm --name "$registry_name" --publish 127.0.0.1:5000:5000 \
    registry:2@sha256:a3d8aaa63ed8681a604f1dea0aa03f100d5895b6a58ace528858a7b332415373 >/dev/null

ready=0
for _ in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30; do
    if curl --fail --silent --show-error http://127.0.0.1:5000/v2/ >/dev/null; then
        ready=1
        break
    fi
    sleep 1
done
if [ "$ready" -ne 1 ]; then
    echo "candidate OCI smoke: local registry did not become ready" >&2
    exit 1
fi

oras cp --to-plain-http --from-oci-layout "$layout:$tag" "127.0.0.1:5000/wowapi:$tag"
reference="127.0.0.1:5000/wowapi@$digest"
script_dir=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
sh "$script_dir/smoke_candidate_arch.sh" "$reference" amd64
sh "$script_dir/smoke_candidate_arch.sh" "$reference" arm64

echo "candidate OCI smoke: exact multi-platform OCI candidate passed"
