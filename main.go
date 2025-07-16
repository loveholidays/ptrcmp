/*
ptrcmp
Copyright (C) 2025  loveholidays
This program is free software; you can redistribute it and/or
modify it under the terms of the GNU Lesser General Public
License as published by the Free Software Foundation; either
version 3 of the License, or (at your option) any later version.
This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
Lesser General Public License for more details.
You should have received a copy of the GNU Lesser General Public License
along with this program; if not, write to the Free Software Foundation,
Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
*/
package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: ptrcmp <directory>")
	}
	dir := os.Args[1]

	results, err := parseDir(dir)
	if err != nil {
		log.Fatalf("Error %v", err)
	}
	for _, result := range results {
		println(result)
	}
}

func parseDir(dir string) ([]string, error) {
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps,
		Dir:  dir,
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return []string{}, fmt.Errorf("failed to load packages: %v", err)
	}

	var errs []error
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			errs = append(errs, err)
		}
	})
	if len(errs) > 0 {
		log.Println("Packages contain errors:")
		for _, err := range errs {
			log.Println(err)
		}
	}

	ptrAnalyzer := NewPtrAnalyzer()
	results := make([]string, 0)

	for _, pkg := range pkgs {
		diagnostics, err := runAnalyzer(ptrAnalyzer, pkg)
		if err != nil {
			return []string{}, fmt.Errorf("failed to run analyzer on package %s: %v", pkg.Name, err)
		}
		for _, diag := range diagnostics {
			pos := pkg.Fset.Position(diag.Pos)
			results = append(results, fmt.Sprintf("%s:%d:%d: %s", pos.Filename, pos.Line, pos.Column, diag.Message))
		}
	}
	return results, nil
}

func runAnalyzer(analyzer *analysis.Analyzer, pkg *packages.Package) ([]analysis.Diagnostic, error) {
	var diagnostics []analysis.Diagnostic

	resultOf := make(map[*analysis.Analyzer]interface{})
	for _, dep := range analyzer.Requires {
		pass := &analysis.Pass{
			Analyzer:   dep,
			Fset:       pkg.Fset,
			Files:      pkg.Syntax,
			OtherFiles: nil,
			Pkg:        pkg.Types,
			TypesInfo:  pkg.TypesInfo,
			TypesSizes: pkg.TypesSizes,
			ResultOf:   resultOf,
			Report:     func(d analysis.Diagnostic) {},
		}

		result, err := dep.Run(pass)
		if err != nil {
			return nil, fmt.Errorf("failed to run dependency %s: %v", dep.Name, err)
		}
		resultOf[dep] = result
	}

	pass := &analysis.Pass{
		Analyzer:   analyzer,
		Fset:       pkg.Fset,
		Files:      pkg.Syntax,
		OtherFiles: nil,
		Pkg:        pkg.Types,
		TypesInfo:  pkg.TypesInfo,
		TypesSizes: pkg.TypesSizes,
		ResultOf:   resultOf,
		Report: func(d analysis.Diagnostic) {
			diagnostics = append(diagnostics, d)
		},
	}

	_, err := analyzer.Run(pass)
	if err != nil {
		return nil, err
	}

	return diagnostics, nil
}

func NewPtrAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "ptrcmp",
		Doc:      "checks that there are no pointer comparisons between basic types",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      run,
	}
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.BinaryExpr)(nil)}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		binaryExpr, ok := n.(*ast.BinaryExpr)
		if !ok {
			return
		}

		if isComparisonOperator(binaryExpr.Op) {
			checkPointerComparison(pass, binaryExpr)
		}
	})

	return nil, nil
}

// isComparisonOperator checks if the token is a comparison operator
func isComparisonOperator(op token.Token) bool {
	switch op {
	case token.EQL, token.NEQ, token.LSS, token.GTR, token.LEQ, token.GEQ:
		return true
	default:
		return false
	}
}

// checkPointerComparison checks if both operands are pointers to basic types
func checkPointerComparison(pass *analysis.Pass, binaryExpr *ast.BinaryExpr) {
	if !isPointerType(pass, binaryExpr.X) || !isPointerType(pass, binaryExpr.Y) {
		return
	}

	leftType := getUnderlyingType(pass, binaryExpr.X)
	rightType := getUnderlyingType(pass, binaryExpr.Y)

	if isBasicType(leftType) && isBasicType(rightType) {
		pass.Report(analysis.Diagnostic{
			Pos:     binaryExpr.Pos(),
			Message: fmt.Sprintf("comparing pointers to basic types: %v and %v", leftType, rightType),
		})
	}
}

func isPointerType(pass *analysis.Pass, expr ast.Expr) bool {
	exprType := pass.TypesInfo.TypeOf(expr)
	if exprType == nil {
		return false
	}

	_, isPtr := exprType.(*types.Pointer)
	return isPtr
}

func getUnderlyingType(pass *analysis.Pass, expr ast.Expr) types.Type {
	exprType := pass.TypesInfo.TypeOf(expr)
	if exprType == nil {
		return nil
	}

	if ptr, ok := exprType.(*types.Pointer); ok {
		return ptr.Elem()
	}

	return exprType
}

func isBasicType(t types.Type) bool {
	if t == nil {
		return false
	}
	_, isBasic := t.Underlying().(*types.Basic)
	return isBasic
}
