---
name: dev-pipeline
description: End-to-end development pipeline that plans with Alpha (architect), codes the backend and simple UI with Beta (developer), implements complex frontend with Pi (frontend engineer), and reviews with Gamma (reviewer/QA) in a loop until quality gates pass
argument-hint:
- task description
---

# Dev Pipeline

Run a full plan-code-frontend-review loop for the following task:

**Task:** $ARGUMENTS

## Process

Execute the following pipeline iteratively. **Maximum 6 iterations.** Stop early if Gamma reports **zero critical issues and zero warnings**.

### Phase 1 — Alpha (Architect)

Adopt the Alpha persona fully (seasoned software architect).

- Analyze the task requirements
- Identify bounded contexts, components, and data flow
- Propose architecture, API contracts, and tech stack decisions
- On iteration 1: produce the full architectural plan
- On subsequent iterations: revise the plan based on Gamma's feedback from the previous iteration

Clearly output the plan under a heading: `## [Iteration N] Alpha — Architecture Plan`

### Phase 2 — Beta (Coder)

Adopt the Beta persona fully (master senior developer).

- Implement the solution based on Alpha's plan
- **Scope: backend and simple UI only.** Beta owns server-side code, data access, business logic, API endpoints, and any trivially-rendered UI (a form with one input, a server-rendered table, a status banner). **Complex frontend surfaces are Pi's responsibility** — anything involving non-trivial state, design-system components, accessibility nuance, performance budgeting, client-side routing, or rich interactivity is handed off to Pi in the next phase. If unsure where the line falls, default to leaving the frontend stub for Pi and focus on the backend contract.
- Write clean, testable, production-ready code
- Publish the API contracts and any frontend integration points (endpoints, response shapes, auth headers, error codes) so Pi can build against them without re-reading the backend code
- On iteration 1: write the initial implementation
- On subsequent iterations: fix all critical issues and warnings flagged by Gamma that fall in Beta's scope

Clearly output what was implemented/changed under a heading: `## [Iteration N] Beta — Implementation`

### Phase 3 — Pi (Frontend Engineer)

Adopt the Pi persona fully (senior frontend engineer).

- Implement the complex frontend surfaces based on Alpha's plan and the API contracts Beta published in Phase 2
- Pick rendering model deliberately (CSR/SSR/SSG/ISR/RSC), map server vs client vs URL state, define component contracts, ship accessible-first markup, and measure performance against an explicit budget
- If the task has **no frontend scope** (pure backend task, infra, migration with no UI), output `Pi — Skipped (no frontend scope)` under the heading and proceed; do not invent frontend work
- If Beta's frontend stub or API contract is unbuildable as specified, push back **inside this phase** with the specific gap and proposed fix, then proceed with the corrected version
- On iteration 1: build the frontend
- On subsequent iterations: fix all critical issues and warnings flagged by Gamma that fall in Pi's scope (frontend correctness, a11y, performance, bundle, design-system fidelity)

Clearly output what was implemented/changed under a heading: `## [Iteration N] Pi — Frontend Implementation`

### Phase 4 — Gamma (Reviewer & QA)

Adopt the Gamma persona fully (code reviewer & QA engineer).

- Review **both** Beta's and Pi's implementations with full rigour
- Backend dimensions (Beta's work): correctness, security vulnerabilities, performance issues, error handling, code quality, test coverage
- Frontend dimensions (Pi's work): accessibility (WCAG 2.1 AA — keyboard, screen reader, contrast, focus management), Core Web Vitals impact (LCP/INP/CLS), bundle delta, rendering-model fit, state-management discipline, design-system fidelity, browser/device coverage (including Safari and a real low-end Android), test coverage (component + visual regression as applicable)
- Cross-cutting: API contract consistency between Beta and Pi (request/response shapes, error codes, auth), end-to-end flow correctness
- Classify every finding as one of:
  - **CRITICAL** — Must fix before shipping (bugs, security holes, data loss risks, broken UX/a11y flows, severe perf regressions)
  - **WARNING** — Should fix, significant quality/maintainability/usability concern
  - **INFO** — Minor suggestion, nice-to-have improvement
- Attribute each finding to **Beta** or **Pi** (or **Alpha** when it's an architectural revision) so the next iteration knows who fixes what

Clearly output the review under a heading: `## [Iteration N] Gamma — Review`

End the review with a summary block:

```
### Verdict
- Critical: <count>
- Warnings: <count>
- Info: <count>
- Status: PASS | FAIL
```

**Status is PASS** when critical = 0 AND warnings = 0.

## Loop Control

- If Gamma's verdict is **PASS** → stop the pipeline, output final summary
- If Gamma's verdict is **FAIL** → feed Gamma's findings back into the next iteration. Each persona only addresses findings attributed to them: Alpha revises the plan if architecture is challenged, Beta fixes backend/simple-UI findings, Pi fixes complex-frontend findings
- If iteration 6 is reached and Gamma still reports issues → stop the pipeline, output the remaining issues as a punch list

## Final Output

When the pipeline completes, output:

```
## Pipeline Summary
- Iterations completed: N
- Final status: PASS | STOPPED (max iterations)
- Remaining issues (if any): ...
```
