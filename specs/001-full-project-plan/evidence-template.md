# Evidence Bundle Template — SănDeal

Based on the Release Contract (specs/001-full-project-plan/contracts/release-contract.md).

## Required Fields

Every release gate evidence bundle MUST contain:

1. **Commit / Build Identifier**
   - Git commit hash or CI build number.

2. **FR IDs Included**
   - List of all FRs in scope for this gate.

3. **Test Command Summary and Pass/Fail Result**
   - Exact commands run.
   - Output summary (pass/fail per test).
   - Link to full test log if available.

4. **Migration Version**
   - Current migration version or sequence number.
   - Confirmation that migration applies cleanly from empty database.

5. **Relevant Dashboard / Log Links or Exported Snapshots**
   - OTel trace IDs or links.
   - Prometheus/Grafana dashboard snapshots.
   - Structured log excerpts.

6. **Security / Compliance Notes**
   - Confirmation of no-cleartext, token-not-on-server, affiliate guardrails, PDPL consent.
   - Any compliance waivers or deviations recorded.

7. **Known Risks and Deferred FRs**
   - Any SHOULD/COULD items not yet done.
   - Known bugs or limitations.
   - FRs explicitly deferred to a later phase.

## Per-Gate Mapping

| Gate | Phase | Evidence File |
|------|-------|--------------|
| P0 Foundation | Phase 2 | `evidence/p0-foundation.md` |
| P1 MVP | Phase 3 | `evidence/p1-mvp.md` |
| P2 Expansion | Phase 4 | `evidence/p2-expansion.md` |
| P3 Growth | Phase 5 | `evidence/p3-growth.md` |
| Final Release | Phase 6 | `evidence/final-release.md` |
