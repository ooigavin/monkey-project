package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"monkey/ast"
	"monkey/code"
	"strings"
)

type ObjectType string

const (
	INTEGER_OBJ       = "INTEGER"
	STRING_OBJ        = "STRING"
	ARRAY_OBJ         = "ARRAY"
	BOOLEAN_OBJ       = "BOOLEAN"
	NULL_OBJ          = "NULL"
	RETURN_VALUE_OBJ  = "RETURN_VALUE"
	ERROR_OBJ         = "ERROR"
	FUNC_OBJ          = "FUNC"
	COMPILED_FUNC_OBJ = "COMPILED_FUNC"
	BUILTIN_OBJ       = "BUILTIN"
	HASH_OBJ          = "HASH"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Hashable interface {
	Hash() HashKey
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

type Environment struct {
	store map[string]Object
	outer *Environment
}

type BuiltinFunction func(args ...Object) Object

func NewEnclosedEnvironment(outer *Environment) *Environment {
	new := NewEnv()
	new.outer = outer
	return new
}

func NewEnv() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s}
}

func (e *Environment) Get(k string) (Object, bool) {
	v, ok := e.store[k]
	if !ok && e.outer != nil {
		return e.outer.Get(k)
	}
	return v, ok
}

func (e *Environment) Set(k string, v Object) Object {
	e.store[k] = v
	return v
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Hash() HashKey {
	return HashKey{Type: INTEGER_OBJ, Value: uint64(i.Value)}
}

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }
func (s *String) Hash() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: STRING_OBJ, Value: h.Sum64()}
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%v", b.Value) }
func (b *Boolean) Hash() HashKey {
	if val := b.Value; val {
		return HashKey{Type: BOOLEAN_OBJ, Value: 1}
	} else {
		return HashKey{Type: BOOLEAN_OBJ, Value: 0}
	}
}

type Array struct {
	Elements []Object
}

func (arr *Array) Type() ObjectType { return ARRAY_OBJ }
func (arr *Array) Inspect() string {
	var out bytes.Buffer
	elements := []string{}
	for _, el := range arr.Elements {
		elements = append(elements, el.Inspect())
	}
	out.WriteString(fmt.Sprintf("[%s]", strings.Join(elements, ", ")))
	return out.String()
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) Inspect() string {
	var out bytes.Buffer
	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}
	out.WriteString(fmt.Sprintf("{%s}", strings.Join(pairs, ", ")))
	return out.String()
}

type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "Error: " + e.Message }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNC_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer
	params := []string{}
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

type CompiledFunction struct {
	Instructions code.Instructions
	NumLocals    int
	NumArgs      int
}

func (cf *CompiledFunction) Type() ObjectType { return COMPILED_FUNC_OBJ }
func (cf *CompiledFunction) Inspect() string  { return fmt.Sprintf("CompiledFunction[%p]", cf) }

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }
