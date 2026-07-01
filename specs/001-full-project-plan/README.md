# SănDeal Full Project Plan

This directory is the Spec Kit execution pack for building the full SănDeal project.

Use it as the control board for implementation and review. Product requirements remain in `docs/feature-requests/` and `docs/non-functional-requirements/`.

## File Map

| File | Purpose |
|------|---------|
| `spec.md` | High-level feature specification for the full project build |
| `plan.md` | Architecture, phases, dependency layers, and monorepo layout |
| `tasks.md` | Ordered implementation task board. Do the next unchecked task first |
| `quickstart.md` | Release gate scenarios for P0/P1/P2/P3/final |
| `research.md` | Planning research and technical decisions |
| `data-model.md` | Spec-level data model summary |
| `contracts/release-contract.md` | Release acceptance contract |
| `constitution-gates.md` | SHIP-GUIDE invariant gates |
| `source-of-truth.md` | Rule for FR/audit docs as authoritative source |
| `status-procedure.md` | Status transition procedure |
| `schema-ownership.md` | One-table-one-owner schema rule |
| `fr-done-checklist.md` | Per-FR completion checklist |
| `execution-ledger.md` | Task execution/review ledger |
| `evidence-template.md` | Template for evidence bundles |
| `evidence/` | Phase evidence bundles |

## Execution Rule

1. Open `tasks.md`.
2. Find the first unchecked task.
3. Read the task's FR/NFR doc and `.audit.md`.
4. Implement only that task's required surface.
5. Run the tests required by the source doc's §5.
6. Update the FR/NFR status and `docs/feature-requests/BACKLOG.md`.
7. Record evidence in the matching `evidence/*.md`.
8. Mark the task `[x]` only after code, tests, status, backlog, and evidence are complete.

## Current State

Completed:

```text
T001-T021
```

Next batch:

```text
T022-T031
```

Next review checkpoint:

```text
T032
```

## Boundaries

- `specs/` contains planning, task tracking, review evidence, and release notes.
- `docs/` contains product requirements and build conventions.
- `services/`, `db/`, `secrets/`, `extension/`, `web/`, `mobile/`, `ml/`, `deploy/`, and `tests/` contain implementation.
- `.agents/`, `.specify/`, and local dependency directories are not project deliverables.
