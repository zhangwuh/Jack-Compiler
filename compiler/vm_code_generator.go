package compiler

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type jackClass struct {
	name         string
	declarations []variable
	subroutines  []subroutine
}

type vmCompiler struct {
	class         jackClass
	classSymTable *symbolTable
	labelCounter  int
}

func NewVmCompiler(class jackClass) *vmCompiler {
	return &vmCompiler{
		class:         class,
		classSymTable: NewClassSymbolTable(),
	}
}

func (vc *vmCompiler) compile() (string, error) {
	err := vc.compileClassDeclarations(vc.class.declarations)
	if err != nil {
		return "", err
	}
	code, err := vc.compileSubRoutines(vc.class.subroutines)
	if err != nil {
		return "", err
	}
	return code, nil
}

//parse tokenized results to structured class
func (w *vmCompiler) compileAndWrite(writer io.Writer) {
	code, err := w.compile()
	if err != nil {
		fmt.Printf("err when compule to vm code:%v", err)
		return
	}
	if _, err = writer.Write([]byte(code)); err != nil {
		fmt.Printf("err when write to output file:%v", err)
	}
}

func (vc *vmCompiler) compileClassDeclarations(declarations []variable) error {
	for _, dec := range declarations {
		if err := vc.classSymTable.add(dec); err != nil {
			return err
		}
	}
	return nil
}

func (vc *vmCompiler) compileSubRoutines(subroutines []subroutine) (string, error) {
	var lines []string
	for _, sub := range subroutines {
		sc := newSubRoutineCompiler(vc.class, vc.classSymTable, vc)
		sl, err := sc.compileSubRoutine(sub)
		if err != nil {
			return "", err
		}
		lines = append(lines, sl...)
	}
	return strings.Join(lines, "\n"), nil
}

type subRoutineCompiler struct {
	class  jackClass
	table  *symbolTable
	parent *vmCompiler
}

func newSubRoutineCompiler(class jackClass, parentTable *symbolTable, parent *vmCompiler) *subRoutineCompiler {
	return &subRoutineCompiler{class: class, table: NewSubroutineSymbolTable(parentTable), parent: parent}
}

func (c *subRoutineCompiler) compileSubRoutine(sub subroutine) ([]string, error) {
	if sub.category == method {
		c.table.asMethod() // 'this' is always the first element in the symbol table
	}

	var varCount int
	for _, dec := range sub.declarations {
		if dec.kind == klocal {
			varCount++
		}
		if err := c.table.add(dec); err != nil {
			return nil, err
		}
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("function %s.%s %d", c.class.name, sub.name, varCount))
	if sub.category == constructor {
		vcount, _ := c.table.parent.count(kfield)
		lines = append(lines, fmt.Sprintf("push constant %d", vcount)) //for memory alloc
		lines = append(lines, "call Memory.alloc 1")
		lines = append(lines, "pop pointer 0") //init address of 'this'
	} else if sub.category == method {
		lines = append(lines, "push argument 0")
		lines = append(lines, "pop pointer 0") //init address of 'this'
	}
	slines := c.compileStatements(sub.statements)
	lines = append(lines, slines...)
	return lines, nil
}

func (c *subRoutineCompiler) compileStatements(statements []Statement) []string {
	var lines []string
	for _, st := range statements {
		switch st.category() {
		case doSc:
			lines = append(lines, c.compileDoStatement(st.(doStatement))...)
		case retSc:
			lines = append(lines, c.compileReturnStatement(st.(retStatement))...)
		case letSc:
			lines = append(lines, c.compileLetStatement(st.(letStatement))...)
		case ifSc:
			lines = append(lines, c.compileIfStatement(st.(ifStatement))...)
		case whileSc:
			lines = append(lines, c.compileWhileStatement(st.(whileStatement))...)
		}
	}
	return lines
}

func (c *subRoutineCompiler) compileDoStatement(st doStatement) []string {
	var lines []string
	lines = append(lines, c.compileSubCall(st.action)...)
	lines = append(lines, "pop temp 0") // store sub call return val to temp
	return lines
}

func (c *subRoutineCompiler) compileSubCall(call subroutineCall) []string {
	var lines []string
	onTarget := call.target //object | method | function
	argSize := len(call.args)
	if len(onTarget) == 0 { //call on `this`
		onTarget = c.class.name
		lines = append(lines, "push pointer 0")
		argSize++
	} else {
		v, ok := c.table.getRecursively(onTarget)
		if ok { //call on `that`
			onTarget = string(v.typ)
			lines = append(lines, fmt.Sprintf("push %s %d", v.memSeg(), v.offset))
			argSize++
		}
	}
	for _, arg := range call.args {
		lines = append(lines, c.compileExpression(arg)...)
	}

	lines = append(lines, fmt.Sprintf("call %s.%s %d", onTarget, call.name, argSize))
	return lines
}

func (c *subRoutineCompiler) compileExpression(exp expression) []string {
	var lines []string

	var left, right Term
	var opIndex int
	for _, term := range exp.terms {
		if left == nil {
			left = term
			lines = append(lines, c.compileTerm(left)...)
		} else {
			right = term
			lines = append(lines, c.compileTerm(right)...)
			lines = append(lines, c.compileOperator(exp.operations[opIndex]))
			opIndex++
			left = nil
			right = nil
		}
	}

	return lines
}

func (c *subRoutineCompiler) compileTerm(term Term) []string {
	var lines []string

	switch term.category() {
	case constantTerm:
		lines = append(lines, c.compileConstTerm(term.(ConstTerm))...)
	case expressionTerm:
		lines = append(lines, c.compileExpression(term.(expression))...)
	case unaryTerm:
		lines = append(lines, c.compileUnaryTerm(term.(UnaryTerm))...)
	case referenceTerm:
		lines = append(lines, c.compileReferenceTerm(term.(ReferenceTerm))...)
	case subCallTerm:
		lines = append(lines, c.compileSubCall(term.(subroutineCall))...)
	}

	return lines
}

func (c *subRoutineCompiler) compileOperator(s string) string {
	return operations[s]
}

func (c *subRoutineCompiler) compileConstTerm(term ConstTerm) []string {
	var lines []string
	switch term.ttype {
	case IntegerConstant:
		lines = append(lines, fmt.Sprintf("push constant %d", term.val.(int)))
	case StringConstant:
		val := term.val.(string)
		lines = append(lines, fmt.Sprintf("push constant %d", len(val)))
		lines = append(lines, "call String.new 1")
		for _, r := range val {
			lines = append(lines, fmt.Sprintf("push constant %d", r))
			lines = append(lines, "call String.appendChar 2") //operate on base address of str and current rune
		}
	case Keyword:
		val := term.val.(string)
		if val == "null" || val == "false" {
			lines = append(lines, "push constant 0")
		} else if val == "true" { //true mapped to -1
			lines = append(lines, "push constant 1")
			lines = append(lines, "neg")
		} else if val == "this" {
			lines = append(lines, "push pointer 0")
		}
	}
	return lines
}

func (c *subRoutineCompiler) compileUnaryTerm(term UnaryTerm) []string {
	var lines []string

	lines = append(lines, c.compileTerm(term.term)...)
	if term.operator == "~" {
		lines = append(lines, "not")
	} else if term.operator == "-" {
		lines = append(lines, "neg")
	}
	return lines
}

func (c *subRoutineCompiler) compileArrayRef(arr variable, exp expression, forAssignment bool) []string {
	var lines []string
	lines = append(lines, fmt.Sprintf("push %s %d", arr.memSeg(), arr.offset))
	lines = append(lines, c.compileExpression(exp)...)
	lines = append(lines, "add")
	if forAssignment {
		lines = append(lines, "pop pointer 1")
		lines = append(lines, "pop that 0")
	} else {
		lines = append(lines, "pop pointer 1")
		lines = append(lines, "push that 0")
	}
	return lines
}

func (c *subRoutineCompiler) compileReferenceTerm(term ReferenceTerm) []string {
	var lines []string

	ref, ok := c.table.getRecursively(term.varName)
	if !ok {
		panic(undeclaredVarErr(term.varName))
	}
	if term.isArrayRef() {
		lines = append(lines, c.compileArrayRef(ref, term.index, false)...)
	} else {
		lines = append(lines, fmt.Sprintf("push %s %d", ref.memSeg(), ref.offset))
	}

	return lines
}

func (c *subRoutineCompiler) compileReturnStatement(statement retStatement) []string {
	var lines []string
	if statement.expression.isEmpty() {
		lines = append(lines, "push constant 0")
	} else {
		lines = append(lines, c.compileExpression(statement.expression)...)
	}
	lines = append(lines, "return")
	return lines
}

func (c *subRoutineCompiler) compileLetStatement(statement letStatement) []string {
	var lines []string
	lines = append(lines, c.compileExpression(statement.expression)...)
	target := statement.target
	v, ok := c.table.getRecursively(target.varName)
	if !ok {
		panic(undeclaredVarErr(target.varName))
	}
	if target.isArrayRef() {
		lines = append(lines, c.compileArrayRef(v, target.index, true)...)
	} else {
		lines = append(lines, fmt.Sprintf("pop %s %d", v.memSeg(), v.offset))
	}
	return lines
}

func (c *subRoutineCompiler) compileIfStatement(statement ifStatement) []string {
	var lines []string
	id := c.parent.labelCounter
	c.parent.labelCounter++

	lines = append(lines, c.compileExpression(statement.condition)...)
	lines = append(lines, fmt.Sprintf("if-goto IF_%d", id))
	lines = append(lines, c.compileStatements(statement.elseStatements)...)
	lines = append(lines, fmt.Sprintf("goto ENDIF_%d", id))
	lines = append(lines, fmt.Sprintf("label IF_%d", id))
	lines = append(lines, c.compileStatements(statement.statements)...)
	lines = append(lines, fmt.Sprintf("label ENDIF_%d", id))

	return lines
}

func (c *subRoutineCompiler) compileWhileStatement(statement whileStatement) []string {
	var lines []string
	id := c.parent.labelCounter
	c.parent.labelCounter++
	lines = append(lines, fmt.Sprintf("label WHILE_%d", id))
	lines = append(lines, c.compileExpression(statement.condition)...)
	lines = append(lines, "not")
	lines = append(lines, fmt.Sprintf("if-goto END_WHILE_%d", id))
	lines = append(lines, c.compileStatements(statement.statements)...)
	lines = append(lines, fmt.Sprintf("goto WHILE_%d", id))
	lines = append(lines, fmt.Sprintf("label END_WHILE_%d", id))
	return lines
}

var emptyClass = jackClass{}

//convert lexial tree to structed class
func parseClass(ts Token) (jackClass, error) {
	root := ts.(*NonTerminalToken)
	name, err := resolveClassName(root)
	if err != nil {
		return emptyClass, nil
	}

	jc := jackClass{
		name: name,
	}

	declarations, err := resolveClassVarDecs(root)
	if err != nil {
		return emptyClass, err
	}
	jc.declarations = append(jc.declarations, declarations...)

	subroutines, err := resolveSubroutines(root)
	if err != nil {
		return emptyClass, err
	}
	jc.subroutines = append(jc.subroutines, subroutines...)

	return jc, nil
}

func resolveSubroutines(root *NonTerminalToken) ([]subroutine, error) {
	var subs []subroutine
	for _, item := range root.subTokens {
		if item.GetType() == SubroutineDec {
			sdc, err := resolveSubroutine(item)
			if err != nil {
				return nil, err
			}
			subs = append(subs, sdc)
		}
	}
	return subs, nil
}

var emptySubroutine = subroutine{}

/*
 input:
<subroutineDec>
        <keyword>function</keyword>
        <keyword>void</keyword>
        <identifier>foo</identifier>
        <symbol>(</symbol>
        <parameterList>
			...
        </parameterList>
        <symbol>)</symbol>
        <subroutineBody>
			...
		</subroutineBody>
</subroutineDec>
*/
func resolveSubroutine(token Token) (subroutine, error) {
	sub := subroutine{}
	sts := token.SubTokens()
	sub.category = subroutineCategory(sts[0].GetVal())
	sub.retType = sts[1].GetVal()
	sub.name = sts[2].GetVal()
	params := match(token, ParameterList)
	if len(params) == 0 {
		return emptySubroutine, newSyntaxError(token)
	}
	sub.declarations = append(sub.declarations, resolveParams(params[0])...)

	body := match(token, SubroutineBody)[0]

	for _, st := range match(body, VarDec) {
		sub.declarations = append(sub.declarations, resolveVarDec(st)...)
	}

	for _, st := range match(body, Statements) {
		statements, err := resolveStatements(st)
		if err != nil {
			return emptySubroutine, err
		}
		sub.statements = append(sub.statements, statements...)
	}

	return sub, nil
}

func resolveVarDec(st Token) []variable {
	var vs []variable
	kind := klocal
	typ := vType(st.SubTokens()[1].GetVal())
	it := NewTokenIterator(st.SubTokens()[2:])
	for it.HasNext() {
		n := it.Next()
		if n.GetType() == Identifier {
			vs = append(vs, variable{
				name: n.GetVal(),
				kind: kind,
				typ:  typ,
			})
		}
	}
	return vs
}

func resolveStatements(st Token) ([]Statement, error) {
	if err := assertToken(st, Statements, ""); err != nil {
		return nil, err
	}
	var sts []Statement
	for _, s := range st.SubTokens() {
		statement, err := resolveStatement(s)
		if err != nil {
			return nil, err
		}
		sts = append(sts, statement)
	}
	return sts, nil
}

func resolveStatement(st Token) (Statement, error) {
	switch st.GetType() {
	case IfStatement:
		return resolveIfStatement(st)
	case LetStatement:
		return resolveLetStatement(st)
	case DoStatement:
		return resolveDoStatement(st)
	case ReturnStatement:
		return resolveReturnStatement(st)
	case WhileStatement:
		return resolveWhileStatement(st)
	default:
		return nil, newSyntaxError(st)
	}
}

func resolveWhileStatement(st Token) (Statement, error) {
	if err := assertToken(st, WhileStatement, ""); err != nil {
		return nil, err
	}
	stat := whileStatement{}
	cond, err := resolveExpression(match(st, Expression)[0])
	if err != nil {
		return nil, err
	}
	stat.condition = cond
	sts, err := resolveStatements(match(st, Statements)[0])
	if err != nil {
		return nil, err
	}
	stat.statements = sts
	return stat, nil
}

func resolveReturnStatement(st Token) (Statement, error) {
	if err := assertToken(st, ReturnStatement, ""); err != nil {
		return nil, err
	}
	stat := retStatement{}
	it := NewTokenIterator(st.SubTokens())
	it.Next() //pop 'return'
	next := it.Next()

	if next.GetType() == Expression {
		exp, err := resolveExpression(next)
		if err != nil {
			return nil, err
		}
		stat.expression = exp
	}
	return stat, nil
}

func resolveDoStatement(st Token) (Statement, error) {
	if err := assertToken(st, DoStatement, ""); err != nil {
		return nil, err
	}
	it := NewTokenIterator(st.SubTokens())
	if err := assertToken(it.Next(), Keyword, "do"); err != nil {
		return nil, err
	}

	target := it.Next()
	if err := assertToken(target, Identifier, ""); err != nil {
		return nil, err
	}

	var name string
	next := it.Peek()
	if next.GetType() == Symbol && next.GetVal() == "." {
		it.Next() //pop '.'
		callee := it.Next()
		if err := assertToken(callee, Identifier, ""); err != nil {
			return nil, err
		}
		name = callee.GetVal()
	}
	it.Next()           //pop (
	params := it.Next() //expression list
	if err := assertToken(params, ExpressionList, ""); err != nil {
		return nil, err
	}
	subcall, err := resolveSubcall(st, target.GetVal(), name)
	if err != nil {
		return nil, err
	}

	if len(subcall.name) == 0 {
		subcall.target = ""
		subcall.name = target.GetVal()
	}
	return doStatement{
		action: subcall,
	}, nil

}

func resolveLetStatement(st Token) (Statement, error) {
	if err := assertToken(st, LetStatement, ""); err != nil {
		return nil, err
	}
	stat := letStatement{}
	it := NewTokenIterator(st.SubTokens())
	it.Next() //let
	varName := it.Next().GetVal()
	if it.Peek().GetVal() == "[" { //ref to array
		it.Next() //pop [
		exp, err := resolveExpression(it.Next())
		if err != nil {
			return nil, err
		}
		stat.target = ReferenceTerm{varName: varName, index: exp}
	} else {
		stat.target = ReferenceTerm{varName: varName}
	}
	for it.HasNext() {
		next := it.Next()
		if next.GetVal() == "=" && next.GetType() == Symbol {
			exp, err := resolveExpression(it.Next())
			if err != nil {
				return nil, err
			}
			stat.expression = exp
			break
		}
	}
	return stat, nil
}

func resolveIfStatement(st Token) (Statement, error) {
	is := ifStatement{}

	it := NewTokenIterator(st.SubTokens())
	for it.HasNext() {
		sub := it.Next()
		if sub.GetType() == Keyword {
			if sub.GetVal() == "if" {
				for it.HasNext() {
					if it.Peek().GetType() == Expression {
						con, err := resolveExpression(it.Next())
						if err != nil {
							return nil, err
						}
						is.condition = con
						continue
					} else if it.Peek().GetType() == Statements {
						statements, err := resolveStatements(it.Next())
						if err != nil {
							return nil, err
						}
						is.statements = statements
						break
					}
					it.Next()
				}
			} else if sub.GetVal() == "else" {
				it.Next() //pop {
				elseStatements, err := resolveStatements(it.Next())
				if err != nil {
					return nil, err
				}
				is.elseStatements = elseStatements
				break
			}
		}
	}
	return is, nil
}

var emptyExpression = expression{}

func resolveExpression(token Token) (expression, error) {
	exp := expression{}
	if err := assertToken(token, Expression, ""); err != nil {
		return emptyExpression, err
	}
	it := NewTokenIterator(token.SubTokens())

	term, err := resolveTerm(it.Next())
	if err != nil {
		return emptyExpression, err
	}
	exp.terms = append(exp.terms, term)
	for it.HasNext() {
		st := it.Next()
		if st.GetType() == Symbol && operations[st.GetVal()] != "" {
			exp.operations = append(exp.operations, st.GetVal())
			term, err := resolveTerm(it.Next())
			if err != nil {
				return emptyExpression, err
			}
			exp.terms = append(exp.terms, term)
		}
	}
	return exp, nil
}

var emptyTerm = ConstTerm{}

func resolveTerm(token Token) (term Term, err error) {
	it := NewTokenIterator(token.SubTokens())
	for it.HasNext() {
		t := it.Next()
		switch t.GetType() {
		case IntegerConstant:
			iv, err := strconv.Atoi(t.GetVal())
			if err != nil {
				return nil, newGrammarError(t, "int required")
			}
			return ConstTerm{ttype: IntegerConstant, val: iv}, nil
		case StringConstant:
			return ConstTerm{ttype: StringConstant, val: t.GetVal()}, nil
		case Keyword:
			if ContainsString(keywordConstants, t.GetVal()) {
				return ConstTerm{ttype: Keyword, val: t.GetVal()}, nil
			}
		case Symbol:
			op := t.GetVal()
			if op == "-" || op == "~" {
				next := it.Next()
				if err = assertToken(next, TokenTerm, ""); err != nil {
					return nil, err
				}
				term, err = resolveTerm(next)
				if err == nil {
					term = UnaryTerm{operator: op, term: term}
				}
				return term, nil
			} else if op == "(" {
				expression, err := resolveExpression(it.Next())
				if err != nil {
					return emptyExpression, err
				}
				return expression, nil
			}
		case Identifier:
			target := t.GetVal()
			if !it.HasNext() {
				return ReferenceTerm{varName: target}, nil
			}
			next := it.Next()
			if next.GetType() == Symbol {
				if next.GetVal() == "." {
					subcallee := it.Next()
					if err := assertToken(subcallee, Identifier, ""); err != nil {
						return emptySubCall, err
					}
					subcall, err := resolveSubcall(token, target, subcallee.GetVal())
					if err != nil {
						return emptyTerm, err
					}
					return subcall, nil
				} else if next.GetVal() == "(" {
					subcall, err := resolveSubcall(token, "", target)
					if err != nil {
						return emptyTerm, err
					}
					return subcall, nil
				} else if next.GetVal() == "[" {
					exp, err := resolveExpression(it.Next())
					if err != nil {
						return emptyTerm, err
					}
					return ReferenceTerm{varName: target, index: exp}, nil
				}
			}
		}
	}
	return nil, newSyntaxError(token)
}

var emptySubCall = subroutineCall{}

func resolveSubcall(parent Token, target string, name string) (subroutineCall, error) {
	args, err := resolveExpressionList(match(parent, ExpressionList)[0])
	if err != nil {
		return emptySubCall, err
	}
	return subroutineCall{
		target: target,
		name:   name,
		args:   args,
	}, nil
}

func resolveExpressionList(ts Token) ([]expression, error) {
	es := match(ts, Expression)
	var exps []expression
	for _, e := range es {
		exp, err := resolveExpression(e)
		if err != nil {
			return nil, err
		}
		exps = append(exps, exp)
	}
	return exps, nil
}

/*
	input:
		<parameterList>
            <identifier>string</identifier>
            <identifier>x</identifier>
            <symbol>,</symbol>
            <keyword>int</keyword>
            <identifier>y</identifier>
        </parameterList>
*/
func resolveParams(params Token) []variable {
	var vs []variable
	it := NewTokenIterator(params.SubTokens())
	for it.HasNext() {
		if it.Peek().GetVal() == "," {
			it.Next()
		}
		vs = append(vs, variable{
			typ:  vType(it.Next().GetVal()),
			name: it.Next().GetVal(),
			kind: kargument,
		})
	}
	return vs
}

func resolveClassVarDecs(root *NonTerminalToken) ([]variable, error) {
	var vars []variable
	for _, item := range root.subTokens {
		if item.GetType() == ClassVarDec {
			vars = append(vars, resolveClassVarDec(item)...)
		}
	}
	return vars, nil
}

func resolveClassVarDec(item Token) []variable {
	var vs []variable
	kind := item.SubTokens()[0]
	typ := item.SubTokens()[1]
	it := NewTokenIterator(item.SubTokens()[2:])
	for it.HasNext() {
		n := it.Next()
		if n.GetType() == Identifier {
			vs = append(vs, variable{
				name: n.GetVal(),
				kind: vKind(kind.GetVal()),
				typ:  vType(typ.GetVal()),
			})
		}
	}
	return vs
}

func resolveClassName(root *NonTerminalToken) (string, error) {
	if err := assertToken(root, Class, ""); err != nil {
		return "", err
	}
	subs := root.subTokens
	if err := assertToken(subs[0], Keyword, "class"); err != nil {
		return "", err
	}
	if err := assertToken(subs[1], Identifier, ""); err != nil {
		return "", err
	}
	return subs[1].GetVal(), nil
}
