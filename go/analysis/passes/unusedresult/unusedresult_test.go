// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unusedresult_test

import (
	"testing"

	"github.com/jackie-feng/tools/go/analysis/analysistest"
	"github.com/jackie-feng/tools/go/analysis/passes/unusedresult"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, unusedresult.Analyzer, "a")
}
