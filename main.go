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

	parseDir(dir)
}

func parseDir(dir string) ([]string, error) {
	// Configure the packages.Load to load the packages in the directory
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps,
		Dir:  dir,
	}

	// Load the packages in the directory
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return []string{}, fmt.Errorf("Failed to load packages: %v", err)
	}

	// Check for any packages with errors
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
		pass := &analysis.Pass{
			Analyzer:   ptrAnalyzer,
			Fset:       pkg.Fset,
			Files:      pkg.Syntax,
			OtherFiles: nil,
			Pkg:        pkg.Types,
			TypesInfo:  pkg.TypesInfo,
			TypesSizes: pkg.TypesSizes,
			ResultOf:   make(map[*analysis.Analyzer]interface{}),
			Report: func(d analysis.Diagnostic) {
				pos := pkg.Fset.Position(d.Pos)
				results = append(results, fmt.Sprintf("%s:%d:%d: %s\n", pos.Filename, pos.Line, pos.Column, d.Message))
			},
		}

		inspectPass := &analysis.Pass{
			Analyzer:   inspect.Analyzer,
			Fset:       pkg.Fset,
			Files:      pkg.Syntax,
			OtherFiles: nil,
			Pkg:        pkg.Types,
			TypesInfo:  pkg.TypesInfo,
			TypesSizes: pkg.TypesSizes,
			ResultOf:   make(map[*analysis.Analyzer]interface{}),
			Report:     func(d analysis.Diagnostic) {},
		}
		result, err := inspect.Analyzer.Run(inspectPass)
		if err != nil {
			log.Printf("Failed to run inspect analyzer on package %s: %v\n", pkg.Name, err)
			continue
		}
		pass.ResultOf[inspect.Analyzer] = result
		// Run our analyzer
		_, err = ptrAnalyzer.Run(pass)
		if err != nil {
			return []string{}, fmt.Errorf("Failed to run analyzer on package %s: %v\n", pkg.Name, err)
		}
	}
	return results, nil
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
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.BinaryExpr)(nil), // Add BinaryExpr to filter to inspect binary expressions
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		if n == nil {
			return
		}

		if binaryExpr, ok := n.(*ast.BinaryExpr); ok {
			switch binaryExpr.Op {
			case token.EQL, token.NEQ, token.LSS, token.GTR, token.LEQ, token.GEQ:
				if isPointerType(pass, binaryExpr.X) && isPointerType(pass, binaryExpr.Y) {
					leftType := getUnderlyingType(pass, binaryExpr.X)
					rightType := getUnderlyingType(pass, binaryExpr.Y)
					if isBasicType(leftType) && isBasicType(rightType) { // Fixed logic: we want to report when BOTH are basic types
						pass.Report(
							analysis.Diagnostic{
								Pos:     binaryExpr.Pos(),
								Message: fmt.Sprintf("comparing pointers to basic types: %v and %v", leftType, rightType),
							},
						)
					}
				}
			default:
			}
		}
	})
	return nil, nil
}

func isPointerType(pass *analysis.Pass, expr ast.Expr) bool {
	// Get the actual type from the type checker
	exprType := pass.TypesInfo.TypeOf(expr)
	if exprType == nil {
		return false
	}

	_, isPtr := exprType.(*types.Pointer)
	return isPtr
}

func getUnderlyingType(pass *analysis.Pass, expr ast.Expr) types.Type {
	// Get the type from the type checker
	exprType := pass.TypesInfo.TypeOf(expr)
	if exprType == nil {
		return nil
	}

	// If it's a pointer, get the element type
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
