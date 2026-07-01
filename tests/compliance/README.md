# Compliance Evidence Index — SănDeal

This directory holds compliance test definitions and evidence artifacts. Each check verifies a SHIP-GUIDE invariant.

## Check Categories

### No Cleartext
- Source: SHIP-GUIDE §1 (Security and PDPL)
- Verifies: No password, secret, or API key appears in plaintext in code, logs, DB, or config.
- FRs: FR-COMPLY-005, FR-TRUST-003, FR-TRUST-002

### Token Not on Server
- Source: SHIP-GUIDE §1
- Verifies: Platform session tokens/cookies (Shopee, TikTok Shop, Lazada) never leave the client extension. `platform_account.ext_user_ref` is anonymous.
- FRs: FR-EXT-003, FR-TRUST-002, FR-AUTH-003

### PDPL Consent
- Source: SHIP-GUIDE §1, FR-COMPLY-001
- Verifies: Consent is recorded with purpose, version, timestamp, IP, user-agent before any processing. User can withdraw.
- FRs: FR-COMPLY-001, FR-COMPLY-002, FR-COMPLY-003

### Affiliate Guardrails
- Source: SHIP-GUIDE §2 (Post-Honey)
- Verifies: Affiliate deep links are created only on explicit user action with disclosure. No cookie-stuffing, auto-redirect, or pop-under.
- FRs: FR-AFFIL-002, FR-AFFIL-004

### Country Gating
- Source: SHIP-GUIDE §8
- Verifies: Default policy is restrictive. MY/PH no-stack rules are enforced where configured. Unknown countries return restrictive defaults.
- FRs: FR-INFRA-005, FR-COMPLY-006, FR-CART-004
