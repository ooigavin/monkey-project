package evaluator

import (
	"monkey/ast"
	"monkey/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	// STATEMENTS
	case *ast.Program:
		return evalStatements(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	// EXPRESSIONS
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		return evaluatePrefixExpression(node.Operator, right)
	}
	return nil
}

func evalStatements(ss []ast.Statement) object.Object {
	var obj object.Object
	for _, stmt := range ss {
		obj = Eval(stmt)
	}
	return obj
}

func nativeBoolToObject(b bool) object.Object {
	if b {
		return TRUE
	}
	return FALSE
}

func evaluatePrefixExpression(op string, right object.Object) object.Object {
	switch op {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusOperatorExpression(right)
	default:
		return NULL
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return NULL
	}

	intObj := right.(*object.Integer)
	negInt := &object.Integer{Value: -intObj.Value}

	return negInt
}
