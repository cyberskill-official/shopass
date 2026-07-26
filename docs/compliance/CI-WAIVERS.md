# Compliance CI waivers

Gates in `.github/workflows/ci.yml` job **Compliance gates (R33)** may be waived
only with an explicit human tag. Agents must not invent waivers.

## Tags

| Gate | Tag | Where |
|------|-----|-------|
| no-cleartext / auditscan | `// audit:allow`, `-- audit:allow`, `# audit:allow` | same line as the match |
| Consent surface deferred | `waiver: reviewed:<id>` column | `docs/compliance/CONSENT-SURFACES.md` |
| DPIA overdue | `waiver: reviewed:<id>` column | `docs/compliance/DPIA.md` |
| Inventory / migration PII | `-- reviewed:<id>` on the SQL line | migration file |

`<id>` should be a human name, ticket, or short decision id (e.g. `Stephen`, `R33-bootstrap`).

## Procedure

1. Human reviews the risk and accepts temporary non-compliance.
2. Add the tag with a short reason in PR description / ledger.
3. Open or link a follow-up to remove the waiver.
4. CI stays green only while the tag is present; removing it without a fix turns the gate red.
