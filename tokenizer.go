package jack_compiler

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strconv"
)

type tokenizer struct {
	file         string
	tokens       []Token
	currentToken []rune
}

type TokensWriter interface {
	Write(writer io.Writer, ts ...Token)
}

type tokensOnlyWriter struct {
}

func (tow *tokensOnlyWriter) Write(writer io.Writer, ts ...Token) {
	writer.Write([]byte("<tokens>\n"))
	defer writer.Write([]byte("</tokens>"))
	for _, ts := range ts {
		writer.Write([]byte(ts.AsText() + "\n"))
	}

}

func (tokenizer *tokenizer) Tokenize(rd io.Reader) error {
	reader := bufio.NewReader(rd)
	var lineCount int
	for {
		lineCount++
		line, _, e := reader.ReadLine()
		if e != nil {
			if e != io.EOF {
				fmt.Println(fmt.Sprintf("read from reader err:%s", e.Error()))
				return e
			}
			//end of file
			return nil
		}
		e = tokenizer.tokenize(string(line))
		if e != nil {
			fmt.Println(fmt.Sprintf("err in file %s, line %d, error:%s", tokenizer.file, lineCount, e.Error()))
			return e
		}
	}
}

func (tokenizer *tokenizer) tokenize(line string) error {
	line = removeComments(line)
	if len(line) == 0 {
		return nil
	}
	return tokenizer.lexicalAnalysis(line)
}

func buildToken(currentToken []rune) (Token, error) {
	typ, err := resolveTokenType(string(currentToken))
	if err != nil {
		return nil, err
	}
	return &TerminalToken{typ, string(currentToken)}, nil
}

func (t *tokenizer) lexicalAnalysis(line string) error {
	for i := 0; i < len(line); i++ {
		r := []rune(line)[i]
		if isWord(r) {
			t.currentToken = append(t.currentToken, r)
		} else if isNumber(r) {
			t.currentToken = append(t.currentToken, r)
		} else if isSymbol(r) {
			err := t.flush()
			if err != nil {
				return err
			}
			tt := &TerminalToken{Symbol, string(r)}
			t.tokens = append(t.tokens, tt)
		} else if r == ' ' {
			err := t.flush()
			if err != nil {
				return err
			}
		} else if r == '"' {
			t.currentToken = append(t.currentToken, '"')
			for i < len(line) {
				i++
				nr := line[i]
				if nr == '"' {
					t.currentToken = append(t.currentToken, '"')
					err := t.flush()
					if err != nil {
						return err
					}
					break
				} else {
					t.currentToken = append(t.currentToken, rune(nr))
				}
			}
		}
	}
	err := t.flush()
	if err != nil {
		return err
	}
	return nil
}

func (t *tokenizer) flush() error {
	if len(t.currentToken) > 0 {
		tt, err := buildToken(t.currentToken)
		if err != nil {
			return err
		}
		t.currentToken = t.currentToken[:0]
		t.tokens = append(t.tokens, tt)
	}
	return nil
}

func resolveTokenType(s string) (TokenType, error) {
	if ContainsString(keywords, s) {
		return Keyword, nil
	}

	if stringReg.MatchString(s) {
		return StringConstant, nil
	}
	if identifierReg.MatchString(s) {
		return Identifier, nil
	}
	if i, err := strconv.Atoi(s); err == nil || i <= math.MaxInt16 {
		return IntegerConstant, nil
	}
	return "", fmt.Errorf("syntax error:%s", s)
}
