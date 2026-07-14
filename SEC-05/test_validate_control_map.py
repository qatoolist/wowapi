#!/usr/bin/env python3
"""Regression tests for SEC-05 control-map validation invariants."""

from __future__ import annotations

import copy
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest

HERE = Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("validate_control_map", HERE / "validate_control_map.py")
assert SPEC is not None and SPEC.loader is not None
VALIDATOR = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(VALIDATOR)


class TestControlMapValidator(unittest.TestCase):
    def setUp(self) -> None:
        self.profile = json.loads((HERE / "control-map.json").read_text(encoding="utf-8"))

    def validate_copy(self, profile: dict) -> dict[str, int]:
        with tempfile.NamedTemporaryFile("w", suffix=".json", encoding="utf-8") as handle:
            json.dump(profile, handle)
            handle.flush()
            return VALIDATOR.validate(Path(handle.name))

    def test_committed_profile_is_complete(self) -> None:
        stats = VALIDATOR.validate()
        self.assertEqual(stats["total"], sum(stats[key] for key in ("applicable", "not_applicable", "waived")))
        self.assertGreater(stats["applicable"], 0)

    def test_rejects_unmapped_control(self) -> None:
        profile = copy.deepcopy(self.profile)
        profile["controls"].pop()
        with self.assertRaisesRegex(VALIDATOR.ValidationError, "inventory mismatch"):
            self.validate_copy(profile)

    def test_rejects_inventory_digest_mismatch(self) -> None:
        profile = copy.deepcopy(self.profile)
        profile["sources"][0]["inventory_sha256"] = "0" * 64
        with self.assertRaisesRegex(VALIDATOR.ValidationError, "inventory digest mismatch"):
            self.validate_copy(profile)

    def test_rejects_source_title_drift(self) -> None:
        profile = copy.deepcopy(self.profile)
        profile["controls"][0]["title"] = "altered source text"
        with self.assertRaisesRegex(VALIDATOR.ValidationError, "source title mismatch"):
            self.validate_copy(profile)

    def test_rejects_dangling_test_reference(self) -> None:
        profile = copy.deepcopy(self.profile)
        control = next(item for item in profile["controls"] if item["disposition"] == "applicable")
        control["verification"][0]["test"] = "TestDoesNotExist"
        with self.assertRaisesRegex(VALIDATOR.ValidationError, "not found"):
            self.validate_copy(profile)

    def test_rejects_invalid_waiver(self) -> None:
        profile = copy.deepcopy(self.profile)
        control = next(item for item in profile["controls"] if item["disposition"] == "applicable")
        control["disposition"] = "waived"
        control["verification"] = []
        control["waiver"] = {
            "status": "approved",
            "owner": "unassigned",
            "approved_by": "TBD",
            "rationale": "This is intentionally invalid for the regression test.",
            "expires_on": "2099-01-01",
        }
        with self.assertRaisesRegex(VALIDATOR.ValidationError, "not genuine"):
            self.validate_copy(profile)


if __name__ == "__main__":
    unittest.main(verbosity=2)
