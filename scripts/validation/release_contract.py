#!/usr/bin/env python3
"""Fail-closed contracts for exact-commit gates and immutable release candidates.

The workflow uses GitHub attestations for the real trust boundary.  The JSON attestation
form accepted here is intentionally limited to scratch/throwaway tests; production jobs run
`gh attestation verify` before invoking the same byte/hash verification commands.
"""
from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile
from typing import Any

SHA40 = re.compile(r"^[0-9a-f]{40}$")
SHA256 = re.compile(r"^[0-9a-f]{64}$")
WORKFLOW_PREFIX = "https://github.com/qatoolist/wowapi/.github/workflows/release.yml@"
OIDC_ISSUER = "https://token.actions.githubusercontent.com"
GATE_REQUIRED = {
    "id", "command", "owner", "required_from_wave", "timeout_minutes",
    "evidence_artifact",
}
GATE_OPTIONAL = {"requires_services", "security_results"}
TOP_REQUIRED = {"schema_version", "completed_wave", "gates"}


class ContractError(Exception):
    pass


def digest_bytes(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def digest(path: Path) -> str:
    return digest_bytes(path.read_bytes())


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text())
    except (OSError, json.JSONDecodeError) as exc:
        raise ContractError(f"cannot parse {path}: {exc}") from exc


def canonical_write(path: Path, value: Any) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(value, sort_keys=True, separators=(",", ":")) + "\n")


def safe_file(root: Path, relative: str) -> Path:
    rel = Path(relative)
    if rel.is_absolute() or ".." in rel.parts or relative in {"", "."}:
        raise ContractError(f"unsafe candidate path: {relative!r}")
    path = root / rel
    resolved_root = root.resolve()
    try:
        resolved = path.resolve(strict=True)
    except OSError as exc:
        raise ContractError(f"missing candidate file: {relative}") from exc
    if resolved_root not in resolved.parents:
        raise ContractError(f"candidate path escapes root: {relative}")
    if path.is_symlink() or not path.is_file():
        raise ContractError(f"candidate path is not a regular file: {relative}")
    return path


def validate_gates_data(data: Any) -> dict[str, Any]:
    if not isinstance(data, dict):
        raise ContractError("gate manifest must be an object")
    missing = TOP_REQUIRED - data.keys()
    unknown = data.keys() - TOP_REQUIRED
    if missing:
        raise ContractError(f"gate manifest missing required field: {sorted(missing)[0]}")
    if unknown:
        raise ContractError(f"gate manifest has unknown field: {sorted(unknown)[0]}")
    if data["schema_version"] != 2:
        raise ContractError("schema_version must equal 2")
    if not isinstance(data["completed_wave"], int) or data["completed_wave"] < 0:
        raise ContractError("completed_wave must be a non-negative integer")
    gates = data["gates"]
    if not isinstance(gates, list) or not gates:
        raise ContractError("gates must be a non-empty array")
    ids: set[str] = set()
    evidence: set[str] = set()
    for index, gate in enumerate(gates):
        if not isinstance(gate, dict):
            raise ContractError(f"gate {index} must be an object")
        missing = GATE_REQUIRED - gate.keys()
        unknown = gate.keys() - GATE_REQUIRED - GATE_OPTIONAL
        if missing:
            raise ContractError(f"gate {index} missing required field: {sorted(missing)[0]}")
        if unknown:
            raise ContractError(f"gate {index} has unknown field: {sorted(unknown)[0]}")
        gate_id = gate["id"]
        if not isinstance(gate_id, str) or not re.fullmatch(r"[a-z0-9][a-z0-9-]*", gate_id):
            raise ContractError(f"gate {index} id is invalid")
        if gate_id in ids:
            raise ContractError(f"duplicate gate id: {gate_id}")
        ids.add(gate_id)
        for field in ("command", "owner", "evidence_artifact"):
            if not isinstance(gate[field], str) or not gate[field].strip():
                raise ContractError(f"gate {gate_id} field {field} must be non-empty")
        if not isinstance(gate["required_from_wave"], int) or gate["required_from_wave"] < 0:
            raise ContractError(f"gate {gate_id} required_from_wave is invalid")
        timeout = gate["timeout_minutes"]
        if not isinstance(timeout, int) or not 1 <= timeout <= 120:
            raise ContractError(f"gate {gate_id} timeout_minutes is invalid")
        artifact = gate["evidence_artifact"]
        if artifact.startswith("/") or ".." in Path(artifact).parts:
            raise ContractError(f"gate {gate_id} evidence_artifact is unsafe")
        if artifact in evidence:
            raise ContractError(f"duplicate evidence_artifact: {artifact}")
        if artifact != f"gate-evidence/{gate_id}.json":
            raise ContractError(
                f"gate {gate_id} evidence_artifact must equal gate-evidence/{gate_id}.json"
            )
        evidence.add(artifact)
        if "requires_services" in gate and not isinstance(gate["requires_services"], bool):
            raise ContractError(f"gate {gate_id} requires_services must be boolean")
        if "security_results" in gate:
            values = gate["security_results"]
            if not isinstance(values, list) or len(values) != len(set(values)):
                raise ContractError(f"gate {gate_id} security_results must be a unique array")
    return data


def command_validate_gates(args: argparse.Namespace) -> None:
    schema = load_json(Path(args.schema))
    if schema.get("$schema") != "https://json-schema.org/draft/2020-12/schema":
        raise ContractError("gate schema must use JSON Schema 2020-12")
    validate_gates_data(load_json(Path(args.manifest)))
    print(f"valid gate manifest: {args.manifest}")


def exact_commit(repo: Path, source_sha: str) -> str:
    if not SHA40.fullmatch(source_sha):
        raise ContractError("source SHA must be a full lowercase 40-hex commit SHA")
    result = subprocess.run(
        ["git", "-C", str(repo), "rev-parse", f"{source_sha}^{{commit}}"],
        text=True, capture_output=True, check=False,
    )
    if result.returncode != 0:
        raise ContractError(f"source SHA is not a commit in repository: {source_sha}")
    peeled = result.stdout.strip()
    if peeled != source_sha:
        raise ContractError(f"tag/source SHA mismatch: requested {source_sha}, resolved {peeled}")
    return peeled

def command_verify_tag(args: argparse.Namespace) -> None:
    repo = Path(args.repo).resolve()
    source_sha = exact_commit(repo, args.source_sha)
    if not re.fullmatch(r"v[0-9A-Za-z][0-9A-Za-z._-]*", args.tag):
        raise ContractError("release tag name is invalid")
    ref = f"refs/tags/{args.tag}^{{commit}}"
    result = subprocess.run(
        ["git", "-C", str(repo), "rev-parse", ref],
        text=True, capture_output=True, check=False,
    )
    if result.returncode != 0:
        raise ContractError(f"release tag does not resolve to a commit: {args.tag}")
    target = result.stdout.strip()
    if target != source_sha:
        raise ContractError(
            f"tag target mismatch: {args.tag} resolves to {target}, expected {source_sha}"
        )
    print(f"release tag target verified: {args.tag} -> {source_sha}")


SEMVER = re.compile(
    r"^v(?P<major>0|[1-9][0-9]*)\.(?P<minor>0|[1-9][0-9]*)\."
    r"(?P<patch>0|[1-9][0-9]*)(?P<pre>-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$"
)


def semver_key(tag: str) -> tuple[int, int, int, tuple[tuple[int, object], ...]]:
    match = SEMVER.fullmatch(tag)
    if not match:
        raise ContractError(f"release tag is not strict semantic version: {tag}")
    pre = match.group("pre")
    # A final release sorts after every prerelease of the same base.
    parts: tuple[tuple[int, object], ...] = ((2, ""),)
    if pre:
        values: list[tuple[int, object]] = []
        for value in pre[1:].split("."):
            if value.isdigit():
                if len(value) > 1 and value.startswith("0"):
                    raise ContractError(f"numeric prerelease identifier has leading zero: {tag}")
                values.append((0, int(value)))
            else:
                values.append((1, value))
        parts = ((1, ""), *values)
    return int(match["major"]), int(match["minor"]), int(match["patch"]), parts


def load_release_policy(path: Path) -> dict[str, Any]:
    policy = load_json(path)
    required = {
        "schema_version", "module_path", "bootstrap_tag", "mode",
        "supported_predecessor", "abandoned_tags",
    }
    if not isinstance(policy, dict) or set(policy) != required:
        raise ContractError("release-line policy has missing or unknown fields")
    if policy["schema_version"] != 1:
        raise ContractError("release-line schema_version must equal 1")
    if policy["module_path"] != "github.com/qatoolist/wowapi":
        raise ContractError("release-line module_path is not the canonical root module")
    if policy["mode"] not in {"bootstrap", "predecessor"}:
        raise ContractError("release-line mode must be bootstrap or predecessor")
    if not isinstance(policy["abandoned_tags"], list) or len(policy["abandoned_tags"]) != len(set(policy["abandoned_tags"])):
        raise ContractError("abandoned_tags must be a unique array")
    semver_key(policy["bootstrap_tag"])
    for tag in policy["abandoned_tags"]:
        semver_key(tag)
    predecessor = policy["supported_predecessor"]
    if policy["mode"] == "bootstrap" and predecessor is not None:
        raise ContractError("bootstrap mode must not declare a predecessor")
    if policy["mode"] == "predecessor":
        if not isinstance(predecessor, str):
            raise ContractError("predecessor mode requires supported_predecessor")
        semver_key(predecessor)
        if predecessor in policy["abandoned_tags"]:
            raise ContractError("supported predecessor is abandoned")
    return policy


def command_verify_release_identity(args: argparse.Namespace) -> None:
    repo = Path(args.repo).resolve()
    source_sha = exact_commit(repo, args.source_sha)
    target_key = semver_key(args.tag)
    if target_key[0] != 1:
        raise ContractError("root module releases must use semantic major v1")
    module = (repo / "go.mod").read_text().splitlines()[0].strip()
    if module != "module github.com/qatoolist/wowapi":
        raise ContractError("go.mod does not declare the canonical root module")
    status = subprocess.run(
        ["git", "-C", str(repo), "status", "--porcelain"], text=True,
        capture_output=True, check=False,
    )
    if status.returncode != 0 or status.stdout:
        raise ContractError("release checkout is not clean")
    command_verify_tag(argparse.Namespace(repo=str(repo), tag=args.tag, source_sha=source_sha))
    policy = load_release_policy(Path(args.policy))
    if args.tag in policy["abandoned_tags"]:
        raise ContractError("abandoned release tag must never be reused")
    if policy["mode"] == "bootstrap":
        if args.tag != policy["bootstrap_tag"]:
            raise ContractError("bootstrap mode is valid only for the declared first production tag")
    else:
        predecessor = policy["supported_predecessor"]
        if semver_key(predecessor) >= target_key:
            raise ContractError("supported predecessor must be older than the target tag")
    tags = subprocess.run(
        ["git", "-C", str(repo), "tag", "--list", "v*"], text=True,
        capture_output=True, check=False,
    )
    if tags.returncode != 0:
        raise ContractError("cannot enumerate repository tags")
    ignored = set(policy["abandoned_tags"]) | {args.tag}
    for tag in tags.stdout.splitlines():
        try:
            key = semver_key(tag)
        except ContractError:
            continue
        if tag not in ignored and key >= target_key:
            raise ContractError(f"release tag is not monotonic; existing tag is {tag}")
    proxy = load_json(Path(args.proxy_state))
    if proxy == {"status": 404}:
        pass
    elif isinstance(proxy, dict) and proxy.get("status") == 200:
        info = proxy.get("info")
        if not isinstance(info, dict) or info.get("Version") != args.tag:
            raise ContractError("public proxy returned mismatched version identity")
        origin = info.get("Origin")
        if not isinstance(origin, dict) or origin.get("Hash") != source_sha:
            raise ContractError("public proxy tag was already published from a different commit")
    else:
        raise ContractError("public proxy state is unavailable or ambiguous")
    print(f"strict release identity verified: {args.tag} -> {source_sha}")


def command_compatibility_policy(args: argparse.Namespace) -> None:
    policy = load_release_policy(Path(args.policy))
    target_key = semver_key(args.tag)
    if args.tag in policy["abandoned_tags"]:
        raise ContractError("abandoned release tag must never be reused")
    if policy["mode"] == "bootstrap":
        if args.tag != policy["bootstrap_tag"]:
            raise ContractError("future releases require predecessor compatibility mode")
        print("bootstrap")
        return
    predecessor = policy["supported_predecessor"]
    if semver_key(predecessor) >= target_key:
        raise ContractError("supported predecessor must be older than the target tag")
    print(predecessor)


def evidence_record(gate: dict[str, Any], source_sha: str, manifest_sha: str,
                    status: str, exit_code: int) -> dict[str, Any]:
    if not SHA40.fullmatch(source_sha) or not SHA256.fullmatch(manifest_sha):
        raise ContractError("gate evidence source or manifest digest is invalid")
    if status not in {"passed", "failed"} or not isinstance(exit_code, int):
        raise ContractError("gate evidence status or exit code is invalid")
    return {
        "schema_version": 1,
        "gate_id": gate["id"],
        "source_sha": source_sha,
        "manifest_sha256": manifest_sha,
        "command_sha256": digest_bytes(gate["command"].encode()),
        "status": status,
        "exit_code": exit_code,
    }


def command_write_gate_evidence(args: argparse.Namespace) -> None:
    manifest_path = Path(args.manifest)
    manifest = validate_gates_data(load_json(manifest_path))
    matches = [gate for gate in manifest["gates"] if gate["id"] == args.gate_id]
    if len(matches) != 1:
        raise ContractError(f"gate is not declared exactly once: {args.gate_id}")
    record = evidence_record(matches[0], args.source_sha, digest(manifest_path), args.status, args.exit_code)
    canonical_write(Path(args.output), record)


def verify_and_assemble_gate_bundle(manifest_path: Path, bundle: Path,
                                    source_sha: str) -> dict[str, Any]:
    if not SHA40.fullmatch(source_sha):
        raise ContractError("source SHA must be a full lowercase 40-hex commit SHA")
    manifest = validate_gates_data(load_json(manifest_path))
    manifest_sha = digest(manifest_path)
    results: list[dict[str, Any]] = []
    expected_paths = {gate["evidence_artifact"] for gate in manifest["gates"]}
    actual_paths = {
        str(path.relative_to(bundle)) for path in bundle.rglob("*")
        if path.is_file() and path.name != "gate-results.json"
    }
    if actual_paths != expected_paths:
        missing = expected_paths - actual_paths
        extra = actual_paths - expected_paths
        detail = f"missing {sorted(missing)[0]}" if missing else f"undeclared {sorted(extra)[0]}"
        raise ContractError(f"gate evidence set mismatch: {detail}")
    for gate in manifest["gates"]:
        path = safe_file(bundle, gate["evidence_artifact"])
        record = load_json(path)
        expected = evidence_record(gate, source_sha, manifest_sha, record.get("status"), record.get("exit_code"))
        if record != expected:
            raise ContractError(f"gate evidence identity mismatch: {gate['id']}")
        results.append({
            "id": gate["id"], "status": record["status"],
            "exit_code": record["exit_code"], "evidence_path": gate["evidence_artifact"],
            "evidence_sha256": digest(path), "command_sha256": record["command_sha256"],
        })
    return {
        "schema_version": 2, "source_sha": source_sha,
        "manifest_sha256": manifest_sha, "completed_wave": manifest["completed_wave"],
        "status": "passed" if all(item["status"] == "passed" for item in results) else "failed",
        "gates": sorted(results, key=lambda item: item["id"]),
    }


def command_assemble_gate_results(args: argparse.Namespace) -> None:
    output = verify_and_assemble_gate_bundle(Path(args.manifest), Path(args.bundle), args.source_sha)
    canonical_write(Path(args.output), output)
    if output["status"] != "passed":
        raise ContractError("one or more required gates failed")


def command_verify_gate_bundle(args: argparse.Namespace) -> None:
    bundle = Path(args.bundle)
    expected = verify_and_assemble_gate_bundle(Path(args.manifest), bundle, args.source_sha)
    actual = load_json(safe_file(bundle, args.results))
    if actual != expected:
        raise ContractError("gate-results summary does not match the evidence bundle")
    if actual["status"] != "passed":
        raise ContractError("required gates did not pass")
    print(f"gate evidence bundle verified for exact SHA {args.source_sha}")



def command_run_gates(args: argparse.Namespace) -> None:
    manifest_path = Path(args.manifest)
    manifest = validate_gates_data(load_json(manifest_path))
    repo = Path(args.repo).resolve()
    source_sha = exact_commit(repo, args.source_sha)
    output_path = Path(args.output).resolve()
    bundle = output_path.parent / f"{output_path.stem}-bundle"
    manifest_sha = digest(manifest_path)
    with tempfile.TemporaryDirectory(prefix="wowapi-exact-sha-") as work:
        archive = subprocess.Popen(["git", "-C", str(repo), "archive", source_sha], stdout=subprocess.PIPE)
        extract = subprocess.run(["tar", "-x", "-C", work], stdin=archive.stdout, capture_output=True, check=False)
        if archive.stdout:
            archive.stdout.close()
        archive_rc = archive.wait()
        if archive_rc != 0 or extract.returncode != 0:
            raise ContractError("failed to materialize exact source SHA")
        for gate in manifest["gates"]:
            env = {**os.environ, "WOWAPI_GATE_SOURCE_SHA": source_sha}
            try:
                result = subprocess.run(
                    ["bash", "-euo", "pipefail", "-c", gate["command"]],
                    cwd=work, env=env, text=True, capture_output=True,
                    timeout=gate["timeout_minutes"] * 60, check=False,
                )
                rc = result.returncode
            except subprocess.TimeoutExpired:
                rc = 124
            record = evidence_record(
                gate, source_sha, manifest_sha,
                "passed" if rc == 0 else "failed", rc,
            )
            canonical_write(bundle / gate["evidence_artifact"], record)
    summary = verify_and_assemble_gate_bundle(manifest_path, bundle, source_sha)
    canonical_write(bundle / "gate-results.json", summary)
    canonical_write(output_path, summary)
    if summary["status"] != "passed":
        failed = ", ".join(item["id"] for item in summary["gates"] if item["status"] == "failed")
        raise ContractError(f"required gates failed: {failed}")
    print(f"required gates passed for exact SHA {source_sha}")


def candidate_kind(relative: str, path: Path) -> str:
    name = Path(relative).name
    lowered = name.lower()
    if "artifact-security" in lowered:
        return "artifact-security"
    if "image-security" in lowered:
        return "image-security"
    if lowered == "wowapi" and path.stat().st_mode & 0o111:
        return "verification-binary"
    if "image-provenance" in lowered:
        return "image-provenance"
    if "image-sbom" in lowered:
        return "image-sbom"
    if lowered.endswith(".oci.tar"):
        return "image"
    if "provenance" in lowered or lowered.endswith(".intoto.jsonl"):
        return "provenance"
    if "sbom" in lowered or lowered.endswith(".spdx.json"):
        return "sbom"
    if lowered == "checksums.txt":
        return "checksums"
    if lowered.endswith(".cosign.bundle"):
        return "signature"
    if lowered.endswith((".tar.gz", ".zip")):
        return "archive"
    return "publisher-metadata"


def command_describe_candidate(args: argparse.Namespace) -> None:
    candidate = Path(args.candidate).resolve()
    if not SHA40.fullmatch(args.source_sha):
        raise ContractError("source SHA must be a full lowercase 40-hex commit SHA")
    if not re.fullmatch(r"sha256:[0-9a-f]{64}", args.image_digest):
        raise ContractError("image digest must be sha256:<64 lowercase hex>")
    gate = safe_file(candidate, args.gate_results)
    output = Path(args.output).resolve()
    artifacts: list[dict[str, Any]] = []
    for path in sorted(candidate.rglob("*")):
        if not path.is_file() or path.resolve() in {gate.resolve(), output}:
            continue
        relative = str(path.relative_to(candidate))
        kind = candidate_kind(relative, path)
        item: dict[str, Any] = {"path": relative, "kind": kind}
        if kind == "image":
            item["digest"] = args.image_digest
            item["platforms"] = ["linux/amd64", "linux/arm64"]
        artifacts.append(item)
    descriptor = {
        "schema_version": 1,
        "version": args.version,
        "source_sha": args.source_sha,
        "gate_results": args.gate_results,
        "artifacts": artifacts,
    }
    canonical_write(output, descriptor)
    print(f"described {len(artifacts)} immutable candidate input(s)")


def descriptor_and_gate(candidate: Path, descriptor_name: str) -> tuple[dict[str, Any], Path, dict[str, Any]]:
    descriptor_path = safe_file(candidate, descriptor_name)
    descriptor = load_json(descriptor_path)
    required = {"schema_version", "version", "source_sha", "gate_results", "artifacts"}
    if not isinstance(descriptor, dict) or required - descriptor.keys():
        raise ContractError("candidate descriptor missing required field")
    if descriptor["schema_version"] != 1 or not isinstance(descriptor["version"], str):
        raise ContractError("candidate descriptor version/schema is invalid")
    if not SHA40.fullmatch(str(descriptor["source_sha"])):
        raise ContractError("candidate descriptor source_sha is invalid")
    gate_path = safe_file(candidate, descriptor["gate_results"])
    gate = load_json(gate_path)
    if gate.get("status") != "passed":
        raise ContractError("required gates did not pass; candidate creation is blocked")
    if gate.get("source_sha") != descriptor["source_sha"]:
        raise ContractError("tag/source SHA does not match gate-results source SHA")
    return descriptor, descriptor_path, gate


def command_create_manifest(args: argparse.Namespace) -> None:
    candidate = Path(args.candidate).resolve()
    descriptor, descriptor_path, gate = descriptor_and_gate(candidate, args.descriptor)
    artifacts: list[dict[str, Any]] = []
    seen: set[str] = set()
    for index, item in enumerate(descriptor["artifacts"]):
        if not isinstance(item, dict) or not {"path", "kind"} <= item.keys():
            raise ContractError(f"descriptor artifact {index} is invalid")
        relative = item["path"]
        if relative in seen:
            raise ContractError(f"duplicate candidate artifact: {relative}")
        seen.add(relative)
        path = safe_file(candidate, relative)
        record = {key: value for key, value in item.items() if key != "sha256"}
        record["sha256"] = digest(path)
        record["size"] = path.stat().st_size
        artifacts.append(record)
    kinds = {item["kind"] for item in artifacts}
    required_kinds = {
        "archive": "archive",
        "checksums": "checksums",
        "signature": "signature",
        "sbom": "SBOM",
        "provenance": "provenance",
        "image": "image",
        "image-sbom": "image SBOM",
        "image-provenance": "image provenance",
        "verification-binary": "verification binary",
        "artifact-security": "artifact security report",
        "image-security": "image security report",
    }
    if missing := required_kinds.keys() - kinds:
        kind = sorted(missing)[0]
        raise ContractError(f"candidate missing required {required_kinds[kind]}")
    manifest = {
        "schema_version": 1,
        "version": descriptor["version"],
        "source_sha": descriptor["source_sha"],
        "gate_results": {"path": descriptor["gate_results"], "sha256": digest(safe_file(candidate, descriptor["gate_results"]))},
        "gate_manifest_sha256": gate.get("manifest_sha256"),
        "completed_wave": gate.get("completed_wave"),
        "descriptor": {"path": args.descriptor, "sha256": digest(descriptor_path)},
        "artifacts": sorted(artifacts, key=lambda item: item["path"]),
    }
    canonical_write(Path(args.output), manifest)
    print(f"created immutable release manifest: {args.output}")


def load_manifest(candidate: Path) -> tuple[Path, dict[str, Any]]:
    path = safe_file(candidate, "release-manifest.json")
    manifest = load_json(path)
    if not isinstance(manifest, dict) or manifest.get("schema_version") != 1:
        raise ContractError("release manifest is invalid")
    return path, manifest


def verify_offline_attestation(candidate: Path, attestation_name: str, manifest_path: Path, manifest: dict[str, Any]) -> None:
    attestation_path = safe_file(candidate, attestation_name)
    attestation = load_json(attestation_path)
    if attestation.get("subject_sha256") != digest(manifest_path):
        raise ContractError("manifest attestation hash mismatch")
    if attestation.get("source_sha") != manifest.get("source_sha"):
        raise ContractError("manifest attestation source SHA mismatch")
    if not str(attestation.get("workflow_identity", "")).startswith(WORKFLOW_PREFIX):
        raise ContractError("manifest attestation workflow identity mismatch")
    if attestation.get("oidc_issuer") != OIDC_ISSUER:
        raise ContractError("manifest attestation OIDC issuer mismatch")


def expected_candidate_files(manifest: dict[str, Any], attestation_name: str) -> set[str]:
    files = {"release-manifest.json", attestation_name, manifest["descriptor"]["path"], manifest["gate_results"]["path"]}
    files.update(item["path"] for item in manifest["artifacts"])
    return files


def verify_candidate(candidate: Path, attestation_name: str) -> dict[str, Any]:
    manifest_path, manifest = load_manifest(candidate)
    verify_offline_attestation(candidate, attestation_name, manifest_path, manifest)
    gate_path = safe_file(candidate, manifest["gate_results"]["path"])
    if digest(gate_path) != manifest["gate_results"]["sha256"]:
        raise ContractError("gate-results hash mismatch")
    gate = load_json(gate_path)
    if gate.get("status") != "passed" or gate.get("source_sha") != manifest.get("source_sha"):
        raise ContractError("gate-results status/source SHA mismatch")
    descriptor_path = safe_file(candidate, manifest["descriptor"]["path"])
    if digest(descriptor_path) != manifest["descriptor"]["sha256"]:
        raise ContractError("candidate descriptor hash mismatch")
    for artifact in manifest["artifacts"]:
        path = safe_file(candidate, artifact["path"])
        if digest(path) != artifact["sha256"] or path.stat().st_size != artifact["size"]:
            raise ContractError(f"artifact hash mismatch: {artifact['path']}")
    actual = {str(path.relative_to(candidate)) for path in candidate.rglob("*") if path.is_file()}
    extra = actual - expected_candidate_files(manifest, attestation_name)
    missing = expected_candidate_files(manifest, attestation_name) - actual
    if extra:
        raise ContractError(f"unmanifested candidate file: {sorted(extra)[0]}")
    if missing:
        raise ContractError(f"manifested candidate file missing: {sorted(missing)[0]}")
    return manifest


def command_verify_candidate(args: argparse.Namespace) -> None:
    manifest = verify_candidate(Path(args.candidate).resolve(), args.attestation)
    print(f"candidate verified: v{manifest['version']} {manifest['source_sha']}")


def copy_verified(candidate: Path, destination: Path, attestation: str) -> dict[str, Any]:
    manifest = verify_candidate(candidate, attestation)
    if destination.exists():
        raise ContractError(f"publish destination already exists: {destination}")
    destination.mkdir(parents=True)
    for relative in sorted(expected_candidate_files(manifest, attestation)):
        source = safe_file(candidate, relative)
        target = destination / relative
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(source, target)
        if source.stat().st_mode & 0o111:
            target.chmod(0o755)
    verify_candidate(destination, attestation)
    return manifest


def command_publish(args: argparse.Namespace) -> None:
    manifest = copy_verified(Path(args.candidate).resolve(), Path(args.destination).resolve(), args.attestation)
    print(f"published attested bytes only: v{manifest['version']}")


def command_promote(args: argparse.Namespace) -> None:
    candidate = Path(args.candidate).resolve()
    registry = Path(args.registry).resolve()
    manifest = verify_candidate(candidate, args.attestation)
    version = manifest["version"]
    version_dir = registry / version
    staging = registry / f".{version}.staging"
    if staging.exists():
        shutil.rmtree(staging)
    copy_verified(candidate, staging, args.attestation)
    if version_dir.exists():
        raise ContractError(f"immutable version already exists: {version}")
    os.replace(staging, version_dir)
    latest_tmp = registry / ".latest.tmp"
    latest_tmp.write_text(version + "\n")
    os.replace(latest_tmp, registry / "latest")
    print(f"promoted immutable candidate: {version}")


def require_kind(manifest: dict[str, Any], kind: str, label: str) -> list[dict[str, Any]]:
    items = [item for item in manifest.get("artifacts", []) if item.get("kind") == kind]
    if not items:
        raise ContractError(f"missing {label}")
    return items

def require_kind_files(
    manifest: dict[str, Any],
    release: Path,
    kind: str,
    label: str,
) -> list[dict[str, Any]]:
    items = require_kind(manifest, kind, label)
    for item in items:
        try:
            safe_file(release, item["path"])
        except ContractError as exc:
            raise ContractError(f"missing {label}: {item['path']}") from exc
    return items



def command_verify_release(args: argparse.Namespace) -> None:
    release = Path(args.release_dir).resolve()
    _, manifest = load_manifest(release)
    if manifest.get("version") != args.version:
        raise ContractError(f"release version mismatch: expected {args.version}")
    if manifest.get("source_sha") != args.source_sha:
        raise ContractError(f"release source SHA mismatch: expected {args.source_sha}")
    signatures = require_kind_files(manifest, release, "signature", "signature")
    require_kind_files(manifest, release, "sbom", "SBOM attestation")
    require_kind_files(manifest, release, "image-sbom", "SBOM attestation")
    require_kind_files(manifest, release, "provenance", "provenance attestation")
    require_kind_files(manifest, release, "image-provenance", "provenance attestation")
    require_kind_files(manifest, release, "artifact-security", "artifact security report")
    require_kind_files(manifest, release, "image-security", "image security report")
    images = require_kind_files(manifest, release, "image", "image")
    expected_platforms = {"linux/amd64", "linux/arm64"}
    for image in images:
        if set(image.get("platforms", [])) != expected_platforms:
            raise ContractError(f"image platform mismatch: expected {sorted(expected_platforms)}")
        if not re.fullmatch(r"sha256:[0-9a-f]{64}", str(image.get("digest", ""))):
            raise ContractError("image digest is invalid")
    manifest = verify_candidate(release, args.attestation)
    checksums_item = require_kind(manifest, "checksums", "checksums")[0]
    checksums_path = safe_file(release, checksums_item["path"])
    signature_path = safe_file(release, signatures[0]["path"])
    if os.environ.get("WOWAPI_OFFLINE_VERIFY") == "1":
        signature = load_json(signature_path)
        if signature.get("subject_sha256") != digest(checksums_path):
            raise ContractError("signature subject hash mismatch")
        if not str(signature.get("certificate_identity", "")).startswith(WORKFLOW_PREFIX) or signature.get("oidc_issuer") != OIDC_ISSUER:
            raise ContractError("signature identity mismatch")
    else:
        cosign = shutil.which("cosign")
        if cosign is None:
            raise ContractError("cosign is required for signature verification")
        identity = f"{WORKFLOW_PREFIX}refs/tags/v{args.version}"
        verified = subprocess.run(
            [
                cosign,
                "verify-blob",
                "--bundle",
                str(signature_path),
                "--certificate-identity",
                identity,
                "--certificate-oidc-issuer",
                OIDC_ISSUER,
                str(checksums_path),
            ],
            text=True,
            capture_output=True,
            check=False,
        )
        if verified.returncode != 0:
            detail = verified.stderr.strip() or verified.stdout.strip() or "cosign rejected bundle"
            raise ContractError(f"signature verification failed: {detail}")
    checksum_entries: dict[str, str] = {}
    for line in checksums_path.read_text().splitlines():
        match = re.fullmatch(r"([0-9a-f]{64})  ([^/]+)", line)
        if not match:
            raise ContractError("checksums file has invalid entry")
        checksum_entries[match.group(2)] = match.group(1)
    for archive in require_kind(manifest, "archive", "archive"):
        path = safe_file(release, archive["path"])
        if checksum_entries.get(archive["path"]) != digest(path):
            raise ContractError(f"published archive checksum mismatch: {archive['path']}")
    for provenance_kind in ("provenance", "image-provenance"):
        for item in require_kind(manifest, provenance_kind, "provenance attestation"):
            text = safe_file(release, item["path"]).read_text()
            if manifest["source_sha"] not in text:
                raise ContractError(f"provenance source SHA mismatch: {item['path']}")
    binary_item = require_kind(manifest, "verification-binary", "verification binary")[0]
    binary = safe_file(release, binary_item["path"])
    result = subprocess.run([str(binary), "version"], text=True, capture_output=True, check=False)
    if result.returncode != 0 or args.version not in result.stdout:
        raise ContractError("CLI version mismatch")
    print(f"published release verified from clean inputs: v{args.version} {args.source_sha}")


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser()
    sub = parser.add_subparsers(dest="command", required=True)
    item = sub.add_parser("validate-gates")
    item.add_argument("--manifest", required=True)
    item.add_argument("--schema", required=True)
    item.set_defaults(func=command_validate_gates)
    item = sub.add_parser("verify-tag")
    item.add_argument("--repo", required=True)
    item.add_argument("--tag", required=True)
    item.add_argument("--source-sha", required=True)
    item.set_defaults(func=command_verify_tag)
    item = sub.add_parser("verify-release-identity")
    item.add_argument("--repo", required=True)
    item.add_argument("--tag", required=True)
    item.add_argument("--source-sha", required=True)
    item.add_argument("--policy", required=True)
    item.add_argument("--proxy-state", required=True)
    item.set_defaults(func=command_verify_release_identity)
    item = sub.add_parser("compatibility-policy")
    item.add_argument("--policy", required=True)
    item.add_argument("--tag", required=True)
    item.set_defaults(func=command_compatibility_policy)
    item = sub.add_parser("write-gate-evidence")
    item.add_argument("--manifest", required=True)
    item.add_argument("--gate-id", required=True)
    item.add_argument("--source-sha", required=True)
    item.add_argument("--status", choices=("passed", "failed"), required=True)
    item.add_argument("--exit-code", type=int, required=True)
    item.add_argument("--output", required=True)
    item.set_defaults(func=command_write_gate_evidence)
    item = sub.add_parser("assemble-gate-results")
    item.add_argument("--manifest", required=True)
    item.add_argument("--bundle", required=True)
    item.add_argument("--source-sha", required=True)
    item.add_argument("--output", required=True)
    item.set_defaults(func=command_assemble_gate_results)
    item = sub.add_parser("verify-gate-bundle")
    item.add_argument("--manifest", required=True)
    item.add_argument("--bundle", required=True)
    item.add_argument("--source-sha", required=True)
    item.add_argument("--results", default="gate-results.json")
    item.set_defaults(func=command_verify_gate_bundle)
    item = sub.add_parser("run-gates")
    item.add_argument("--manifest", required=True)
    item.add_argument("--source-sha", required=True)
    item.add_argument("--repo", required=True)
    item.add_argument("--output", required=True)
    item.set_defaults(func=command_run_gates)
    item = sub.add_parser("describe-candidate")
    item.add_argument("--candidate", required=True)
    item.add_argument("--version", required=True)
    item.add_argument("--source-sha", required=True)
    item.add_argument("--gate-results", required=True)
    item.add_argument("--image-digest", required=True)
    item.add_argument("--output", required=True)
    item.set_defaults(func=command_describe_candidate)
    item = sub.add_parser("create-manifest")
    item.add_argument("--candidate", required=True)
    item.add_argument("--descriptor", required=True)
    item.add_argument("--output", required=True)
    item.set_defaults(func=command_create_manifest)
    item = sub.add_parser("verify-candidate")
    item.add_argument("--candidate", required=True)
    item.add_argument("--attestation", required=True)
    item.set_defaults(func=command_verify_candidate)
    item = sub.add_parser("publish")
    item.add_argument("--candidate", required=True)
    item.add_argument("--destination", required=True)
    item.add_argument("--attestation", required=True)
    item.set_defaults(func=command_publish)
    item = sub.add_parser("promote")
    item.add_argument("--candidate", required=True)
    item.add_argument("--registry", required=True)
    item.add_argument("--attestation", required=True)
    item.set_defaults(func=command_promote)
    item = sub.add_parser("verify-release")
    item.add_argument("--release-dir", required=True)
    item.add_argument("--version", required=True)
    item.add_argument("--source-sha", required=True)
    item.add_argument("--attestation", default="release-manifest.attestation.json")
    item.set_defaults(func=command_verify_release)
    return parser


def main() -> int:
    try:
        args = build_parser().parse_args()
        args.func(args)
        return 0
    except ContractError as exc:
        print(f"release contract violation: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
