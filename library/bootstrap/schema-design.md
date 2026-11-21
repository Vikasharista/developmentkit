# Schema design

## Entity dialogue

Derive a first draft from the requirement, then let the developer correct it. For each candidate field
settle:

| Question | Why it matters |
|---|---|
| What is it called, in snake_case? Is it the business name people use? | Column, GraphQL field and JSON key all derive from it |
| Type (table below)? For strings, the real maximum length | `varchar(n)` + validation; samples in tests fit the width |
| Can it be absent? (nullable) | Nullable → pointer in Go, optional in GraphQL |
| Does the database supply a value when the caller does not? (default) | Default → optional on create, pointer + GORM `default:` tag |
| Unique within a namespace? Searched or sorted often? | Partial unique / plain index leading with `namespace` |
| A fixed set of values? | `enum`: varchar + CHECK + GraphQL enum + service check; values UPPER_SNAKE |
| Points at another table? | `ref`: FK + index, `ref_id_type` `string` (varchar(64)) or `int` (integer) to match the parent; tests need a parent row when NOT NULL |
| Business-generated code (`VND-0001`)? | Not templated: add the backend-b `business_code_configs` trigger by hand |

Then settle the entity: singular/plural names, table name, ID type (UUID string or int), versioned or
not, which operations, which need a permission, whether it is a federation entity, and a one-line
description (becomes the table comment and the GraphQL type doc).

Never add columns the base model owns: `id`, `namespace`, `version`, `is_active`, `created_*`,
`updated_*`, `deleted_*`. The renderer also rejects SQL keywords (`order`, `user`, `from`, …) and `key`
as field names, entity names with consecutive capitals (`QrCode`, not `QRCode`), entity or enum type
names that clash with the module schema or gqlgen (`Filter`, `Sort`, `Config`, `Query`, …), Go keywords,
duplicate enum values, quotes in defaults and identifiers longer than Postgres's 63 bytes.

Spot and raise: money without a currency, quantities without a unit, free text that should be an enum,
a status without its transitions, a reference to an entity in another service (store the ID as `string`,
not `ref`; there is no cross-service FK), personal data that needs masking or retention rules.

## Field types

| Spec type | Go (model / DTO) | GraphQL | Postgres | Notes |
|---|---|---|---|---|
| `string` | `string` | `String` | `varchar(max_len)` (default 256) | length checked on create and update |
| `text` | `string` | `String` | `text` | |
| `int` | `int` | `Int` | `integer` | |
| `decimal` | `float64` | `Float` | `numeric(precision,scale)` (default 18,3) | money: say so, consider minor units |
| `bool` | `bool` | `Boolean` | `boolean` | give it a default |
| `timestamp` | `time.Time` | `Time` | `timestamptz` | |
| `enum` | `string` | `<Entity><Field>` enum | `varchar(32)` (wider if a value is longer) + CHECK | |
| `ref` | `string` / `int` | `String` / `Int` | `varchar(64)` / `integer` + FK `REFERENCES schema.table(id)` + index | same-database parents only |
| `jsonb` | `model.Map` (DTO `map[string]interface{}`) | `Map` | `jsonb` | no unique/index here; add a GIN index by hand if queried |

Nullable or defaulted fields become pointers. Required numbers, booleans and times are pointers on the
create request only, so a missing value is rejected. Go field names follow gqlgen v0.17.86's initialisms
(`customer_id` → `CustomerID`, `image_url` → `ImageURL`; `sku` stays `Sku`), so copier matches model, DTO
and gqlgen types by name.

## DDL rules (what the migration template emits)

- `CREATE SCHEMA IF NOT EXISTS` when the module owns a schema.
- `CREATE TABLE` without `IF NOT EXISTS`: an existing table of that name makes the migration fail instead
  of being silently adopted (its down migration would drop it).
- `id` `varchar(64)` (UUID set by core-lib) or `serial`; `namespace varchar(32) NOT NULL`.
- Entity columns, then `is_active boolean NOT NULL DEFAULT true`, `version int NOT NULL DEFAULT 1`
  (versioned), `created_by NOT NULL`, `updated_by`, `deleted_by`, `created_at timestamptz NOT NULL DEFAULT now()`,
  `updated_at`, `deleted_at`, `PRIMARY KEY (id)`.
- Named constraints: `chk_<table>_<column>`, `fk_<table>_<column>`.
- `COMMENT ON TABLE/COLUMN` from the descriptions.
- Indexes: `<table>_namespace_created_at_idx` (default list order), `<table>_namespace_<col>_uidx` for
  unique, `<table>_namespace_<col>_idx` for indexed and reference columns, all `WHERE deleted_at IS NULL`.
- `DROP TRIGGER IF EXISTS` + `CREATE TRIGGER log_audit_trigger AFTER UPDATE OR DELETE … log_audit()`.
- No `BEGIN`/`COMMIT`: golang-migrate wraps the file in one transaction.
- Down: drop trigger, drop table. A schema is left in place (other tables may live in it).

## Double confirmation

Show the rendered up and down migration verbatim. Point out anything the developer may not have
noticed: defaults, partial unique semantics (a soft-deleted row frees its code), the audit trigger's
`version` dependency on the host, index choices, and the fact that changing the schema after the
migration runs anywhere means a new migration. Ask to confirm; then ask once more whether it is final.
Do not render anything else before the second yes.
