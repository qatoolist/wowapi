#!/usr/bin/env python3
"""Regenerate impl/tracking/status-register.md from canonical front-matter statuses.

Canonical status lives in each wave.md / epic.md / story.md / task-*.md front matter
(mandate §6); this register is a derived roll-up. Run from the repo root:

    python3 miscellaneous/regen_status_register.py            # validate + regenerate
    python3 miscellaneous/regen_status_register.py --check    # validate + drift-check, no write (CI)

Validation (mandate §6/§7, status-model.md — added by the 2026-07-16 findings
remediation, preventing recurrence of vocabulary drift and false acceptance roll-ups):
  - every wave/epic/story/task front-matter status must be a permitted token for its level;
  - a story `accepted` or `verified` must have every task `done` or `cancelled`;
  - an epic `accepted` must have every story `accepted`, `deferred`, or `cancelled`;
  - a wave `accepted` must have every epic `accepted`, `deferred`, or `cancelled`;
  - a wave/epic `partially-accepted` must have at least one accepted child, at least one
    non-accepted/non-cancelled child (otherwise the token is wrong), and its own file text
    must cite the deviation/decision record (DEV-*/DEC-*) disposing of the exceptions.
Violations exit non-zero. Evidence/artifact records keep their own vocabulary (mandate
§10) and are deliberately NOT scanned here.

Introduced by the 2026-07-16 implementation-autopsy remediation (finding C-5: the
hand-maintained register had drifted four ways from canonical statuses).
"""

from __future__ import annotations

import datetime
import pathlib
import re
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent
WAVES = ROOT / "impl" / "waves"
OUT = ROOT / "impl" / "tracking" / "status-register.md"


def front_matter(path: pathlib.Path) -> dict[str, str]:
    text = path.read_text(encoding="utf-8")
    m = re.match(r"---\n(.*?)\n---", text, re.S)
    fields: dict[str, str] = {}
    if m:
        for line in m.group(1).splitlines():
            if ":" in line and not line.startswith(" "):
                k, _, v = line.partition(":")
                fields[k.strip()] = v.strip().strip('"')
    return fields


# Permitted vocabulary per level — status-model.md §7.1 (wave/epic), §7.2 (story), §7.3 (task).
WAVE_EPIC_STATUSES = {
    "proposed",
    "planned",
    "ready",
    "in-progress",
    "blocked",
    "verification",
    "accepted",
    "partially-accepted",
    "deferred",
    "cancelled",
}
STORY_STATUSES = {
    "draft",
    "planned",
    "ready",
    "in-progress",
    "implemented",
    "verification",
    "verified",
    "accepted",
    "blocked",
    "deferred",
    "cancelled",
}
TASK_STATUSES = {
    "todo",
    "ready",
    "in-progress",
    "blocked",
    "implemented",
    "verified",
    "done",
    "cancelled",
}
# A story claiming completion must have no unfinished tasks (status-model.md §7.2,
# definition-of-done.md "Story-level done").
STORY_COMPLETE = {"accepted", "verified"}
TASK_FINISHED = {"done", "cancelled"}
# A parent claiming full acceptance must have no child outside these tokens
# (status-model.md §7.1: `partially-accepted` exists precisely for the mixed case).
CHILD_ACCEPTED_OK = {"accepted", "deferred", "cancelled"}
# Deviation/decision citation required in a partially-accepted item's own text (mandate
# §1.2/§1.4: exceptions carry a recorded disposition, never a silent remainder).
DISPOSITION_REF = re.compile(r"\bDE[VC]-[A-Z0-9][A-Z0-9-]*\b")


def validate() -> list[str]:
    """Return a list of violations (empty = clean)."""
    violations: list[str] = []

    def check(path: pathlib.Path, level: str, allowed: set[str]) -> str:
        status = front_matter(path).get("status", "")
        if status not in allowed:
            violations.append(
                f"{path.relative_to(ROOT)}: {level} status `{status or '(missing)'}`"
                f" is not a permitted token ({'/'.join(sorted(allowed))})"
            )
        return status

    def check_rollup(
        path: pathlib.Path, level: str, status: str, children: list[str]
    ) -> None:
        """Enforce accepted/partially-accepted roll-up consistency for a wave or epic."""
        child_level = "epic" if level == "wave" else "story"
        exceptions = [
            c for c in children if c.split(":", 1)[1] not in CHILD_ACCEPTED_OK
        ]
        if status == "accepted" and exceptions:
            violations.append(
                f"{path.relative_to(ROOT)}: {level} is `accepted` but {len(exceptions)}"
                f" {child_level}(s) are not accepted/deferred/cancelled: {', '.join(exceptions)}"
            )
        if status == "partially-accepted":
            accepted = [c for c in children if c.split(":", 1)[1] == "accepted"]
            open_children = [
                c
                for c in children
                if c.split(":", 1)[1] not in {"accepted", "cancelled"}
            ]
            if not accepted:
                violations.append(
                    f"{path.relative_to(ROOT)}: {level} is `partially-accepted` but no"
                    f" {child_level} is accepted"
                )
            if not open_children:
                violations.append(
                    f"{path.relative_to(ROOT)}: {level} is `partially-accepted` but every"
                    f" {child_level} is accepted/cancelled — should be `accepted`"
                )
            if not DISPOSITION_REF.search(path.read_text(encoding="utf-8")):
                violations.append(
                    f"{path.relative_to(ROOT)}: {level} is `partially-accepted` but cites no"
                    " DEV-*/DEC-* record disposing of the non-accepted remainder"
                )

    for wave_dir in sorted(WAVES.iterdir()):
        wave_md = wave_dir / "wave.md"
        if not wave_md.is_file():
            continue
        wave_status = check(wave_md, "wave", WAVE_EPIC_STATUSES)
        epic_children: list[str] = []
        for epic_dir in sorted((wave_dir / "epics").glob("epic-*")):
            epic_md = epic_dir / "epic.md"
            epic_status = ""
            if epic_md.is_file():
                epic_status = check(epic_md, "epic", WAVE_EPIC_STATUSES)
                epic_children.append(f"{epic_dir.name}:{epic_status}")
            story_children: list[str] = []
            for story_dir in sorted(epic_dir.glob("stories/story-*")):
                story_md = story_dir / "story.md"
                if not story_md.is_file():
                    continue
                story_status = check(story_md, "story", STORY_STATUSES)
                unfinished = []
                for task_md in sorted(story_dir.glob("tasks/task-*.md")):
                    task_status = check(task_md, "task", TASK_STATUSES)
                    if task_status not in TASK_FINISHED:
                        unfinished.append(f"{task_md.name}:{task_status}")
                if story_status in STORY_COMPLETE and unfinished:
                    violations.append(
                        f"{story_md.relative_to(ROOT)}: story is `{story_status}` but"
                        f" {len(unfinished)} task(s) are not done/cancelled: {', '.join(unfinished)}"
                    )
                story_children.append(f"{story_dir.name}:{story_status}")
            if epic_md.is_file():
                check_rollup(epic_md, "epic", epic_status, story_children)
        check_rollup(wave_md, "wave", wave_status, epic_children)
    return violations


def main() -> int:
    check_only = "--check" in sys.argv
    violations = validate()
    for v in violations:
        print(f"VIOLATION: {v}", file=sys.stderr)

    today = datetime.date.today().isoformat()
    lines = [
        "---",
        "id: TRACK-STATUS-REGISTER",
        "type: register",
        "title: Status register — derived roll-up of canonical front-matter statuses",
        "status: active",
        "created_at: 2026-07-12",
        f"updated_at: {today}",
        "derived: true",
        "generated_by: miscellaneous/regen_status_register.py",
        "---",
        "",
        "# Status register",
        "",
        f"**DERIVED VIEW — generated {today} by `miscellaneous/regen_status_register.py`.**",
        "Canonical status lives in each wave/epic/story's own front matter (mandate §6). Do not",
        "hand-edit this file; regenerate it after any canonical status change.",
        "",
    ]
    counts: dict[str, int] = {}
    for wave_dir in sorted(WAVES.iterdir()):
        wave_md = wave_dir / "wave.md"
        if not wave_md.is_file():
            continue
        wf = front_matter(wave_md)
        lines += [f"## {wave_dir.name} — wave status: `{wf.get('status', '?')}`", ""]
        lines += ["| Item | Level | Title | Status |", "|---|---|---|---|"]
        for epic_dir in sorted((wave_dir / "epics").glob("epic-*")):
            ef = front_matter(epic_dir / "epic.md")
            lines.append(
                f"| {ef.get('id', epic_dir.name)} | epic | {ef.get('title', epic_dir.name)}"
                f" | {ef.get('status', '?')} |"
            )
            for story_dir in sorted(epic_dir.glob("stories/story-*")):
                sf = front_matter(story_dir / "story.md")
                status = sf.get("status", "?")
                counts[status] = counts.get(status, 0) + 1
                lines.append(
                    f"| {sf.get('id', story_dir.name)} | story | {sf.get('title', story_dir.name)}"
                    f" | {status} |"
                )
        lines.append("")
    lines += ["## Story status totals", ""]
    for status in sorted(counts):
        lines.append(f"- `{status}`: {counts[status]}")
    lines.append("")
    content = "\n".join(lines)

    def stable(text: str) -> str:
        # Ignore the two date-bearing lines so --check doesn't fail on regeneration date alone.
        return "\n".join(
            ln
            for ln in text.splitlines()
            if not ln.startswith("updated_at:")
            and not ln.startswith("**DERIVED VIEW — generated")
        )

    if check_only:
        existing = OUT.read_text(encoding="utf-8") if OUT.is_file() else ""
        drifted = stable(existing) != stable(content)
        if drifted:
            print(
                f"DRIFT: {OUT.relative_to(ROOT)} does not match canonical front matter —"
                " run miscellaneous/regen_status_register.py and commit the result",
                file=sys.stderr,
            )
        if violations or drifted:
            return 1
        print(f"status register check OK ({sum(counts.values())} stories: {counts})")
        return 0

    OUT.write_text(content, encoding="utf-8")
    print(f"wrote {OUT} ({sum(counts.values())} stories: {counts})")
    return 1 if violations else 0


if __name__ == "__main__":
    sys.exit(main())
