// Package analyzers используется для подготовки списка статических анализаторов.
// Для получения списка используемых анализаторов необходимо вызвать функию
//
//	GetAnalyzers()
//
// В качестве анализаторов используются:
//
// Printf - checks consistency of Printf format strings and arguments.
// Shadow - Analyzer that checks for shadowed variables.
// Shift - checks for shifts that exceed the width of an integer.
// Structtag - an Analyzer that checks struct field tags are well formed.
// Staticcheck - analyzes that find bugs and performance issues (SAxxxx).
// Quickfix - analyzes that implement code refactorings (QF1xxx).
// Simple - analyzes that simplify code (S1xxx).
// Stylecheck - analyzes  that enforce style rules (STxxxx).
// Unused - contains code for finding unused code (U1000).
// Osexit - privent using os.Exit function in main package.
package analyzers

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"

	"github.com/gostuding/go-metrics/cmd/staticlint/analyzers/osexit"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
	"honnef.co/go/tools/unused"
)

// GetAnalyzers creates []*analysis.Analyzer for multichecker.
func GetAnalyzers() []*analysis.Analyzer {
	count := 6 + len(staticcheck.Analyzers) + len(simple.Analyzers) +
		len(quickfix.Analyzers) + len(stylecheck.Analyzers)
	checks := make([]*analysis.Analyzer, 0, count)
	checks = append(checks, osexit.OsExitAnalyzer, printf.Analyzer, shadow.Analyzer,
		structtag.Analyzer, shift.Analyzer, unused.Analyzer.Analyzer)
	for _, v := range staticcheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}
	for _, v := range simple.Analyzers {
		checks = append(checks, v.Analyzer)
	}
	for _, v := range quickfix.Analyzers {
		checks = append(checks, v.Analyzer)
	}
	for _, v := range stylecheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}
	return checks
}
