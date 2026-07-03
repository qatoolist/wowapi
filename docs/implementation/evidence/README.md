# Evidence Bundles

One directory per phase (`phase-00/`, `phase-01/`, …). Required files per Goal 2 §Evidence Gate:

- `proof-bundle.md` — decisions (links into ../decisions.md), discussions/agent debates,
  implementation inventory (files/packages/tests added), acceptance checklist status.
- `review-findings.md` — critique/review output: finding, severity, file:line, resolution
  (fixed→commit/patch, or rejected→reason). No finding may be silently dropped.
- `command-log.md` — exact commands, exit codes, summarized output. Commands that could not run
  are listed with the reason and the residual risk.
- `acceptance-map.md` — phase acceptance criteria → code/test/command evidence.

"Reviewed/tested/verified" claims without an entry here are treated as not done.
