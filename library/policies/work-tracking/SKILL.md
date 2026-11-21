---
name: work-tracking
description: Directive for grounding, tracking and closing out work — before any development, design or product task, confirm a Jira ticket and a spec doc exist (and propose creating them if not); place the work in the Initiative > Feature > Epic > Task hierarchy; on the Task, set assignee, sprint and original estimate before starting; move the ticket one state per real event as the work proceeds, rolling parents up; on completion post an ultra-short summary comment and log effort; a release Epic is Done only once main carries the pushed annotated tag vX.Y.Z and the Confluence release history is updated. Read at the start of any build/design/product task, whichever devkit verb started it.
---

# Work tracking directive

Four gates around every development, design or product task: **ground it before starting**
(Gate 1), **open it as work begins** (Gate 1b), **track it against actuals while it runs** (Gate 2),
**close it out after finishing** (Gate 3). All Atlassian calls go through
`${CLAUDE_PLUGIN_ROOT}/library/policies/service-tooling/` (twg CLI first, except for the calls listed
under *Known rung-1 gaps* in `atlassian.md` — estimate, assignee and sprint are all on that list).

**Every devkit verb applies this directive** — `ship`, `support`, `product`, `review`, `understand`,
`deliver`, and any verb added later. A verb may declare a gate inapplicable for its kind of work
(`review` is read-only, so it skips Gates 1b and 2; a live incident under `support` grounds after
the fact) but it does not get to substitute a different Jira workflow. The hierarchy, the three
Gate 1b defaults and the parent rollups below hold wherever a Task is being worked.

Workflow shapes, per-type state chains and transition IDs differ per Jira site and issue type. Map them on
your own site (walk real tickets with `transitions query`) before the first transition of a session. Status names in this policy (`To Do`, `In-Progress`,
`In Review`, `Done`, `RESOLVED`, `Blocked`, `UAT`) are examples: substitute your site's names.

---

## The hierarchy — where a ticket lives

PROJ issue types, with Jira's `hierarchyLevel` in brackets:

```
Initiative [3]
└── Feature [2]
    └── Epic [1]
        └── Task · BE Task · FE Task · QA Task · Design Task · PRD Task · Bug [0]
            └── Sub Task [-1]
```

| Level | One per | Naming | Example |
|---|---|---|---|
| **Initiative** | programme | the programme's name | `Example Programme` |
| **Feature** | product or site deliverable | `[<site or product>] <deliverable>` | `[Site A] Example Product` |
| **Epic** | incrementally viable, versioned release of the Feature | `[<short product>] vX.Y.Z - <what it makes viable>` | `[Example Product] v0.1.0 - M0 baseline` |
| **Task** (level 0) | unit of execution | plain summary, prefixed with the Epic's short tag | `[EP v0.1.0] Core pipeline: ingest, validate, store` |

Rules:

1. **Identify or create the Initiative** first. If the programme has one, use it; do not create a
   second.
2. **Identify or create the Feature** under it.
3. **Epics are versioned releases** under the Feature. If the user did not name an Epic, **ask which
   Epic the work belongs to, or propose the next version** (`vX.Y.Z` bumped from the latest Epic under
   the Feature) with one line on what that version makes viable. Do not park a Task directly under a
   Feature.
4. **Everything operational happens on level-0 items only.** Estimates, worklogs, sprint, assignee
   and status transitions go on the Task (or Bug). Create the Task under the Epic if none exists.
   **Never log time or set an estimate on an Epic, Feature or Initiative** — those states are derived
   from their children (see *Parent rollups* under Gate 2).
5. **Parent linking is the `parent` field**, at every level. Create works top-down:

   ```bash
   twg jira workitem create --space PROJ --type <Initiative|Feature|Epic|Task> \
     --summary "<per the naming column>" --parent <PARENT-KEY> \
     --description "<derived from the spec>" --description-format markdown \
     --labels a,b --output json          # the new key is "key" in the JSON
   ```

   Omit `--parent` only on an Initiative. `--reporter` is rejected on some projects (`Field
   'reporter' cannot be set. It is not on the appropriate screen`); omit it, the creating account
   becomes reporter anyway.

Worked example:

```
PROJ-100  Example Programme                                         Initiative
└── PROJ-101  [Site A] Example Product                              Feature
    └── PROJ-102  [Example Product] v0.1.0 - M0 baseline              Epic
        ├── PROJ-103  [EP v0.1.0] Core pipeline: …                   Task  ← estimated, assigned,
        ├── PROJ-104  …                                              Task    sprint 42, in review,
        ├── PROJ-105  …                                              Task    comment + worklog
        └── PROJ-106  …                                              Task
```

---

## Gate 1 — Ground the work (before implementing)

Ask, in **one** message bundled with any other clarifier:

1. **Is there a Jira ticket?** (key or URL)
2. **Is there a spec?** — PRD, Confluence page, design doc, RFC, ADR, or any written requirement.

Then:

- **Both exist** → read them via the ladder (`twg jira workitem get <KEY>`, `twg confluence page get`)
  and treat them as the source of truth for the acceptance contract. Flag any conflict between the
  ticket, the doc and the user's request instead of silently picking one. If the ticket is an Epic,
  Feature or Initiative, find or create the level-0 Task the session will actually execute.
- **Either missing** → **propose** creating it, placed in the hierarchy above. Do not create silently,
  and do not refuse to work.
- **User declines, or the task is exempt** → proceed, and say in the final summary that the work ran
  without a ticket/spec.

### Proposal routes

| Missing | Route |
|---|---|
| Spec / requirements | `superpowers:brainstorming` first (mandatory before creative work), then `devkit:prd-architect` for a PRD or `superpowers:writing-plans` for an implementation plan; publish to Confluence via the ladder |
| Jira ticket | Walk the hierarchy: Initiative → Feature → Epic → Task, creating only the missing levels with `twg jira workitem create … --parent <KEY>` as above; description derived from the spec — not from a guess |
| Both | Spec first; the ticket's description then comes from it |

### Exempt from Gate 1

Trivial one-line fixes • read-only investigation or questions • live incident work (ticket after, not
before) • work the user calls throwaway/experimental • a task the user has already scoped in this
thread with a ticket in hand.

Ask **once per task**, not per step. Never turn this into a nag.

---

## Gate 1b — Open the work (as execution starts)

Once, after grounding and **before** the first substantive change, on the **level-0 Task**. Four
actions: three field defaults, then the move to in-progress.

### 1b.a Three defaults on the Task

Read the Task first (`twg jira workitem get <KEY> --fields assignee,customfield_10021,timetracking,status,parent -o json --output-summary none`;
twg flattens the fields onto `data[0]`), then fill whatever is empty:

| | Field | If empty | If already set |
|---|---|---|---|
| **A** | **Assignee** (`assignee.accountId`) | Assign to the current user. `twg whoami -o json` returns it as `data.accountId`. | Someone else → **ask before reassigning**. The user may only be evaluating or commenting on that person's ticket. |
| **B** | **Sprint** (`customfield_10021`, numeric sprint id; `null` clears) | Add to the board's **active** sprint. | A **future** sprint → **ask before moving**. The active sprint → leave it. |
| **C** | **Original estimate** (`timetracking.originalEstimate`) | Propose from your own read of the task, round to a bucket (`15m` / `30m` / `1h` / `2h` / `4h` / `1d` / `2d`), write it, and **state the number** so the user can correct it. Do not stall the work to ask. | Leave it. It is someone's forecast, not yours to revise. |

Board and sprint discovery, once per session:

```bash
twg jira board query --project PROJ -o json --output-summary none                 # data.boards: PROJ 12 "PROJ Board" (scrum)
twg jira board sprints query --id 12 --state active -o json --output-summary none  # data.sprints[0].id is the sprint id to write
```

**The write path for all three is one call with one body, then a read-back.** This is a documented
rung-1 gap — see *Known rung-1 gaps* in `atlassian.md`: `twg jira workitem update` silently no-opped
on `--assignee`, `--sprint` and `--original-estimate` on these projects. The body:

```json
{"fields": {"assignee": {"accountId": "<from whoami>"},
            "customfield_10021": 1588,
            "timetracking": {"originalEstimate": "2h"}}}
```

Include only the fields you are actually setting. Where it goes depends on the installed twg
(`twg api --help` exits 1 with `unknown command 'api'` when the escape hatch is gone, as on 1.3.6):

- **twg with `api`** (older builds): `twg api "jira:/rest/api/3/issue/<KEY>" --method PUT --input body.json`.
  The flag is `--input <file|->`; there is no `--body`.
- **twg without `api`** (1.3.6 and later): Atlassian MCP `editJiraIssue` (cloudId
  `your-org.atlassian.net`) with the same `fields` object. It works but returns about 4 KB per call.
  The native `workitem update` flags are not re-verified on 1.3.6; do not use them for these three.

**Always read back after writing** — the no-op is silent and looks like success:

```bash
twg jira workitem get <KEY> --fields assignee,customfield_10021,timetracking -o json --output-summary none
```

`timetracking.originalEstimate` is what the Jira UI labels "Original estimate"; it is Jira's built-in
time-tracking field, not a `customfield_*`, and it is **not settable at create time** — always
create-then-edit.

### 1b.b Move the Task to in-progress

Starting work *is* the event, so this gate — not Gate 2 — owns the move to the in-progress state.
Walk the ticket there from wherever it sits, one hop at a time, resolving each transition **by the
name of the target status**. The helper does exactly that, with a read-back
after every hop and a non-zero exit when a hop is not offered:

```bash
python3 "${CLAUDE_PLUGIN_ROOT}/library/scripts/jira-walk.py" walk <KEY> "To Do" "In-Progress"
```

| Family | Walk |
|---|---|
| dev chain (Task, BE/FE/QA/Design/PRD Task), Sub Task | `Open → To Do → In-Progress` |
| Bug | `Open → To Do → Work In-Progress` |

`To Do` is a required waypoint, not a separate event — no workflow here has a forward edge that
skips it. Gate 2 picks up from in-progress onward, including the parent rollups.

### Skip Gates 1b and 2 when

There is no ticket — Gate 1 was exempt, or the user declined creating one. Nothing to estimate and
nothing to transition; say so in the summary and go straight to the work. Also skip on someone
else's ticket: never set an estimate, reassign, or move a state on a work item you are only reviewing
or commenting on.

---

## Gate 2 — Track state against actuals (during the work)

**Do not batch transitions.** A ticket that jumps from Open to Done in one stretch at the end records
nothing useful — the board was wrong for the whole duration, and cycle-time data is destroyed. Move it
when the thing actually happens.

Gate 1b already left the Task in-progress. From there:

| Real event | Target state (dev-chain families) | Bug equivalent |
|---|---|---|
| **PR raised** | `In Review` | `In Review` |
| Merged **and** verified by the user | `Done` | `RESOLVED` |
| Waiting on someone or something external | `Blocked` | `Blocked` |

`Blocked` does not exist on `Sub Task`, `Feature` or `Initiative` — those types have no blocked
state, so say the work is blocked rather than inventing a transition for it.

### Parent rollups — derived, never worked directly

Parents carry no estimate, worklog, sprint or assignee. Their status is rolled up from their Tasks,
by the same one-hop walk:

| Parent | Move it to | When |
|---|---|---|
| **Initiative**, **Feature** | `In-Progress` (via `To Do`) | the first Task anywhere beneath it starts (Gate 1b) |
| **Epic** | `PRD In-Progress` (via `To Do`) | its first Task starts (Gate 1b) |
| **Epic** | `UAT` (via `PRD Ready`, `In Dev`) | **all** its Tasks are in review |
| **Epic**, **Feature**, **Initiative** | `Done` | needs the same explicit yes as a Task (Gate 3c) |

Check the parent's current status before walking; if a sibling Task already moved it, there is
nothing to do. The `UAT` rollup needs a read of every sibling
(`twg jira workitem query --jql "parent = <EPIC>" --fields status -n 50 -o json --output-summary none`), not an
assumption from the one Task in hand.

Rules:

- **One hop per event, in order.** No workflow here has a *forward* edge that skips a state, so a
  two-state move is genuinely two calls with a re-discovery between them. (Several **global**
  transitions do jump from anywhere — `151` back to `Open`, `141` to `INVALID` — but those are
  escapes, never forward progress. Do not mistake one for a shortcut.)
- **Resolve transitions by name, never by memorised ID.** IDs are not stable across issue types —
  the same numeric ID can mean "Done" on one issue type and "Blocked" on another.
- **Re-discover before every hop** and **confirm the landing status** before the next one. `jira-walk.py`
  does both; if you transition by hand, `twg jira workitem transitions query --id <KEY> -o json` is
  the discovery call (match `toName` to the target status) and
  `twg jira workitem transition --id <KEY> --transition-id <id from that set>` is the hop.
- If a hop you expect is missing from the discovered set, say so rather than reaching for a
  same-numbered transition that happens to be offered. The helper exits non-zero in that case and
  fires nothing further.
- Bugs terminate at `RESOLVED` and have no `Done` at all.

Moving a ticket **forward through in-progress and review states is default-on** — it is a record of
what happened. Only the final state needs a yes (see below).

---

## Gate 3 — Close it out (on completion)

Default-on for any completed task that has a Jira ticket. All via the ladder.

### 3a. Summary comment — always

`twg jira workitem comment create --issue-id <KEY> --body <text> --body-format markdown`
(MCP fallback: `addCommentToJiraIssue`).

Ultra short. Six lines maximum:

```markdown
**Done:** <one line — what actually changed>
**Watch for:** <risks, edge cases, follow-ups — omit the line if genuinely none>
**Links:** PR • doc • test run
```

Not the transcript, not the reasoning, not a changelog of every file.

### 3b. Effort log — on completion

`twg jira workitem worklog add --issue-id <KEY> --time-spent "<2h 30m>" --comment "<what the time
went on>"` (MCP fallback: `addWorklogToJiraIssue`). Jira's standard time format: `2h 30m`, `1d`, `45m`.

Read it back with `twg jira workitem worklog query --issue-id <KEY>`. Note that the issue's top-level
`timespent` field reads `null` under twg's default field set even when the worklog exists — check
`timetracking.timeSpent` or the worklog query, not `timespent`.

### 3c. Final transition — needs an explicit yes

`Done` (or `RESOLVED` for a Bug) is a **status claim**, not a note. Propose it; do not fire it
unprompted. The user's confirmation that the work is verified *is* the yes.

Everything before it — `To Do`, in-progress, review — is default-on under Gate 2.

### 3d. Release close-out — when the Epic is a release

A release Epic (`[<product>] vX.Y.Z - <what it makes viable>`) is **not Done until `main` carries an
annotated git tag `vX.Y.Z` on the merged commit and the tag is pushed to origin**. Tag message: one
line, `vX.Y.Z - <what it makes viable>`, plus the Epic key. Agreed with TJ, 2026-09-19.

- **The version bump ships in the same release PR** — `pyproject.toml`, the package `__version__`, or
  the stack's equivalent — so `<cli> --version` and any `model_version` / `schema_version` field agree
  with the tag.
- **Order at close-out**, each step only after the previous one is verified:
  1. PR merged.
  2. Version bump present on the merged commit.
  3. Tag created and pushed.
  4. Confluence **Release history** entry added — `${CLAUDE_PLUGIN_ROOT}/library/policies/docs-split/SKILL.md`.
  5. Epic comment with the tag and the commit SHA.
  6. Epic to `Done` — the explicit yes of 3c still applies.
- Local git via gitkraken MCP where available. The MCP has no tag command, so the shell is acceptable
  for the tag itself:

  ```bash
  git tag -a vX.Y.Z <sha> -m "vX.Y.Z - <what it makes viable> (PROJ-NNN)"
  git push origin vX.Y.Z
  ```

Worked example: `example-product` `v0.1.0` on its merged commit, Epic PROJ-102.

Once `vX.Y.Z` is tagged, the next alpha build is `vX.Y.(Z+1)-alpha-1`, never a continuation of an
older alpha series.

### Flag names — get these right

The flag names are inconsistent across the CLI's own subcommands. Read `--help` rather than assuming:

| Command | Ticket flag |
|---|---|
| `workitem get` | positional (`get AF-1926`) |
| `workitem update` / `transition` / `transitions query` | `--id` |
| `workitem comment create` | `--issue-id` |
| `workitem worklog add` / `worklog query` | `--issue-id` |
| `workitem changelog query` | `--issue-id` |
| `workitem create` | `--space` (project, not a ticket) |

### Numbers are records — do not fabricate them

Estimates and worklogs feed capacity planning, billing and performance reviews. Real humans report on
them.

- The **estimate** (Gate 1b) is a forecast made before the work: propose it from your own read of the
  task, round to a bucket, and state it so the user can correct it.
- The **worklog** (Gate 3b) is an observation: base it on effort actually seen in the session. If
  elapsed effort is not determinable, **ask the user for the number**. Never guess, never inflate,
  never pad to look busy.
- No false precision. `2h`, not `1h 47m`.
- **State every value written** in the response.

### Skip Gate 3 when

No ticket exists (and the user declined creating one) • the task was read-only • the user opted out •
you only reviewed someone else's ticket (a comment is fine, a worklog is not — log time on your own
work only).

### Still needs an explicit yes

Default-on covers internal comments, the estimate, the worklog, the Gate 1b defaults (self-assign an
unassigned Task, add an unsprinted Task to the active sprint), forward transitions up to and
including review, and the parent rollups short of Done. These do not:

- **Customer-facing JSM comments** — external visibility.
- **The terminal transition** — `Done` / `RESOLVED`, per Gate 3c, on the Task and on any parent.
- **Cancelling or invalidating** a ticket (`Cancelled`, `INVALID`, `DEFERRED`) — a judgement
  about someone's work item, not a progress record.
- **Reassigning** a ticket that already has an assignee, or **moving** one out of a future sprint —
  Gate 1b asks first.
- Editing anything else on a ticket (summary, labels, priority, parent of an existing item).
