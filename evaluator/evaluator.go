package evaluator

import (
	"monkey/ast"
	"monkey/object"
)

func Eval(node ast.Node) object.Object {
	switch nodeType := node.(type) {
	case *ast.Program:
		return evalStatements(nodeType.Statements)
	case *ast.ExpressionStatement:
		return Eval(nodeType.Expression)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: nodeType.Value}
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
