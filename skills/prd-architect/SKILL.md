---
name: prd-architect
description: Convert business context into an implementation-ready Product Requirements Document detailed enough to derive a Low-Level Design without assumptions. Use whenever the user needs a PRD, wants to spec product requirements, turn a business problem or feature idea into developer-ready documentation, define functional and non-functional requirements, or write acceptance criteria for engineering handoff — even if they never say the word "PRD".
---

# prd-architect

**Before writing** — `${CLAUDE_PLUGIN_ROOT}/library/policies/work-tracking/SKILL.md`, Gate 1: check whether a spec already exists (Confluence page, older PRD, design doc) and whether a Jira ticket covers this. Extend what exists instead of duplicating it; propose creating the missing piece. On completion, publish the PRD via the service ladder and close out the ticket per Gate 3.

Turn business context into a PRD an engineer can build from with zero guesswork. Self-contained — depends on no other installed skill. Pairs with `product` (upstream discovery) and `ship` (downstream build; the acceptance criteria here become ship's exit contract).

## Principles

- **No assumptions.** Every requirement is explicit enough to derive a Low-Level Design directly. Where a decision is open, list it under Open Questions with a proposed default — don't silently assume.
- **Testable acceptance criteria.** Each requirement carries criteria phrased as verifiable pass/fail statements (Given/When/Then where useful). These drive QA and the `ship` loop's exit gate.
- **Right altitude.** Enough detail to remove ambiguity; not so much that it becomes the design doc. Capture the "what" and "why"; leave the "how" to LLD, but constrain the "how" where it affects contracts (APIs, data shapes, SLAs).

## Interview (only what's missing)

Gather, asking the user only for what you can't infer from context/codebase:
problem & who has it · goal & success metric · in-scope / explicitly out-of-scope · primary user flows · data & entities · external systems/integrations · non-functional needs (perf, scale, security, compliance, availability) · constraints & assumptions · rollout/flags.

For an existing system, inspect the codebase first (routes, models, integrations) so requirements match reality.

Modes: **full** PRD for a feature/system; **lightweight** contract when invoked by `ship` for a small task (just requirements + acceptance criteria + constraints).

## PRD structure

```markdown
# PRD: <title>

## 1. Summary
Problem, target user, and the outcome in 3–5 sentences.

## 2. Goals & Success Metrics
- Goal → metric (with baseline/target where known)
- Non-goals (explicitly out of scope)

## 3. Users & Use Cases
Personas and the concrete jobs-to-be-done this addresses.
When users act on data, add the **access matrix** (use case → actor → suggested role → permission `org::app::client::entity::action` → scope) with explicit allow and deny criteria; security NFRs in §5 then cover tenancy, session expiry, revocation lag and auth outage.

## 4. Functional Requirements
For each: `FR-n — <requirement>`
- Behavior (happy path + variations)
- Inputs / outputs / data shapes
- Acceptance criteria (Given/When/Then, pass/fail)

## 5. Non-Functional Requirements
Performance, scale, availability, security, privacy/compliance, observability, accessibility. Each with a measurable target.

## 6. Data & Domain
Entities, key fields, relationships, lifecycle (create/update/delete/soft-delete), retention, PII tagging.

## 7. Integrations & Contracts
External systems, API contracts (request/response/errors/auth), events, idempotency.

## 8. Edge Cases & Error Handling
Boundaries, empty/zero/negative, concurrency, partial failure, backward compatibility.

## 9. Rollout
Feature flags, migration/backfill, phased release, kill switch, metrics to watch.

## 10. Open Questions
Unresolved decisions — each with a proposed default.

## 11. Acceptance Criteria (consolidated)
The full testable checklist QA / `ship` verifies against. This is the definition of done.
```

## Handoff

End by stating: acceptance criteria are the build contract → run `ship` to implement, or `product` if strategy/discovery is still open. Keep section 11 copy-pasteable as the loop's exit gate.
