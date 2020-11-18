package evaluator

import (
	"monkey/ast"
	"monkey/object"
)

var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalStatements(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToObject(node.Value)
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
