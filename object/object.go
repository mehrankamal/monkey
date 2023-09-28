package object

import "fmt"

type Type string

type Object interface {
	Type() Type
	Inspect() string
}

const (
	INTEGER Type = "INTEGER"
	BOOLEAN      = "BOOLEAN"
	NULL         = "NULL"
)

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() Type      { return INTEGER }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() Type      { return BOOLEAN }
func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }

type Null struct{}

func (n *Null) Type() Type      { return NULL }
func (n *Null) Inspect() string { return "null" }
