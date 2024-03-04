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
	req     map[string]bool
	posn    map[string]token.Pos
	related map[string][]analysis.RelatedInformation
}

func (*dataFact) AFact() {}

func run(pass *analysis.Pass) (interface{}, error) {

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	fact := new(dataFact)
	fact.req = make(map[string]bool)
	fact.posn = make(map[string]token.Pos)
	fact.related = make(map[string][]analysis.RelatedInformation)

	inspect.Preorder(nodeFilter, func(node ast.Node) {

		funcDecl := node.(*ast.FuncDecl)
		funcName := funcDecl.Name.Name

		ast.Inspect(funcDecl, func(n ast.Node) bool {

			pkgFact := new(dataFact)
			pkgFact.req = make(map[string]bool)
			pkgFact.posn = make(map[string]token.Pos)

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
					fact.req[funcName] = true
					fact.posn[funcName] = funcDecl.Pos()
					if c.Pos() != 0 {
						fact.related[funcName] = append(fact.related[funcName],
							analysis.RelatedInformation{Pos: c.Pos(), Message: "first related call"})
					}
				} else if pass.ImportPackageFact(pass.TypesInfo.Uses[f.Sel].Pkg(), pkgFact) {
					fact.req[funcName] = pkgFact.req[f.Sel.Name]
					if pkgFact.req[f.Sel.Name] && pkgFact.posn[f.Sel.Name] != 0 {
						fact.related[funcName] = append(pkgFact.related[f.Sel.Name],
							analysis.RelatedInformation{Pos: pkgFact.posn[f.Sel.Name], Message: "related call"})
					}
				}
			}
			if fact.req[funcName] {
				pass.Report(analysis.Diagnostic{
					Pos:     funcDecl.Pos(),
					Message: "call NewRequest",
					Related: fact.related[funcName],
				})
			}
			return true
		})
	})

	pass.ExportPackageFact(fact)

	return nil, nil
}
