# Source of Truth — SănDeal

## Rule

Each FR document and its companion `.audit.md` file are the authoritative source of truth for that feature's implementation.

## Implications

1. **Implementation follows the FR doc, not external assumptions.** The FR §1 (normative description), §3 (API/DDL contracts), §4 (acceptance criteria), §5 (tests), and §6 (implementation outline) define what must be built.

2. **The `.audit.md` file records audit evidence.** Every acceptance criterion, normative statement, and risk has been independently audited 10/10. Implementation must preserve this traceability.

3. **No implementation deviates from the FR without a recorded decision.** If the implementing agent finds a conflict, ambiguity, or error, it must:
   - Stop and report the issue.
   - Record the proposed deviation.
   - Wait for resolution (either a re-audit or an explicit override).

4. **Cross-FR references are resolved by `depends_on` and `blocks` fields.** Never duplicate schema owned by another FR — reference it via FK.

5. **NFR files supplement but do not override FRs.** If an NFR requirement conflicts with a FR requirement, the FR requirement takes precedence unless the NFR explicitly overrides via cross-reference.

## Document Order (descending authority)

1. User command in chat (if applicable)
2. AGENTS.md (CyberOS memory protocol)
3. FR doc + `.audit.md` (per-feature source of truth)
4. SHIP-GUIDE.md (cross-cutting invariants)
5. DATA-MODEL.md (schema catalog)
6. BACKLOG.md / IMPLEMENTATION-ORDER.md (execution order)
7. STATUS-REFERENCE.md (status lifecycle)
8. NFR files (non-functional constraints)
