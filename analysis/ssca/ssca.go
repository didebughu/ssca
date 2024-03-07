package ssca

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `analyze call link.`

var Analyzer = &analysis.Analyzer{
	Name:      "ssca",
	Doc:       Doc,
	URL:       "",
	Run:       run,
	FactTypes: []analysis.Fact{new(dataFact)},
	Requires:  []*analysis.Analyzer{inspect.Analyzer},
}

type dataFact struct {
	req     bool
	related token.Pos
}

func (*dataFact) AFact() {}

func run(pass *analysis.Pass) (interface{}, error) {

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspect.Preorder(nodeFilter, func(node ast.Node) {

		funcDecl := node.(*ast.FuncDecl)

		if funcDecl.Name.Obj == nil {
			return
		}

		fact := new(dataFact)
		fact.req = false
		fact.related = 0

		ast.Inspect(funcDecl, func(n ast.Node) bool {

			c, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			switch f := c.Fun.(type) {
			case *ast.SelectorExpr:
				if pass.TypesInfo == nil {
					return true
				}
				if pass.TypesInfo.Uses[f.Sel].Pkg() == nil {
					return true
				}
				if f.Sel.Name == "Get" && pass.TypesInfo.Uses[f.Sel].Pkg().Path() == "net/http" {
					fact.req = true
					fact.related = c.Pos()
					pass.ExportObjectFact(pass.TypesInfo.ObjectOf(funcDecl.Name), fact)
					return false
				}
				pkgFact := new(dataFact)
				if pass.ImportObjectFact(pass.TypesInfo.ObjectOf(f.Sel), pkgFact) && pkgFact.req {
					fact.req = true
					fact.related = c.Pos()
					pass.ExportObjectFact(pass.TypesInfo.ObjectOf(funcDecl.Name), fact)
					return false
				}
			}
			return true
		})
		if fact.req {
			pass.Report(analysis.Diagnostic{
				Pos:     funcDecl.Pos(),
				Message: "call NewRequest",
				Related: []analysis.RelatedInformation{{
					Pos:     fact.related,
					Message: "related call",
				}},
			})
		}
	})

	return nil, nil
}
