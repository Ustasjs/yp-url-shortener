// Command reset generates Reset() methods for annotated structs.
//
// It scans every package in the module, starting from the root directory and
// below, and looks for struct declarations preceded by a
//
//	// generate:reset
//
// comment. For each such struct it generates a Reset() method that resets the
// value to its zero state and writes all generated methods for a package into a
// reset.gen.go file in that same package.
//
// The reset rules are:
//
//   - primitives are set to their zero value (int to 0, string to "", bool to
//     false and so on);
//   - slices are truncated to zero length but not niled (s = s[:0]);
//   - maps are emptied with the built-in clear;
//   - fields whose type has a Reset() method call that method;
//   - non-nil pointers reset their pointed-to value by the same rules.
//
// Usage:
//
//	go run ./cmd/reset [-dir root]
//
// The -dir flag selects the root directory to scan and defaults to the current
// directory.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// marker is the comment that opts a struct into Reset() generation.
const marker = "generate:reset"

// generatedFile is the name of the file written into each package that has at
// least one annotated struct.
const generatedFile = "reset.gen.go"

func main() {
	dir := flag.String("dir", ".", "root directory to scan for annotated structs")
	flag.Parse()

	if err := run(*dir); err != nil {
		log.Fatalf("reset: %v", err)
	}
}

// target groups the annotated struct types found in a single package together
// with the directory their generated file must be written to.
type target struct {
	pkgName string
	dir     string
	structs []*types.Named
}

func run(root string) error {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports |
			packages.NeedDeps,
		Dir: root,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return fmt.Errorf("load packages: %w", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		// Type information may be incomplete, but generation can still proceed
		// for the packages that loaded cleanly, so only warn.
		log.Print("reset: some packages contain errors; generated output may be incomplete")
	}

	g := &generator{marked: map[*types.TypeName]bool{}}
	targets := g.collect(pkgs)

	for _, t := range targets {
		if err := g.write(t); err != nil {
			return err
		}
	}
	return nil
}

// generator holds the set of struct types that are known to receive a generated
// Reset() method
type generator struct {
	marked map[*types.TypeName]bool
}

// collect scans the loaded packages for annotated structs. It records every
// annotated struct in the marked set (so cross-references can be resolved) and
// returns the per-package targets
func (g *generator) collect(pkgs []*packages.Package) []*target {
	byPkg := map[*packages.Package]*target{}

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				gd, ok := decl.(*ast.GenDecl)
				if !ok || gd.Tok != token.TYPE {
					continue
				}
				for _, spec := range gd.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if !isMarked(gd.Doc) && !isMarked(ts.Doc) {
						continue
					}

					named, ok := namedStruct(pkg, ts)
					if !ok {
						log.Printf("reset: %s: %s is annotated but is not a struct; skipping",
							pkg.PkgPath, ts.Name.Name)
						continue
					}

					g.marked[named.Obj()] = true

					t := byPkg[pkg]
					if t == nil {
						t = &target{
							pkgName: pkg.Name,
							dir:     filepath.Dir(pkg.Fset.Position(ts.Pos()).Filename),
						}
						byPkg[pkg] = t
					}
					t.structs = append(t.structs, named)
				}
			}
		}
	}

	targets := make([]*target, 0, len(byPkg))
	for _, t := range byPkg {
		sort.Slice(t.structs, func(i, j int) bool {
			return t.structs[i].Obj().Name() < t.structs[j].Obj().Name()
		})
		targets = append(targets, t)
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].dir < targets[j].dir })
	return targets
}

// namedStruct resolves the type spec to its named struct type, reporting false
// when the annotated declaration is not a struct.
func namedStruct(pkg *packages.Package, ts *ast.TypeSpec) (*types.Named, bool) {
	obj, ok := pkg.TypesInfo.Defs[ts.Name].(*types.TypeName)
	if !ok {
		return nil, false
	}
	named, ok := obj.Type().(*types.Named)
	if !ok {
		return nil, false
	}
	if _, ok := named.Underlying().(*types.Struct); !ok {
		return nil, false
	}
	return named, true
}

// isMarked reports whether a comment group contains the generation marker.
func isMarked(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}
	for _, c := range doc.List {
		text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
		if text == marker {
			return true
		}
	}
	return false
}

// write renders and writes the reset.gen.go file for a single target package.
func (g *generator) write(t *target) error {
	var b strings.Builder
	fmt.Fprintf(&b, "// Code generated by \"go run ./cmd/reset\"; DO NOT EDIT.\n\n")
	fmt.Fprintf(&b, "package %s\n\n", t.pkgName)
	for i, named := range t.structs {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(g.method(named))
	}

	src, err := format.Source([]byte(b.String()))
	if err != nil {
		return fmt.Errorf("format %s: %w", t.dir, err)
	}

	path := filepath.Join(t.dir, generatedFile)
	if err := os.WriteFile(path, src, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	log.Printf("reset: wrote %s (%d struct(s))", path, len(t.structs))
	return nil
}

// method renders the Reset() method for a single struct. gofmt normalises the
// indentation afterwards, so the raw body only needs to be syntactically valid.
func (g *generator) method(named *types.Named) string {
	recv := receiverName(named.Obj().Name())
	st := named.Underlying().(*types.Struct)

	var body []string
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		if f.Name() == "_" {
			continue
		}
		body = append(body, g.fieldStmts(recv, f.Name(), f.Type())...)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "func (%s *%s) Reset() {\n", recv, named.Obj().Name())
	fmt.Fprintf(&b, "if %s == nil {\nreturn\n}\n", recv)
	if len(body) > 0 {
		b.WriteString("\n")
		for _, line := range body {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	b.WriteString("}\n")
	return b.String()
}

// fieldStmts returns the statements that reset a single field.
func (g *generator) fieldStmts(recv, name string, t types.Type) []string {
	expr := recv + "." + name

	if p, ok := t.(*types.Pointer); ok {
		if g.hasReset(p.Elem()) {
			return wrapNotNil(expr, []string{expr + ".Reset()"})
		}
		if stmt, ok := primitiveReset("(*"+expr+")", p.Elem()); ok {
			return wrapNotNil(expr, []string{stmt})
		}
		return nil
	}

	if g.hasReset(t) {
		return []string{expr + ".Reset()"}
	}
	if stmt, ok := primitiveReset(expr, t); ok {
		return []string{stmt}
	}
	return nil
}

// primitiveReset returns the statement that resets a non-pointer primitive,
// slice or map expression, reporting false for any other type.
func primitiveReset(expr string, t types.Type) (string, bool) {
	switch u := t.Underlying().(type) {
	case *types.Basic:
		lit, ok := zeroValue(u)
		if !ok {
			return "", false
		}
		return expr + " = " + lit, true
	case *types.Slice:
		return expr + " = " + expr + "[:0]", true
	case *types.Map:
		return "clear(" + expr + ")", true
	}
	return "", false
}

// zeroValue returns the literal zero value for a basic type.
func zeroValue(b *types.Basic) (string, bool) {
	switch info := b.Info(); {
	case info&types.IsBoolean != 0:
		return "false", true
	case info&types.IsString != 0:
		return `""`, true
	case info&types.IsNumeric != 0:
		return "0", true
	}
	return "", false
}

// hasReset reports whether the given type will have a Reset() method, either
// because it is one of the annotated structs or because it already declares a
// Reset() method that takes no arguments and returns nothing.
func (g *generator) hasReset(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	if g.marked[named.Obj()] {
		return true
	}
	ms := types.NewMethodSet(types.NewPointer(named))
	for i := 0; i < ms.Len(); i++ {
		fn := ms.At(i).Obj()
		if fn.Name() != "Reset" {
			continue
		}
		sig, ok := fn.Type().(*types.Signature)
		if ok && sig.Params().Len() == 0 && sig.Results().Len() == 0 {
			return true
		}
	}
	return false
}

// wrapNotNil guards the given statements with a nil check on expr.
func wrapNotNil(expr string, stmts []string) []string {
	out := make([]string, 0, len(stmts)+2)
	out = append(out, "if "+expr+" != nil {")
	out = append(out, stmts...)
	out = append(out, "}")
	return out
}

// receiverName derives a short method receiver name from a type name.
func receiverName(typeName string) string {
	if typeName == "" {
		return "r"
	}
	return strings.ToLower(typeName[:1])
}
