# Evidence Bundles

Active evidence lives here; **historical bundles are archived** (2026-07-11) in the `wowapi2`
documentation archive under `archive/evidence/` — see its `ARCHIVE-INDEX.md` for the full map:

- `phase-00/` … `phase-12/` — Goal-2 build-phase exit bundles (all phases closed).
- `hardening-H1/` … `hardening-H5/`, `hardening-P1/` — hardening-tranche proof bundles (closed).
- `wowsociety-gaps/` — GAP-001..008 proof bundle (closed).

References to those paths from historical entries in `../decisions.md`, `phase-plan.md` (itself
archived), and the CHANGELOG resolve via the archive at the same relative names.

Still active here:

- `architecture-review-2026-07-11/` — evidence for the in-flight Fable 5 architecture-review
  programme (`../fable5-final-architecture-review-2026-07-11.md`).

Required files per bundle (Goal 2 §Evidence Gate; applies to any future bundle added here):

- `proof-bundle.md` — decisions (links into ../decisions.md), discussions/agent debates,
  implementation inventory (files/packages/tests added), acceptance checklist status.
- `review-findings.md` — critique/review output: finding, severity, file:line, resolution
  (fixed→commit/patch, or rejected→reason). No finding may be silently dropped.
- `command-log.md` — exact commands, exit codes, summarized output. Commands that could not run
  are listed with the reason and the residual risk.
- `acceptance-map.md` — phase acceptance criteria → code/test/command evidence.

"Reviewed/tested/verified" claims without an entry here are treated as not done.
