package artifacts

import (
	"errors"
	"slices"
	"testing"
)

func testSourceStateManifestRejectsMalformedRelations(t *testing.T) {
	tests := []struct {
		name   string
		mutate func([]sourceStateRelation) []sourceStateRelation
	}{
		{
			name: "empty manifest",
			mutate: func(_ []sourceStateRelation) []sourceStateRelation {
				return nil
			},
		},
		{
			name: "empty table identifier",
			mutate: func(relations []sourceStateRelation) []sourceStateRelation {
				relations[0].tableName = ""
				return relations
			},
		},
		{
			name: "unsafe table identifier",
			mutate: func(relations []sourceStateRelation) []sourceStateRelation {
				relations[0].tableName = `artifacts"; DROP TABLE records; --`
				return relations
			},
		},
		{
			name: "duplicate table identifier",
			mutate: func(relations []sourceStateRelation) []sourceStateRelation {
				relations[1].tableName = relations[0].tableName
				return relations
			},
		},
		{
			name: "empty logical path",
			mutate: func(relations []sourceStateRelation) []sourceStateRelation {
				relations[0].logicalBundlePath = ""
				return relations
			},
		},
		{
			name: "unsafe logical path",
			mutate: func(relations []sourceStateRelation) []sourceStateRelation {
				relations[0].logicalBundlePath = "data/../artifacts.ndjson"
				return relations
			},
		},
		{
			name: "duplicate logical path",
			mutate: func(relations []sourceStateRelation) []sourceStateRelation {
				relations[1].logicalBundlePath = relations[0].logicalBundlePath
				return relations
			},
		},
		{
			name: "empty supported versions",
			mutate: func(relations []sourceStateRelation) []sourceStateRelation {
				relations[0].supportedVersions = nil
				return relations
			},
		},
		{
			name: "non-increasing supported versions",
			mutate: func(relations []sourceStateRelation) []sourceStateRelation {
				relations[0].supportedVersions = []int{2, 1}
				return relations
			},
		},
		{
			name: "empty stable identity",
			mutate: func(relations []sourceStateRelation) []sourceStateRelation {
				relations[0].stableIdentity = nil
				return relations
			},
		},
		{
			name: "unsafe stable identity",
			mutate: func(relations []sourceStateRelation) []sourceStateRelation {
				relations[0].stableIdentity = []string{"record-id"}
				return relations
			},
		},
		{
			name: "duplicate stable identity",
			mutate: func(relations []sourceStateRelation) []sourceStateRelation {
				relations[0].stableIdentity = []string{"record_id", "record_id"}
				return relations
			},
		},
		{
			name: "empty required import columns",
			mutate: func(relations []sourceStateRelation) []sourceStateRelation {
				relations[0].requiredImportColumns = nil
				return relations
			},
		},
		{
			name: "unsafe required import column",
			mutate: func(relations []sourceStateRelation) []sourceStateRelation {
				relations[0].requiredImportColumns = []string{"record_id", "incident-id"}
				return relations
			},
		},
		{
			name: "duplicate required import column",
			mutate: func(relations []sourceStateRelation) []sourceStateRelation {
				relations[0].requiredImportColumns = []string{"record_id", "record_id"}
				return relations
			},
		},
		{
			name: "stable identity not required",
			mutate: func(relations []sourceStateRelation) []sourceStateRelation {
				relations[0].requiredImportColumns = []string{"incident_id"}
				return relations
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := newSourceStateManifest(test.mutate(cloneSourceStateRelations(authoredSourceStateRelations)))
			if !errors.Is(err, errInvalidSourceStateManifest) {
				t.Fatalf("newSourceStateManifest() error = %v, want errInvalidSourceStateManifest", err)
			}
		})
	}
}

func testSourceStateManifestDerivesExactRecoveryAndPortabilityState(t *testing.T) {
	manifest, err := loadSourceStateManifest()
	if err != nil {
		t.Fatalf("load source-state manifest: %v", err)
	}

	wantTables := []string{
		"artifacts",
		"artifact_findings",
		"artifact_investigative_queries",
		"artifact_forensic_keywords",
		"handoff_risk_refs",
	}
	tables := manifest.recoveryTables()
	if len(tables) != len(wantTables) {
		t.Fatalf("recovery table count = %d, want %d", len(tables), len(wantTables))
	}
	for index, want := range wantTables {
		if tables[index].TableName != want {
			t.Fatalf("recovery table %d = %q, want %q", index, tables[index].TableName, want)
		}
	}

	wantPaths := []struct {
		logicalPath    string
		stableIdentity string
	}{
		{"data/artifacts.ndjson", "record_id"},
		{"data/artifact_findings.ndjson", "record_id"},
		{"data/artifact_investigative_queries.ndjson", "record_id"},
		{"data/artifact_forensic_keywords.ndjson", "record_id"},
		{"data/handoff_risk_refs.ndjson", "risk_ref_id"},
	}
	paths := manifest.sourcePortPaths()
	if len(paths) != len(wantPaths) {
		t.Fatalf("source-port path count = %d, want %d", len(paths), len(wantPaths))
	}
	for index, want := range wantPaths {
		got := paths[index]
		if got.LogicalPath != want.logicalPath || got.ContentRole != "source_rows" ||
			!slices.Equal(got.Versions, []int{2}) || !slices.Equal(got.StableIdentity, []string{want.stableIdentity}) {
			t.Fatalf("source-port path %d = %#v, want %#v", index, got, want)
		}
	}

	wantExportQueries := []string{
		`SELECT to_jsonb(t) FROM "artifacts" AS t WHERE "incident_id" = $1 ORDER BY "record_id"`,
		`SELECT to_jsonb(t) FROM "artifact_findings" AS t WHERE "incident_id" = $1 ORDER BY "record_id"`,
		`SELECT to_jsonb(t) FROM "artifact_investigative_queries" AS t WHERE "incident_id" = $1 ORDER BY "record_id"`,
		`SELECT to_jsonb(t) FROM "artifact_forensic_keywords" AS t WHERE "incident_id" = $1 ORDER BY "record_id"`,
		`SELECT to_jsonb(t) FROM "handoff_risk_refs" AS t WHERE "incident_id" = $1 ORDER BY "risk_ref_id"`,
	}
	exportSpecifications := manifest.exportSpecifications()
	for index, specification := range exportSpecifications {
		if specification.logicalBundlePath != wantPaths[index].logicalPath || specification.query != wantExportQueries[index] {
			t.Fatalf("export specification %d = %#v, want path %q query %q", index, specification, wantPaths[index].logicalPath, wantExportQueries[index])
		}
	}

	wantRequiredColumns := [][]string{
		{"record_id", "incident_id"},
		{"record_id", "incident_id"},
		{"record_id", "incident_id"},
		{"record_id", "incident_id"},
		{"risk_ref_id", "handoff_record_id"},
	}
	importSpecifications := manifest.importSpecifications()
	for index, specification := range importSpecifications {
		wantTable := wantTables[index]
		wantInsertSQL := `INSERT INTO "` + wantTable + `" SELECT * FROM jsonb_populate_record(NULL::"` + wantTable + `", $1::jsonb)`
		if specification.LogicalBundlePath != wantPaths[index].logicalPath ||
			specification.AttributionTable != wantTable ||
			!slices.Equal(specification.StableIdentity, []string{wantPaths[index].stableIdentity}) ||
			!slices.Equal(specification.RequiredColumns, wantRequiredColumns[index]) ||
			specification.InsertSQL != wantInsertSQL {
			t.Fatalf("import specification %d = %#v", index, specification)
		}
	}
}

func testSourceStateManifestOwnsDefensiveRelationCopies(t *testing.T) {
	input := cloneSourceStateRelations(authoredSourceStateRelations)
	manifest, err := newSourceStateManifest(input)
	if err != nil {
		t.Fatalf("build source-state manifest: %v", err)
	}
	input[0].supportedVersions[0] = 99
	input[0].stableIdentity[0] = "mutated_identity"
	input[0].requiredImportColumns[0] = "mutated_column"
	if manifest.relations[0].supportedVersions[0] != 2 ||
		manifest.relations[0].stableIdentity[0] != "record_id" ||
		manifest.relations[0].requiredImportColumns[0] != "record_id" {
		t.Fatalf("manifest retained caller-owned slices: %#v", manifest.relations[0])
	}

	paths := manifest.sourcePortPaths()
	paths[0].Versions[0] = 99
	paths[0].StableIdentity[0] = "mutated_identity"
	imports := manifest.importSpecifications()
	imports[0].StableIdentity[0] = "mutated_identity"
	imports[0].RequiredColumns[0] = "mutated_column"
	if got := manifest.sourcePortPaths()[0]; got.Versions[0] != 2 || got.StableIdentity[0] != "record_id" {
		t.Fatalf("source-port projection exposed manifest slices: %#v", got)
	}
	if got := manifest.importSpecifications()[0]; got.StableIdentity[0] != "record_id" || got.RequiredColumns[0] != "record_id" {
		t.Fatalf("import projection exposed manifest slices: %#v", got)
	}
}

func cloneSourceStateRelations(relations []sourceStateRelation) []sourceStateRelation {
	cloned := make([]sourceStateRelation, len(relations))
	for index, relation := range relations {
		cloned[index] = cloneSourceStateRelation(relation)
	}
	return cloned
}
