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
	default:
		return newError("argument to `len` not supported, got %s", args[0].Type())
	}
}

func init() {
	registerBuiltin(&builtins, "len", builtinLen)
}

func registerBuiltin(store *map[string]*object.Builtin, name string, function object.BuiltinFunction) bool {
	_, ok := (*store)[name]
	if ok {
		return false
	}

	(*store)[name] = &object.Builtin{Fn: function}
	return true
}
