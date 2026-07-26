# Improvement ledger (append-only)

Implementing agents append one entry per task state change here. Never edit or delete prior entries. Humans add review entries the same way.

Entry format:

```
## [date] R<n> - <event: started | evidence | awaiting_review | needs_stephen | review_pass | review_fail>
- agent/human: <who>
- branch/commit: <branch> @ <sha>
- evidence: <commands run + key output lines, links, file paths>
- stephen_ask: <exact ask, only for needs_stephen>
- notes: <decisions made, deviations from the card, follow-ups discovered>
```

Follow-ups discovered mid-task get their own line here and a new row in `BACKLOG.md` (id style: R<n>.1) rather than silently widening the task.

---

## [2026-07-26] staging - review_pass (Paths 1–2 deferred-externals DoD)
- agent/human: operator (via implement next-steps plan) / Auto
- branch/commit: main @ 4f792c2 (+ local staging helpers pending PR)
- evidence: `make simulate-prices` wrote 199000→179000→149000; free user bottom_predicted 402; `make grant-premium USER_ID=5252` → active plan_id=2; retry 201. FCM/proxy/real pay deferred per `docs/STAGING-MONETIZED-SLICE.md`.
- stephen_ask: -
- notes: Unlocks Wave-1 hardening pickup (R1). Does not mark product tasks done.

## [2026-07-26] R1 - started (converted to TASK-INFRA-006)
- agent/human: Auto
- branch/commit: hardening/infra-006-wire-gateway-compose
- evidence: Created `docs/tasks/infra/TASK-INFRA-006-wire-gateway-compose/spec.md` (status: implementing); wired redis+gateway into `deploy/docker-compose.yml`; unpublished pricesvc/dealsvc/authsvc/tracksvc/billsvc/bff host ports; web gets `GATEWAY_INTERNAL_BASE_URL`.
- stephen_ask: -
- notes: Halt at reviewing→ready_to_test and testing→done for HITL; do not self-set done.

## [2026-07-26] R1 - evidence (ready_to_review)
- agent/human: Auto
- branch/commit: hardening/infra-006-wire-gateway-compose
- evidence: Host ports only db/gateway/web; `curl -H 'X-User-Id: 1' http://127.0.0.1:8080/v1/alerts` → 401; tracksvc:8083 unreachable; `make simulate-prices` OK via compose network; `go test ./services/gateway/...` ok.
- stephen_ask: HITL review accept for reviewing → ready_to_test (do not mark done yet).
- notes: Status moved implementing → ready_to_review.

## [2026-07-26] R1 - review_pass + done (TASK-INFRA-006)
- agent/human: operator (“Approve, merge then continue”) / Auto
- branch/commit: main @ fd708c6 (#57 squash); helpers #56 @ c2d2dcb
- evidence: PRs merged; CI green on #57; forged X-User-Id 401 verified pre-merge.
- stephen_ask: -
- notes: HITL covered review + final accept. Task frontmatter → done.

## [2026-07-26] R2 - review_pass + done (already satisfied)
- agent/human: Auto
- branch/commit: main @ fd708c6
- evidence: `git ls-files | grep -c node_modules` → 0; root `.gitignore` has `node_modules/` and `**/.venv/`.
- stephen_ask: -
- notes: No purge commit needed; history may still contain old blobs (BFG optional before R36).

## [2026-07-26] R4 - review_pass + done (already satisfied)
- agent/human: Auto
- branch/commit: main @ fd708c6
- evidence: `web/app/layout.tsx` title/description Shopass (not Create Next App); `web/test/root-metadata.test.ts` asserts that.
- stephen_ask: -
- notes: Reconcile-only status update.

## [2026-07-26] R5 - review_pass + done (already satisfied)
- agent/human: Auto
- branch/commit: main @ fd708c6
- evidence: `web/middleware.ts` protects /dashboard,/wishlist,/alerts,/billing,/products,/capture*; `web/test/middleware-guard.test.ts`.
- stephen_ask: -
- notes: Reconcile-only status update.

## [2026-07-26] R6 - started
- agent/human: Auto
- branch/commit: hardening/wave1-reconcile-and-r6
- evidence: Origin checks + SameSite cookies + gateway login rate-limit already present; closing gaps (refresh limit + conventions doc).
- stephen_ask: -
- notes: -

## [2026-07-26] R6 - evidence (ready_to_review)
- agent/human: Auto
- branch/commit: hardening/wave1-reconcile-and-r6
- evidence: Added `docs/conventions/AUTH-ORIGIN-CSRF.md` (origin-check strategy); gateway `/v1/auth/refresh` limit=10 + tests; login remains 5/min. Pre-existing: web origin checks on login/register/refresh/logout; SameSite=Strict cookies.
- stephen_ask: HITL review accept → done.
- notes: -

## [2026-07-26] R6 - review_pass + done
- agent/human: operator (continue after approve/merge) / Auto
- branch/commit: main (PR #58)
- evidence: AUTH-ORIGIN-CSRF.md; refresh rate-limit tests; CI green on #58.
- stephen_ask: -
- notes: Wave 1 A-blockers R1–R6 closed except R3 (needs Stephen license decision).
