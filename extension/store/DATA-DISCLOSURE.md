# Chrome data-use / disclosure worksheet (R49)

Privacy policy URL: https://shopass.cyberskill.world/chinh-sach-bao-mat  
Consent purposes: see monorepo `docs/compliance/CONSENT-SURFACES.md` and `DATA-INVENTORY.md`.

## Data collected / accessed

| Category | Collected? | Purpose | Consent / note |
|----------|------------|---------|----------------|
| Personally identifiable info (account email) | Via Shopass web login, not scraped from marketplace | Account | Account signup |
| Health | No | — | — |
| Financial / payment | No in extension | — | Billing is web-only |
| Authentication info (marketplace passwords/tokens) | **Never** — must not leave client | — | Trust invariant |
| Personal communications | No | — | — |
| Location | No | — | — |
| Web history | No (only matched host pages when script runs) | Price helpers | Host permissions below |
| User activity (cart/voucher on marketplace pages) | Yes, **only after** `cart_read` consent | Optimize vouchers / user-requested assist | Extension consent store → backend when logged in |
| Website content (public listing price signals) | Yes on matched product pages | Price capture for tracking | User-initiated track |

## Host permission justifications

| Pattern | Why |
|---------|-----|
| `https://shopee.vn/*` | Content script + price/cart helpers on Shopee VN |
| `https://*.tiktok.com/*` | TikTok Shop content script |
| `https://www.lazada.vn/*` | Lazada VN content script |
| `https://shopass.cyberskill.world/*` | Same-origin Shopass API (`/v1/*`) for sync/alerts/consent |
| `http://127.0.0.1:8080/*` | Local/dev gateway only (omit from store build if packaging prod-only) |

## Remote code

No remote code execution. All JS ships in the package (MV3).

## Affiliate

User-initiated deep links only; see `/minh-bach` and Chrome affiliate checklist in `docs/compliance/CHROME-WEBSTORE-AFFILIATE-CHECKLIST.md`.
