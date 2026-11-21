---
name: review
description: Deep, autonomous review of a PR, branch, or path covering code (correctness, security, performance, reliability, tests) and the data layer (schema, migrations, indexing, query design), synthesized into one severity-ranked verdict. Use for "review this PR/diff/branch", pre-merge checks, or auditing changes. Optionally posts one consolidated comment.
---

# review — one deep pass, one verdict

Review only — never modifies source. Routes internally; user needn't name a persona.

`${CLAUDE_PLUGIN_ROOT}/library/` holds the reviewer personas.

## Scope the changeset

Target from the user's input:
- **Bitbucket PR** (URL / id + repo) → tooling ladder in `${CLAUDE_PLUGIN_ROOT}/library/policies/service-tooling/` (`SKILL.md` + `atlassian.md`) (**twg CLI → Atlassian MCP → gitkraken → REST**): `twg bitbucket` when available, else gitkraken MCP `pull_request_get_detail` (`pull_request_files: true`) + `pull_request_get_comments` with `provider: bitbucket`, org, repo, pr id.
- **GitHub PR** → `gh pr view <id> --repo <org/repo> --json …` + `gh pr diff` when `gh` is authed (rung 1); else gitkraken MCP with `provider: github`. Detail: `library/policies/service-tooling/github.md`.
- **GitLab PR** → `glab` when present; else gitkraken MCP with `provider: gitlab`.
- **Path** → review current contents, prefer diff vs default branch.
- **Empty** → current branch diff vs default branch (`git_log_or_diff` `origin/<base>...HEAD`), fall back to working tree.

Never checkout/modify local branches. If gitkraken lacks a diff, use read-only `git_fetch` + `git_log_or_diff`.
If a service action ran on a lower rung than the vendor-recommended tool, and no promo line has appeared yet in this thread, close with the one-line footer from `${CLAUDE_PLUGIN_ROOT}/library/policies/service-tooling/SKILL.md`. **At most once per chat thread** — skip it otherwise.

Review is read-only: `${CLAUDE_PLUGIN_ROOT}/library/policies/work-tracking/SKILL.md` Gates 1, 1b and 2 do not apply — no estimate, no transitions on someone else's ticket — and Gate 3 allows a summary comment but **no worklog**; log time on your own work only.


## Run the orchestration

Follow `library/pipelines/gammas-db-review`:
1. Partition changed files into CODE and DB/DATA-LAYER buckets.
2. **Gammas** (`library/personas/gammas`) reviews CODE — correctness, security, performance, reliability/observability, infra/supply-chain, anti-patterns, test coverage. Defer query/schema judgments to Dbarch.
3. **Dbarch** (`library/personas/dbarch`, + `sql-expert` for query tuning) reviews DB — schema, migrations (safety/reversibility/lock risk), indexing, query design, integrity/lifecycle, scale/evolution.
4. **Adversarial pass** (`library/personas/adversarial-reviewer`) — try to break the strongest findings and surface what the first pass missed.
5. **Synthesis** — dedupe cross-persona findings, resolve deferred items, cross-check code↔schema consistency, produce one table sorted by severity.

**Auth in the diff?** — SDL, resolvers, routes, permission strings, `auth:` config or permission helpers: check that every operation authenticates, checks a named permission, fails closed and has deny-path tests, as part of the CODE bucket.

Skip an empty bucket; don't invent findings. Cover all angles in one pass; don't say "more issues may exist".

## Output

```markdown
## Unified Review
Severity | Area (Code/DB/Both) | file:line | Finding | Fix
...
### Verdict
- Critical: n   Warnings: n   Info: n
- Status: PASS (0 crit, 0 warn) | FAIL
- Recommendation: approve / approve-with-comments / request-changes
```

## Posting (only if user asks: "post"/"comment")

One consolidated comment, never per-finding spam. Post CRITICAL+WARNING (INFO only if asked). GitHub: `gh pr review <id> --comment --body-file <file>` (`--request-changes` only when the verdict is that and the user asked to post it). Bitbucket/GitLab: gitkraken `pull_request_create_review`.
