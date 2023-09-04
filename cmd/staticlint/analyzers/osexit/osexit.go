// Package osexit contains an Analyzer that find using os.Exit func in main package.
// Analuzer implements only to functin main in package main.
package osexit

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

const (
	pkgName  = "main"    // search package name
	funcName = "main"    // in function search name
	osExit   = "os.Exit" // search function name
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
			switch x := node.(type) {
			case *ast.CallExpr:
				fn, ok := typeutil.Callee(pass.TypesInfo, x).(*types.Func)
				if ok && fn.FullName() == osExit {
					pass.Reportf(x.Pos(), "using 'os.Exit' function in main package and main function detected")
				}
			case *ast.FuncDecl:
				if x.Name.Name != funcName {
					return false
				}
			}
			return true
		},
		)
	}
	return nil, nil
}
