# Data inventory

Minimum inventory for the R32 PDPL set. Retention periods are product defaults
pending counsel review and production policy approval.

| Store / field | Data class | Purpose | Retention / erasure rule |
|---|---|---|---|
| `app_user.id` | Account identifier | Join user-owned data across services | Retain as tombstone key after erase |
| `app_user.email` | Personal data | Login, account contact, DSAR identity | Anonymize to `deleted_<id>@tombstone.local` on erase |
| `app_user.phone` | Personal data | Account recovery / contact when present | Set `NULL` on erase |
| `app_user.display_name` | Personal data | User profile display | Set `NULL` on erase |
| `app_user.locale` | Preference | Localization | Retain unless user deletes account; not sensitive alone |
| `consent_policy.*` | Compliance record | Versioned consent text and purpose basis | Immutable; retain indefinitely for legal proof |
| `consent_record.user_id` | Compliance record | Reconstruct grant/withdraw history | Retain after erase as legal evidence |
| `consent_record.purpose_key` | Compliance record | Single-purpose consent basis | Retain after erase |
| `consent_record.ip` | Personal data / audit metadata | Prove consent event context | Retain with consent record; minimize access |
| `consent_record.user_agent` | Personal data / audit metadata | Prove consent event context | Retain with consent record; minimize access |
| `user_tracked_product.user_id` | Personal data | Price tracking by account | Hard-delete on erase |
| `tracked_product.*` | Product data | Shared product catalog and price tracking | Retain; not user-specific without join table |
| `wishlist.user_id` | Personal data | User wishlist | Hard-delete on erase |
| `wishlist_item.*` | Personal data by relation | Wishlist product and target price | Hard-delete through wishlist on erase |
| `alert_rule.user_id` | Personal data | User alert configuration | Hard-delete on erase |
| `alert.payload` | Potential personal data | Notification/audit payload for alert firing | Hard-delete through user alert rules on erase |
| `alert_fired_state.*` | Personal data by relation | Alert de-duplication state | Hard-delete through user alert rules on erase |
| `subscription.user_id` | Accounting / entitlement data | Paid plan entitlement and legal accounting | Retain; keep account key, remove payment PII |
| `payment.order_ref` | Personal data when `order_<userID>_*` | Payment reconciliation | Replace with `erased_<payment_id>` on erase |
| `payment.transaction_id` | Payment PII / partner identifier | Payment reconciliation and support | Set `NULL` on erase when tied to user |
| `dsar_request.user_id` | Compliance record | Track data-rights requests and SLA | Retain as legal evidence |
| `dsar_request.kind/status/timestamps` | Compliance record | Prove DSAR handling | Retain as legal evidence |
| `breach_incident.*` | Compliance/security record | 72h breach workflow and evidence timeline | Retain as incident evidence |
| `processing_activity.*` | Compliance record | DPIA/TIA processing inventory | Retain while processing exists + audit horizon |
| `dpia.*` / `tia.*` | Compliance record | Risk assessment and cross-border safeguards | Retain by version; do not overwrite history |

## Processing purposes

| Purpose key | Description | Primary stores |
|---|---|---|
| `cart_read` | Extension reads cart/voucher data on the client for user-requested optimization | `consent_record`; no server token storage |
| `price_tracking` | Track selected products, wishlists, and alert rules | `user_tracked_product`, `wishlist`, `alert_rule` |
| `marketing_notification` | Optional sale or marketing notifications | `consent_record`, notification channel stores |
| `analytics_b2b` | Contribute anonymized market trend data | Aggregate stores only; raw personal joins must be removed |

## Erasure summary

1. Hard-delete user-owned tracking, wishlist, alert, and alert state rows.
2. Anonymize payment identifiers that directly encode or identify the user.
3. Anonymize account PII in `app_user`.
4. Retain consent, DSAR, breach, and DPIA records as legal evidence.
5. Re-running erase must be safe and must not restore or expose deleted data.
