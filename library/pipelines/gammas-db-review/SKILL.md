---
name: gammas-db-review
description: Comprehensive, persona-orchestrated review of a PR or codebase. Runs Gammas (elite code/security/performance/infra/compliance auditor) over all code changes and Dbarch (data-layer architect — schema, migrations, normalization, indexing) over all database/data-layer changes, then synthesizes a single unified verdict. Use when the user wants a deep review of changes that touch both application code and the data layer.
argument-hint: "[PR number/URL | path | empty for current branch diff] [--comment to post to PR]"
---

# Gamma + Dbarch Review

Run a comprehensive review of the target changes by orchestrating two expert personas:

- **Gammas** — elite code reviewer, QA engineer & security/infra/compliance auditor (`gammas` skill)
- **Dbarch** — expert database architect (`dbarch` skill)

**Target:** $ARGUMENTS

This skill **reviews only** — it does not modify source code. (It may post review comments if `--comment` is passed.) Fixing is a separate step the user runs afterwards.

---

## Phase 0 — Scope the changeset

Determine what to review from `$ARGUMENTS`:

1. **A PR** (a number like `123` or a GitHub URL) → use `gh pr diff <n>` for the diff and `gh pr view <n>` for context (title, description, linked issues). Note the base branch.
2. **A path** (a file or directory that exists) → review the current contents of that path (codebase-review mode), and if it is inside a git repo, prefer the diff of that path vs the default branch when one exists.
3. **Empty** → review the current branch's diff against the default branch (`git merge-base` against `main`/`master`), falling back to uncommitted working-tree changes (`git diff HEAD`) if the branch has no commits ahead.

Gather the diff/changeset, then **partition every changed file** into two buckets (a file may legitimately appear in both):

- **CODE bucket** — application/business logic, API handlers, services, clients, config, infra (Dockerfiles, CI, IaC), tests, frontend.
- **DB / DATA-LAYER bucket** — migration files (`*.sql`, `migrations/**`, Flyway/Liquibase/Alembic/Prisma), ORM models / entity definitions / schema structs, DAOs / repositories / query builders, raw SQL, index definitions, connection/pool config, and any code that constructs queries or transactions.

Output the partition under a heading: `## Review Scope` — list the target, the base, and the two file buckets. If a bucket is empty, say so and skip that persona's phase.

---

## Phase 1 — Gammas reviews the CODE bucket

Adopt the **Gammas** persona fully (elite code reviewer, QA & security/infra/compliance auditor). Review every file in the CODE bucket with full rigour across its mandate:

- Correctness & logic (off-by-one, nil/null, race conditions, error handling, edge cases)
- Security (injection, authn/authz, secrets, SSRF, data exposure, input validation)
- Performance & scalability (N+1, unbounded loops/allocations, missing pagination, hot paths)
- Reliability & observability (timeouts, retries, idempotency, logging/tracing, graceful degradation)
- Infrastructure & supply chain (containers, CI, dependency risk) where touched
- Language-specific anti-patterns and idioms
- Test coverage of new/changed code paths

For findings that touch a **query, transaction, or schema-dependent assumption**, note them and **defer the data-layer judgment to Dbarch** (Phase 2) rather than ruling definitively — flag them as "→ confirm with Dbarch".

Output under: `## Gammas — Code Review`. Classify every finding **CRITICAL / WARNING / INFO**, each with `file:line`, the problem, and a concrete fix. End with a per-persona count.

---

## Phase 2 — Dbarch reviews the DB / DATA-LAYER bucket

Adopt the **Dbarch** persona fully (expert database architect). Review every file in the DB bucket plus any query/transaction code Gammas flagged for confirmation:

- **Schema design** — data types, nullability, normalization vs. justified denormalization, constraints (PK/FK/unique/check), referential integrity, JSONB-vs-column choices, array-vs-junction choices
- **Migrations** — forward/backward safety, reversibility (down migrations), idempotency, lock/downtime risk on large tables, ordering/dependencies, default backfills, contiguous versioning
- **Indexing** — presence and correctness for the query patterns introduced, index type fit (B-tree/GIN/partial/composite), redundant or missing indexes, FK-target validity (e.g. partial-unique indexes cannot be FK targets)
- **Query design** — N+1, full scans, missing `WHERE` on soft-delete columns, transaction scope/atomicity, parameterization (injection), pagination
- **Data integrity & lifecycle** — soft vs hard delete consistency, audit columns, idempotency keys, data durability and mobility, PII tagging/compliance
- **Scale & evolution** — partitioning/sharding implications, backward-compatible schema evolution, multi-region/CDC considerations where relevant

Output under: `## Dbarch — Data-Layer Review`. Classify every finding **CRITICAL / WARNING / INFO** with `file:line`, problem, and recommended change (schema/index/migration). End with a per-persona count.

---

## Phase 3 — Synthesis

Reconcile both reviews into one report under: `## Unified Review`.

1. **Deduplicate / merge** findings the two personas raised on the same line (e.g. Gammas flags a raw-SQL injection risk and Dbarch flags the same query's missing index — present once, attributed to both).
2. **Resolve deferred items** — for each Gammas "→ confirm with Dbarch", state Dbarch's verdict.
3. **Cross-cutting check** — code/schema consistency: do the ORM models match the migration columns? Do queries reference columns that exist? Do soft-delete reads filter the same flag the writes set? Are DTO/API shapes consistent with the stored schema?
4. Present a single findings table sorted by severity: `Severity | Area (Code/DB/Both) | file:line | Finding | Recommended fix`.

End with a verdict block:

```
### Verdict
- Critical: <count>
- Warnings: <count>
- Info: <count>
- Status: PASS | FAIL
```

**Status is PASS** when critical = 0 AND warnings = 0.

---

## Posting comments (optional)

If `--comment` was passed **and** the target is a PR: post each CRITICAL and WARNING as an inline PR review comment at its `file:line` via `gh` (e.g. `gh pr review` / the GitHub review API), prefixed with the persona (`[Gammas]` / `[Dbarch]`) and severity. Post INFO items only if explicitly requested. Without `--comment`, do not post anything — output the report to the conversation only.

---

## Notes

- This is a **review**, not a pipeline — run each persona once. If the user wants the findings fixed, hand off to `/beta` (code) and a follow-up `/dbarch` design pass (schema/migration changes), or re-run `dev-pipeline`.
- If a bucket is empty, skip that persona's phase and say so — do not invent findings.
- Keep findings concrete and actionable: every CRITICAL/WARNING must name the file, the line, the risk, and the fix.
