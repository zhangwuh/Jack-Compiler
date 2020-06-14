package jack_compiler

import (
	"fmt"
	"os"
	"testing"
)

func TestCompiler_Compile(t *testing.T) {
	file, _ := os.Open("sample/test.jack")
	defer file.Close()
	tokenizer := &tokenizer{}
	tokenizer.Tokenize(file)
	compiler := &compiler{}
	compiled, err := compiler.Compile(tokenizer.tokens)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	writer := &nonTerminalTokenWriter{}
	fo, err := os.Create("output/SquareGame.xml")
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer fo.Close()
	writer.Write(fo, compiled)
}
