---
id: DEV-W06-E04-S002
type: deviations-record
parent_story: W06-E04-S002
status: recorded
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations record — W06-E04-S002

## DEV-W06-E04-S002-001 — T4 began before W05 lifecycle bookkeeping reached accepted

- **Approved plan:** W06-E04-S002-T001 begins only after W05-E03 reaches lifecycle status `accepted`.
- **Actual implementation:** The shared working tree contained W05 AR-03's
  `kernel/appmodel.GenerateProjections`, including its `Doc` export and AR-03 golden tests. The W05
  owner explicitly confirmed that function as the byte-authoritative export and confirmed the tests
  passed, but also confirmed W05's story/task records would remain draft/todo this session. Under the
  user's explicit direction to implement T4 when the exact export prerequisite is available,
  W06-E04-S002-T001 consumed the live export and added an independent byte-equality test.
- **Reason:** The technical prerequisite was genuinely present; only upstream lifecycle bookkeeping
  lagged. Treating T4 as absent/blocked would have contradicted the user's available-export rule.
- **Impact:** No production runtime impact. Traceability cannot claim that W05-E03's lifecycle record
  was `accepted` at execution time; it can and does prove the exact delivered export bytes consumed.
- **Risks:** W05 may revise its export before its own acceptance. `make docs-check` will then fail until
  the generated table is intentionally regenerated, preserving safety.
- **Approval:** User instruction authorized implementation when the exact export is available;
  W06-E01-E04-Execution.W06E04ReviewR independently accepted the recorded deviation with no issues.
- **Compensating controls:** Direct import of AR-03's export function, byte-for-byte golden comparison,
  generated-file currency check, CI wiring, and explicit revision/evidence records.
- **Follow-up work:** W05 owner/conductor updates W05-E03 lifecycle records; no W06 code change is
  required unless the accepted export bytes change.
