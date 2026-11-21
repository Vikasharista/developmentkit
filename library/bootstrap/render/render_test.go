package main

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const examplesDir = "../examples"
const templatesRoot = "../templates"

func writeSpec(t *testing.T, mutate func(*Spec)) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(examplesDir, "widget-service.json"))
	if err != nil {
		t.Fatal(err)
	}
	spec := &Spec{}
	if err := json.Unmarshal(raw, spec); err != nil {
		t.Fatal(err)
	}
	if mutate != nil {
		mutate(spec)
	}
	out, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "spec.json")
	if err := os.WriteFile(path, out, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func render(t *testing.T, specPath, outDir string, only ...string) error {
	t.Helper()
	return run(specPath, templatesRoot, outDir, false, only, io.Discard)
}

func TestNaming(t *testing.T) {
	cases := []struct {
		fn       func(string) string
		in, want string
	}{
		{pascal, "sku_code", "SkuCode"}, // SKU is not a gqlgen initialism
		{pascal, "image_url", "ImageURL"},
		{pascal, "customer_id", "CustomerID"},
		{pascal, "purchase-planner", "PurchasePlanner"},
		{camel, "sku_code", "skuCode"},
		{camel, "customer_id", "customerId"},
		{snake, "PurchaseOrder", "purchase_order"},
		{kebab, "PurchaseOrder", "purchase-order"},
		{lowerFirst, "PurchaseOrder", "purchaseOrder"},
	}
	for _, tc := range cases {
		if got := tc.fn(tc.in); got != tc.want {
			t.Errorf("%q: got %q want %q", tc.in, got, tc.want)
		}
	}
}

func TestFieldTypes(t *testing.T) {
	cases := []struct {
		field                Field
		goType, gql, sqlType string
	}{
		{Field{Name: "code", Type: "string", MaxLen: 64}, "string", "String!", "varchar(64)"},
		{Field{Name: "note", Type: "text", Nullable: true}, "*string", "String", "text"},
		{Field{Name: "qty", Type: "int", Default: "0"}, "*int", "Int!", "integer"},
		{Field{Name: "price", Type: "decimal", Precision: 12, Scale: 2}, "float64", "Float!", "numeric(12,2)"},
		{Field{Name: "status", Type: "enum", Enum: []string{"A"}}, "string", "WidgetStatus!", "varchar(32)"},
		{Field{Name: "meta", Type: "jsonb", Nullable: true}, "model.Map", "Map", "jsonb"},
		{Field{Name: "at", Type: "timestamp"}, "time.Time", "Time!", "timestamptz"},
		{Field{Name: "owner_id", Type: "ref", Ref: "public.users"}, "string", "String!", "varchar(64)"},
	}
	for _, tc := range cases {
		if got := tc.field.GoType(); got != tc.goType {
			t.Errorf("%s GoType: got %q want %q", tc.field.Name, got, tc.goType)
		}
		if got := tc.field.GQLType("Widget"); got != tc.gql {
			t.Errorf("%s GQLType: got %q want %q", tc.field.Name, got, tc.gql)
		}
		if got := tc.field.SQLType(); got != tc.sqlType {
			t.Errorf("%s SQLType: got %q want %q", tc.field.Name, got, tc.sqlType)
		}
	}
}

func TestSeedStringFitsColumn(t *testing.T) {
	long := Field{Name: "code", Type: "string", MaxLen: 64}
	short := Field{Name: "un_locode", Type: "string", MaxLen: 5}
	if got := long.seedString(); got != `fmt.Sprintf("code-%d", seed)` {
		t.Errorf("long: %s", got)
	}
	if got := short.seedString(); got != `fmt.Sprintf("%05d", seed)` {
		t.Errorf("short: %s", got)
	}
}

func TestValidate(t *testing.T) {
	cases := map[string]func(*Spec){
		"bad mode":         func(s *Spec) { s.Mode = "app" },
		"reserved column":  func(s *Spec) { s.Entity.Fields[0].Name = "namespace" },
		"duplicate field":  func(s *Spec) { s.Entity.Fields[1].Name = s.Entity.Fields[0].Name },
		"empty enum":       func(s *Spec) { s.Entity.Fields[5].Enum = nil },
		"lower enum value": func(s *Spec) { s.Entity.Fields[5].Enum = []string{"draft"} },
		"bad ref":          func(s *Spec) { s.Entity.Fields[0].Type, s.Entity.Fields[0].Ref = "ref", "users" },
		"no transport":     func(s *Spec) { s.Module.GraphQL, s.Module.REST = false, false },
		"loader sans get":  func(s *Spec) { s.Entity.Operations = []string{"create", "list"} },
		"stray permission": func(s *Spec) { s.Entity.Operations = []string{"create", "get"} },
		"audit migration":  func(s *Spec) { s.Migration.Number = 1 },
		"indexed jsonb":    func(s *Spec) { s.Entity.Fields[6].Index = true },
		"sql keyword":      func(s *Spec) { s.Entity.Fields[1].Name = "order" },
		"key field":        func(s *Spec) { s.Entity.Fields[1].Name = "key" },
		"double capitals":  func(s *Spec) { s.Entity.Name = "QRCode" },
		"gqlgen type":      func(s *Spec) { s.Entity.Name, s.Entity.Plural = "Config", "Configs" },
		"go keyword":       func(s *Spec) { s.Entity.Name, s.Entity.Plural = "Package", "Packages" },
		"duplicate enum":   func(s *Spec) { s.Entity.Fields[5].Enum = []string{"A", "A"} },
		"quoted default":   func(s *Spec) { s.Entity.Fields[3].Default = `"0"` },
		"long identifier":  func(s *Spec) { s.Entity.Fields[1].Name = strings.Repeat("n", 60) },
		"bad ref id type": func(s *Spec) {
			s.Entity.Fields[0].Type, s.Entity.Fields[0].Ref, s.Entity.Fields[0].RefIDType = "ref", "public.users", "uuid"
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := loadSpec(writeSpec(t, mutate)); err == nil {
				t.Fatal("want a validation error")
			}
		})
	}
	if _, err := loadSpec(writeSpec(t, nil)); err != nil {
		t.Fatalf("example spec must be valid: %v", err)
	}
}

func TestRenderService(t *testing.T) {
	out := t.TempDir()
	if err := render(t, writeSpec(t, nil), out); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		"go.mod", "server.go", "Makefile", "docker-compose.yml", "AGENTS.md", "README.md",
		"migrations/000001_audit_schema.up.sql", "migrations/000002_catalog_widgets.up.sql",
		"internal/app/catalog/graphql/schema/widget.graphqls", "internal/app/catalog/rest/widget_controller.go",
		"internal/app/catalog/service/widget_test.go", "internal/testutil/pgtest/pgtest.go",
	} {
		if _, err := os.Stat(filepath.Join(out, path)); err != nil {
			t.Errorf("missing %s", path)
		}
	}
	if _, err := os.Stat(filepath.Join(out, ".bootstrap", "patches")); err == nil {
		t.Error("service mode must not emit patches")
	}
	info, err := os.Stat(filepath.Join(out, "scripts", "smoke.sh"))
	if err != nil || info.Mode()&0o100 == 0 {
		t.Error("shell scripts must be executable")
	}
	schema, _ := os.ReadFile(filepath.Join(out, "internal/app/catalog/graphql/schema/widget.graphqls"))
	if !strings.Contains(string(schema), `deleteWidget(id: String!): Widget! @authenticate @hasPermission(permission: "your-org::widget-service::widget-service::widget::delete")`) {
		t.Errorf("permission directive missing:\n%s", schema)
	}

	// A second run must refuse to overwrite.
	if err := render(t, writeSpec(t, nil), out); err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("want refusal, got %v", err)
	}
}

func TestRenderEntityIntoExistingModule(t *testing.T) {
	out := t.TempDir()
	if err := render(t, writeSpec(t, nil), out); err != nil {
		t.Fatal(err)
	}
	second := writeSpec(t, func(s *Spec) {
		s.Mode = "entity"
		s.Entity.Name, s.Entity.Plural, s.Entity.Table = "Gadget", "Gadgets", "gadgets"
		s.Migration = Migration{Number: 3, Description: "catalog_gadgets"}
	})
	if err := render(t, second, out); err != nil {
		t.Fatalf("shared .once files must be kept, not refused: %v", err)
	}
	patches, err := os.ReadFile(filepath.Join(out, ".bootstrap", "patches", "catalog.gadget.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"GetGadgetRepository", "GadgetService service.GadgetService", "FindGadgetByID", "registerGadgetRoutes"} {
		if !strings.Contains(string(patches), want) {
			t.Errorf("patch list lacks %s", want)
		}
	}
	if _, err := os.Stat(filepath.Join(out, "internal/app/catalog/service/gadget.go")); err != nil {
		t.Error("entity files missing")
	}
	requireNoDuplicateDeclarations(t, out)
}

// requireNoDuplicateDeclarations fails when two files of one package (test files
// counted with the package) declare the same top-level name, which is what a second
// entity in a module would hit if a template used an unprefixed helper name.
func requireNoDuplicateDeclarations(t *testing.T, root string) {
	t.Helper()
	declared := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		for _, decl := range file.Decls {
			for _, name := range topLevelNames(decl) {
				key := filepath.Dir(path) + ":" + file.Name.Name + ":" + name
				if other, seen := declared[key]; seen {
					t.Errorf("%s declared in %s and %s", name, other, path)
				}
				declared[key] = path
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func topLevelNames(decl ast.Decl) []string {
	var names []string
	switch d := decl.(type) {
	case *ast.FuncDecl:
		if d.Recv == nil && d.Name.Name != "init" {
			names = append(names, d.Name.Name)
		}
	case *ast.GenDecl:
		for _, spec := range d.Specs {
			switch s := spec.(type) {
			case *ast.TypeSpec:
				names = append(names, s.Name.Name)
			case *ast.ValueSpec:
				for _, n := range s.Names {
					if n.Name != "_" {
						names = append(names, n.Name)
					}
				}
			}
		}
	}
	return names
}

func TestRenderFailsLoudly(t *testing.T) {
	if err := run(writeSpec(t, nil), "/no/such/templates", t.TempDir(), false, nil, io.Discard); err == nil {
		t.Error("a missing templates root must fail")
	}
	if err := render(t, writeSpec(t, nil), t.TempDir(), "no-such-file"); err == nil {
		t.Error("-only matching nothing must fail")
	}
}

func TestRenderOnlyMakefileForExistingRepo(t *testing.T) {
	out := t.TempDir()
	spec := writeSpec(t, func(s *Spec) { s.Mode = "entity" })
	if err := render(t, spec, out, "Makefile"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "Makefile")); err != nil {
		t.Error("-only must reach the service layer in entity mode")
	}
}

func TestPatchFileIsNotOverwritten(t *testing.T) {
	out := t.TempDir()
	spec := writeSpec(t, func(s *Spec) { s.Mode = "entity" })
	if err := render(t, spec, out); err != nil {
		t.Fatal(err)
	}
	if err := render(t, spec, out); err == nil {
		t.Error("a second entity-mode run must refuse, patch file included")
	}
}

func TestIntRefAndCreatePointers(t *testing.T) {
	ref := Field{Name: "store_id", Type: "ref", Ref: "public.stores", RefIDType: "int"}
	if ref.GoType() != "int" || ref.SQLType() != "integer" || ref.GQLBaseType("X") != "Int" || ref.SeedGo() != "seed" {
		t.Errorf("int ref: %s %s %s %s", ref.GoType(), ref.SQLType(), ref.GQLBaseType("X"), ref.SeedGo())
	}
	if !ref.CreatePtr() || ref.CreateType() != "*int" {
		t.Error("required int ref must be a pointer on create")
	}
	at := Field{Name: "at", Type: "timestamp"}
	if at.CreateType() != "*time.Time" || at.ValidateTag() != ` validate:"nonzero"` {
		t.Errorf("required timestamp: %s %s", at.CreateType(), at.ValidateTag())
	}
	long := Field{Name: "s", Type: "enum", Enum: []string{strings.Repeat("A", 40)}}
	if long.SQLType() != "varchar(40)" {
		t.Errorf("enum width: %s", long.SQLType())
	}
}

func TestRenderOnly(t *testing.T) {
	out := t.TempDir()
	if err := render(t, writeSpec(t, nil), out, "Makefile", "scripts/*"); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(out)
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if strings.Join(names, ",") != "Makefile,scripts" {
		t.Fatalf("got %v", names)
	}
}

func TestRestOnlyServiceHasNoGraphQL(t *testing.T) {
	out := t.TempDir()
	spec := writeSpec(t, func(s *Spec) { s.Module.GraphQL = false; s.Entity.Loader = false })
	if err := render(t, spec, out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "internal/app/catalog/graphql")); err == nil {
		t.Error("REST-only module must not get a graphql directory")
	}
	if _, err := os.Stat(filepath.Join(out, "internal/app/catalog/gqlgen.yml")); err == nil {
		t.Error("REST-only module must not get gqlgen.yml")
	}
}
