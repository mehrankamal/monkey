package evaluator

import "github.com/mehrankamal/monkey/object"

var builtins = make(map[string]*object.Builtin)

func builtinLen(args ...object.Object) object.Object {
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

func builtinFirst(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	switch arg := args[0].(type) {
	case *object.Array:
		return arrayFirst(arg)
	default:
		return newError("argument to `first` not supported, got %s", args[0].Type())
	}
}

func arrayFirst(arg *object.Array) object.Object {
	if len(arg.Elements) > 0 {
		return arg.Elements[0]
	}
	return NULL
}

func builtinLast(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	switch arg := args[0].(type) {
	case *object.Array:
		return lastArray(arg)
	default:
		return newError("argument to `last` not supported, got %s", args[0].Type())
	}
}

func lastArray(arg *object.Array) object.Object {
	if len(arg.Elements) > 0 {
		return arg.Elements[len(arg.Elements)-1]
	}
	return NULL
}

func init() {
	registerBuiltin(&builtins, "len", builtinLen)
	registerBuiltin(&builtins, "first", builtinFirst)
	registerBuiltin(&builtins, "last", builtinLast)
	registerBuiltin(&builtins, "rest", builtinRest)
	registerBuiltin(&builtins, "push", builtinPush)

}

func builtinPush(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}

	switch arg := args[0].(type) {
	case *object.Array:
		return arrayPush(arg, args[1])
	default:
		return newError("argument to `push` not supported, got %s", args[0].Type())
	}
}

func arrayPush(arr *object.Array, newElem object.Object) object.Object {
	length := len(arr.Elements)

	newElements := make([]object.Object, length+1)
	copy(newElements, arr.Elements)

	newElements[length] = newElem
	return &object.Array{Elements: newElements}
}

func builtinRest(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	switch arg := args[0].(type) {
	case *object.Array:
		return restArray(arg)
	default:
		return newError("argument to `last` not supported, got %s", args[0].Type())
	}
}

func restArray(arg *object.Array) object.Object {
	length := len(arg.Elements)
	if len(arg.Elements) > 0 {
		newElems := make([]object.Object, length-1)
		copy(newElems, arg.Elements[1:])

		return &object.Array{Elements: newElems}
	}

	return NULL
}

func registerBuiltin(store *map[string]*object.Builtin, name string, function object.BuiltinFunction) bool {
	_, ok := (*store)[name]
	if ok {
		return false
	}

	(*store)[name] = &object.Builtin{Fn: function}
	return true
}
