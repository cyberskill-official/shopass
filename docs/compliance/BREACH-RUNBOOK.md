# Breach notification runbook

Scope: Shopass/SanDeal personal-data incidents under Vietnam PDPL Law
91/2025/QH15 and the internal TASK-COMPLY-004 state machine.

## 72 hour rule

- The 72 hour clock starts at `breach_incident.acknowledged_at`: when Shopass
  becomes aware of a likely personal-data breach.
- Do not count from `occurred_at`; that field is for investigation only.
- Authority notice target: before `acknowledged_at + 72h`.
- `complysvc` tracks states:
  `detected -> triaged -> notified_authority -> notified_subjects -> closed`.
- High and critical incidents must not close before data subjects are notified.

## Roles

| Role | Primary duty | Backup |
|---|---|---|
| Incident commander | Own timeline, severity, and decisions | CTO |
| Privacy lead | PDPL assessment, MPS notice, user notice wording | Counsel |
| Security lead | Containment, forensics, evidence preservation | Platform lead |
| Customer lead | User support, public FAQ, status page copy | Founder |
| Engineering lead | Fix forward, deploy, verify no recurrence | Service owner |

## First hour checklist

1. Open an incident: `POST /v1/comply/breach/open`.
2. Preserve evidence: logs, traces, database snapshots, deployed commit SHAs,
   alert links, access records.
3. Contain: revoke leaked credentials, isolate affected service, disable risky
   jobs, rotate tokens where needed.
4. Triage severity:
   - low: no personal data exposed, no unauthorized access confirmed.
   - medium: limited personal data exposure, contained quickly.
   - high: confirmed personal data exposure or account takeover risk.
   - critical: sensitive data, broad exposure, ongoing abuse, or regulatory risk.
5. Advance to `triaged` once severity and initial scope are recorded.

## MPS / authority contact record

Keep current contact details in the private incident-response vault. The
runbook records the accountable agency and the fields to prepare, not secrets
or personal phone numbers.

| Contact | Purpose | Evidence packet |
|---|---|---|
| Ministry of Public Security (MPS), personal-data protection channel | 72h authority notice | Incident summary, dates, data categories, affected count, containment, remedial steps |
| Local counsel | Legal review before notice | Draft authority notice, user notice, evidence packet |
| Payment partners, if affected | Contractual/payment-risk notice | Payment scope, transaction ids anonymized where possible |

## Authority notice template

Subject: Personal data breach notice - Shopass/SanDeal - incident `<id>`

1. Controller / processor: Shopass / SanDeal.
2. Contact: `<privacy lead name, role, email, phone>`.
3. Acknowledged at: `<timestamp with timezone>`.
4. Incident summary: `<plain-language summary>`.
5. Systems affected: `<services, databases, regions>`.
6. Data categories: `<email, phone, account, tracking, payment metadata, etc.>`.
7. Estimated affected subjects: `<count or range and confidence>`.
8. Likely consequences: `<risk to users>`.
9. Containment completed: `<actions and timestamps>`.
10. Remediation planned: `<fixes, rotations, monitoring>`.
11. Subject notification plan: `<required/not required; timing>`.
12. Evidence references: `<trace ids, alert ids, commit/deploy ids>`.

## Data-subject notice template

Subject: Important notice about your SanDeal account data

Hello,

We identified a security incident on `<date>` involving `<short description>`.
The data involved may include `<data categories>`. We have contained the issue
by `<actions>`.

What you should do:

- `<recommended user action 1>`
- `<recommended user action 2>`

What we are doing:

- `<remediation 1>`
- `<monitoring/support 2>`

If you have questions or want to exercise your data rights, contact
`<privacy contact>`.

## Closure criteria

- Authority notice sent or formally assessed as not required.
- Subject notice sent for high/critical incidents, or counsel records why not.
- Root cause fixed and verified.
- Evidence packet stored in the incident vault.
- `complysvc` incident closed.
- Follow-up tasks filed for prevention and monitoring gaps.
