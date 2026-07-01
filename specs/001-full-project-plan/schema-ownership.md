# Schema Ownership — One Table One Owner

Derived from `docs/feature-requests/DATA-MODEL.md`.

## Rule

Each database table has exactly one owner FR. The owner FR issues the `CREATE TABLE` statement. Other modules reference the table via foreign key (FK) or extend it via `ALTER TABLE`. No module re-creates a table owned by another FR.

## Owner Map

| Table | Owner FR | Notes |
|-------|----------|-------|
| `platform` | FR-INFRA-002 | Foundation FK target |
| `app_user` | FR-INFRA-002 | Extended by FR-AUTH-001 (`pwd_hash`, `referral_code_id`) |
| `platform_account` | FR-AUTH-003 | Anonymous `ext_user_ref`, no token/cookie stored |
| `refresh_token` | FR-AUTH-002 | Token hash only, no cleartext |
| `social_identity` | FR-AUTH-004 | Provider + subject unique |
| `password_reset` | FR-AUTH-005 | Token hash, expires_at |
| `tracked_product` | FR-PRICE-001 | Canonical key for cross-platform matching |
| `price_snapshot` | FR-PRICE-002 | TimescaleDB hypertable, delta-only |
| `price_daily` | FR-PRICE-002 | Continuous aggregate, not a physical table |
| `canonical_review_queue` | FR-PRICE-005 | Queue for low-confidence canonical matches |
| `user_tracked_product` | FR-TRACK-001 | User-product follow link |
| `wishlist` / `wishlist_item` | FR-TRACK-002 | Target price as BIGINT VND |
| `alert_rule` / `alert` | FR-TRACK-003 | Price thresholds, BIGINT only |
| `alert_fired_state` | FR-TRACK-004 | Edge-triggered anti-spam |
| `voucher_catalog` | FR-CART-001 | Per-platform voucher definitions |
| `cart_snapshot` / `cart_item` | FR-CART-002 | Captured cart state |
| `user_activity_coin_task` | FR-CART-006 | Coin/activity checklist |
| `affiliate_click` / `affiliate_conversion` | FR-AFFIL-001 | Tracking with anonymous `sub_id` |
| `affiliate_network` / `affiliate_postback_log` | FR-AFFIL-003 | Network config + postback verification |
| `cashback_entry` / `payout_request` | FR-AFFIL-005 | Cashback layering |
| `plan_catalog` / `subscription` | FR-BILL-001 | Normalized: price in plan_catalog |
| `payment` | FR-BILL-003 | Gateway transaction + reconciliation |
| `referral_code` | FR-BILL-004 | One code per user |
| `plan_feature` | FR-BILL-005 | Tier feature gating |
| `notification` | FR-NOTIF-001 | Extended by FR-NOTIF-003 (attempts, lease_until, last_error) |
| `user_channel_token` | FR-NOTIF-001 | Per-channel push/email/sms address |
| `notification_dlq` | FR-NOTIF-003 | Dead-letter queue |
| `consent_policy` / `consent_record` | FR-COMPLY-001 | PDPL consent framework |
| `processing_activity` / `dpia` / `tia` | FR-COMPLY-002 | DPIA/TIA register |
| `dsar_request` | FR-COMPLY-003 | Data subject rights |
| `breach_incident` | FR-COMPLY-004 | 72h breach notification |
| `country_rule` | FR-COMPLY-006 | Per-country policy gating |
| `ecommerce_obligation` / `yearly_transaction_count` / `compliance_threshold` | FR-COMPLY-008 | VN ecommerce law |
| `fraud_signal` / `account_link_edge` | FR-TRUST-004 | Anti-fraud engine |
| `payout_hold` | FR-TRUST-005 | Attribution gaming detection |
| `device_fingerprint` | FR-TRUST-006 | Multi-account detection |
| `market_trend_daily` | FR-B2B-001 | Anonymized k-anonymity |
| `b2b_subscription` | FR-B2B-002 | B2B insights reports |
| `seller_owned_sku` | FR-B2B-003 | Seller competitor analytics |
| `api_key` / `api_usage` | FR-B2B-004 | Premium API access |
| `scrape_job` | FR-SCRAPE-001 | Scraping orchestrator |
| `proxy_usage` | FR-SCRAPE-004 | Proxy cost tracking |
| `price_forecast` | FR-DEAL-004 | ML price predictions |
| `feature_store` | FR-DEAL-005 | Feature engineering for ML |
| `bottom_alert_log` | FR-DEAL-006 | Bottom price alerts |

## Extension Rules

- Extending a table owned by another FR uses `ALTER TABLE` only.
- The extending FR must list the owner FR in its `depends_on`.
- Never redefine columns owned by another FR.
- Reference tables via FK; do not duplicate schema.
