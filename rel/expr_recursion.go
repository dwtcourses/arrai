package rel

import (
	"fmt"

	"github.com/arr-ai/wbnf/parser"
	"github.com/go-errors/errors"
)

type RecursionExpr struct {
	ExprScanner
	name      Pattern
	fn        Expr
	fix, fixt Value
}

func NewRecursionExpr(scanner parser.Scanner, name Pattern, fn Expr, fix, fixt Value) Expr {
	return RecursionExpr{ExprScanner{scanner}, name, fn, fix, fixt}
}

func (r RecursionExpr) Eval(local Scope) (Value, error) {
	name, isIdent := r.name.(ExprPattern).expr.(IdentExpr)
	if !isIdent {
		return nil, wrapContext(errors.Errorf("Does not evaluate to a variable name: %v", r.name), r, local)
	}

	val, err := r.fn.Eval(local)
	if err != nil {
		return nil, wrapContext(err, r, local)
	}

	argName := NewIdentExpr(name.Source(), name.String())

	//TODO: optimise, get it to load either fix or fixt not both
	switch f := val.(type) {
	case Tuple:
		t := f
		for e := f.Enumerator(); e.MoveNext(); {
			attr, val := e.Current()
			if fn, isFunction := val.(Closure); isFunction {
				f = f.With(attr, NewClosure(local, NewFunction(fn.Source(), argName, fn.f).(*Function)))
				continue
			}
			return nil, wrapContext(errors.Errorf("Recursion requires a tuple of functions: %v", t.String()), r, local)
		}
		return Call(r.fixt, f, local)
	case Closure:
		return Call(r.fix, NewClosure(local, NewFunction(f.Source(), argName, f.f).(*Function)), local)
	}
	return nil, wrapContext(errors.Errorf("Recursion does not support %T", val), r, local)
}

func (r RecursionExpr) String() string {
	return fmt.Sprintf("\\%s %s", r.name, r.fn)
}