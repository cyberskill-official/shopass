# Project Structure

This repository is organized as a product monorepo. Keep product requirements, execution planning, and implementation code separate.

## Read Order

1. `README.md` - project landing page and current execution entry point.
2. `docs/feature-requests/SHIP-GUIDE.md` - non-negotiable build rules.
3. `docs/feature-requests/BACKLOG.md` - FR/NFR status board.
4. `specs/001-full-project-plan/README.md` - Spec Kit execution pack.
5. `specs/001-full-project-plan/tasks.md` - task board. The next unchecked task is the next implementation task.

## Top-Level Layout

```text
shopass/
  README.md
  AGENTS.md
  docs/
  specs/
  db/
  services/
  secrets/
  extension/
  web/
  mobile/
  ml/
  deploy/
  tests/
```

## What Belongs Where

| Path | Purpose | Commit? |
|------|---------|---------|
| `docs/` | Product docs, audited FR/NFR source docs, conventions, project structure docs | Yes |
| `docs/feature-requests/` | 90 Feature Requests, audits, backlog, implementation order, data model | Yes |
| `docs/non-functional-requirements/` | 10 NFR docs and audits | Yes |
| `specs/001-full-project-plan/` | Spec Kit implementation plan, task board, evidence, release gates | Yes |
| `db/` | Database migrations, seed data, migration tests | Yes |
| `services/` | Backend services by bounded module | Yes |
| `secrets/` | Shared secret provider package and tests | Yes |
| `extension/` | Chrome MV3 extension source, manifest, tests | Yes, except `node_modules/` |
| `web/` | Next.js web app | Yes |
| `mobile/` | Mobile app | Yes |
| `ml/` | ML jobs/models for deal scoring | Yes |
| `deploy/` | Local/dev/prod deployment assets | Yes |
| `tests/` | Cross-service contract, integration, e2e, compliance, performance tests | Yes |
| `.agents/` | Local Codex/agent skills | No |
| `.specify/` | Local Spec Kit CLI/runtime state | No |
| `.codex/` | Local Codex state | No |
| `node_modules/`, `dist/`, `build/`, `.next/`, `coverage/` | Generated outputs | No |

## Implementation Rules

- Do not put implementation code in `docs/` or `specs/`.
- Do not put planning/status files inside service folders.
- Every FR implementation must update its source FR status and `docs/feature-requests/BACKLOG.md`.
- Every review task must attach evidence under `specs/001-full-project-plan/evidence/`.
- Keep `.gitkeep` files out of commits; `.gitignore` already ignores them.

## Current Execution State

Layer 0 foundation is complete through `T021`.

Next implementation batch:

```text
T022-T031
```

Review checkpoint after that batch:

```text
T032
```
