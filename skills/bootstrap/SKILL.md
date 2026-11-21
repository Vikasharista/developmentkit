---
name: bootstrap
description: Scaffold an in-house Go backend the way backend-a and backend-b are built, either a whole new service or a new entity (or module) in an existing one. Walks the developer through the entity design, double-confirms the DB schema, settles conventions (GraphQL subgraph and/or gin REST, auth, IDs, versioning), then generates migration, model, repository, service, copier mappers, resolvers/controllers, factories, docs and a Makefile, and drives it until build, 100% coverage and end-to-end tests pass. Use for "new backend service", "new microservice", "add an entity / table / CRUD", "new module", "scaffold". Frontend scaffolding is not available yet.
---

# bootstrap — requirement in, running service or entity out

Builds Go backends on the in-house stack (core-lib, gin, gqlgen federation v2, GORM, golang-migrate)
from a requirement. A deterministic renderer writes the first draft from templates; **you** then
validate every file against the target repo and the agreed spec, fix what does not fit, and prove it
runs. The templates are a base, never the verdict.

Library: `${CLAUDE_PLUGIN_ROOT}/library/bootstrap/` — read `README.md` there first. It indexes
`conventions.md` (house rules, backend-a vs backend-b deltas, host detection), `schema-design.md`
(entity dialogue, type map, DDL rules), `validation.md` (the definition of done), the renderer
(`render/`), the templates and a worked example spec (`examples/widget-service.json`).

Stop points are marked ⏸: do not continue past one without the developer's explicit answer.

## 0 — Route and ground

- **Frontend** request (UI, web app, mobile screen, "FE for this"): reply that frontend bootstrap is
  coming soon, offer to scaffold the backend it needs, and stop there for the frontend part.
- **Backend**: apply `${CLAUDE_PLUGIN_ROOT}/library/policies/work-tracking/SKILL.md` Gate 1 (ticket +
  spec) and Gate 1b once execution starts, exactly as `ship` does.
- Decide the **mode**:
  - `service` — a new repo (service + first module + first entity).
  - `module` — a new module with its first entity, in an existing repo.
  - `entity` — a new entity in an existing module.
  - Ask if it is not obvious from the request and the working directory.

## 1 — Read the host (module and entity modes)

Follow `conventions.md` § *Detecting the host*: read `go.mod`, `AGENTS.md` / `CLAUDE.md`, the target module
(or the closest sibling) end to end, the newest migrations and the `log_audit` function. Write down the
fingerprint: core-lib version and auth wiring, ID type, base model, mapper style, transports, migration
number, supergraph file, test setup. **The host wins** wherever it differs from the templates; you will
adapt the rendered files to it in step 6.

## 2 — ⏸ Entity design

Turn the requirement into a field table with `schema-design.md` § *Entity dialogue*. Propose; do not
interrogate. Present it as one table (field, type, nullable, default, unique/index, enum values,
reference, comment) plus your assumptions, and ask the developer to correct or confirm. Settle here too
everything else that shapes the table: entity and table names, ID type, versioning (default yes; required
where the host's `log_audit` reads `OLD.version`) and, for a new module, the Postgres schema. Iterate
until they confirm.

## 3 — ⏸⏸ DB schema, confirmed twice

Write the draft spec to `<scratch>/spec.json` (schema: `render/spec.go`, example:
`examples/widget-service.json`; operations and transports can stay at their defaults for now, they do
not change the DDL). The renderer validates names before anything is shown. Render only the migration:

```bash
cd ${CLAUDE_PLUGIN_ROOT}/library/bootstrap/render   # the renderer is its own Go module
go run . -spec <scratch>/spec.json -out <scratch>/ddl -only 'migrations/*'
```

Use absolute paths for `-spec` and `-out`.

Show the exact up and down SQL. Walk through table name and schema, ID type, every column, constraints,
indexes, the audit trigger and the down migration. Ask for confirmation, then ask once more: "Confirm
this schema is final? Changing it after the migration runs anywhere needs a new migration." Only a
second yes moves on. Any change sends you back to step 2.

## 4 — ⏸ Conventions Q&A

One `AskUserQuestion` round (two at most), recommendations first. Only ask what the host or the request
does not already answer, and nothing that changes the confirmed DDL:

- Transport: GraphQL subgraph, REST, or both (new service / new module only).
- Operations: create, update, delete, get, list, and the permission each one checks. Every operation carries
  `@authenticate @hasPermission(...)` (one `permissions` entry per operation in the spec), permission strings
  are `org::app::client::entity::action`, and each new one is declared in the app's `auth/manifest.yaml`
  in the same change. An operation left without a permission is public by decision only: record the reason
  as a `# public: <reason>` comment above it in the schema. `@authenticate` alone does not check the caller
  belongs to the `X-Namespace-Id` they send (`conventions.md` § *Trust boundary*).
- Federation entity + dataloader (default yes when GraphQL and get are on).
- New service: name, Go module path, port, route prefix. Existing repo without a Makefile: add one?

## 5 — Write the spec and render

Complete the spec with the step 4 answers and save it as `.bootstrap/<module>.<entity>.json` in the
target repo. Then:

```bash
cd ${CLAUDE_PLUGIN_ROOT}/library/bootstrap/render
go run . -spec <repo>/.bootstrap/<module>.<entity>.json -out <repo> -dry-run   # review the file list
go run . -spec <repo>/.bootstrap/<module>.<entity>.json -out <repo>
```

The renderer only creates files and refuses to overwrite. Edits to existing files land in
`.bootstrap/patches/<module>.<entity>.md`: apply each one with Edit, then delete the patch file.
For a Makefile in an existing repo, run once more with
`-only 'Makefile,docker-compose.yml,scripts/coverage.sh,scripts/smoke.sh'` (any mode; rename anything that
collides, and scope the coverage gate with `COVERAGE_INCLUDE` as `validation.md` says).

Then, in the target repo:

```bash
go mod tidy -e                     # tolerates the not-yet-generated graphql package
make gql MODULE=<module>           # or: cd internal/app/<module> && go run github.com/99designs/gqlgen generate
go mod tidy && go build ./... && go vet ./... && go vet -tags integration ./...
```

## 6 — Validate before calling anything done

Mandatory. Work through `validation.md` in order, which comes down to:

1. **Read every generated and patched file** against the host fingerprint and the spec. Fix drift from
   host conventions (naming, error and log style, factory shape, auth wiring, test style). Fix anything the
   templates got wrong for this entity (business rules the requirement implies, cross-field validation,
   references needing a parent row in tests).
2. **Run the gates**: `make lint`, `make cover` (100% of hand-written code; in existing repos, of the new
   files), `make e2e`, migrate up → down 1 → up against local Postgres, `make serve` + `make smoke`.
3. **Review**: run the `ship` review pass on the diff — Gamma (`library/personas/gamma-reviewer-tester`) and
   the adversarial reviewer — plus `dbarch` on the migration. Fix every CRITICAL and WARNING and re-run the gates.
4. **Cross-check the contract**: every field, constraint, operation and permission from steps 2–4 is
   present, and nothing extra slipped in.

Any failure loops back into this step. Maximum six loops, as in `ship`; then stop and report the punch list.

## 7 — Docs, hand-off, PR

- Module README, root README module map and `AGENTS.md` module index updated (the templates do this for new
  services; patches tell you where for existing repos). Keep `.bootstrap/*.json`.
- GitOps / Apollo Router / deploy tool changes outside this repo: print the exact snippet for the developer.
  Never edit those repos from here.
- Git, PR and Jira state follow `ship` § 3 (in-review when the PR is raised).

## Summary to the developer

```
## Bootstrap summary
- Mode / target: <service|module|entity> in <repo>, module <m>, entity <E>
- Schema: <schema.table>, migration <NNNNNN_desc> (confirmed twice)
- Surface: <GraphQL ops> / <REST routes>, auth <directives>
- Gates: lint ✓ · coverage <n>% ✓ · e2e ✓ · migrate up/down/up ✓ · serve + smoke ✓ · review PASS
- Adapted from templates: <what you changed to fit the host or the requirement>
- Outside this repo: <snippets handed over, or none>
```
