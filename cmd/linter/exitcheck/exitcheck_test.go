package exitcheck_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"Ustasjs/yp-url-shortener/cmd/linter/exitcheck"
)

func TestExitCheck(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), exitcheck.Analyzer, "withexit", "clean", "notmain")
}
