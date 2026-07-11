# wowapi — Codex Workflow

This repository is a domain-neutral Go platform kernel; `wowsociety` is the product built on it.
This file is the Codex-facing entrypoint for working here. Keep it aligned with the real codebase and
the living working docs in `docs/working/`.

## How to work here

1. Read the relevant working docs first:
   - [`docs/working/README.md`](docs/working/README.md)
   - [`docs/working/skills-and-knowledge-map.md`](docs/working/skills-and-knowledge-map.md)
   - [`docs/working/best-practices.md`](docs/working/best-practices.md)
   - [`docs/working/quality-gate-checklist.md`](docs/working/quality-gate-checklist.md)
2. Inspect existing code before changing anything. Prefer `rg`, `rg --files`, `git show`, and the graph
   outputs over assumptions.
3. Reuse the repo’s existing patterns, packages, and tests. Do not invent a parallel style when one already exists.
4. Wire changes end to end. A kernel primitive without its adapter, boot wiring, and tests is not done.
5. Record load-bearing decisions in [`docs/implementation/decisions.md`](docs/implementation/decisions.md).
6. Finish with tests and the review gate. Do not mark a goal done until the change has been verified.

## Tool routing

Use the cheapest tool that answers the question.

| Task | Preferred tool |
|---|---|
| Find a symbol, package, or usage | `rg` / `rg --files` |
| Understand cross-file structure | Graphify outputs in `graphify-out/` |
| Inspect central bridges | [`scripts/graphify_bridge_map.sh`](scripts/graphify_bridge_map.sh) |
| Load graph into Neo4j | `make graph-neo4j` |
| Review blast radius | `graphify-out/GRAPH_REPORT.md` + `graphify-out/graph.json` |
| Compare recent changes | `git diff`, `git log`, `git show` |
| Run checks | `make check`, `make ci`, `make ci-container` |

## Graph and review discipline

- Keep `graphify-out/` current when architecture or code shape changes.
- Use `make graph-update` for normal code changes.
- Use `graphify refresh`/extract only when the change is large enough to justify semantic rebuilding.
- Use `make graph-neo4j` when you want interactive queries over the current graph.
- Prefer the bridge-map script for a quick cross-community view before deeper graph work.

## Review expectations

- Review the changed code, not just the diff summary.
- Verify claims with source, tests, and the current graph.
- Call out missing wiring, missing tests, or implicit assumptions.
- Distinguish done from deferred. “Planned later” is not the same as complete.

## Working constraints

- Keep changes ASCII unless the file already uses another character set.
- Use `apply_patch` for file edits.
- Do not revert user changes you did not make.
- Prefer small, scoped changes over broad refactors.

## If the task is architecture-related

Start with:
1. [`graphify-out/GRAPH_REPORT.md`](graphify-out/GRAPH_REPORT.md)
2. [`docs/implementation/premier-framework-implementation-plan.md`](docs/implementation/premier-framework-implementation-plan.md) §5 (task backlog) + [`docs/implementation/fable5-closure-depth-matrix-2026-07-11.md`](docs/implementation/fable5-closure-depth-matrix-2026-07-11.md) (closure specs)
3. the relevant package source

Then inspect the shortest path through the graph, not the whole tree.
