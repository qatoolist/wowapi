# Graphify Project Graph

This repository is expected to grow as implementation starts. Keep a Graphify knowledge graph current so architecture, docs, and later code remain navigable across sessions.

## Current Setup

- Graphify Codex integration is installed.
- Git hooks are installed (via `core.hooksPath = .githooks`):
  - `.githooks/pre-commit` (gitleaks secret scan, fmt/lint of staged Go changes, `code-review-graph update`)
  - `.githooks/pre-push` (gitleaks secret scan, vet, lint, tests, tidy check — no graphify step)
- The current corpus is documentation-only: 17 supported document files, about 35k words.
- Full semantic extraction currently requires an LLM backend key for Graphify.

## First Full Graph

Set one supported Graphify backend key, then run:

```bash
graphify extract . --out . --max-concurrency 2
```

Supported keys include:

- `GEMINI_API_KEY` or `GOOGLE_API_KEY`
- `MOONSHOT_API_KEY`
- `ANTHROPIC_API_KEY`
- `OPENAI_API_KEY`

Expected outputs:

```text
graphify-out/graph.html
graphify-out/GRAPH_REPORT.md
graphify-out/graph.json
```

## Regular Updates

For normal code changes:

```bash
scripts/graphify_refresh.sh update
```

To export the current graph into the local Neo4j container:

```bash
make graph-neo4j
```

For a bridge map over the current graph:

```bash
sh scripts/graphify_bridge_map.sh node
sh scripts/graphify_bridge_map.sh community
```

For doc/architecture changes that need semantic extraction:

```bash
scripts/graphify_refresh.sh extract
```

For reclustering an existing graph:

```bash
scripts/graphify_refresh.sh cluster
```

For graph database export and deeper cross-community querying:

```bash
graphify export neo4j
```

## Policy

- Do not hand-edit `graphify-out/graph.json`.
- Prefer `graphify update .` for code-only changes.
- Use full `graphify extract . --out .` after substantial documentation, architecture, or product-shaping changes.
- Keep generated graph artifacts out of commits unless the team explicitly decides to version them.
- If Graphify reports that an update is needed, refresh before major design or implementation work.
