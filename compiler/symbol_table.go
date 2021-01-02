package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
	LocalScope  SymbolScope = "LOCAL"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	Outer  *SymbolTable
	store  map[string]Symbol
	numDef int
}

// NewSymbolTable creates a new symbol table & returns its pointer
func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

// NewEnclosedSymbolTable takes an Outer symbol table & creates a new enclosed table
func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	st := NewSymbolTable()
	st.Outer = outer

	return st
}

// Define takes a name and maps it to the symbol table.
// If symbol table's Outer value is not nil, scope is local
func (st *SymbolTable) Define(name string) Symbol {
	scope := GlobalScope
	if st.Outer != nil {
		scope = LocalScope
	}

	s := Symbol{Name: name, Index: st.numDef, Scope: scope}
	st.store[name] = s
	st.numDef++
	return s
}

func (st *SymbolTable) Resolve(id string) (Symbol, bool) {
	sym, ok := st.store[id]
	if !ok && st.Outer != nil {
		return st.Outer.Resolve(id)
	}
	return sym, ok
}
