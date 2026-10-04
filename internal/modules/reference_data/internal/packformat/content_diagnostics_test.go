package packformat

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func testContentSemanticFindings(t *testing.T) {
	ctx := context.Background()
	var fixture contentFixture
	for _, candidate := range contentFixtures(t) {
		if candidate.Key == "type_registry.host" {
			fixture = candidate
		}
	}
	original := strings.Split(fixture.Members["payload/entries.ndjson"], "\n")[0]
	lines := []string{}
	for _, id := range []string{"first", "second"} {
		value, err := canonicaljson.DecodeStrict([]byte(original))
		if err != nil {
			t.Fatal(err)
		}
		row := value.(map[string]any)
		row["entry_id"], row["display_label"], row["description"] = id, " hostile secret ", " forbidden value "
		row["aliases"] = []string{"same", "same"}
		row["source_refs"] = []any{map[string]any{"artifact_index": 2, "locator": " secret "}}
		row["deprecated"], row["replacement_entry_id"] = true, "missing"
		data, err := canonicaljson.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, string(data))
	}
	fixture.Members = map[string]string{"payload/entries.ndjson": strings.Join(lines, "\n") + "\n"}
	m := fixtureManifest(fixture)
	if err := ValidateContentSchema(ctx, m, contentMemory(fixture.Members)); err != nil {
		t.Fatal(err)
	}
	summary, err := CheckSummary(ctx, "content_semantics", newDiagnosticTestScratch(), func(ctx context.Context, emit FindingSink) error {
		return ContentSemanticsFindings(ctx, m, contentMemory(fixture.Members), emit)
	})
	if err != nil {
		t.Fatal(err)
	}
	// Five row-local findings in row 0; six in row 1 because its first alias
	// also collides; both replacement sources independently lack a target.
	want := []string{"$.entries", "$.entries", "$.entries[0].aliases[1]", "$.entries[0].description", "$.entries[0].display_label", "$.entries[0].source_refs[0].artifact_index", "$.entries[0].source_refs[0].locator", "$.entries[1].aliases[0]", "$.entries[1].aliases[1]", "$.entries[1].description", "$.entries[1].display_label", "$.entries[1].source_refs[0].artifact_index", "$.entries[1].source_refs[0].locator"}
	paths := []string{}
	for _, issue := range summary.Issues {
		paths = append(paths, issue.Path)
	}
	if summary.Total != len(want) || !slices.Equal(paths, want) {
		t.Fatalf("incomplete findings: %#v", summary)
	}
	encoded, err := canonicaljson.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "secret") || strings.Contains(string(encoded), "forbidden") {
		t.Fatal("semantic findings echoed source values")
	}
	if _, err := DecodeValidationSummary(encoded); err != nil {
		t.Fatal(err)
	}
	fault := errors.New("scratch unavailable")
	if err := ContentSemanticsFindings(ctx, m, contentMemory(fixture.Members), func(Finding) error { return fault }); !errors.Is(err, fault) {
		t.Fatal("diagnostic storage failure became content rejection", err)
	}
}

func testRegistrySemanticFindings(t *testing.T) {
	ctx := context.Background()
	var fixture contentFixture
	for _, candidate := range contentFixtures(t) {
		if candidate.Key == "type_registry.indicator" {
			fixture = candidate
		}
	}
	lines := strings.Split(strings.TrimSuffix(fixture.Members["payload/entries.ndjson"], "\n"), "\n")
	for i, line := range lines {
		value, err := canonicaljson.DecodeStrict([]byte(line))
		if err != nil {
			t.Fatal(err)
		}
		row := value.(map[string]any)
		row["normalization_algorithm_id"] = "unregistered_secret_algorithm"
		row["validation_algorithm_id"] = "unregistered_secret_algorithm"
		encoded, err := canonicaljson.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		lines[i] = string(encoded)
	}
	fixture.Members = map[string]string{"payload/entries.ndjson": strings.Join(lines, "\n") + "\n"}
	m := fixtureManifest(fixture)
	if err := ValidateContentSchema(ctx, m, contentMemory(fixture.Members)); err != nil {
		t.Fatal(err)
	}
	if err := ValidateContentSemantics(ctx, m, contentMemory(fixture.Members), nil); err != nil {
		t.Fatal(err)
	}
	summary, err := CheckSummary(ctx, "type_registry", newDiagnosticTestScratch(), func(ctx context.Context, emit FindingSink) error {
		return RegistryCompatibilityFindings(ctx, m, contentMemory(fixture.Members), emit)
	})
	if err != nil {
		t.Fatal(err)
	}
	if summary.Total != 18 {
		t.Fatalf("registry did not enumerate all nine policy pairs: %#v", summary)
	}
	for _, issue := range summary.Issues {
		if !strings.HasSuffix(issue.Path, ".normalization_algorithm_id") && !strings.HasSuffix(issue.Path, ".validation_algorithm_id") {
			t.Fatal("unexpected registry issue", issue)
		}
	}
}

func testContentSchemaFindings(t *testing.T) {
	var fixture contentFixture
	for _, candidate := range contentFixtures(t) {
		if candidate.Key == "type_registry.host" {
			fixture = candidate
			break
		}
	}
	valid := strings.Split(fixture.Members["payload/entries.ndjson"], "\n")[0]
	rows := []string{}
	for i := range 3 {
		var row map[string]any
		if err := json.Unmarshal([]byte(valid), &row); err != nil {
			t.Fatal(err)
		}
		switch i {
		case 0:
			delete(row, "display_label")
			delete(row, "deprecated")
		case 1:
			row["entry_id"] = 123
			row["private-hostile-name"] = true
			row["another-secret-name"] = false
		}
		data, err := canonicaljson.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		if i == 2 {
			data = append([]byte(" "), data...)
		}
		rows = append(rows, string(data))
	}
	fixture.Members = map[string]string{"payload/entries.ndjson": strings.Join(rows, "\n") + "\n"}
	summary, err := ContentSchemaSummary(context.Background(), fixtureManifest(fixture), contentMemory(fixture.Members), newDiagnosticTestScratch())
	if err != nil {
		t.Fatal(err)
	}
	if summary.Total != 5 || summary.Retained != 5 || summary.Truncated {
		t.Fatalf("wrong complete schema count: %#v", summary)
	}
	want := []string{"$.entries[0].deprecated", "$.entries[0].display_label", "$.entries[1].entry_id", `$.entries[1]["<unknown>"]`, "$.entries[2]"}
	for i, issue := range summary.Issues {
		if issue.Path != want[i] || issue.CheckID != "content_schema" {
			t.Fatalf("issue %d: %#v", i, issue)
		}
	}
	encoded, err := canonicaljson.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "hostile") || strings.Contains(string(encoded), "secret") {
		t.Fatal("schema diagnostic leaked arbitrary input names")
	}
	if _, err := DecodeValidationSummary(encoded); err != nil {
		t.Fatal(err)
	}
}

func testValidationSummaryAdmission(t *testing.T) {
	summary, err := CheckSummary(context.Background(), "manifest_schema", nil, func(_ context.Context, emit FindingSink) error { return emit(Finding{Path: "$.pack_key"}) })
	if err != nil {
		t.Fatal(err)
	}
	data, err := canonicaljson.Marshal(summary)
	// Independently calculated with Python stdlib sorted compact JSON and
	// hashlib over this ASCII-only RFC 8785 preimage, not the owner encoder.
	if *summary.PrimaryIssueID != "rpi_4f807e465331c07aee073373f298e0a8e9a851fdfa18718ceecc752cad6a9e6d" {
		t.Fatal("issue identity differs from independent vector")
	}
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeValidationSummary(data); err != nil {
		t.Fatal(err)
	}
	var original map[string]any
	if err := json.Unmarshal(data, &original); err != nil {
		t.Fatal(err)
	}
	// Every nested boundary is required, including explicitly nullable members.
	for _, boundary := range []string{"summary", "issue", "details"} {
		var source map[string]any
		switch boundary {
		case "summary":
			source = original
		case "issue":
			source = original["issues"].([]any)[0].(map[string]any)
		case "details":
			source = original["issues"].([]any)[0].(map[string]any)["safe_details"].(map[string]any)
		}
		for name := range source {
			for _, mutation := range []string{"omit", "wrong_type", "null"} {
				if mutation == "null" && source[name] == nil {
					continue
				}
				var value map[string]any
				_ = json.Unmarshal(data, &value)
				node := value
				if boundary != "summary" {
					node = value["issues"].([]any)[0].(map[string]any)
				}
				if boundary == "details" {
					node = node["safe_details"].(map[string]any)
				}
				if mutation == "omit" {
					delete(node, name)
				} else if mutation == "null" {
					node[name] = nil
				} else {
					node[name] = map[string]any{}
				}
				encoded, _ := canonicaljson.Marshal(value)
				if _, err := DecodeValidationSummary(encoded); err == nil {
					t.Fatalf("admitted %s.%s %s", boundary, name, mutation)
				}
			}
		}
	}
	for _, boundary := range []string{"summary", "issue", "details"} {
		var value map[string]any
		_ = json.Unmarshal(data, &value)
		node := value
		if boundary != "summary" {
			node = value["issues"].([]any)[0].(map[string]any)
		}
		if boundary == "details" {
			node = node["safe_details"].(map[string]any)
		}
		node["hostile_secret_member"] = true
		encoded, _ := canonicaljson.Marshal(value)
		if _, err := DecodeValidationSummary(encoded); err == nil {
			t.Fatal("unknown nested member admitted", boundary)
		}
	}
	for _, size := range []int{0, 16777216, 16777217} {
		if got := admittedValidationSummarySize(size); got != (size == 16777216) {
			t.Fatal("summary byte guard", size, got)
		}
	}
	for _, mutate := range []func(map[string]any){
		func(v map[string]any) { v["issues_truncated"] = true },
		func(v map[string]any) { v["total_issue_count"] = 2 },
		func(v map[string]any) { v["result"] = "succeeded" },
		func(v map[string]any) { v["primary_issue_id"] = nil },
		func(v map[string]any) { v["issues"].([]any)[0].(map[string]any)["check_id"] = "content_schema" },
		func(v map[string]any) { v["issues"].([]any)[0].(map[string]any)["reason_code"] = nil },
		func(v map[string]any) { v["issues"].([]any)[0].(map[string]any)["path"] = "$.private_secret" },
	} {
		var value map[string]any
		_ = json.Unmarshal(data, &value)
		mutate(value)
		encoded, _ := canonicaljson.Marshal(value)
		if _, err := DecodeValidationSummary(encoded); err == nil {
			t.Fatal("admitted inconsistent summary")
		}
	}
	if _, err := DecodeValidationSummary(append(data, '\n')); err == nil {
		t.Fatal("noncanonical retained bytes admitted")
	}
	success, _ := canonicaljson.Marshal(successSummary())
	if _, err := DecodeValidationSummary(success); err != nil {
		t.Fatal(err)
	}
}
