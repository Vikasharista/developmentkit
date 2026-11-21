---
name: product
description: Take a product problem or feature idea from discovery to a buildable spec — validate the opportunity, shape strategy, then produce an implementation-ready PRD with testable acceptance criteria that engineering can execute via `ship`. Use for product discovery, opportunity validation, PRDs/specs, experiments/A-B design, product analytics/KPIs, competitive teardowns, roadmap narratives, and UX research/design. Routes internally by need.
---

# product — problem to buildable spec

The product-role entrypoint. Routes internally to the toolkit in `${CLAUDE_PLUGIN_ROOT}/library/product/` and the PM persona in `library/personas/epsilon-pm`. Don't make the user pick a sub-skill.

**Ground the work first** — `${CLAUDE_PLUGIN_ROOT}/library/policies/work-tracking/SKILL.md`, Gate 1: ask once whether a Jira ticket and an existing spec/doc cover this, before producing a new one. If a doc already exists, extend it rather than writing a parallel spec. On completion, Gate 3 applies to any linked ticket.

**Docs split** — `${CLAUDE_PLUGIN_ROOT}/library/policies/docs-split/SKILL.md`: Confluence is the product view, not a mirror of Epics and Tasks; `product` owns the **Product definition** page and **Release history** entries under the product folder, Jira stays the engineering view.

## Route by intent

| User need | Route to |
|---|---|
| Validate an opportunity, test problem-solution fit, discovery sprint | `library/product/product-discovery` |
| OKRs, quarterly planning, competitive landscape, strategy | `library/product/product-strategist` |
| **Write a PRD / spec / requirements / acceptance criteria** | **`${CLAUDE_PLUGIN_ROOT}/skills/prd-architect/SKILL.md`** (bundled — the PRD engine) |
| Reverse-engineer an existing codebase into a PRD | `library/product/code-to-prd` |
| Design an experiment / hypothesis / A-B / sample size | `library/product/experiment-designer` |
| Define KPIs, cohort/retention/adoption analysis | `library/product/product-analytics` |
| Competitor / market teardown | `library/product/competitive-teardown` |
| Roadmap narrative, release notes, changelog, stakeholder update | `library/product/roadmap-communicator` |
| Personas, journey maps, usability, UX research | `library/product/ux-researcher-designer` |
| Design tokens, component docs, UI system | `library/product/ui-design-system` |
| iOS/macOS HIG audit/design | `library/product/apple-hig-expert` |
| General multi-tool product coordination | `library/product/product-skills` or `product-manager-toolkit` |
| Idea/spec → scaffold a starter repo | `library/product/spec-to-repo` / `saas-scaffolder` (then hand to `ship`) |
| Summarize research/papers/reports | `library/product/research-summarizer` |

## The product → engineering handoff (keeps the loop tight)

Whatever the entry point, when the goal is "get this built", converge on a **PRD via the bundled `prd-architect`** whose acceptance criteria are testable. Those criteria become the exit contract for `ship` — engineering iterates until they pass, so there's no spec↔build ping-pong. Explicitly hand off: "acceptance criteria ready → run `ship`".

**Who can do what** — whenever a use case has users acting on data discovery and the PRD carry an access matrix (actor → role → permission) with deny-case criteria. Validate the matrix with the stakeholder during discovery, not after the build.

**When the build described by the PRD spans more than one Epic** — `${CLAUDE_PLUGIN_ROOT}/library/policies/task-chips/SKILL.md`: break it down in Jira (Initiative > Feature > Epic > Task) in this session, then hand each Epic to its own session as a self-contained task chip that opens with `/devkit:ship`. Do not chain the builds into this conversation.
