package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	store  map[string]Symbol
	numDef int
}

// NewSymbolTable creates a new symbol table & returns its pointer
func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

// Define takes a name and maps it to the symbol table
func (st *SymbolTable) Define(name string) Symbol {
	s := Symbol{Name: name, Index: st.numDef, Scope: GlobalScope}
	st.store[name] = s
	st.numDef++
	return s
}

func (st *SymbolTable) Resolve(id string) (Symbol, bool) {
	sym, ok := st.store[id]
	return sym, ok
}
