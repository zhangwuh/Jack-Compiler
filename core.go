package jack_compiler

import (
	"fmt"
	"io"
	"strings"
)

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
	Expression      TokenType = "expression"
	Term            TokenType = "term"
	ExpressionList  TokenType = "expressionList"
)

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

type Token interface {
	GetType() TokenType
	GetVal() string
	SubTokens() []Token
	IsTerminal() bool
	AsText() string
}
type TerminalToken struct {
	tokenType TokenType
	val       string
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

func (tt *TerminalToken) AsText() string {
	if tt.tokenType == StringConstant {
		tt.val = strings.ReplaceAll(tt.val, "\"", "")
	}
	return fmt.Sprintf("<%s>%s</%s>", tt.tokenType, EscapeXml(tt.val), tt.tokenType)
}

func (tt *TerminalToken) IsTerminal() bool {
	return true
}

type NonTerminalToken struct {
	tokenType TokenType
	tokens    []Token
}

func (tt *NonTerminalToken) GetType() TokenType {
	return tt.tokenType
}

func (tt *NonTerminalToken) GetVal() string {
	return ""
}

func (tt *NonTerminalToken) SubTokens() []Token {
	return tt.tokens
}

func (tt *NonTerminalToken) AsText() string {
	body := ""
	for _, t := range tt.tokens {
		body += EscapeXml(t.AsText())
	}
	return fmt.Sprintf("<%s>%s</%s>", tt.tokenType, body, tt.tokenType)
}

func (tt *NonTerminalToken) IsTerminal() bool {
	return false
}

type Tokenizer interface {
	Tokenize(rd io.Reader) []TerminalToken
}

type Analyser interface {
	Analysis(tokens []Token) NonTerminalToken
}
