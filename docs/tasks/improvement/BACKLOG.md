# Improvement backlog - status index

Source report: `docs/strategy/shopass-strengthening-audit-2026-07-06.md`. Cards live in `TASKS-*.md`. This table is the single source of truth for status; agents update the Status column here and append evidence to `LEDGER.md`.

Stephen-input legend: `-` none, `decision` (choice/approval), `account` (external account signup), `creds` (secret/credential handover), `budget` (spend approval), `outreach` (human contact work).

## Wave 1 - unblock and harden (days 0-30)

| ID | Title | Card | Effort | Depends on | Stephen input | Status |
|----|-------|------|--------|-----------|---------------|--------|
| R1 | Wire gateway into compose, kill X-User-Id trust | A | M | - | - | done (TASK-INFRA-006) |
| R2 | Purge node_modules from git | A | S | - | - | done (already untracked; node_modules/ in .gitignore) |
| R3 | Add LICENSE (proprietary core + OSS extension) | A | S | - | decision | done (PR #60) |
| R4 | Real root metadata, kill "Create Next App" | A | S | - | - | done (web/app/layout.tsx + root-metadata.test.ts) |
| R5 | Auth guard for the whole (app) route group | A | S | - | - | done (middleware + middleware-guard.test.ts) |
| R6 | CSRF + origin checks + login rate limit | A | M | R1 | - | done |
| R7 | Session restore on page reload | A | S | - | - | done (PR #60) |
| R8 | Extension env-config endpoints + domain decision | A | S | - | decision | done (PR #60) |
| R9 | Guard demo seed against prod | A | S | - | - | done (PR #60) |
| R10 | Go toolchain alignment + govulncheck in CI | A | S | - | - | done (PR #61) |
| R11 | TLS + reverse proxy (Caddy) | B | M | R1 | creds (domain DNS) | needs_stephen |
| R12 | Automated backups + restore drill | B | M | - | account, creds (object storage) | needs_stephen |
| R13 | Prometheus + alert rules for declared NFRs | B | M | - | - | done (PR #62) |
| R14 | Centralized logs (Loki) | B | M | - | - | done (PR #62) |
| R15 | CI/CD: GHCR image publish + SSH deploy | B | M | R2 | creds (VPS, GHCR secrets) | needs_stephen |
| R16 | Zero-downtime deploys + migration guards | B | M | R15 | - | blocked |
| R17 | Production scheduling for scrape + forecast jobs | B | S | - | - | done (PR #62) |
| R23 | Zalo ZNS + email senders beside FCM | C | M | - | account, creds (Zalo OA, SMTP) | needs_stephen |
| R34 | Public privacy policy + terms pages (VN + EN) | D | S | - | - | done (PR #62) |
| R40 | Analytics + funnel events + UTM discipline | E | S | - | decision (GA4 vs Plausible) | needs_stephen |
| R49 | Extension store kit (icons, locales, listing, screenshots) | F | M | R3, R34 | account (Chrome dev, Cốc Cốc), budget ($5) | needs_stephen |

## Wave 2 - real data and first users (days 31-60)

| ID | Title | Card | Effort | Depends on | Stephen input | Status |
|----|-------|------|--------|-----------|---------------|--------|
| R18 | Wire BFF behind gateway or remove dead path | B | M | R1 | - | done (PR #63) |
| R19 | Data retention + chunk policy decision | B | S | - | decision | todo |
| R24 | Battle-test live scraping behind residential proxy | C | L | R17 | budget (proxy), creds | todo |
| R25 | Pluggable CAPTCHA path (manual queue first) | C | M | R24 | budget (optional solver) | todo |
| R26 | ML model versioning + evaluation gate | C | M | - | - | done (PR #65) |
| R27 | Cold-start backfill + honest history-depth UI | C | M | R24 | - | todo |
| R30 | Google OAuth live end-to-end test | C | S | R11 | creds (Google client) | todo |
| R31 | DNR rules integration tests | C | S | - | - | done (PR #63) |
| R35 | Transparency page (anti-Honey positioning) | D | S | R34 | - | done (PR #63) |
| R36 | Open-source the extension publicly | D | M | R2, R3, R31 | decision | todo |
| R38 | Real landing page (hero, demo, trust strip, FAQ) | E | M | R4 | - | done (PR #65) |
| R39 | Pricing page + Premium waitlist capture | E | S | R38 | - | done |
| R41 | Programmatic SEO: public price-history pages (pilot) | E | L | R24 | - | todo |
| R43 | Lead-magnet tools (fake-sale checker first) | E | M | R38 | - | done |
| R45 | Onboarding to first-alert aha moment | E | M | R38 | - | done |
| R46 | Track-by-email without an account | E | M | R23 | - | todo |
| R55 | Closed beta program (50-200 users) | F | M | R11, R23 | outreach | todo |

## Wave 3 - launch and monetize (days 61-90)

| ID | Title | Card | Effort | Depends on | Stephen input | Status |
|----|-------|------|--------|-----------|---------------|--------|
| R20 | k6 load test gate vs NFR p95 targets | B | M | R13 | - | todo |
| R21 | Dependency, image scanning, SBOM | B | S | R15 | - | todo |
| R22 | Nightly end-to-end smoke in CI | B | M | R15 | - | todo |
| R28 | Payments sandbox: MoMo/ZaloPay/VNPay real flows | C | L | R11 | account (merchant), creds | todo |
| R29 | Affiliate programs live + attribution logging | C | M | R24 | account (Shopee/TikTok affiliate) | todo |
| R32 | PDPL minimum set: consent, DSAR, breach runbook | D | L | - | - | todo |
| R33 | CI compliance gates (no-cleartext, consent, DPIA) | D | M | R32 | - | todo |
| R37 | Scraping legal posture memo | D | S | - | decision (counsel review) | todo |
| R42 | Keyword cluster expansion (30-50 pages) | E | M | R38 | - | done |
| R44 | Comparison pages (vs BeeCost, Honey alternative) | E | S | R38 | - | done |
| R47 | Blog + changelog + RSS | E | S | R38 | - | done |
| R48 | Web performance + a11y pass, Lighthouse in CI | E | M | R38 | - | todo |
| R50 | Zalo OA + Telegram public deal channels | F | M | R23, R41 | account | todo |
| R51 | TikTok content engine (chart-verdict shorts) | F | M | R41 | account, outreach | todo |
| R52 | Referral program UI surface | F | M | R38 | - | todo |
| R53 | Launch calendar + press kit (timed to double-day sale) | F | M | R38, R49 | outreach | todo |
| R54 | B2B lead magnet: VN price index report (gated) | F | M | R24 | - | todo |
| R56 | Partnership + KOC affiliate probes | F | M | R52 | outreach | todo |
| R57 | Support + feedback rails (Zalo OA inbox, in-app form, roadmap page) | F | S | R23 | - | todo |
| R58 | Brand consistency sweep + handle registration | F | S | R8 | decision, account | todo |
