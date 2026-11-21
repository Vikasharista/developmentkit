# Validation — the definition of done

The renderer produces a strong first draft, not a finished change. A bootstrap run is done only when
every section below holds, with evidence (command output) to show for the gates. Loop until it does,
at most six times, then report what is left.

## 1. Read what was generated

Open every created file and every file a patch touched. Check each against:

- **The spec**: every field, type, nullability, default, constraint, index, enum value, operation and
  permission the developer confirmed is present; nothing they did not ask for slipped in.
- **The host fingerprint** (`conventions.md` § *Detecting the host*): base model and ID type, factory
  type name, repository signatures used by sibling code, mapper style, log message format, error
  helpers, directive wiring, REST layout, file naming. Adapt the generated code to the host; keep the
  deliberate improvements listed in `conventions.md`.
- **The requirement**: business rules the templates cannot know — cross-field checks, state
  transitions, derived values, authorisation beyond the directive, side effects on other modules. Add
  them in the service with tests.
- **Tests that need data the templates cannot invent**: NOT NULL `ref` fields need a parent row seeded in
  the repository integration test and both e2e tests.
- **Auth**: every operation has a permission, or a `# public: <reason>` comment; each new
  permission string is declared in `auth/manifest.yaml`.
- **Leftovers**: `.bootstrap/patches/` applied and deleted; no `panic("not implemented")` from gqlgen; no
  commented-out code gqlgen moved to the end of a resolvers file.

## 2. Gates

Run from the target repo. All must pass.

| Gate | Command | Passes when |
|---|---|---|
| Build | `go build ./...` | no errors |
| Vet / lint | `make lint` (or `go vet ./... && go vet -tags integration ./...`) | clean |
| Unit | `make test` | all pass, no database needed |
| Coverage | `make cover` | 100% of hand-written code; in an existing repo, of the new files (below) |
| End to end | `make e2e` | `TestE2E*` pass: every operation over HTTP against Postgres, error codes, federation `_entities` |
| Migrations | `make migrate-up && make migrate-down && make migrate-up` | each step succeeds on the local database |
| Runtime | `make serve`, then `make smoke` | health, `/healthz` with the database up, the GraphQL probe |
| Supergraph | `make supergraph` (or the host's supergraph test) | composes, when a GraphQL module was added |

In a repo whose older code is below 100%, scope the gate to the files this run added and say so in the
summary:

```bash
COVERAGE_INCLUDE='/internal/app/<module>/.*(<entity>|query_column|copier)' make cover
```

For a repo with no Makefile and none wanted, run the equivalent commands; the targets are listed in
`templates/service/Makefile.tmpl`.

Stop any server or debugger you started (by the PID you started, not by pattern) and `make db-down`
when finished.

## 3. Review

Run the `ship` review pass on the full diff:

- Gamma (`library/personas/gamma-reviewer-tester`): correctness, security, error handling, tests.
- Adversarial reviewer (`library/personas/adversarial-reviewer`): try to break tenancy (cross-namespace
  read/update/delete), injection through filters and sorts, lost updates, missing validation.
- `dbarch` on the migration: types, constraints, index fit for the expected queries, lock impact of the
  DDL on a large existing table.

Fix every CRITICAL and WARNING, then repeat the gates.

## 4. Contract cross-check

Tick each item from the confirmed entity table and conventions answers against the code and tests, and
list in the summary anything you adapted from the templates and why.
