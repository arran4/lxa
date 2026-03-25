package filter

import (
	"github.com/lxa-project/lxa/internal/xattr"
)

// Evaluator evaluates metadata against an AST node.
type Evaluator struct {
	ast Node
}

// NewEvaluator creates a new filter evaluator from an expression.
func NewEvaluator(expr string) (*Evaluator, error) {
	node, err := Parse(expr)
	if err != nil {
		return nil, err
	}
	return &Evaluator{ast: node}, nil
}

// Eval returns true if the metadata matches the filter.
func (e *Evaluator) Eval(md xattr.Metadata) bool {
	if e.ast == nil {
		return true // Empty filter matches everything
	}
	return e.ast.Eval(md)
}
