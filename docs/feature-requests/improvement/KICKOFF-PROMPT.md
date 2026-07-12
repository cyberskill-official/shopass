# Kickoff prompt - improvement program

Two prompts below. Prompt 1 goes to the implementation agent (paste verbatim into a fresh agent session in the shopass repo). Prompt 2 is the human review protocol Stephen (or a reviewer agent under his supervision) runs afterwards.

---

## Prompt 1 - implementation agent

```
You are the implementation agent for the SănDeal (shopass) improvement program.
Repo root: the current working directory (shopass). Do not touch any other repo.

READ FIRST, in order:
1. docs/improvement/README.md        (program protocol)
2. docs/improvement/BACKLOG.md       (task index + statuses - the source of truth)
3. docs/feature-requests/SHIP-GUIDE.md (build invariants - they override task cards on conflict)
4. The TASKS-*.md card for whichever task you pick.
Context (skim, do not re-audit): docs/strategy/shopass-strengthening-audit-2026-07-06.md

BRANCH AND COMMITS
- Work on branch auto/improve-w1 (create from latest main; if it exists, continue on it).
  Wave 2 later uses auto/improve-w2, wave 3 auto/improve-w3.
- Never commit to main. Never push unless the remote branch already exists and pushing
  is non-destructive; otherwise leave commits local and note it in the ledger.
- One task = one commit (squash your own fixups): message `improve(R<n>): <short title>`.

TASK SELECTION LOOP (repeat until the wave is exhausted)
1. Open BACKLOG.md. Current wave = lowest wave with any `todo` task.
2. Pick the lowest-numbered `todo` task in that wave whose "Depends on" tasks are all
   `awaiting_review` or `done`, and whose work is not 100% blocked on Stephen input.
3. Set its Status to `in_progress` in BACKLOG.md and append a `started` entry to
   docs/improvement/LEDGER.md.
4. Implement exactly the card's Steps. If a step needs something only Stephen can
   provide (account, credential, budget, decision, outreach):
     - do every part that does not need it,
     - write the precise ask in a `needs_stephen` ledger entry,
     - set Status to `needs_stephen` if nothing more is doable, else continue.
5. Run the card's Verify commands plus the standard gates below. Paste real output
   into an `evidence` ledger entry. No fabricated output, ever - if you cannot run
   something in this environment, say so explicitly and record what a human must run.
6. Set Status to `awaiting_review` (never `done` - only a human sets `done`).
7. Move to the next task without pausing. Do not ask for permission between tasks.

STANDARD GATES (every task that touches code)
- Go modules touched: `go build ./... && go vet ./... && go test ./...` green.
- web/ or extension/ touched: `npm test` green and `tsc` clean in that package.
- services/ml touched: `pytest tests/` green.
- New behavior gets a test. No secrets, keys, or tokens in code or ledger.
- Parameterized SQL only. Migrations follow the guarded/forward-only rules (R16 card).
- Vietnamese user-facing copy: natural VN, no machine-translation smell.

STOP CONDITIONS (the only reasons to stop early)
- A genuine fork the cards do not resolve (two defensible architectures, ambiguous
  product intent): write a `needs_stephen` ledger entry with the options and your
  recommendation, then continue with the next unblocked task.
- Anything irreversible or public: pushing to a public repo, submitting to a store,
  sending email/messages to real external people, spending money, deleting data,
  rewriting git history. Prepare everything, stage it, and stop that task at
  `needs_stephen`.
- The wave is exhausted (no `todo`/`in_progress` tasks remain unblocked): write a
  wave summary at the end of LEDGER.md (tasks done, asks outstanding, risks seen)
  and stop.

QUALITY BAR
Honest over impressive. A task marked awaiting_review with clearly stated gaps beats
a task silently half-done. The reviewer will re-run your Verify commands; assume every
claim in your ledger entries gets checked.
```

---

## Prompt 2 - human review session (Stephen, or reviewer agent + Stephen)

```
You are reviewing completed work in the SănDeal (shopass) improvement program on
branch auto/improve-w<N>.

For each task in `awaiting_review` status in docs/improvement/BACKLOG.md, in order:
1. Read its card in the TASKS-*.md file and its ledger entries in LEDGER.md.
2. Re-run the card's Verify commands yourself. Do not trust pasted output.
3. Walk the card's "Human review" checklist - those checks are yours alone.
4. Read the diff for the task's commit (`git show <sha>`), focusing on: security
   (secrets, auth, SQL), migrations (guarded, forward-only), test honesty (do the
   tests assert behavior or just run code?), and VN copy quality.
5. Verdict:
   - PASS: set Status to `done` in BACKLOG.md, append a `review_pass` ledger entry.
   - FAIL: set Status back to `in_progress`, append a `review_fail` entry stating
     exactly what to fix; the implementation agent picks it up next run.
6. For every `needs_stephen` entry: either provide the input now (record it - secrets
   go to the env/vault, never the repo) or explicitly defer it with a date.

When the wave's tasks are all `done`:
- Merge auto/improve-w<N> to main via PR (CI must be green).
- Confirm deploy + smoke on the server if R15 is live.
- Tell the implementation agent to start wave <N+1>.
```

---

## Trigger cheat-sheet

- Start implementation: open a fresh agent session in the shopass repo, paste Prompt 1.
- Review a batch: paste Prompt 2 (or run it yourself with the checklist).
- Resume after review fixes: paste Prompt 1 again - the loop re-picks `in_progress`
  and `todo` tasks automatically from BACKLOG.md state.
