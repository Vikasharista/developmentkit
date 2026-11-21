---
name: docs-split
description: Directive for where documentation lives — Jira mirrors the implementation (Initiative > Feature > release Epics > Tasks, in engineering language); Confluence mirrors the product (Product definition, Usage guide, SOPs, Release history, site pages) and does not copy the Epic or Task structure. Release history is the one deliberate alignment point between the two. Read before writing or updating any Confluence page or Jira description for a product.
---

# Docs split directive — Jira mirrors implementation, Confluence mirrors the product

Two systems, two views, one join. Agreed with TJ, 2026-09-19. All Atlassian calls go through
`${CLAUDE_PLUGIN_ROOT}/library/policies/service-tooling/` (twg CLI first).

## Jira — the implementation and tracking view

Initiative > Feature > release Epics > Tasks, per
`${CLAUDE_PLUGIN_ROOT}/library/policies/work-tracking/SKILL.md`. Epics are versioned releases; the
language is engineering. Nothing in Jira tries to describe the product to an operator.

## Confluence — the product view

Confluence does **not** mirror Epics or Tasks. One folder per product with a standard page set, plus
one folder per site for pages that vary by site:

| Page | Holds |
|---|---|
| **Product definition** | What it is, for whom, what it deliberately does not do, event or data model in plain language, architecture at a glance, glossary |
| **Usage guide** | Install, configure, read the outputs, troubleshoot |
| **Standard operating procedures** | Numbered SOPs, each with purpose, trigger, roles, steps, evidence, escalation |
| **Release history** | One entry per version — see below |
| **Site pages** (site folder) | Registries, calibration values, anything site-specific |

Feasibility studies and decision documents live under the **programme** folder and are linked from
the Initiative and Feature descriptions in Jira.

## Release history — the one alignment point

The only place the two systems deliberately meet. One entry per version, **oldest first**, each with:

- tag and date
- what became possible, in product language
- schema or data-model version
- operator-facing changes
- known limits
- link to the Jira release Epic

Updated on **every** release, as a step of release close-out (`work-tracking/SKILL.md`, Gate 3d).
Future versions may appear as **Planned**, two or three sentences each.

## Who writes what

| Devkit verb | Writes |
|---|---|
| `product` | Product definition; Release history entries |
| `support` (xi-techwriter) | Usage guide; Standard operating procedures |
| `ship` | Release history entry and Usage guide deltas, at release close-out |

## Worked example

Confluence space **PROJ**, folder *Example Programme* > *Example Product* (product pages) and
*Site A* (site pages). Jira: PROJ-100 (Initiative) > PROJ-101 (Feature) > PROJ-102 and following
(release Epics).

## Red flags

| Thought | Reality |
|---|---|
| "One Confluence page per Epic" | No. Epics are releases; the product view carries one Release history entry per version and nothing else per Epic. |
| "Paste the Task list into the Usage guide" | Task language is engineering. Rewrite as what the operator can now do. |
| "Skip Release history, the tag says it all" | The tag is for git. The entry is where the product side learns what changed. |
