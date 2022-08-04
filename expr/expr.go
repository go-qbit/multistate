package expr

type Expression interface {
	Eval(v uint64) bool
}

type andExpr []Expression

func And(e1, e2 Expression, eN ...Expression) andExpr {
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

type orExpr []Expression

func Or(e1, e2 Expression, eN ...Expression) orExpr {
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

type xorExpr []Expression

func Xor(e1, e2 Expression, eN ...Expression) xorExpr {
	if len(eN) == 0 {
		return xorExpr{e1, e2}
	}
	return append(xorExpr{e1, e2}, eN...)
}

func (e xorExpr) Eval(v uint64) bool {
	var c int
	for _, expr := range e {
		if expr.Eval(v) {
			c++
		}
		if c > 1 {
			return false
		}
	}
	return c == 1
}

type notExpr struct {
	e Expression
}

func Not(e Expression) notExpr {
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
