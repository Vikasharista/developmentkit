---
name: understand
description: Understand any codebase — architecture, file relationships, business domain, a specific file/function, PR/diff risk, or new-hire onboarding — via a persistent knowledge graph. Use for "explain this codebase", "how does X work", "what changed in this PR and what's the risk", "onboard me to this repo", or any question about project structure and content.
---

# understand — codebase intelligence

Router over the knowledge-graph suite in `${CLAUDE_PLUGIN_ROOT}/library/understand/`. Pick the sub-skill by question type; read and follow that SKILL.md (each has its own scripts/tools).

| User wants | Sub-skill |
|---|---|
| Build/refresh the knowledge graph for a repo | `understand` |
| Ask questions about the codebase | `understand-chat` (or query an existing `graphify-out/`) |
| Deep-dive a specific file/function/module | `understand-explain` |
| Analyze a git diff / PR — what changed, impact, risk | `understand-diff` |
| Business-domain map / domain flow graph | `understand-domain` |
| Onboarding guide for new team members | `understand-onboard` |
| Interactive web dashboard of the graph | `understand-dashboard` |
| Analyze a Karpathy-pattern LLM wiki knowledge base | `understand-knowledge` |
| Turn arbitrary input (code/docs/papers/media) into a graph | `graphify` |

Default: if `graphify-out/` exists, treat the question as a graph query first (`understand-chat`). Otherwise build the graph (`understand`) then answer.

Do not surface these as separate choices — infer the right one and go.

**Auth questions** ("how does login / permission checking work here", "what can role X do", "why is this screen hidden") — explain against the repo's own SDL guards, route middleware and permission manifest; flag any check that fails open.
