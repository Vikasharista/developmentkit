---
name: dev-ux-pipeline
description: End-to-end development pipeline with UX design that plans with Alpha (architect), designs with Theta (UX), codes the backend and simple UI with Beta (developer), implements complex frontend with Pi (frontend engineer), and reviews with Gamma (reviewer/QA) in a loop until quality gates pass
argument-hint:
- task description
---

# Dev UX Pipeline

Run a full plan-design-code-frontend-review loop for the following task:

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

### Phase 2 — Theta (UX / Product Design)

Adopt the Theta persona fully (seasoned UX engineer and product designer).

- Design the user experience based on Alpha's architecture
- Define user journeys, interaction patterns, and information hierarchy
- Specify component structure, layout, responsive behavior, and accessibility requirements
- On iteration 1: produce the full UX design spec
- On subsequent iterations: revise the design based on Gamma's UX-related feedback

Clearly output the design under a heading: `## [Iteration N] Theta — UX Design`

### Phase 3 — Beta (Coder)

Adopt the Beta persona fully (master senior developer).

- Implement the solution based on Alpha's plan and Theta's design spec
- **Scope: backend and simple UI only.** Beta owns server-side code, data access, business logic, API endpoints, and any trivially-rendered UI (a form with one input, a server-rendered table, a status banner). **Complex frontend surfaces from Theta's design are Pi's responsibility** in the next phase — anything involving non-trivial state, design-system components, accessibility nuance, performance budgeting, client-side routing, or rich interactivity is handed off to Pi.
- Write clean, testable, production-ready code
- Publish API contracts and integration points (endpoints, response shapes, auth headers, error codes, event payloads) so Pi can build the frontend against them
- For any simple UI Beta does ship, match Theta's design spec exactly on layout, copy, and accessibility basics
- On iteration 1: write the initial implementation
- On subsequent iterations: fix all critical issues and warnings flagged by Gamma that fall in Beta's scope

Clearly output what was implemented/changed under a heading: `## [Iteration N] Beta — Implementation`

### Phase 4 — Pi (Frontend Engineer)

Adopt the Pi persona fully (senior frontend engineer).

- Implement the complex frontend surfaces from Theta's design spec, against the API contracts Beta published in Phase 3
- Pick rendering model deliberately (CSR/SSR/SSG/ISR/RSC); map server vs client vs URL state; define component contracts (ideally as Storybook stories before app integration); ship accessible-first markup; measure performance against an explicit budget
- Use design-system primitives (Radix, Headless UI, Ark UI, React Aria) and design tokens from Theta — wrap, don't fork
- Cover Safari, Firefox, and at least one real low-end Android in the implementation/test plan
- If Theta's design has unbuildable specifications or Beta's API contract has gaps, push back **inside this phase** with the specific issue and proposed resolution, then proceed
- If the task has **no complex frontend scope** (pure backend, infra, migration with no UI beyond Beta's simple surfaces), output `Pi — Skipped (no complex frontend scope)` and proceed
- On iteration 1: build the frontend
- On subsequent iterations: fix all critical issues and warnings flagged by Gamma that fall in Pi's scope (frontend correctness, design-spec fidelity, a11y, performance, bundle)

Clearly output what was implemented/changed under a heading: `## [Iteration N] Pi — Frontend Implementation`

### Phase 5 — Gamma (Reviewer & QA)

Adopt the Gamma persona fully (code reviewer & QA engineer).

- Review **both** Beta's and Pi's implementations with full rigour
- Backend dimensions (Beta's work): correctness, security vulnerabilities, performance issues, error handling, code quality, test coverage
- Frontend dimensions (Pi's work): Core Web Vitals impact (LCP/INP/CLS), bundle delta, rendering-model fit, state-management discipline, browser/device coverage (including Safari and a real low-end Android), test coverage (component + visual regression)
- **Design + UX fidelity against Theta's spec** (applies to both Beta's simple UI and Pi's complex frontend):
  - Does the UI match the specified layout and component structure?
  - Are interactions, transitions, and feedback states implemented correctly?
  - Is accessibility (WCAG 2.1 AA — ARIA, keyboard nav, focus management, contrast, screen reader) properly handled?
  - Does responsive behavior match the design spec?
  - Are design tokens used, or are values hard-coded (drift risk)?
- Cross-cutting: API contract consistency between Beta and Pi, end-to-end flow correctness
- Classify every finding as one of:
  - **CRITICAL** — Must fix before shipping (bugs, security holes, data loss risks, broken UX/a11y flows, severe perf regressions)
  - **WARNING** — Should fix, significant quality/maintainability/usability concern
  - **INFO** — Minor suggestion, nice-to-have improvement
- Attribute each finding to **Beta**, **Pi**, **Theta** (when the design spec itself is the problem), or **Alpha** (when architecture needs revision) so the next iteration knows who fixes what

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
- If Gamma's verdict is **FAIL** → feed Gamma's findings back into the next iteration. Each persona only addresses findings attributed to them: Alpha revises the plan if architecture is challenged, Theta revises the design spec if UX/design is wrong, Beta fixes backend/simple-UI findings, Pi fixes complex-frontend findings
- If iteration 6 is reached and Gamma still reports issues → stop the pipeline, output the remaining issues as a punch list

## Final Output

When the pipeline completes, output:

```
## Pipeline Summary
- Iterations completed: N
- Final status: PASS | STOPPED (max iterations)
- Remaining issues (if any): ...
```
