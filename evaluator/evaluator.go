package evaluator

import (
	"fmt"
	"monkey/ast"
	"monkey/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	// STATEMENTS
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.BlockStatement:
		return evalBlockStatements(node.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.ReturnStatement:
		obj := Eval(node.ReturnValue, env)
		if isError(obj) {
			return obj
		}
		return &object.ReturnValue{Value: obj}
	case *ast.LetStatement:
		obj := Eval(node.Value, env)
		if isError(obj) {
			return obj
		}
		env.Set(node.Name.Value, obj)
	// EXPRESSIONS
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		right := Eval(node.Right, env)
		if isError(left) {
			return left
		}
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.FuncLiteral:
		return &object.Function{Env: env, Body: node.Body, Parameters: node.Parameters}
	case *ast.CallExpression:
		// find the function
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunc(function, args)
	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}
	case *ast.HashLiteral:
		return evalHashExpression(node, env)
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		index := Eval(node.Index, env)
		if isError(left) {
			return left
		}
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)
	}

	return nil
}

func nativeBoolToObject(b bool) object.Object {
	if b {
		return TRUE
	}
	return FALSE
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func applyFunc(obj object.Object, args []object.Object) object.Object {
	switch fn := obj.(type) {
	case *object.Function:
		extendedEnv := extendEnv(fn, args)
		returnVal := Eval(fn.Body, extendedEnv)
		return unwrapReturn(returnVal)
	case *object.Builtin:
		if res := fn.Fn(args...); res == nil {
			return NULL
		} else {
			return res
		}
	default:
		return newError("not a function: %s", obj.Type())
	}

}

func extendEnv(fn *object.Function, args []object.Object) *object.Environment {
	extendedEnv := object.NewEnclosedEnvironment(fn.Env)
	for paramIndex, param := range fn.Parameters {
		extendedEnv.Set(param.Value, args[paramIndex])
	}
	return extendedEnv
}

func unwrapReturn(obj object.Object) object.Object {
	if returnObj, ok := obj.(*object.ReturnValue); ok {
		return returnObj.Value
	}
	return obj
}
