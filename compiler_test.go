package jack_compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompileDir(t *testing.T) {
	err := compileDir("/Users/zhangwuh/dev/playground/jack-compiler/sample/Square", "")
	assert.Nil(t, err)
}
