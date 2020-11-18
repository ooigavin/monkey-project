package evaluator

import (
	"monkey/ast"
	"monkey/object"
)

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
