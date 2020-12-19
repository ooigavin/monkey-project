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
)

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack   []object.Object
	sp      int             // points to the next value in the stack
	globals []object.Object // slice to store all global objects
}

func New(bc *compiler.Bytecode) *VM {
	return &VM{
		instructions: bc.Instructions,
		constants:    bc.Constants,
		stack:        make([]object.Object, StackSize),
		sp:           0,
		globals:      make([]object.Object, GlobalSize),
	}
}

func NewWithState(bc *compiler.Bytecode, globals []object.Object) *VM {
	vm := New(bc)
	vm.globals = globals
	return vm
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])
		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2
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
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2
			// pos is the position you want to jump to, but you need to decrement by 1 as the loop will increment at the end
			if !isTruthy(vm.pop()) {
				ip = pos - 1
			}
		case code.OpJump:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip = pos - 1
		case code.OpSetGlobal:
			i := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2
			vm.globals[i] = vm.pop()
		case code.OpGetGlobal:
			i := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2
			if err := vm.push(vm.globals[i]); err != nil {
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
