package main

import (
	"fmt"
	"strings"
	"text/template"
	"unicode"
)

// commonInitialisms mirrors the list gqlgen uses when it turns a GraphQL field
// name into a Go field name. copier matches fields by exact Go name, so the
// model, the DTO and the gqlgen model must agree on every initialism.
// Copied from gqlgen v0.17.86 codegen/templates/templates.go.
var commonInitialisms = map[string]bool{
	"ACL": true, "API": true, "ASCII": true, "AWS": true, "CPU": true, "CSS": true,
	"CSV": true, "DNS": true, "EOF": true, "GCP": true, "GUID": true, "HTML": true,
	"HTTP": true, "HTTPS": true, "ICMP": true, "ID": true, "IP": true, "JSON": true,
	"KVK": true, "LHS": true, "PDF": true, "PGP": true, "QPS": true, "QR": true,
	"RAM": true, "RHS": true, "RPC": true, "SLA": true, "SMTP": true, "SQL": true,
	"SSH": true, "SVG": true, "TCP": true, "TLS": true, "TTL": true, "UDP": true,
	"UI": true, "UID": true, "URI": true, "URL": true, "UTF8": true, "UUID": true,
	"VM": true, "XML": true, "XMPP": true, "XSRF": true, "XSS": true,
}

func words(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == '-' || r == ' ' })
}

// pascal turns snake or kebab case into an exported Go name: sku_code -> SKUCode.
func pascal(s string) string {
	var b strings.Builder
	for _, w := range words(s) {
		if commonInitialisms[strings.ToUpper(w)] {
			b.WriteString(strings.ToUpper(w))
			continue
		}
		b.WriteString(strings.ToUpper(w[:1]) + w[1:])
	}
	return b.String()
}

// camel turns snake or kebab case into a GraphQL / JSON field name: sku_code -> skuCode.
func camel(s string) string {
	ws := words(s)
	for i := range ws {
		if i == 0 {
			ws[i] = strings.ToLower(ws[i])
			continue
		}
		ws[i] = strings.ToUpper(ws[i][:1]) + ws[i][1:]
	}
	return strings.Join(ws, "")
}

// lowerFirst turns a PascalCase name into a Go local identifier: PurchaseOrder -> purchaseOrder.
func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// snake turns PascalCase into snake_case: PurchaseOrder -> purchase_order.
func snake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func kebab(s string) string { return strings.ReplaceAll(snake(s), "_", "-") }

// Field helpers, called from templates as {{ .GoName }} and so on.

func (f Field) GoName() string  { return pascal(f.Name) }
func (f Field) GQLName() string { return camel(f.Name) }

// GoType is the type on the GORM model and the service DTO.
func (f Field) GoType() string {
	var t string
	switch f.Type {
	case "string", "text", "enum":
		t = "string"
	case "ref":
		t = f.refGoType()
	case "int":
		t = "int"
	case "decimal":
		t = "float64"
	case "bool":
		t = "bool"
	case "timestamp":
		t = "time.Time"
	case "jsonb":
		return "model.Map"
	}
	if f.Ptr() {
		return "*" + t
	}
	return t
}

func (f Field) refGoType() string {
	if f.RefIDType == "int" {
		return "int"
	}
	return "string"
}

// enumWidth is 32, or the longest value when one is longer.
func (f Field) enumWidth() int {
	width := 32
	for _, v := range f.Enum {
		if len(v) > width {
			width = len(v)
		}
	}
	return width
}

// Ptr is true when the Go field is a pointer: the column is nullable, or it has a
// database default that a nil value must leave in charge.
func (f Field) Ptr() bool { return f.Nullable || f.Default != "" }

// DTOType is GoType with core-lib's model.Map spelled as a plain map so the dto
// package does not import the persistence model package.
func (f Field) DTOType() string {
	if f.Type == "jsonb" {
		return "map[string]interface{}"
	}
	return f.GoType()
}

// CreatePtr is true for a required non-text field that the create request carries
// as a pointer, so a missing value is rejected instead of stored as zero.
func (f Field) CreatePtr() bool {
	if f.Ptr() {
		return false
	}
	switch f.Type {
	case "int", "decimal", "bool", "timestamp":
		return true
	case "ref":
		return f.RefIDType == "int"
	}
	return false
}

// CreateType is the field's type on the create request.
func (f Field) CreateType() string {
	if f.CreatePtr() {
		return "*" + f.GoType()
	}
	return f.DTOType()
}

// SeedCreateGo is SeedGo as the create request carries it.
func (f Field) SeedCreateGo() string {
	if f.CreatePtr() {
		return "utility.GetPointerToPrimitiveType(" + f.SeedGo() + ")"
	}
	return f.SeedGo()
}

// IsText is true for fields stored as text that a patch must not blank when required.
func (f Field) IsText() bool {
	return f.Type == "string" || f.Type == "text" || f.Type == "enum" || (f.Type == "ref" && f.RefIDType != "int")
}

// PatchType is the pointer type an update request uses: nil means "leave as is".
func (f Field) PatchType() string {
	t := f.DTOType()
	if strings.HasPrefix(t, "*") || strings.HasPrefix(t, "map[") {
		return t
	}
	return "*" + t
}

// GQLBaseType is the GraphQL named type without nullability.
func (f Field) GQLBaseType(entity string) string {
	switch f.Type {
	case "string", "text":
		return "String"
	case "ref":
		if f.RefIDType == "int" {
			return "Int"
		}
		return "String"
	case "enum":
		return entity + pascal(f.Name)
	case "int":
		return "Int"
	case "decimal":
		return "Float"
	case "bool":
		return "Boolean"
	case "timestamp":
		return "Time"
	case "jsonb":
		return "Map"
	}
	return "String"
}

// GQLType is the output type, non-null when the column is NOT NULL.
func (f Field) GQLType(entity string) string {
	if f.Nullable {
		return f.GQLBaseType(entity)
	}
	return f.GQLBaseType(entity) + "!"
}

// GQLCreateType is the create input type: required unless nullable or defaulted.
func (f Field) GQLCreateType(entity string) string {
	if f.Nullable || f.Default != "" {
		return f.GQLBaseType(entity)
	}
	return f.GQLBaseType(entity) + "!"
}

// SQLType is the column type in the migration.
func (f Field) SQLType() string {
	switch f.Type {
	case "string":
		return fmt.Sprintf("varchar(%d)", f.MaxLen)
	case "text":
		return "text"
	case "enum":
		return fmt.Sprintf("varchar(%d)", f.enumWidth())
	case "ref":
		if f.RefIDType == "int" {
			return "integer"
		}
		return "varchar(64)"
	case "int":
		return "integer"
	case "decimal":
		return fmt.Sprintf("numeric(%d,%d)", f.Precision, f.Scale)
	case "bool":
		return "boolean"
	case "timestamp":
		return "timestamptz"
	case "jsonb":
		return "jsonb"
	}
	return "text"
}

// GormTag is the struct tag for the GORM model.
func (f Field) GormTag() string {
	tag := "column:" + f.Name
	if f.Type == "jsonb" {
		tag += ";type:jsonb"
	}
	if f.Default != "" {
		tag += ";default:" + f.Default
	}
	return fmt.Sprintf("`gorm:\"%s\"`", tag)
}

// Required is true when a create request must carry the field.
func (f Field) Required() bool { return !f.Nullable && f.Default == "" }

// ValidateTag is the core-lib validator tag on the create request, as a struct tag
// fragment with a leading space. Nullable
// fields carry none: the column constraints are the backstop for those.
func (f Field) ValidateTag() string {
	if f.Ptr() {
		return ""
	}
	if f.CreatePtr() {
		return " validate:\"nonzero\""
	}
	if f.Required() && (f.Type == "string" || f.Type == "text" || f.Type == "enum" || f.Type == "ref") {
		if f.Type == "string" {
			return fmt.Sprintf(" validate:\"required,max=%d\"", f.MaxLen)
		}
		return " validate:\"required\""
	}
	if f.Type == "string" {
		return fmt.Sprintf(" validate:\"max=%d\"", f.MaxLen)
	}
	return ""
}

// SkipInSamples is true for fields generated tests leave unset: nullable
// references, which would need a parent row.
func (f Field) SkipInSamples() bool { return f.Type == "ref" && f.Nullable }

// SeedGo is a Go expression for the field's value in generated tests, built from
// an int variable named seed so two seeds never collide on unique columns.
func (f Field) SeedGo() string {
	var v string
	switch f.Type {
	case "string", "text", "ref":
		v = f.seedString()
		if f.Type == "ref" && f.RefIDType == "int" {
			v = "seed"
		}
	case "enum":
		v = fmt.Sprintf("%q", f.Enum[0])
	case "int":
		v = "seed"
	case "decimal":
		v = "float64(seed) + 0.5"
	case "bool":
		v = "true"
	case "timestamp":
		v = "time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)"
	case "jsonb":
		return "map[string]interface{}{\"seed\": float64(seed)}"
	}
	if f.Ptr() {
		return "utility.GetPointerToPrimitiveType(" + v + ")"
	}
	return v
}

// seedString formats seed into a value that fits the column: "<name>-<seed>" when
// there is room, otherwise zero-padded digits as wide as the column.
func (f Field) seedString() string {
	prefix := f.Name
	if f.Type == "ref" {
		prefix = "ref"
	}
	if f.Type != "string" || f.MaxLen >= len(prefix)+6 {
		return fmt.Sprintf("fmt.Sprintf(\"%s-%%d\", seed)", prefix)
	}
	return fmt.Sprintf("fmt.Sprintf(\"%%0%dd\", seed)", f.MaxLen)
}

// SeedPatchGo is SeedGo as the pointer an update request carries.
func (f Field) SeedPatchGo() string {
	if f.Type == "jsonb" || f.Ptr() {
		return f.SeedGo()
	}
	return "utility.GetPointerToPrimitiveType(" + f.SeedGo() + ")"
}

// SeedJSONGo is a Go expression for the field's value in a GraphQL variables or
// REST JSON payload, built from seed.
func (f Field) SeedJSONGo() string {
	switch f.Type {
	case "string", "text", "ref":
		if f.Type == "ref" && f.RefIDType == "int" {
			return "seed"
		}
		return f.seedString()
	case "enum":
		return fmt.Sprintf("%q", f.Enum[0])
	case "int":
		return "seed"
	case "decimal":
		return "float64(seed) + 0.5"
	case "bool":
		return "true"
	case "timestamp":
		return "\"2026-01-02T03:04:05Z\""
	case "jsonb":
		return "map[string]interface{}{\"seed\": float64(seed)}"
	}
	return "nil"
}

// IsTime reports whether the field needs the time import.
func (f Field) IsTime() bool { return f.Type == "timestamp" }

// Entity helpers.

func (e Entity) Var() string         { return lowerFirst(e.Name) }
func (e Entity) PluralVar() string   { return lowerFirst(e.Plural) }
func (e Entity) Snake() string       { return snake(e.Name) }
func (e Entity) PluralSnake() string { return snake(e.Plural) }
func (e Entity) Kebab() string       { return kebab(e.Name) }
func (e Entity) PluralKebab() string { return kebab(e.Plural) }

func (e Entity) IDGo() string {
	if e.IDType == "int" {
		return "int"
	}
	return "string"
}

func (e Entity) IDGQL() string {
	if e.IDType == "int" {
		return "Int"
	}
	return "String"
}

// IDSample is a well-formed ID that does not exist, used by not-found tests.
func (e Entity) IDSample() string {
	if e.IDType == "int" {
		return "987654"
	}
	return "\"00000000-0000-0000-0000-000000000000\""
}

func (e Entity) BaseModel() string {
	switch {
	case e.IDType == "int" && e.Versioned:
		return "VersionedNamespaceModelIDInt"
	case e.IDType == "int":
		return "NamespaceModelIDInt"
	case e.Versioned:
		return "VersionedNamespaceModelIDString"
	default:
		return "NamespaceModelIDString"
	}
}

func (e Entity) NeedsTime() bool {
	for _, f := range e.Fields {
		if f.IsTime() {
			return true
		}
	}
	return false
}

// NeedsUtility is true when generated test samples wrap a value in a pointer.
func (e Entity) NeedsUtility() bool {
	for _, f := range e.Fields {
		if (f.Ptr() || f.CreatePtr()) && f.Type != "jsonb" && !f.SkipInSamples() {
			return true
		}
	}
	return false
}

// RowNeedsUtility is true when sample model rows wrap a value in a pointer.
func (e Entity) RowNeedsUtility() bool {
	for _, f := range e.Fields {
		if f.Ptr() && f.Type != "jsonb" && !f.SkipInSamples() {
			return true
		}
	}
	return false
}

// NeedsFmt is true when generated test samples format a seed into a string.
func (e Entity) NeedsFmt() bool {
	for _, f := range e.Fields {
		if (f.Type == "string" || f.Type == "text" || (f.Type == "ref" && f.RefIDType != "int")) && !f.SkipInSamples() {
			return true
		}
	}
	return false
}

// HasType reports whether any field has the given type.
func (e Entity) HasType(t string) bool {
	for _, f := range e.Fields {
		if f.Type == t {
			return true
		}
	}
	return false
}

// SampleFields are the fields generated tests populate.
func (e Entity) SampleFields() []Field {
	var out []Field
	for _, f := range e.Fields {
		if !f.SkipInSamples() {
			out = append(out, f)
		}
	}
	return out
}

func (e Entity) Enums() []Field {
	var out []Field
	for _, f := range e.Fields {
		if f.Type == "enum" {
			out = append(out, f)
		}
	}
	return out
}

// Sortable lists columns a list query may filter or sort on. Everything except
// jsonb, plus the audit columns.
func (e Entity) Sortable() []string {
	cols := []string{"id", "is_active", "created_at", "updated_at", "created_by", "updated_by"}
	if e.Versioned {
		cols = append(cols, "version")
	}
	for _, f := range e.Fields {
		if f.Type != "jsonb" {
			cols = append(cols, f.Name)
		}
	}
	return cols
}

func (e Entity) Has(op string) bool {
	for _, o := range e.Operations {
		if o == op {
			return true
		}
	}
	return false
}

// Directive is the GraphQL auth for an operation. A permission check always runs
// after @authenticate, so the audit columns come from the verified session and
// never from the X-User-Id header.
func (e Entity) Directive(op string) string {
	if p, ok := e.Permissions[op]; ok {
		return fmt.Sprintf("@authenticate @hasPermission(permission: %q)", p)
	}
	return "@authenticate"
}

// Permission is the permission string for an operation, empty when it only needs authentication.
func (e Entity) Permission(op string) string { return e.Permissions[op] }

// UsesPermissions reports whether any operation carries a permission.
func (e Entity) UsesPermissions() bool { return len(e.Permissions) > 0 }

func (e Entity) UniqueFields() []Field {
	var out []Field
	for _, f := range e.Fields {
		if f.Unique {
			out = append(out, f)
		}
	}
	return out
}

// Spec-level helpers.

// Table is the schema-qualified table name used by the repository and the migration.
func (s *Spec) Table() string { return s.Module.PgSchema + "." + s.Entity.Table }

// MigrationName is the 6-digit migration file stem.
func (s *Spec) MigrationName() string {
	return fmt.Sprintf("%06d_%s", s.Migration.Number, s.Migration.Description)
}

// ModuleImport is the import path of the module directory.
func (s *Spec) ModuleImport() string {
	return s.Service.ModulePath + "/internal/app/" + s.Module.Name
}

// ModuleAlias is a Go identifier for importing the module's graphql or rest package.
func (s *Spec) ModuleAlias() string { return lowerFirst(pascal(s.Module.Name)) }

// ModuleTitle is a human title for docs: purchase_planner -> Purchase Planner.
func (s *Spec) ModuleTitle() string {
	ws := words(s.Module.Name)
	for i := range ws {
		ws[i] = strings.ToUpper(ws[i][:1]) + ws[i][1:]
	}
	return strings.Join(ws, " ")
}

// LogPrefix is the zap message prefix: <Module>.<Entity>.
func (s *Spec) LogPrefix() string { return pascal(s.Module.Name) + "." + s.Entity.Name }

func funcMap() template.FuncMap {
	return template.FuncMap{
		"add":      func(a, b int) int { return a + b },
		"sqlquote": func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" },
		"sqlList": func(vs []string) string {
			out := make([]string, len(vs))
			for i, v := range vs {
				out[i] = "'" + v + "'"
			}
			return strings.Join(out, ", ")
		},
	}
}
