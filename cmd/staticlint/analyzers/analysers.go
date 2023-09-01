// Package analyzers используется для подготовки списка статических анализаторов.
// Для получения списка используемых анализаторов необходимо вызвать функию
//
//	 GetAnalyzers()
//
// В качестве анализаторов используются:
//
// golang.org/x/tools/go/analysis/passes/printf - an Analyzer that checks consistency of Printf format strings and arguments.
// golang.org/x/tools/go/analysis/passes/shadow - Analyzer that checks for shadowed variables.
// golang.org/x/tools/go/analysis/passes/shift - an Analyzer that checks for shifts that exceed the width of an integer.
// golang.org/x/tools/go/analysis/passes/structtag - an Analyzer that checks struct field tags are well formed.
// honnef.co/go/tools/staticcheck - analyzes that find bugs and performance issues (SAxxxx).
// honnef.co/go/tools/quickfix - analyzes that implement code refactorings (QF1xxx).
// honnef.co/go/tools/simple - analyzes that simplify code (S1xxx).
// honnef.co/go/tools/stylecheck - analyzes  that enforce style rules (STxxxx).
// honnef.co/go/tools/unused - contains code for finding unused code (U1000).
// github.com/gostuding/go-metrics/cmd/staticlint/analyzers/osexit - an Analyzer for privent using os.Exit function in main package
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
	cap := 6 + len(staticcheck.Analyzers) + len(simple.Analyzers) + len(quickfix.Analyzers) + len(stylecheck.Analyzers)
	checks := make([]*analysis.Analyzer, 0, cap)
	checks = append(checks, osexit.OsExitAnalyzer)
	checks = append(checks, printf.Analyzer)
	checks = append(checks, shadow.Analyzer)
	checks = append(checks, structtag.Analyzer)
	checks = append(checks, shift.Analyzer)
	checks = append(checks, unused.Analyzer.Analyzer)
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
