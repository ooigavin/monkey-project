package evaluator

import "monkey/object"

var builtins = map[string]*object.Builtin{
	"len":   &object.Builtin{Fn: lenBn},
	"first": &object.Builtin{Fn: firstBn},
	"last":  &object.Builtin{Fn: lastBn},
	"rest":  &object.Builtin{Fn: restBn},
	"push":  &object.Builtin{Fn: pushBn},
}

func lenBn(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	switch arg := args[0].(type) {
	case *object.String:
		return &object.Integer{Value: int64(len(arg.Value))}
	case *object.Array:
		return &object.Integer{Value: int64(len(arg.Elements))}
	default:
		return newError("argument to `len` not supported, got %s", args[0].Type())
	}
}

func handleArr(expectedArgs int, args ...object.Object) (*object.Array, object.Object) {
	if len(args) != expectedArgs {
		return nil, newError("wrong number of arguments. got=%d, want=%d", len(args), expectedArgs)
	}
	arr, ok := args[0].(*object.Array)
	if !ok {
		return nil, newError("method expected an array, got %T", args[0])
	}
	return arr, nil
}

func firstBn(args ...object.Object) object.Object {
	if arr, errObj := handleArr(1, args...); errObj != nil {
		return errObj
	} else if len(arr.Elements) > 0 {
		return arr.Elements[0]
	}
	return NULL
}

func lastBn(args ...object.Object) object.Object {
	if arr, errObj := handleArr(1, args...); errObj != nil {
		return errObj
	} else if len(arr.Elements) > 0 {
		return arr.Elements[len(arr.Elements)-1]
	}
	return NULL
}

func restBn(args ...object.Object) object.Object {
	if arr, errObj := handleArr(1, args...); errObj != nil {
		return errObj
	} else if length := len(arr.Elements); length > 0 {
		newArr := make([]object.Object, length-1, length-1)
		copy(newArr, arr.Elements[1:length])
		return &object.Array{Elements: newArr}
	}
	return NULL
}

func pushBn(args ...object.Object) object.Object {
	if arr, errObj := handleArr(2, args...); errObj != nil {
		return errObj
	} else if length := len(arr.Elements); length > 0 {
		newArr := make([]object.Object, length+1, length+1)
		copy(newArr, arr.Elements)
		newArr[length] = args[1]
		return &object.Array{Elements: newArr}
	}
	return NULL
}
