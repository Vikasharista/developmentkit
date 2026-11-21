// Command render turns a bootstrap spec into Go service, module or entity files.
//
//	go run . -spec /path/to/repo/.bootstrap/catalog.widget.json -out /path/to/repo
//
// It only creates files. A file that already exists is never overwritten; the run
// stops before writing anything if one would be. Edits to files that already exist
// (factories, resolver root, server wiring, indexes) are rendered as a patch list
// under .bootstrap/patches/ for the agent to apply.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"text/template"
)

func main() {
	specPath := flag.String("spec", "", "path to the spec JSON")
	outDir := flag.String("out", ".", "repository root to write into")
	templatesDir := flag.String("templates", defaultTemplatesDir(), "templates root")
	dryRun := flag.Bool("dry-run", false, "list the files without writing them")
	only := flag.String("only", "", "comma-separated glob patterns; render only matching output paths (e.g. Makefile,scripts/*)")
	flag.Parse()

	if err := run(*specPath, *templatesDir, *outDir, *dryRun, splitPatterns(*only), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "render:", err)
		os.Exit(1)
	}
}

// defaultTemplatesDir is library/bootstrap/templates, found from this source file so
// `go run` works from any directory.
func defaultTemplatesDir() string {
	_, self, _, ok := runtime.Caller(0)
	if !ok {
		return "templates"
	}
	return filepath.Join(filepath.Dir(self), "..", "templates")
}

// layers returns the template directories a mode renders, in order. With -only every
// layer is rendered and filtered, so a Makefile can be taken into an existing repo.
func layers(mode string, only []string) []string {
	if len(only) > 0 {
		return []string{"service", "module", "entity"}
	}
	switch mode {
	case "service":
		return []string{"service", "module", "entity"}
	case "module":
		return []string{"module", "entity"}
	default:
		return []string{"entity"}
	}
}

// patchLayer is the directory holding edits to files that already exist in the target.
func patchLayer(mode string) string {
	switch mode {
	case "module":
		return "module/_patches"
	case "entity":
		return "entity/_patches"
	}
	return ""
}

type output struct {
	path    string
	content []byte
	once    bool // shared helper: create when missing, keep when present
}

func splitPatterns(raw string) []string {
	var patterns []string
	for _, p := range strings.Split(raw, ",") {
		if p = strings.TrimSpace(p); p != "" {
			patterns = append(patterns, p)
		}
	}
	return patterns
}

// selected reports whether path matches one of the patterns; no patterns selects everything.
func selected(path string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, p := range patterns {
		if ok, _ := filepath.Match(p, path); ok {
			return true
		}
	}
	return false
}

func run(specPath, templatesDir, outDir string, dryRun bool, only []string, stdout io.Writer) error {
	if specPath == "" {
		return errors.New("-spec is required")
	}
	spec, err := loadSpec(specPath)
	if err != nil {
		return err
	}

	if _, err := os.Stat(templatesDir); err != nil {
		return fmt.Errorf("templates not found at %s: %w", templatesDir, err)
	}

	var outputs []output
	for _, layer := range layers(spec.Mode, only) {
		rendered, err := renderTree(filepath.Join(templatesDir, layer), spec)
		if err != nil {
			return err
		}
		for _, o := range rendered {
			if selected(o.path, only) {
				outputs = append(outputs, o)
			}
		}
	}

	var patches []output
	if layer := patchLayer(spec.Mode); layer != "" && len(only) == 0 {
		patches, err = renderTree(filepath.Join(templatesDir, layer), spec)
		if err != nil {
			return err
		}
	}

	if len(patches) > 0 {
		outputs = append(outputs, output{
			path:    filepath.Join(".bootstrap", "patches", spec.Module.Name+"."+spec.Entity.Snake()+".md"),
			content: joinPatches(patches),
		})
	}

	if len(outputs) == 0 {
		return errors.New("nothing to render: check -only and the spec mode")
	}

	var conflicts []string
	kept := outputs[:0]
	for _, o := range outputs {
		_, err := os.Stat(filepath.Join(outDir, o.path))
		switch {
		case err == nil && o.once:
			fmt.Fprintln(stdout, "keep  ", o.path)
		case err == nil:
			conflicts = append(conflicts, o.path)
		default:
			kept = append(kept, o)
		}
	}
	outputs = kept
	if len(conflicts) > 0 {
		return fmt.Errorf("refusing to overwrite existing files:\n  %s", strings.Join(conflicts, "\n  "))
	}

	for _, o := range outputs {
		fmt.Fprintln(stdout, "create", o.path)
		if dryRun {
			continue
		}
		target := filepath.Join(outDir, o.path)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		mode := os.FileMode(0o644)
		if strings.HasSuffix(o.path, ".sh") {
			mode = 0o755
		}
		if err := os.WriteFile(target, o.content, mode); err != nil {
			return err
		}
	}
	return nil
}

// renderTree renders every file under root. Path segments __module__, __entity__,
// __migration__ and __service__ are substituted. Files ending in .tmpl are
// executed as text/template against the spec and lose the suffix; a template
// that renders to whitespace only is skipped, which is how optional files
// (REST-only, GraphQL-only) switch off. A .once.tmpl file is a shared helper:
// created when missing, left alone when present.
func renderTree(root string, spec *Spec) ([]output, error) {
	if _, err := os.Stat(root); errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	replacer := strings.NewReplacer(
		"__module__", spec.Module.Name,
		"__entity__", spec.Entity.Snake(),
		"__migration__", spec.MigrationName(),
		"__service__", spec.Service.Name,
	)

	var outputs []output
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		if d.IsDir() {
			if d.Name() == "_patches" && path != root {
				return filepath.SkipDir
			}
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := raw
		once := false
		if strings.HasSuffix(rel, ".once.tmpl") {
			rel = strings.TrimSuffix(rel, ".once.tmpl") + ".tmpl"
			once = true
		}
		if strings.HasSuffix(rel, ".tmpl") {
			rel = strings.TrimSuffix(rel, ".tmpl")
			tmpl, err := template.New(rel).Funcs(funcMap()).Option("missingkey=error").Parse(string(raw))
			if err != nil {
				return fmt.Errorf("parse %s: %w", path, err)
			}
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, spec); err != nil {
				return fmt.Errorf("execute %s: %w", path, err)
			}
			if strings.TrimSpace(buf.String()) == "" {
				return nil
			}
			content = buf.Bytes()
		}
		rel = replacer.Replace(rel)
		if strings.HasSuffix(rel, ".go") {
			formatted, err := format.Source(content)
			if err != nil {
				return fmt.Errorf("gofmt %s: %w\n%s", rel, err, numbered(content))
			}
			content = formatted
		}
		outputs = append(outputs, output{path: rel, content: content, once: once})
		return nil
	})
	sort.Slice(outputs, func(i, j int) bool { return outputs[i].path < outputs[j].path })
	return outputs, err
}

func joinPatches(patches []output) []byte {
	var b bytes.Buffer
	b.WriteString("# Edits to existing files\n\nApply each edit, then delete this file. Generated by bootstrap.\n")
	for _, p := range patches {
		b.WriteString("\n")
		b.Write(bytes.TrimSpace(p.content))
		b.WriteString("\n")
	}
	return b.Bytes()
}

func numbered(src []byte) string {
	var b strings.Builder
	for i, line := range strings.Split(string(src), "\n") {
		fmt.Fprintf(&b, "%4d  %s\n", i+1, line)
	}
	return b.String()
}
