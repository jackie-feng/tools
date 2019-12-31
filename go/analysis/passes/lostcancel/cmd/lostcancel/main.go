// The lostcancel command applies the github.com/jackie-feng/tools/go/analysis/passes/lostcancel
// analysis to the specified packages of Go source code.
package main

import (
	"github.com/jackie-feng/tools/go/analysis/passes/lostcancel"
	"github.com/jackie-feng/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(lostcancel.Analyzer) }
