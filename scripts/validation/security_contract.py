#!/usr/bin/env python3
"""Fail-closed security gate, waiver, visibility, and private-fallback contracts."""

from __future__ import annotations

import argparse
from datetime import date
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import time
from typing import Any

REQUIRED_WAIVER_FIELDS = {
    "id",
    "scanner",
    "scope",
    "owner",
    "rationale",
    "expires",
    "remediation_link",
}
ALLOWED_WAIVER_FIELDS = REQUIRED_WAIVER_FIELDS
ALLOWED_SCANNERS = {
    "trivy",
    "gitleaks",
    "govulncheck",
    "dependency-review",
    "actionlint",
    "codeql",
    "scorecard",
    "local-sast",
}
REQUIRED_SECURITY_RESULTS = {
    "secret",
    "reachable-vulnerability",
    "disallowed-license",
    "dependency",
    "workflow",
    "critical-high-config",
}
HOSTED_SCANNERS = ("dependency-review", "codeql", "scorecard")


class ContractError(Exception):
    pass


def load_json(path: Path) -> Any:
    try:
        return json.loads(path.read_text())
    except (OSError, json.JSONDecodeError) as exc:
        raise ContractError(f"cannot parse {path}: {exc}") from exc


def parse_day(raw: str, field: str) -> date:
    try:
        return date.fromisoformat(raw)
    except (TypeError, ValueError) as exc:
        raise ContractError(f"{field} must be an ISO-8601 date") from exc


def validate_waivers(data: Any, today: date) -> list[dict[str, str]]:
    if not isinstance(data, dict) or set(data) != {"schema_version", "waivers"}:
        raise ContractError(
            "waiver registry must contain only schema_version and waivers"
        )
    if data["schema_version"] != 1 or not isinstance(data["waivers"], list):
        raise ContractError("waiver registry schema_version/waivers is invalid")
    seen: set[tuple[str, str, str]] = set()
    validated: list[dict[str, str]] = []
    for index, waiver in enumerate(data["waivers"]):
        if not isinstance(waiver, dict):
            raise ContractError(f"waiver {index} must be an object")
        missing = REQUIRED_WAIVER_FIELDS - waiver.keys()
        unknown = waiver.keys() - ALLOWED_WAIVER_FIELDS
        if missing:
            raise ContractError(
                f"waiver {index} missing required field: {sorted(missing)[0]}"
            )
        if unknown:
            raise ContractError(
                f"waiver {index} has unknown field: {sorted(unknown)[0]}"
            )
        for field in REQUIRED_WAIVER_FIELDS:
            if not isinstance(waiver[field], str) or not waiver[field].strip():
                raise ContractError(f"waiver {index} field {field} must be non-empty")
        if waiver["scanner"] not in ALLOWED_SCANNERS:
            raise ContractError(f"waiver {index} scanner is unsupported")
        if len(waiver["rationale"].strip()) < 10:
            raise ContractError(f"waiver {index} rationale is too short")
        if not waiver["remediation_link"].startswith("https://"):
            raise ContractError(f"waiver {index} remediation_link must use https")
        expiry = parse_day(waiver["expires"], f"waiver {index} expires")
        if expiry < today:
            raise ContractError(
                f"waiver {waiver['id']} expired on {expiry.isoformat()}"
            )
        key = (waiver["scanner"], waiver["id"], waiver["scope"])
        if key in seen:
            raise ContractError(f"duplicate waiver scope: {key}")
        seen.add(key)
        validated.append(waiver)
    return validated


def requested_today(raw: str | None) -> date:
    return parse_day(raw, "today") if raw else date.today()


def command_validate_waivers(args: argparse.Namespace) -> None:
    schema = load_json(Path(args.schema))
    if schema.get("$schema") != "https://json-schema.org/draft/2020-12/schema":
        raise ContractError("waiver schema must use JSON Schema 2020-12")
    waivers = validate_waivers(
        load_json(Path(args.waivers)), requested_today(args.today)
    )
    print(f"valid waiver registry: {len(waivers)} active scoped waiver(s)")


def command_enforce_results(args: argparse.Namespace) -> None:
    today = requested_today(args.today)
    waivers = validate_waivers(load_json(Path(args.waivers)), today)
    registry = {(item["scanner"], item["id"], item["scope"]): item for item in waivers}
    document = load_json(Path(args.results))
    if (
        not isinstance(document, dict)
        or document.get("schema_version") != 1
        or not isinstance(document.get("results"), list)
    ):
        raise ContractError("scanner results document is invalid")
    failures: list[str] = []
    for index, result in enumerate(document["results"]):
        if not isinstance(result, dict) or result.get("status") not in {
            "passed",
            "failed",
        }:
            raise ContractError(f"scanner result {index} is invalid")
        result_class = result.get("class")
        if result_class not in REQUIRED_SECURITY_RESULTS:
            raise ContractError(f"scanner result {index} has unsupported class")
        if result["status"] == "passed":
            continue
        key = (
            result.get("scanner", ""),
            result.get("finding_id", ""),
            result.get("scope", ""),
        )
        if key not in registry:
            failures.append(
                f"{result_class}: unwaived finding {key[1] or '<missing-id>'} in {key[2] or '<missing-scope>'}"
            )
    if failures:
        raise ContractError("; ".join(failures))
    print("security scanner results passed or matched exact active waiver scopes")


def command_visibility_guard(args: argparse.Namespace) -> None:
    if args.visibility not in {"public", "private", "internal"}:
        raise ContractError("visibility must be public, private, or internal")
    results = load_json(Path(args.hosted_results))
    if not isinstance(results, dict):
        raise ContractError("hosted scanner results must be an object")
    if args.visibility == "public":
        failed = [name for name in HOSTED_SCANNERS if results.get(name) != "success"]
        if failed:
            raise ContractError(
                "public repository missing successful hosted scanner: "
                + ", ".join(failed)
            )
        print("public hosted scanners all ran successfully")
    else:
        print(
            "fallback-required: hosted scanners are not trusted for non-public visibility"
        )


def unsafe_findings(root: Path) -> list[str]:
    patterns = (
        (
            re.compile(r'exec\.Command\(\s*"(?:sh|bash)"\s*,\s*"-c"\s*,'),
            "unsafe shell command construction",
        ),
        (re.compile(r"InsecureSkipVerify\s*:\s*true"), "TLS verification disabled"),
        (re.compile(r"md5\.New\s*\("), "cryptographic MD5 use"),
    )
    findings: list[str] = []
    for path in sorted(root.rglob("*.go")):
        if any(part in {"vendor", ".git", "testdata"} for part in path.parts):
            continue
        text = path.read_text(errors="replace")
        for pattern, message in patterns:
            for match in pattern.finditer(text):
                line = text.count("\n", 0, match.start()) + 1
                findings.append(f"{path}:{line}: {message}")
    return findings


def command_local_sast(args: argparse.Namespace) -> None:
    findings = unsafe_findings(Path(args.path))
    if findings:
        raise ContractError(
            "local SAST unsafe shell/security pattern: " + "; ".join(findings)
        )
    print("local SAST fallback found no seeded unsafe pattern")


def posture_findings(workflow: Path) -> list[str]:
    findings: list[str] = []
    text = workflow.read_text(errors="replace")
    for line_number, line in enumerate(text.splitlines(), start=1):
        match = re.search(r"\buses:\s*([^\s#]+)", line)
        if not match:
            continue
        action = match.group(1)
        if action.startswith("./") or action.startswith("docker://"):
            continue
        if "@" not in action or not re.fullmatch(r"[^@]+@[0-9a-fA-F]{40}", action):
            findings.append(
                f"{workflow}:{line_number}: action is not pinned to a full commit SHA: {action}"
            )
    if "permissions:" not in text:
        findings.append(f"{workflow}: missing explicit permissions")
    return findings


def command_repository_posture(args: argparse.Namespace) -> None:
    paths: list[Path]
    if args.workflow:
        paths = [Path(args.workflow)]
    else:
        root = Path(args.path)
        paths = sorted((root / ".github/workflows").glob("*.yml"))
    findings = [finding for path in paths for finding in posture_findings(path)]
    if findings:
        raise ContractError(
            "scorecard-equivalent posture failure (full commit SHA/least privilege): "
            + "; ".join(findings)
        )
    print(f"scorecard-equivalent posture passed for {len(paths)} workflow(s)")


def command_cross_reference(args: argparse.Namespace) -> None:
    manifest = load_json(Path(args.manifest))
    gates = manifest.get("gates") if isinstance(manifest, dict) else None
    if not isinstance(gates, list):
        raise ContractError("release gate manifest is invalid")
    ids: set[str] = set()
    counts = {name: 0 for name in REQUIRED_SECURITY_RESULTS}
    for gate in gates:
        gate_id = gate.get("id") if isinstance(gate, dict) else None
        if not isinstance(gate_id, str) or gate_id in ids:
            raise ContractError(f"duplicate or invalid gate id: {gate_id}")
        ids.add(gate_id)
        for result_class in gate.get("security_results", []):
            if result_class in counts:
                counts[result_class] += 1
    wrong = {name: count for name, count in counts.items() if count != 1}
    if wrong:
        raise ContractError(
            "security result manifest cardinality is not exactly one: "
            + json.dumps(wrong, sort_keys=True)
        )
    required_ids = {
        "trivy-blocking",
        "security-waivers",
        "hosted-scanner-meta",
        "private-scanner-fallback",
    }
    if missing := required_ids - ids:
        raise ContractError(
            "REL-02 manifest entry missing: " + ", ".join(sorted(missing))
        )
    print(
        "every required security result and REL-02 check appears exactly once in the gate manifest"
    )


def gh_output(args: list[str], *, retries: int = 0, interval: float = 3.0) -> str:
    """Run a gh command, returning stdout. Fails CLOSED (raises) on error.

    retries>0 re-attempts on a transient failure — a non-zero exit, OR a 200
    whose body is an HTML error/maintenance page (GitHub occasionally serves
    these during API incidents, which then fail a downstream `--jq` parse with
    'invalid character <'). This does NOT weaken the gate: it still raises after
    the last attempt, and the facts these calls read (repo visibility, hosted
    scan conclusions) are stable, so retrying a blip is strictly correct.
    """
    last = ""
    for attempt in range(retries + 1):
        result = subprocess.run(args, text=True, capture_output=True, check=False)
        if result.returncode == 0 and not result.stdout.lstrip().startswith("<"):
            return result.stdout
        last = (result.stderr.strip() or result.stdout.strip())[:200]
        if attempt < retries:
            time.sleep(interval)
    raise ContractError(
        f"command failed closed after {retries + 1} attempt(s): {' '.join(args)}: {last}"
    )


def repository_visibility() -> str:
    repo = os.environ.get("GITHUB_REPOSITORY", "qatoolist/wowapi")
    # Bounded retries: a transient API blip must not fail the gate irrecoverably
    # (mirrors the hosted-scanner meta-check's own retry loop below).
    attempts = int(os.environ.get("WOWAPI_GH_API_RETRIES", "5"))
    return gh_output(
        ["gh", "api", f"repos/{repo}", "--jq", ".visibility"], retries=attempts
    ).strip()


def hosted_results_for_sha(source_sha: str) -> dict[str, str]:
    if not re.fullmatch(r"[0-9a-f]{40}", source_sha):
        raise ContractError("source SHA must be full lowercase 40-hex")
    workflows = {
        "dependency-review": "security-scan.yml",
        "codeql": "codeql.yml",
        "scorecard": "scorecard.yml",
    }
    results: dict[str, str] = {}
    for name, workflow in workflows.items():
        raw = gh_output(
            [
                "gh",
                "run",
                "list",
                "--workflow",
                workflow,
                "--commit",
                source_sha,
                "--limit",
                "20",
                "--json",
                "conclusion",
            ]
        )
        runs = json.loads(raw)
        results[name] = (
            "success"
            if any(item.get("conclusion") == "success" for item in runs)
            else "missing"
        )
    return results


def command_workflow_policy(args: argparse.Namespace) -> None:
    visibility = (
        repository_visibility() if args.visibility == "auto" else args.visibility
    )
    if visibility != "public":
        print("fallback-required: non-public repository must run local scanners")
        return
    attempts = int(os.environ.get("WOWAPI_HOSTED_SCAN_ATTEMPTS", "30"))
    interval = int(os.environ.get("WOWAPI_HOSTED_SCAN_INTERVAL", "30"))
    if attempts < 1 or attempts > 60 or interval < 0 or interval > 120:
        raise ContractError("hosted scanner retry configuration is invalid")
    failed = list(HOSTED_SCANNERS)
    for attempt in range(attempts):
        results = hosted_results_for_sha(args.source_sha)
        failed = [name for name in HOSTED_SCANNERS if results.get(name) != "success"]
        if not failed:
            print("public exact-SHA hosted scanner meta-check passed")
            return
        if attempt + 1 < attempts:
            time.sleep(interval)
    raise ContractError(
        "public exact-SHA hosted scanner missing/success not observed: "
        + ", ".join(failed)
    )


def run_checked(command: list[str], cwd: Path) -> None:
    result = subprocess.run(command, cwd=cwd, check=False)
    if result.returncode != 0:
        raise ContractError("private fallback command failed: " + " ".join(command))


def command_private_fallback(args: argparse.Namespace) -> None:
    root = Path(args.path).resolve()
    if (
        not args.force
        and os.environ.get("GITHUB_ACTIONS") == "true"
        and repository_visibility() == "public"
    ):
        print("private fallback not required for public repository")
        return
    findings = unsafe_findings(root)
    if findings:
        raise ContractError("private local SAST failed: " + "; ".join(findings))
    workflow_findings = [
        finding
        for path in sorted((root / ".github/workflows").glob("*.yml"))
        for finding in posture_findings(path)
    ]
    if workflow_findings:
        raise ContractError(
            "private scorecard-equivalent failed: " + "; ".join(workflow_findings)
        )
    run_checked(
        ["go", "run", "github.com/rhysd/actionlint/cmd/actionlint@v1.7.12", "-color"],
        root,
    )
    run_checked(
        ["go", "run", "golang.org/x/vuln/cmd/govulncheck@v1.1.4", "./..."], root
    )
    run_checked(
        [
            "trivy",
            "fs",
            "--scanners",
            "vuln,secret,misconfig,license",
            "--severity",
            "CRITICAL,HIGH",
            "--ignorefile",
            ".trivyignore.yaml",
            "--exit-code",
            "1",
            ".",
        ],
        root,
    )
    print(
        "private fallback scanners passed (local SAST/posture/actionlint/govulncheck/Trivy)"
    )


def command_validate_trivy_ignore(args: argparse.Namespace) -> None:
    waivers = validate_waivers(
        load_json(Path(args.waivers)), requested_today(args.today)
    )
    expected = {
        (item["id"], item["scope"], item["expires"])
        for item in waivers
        if item["scanner"] == "trivy"
    }
    actual: set[tuple[str, str, str]] = set()
    current_id = ""
    current_path = ""
    for line in Path(args.ignore).read_text().splitlines():
        id_match = re.match(r"^\s*-\s+id:\s*(\S+)\s*$", line)
        if id_match:
            current_id = id_match.group(1)
            current_path = ""
            continue
        path_match = re.match(r"^\s*-\s+([^:\s][^:]*)\s*$", line)
        if path_match and current_id:
            current_path = path_match.group(1).strip()
            continue
        expiry_match = re.match(r"^\s*expired_at:\s*(\d{4}-\d{2}-\d{2})\s*$", line)
        if expiry_match and current_id and current_path:
            actual.add((current_id, current_path, expiry_match.group(1)))
            current_id = ""
            current_path = ""
    if actual != expected:
        raise ContractError(
            "Trivy ignore file does not match active exact-scope waivers: "
            f"expected={sorted(expected)!r} actual={sorted(actual)!r}"
        )
    print(f"Trivy ignore file matches {len(expected)} active exact-scope waiver(s)")


def command_render_trivy_ignore(args: argparse.Namespace) -> None:
    waivers = validate_waivers(
        load_json(Path(args.waivers)), requested_today(args.today)
    )
    entries = sorted(
        (item for item in waivers if item["scanner"] == "trivy"),
        key=lambda item: (item["id"], item["scope"]),
    )
    lines = ["misconfigurations:"]
    for item in entries:
        lines.extend(
            [
                f"  - id: {item['id']}",
                "    paths:",
                f"      - {item['scope']}",
                f"    expired_at: {item['expires']}",
                f'    statement: "Owner {item["owner"]}; {item["rationale"]} See {item["remediation_link"]}"',
            ]
        )
    Path(args.output).write_text("\n".join(lines) + "\n")
    print(f"rendered {len(entries)} active scoped Trivy waiver(s)")


def parser() -> argparse.ArgumentParser:
    result = argparse.ArgumentParser()
    sub = result.add_subparsers(dest="command", required=True)
    item = sub.add_parser("validate-waivers")
    item.add_argument("--waivers", required=True)
    item.add_argument("--schema", required=True)
    item.add_argument("--today")
    item.set_defaults(func=command_validate_waivers)
    item = sub.add_parser("enforce-results")
    item.add_argument("--results", required=True)
    item.add_argument("--waivers", required=True)
    item.add_argument("--today")
    item.set_defaults(func=command_enforce_results)
    item = sub.add_parser("visibility-guard")
    item.add_argument("--visibility", required=True)
    item.add_argument("--hosted-results", required=True)
    item.set_defaults(func=command_visibility_guard)
    item = sub.add_parser("local-sast")
    item.add_argument("--path", required=True)
    item.set_defaults(func=command_local_sast)
    item = sub.add_parser("repository-posture")
    group = item.add_mutually_exclusive_group(required=True)
    group.add_argument("--workflow")
    group.add_argument("--path")
    item.set_defaults(func=command_repository_posture)
    item = sub.add_parser("cross-reference-gates")
    item.add_argument("--manifest", required=True)
    item.set_defaults(func=command_cross_reference)
    item = sub.add_parser("workflow-policy")
    item.add_argument(
        "--visibility",
        choices=["auto", "public", "private", "internal"],
        default="auto",
    )
    item.add_argument("--source-sha", required=True)
    item.set_defaults(func=command_workflow_policy)
    item = sub.add_parser("private-fallback")
    item.add_argument("--path", required=True)
    item.add_argument("--force", action="store_true")
    item.set_defaults(func=command_private_fallback)
    item = sub.add_parser("validate-trivy-ignore")
    item.add_argument("--waivers", required=True)
    item.add_argument("--ignore", required=True)
    item.add_argument("--today")
    item.set_defaults(func=command_validate_trivy_ignore)
    item = sub.add_parser("render-trivy-ignore")
    item.add_argument("--waivers", required=True)
    item.add_argument("--output", required=True)
    item.add_argument("--today")
    item.set_defaults(func=command_render_trivy_ignore)
    return result


def main() -> int:
    try:
        args = parser().parse_args()
        args.func(args)
        return 0
    except ContractError as exc:
        print(f"security contract violation: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
