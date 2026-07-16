# SEC-05 — Versioned Security Verification Profile

## What is this directory?

SEC-05 is a quality-gate requirement for a **versioned security verification profile** — a machine-readable map of which security controls are tested, how thoroughly, and with what evidence. The directory contains:

- **Control map** (`control-map.md`, `control-map.json`): authoritative registry of security controls and their verification status
- **Python validation tooling** (`validate_control_map.py`, `test_validate_control_map.py`): ensures the control map is internally consistent and complete
- **Prerequisite verification** (`verify_prerequisites.py`): checks that the framework and test environment meet security baseline requirements
- **Evidence sources** (`sources/`): supporting documentation and test output references
- **External assessment status** (`external-assessment-status.md`): tracking of professional security review engagement

The control map and validation scripts are run as part of the final verification gate (Wave 07, Epic 02) to prove that security coverage is complete and evidence is sound.

## Which requirements/stories reference this?

- **Wave 07, Epic 02, Story 001** (`W07-E02-S001`): the plan for executing this requirement (currently blocked pending professional assessment)
- **Wave 07, Epic 02** general verification hardening and coverage truthfulness
- Cross-references in the implementation ledger: `impl/index.md` (wave allocation), `impl/analysis/requirement-inventory.md` (SEC-05 entry)

## Why does SEC-05 live at the repo root?

SEC-05 evidence bundles and control-map pinning (to specific commits) are recorded in `impl/waves/wave-07.../epic-002/.../evidence/index.md`. **Relocating this directory would break all pinned evidence references** in the implementation programme.

Relocation is a tracked future refactoring. It requires:
1. Update all SEC-05 control-map and evidence paths in `impl/waves/wave-07.../`
2. Update external-assessment-status references if any exist in other waves
3. Re-run and re-pin evidence bundles against the new location

## Testing the security verification profile

Run the control-map validator:
```bash
cd SEC-05
python3 validate_control_map.py
python3 -m pytest test_validate_control_map.py -v
```

Run prerequisite checks:
```bash
python3 verify_prerequisites.py
```

The full control map is in `control-map.json` (machine-readable) and `control-map.md` (human-readable). Both are kept in sync.

## Note on Python tooling

The framework is primarily Go; SEC-05's Python tooling is an exception for security control mapping (JSON schema validation, audit control matrix enumeration). The `.ruff_cache` directory in this tree is Python linting cache and should be `.gitignore`d at the repo root.
