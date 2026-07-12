# Part D - compliance and trust (R32-R37)

PDPL (Law 91/2025/QH15) has been effective since 2026-01-01; penalties reach 5% of prior-year revenue. Trust artifacts double as marketing (anti-Honey positioning).

---

## R32 - PDPL minimum set: consent, DSAR, breach runbook

Wave 3 | Effort L | Depends: - | Stephen input: -

Why: comply-service scaffolding and the extension consent framework exist, but DSAR enforcement, breach notification, and versioned consent records are stubs per `docs/FR-COVERAGE.md`. Legal exposure plus a blocked sales asset ("PDPL-ready by design").

Steps:
1. Consent records: persist user consent with policy version, timestamp, scope (account data, price alerts, extension cart read); re-consent flow on policy version bump.
2. DSAR endpoints: export (JSON bundle of user rows across services) and delete (cascade or anonymize; keep financial rows per retention law) - wire the existing comply stubs to real queries; admin audit log of DSAR executions.
3. Breach runbook: `docs/compliance/BREACH-RUNBOOK.md` with the 72h notification path, roles, MPS contact points, template letters (VN).
4. Data inventory: one table in docs listing every personal-data field, store, purpose, retention (feeds R33's gates and the R35 transparency page).
5. Tests: DSAR export completeness against the inventory; delete leaves no orphaned PII.

Acceptance: a test user can be exported and deleted end to end; consent versions recorded; runbook reviewed.

Verify: DSAR export sample (redacted) + delete-then-scan proof in ledger.

Human review: walk the breach runbook once as a tabletop; confirm the data inventory matches reality (spot-check 3 fields).

---

## R33 - CI compliance gates (no-cleartext, consent, DPIA)

Wave 3 | Effort M | Depends: R32 | Stephen input: -

Why: NFR-COMPLY-001 promises automated gates (no-cleartext scan, consent coverage, DPIA-overdue detection); none exist in `.github/workflows/ci.yml`.

Steps:
1. No-cleartext gate: CI script scanning migrations + code for PII columns lacking the documented protection level (drive it from R32's data inventory file so it cannot drift).
2. Consent-coverage gate: test asserting every data-collecting endpoint/extension pipeline checks a consent scope (table-driven from the inventory).
3. DPIA tracker: `docs/compliance/DPIA.md` with review dates; CI warns when a review date passes.
4. Wire all three into CI as required checks; document waiver procedure (`-- reviewed:` style tag like R16).

Acceptance: seeding a violation (test PII column, unconsented endpoint, stale DPIA date) turns CI red; clean repo is green.

Verify: red-run and green-run CI links in ledger.

Human review: read the gate scripts once; confirm the waiver path requires a human tag.

---

## R34 - public privacy policy + terms pages (VN + EN)

Wave 1 | Effort S | Depends: - | Stephen input: -

Why: no policy pages exist; the Chrome Web Store (R49) requires a privacy-policy URL, and PDPL requires disclosure. Blocks store submission and looks untrustworthy to exactly the audience the product courts.

Steps:
1. Draft `/chinh-sach-bao-mat` (privacy, VN primary + EN) and `/dieu-khoan` (terms) as static pages in `web/app/(marketing)/`: data collected (account, tracked products, extension cart/voucher reads under consent), purposes, retention, DSAR contact (info@cyberskill.world or product address), affiliate disclosure.
2. Source the substance from R32's inventory (draft now, refine when R32 lands - do not block on it).
3. Footer links on all pages; sitemap entries; company block: CyberSkill Software Solutions Consultancy and Development JSC, 1st Floor, 207A Nguyen Van Thu, Tan Dinh Ward, HCMC.
4. Legal-review flag: mark draft-pending-counsel in ledger (ties to R37's decision).

Acceptance: both pages live, linked in footer, indexed, VN + EN toggles working (static duplication acceptable pre-i18n).

Verify: URLs + rendered screenshots in ledger.

Human review: Stephen reads both drafts (he is the accountable signatory); optionally forwards to counsel with R37.

---

## R35 - transparency page (anti-Honey positioning)

Wave 2 | Effort S | Depends: R34 | Stephen input: -

Why: Honey was cut off by Rakuten (2026-01-12) and Impact.com over hidden attribution rewriting. SănDeal's code already takes the opposite stance (user-initiated affiliate only, consent-gated reads, DNR allowlist). The market cannot reward what it cannot see.

Steps:
1. Build `/minh-bach` (transparency) in `(marketing)`: what the extension reads (cart, vouchers - with annotated screenshots), what it never reads (passwords, messages, other tabs), when affiliate links activate (only on explicit click - link the R29 guardrail test), how the company earns (affiliate + Premium), data flow diagram.
2. Link the open-source repo (R36) and the DNR allowlist as proof artifacts.
3. VN primary, EN secondary; internal links from landing (R38) trust strip and store listing (R49).

Acceptance: page live with the guardrail-test link and repo link (placeholder until R36 lands).

Verify: URL + screenshot in ledger.

Human review: read as a skeptical user; every claim must be checkable via a link, not asserted.

---

## R36 - open-source the extension publicly

Wave 2 | Effort M | Depends: R2, R3, R31 | Stephen input: decision (go-public approval)

Why: the PRD's trust moat depends on a verifiable extension; a public repo turns the transparency claims into checkable facts and is press-worthy in the post-Honey window.

Steps:
1. Pre-flight: R2 done (no node_modules), R3 extension LICENSE (MIT), secret scan over full history of `extension/` paths (gitleaks); decide split-repo (recommended: `cyberskill/sandeal-extension` public mirror via subtree/filter) vs monorepo-public.
2. Add `extension/README.md` (build, test, load-unpacked), `SECURITY.md` (report channel), `CONTRIBUTING.md` (light).
3. Reproducible build notes: steps to diff a store package against a source build (store reviewers and skeptics both use this).
4. Public CI on the mirror running the jest suite.
5. Stephen ask: approve going public + repo location/name.

Acceptance: public repo with green CI, license, security policy; transparency page (R35) links it.

Verify: repo URL + CI badge in ledger.

Human review: Stephen approves publication; reviewer runs gitleaks once more on the final public history.

---

## R37 - scraping legal posture memo

Wave 3 | Effort S | Depends: - | Stephen input: decision (whether to engage counsel)

Why: scraping Shopee/TikTok/Lazada has ToS friction; the first press cycle (R53) will ask. An unprepared answer becomes the story.

Steps:
1. Memo in `docs/compliance/SCRAPING-POSTURE.md`: per-platform ToS analysis, public-data scope, robots stance (from R24 evidence), rate budgets, data minimization (prices only, no personal data scraped), takedown-response process with owner + SLA.
2. Fold in the affiliate-program relationship (R29): registered partners scraping public listing prices sit differently than anonymous bots - document it.
3. FAQ answers for press/platform inquiries (3 questions, 3 answers, VN + EN).
4. Stephen ask: decide whether a VN counsel review is worth it pre-launch (recommended, half-day engagement).

Acceptance: memo merged; takedown process has a named owner and inbox.

Verify: doc link in ledger.

Human review: Stephen reads the memo and decides on counsel; confirm the takedown inbox actually routes to someone.
