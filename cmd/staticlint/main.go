// Package main contains static analyzers.
// To build executable file in all project folder use command:
//
//	go build -o staticlint cmd/staticlint/main.go
//
// After that use command:
//
//	staticlint ./...
package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/gostuding/go-metrics/cmd/staticlint/analyzers"
)

func main() {
	multichecker.Main(
		analyzers.GetAnalyzers()...,
	)
}
