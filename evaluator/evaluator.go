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
		return evalProgram(node.Statements)
	case *ast.BlockStatement:
		return evalBlockStatements(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.ReturnStatement:
		obj := Eval(node.ReturnValue)
		return &object.ReturnValue{Value: obj}
	// EXPRESSIONS
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left)
		right := Eval(node.Right)
		return evalInfixExpression(node.Operator, left, right)
	case *ast.IfExpression:
		return evalIfExpression(node)
	}
	return nil
}

func evalProgram(ss []ast.Statement) object.Object {
	var obj object.Object
	for _, stmt := range ss {
		obj = Eval(stmt)
		if returnObj, ok := obj.(*object.ReturnValue); ok {
			return returnObj.Value
		}
	}
	return obj
}

func evalBlockStatements(ss []ast.Statement) object.Object {
	var obj object.Object
	for _, stmt := range ss {
		obj = Eval(stmt)
		if obj != nil && obj.Type() == object.RETURN_VALUE_OBJ {
			return obj
		}
	}
	return obj
}

func nativeBoolToObject(b bool) object.Object {
	if b {
		return TRUE
	}
	return FALSE
}

func evalPrefixExpression(op string, right object.Object) object.Object {
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

func evalInfixExpression(op string, left object.Object, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(op, left, right)
	case op == "==":
		return nativeBoolToObject(left == right)
	case op == "!=":
		return nativeBoolToObject(left != right)
	default:
		return NULL
	}
}

func evalIntegerInfixExpression(op string, left object.Object, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch op {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case ">":
		return nativeBoolToObject(leftVal > rightVal)
	case ">=":
		return nativeBoolToObject(leftVal >= rightVal)
	case "<":
		return nativeBoolToObject(leftVal < rightVal)
	case "<=":
		return nativeBoolToObject(leftVal <= rightVal)
	case "==":
		return nativeBoolToObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToObject(leftVal != rightVal)
	default:
		return NULL
	}
}

func evalIfExpression(ie *ast.IfExpression) object.Object {
	cond := Eval(ie.Condition)
	// if cond is truthy eval consequence
	if isTruthy(cond) {
		return Eval(ie.Consequence)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative)
	} else {
		return NULL
	}
}

func isTruthy(condition object.Object) bool {
	switch condition {
	case NULL:
		return false
	case FALSE:
		return false
	default:
		return true
	}
}
