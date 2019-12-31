package sortslice_test

import (
	"testing"

	"github.com/jackie-feng/tools/go/analysis/analysistest"
	"github.com/jackie-feng/tools/go/analysis/passes/sortslice"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, sortslice.Analyzer, "a")
}
