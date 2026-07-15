#!/usr/bin/env python3
"""Verify the lifecycle state required before SEC-05 can close."""

from __future__ import annotations

from pathlib import Path
import re

ROOT = Path(__file__).resolve().parent.parent
CHECKS = {
    "SEC-01/W03-E01-S001": "impl/waves/wave-03-identity-and-session-security/epics/epic-001-server-side-session-state/stories/story-001-grant-schema-and-membership",
    "SEC-01/W03-E01-S002": "impl/waves/wave-03-identity-and-session-security/epics/epic-001-server-side-session-state/stories/story-002-capacity-and-privileged-resolver",
    "SEC-01/W03-E01-S003": "impl/waves/wave-03-identity-and-session-security/epics/epic-001-server-side-session-state/stories/story-003-assurance-and-credential-schemes",
    "SEC-01/W03-E01-S004": "impl/waves/wave-03-identity-and-session-security/epics/epic-001-server-side-session-state/stories/story-004-cross-repo-cutover-plan",
    "SEC-06/W03-E02-S001": "impl/waves/wave-03-identity-and-session-security/epics/epic-002-outbound-security-governance/stories/story-001-outbound-security-governance",
    "SEC-03/W03-E03-S001": "impl/waves/wave-03-identity-and-session-security/epics/epic-003-webhook-authenticated-replay/stories/story-001-webhook-authenticated-replay",
    "SEC-04/W05-E04-S002": "impl/waves/wave-05-application-model-and-layering/epics/epic-004-wiring-and-cache-hygiene/stories/story-002-authz-cache-bounding",
}


def frontmatter_status(path: Path) -> str:
    text = path.read_text(encoding="utf-8")
    match = re.search(r"(?m)^status:\s*([^\n]+)$", text)
    if match is None:
        return "missing"
    return match.group(1).strip()


def main() -> int:
    failures = 0
    for label, directory in CHECKS.items():
        story_path = ROOT / directory / "story.md"
        closure_path = ROOT / directory / "closure.md"
        story_status = frontmatter_status(story_path)
        closure_status = frontmatter_status(closure_path)
        consistent = story_status == closure_status
        accepted = story_status == closure_status == "accepted"
        result = "PASS" if accepted else "FAIL"
        if not accepted:
            failures += 1
        consistency = "consistent" if consistent else "INCONSISTENT"
        print(
            f"{result} {label}: story={story_status} closure={closure_status} {consistency} "
            f"story_path={story_path.relative_to(ROOT)} closure_path={closure_path.relative_to(ROOT)}"
        )
    if failures:
        print(f"SEC-05 prerequisite check failed: {failures}/{len(CHECKS)} lifecycle records are not consistently accepted")
        return 1
    print(f"SEC-05 prerequisite check passed: {len(CHECKS)}/{len(CHECKS)} lifecycle records consistently accepted")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
