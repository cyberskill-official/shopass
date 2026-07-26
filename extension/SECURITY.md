# Security policy

## Supported versions

| Version | Supported |
|---------|-----------|
| `1.x` (MV3, this repo) | Yes |

## Reporting a vulnerability

Email **info@cyberskill.world** with subject `[Shopass extension security]`.

Please include:

1. Extension version (`manifest.json` → `version`) and browser/OS
2. Steps to reproduce
3. Impact (data exposure, privilege escalation, network abuse)
4. Whether you plan a public disclosure date

We aim to acknowledge within **48 hours** and provide a substantive update within **7 days**.

Do **not** open a public GitHub issue for security findings until we confirm a fix or coordinated disclosure plan.

## Scope notes

- Marketplace session tokens must never leave the client (see product trust docs / DNR allowlist).
- Affiliate activation is user-initiated only.
- Consent scopes for cart/voucher reads are documented in `docs/compliance/` (monorepo) and the store data-disclosure worksheet.
