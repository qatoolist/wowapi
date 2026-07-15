#!/usr/bin/env python3
from __future__ import annotations

import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[3]
TOOL = ROOT / "scripts/validation/security_contract.py"
GATES = ROOT / "ci/release-gates.yaml"
WAIVERS = ROOT / "ci/security-waivers.yaml"
WAIVER_SCHEMA = ROOT / "ci/security-waivers.schema.json"
FIXTURES = ROOT / "scripts/validation/fixtures/security"


class SecurityContractTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temp = Path(tempfile.mkdtemp(prefix="wowapi-security-contract-"))

    def tearDown(self) -> None:
        shutil.rmtree(self.temp)

    def run_tool(self, *args: str, ok: bool = True) -> subprocess.CompletedProcess[str]:
        result = subprocess.run(["python3", str(TOOL), *args], cwd=ROOT, text=True, capture_output=True, env=os.environ, check=False)
        if ok and result.returncode != 0:
            self.fail(f"command failed ({result.returncode}): {result.stderr}\n{result.stdout}")
        if not ok and result.returncode == 0:
            self.fail(f"command unexpectedly passed: {result.stdout}")
        return result

    def test_trivy_ignore_file_exactly_matches_active_scoped_waivers(self) -> None:
        ignore = ROOT / ".trivyignore.yaml"
        self.run_tool("validate-trivy-ignore", "--waivers", str(WAIVERS), "--ignore", str(ignore), "--today", "2026-07-13")
        altered = self.temp / ".trivyignore.yaml"
        altered.write_text(ignore.read_text().replace("Dockerfile", "Otherfile", 1))
        self.assertIn("does not match", self.run_tool("validate-trivy-ignore", "--waivers", str(WAIVERS), "--ignore", str(altered), "--today", "2026-07-13", ok=False).stderr)

    def test_waiver_schema_accepts_scoped_unexpired_and_rejects_missing_or_expired(self) -> None:
        self.assertIn("valid", self.run_tool("validate-waivers", "--waivers", str(WAIVERS), "--schema", str(WAIVER_SCHEMA), "--today", "2026-07-13").stdout)
        valid = {
            "schema_version": 1,
            "waivers": [{
                "id": "CVE-2099-0001", "scanner": "trivy", "scope": "fixtures/image@sha256:abc",
                "owner": "release-security", "rationale": "Fixture-only false positive",
                "expires": "2026-08-01", "remediation_link": "https://github.com/qatoolist/wowapi/issues/1",
            }],
        }
        missing = self.temp / "missing.json"
        missing_data = json.loads(json.dumps(valid))
        del missing_data["waivers"][0]["owner"]
        missing.write_text(json.dumps(missing_data))
        self.assertIn("owner", self.run_tool("validate-waivers", "--waivers", str(missing), "--schema", str(WAIVER_SCHEMA), "--today", "2026-07-13", ok=False).stderr)
        expired = self.temp / "expired.json"
        expired_data = json.loads(json.dumps(valid))
        expired_data["waivers"][0]["expires"] = "2026-07-12"
        expired.write_text(json.dumps(expired_data))
        self.assertIn("expired", self.run_tool("validate-waivers", "--waivers", str(expired), "--schema", str(WAIVER_SCHEMA), "--today", "2026-07-13", ok=False).stderr)

    def test_each_seeded_security_defect_fails_closed(self) -> None:
        seeded = json.loads((FIXTURES / "seeded-failures.json").read_text())
        required = tuple(item["class"] for item in seeded["results"])
        for defect in required:
            results = self.temp / f"{defect}.json"
            document = json.loads(json.dumps(seeded))
            for item in document["results"]:
                item["status"] = "failed" if item["class"] == defect else "passed"
            results.write_text(json.dumps(document))
            with self.subTest(defect=defect):
                failure = self.run_tool("enforce-results", "--results", str(results), "--waivers", str(WAIVERS), "--today", "2026-07-13", ok=False)
                self.assertIn(defect, failure.stderr)

    def test_valid_waiver_is_exact_scope_not_global(self) -> None:
        waivers = self.temp / "waivers.json"
        waivers.write_text(json.dumps({"schema_version": 1, "waivers": [{
            "id": "CVE-2099-0001", "scanner": "trivy", "scope": "module/example@v1.2.3",
            "owner": "release-security", "rationale": "Temporary mitigation is deployed",
            "expires": "2026-08-01", "remediation_link": "https://github.com/qatoolist/wowapi/issues/1",
        }]}))
        scoped = self.temp / "scoped.json"
        scoped.write_text(json.dumps({"schema_version": 1, "source_sha": "a" * 40, "results": [{
            "class": "reachable-vulnerability", "scanner": "trivy", "finding_id": "CVE-2099-0001",
            "scope": "module/example@v1.2.3", "status": "failed",
        }]}))
        self.run_tool("enforce-results", "--results", str(scoped), "--waivers", str(waivers), "--today", "2026-07-13")
        wrong_scope = json.loads(scoped.read_text())
        wrong_scope["results"][0]["scope"] = "module/other@v1.2.3"
        scoped.write_text(json.dumps(wrong_scope))
        self.assertIn("unwaived", self.run_tool("enforce-results", "--results", str(scoped), "--waivers", str(waivers), "--today", "2026-07-13", ok=False).stderr)

    def test_public_visibility_requires_all_hosted_scanners_and_private_activates_fallback(self) -> None:
        all_success = self.temp / "all-success.json"
        all_success.write_text(json.dumps({"dependency-review": "success", "codeql": "success", "scorecard": "success"}))
        self.run_tool("visibility-guard", "--visibility", "public", "--hosted-results", str(all_success))
        missing = self.temp / "missing.json"
        missing.write_text(json.dumps({"dependency-review": "success", "codeql": "success"}))
        self.assertIn("scorecard", self.run_tool("visibility-guard", "--visibility", "public", "--hosted-results", str(missing), ok=False).stderr)
        private = self.run_tool("visibility-guard", "--visibility", "private", "--hosted-results", str(missing))
        self.assertIn("fallback-required", private.stdout)

    def test_public_workflow_policy_retries_concurrent_hosted_runs(self) -> None:
        fake_bin = self.temp / "bin"
        fake_bin.mkdir()
        counter = self.temp / "calls"
        gh = fake_bin / "gh"
        gh.write_text(
            "#!/bin/sh\n"
            "if [ \"$1\" = api ]; then echo public; exit 0; fi\n"
            f"count=$(cat '{counter}' 2>/dev/null || echo 0)\n"
            "count=$((count + 1))\n"
            f"echo \"$count\" > '{counter}'\n"
            "if [ \"$count\" -eq 1 ]; then echo '[{\"conclusion\":\"\"}]'; "
            "else echo '[{\"conclusion\":\"success\"}]'; fi\n"
        )
        gh.chmod(0o755)
        prior = dict(os.environ)
        try:
            os.environ["PATH"] = str(fake_bin) + os.pathsep + prior["PATH"]
            os.environ["GITHUB_REPOSITORY"] = "qatoolist/wowapi"
            os.environ["WOWAPI_HOSTED_SCAN_ATTEMPTS"] = "2"
            os.environ["WOWAPI_HOSTED_SCAN_INTERVAL"] = "0"
            result = self.run_tool("workflow-policy", "--visibility", "auto", "--source-sha", "a" * 40)
            self.assertIn("passed", result.stdout)
        finally:
            os.environ.clear()
            os.environ.update(prior)

    def test_private_fallback_catches_seeded_unsafe_pattern_and_posture_defect(self) -> None:
        source = self.temp / "unsafe.go"
        source.write_text((FIXTURES / "unsafe.go.seeded").read_text())
        self.assertIn("unsafe shell", self.run_tool("local-sast", "--path", str(self.temp), ok=False).stderr)
        source.write_text("package fixture\nfunc safe() {}\n")
        self.run_tool("local-sast", "--path", str(self.temp))
        workflow = self.temp / "workflow.yml"
        workflow.write_text((FIXTURES / "unpinned-workflow.yml.seeded").read_text())
        self.assertIn("full commit SHA", self.run_tool("repository-posture", "--workflow", str(workflow), ok=False).stderr)

    def test_every_required_security_result_has_exactly_one_manifest_entry(self) -> None:
        result = self.run_tool("cross-reference-gates", "--manifest", str(GATES))
        self.assertIn("exactly once", result.stdout)
        data = json.loads(GATES.read_text())
        duplicate = self.temp / "duplicate.json"
        data["gates"].append(dict(data["gates"][0]))
        duplicate.write_text(json.dumps(data))
        self.run_tool("cross-reference-gates", "--manifest", str(duplicate), ok=False)


if __name__ == "__main__":
    unittest.main(verbosity=2)
