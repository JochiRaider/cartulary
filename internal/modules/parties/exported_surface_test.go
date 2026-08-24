package parties

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

var partyExportedSurfaceAllowlist = []string{
	"const:MutationCreated",
	"const:MutationKeptSaved",
	"const:MutationReplayed",
	"const:MutationReused",
	"const:MutationUpdated",
	"const:PartyMatchCrossKeyExactMatch",
	"const:PartyMatchExactKeyClaimed",
	"const:StoredMutationCreate",
	"const:StoredMutationPatch",
	"const:ViewSchemaID",
	"func:AdmitConflictResolveJSON",
	"func:AdmitCreateJSON",
	"func:AdmitPatchJSON",
	"func:IncidentBundleSubtypeContribution",
	"func:NewImportContribution",
	"func:NewIncidentBundleContribution",
	"func:NewMutationContribution",
	"func:NewProjectionContribution",
	"func:NewReportingContribution",
	"func:NewRevisionContribution",
	"func:NewStoredCreateResult",
	"func:NewStoredPatchResult",
	"func:RecoveryStateContribution",
	"method:AdmissionError.Error",
	"method:AdmissionError.Limit",
	"method:ConflictResolveAdmission.ClientTransactionID",
	"method:CreateAdmission.ClientTransactionID",
	"method:MutationFacade.Create",
	"method:MutationFacade.Patch",
	"method:MutationFacade.ResolveConflict",
	"method:PatchAdmission.AdmittedBaseRowVersion",
	"method:PatchAdmission.ClientTransactionID",
	"method:RowVersionConflictError.Error",
	"method:SameFieldConflictError.Error",
	"method:StoredMutationResult.Kind",
	"method:StoredMutationResult.RowMutationResult",
	"method:ValidationError.Error",
	"type:AdmissionError",
	"type:ConflictClaims",
	"type:ConflictCommand",
	"type:ConflictResolveAdmission",
	"type:CreateAdmission",
	"type:CreateCommand",
	"type:IdempotencyCapability",
	"type:IdempotencyKey",
	"type:ImportDependencies",
	"type:ImportRecordEnvelopeCapability",
	"type:IncidentStateCapability",
	"type:KeepSavedCapability",
	"type:KeepSavedResult",
	"type:MutationDependencies",
	"type:MutationFacade",
	"type:MutationOutcome",
	"type:MutationResult",
	"type:OptionalConflictValue",
	"type:PartyMatchConflictError",
	"type:PatchAdmission",
	"type:PatchCommand",
	"type:RecordEnvelopeCapability",
	"type:RevisionCapability",
	"type:RowVersionConflictError",
	"type:SameFieldConflict",
	"type:SameFieldConflictError",
	"type:StoredMutationKind",
	"type:StoredMutationResult",
	"type:StoredRowMutationResult",
	"type:ValidationError",
	"var:ErrClientTxnConflict",
	"var:ErrStoredMutationKindMismatch",
}

func TestPartyExportedSurfaceLock_Unit(t *testing.T) {
	actual := partyExportedDeclarations(t)
	if !slices.Equal(actual, partyExportedSurfaceAllowlist) {
		t.Fatalf("Party exported surface mismatch:\nactual=%q\nwant=%q", actual, partyExportedSurfaceAllowlist)
	}

	assertPartyStructFields(t, reflect.TypeOf(CreateCommand{}), []string{
		"ActorUserID", "IncidentID", "Admission", "RequestID", "Now",
	})
	assertPartyStructFields(t, reflect.TypeOf(PatchCommand{}), []string{
		"ActorUserID", "RecordID", "Admission", "RequestID", "Now",
	})
	assertPartyStructFields(t, reflect.TypeOf(ConflictCommand{}), []string{
		"ActorUserID", "Admission", "RequestID", "Now",
	})
	assertPartyStructFields(t, reflect.TypeOf(MutationResult{}), []string{
		"Outcome", "Row", "IncidentID", "RecordID", "ChangeSetID", "RowVersion", "ChangedFieldKeys",
	})
}

func partyExportedDeclarations(t testing.TB) []string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read Parties root: %v", err)
	}
	set := token.NewFileSet()
	var result []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(set, filepath.Join(".", entry.Name()), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", entry.Name(), err)
		}
		for _, declaration := range file.Decls {
			switch typed := declaration.(type) {
			case *ast.GenDecl:
				for _, spec := range typed.Specs {
					switch value := spec.(type) {
					case *ast.TypeSpec:
						if value.Name.IsExported() {
							result = append(result, "type:"+value.Name.Name)
						}
					case *ast.ValueSpec:
						for _, name := range value.Names {
							if name.IsExported() {
								result = append(result, strings.ToLower(typed.Tok.String())+":"+name.Name)
							}
						}
					}
				}
			case *ast.FuncDecl:
				if !typed.Name.IsExported() {
					continue
				}
				if typed.Recv == nil {
					result = append(result, "func:"+typed.Name.Name)
					continue
				}
				receiver := partyExportedReceiver(typed.Recv.List[0].Type)
				if receiver != "" {
					result = append(result, "method:"+receiver+"."+typed.Name.Name)
				}
			}
		}
	}
	slices.Sort(result)
	return result
}

func partyExportedReceiver(expression ast.Expr) string {
	switch typed := expression.(type) {
	case *ast.Ident:
		if typed.IsExported() {
			return typed.Name
		}
	case *ast.StarExpr:
		return partyExportedReceiver(typed.X)
	}
	return ""
}

func assertPartyStructFields(t testing.TB, structure reflect.Type, expected []string) {
	t.Helper()
	actual := make([]string, 0, structure.NumField())
	for index := range structure.NumField() {
		field := structure.Field(index)
		if field.IsExported() {
			actual = append(actual, field.Name)
		}
	}
	if !slices.Equal(actual, expected) {
		t.Fatalf("%s exported fields = %v, want %v", structure.Name(), actual, expected)
	}
}
