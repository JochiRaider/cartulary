package sourcecheck

import "go/ast"

// bindingKind classifies the declared object an identifier use denotes.
type bindingKind uint8

const (
	unbound bindingKind = iota // functions, package-level, universe and package names; unresolved uses
	variableBinding
	constantBinding
	typeBinding
	otherBinding // labels and any other local object
)

// binding is a binder's answer for one identifier use: the kind of local
// declaration it denotes and, when the binder knows it, the declaring node
// (*ast.ValueSpec, *ast.AssignStmt, *ast.Field, *ast.TypeSpec, ...).
type binding struct {
	kind bindingKind
	decl ast.Node
}

func (b binding) bound() bool                   { return b.kind != unbound }
func (b binding) is(kind bindingKind) bool      { return b.kind == kind }
func (b binding) declaredBy(node ast.Node) bool { return b.bound() && b.decl == node }

// binder resolves identifier uses for one AnalyzeSQL batch. It is threaded
// explicitly through the analyzers and the type index, never held globally.
type binder interface {
	bind(*ast.Ident) binding
}
