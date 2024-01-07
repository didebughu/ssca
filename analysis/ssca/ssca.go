package ssca

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

const Doc = `check var.`

var Analyzer = &analysis.Analyzer{
	Name:     "ssca",
	Doc:      Doc,
	URL:      "",
	Run:      run,
	Requires: []*analysis.Analyzer{buildssa.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {

	ssainput := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)

	for _, fn := range ssainput.SrcFuncs {

		for _, b := range fn.Blocks {
			for _, instr := range b.Instrs {
				switch instr.(type) {
				case *ssa.Store:
				case *ssa.MapUpdate:
				}

			}
		}
	}

	return nil, nil
}
