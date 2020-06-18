package jack_compiler

import "fmt"

type symbolTable struct {
	table map[string]variable
	counter map[vKind]int
}

func NewClassLevelTable() *symbolTable {
	return &symbolTable{
		table: map[string]variable{},
		counter: map[vKind]int{
			kfield: 0, kstatic: 0,
		},
	}
}

var thisRef = variable{
	name: "this",
	typ: vpointer,
	kind: kargument,
}

func NewSubroutineLevelTable() (st *symbolTable) {
	defer func() {
		st.add(thisRef) // 'this' is always the first element in the symbol table
	}()
	return &symbolTable{
		table: map[string]variable{},
		counter: map[vKind]int{
			klocal: 0, kargument: 0,
		},
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
	klocal    vKind = "local" //local var inside method or function
	kfield    vKind = "field" // field in class
	kstatic   vKind = "static" // static var
)

type variable struct {
	name string
	typ vType
	kind vKind
	offset int
}

var emptyVar = variable{}

func (t *symbolTable) get(name string) (variable, bool) {
	if v, ok := t.table[name];ok {
		return v, true
	}
	return emptyVar, false
}

func (t *symbolTable) getNumber(name string) (int, bool) {
	v, ok := t.get(name)
	return v.offset, ok
}

func (t *symbolTable) add(v variable) error {
	if _, ok := t.get(v.name);ok {
		return fmt.Errorf("%s already defined", v.name)
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
	if _,ok := t.counter[kind]; ok {
		t.counter[kind]++
	}
}

func (t *symbolTable) count(kind vKind) (int, error) {
	if c, ok := t.counter[kind];ok {
		return c, nil
	}
	return 0, fmt.Errorf("unsupported kind:%v", kind)
}

type linkedListNode struct {
	table *symbolTable
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
