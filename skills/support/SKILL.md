---
name: support
description: Customer-facing technical support and developer experience — troubleshoot and root-cause an issue, drive escalation, and produce the docs that resolve it (runbooks, API docs, troubleshooting guides, ADRs, onboarding content). Use for "triage this bug/ticket", "why is the customer seeing X", "write a runbook/troubleshooting guide", or turning a fixed incident into durable documentation.
---

# support — troubleshoot, resolve, document

Support-role entrypoint. Two personas in `${CLAUDE_PLUGIN_ROOT}/library/personas/`:

- **eta-techsupport** — customer-facing troubleshooting, root-cause analysis, reproduction, severity assessment, escalation management, workaround vs fix.
- **xi-techwriter** — API docs, SDK guides, runbooks, ADRs, troubleshooting guides, onboarding content, developer portals.

## Route by intent

| User need | Adopt |
|---|---|
| Diagnose a live issue / customer report, find root cause, decide escalation | `eta-techsupport` |
| Write/update a runbook, troubleshooting guide, API/SDK doc, ADR, onboarding | `xi-techwriter` |
| Both (resolve then capture) | `eta-techsupport` to resolve → hand the root cause + fix to `xi-techwriter` for a durable runbook/KB entry |

**Work tracking** — `${CLAUDE_PLUGIN_ROOT}/library/policies/work-tracking/SKILL.md`: live incident work is exempt from Gates 1 and 1b up front (raise the ticket after, not before), but Gate 3 still applies — short summary comment with the root cause and what to watch for, plus effort logged. A Bug's terminal state is `RESOLVED`, not Done. Customer-facing JSM comments need an explicit yes.

**Docs split** — `${CLAUDE_PLUGIN_ROOT}/library/policies/docs-split/SKILL.md`: `xi-techwriter` owns the **Usage guide** and the **Standard operating procedures** under the product's Confluence folder; Jira stays the engineering view and is not copied into Confluence.

## Escalation & evidence

- Reproduce before escalating; capture exact error strings, versions, and steps.
- Auth symptoms (401/403, `UNAUTHENTICATED`/`UNAUTHORISED`, session expiry loops, missing menu or screen, service credentials rejected): check session validity, the permission the call needs versus the caller's roles, and cookie/CORS scope first; the fix goes through `ship`.
- For code-level root cause, pull in `library/personas/gammas` (or the `review` skill) on the suspect diff; for data issues, `library/personas/dbarch`.
- When the fix needs engineering, hand off to `ship` with the reproduction as acceptance criteria.
- When the follow-up splits into several items (a fix plus a runbook, or fixes across repos), apply `${CLAUDE_PLUGIN_ROOT}/library/policies/task-chips/SKILL.md`: decompose in Jira here, execute at most one Epic in this session, spawn the rest as self-contained task chips that open with a devkit verb.
- Publishing customer-facing content or sending a customer message needs explicit user approval first.
