// The shadow command runs the shadow analyzer.
package main

import (
	"github.com/jackie-feng/tools/go/analysis/passes/shadow"
	"github.com/jackie-feng/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(shadow.Analyzer) }
