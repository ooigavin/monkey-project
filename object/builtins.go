package object

import "fmt"

// Builtins is a slice of structs, Name: Builtin fn
// Slice is used to allow a stable iteration
// Name is used to identify the fn
var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{"len", &Builtin{Fn: lenBn}},
	{"print", &Builtin{Fn: printBn}},
}

// GetBuiltinByName finds a builtin func from its name
// Returns nil if none is found
func GetBuiltinByName(name string) *Builtin {
	for _, def := range Builtins {
		if def.Name == name {
			return def.Builtin
		}
	}
	return nil
}

func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func lenBn(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	switch arg := args[0].(type) {
	case *String:
		return &Integer{Value: int64(len(arg.Value))}
	case *Array:
		return &Integer{Value: int64(len(arg.Elements))}
	default:
		return newError("argument to `len` not supported, got %s", args[0].Type())
	}
}

func printBn(args ...Object) Object {
	for _, arg := range args {
		fmt.Println(arg.Inspect())
	}
	return nil
}

func handleArr(expectedArgs int, args ...Object) (*Array, Object) {
	if len(args) != expectedArgs {
		return nil, newError("wrong number of arguments. got=%d, want=%d", len(args), expectedArgs)
	}
	arr, ok := args[0].(*Array)
	if !ok {
		return nil, newError("method expected an array, got %T", args[0])
	}
	return arr, nil
}

func firstBn(args ...Object) Object {
	if arr, errObj := handleArr(1, args...); errObj != nil {
		return errObj
	} else if len(arr.Elements) > 0 {
		return arr.Elements[0]
	}
	return nil
}

func lastBn(args ...Object) Object {
	if arr, errObj := handleArr(1, args...); errObj != nil {
		return errObj
	} else if len(arr.Elements) > 0 {
		return arr.Elements[len(arr.Elements)-1]
	}
	return nil
}

func restBn(args ...Object) Object {
	if arr, errObj := handleArr(1, args...); errObj != nil {
		return errObj
	} else if length := len(arr.Elements); length > 0 {
		newArr := make([]Object, length-1, length-1)
		copy(newArr, arr.Elements[1:length])
		return &Array{Elements: newArr}
	}
	return nil
}

func pushBn(args ...Object) Object {
	if arr, errObj := handleArr(2, args...); errObj != nil {
		return errObj
	} else if length := len(arr.Elements); length > 0 {
		newArr := make([]Object, length+1, length+1)
		copy(newArr, arr.Elements)
		newArr[length] = args[1]
		return &Array{Elements: newArr}
	}
	return nil
}
