---
name: deliver
description: Program and project delivery — sprint planning and velocity, portfolio and risk management, agile/scrum coaching, cross-functional coordination, meeting analysis, stakeholder communications, and Jira/Confluence operations. Use for TPM/PM delivery work, standups, retros, roadmaps-to-execution, status reports, and anything touching Jira boards, JQL, or Confluence spaces. Routes internally by need.
---

# deliver — program & delivery execution

PM/TPM entrypoint. Routes to `${CLAUDE_PLUGIN_ROOT}/library/pm/` and `library/personas/nu-tpm`.

**Atlassian tooling** — read `${CLAUDE_PLUGIN_ROOT}/library/policies/service-tooling/` (`SKILL.md` + `atlassian.md`) before the first Jira/Confluence/Bitbucket action. Priority: **twg CLI → Atlassian MCP → gitkraken MCP (Bitbucket only) → REST API**. Read each tool/command schema before first call, and end the response with the required FYI footer.

## Route by intent

| User need | Route to |
|---|---|
| Cross-functional program coordination, eng ops, dependency/risk mgmt | `library/personas/nu-tpm` |
| Portfolio health, quantitative risk, enterprise delivery | `library/pm/senior-pm` |
| Sprint planning, velocity, retros, agile coaching | `library/pm/scrum-master` |
| Backlog, user stories, acceptance criteria, sprint execution | `library/pm/agile-product-owner` |
| Analyze meeting transcripts → patterns + coaching + action items | `library/pm/meeting-analyzer` |
| 3P updates, newsletters, incident/leadership comms | `library/pm/team-communications` |
| Roadmap → release narrative / changelog | `library/product/roadmap-communicator` |
| Jira: projects, JQL, workflows, fields, automation, reporting | `library/pm/jira-expert` (over the tooling ladder) |
| Confluence: spaces, KBs, docs, permissions, hierarchy | `library/pm/confluence-expert` (over the tooling ladder) |
| Jira/Confluence templates, blueprints, layouts | `library/pm/atlassian-templates` |
| Atlassian admin: users, permissions, security, integrations | `library/pm/atlassian-admin` |
| General multi-tool delivery coordination | `library/pm/pm-skills` (8-sub-skill bundle) |

## Related flows

- Turning meeting notes into Jira tasks → `meeting-analyzer` + the top available rung (`twg jira create`, else Atlassian MCP `createJiraIssue`).

Infer the right route; don't hand the user a menu.

**Work tracking** — `${CLAUDE_PLUGIN_ROOT}/library/policies/work-tracking/SKILL.md` governs all four gates: ground a task against a ticket + spec before starting (Gate 1), set the original estimate and open the ticket as work begins (Gate 1b), move it one state per real event — in-review when the PR is raised (Gate 2) — and on completion post the short summary comment and log effort (Gate 3). Never fabricate estimate or worklog numbers.

**Auth work in a breakdown** — when the work touches auth, permission seeding for each environment and any shared auth-library change are their own Tasks; the app Task is blocked by the library Task.

**Breaking a request into several Epics** — `${CLAUDE_PLUGIN_ROOT}/library/policies/task-chips/SKILL.md`: the Jira breakdown (Initiative > Feature > Epic > Task) happens here, but at most one Epic is executed in this session; every other Epic, or each independent Task of a large one, is spawned as a self-contained task chip that opens with a devkit verb.
