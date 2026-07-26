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
