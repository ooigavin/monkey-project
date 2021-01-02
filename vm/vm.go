package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

const (
	StackSize  = 2048
	GlobalSize = 65536
	MaxFrames  = 1024
)

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

type Frame struct {
	fn          *object.CompiledFunction
	ip          int
	basePointer int // refers to the location in the stack before the func frame was pushed
}

func NewFrame(fn *object.CompiledFunction, basePointer int) *Frame {
	return &Frame{
		fn:          fn,
		ip:          -1,
		basePointer: basePointer,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}

type VM struct {
	constants []object.Object

	stack   []object.Object
	sp      int             // points to the next value in the stack
	globals []object.Object // slice to store all global objects

	frames      []*Frame
	framesIndex int
}

func New(bc *compiler.Bytecode) *VM {
	mainFrame := NewFrame(&object.CompiledFunction{Instructions: bc.Instructions}, 0)
	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants:   bc.Constants,
		stack:       make([]object.Object, StackSize),
		sp:          0, // need to -1 when access objects from stack
		globals:     make([]object.Object, GlobalSize),
		frames:      frames,
		framesIndex: 1,
	}
}

func NewWithState(bc *compiler.Bytecode, globals []object.Object) *VM {
	vm := New(bc)
	vm.globals = globals
	return vm
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func (vm *VM) LastPoppedElem() object.Object {
	return vm.stack[vm.sp]
}

// Run iterates thru the slice of bytecode instructions and executes them
func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			if err := vm.push(vm.constants[constIndex]); err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			if err := vm.executeBinaryOperation(op); err != nil {
				return err
			}
		case code.OpGreaterThan, code.OpEqual, code.OpNotEqual:
			if err := vm.executeCompairison(op); err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		case code.OpNull:
			if err := vm.push(Null); err != nil {
				return err
			}
		case code.OpTrue:
			if err := vm.push(True); err != nil {
				return err
			}
		case code.OpFalse:
			if err := vm.push(False); err != nil {
				return err
			}
		case code.OpMinus:
			if err := vm.executeMinus(); err != nil {
				return err
			}
		case code.OpBang:
			if err := vm.executeBang(); err != nil {
				return err
			}
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			// pos is the position you want to jump to, but you need to decrement by 1 as the loop will increment at the end
			if !isTruthy(vm.pop()) {
				vm.currentFrame().ip = pos - 1
			}
		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1
		case code.OpSetGlobal:
			i := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.globals[i] = vm.pop()
		case code.OpGetGlobal:
			i := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			if err := vm.push(vm.globals[i]); err != nil {
				return err
			}
		case code.OpSetLocal:
			// get the index of the symbol
			i := int(code.ReadUint8(ins[ip+1:]))
			// increment the instruction pointer by one as the operand size is one byte
			vm.currentFrame().ip++
			// pop the value off the stack
			// save it to the preallocated space on the stack for the current frames local bindings
			vm.stack[vm.currentFrame().basePointer+i] = vm.pop()
		case code.OpGetLocal:
			i := int(code.ReadUint8(ins[ip+1:]))
			vm.currentFrame().ip++
			if err := vm.push(vm.stack[vm.currentFrame().basePointer+i]); err != nil {
				return err
			}
		case code.OpArray:
			noElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			arr := vm.buildArray(noElements)
			if err := vm.push(arr); err != nil {
				return err
			}
		case code.OpHash:
			noElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			hash, err := vm.buildHash(noElements)
			if err != nil {
				return err
			}
			if err := vm.push(hash); err != nil {
				return err
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()
			if err := vm.executeIndexExpression(left, index); err != nil {
				return err
			}
		case code.OpCall:
			fn, ok := vm.stack[vm.sp-1].(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("Not a callable function")
			}

			newFrame := NewFrame(fn, vm.sp)
			vm.pushFrame(newFrame)
			// allocate space for local bindings in the stack
			vm.sp += fn.NumLocals
		case code.OpReturnValue:
			// get the return value from current frame
			retVal := vm.pop()

			// pop off the current frame
			poppedFrame := vm.popFrame()
			// reset stack pointer to before call frame
			// -1 to pop off the compiled func as well
			vm.sp = poppedFrame.basePointer - 1

			// push retval to stack
			if err := vm.push(retVal); err != nil {
				return err
			}

		case code.OpReturn:
			poppedFrame := vm.popFrame()
			vm.sp = poppedFrame.basePointer - 1

			if err := vm.push(Null); err != nil {
				return err
			}
		}
	}
	return nil
}

func (vm *VM) push(obj object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stackoverflow.com")
	}

	vm.stack[vm.sp] = obj
	vm.sp++
	return nil
}

func (vm *VM) pop() object.Object {
	obj := vm.stack[vm.sp-1]
	vm.sp--
	return obj
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	rightType := right.Type()
	left := vm.pop()
	leftType := left.Type()

	if rightType == object.INTEGER_OBJ && leftType == object.INTEGER_OBJ {
		return vm.executeBinaryIntegerOperation(op, left, right)
	}
	if rightType == object.STRING_OBJ && leftType == object.STRING_OBJ {
		return vm.executeBinaryStringOperation(op, left, right)
	}
	return fmt.Errorf("unsupported types for binary operation: %s, %s", leftType, rightType)
}

func (vm *VM) executeBinaryIntegerOperation(op code.Opcode, left object.Object, right object.Object) error {
	rightValue := right.(*object.Integer).Value
	leftValue := left.(*object.Integer).Value

	switch op {
	case code.OpAdd:
		return vm.push(&object.Integer{Value: leftValue + rightValue})
	case code.OpSub:
		return vm.push(&object.Integer{Value: leftValue - rightValue})
	case code.OpMul:
		return vm.push(&object.Integer{Value: leftValue * rightValue})
	case code.OpDiv:
		return vm.push(&object.Integer{Value: leftValue / rightValue})
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}
}

func (vm *VM) executeBinaryStringOperation(op code.Opcode, left object.Object, right object.Object) error {
	rightValue := right.(*object.String).Value
	leftValue := left.(*object.String).Value
	if op == code.OpAdd {
		return vm.push(&object.String{Value: leftValue + rightValue})
	}
	return fmt.Errorf("unknown integer operator: %d", op)
}

func (vm *VM) executeCompairison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if right.Type() == object.INTEGER_OBJ && left.Type() == object.INTEGER_OBJ {
		return vm.compareIntegers(left, right, op)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToObject(left == right))
	case code.OpNotEqual:
		return vm.push(nativeBoolToObject(left != right))
	default:
		return fmt.Errorf("unknown operator: %d, %s %s", op, left.Type(), right.Type())
	}
}

func (vm *VM) compareIntegers(left object.Object, right object.Object, op code.Opcode) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch op {
	case code.OpGreaterThan:
		return vm.push(nativeBoolToObject(leftVal > rightVal))
	case code.OpEqual:
		return vm.push(nativeBoolToObject(leftVal == rightVal))
	case code.OpNotEqual:
		return vm.push(nativeBoolToObject(leftVal != rightVal))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func nativeBoolToObject(nativeBool bool) object.Object {
	if nativeBool {
		return True
	}
	return False
}

func (vm *VM) executeMinus() error {
	obj := vm.pop()
	if obj.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type for negation: %s", obj.Type())
	}

	val := obj.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -val})
}

func (vm *VM) executeBang() error {
	obj := vm.pop()

	if isTruthy(obj) {
		return vm.push(False)
	}
	return vm.push(True)
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case True:
		return true
	case False, Null:
		return false
	default:
		return true
	}
}

func (vm *VM) buildArray(noElements int) object.Object {
	start, end := vm.sp-noElements, vm.sp
	elements := make([]object.Object, noElements)

	for i := start; i < end; i++ {
		elements[i-start] = vm.stack[i]
	}
	vm.sp = start
	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(noElements int) (object.Object, error) {
	start, end := vm.sp-noElements, vm.sp
	pairs := make(map[object.HashKey]object.HashPair)

	for i := start; i < end; i += 2 {
		key := vm.stack[i]
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}
		val := vm.stack[i+1]
		pairs[hashKey.Hash()] = object.HashPair{Key: key, Value: val}
	}
	vm.sp = start
	return &object.Hash{Pairs: pairs}, nil
}

func (vm *VM) executeIndexExpression(left object.Object, index object.Object) error {
	switch left.Type() {
	case object.ARRAY_OBJ:
		return vm.executeArrayIndex(left, index)
	case object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("Unsupported index operation %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(left object.Object, index object.Object) error {
	// convert to array
	arr := left.(*object.Array)
	indexObj, ok := index.(*object.Integer)
	if !ok {
		return fmt.Errorf("Invalid integer")
	}

	idx := int(indexObj.Value)
	if idx >= len(arr.Elements) || idx < 0 {
		return vm.push(Null)
	}
	return vm.push(arr.Elements[idx])
}

func (vm *VM) executeHashIndex(left object.Object, index object.Object) error {
	// check if it can be hashed
	hash := left.(*object.Hash)
	hashObj, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("object %s is not hashable", left.Type())
	}
	// check for index error
	pair, ok := hash.Pairs[hashObj.Hash()]
	if !ok {
		return vm.push(Null)
	}
	// push obj to stack
	return vm.push(pair.Value)
}
