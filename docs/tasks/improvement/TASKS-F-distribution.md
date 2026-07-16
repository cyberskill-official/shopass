# Part F - distribution, community, B2B leads (R49-R58)

The extension is the front door; channels amplify what R41-R47 build; B2B leads come from data exhaust.

---

## R49 - extension store kit

Wave 1 | Effort M | Depends: R3, R34 | Stephen input: account (Chrome Web Store dev $5, Edge, Cốc Cốc) + budget

Why: `extension/manifest.json` has no `icons` field and no `default_locale`; no privacy policy, screenshots, or listing copy exist. The product's main distribution surface cannot ship at all today.

Steps:
1. Icons: 16/32/48/128 PNG from the SănDeal mark (if no mark exists yet, flag to R58 and use a clean wordmark placeholder Stephen approves); add `icons` to manifest + action icons.
2. Locales: `_locales/vi/messages.json` + `_locales/en/`, `default_locale: "vi"`; move name/description strings into messages.
3. Listing pack in `extension/store/`: VN + EN descriptions (short + full), 5 screenshots at 1280x800 (staged flows: price history on Shopee, fake-sale badge, TikTok Shop compare, alert setup, consent screen), promo tile 440x280.
4. Data-disclosure worksheet: Chrome's data-use form answers derived from the consent scopes (cart, vouchers, sync) - store as `extension/store/DATA-DISCLOSURE.md`; privacy-policy URL from R34.
5. Permission justifications paragraph per host permission.
6. Dry-run: `web-ext lint` (or Chrome's validator) clean; package zip reproducible (ties R36).
7. Stephen asks: create the dev accounts (Chrome $5 one-time, Edge free, Cốc Cốc store contact), then submit wave-2 end.

Acceptance: a submission-ready zip + complete listing pack; validator clean; only account creation stands between the repo and review queues.

Verify: zip manifest listing + validator output + screenshot set in ledger.

Human review: Stephen approves icons/branding and creates accounts; reviewer checks the disclosure worksheet matches the consent scopes exactly (mismatch = store rejection).

---

## R50 - Zalo OA + Telegram public deal channels

Wave 3 | Effort M | Depends: R23, R41 | Stephen input: account (public OA, Telegram channel)

Why: your own scoring already finds "real drops" nightly; publishing the top N into public channels turns the alert engine into an acquisition channel where every post links a product page (R41).

Steps:
1. Channel-publisher job in notifsvc (or a small worker): daily top-10 real drops (dedupe against yesterday, category diversity rule), formatted VN post with price-was/now, verdict badge, product-page link with UTM (R40 taxonomy).
2. Telegram first (bot API, trivial); Zalo OA broadcast second (respect OA broadcast rules/quotas).
3. Content guardrails: only verdict-passing drops; no affiliate params in channel links unless clicked through the product page (keep the R29 guardrail honest).
4. Ops: publish log table, kill switch env var, alert on publish failure.

Acceptance: 7 consecutive daily posts on both channels from the real pipeline, with click-through visible in analytics.

Verify: channel links + publish log + UTM click sample in ledger.

Human review: Stephen creates the public channels; read a week of posts - would you subscribe? Cut anything spammy.

---

## R51 - TikTok content engine

Wave 3 | Effort M (ongoing) | Depends: R41 | Stephen input: account + outreach (creator time or KOC)

Why: the audience that made TikTok Shop pass Shopee lives on TikTok; "sale thật hay ảo" verdict shorts using your own charts are native content with built-in proof.

Steps:
1. Format template: 15-30s short - product hook, chart zoom (screen capture of the product page), verdict reveal, "link in bio / comment" to the product page.
2. Produce a starter batch of 10 around the next double-day sale (calendar from R42); tooling script to export chart clips (even simple screen-record checklist is fine - document it in `docs/growth/TIKTOK-PLAYBOOK.md`).
3. Bio/link strategy: link-in-bio page (`/tiktok`) listing today's verdicts (auto from R50's publish log).
4. Measure with UTM + R40; iterate hooks by retention stats.
5. Stephen asks: who fronts the content (his time, a hire, or KOC per R56) + the TikTok account handle (R58).

Acceptance: 10 shorts published around one sale event; `/tiktok` page live; traffic attributable in analytics.

Verify: post links + analytics screenshot in ledger.

Human review: Stephen picks the face/voice of the channel - this one is a human-led task with agent support, not the reverse.

---

## R52 - referral program UI surface

Wave 3 | Effort M | Depends: R38 | Stephen input: -

Why: TASK-BILL-004 referral logic (codes, attribution constraints) exists server-side with no user-facing surface - a built growth loop that nobody can use.

Steps:
1. Dashboard block: your code, share buttons (Zalo/copy), reward status per the existing attribution rules; T&C line (R34 terms addendum).
2. Alert-email/Zalo footers carry the referral line ("Mời bạn - cả hai nhận Premium") once rewards are defined.
3. Stephen decision embedded: confirm the reward (e.g., 1 month Premium both sides) - default proposal in ledger if unanswered.
4. R40 events: share-click, referred-signup, reward-granted; fraud guard sanity (self-referral blocked - test the existing constraint).

Acceptance: a referred staging signup credits both accounts per the rules and shows in both dashboards.

Verify: two-account transcript in ledger.

Human review: approve the reward economics; try a self-referral and confirm it fails.

---

## R53 - launch calendar + press kit

Wave 3 | Effort M | Depends: R38, R49 | Stephen input: outreach (press, communities)

Why: BeeCost earned coverage on GenK, Tiên Phong, Tinh tế - the channel is proven for this category. The Honey story gives a fresh angle: "the trust-first deal extension, open source, made in VN". Unprepared launches waste the one news cycle.

Steps:
1. Press kit in `docs/growth/PRESS-KIT/`: VN press release (angle: TikTok Shop price history first + open-source trust post-Honey), founder quote, product screenshots, logo pack, fact sheet (platforms, data points tracked, PDPL stance), contact.
2. Target list with per-outlet pitch notes: GenK, Tinh tế, Tiên Phong, VnExpress Số hóa, CafeBiz; communities: VOZ, Facebook deal groups (Nghiện Shopee etc.), Reddit r/Vietnam; EN: Product Hunt for the open-source extension (R36).
3. Timing plan: T-14 to T+7 around the chosen double-day sale (first viable after wave-2 exit: 8.8 or 9.9); coordinate R50/R51 output that week.
4. Beta testimonials (R55) folded into the kit.
5. Stephen owns sending; agent drafts everything sendable.

Acceptance: kit complete and sendable; calendar with owners and dates agreed.

Verify: kit files + target list in ledger.

Human review: Stephen edits the founder quote and approves the angle; dry-read the release as a journalist - is there a story in the first two lines?

---

## R54 - B2B lead magnet: VN price index report

Wave 3 | Effort M | Depends: R24 | Stephen input: -

Why: the PRD's phase-3 B2B data revenue starts as a phase-1 lead magnet: brands/sellers trade work emails for market data nobody else publishes (real-vs-fake discount rates by category, platform price-war stats).

Steps:
1. Report v1 (quarterly): aggregate-only stats from the snapshot corpus - median discount depth by category, % listings failing the fake-sale test, Shopee vs TikTok Shop price-gap distribution; strict k-anonymity (no shop-level data) - document the aggregation rules (PDPL-clean, R32 inventory).
2. Landing `/bao-cao-thi-truong`: preview charts + gated PDF download (work email form, tags source=b2b).
3. PDF build scripted (repeatable next quarter); VN primary, EN executive summary.
4. Follow-up asset: seller-dashboard waitlist question in the download form ("Bạn muốn theo dõi giá đối thủ?" yes -> waitlist row).
5. R40 events + a `b2b_leads` view for Stephen.

Acceptance: report downloadable behind the gate; 1 lead flows through to the store; aggregation rules documented.

Verify: PDF + lead row + rules doc in ledger.

Human review: check no stat could identify a single shop; Stephen approves the numbers before publication (they will be quoted).

---

## R55 - closed beta program

Wave 2 | Effort M | Depends: R11, R23 | Stephen input: outreach (recruit 50-200 users)

Why: live scraping (R24), alert quality, and onboarding (R45) need real users before the press cycle (R53); testimonials feed the landing trust strip (R38).

Steps:
1. Beta scaffold: invite-code gate on signup (env-flagged), `/beta` page explaining the deal (free Premium during beta for feedback), feedback form in-app (R57's form, earlier if trivial).
2. Instrument the beta cohort in analytics (R40 property) and a weekly usage digest (tracked products, alerts fired, WAU) posted where Stephen reads it.
3. Structured feedback: 3 questions at day 7 (in-product prompt) - what's confusing, what's missing, would you pay 29k?
4. Testimonial consent checkbox on the feedback form (name + quote usage, PDPL-clean).
5. Stephen asks: recruit via VOZ/FB deal groups/personal network; approve the beta-perk terms.

Acceptance: 50+ activated beta users, weekly digest flowing, >= 5 usable testimonials with consent.

Verify: cohort dashboard + digest sample + testimonial rows in ledger.

Human review: Stephen reads week-1 feedback raw (unfiltered) and marks the top 3 fixes as new backlog entries.

---

## R56 - partnership + KOC affiliate probes

Wave 3 | Effort M | Depends: R52 | Stephen input: outreach

Why: non-competing products share the exact audience (cashback apps, personal-finance communities); KOCs paid via your own referral codes (R52) scale content without upfront fees.

Steps:
1. One-pager for partners (`docs/growth/PARTNER-ONE-PAGER.md`, VN): what SănDeal does, audience, integration/co-promo options (deal widget, co-branded sale-calendar embed, channel cross-posts).
2. KOC kit: referral-code onboarding doc + content examples (from R51's playbook) + reward table.
3. Target list: 10 partner candidates + 20 KOC candidates with contact notes; agent researches, Stephen contacts.
4. Track partner-attributed signups via UTM + referral codes.

Acceptance: kits complete; target list delivered; first 3 conversations logged (Stephen's outcome, not the agent's).

Verify: docs + list in ledger.

Human review: Stephen prunes the target list before any outreach happens under the company name.

---

## R57 - support + feedback rails

Wave 3 | Effort S-M | Depends: R23 | Stephen input: -

Why: launch without a support channel burns trust exactly where the brand claims it; feedback needs a queue, not a DM inbox.

Steps:
1. Support channel: Zalo OA inbox as primary (link in footer + app shell), support@ alias as secondary; response-time promise on the contact page (48h beta, 24h paid).
2. In-app feedback/bug form writing to a tracked table + Telegram notification (R13's alert bot); include app version + page context automatically.
3. Public roadmap page (`/lo-trinh`): now/next/later from this backlog's public-safe subset; changelog link (R47).
4. Canned-answer doc for the top 10 expected questions (VN) in `docs/growth/SUPPORT-FAQ.md`.

Acceptance: a test feedback submission lands in the queue + Telegram; roadmap page live; footer links present.

Verify: submission transcript + page URLs in ledger.

Human review: send one real support message via Zalo and time the visibility path; approve the response-time promise before publishing it.

---

## R58 - brand consistency sweep + handle registration

Wave 3 | Effort S | Depends: R8 | Stephen input: decision (final name) + account (handles)

Why: three names coexist today (repo "shopass", product "SănDeal", domain "sandeal.vn" hardcoded in the extension). Store listings, press, and channels will freeze whichever name ships - drift after that is expensive.

Steps:
1. Stephen decision: confirm SănDeal + final domain (from R8).
2. Sweep: README, web metadata (R4), extension name/locales (R49), OG images, press kit (R53), policy pages (R34) - one name, one domain, one logo everywhere; grep for stragglers ("shopass" stays as internal repo name only, noted in README).
3. Register handles: TikTok, Facebook, Telegram, Zalo OA name, GitHub org repo name (R36), X if desired; record credentials in Stephen's vault (not the repo).
4. Simple brand sheet `docs/growth/BRAND-SHEET.md`: logo files, colors, name usage rules (reference the CyberSkill design system only as the parent brand).

Acceptance: grep across repo + live pages shows one consistent public name/domain; handle list registered and recorded.

Verify: grep output + handle checklist in ledger.

Human review: Stephen owns the final call; check the store listing drafts (R49) match before submission.
