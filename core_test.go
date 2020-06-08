package jack_compiler

import (
	"fmt"
	"testing"
)

func TestNonTerminalToken_AsText(t *testing.T) {
	token := &NonTerminalToken{
		tokenType: Class,
		tokens: []Token{
			&TerminalToken{
				tokenType: Keyword,
				val:       "function",
			},
			&NonTerminalToken{
				tokenType: VarDec,
				tokens: []Token{
					&TerminalToken{
						tokenType: Keyword,
						val:       "var",
					},
					&TerminalToken{
						tokenType: Keyword,
						val:       "int",
					},
					&TerminalToken{
						tokenType: Identifier,
						val:       "length",
					},
					&TerminalToken{
						tokenType: Symbol,
						val:       ";",
					},
				},
			},
		},
	}

	fmt.Println(token.AsText())
}
