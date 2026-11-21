package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Spec is the agreed design for one bootstrap run. The skill writes it to
// .bootstrap/<module>.<entity>.json after the user confirms the entity, the DB
// schema and the conventions, so every run can be reproduced and reviewed.
type Spec struct {
	// Mode is service (new repo: service + module + entity), module (new module
	// + entity in an existing repo) or entity (entity in an existing module).
	Mode      string    `json:"mode"`
	Service   Service   `json:"service"`
	Module    Module    `json:"module"`
	Entity    Entity    `json:"entity"`
	Migration Migration `json:"migration"`
}

type Service struct {
	Name        string `json:"name"`         // kebab-case, e.g. widget-service
	ModulePath  string `json:"module_path"`  // go module path
	Description string `json:"description"`  // one line, used in README
	GoVersion   string `json:"go_version"`   // e.g. 1.25.5
	CoreVersion string `json:"core_version"` // core library version, e.g. v0.2.0
	CoreModule  string `json:"core_module"`  // Go module path of your core library, e.g. bitbucket.org/your-org/core-lib
	OrgSlug     string `json:"org_slug"`     // Bitbucket workspace / org slug, default your-org
	OrgDomain   string `json:"org_domain"`   // base domain for auth and gateway hosts, default example.com
	Port        int    `json:"port"`         // HTTP port, default 8080
	DBName      string `json:"db_name"`      // local Postgres database name
	DBPort      int    `json:"db_port"`      // local Postgres host port, default 5432
	RoutePrefix string `json:"route_prefix"` // gateway prefix, e.g. /widget-service
	Namespace   string `json:"namespace"`    // namespace used by local tests and smoke runs
}

type Module struct {
	Name        string `json:"name"`        // snake_case package dir under internal/app
	Route       string `json:"route"`       // URL segment, kebab-case
	PgSchema    string `json:"pg_schema"`   // postgres schema, "public" or a bounded-context schema
	GraphQL     bool   `json:"graphql"`     // expose a federation v2 subgraph
	REST        bool   `json:"rest"`        // expose gin REST routes
	Auth        string `json:"auth"`        // config (authdirective.New) or verifier (NewVerifierDirective)
	Description string `json:"description"` // one line, used in module README
}

type Entity struct {
	Name       string   `json:"name"`       // PascalCase singular, e.g. Widget
	Plural     string   `json:"plural"`     // PascalCase plural, e.g. Widgets
	Table      string   `json:"table"`      // snake_case plural table name
	IDType     string   `json:"id_type"`    // string (varchar(64) uuid) or int (serial)
	Versioned  bool     `json:"versioned"`  // version column + optimistic locking
	Loader     bool     `json:"loader"`     // dataloader + federation entity resolver
	Operations []string `json:"operations"` // subset of create, update, delete, get, list
	// Permissions maps an operation to a permission string. Mapped operations use
	// @hasPermission(permission: ...); the rest use @authenticate.
	Permissions map[string]string `json:"permissions"`
	Description string            `json:"description"`
	Fields      []Field           `json:"fields"`
}

type Field struct {
	Name      string   `json:"name"`        // snake_case column name
	Type      string   `json:"type"`        // see fieldTypes
	MaxLen    int      `json:"max_len"`     // string only, default 256
	Precision int      `json:"precision"`   // decimal only, default 18
	Scale     int      `json:"scale"`       // decimal only, default 3
	Nullable  bool     `json:"nullable"`    // column allows NULL
	Unique    bool     `json:"unique"`      // unique per namespace among live rows
	Index     bool     `json:"index"`       // plain index per namespace among live rows
	Enum      []string `json:"enum"`        // enum only, UPPER_SNAKE values
	Default   string   `json:"default"`     // raw SQL default, optional
	Ref       string   `json:"ref"`         // ref only, schema-qualified table it points at
	RefIDType string   `json:"ref_id_type"` // ref only, the parent's id type: string (default) or int
	Comment   string   `json:"comment"`     // becomes COMMENT ON COLUMN and doc text
}

type Migration struct {
	Number      int    `json:"number"`      // next migration number in the target repo
	Description string `json:"description"` // snake_case, e.g. catalog_widgets
}

var fieldTypes = map[string]bool{
	"string": true, "text": true, "int": true, "decimal": true, "bool": true,
	"timestamp": true, "jsonb": true, "enum": true, "ref": true,
}

var operations = map[string]bool{"create": true, "update": true, "delete": true, "get": true, "list": true}

var (
	snakeRe = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	// One capital per word (QrCode, not QRCode): names derive routes, columns and GraphQL fields.
	pascalRe = regexp.MustCompile(`^[A-Z][a-z0-9]+([A-Z][a-z0-9]+)*$`)
	kebabRe  = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	enumRe   = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
	tableRe  = regexp.MustCompile(`^[a-z][a-z0-9_]*\.[a-z][a-z0-9_]*$`)
)

// reservedTypeNames are GraphQL types the module schema or gqlgen already define.
var reservedTypeNames = map[string]bool{
	"Query": true, "Mutation": true, "Subscription": true, "Entity": true, "Service": true,
	"Filter": true, "FilterOperator": true, "Sort": true, "SortOrder": true, "PageInput": true,
	"Any": true, "Map": true, "Time": true, "Int": true, "Float": true, "String": true, "Boolean": true,
	"ID": true, "Config": true, "ResolverRoot": true, "DirectiveRoot": true, "ComplexityRoot": true,
	"User": true,
}

// reservedFieldNames break generated code: SQL keywords that rqp and the DDL leave
// unquoted, and key, which clashes with the DTO's Key() method.
var reservedFieldNames = map[string]bool{
	"key": true, "all": true, "and": true, "any": true, "as": true, "asc": true, "between": true,
	"both": true, "case": true, "cast": true, "check": true, "column": true, "constraint": true,
	"create": true, "default": true, "desc": true, "distinct": true, "do": true, "else": true,
	"end": true, "except": true, "false": true, "for": true, "foreign": true, "from": true,
	"grant": true, "group": true, "having": true, "in": true, "into": true, "is": true, "join": true,
	"like": true, "limit": true, "not": true, "null": true, "offset": true, "on": true, "only": true,
	"or": true, "order": true, "primary": true, "references": true, "select": true, "table": true,
	"then": true, "to": true, "true": true, "union": true, "unique": true, "user": true,
	"using": true, "when": true, "where": true, "with": true,
}

// goKeywords cannot be the lower-cased entity name, which generated code uses as a variable.
var goKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true, "continue": true, "default": true,
	"defer": true, "else": true, "fallthrough": true, "for": true, "func": true, "go": true,
	"goto": true, "if": true, "import": true, "interface": true, "map": true, "package": true,
	"range": true, "return": true, "select": true, "struct": true, "switch": true, "type": true,
	"var": true,
}

// postgresIdentifierMax is the length past which Postgres silently truncates names.
const postgresIdentifierMax = 63

// reservedColumns are owned by the core-lib base models and the audit trigger.
var reservedColumns = map[string]bool{
	"id": true, "namespace": true, "version": true, "is_active": true,
	"created_at": true, "updated_at": true, "deleted_at": true,
	"created_by": true, "updated_by": true, "deleted_by": true,
}

func loadSpec(path string) (*Spec, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	spec := &Spec{}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(spec); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	spec.applyDefaults()
	if err := spec.validate(); err != nil {
		return nil, err
	}
	return spec, nil
}

func (s *Spec) applyDefaults() {
	if s.Service.Port == 0 {
		s.Service.Port = 8080
	}
	if s.Service.DBPort == 0 {
		s.Service.DBPort = 5432
	}
	if s.Service.GoVersion == "" {
		s.Service.GoVersion = "1.25.5"
	}
	if s.Service.CoreVersion == "" {
		s.Service.CoreVersion = "v0.2.0"
	}
	if s.Service.OrgSlug == "" {
		s.Service.OrgSlug = "your-org"
	}
	if s.Service.OrgDomain == "" {
		s.Service.OrgDomain = "example.com"
	}
	if s.Service.Namespace == "" {
		s.Service.Namespace = "local"
	}
	if s.Service.DBName == "" {
		s.Service.DBName = strings.ReplaceAll(s.Service.Name, "-", "_")
	}
	if s.Service.RoutePrefix == "" && s.Service.Name != "" {
		s.Service.RoutePrefix = "/" + s.Service.Name
	}
	if s.Module.Route == "" {
		s.Module.Route = strings.ReplaceAll(s.Module.Name, "_", "-")
	}
	if s.Module.PgSchema == "" {
		s.Module.PgSchema = "public"
	}
	if s.Module.Auth == "" {
		s.Module.Auth = "config"
	}
	if s.Entity.IDType == "" {
		s.Entity.IDType = "string"
	}
	if len(s.Entity.Operations) == 0 {
		s.Entity.Operations = []string{"create", "update", "delete", "get", "list"}
	}
	for i := range s.Entity.Fields {
		f := &s.Entity.Fields[i]
		if f.Type == "string" && f.MaxLen == 0 {
			f.MaxLen = 256
		}
		if f.Type == "decimal" && f.Precision == 0 {
			f.Precision, f.Scale = 18, 3
		}
	}
}

func (s *Spec) validate() error {
	var problems []string
	add := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }

	switch s.Mode {
	case "service", "module", "entity":
	default:
		add("mode must be service, module or entity, got %q", s.Mode)
	}
	if !kebabRe.MatchString(s.Service.Name) {
		add("service.name must be kebab-case, got %q", s.Service.Name)
	}
	if s.Service.ModulePath == "" {
		add("service.module_path is required")
	}
	if s.Service.CoreModule == "" {
		add("service.core_module is required: the Go module path of your core library (see library/bootstrap/README.md)")
	}
	if !snakeRe.MatchString(s.Module.Name) {
		add("module.name must be snake_case, got %q", s.Module.Name)
	}
	if !kebabRe.MatchString(s.Module.Route) {
		add("module.route must be kebab-case, got %q", s.Module.Route)
	}
	if !snakeRe.MatchString(s.Module.PgSchema) {
		add("module.pg_schema must be snake_case, got %q", s.Module.PgSchema)
	}
	if !s.Module.GraphQL && !s.Module.REST {
		add("module needs graphql, rest or both")
	}
	if s.Module.Auth != "config" && s.Module.Auth != "verifier" {
		add("module.auth must be config or verifier, got %q", s.Module.Auth)
	}
	if !pascalRe.MatchString(s.Entity.Name) || !pascalRe.MatchString(s.Entity.Plural) {
		add("entity.name and entity.plural must be PascalCase with one capital per word (QrCode, not QRCode)")
	}
	for _, name := range []string{s.Entity.Name, s.Entity.Plural, "Create" + s.Entity.Name + "Input", "Update" + s.Entity.Name + "Input"} {
		if reservedTypeNames[name] {
			add("GraphQL type %q clashes with a type the module or gqlgen defines; rename the entity", name)
		}
	}
	if goKeywords[lowerFirst(s.Entity.Name)] || goKeywords[lowerFirst(s.Entity.Plural)] {
		add("entity %q lower-cases to a Go keyword; rename it", s.Entity.Name)
	}
	if s.Entity.Name == s.Entity.Plural {
		add("entity.plural must differ from entity.name")
	}
	if !snakeRe.MatchString(s.Entity.Table) {
		add("entity.table must be snake_case, got %q", s.Entity.Table)
	}
	if s.Entity.IDType != "string" && s.Entity.IDType != "int" {
		add("entity.id_type must be string or int, got %q", s.Entity.IDType)
	}
	for _, op := range s.Entity.Operations {
		if !operations[op] {
			add("unknown operation %q", op)
		}
	}
	for op := range s.Entity.Permissions {
		if !s.hasOp(op) {
			add("permission set for %q, which is not an enabled operation", op)
		}
	}
	if s.Entity.Loader && !s.hasOp("get") {
		add("entity.loader needs the get operation")
	}
	if s.Migration.Number < 1 {
		add("migration.number must be >= 1")
	}
	if s.Mode == "service" && s.Migration.Number < 2 {
		add("migration.number must be >= 2 in service mode; 000001 is the audit schema")
	}
	if !snakeRe.MatchString(s.Migration.Description) {
		add("migration.description must be snake_case, got %q", s.Migration.Description)
	}
	if len(s.Entity.Fields) == 0 {
		add("entity needs at least one field")
	}
	seen := map[string]bool{}
	for _, f := range s.Entity.Fields {
		if !snakeRe.MatchString(f.Name) {
			add("field %q must be snake_case", f.Name)
		}
		if reservedColumns[f.Name] {
			add("field %q is owned by the core-lib base model", f.Name)
		}
		if reservedFieldNames[f.Name] {
			add("field %q is a SQL keyword or clashes with generated code; rename it (e.g. %s_value)", f.Name, f.Name)
		}
		if strings.ContainsAny(f.Default, "\"`") {
			add("field %q default must not contain quotes or backticks other than SQL single quotes", f.Name)
		}
		for _, ident := range []string{
			s.Entity.Table + "_namespace_" + f.Name + "_uidx",
			"chk_" + s.Entity.Table + "_" + f.Name,
		} {
			if len(ident) > postgresIdentifierMax {
				add("identifier %q is longer than %d bytes; shorten the table or field name", ident, postgresIdentifierMax)
			}
		}
		if seen[f.Name] {
			add("field %q is declared twice", f.Name)
		}
		seen[f.Name] = true
		if !fieldTypes[f.Type] {
			add("field %q has unknown type %q", f.Name, f.Type)
		}
		if f.Type == "enum" {
			if len(f.Enum) == 0 {
				add("enum field %q needs values", f.Name)
			}
			values := map[string]bool{}
			for _, v := range f.Enum {
				if !enumRe.MatchString(v) {
					add("enum value %q on %q must be UPPER_SNAKE", v, f.Name)
				}
				if values[v] {
					add("enum value %q on %q is listed twice", v, f.Name)
				}
				values[v] = true
			}
			if reservedTypeNames[s.Entity.Name+pascal(f.Name)] {
				add("enum type %q clashes with a type the module defines; rename the field", s.Entity.Name+pascal(f.Name))
			}
		}
		if f.Type == "ref" && !tableRe.MatchString(f.Ref) {
			add("ref field %q needs ref as schema.table, got %q", f.Name, f.Ref)
		}
		if f.RefIDType != "" && f.RefIDType != "string" && f.RefIDType != "int" {
			add("ref field %q ref_id_type must be string or int", f.Name)
		}
		if f.Type == "jsonb" && (f.Unique || f.Index) {
			add("jsonb field %q cannot be unique or indexed here; add a GIN index by hand", f.Name)
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("invalid spec:\n  - %s", strings.Join(problems, "\n  - "))
	}
	return nil
}

func (s *Spec) hasOp(op string) bool {
	for _, o := range s.Entity.Operations {
		if o == op {
			return true
		}
	}
	return false
}
