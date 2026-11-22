package main

import (
	"go/constant"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ssa"
)

// Trace one input, following only resolved control flow. A missing fact or an
// instruction outside the supported subset ends this attempt. The budget is
// shared by all calls and loop iterations.
type requestTrace struct {
	probe     *admittedProbe
	config    Config
	remaining int
	write     writtenError
}

func conditionalWrite(handler *ssa.Function, config Config, probe *admittedProbe) (writtenError, bool) {
	trace := requestTrace{probe: probe, config: config, remaining: 256}
	trace.follow(handler, []any{originalWriter, originalRequest})
	return trace.write, trace.write.Writer != ""
}

type traceToken int

const (
	unresolved traceToken = iota
	originalWriter
	originalRequest
	requestURL
	requestQuery
	localMap
	boxedString
)

type frame map[ssa.Value]any

func (f frame) value(v ssa.Value) any {
	if value, ok := f[v]; ok {
		return value
	}
	if c, ok := v.(*ssa.Const); ok {
		if c.IsNil() {
			return nil
		}
		if c.Value != nil {
			switch c.Value.Kind() {
			case constant.String:
				return constant.StringVal(c.Value)
			case constant.Bool:
				return constant.BoolVal(c.Value)
			case constant.Int:
				if n, exact := constant.Int64Val(c.Value); exact {
					return n
				}
			}
		}
	}
	return unresolved
}

// Only fresh, single-assignment local storage can supply facts. Opaque
// values can be carried (e.g. the authenticated principal), never inspected.
type traceCell struct {
	value    any
	written  bool
	children map[int]*traceCell
}

func (c *traceCell) child(index int) *traceCell {
	if c.children == nil {
		c.children = make(map[int]*traceCell)
	}
	if c.children[index] == nil {
		c.children[index] = &traceCell{}
	}
	return c.children[index]
}

func (t *requestTrace) follow(fn *ssa.Function, args []any) (any, bool) {
	if fn == nil || len(fn.Blocks) == 0 || len(fn.Params) != len(args) {
		return nil, false
	}
	f := frame{}
	for i, p := range fn.Params {
		f[p] = args[i]
	}
	block := fn.Blocks[0]
	var previous *ssa.BasicBlock
	for {
		var next *ssa.BasicBlock
		// Multiple phis require simultaneous transfer. Leave them unresolved.
		if len(block.Instrs) > 1 {
			if _, multi := block.Instrs[1].(*ssa.Phi); multi {
				return nil, false
			}
		}
		for _, instruction := range block.Instrs {
			t.remaining--
			if t.remaining < 0 {
				return nil, false
			}
			switch i := instruction.(type) {
			case *ssa.Jump:
				next = block.Succs[0]
			case *ssa.If:
				condition, ok := f.value(i.Cond).(bool)
				if !ok {
					return nil, false
				}
				next = block.Succs[0]
				if !condition {
					next = block.Succs[1]
				}
			case *ssa.Phi:
				for index, pred := range block.Preds {
					if pred == previous {
						f[i] = f.value(i.Edges[index])
					}
				}
			case *ssa.Return:
				results := make([]any, len(i.Results))
				for n, v := range i.Results {
					results[n] = f.value(v)
				}
				if len(results) == 1 {
					return results[0], true
				}
				return results, true
			case *ssa.Store:
				cell, local := f.value(i.Addr).(*traceCell)
				value := f.value(i.Val)
				_, pointer := value.(*traceCell)
				if !local || cell.written || cell.children != nil || pointer {
					return nil, false
				}
				cell.value, cell.written = value, true
			case *ssa.MapUpdate:
				// Only inert string metadata in a fresh local map is admitted.
				if f.value(i.Map) != localMap {
					return nil, false
				}
				if _, ok := f.value(i.Key).(string); !ok {
					return nil, false
				}
				if f.value(i.Value) != boxedString {
					return nil, false
				}
			case ssa.Value:
				value, ok := t.evaluate(f, i)
				if !ok || t.write.Writer != "" {
					return nil, false
				}
				f[i] = value
			default:
				return nil, false
			}
		}
		if next == nil {
			return nil, false
		}
		previous, block = block, next
	}
}

type stringArray struct {
	storage *traceCell
	length  int64
}

func (t *requestTrace) evaluate(f frame, v ssa.Value) (any, bool) {
	switch i := v.(type) {
	case *ssa.Call:
		return t.call(f, i)
	case *ssa.Alloc:
		return &traceCell{}, true
	case *ssa.FieldAddr:
		base := f.value(i.X)
		field := i.X.Type().Underlying().(*types.Pointer).Elem().Underlying().(*types.Struct).Field(i.Field)
		if base == originalRequest && field.Name() == "URL" {
			return requestURL, true // read-only address
		}
		if c, ok := base.(*traceCell); ok && !c.written {
			return c.child(i.Field), true
		}
	case *ssa.UnOp:
		if i.Op == token.MUL {
			addr := f.value(i.X)
			if addr == requestURL {
				return requestURL, true
			}
			if c, ok := addr.(*traceCell); ok {
				if c.written {
					return c.value, true
				}
				// The message of a fresh APIError commonly starts empty.
				if basic, ok := i.Type().Underlying().(*types.Basic); ok && basic.Kind() == types.String {
					return "", true
				}
			}
		}
	case *ssa.BinOp:
		if i.Op == token.ADD && !types.Identical(i.Type(), types.Typ[types.Int]) {
			return nil, false
		}
		return compareOrAdd(i.Op, f.value(i.X), f.value(i.Y))
	case *ssa.Extract:
		if tuple, ok := f.value(i.Tuple).([]any); ok && i.Index < len(tuple) {
			return tuple[i.Index], true
		}
	case *ssa.MakeMap:
		if m, ok := i.Type().Underlying().(*types.Map); ok && m.Key().String() == "string" && i.Reserve != nil {
			if size, ok := f.value(i.Reserve).(int64); ok && size >= 0 && size <= 256 {
				return localMap, true
			}
		}
	case *ssa.MakeInterface:
		if _, ok := f.value(i.X).(string); ok {
			return boxedString, true
		}
	case *ssa.Slice:
		if i.Low == nil && i.High == nil && i.Max == nil {
			if a, ok := f.array(i.X); ok {
				return a, true
			}
		}
	case *ssa.IndexAddr:
		a, ok := f.array(i.X)
		index, known := f.value(i.Index).(int64)
		if ok && known && index >= 0 && index < a.length {
			return a.storage.child(int(index)), true
		}
	case *ssa.Lookup:
		key, known := f.value(i.Index).(string)
		if i.CommaOk && f.value(i.X) == requestQuery && known {
			_, present := t.probe.query[key]
			return []any{unresolved, present}, true
		}
	}
	return nil, false
}

func (f frame) array(v ssa.Value) (stringArray, bool) {
	if array, ok := f.value(v).(stringArray); ok {
		return array, true
	}
	if p, ok := v.Type().Underlying().(*types.Pointer); ok {
		if array, ok := p.Elem().Underlying().(*types.Array); ok && array.Elem().String() == "string" {
			if cell, ok := f.value(v).(*traceCell); ok && !cell.written {
				return stringArray{cell, array.Len()}, true
			}
		}
	}
	return stringArray{}, false
}

func compareOrAdd(op token.Token, left, right any) (any, bool) {
	if op == token.ADD || op == token.LSS {
		a, aOK := left.(int64)
		b, bOK := right.(int64)
		if !aOK || !bOK {
			return nil, false
		}
		if op == token.LSS {
			return a < b, true
		}
		// Only the small indices of a literal range are qualified.
		if a < -1 || a > 256 || b != 1 {
			return nil, false
		}
		return a + b, true
	}
	if op != token.EQL && op != token.NEQ {
		return nil, false
	}
	// Pointer/null checks and concrete string comparisons only.
	equal := false
	switch a := left.(type) {
	case *traceCell, nil:
		if right != nil {
			return nil, false
		}
		equal = left == nil
	case string:
		b, ok := right.(string)
		if !ok {
			return nil, false
		}
		equal = a == b
	default:
		return nil, false
	}
	return equal == (op == token.EQL), true
}

func (t *requestTrace) call(f frame, call *ssa.Call) (any, bool) {
	args := make([]any, len(call.Call.Args))
	for n, arg := range call.Call.Args {
		args[n] = f.value(arg)
	}
	if builtin, ok := call.Call.Value.(*ssa.Builtin); ok {
		if builtin.Name() == "len" && len(args) == 1 {
			if array, ok := args[0].(stringArray); ok {
				return array.length, true
			}
		}
		return nil, false
	}
	fn := call.Call.StaticCallee()
	if fn == nil {
		return nil, false
	}
	name := qualifiedFunction(fn)
	if index, sink := t.config.Writers[name]; sink {
		if index == 3 && len(args) > index && args[0] == originalWriter && args[1] == originalRequest {
			status, statusOK := args[2].(int64)
			code, codeOK := args[index].(string)
			if statusOK && status >= 400 && status <= 599 && codeOK {
				t.write = writtenError{Status: int(status), Code: code, Writer: name}
			}
		}
		return nil, false
	}
	switch name {
	case t.config.AdmittedPrincipal:
		if len(args) == 1 && args[0] == originalRequest {
			return []any{unresolved, true}, true
		}
	case t.config.AdmittedPathString:
		if t.probe.parameter != "" && len(args) == 2 && args[0] == originalRequest && args[1] == t.probe.parameter {
			return []any{t.probe.value, true}, true
		}
	case "(*net/url.URL).Query":
		if len(args) == 1 && args[0] == requestURL {
			return requestQuery, true
		}
	default:
		return t.follow(fn, args)
	}
	return nil, false
}
