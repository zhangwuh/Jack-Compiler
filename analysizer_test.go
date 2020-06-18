package jack_compiler

import (
	"fmt"
	"os"
	"testing"
)

func TestCompiler_Compile(t *testing.T) {
	file, _ := os.Open("foo.jack")
	defer file.Close()
	tokenizer := &tokenizer{}
	tokenizer.Tokenize(file)
	compiler := &analysizer{}
	compiled, err := compiler.LexialAnalysis(tokenizer.tokens)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	writer := &nonTerminalTokenWriter{}
	fo, err := os.Create("output/foo.xml")
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer fo.Close()
	writer.Write(fo, compiled)
}
