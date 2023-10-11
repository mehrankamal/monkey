package object

import (
	"bytes"
	"fmt"
	"github.com/mehrankamal/monkey/ast"
	"github.com/mehrankamal/monkey/code"
	"hash/fnv"
	"strings"
)

type Type string

type Object interface {
	Type() Type
	Inspect() string
}

type Hashable interface {
	HashKey() HashKey
}

const (
	INTEGER           Type = "INTEGER"
	BOOLEAN                = "BOOLEAN"
	NULL                   = "NULL"
	RETURN_VALUE           = "RETURN_VALUE"
	ERROR                  = "ERROR"
	FUNCTION               = "FUNCTION"
	STRING                 = "STRING"
	BUILTIN                = "BUILTIN"
	ARRAY                  = "ARRAY"
	HASH                   = "HASH"
	COMPILED_FUNCTION      = "COMPILED_FUNCTION"
)

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() Type      { return INTEGER }
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() Type      { return BOOLEAN }
func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) HashKey() HashKey {
	var hashValue uint64

	if b.Value {
		hashValue = 1
	} else {
		hashValue = 0
	}

	return HashKey{Type: b.Type(), Value: hashValue}
}

type Null struct{}

func (n *Null) Type() Type      { return NULL }
func (n *Null) Inspect() string { return "null" }

type CompiledFunction struct {
	Instructions  code.Instructions
	NumLocals     int
	NumParameters int
}

func (cf *CompiledFunction) Type() Type { return COMPILED_FUNCTION }
func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%p]", cf)
}

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() Type      { return RETURN_VALUE }
func (rv *ReturnValue) Inspect() string { return rv.Value.Inspect() }

type Error struct {
	Message string
}

func (e *Error) Type() Type      { return ERROR }
func (e *Error) Inspect() string { return "ERROR: " + e.Message }

type Environment struct {
	store map[string]Object
	outer *Environment
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)

	return &Environment{store: s, outer: nil}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() Type { return FUNCTION }
func (f *Function) Inspect() string {
	var out bytes.Buffer
	params := make([]string, 0)

	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")
	return out.String()
}

type String struct {
	Value string
}

func (s *String) Type() Type      { return STRING }
func (s *String) Inspect() string { return s.Value }
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))

	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() Type      { return BUILTIN }
func (b *Builtin) Inspect() string { return "builtin function" }

type Array struct {
	Elements []Object
}

func (ao *Array) Type() Type { return ARRAY }
func (ao *Array) Inspect() string {
	var out bytes.Buffer

	elements := make([]string, 0)
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type HashKey struct {
	Type  Type
	Value uint64
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() Type { return HASH }
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
