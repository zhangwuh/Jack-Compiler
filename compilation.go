package jack_compiler

import (
	"fmt"
	"io"
)

func newCompileError(t Token, msg string) error {
	return fmt.Errorf("%s, line:%d", msg, t.Position())
}

func assertToken(t Token, typ TokenType, val string) error {
	if typ != t.GetType() || (len(val) > 0 && t.GetVal() != val) {
		return newCompileError(t, fmt.Sprintf("compile error, encountered:%s, expected:%s", t.GetVal(), val))
	}
	return nil
}

type nonTerminalTokenWriter struct{}

func (ntw *nonTerminalTokenWriter) Write(writer io.Writer, ts ...Token) {
	for _, ts := range ts {
		writer.Write([]byte(ts.AsText()))
	}
}

type compiler struct {
}

type blockEndChecker func(it *TokenIterator) bool

var statementEndChecker = func(it *TokenIterator) bool {
	return it.Peek().GetVal() == ";"
}

var parameterListEndChecker = func(it *TokenIterator) bool {
	return it.Peek().GetVal() == ")"
}

var bodyEndChecker = func(it *TokenIterator) bool {
	return it.Peek().GetVal() == "}"
}

func (cp *compiler) Compile(tokens []Token) (*NonTerminalToken, error) {
	it := &TokenIterator{tokens: tokens}
	if it.Size() == 0 {
		return nil, nil
	}
	return compileClass(it)
}

func compileClass(it *TokenIterator) (nt *NonTerminalToken, err error) {
	nt = &NonTerminalToken{
		tokenType: Class,
	}
	token := it.Next()
	if err = assertToken(token, Keyword, "class"); err != nil {
		return
	}
	nt.AddSubToken(token)

	token = it.Next()
	if err = assertToken(token, Identifier, ""); err != nil {
		return
	}
	nt.AddSubToken(token)
	token = it.Next()
	if err = assertToken(token, Symbol, "{"); err != nil {
		return
	}
	nt.AddSubToken(token)
	for it.HasNext() {
		token = it.Peek()
		switch token.GetVal() {
		case "field", "static":
			nt.AddSubToken(compileClassVarDec(it))
			continue
		case "constructor", "function", "method":
			sub, se := compileSubRoutineDec(it)
			if se != nil {
				return nil, se
			}
			nt.AddSubToken(sub)
			continue
		default:
			break
		}
		if err = assertToken(token, Symbol, "}"); err != nil {
			return
		}
		nt.AddSubToken(it.Next())
		break
	}
	return
}

func compileSubRoutineDec(it *TokenIterator) (*NonTerminalToken, error) {
	nt := &NonTerminalToken{
		tokenType: SubroutineDec,
	}

	nt.AddSubToken(it.Next()) //constructor, method, function
	nt.AddSubToken(it.Next()) //class name in constructor, void, return type
	nt.AddSubToken(it.Next()) //func,method name

	ts, _ := withParentheses(it, compileParameters)
	nt.AddSubToken(ts...)

	st, err := compileSubRoutineBody(it)
	if err != nil {
		return nil, err
	}
	nt.AddSubToken(st)
	return nt, nil
}

type symbolPair struct {
	left, right string
}

var parentheses = symbolPair{"(", ")"}
var curlyBrackets = symbolPair{"{", "}"}

func withSymbolBlock(it *TokenIterator, pair symbolPair, builder func(it *TokenIterator) (*NonTerminalToken, error)) (ts []Token, err error) {
	token := it.Next()
	if e := assertToken(token, Symbol, pair.left); e != nil {
		return nil, e
	}
	ts = append(ts, token)

	st, e := builder(it)
	if e != nil {
		return nil, e
	}
	ts = append(ts, st)

	token = it.Next() //}
	if e := assertToken(token, Symbol, pair.right); e != nil {
		return nil, e
	}
	ts = append(ts, token)
	return ts, nil
}

func withParentheses(it *TokenIterator, builder func(it *TokenIterator) (*NonTerminalToken, error)) (ts []Token, err error) {
	return withSymbolBlock(it, parentheses, builder)
}

func withCurlyBrackets(it *TokenIterator, builder func(it *TokenIterator) (*NonTerminalToken, error)) (ts []Token, err error) {
	return withSymbolBlock(it, curlyBrackets, builder)
}

func compileExpressionList(it *TokenIterator) (nt *NonTerminalToken, err error) {
	nt = &NonTerminalToken{
		tokenType: ExpressionList,
	}
	for it.HasNext() {
		if parameterListEndChecker(it) {
			return
		}

		exp, err := compileExpression(it)
		if err != nil {
			return nil, err
		}
		nt.AddSubToken(exp)
		next := it.Peek()
		if next.GetVal() == "," {
			nt.AddSubToken(it.Next())
		}
	}
	return
}

func compileExpression(it *TokenIterator) (nt *NonTerminalToken, err error) {
	nt = &NonTerminalToken{
		tokenType: Expression,
	}
	leftTerm, e := compileTerm(it)
	if e != nil {
		return nil, e
	}
	nt.AddSubToken(leftTerm)

	token := it.Peek()
	if operations[token.GetVal()] == "" {
		return
	}

	nt.AddSubToken(it.Next()) //operation symbol

	rightTerm, e := compileTerm(it)
	if e != nil {
		return nil, e
	}
	nt.AddSubToken(rightTerm)
	return
}

func compileTerm(it *TokenIterator) (*NonTerminalToken, error) {
	term := &NonTerminalToken{
		tokenType: Term,
	}
	next := it.Peek()
	if next.GetVal() == "(" { //expression
		st, err := withParentheses(it, compileExpression)
		if err != nil {
			return nil, err
		}
		term.AddSubToken(st...)

	} else if next.GetType() == Identifier { // x , x.a()
		token, err := compileIdentifier(it)
		if err != nil {
			return nil, err
		}
		term.AddSubToken(token...)
	} else if next.GetType() == Symbol { //-1, -i
		term.AddSubToken(it.Next())
		st, err := compileTerm(it)
		if err != nil {
			return nil, err
		}
		term.AddSubToken(st)
	} else if next.GetType() == IntegerConstant || next.GetType() == StringConstant || isKeywordConstant(next) {
		term.AddSubToken(it.Next())
	} else {
		return nil, newCompileError(next, fmt.Sprintf("invalid grammar error in if statement:%s", next.AsText()))
	}
	return term, nil
}

func compileSubRoutineBody(it *TokenIterator) (Token, error) {
	nt := &NonTerminalToken{
		tokenType: SubroutineBody,
	}

	token := it.Next()
	if e := assertToken(token, Symbol, "{"); e != nil {
		return nil, e
	}
	nt.AddSubToken(token)

	for it.HasNext() {
		t := it.Peek()
		if t.GetVal() == "}" {
			nt.AddSubToken(it.Next())
			return nt, nil
		}
		tt := typeOf(t)
		if tt == IfStatement || tt == LetStatement || tt == WhileStatement || tt == DoStatement || tt == ReturnStatement {
			st, err := compileStatements(it)
			if err != nil {
				return nil, err
			}
			nt.AddSubToken(st)
		} else if tt == VarStatement {
			ts, err := compileVarDec(it)
			if err != nil {
				return nil, err
			}
			nt.AddSubToken(ts)
		} else {
			return nil, newCompileError(it.Peek(), fmt.Sprintf("grammar error, unknow token:%s", t.AsText()))
		}
	}

	token = it.Next() //}
	if e := assertToken(token, Symbol, "}"); e != nil {
		return nil, e
	}
	nt.AddSubToken(token)
	return nt, nil
}

func compileStatements(it *TokenIterator) (nt *NonTerminalToken, err error) {
	nt = &NonTerminalToken{
		tokenType: Statements,
	}

	for it.HasNext() {
		token := it.Peek()
		switch typeOf(token) {
		case IfStatement:
			ts, err := compileIfStatement(it)
			if err != nil {
				return nil, err
			}
			nt.AddSubToken(ts)
			continue
		case WhileStatement:
			ts, err := compileWhileStatement(it)
			if err != nil {
				return nil, err
			}
			nt.AddSubToken(ts)
			continue
		case LetStatement:
			ts, err := compileLetStatement(it)
			if err != nil {
				return nil, err
			}
			nt.AddSubToken(ts)
			continue
		case DoStatement:
			ts, err := compileDoStatement(it)
			if err != nil {
				return nil, err
			}
			nt.AddSubToken(ts)
			continue
		case ReturnStatement:
			ts, err := compileReturnStatement(it)
			if err != nil {
				return nil, err
			}
			nt.AddSubToken(ts)
			continue
		}
		break
	}

	return
}

/* handle variable ref like: a, a[0], a[0+1]*/
func compileIdentifier(it *TokenIterator) (ts []Token, err error) {
	t := it.Next()
	if err := assertToken(t, Identifier, ""); err != nil {
		return nil, err
	}
	ts = append(ts, t)

	t = it.Peek()
	if t.GetType() == Symbol && t.GetVal() == "[" {
		ts = append(ts, it.Next())

		exp, err := compileExpression(it)
		if err != nil {
			return nil, err
		}
		ts = append(ts, exp)

		t = it.Next()
		if err := assertToken(t, Symbol, "]"); err != nil {
			return nil, err
		}
		ts = append(ts, t)
	} else {
		if t.GetType() == Symbol && t.GetVal() == "." { //subroutine call
			ts = append(ts, it.Next()) //add "."
			st, err := compileSubCall(it)
			if err != nil {
				return nil, err
			}
			ts = append(ts, st...)
		}
	}

	return
}

func compileIfStatement(it *TokenIterator) (nt *NonTerminalToken, err error) {
	nt = &NonTerminalToken{
		tokenType: IfStatement,
	}

	token := it.Next()
	if token.GetVal() != "if" {
		err := assertToken(token, Keyword, "if")
		if err != nil {
			return nil, err
		}
	}
	nt.AddSubToken(token) //if
	if token.GetVal() == "if" {
		st, e := withParentheses(it, compileExpression) // (expression)
		if e != nil {
			return nil, e
		}
		nt.AddSubToken(st...)
	}

	sts, ee := withCurlyBrackets(it, compileStatements) //{statements}
	if ee != nil {
		return nil, ee
	}
	nt.AddSubToken(sts...)

	next := it.Peek()
	if next.GetVal() == "else" {
		nt.AddSubToken(it.Next())
		sts, ee := withCurlyBrackets(it, compileStatements) //{statements}
		if ee != nil {
			return nil, ee
		}
		nt.AddSubToken(sts...)
	}
	return nt, nil
}

func compileWhileStatement(it *TokenIterator) (nt *NonTerminalToken, err error) {
	nt = &NonTerminalToken{
		tokenType: WhileStatement,
	}

	token := it.Next()
	if err = assertToken(token, Keyword, "while"); err != nil {
		return
	}
	nt.AddSubToken(token) //while

	st, e := withParentheses(it, compileExpression)
	if e != nil {
		return nil, e
	}
	nt.AddSubToken(st...)

	sts, ee := withCurlyBrackets(it, compileStatements) //{statements}
	if ee != nil {
		return nil, ee
	}
	nt.AddSubToken(sts...)
	return
}

func compileLetStatement(it *TokenIterator) (nt *NonTerminalToken, err error) {
	nt = &NonTerminalToken{
		tokenType: LetStatement,
	}

	token := it.Next()
	if err = assertToken(token, Keyword, "let"); err != nil {
		return nil, err
	}
	nt.AddSubToken(token) //let

	varname, err := compileIdentifier(it)
	if err != nil {
		return nil, err
	}
	nt.AddSubToken(varname...) //varname

	token = it.Next()
	if err = assertToken(token, Symbol, "="); err != nil {
		return nil, err
	}
	nt.AddSubToken(token)

	exp, err := compileExpression(it)
	if err != nil {
		return nil, err
	}
	nt.AddSubToken(exp) //expression

	token = it.Next()
	if err = assertToken(token, Symbol, ";"); err != nil {
		return nil, err
	}
	nt.AddSubToken(token)

	return
}

func compileDoStatement(it *TokenIterator) (nt *NonTerminalToken, err error) {
	nt = &NonTerminalToken{
		tokenType: DoStatement,
	}

	token := it.Next()
	if err = assertToken(token, Keyword, "do"); err != nil {
		return nil, err
	}
	nt.AddSubToken(token) //do

	token = it.Next()
	if err = assertToken(token, Identifier, ""); err != nil {
		return nil, err
	}
	nt.AddSubToken(token)

	token = it.Peek()

	if IsDot(token) { //method call
		nt.AddSubToken(it.Next())

		token = it.Next()
		if err = assertToken(token, Identifier, ""); err != nil {
			return nil, err
		}
		nt.AddSubToken(token)
	}

	ts, err := withParentheses(it, compileExpressionList)
	if err != nil {
		return nil, err
	}
	nt.AddSubToken(ts...)

	token = it.Next()
	if err = assertToken(token, Symbol, ";"); err != nil {
		return nil, err
	}
	nt.AddSubToken(token)

	return
}

func compileReturnStatement(it *TokenIterator) (nt *NonTerminalToken, err error) {
	nt = &NonTerminalToken{
		tokenType: ReturnStatement,
	}
	token := it.Next()
	if err = assertToken(token, Keyword, "return"); err != nil {
		return nil, err
	}
	nt.AddSubToken(token) //return

	if it.Peek().GetType() != Symbol && it.Peek().GetVal() != ";" {
		ct, err := compileExpression(it)
		if err != nil {
			return nil, err
		}
		nt.AddSubToken(ct)
	}

	token = it.Next()
	if err = assertToken(token, Symbol, ";"); err != nil {
		return nil, err
	}
	nt.AddSubToken(token)
	return
}

func compileParameters(it *TokenIterator) (*NonTerminalToken, error) {
	ts := &NonTerminalToken{
		tokenType: ParameterList,
	}
	for it.HasNext() {
		if parameterListEndChecker(it) {
			return ts, nil
		}
		ts.AddSubToken(it.Next())
	}
	return ts, nil
}

func compileClassVarDec(it *TokenIterator) *NonTerminalToken {
	nt := &NonTerminalToken{
		tokenType: ClassVarDec,
	}
	for it.HasNext() {
		t := it.Next()
		nt.AddSubToken(t)
		if statementEndChecker(it) {
			nt.AddSubToken(it.Next())
			return nt
		}
	}
	return nt
}

var langSupportedTypes = []string{"int", "string", "Array", "char", "boolean"}

func compileVarDec(it *TokenIterator) (nt *NonTerminalToken, err error) {
	nt = &NonTerminalToken{
		tokenType: VarDec,
	}

	token := it.Next()
	if err = assertToken(token, Keyword, "var"); err != nil {
		return nil, err
	}
	nt.AddSubToken(token) //var

	token = it.Next()
	if !ContainsString(langSupportedTypes, token.GetVal()) && token.GetType() != Identifier {
		return nil, newCompileError(token, fmt.Sprintf("invalid var type:%s", token.GetVal()))
	}
	nt.AddSubToken(token) //int,string,class

	for it.HasNext() {
		token = it.Next()
		if err = assertToken(token, Identifier, ""); err != nil {
			return nil, err
		}
		nt.AddSubToken(token)

		next := it.Peek()

		if next.GetVal() == ";" && next.GetType() == Symbol {
			nt.AddSubToken(it.Next())
			return
		}

		token = it.Next()
		if err = assertToken(token, Symbol, ","); err != nil {
			return nil, err
		}
		nt.AddSubToken(token) //','
	}
	return
}

//handle a.b.c(exp)
func compileSubCall(it *TokenIterator) (ts []Token, err error) {
	for it.HasNext() {
		token := it.Next()
		if err = assertToken(token, Identifier, ""); err != nil {
			return nil, err
		}
		ts = append(ts, token)

		next := it.Peek()

		if next.GetVal() == "(" && next.GetType() == Symbol {
			el, e := withParentheses(it, compileExpressionList)
			if e != nil {
				return nil, err
			}
			ts = append(ts, el...)
			return ts, nil
		}

		if err = assertToken(next, Symbol, "."); err != nil {
			return nil, err
		}
		ts = append(ts, it.Next())
	}
	return
}
