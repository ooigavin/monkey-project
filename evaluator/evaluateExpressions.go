package evaluator

import (
	"monkey/ast"
	"monkey/object"
)

func evalPrefixExpression(op string, right object.Object) object.Object {
	switch op {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", op, right.Type())
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
		return newError("unknown operator: -%s", right.Type())
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
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), op, right.Type())
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(op, left, right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
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
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func evalStringInfixExpression(op string, left object.Object, right object.Object) object.Object {
	if op != "+" {
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

func evalIdentifier(id *ast.Identifier, env *object.Environment) object.Object {
	if obj, ok := env.Get(id.Value); ok {
		return obj
	}
	if builtin, ok := builtins[id.Value]; ok {
		return builtin
	}

	return newError("identifier not found: %s", id.Value)
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	cond := Eval(ie.Condition, env)
	if isError(cond) {
		return cond
	}
	// if cond is truthy eval consequence
	if isTruthy(cond) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
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

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var objs []object.Object
	for _, exp := range exps {
		obj := Eval(exp, env)
		if isError(obj) {
			return []object.Object{obj}
		}
		objs = append(objs, obj)
	}
	return objs
}

func evalIndexExpression(left object.Object, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(left object.Object, index object.Object) object.Object {
	arr := left.(*object.Array)
	idx := index.(*object.Integer).Value
	if 0 > idx || idx > int64(len(arr.Elements)-1) {
		return NULL
	}
	return arr.Elements[idx]
}

func evalHashExpression(node *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)
	for k, v := range node.Pairs {
		key := Eval(k, env)
		if isError(key) {
			return key
		}
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("type %s is not hashable", key.Type())
		}
		val := Eval(v, env)
		if isError(val) {
			return val
		}
		pairs[hashKey.Hash()] = object.HashPair{Key: key, Value: val}
	}
	return &object.Hash{Pairs: pairs}

}

func evalHashIndexExpression(left object.Object, index object.Object) object.Object {
	hash := left.(*object.Hash)
	idx, ok := index.(object.Hashable)
	if !ok {
		return newError("type %s is not hashable", index.Type())
	}
	if pair, ok := hash.Pairs[idx.Hash()]; ok {
		return pair.Value
	} else {
		return NULL
	}
}
