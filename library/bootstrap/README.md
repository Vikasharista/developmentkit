# bootstrap library

Building blocks for the `bootstrap` verb (`skills/bootstrap/SKILL.md`).

**Before you generate anything:** the templates import a shared Go core library (`core-lib`) for config,
logging, Postgres, transactions, DTOs, errors and the GraphQL auth directive (`authdirective`). Set
`service.core_module` in the spec to the module path of a library that provides those packages and that
API. Without one, the renderer still runs but the generated service will not build. `org_slug` (default
`your-org`) and `org_domain` (default `example.com`) fill the Bitbucket workspace and the auth/gateway
hosts in the generated config; set them per project.

| Path | What it is |
|---|---|
| [conventions.md](conventions.md) | House rules for Go backends, backend-a vs backend-b deltas, how to fingerprint a host repo |
| [schema-design.md](schema-design.md) | Entity dialogue, field types and their Go / GraphQL / SQL mapping, DDL rules, the double confirmation |
| [validation.md](validation.md) | Definition of done: the review and gates a run must pass before it is called finished |
| [render/](render/) | Stdlib-only Go renderer: spec JSON + templates → files, plus patch notes for existing files |
| [templates/](templates/) | `service/` (repo skeleton), `module/` (module shell), `entity/` (one entity end to end); `_patches/` hold edits to existing files |
| [examples/](examples/) | `widget-service.json` (every field type, GraphQL + REST, string IDs), `inventory-service.json` (REST only, int IDs, verifier auth), `logistics-service.json` (GraphQL only, int IDs) |
| [scripts/golden.sh](scripts/golden.sh) | Renders an example and runs build, coverage gate and e2e against docker Postgres |

## Renderer

```bash
cd library/bootstrap/render
go run . -spec <spec.json> -out <repo root> [-dry-run] [-only 'glob,glob']
go test ./...
```

- `mode` in the spec picks the layers: `service` renders service + module + entity, `module` renders
  module + entity, `entity` renders the entity only.
- Templates are Go `text/template` files ending in `.tmpl`; `__module__`, `__entity__`, `__migration__`
  and `__service__` in paths are substituted. A template that renders to nothing is skipped (that is how
  GraphQL-only and REST-only files switch off). `.once.tmpl` files are shared helpers: created when
  missing, kept when present.
- Existing files are never overwritten; the run aborts listing them. In `module` and `entity` modes the
  edits existing files need are written to `.bootstrap/patches/<module>.<entity>.md`.
- `.go` output is gofmt'ed; a template that produces invalid Go fails the run with numbered source.

Writing templates: never put `{{` inside Go code (composite literals such as `[]T{{...}}`); break the
literal over lines instead. Keep generated code free of branches tests cannot reach: the coverage gate is
100%.

## Golden run (how the templates are verified)

Each template change is checked with `scripts/golden.sh` on the three examples (string IDs + both transports; integer IDs + REST only
+ `verifier` auth; integer IDs + GraphQL only), which runs:

```bash
go mod tidy -e && (cd internal/app/<module> && go run github.com/99designs/gqlgen generate) && go mod tidy
go build ./... && go vet ./... && go vet -tags integration ./...
make db-up db-create-test && make cover && make e2e
```

All three reach 100% coverage with their end-to-end suites passing against Postgres 16. The widget run
also exercises `make migrate-up`, `migrate-down`, `migrate-new`, `serve`, `smoke` and `debug`.
