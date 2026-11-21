# Conventions

What every in-house Go backend shares, where backend-a and backend-b differ, and how to read a
host repo before generating into it. Surveyed from two reference backends (backend-a and backend-b),
2026-10-30.

## Shared stack

| Concern | Convention |
|---|---|
| Entry point | `server.go`: core-lib `config.Init` → `logger.Init` → `cmd.Execute{Dsn, ServeFunc: internal.Init}`. core-lib owns the cobra root: `serve`, `migrate up [n]`, `migrate down [n]` |
| Migrations | golang-migrate, `migrations/NNNNNN_snake_desc.{up,down}.sql`, six digits, forward-only, run as `./app migrate up` in a deploy job (`serve` never migrates) |
| Layout | `internal/app/<module>/{model,dto,repository,service,mapper,graphql/{schema,generated,resolver,mapper,loader},rest or controller+router}` |
| DI | `sync.Once` factories per layer (`repository.GetFactory().GetXRepository()`), zero-arg constructors |
| Models | embed a core-lib base model (`VersionedNamespaceModelIDString`, `NamespaceModelIDInt`, …); hooks fill `namespace`, `created_by`, `updated_by`, `version`; `SoftDelete` stamps `deleted_at` + `deleted_by` |
| Tenancy | `namespace varchar(32)`; reads call `utility.AppendNamespaceFilterToQuery` and add `deleted_at IS NULL`; indexes lead with `namespace` and are partial on live rows |
| Audit | `log_audit_trigger AFTER UPDATE OR DELETE … EXECUTE FUNCTION log_audit()` on every table |
| Queries | GraphQL `filters/page/sort` → `coredto.QueryBuilder` → `ToRQP()` → GORM `Where(query.Where(), query.Args()...)`; `{data, totalCount}` pages |
| Errors | core-lib `errors.New*Error`, `RepositoryError` for DB errors; `graphql.PresentGqlError` puts the code in `extensions.code` |
| GraphQL | gqlgen per module (`gqlgen.yml` byte-identical across modules), federation v2, `@key(fields: "id")`, `//go:generate` in `resolver.go`, shared schema files copied per module |
| Logging | `logger.GetLogger()` zap, message `<Module>.<Entity>.<Layer>.<Method>.Error` |
| Tests | stdlib + testify, hand-written stubs (no mockery/gomock), `//go:build integration` + env DSN for DB tests |
| Config | `config/default.yaml` + override file + `APP_*` env vars; `config/local.yaml` gitignored |
| Build | Dockerfile with `GOPRIVATE=bitbucket.org/<your-org>/core-lib` and an SSH key arg; deploy tool (`deploy.yaml`, `.deploy/values.yaml`); Bitbucket pipelines on tags |

## backend-a vs backend-b

| | backend-a | backend-b | Templates default |
|---|---|---|---|
| core-lib | v0.1.7-dev-1 | v0.2.0 | v0.2.0 |
| Auth wiring | `authdirective.NewVerifierDirective(services.gateway.url)`; admin via `VerifyUserPermission(permission.AdminUser())` | `authdirective.New(ConfigFromViper(config.Viper, "auth"))`, provider permsvc; `HasPermission()` | `config` (backend-b); `verifier` for backend-a |
| IDs | `SERIAL` int mostly (`NamespaceModelIDInt`, `VersionedNamespaceModelIDInt`) | `varchar(64)` UUID set in Go (`VersionedNamespaceModelIDString`) | string UUID, versioned |
| Mappers | copier (`copier.Option{DeepCopy, IgnoreEmpty}`) | hand-written | copier, with a swappable `CopyWithOption` for tests |
| Repository `Get` | returns `(rows, total, err)` | separate `Get` and `Count` | `(rows, total, err)` |
| Update | `Save` | `Updates` | `Select("*").Omit(immutable).Updates` + version check |
| Schemas | `public` | `public`, newer contexts own a schema (`quality`) | per spec |
| `log_audit()` | reads `OLD.version`, no namespace column → **every audited table needs `version`** | reads `OLD.version`, writes namespace + schema | reads version through jsonb, so unversioned tables work |
| Supergraph | `apollo-router/supergraph-local.yaml` + GitOps `gitops-non-prod/apollo-router/values.yaml` | `scripts/supergraph/supergraph.yaml` + `supergraph_test.go`; supergraph composed in the Dockerfile | `scripts/supergraph/` |
| Gateway prefix | `withGatewayPrefix` serves every route under `/backend-a` too | webhooks mounted under `""` and `/backend-b` | every route under `""` and `/<service>` |
| DB tests | GORM DryRun SQL assertions; env-gated `DELIVERY_INTEGRATION_DSN` | `QUALITY_TEST_DSN`, migrate once, isolate by namespace | `TEST_DATABASE_DSN` via `internal/testutil/pgtest`, isolate by namespace |
| Makefile | none | none | generated; offered for existing repos |

## Deliberate improvements over both repos

The templates keep the shape above but fix known problems. Keep these when adapting to a host:

- **List allow-lists.** `filters`/`sort` field names are checked against the entity's columns before they
  reach SQL (both repos pass them through to rqp unchecked).
- **Column names.** core-lib v0.2.0 `utility.ToSnakeCase` discards its camel→snake result (`createdBy`
  becomes `createdby`), so `ToRQP` breaks camelCase GraphQL filters. The service normalises names itself
  (`service/query_column.go`).
- **Per-request dataloaders**, never a process-wide singleton (backend-b's `WithLoader` is one).
- **Namespace-guarded delete** and **optimistic locking** on update (the client may send the `version` it read).
- **Required numbers, booleans and times** are pointers on the create request, so REST cannot store a
  missing value as zero; required text cannot be blanked by a patch.
- **REST shares GraphQL auth**: the same directive functions run as gin middleware (`auth.Gin`).
- **No secrets in committed config**; `APP_*` env vars or the mounted override only.
- **CI runs the tests** (both repos' pipelines only build).

## Trust boundary

core-lib's `EnrichContext` copies `X-Namespace-Id` and `X-User-Id` from the request into the context
unchecked. `@authenticate` (permsvc provider) verifies the session or client credentials; it does **not**
check that the caller belongs to the namespace they name. `@hasPermission` sends the namespace to
permission-service and does. backend-a and backend-b live with this behind the Apollo Router and gateway; a
generated service inherits it. So:

- Recommend a permission for every write, and for reads of anything tenant-sensitive.
- Permission checks always run after `@authenticate`, so `created_by` / `updated_by` / `deleted_by`
  come from the verified session, never from the `X-User-Id` header.
- Federation `_entities` runs `Authenticate` too (a hook in `graphql/init.go`); without it any caller who
  reaches the subgraph could read rows by ID.
- Never expose a subgraph route publicly without the router in front. Permission naming and seeding follow
  your organisation's auth platform.

## Known limits of the generated code

- A patch cannot set a nullable column back to NULL (nil means "keep"). Add a dedicated mutation when a
  field must be clearable.
- Unversioned entities overwrite every column on update, so two concurrent patches to different fields can
  lose one. Versioned entities (the default) refuse a stale write when the caller sends the `version` it read.
- `ref` columns use a plain FK to `(id)`: the database does not stop a reference to a parent in another
  namespace. Add a parent lookup in the service (same namespace) during validation.

## Detecting the host

Before generating into an existing repo, read and note:

1. `go.mod`: module path, Go version, core-lib version (decides `auth: config` vs `verifier` and whether
   `HasPermission` exists), gqlgen version.
2. `AGENTS.md`, `CLAUDE.md`, `README.md`, and the target module's README: explicit rules win over templates.
3. The target module, or the closest sibling, end to end: base model and ID type, mapper style, repository
   signatures, factory type name (`factory` or exported `Factory`), log message style, how `graphql/init.go`
   wires directives and loaders, REST layout (`controller/` + `router/` in backend-a, `webhook/` or
   `rest/` elsewhere).
4. `ls migrations | tail`: next number, naming, whether new contexts use their own schema; the newest
   `log_audit` definition (does it need `version`? a namespace column?).
5. `internal/server.go`: how modules are mounted, gateway prefixes, middleware order.
6. Supergraph listing file and its test, if any.
7. Existing test setup: integration DSN env var, `TestMain`s (the templates add one per package through
   `main_integration_test.go`; merge into an existing one instead of adding a second).

Record the fingerprint in your working notes and adapt the rendered files to it in the validation step.
Where the host and the templates disagree, the host wins, except for the improvements listed above,
which you keep and point out to the developer.
