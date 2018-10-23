package expr

type IExpression interface {
	Eval(v uint64) bool
}

type andExpr []IExpression

func And(e1, e2 IExpression, eN ...IExpression) andExpr {
	if len(eN) == 0 {
		return andExpr{e1, e2}
	}

	return append(andExpr{e1, e2}, eN...)
}

func (e andExpr) Eval(v uint64) bool {
	for _, expr := range e {
		if !expr.Eval(v) {
			return false
		}
	}

	return true
}

type orExpr []IExpression

func Or(e1, e2 IExpression, eN ...IExpression) orExpr {
	if len(eN) == 0 {
		return orExpr{e1, e2}
	}

	return append(orExpr{e1, e2}, eN...)
}

func (e orExpr) Eval(v uint64) bool {
	for _, expr := range e {
		if expr.Eval(v) {
			return true
		}
	}

	return false
}

type notExpr struct {
	e IExpression
}

func Not(e IExpression) notExpr {
	return notExpr{e}
}

func (e notExpr) Eval(v uint64) bool {
	return !e.e.Eval(v)
}

type exprAny struct{}

func Any() exprAny {
	return exprAny{}
}

func (e exprAny) Eval(uint64) bool {
	return true
}

type exprEmpty struct{}

func Empty() exprEmpty {
	return exprEmpty{}
}

func (e exprEmpty) Eval(v uint64) bool {
	return v == 0
}
