# DPIA register (CI-tracked)

Source of truth for R33 DPIA review-date gate. Each row is a processing activity.
`review_due` is ISO-8601 (UTC date). CI fails when `review_due < today` unless the
row carries `waiver: reviewed:<ticket-or-name>`.

| id | activity | filing_date | review_due | owner | waiver |
|----|----------|-------------|------------|-------|--------|
| DPIA-001 | Account registration and session (email, locale) | 2026-06-01 | 2026-12-01 | Stephen Cheng | |
| DPIA-002 | Price tracking, wishlist, alert rules | 2026-06-01 | 2026-12-01 | Stephen Cheng | |
| DPIA-003 | Extension cart/voucher read (client-only; consent `cart_read`) | 2026-06-15 | 2026-12-15 | Stephen Cheng | |
| DPIA-004 | Billing / payments (order refs, transaction ids) | 2026-06-20 | 2026-12-20 | Stephen Cheng | |
| DPIA-005 | Consent, DSAR, breach compliance records | 2026-07-26 | 2027-01-26 | Stephen Cheng | |

## Rules

1. New personal-data processing → add a row within 60 days of start (`filing_date`).
2. Re-review at least every 6 months (`review_due`).
3. Cross-border transfers need a TIA note in `docs/compliance/DATA-INVENTORY.md` before enablement.
4. Waivers are temporary; remove `waiver:` once the review is filed.
