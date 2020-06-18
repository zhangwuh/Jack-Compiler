package jack_compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSymbolTable(t *testing.T) {
	classSt := NewClassLevelTable()
	assert.Nil(t, classSt.add(variable{name :"x", typ: vinteger, kind: kfield}))
	assert.Nil(t, classSt.add(variable{name :"y", typ: vstring, kind: kfield}))
	assert.Nil(t, classSt.add(variable{name :"counter", typ: vinteger, kind: kstatic}))
	assert.NotNil(t, classSt.add(variable{name :"counter", typ: vinteger, kind: kargument}))

	v, ok := classSt.get("y")
	assert.True(t, ok)
	assert.Equal(t, v.offset, 1)
	assert.Equal(t, v.kind, kfield)

	v, ok = classSt.get("x")
	assert.True(t, ok)
	assert.Equal(t, v.offset, 0)
	assert.Equal(t, v.kind, kfield)
	assert.Equal(t, v.typ, vinteger)

	v, ok = classSt.get("counter")
	assert.True(t, ok)
	assert.Equal(t, v.offset, 0)
	assert.Equal(t, v.kind, kstatic)


	subSt := NewSubroutineLevelTable()
	assert.Nil(t, subSt.add(variable{name :"other", typ: vpointer, kind:kargument}))
	assert.Nil(t, subSt.add(variable{name :"dx", typ: vinteger, kind: klocal}))
	assert.Nil(t, subSt.add(variable{name :"dy", typ: vboolean, kind: klocal}))
	assert.NotNil(t, subSt.add(variable{name :"counter", typ: vinteger, kind: kstatic}))

	v, ok = subSt.get("this")
	assert.True(t, ok)
	assert.Equal(t, v.offset, 0)
	assert.Equal(t, v.kind, kargument)
	assert.Equal(t, v.typ, vpointer)


	v, ok = subSt.get("other")
	assert.True(t, ok)
	assert.Equal(t, v.offset, 1)
	assert.Equal(t, v.kind, kargument)
	assert.Equal(t, v.typ, vpointer)

	v, ok = subSt.get("dx")
	assert.True(t, ok)
	assert.Equal(t, v.offset, 0)
	assert.Equal(t, v.kind, klocal)

}
