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
)

// Definition defines the structure of an opcode.
// holds the name as well as a slice of variable args the opcode will take
// each slice represents the size of the arg in bytes
type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpConstant:      {"OpConstant", []int{2}},
	OpAdd:           {"OpAdd", []int{}},
	OpSub:           {"OpSub", []int{}},
	OpMul:           {"OpMul", []int{}},
	OpDiv:           {"OpDiv", []int{}},
	OpPop:           {"OpPop", []int{}},
	OpTrue:          {"OpTrue", []int{}},
	OpFalse:         {"OpFalse", []int{}},
	OpNull:          {"OpNull", []int{}},
	OpGreaterThan:   {"OpGreaterThan", []int{}},
	OpEqual:         {"OpEqual", []int{}},
	OpNotEqual:      {"OpNotEqual", []int{}},
	OpMinus:         {"OpMinus", []int{}},
	OpBang:          {"OpBang", []int{}},
	OpJump:          {"OpJump", []int{2}},
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}},
	OpSetGlobal:     {"OpSetGlobal", []int{2}},
	OpGetGlobal:     {"OpGetGlobal", []int{2}},
	OpSetLocal:      {"OpSetLocal", []int{1}},
	OpGetLocal:      {"OpGetLocal", []int{1}},
	OpArray:         {"OpArray", []int{2}},
	OpHash:          {"OpHash", []int{2}},
	OpIndex:         {"OpNull", []int{}},
	// returns the no of args a function call has
	OpCall:        {"OpCall", []int{1}},
	OpReturn:      {"OpReturn", []int{}},
	OpReturnValue: {"OpReturnValue", []int{}},
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
