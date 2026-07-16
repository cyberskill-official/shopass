# Part E - web, SEO, conversion (R38-R48)

The lead engine. R40 (analytics) instruments everything else - never ship an E task without its events.

---

## R38 - real landing page

Wave 2 | Effort M | Depends: R4 | Stephen input: -

Why: `web/app/page.tsx` only redirects to `/dashboard`; logged-out visitors bounce off a login wall. There is no page that sells anything.

Steps:
1. Build the landing at `/` (marketing group, no auth): hero one-liner ("Biết khi nào giá chạm đáy - trên Shopee, TikTok Shop, Lazada") + subline on fake-sale detection; live example price chart (reuse the recharts component with demo data); 3-step how-it-works; extension install CTA + email fallback (R46 form); trust strip (open source R36, PDPL R34, no cookie-stuffing R35 - link each); FAQ (5 questions incl. "Sale thật hay ảo là gì?"); footer with policy links.
2. Logged-in users hitting `/` still bounce to `/dashboard` (cookie check server-side).
3. Apply CyberSkill design tokens where sensible but keep SănDeal its own product brand (Stephen note in ledger if a brand kit is wanted).
4. Metadata per R4 pattern; JSON-LD Organization + WebSite; OG image specific to the hero.
5. R40 events on every CTA (install-click, signup-click, email-submit).

Acceptance: logged-out `/` renders the full page; Lighthouse SEO >= 95; CTAs fire events.

Verify: screenshot, Lighthouse JSON, event log sample in ledger.

Human review: Stephen reviews VN copy and visual quality bar personally (this page is the company's face for the product).

---

## R39 - pricing page + Premium waitlist

Wave 2 | Effort S | Depends: R38 | Stephen input: -

Why: no pricing page exists; monetization is invisible. Payments are stubbed (R28), so the page should sell the tiers and capture intent now.

Steps:
1. `/bang-gia` page: Free vs Premium (29k) vs Pro (79k VND/thang) per the PRD tiers; feature matrix from TASK-BILL-005 gating list; FAQ on billing.
2. Premium CTA opens a waitlist modal (email + optional Zalo number) writing to the leads store; tag source=pricing.
3. R40 events: view, tier-click, waitlist-submit.
4. When R28 lands later, the CTA flips to checkout - leave the flag in code.

Acceptance: page live, waitlist rows written, events firing.

Verify: lead row sample + screenshot in ledger.

Human review: Stephen confirms the tier prices still match his current pricing intent before indexing.

---

## R40 - analytics + funnel events + UTM discipline

Wave 1 | Effort S | Depends: - | Stephen input: decision (GA4 free vs Plausible paid ~9 EUR/mo)

Why: zero analytics anywhere in web/ (verified). Every growth task after this is blind guessing without it.

Steps:
1. Stephen decision: GA4 (free, heavier, consent banner needed) vs Plausible (paid, cookieless, lighter consent story - recommended for the trust brand). Record in ledger; default to Plausible if no answer within a wave.
2. Add the script via a typed analytics module (`web/lib/analytics.ts`) with a single `track(event, props)` API; no vendor calls scattered in components.
3. Canonical event set (document in the module): `ext_install_click`, `signup_start`, `signup_done`, `first_track`, `first_alert`, `waitlist_submit`, `email_track_submit`, `upgrade_intent`, `affiliate_click`.
4. UTM convention doc (`docs/conventions/UTM.md`): source/medium/campaign taxonomy for R50-R53 channels.
5. Consent alignment: if GA4, gate behind the consent banner (R32/R34); Plausible runs consentless with a disclosure line in the privacy policy.
6. Server-side counterpart: gateway/services already emit obs metrics; do not duplicate - product analytics only in web.

Acceptance: events visible in the analytics dashboard from a staging click-through of the whole funnel.

Verify: dashboard screenshot + event log in ledger.

Human review: click the funnel once and watch events arrive; approve the taxonomy doc.

---

## R41 - programmatic SEO: public price-history pages (pilot)

Wave 2 | Effort L | Depends: R24 | Stephen input: -

Why: the single highest-traffic lever in the plan. Every scraped product can be an indexable page targeting "giá [product]" and "lịch sử giá [product]" long-tails - the Keepa/BeeCost growth engine. Requires live data (R24).

Steps:
1. Route `web/app/(marketing)/san-pham/[slug]/page.tsx`: SSR/ISR page per tracked product - name, current price, min/median (90d), sparkline chart, fake-sale verdict badge, platform links (user-initiated affiliate per R29 guardrail), "Theo dõi giá" CTA (signup or R46 email form).
2. Slug strategy: `ten-san-pham-p<product_id>`; canonical URLs; noindex products with < 7 days of history (honesty per R27).
3. ISR revalidation on the daily aggregate cadence; page reads from `price_daily` continuous aggregate (fast path per NFR).
4. Sitemap: chunked sitemap index generated from the DB (extend `web/app/sitemap.ts`); cap pilot at 10k products.
5. Internal linking: category hub pages linking product pages; product pages link category + related products.
6. JSON-LD Product + Offer with price history where valid.
7. Guard: bot-safe (no auth), rate-limit friendly, cached at the gateway/CDN layer.
8. R40 events: page views land in analytics with product dimension; CTA conversions tagged source=pseo.

Acceptance: 5-10k product pages live in staging, indexed sitemap submitted (Search Console once domain exists), CTR events flowing, p95 page render within NFR chart target.

Verify: sitemap sample, 3 rendered pages, Search Console submission proof in ledger.

Human review: spot-check 5 pages for copy quality and honest data depth; confirm noindex logic on thin pages; approve opening the crawl tap (robots).

---

## R42 - keyword cluster expansion (30-50 pages)

Wave 3 | Effort M (ongoing) | Depends: R38 | Stephen input: -

Why: `web/lib/seo/keywords.ts` holds exactly 4 pages sharing one template, with no internal links between them. Thin coverage of a keyword space the PRD explicitly targets.

Steps:
1. Keyword map first (`docs/growth/KEYWORD-MAP.md`): clusters - sale calendar (lịch sale 7.7/8.8/9.9/10.10/11.11/12.12 + per-month pages), coupon/freeship (mã freeship hôm nay, mã giảm giá per platform), tactics (cách săn xu, săn deal flash sale), verdicts (sale thật hay ảo, giá gốc là gì), comparisons (feeds R44). 30-50 targets with intent notes.
2. Upgrade the page template: unique intro per page (no shared boilerplate), FAQ block with JSON-LD FAQPage, related-pages block (internal links within cluster + to R41 product pages + landing).
3. Sale-calendar pages get structured dates and a reminder CTA (R43's calendar tool).
4. Cadence: ship 10 pages per batch with real content, measure via R40 before the next batch.

Acceptance: first 10 new pages live with unique content and interlinks; impressions visible in Search Console within 2 weeks of indexing.

Verify: page list + interlink graph sketch in ledger.

Human review: read 3 random pages - would a real deal-hunter find them useful? Reject template-smell.

---

## R43 - lead-magnet tools (fake-sale checker first)

Wave 2 | Effort M | Depends: R38 | Stephen input: -

Why: free tools convert cold SEO traffic into leads and demos. The APIs already exist (price history, median, verdict logic in dealsvc).

Steps:
1. Tool 1 - fake-sale checker at `/kiem-tra-sale-ao`: paste a product URL -> resolve platform + product -> if tracked, instant verdict (current vs 90d median, badge + chart snippet); if untracked, enqueue (R27 demand lane) + capture email for "verdict when ready" (writes lead, tags source=tool).
2. Tool 2 - sale calendar at `/lich-sale` upgrade: next big sale countdown + email/Zalo reminder subscribe.
3. Tool 3 (optional, if cart API surface allows) - coin/voucher stack calculator; skip if it needs auth'd platform data.
4. Rate-limit the resolver at the gateway; cache verdicts.
5. R40 events per tool: submit, verdict-shown, lead-captured.

Acceptance: checker returns a live verdict for a tracked product and captures a lead for an untracked one.

Verify: two transcripts (tracked/untracked) + lead rows in ledger.

Human review: paste 5 real product URLs yourself; check verdict credibility and VN wording.

---

## R44 - comparison pages

Wave 3 | Effort S | Depends: R38 | Stephen input: -

Why: buyers search the incumbent's name. Honest comparison pages capture that intent (BeeCost lacks TikTok Shop; SănDeal forecasts, BeeCost only shows history).

Steps:
1. `/so-sanh/sandeal-vs-beecost`: factual feature table (platforms incl. TikTok Shop, price history, forecast/p_bottom, cart optimizer, open source, PDPL stance), screenshots, no disparagement - factual rows only.
2. `/thay-the-honey` (Honey alternative for VN): the trust story with sources (Rakuten/Impact.com removals), how SănDeal differs by architecture (link R35).
3. JSON-LD, interlinks with R42 cluster, R40 events on CTAs.

Acceptance: both pages live, factually accurate at publish date, sources linked.

Verify: URLs + fact-check list in ledger.

Human review: verify every comparison row personally - a wrong row invites a public callout.

---

## R45 - onboarding to first-alert aha moment

Wave 2 | Effort M | Depends: R38 | Stephen input: -

Why: after signup the dashboard is empty; there is no guided path to the product's aha (seeing a price history + setting an alert). Activation dies here.

Steps:
1. Post-signup flow: one screen - "Dán link sản phẩm Shopee/TikTok/Lazada bất kỳ" -> resolve -> show chart (or honest thin-history state per R27) -> one-click "Báo tôi khi chạm đáy" default alert rule.
2. Empty states on dashboard/wishlist/alerts all deep-link back into this flow.
3. Measure `first_track` and `first_alert` (R40); target time-to-first-alert < 2 minutes.
4. Jest tests for the flow states.

Acceptance: a new staging account reaches first alert in under 2 minutes without help.

Verify: recorded funnel timings for 3 fresh accounts in ledger.

Human review: run it once yourself on a fresh account; note any hesitation point.

---

## R46 - track-by-email without an account

Wave 2 | Effort M | Depends: R23 | Stephen input: -

Why: the lowest-friction lead capture: product URL + email on any public page -> alert on drop -> upgrade path to account. Feeds the lead store and the R27 demand queue.

Steps:
1. Public form component (landing, product pages, checker tool): URL + email; double-opt-in confirmation email (R23 sender); creates a lightweight lead-track row (no full account).
2. Alert fire for lead-tracks routes via email with an upgrade CTA ("tạo tài khoản để theo dõi nhiều sản phẩm + chọn kênh Zalo").
3. Rate limits + disposable-email guard at the gateway; unsubscribe honored (List-Unsubscribe + link).
4. Conversion path: signup with the same email adopts existing lead-tracks.
5. R40 events: submit, confirm, alert-sent, lead-to-account conversion.

Acceptance: end-to-end on staging - submit, confirm, price-drop email received, upgrade adopts the track.

Verify: full transcript in ledger.

Human review: run the flow with a personal email; check the emails do not smell like spam (sender, tone, unsubscribe).

---

## R47 - blog + changelog + RSS

Wave 3 | Effort S (setup) | Depends: R38 | Stephen input: -

Why: long-tail content and launch stories need a home; a changelog shows the product is alive (trust signal for both users and press).

Steps:
1. `/blog` with MDX or file-based content in `(marketing)`; author/date/tags; RSS feed; JSON-LD Article.
2. `/changelog` fed from a simple markdown file per release (hook into R15's deploy notes if easy).
3. Two seed posts: "Sale thật hay sale ảo - cách kiểm tra trong 10 giây" (links R43 tool) and the SănDeal introduction/launch post (feeds R53).
4. Interlink with R42 cluster; R40 events.

Acceptance: blog live with 2 posts, RSS validates, changelog shows the current release.

Verify: URLs + RSS validator output in ledger.

Human review: approve the two posts' voice - they set the content bar.

---

## R48 - web performance + a11y pass, Lighthouse in CI

Wave 3 | Effort M | Depends: R38 | Stephen input: -

Why: `next.config.mjs` is minimal (no image config); chart pages lack skeletons; no Lighthouse gate exists (the landing-page repo already runs one - reuse the pattern).

Steps:
1. next.config: image domains/formats, compression; font loading via `next/font`; bundle-analyze the app group and trim heavy imports (recharts only on chart routes via dynamic import).
2. Loading skeletons for chart/wishlist/alerts; error boundaries with retry (the chart fetch currently shows a bare VN error line).
3. A11y sweep on public pages: landmarks, contrast, focus order, alt text; axe checks in jest for landing + product page templates.
4. Lighthouse CI (mobile) on landing, one product page, one keyword page: perf >= 85, SEO >= 95, a11y >= 90 budgets; wire into CI as a soft gate first, hard gate after two green weeks.

Acceptance: budgets met on the three page types; CI publishes scores per PR.

Verify: Lighthouse CI report links in ledger.

Human review: open the product page on a mid-range Android over 4G once - the target user's reality.
