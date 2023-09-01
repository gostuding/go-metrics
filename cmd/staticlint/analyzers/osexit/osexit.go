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
	pkgName = "main"    // search package name
	osExit  = "os.Exit" // search function name
)

// OsExitAnalyzer is an analyzer.
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osexitanalyzer",
	Doc:  "analyze to privent using os.Exit in main package and main function",
	Run:  CheckOsExit,
}

// CheckOsExit checks the ast.Tree.
func CheckOsExit(pass *analysis.Pass) (interface{}, error) {
	fMain := false
	for _, file := range pass.Files {
		if file.Name.Name == pkgName {
			fMain = false
			ast.Inspect(file, func(node ast.Node) bool {
				switch x := node.(type) {
				case *ast.CallExpr:
					if fMain {
						fn, _ := typeutil.Callee(pass.TypesInfo, x).(*types.Func)
						if fn != nil && fn.FullName() == osExit {
							pass.Reportf(x.Pos(), "using 'os.Exit' function in main package detected")
						}
					}
				case *ast.FuncDecl:
					if x.Name.Name == "main" {
						fMain = true
					} else {
						fMain = false
					}
				}
				return true
			},
			)

		}
	}
	return nil, nil
}
