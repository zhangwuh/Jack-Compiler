package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenIterator(t *testing.T) {
	it := &TokenIterator{
		tokens: []Token{
			&TerminalToken{}, &TerminalToken{}, &TerminalToken{},
		},
	}
	assert.True(t, it.HasNext())
	it.Next()
	it.Next()
	assert.NotNil(t, it.Peek())
	assert.NotNil(t, it.Peek())
	assert.True(t, it.HasNext())
	it.Next()
	assert.False(t, it.HasNext())
	assert.Nil(t, it.Next())
	assert.Nil(t, it.Next())
	assert.Nil(t, it.Peek())
}
