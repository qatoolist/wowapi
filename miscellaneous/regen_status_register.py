#!/usr/bin/env python3
"""Regenerate impl/tracking/status-register.md from canonical front-matter statuses.

Canonical status lives in each wave.md / epic.md / story.md front matter (mandate §6);
this register is a derived roll-up. Run from the repo root:

    python3 miscellaneous/regen_status_register.py

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


def main() -> int:
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
    OUT.write_text("\n".join(lines), encoding="utf-8")
    print(f"wrote {OUT} ({sum(counts.values())} stories: {counts})")
    return 0


if __name__ == "__main__":
    sys.exit(main())
