package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompileDir(t *testing.T) {
	err := CompileDir("/Users/zhangwuh/dev/playground/jack-compiler/sample/fibonacci", "/Users/zhangwuh/dev/playground/jack-compiler/output/vm/fibonacci")
	assert.Nil(t, err)
}

func TestCompileFile(t *testing.T) {
	CompileFile("/Users/zhangwuh/dev/playground/jack-compiler/sample/Square/SquareGame.jack", ".")
}
