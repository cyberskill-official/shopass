# Data Model: SănDeal Full Project Plan

This file models the delivery program and references the product data model in `docs/feature-requests/DATA-MODEL.md`. It does not replace product DDL. Product tables remain owned by their original FRs.

## Program Entities

### ImplementationProgram

Represents the full project delivery.

Fields:

- `id`: stable string, `sandeal-full-project`
- `name`: `SănDeal Full Project`
- `source_docs`: list of source docs used for planning
- `phases`: P0, P1, P2, P3
- `total_fr`: 90
- `total_nfr`: 10
- `estimated_hours`: 627
- `active_constitution`: `docs/feature-requests/SHIP-GUIDE.md`

Validation:

- `total_fr` must match BACKLOG count.
- `estimated_hours` must match IMPLEMENTATION-ORDER total unless backlog changes.

### FeatureRequest

Represents one audited FR markdown file.

Fields:

- `id`: e.g. `FR-PRICE-002`
- `title`
- `module`
- `phase`
- `priority`: MUST, SHOULD, COULD, MAY
- `status`: one of `draft`, `ready_to_implement`, `implementing`, `ready_to_review`, `reviewing`, `ready_to_test`, `testing`, `done`, `on_hold`, `closed`
- `depends_on`: list of FR ids
- `blocks`: list of FR ids
- `effort_hours`
- `source_path`
- `audit_path`
- `service`
- `new_files`
- `modified_files`
- `test_references`

Validation:

- `depends_on` must be acyclic.
- `blocks` must be reciprocal with reverse dependencies.
- A FR can move to `done` only after its tests and acceptance criteria pass.
- A FR can start only when dependencies are `done`, unless human override is recorded.

State transitions:

```text
ready_to_implement -> implementing -> ready_to_review -> reviewing -> ready_to_test -> testing -> done
```

Off-ramp states:

```text
ready_to_implement|implementing|reviewing|testing -> on_hold
ready_to_implement|on_hold -> closed
closed|on_hold -> ready_to_implement
```

### Module

Represents a product/service area.

Fields:

- `code`: infra, auth, scrape, price, ext, track, deal, notif, web, comply, trust, cart, affil, bill, b2b, mobile
- `stack`
- `owned_fr_ids`
- `owned_tables`
- `release_phase`

Validation:

- Every FR belongs to exactly one module.
- Every product table has exactly one owner FR.

### DependencyLayer

Represents one topological build layer.

Fields:

- `index`: 0 through 7
- `fr_ids`
- `total_effort_hours`
- `can_parallelize`: true once all earlier layers are complete

Validation:

- A layer cannot contain a FR that depends on another FR in the same or later layer.

### ReleasePhase

Represents a business release gate.

Fields:

- `code`: P0, P1, P2, P3
- `name`
- `scope_modules`
- `fr_ids`
- `exit_gate`
- `demo_scenarios`
- `compliance_evidence`

Validation:

- P1 cannot release before all P0 MUST FRs are `done`.
- P2 cannot release before all P1 MUST FRs are `done`.
- P3 cannot release before all P2 MUST FRs that feed growth/fraud/compliance are `done`.

### GateEvidence

Represents proof that a FR or phase is shippable.

Fields:

- `subject_type`: `fr` or `phase`
- `subject_id`
- `test_report_paths`
- `contract_report_paths`
- `migration_status`
- `observability_links`
- `security_audit_notes`
- `compliance_notes`
- `reviewer`
- `recorded_at`

Validation:

- Phase evidence must include end-to-end demo result.
- Compliance evidence is mandatory for P1, P2 and P3.

## Product Data Model Reference

Product DDL remains in the owner FR files and summarized in `docs/feature-requests/DATA-MODEL.md`.

Core ownership map:

- `platform`, `app_user`: FR-INFRA-002
- `platform_account`: FR-AUTH-003
- `refresh_token`: FR-AUTH-002
- `tracked_product`: FR-PRICE-001
- `price_snapshot`, `price_daily`: FR-PRICE-002
- `canonical_review_queue`: FR-PRICE-005
- `wishlist`, `wishlist_item`, `alert_rule`, `alert`: FR-TRACK-002/003
- `voucher_catalog`: FR-CART-001
- `cart_snapshot`, `cart_item`: FR-CART-002
- `affiliate_click`, `affiliate_conversion`: FR-AFFIL-001
- `subscription`, `payment`, `referral_code`: FR-BILL-001/003/004
- `notification`, `user_channel_token`, `notification_dlq`: FR-NOTIF-001/003
- `consent_policy`, `consent_record`, `dpia`, `tia`, `dsar_request`, `breach_incident`, `country_rule`: FR-COMPLY-001/002/003/004/006
- `fraud_signal`, `account_link_edge`, `payout_hold`, `device_fingerprint`: FR-TRUST-004/005/006
- `market_trend_daily`, `b2b_subscription`, `seller_owned_sku`: FR-B2B-001/002/003

Global validation rules:

- Money values are BIGINT VND.
- `price_snapshot` is delta-only and read-heavy flows use `price_daily`.
- Extension-origin payloads must not contain platform cookie/token/password.
- Tables are created only by their owner FR; later FRs use FK or ALTER.
