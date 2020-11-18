package evaluator

import (
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

func testEval(in string) object.Object {
	l := lexer.New(in)
	p := parser.New(l)
	program := p.ParseProgram()
	return Eval(program)
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("Object is not an integer, we got: %T, (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value, expected: %d, got: %d", expected, result.Value)
		return false
	}
	return true
}

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		in       string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.in)
		testIntegerObject(t, evaluated, tt.expected)
	}
}