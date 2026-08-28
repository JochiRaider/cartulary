package artifacts

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"sort"
	"strings"
	"testing"
)

type exportDisposition string

const (
	exportRetain   exportDisposition = "retain"
	exportRemove   exportDisposition = "remove"
	exportUnexport exportDisposition = "unexport"
	exportRename   exportDisposition = "rename"
)

// artifactRootExportClassifications is the reviewed Iteration 3 production
// boundary. Owner-local and caller-free entries are absent; new root exports
// always require an explicit owner decision.
var artifactRootExportClassifications = map[string]exportDisposition{
	"AdmissionError":                       exportRetain,
	"AdmissionError.CollectionField":       exportRetain,
	"AdmissionError.Error":                 exportRetain,
	"AdmissionError.Field":                 exportRetain,
	"AdmissionError.MaximumCount":          exportRetain,
	"AdmissionError.ReasonCode":            exportRetain,
	"AdmissionError.RequestedCount":        exportRetain,
	"AdmissionReasonCode":                  exportRetain,
	"AdmitConflictResolution":              exportRetain,
	"AdmitContextualNote":                  exportRetain,
	"AdmitCreate":                          exportRetain,
	"AdmitPatch":                           exportRetain,
	"CommLogViewSchemaID":                  exportRetain,
	"ConflictAdmissionContext":             exportRetain,
	"ConflictCommand":                      exportRetain,
	"ConflictResolveAdmission":             exportRetain,
	"ConflictResolveAdmission.ClientTxnID": exportRetain,
	"ContextualLink":                       exportRetain,
	"ContextualNoteAdmission":              exportRetain,
	"ContextualNoteAdmission.ClientTxnID":  exportRetain,
	"ContextualNoteCreateCommand":          exportRetain,
	"CreateAdmission":                      exportRetain,
	"CreateAdmission.ClientTxnID":          exportRetain,
	"CreateAdmission.ViewSchemaID":         exportRetain,
	"CreateCommand":                        exportRetain,
	"ErrClientTxnConflict":                 exportRetain,
	"ErrStoredMutationKindMismatch":        exportRetain,
	"FindingsViewSchemaID":                 exportRetain,
	"ForensicKeywordsViewSchemaID":         exportRetain,
	"HandoffViewSchemaID":                  exportRetain,
	"IdempotencyCapability":                exportRetain,
	"IdempotencyKey":                       exportRetain,
	"ImportDependencies":                   exportRetain,
	"IncidentBundleSubtypeContribution":    exportRetain,
	"IncidentStateCapability":              exportRetain,
	"InvestigativeQueriesViewSchemaID":     exportRetain,
	"LessonViewSchemaID":                   exportRetain,
	"LinkCapability":                       exportRetain,
	"MemberReferenceCapability":            exportRetain,
	"MutationDependencies":                 exportRetain,
	"MutationFacade":                       exportRetain,
	"MutationFacade.Create":                exportRetain,
	"MutationFacade.CreateContextualNote":  exportRetain,
	"MutationFacade.Patch":                 exportRetain,
	"MutationFacade.ResolveConflict":       exportRetain,
	"MutationResult":                       exportRetain,
	"MutationOutcome":                      exportRetain,
	"MutationOutcomeCreated":               exportRetain,
	"MutationOutcomeKeptSaved":             exportRetain,
	"MutationOutcomeReplayed":              exportRetain,
	"MutationOutcomeUpdated":               exportRetain,
	"NewImportContribution":                exportRetain,
	"NewIncidentBundleSourcePort":          exportRetain,
	"NewMemberReferenceCapability":         exportRetain,
	"NewMutationContribution":              exportRetain,
	"NewProjectionContribution":            exportRetain,
	"NewReportingContribution":             exportRetain,
	"NewRevisionContribution":              exportRetain,
	"NewStoredCreateResult":                exportRetain,
	"NewStoredLinkedNoteResult":            exportRetain,
	"NewStoredPatchResult":                 exportRetain,
	"NotesViewSchemaID":                    exportRetain,
	"OperationConflictResolve":             exportRetain,
	"OperationCreate":                      exportRetain,
	"OperationID":                          exportRetain,
	"OperationLinkedNoteCreate":            exportRetain,
	"OperationPatch":                       exportRetain,
	"OptionalConflictValue":                exportRetain,
	"PatchAdmission":                       exportRetain,
	"PatchAdmission.BaseRowVersion":        exportRetain,
	"PatchAdmission.ClientTxnID":           exportRetain,
	"PatchAdmission.ViewSchemaID":          exportRetain,
	"PatchCommand":                         exportRetain,
	"RecordEnvelopeCapability":             exportRetain,
	"RecoveryStateContribution":            exportRetain,
	"RevisionCapability":                   exportRetain,
	"RowVersionConflictError":              exportRetain,
	"RowVersionConflictError.Error":        exportRetain,
	"SameFieldConflictError":               exportRetain,
	"SameFieldConflictError.Error":         exportRetain,
	"SameFieldConflict":                    exportRetain,
	"StatusReviewViewSchemaID":             exportRetain,
	"StoredMutationCreate":                 exportRetain,
	"StoredMutationKind":                   exportRetain,
	"StoredMutationLinkedNote":             exportRetain,
	"StoredMutationPatch":                  exportRetain,
	"StoredMutationPayload":                exportRetain,
	"StoredMutationResult":                 exportRetain,
	"StoredMutationResult.Kind":            exportRetain,
	"StoredMutationResult.Payload":         exportRetain,
	"ValidationError":                      exportRetain,
	"ValidationError.Error":                exportRetain,
}

var workbookProjectionExportClassifications = map[string]exportDisposition{
	"CommunicationLogVariant":              exportRetain,
	"Contribution":                         exportRetain,
	"Contribution.ProjectionContribution":  exportRetain,
	"Contribution.Source":                  exportRetain,
	"DerivedFact":                          exportRetain,
	"FindingVariant":                       exportRetain,
	"ForensicKeywordVariant":               exportRetain,
	"HandoffVariant":                       exportRetain,
	"InvestigativeQueryVariant":            exportRetain,
	"LessonVariant":                        exportRetain,
	"NewCommunicationLogProjectionInput":   exportRetain,
	"NewContribution":                      exportRetain,
	"NewFindingProjectionInput":            exportRetain,
	"NewForensicKeywordProjectionInput":    exportRetain,
	"NewHandoffProjectionInput":            exportRetain,
	"NewInvestigativeQueryProjectionInput": exportRetain,
	"NewLessonProjectionInput":             exportRetain,
	"NewNoteProjectionInput":               exportRetain,
	"NewProjectionEnvelope":                exportRetain,
	"NewProjectionInputPage":               exportRetain,
	"NewStatusReviewProjectionInput":       exportRetain,
	"NoteVariant":                          exportRetain,
	"Ports":                                exportRetain,
	"ProjectionEnvelope":                   exportRetain,
	"ProjectionEnvelope.Body":              exportRetain,
	"ProjectionEnvelope.CreatedAt":         exportRetain,
	"ProjectionEnvelope.CreatedByUserID":   exportRetain,
	"ProjectionEnvelope.IncidentID":        exportRetain,
	"ProjectionEnvelope.LinkedRecordCount": exportRetain,
	"ProjectionEnvelope.RecordID":          exportRetain,
	"ProjectionEnvelope.RowVersion":        exportRetain,
	"ProjectionEnvelope.TimestampDay":      exportRetain,
	"ProjectionEnvelope.TimestampUTC":      exportRetain,
	"ProjectionEnvelope.Title":             exportRetain,
	"ProjectionEnvelope.UpdatedAt":         exportRetain,
	"ProjectionInput":                      exportRetain,
	"ProjectionInput.ArtifactType":         exportRetain,
	"ProjectionInput.Envelope":             exportRetain,
	"ProjectionInput.Variant":              exportRetain,
	"ProjectionInputPage":                  exportRetain,
	"ProjectionInputPage.Inputs":           exportRetain,
	"ProjectionInputPage.NextRecordID":     exportRetain,
	"ProjectionVariant":                    exportRetain,
	"Reader":                               exportRetain,
	"Rows":                                 exportRetain,
	"SourceReader":                         exportRetain,
	"StatusReviewVariant":                  exportRetain,
}

func TestArtifactExportedSurfaceReachabilityLock(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve Artifacts package path")
	}
	root := filepath.Dir(currentFile)
	assertExportedSurface(t, root, artifactRootExportClassifications)
	assertExportedSurface(t, filepath.Join(root, "workbookprojection"), workbookProjectionExportClassifications)

	t.Run("negative fixture detects an unapproved export", func(t *testing.T) {
		file, err := parser.ParseFile(token.NewFileSet(), "negative_fixture.go", "package negativefixture\nfunc UnexpectedExport() {}\n", 0)
		if err != nil {
			t.Fatalf("parse negative fixture: %v", err)
		}
		unexpected, missing := exportSurfaceDelta(exportedDeclarations(file), nil)
		if !reflect.DeepEqual(unexpected, []string{"UnexpectedExport"}) || len(missing) != 0 {
			t.Fatalf("negative fixture was not rejected: unexpected=%v missing=%v", unexpected, missing)
		}
	})

	t.Run("mutation and integration declarations retain cohesive file ownership", func(t *testing.T) {
		assertAbsentFiles(t, root,
			"workbook_facade.go",
			"workbook_conflict.go",
			"mutation_kernel.go",
			"mutation_ports.go",
			"artifact_contract_integration_test.go",
		)
		assertFileDeclarations(t, root, map[string][]string{
			"mutation_facade.go":                           {"MutationFacade", "NewMutationContribution"},
			"mutation_create.go":                           {"MutationFacade.Create", "MutationFacade.create", "MutationFacade.executeCreateTx"},
			"mutation_patch.go":                            {"MutationFacade.Patch", "MutationFacade.applyPatchTx"},
			"mutation_conflict.go":                         {"MutationFacade.ResolveConflict", "MutationFacade.loadConflictTarget", "buildArtifactSameFieldConflict"},
			"mutation_idempotency.go":                      {"StoredMutationResult", "MutationFacade.replayStoredMutation"},
			"mutation_collections.go":                      {"MutationFacade.applyCollectionsTx", "MutationFacade.applyCollectionTx", "MutationFacade.appendCollectionMutationsTx"},
			"mutation_shared.go":                           {"MutationDependencies", "RecordEnvelopeCapability", "MutationFacade.loadArtifactRecordMetaForUpdateTx"},
			"artifact_source_mutation_integration_test.go": {"TestArtifactWorkbookMutationContractMatrix"},
			"artifact_rollback_integration_test.go":        {"TestArtifactRollbackProviderExecutionMatrix"},
			"artifact_patch_envelope_integration_test.go":  {"TestArtifactPatchEnvelopeOutcomes"},
			"artifact_collection_integration_test.go":      {"TestArtifactCollectionMutationContractMatrix"},
			"artifact_import_integration_test.go":          {"TestArtifactImportCreateFacadeContract"},
			"artifact_conflict_integration_test.go":        {"TestArtifactConflictSourceRevalidation"},
			"artifact_contract_support_test.go":            {"mustArtifactMutationFacade"},
		})
	})
}

func assertAbsentFiles(t testing.TB, root string, names ...string) {
	t.Helper()
	for _, name := range names {
		if _, err := os.Stat(filepath.Join(root, name)); !os.IsNotExist(err) {
			t.Fatalf("retired cohesion file %s exists or cannot be checked: %v", name, err)
		}
	}
}

func assertFileDeclarations(t testing.TB, root string, expected map[string][]string) {
	t.Helper()
	fset := token.NewFileSet()
	for name, want := range expected {
		path := filepath.Join(root, name)
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parse cohesive file %s: %v", name, err)
		}
		actual := declarationNames(file)
		for _, declaration := range want {
			if !slices.Contains(actual, declaration) {
				t.Fatalf("cohesive file %s declarations = %v; missing %s", name, actual, declaration)
			}
		}
	}
}

func declarationNames(file *ast.File) []string {
	var names []string
	for _, declaration := range file.Decls {
		switch typed := declaration.(type) {
		case *ast.FuncDecl:
			name := typed.Name.Name
			if typed.Recv != nil {
				receiver := receiverName(typed.Recv.List[0].Type)
				if receiver != "" {
					name = receiver + "." + name
				}
			}
			names = append(names, name)
		case *ast.GenDecl:
			for _, specification := range typed.Specs {
				if spec, ok := specification.(*ast.TypeSpec); ok {
					names = append(names, spec.Name.Name)
				}
			}
		}
	}
	return names
}

func assertExportedSurface(t testing.TB, directory string, expected map[string]exportDisposition) {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read package directory %s: %v", directory, err)
	}
	fset := token.NewFileSet()
	var actual []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		file, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			t.Fatalf("parse %s: %v", path, parseErr)
		}
		actual = append(actual, exportedDeclarations(file)...)
	}
	allowed := make([]string, 0, len(expected))
	for name := range expected {
		allowed = append(allowed, name)
	}
	unapproved, missing := exportSurfaceDelta(actual, allowed)
	if len(unapproved) != 0 || len(missing) != 0 {
		t.Fatalf("exported surface mismatch for %s: unapproved=%v missing=%v", directory, unapproved, missing)
	}
}

func exportedDeclarations(file *ast.File) []string {
	var names []string
	for _, declaration := range file.Decls {
		switch typed := declaration.(type) {
		case *ast.FuncDecl:
			if !ast.IsExported(typed.Name.Name) {
				continue
			}
			name := typed.Name.Name
			if typed.Recv != nil {
				receiver := exportedReceiverName(typed.Recv.List[0].Type)
				if receiver == "" {
					continue
				}
				name = receiver + "." + name
			}
			names = append(names, name)
		case *ast.GenDecl:
			for _, specification := range typed.Specs {
				switch spec := specification.(type) {
				case *ast.TypeSpec:
					if ast.IsExported(spec.Name.Name) {
						names = append(names, spec.Name.Name)
					}
				case *ast.ValueSpec:
					for _, name := range spec.Names {
						if ast.IsExported(name.Name) {
							names = append(names, name.Name)
						}
					}
				}
			}
		}
	}
	return names
}

func exportedReceiverName(expression ast.Expr) string {
	name := receiverName(expression)
	if ast.IsExported(name) {
		return name
	}
	return ""
}

func receiverName(expression ast.Expr) string {
	switch typed := expression.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return receiverName(typed.X)
	}
	return ""
}

func exportSurfaceDelta(actual, expected []string) ([]string, []string) {
	actualSet := make(map[string]struct{}, len(actual))
	for _, name := range actual {
		actualSet[name] = struct{}{}
	}
	expectedSet := make(map[string]struct{}, len(expected))
	for _, name := range expected {
		expectedSet[name] = struct{}{}
	}
	var unexpected []string
	for name := range actualSet {
		if _, ok := expectedSet[name]; !ok {
			unexpected = append(unexpected, name)
		}
	}
	var missing []string
	for name := range expectedSet {
		if _, ok := actualSet[name]; !ok {
			missing = append(missing, name)
		}
	}
	sort.Strings(unexpected)
	sort.Strings(missing)
	return unexpected, missing
}
