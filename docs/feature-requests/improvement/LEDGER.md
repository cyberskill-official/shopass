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
