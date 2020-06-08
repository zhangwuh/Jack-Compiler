package jack_compiler

import (
	"bufio"
	"fmt"
	"io"
)

type tokenizer struct {
}

func (tokenizer *tokenizer) Tokenize(rd io.Reader) (tokens []TerminalToken) {
	reader := bufio.NewReader(rd)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			fmt.Println(fmt.Sprintf("read from reader err:%s", err.Error()))
			return
		}
		tokens = append(tokens, tokenize(line))
	}
}

func tokenize(bs []byte) TerminalToken {

}
