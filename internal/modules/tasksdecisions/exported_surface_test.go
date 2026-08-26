package tasksdecisions_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestTasksDecisionsExportedSurfaceAllowlist_Unit(t *testing.T) {
	t.Parallel()
	want := []string{
		"AdmissionFailure",
		"AdmitConflictResolveJSON",
		"AdmitCreateJSON",
		"AdmitPatchJSON",
		"AdmitSupersedeJSON",
		"CollectionAction",
		"CollectionActionPayload",
		"ConflictClaims",
		"ConflictCommand",
		"ConflictResolveRequest",
		"ConflictResolveRequestHash",
		"CreateCommand",
		"CreateRequest",
		"CreateRequestHash",
		"DecisionsViewSchemaID",
		"ErrClientTxnConflict",
		"ErrIdempotencyNotFound",
		"ErrStoredMutationKindMismatch",
		"FieldValue",
		"IdempotencyCapability",
		"IdempotencyKey",
		"IdempotencyRecord",
		"ImportCreateCommand",
		"ImportDependencies",
		"ImportLinkCapability",
		"ImportRecordEnvelopeCapability",
		"ImportRevisionCapability",
		"IncidentBundleContribution",
		"IncidentStateCapability",
		"LifecycleValidationError",
		"LinkCapability",
		"MemberReferenceCapability",
		"MutationDependencies",
		"MutationFacade",
		"MutationResult",
		"NewImportContribution",
		"NewIncidentBundleContribution",
		"NewMemberReferenceCapability",
		"NewMutationContribution",
		"NewProjectionContribution",
		"NewRecoveryContribution",
		"NewReportingContribution",
		"NewRevisionContribution",
		"NewStoredCreateResult",
		"NewStoredDecisionSupersessionResult",
		"NewStoredPatchResult",
		"OptionalConflictValue",
		"PatchChange",
		"PatchCommand",
		"PatchRequest",
		"PatchRequestHash",
		"RecordEnvelopeCapability",
		"ReportingContribution",
		"RevisionCapability",
		"RowVersionConflictError",
		"SameFieldConflict",
		"SameFieldConflictError",
		"StoredDecisionSupersessionResult",
		"StoredMutationCreate",
		"StoredMutationDecisionSupersession",
		"StoredMutationKind",
		"StoredMutationPatch",
		"StoredMutationResult",
		"StoredRowMutationResult",
		"SupersedeCommand",
		"SupersedeFacts",
		"SupersedeMutationResult",
		"SupersedeRequest",
		"SupersedeRequestHash",
		"SupersedeRowVersionConflictError",
		"TaskRequestsViewSchemaID",
		"ValidationError",
	}
	slices.Sort(want)

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate Tasks/Decisions package")
	}
	entries, err := filepath.Glob(filepath.Join(filepath.Dir(currentFile), "*.go"))
	if err != nil {
		t.Fatalf("list Tasks/Decisions sources: %v", err)
	}
	files := token.NewFileSet()
	got := make([]string, 0, len(want))
	for _, path := range entries {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(files, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, declaration := range file.Decls {
			switch current := declaration.(type) {
			case *ast.GenDecl:
				for _, specification := range current.Specs {
					switch value := specification.(type) {
					case *ast.TypeSpec:
						if value.Name.IsExported() {
							got = append(got, value.Name.Name)
						}
					case *ast.ValueSpec:
						for _, name := range value.Names {
							if name.IsExported() {
								got = append(got, name.Name)
							}
						}
					}
				}
			case *ast.FuncDecl:
				if current.Recv == nil && current.Name.IsExported() {
					got = append(got, current.Name.Name)
				}
			}
		}
	}
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Fatalf("Tasks/Decisions exported surface drifted\n got: %v\nwant: %v", got, want)
	}
}
