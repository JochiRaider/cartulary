package networkflow

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/contractnetworkflow"
)

// reasonConstructors names each semantic failure constructor's error code and
// the position of its literal reason argument.
var reasonConstructors = map[string]struct {
	code     string
	argument int
}{
	"invalidNetworkFlowRequest": {"network_flow_invalid_request", 1},
	"invalidSortAt":             {"network_flow_invalid_sort", 0},
	"invalidQueryLimit":         {"network_flow_invalid_limit", 0},
	"invalidLimitOverride":      {"network_flow_invalid_limit_override", 1},
	"cursorInvalid":             {"network_flow_cursor_invalid", 0},
	"invalidFilter":             {"network_flow_invalid_filter", 1},
	"invalidTableScope":         {"network_flow_invalid_table_scope", 1},
	"invalidTimeRange":          {"network_flow_invalid_time_range", 1},
	"invalidGraphAggregation":   {"network_flow_invalid_graph_aggregation", 1},
	"graphQueryStale":           {"network_flow_graph_query_stale", 0},
	"graphLimitExceeded":        {"network_flow_graph_limit_exceeded", 0},
	"graphLimitExceededAt":      {"network_flow_graph_limit_exceeded", 0},
	"invalidIndicatorSelector":  {"network_flow_invalid_indicator_selector", 1},
	"invalidIndicatorTarget":    {"network_flow_invalid_indicator_target", 1},
	"indicatorLinkForbidden":    {"network_flow_indicator_link_forbidden", 3},
	"indicatorLinkAmbiguous":    {"network_flow_indicator_link_ambiguous", 2},
}

// assertEmittedReasonCodesRegistered covers NF NLSpec Table 21-A1 statically:
// every literal reason code a Network Flow semantic failure is built with is
// permitted for its error code by the packaged reason registry, including on
// paths no test exercises.
func assertEmittedReasonCodesRegistered(t *testing.T) {
	t.Helper()
	var registry struct {
		ReasonRegistries []struct {
			ErrorCode   string   `json:"error_code"`
			ReasonCodes []string `json:"reason_codes"`
		} `json:"reason_registries"`
	}
	if err := json.Unmarshal([]byte(contractnetworkflow.Index["contracts/network-flow/errors.v2.json"].JSON), &registry); err != nil {
		t.Fatal(err)
	}
	permitted := map[string]map[string]bool{}
	for _, family := range registry.ReasonRegistries {
		codes := strings.Split(family.ErrorCode, ",")
		if family.ErrorCode == "indicator-link errors" {
			codes = []string{"network_flow_indicator_link_ambiguous", "network_flow_invalid_indicator_selector", "network_flow_invalid_indicator_target", "network_flow_indicator_link_forbidden"}
		}
		for _, code := range codes {
			if permitted[code] == nil {
				permitted[code] = map[string]bool{}
			}
			for _, reason := range family.ReasonCodes {
				permitted[code][reason] = true
			}
		}
	}
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve the Network Flow package")
	}
	directory := filepath.Dir(currentFile)
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	fileSet := token.NewFileSet()
	files := []*ast.File{}
	failureKinds := map[string]string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(fileSet, filepath.Join(directory, entry.Name()), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, parsed)
		for _, declaration := range parsed.Decls {
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.CONST {
				continue
			}
			for _, spec := range general.Specs {
				value, ok := spec.(*ast.ValueSpec)
				if !ok || len(value.Values) != len(value.Names) {
					continue
				}
				for index, name := range value.Names {
					if literal, ok := value.Values[index].(*ast.BasicLit); ok && strings.HasPrefix(name.Name, "failure") {
						if code, err := strconv.Unquote(literal.Value); err == nil {
							failureKinds[name.Name] = code
						}
					}
				}
			}
		}
	}
	literal := func(expression ast.Expr) (string, bool) {
		value, ok := expression.(*ast.BasicLit)
		if !ok || value.Kind != token.STRING {
			return "", false
		}
		text, err := strconv.Unquote(value.Value)
		return text, err == nil
	}
	check := func(position token.Pos, code, reason string) {
		if !strings.HasPrefix(code, "network_flow_") {
			return
		}
		if code == "network_flow_invalid_filter" && (reason == "value_forbidden" || reason == "unknown_member") {
			reason = "invalid_value" // invalidFilter's own remapping
		}
		if !permitted[code][reason] {
			t.Errorf("%s: %s reason %q is not in its Table 21-A1 reason registry", fileSet.Position(position), code, reason)
		}
	}
	checked := 0
	for _, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch value := node.(type) {
			case *ast.CallExpr:
				name, ok := value.Fun.(*ast.Ident)
				if !ok {
					return true
				}
				if constructor, known := reasonConstructors[name.Name]; known && len(value.Args) > constructor.argument {
					if reason, ok := literal(value.Args[constructor.argument]); ok {
						check(value.Pos(), constructor.code, reason)
						checked++
					}
				}
				if name.Name == "newSemanticFailure" && len(value.Args) == 3 {
					kind, _ := value.Args[0].(*ast.Ident)
					if reason, ok := literal(value.Args[2]); ok && kind != nil {
						check(value.Pos(), failureKinds[kind.Name], reason)
						checked++
					}
				}
			case *ast.CompositeLit:
				typeName, ok := value.Type.(*ast.Ident)
				if !ok || typeName.Name != "semanticFailure" {
					return true
				}
				var kind, reason string
				for _, element := range value.Elts {
					field, ok := element.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					switch key, _ := field.Key.(*ast.Ident); {
					case key != nil && key.Name == "kind":
						if ident, ok := field.Value.(*ast.Ident); ok {
							kind = failureKinds[ident.Name]
						}
					case key != nil && key.Name == "reason":
						reason, _ = literal(field.Value)
					}
				}
				if kind != "" && reason != "" {
					check(value.Pos(), kind, reason)
					checked++
				}
			}
			return true
		})
	}
	if checked < 100 {
		t.Fatalf("checked only %d literal reason codes; the scan no longer sees the constructors", checked)
	}
}
