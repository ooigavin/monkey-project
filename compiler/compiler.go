package compiler

import (
	"fmt"
	"monkey/ast"
	"monkey/code"
	"monkey/object"
	"sort"
)

type Compiler struct {
	constants   []object.Object
	scopeIndex  int
	scopes      []CompilationScope
	symbolTable *SymbolTable
}

type CompilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	st := NewSymbolTable()
	for i, builtin := range object.Builtins {
		st.DefineBuiltin(i, builtin.Name)
	}

	return &Compiler{
		constants:   []object.Object{},
		symbolTable: st,
		scopeIndex:  0,
		scopes:      []CompilationScope{mainScope},
	}
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	c := New()
	c.constants = constants
	c.symbolTable = s
	return c
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

// Compile recursively walks thru the ast and adds byte code instructions to be executed by the vm
func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		if err := c.Compile(node.Expression); err != nil {
			return err
		}
		c.emit(code.OpPop)
	case *ast.LetStatement:
		if err := c.Compile(node.Value); err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Name.Value)
		if c.symbolTable.Outer != nil {
			c.emit(code.OpSetLocal, symbol.Index)
		} else {
			c.emit(code.OpSetGlobal, symbol.Index)
		}
	case *ast.Identifier:
		if symbol, ok := c.symbolTable.Resolve(node.Value); ok {
			c.loadSymbol(symbol)
		} else {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
	case *ast.BlockStatement:
		for _, s := range node.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}
	case *ast.InfixExpression:
		if node.Operator == "<" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}
			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
			c.emit(code.OpGreaterThan)
			return nil
		}

		if err := c.Compile(node.Left); err != nil {
			return err
		}
		if err := c.Compile(node.Right); err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case ">":
			c.emit(code.OpGreaterThan)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		case "<":
			c.emit(code.OpGreaterThan)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.PrefixExpression:
		if err := c.Compile(node.Right); err != nil {
			return err
		}

		switch node.Operator {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		}
	case *ast.IfExpression:
		// compile the condition
		if err := c.Compile(node.Condition); err != nil {
			return err
		}
		// add the not truthy jump instruction
		notTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)
		// compile the consequence & remove the pop instruction if emitted
		if err := c.Compile(node.Consequence); err != nil {
			return err
		}
		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}
		// emit the jump instruction to the end of the if expression
		jumpPos := c.emit(code.OpJump, 9999)
		// you want to jump to the start of the alternative block statement, which is after the OpJump code
		posAfterConsequence := len(c.currentInstructions())
		c.changeOperand(notTruthyPos, posAfterConsequence)
		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			if err := c.Compile(node.Alternative); err != nil {
				return err
			}
			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			}
		}
		posAfterAlt := len(c.currentInstructions())
		c.changeOperand(jumpPos, posAfterAlt)
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))
	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	case *ast.ArrayLiteral:
		for _, el := range node.Elements {
			if err := c.Compile(el); err != nil {
				return nil
			}
		}
		c.emit(code.OpArray, len(node.Elements))
	case *ast.HashLiteral:
		keys := []ast.Expression{}
		for key := range node.Pairs {
			keys = append(keys, key)
		}

		sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
		for _, key := range keys {
			if err := c.Compile(key); err != nil {
				return err
			}
			if err := c.Compile(node.Pairs[key]); err != nil {
				return err
			}
		}
		c.emit(code.OpHash, len(keys)*2)
	case *ast.IndexExpression:
		if err := c.Compile(node.Left); err != nil {
			return err
		}
		if err := c.Compile(node.Index); err != nil {
			return err
		}
		c.emit(code.OpIndex)
	case *ast.FuncLiteral:
		c.enterScope()

		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value)
		}

		if err := c.Compile(node.Body); err != nil {
			return err
		}
		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}
		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}

		// ensure that these values are taken from the nested scope before leaving
		numLocals := c.symbolTable.numDef
		freeSyms := c.symbolTable.FreeSymbols
		instructions := c.leaveScope()

		for _, sym := range freeSyms {
			c.loadSymbol(sym)
		}

		// a compiled func is seen as an obj by the compiler & is emited as an OpConstant
		compiledFn := &object.CompiledFunction{
			Instructions: instructions,
			NumLocals:    numLocals,
			NumArgs:      len(node.Parameters),
		}
		c.emit(code.OpClosure, c.addConstant(compiledFn), len(freeSyms))
	case *ast.CallExpression:
		if err := c.Compile(node.Function); err != nil {
			return err
		}
		for _, a := range node.Arguments {
			if err := c.Compile(a); err != nil {
				return err
			}
		}
		c.emit(code.OpCall, len(node.Arguments))
	case *ast.ReturnStatement:
		if err := c.Compile(node.ReturnValue); err != nil {
			return err
		}
		c.emit(code.OpReturnValue)
	}
	return nil
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

// emit generates a bytecode instruction from the opcode & its operands
// returns the index of the instruction in the array of compiler instructions
func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	return pos
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	c.scopes[c.scopeIndex].previousInstruction = c.scopes[c.scopeIndex].lastInstruction
	c.scopes[c.scopeIndex].lastInstruction = EmittedInstruction{Opcode: op, Position: pos}
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.currentInstructions())
	instructions := append(c.currentInstructions(), ins...)
	c.scopes[c.scopeIndex].instructions = instructions
	return posNewInstruction
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	currScope := c.scopes[c.scopeIndex]
	instructions := currScope.instructions[:currScope.lastInstruction.Position]
	c.scopes[c.scopeIndex].instructions = instructions
	c.scopes[c.scopeIndex].lastInstruction = currScope.previousInstruction
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.currentInstructions()[opPos])
	newInstruction := code.Make(op, operand)
	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))
	c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.scopes[c.scopeIndex].instructions[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) enterScope() {
	st := NewEnclosedSymbolTable(c.symbolTable)
	c.symbolTable = st
	scope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
}

func (c *Compiler) leaveScope() code.Instructions {
	c.symbolTable = c.symbolTable.Outer
	instructions := c.scopes[c.scopeIndex].instructions
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--

	return instructions
}

func (c *Compiler) loadSymbol(sym Symbol) {
	// symbol could belong to outer or even the global scope
	// cannot just check if the current symboltable is global or local
	switch sym.Scope {
	case GlobalScope:
		c.emit(code.OpGetGlobal, sym.Index)
	case LocalScope:
		c.emit(code.OpGetLocal, sym.Index)
	case BuiltinScope:
		c.emit(code.OpGetBuiltin, sym.Index)
	case FreeScope:
		c.emit(code.OpGetFree, sym.Index)
	}
}
