package evaluator

import (
	"monkey/ast"
	"monkey/object"
)

func evalProgram(ss []ast.Statement) object.Object {
	var evalObj object.Object
	for _, stmt := range ss {
		evalObj = Eval(stmt)

		switch obj := evalObj.(type) {
		case *object.ReturnValue:
			return obj.Value
		case *object.Error:
			return obj
		}
	}
	return evalObj
}

func evalBlockStatements(ss []ast.Statement) object.Object {
	var obj object.Object
	for _, stmt := range ss {
		obj = Eval(stmt)
		rt := obj.Type()
		if obj != nil && rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
			return obj
		}
	}
	return obj
}
