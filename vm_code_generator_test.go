package jack_compiler

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassWriter_Write(t *testing.T) {
	file, _ := os.Open("sample/Square/Square.jack")
	defer file.Close()
	tokenizer := &tokenizer{}
	tokenizer.Tokenize(file)
	analysizer := &analysizer{}
	output, err := analysizer.LexialAnalysis(tokenizer.tokens)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	cw := &classWriter{}
	fo, err := os.Create("output/Square.vm")
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer fo.Close()
	cw.Write(fo, output)
}

func Test_parseClass(t *testing.T) {
	file, err := os.Open("foo.jack")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer file.Close()
	tokenizer := &tokenizer{}
	tokenizer.Tokenize(file)
	analysizer := &analysizer{}
	output, err := analysizer.LexialAnalysis(tokenizer.tokens)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	class, err := parseClass(output)
	assert.Nil(t, err)
	assert.Equal(t, class.name, "Square")

	assert.Equal(t, len(class.declarations), 2)
	assert.Equal(t, class.declarations[1].typ, vinteger)
	assert.Equal(t, class.declarations[1].kind, kfield)
	assert.Equal(t, class.declarations[1].name, "size")

	assert.Equal(t, len(class.subroutines), 4)
	csct := class.subroutines[0]
	assert.Equal(t, csct.category, constructor)
	assert.Equal(t, csct.retType, "Square")
	assert.Equal(t, csct.name, "new")
	assert.Equal(t, len(csct.declarations), 4)
	assert.Equal(t, csct.declarations[2].name, "Asize")
	assert.Equal(t, csct.declarations[2].typ, vinteger)
	assert.Equal(t, csct.declarations[2].kind, kargument)

	assert.Equal(t, csct.declarations[3].name, "i")
	assert.Equal(t, csct.declarations[3].typ, vinteger)
	assert.Equal(t, csct.declarations[3].kind, klocal)
}
