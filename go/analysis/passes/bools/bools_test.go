// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bools_test

import (
	"testing"

	"github.com/jackie-feng/tools/go/analysis/analysistest"
	"github.com/jackie-feng/tools/go/analysis/passes/bools"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, bools.Analyzer, "a")
}
