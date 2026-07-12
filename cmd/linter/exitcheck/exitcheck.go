// Package exitcheck defines an analyzer that reports a direct call to
// os.Exit inside the main function of package main.
package exitcheck

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// Analyzer reports a direct call to os.Exit in the main function of the main
// package.
var Analyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "reports a direct call to os.Exit in the main function of package main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name.Name != "main" || fn.Body == nil {
				continue
			}
			checkMainBody(pass, fn.Body)
		}
	}

	return nil, nil
}

// checkMainBody walks the body of main, reporting os.Exit calls but not
// descending into nested function literals.
func checkMainBody(pass *analysis.Pass, body *ast.BlockStmt) {
	ast.Inspect(body, func(n ast.Node) bool {
		// Do not descend into nested closures: a call there is out of scope.
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if isOSExit(pass, call) {
			pass.Reportf(call.Pos(), "direct call to os.Exit in main function of package main is forbidden")
		}
		return true
	})
}

// isOSExit reports whether the call expression is a call to os.Exit, resolving
// the package through type information so an aliased import is still caught.
func isOSExit(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Exit" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	pkgName, ok := pass.TypesInfo.Uses[ident].(*types.PkgName)
	if !ok {
		return false
	}
	return pkgName.Imported().Path() == "os"
}
