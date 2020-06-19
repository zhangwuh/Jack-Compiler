package compiler

type subroutineCategory string

const (
	constructor subroutineCategory = "constructor"
	method      subroutineCategory = "method"
	function    subroutineCategory = "function"
)

type subroutine struct {
	name         string
	category     subroutineCategory
	declarations []variable
	statements   []Statement
	retType      string
}

type subroutineCall struct {
	target string
	name   string
	args   []expression
}

func (sc subroutineCall) category() termCategory {
	return subCallTerm
}
