package jack_compiler

import (
	"fmt"
	"io"
)

type jackClass struct {
	name string
	declarations []variable
	subroutines []subroutine
}

func (c *jackClass) asText() string {
	return c.name
}


type classWriter struct {
}

//parse tokenized results to structured class
func (w *classWriter) Write(writer io.Writer, ts ...Token) {
	if len(ts) == 0 {
		fmt.Println("no input token")
		return
	}
	jc, err := parseClass(ts[0])
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	writer.Write([]byte(jc.asText()))
}

var emptyClass = jackClass{}

//convert lexial tree to structed class
func parseClass(ts Token) (jackClass, error) {
	root := ts.(*NonTerminalToken)
	name, err := resolveClassName(root)
	if err != nil {
		return emptyClass, nil
	}

	jc := jackClass{
		name: name,
	}

	declarations, err := resolveClassVarDecs(root)
	if err != nil {
		return emptyClass, err
	}
	jc.declarations = append(jc.declarations, declarations...)

	subroutines, err := resolveSubroutines(root)
	if err != nil {
		return emptyClass, err
	}
	jc.subroutines = append(jc.subroutines, subroutines...)

	return jc, nil
}

func resolveSubroutines(root *NonTerminalToken) ([]subroutine, error) {
	var subs []subroutine
	for _, item := range root.subTokens {
		if item.GetType() == SubroutineDec {
			sdc, err := resolveSubroutine(item)
			if err != nil {
				return nil, err
			}
			subs = append(subs, sdc)
		}
	}
	return subs, nil
}

var emptySubroutine = subroutine{}

/*
 input:
<subroutineDec>
        <keyword>function</keyword>
        <keyword>void</keyword>
        <identifier>foo</identifier>
        <symbol>(</symbol>
        <parameterList>
			...
        </parameterList>
        <symbol>)</symbol>
        <subroutineBody>
			...
		</subroutineBody>
</subroutineDec>
*/
func resolveSubroutine(token Token) (subroutine, error) {
	sub := subroutine{}
	sts := token.SubTokens()
	sub.category = subroutineCategory(sts[0].GetVal())
	sub.retType = sts[1].GetVal()
	sub.name = sts[2].GetVal()
	params := match(token, ParameterList)
	if len(params) == 0 {
		return emptySubroutine, newSyntaxError(token)
	}
	sub.declarations = append(sub.declarations, resolveParams(params[0])...)

	body := match(token, SubroutineBody)[0]

	for _, st := range match(body, VarDec) {
		sub.declarations = append(sub.declarations, resolveVarDec(st))
	}

	for _, st := range match(body, Statements) {
		statements, err := resolveStatements(st)
		if err != nil {
			return emptySubroutine, nil
		}
		sub.statements = append(sub.statements, statements...)
	}

	return sub, nil
}

func resolveVarDec(st Token) variable {
	return variable{
		typ: vType(st.SubTokens()[1].GetVal()),
		name: st.SubTokens()[2].GetVal(),
		kind: klocal,
	}
}

func resolveStatements(st Token) ([]Statement, error) {
	if err := assertToken(st, Statements, "");err != nil {
		return nil, err
	}
	var sts []Statement
	for _, s := range st.SubTokens() {
		statement, err := resolveStatement(s)
		if err != nil {
			return nil, err
		}
		sts = append(sts, statement)
	}
	return sts, nil
}

func resolveStatement(st Token) (Statement, error) {
	switch st.GetType() {
	case IfStatement:
		return resolveIfStatement(st)
	case LetStatement:
		return resolveLetStatement(st)
	case DoStatement:
		return resolveDoStatement(st)
	case ReturnStatement:
		return resolveReturnStatement(st)
	default:
		return nil, newSyntaxError(st)
	}
}

func resolveReturnStatement(st Token) (Statement, error) {
	if err := assertToken(st, ReturnStatement, "");err != nil {
		return nil, err
	}
	stat := retStatement{}
	it := NewTokenIterator(st.SubTokens())
	it.Next() //pop 'return'
	next := it.Next()

	if next.GetType() == Expression {
		exp, err := resolveExpression(next)
		if err != nil {
			return nil, err
		}
		stat.expression = exp
	}
	return stat, nil
}

func resolveDoStatement(st Token) (Statement, error) {
	if err := assertToken(st, DoStatement, "");err != nil {
		return nil, err
	}
	it := NewTokenIterator(st.SubTokens())
	if err := assertToken(it.Next(), Keyword, "do");err != nil {
		return nil, err
	}

	target := it.Next()
	if err := assertToken(target, Identifier, "");err != nil {
		return nil, err
	}

	var name string
	next := it.Peek()
	if next.GetType() == Symbol && next.GetVal() == "." {
		it.Next() //pop '.'
		callee := it.Next()
		if err := assertToken(callee, Identifier, "");err != nil {
			return nil, err
		}
		name = callee.GetVal()
	}
	it.Next()//pop (
	params := it.Next()//expression list
	if err := assertToken(params, ExpressionList, "");err != nil {
		return nil, err
	}
	subcall, err := resolveSubcall(st, target.GetVal(), name)
	if err != nil {
		return nil, err
	}

	return doStatement{
		action: subcall,
	},nil


}

func resolveLetStatement(st Token) (Statement, error) {
	if err := assertToken(st, LetStatement, "");err != nil {
		return nil, err
	}
	stat := letStatement{}
	it := NewTokenIterator(st.SubTokens())
	it.Next()//let
	varName := it.Next().GetVal()
	if it.Peek().GetVal() == "[" {//ref to array
		it.Next()//pop [
		exp,err := resolveExpression(it.Next())
		if err != nil {
			return nil, err
		}
		stat.target = ReferenceTerm{varName: varName, index: exp}
	} else {
		stat.target = ReferenceTerm{varName: varName}
	}
	for it.HasNext() {
		next := it.Next()
		if next.GetVal() == "=" &&  next.GetType() == Symbol {
			exp, err := resolveExpression(it.Next())
			if err != nil {
				return nil, err
			}
			stat.expression = exp
			break
		}
	}
	return stat, nil
}

func resolveIfStatement(st Token) (Statement, error) {
	is := ifStatement{}

	it := NewTokenIterator(st.SubTokens())
	for it.HasNext() {
		sub := it.Next()
		if sub.GetType() == Keyword {
			if sub.GetVal() == "if" {
				for it.HasNext() {
					if it.Peek().GetType() == Expression {
						con, err := resolveExpression(it.Next())
						if err != nil {
							return nil, err
						}
						is.condition = con
						continue
					} else if it.Peek().GetType() == Statements {
						statements, err := resolveStatements(it.Next())
						if err != nil {
							return nil, err
						}
						is.statements = statements
						break
					}
					it.Next()
				}
			} else if sub.GetVal() == "else" {
				it.Next() //pop {
				elseStatements, err := resolveStatements(it.Next())
				if err != nil {
					return nil, err
				}
				is.elseStatements = elseStatements
				break
			}
		}
	}
	return is, nil
}

var emptyExpression = expression{}

func resolveExpression(token Token) (expression, error) {
	exp := expression{}
	if err := assertToken(token, Expression, ""); err != nil {
		return emptyExpression, err
	}
	it := NewTokenIterator(token.SubTokens())

	term, err := resolveTerm(it.Next())
	if err != nil {
		return emptyExpression, err
	}
	exp.terms = append(exp.terms, term)
	for it.HasNext() {
		st := it.Next()
		if st.GetType() == Symbol && operations[st.GetVal()] != "" {
			exp.operations = append(exp.operations, st.GetVal())
			term, err := resolveTerm(it.Next())
			if err != nil {
				return emptyExpression, err
			}
			exp.terms = append(exp.terms, term)
		}
	}
	return exp, nil
}

var emptyTerm = ConstTerm{}

func resolveTerm(token Token) (term Term, err error) {
	it := NewTokenIterator(token.SubTokens())
	for it.HasNext() {
		t := it.Next()
		switch t.GetType() {
		case IntegerConstant:
			return ConstTerm{ttype: IntegerConstant, val: t.GetVal()}, nil
		case StringConstant:
			return ConstTerm{ttype: StringConstant, val: t.GetVal()}, nil
		case Keyword:
			if ContainsString(keywordConstants, t.GetVal()) {
				return ConstTerm{ttype: Keyword, val: t.GetVal()}, nil
			}
		case Symbol:
			op := t.GetVal()
			if op == "-" || op == "~" {
				next := it.Next()
				if err = assertToken(next, TokenTerm, "");err != nil {
					return nil, err
				}
				term, err = resolveTerm(next)
				if err == nil {
					term = UnaryTerm{operator: op, term: term}
				}
			} else if op == "(" {
				expression, err := resolveExpression(it.Next())
				if err != nil {
					return emptyExpression, err
				}
				return expression, nil
			}
		case Identifier:
			target := t.GetVal()
			if !it.HasNext() {
				return ReferenceTerm{varName: target}, nil
			}
			next := it.Next()
			if next.GetType() == Symbol {
				if next.GetVal() == "." {
					subcallee := it.Next()
					if err := assertToken(subcallee, Identifier, "");err != nil {
						return emptySubCall, err
					}
					subcall, err := resolveSubcall(token, target, subcallee.GetVal())
					if err != nil {
						return emptyTerm, err
					}
					return subcall, nil
				} else if next.GetVal() == "(" {
					subcall, err := resolveSubcall(token, "", target)
					if err != nil {
						return emptyTerm, err
					}
					return subcall, nil
				} else if next.GetVal() == "[" {
					exp, err := resolveExpression(it.Next())
					if err != nil {
						return emptyTerm, err
					}
					return ReferenceTerm{varName:target, index: exp}, nil
				}
			}
		}
	}
	return nil, newSyntaxError(token)
}

var emptySubCall = subroutineCall{}

func resolveSubcall(parent Token, target string, name string) (subroutineCall, error) {
	args, err := resolveExpressionList(match(parent, ExpressionList)[0])
	if err != nil {
		return emptySubCall, err
	}
	return subroutineCall{
		target: target,
		name: name,
		args: args,
	}, nil
}


func resolveExpressionList(ts Token) ([]expression, error) {
	es := match(ts, Expression)
	var exps []expression
	for _, e := range es {
		exp, err := resolveExpression(e)
		if err != nil {
			return nil, err
		}
		exps = append(exps, exp)
	}
	return exps, nil
}

/*
	input:
		<parameterList>
            <identifier>string</identifier>
            <identifier>x</identifier>
            <symbol>,</symbol>
            <keyword>int</keyword>
            <identifier>y</identifier>
        </parameterList>
*/
func resolveParams(params Token) []variable {
	var vs []variable
	it := NewTokenIterator(params.SubTokens())
	for it.HasNext() {
		if it.Peek().GetVal() == "," {
			it.Next()
		}
		vs = append(vs,  variable{
			typ: vType(it.Next().GetVal()),
			name: it.Next().GetVal(),
			kind: kargument,
		})
	}
	return vs
}

func resolveClassVarDecs(root *NonTerminalToken) ([]variable, error) {
	var vars []variable
	for _, item := range root.subTokens {
		if item.GetType() == ClassVarDec {
			vars = append(vars, resolveClassVarDec(item))
		}
	}
	return vars, nil
}

func resolveClassVarDec(item Token) variable {
	kind := item.SubTokens()[0]
	typ := item.SubTokens()[1]
	name := item.SubTokens()[2]
	return variable{
		name: name.GetVal(),
		kind: vKind(kind.GetVal()),
		typ: vType(typ.GetVal()),
	}
}

func resolveClassName(root *NonTerminalToken) (string, error) {
	if err := assertToken(root, Class, "");err != nil {
		return "", err
	}
	subs := root.subTokens
	if err := assertToken(subs[0], Keyword, "class"); err != nil {
		return "", err
	}
	if err := assertToken(subs[1], Identifier, ""); err != nil {
		return "", err
	}
	return subs[1].GetVal(), nil
}