// The unmarshal command runs the unmarshal analyzer.
package main

import (
	"github.com/jackie-feng/tools/go/analysis/passes/unmarshal"
	"github.com/jackie-feng/tools/go/analysis/singlechecker"
)

func main() { singlechecker.Main(unmarshal.Analyzer) }
