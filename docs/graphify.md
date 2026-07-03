# Graphify Project Graph

This repository is expected to grow as implementation starts. Keep a Graphify knowledge graph current so architecture, docs, and later code remain navigable across sessions.

## Current Setup

- Graphify Codex integration is installed.
- Git hooks are installed:
  - `.git/hooks/post-commit`
  - `.git/hooks/post-checkout`
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

For doc/architecture changes that need semantic extraction:

```bash
scripts/graphify_refresh.sh extract
```

For reclustering an existing graph:

```bash
scripts/graphify_refresh.sh cluster
```

## Policy

- Do not hand-edit `graphify-out/graph.json`.
- Prefer `graphify update .` for code-only changes.
- Use full `graphify extract . --out .` after substantial documentation, architecture, or product-shaping changes.
- Keep generated graph artifacts out of commits unless the team explicitly decides to version them.
- If Graphify reports that an update is needed, refresh before major design or implementation work.
