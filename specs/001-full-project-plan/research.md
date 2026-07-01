# Research: SănDeal Full Project Plan

## Decision: Treat existing FR docs as authoritative implementation units

**Rationale**: The repo already contains 90 FR + 10 NFR, each audited independently and organized by phase, module, dependency, status and effort. Rewriting them into one giant spec would lose traceability and create drift.

**Alternatives considered**:

- Create one monolithic spec replacing the backlog. Rejected because it would duplicate audited source-of-truth documents.
- Generate separate Spec Kit feature folders for all 90 FR now. Rejected for the planning step because the FR files already contain the detailed §1-§11 contract; generating 90 folders is better as an automation follow-up if desired.

## Decision: Use `SHIP-GUIDE.md` as active constitution

**Rationale**: `.specify/memory/constitution.md` is still a placeholder template, while `SHIP-GUIDE.md` contains the actual non-negotiable invariants, stack and workflow for SănDeal.

**Alternatives considered**:

- Block planning until constitution is rewritten. Rejected because the user asked to plan now and the repo already has project-specific governance.
- Ignore constitution gates. Rejected because compliance/security/affiliate invariants are central to the product.

## Decision: Build by dependency layer, not by phase alone

**Rationale**: `IMPLEMENTATION-ORDER.md` contains a verified acyclic DAG with 8 layers. Layer execution allows safe parallelism and prevents starting dependent FRs early. Phase gates still matter for release validation.

**Alternatives considered**:

- Build P0, then P1, then P2, then P3 strictly. Rejected because some dependency-free P1/P2 scaffolds can safely start early and unblock later work.
- Build by module. Rejected because cross-module dependencies are strong, especially infra/auth/price/ext/comply.

## Decision: Prioritize critical price-history path and scraping cold-start

**Rationale**: Sale-realness and bottom-price features require accumulated history. Starting scrape and delta-only writes early lets the data clock run while UI, auth and trust surfaces are built.

**Alternatives considered**:

- Build web UI first with mock data. Rejected as a primary path because it delays the 90-day history problem.
- Build ML before price foundation. Rejected because ML depends on `price_daily` and history quality.

## Decision: Use monorepo layout

**Rationale**: The product spans backend services, extension, web, ML, mobile and deploy assets with shared contracts and compliance gates. A monorepo keeps cross-FR work visible and reduces schema/contract drift.

**Alternatives considered**:

- One repo per service. Rejected for this stage because the project starts from a docs-only repo and needs coordinated implementation.
- Single backend service only. Rejected because extension/web/mobile/ML are first-class surfaces in the audited backlog.

## Decision: Default mobile framework to React Native until FR-MOBILE-001

**Rationale**: The docs allow React Native or Flutter. React Native aligns with existing TypeScript choices for extension and web, reducing language/context switching.

**Alternatives considered**:

- Choose Flutter now. Valid option, but not necessary until P3 and would add a second frontend language.
- Leave as unknown. Rejected because a plan should carry an executable assumption; the FR can still override.

## Decision: Release gates by P0/P1/P2/P3

**Rationale**: BACKLOG defines phase exit gates. Gates are necessary because "done FRs" alone do not prove end-to-end product behavior, compliance evidence or operational readiness.

**Alternatives considered**:

- Gate only at the end of all 90 FR. Rejected because it delays integration discovery.
- Gate every FR only. Rejected because FR-level tests miss cross-service behavior.
