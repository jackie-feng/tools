// The findcall command runs the findcall analyzer.
package main

import (
	"github.com/jackie-feng/tools/go/analysis/passes/findcall"
	"github.com/jackie-feng/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(findcall.Analyzer) }
