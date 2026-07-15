---
id: W05-EPICS-INDEX
type: epics-index
wave: W05
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05 — Epics index

| Epic | Title | Status | Stories | Objective |
|---|---|---|---|---|
| [W05-E01](epic-001-application-model/epic.md) | application-model | planned | 4 | Build the ownership-bound, immutable `ApplicationModel` from zero: lifecycle skeleton, owner-bound `Registrar` capability type, per-registry ownership wrappers, snapshot immutability, deterministic model hash, race safety, and the legacy compatibility adapter (AR-01) |
| [W05-E02](epic-002-typed-ports/epic.md) | typed-ports | planned | 3 | Build the typed port-key API and compiled provider graph, boot-time validated and free of hot-path reflection, retiring the hand-maintained lifecycle manifest (AR-02) |
| [W05-E03](epic-003-authoritative-declarations/epic.md) | authoritative-declarations | planned | 2 | Establish the single authoritative module manifest with derived projections proven by a golden-delta gate, and close the remaining boot-time silent-behaviour gaps with a shared waiver mechanism (AR-03 + AR-04 remainder) |
| [W05-E04](epic-004-wiring-and-cache-hygiene/epic.md) | wiring-and-cache-hygiene | planned | 2 | Close the remaining kernel constructor-bypass surface and bound the authorization cache with epoch-based cross-pod invalidation (AR-06 remainder + SEC-04) |
| [W05-E05](epic-005-kernel-re-home/epic.md) | kernel-re-home | planned | 2 | Re-home the nine misplaced app-foundation/adapter packages out of `kernel/` to `foundation/`, with a deprecated forwarding shim for `kernel/mfa` and extended layering enforcement (FBL-01) |
