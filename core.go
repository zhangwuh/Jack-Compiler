package jack_compiler

import (
	"fmt"
	"io"
	"strings"
)

type Tokenizer interface {
	Tokenize(rd io.Reader) []TerminalToken
}

type Compiler interface {
	LexialAnalysis(tokens []Token) (*NonTerminalToken, error)
}

type TokenType string

const (
	//terminalElements
	Keyword         TokenType = "keyword"
	Identifier      TokenType = "identifier"
	Symbol          TokenType = "symbol"
	IntegerConstant TokenType = "integerConstant"
	StringConstant  TokenType = "stringConstant"

	//nonTerminalElements
	Class          TokenType = "class"
	ClassVarDec    TokenType = "classVarDec"
	SubroutineDec  TokenType = "subroutineDec"
	ParameterList  TokenType = "parameterList"
	SubroutineBody TokenType = "subroutineBody"

	VarDec          TokenType = "varDec"
	Statements      TokenType = "statements"
	LetStatement    TokenType = "letStatement"
	IfStatement     TokenType = "ifStatement"
	WhileStatement  TokenType = "whileStatement"
	DoStatement     TokenType = "doStatement"
	ReturnStatement TokenType = "returnStatement"
	VarStatement    TokenType = "varStatement"
	Expression      TokenType = "expression"
	TokenTerm       TokenType = "term"
	ExpressionList  TokenType = "expressionList"
)

func typeOf(t Token) TokenType {
	switch t.GetVal() {
	case "if":
		return IfStatement
	case "while":
		return WhileStatement
	case "do":
		return DoStatement
	case "return":
		return ReturnStatement
	case "let":
		return LetStatement
	case "var":
		return VarStatement
	}
	return t.GetType()
}

var operations = map[string]string{
	"+": "add",
	"-": "sub",
	"&": "and",
	"|": "or",
	"<": "lt",
	">": "gt",
	"=": "eq",
	"/": "call Math.divide 2",
	"*": "call Math.multiply 2",
}

var terminalElements = []TokenType{Keyword, Identifier, Symbol, IntegerConstant, StringConstant}

func isTerminalElement(t TokenType) bool {
	for _, te := range terminalElements {
		if te == t {
			return true
		}
	}
	return false
}

func isSymbol(r rune) bool {
	for _, sr := range symbols {
		if sr == r {
			return true
		}
	}
	return false
}

var keywords = []string{"class", "constructor", "function", "method", "field", "static", "var", "int", "char", "boolean",
	"void", "true", "false", "null", "this", "let", "do", "if", "else", "while", "return"}
var symbols = []rune{'{', '}', '(', ')', '[', ']', '.', ',', ';', '+', '-', '*', '/', '&', '|', '<', '>', '=', '~', '<', '>', '&'}

var keywordConstants = []string{"null", "this", "true", "false"}

func isKeywordConstant(key Token) bool {
	return ContainsString(keywordConstants, key.GetVal()) && key.GetType() == Keyword
}

type Token interface {
	GetType() TokenType
	GetVal() string
	SubTokens() []Token
	IsTerminal() bool
	AsText() string
	AddSubToken(ts ...Token)
	Position() int //line number of source code
}
type TerminalToken struct {
	tokenType TokenType
	val       string
	line      int
}

func (tt *TerminalToken) GetType() TokenType {
	return tt.tokenType
}

func (tt *TerminalToken) GetVal() string {
	return tt.val
}

func (tt *TerminalToken) SubTokens() []Token {
	return nil
}

func (tt *TerminalToken) AddSubToken(ts ...Token) {
}

func (tt *TerminalToken) AsText() string {
	if tt.tokenType == StringConstant {
		tt.val = strings.ReplaceAll(tt.val, "\"", "")
	}
	return fmt.Sprintf("<%s>%s</%s>", tt.tokenType, EscapeXml(tt.val), tt.tokenType)
}

func (tt *TerminalToken) IsTerminal() bool {
	return true
}

func (tt *TerminalToken) Position() int {
	return tt.line
}

type NonTerminalToken struct {
	tokenType TokenType
	subTokens []Token
}

func (tt *NonTerminalToken) GetType() TokenType {
	return tt.tokenType
}

func (tt *NonTerminalToken) GetVal() string {
	return ""
}

func (tt *NonTerminalToken) SubTokens() []Token {
	return tt.subTokens
}

func (tt *NonTerminalToken) AsText() string {
	body := ""
	for _, t := range tt.subTokens {
		body += t.AsText()
	}
	return fmt.Sprintf("<%s>%s</%s>", tt.tokenType, body, tt.tokenType)
}

func (tt *NonTerminalToken) IsTerminal() bool {
	return false
}

func (tt *NonTerminalToken) AddSubToken(t ...Token) {
	tt.subTokens = append(tt.subTokens, t...)
}

func (tt *NonTerminalToken) Position() int {
	if len(tt.SubTokens()) == 0 {
		return -1
	}
	return tt.subTokens[0].Position()
}
