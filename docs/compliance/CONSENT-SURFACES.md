# Consent surfaces (CI-tracked)

Table-driven consent coverage for R33. Every personal-data collecting surface must
appear here with a PDPL purpose and an enforcement status.

Status values:

- `api` — server handler must call `consent.IsAllowed` (or be waived)
- `client` — client-only processing; server must not receive the raw payload
- `deferred` — known gap; requires `waiver: reviewed:<who>` until wired

| id | surface | purpose | status | code_anchor | waiver |
|----|---------|---------|--------|-------------|--------|
| CS-001 | Extension cart/voucher read | `cart_read` | client | `extension/src/consent/consent-store.ts` | |
| CS-002 | Track product / wishlist / alerts | `price_tracking` | deferred | `services/track/cmd/tracksvc` | reviewed:R33-bootstrap |
| CS-003 | Marketing / sale notifications | `marketing_notification` | deferred | `services/notif` | reviewed:R33-bootstrap |
| CS-004 | Aggregated B2B analytics export | `analytics_b2b` | deferred | `services/b2b` | reviewed:R33-bootstrap |
| CS-005 | Consent grant/withdraw API | `cart_read` | api | `services/comply/internal/api/consent.go` | |

## Purpose allowlist

Must match `services/comply/internal/consent/types.go`:

- `cart_read`
- `price_tracking`
- `marketing_notification`
- `analytics_b2b`
