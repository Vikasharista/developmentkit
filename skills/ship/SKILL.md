---
name: ship
description: Ship a feature or fix end-to-end from a single task description — plan, code, and review in an iteration loop until acceptance criteria and quality gates pass, then hand back a polished, verified result. Use when the user gives a build/implement/fix task and wants it taken to done without repeated back-and-forth. Routes internally to architect/backend/frontend/mobile/reviewer personas by stack; no need to name a persona.
---

# ship — task in, verified result out

The goal-loop entrypoint. User gives a task; you drive it to done. Minimize to-and-fro: derive the acceptance contract up front, then loop until it is provably met.

`${CLAUDE_PLUGIN_ROOT}/library/` holds the personas and pipelines this skill orchestrates. Read the relevant files; do not expose them as separate choices to the user.

## 0 — Acceptance contract (do this first, once)

**Ground the work first** — `${CLAUDE_PLUGIN_ROOT}/library/policies/work-tracking/SKILL.md`, Gate 1: in one message, ask whether a Jira ticket and a spec (PRD / Confluence / design doc) already exist. Both exist → read them and use them as the contract. Either missing → propose creating it (`superpowers:brainstorming` → `prd-architect` for the spec; `twg jira workitem create` for the ticket). User declines or the task is exempt (trivial fix, live incident, read-only) → proceed and say so in the summary.

**More than one Epic (or independent Task) falls out of the contract?** — `${CLAUDE_PLUGIN_ROOT}/library/policies/task-chips/SKILL.md`: decompose in Jira here, execute at most one Epic in this session, and spawn every other item as a self-contained task chip that opens with a devkit verb.

Establish the exit criteria before coding, so the loop has a definition of done:

- If a PRD / acceptance criteria already exist (from `product` or `prd-architect`), use them verbatim.
- Else derive a short acceptance contract with `${CLAUDE_PLUGIN_ROOT}/skills/prd-architect/SKILL.md` (lightweight mode): functional requirements, non-functional constraints, and testable acceptance criteria. Keep it tight for small tasks.
- Confirm the contract with the user only if the task is ambiguous or high-risk; otherwise state it and proceed.
- **Touches auth?** — the task adds or changes an API operation, route, screen or role, or touches login, sessions, permissions or service credentials: add an access matrix (actor → role → permission, with deny cases) to the acceptance contract and pull in `zeta-security`.

**Open the ticket** — same policy, Gate 1b, before the first substantive change, on the **level-0 Task** (Initiative > Feature > Epic > Task; if the ticket in hand is an Epic or higher, find or create the Task under it, and if no Epic was named, ask which one or propose the next version). Three defaults, each only if empty: **assignee** → the current user (someone else already on it → ask before reassigning); **sprint** → the board's active sprint (a future sprint → ask before moving); **original estimate** → propose the number from your read of the contract above, round to a bucket, state it so the user can correct it. All three go in one write with one `fields` body, since `twg jira workitem update` silently no-ops on every one of them: `twg api … --method PUT --input body.json` where the installed twg still has `api`, else Atlassian MCP `editJiraIssue` (twg 1.3.6 dropped `api`; see Gate 1b) — then read the values back with `twg jira workitem get`. Then walk the Task to in-progress one hop at a time with `library/scripts/jira-walk.py`, and roll the parents up (Initiative and Feature to in-progress, Epic to PRD in-progress) if this is the first Task under them to start.

## 1 — Pick the pipeline (auto, by scope)

**New Go backend service, or a new entity / module in an existing service built the same way?**
Hand over to `${CLAUDE_PLUGIN_ROOT}/skills/bootstrap/SKILL.md`: it runs the entity and schema dialogue, generates
every layer from templates and then rejoins this loop for review, gates and the PR.

Detect scope from the task + repo, then adopt the matching pipeline from `library/pipelines/`:

| Scope | Pipeline |
|---|---|
| Backend / API / infra, or backend + simple UI | `dev-pipeline` |
| Has non-trivial UX/design work | `dev-ux-pipeline` (adds Theta UX phase) |
| Native mobile / IoT / kiosk / edge client | `dev-pipe-mobile` (adds Lambda phase) |

The pipeline defines the phase order. Personas live in `library/personas/`:
Alpha (`alpha-architect`) plans · Beta (`beta-coder`) backend+simple UI · Pi (`pi-frontend`) complex frontend · Lambda (`lambda-mobile-iot`) mobile/IoT · Theta (`theta-ux`) UX · Gamma (`gamma-reviewer-tester`) review/QA. Pull in `dbarch`/`sql-expert` when the data layer changes, `zeta-security` for security-sensitive work (always when the task touches auth), `delta-devops`/`mu-sre` for infra/reliability.

## 2 — The loop (max 6 iterations)

Run the pipeline iteratively. Each iteration: plan/revise → implement → review. Reviewer (Gamma) classifies findings CRITICAL / WARNING / INFO and attributes each to the owning persona; the next iteration fixes only what's attributed.

**Harden the verify phase** (this is what removes the back-and-forth):
- Enforce tests via `library/personas/tdd-guide` — new/changed paths must have tests; run them.
- Add an adversarial pass with `library/personas/adversarial-reviewer` on the diff each iteration.
- For E2E/UI, generate + run checks via the `playwright-pro` skill.
- Cross-check every acceptance criterion from step 0 explicitly.

**Exit when:** Gamma verdict PASS (0 critical, 0 warnings) **and** every acceptance criterion is demonstrably satisfied (tests/output shown), **or** 6 iterations reached → stop and report the remaining punch list. Do not claim done without evidence (`superpowers:verification-before-completion` applies).

## 3 — Git / PR

When the work should land: local git via gitkraken MCP (`mcp__plugin_gitkraken_gitkraken__git_*`) — never curl.

Anything touching a hosted service follows `${CLAUDE_PLUGIN_ROOT}/library/policies/service-tooling/`: **GitHub → `gh` CLI, then gitkraken MCP** (`github.md`); **Atlassian → twg CLI, then Atlassian MCP, then gitkraken for Bitbucket** (`atlassian.md`); REST last in both cases.

**Once the PR is raised**, move the ticket to in-review — Gate 2 of the work-tracking policy. That is the event the state records; do not defer it to close-out.

**Release Epic?** Its close-out needs an annotated tag `vX.Y.Z` on the merged commit, pushed to origin, with the version bump in the same PR — Gate 3d (*Release close-out*) of the work-tracking policy. `git tag -a` / `git push origin <tag>` through the shell is fine where the gitkraken MCP lacks a tag command.

## Final output

```
## Ship Summary
- Acceptance criteria: <n met / n total>
- Iterations: N
- Status: PASS | STOPPED (max iterations)
- Evidence: tests run, key outputs
- Remaining (if any): punch list
```

**Close out** — `${CLAUDE_PLUGIN_ROOT}/library/policies/work-tracking/SKILL.md`, Gate 3: when the task had a Jira ticket, post the ultra-short summary comment (`Done` / `Watch for` / `Links`) and log effort based on effort actually observed, rounded to a bucket, stating what you wrote. The terminal transition — Done, or **Resolved** for a Bug — needs an explicit yes; the user confirming the work is verified is that yes.

Release close-out also touches Confluence: `ship` adds the **Release history** entry and any **Usage guide** delta — `${CLAUDE_PLUGIN_ROOT}/library/policies/docs-split/SKILL.md` (Jira mirrors implementation, Confluence mirrors the product).

If a service action ran on a lower rung than the vendor-recommended tool, and no promo line has appeared yet in this thread, close with the one-line footer from `${CLAUDE_PLUGIN_ROOT}/library/policies/service-tooling/SKILL.md`. **At most once per chat thread** — skip it otherwise.
