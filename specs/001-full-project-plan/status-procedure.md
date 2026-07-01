# Status Transition Procedure — SănDeal

Derived from `docs/feature-requests/STATUS-REFERENCE.md`.

## Status Enum (10 states)

All lowercase snake_case, linear lifecycle.

| # | Status | Meaning | Default Author |
|---|--------|---------|---------------|
| 1 | `draft` | Spec being written; not yet audited | FR author |
| 2 | `ready_to_implement` | Audited 10/10; ready for build queue. Also the fallback when implementing/reviewing/testing fails | Audit / rework |
| 3 | `implementing` | Code being written; tests partially present | Workflow ship |
| 4 | `ready_to_review` | Code + tests complete; awaiting reviewer | Workflow ship |
| 5 | `reviewing` | Reviewer reading diff against §1 + §4 AC matrix | Workflow ship |
| 6 | `ready_to_test` | Reviewer approved; awaiting tester | Workflow ship |
| 7 | `testing` | Tester running coverage gate (§1 -> §4 -> §5) | Workflow ship |
| 8 | `done` | All §1 statements trace to passing tests; terminal success | Workflow ship |

### Off-Ramp States

| # | Status | Meaning |
|---|--------|---------|
| 9 | `on_hold` | Deliberately paused; revisit later. Skipped by build queue. |
| 10 | `closed` | Terminal kill — won't build (rejected, superseded, duplicate, wontfix). |

## Transition Rules

- `[FAILED: ...]` and `[BLOCKED: ...]` are routing decisions, not states.
- Fail circuit-breaker during `implementing` (e.g., 5 consecutive test failures) → status drops to `ready_to_implement`.
- Non-fatal blocker found during `reviewing`/`testing` (ambiguous spec, missing dependency) → status drops to `ready_to_implement`.
- Reason recorded in a comment on BACKLOG or an external issue.

## HITL (Human in the Loop) — Optional

Workflow ship auto-advances status through §1.1 when each gate passes. Operator may override any status to any other at any time. Common overrides:

- Re-audit done FR: `done` → `ready_to_review`
- Skip review for trivial FR: `ready_to_review` → `ready_to_test`
- Park in-progress FR: `implementing` → `on_hold`
- Revive closed FR: `closed` → `ready_to_implement`

## Initial State

All 90 FRs in the SănDeal backlog start at `ready_to_implement` (audited 10/10, ready for build queue).

## How an Agent Ships a FR

1. Read IMPLEMENTATION-ORDER.md; pick lowest-layer FR where all `depends_on` are `done`.
2. Within same layer: MUST > SHOULD > COULD priority.
3. Flip FR status to `implementing`.
4. Build per `new_files`/`sub_tasks`; run §5 tests; verify §1 -> §4.
5. Advance through lifecycle to `done`; update BACKLOG status.
