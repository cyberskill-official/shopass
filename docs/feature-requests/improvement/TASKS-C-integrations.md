# Part C - stubs to real (R23-R31)

Revenue-ordered: alerts reach (R23) and live data (R24) come before payments (R28). Affiliate (R29) monetizes before Premium.

---

## R23 - Zalo ZNS + email senders beside FCM

Wave 1 | Effort M | Depends: - | Stephen input: account + creds (Zalo OA/ZNS, SMTP provider)

Why: `services/notif/internal/` implements only `fcm/`; `routing.go` ranks an `email` channel with no sender behind it. In Vietnam, purchase alerts live on Zalo; FCM-only limits reach to web-push grantors. Alerts are the habit loop - this is retention infrastructure.

Steps:
1. Define a `Sender` interface in notifsvc matching the existing fanout dispatch; adapt `fcm` to it.
2. Add `email` sender (SES or Postmark, per Stephen's account choice) with templated VN messages (price-drop, bottom-predicted), List-Unsubscribe header, send-log rows mirroring FCM's.
3. Add `zalo` sender: OA free-form message first (48h-window rules), ZNS template flow behind a feature flag (templates need Zalo approval - draft them, record in ledger for Stephen to submit).
4. Extend `routing.go`: zalo > fcm > email default for VN; add per-user channel preference column if missing (guarded migration per R16).
5. Apply the existing midnight-spike flattening to all channels; per-channel rate budgets in config.
6. Tests: routing matrix, template rendering, dedupe unaffected; fake transport in unit tests, sandbox send in staging.
7. Stephen asks recorded in ledger: Zalo OA + ZNS account, SMTP creds, sender-domain SPF/DKIM once the R11 domain exists.

Acceptance: a bottom-price alert in staging lands in an email inbox and a Zalo test OA; FCM unchanged; per-channel send logs written.

Verify: staging alert transcript per channel + routing unit tests in ledger.

Human review: receive one real alert on your own Zalo + email; check VN copy; confirm unsubscribe works.

---

## R24 - battle-test live scraping behind residential proxy

Wave 2 | Effort L | Depends: R17 | Stephen input: budget (residential proxy) + creds

Why: per `docs/FR-COVERAGE.md` the Shopee adapter parses real JSON but has never run live; the Playwright farm is untested on real pages; block/CAPTCHA behavior unknown. No live data, no product.

Steps:
1. Stephen ask: approve a small residential proxy plan (PRD estimates $30-60 per 1k users/mo) and supply creds for the existing `HTTPS_PROXY` hook.
2. Seed list: top 200 products across 5 categories as a config file, tiered cadence via the existing rescheduler (`services/scrape/cmd/scrapesvc/main.go:61`).
3. Run scrapesvc on schedule 7 days against Shopee; measure success/block/CAPTCHA rates and latency (add obs counters where missing).
4. Implement the documented pacing/jitter budgets in `region/` per-platform config; record the robots.txt stance (feeds R37).
5. Extend to TikTok Shop and Lazada via the farm; fix selectors against live DOM; capture fixture updates from real payloads so tests track reality.
6. Write an IP-ban recovery runbook: rotate proxy pool, back off tier cadence, alert threshold on block rate > 20%.

Acceptance: 7 consecutive days of scheduled Shopee scraping at >90% success; TikTok + Lazada each proven on >=50 products; block-rate alert wired.

Verify: success-rate dashboard export + queue stats in ledger.

Human review: Stephen approves the proxy spend; reviewer checks pacing config against the platform budgets in `region/` and skims the ban runbook.

---

## R25 - pluggable CAPTCHA path (manual queue first)

Wave 2 | Effort M | Depends: R24 | Stephen input: budget (optional commercial solver)

Why: `services/scrape/internal/captcha/solver.go` is simulated per `docs/AUDIT-REPORT.md`. Live scraping (R24) will hit CAPTCHA walls; without a path, coverage silently decays.

Steps:
1. Define a `Solver` interface where the simulated solver sits; keep the simulation as the test implementation.
2. Implement a manual-solve queue first: blocked jobs park in a table, a minimal admin page (or CLI) shows the challenge, a human solves, job resumes. Cheap and legally clean.
3. Add a commercial-solver adapter (2Captcha/CapSolver style) behind a feature flag and env creds; leave disabled until Stephen approves spend.
4. Metrics: challenges seen, solve rate, time-to-solve per path (feeds R13 dashboards).

Acceptance: a real CAPTCHA encountered in staging parks the job, gets solved manually, and the job completes; solver metrics visible.

Verify: one end-to-end parked-solved-resumed transcript in ledger.

Human review: decide whether the commercial adapter gets a budget now or later; check the parked-job table cannot leak challenge images publicly.

---

## R26 - ML model versioning + evaluation gate

Wave 2 | Effort M | Depends: - | Stephen input: -

Why: Prophet + LightGBM code is real but nothing versions models or gates quality; a silently degraded model would emit wrong "buy now" signals - the fastest way to burn the trust moat.

Steps:
1. Add a `model_run` table (guarded migration): version, model_kind, training window, feature-set hash, per-category backtest MAPE and hit-rate, artifact path.
2. Persist artifacts per run under a versioned path; `price_forecast` rows already carry `model_kind` - add `model_run_id`.
3. Backtest gate in the mlforecast job: publish forecasts only if backtest beats threshold (start: MAPE within 1.2x of trailing 30-day best; else fall back to Prophet baseline or suppress `p_bottom`).
4. Emit evaluation metrics to obs; alert on gate trips (R13).
5. pytest coverage for the gate logic and fallback.

Acceptance: two consecutive forecast runs recorded with versions + metrics; a forced-bad model triggers fallback and alert, and its forecasts are not published.

Verify: `model_run` rows + forced-failure transcript in ledger.

Human review: sanity-check thresholds; confirm suppressed forecasts render honestly in the chart UI (no stale p_bottom shown).

---

## R27 - cold-start backfill + honest history-depth UI

Wave 2 | Effort M | Depends: R24 | Stephen input: -

Why: forecasts need months of history; at launch most products have days. The PRD's answer (prioritized seeding + crowdsourced backfill) exists only as strategy text. Fake-looking full charts on thin data would break trust.

Steps:
1. Prioritized seeding: category x shop seed lists driving the R24 schedule, expanding by observed demand (tracked products, R41/R46 requests auto-enqueue scraping).
2. Extension-triggered backfill: on first product visit, enqueue a history-priority job (respect consent gates; queue exists - add the priority lane).
3. UI honesty: chart shows tracked-since date and a "history depth" note; suppress p_bottom under a minimum-days threshold (align with R26's gate).
4. Tests: enqueue-on-demand path; UI renders the thin-history state correctly.

Acceptance: pasting a never-seen product URL gets first data within one scrape cycle and the chart states its real depth.

Verify: demand-enqueue transcript + screenshot in ledger.

Human review: check the thin-history wording sets expectations without scaring users off.

---

## R28 - payments sandbox: MoMo/ZaloPay/VNPay real flows

Wave 3 | Effort L | Depends: R11 | Stephen input: account (merchant registrations) + creds

Why: `services/bill/internal/pay/gateway.go:47` is a placeholder; signature checks are trivial. No real checkout, no Premium revenue, and unverifiable webhooks are a fraud surface.

Steps:
1. Stephen ask first (long lead time): register merchant/sandbox accounts for MoMo, ZaloPay, VNPay; deliver creds as env secrets.
2. Implement one provider end to end first (pick by fastest sandbox approval): create order, redirect/deeplink, IPN/webhook with real signature verification (HMAC per provider spec), idempotent state machine (created -> pending -> paid/failed/expired), refund hook stub.
3. Reconciliation job: daily compare provider statements vs local orders; report mismatches (bill service has reconciliation scaffolding - wire it).
4. Feature-gate Premium activation on paid status (gating logic exists per FR-BILL-005; connect it).
5. Repeat for the other two providers; contract tests with recorded sandbox payloads.

Acceptance: sandbox payment activates Premium on a test account for each wired provider; tampered webhook signatures rejected; reconciliation report clean.

Verify: sandbox transaction IDs + webhook logs + reconciliation output in ledger.

Human review: Stephen completes merchant KYC; reviewer replays a tampered webhook and confirms rejection; check refund path is at least stubbed with an owner alert.

---

## R29 - affiliate programs live + attribution logging

Wave 3 | Effort M | Depends: R24 | Stephen input: account (Shopee/TikTok affiliate registrations)

Why: affiliate is revenue with zero paywall friction and precedes Premium in the PRD's model (EPC ~$0.03, category commissions 2.5-12%). Code enforces user-initiated links only (the post-Honey compliance stance) but no program registration or attribution verification exists.

Steps:
1. Stephen ask: register Shopee affiliate and TikTok Shop affiliate accounts; deliver IDs/tokens.
2. Verify per-platform deep-link formats and commission windows; encode them in `region/` config with tests (they differ per country - the config layer exists for this).
3. Attribution event log: every user-initiated affiliate click writes an event (user, product, platform, timestamp, link id) for later statement reconciliation.
4. Monthly reconciliation report: platform statement vs click log; store in the bill service alongside R28's reconciliation.
5. Guardrail test: no affiliate parameter is ever attached without an explicit user click (regression-protect the anti-cookie-stuffing stance - this is the R35/R36 marketing claim).

Acceptance: a real tracked click-through earns a sandbox/live commission entry that matches our event log.

Verify: click event rows + platform dashboard screenshot in ledger.

Human review: Stephen completes registrations; reviewer re-runs the guardrail test and reads its assertion.

---

## R30 - Google OAuth live end-to-end test

Wave 2 | Effort S | Depends: R11 | Stephen input: creds (Google OAuth client)

Why: the JWKS path is only tested against a self-signed token (`services/auth/internal/auth/oauth_google_test.go`); the live flow has never run. cyberos already hardened this path once - reuse those lessons (JWKS verify, JIT provisioning).

Steps:
1. Stephen ask: create the Google OAuth client (or reuse the cyberos console project) with the R11 domain redirect URI; deliver `GOOGLE_CLIENT_ID/SECRET`.
2. Run the full login on staging; verify JWKS fetch, audience check, JIT user creation, refresh issuance.
3. Add a staging checklist entry to `deploy/README.md`; keep the fake-token unit tests as-is.

Acceptance: a real Google account logs in on staging and lands authenticated on the dashboard.

Verify: redacted flow trace (no tokens) in ledger.

Human review: log in yourself with a personal Google account; confirm account row + no duplicate on second login.

---

## R31 - DNR rules integration tests

Wave 2 | Effort S | Depends: - | Stephen input: -

Why: `extension/manifest.json` points at `dnr/rules.json`, and "we technically cannot exfiltrate to unknown hosts" is a trust-critical claim (R35/R36), but no test asserts the rules actually block non-allowlisted hosts.

Steps:
1. Add tests validating rules.json structure and semantics: allowlisted API/WS hosts permitted, everything else blocked for extension-initiated requests.
2. If feasible in the existing jest harness, simulate matches with the declarativeNetRequest rule-matching logic (or a table-driven evaluator mirroring it); otherwise add a Playwright-driven check gated like the farm tests.
3. Keep rules and R8's env config generated from one source so they cannot drift.

Acceptance: a test fails if anyone adds a new outbound host without updating the allowlist consciously.

Verify: test run output in ledger.

Human review: read the allowlist; confirm it contains only the final domain set from R8.
