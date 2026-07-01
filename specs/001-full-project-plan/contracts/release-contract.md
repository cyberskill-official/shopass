# Release Contract: SănDeal Full Project Plan

This contract defines what must be true before a phase or full project can be called complete.

## FR Completion Contract

For each FR:

- `status` is advanced according to `docs/feature-requests/STATUS-REFERENCE.md`.
- All `depends_on` FRs are `done` before work starts, unless an override is recorded.
- All files listed in `new_files` exist or the FR records an approved deviation.
- All §1 MUST/SHOULD statements trace to §4 acceptance criteria and §5 tests.
- Tests pass locally and in CI.
- The companion `.audit.md` remains traceable after implementation changes.
- BACKLOG status matches FR frontmatter status.

## Phase Release Contract

For each phase:

- All MUST FRs in scope are `done`.
- SHOULD/COULD items are either `done` or explicitly deferred.
- End-to-end quickstart scenario passes.
- Migration state is reproducible from empty database.
- Observability covers request path, background jobs, errors and external integrations.
- Compliance evidence is attached for data processing, secrets, affiliate behavior and country rules where applicable.

## Security and Compliance Contract

These are hard failures:

- Server stores platform cookie/session token/password.
- Extension sends cookie/token to backend.
- Secret appears cleartext in repo, logs, DB or config.
- Money stored as float for product price, order value, commission, cashback, payment or subscription price.
- Affiliate link is generated without explicit user action and disclosure.
- Consent-dependent data processing runs without active consent.
- Country rule missing and behavior defaults permissive.

## API/Integration Contract Families

Concrete endpoint contracts stay in each FR. The full project must expose these families by the end of the relevant phase:

- P0: gateway health, auth verification hooks, migration status, observability endpoints.
- P1: auth/session, platform account link, product tracking, price history, chart data, alert rules, notification registration, consent/DSAR.
- P2: cart snapshot, voucher optimizer, affiliate deeplink, payment webhook/reconciliation, multi-platform scrape adapters.
- P3: cashback payout, B2B trend exports, mobile auth/push, anti-fraud signals, SEA country gates.

## Evidence Contract

Every release gate must produce an evidence bundle containing:

- commit or build identifier
- list of FR ids included
- test command summary and pass/fail result
- migration version
- relevant dashboard/log links or exported snapshots
- security/compliance notes
- known risks and deferred FRs
