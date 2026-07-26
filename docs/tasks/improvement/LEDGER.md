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

## [2026-07-26] operator decisions (R3/R8 brand+domain)
- agent/human: Stephen / Auto
- branch/commit: hardening/wave1-r3-r7-r8-r9
- evidence: Decisions — (1) proprietary core + MIT extension (agent judgment confirmed); (2) domain shopass.cyberskill.world, Chrome Web Store intent now; (3) proceed A+B (R7/R9 + R3/R8); (4) brand Shopass.
- stephen_ask: -
- notes: Legal entity CyberSkill Software Solutions Consultancy and Development JSC.

## [2026-07-26] R3 - evidence (ready_to_review)
- agent/human: Auto
- evidence: Root `LICENSE` proprietary; `extension/LICENSE` MIT; `NOTICE.md` split + dependency summary instructions.
- stephen_ask: HITL accept → done.
- notes: -

## [2026-07-26] R7 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `SessionGuard` already bootstraps refresh (single-flight via tryRefreshOnce); added `web/test/session-guard.test.tsx`.
- stephen_ask: HITL accept → done.
- notes: Reconcile + test coverage.

## [2026-07-26] R8 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `extension/src/shared/config.ts`; build.mjs `SHOPASS_ENV`; npm scripts build/build:dev/build:staging; sender/ws/settings use config; manifest host_permissions + homepage shopass.cyberskill.world; no sandeal.vn in sync URLs.
- stephen_ask: HITL accept → done; DNS for shopass.cyberskill.world / api. still operator infra.
- notes: Chrome store intent recorded; listing copy still references github.com/shopass/sandeal-extension (follow-up).

## [2026-07-26] R9 - evidence (ready_to_review)
- agent/human: Auto
- evidence: Makefile seed/smoke refuse without APP_ENV=dev|development or ALLOW_SEED=1; deploy/README documents demo IDs 999/100.
- stephen_ask: HITL accept → done.
- notes: -

## [2026-07-26] R3/R7/R8/R9 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: main (PR #60 merged)
- evidence: Operator "Merge then continue"; CI green on #60; proprietary+MIT license, Shopass domain endpoints, SessionGuard tests, seed guard.
- stephen_ask: -
- notes: DNS for shopass.cyberskill.world / api still operator infra; CWS listing assets remain R49.

## [2026-07-26] R10 - evidence (ready_to_review)
- agent/human: Auto
- evidence: SHIP-GUIDE → Go 1.25.12; `toolchain go1.25.12` in every go.mod; CI setup-go 1.25.12; Dockerfile.go golang:1.25.12; govulncheck step (all modules incl. b2b/trust) before Go tests; local govulncheck clean on 1.25.12.
- stephen_ask: HITL accept → done (skim triage).
- notes: Version decision = keep 1.25 line, pin patch 1.25.12 (stdlib CVEs fixed through 1.25.12). Dep bumps for reachable findings: go-jose/v4 → v4.1.4 (secrets); golang.org/x/text → v0.39.0 (scrape/deal/bill).

## [2026-07-26] R10 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: main (PR #61 merged)
- evidence: Operator "Merge then continue remaining tasks in batches"; CI green on #61.
- stephen_ask: -
- notes: -

## [2026-07-26] Wave-1 Stephen asks (needs_stephen)
- agent/human: Auto
- evidence: Marked R11 (DNS A/AAAA for shopass.cyberskill.world + api), R12 (object-storage bucket + keys), R15 (VPS SSH + GHCR), R23 (Zalo OA + SMTP), R40 (GA4 vs Plausible), R49 (Chrome/Cốc Cốc $5) as needs_stephen; R16 blocked on R15.
- stephen_ask: Provide the above when ready; agent continues unblocked work.
- notes: -

## [2026-07-26] R13 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `deploy/docker-compose.observability.yml` (prometheus, blackbox, pushgateway, alertmanager, grafana); rules in `deploy/prometheus/rules/shopass.yml`; gateway metrics host :9094; authsvc `/metrics`; `obs/README.md` SLO table.
- stephen_ask: HITL accept; Telegram bot token/chat for live alert delivery (optional overlay).
- notes: Forced Telegram fire needs credentials — noop webhook until then.

## [2026-07-26] R14 - evidence (ready_to_review)
- agent/human: Auto
- evidence: Loki + Promtail in observability profile; 14d retention; Grafana dashboard `errors-across-services.json`; datasources provisioned.
- stephen_ask: HITL accept; smoke search a known log line in Grafana Explore.
- notes: -

## [2026-07-26] R17 - evidence (ready_to_review)
- agent/human: Auto
- evidence: Existing systemd timers polished (Shopass naming); `job-heartbeat.sh` + ExecStartPost; stale-job alerts in R13 rules; docs in deploy/systemd/README + deploy/README.
- stephen_ask: HITL accept; 3-day prod run still requires host install (R15/VPS).
- notes: Overlap safety via flock + pg queue lease (pre-existing).

## [2026-07-26] R34 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `/chinh-sach-bao-mat` + `/dieu-khoan` (VN primary + EN summary); draft-pending-counsel banner; footer links; sitemap entries + test.
- stephen_ask: Stephen reads drafts as signatory; counsel optional via R37.
- notes: legal-review: draft-pending-counsel.

## [2026-07-26] R13/R14/R17/R34 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: main (PR #62 merged)
- evidence: Operator "Merge then continue remaining tasks in batches"; CI green on #62.
- stephen_ask: Telegram Alertmanager creds still optional.
- notes: -

## [2026-07-26] R18 - evidence (ready_to_review)
- agent/human: Auto
- evidence: Decision (b) experimental BFF; compose profile `bff`; removed from gateway depends_on; `docs/architecture/web-data-path.md`; TASK-COVERAGE note.
- stephen_ask: HITL accept decision (b).
- notes: Default services list no longer includes bff.

## [2026-07-26] R31 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `extension/src/shared/allowlist.ts`; non-empty DNR allow rules; manifest DNR resource; `allowlist.test.ts` fails on host_permissions drift; build copies rules.json.
- stephen_ask: HITL read allowlist (api.shopass.cyberskill.world + 127.0.0.1:8080).
- notes: Hard exfil boundary remains host_permissions; DNR is allow-documented (no marketplace block).

## [2026-07-26] R35 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `/minh-bach` transparency page; links DISCLOSURE, allowlist, legal pages, repo; footer + sitemap.
- stephen_ask: HITL skeptical read.
- notes: R36 public mirror still placeholder.

## [2026-07-26] R18/R31/R35 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: main (PR #63 merged)
- evidence: Operator standing "merge then continue remaining tasks in batches"; CI green on #63.
- stephen_ask: -
- notes: Next unblocked Wave-2 candidates: R26 (ML versioning), R38 (landing). Still needs_stephen: R11/R12/R15/R19/R23/R24/R36/R40/R49.

## [2026-07-26] R26 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `model_run` migration 0003; gate MAPE≤1.2× trailing best; suppress p_bottom / category_prior fallback; artifact stub; Pushgateway counter + `ShopassModelGateTripped` alert; pytest `test_gate.py` + consecutive runs in forecast writer tests.
- stephen_ask: HITL skim thresholds; confirm chart honesty when p_bottom suppressed.
- notes: -

## [2026-07-26] R38 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `/` landing with R38 hero copy, recharts demo, how-it-works, trust strip (R34/R35/GitHub), FAQ+JSON-LD, signup/install CTAs with analytics stub; signed-in refresh cookie → `/dashboard`.
- stephen_ask: HITL visual/copy review (company face).
- notes: R40 events are buffered stubs until analytics vendor decision.

## [2026-07-26] R26/R38 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: main (PR #65 merged)
- evidence: Operator standing continue-in-batches; CI green on #65.
- stephen_ask: -
- notes: Next unblocked Wave-2: R39 (pricing, depends R38), R45/R43 (depend R38). R27 blocked on R24.

## [2026-07-26] R39 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `/bang-gia` Free/Premium/Plus/Pro + FAQ; waitlist modal → `POST /api/waitlist` → gateway `POST /v1/leads/waitlist` → bill `marketing_lead` (migration 0006); public JWT path; analytics stubs `pricing-view`/`tier-click`/`waitlist-submit`; `NEXT_PUBLIC_CHECKOUT_LIVE` flip for R28; landing nav/footer + sitemap.
- stephen_ask: HITL confirm tier prices (29k/49k/79k) still match intent before indexing.
- notes: -

## [2026-07-26] R39 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: main (PR #67 merged)
- evidence: Operator standing continue-in-batches; CI green on #67.
- stephen_ask: -
- notes: Next unblocked Wave-2: R43 (fake-sale checker), R45 (onboarding). Still needs_stephen: R11/R12/R15/R19/R23/R24/R36/R40/R49.

## [2026-07-26] R43 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `/kiem-tra-sale-ao` + public `POST /v1/tools/fake-sale-check` (deal, no JWT); tracked → verdict/median/chart snippet; untracked → lead via waitlist `source=tool`; `/lich-sale` countdown + reminder subscribe `source=sale-calendar`; analytics `tool-submit`/`verdict-shown`/`lead-captured`.
- stephen_ask: HITL paste 5 real product URLs; check verdict credibility + VN wording.
- notes: Untracked enqueue deferred (R27/R24); lead capture only for now.

## [2026-07-26] R43 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: main (PR #69 merged)
- evidence: Operator standing continue-in-batches; CI green on #69.
- stephen_ask: -
- notes: Next unblocked Wave-2: R45 (onboarding). Still needs_stephen: R11/R12/R15/R19/R23/R24/R36/R40/R49.
