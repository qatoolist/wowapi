# golangci-lint backlog

golangci-lint was only wired into the enforced gate on 2026-07-04. Before that a `.golangci.yml` existed
but nothing ran the *full* set in CI, so a backlog of pre-existing issues accumulated. The gate now blocks
on **changed code only** (`make lint-new`), so this backlog does not block work — but it is tracked here
and burned down incrementally. See [quality-gates.md](quality-gates.md).

> **Scope note (2026-07-12).** "CLOSED" below refers only to the historical **B-1** burn-down of the
> *currently-enabled* linter set. It does **not** mean lint work is finished: the active lint programme
> is now **FBL-05/FBL-07** (enable `sqlclosecheck`/`rowserrcheck`/`bodyclose`/`noctx`/`gosec`/etc.) —
> see `docs/implementation/fable5-closure-depth-matrix-2026-07-11.md` CS-23. Do not read this file as
> a reason to skip those tasks.

## Status: CLOSED (2026-07-05) — `make lint` reports **0 issues** (D-0087)

The full advisory backlog (B-1) is burned down. Closed with **no behavior change** — full test suite green,
coverage 91.6% ≥ 90% floor.

- **149 `internal/cli` `fmt.Fprint*` issues → one path-scoped exclusion.** Every one was a best-effort write
  to a stdout/stderr `io.Writer` (os.Stdout/os.Stderr in prod, bytes.Buffer in tests) where a failed terminal
  write has no recovery — the canonical errcheck-exempt case (mirrors the stdlib's own `fmt.Print*` exclusion).
  A single scoped rule in `.golangci.yml` (`path: internal/cli/` + `source: fmt\.Fprint`) covers them, so
  genuine errcheck issues in that package (pool/file/exec errors) are still caught. This is the burn-down
  plan's sanctioned "scoped exclusion for a proven-safe stdlib call."
- **1 `internal/cli` `tw.Flush`** → explicit `_ =` with a one-line reason.
- **10 scattered issues → real code fixes**, each verified behavior-preserving: read-only `defer f.Close()`
  (`buildinfo`, `benchbudget`) → `defer func() { _ = f.Close() }()`; `Secret.Format` writes (a `fmt.Formatter`
  cannot return an error) → `_, _ =`; `unparam` (`webhook.truncate` dropped its always-500 `max`;
  `workflow.createTask` dropped its unused `def`; `testkit.newPoolDB` dropped its always-nil `opts`);
  `unused` (dead `type want` + `func run` in tests); `unconvert` in `benchbudget`.

**Snapshot before closure (2026-07-04):** ~160 — 154 `errcheck`, 3 `unused`, 2 `unparam`, 1 `unconvert`. An
earlier wiring pass had already cleared 31 `depguard` false positives, 10 `gofumpt`, `misspell`, and most
`staticcheck` quick-fixes.

**Keeping it closed:** `make lint-new` blocks any new/changed code that adds an issue, so the tree stays at 0
for normal development. Promoting `make lint` (full tree) into the enforced CI gate (step 3 below) is a
recommended follow-up — pin golangci-lint to a fixed version first, since CI installs `@latest` and a
full-tree enforced gate would otherwise break on any new upstream check.

## Burn-down plan

1. **errcheck** — tackle package-by-package. For each: check the error, or `_ =` with a one-line reason
   for genuinely best-effort calls (e.g. `defer rows.Close()`), or add a scoped `errcheck.exclude-functions`
   entry in `.golangci.yml` for a proven-safe stdlib call. Do NOT blanket-`//nolint`.
2. **unused / unparam / unconvert** — cheap; fix opportunistically when touching the file.
3. As categories reach zero, promote them from advisory to enforced on the *whole* tree (drop the
   `--new-from-merge-base` scoping for that linter, or fail `make lint`).

## Rule

Do not let the backlog grow: `make lint-new` (the gate) ensures new/changed code is fully clean, so this
number should only ever go **down**.
