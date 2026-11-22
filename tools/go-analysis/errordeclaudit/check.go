package main

import (
	"go/constant"
	"slices"
	"sort"

	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

type writtenError struct {
	Operation string
	Status    int
	Code      string
	Writer    string
}

type writeCheck struct {
	Writes     []writtenError
	Violations []writtenError
}

// checkWrites accepts only a fresh map literal passed to the route binder and
// an unambiguous operation binding. The broader VTA audit has separate roots.
func checkWrites(config Config, binder string) (writeCheck, error) {
	if err := resolveProbes(config.Catalog); err != nil {
		return writeCheck{}, err
	}
	program, err := loadProgram(config)
	if err != nil {
		return writeCheck{}, err
	}
	functions := ssautil.AllFunctions(program)
	roots := handlerRoots(config, functions)
	result := writeCheck{}
	for fn := range functions {
		for _, block := range fn.Blocks {
			for _, instr := range block.Instrs {
				call, ok := instr.(*ssa.Call)
				if !ok || call.Call.StaticCallee() == nil || qualifiedFunction(call.Call.StaticCallee()) != binder || len(call.Call.Args) < 4 {
					continue
				}
				m, ok := call.Call.Args[3].(*ssa.MakeMap)
				if !ok || m.Block() != block {
					continue
				}
				bindings := literalHandlers(m, call)
				for operation, handler := range bindings {
					declaration, exists := config.Catalog[operation]
					if !exists || handler == nil || len(roots[operation]) != 1 || len(handler.Params) != 2 {
						continue
					}
					written, ok := firstResponseWrite(handler, handler.Params[0], handler.Params[1], nil, config.Writers, map[*ssa.Function]bool{})
					if !ok && declaration.probe != nil && declaration.probe.owner == stringConstant(call.Call.Args[2]) {
						written, ok = conditionalWrite(handler, config, declaration.probe)
					}
					if !ok {
						continue
					}
					written.Operation = operation
					result.Writes = append(result.Writes, written)
					if !slices.Contains(declaration.codes[written.Status], written.Code) {
						result.Violations = append(result.Violations, written)
					}
				}
			}
		}
	}
	sort.Slice(result.Writes, func(i, j int) bool { return result.Writes[i].Operation < result.Writes[j].Operation })
	sort.Slice(result.Violations, func(i, j int) bool { return result.Violations[i].Operation < result.Violations[j].Operation })
	return result, nil
}

// An escaped map, a dynamic key or another mutation cannot certify which
// handler receives an admitted request.
func literalHandlers(m *ssa.MakeMap, binder *ssa.Call) map[string]*ssa.Function {
	handlers := map[string]*ssa.Function{}
	for _, use := range *m.Referrers() {
		if use == binder {
			continue
		}
		update, ok := use.(*ssa.MapUpdate)
		if !ok || update.Map != m || update.Block() != binder.Block() {
			return nil
		}
		key, ok := update.Key.(*ssa.Const)
		if !ok || key.Value == nil || key.Value.Kind() != constant.String {
			return nil
		}
		name := constant.StringVal(key.Value)
		if _, duplicate := handlers[name]; duplicate {
			return nil
		}
		handlers[name] = handlerFunction(update.Value)
	}
	// Reject a map updated after registration.
	for i, instr := range binder.Block().Instrs {
		if instr != binder {
			continue
		}
		for _, later := range binder.Block().Instrs[i+1:] {
			if update, ok := later.(*ssa.MapUpdate); ok && update.Map == m {
				return nil
			}
		}
	}
	return handlers
}

// Follow the first unconditional call only. Any data-dependent branch,
// earlier effect, unknown dispatch or alternate response writer loses proof.
func firstResponseWrite(fn *ssa.Function, w, r ssa.Value, args map[ssa.Value]ssa.Value, writers map[string]int, seen map[*ssa.Function]bool) (writtenError, bool) {
	if seen[fn] || len(fn.Blocks) == 0 || len(fn.Blocks[0].Instrs) == 0 {
		return writtenError{}, false
	}
	seen[fn] = true
	value := func(v ssa.Value) ssa.Value {
		if actual, ok := args[v]; ok {
			return actual
		}
		return v
	}
	call, ok := fn.Blocks[0].Instrs[0].(*ssa.Call)
	if !ok {
		return writtenError{}, false
	}
	callee := call.Call.StaticCallee()
	if callee == nil {
		return writtenError{}, false
	}
	name := qualifiedFunction(callee)
	actual := call.Call.Args
	if codeIndex, ok := writers[name]; ok {
		if codeIndex != 3 || len(actual) <= codeIndex || value(actual[0]) != w || value(actual[1]) != r {
			return writtenError{}, false
		}
		status, statusOK := value(actual[2]).(*ssa.Const)
		code, codeOK := value(actual[3]).(*ssa.Const)
		if !statusOK || !codeOK || status.Value == nil || code.Value == nil || status.Value.Kind() != constant.Int || code.Value.Kind() != constant.String {
			return writtenError{}, false
		}
		n, exact := constant.Int64Val(status.Value)
		if !exact || n < 400 || n > 599 {
			return writtenError{}, false
		}
		return writtenError{Status: int(n), Code: constant.StringVal(code.Value), Writer: name}, true
	}
	if len(actual) != len(callee.Params) {
		return writtenError{}, false
	}
	next := make(map[ssa.Value]ssa.Value, len(actual))
	for i, arg := range actual {
		next[callee.Params[i]] = value(arg)
	}
	return firstResponseWrite(callee, w, r, next, writers, seen)
}
