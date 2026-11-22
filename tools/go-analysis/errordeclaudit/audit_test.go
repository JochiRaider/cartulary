package main

import (
	"go/ast"
	"go/parser"
	"reflect"
	"strings"
	"testing"
)

// TestAuditReportsSeededUndeclaredCode is the fixture proof the audit runs
// before analyzing the server (TH-HARNESS-REQ-867). Declared codes produced
// directly, through a constant phi and through a forwarding chain are not
// findings; the seeded undeclared code is, at the function that produces it;
// and an operation with no bound handler is unrooted.
func TestAuditReportsSeededUndeclaredCode(t *testing.T) {
	catalog, err := readCatalog("testdata/fixture/catalog/catalog.go")
	if err != nil {
		t.Fatal(err)
	}
	report, err := Run(Config{
		Dir:         "testdata/fixture",
		Patterns:    []string{"./server"},
		APIError:    "example.test/fixture/httpapi.APIError",
		CodeField:   "Code",
		Writers:     map[string]int{"example.test/fixture/httpapi.WriteError": 3},
		HandlerType: "net/http.HandlerFunc",
		Catalog:     catalog,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := Report{
		SchemaID:           reportSchemaID,
		OperationCount:     3,
		RootedCount:        2,
		UnrootedOperations: []string{"lostThing"},
		FindingCount:       1,
		Findings: []Finding{{
			OperationID: "putThing",
			Code:        "seeded_undeclared",
			Functions:   []string{"(*example.test/fixture/server.service).handlePut"},
		}},
	}
	if !reflect.DeepEqual(report, want) {
		t.Fatalf("fixture report = %#v, want %#v", report, want)
	}
}

func TestCatalogRejectsMalformedDeclarations(t *testing.T) {
	if _, err := readCatalog("testdata/fixture/httpapi/httpapi.go"); err == nil {
		t.Fatal("a source with no operations was accepted as a catalog")
	}
}

func TestCheckUsesProvenWriterAndDeclaredStatus(t *testing.T) {
	catalog := declarationCatalog{}
	for _, op := range []string{"declared", "wrongStatus", "undeclared", "conditional", "unrooted"} {
		catalog[op] = operationDeclaration{codes: map[int][]string{404: {"missing"}}}
	}
	result, err := checkWrites(Config{
		Dir: "testdata/fixture", Patterns: []string{"./qualified"}, Catalog: catalog,
		Writers: map[string]int{"example.test/fixture/httpapi.WriteError": 3}, HandlerType: "net/http.HandlerFunc",
	}, "example.test/fixture/qualified.BindRoutes")
	if err != nil {
		t.Fatal(err)
	}
	want := []writtenError{
		{Operation: "undeclared", Status: 404, Code: "unknown", Writer: "example.test/fixture/httpapi.WriteError"},
		{Operation: "wrongStatus", Status: 409, Code: "missing", Writer: "example.test/fixture/httpapi.WriteError"},
	}
	if len(result.Writes) != 3 || !reflect.DeepEqual(result.Violations, want) {
		t.Fatalf("qualified writes = %#v, want 3 with violations %#v", result, want)
	}
}

func TestConditionalProofUsesOnlyTheAdmittedRequest(t *testing.T) {
	names := strings.Fields("queryDeclared queryWrongStatus queryUnknown infeasible discarded reassigned escaped otherRequest otherWriter earlierResponse ambiguous absentParameter")
	catalog := declarationCatalog{}
	for _, op := range names {
		catalog[op] = operationDeclaration{codes: map[int][]string{404: {"missing"}}, pattern: "GET /" + op, probe: &admittedProbe{owner: "example", path: "/" + op}}
	}
	const writer = "example.test/fixture/httpapi.WriteError"
	result, err := checkWrites(Config{Dir: "testdata/fixture", Patterns: []string{"./qualified"}, Catalog: catalog,
		AdmittedPathString: "example.test/fixture/qualified.PathString", APIError: "example.test/fixture/httpapi.APIError", Writers: map[string]int{writer: 3}, HandlerType: "net/http.HandlerFunc"}, "example.test/fixture/qualified.BindRoutes")
	if err != nil {
		t.Fatal(err)
	}
	want := []writtenError{{"queryDeclared", 404, "missing", writer}, {"queryUnknown", 404, "unknown", writer}, {"queryWrongStatus", 409, "missing", writer}}
	if !reflect.DeepEqual(result.Writes, want) || !reflect.DeepEqual(result.Violations, want[1:]) {
		t.Fatalf("conditional writes = %#v", result)
	}
}

func TestSingletonProbeRequiresSessionAndAdmissiblePath(t *testing.T) {
	const op = `Operation{OwnerID:"owner", Method:"GET", Availability:"base", Pattern:"GET /schema/{id}", PathTemplate:"/schema/{id}", Security:[][]string{{"sessionCookie"}}, PathParameters:[]Param{{Name:"id", Pattern:"^cartulary[.]", MaxLength:100}}}`
	for _, tc := range []struct {
		old, replacement string
		want             bool
	}{
		{"GET", "GET", true}, {"GET", "POST", false}, {"sessionCookie", "credentialBootstrapBearer", false},
		{`{"sessionCookie"}`, `{}, {"sessionCookie"}`, false}, {"^cartulary", "^other", false}, {"100", "3", false},
	} {
		expr, err := parser.ParseExpr(strings.ReplaceAll(op, tc.old, tc.replacement))
		if err != nil {
			t.Fatal(err)
		}
		probe := singletonProbe(literalFields(expr.(*ast.CompositeLit)))
		if (probe != nil) != tc.want {
			t.Fatalf("replacement %q: probe=%#v", tc.replacement, probe)
		}
		if probe != nil {
			catalog := declarationCatalog{"member": {probe: probe, pattern: "GET /schema/{id}"}, "literal": {pattern: "GET " + probe.path}}
			if err := resolveProbes(catalog); err != nil {
				t.Fatal(err)
			}
			if catalog["member"].probe != nil {
				t.Fatal("literal sibling shadows witness request")
			}
		}
	}
}

func TestLiteralProbeCarriesOnlyTheSessionAndQuery(t *testing.T) {
	const op = `Operation{OwnerID:"owner", Method:"GET", Availability:"base", Pattern:"GET /things", PathTemplate:"/things", Security:[][]string{{"sessionCookie"}}`
	for _, suffix := range []string{"}", ", PathParameters:nil}", ", PathParameters:[]Param{}}", ", PathParameters:unknown}"} {
		expr, err := parser.ParseExpr(op + suffix)
		if err != nil {
			t.Fatal(err)
		}
		fields := literalFields(expr.(*ast.CompositeLit))
		probe := singletonProbe(fields)
		if suffix == ", PathParameters:unknown}" {
			if probe != nil {
				t.Fatal("unresolved path parameters qualified")
			}
			continue
		}
		if probe == nil || probe.parameter != "" || probe.path != "/things" {
			t.Fatalf("literal request with %s: %#v", suffix, probe)
		}
		fields["PathTemplate"], _ = parser.ParseExpr(`"/things/{id}"`)
		fields["Pattern"], _ = parser.ParseExpr(`"GET /things/{id}"`)
		if singletonProbe(fields) != nil {
			t.Fatal("unbound template parameter qualified")
		}
	}
}
