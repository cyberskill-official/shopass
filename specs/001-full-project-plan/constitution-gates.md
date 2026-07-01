# Constitution Gates — SănDeal

Active constitution rules derived from `docs/feature-requests/SHIP-GUIDE.md`. These 9 non-negotiable invariants MUST be preserved by every FR implementation. Violation is a hard failure.

## 1. Security and PDPL
- No cleartext credentials anywhere (code, logs, DB, config).
- Platform session tokens/cookies NEVER leave the client.
- `platform_account.ext_user_ref` is an anonymous identifier, not a token.
- Passwords hashed with argon2id in PHC format (parameters stored with hash).
- Secrets in Vault / AWS Secrets Manager only.
- Personal data processing requires PDPL consent (voluntary, specific, single-purpose; silence != consent).
- DPIA submitted within 60 days.
- Breach notification within 72 hours.
- Penalties: up to 5% revenue / 10x illicit profit / 3 billion VND.

## 2. Post-Honey Affiliate Ethics
- Affiliate deep links ONLY created when user ACTIVELY clicks.
- Disclosure clearly shown.
- No cookie-stuffing, dropping, auto-redirect, or pop-under.
- Extension scraping MUST NOT be used to attach affiliate tags.
- Chrome Web Store policy (updated 3/2025, enforced 10/06/2025) must be followed.
- Cashback only on confirmed conversion with hold-then-release and delayed payout.

## 3. Money — BIGINT VND
- Every money column is `BIGINT` in VND.
- No float or numeric types for monetary values.
- Avoids rounding errors in percentage calculations (fake sale, optimizer).

## 4. Price — Delta-Only Time-Series
- `price_snapshot` records delta-only (only INSERT when price changes).
- `price_snapshot` is a TimescaleDB hypertable with compression.
- `price_daily` is a continuous aggregate.
- Charts/history read from `price_daily`, never scan raw snapshots.

## 5. Extension MV3
- Service worker is ephemeral (no state in global variables; use `chrome.storage`).
- `chrome.alarms` minimum 30s interval; no `setInterval`.
- Heavy tasks delegated to backend.
- Offscreen API for DOM/clipboard access.
- `declarativeNetRequest` replaces webRequest blocking.
- Content script ONLY sends productId/price/qty to backend.
- No cookie/token in content script payloads.
- Prefer DOM reading for TikTok Shop (avoid msToken/_signature/X-Bogus) and Lazada (Akamai).

## 6. Notification — Cost Model
- Priority order: push > email > SMS.
- FCM HTTP v1, quota 600,000 msg/min/project.
- Handle 429 RESOURCE_EXHAUSTED with backoff.
- Midnight spike flattened with jitter + per-minute bucket.
- Push channel split by `user_channel_token.platform`: FCM for `platform IN ('android','web')`, APNs for `platform='ios'`.

## 7. Data Model — One Table One Owner
- Each table has exactly one owner FR (see DATA-MODEL.md).
- Other modules reference via FK or extend via ALTER TABLE.
- Never re-create a table owned by another FR.
- Extensions use ALTER TABLE (e.g., AUTH-001 adds `pwd_hash` to INFRA-002's `app_user`).

## 8. Per-Country Gating
- VN first; MY/PH have stacking voucher rules (freeship grouped by platform).
- Data protection laws differ per country.
- Read CountryPolicy; default restrictive (no-stack) for unconfigured countries.

## 9. Scraping
- Residential proxy required (datacenter ineffective against Cloudflare/Akamai).
- Random pacing + jitter.
- DOM drift monitoring required.
- Coin/voucher automation is high-risk: only checklist reminders + user-initiated auto-test.
- Sleep 2.5-5s between actions, revert on failure, never auto-checkout.
