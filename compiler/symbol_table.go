package compiler

type SymbolScope string

const (
	GlobalScope   SymbolScope = "GLOBAL"
	LocalScope    SymbolScope = "LOCAL"
	BuiltinScope  SymbolScope = "BUILTIN"
	FreeScope     SymbolScope = "FREE"
	FunctionScope SymbolScope = "FUNCTION"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	Outer       *SymbolTable
	FreeSymbols []Symbol
	store       map[string]Symbol
	numDef      int
}

// NewSymbolTable creates a new symbol table & returns its pointer
func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	free := []Symbol{}
	return &SymbolTable{store: s, FreeSymbols: free}
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

func (st *SymbolTable) DefineBuiltin(i int, name string) Symbol {
	s := Symbol{Name: name, Index: i, Scope: BuiltinScope}
	st.store[name] = s
	return s
}

func (st *SymbolTable) DefineFunctionName(name string) Symbol {
	// index is always 0 as thr will only be one symbol in this scope
	s := Symbol{Name: name, Index: 0, Scope: FunctionScope}
	st.store[name] = s
	return s
}

func (st *SymbolTable) defineFree(original Symbol) Symbol {
	st.FreeSymbols = append(st.FreeSymbols, original)
	s := Symbol{Name: original.Name, Index: len(st.FreeSymbols) - 1, Scope: FreeScope}
	st.store[original.Name] = s
	return s
}

func (st *SymbolTable) Resolve(id string) (Symbol, bool) {
	sym, ok := st.store[id]
	if !ok && st.Outer != nil {
		// recursively resolve the symbol
		sym, ok = st.Outer.Resolve(id)
		// if unable to resolve OR the scope of the symbol is global or builtin
		// return the symbol & the bool
		if !ok || sym.Scope == GlobalScope || sym.Scope == BuiltinScope {
			return sym, ok
		}
		// else the symbol is a free variable, define it and return it as such
		free := st.defineFree(sym)
		return free, true
	}
	// return the symbol & bool if scope has no enclosing scope
	return sym, ok
}
