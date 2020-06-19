package compiler

import (
	"errors"
	"fmt"
)

type symbolTable struct {
	table   map[string]variable
	counter map[vKind]int
	parent  *symbolTable
}

func NewClassSymbolTable() *symbolTable {
	return &symbolTable{
		table: map[string]variable{},
		counter: map[vKind]int{
			kfield: 0, kstatic: 0,
		},
	}
}

var thisRef = variable{
	name: "this",
	typ:  vpointer,
	kind: kargument,
}

func NewSubroutineSymbolTable(parent *symbolTable) (st *symbolTable) {
	return &symbolTable{
		table: map[string]variable{},
		counter: map[vKind]int{
			klocal: 0, kargument: 0,
		},
		parent: parent,
	}
}

type vType string
type vKind string

const (
	vpointer vType = "pointer"
	vinteger vType = "int"
	vstring  vType = "string"
	vboolean vType = "boolean"

	kargument vKind = "argument" //arguments
	klocal    vKind = "local"    //local var inside method or function
	kfield    vKind = "field"    // field in class
	kstatic   vKind = "static"   // static var
)

type variable struct {
	name   string
	typ    vType
	kind   vKind
	offset int
}

func (v variable) memSeg() string {
	if v.kind == kfield {
		return "this"
	}
	return string(v.kind)
}

var emptyVar = variable{}

func (t *symbolTable) get(name string) (variable, bool) {
	if v, ok := t.table[name]; ok {
		return v, true
	}
	return emptyVar, false
}

func (t *symbolTable) getRecursively(name string) (variable, bool) {
	if v, ok := t.get(name); ok {
		return v, true
	}
	if t.parent != nil {
		return t.parent.get(name)
	}
	return emptyVar, false
}

func (t *symbolTable) getNumber(name string) (int, bool) {
	v, ok := t.get(name)
	return v.offset, ok
}

func redeclaredVarErr(dec variable) error {
	return errors.New(fmt.Sprintf("redeclared var:%s", dec.name))
}

func undeclaredVarErr(name string) error {
	return errors.New(fmt.Sprintf("undefined var %s", name))
}

func (t *symbolTable) add(v variable) error {
	if _, ok := t.get(v.name); ok {
		return redeclaredVarErr(v)
	}
	defer t.incCounter(v.kind)
	offset, err := t.count(v.kind)
	if err != nil {
		return err
	}
	v.offset = offset
	t.table[v.name] = v
	return nil
}

func (t *symbolTable) incCounter(kind vKind) {
	if _, ok := t.counter[kind]; ok {
		t.counter[kind]++
	}
}

func (t *symbolTable) count(kind vKind) (int, error) {
	if c, ok := t.counter[kind]; ok {
		return c, nil
	}
	return 0, fmt.Errorf("unsupported kind")
}

func (t *symbolTable) asMethod() *symbolTable {
	t.add(thisRef) // 'this' is always the first element in the symbol table
	return t
}

type linkedListNode struct {
	table  *symbolTable
	nextTo *linkedListNode
}

func (l *linkedListNode) next() *linkedListNode {
	if l == nil {
		return nil
	}
	return l.nextTo
}

type linkedList struct {
	head *linkedListNode
}

func (l *linkedList) isEmpty() bool {
	return l == nil || l.head == nil
}
