package compiler

import (
	"fmt"
	"io"
	"strings"
)

func assertToken(t Token, typ TokenType, val string) error {
	if typ != t.GetType() || (len(val) > 0 && t.GetVal() != val) {
		return newGrammarError(t, fmt.Sprintf("analysizer error, encountered:%s, expected:%s", t.GetVal(), val))
	}
	return nil
}

type nonTerminalTokenWriter struct{}

func (ntw *nonTerminalTokenWriter) Write(writer io.Writer, ts ...Token) {
	for _, ts := range ts {
		writer.Write([]byte(ts.AsText()))
	}
}

type analysizer struct {
}

var statementEndChecker = func(it *TokenIterator) bool {
	return it.Peek().GetVal() == ";"
}

var parameterListEndChecker = func(it *TokenIterator) bool {
	return it.Peek().GetVal() == ")"
}

var bodyEndChecker = func(it *TokenIterator) bool {
	return it.Peek().GetVal() == "}"
}

func (cp *analysizer) LexialAnalysis(tokens []Token) (*NonTerminalToken, error) {
	it := &TokenIterator{tokens: tokens}
	if it.Size() == 0 {
		return nil, nil
	}
	return analysisClass(it)
}

func analysisClass(it *TokenIterator) (nt *NonTerminalToken, err error) {
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
			nt.AddSubToken(analysisClassVarDec(it))
			continue
		case "constructor", "function", "method":
			sub, se := analysizerSubRoutineDec(it)
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

func analysizerSubRoutineDec(it *TokenIterator) (*NonTerminalToken, error) {
	nt := &NonTerminalToken{
		tokenType: SubroutineDec,
	}

	nt.AddSubToken(it.Next()) //constructor, method, function
	nt.AddSubToken(it.Next()) //class name in constructor, void, return type
	nt.AddSubToken(it.Next()) //func,method name

	ts, _ := withParentheses(it, analysisParameters)
	nt.AddSubToken(ts...)

	st, err := analysisSubRoutineBody(it)
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

func analysisExpressionList(it *TokenIterator) (nt *NonTerminalToken, err error) {
	nt = &NonTerminalToken{
		tokenType: ExpressionList,
	}
	for it.HasNext() {
		if parameterListEndChecker(it) {
			return
		}

		exp, err := analysisExpression(it)
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

func analysisExpression(it *TokenIterator) (nt *NonTerminalToken, err error) {
	nt = &NonTerminalToken{
		tokenType: Expression,
	}
	leftTerm, e := analysisTerm(it)
	if e != nil {
		return nil, e
	}
	nt.AddSubToken(leftTerm)

	token := it.Peek()
	if operations[token.GetVal()] == "" {
		return
	}

	nt.AddSubToken(it.Next()) //operation symbol

	rightTerm, e := analysisTerm(it)
	if e != nil {
		return nil, e
	}
	nt.AddSubToken(rightTerm)
	return
}

func analysisTerm(it *TokenIterator) (*NonTerminalToken, error) {
	term := &NonTerminalToken{
		tokenType: TokenTerm,
	}
	next := it.Peek()
	if next.GetVal() == "(" { //expression
		st, err := withParentheses(it, analysisExpression)
		if err != nil {
			return nil, err
		}
		term.AddSubToken(st...)

	} else if next.GetType() == Identifier { // x , x.a()
		token, err := analysisIdentifier(it)
		if err != nil {
			return nil, err
		}
		term.AddSubToken(token...)
	} else if next.GetType() == Symbol { //-1, -i
		term.AddSubToken(it.Next())
		st, err := analysisTerm(it)
		if err != nil {
			return nil, err
		}
		term.AddSubToken(st)
	} else if next.GetType() == StringConstant {
		next := it.Next()
		(next.(*TerminalToken)).val = strings.ReplaceAll(next.GetVal(), "\"", "")
		term.AddSubToken(next)
	} else if next.GetType() == IntegerConstant || isKeywordConstant(next) {
		term.AddSubToken(it.Next())
	} else {
		return nil, newGrammarError(next, fmt.Sprintf("invalid grammar error in if statement:%s", next.AsText()))
	}
	return term, nil
}

func analysisSubRoutineBody(it *TokenIterator) (Token, error) {
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
			st, err := analysisStatements(it)
			if err != nil {
				return nil, err
			}
			nt.AddSubToken(st)
		} else if tt == VarStatement {
			ts, err := analysisVarDec(it)
			if err != nil {
				return nil, err
			}
			nt.AddSubToken(ts)
		} else {
			return nil, newGrammarError(it.Peek(), fmt.Sprintf("grammar error, unknow token:%s", t.AsText()))
		}
	}

	token = it.Next() //}
	if e := assertToken(token, Symbol, "}"); e != nil {
		return nil, e
	}
	nt.AddSubToken(token)
	return nt, nil
}

func analysisStatements(it *TokenIterator) (nt *NonTerminalToken, err error) {
	nt = &NonTerminalToken{
		tokenType: Statements,
	}

	for it.HasNext() {
		token := it.Peek()
		switch typeOf(token) {
		case IfStatement:
			ts, err := analysisIfStatement(it)
			if err != nil {
				return nil, err
			}
			nt.AddSubToken(ts)
			continue
		case WhileStatement:
			ts, err := analysisWhileStatement(it)
			if err != nil {
				return nil, err
			}
			nt.AddSubToken(ts)
			continue
		case LetStatement:
			ts, err := analysisLetStatement(it)
			if err != nil {
				return nil, err
			}
			nt.AddSubToken(ts)
			continue
		case DoStatement:
			ts, err := analysisDoStatement(it)
			if err != nil {
				return nil, err
			}
			nt.AddSubToken(ts)
			continue
		case ReturnStatement:
			ts, err := analysisReturnStatement(it)
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
func analysisIdentifier(it *TokenIterator) (ts []Token, err error) {
	t := it.Next()
	if err := assertToken(t, Identifier, ""); err != nil {
		return nil, err
	}
	ts = append(ts, t)

	t = it.Peek()
	if t.GetType() == Symbol && t.GetVal() == "[" {
		ts = append(ts, it.Next())

		exp, err := analysisExpression(it)
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
			st, err := analysisSubCall(it)
			if err != nil {
				return nil, err
			}
			ts = append(ts, st...)
		}
	}

	return
}

func analysisIfStatement(it *TokenIterator) (nt *NonTerminalToken, err error) {
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
		st, e := withParentheses(it, analysisExpression) // (expression)
		if e != nil {
			return nil, e
		}
		nt.AddSubToken(st...)
	}

	sts, ee := withCurlyBrackets(it, analysisStatements) //{statements}
	if ee != nil {
		return nil, ee
	}
	nt.AddSubToken(sts...)

	next := it.Peek()
	if next.GetVal() == "else" {
		nt.AddSubToken(it.Next())
		sts, ee := withCurlyBrackets(it, analysisStatements) //{statements}
		if ee != nil {
			return nil, ee
		}
		nt.AddSubToken(sts...)
	}
	return nt, nil
}

func analysisWhileStatement(it *TokenIterator) (nt *NonTerminalToken, err error) {
	nt = &NonTerminalToken{
		tokenType: WhileStatement,
	}

	token := it.Next()
	if err = assertToken(token, Keyword, "while"); err != nil {
		return
	}
	nt.AddSubToken(token) //while

	st, e := withParentheses(it, analysisExpression)
	if e != nil {
		return nil, e
	}
	nt.AddSubToken(st...)

	sts, ee := withCurlyBrackets(it, analysisStatements) //{statements}
	if ee != nil {
		return nil, ee
	}
	nt.AddSubToken(sts...)
	return
}

func analysisLetStatement(it *TokenIterator) (nt *NonTerminalToken, err error) {
	nt = &NonTerminalToken{
		tokenType: LetStatement,
	}

	token := it.Next()
	if err = assertToken(token, Keyword, "let"); err != nil {
		return nil, err
	}
	nt.AddSubToken(token) //let

	varname, err := analysisIdentifier(it)
	if err != nil {
		return nil, err
	}
	nt.AddSubToken(varname...) //varname

	token = it.Next()
	if err = assertToken(token, Symbol, "="); err != nil {
		return nil, err
	}
	nt.AddSubToken(token)

	exp, err := analysisExpression(it)
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

func analysisDoStatement(it *TokenIterator) (nt *NonTerminalToken, err error) {
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

	ts, err := withParentheses(it, analysisExpressionList)
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

func analysisReturnStatement(it *TokenIterator) (nt *NonTerminalToken, err error) {
	nt = &NonTerminalToken{
		tokenType: ReturnStatement,
	}
	token := it.Next()
	if err = assertToken(token, Keyword, "return"); err != nil {
		return nil, err
	}
	nt.AddSubToken(token) //return

	if it.Peek().GetType() != Symbol && it.Peek().GetVal() != ";" {
		ct, err := analysisExpression(it)
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

func analysisParameters(it *TokenIterator) (*NonTerminalToken, error) {
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

func analysisClassVarDec(it *TokenIterator) *NonTerminalToken {
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

func analysisVarDec(it *TokenIterator) (nt *NonTerminalToken, err error) {
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
		return nil, newGrammarError(token, fmt.Sprintf("invalid var type:%s", token.GetVal()))
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
func analysisSubCall(it *TokenIterator) (ts []Token, err error) {
	for it.HasNext() {
		token := it.Next()
		if err = assertToken(token, Identifier, ""); err != nil {
			return nil, err
		}
		ts = append(ts, token)

		next := it.Peek()

		if next.GetVal() == "(" && next.GetType() == Symbol {
			el, e := withParentheses(it, analysisExpressionList)
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
