package jack_compiler

import (
	"fmt"
	"io"
)

func assertToken(t Token, typ TokenType, val string) error {
	if typ != t.GetType() || (len(val) > 0 && t.GetVal() != val) {
		return fmt.Errorf("invalid grammar error, wrong type:%s", val)
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
	token := it.Next()
	if err = assertToken(token, Keyword, "class"); err != nil {
		return
	}

	nt = &NonTerminalToken{
		tokenType: Class,
	}
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
	if token.GetVal() == ")" {
		return
	}

	if operations[token.GetVal()] != "" { //op
		nt.AddSubToken(it.Next())
	}

	rightTerm, e := compileTerm(it)
	if e != nil {
		return nil, e
	}
	nt.AddSubToken(rightTerm)

	token = it.Peek()
	if token.GetVal() == ")" {
		return
	} else {
		return nil, fmt.Errorf("invalid grammar error, err op:%s", token.GetVal())
	}
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

	} else if next.GetType() == Identifier || next.GetType() == IntegerConstant || next.GetType() == StringConstant {
		term.AddSubToken(it.Next())
	} else {
		return nil, fmt.Errorf("invalid grammar error in if statement:%s", next.AsText())
	}
	return term, nil
}

func compileSubRoutineBody(it *TokenIterator) (Token, error) {
	nt := &NonTerminalToken{
		tokenType: SubroutineBody,
	}

	st, err := withCurlyBrackets(it, compileStatements)
	if err != nil {
		return nil, err
	}
	nt.AddSubToken(st...)
	return nt, nil
}

func compileStatements(it *TokenIterator) (*NonTerminalToken, error) {
	st := &NonTerminalToken{
		tokenType: Statements,
	}

	for it.HasNext() {
		token := it.Peek()
		switch of(token) {
		case IfStatement:
			ts, err := compileIfStatement(it)
			if err != nil {
				return nil, err
			}
			st.AddSubToken(ts)
		default:
			it.Next()
		}
		if bodyEndChecker(it) {
			break
		}
	}

	return st, nil
}

func compileIfStatement(it *TokenIterator) (nt *NonTerminalToken, err error) {
	t := &NonTerminalToken{
		tokenType: IfStatement,
	}

	token := it.Next()
	if err = assertToken(token, Keyword, "if"); err != nil {
		return
	}
	t.AddSubToken(token) //if

	st, e := withParentheses(it, compileExpression) // (expression)
	if e != nil {
		return nil, e
	}
	t.AddSubToken(st...)

	sts, ee := withCurlyBrackets(it, compileStatements) //{statements}
	if ee != nil {
		return nil, ee
	}
	t.AddSubToken(sts...)
	return t, nil
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
