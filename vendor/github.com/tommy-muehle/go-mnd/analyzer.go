package magic_numbers

import (
	"flag"
	"go/ast"

	"github.com/tommy-muehle/go-mnd/checks"
	"github.com/tommy-muehle/go-mnd/config"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `magic number detector`

var Analyzer = &analysis.Analyzer{
	Name:             "mnd",
	Doc:              Doc,
	Run:              run,
	Flags:            options(),
	Requires:         []*analysis.Analyzer{inspect.Analyzer},
	RunDespiteErrors: true,
}

type Checker interface {
	NodeFilter() []ast.Node
	Check(n ast.Node)
}

func options() flag.FlagSet {
	options := flag.NewFlagSet("", flag.ExitOnError)
	options.String("ignored-numbers", "", "comma separated list of numbers excluded from analysis")
	options.String(
		"checks",
		checks.ArgumentCheck+","+
			checks.CaseCheck+","+
			checks.ConditionCheck+","+
			checks.OperationCheck+","+
			checks.ReturnCheck+","+
			checks.AssignCheck,
		"comma separated list of checks",
	)

	return *options
}

func run(pass *analysis.Pass) (interface{}, error) {
	conf := config.WithOptions(
		config.WithCustomChecks(pass.Analyzer.Flags.Lookup("checks").Value.String()),
		config.WithIgnoredNumbers(pass.Analyzer.Flags.Lookup("ignored-numbers").Value.String()),
	)

	var checker []Checker
	if conf.IsCheckEnabled(checks.ArgumentCheck) {
		checker = append(checker, checks.NewArgumentAnalyzer(pass, conf))
	}
	if conf.IsCheckEnabled(checks.CaseCheck) {
		checker = append(checker, checks.NewCaseAnalyzer(pass, conf))
	}
	if conf.IsCheckEnabled(checks.ConditionCheck) {
		checker = append(checker, checks.NewConditionAnalyzer(pass, conf))
	}
	if conf.IsCheckEnabled(checks.OperationCheck) {
		checker = append(checker, checks.NewOperationAnalyzer(pass, conf))
	}
	if conf.IsCheckEnabled(checks.ReturnCheck) {
		checker = append(checker, checks.NewReturnAnalyzer(pass, conf))
	}
	if conf.IsCheckEnabled(checks.AssignCheck) {
		checker = append(checker, checks.NewAssignAnalyzer(pass, conf))
	}

	i := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	for _, c := range checker {
		i.Preorder(c.NodeFilter(), func(node ast.Node) {
			c.Check(node)
		})
	}

	return nil, nil
}