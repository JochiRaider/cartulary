package assessments_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestAssessmentExportSurface(t *testing.T) {
	t.Parallel()

	for _, packageTest := range []struct {
		name        string
		directory   string
		packageName string
		expected    []string
	}{
		{
			name:        "root",
			directory:   ".",
			packageName: "assessments",
			expected:    assessmentRootExportSurface,
		},
		{
			name:        "workbook projection",
			directory:   "workbookprojection",
			packageName: "workbookprojection",
			expected:    assessmentWorkbookProjectionExportSurface,
		},
	} {
		packageTest := packageTest
		t.Run(packageTest.name, func(t *testing.T) {
			t.Parallel()
			actual, err := exportedSurfaceFromDirectory(packageTest.directory, packageTest.packageName)
			if err != nil {
				t.Fatalf("inspect exported surface: %v", err)
			}
			assertExactExportSurface(t, packageTest.expected, actual)
		})
	}
}

func TestAssessmentExportSurfaceRejectsInjectedExport(t *testing.T) {
	t.Parallel()

	file, err := parser.ParseFile(token.NewFileSet(), "injected.go", `package assessments

func InjectedConvenienceExport() {}
`, 0)
	if err != nil {
		t.Fatalf("parse synthetic negative fixture: %v", err)
	}
	actual := exportedSurface("assessments", []*ast.File{file})
	if differences := exportSurfaceDifferences([]string{"package:assessments"}, actual); len(differences) != 1 || differences[0] != "unexpected func:InjectedConvenienceExport" {
		t.Fatalf("synthetic export differences = %v", differences)
	}
	if differences := exportSurfaceDifferences(actual, []string{"package:assessments"}); len(differences) != 1 || differences[0] != "missing func:InjectedConvenienceExport" {
		t.Fatalf("synthetic missing-export differences = %v", differences)
	}
}

func exportedSurfaceFromDirectory(directory string, packageName string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	files := make([]*ast.File, 0, len(entries))
	fileSet := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, parseErr := parser.ParseFile(fileSet, filepath.Join(directory, entry.Name()), nil, 0)
		if parseErr != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), parseErr)
		}
		if file.Name.Name != packageName {
			return nil, fmt.Errorf("%s declares package %q, want %q", entry.Name(), file.Name.Name, packageName)
		}
		files = append(files, file)
	}
	return exportedSurface(packageName, files), nil
}

func exportedSurface(packageName string, files []*ast.File) []string {
	entries := map[string]struct{}{"package:" + packageName: {}}
	for _, file := range files {
		for _, declaration := range file.Decls {
			switch typed := declaration.(type) {
			case *ast.GenDecl:
				collectExportedGeneralDeclaration(entries, typed)
			case *ast.FuncDecl:
				if !typed.Name.IsExported() {
					continue
				}
				if typed.Recv == nil {
					entries["func:"+typed.Name.Name] = struct{}{}
					continue
				}
				entries["method:"+receiverTypeName(typed.Recv.List[0].Type)+"."+typed.Name.Name] = struct{}{}
			}
		}
	}
	result := make([]string, 0, len(entries))
	for entry := range entries {
		result = append(result, entry)
	}
	sort.Strings(result)
	return result
}

func collectExportedGeneralDeclaration(entries map[string]struct{}, declaration *ast.GenDecl) {
	for _, specification := range declaration.Specs {
		switch typed := specification.(type) {
		case *ast.ValueSpec:
			for _, name := range typed.Names {
				if name.IsExported() {
					entries[declaration.Tok.String()+":"+name.Name] = struct{}{}
				}
			}
		case *ast.TypeSpec:
			if !typed.Name.IsExported() {
				continue
			}
			entries["type:"+typed.Name.Name] = struct{}{}
			switch body := typed.Type.(type) {
			case *ast.StructType:
				collectExportedFields(entries, "field:"+typed.Name.Name+".", body.Fields)
			case *ast.InterfaceType:
				collectExportedFields(entries, "interface:"+typed.Name.Name+".", body.Methods)
			}
		}
	}
}

func collectExportedFields(entries map[string]struct{}, prefix string, fields *ast.FieldList) {
	for _, field := range fields.List {
		if len(field.Names) == 0 {
			name := receiverTypeName(field.Type)
			if ast.IsExported(name) {
				entries[prefix+name] = struct{}{}
			}
			continue
		}
		for _, name := range field.Names {
			if name.IsExported() {
				entries[prefix+name.Name] = struct{}{}
			}
		}
	}
}

func receiverTypeName(expression ast.Expr) string {
	switch typed := expression.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return receiverTypeName(typed.X)
	case *ast.IndexExpr:
		return receiverTypeName(typed.X)
	case *ast.IndexListExpr:
		return receiverTypeName(typed.X)
	case *ast.SelectorExpr:
		return typed.Sel.Name
	default:
		return fmt.Sprintf("%T", expression)
	}
}

func assertExactExportSurface(t *testing.T, expected []string, actual []string) {
	t.Helper()
	if differences := exportSurfaceDifferences(expected, actual); len(differences) != 0 {
		t.Fatalf("exported surface differs:\n%s", strings.Join(differences, "\n"))
	}
}

func exportSurfaceDifferences(expected []string, actual []string) []string {
	expectedSet := make(map[string]struct{}, len(expected))
	actualSet := make(map[string]struct{}, len(actual))
	for _, entry := range expected {
		expectedSet[entry] = struct{}{}
	}
	for _, entry := range actual {
		actualSet[entry] = struct{}{}
	}
	differences := make([]string, 0)
	for entry := range expectedSet {
		if _, ok := actualSet[entry]; !ok {
			differences = append(differences, "missing "+entry)
		}
	}
	for entry := range actualSet {
		if _, ok := expectedSet[entry]; !ok {
			differences = append(differences, "unexpected "+entry)
		}
	}
	sort.Strings(differences)
	return differences
}

var assessmentRootExportSurface = strings.Fields(`
const:AssessmentsViewSchemaID
const:CreateOutcomeCommitted
const:CreateOutcomeReplayed
field:CreateCommand.ActorUserID
field:CreateCommand.Idempotency
field:CreateCommand.IncidentID
field:CreateCommand.Input
field:CreateCommand.Now
field:CreateCommand.RequestID
field:CreateIdempotencyKey.ActorUserID
field:CreateIdempotencyKey.ClientTxnID
field:CreateIdempotencyKey.RequestHash
field:CreateIdempotencyKey.RouteKey
field:CreateIdempotencyKey.ScopeKey
field:CreateIdempotencyRecord.RequestHash
field:CreateIdempotencyRecord.Result
field:CreateInput.AssessedAt
field:CreateInput.AssessmentState
field:CreateInput.Assessor
field:CreateInput.ClientTxnID
field:CreateInput.ConfidenceScore
field:CreateInput.Rationale
field:CreateInput.SubjectRef
field:CreateInput.SubjectType
field:CreateInput.SupportRefs
field:CreateResult.CanonicalRow
field:CreateResult.ChangeSetID
field:CreateResult.Outcome
field:CreateResult.RecordID
field:CreateResult.RowVersion
field:CreateRevision.ActorUserID
field:CreateRevision.AfterVersion
field:CreateRevision.CanonicalRow
field:CreateRevision.ClientTxnID
field:CreateRevision.CreatedAt
field:CreateRevision.IncidentID
field:CreateRevision.OperationKind
field:CreateRevision.RecordID
field:CreateRevision.RequestID
field:CreateRevision.RouteKey
field:CreateRevision.RowVersion
field:CreateRevision.TargetKind
field:CreateValidationError.Field
field:CreateValidationError.ReasonCode
field:FacadeDependencies.Assessors
field:FacadeDependencies.Idempotency
field:FacadeDependencies.Projections
field:FacadeDependencies.Records
field:FacadeDependencies.Revisions
field:FacadeDependencies.Subjects
field:FacadeDependencies.SupportLinks
field:FacadeDependencies.SupportTargets
field:ImportCreateDependencies.Assessors
field:ImportCreateDependencies.Projections
field:ImportCreateDependencies.Records
field:ImportCreateDependencies.Revisions
field:ImportCreateDependencies.Subjects
field:MergeMutation.AfterSnapshot
field:MergeMutation.AfterValue
field:MergeMutation.BeforeSnapshot
field:MergeMutation.BeforeValue
field:MergeMutation.OperationKind
field:MergeMutation.TargetID
field:MergeMutation.TargetKind
field:MergeProtectedSetChangedError.RecordID
field:ProjectionContributionDependencies.Envelopes
field:ProjectionContributionDependencies.Support
field:RecordEnvelopeCreate.ActorID
field:RecordEnvelopeCreate.IncidentID
field:RecordEnvelopeCreate.Now
field:RecordEnvelopeCreate.RecordType
field:RecordEnvelopeCreate.RowVersion
func:IncidentBundleSubtypeContribution
func:NewFacade
func:NewImportCreateFacade
func:NewIncidentBundleSourcePort
func:NewMergeEffects
func:NewProjectionContribution
func:RecoveryStateContribution
func:RevisionProviderContribution
interface:AssessmentProjectionPort.RefreshAndLoadAssessmentRowTx
interface:AssessorValidator.ValidateAssessmentAssessorTx
interface:CreateIdempotencyPort.LookupCreate
interface:CreateIdempotencyPort.StoreCreateTx
interface:CreateRevisionAppender.AppendAssessmentCreateRevisionTx
interface:InitialSupportLinkApplier.ApplyInitialAssessmentSupportLinksTx
interface:MergeProjectionPort.RefreshAssessmentProjectionTx
interface:MergeSnapshotCapturePort.CaptureRecordSnapshotTx
interface:RecordEnvelopeCreator.CreateAssessmentEnvelopeTx
interface:SubjectValidator.ValidateAssessmentSubjectTx
interface:SupportTargetValidator.ValidateAssessmentSupportTargetsTx
method:CreateValidationError.Error
method:Facade.Create
method:MergeEffects.LoadProtectedRecordIDsTx
method:MergeEffects.RepointTx
method:MergeProtectedSetChangedError.Error
method:assessmentSourceRepository.InsertTx
method:importCreateFacade.CreateImportRowTx
package:assessments
type:AssessmentProjectionPort
type:AssessorValidator
type:CreateCommand
type:CreateIdempotencyKey
type:CreateIdempotencyPort
type:CreateIdempotencyRecord
type:CreateInput
type:CreateOutcome
type:CreateResult
type:CreateRevision
type:CreateRevisionAppender
type:CreateValidationError
type:Facade
type:FacadeDependencies
type:ImportCreateDependencies
type:InitialSupportLinkApplier
type:MergeEffects
type:MergeMutation
type:MergeProjectionPort
type:MergeProtectedSetChangedError
type:MergeSnapshotCapturePort
type:ProjectionContributionDependencies
type:RecordEnvelopeCreate
type:RecordEnvelopeCreator
type:SubjectValidator
type:SupportTargetValidator
var:ErrClientTxnConflict
`)

var assessmentWorkbookProjectionExportSurface = strings.Fields(`
const:ProjectionMutationDelete
const:ProjectionMutationUpsert
field:Envelope.DeletedAt
field:Envelope.IncidentID
field:Envelope.RecordID
field:Envelope.RecordType
field:Envelope.RowVersion
field:Ports.Rebuilder
field:Ports.Rows
field:ProjectionInput.AssessedAt
field:ProjectionInput.AssessmentState
field:ProjectionInput.Assessor
field:ProjectionInput.ConfidenceBand
field:ProjectionInput.ConfidenceScore
field:ProjectionInput.IncidentID
field:ProjectionInput.Rationale
field:ProjectionInput.RecordID
field:ProjectionInput.RowVersion
field:ProjectionInput.SubjectRef
field:ProjectionInput.SubjectType
field:ProjectionInput.SupportingLinkCount
field:ProjectionInputPage.Inputs
field:ProjectionInputPage.NextRecordID
field:ProjectionMutation.Input
field:ProjectionMutation.Kind
field:ProjectionMutation.RecordID
field:SupportFacts.ActiveTargetCount
func:Descriptor
func:NewContribution
func:SurfaceIntent
interface:EnvelopeReader.LoadAssessmentProjectionEnvelopeTx
interface:Rebuilder.RebuildAssessments
interface:Rows.ApplyAssessmentMutationTx
interface:Rows.LoadAssessmentTx
interface:Rows.RebuildAssessmentsTx
interface:Rows.RefreshAssessmentTx
interface:SourceReader.BuildProjectionMutationTx
interface:SourceReader.ListProjectionInputsTx
interface:SupportFactReader.LoadAssessmentProjectionSupportFactsTx
method:Contribution.ProjectionContribution
method:Contribution.Source
method:ProjectionInput.Validate
method:ProjectionMutation.Validate
package:workbookprojection
type:Contribution
type:Envelope
type:EnvelopeReader
type:Ports
type:ProjectionInput
type:ProjectionInputPage
type:ProjectionMutation
type:ProjectionMutationKind
type:Rebuilder
type:Rows
type:SourceReader
type:SupportFactReader
type:SupportFacts
`)
