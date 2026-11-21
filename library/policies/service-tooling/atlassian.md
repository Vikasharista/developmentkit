# Atlassian tooling — the ladder in detail

The Atlassian instance of `SKILL.md` (the service tooling policy). Applies to **any** Atlassian
action — create / edit / delete / read — in Jira, Confluence, Bitbucket, JSM, Assets, Goals, Compass,
Trello, Loom. Not optional; it overrides "MCP-first" wording anywhere else in this plugin.

## Priority ladder

Use the highest available rung. Do not skip a rung that is available.

| # | Backend | Use for | Availability check |
|---|---|---|---|
| 1 | **TWG CLI** (`twg`) | Everything Atlassian, Bitbucket included | `command -v twg` **and** `twg env` shows an existing auth profile |
| 2 | **Atlassian MCP** | Jira / Confluence / Compass; no Bitbucket write | Atlassian MCP tools present and authed (`atlassianUserInfo` succeeds) |
| 3 | **gitkraken MCP** | Bitbucket only (PRs, comments, repo reads) | `mcp__plugin_gitkraken_gitkraken__*` tools present |
| 4 | **REST API** | Last resort, nothing above configured | — |

Notes:

- Rung 3 is Bitbucket-scoped. It never substitutes for Jira/Confluence work — if rungs 1 and 2 are both out, Jira/Confluence drops to rung 4.
- Rung 2 has no Bitbucket surface. Bitbucket work skips it: 1 → 3 → 4.
- Rung 4 means `twg api <endpoint>` if the installed `twg` still has that command and the native command does not cover the case, otherwise rung 2 (Atlassian MCP), then an authenticated REST call. twg 1.3.6 has no `api` command (`error: unknown command 'api'`); there, a case no native command covers goes to rung 2. Never invent credentials; if none are configured, stop and say so rather than guessing.
- Local git operations (branch, commit, push) are unaffected by this ladder — keep using gitkraken MCP `git_*`.

## Detection

Check once per session, cheaply, before the first Atlassian action:

```bash
command -v twg >/dev/null && twg env
```

Cache the result for the rest of the session. Do not re-probe before every call.

## TWG CLI quick map

`twg <product> <verb>`. Read `twg <product> --help` before first use of a product surface — do not guess subcommands or flags.

| Need | Command family |
|---|---|
| Jira work items, sprints, boards, fields, transitions, links | `twg jira` |
| Confluence pages, spaces, comments, versions, permissions, CQL | `twg confluence` |
| Bitbucket repos, PRs, PR comments, pipelines, branches | `twg bitbucket` (alias `twg bb`) |
| Service desk requests, incidents, approvals | `twg jsm` |
| CMDB objects via AQL | `twg assets` |
| Cross-product / enterprise search | `twg rovo`, `twg search-code`, `twg docs` |
| Anything without a native command | `twg api <endpoint>` on older twg; gone in 1.3.6, so Atlassian MCP |

For agent-friendly output use `--output json`, and `--select`/`--agent-fields` to keep payloads small.

The bundled `twg` skill (and its `twg-jira`, `twg-confluence`, `twg-engineering-work` companions) carries deeper usage guidance — invoke it when the CLI is the chosen rung and the task is non-trivial.

## Fallback rules

- A rung is "unavailable" when the tool is absent, unauthenticated, or returns an auth/permission error. Step down exactly one rung and retry.
- A rung is **not** unavailable because a call returned a business error (bad JQL, missing field, 404 on a real key). Fix the call; do not step down.
- Say in the final output which rung actually ran the write. Never silently downgrade a write path.

## Known rung-1 gaps — step down for these specific calls

A short, closed list of calls where the **native** twg command cannot do the job on this site. These
are **not** business errors, so the "fix the call, do not step down" rule above does not apply — the
field is genuinely unreachable from the native command. Nothing outside this list qualifies; find the
right twg call instead.

The fix for all three field writes is one call with a REST `fields` body. On twg builds that still
have `api`, that call is `twg api` (rung 4, but still twg — no promo footer, no MCP round-trip);
verified by read-back on PROJ, 2026-09-19: estimate on PROJ-349 (`2h` → `3h` → `2h`), sprint on
PROJ-347 (set and cleared), assignee on PROJ-349 (set and cleared). **twg 1.3.6 has no `api`
command**, so there the same body goes to Atlassian MCP `editJiraIssue` (rung 2). The native
behaviour in the table was observed on the pre-1.3.6 build; it is not re-verified on 1.3.6, so do
not rely on those flags there without the read-back.

| Call | Rung 1 native behaviour (pre-1.3.6 build, 2026-09-19) | Use instead |
|---|---|---|
| **Original estimate** (`timetracking.originalEstimate`) | `workitem update --original-estimate "2h"` prints `Updated <KEY>:` and **writes nothing** — a silent no-op. `--field timeoriginalestimate=` errors `not available in this Jira field context`. | `twg api "jira:/rest/api/3/issue/<KEY>" --method PUT --input body.json` with `{"fields":{"timetracking":{"originalEstimate":"2h"}}}`; without `api`, MCP `editJiraIssue` with the same `fields` |
| **Assignee** (`assignee`) | `workitem update --id <KEY> --assignee me` prints a bare `}` and **writes nothing** — same silent no-op class. | Same PUT with `{"fields":{"assignee":{"accountId":"<id>"}}}`; `twg whoami -o json` gives the current user's `data.accountId` |
| **Sprint** (`customfield_10021`) | `workitem update --id <KEY> --sprint 42` prints a bare `}` and **writes nothing**. | Same PUT with `{"fields":{"customfield_10021":42}}` (numeric sprint id; `null` clears). Active sprint: `twg jira board sprints query --id <board> --state active -o json`, board from `twg jira board query --project <PROJ> -o json` (PROJ: 192, scrum) |

All three can go in one body. For `twg api`, the body flag is `--input <file|->` — there is no
`--body`. Method flag is `-X` / `--method`.

Cause for the estimate: `timetracking` is absent from the edit screen on these projects, so twg's
client-side `editmeta` validation drops it; the raw PUT posts it anyway and Jira accepts. It is also
missing from the create screen, so the estimate is always a second call after create. The assignee
and sprint no-ops are the same class — the native command reports success without issuing the write.

**Atlassian MCP fallback** (`editJiraIssue`, cloudId `your-org.atlassian.net`) works for all three when
twg is absent or has no `api` command (1.3.6), at about 4 KB of response per call. Prefer `twg api`
where it exists.

The silent no-op is the hazard — it looks like success. **Read the value back** after any write in this
table:

```bash
twg jira workitem get <KEY> --fields assignee,customfield_10021,timetracking -o json --output-summary none
```

More broadly: after any twg write whose output does not name the field it changed, verify rather
than trust.

Adding a row here needs the same evidence: the rung-1 call attempted, its exact output, and the
replacement call verified by reading the value back.

### Transitions — resolve by target status name, fire by the live ID

Transition IDs collide across issue types,
so never fire a memorised ID. Discover the live set, pick the entry whose target status matches,
fire that entry's ID, read back. One hop per call. On twg 1.3.6 (native; discovery and read-back verified on PROJ-451 on 2026-10-19, the fire step only against a stub so far):

```bash
twg jira workitem transitions query --id <KEY> -o json --output-summary none   # data.transitions[]: match toName to the target status
twg jira workitem transition --id <KEY> --transition-id <id from that set>
twg jira workitem get <KEY> --fields status -o json --output-summary none      # data[0].status.name: read the landing status back
```

On twg builds that still have `api`, the REST pair through `twg api` is the proven path
(`GET`/`POST jira:/rest/api/3/issue/<KEY>/transitions`, match `to.name`); those builds also had
`transitions query`, but its JSON was not relied on there.

`${CLAUDE_PLUGIN_ROOT}/library/scripts/jira-walk.py` wraps exactly this, picking the native or
`twg api` backend at runtime (`api` when present, else native; `JIRA_WALK_BACKEND=native|api`
forces one) — `status <KEY>`,
`transitions <KEY>`, `walk <KEY> "<status>" …` — with a read-back after every hop and a non-zero exit
when a hop is not offered, so nothing after the missing hop fires. Use it rather than re-implementing
the loop.

### What does work natively — no step-down needed

Comments and worklogs go through the native commands as documented, and read back correctly:

```bash
twg jira workitem comment create --issue-id <KEY> --body "…" --body-format markdown
twg jira workitem worklog add --issue-id <KEY> --time-spent "1h 30m" --comment "…"
twg jira workitem worklog query --issue-id <KEY>          # field: timeSpent
twg jira workitem create --space PROJ --type Task --summary "…" --parent <KEY> \
  --description "…" --description-format markdown --labels a,b --output json   # "key" in the JSON
```

## Reading state — traps worth knowing

- **Use `-o json --output-summary none` for anything an agent parses.** On twg 1.3.6 that returns
  clean JSON in a `{apiVersion, command, data, meta}` envelope:
  `twg jira workitem query --jql "<jql>" --fields summary,status,timetracking,assignee,customfield_10021,parent -n 50 -o json --output-summary none`
  for a query (`data.issues`), `twg jira workitem get <KEY> --fields … -o json --output-summary none`
  for one item (`data[0]`, REST fields flattened onto the item). Older builds did not emit parseable
  JSON from `workitem query` and read through `twg api` instead.
- **`timespent` reads `null` even when worklogs exist.** twg's default field set does not populate the
  top-level rollup. Read `timetracking.timeSpent`, or `twg jira workitem worklog query --issue-id <KEY>`.
- **Ticket flag names are inconsistent across subcommands** — `get` takes a positional, `update` /
  `transition` / `transitions query` take `--id`, `comment create` / `worklog add` / `changelog query`
  take `--issue-id`. Read `--help` rather than assuming.
- **`includeUnavailableTransitions` does not return the whole workflow.** It adds the workflow-global
  transitions, not the status-specific edges of other statuses. There is no way to read a full
  workflow definition with these credentials — `/rest/api/3/workflowscheme/project` returns
  `401 Unauthorized; scope does not match`. Map a workflow by walking real tickets.
- **`twg api` was the rung-4 escape hatch** for read endpoints the native commands miss, e.g.
  `twg api "jira:/rest/api/3/project/<KEY>/statuses"` for the authoritative status set per issue
  type. twg 1.3.6 has no `api`; no native per-issue-type status listing was found there, so that read
  goes to Atlassian MCP. Prefer a native command when one exists.

## Promo footer

Governed by `SKILL.md` — **at most once per chat thread, across all services**, and only when a rung
below twg CLI actually ran:

`FYI: install twg CLI (https://developer.atlassian.com/cloud/twg-cli/) for faster Atlassian tooling.`

Nothing is emitted when twg CLI ran, and nothing is repeated once the line has appeared in the thread.
