package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte

func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "Error: %s\n", err)
			continue
		}
		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))
		i += 1 + read
	}
	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("Error: operand len %d does not match defined %d\n", len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	}
	return fmt.Sprintf("Error: unhandled operand count for %s", def.Name)
}

type Opcode byte

const (
	OpConstant Opcode = iota
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpPop
	OpTrue
	OpFalse
	OpGreaterThan
	OpEqual
	OpNotEqual
	OpMinus
	OpBang
	OpJump
	OpJumpNotTruthy
	OpNull
	OpSetGlobal
	OpGetGlobal
	OpSetLocal
	OpGetLocal
	OpArray
	OpHash
	OpIndex
	OpCall
	OpReturn
	OpReturnValue
	OpGetBuiltin
	OpClosure
	OpGetFree
	OpCurrentClosure
)

// Definition defines the structure of an opcode.
// holds the name as well as a slice of variable args the opcode will take
// each slice represents the size of the arg in bytes
type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	// pushes a constant from the slice of constants onto the stack
	// the arg is the index of the constant on the slice
	OpConstant: {"OpConstant", []int{2}},
	// infix binary operators takes no args
	// pop the next 2 values off the stack and perform the operations
	OpAdd:         {"OpAdd", []int{}},
	OpSub:         {"OpSub", []int{}},
	OpMul:         {"OpMul", []int{}},
	OpDiv:         {"OpDiv", []int{}},
	OpPop:         {"OpPop", []int{}},
	OpTrue:        {"OpTrue", []int{}},
	OpFalse:       {"OpFalse", []int{}},
	OpNull:        {"OpNull", []int{}},
	OpGreaterThan: {"OpGreaterThan", []int{}},
	OpEqual:       {"OpEqual", []int{}},
	OpNotEqual:    {"OpNotEqual", []int{}},
	// prefix operators
	// pops the next value off the stack and performs its operation
	OpMinus: {"OpMinus", []int{}},
	OpBang:  {"OpBang", []int{}},
	// jump operators, used for conditionals to move instruction pointer
	OpJump:          {"OpJump", []int{2}},
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}},
	// pop the last item off the stack and assign it to the globals slice at index i (arg)
	OpSetGlobal: {"OpSetGlobal", []int{2}},
	// arg is the index of the global variable to fetch
	// gets the obj at index i and pushes it onto the stack
	OpGetGlobal: {"OpGetGlobal", []int{2}},
	OpSetLocal:  {"OpSetLocal", []int{1}},
	OpGetLocal:  {"OpGetLocal", []int{1}},
	// arg represents the no of elements in the array
	// OpArray builds the arr object from the no of elements & pushes it onto the stack
	OpArray: {"OpArray", []int{2}},
	// ophash does the same as oparray but doubles its values for the key-val pairings
	OpHash: {"OpHash", []int{2}},
	// pushes a null object onto the stack
	OpIndex: {"OpNull", []int{}},
	// returns the no of args a function call has
	OpCall: {"OpCall", []int{1}},
	// pops off the current frame and pushes null onto the stack
	OpReturn: {"OpReturn", []int{}},
	// gets the last value from the current frame then pops off the frame
	OpReturnValue: {"OpReturnValue", []int{}},
	OpGetBuiltin:  {"OpGetBuiltin", []int{1}},
	// Opclosure has 2 args, first arg is 2 bytes wide index that points to the location of the compiled fn in the constant stack
	// the 2nd arg is the no of free variables used in this closure
	// these free variables will hv to be pushed onto the stack before hand
	OpClosure: {"OpClosure", []int{2, 1}},
	// gets the free variable from the current closure and pushes it onto the stack
	OpGetFree: {"OpGetFree", []int{1}},
	// pushes current closure onto the stack to allow for recursion
	OpCurrentClosure: {"OpCurrentClosure", []int{}},
}

func Lookup(op byte) (*Definition, error) {
	if def, ok := definitions[Opcode(op)]; ok {
		return def, nil
	}
	return nil, fmt.Errorf("opcode %d undefined", op)
}

// Make takes an opcode & slice of operands and returns bytecode instructions
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case 1:
			instruction[offset] = byte(o)
		}
		offset += width
	}

	return instruction
}

// ReadOperands takes a pointer to a opcode definition and a slice of bytes, returning a slice of operands and the no of operands read
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		case 1:
			operands[i] = int(ReadUint8(ins[offset:]))
		}
		offset += width
	}
	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}
