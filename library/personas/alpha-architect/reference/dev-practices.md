# Development Practices & Engineering Culture

Read when assessing or recommending how teams build and ship software.

- **Test strategy** — Test pyramid: unit tests for logic, integration tests for contracts, end-to-end tests for critical user journeys only. TDD where domain logic is complex. BDD when shared understanding between engineers and product is critical. No untested code in production.
- **DevSecOps** — Shift security left. SAST (SonarQube, Semgrep) and dependency scanning (Dependabot, Snyk) in CI. Container image scanning before push. Security review is part of the definition of done, not a gate after the fact.
- **Feature flags & progressive delivery** — Feature flags (LaunchDarkly, Unleash, Flagsmith) for decoupling deploy from release. Canary deployments with automated rollback on error rate spikes. Blue-green for zero-downtime database migrations.
- **Infrastructure immutability** — Cattle, not pets. Servers are replaced, not patched. Container images are built once and promoted across environments. No SSH into production unless incident demands it.
- **Platform engineering** — Internal developer platforms (IDPs) reduce cognitive load. Golden paths for service creation (templates, scaffolding, CI pipelines). Self-service infrastructure provisioning with guardrails.
- **SRE principles** — Error budgets drive release velocity decisions. SLOs are defined before launch, not after an incident. Toil is tracked and reduced. Blameless post-mortems are mandatory for SEV1/SEV2.

> For hands-on CI/CD, containers, and production ops, hand off to the `delta`
> DevOps persona. Alpha decides the strategy; Delta builds the pipeline.
