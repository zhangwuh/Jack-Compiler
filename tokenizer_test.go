package jack_compiler

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_removeComments(t *testing.T) {
	assert.Equal(t, removeComments("//ttt xxx"), "")
	assert.Equal(t, removeComments("///xxxx"), "")
	assert.Equal(t, removeComments("aa///"), "aa")
	assert.Equal(t, removeComments("bbb //ttt xxx //xxx"), "bbb")
	assert.Equal(t, removeComments("xxx"), "xxx")

	assert.Equal(t, removeComments("/*ttt xxx */"), "")
	assert.Equal(t, removeComments("aa/*ttt xxx */"), "aa")
	assert.Equal(t, removeComments("aaa/*ttt xxx */ bbb"), "aaa bbb")
	assert.Equal(t, removeComments("  /**ttt xxx ****/  "), "")
	assert.Equal(t, removeComments("  * ttt xxx ****  "), "")
	assert.Equal(t, removeComments("  /** ttt xxx ****  "), "")
}

func TestTokenizer_Tokenize(t *testing.T) {
	file, _ := os.Open("Square/SquareGame.jack")
	defer file.Close()
	tokenizer := &tokenizer{}
	tokenizer.Tokenize(file)
	writer := &tokensOnlyWriter{}
	fo, err := os.Create("output/SquareGameT.xml")
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer fo.Close()
	writer.Write(tokenizer.tokens, fo)
}
