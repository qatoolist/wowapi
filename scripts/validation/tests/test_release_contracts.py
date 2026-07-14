#!/usr/bin/env python3
from __future__ import annotations

import copy
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[3]
TOOL = ROOT / "scripts/validation/release_contract.py"
VERIFY = ROOT / "scripts/validation/verify_release.sh"
MANIFEST = ROOT / "ci/release-gates.yaml"
SCHEMA = ROOT / "ci/release-gates.schema.json"


def sha256(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


class ReleaseContractTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temp = Path(tempfile.mkdtemp(prefix="wowapi-release-contract-"))
        self.candidate = self.temp / "candidate"
        self.candidate.mkdir()
        self.source_sha = "a" * 40
        self.version = "1.2.3"
        (self.candidate / "gate-results.json").write_text(json.dumps({
            "schema_version": 1,
            "source_sha": self.source_sha,
            "manifest_sha256": "b" * 64,
            "completed_wave": 6,
            "status": "passed",
            "gates": [{"id": "unit", "status": "passed", "evidence_sha256": "c" * 64}],
        }, sort_keys=True) + "\n")
        (self.candidate / "wowapi_1.2.3_linux_amd64.tar.gz").write_bytes(b"linux archive\n")
        (self.candidate / "wowapi_1.2.3_darwin_arm64.tar.gz").write_bytes(b"darwin archive\n")
        (self.candidate / "wowapi_1.2.3_windows_amd64.zip").write_bytes(b"windows archive\n")
        checksums = "".join(
            f"{sha256(self.candidate / name)}  {name}\n"
            for name in (
                "wowapi_1.2.3_linux_amd64.tar.gz",
                "wowapi_1.2.3_darwin_arm64.tar.gz",
                "wowapi_1.2.3_windows_amd64.zip",
            )
        )
        (self.candidate / "checksums.txt").write_text(checksums)
        (self.candidate / "checksums.txt.cosign.bundle").write_text(json.dumps({
            "subject_sha256": hashlib.sha256(checksums.encode()).hexdigest(),
            "certificate_identity": "https://github.com/qatoolist/wowapi/.github/workflows/release.yml@refs/tags/v1.2.3",
            "oidc_issuer": "https://token.actions.githubusercontent.com",
        }, sort_keys=True) + "\n")
        (self.candidate / "archive-sbom.spdx.json").write_text('{"spdxVersion":"SPDX-2.3"}\n')
        (self.candidate / "archive-provenance.intoto.jsonl").write_text(json.dumps({
            "subject_sha256": sha256(self.candidate / "wowapi_1.2.3_linux_amd64.tar.gz"),
            "source_sha": self.source_sha,
            "builder": "https://github.com/qatoolist/wowapi/.github/workflows/release.yml",
        }) + "\n")
        (self.candidate / "wowapi-image.oci.tar").write_bytes(b"oci image layout\n")
        (self.candidate / "image-sbom.spdx.json").write_text('{"spdxVersion":"SPDX-2.3"}\n')
        (self.candidate / "image-provenance.intoto.jsonl").write_text(json.dumps({
            "subject_digest": "sha256:" + "d" * 64,
            "source_sha": self.source_sha,
        }) + "\n")
        (self.candidate / "artifact-security.trivy.json").write_text('{"Results":[]}\n')
        (self.candidate / "image-security.trivy.json").write_text('{"Results":[]}\n')
        binary = self.candidate / "wowapi"
        binary.write_text("#!/bin/sh\nprintf 'wowapi 1.2.3\\n'\n")
        binary.chmod(0o755)
        descriptor = {
            "schema_version": 1,
            "version": self.version,
            "source_sha": self.source_sha,
            "gate_results": "gate-results.json",
            "artifacts": [
                {"path": "wowapi_1.2.3_linux_amd64.tar.gz", "kind": "archive", "platform": "linux/amd64"},
                {"path": "wowapi_1.2.3_darwin_arm64.tar.gz", "kind": "archive", "platform": "darwin/arm64"},
                {"path": "wowapi_1.2.3_windows_amd64.zip", "kind": "archive", "platform": "windows/amd64"},
                {"path": "checksums.txt", "kind": "checksums"},
                {"path": "checksums.txt.cosign.bundle", "kind": "signature"},
                {"path": "archive-sbom.spdx.json", "kind": "sbom"},
                {"path": "archive-provenance.intoto.jsonl", "kind": "provenance"},
                {"path": "wowapi-image.oci.tar", "kind": "image", "platforms": ["linux/amd64", "linux/arm64"], "digest": "sha256:" + "d" * 64},
                {"path": "image-sbom.spdx.json", "kind": "image-sbom"},
                {"path": "image-provenance.intoto.jsonl", "kind": "image-provenance"},
                {"path": "wowapi", "kind": "verification-binary", "platform": "linux/amd64"},
                {"path": "artifact-security.trivy.json", "kind": "artifact-security"},
                {"path": "image-security.trivy.json", "kind": "image-security"},
            ],
        }
        (self.candidate / "candidate-descriptor.json").write_text(json.dumps(descriptor, sort_keys=True) + "\n")

    def tearDown(self) -> None:
        shutil.rmtree(self.temp)

    def run_tool(self, *args: str, ok: bool = True, env: dict[str, str] | None = None) -> subprocess.CompletedProcess[str]:
        result = subprocess.run(
            ["python3", str(TOOL), *args], cwd=ROOT, text=True, capture_output=True,
            env={**os.environ, **(env or {})}, check=False,
        )
        if ok and result.returncode != 0:
            self.fail(f"command failed ({result.returncode}): {result.stderr}\n{result.stdout}")
        if not ok and result.returncode == 0:
            self.fail(f"command unexpectedly passed: {result.stdout}")
        return result

    def create_manifest(self) -> Path:
        manifest = self.candidate / "release-manifest.json"
        self.run_tool("create-manifest", "--candidate", str(self.candidate), "--descriptor", "candidate-descriptor.json", "--output", str(manifest))
        attestation = self.candidate / "release-manifest.attestation.json"
        attestation.write_text(json.dumps({
            "subject_sha256": sha256(manifest),
            "source_sha": self.source_sha,
            "workflow_identity": "https://github.com/qatoolist/wowapi/.github/workflows/release.yml@refs/tags/v1.2.3",
            "oidc_issuer": "https://token.actions.githubusercontent.com",
        }, sort_keys=True) + "\n")
        return manifest

    def verify_candidate(self, ok: bool = True) -> subprocess.CompletedProcess[str]:
        return self.run_tool(
            "verify-candidate", "--candidate", str(self.candidate),
            "--attestation", "release-manifest.attestation.json", ok=ok,
        )

    def test_describe_candidate_registers_every_input_and_required_kind(self) -> None:
        (self.candidate / "candidate-descriptor.json").unlink()
        self.run_tool(
            "describe-candidate",
            "--candidate", str(self.candidate),
            "--version", self.version,
            "--source-sha", self.source_sha,
            "--gate-results", "gate-results.json",
            "--image-digest", "sha256:" + "d" * 64,
            "--output", str(self.candidate / "candidate-descriptor.json"),
        )
        descriptor = json.loads((self.candidate / "candidate-descriptor.json").read_text())
        paths = {item["path"] for item in descriptor["artifacts"]}
        self.assertEqual(
            {path.name for path in self.candidate.iterdir() if path.is_file()} - {"gate-results.json", "candidate-descriptor.json"},
            paths,
        )
        self.assertTrue({"archive", "signature", "sbom", "provenance", "image", "image-sbom", "image-provenance", "verification-binary", "artifact-security", "image-security"} <= {item["kind"] for item in descriptor["artifacts"]})

    def test_candidate_without_blocking_security_reports_is_rejected(self) -> None:
        descriptor_path = self.candidate / "candidate-descriptor.json"
        descriptor = json.loads(descriptor_path.read_text())
        descriptor["artifacts"] = [
            item for item in descriptor["artifacts"]
            if item["kind"] != "image-security"
        ]
        descriptor_path.write_text(json.dumps(descriptor, sort_keys=True) + "\n")
        failure = self.run_tool(
            "create-manifest",
            "--candidate", str(self.candidate),
            "--descriptor", "candidate-descriptor.json",
            "--output", str(self.candidate / "release-manifest.json"),
            ok=False,
        )
        self.assertIn("image security report", failure.stderr)

    def test_manifest_schema_rejects_missing_required_field(self) -> None:
        valid = self.run_tool("validate-gates", "--manifest", str(MANIFEST), "--schema", str(SCHEMA))
        self.assertIn("valid", valid.stdout)
        malformed = self.temp / "malformed.json"
        data = json.loads(MANIFEST.read_text())
        del data["gates"][0]["owner"]
        malformed.write_text(json.dumps(data))
        failure = self.run_tool("validate-gates", "--manifest", str(malformed), "--schema", str(SCHEMA), ok=False)
        self.assertIn("owner", failure.stderr)

    def test_moving_tag_target_is_rejected(self) -> None:
        repo = self.temp / "tag-repo"
        repo.mkdir()
        subprocess.run(["git", "init", "-q", str(repo)], check=True)
        subprocess.run(["git", "-C", str(repo), "config", "user.email", "fixture@example.test"], check=True)
        subprocess.run(["git", "-C", str(repo), "config", "user.name", "fixture"], check=True)
        (repo / "value").write_text("first")
        subprocess.run(["git", "-C", str(repo), "add", "value"], check=True)
        subprocess.run(["git", "-C", str(repo), "commit", "-qm", "first"], check=True)
        original = subprocess.check_output(["git", "-C", str(repo), "rev-parse", "HEAD"], text=True).strip()
        subprocess.run(["git", "-C", str(repo), "-c", "tag.gpgSign=false", "tag", "v1.2.3"], check=True)
        self.run_tool("verify-tag", "--repo", str(repo), "--tag", "v1.2.3", "--source-sha", original)
        (repo / "value").write_text("second")
        subprocess.run(["git", "-C", str(repo), "commit", "-qam", "second"], check=True)
        subprocess.run(["git", "-C", str(repo), "-c", "tag.gpgSign=false", "tag", "-f", "v1.2.3"], check=True, capture_output=True)
        failure = self.run_tool("verify-tag", "--repo", str(repo), "--tag", "v1.2.3", "--source-sha", original, ok=False)
        self.assertIn("tag target", failure.stderr)

    def test_failing_exact_sha_gate_is_attested_and_blocks_candidate(self) -> None:
        repo = self.temp / "repo"
        repo.mkdir()
        subprocess.run(["git", "init", "-q", str(repo)], check=True)
        subprocess.run(["git", "-C", str(repo), "config", "user.email", "fixture@example.test"], check=True)
        subprocess.run(["git", "-C", str(repo), "config", "user.name", "fixture"], check=True)
        (repo / "fail.sh").write_text("#!/bin/sh\nexit 23\n")
        (repo / "fail.sh").chmod(0o755)
        subprocess.run(["git", "-C", str(repo), "add", "fail.sh"], check=True)
        subprocess.run(["git", "-C", str(repo), "commit", "-qm", "seed failing gate"], check=True)
        sha = subprocess.check_output(["git", "-C", str(repo), "rev-parse", "HEAD"], text=True).strip()
        gates = self.temp / "gates.json"
        gates.write_text(json.dumps({"schema_version": 1, "completed_wave": 6, "gates": [{
            "id": "seeded-failure", "job_ref": "fixture", "command": "./fail.sh", "owner": "release-security",
            "required_from_wave": 0, "timeout_minutes": 1, "evidence_artifact": "seeded-failure.json",
        }]}, sort_keys=True))
        output = self.temp / "gate-results.json"
        result = self.run_tool("run-gates", "--manifest", str(gates), "--source-sha", sha, "--repo", str(repo), "--output", str(output), ok=False)
        self.assertIn("seeded-failure", result.stderr)
        gate_results = json.loads(output.read_text())
        self.assertEqual(sha, gate_results["source_sha"])
        self.assertEqual("failed", gate_results["status"])
        shutil.copy2(output, self.candidate / "gate-results.json")
        descriptor = json.loads((self.candidate / "candidate-descriptor.json").read_text())
        descriptor["source_sha"] = sha
        (self.candidate / "candidate-descriptor.json").write_text(json.dumps(descriptor))
        blocked = self.run_tool("create-manifest", "--candidate", str(self.candidate), "--descriptor", "candidate-descriptor.json", "--output", str(self.candidate / "release-manifest.json"), ok=False)
        self.assertIn("required gates did not pass", blocked.stderr)

    def test_same_sha_gate_results_are_byte_identical(self) -> None:
        repo = self.temp / "repo"
        repo.mkdir()
        subprocess.run(["git", "init", "-q", str(repo)], check=True)
        subprocess.run(["git", "-C", str(repo), "config", "user.email", "fixture@example.test"], check=True)
        subprocess.run(["git", "-C", str(repo), "config", "user.name", "fixture"], check=True)
        (repo / "pass.sh").write_text("#!/bin/sh\nprintf ok\n")
        (repo / "pass.sh").chmod(0o755)
        subprocess.run(["git", "-C", str(repo), "add", "pass.sh"], check=True)
        subprocess.run(["git", "-C", str(repo), "commit", "-qm", "seed passing gate"], check=True)
        sha = subprocess.check_output(["git", "-C", str(repo), "rev-parse", "HEAD"], text=True).strip()
        gates = self.temp / "gates.json"
        gates.write_text(json.dumps({"schema_version": 1, "completed_wave": 6, "gates": [{
            "id": "seeded-pass", "job_ref": "fixture", "command": "./pass.sh", "owner": "release-security",
            "required_from_wave": 0, "timeout_minutes": 1, "evidence_artifact": "seeded-pass.json",
        }]}, sort_keys=True))
        first, second = self.temp / "first.json", self.temp / "second.json"
        self.run_tool("run-gates", "--manifest", str(gates), "--source-sha", sha, "--repo", str(repo), "--output", str(first))
        self.run_tool("run-gates", "--manifest", str(gates), "--source-sha", sha, "--repo", str(repo), "--output", str(second))
        self.assertEqual(first.read_bytes(), second.read_bytes())

    def test_candidate_rejects_altered_gate_tag_manifest_archive_and_image(self) -> None:
        manifest = self.create_manifest()
        self.verify_candidate()
        pristine = {p.name: p.read_bytes() for p in self.candidate.iterdir() if p.is_file()}
        mutations = {
            "gate-results.json": lambda p: p.write_bytes(p.read_bytes() + b" "),
            "release-manifest.json": lambda p: p.write_bytes(p.read_bytes().replace(self.source_sha.encode(), ("e" * 40).encode(), 1)),
            "wowapi_1.2.3_linux_amd64.tar.gz": lambda p: p.write_bytes(p.read_bytes() + b"tamper"),
            "wowapi-image.oci.tar": lambda p: p.write_bytes(p.read_bytes() + b"tamper"),
        }
        for name, mutate in mutations.items():
            for restore_name, content in pristine.items():
                (self.candidate / restore_name).write_bytes(content)
            mutate(self.candidate / name)
            with self.subTest(name=name):
                self.verify_candidate(ok=False)
        for restore_name, content in pristine.items():
            (self.candidate / restore_name).write_bytes(content)
        data = json.loads((self.candidate / "candidate-descriptor.json").read_text())
        data["source_sha"] = "f" * 40
        (self.candidate / "candidate-descriptor.json").write_text(json.dumps(data))
        mismatch = self.run_tool("create-manifest", "--candidate", str(self.candidate), "--descriptor", "candidate-descriptor.json", "--output", str(manifest), ok=False)
        self.assertIn("tag/source SHA", mismatch.stderr)

    def test_publish_accepts_only_attested_manifested_bytes(self) -> None:
        self.create_manifest()
        destination = self.temp / "published"
        self.run_tool("publish", "--candidate", str(self.candidate), "--destination", str(destination), "--attestation", "release-manifest.attestation.json")
        self.assertEqual(sha256(self.candidate / "wowapi-image.oci.tar"), sha256(destination / "wowapi-image.oci.tar"))
        (self.candidate / "unmanifested.bin").write_bytes(b"not allowed")
        rejected = self.run_tool("publish", "--candidate", str(self.candidate), "--destination", str(self.temp / "rejected"), "--attestation", "release-manifest.attestation.json", ok=False)
        self.assertIn("unmanifested", rejected.stderr)

    def run_verify(self, release_dir: Path, source_sha: str | None = None, ok: bool = True) -> subprocess.CompletedProcess[str]:
        result = subprocess.run(
            [str(VERIFY), self.version, source_sha or self.source_sha], cwd=ROOT, text=True, capture_output=True,
            env={**os.environ, "WOWAPI_RELEASE_DIR": str(release_dir), "WOWAPI_OFFLINE_VERIFY": "1"}, check=False,
        )
        if ok and result.returncode != 0:
            self.fail(f"verify failed ({result.returncode}): {result.stderr}\n{result.stdout}")
        if not ok and result.returncode == 0:
            self.fail("verify unexpectedly passed")
        return result

    def run_online_verify(
        self,
        release_dir: Path,
        *,
        cosign_status: int = 0,
        ok: bool = True,
    ) -> tuple[subprocess.CompletedProcess[str], list[str]]:
        fake_bin = self.temp / "fake-bin"
        fake_bin.mkdir(exist_ok=True)
        for name, body in {
            "gh": "#!/bin/sh\nexit 0\n",
            "cosign": (
                "#!/bin/sh\n"
                "printf '%s\\n' \"$@\" > \"$FAKE_COSIGN_ARGS\"\n"
                "if [ \"${FAKE_COSIGN_STATUS:-0}\" -ne 0 ]; then\n"
                "  echo 'cosign rejected signature bundle' >&2\n"
                "  exit \"$FAKE_COSIGN_STATUS\"\n"
                "fi\n"
            ),
        }.items():
            tool = fake_bin / name
            tool.write_text(body)
            tool.chmod(0o755)
        args_file = self.temp / "cosign-args"
        result = subprocess.run(
            [str(VERIFY), self.version, self.source_sha],
            cwd=ROOT,
            text=True,
            capture_output=True,
            env={
                **os.environ,
                "PATH": f"{fake_bin}{os.pathsep}{os.environ['PATH']}",
                "WOWAPI_RELEASE_DIR": str(release_dir),
                "FAKE_COSIGN_ARGS": str(args_file),
                "FAKE_COSIGN_STATUS": str(cosign_status),
            },
            check=False,
        )
        if ok and result.returncode != 0:
            self.fail(f"online verify failed ({result.returncode}): {result.stderr}\n{result.stdout}")
        if not ok and result.returncode == 0:
            self.fail("online verify unexpectedly passed")
        args = args_file.read_text().splitlines() if args_file.exists() else []
        return result, args

    def test_online_verifier_uses_cosign_and_accepts_real_bundle_schema(self) -> None:
        (self.candidate / "checksums.txt.cosign.bundle").write_text(
            '{"mediaType":"application/vnd.dev.sigstore.bundle.v0.3+json","verificationMaterial":{}}\n'
        )
        self.create_manifest()
        published = self.temp / "published"
        self.run_tool(
            "publish",
            "--candidate",
            str(self.candidate),
            "--destination",
            str(published),
            "--attestation",
            "release-manifest.attestation.json",
        )
        _, args = self.run_online_verify(published)
        self.assertEqual("verify-blob", args[0])
        self.assertIn("--bundle", args)
        self.assertIn("--certificate-identity", args)
        self.assertIn("--certificate-oidc-issuer", args)
        self.assertEqual(
            f"https://github.com/qatoolist/wowapi/.github/workflows/release.yml@refs/tags/v{self.version}",
            args[args.index("--certificate-identity") + 1],
        )
        self.assertEqual("https://token.actions.githubusercontent.com", args[args.index("--certificate-oidc-issuer") + 1])
        self.assertEqual(
            published.joinpath("checksums.txt.cosign.bundle").resolve(),
            Path(args[args.index("--bundle") + 1]).resolve(),
        )
        self.assertEqual(published.joinpath("checksums.txt").resolve(), Path(args[-1]).resolve())

    def test_online_verifier_fails_closed_when_cosign_rejects_bundle(self) -> None:
        self.create_manifest()
        published = self.temp / "published"
        self.run_tool(
            "publish",
            "--candidate",
            str(self.candidate),
            "--destination",
            str(published),
            "--attestation",
            "release-manifest.attestation.json",
        )
        result, _ = self.run_online_verify(published, cosign_status=23, ok=False)
        self.assertIn("cosign rejected signature bundle", result.stderr)

    def test_clean_verifier_golden_failures(self) -> None:
        self.create_manifest()
        published = self.temp / "published"
        self.run_tool("publish", "--candidate", str(self.candidate), "--destination", str(published), "--attestation", "release-manifest.attestation.json")
        self.run_verify(published)
        pristine = {p.name: p.read_bytes() for p in published.iterdir() if p.is_file()}

        with self.subTest(case="wrong source SHA"):
            self.assertIn("source SHA", self.run_verify(published, "9" * 40, ok=False).stderr)
        cases = {
            "stripped signature": ("checksums.txt.cosign.bundle", lambda p: p.unlink(), "signature"),
            "missing SBOM attestation": ("image-sbom.spdx.json", lambda p: p.unlink(), "SBOM"),
            "missing provenance attestation": ("image-provenance.intoto.jsonl", lambda p: p.unlink(), "provenance"),
            "tampered provenance source SHA": ("archive-provenance.intoto.jsonl", lambda p: p.write_bytes(p.read_bytes().replace(self.source_sha.encode(), ("8" * 40).encode())), "provenance"),
            "wrong platforms": ("release-manifest.json", lambda p: p.write_bytes(p.read_bytes().replace(b'"linux/arm64"', b'"linux/s390x"')), "platform"),
            "tampered manifest hash": ("release-manifest.attestation.json", lambda p: p.write_bytes(p.read_bytes().replace(sha256(published / "release-manifest.json").encode(), ("0" * 64).encode())), "attestation"),
        }
        for label, (name, mutate, expected) in cases.items():
            for restore_name, content in pristine.items():
                (published / restore_name).write_bytes(content)
            mutate(published / name)
            with self.subTest(case=label):
                self.assertIn(expected.lower(), self.run_verify(published, ok=False).stderr.lower())

    def test_corrupted_publish_does_not_move_latest(self) -> None:
        self.create_manifest()
        registry = self.temp / "registry"
        registry.mkdir()
        (registry / "latest").write_text("1.2.2\n")
        (self.candidate / "wowapi-image.oci.tar").write_bytes(b"corrupt after manifest")
        result = self.run_tool("promote", "--candidate", str(self.candidate), "--registry", str(registry), "--attestation", "release-manifest.attestation.json", ok=False)
        self.assertIn("hash mismatch", result.stderr)
        self.assertEqual("1.2.2\n", (registry / "latest").read_text())


if __name__ == "__main__":
    unittest.main(verbosity=2)
