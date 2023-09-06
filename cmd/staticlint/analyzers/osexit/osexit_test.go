package osexit

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestCheckOsExit(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), OsExitAnalyzer, "./...")
}
