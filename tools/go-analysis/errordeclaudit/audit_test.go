package main

import (
	"reflect"
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
		catalog[op] = map[int][]string{404: {"missing"}}
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
	if result.Writes != 3 || !reflect.DeepEqual(result.Violations, want) {
		t.Fatalf("qualified writes = %#v, want 3 with violations %#v", result, want)
	}
}
