# Graphify Project Graph

Keep the Graphify knowledge graph current so architecture, implementation, tests, and documentation remain navigable across sessions.

## Current Setup

- Graphify Codex integration is installed.
- Git hooks are installed (via `core.hooksPath = .githooks`):
  - `.githooks/pre-commit` (gitleaks secret scan, fmt/lint of staged Go changes, `code-review-graph update`)
  - `.githooks/pre-push` (gitleaks secret scan, vet, lint, tests, tidy check — no graphify step)
- The current graph covers more than 2,200 repository files and 5,000 symbols/concepts. Treat `graphify-out/graph.json` and `GRAPH_REPORT.md` as the authoritative current counts because they change with the corpus.
- Full semantic extraction uses Google Gemini by repository default and requires `GEMINI_API_KEY` or `GOOGLE_API_KEY`.

## First Full Graph

Set `GEMINI_API_KEY` (or `GOOGLE_API_KEY`), then run:

```bash
scripts/graphify_refresh.sh extract
```

The wrapper deliberately pins `--backend gemini`. For an intentional alternate backend, set `GRAPHIFY_BACKEND` and its corresponding key. Graphify itself supports keys including:

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
- Use `scripts/graphify_refresh.sh extract` after substantial documentation, architecture, or product-shaping changes so the semantic engine is explicit.
- Record the source commit and extraction backend/model in review or release evidence until Graphify persists backend/model provenance directly in graph metadata.
- Treat a reported semantic-chunk connection failure as an incomplete refresh and rerun it; an updated AST graph alone is not proof that changed documentation received semantic extraction.
- A completed semantic refresh must report each chunk as done and include nonzero input/output token accounting in the command evidence.
- Keep generated graph artifacts out of commits unless the team explicitly decides to version them.
- If Graphify reports that an update is needed, refresh before major design or implementation work.
