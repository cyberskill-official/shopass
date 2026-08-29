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
- evidence: SHIP-GUIDE → Go 1.25.13; `toolchain go1.25.13` in every go.mod; CI setup-go 1.25.13; Dockerfile.go golang:1.25.13; govulncheck step (all modules incl. b2b/trust) before Go tests; local govulncheck clean on 1.25.13.
- stephen_ask: HITL accept → done (skim triage).
- notes: Version decision = keep 1.25 line, pin patch 1.25.13 (stdlib CVEs fixed through 1.25.13). Dep bumps for reachable findings: go-jose/v4 → v4.1.4 (secrets); golang.org/x/text → v0.39.0 (scrape/deal/bill).

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

## [2026-07-26] HITL confirm (R39 prices + R43 tools)
- agent/human: Stephen
- evidence: Operator "Confirm, continue the rest" — pricing tiers and tool direction accepted; continue unblocked Wave-2 (R45 next).
- stephen_ask: -
- notes: -

## [2026-07-26] R45 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `/onboarding` paste→chart→「Báo tôi khi chạm đáy」; `createAhaAlert` prefers `bottom_predicted`, Free 402→`real_sale`; signup auto-login + `next=/onboarding`; empty states (dashboard/wishlist/alerts) deep-link; chart `BottomAlertCta`; analytics `first-track`/`first-alert`.
- stephen_ask: HITL run once on a fresh account; note hesitation points.
- notes: -

## [2026-07-26] R45 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: main (PR #71 merged)
- evidence: Operator "Confirm, continue the rest"; CI green on #71.
- stephen_ask: -
- notes: Wave-2 agent-unblocked exhausted. Next: Wave-3 no-Stephen (R44 comparison, R47 blog, R42 keywords, R48 perf, R32 PDPL, R20 k6).

## [2026-07-26] R44 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `/so-sanh/shopass-vs-beecost` feature matrix (10 rows) + JSON-LD; `/thay-the-honey` trust story with Verge/Rakuten/Impact citations + link `/minh-bach`; alias redirect `sandeal-vs-beecost` → shopass; sitemap + landing footer.
- stephen_ask: HITL verify every comparison row before indexing.
- notes: Fact-check list = `SHOPASS_VS_BEECOST` in `web/lib/compare/beecost.ts` + Honey source URLs on page.

## [2026-07-26] R44 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: main (PR #73 merged)
- evidence: Operator continue-the-rest; CI green on #73.
- stephen_ask: -
- notes: Next Wave-3 unblocked: R47 blog/changelog/RSS.

## [2026-07-26] R47 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `/blog` + 2 seed posts (sale-ảo 10s, chào Shopass); `/changelog` v0.3.0; `/rss.xml` RSS 2.0; sitemap + footer links; file-based content (no MDX dep).
- stephen_ask: HITL approve post voice (sets content bar).
- notes: -

## [2026-07-26] R47 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: main (PR #75 merged)
- evidence: Operator continue-the-rest; CI green on #75.
- stephen_ask: -
- notes: Agent-unblocked Wave-2/3 growth batch paused. Remaining need Stephen (R11/R12/R15/R19/R23/R24/R36/R40/R49) or heavy Wave-3 (R32/R42/R48/R20).

## [2026-07-26] R42 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `docs/growth/KEYWORD-MAP.md` (~35 targets); batch-1 = 10 new keyword pages (13 total) with unique intros/FAQs/related links; template upgrade + calendar CTA→`/lich-sale`; JSON-LD FAQ/ItemList.
- stephen_ask: HITL read 3 random pages for template-smell.
- notes: Interlink sketch in KEYWORD-MAP.md.

## [2026-07-26] R42 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: main (PR #77 merged)
- evidence: Operator "Go"; CI green on #77.
- stephen_ask: -
- notes: Next: R52 referral UI.

## [2026-07-26] R52 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `GET /v1/referral/me` + `POST /v1/referral/attribute` (bill PG repo); dashboard ReferralCard (code/copy/Zalo); signup `?ref=` → attribute after auto-login; self-referral blocked; analytics `share-click`/`referred-signup`; default reward proposal 1mo Premium both sides (pending Stephen economics).
- stephen_ask: Approve reward economics; try self-referral fails.
- notes: Rewards still delayed via anti-fraud event (DEC-BILL-19) — UI shows status/uses only.

## [2026-07-26] R52 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: main (PR #79 merged)
- evidence: Operator "Go"; CI green on #79.
- stephen_ask: -
- notes: Next agent-doable: R48 Lighthouse/a11y, R20 k6, R32 PDPL.

## [2026-07-26] R48 - evidence (ready_to_review)
- agent/human: Auto
- evidence: next.config compress + image formats/remotePatterns; dynamic recharts on landing + chart; ListSkeleton/RouteError on alerts/chart; marketing skip-link + single h1; jest-axe on landing + keyword fixture; LHCI soft workflow (continue-on-error) for `/`, keyword, `/bang-gia`.
- stephen_ask: Spot-check product chart on mid-range Android/4G once; promote LHCI to hard gate after two green weeks.
- notes: Soft gate only — budgets warn, job does not block merge yet.

## [2026-07-26] R48 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: main (PR #81 merged)
- evidence: Operator "Go do all remainings"; CI green on #81 (incl. soft Lighthouse).
- stephen_ask: -
- notes: Next agent-doable: R20 k6, R32 PDPL.

## [2026-07-26] R20 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `loadtests/k6/nfr-infra-001.js` mixed 50 RPS chart+tracked-products; thresholds p95&lt;500/300; weekly workflow Mon 01:00 UTC (skips if secrets missing); baseline log `docs/non-functional-requirements/infra/NFR-INFRA-001-k6-baseline.md`. Compare/wishlist omitted (gateway 404 beta).
- stephen_ask: Add `K6_ACCESS_TOKEN(S)` + `K6_PRODUCT_ID` secrets; prefer token pool vs rate-limit raise for 50 RPS.
- notes: First green weekly row pending secrets + live API (R11).

## [2026-07-26] R20 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: main (PR #83 merged)
- evidence: Operator "Go do all remainings"; CI green on #83. Weekly job skips until K6 secrets set.
- stephen_ask: Configure K6_ACCESS_TOKEN(S) + K6_PRODUCT_ID for first green baseline row.
- notes: Next agent-doable: R32 PDPL (PR #84), then R33.

## [2026-07-26] R32 - evidence (ready_to_review)
- agent/human: Auto
- evidence: complysvc HTTP added for consent/DSAR/breach; same-DB DSAR adapters for app_user, track/wishlist/alerts, and bill payment PII; gateway/compose/prometheus wired; breach runbook and data inventory docs added; `go test ./...` in services/comply and services/gateway passed; CyberOS gate green with `PYENV_VERSION=3.12.13 PYTHONPATH=.`.
- stephen_ask: HITL review acceptance for R32; confirm production incident contacts before use.
- notes: Did not set done. First unqualified CyberOS gate failed because local pyenv lacks Python 3.11 required by services/ml/.python-version; rerun passed using installed Python 3.12.13 and PYTHONPATH.

## [2026-07-26] R32 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: main (PR #84 merged)
- evidence: Operator "Go do all remainings"; CI green on #84.
- stephen_ask: Confirm production incident contacts; tabletop breach runbook; spot-check 3 inventory fields.
- notes: Next agent-doable: R33 CI compliance gates.

## [2026-07-26] R33 - evidence (ready_to_review)
- agent/human: Auto
- evidence: CI job `Compliance gates (R33)` wires no_cleartext + consent_coverage + inventory_pii + DPIA review; docs `DPIA.md`, `CONSENT-SURFACES.md`, `CI-WAIVERS.md`; overdue DPIA red-path asserted in CI; deferred consent surfaces waived `reviewed:R33-bootstrap` until IsAllowed wired on track/notif/b2b.
- stephen_ask: Read gate scripts once; confirm waiver tags require a human id; plan removal of R33-bootstrap waivers.
- notes: Local green + overdue red via `DPIA_TODAY=2099-01-01`.

## [2026-07-26] R33 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: main (PR #88 merged)
- evidence: Operator "Go do all remainings"; CI green on #88 including Compliance gates (R33) job + overdue DPIA red-path.
- stephen_ask: Plan removal of reviewed:R33-bootstrap consent waivers when IsAllowed lands on track/notif/b2b.
- notes: Agent-unblocked Wave-3 compliance batch complete. Remaining need Stephen (R11/R12/R15/R19/R23/R24/R36/R40/R49) or account/outreach tasks.

## [2026-07-26] R37 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `docs/compliance/SCRAPING-POSTURE.md` — per-platform ToS (Shopee/TikTok/Lazada VN), public-data scope aligned with DATA-INVENTORY, robots/proxy/pacing/deny-by-default stance (scrapesvc), price_snapshot retention pointer, takedown owner Stephen Cheng + info@cyberskill.world (48h/7d SLA), affiliate R29 posture, VN+EN press FAQ (3×3), Stephen ask (VN counsel half-day).
- stephen_ask: Read memo; confirm takedown inbox routing; decide VN counsel pre-launch (recommended yes, half-day).
- notes: Draft — not counsel-reviewed. No fake case citations.

## [2026-07-26] R19 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `docs/ops/RETENTION-POLICY.md` — propose 18mo raw / forever `price_daily` / compress 30d / chunk 7d; fixture cost estimate; notes `0020_compression_policy.up.sql` already encodes the policy.
- stephen_ask: Sign off horizons; confirm jobs visible in `timescaledb_information.jobs` on staging/prod.
- notes: No new migration — already present; this is the explicit product decision record.

## [2026-07-26] R36 - evidence (ready_to_review)
- agent/human: Auto
- evidence: Extension OSS prep — `README.md`, `SECURITY.md`, `CONTRIBUTING.md`, `REPRODUCIBLE-BUILD.md`; MIT LICENSE already present. Publish blocked on founder approval + mirror location.
- stephen_ask: Approve go-public + repo name (`shopass-extension` recommended); run gitleaks on final public history before push.
- notes: Docs only — no public mirror created.

## [2026-07-26] R49 - evidence (ready_to_review)
- agent/human: Auto
- evidence: Icons 16/32/48/128 + promo tile; `_locales` vi/en + `default_locale`; manifest icons/action; `store/` LISTING-VI/EN, DATA-DISCLOSURE, SCREENSHOTS checklist, `package.sh`; build copies icons/locales.
- stephen_ask: Approve placeholder mark (or supply brand); capture 5×1280 screenshots; create Chrome ($5)/Edge/Cốc Cốc accounts and submit.
- notes: Screenshots still TODO capture; localhost host stripped only in store zip script.

## [2026-07-26] R56 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `docs/growth/PARTNER-ONE-PAGER.md`, `KOC-KIT.md`, `PARTNER-KOC-TARGETS.md` (10 partners + 20 KOC slots + UTM/ref tracking).
- stephen_ask: Prune target list before any outreach; log first 3 conversations.
- notes: Research kits only — no outreach performed.

## [2026-07-26] R58 - evidence (ready_to_review)
- agent/human: Auto
- evidence: `docs/growth/BRAND-SHEET.md` — Shopass public name, colors, handle checklist; notes placeholder extension mark.
- stephen_ask: Confirm final name/domain; register handles; supply final logo.
- notes: Handle registration + full repo grep sweep incomplete until name locked.

## [2026-07-26] R37 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: main (PR #90 merged)
- evidence: Operator "Go do all remainings"; memo + takedown owner/inbox present; CI green on #90.
- stephen_ask: Optional VN counsel half-day; confirm inbox routing.
- notes: Counsel review still recommended, not required for memo acceptance.

## [2026-07-26] R56 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: main (PR #90 merged)
- evidence: Operator continue/go; kits + target list delivered.
- stephen_ask: Prune list; first 3 conversations.
- notes: R19/R36/R49/R58 remain ready_to_review pending Stephen decisions/accounts/screenshots/handles.

## [2026-07-27] R19 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: docs/hitl-accept-improvement-r19-r36-r49-r58 (this PR)
- evidence: Operator "accept" for improvement ready_to_review queue; retention policy doc + existing compression migration accepted as the product decision record.
- stephen_ask: Confirm `timescaledb_information.jobs` shows retention/compress jobs on staging/prod when those envs exist.
- notes: No new migration required — policy already encoded in `0020_compression_policy.up.sql`.

## [2026-07-27] R36 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: docs/hitl-accept-improvement-r19-r36-r49-r58 (this PR)
- evidence: Operator "accept"; extension OSS prep docs (README/SECURITY/CONTRIBUTING/REPRODUCIBLE-BUILD + MIT) accepted as the go-public decision/kit.
- stephen_ask: Create public mirror (`shopass-extension` recommended); run gitleaks on final public history before push; link from transparency page (R35).
- notes: Docs/kit accepted — public GitHub mirror + CI on that mirror remain ops follow-up.

## [2026-07-27] R49 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: docs/hitl-accept-improvement-r19-r36-r49-r58 (this PR)
- evidence: Operator "accept"; store kit (icons, locales, LISTING-VI/EN, DATA-DISCLOSURE, SCREENSHOTS checklist, package.sh) accepted.
- stephen_ask: Capture 5×1280 screenshots; create Chrome ($5)/Edge/Cốc Cốc developer accounts; submit listing; approve or replace placeholder mark.
- notes: Kit/docs done — store account creation, screenshots, and submission remain ops.

## [2026-07-27] R58 - HITL accept → done
- agent/human: Stephen / Auto
- branch/commit: docs/hitl-accept-improvement-r19-r36-r49-r58 (this PR)
- evidence: Operator "accept"; `docs/growth/BRAND-SHEET.md` + handle checklist accepted as brand consistency decision surface.
- stephen_ask: Confirm final public name/domain if still undecided; register TikTok/Facebook/Telegram/Zalo OA/GitHub/X handles; supply final logo; vault credentials outside repo.
- notes: Brand sheet accepted — live handle registration + final logo swap remain ops.

## [2026-07-26] Staging next-steps plan — HITL + sequence check
- agent/human: Stephen / Auto
- evidence: Operator implement “Suggested next steps”. Paths 1–2 HITL reaffirmed under deferred FCM/proxy/pay. Helpers already merged (`c2d2dcb` / PR #56). R1 (TASK-INFRA-006 / PR #57) and Wave‑1 R2/R4/R5/R17 already `done` on backlog — no new engineering PR required for that sequence.
- stephen_ask: Pick post‑R1 branch (Trust / Comply / Channels) or provision secrets (FCM / HTTPS_PROXY / pay sandboxes).
- notes: Still banned until asked: Lazada/TikTok live, Vault `_FILE`, Google OAuth.

