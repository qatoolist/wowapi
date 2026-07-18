# Graphify extraction provenance

This file records the semantic-extraction evidence for the clean-V1 review. Generated files under `graphify-out/` remain local and ignored; this small record is the auditable repository artifact.

## Clean-V1 semantic run

- Implementation revision: `8b2188f29a2d96536004b6bcf619c5cac7b8a1dd`; the final report/provenance revision is repository `HEAD` at handoff.
- Graphify: `0.7.19`.
- Invocation: `scripts/graphify_refresh.sh extract`, followed by `scripts/graphify_refresh.sh cluster`.
- Backend: Google Gemini, pinned by the repository wrapper.
- Model: `gemini-3-flash-preview`, Graphify 0.7.19's Gemini default; no `GRAPHIFY_MODEL` override was supplied.
- Semantic scope: 46 deliberately invalidated changed documents/configuration files, split into three chunks.
- Completion: 3/3 semantic chunks succeeded after rerunning outside the network-restricted sandbox.
- Raw usage: 146,703 input tokens and 9,786 output tokens in `graphify-out/.graphify_analysis.json`.
- Result before the final incremental refresh: 8,489 nodes, 21,483 edges, and 580 communities.

Graphify's cluster/report regeneration currently writes zero token usage into `graphify-out/GRAPH_REPORT.md`, even though the raw analysis file contains the nonzero totals above. The raw analysis is therefore the usage authority; the generated report is the node/edge/community authority. Seven warnings about missing `source_file` provenance remain on historical `impl/` inferred edges. They were not hand-edited because generated graph data is not a source artifact.

The final handoff records the post-report incremental counts and confirms that structural freshness is current. Semantic extraction is considered complete only when every requested chunk succeeds and raw input/output usage is nonzero.

## Final report refresh

After `ComprehensiveReport.md` received Fable's final verdict, committed revision `bd650be` was structurally refreshed and the report plus this provenance document were deliberately invalidated from the semantic cache. Google Gemini re-extracted 2/2 files in one successful chunk using 5,258 input and 2,456 output tokens. Reclustering produced 8,541 nodes, 21,652 edges, and 599 communities. The same seven historical `impl/` edge warnings remained; no semantic chunk failed.
