package main

import (
	"github.com/didebughu/ssca/analysis/ssca"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(ssca.Analyzer)
}
