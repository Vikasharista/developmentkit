# Architecture Governance

Read when establishing or reviewing how architectural decisions are made and
maintained.

- **Architecture Decision Records (ADRs)** — Every significant decision (new service, tech adoption, pattern change) is documented as an ADR: context, decision, consequences, status. ADRs are stored in the repo, versioned with the code.
- **Technical debt tracking** — Debt is logged, categorised (deliberate vs. accidental, reckless vs. prudent), and prioritised alongside feature work. Every sprint carries a debt budget. Invisible debt is the most expensive kind.
- **Architecture review process** — Lightweight review for routine changes (PR review with architectural checklist). Formal review for boundary-crossing changes (new service, data model changes, public API changes). Reviews produce ADRs, not just approvals.
- **Fitness functions** — Automated architecture tests that run in CI: dependency direction checks (ArchUnit, DepCheck), module boundary enforcement, API contract validation, performance budget checks. Architecture that isn't tested drifts.
- **Documentation as code** — C4 model diagrams (system context, container, component, deployment views) maintained in the repo. Architecture overviews in markdown, not wiki pages that decay. Diagrams are generated from code or structured definitions where possible.

## Review gate

- [ ] Is the decision documented as an ADR (context, decision, consequences)?
- [ ] Is known technical debt logged, categorised, and scheduled for resolution?
- [ ] Are architectural fitness functions running in CI (dependency checks, boundary enforcement)?
- [ ] Are C4 diagrams maintained in the repo and up to date?
- [ ] Is there a defined review threshold (which changes require formal arch review)?
