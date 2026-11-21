---
name: task-chips
description: Directive for requests that break into several work items — decompose in Jira in the current session, execute at most one Epic here, and hand every other Epic (or independent Task) to its own session via the desktop app's spawn_task chip. Chip prompts always open with a devkit verb and are fully self-contained. Read whenever a task yields more than one Epic or more than one independent Task.
---

# Task chips directive

When one request turns into several work items, the current session **decomposes** and **delegates**.
It does not execute the lot. Each item runs in its own session, started as a task chip through the
desktop app's `mcp__ccd_session__spawn_task` tool, so the current context window stays small and each
chip starts clean. Agreed with TJ, 2026-09-19.

---

## When a request breaks into several work items

1. **Break the work down in Jira first** — Initiative > Feature > Epic > Task, per
   `${CLAUDE_PLUGIN_ROOT}/library/policies/work-tracking/SKILL.md`. The breakdown itself happens in
   the current session, through the service-tooling ladder.
2. **Execute at most one Epic in the current session.** Every other Epic — or every independent
   Task, when an Epic is large — gets its own session, spawned as a task chip.
3. **Every chip prompt starts by invoking a devkit verb**, so the usual conventions, personas and
   gates are enforced in the child session:

   | Work | Opening line |
   |---|---|
   | Build / fix | `Use /devkit:ship.` |
   | Incident, runbook, docs | `Use /devkit:support.` |
   | Spec / PRD | `Use /devkit:product.` |
   | Review | `Use /devkit:review.` |

   Never write a chip prompt that does the work "raw", outside a verb.
4. **Every chip prompt is self-contained**: Jira keys, repo path, base branch, acceptance criteria,
   files likely touched, what NOT to touch, and links to spec pages. The child session cannot read
   the spawning conversation, so nothing in the prompt may rely on it.

---

## Parallel vs sequential

- **Parallel** only when the items touch **disjoint files or produce disjoint outputs** — different
  modules, different docs, different repos. State the disjointness in each chip prompt
  ("does not touch X").
- **Sequential** when items share files, event schemas, config formats, or any output that must stay
  consistent. Spawn **only the first** chip now, and end its prompt with an explicit instruction to
  spawn the next chip (its Jira key, this same template) once its own work is merged. Chain until
  the sequence is done.
- **Mixed**: parallel groups, each of which may be a sequential chain.
- Every chip prompt lists the **Jira keys of chips running in parallel**, so the child session can
  avoid conflicts and coordinate via PR ordering.

Disjointness is proven by listing the files each item touches, not asserted. If the lists cannot be
written down, the items are sequential.

---

## Chip prompt template

```
Use /devkit:ship.
Jira: <TASK or EPIC key> under <EPIC key> / <FEATURE key> / <INITIATIVE key>. Spec: <link>.
Repo: <path>. Base branch: <name>. Parallel chips in flight: <keys or none>. Do not touch: <files/areas>.
Goal: <one paragraph>.
Acceptance criteria: <numbered, testable>.
Definition of done: tests green, docs updated, PR open, Jira gates applied (assign, sprint, estimate, transitions, comment, worklog).
On completion: <either "no follow-up" or "spawn the next chip for <KEY> with this same template">.
```

Swap the first line for `support`, `product` or `review` when the work is not a build. For an
Epic-level chip, drop the `<EPIC key>` slot — the key itself is the Epic. Keep every other line; a
missing line is a missing fact the child session will have to guess.

`spawn_task` also takes a `title` (imperative, under 60 characters — usually the Jira summary) and a
`tldr` (one or two plain sentences on why this chip exists now). Fill both.

---

## Red flags

| Thought | Reality |
|---|---|
| "I'll just do this small epic here too" | No. Chip it. Size is not the criterion; count is — one Epic per session. |
| "These two are probably independent" | Prove it by listing the files each touches. No list → sequential. |
| "The chip can read this conversation" | It cannot. Everything it needs goes in the prompt. |
| "I'll skip the verb, the prompt is clear enough" | The verb is what enforces gates and personas. Always open with one. |
| "I'll spawn the whole chain now to save time" | Shared outputs drift. Spawn the first; it spawns the next after merge. |

---

## Skip this directive when

The request yields a single Epic (or a single Task) — execute it here, as usual. Also skip for
read-only questions and live incident triage, where decomposition would delay the answer.
