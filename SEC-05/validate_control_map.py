#!/usr/bin/env python3
"""Validate and optionally execute the SEC-05 version-pinned control map."""

from __future__ import annotations

import argparse
import csv
import datetime as dt
import json
import hashlib
import os
from pathlib import Path
import re
import subprocess
import sys
from typing import Any

ROOT = Path(__file__).resolve().parent.parent
PROFILE_DIR = ROOT / "SEC-05"
MAP_PATH = PROFILE_DIR / "control-map.json"


class ValidationError(Exception):
    """A control-map invariant was violated."""


def _load_json(path: Path) -> dict[str, Any]:
    with path.open(encoding="utf-8") as handle:
        return json.load(handle)


def _source_inventory(source: dict[str, Any]) -> dict[str, str]:
    path = ROOT / source["inventory_path"]
    if not path.is_file():
        raise ValidationError(f"missing source inventory: {source['inventory_path']}")
    actual_digest = hashlib.sha256(path.read_bytes()).hexdigest()
    if actual_digest != source["inventory_sha256"]:
        raise ValidationError(
            f"inventory digest mismatch for {source['standard']}: "
            f"expected={source['inventory_sha256']} actual={actual_digest}"
        )
    if source["standard"] == "ASVS":
        with path.open(newline="", encoding="utf-8") as handle:
            rows = list(csv.DictReader(handle))
        return {row["req_id"]: row["req_description"] for row in rows}
    payload = _load_json(path)
    return {item["id"]: item["title"] for item in payload["controls"]}


def _validate_test_reference(reference: dict[str, Any]) -> None:
    path = ROOT / reference["file"]
    if not path.is_file():
        raise ValidationError(f"test file does not exist: {reference['file']}")
    body = path.read_text(encoding="utf-8")
    name = reference["test"]
    kind = reference["kind"]
    if kind == "go-test":
        if not reference.get("package", "").startswith("./"):
            raise ValidationError(f"invalid Go package for {name}")
        pattern = rf"(?m)^func\s+{re.escape(name)}\s*\("
    elif kind == "python-unittest":
        pattern = rf"(?m)^\s*def\s+{re.escape(name)}\s*\("
    else:
        raise ValidationError(f"unsupported test kind: {kind}")
    if re.search(pattern, body) is None:
        raise ValidationError(f"test {name} not found in {reference['file']}")


def validate(map_path: Path = MAP_PATH) -> dict[str, int]:
    profile = _load_json(map_path)
    if profile.get("schema_version") != 1:
        raise ValidationError("schema_version must be 1")
    if profile.get("profile_id") != "SEC-05":
        raise ValidationError("profile_id must be SEC-05")

    sources = profile.get("sources", [])
    source_by_standard = {source["standard"]: source for source in sources}
    if set(source_by_standard) != {"ASVS", "OWASP-API-TOP-10", "NIST-SP-800-63"}:
        raise ValidationError("exactly the three mandated standards must be pinned")

    inventories: dict[str, dict[str, str]] = {}
    for standard, source in source_by_standard.items():
        for field in ("version", "canonical_uri", "inventory_path", "inventory_sha256", "source_pin"):
            if not source.get(field):
                raise ValidationError(f"{standard} source is missing {field}")
        inventory = _source_inventory(source)
        if len(inventory) != source["inventory_count"]:
            raise ValidationError(
                f"{standard} inventory_count={source['inventory_count']} but file has {len(inventory)}"
            )
        inventories[standard] = inventory

    controls = profile.get("controls", [])
    seen: set[tuple[str, str]] = set()
    mapped_ids: dict[str, set[str]] = {standard: set() for standard in inventories}
    stats = {"applicable": 0, "not_applicable": 0, "waived": 0}
    as_of = dt.date.fromisoformat(profile["as_of"])

    for control in controls:
        standard = control["standard"]
        control_id = control["control_id"]
        key = (standard, control_id)
        if key in seen:
            raise ValidationError(f"duplicate control: {standard}/{control_id}")
        seen.add(key)
        if standard not in inventories or control_id not in inventories[standard]:
            raise ValidationError(f"unknown control: {standard}/{control_id}")
        if control.get("title") != inventories[standard][control_id]:
            raise ValidationError(f"source title mismatch: {standard}/{control_id}")
        if control.get("version") != source_by_standard[standard]["version"]:
            raise ValidationError(f"version mismatch: {standard}/{control_id}")
        mapped_ids[standard].add(control_id)

        disposition = control.get("disposition")
        if disposition not in stats:
            raise ValidationError(f"invalid disposition for {standard}/{control_id}: {disposition}")
        stats[disposition] += 1
        rationale = control.get("rationale", "")
        references = control.get("verification", [])
        waiver = control.get("waiver")

        if disposition == "applicable":
            if len(rationale.strip()) < 20:
                raise ValidationError(f"applicable control lacks scope rationale: {standard}/{control_id}")
            if not references:
                raise ValidationError(f"applicable control has no executable test: {standard}/{control_id}")
            if waiver is not None:
                raise ValidationError(f"applicable control unexpectedly has waiver: {standard}/{control_id}")
            for reference in references:
                _validate_test_reference(reference)
        elif disposition == "not_applicable":
            if len(rationale.strip()) < 20:
                raise ValidationError(f"N/A control lacks rationale: {standard}/{control_id}")
            if references or waiver is not None:
                raise ValidationError(f"N/A control has verification or waiver: {standard}/{control_id}")
        else:
            if references:
                raise ValidationError(f"waived control also has test mapping: {standard}/{control_id}")
            if not isinstance(waiver, dict):
                raise ValidationError(f"waived control lacks waiver: {standard}/{control_id}")
            required = ("status", "owner", "approved_by", "rationale", "expires_on")
            if any(not str(waiver.get(field, "")).strip() for field in required):
                raise ValidationError(f"waiver fields incomplete: {standard}/{control_id}")
            if waiver["status"] != "approved":
                raise ValidationError(f"waiver is not approved: {standard}/{control_id}")
            forbidden = {"tbd", "unassigned", "unknown", "placeholder"}
            if waiver["owner"].strip().lower() in forbidden or waiver["approved_by"].strip().lower() in forbidden:
                raise ValidationError(f"waiver owner/approver is not genuine: {standard}/{control_id}")
            if len(waiver["rationale"].strip()) < 20:
                raise ValidationError(f"waiver rationale is too short: {standard}/{control_id}")
            if dt.date.fromisoformat(waiver["expires_on"]) <= as_of:
                raise ValidationError(f"waiver is expired at profile date: {standard}/{control_id}")

    for standard, inventory in inventories.items():
        missing = set(inventory) - mapped_ids[standard]
        extra = mapped_ids[standard] - set(inventory)
        if missing or extra:
            raise ValidationError(
                f"{standard} inventory mismatch: missing={sorted(missing)} extra={sorted(extra)}"
            )

    stats["total"] = len(controls)
    return stats


def run_tests(map_path: Path = MAP_PATH) -> None:
    profile = _load_json(map_path)
    go_tests: dict[str, set[str]] = {}
    python_files: set[str] = set()
    for control in profile["controls"]:
        if control["disposition"] != "applicable":
            continue
        for reference in control["verification"]:
            if reference["kind"] == "go-test":
                go_tests.setdefault(reference["package"], set()).add(reference["test"])
            elif reference["kind"] == "python-unittest":
                python_files.add(reference["file"])

    for package, tests in sorted(go_tests.items()):
        pattern = "^(?:" + "|".join(sorted(map(re.escape, tests))) + ")$"
        command = ["go", "test", package, "-run", pattern, "-count=1"]
        print("+", " ".join(command), flush=True)
        subprocess.run(command, cwd=ROOT, env=os.environ.copy(), check=True)
    for file_name in sorted(python_files):
        command = [sys.executable, file_name]
        print("+", " ".join(command), flush=True)
        subprocess.run(command, cwd=ROOT, env=os.environ.copy(), check=True)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--map", type=Path, default=MAP_PATH)
    parser.add_argument("--run-tests", action="store_true")
    args = parser.parse_args()
    try:
        stats = validate(args.map)
        print(
            "control-map valid: "
            f"total={stats['total']} applicable={stats['applicable']} "
            f"not-applicable={stats['not_applicable']} waived={stats['waived']}"
        )
        if args.run_tests:
            run_tests(args.map)
    except (ValidationError, KeyError, TypeError, ValueError, json.JSONDecodeError) as exc:
        print(f"control-map invalid: {exc}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
