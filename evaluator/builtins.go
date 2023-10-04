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

func init() {
	registerBuiltin(&builtins, "len", builtinLen)
	registerBuiltin(&builtins, "first", builtinFirst)
}

func registerBuiltin(store *map[string]*object.Builtin, name string, function object.BuiltinFunction) bool {
	_, ok := (*store)[name]
	if ok {
		return false
	}

	(*store)[name] = &object.Builtin{Fn: function}
	return true
}
