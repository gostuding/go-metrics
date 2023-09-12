// Package osexit contains an Analyzer that find using os.Exit func in main package.
// Analuzer implements only to functin main in package main.
package osexit

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

const (
	pkgName      = "main"    // search package name
	funcName     = "main"    // in function search name
	osExit       = "os.Exit" // search function name
	errorMessage = "using 'os.Exit' function in main package and main function detected"
)

// OsExitAnalyzer is an analyzer.
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osexitanalyzer",
	Doc:  "analyze to privent using os.Exit in main package and main function",
	Run:  CheckOsExit,
}

// CheckOsExit checks the ast.Tree.
func CheckOsExit(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != pkgName {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			x, ok := node.(*ast.FuncDecl)
			if ok {
				if x.Name.Name == funcName {
					ast.Inspect(x, func(n ast.Node) bool {
						f, ok := node.(*ast.CallExpr)
						if ok {
							fun, ok := f.Fun.(*ast.SelectorExpr)
							if ok && fmt.Sprintf("%s.%s", fun.X, fun.Sel.Name) == osExit {
								pass.Reportf(f.Pos(), errorMessage)
							}
						}
						return true
					})
					return false
				}
			}
			return true
		},
		)
	}
	return nil, nil //nolint:all //<-senselessly
}
