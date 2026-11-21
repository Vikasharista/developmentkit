# devkit

Engineering + product + support toolkit. **Seven goal-oriented verbs** are all a user needs to learn — each routes internally (by role, stack, or use-case) to a large library of senior-role personas and pipelines, so nobody has to pick from a wall of skills.

## The surface — what you invoke

| Verb | Who | Give it… | It does |
|---|---|---|---|
| **ship** | Eng | a build/fix task | Plans → codes → reviews in a loop until acceptance criteria + quality gates pass. Task in, verified result out. |
| **bootstrap** | Eng | a new Go service or entity | Designs the entity with you, double-confirms the DB schema, then generates migration → repository → service → mappers → GraphQL subgraph and/or gin REST → tests (100% coverage, e2e) → Makefile → docs, following the conventions in `library/bootstrap/conventions.md`. Frontend: coming soon. |
| **review** | Eng | a PR / branch / path | Deep code + data-layer + adversarial review → one severity-ranked verdict. |
| **understand** | Eng | a codebase / PR / question | Knowledge-graph: architecture, domain, file/PR explain, onboarding. |
| **product** | Product | a problem / idea | Discovery → strategy → **PRD with testable acceptance criteria** → hands off to `ship`. |
| **deliver** | PM/TPM | delivery / Atlassian work | Sprints, portfolio, risk, agile coaching, Jira/Confluence (via the service tooling ladder). |
| **support** | Support | a ticket / issue | Troubleshoot + root-cause + escalate, then write the runbook/docs. |

Plus focused helpers: **prd-architect** (bundled PRD engine), **playwright-pro** (E2E), **pr-review**.

## The goal-loop, end to end

`product` → `prd-architect` produces **acceptance criteria** → those become the **exit contract** for `ship`, which iterates (plan/code/review + TDD + adversarial + Playwright) until every criterion demonstrably passes. This is the "give a task, get a polished verified result, no repeated back-and-forth" path.

## Design: concise surface, deep internals

- **`skills/`** holds only the invocable verbs + helpers (10). This is the entire menu users see.
- **`library/`** holds the building blocks the verbs read and orchestrate — never shown as separate choices:
  - `library/personas/` — 19 senior-role personas (architect, backend, frontend, mobile, dbarch, sql, data, ML, UX, devops, sre, security, reviewer, sdet, PM, TPM, support, tech-writer) + `adversarial-reviewer`, `tdd-guide`, `code-reviewer`.
  - `library/pipelines/` — `dev-pipeline`, `dev-ux-pipeline`, `dev-pipe-mobile`, `gammas-db-review`.
  - `library/understand/` — graphify + the 8-skill understand suite.
  - `library/product/`, `library/pm/` — product & delivery toolkits (incl. `product-skills`, `pm-skills` bundles).
  - `library/playwright/` — Playwright sub-skills.
  - `library/bootstrap/` — Go backend conventions, schema-design and validation guides, a stdlib Go renderer and the service / module / entity templates `bootstrap` drives.
  - `library/policies/` — cross-cutting rules every verb obeys (`service-tooling` + per-service detail, `work-tracking`, `task-chips`, `docs-split`).

## Service tooling order

Every external service is reached through the tooling **its own vendor recommends**, at the version a
skill pins. Rule + registry: `library/policies/service-tooling/SKILL.md`.

| # | Rung | Meaning |
|---|---|---|
| 1 | **Vendor-recommended CLI / SDK** | The vendor's own first-party tool |
| 2 | **Vendor's own MCP server** | First-party MCP from the same vendor |
| 3 | **Third-party MCP / client** | Someone else's wrapper over that service |
| 4 | **Raw REST / GraphQL API** | Last resort |

Highest available rung wins; step down only when a rung is absent or unauthenticated, never on a
business error. Never invent a vendor tool — when none is listed, the API *is* rung 1.

| Service | Rung 1 | Rung 2 | Rung 3 |
|---|---|---|---|
| Atlassian (Jira, Confluence, Bitbucket, JSM) | **[twg CLI](https://developer.atlassian.com/cloud/twg-cli/)** | Atlassian MCP | gitkraken MCP (Bitbucket only) |
| GitHub (PRs, issues, reviews, checks) | **[gh CLI](https://cli.github.com)** | gitkraken MCP | — |
| Playwright | `@playwright/test` | `@playwright/mcp` | — |
| BrowserStack | BrowserStack SDK | BrowserStack MCP | — |
| TestRail | REST API (no first-party tool) | — | — |

Absent rungs collapse rather than block: GitHub has no first-party MCP, so gitkraken moves up to rung
2; TestRail has no first-party tool at all, so its API *is* rung 1. Local git operations stay on
gitkraken MCP, outside the ladder.

When a lower rung ran, the response ends with one ultra-short FYI promoting the vendor tool — **at
most once per chat thread**, then never again in that thread. Per-service detail:
`library/policies/service-tooling/atlassian.md`, `…/github.md`.

## Work tracking

`library/policies/work-tracking/SKILL.md` puts four gates around every dev / design / product task:

- **Gate 1 — ground it.** Before implementing, ask once whether a **Jira ticket** and a **spec**
  (PRD, Confluence page, design doc, RFC) already exist. Both there → they become the acceptance
  contract. Missing → propose creating it (`superpowers:brainstorming` → `prd-architect` for the spec,
  `twg jira workitem create` for the ticket). Exempt: trivial fixes, read-only work, live incidents.
- **Gate 1b — open it.** As execution starts, on the level-0 Task (Initiative > Feature > Epic >
  Task): default the **assignee** to the current user, the **sprint** to the board's active one and
  the **original estimate** to a proposed bucket (a forecast, so it goes in up front) — each only if
  empty, asking before reassigning or moving out of a future sprint — then move the Task to
  in-progress and roll its parents up.
- **Gate 2 — track it.** One state change per real event from there on — in-review when the **PR is
  raised**, done only once it is merged and verified. Never a single Open→Done jump at the end; that
  leaves the board wrong for the whole duration and destroys cycle-time data.
- **Gate 3 — close it out.** On completion, default-on for any linked ticket: an ultra-short
  `Done` / `Watch for` / `Links` comment and effort logged.
  A **release Epic** is not Done until `main` carries an annotated tag `vX.Y.Z` on the merged commit,
  pushed to origin, with the version bump in the same PR and the Confluence Release history updated.

Estimates and worklogs are records people report on, so they are never fabricated — the estimate is a
proposed forecast stated back for correction, the worklog is derived from observed effort, both
rounded to a bucket. The terminal transition (Done, or **Resolved** for a Bug) and customer-facing
JSM comments still need an explicit yes.

Per-type workflow shapes and the transition-ID hazard — the same numeric ID means "Done" on one issue
type and "Blocked" on another, so transitions are always resolved by name — differ per Jira site, so map them on your own site before the first transition.

## Task chips — one Epic per session

`library/policies/task-chips/SKILL.md` governs requests that break into several work items. The
current session **decomposes** the work in Jira (Initiative > Feature > Epic > Task) and executes at
most **one Epic**; every other Epic, or each independent Task of a large Epic, is handed to its own
session as a task chip via the desktop app's `spawn_task` tool. That keeps each context window small.

Every chip prompt opens with a devkit verb (`/devkit:ship`, `support`, `product` or
`review`) so gates and personas are enforced in the child session, and is fully self-contained — Jira
keys, repo, base branch, acceptance criteria, files to leave alone, spec links — because the child
cannot read the spawning conversation. Chips run in **parallel** only when their files and outputs are
provably disjoint; otherwise the first chip's prompt ends by spawning the next once its work merges.

## Docs split — Jira mirrors implementation, Confluence mirrors the product

`library/policies/docs-split/SKILL.md`: Jira (Initiative > Feature > release Epics > Tasks) is the
implementation view, in engineering language. Confluence is the product view and does not copy Epics
or Tasks — one folder per product with **Product definition**, **Usage guide**, **Standard operating
procedures** and **Release history**, plus site-specific pages under a site folder. Release history is
the one deliberate join: one entry per version with tag, date, what became possible, schema version,
operator-facing changes, known limits and the Epic link, updated at every release close-out.
`product` writes the definition and release entries, `support` the usage guide and SOPs, `ship` the
release deltas.

## Install

```bash
/plugin marketplace add https://github.com/Vikasharista/developmentkit
/plugin install devkit@developmentkit
```

(Requires read access to the private repo and GitHub auth.) In cloud sessions, the environment setup
script installs it automatically when this repo is attached to the session.

## Attribution

Bundles adapted skills from the open-source [claude-skills](https://github.com/) collection (MIT, © 2025 Alireza Rezvani — see `NOTICE-third-party-LICENSE`) and an internal persona/pipeline set. `prd-architect` is an independent implementation of the PRD-authoring concept and depends on no external skill.
