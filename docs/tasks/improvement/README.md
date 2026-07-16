# docs/improvement - strengthening program for SănDeal (shopass)

Source: `docs/strategy/shopass-strengthening-audit-2026-07-06.md` (the audit). Every task here maps 1:1 to a recommendation R1-R58 from that report and keeps the same number. This folder is the working surface: agents implement from it, humans review from it.

## Files

- `BACKLOG.md` - index of all 58 tasks: wave, effort, dependencies, Stephen-input flag, live status. The single source of truth for status.
- `TASKS-A-blockers.md` - R1-R10 detailed cards (ship blockers).
- `TASKS-B-operations.md` - R11-R22 (production operations).
- `TASKS-C-integrations.md` - R23-R31 (stubs to real: notifications, scraping, ML, payments, affiliate).
- `TASKS-D-compliance-trust.md` - R32-R37 (PDPL, policies, transparency, open source).
- `TASKS-E-web-growth.md` - R38-R48 (landing, SEO, conversion, lead capture).
- `TASKS-F-distribution.md` - R49-R58 (stores, channels, B2B leads, launch).
- `KICKOFF-PROMPT.md` - the prompt to paste into an implementation agent. Contains the full working protocol and the human review protocol.
- `LEDGER.md` - append-only evidence log written by the implementing agent, read by the human reviewer.

## Status enum

`todo` -> `in_progress` -> `awaiting_review` -> `done`. Side states: `blocked` (dependency not met), `needs_stephen` (requires a decision, account, credential, budget, or outreach only Stephen can provide - the precise ask must be written in the ledger), `dropped`.

Only a human flips a task to `done`. Agents stop at `awaiting_review`.

## Waves

- Wave 1 (days 0-30, unblock + harden): R1-R17, R23, R34, R40, R49.
- Wave 2 (days 31-60, real data + first users): R18, R19, R24-R27, R30, R31, R35, R36, R38, R39, R41, R43, R45, R46, R55.
- Wave 3 (days 61-90, launch + monetize): R20-R22, R28, R29, R32, R33, R37, R42, R44, R47, R48, R50-R54, R56-R58.

Order inside a wave: lowest task number first unless dependencies say otherwise. A task whose Stephen-input is pending does not block the next task - the agent records the ask and moves on.

## Non-negotiables (inherited)

The build invariants in `docs/tasks/SHIP-GUIDE.md` still apply to all code produced here (stack choices, testing discipline, no secrets in code, parameterized SQL, consent-first extension rules). Where a card here conflicts with SHIP-GUIDE, SHIP-GUIDE wins and the conflict gets logged in the ledger.
