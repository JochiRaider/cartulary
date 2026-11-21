package main

import (
	"fmt"
	"go/constant"
	"go/types"
	"sort"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/callgraph/vta"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

const (
	reportSchemaID      = "cartulary.harness_error_declaration_audit.v1"
	maxFindingFunctions = 5
)

// Config names the program, the error vocabulary carriers and the declared
// catalog. Type and function names are "import/path.Name".
type Config struct {
	Dir         string
	Patterns    []string
	APIError    string
	CodeField   string
	Writers     map[string]int
	HandlerType string
	Catalog     map[string][]string
}

type Finding struct {
	OperationID      string   `json:"operation_id"`
	Code             string   `json:"code"`
	Functions        []string `json:"functions"`
	FunctionsOmitted int      `json:"functions_omitted"`
}

type Report struct {
	SchemaID           string    `json:"schema_id"`
	OperationCount     int       `json:"operation_count"`
	RootedCount        int       `json:"rooted_operation_count"`
	UnrootedOperations []string  `json:"unrooted_operations"`
	FindingCount       int       `json:"finding_count"`
	Findings           []Finding `json:"findings"`
}

// Run builds the program, roots one traversal at each bound operation handler
// and reports the codes each operation can produce but does not declare.
func Run(config Config) (Report, error) {
	loaded, err := packages.Load(&packages.Config{Mode: packages.LoadAllSyntax, Dir: config.Dir}, config.Patterns...)
	if err != nil {
		return Report{}, fmt.Errorf("load program: %w", err)
	}
	if packages.PrintErrors(loaded) > 0 {
		return Report{}, fmt.Errorf("load program: package errors")
	}
	program, _ := ssautil.AllPackages(loaded, ssa.InstantiateGenerics)
	program.Build()
	functions := ssautil.AllFunctions(program)
	graph := vta.CallGraph(functions, cha.CallGraph(program))

	produced := producedCodes(config, functions)
	roots := handlerRoots(config, functions)

	report := Report{SchemaID: reportSchemaID, OperationCount: len(config.Catalog), UnrootedOperations: []string{}, Findings: []Finding{}}
	operations := make([]string, 0, len(config.Catalog))
	for operationID := range config.Catalog {
		operations = append(operations, operationID)
	}
	sort.Strings(operations)
	for _, operationID := range operations {
		handlers := roots[operationID]
		if len(handlers) == 0 {
			report.UnrootedOperations = append(report.UnrootedOperations, operationID)
			continue
		}
		report.RootedCount++
		declared := map[string]bool{}
		for _, code := range config.Catalog[operationID] {
			declared[code] = true
		}
		sites := map[string]map[string]bool{}
		for function := range reachable(graph, handlers) {
			for code := range produced[function] {
				if declared[code] {
					continue
				}
				if sites[code] == nil {
					sites[code] = map[string]bool{}
				}
				sites[code][function.String()] = true
			}
		}
		codes := make([]string, 0, len(sites))
		for code := range sites {
			codes = append(codes, code)
		}
		sort.Strings(codes)
		for _, code := range codes {
			names := make([]string, 0, len(sites[code]))
			for name := range sites[code] {
				names = append(names, name)
			}
			sort.Strings(names)
			finding := Finding{OperationID: operationID, Code: code, Functions: names}
			if len(names) > maxFindingFunctions {
				finding.Functions, finding.FunctionsOmitted = names[:maxFindingFunctions], len(names)-maxFindingFunctions
			}
			report.Findings = append(report.Findings, finding)
		}
	}
	report.FindingCount = len(report.Findings)
	return report, nil
}

// producedCodes finds, per function, the constant codes it produces. A function
// that forwards a parameter into a code sink is a carrier, and a caller that
// passes a constant to that parameter produces the constant. Carriers are
// resolved to a fixed point, so forwarding chains of any depth are followed.
func producedCodes(config Config, functions map[*ssa.Function]bool) map[*ssa.Function]map[string]bool {
	produced := map[*ssa.Function]map[string]bool{}
	carriers := map[*ssa.Function]map[int]bool{}
	for changed := true; changed; {
		changed = false
		for function := range functions {
			codes, params := functionSinks(config, function, carriers)
			if len(codes) > 0 {
				if produced[function] == nil {
					produced[function] = map[string]bool{}
				}
				for code := range codes {
					produced[function][code] = true
				}
			}
			for index := range params {
				if carriers[function] == nil {
					carriers[function] = map[int]bool{}
				}
				if !carriers[function][index] {
					carriers[function][index] = true
					changed = true
				}
			}
		}
	}
	return produced
}

func functionSinks(config Config, function *ssa.Function, carriers map[*ssa.Function]map[int]bool) (map[string]bool, map[int]bool) {
	codes, params := map[string]bool{}, map[int]bool{}
	resolve := func(value ssa.Value) { resolveCode(function, value, map[ssa.Value]bool{}, codes, params) }
	for _, block := range function.Blocks {
		for _, instruction := range block.Instrs {
			switch instruction := instruction.(type) {
			case *ssa.Store:
				if field, ok := instruction.Addr.(*ssa.FieldAddr); ok && isCodeField(config, field) {
					resolve(instruction.Val)
				}
			case ssa.CallInstruction:
				common := instruction.Common()
				callee := common.StaticCallee()
				if callee == nil {
					continue
				}
				if index, ok := config.Writers[qualifiedFunction(callee)]; ok && index < len(common.Args) {
					resolve(common.Args[index])
				}
				for index := range carriers[callee] {
					if index < len(common.Args) {
						resolve(common.Args[index])
					}
				}
			}
		}
	}
	return codes, params
}

// resolveCode follows a code value to its constants and to the function's own
// parameters. A code read from any other field or call result is not followed.
func resolveCode(function *ssa.Function, value ssa.Value, seen map[ssa.Value]bool, codes map[string]bool, params map[int]bool) {
	if seen[value] {
		return
	}
	seen[value] = true
	switch value := value.(type) {
	case *ssa.Const:
		if value.Value != nil && value.Value.Kind() == constant.String {
			codes[constant.StringVal(value.Value)] = true
		}
	case *ssa.Parameter:
		for index, parameter := range function.Params {
			if parameter == value {
				params[index] = true
			}
		}
	case *ssa.Phi:
		for _, edge := range value.Edges {
			resolveCode(function, edge, seen, codes, params)
		}
	case *ssa.ChangeType:
		resolveCode(function, value.X, seen, codes, params)
	case *ssa.Convert:
		resolveCode(function, value.X, seen, codes, params)
	}
}

func isCodeField(config Config, field *ssa.FieldAddr) bool {
	pointer, ok := field.X.Type().Underlying().(*types.Pointer)
	if !ok {
		return false
	}
	named, ok := pointer.Elem().(*types.Named)
	if !ok || qualifiedType(named) != config.APIError {
		return false
	}
	structure, ok := named.Underlying().(*types.Struct)
	return ok && field.Field < structure.NumFields() && structure.Field(field.Field).Name() == config.CodeField
}

// handlerRoots finds each operation's handler where an owner binds it: a
// constant operation identifier stored into a map of the handler type.
func handlerRoots(config Config, functions map[*ssa.Function]bool) map[string][]*ssa.Function {
	roots := map[string][]*ssa.Function{}
	for function := range functions {
		for _, block := range function.Blocks {
			for _, instruction := range block.Instrs {
				update, ok := instruction.(*ssa.MapUpdate)
				if !ok {
					continue
				}
				mapType, ok := update.Map.Type().Underlying().(*types.Map)
				if !ok {
					continue
				}
				named, ok := mapType.Elem().(*types.Named)
				if !ok || qualifiedType(named) != config.HandlerType {
					continue
				}
				key, ok := update.Key.(*ssa.Const)
				if !ok || key.Value == nil || key.Value.Kind() != constant.String {
					continue
				}
				if handler := handlerFunction(update.Value); handler != nil {
					operationID := constant.StringVal(key.Value)
					roots[operationID] = append(roots[operationID], handler)
				}
			}
		}
	}
	return roots
}

func handlerFunction(value ssa.Value) *ssa.Function {
	switch value := value.(type) {
	case *ssa.ChangeType:
		return handlerFunction(value.X)
	case *ssa.MakeClosure:
		function, _ := value.Fn.(*ssa.Function)
		return function
	case *ssa.Function:
		return value
	}
	return nil
}

func reachable(graph *callgraph.Graph, roots []*ssa.Function) map[*ssa.Function]bool {
	seen := map[*ssa.Function]bool{}
	queue := append([]*ssa.Function(nil), roots...)
	for len(queue) > 0 {
		function := queue[0]
		queue = queue[1:]
		if seen[function] {
			continue
		}
		seen[function] = true
		node := graph.Nodes[function]
		if node == nil {
			continue
		}
		for _, edge := range node.Out {
			if edge.Callee != nil && edge.Callee.Func != nil && !seen[edge.Callee.Func] {
				queue = append(queue, edge.Callee.Func)
			}
		}
	}
	return seen
}

func qualifiedType(named *types.Named) string {
	object := named.Obj()
	if object.Pkg() == nil {
		return object.Name()
	}
	return object.Pkg().Path() + "." + object.Name()
}

func qualifiedFunction(function *ssa.Function) string {
	if origin := function.Origin(); origin != nil {
		function = origin
	}
	if function.Pkg == nil || function.Signature.Recv() != nil {
		return function.String()
	}
	return function.Pkg.Pkg.Path() + "." + function.Name()
}
