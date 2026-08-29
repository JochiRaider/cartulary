package evidence

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
)

type exportDisposition string

const (
	exportRetain exportDisposition = "retain"
)

// evidenceExportClassifications is the steady-state reviewed production
// boundary. Every entry has an independent production role; any new export
// requires an explicit owner decision.
var evidenceExportClassifications = map[string]map[string]exportDisposition{
	".": {
		"AdmissionFailure":                            exportRetain,
		"AdmissionFailure.CollectionField":            exportRetain,
		"AdmissionFailure.Error":                      exportRetain,
		"AdmissionFailure.Field":                      exportRetain,
		"AdmissionFailure.MaximumCount":               exportRetain,
		"AdmissionFailure.ReasonCode":                 exportRetain,
		"AdmissionFailure.RequestedCount":             exportRetain,
		"AdmissionReasonCode":                         exportRetain,
		"AdmitConflictResolveJSON":                    exportRetain,
		"AdmitCreateJSON":                             exportRetain,
		"AdmitPatchJSON":                              exportRetain,
		"AttachReasonBlobNotVisible":                  exportRetain,
		"AttachRejectedError":                         exportRetain,
		"AttachRejectedError.Error":                   exportRetain,
		"AttachRejectedError.Unwrap":                  exportRetain,
		"CleanupDispatcher":                           exportRetain,
		"CleanupDispatcher.Close":                     exportRetain,
		"CleanupDispatcher.Start":                     exportRetain,
		"CleanupObserver":                             exportRetain,
		"CleanupSweepObservation":                     exportRetain,
		"ConflictAdmissionContext":                    exportRetain,
		"ConflictResolveAdmission":                    exportRetain,
		"ConflictResolveAdmission.ClientTxnID":        exportRetain,
		"ErrBlobNotFound":                             exportRetain,
		"ErrEvidenceNotFound":                         exportRetain,
		"ErrIncidentMismatch":                         exportRetain,
		"ErrObjectStoreUnavailable":                   exportRetain,
		"ErrClientTxnConflict":                        exportRetain,
		"ErrStoredMutationKindMismatch":               exportRetain,
		"IncidentBundleBlobPortability":               exportRetain,
		"IncidentBundleSubtypeContribution":           exportRetain,
		"LifecycleValidationError":                    exportRetain,
		"LifecycleValidationError.Error":              exportRetain,
		"LifecycleIdempotencyCapability":              exportRetain,
		"LifecycleIdempotencyKey":                     exportRetain,
		"LifecycleOperationBlobAttach":                exportRetain,
		"LifecycleOperationBlobCreate":                exportRetain,
		"LifecycleOperationID":                        exportRetain,
		"OptionalConflictValue":                       exportRetain,
		"NewIncidentBundleBlobPortability":            exportRetain,
		"NewIncidentBundleSourcePort":                 exportRetain,
		"NewProjectionContribution":                   exportRetain,
		"NewRecoveryProvider":                         exportRetain,
		"NewReportingFieldContribution":               exportRetain,
		"NewReportingLogicalTargetContribution":       exportRetain,
		"NewOwnerRuntime":                             exportRetain,
		"OwnerRuntime":                                exportRetain,
		"OwnerRuntime.CleanupDispatcher":              exportRetain,
		"OwnerRuntime.ImportCreateFacade":             exportRetain,
		"OwnerRuntime.RouteRegistrar":                 exportRetain,
		"OwnerRuntime.TimelineAttachmentContribution": exportRetain,
		"OwnerRuntime.MutationContribution":           exportRetain,
		"OwnerRuntimeDependencies":                    exportRetain,
		"PersistedObjectBlobStorageKeyErrorReason":    exportRetain,
		"RecoveryStateContribution":                   exportRetain,
		"RevisionProviderContribution":                exportRetain,
		"RowVersionConflictError":                     exportRetain,
		"RowVersionConflictError.Error":               exportRetain,
		"SameFieldConflictError":                      exportRetain,
		"SameFieldConflictError.Error":                exportRetain,
		"SameFieldConflict":                           exportRetain,
		"Settings":                                    exportRetain,
		"TimelineAttachmentContribution":              exportRetain,
		"TimelineFact":                                exportRetain,
		"TimelineFactReader":                          exportRetain,
		"TimelineFactReader.LoadTx":                   exportRetain,
		"VNextRecoveryObjectInventory":                exportRetain,
		"ValidationError":                             exportRetain,
		"ValidationError.Error":                       exportRetain,
		"ViewSchemaID":                                exportRetain,
		"ConflictCommand":                             exportRetain,
		"MutationContribution":                        exportRetain,
		"MutationDependencies":                        exportRetain,
		"MutationOutcome":                             exportRetain,
		"MutationOutcomeCreated":                      exportRetain,
		"MutationOutcomeKeptSaved":                    exportRetain,
		"MutationOutcomeReplayed":                     exportRetain,
		"MutationOutcomeUpdated":                      exportRetain,
		"CreateCommand":                               exportRetain,
		"CreateAdmission":                             exportRetain,
		"CreateAdmission.ClientTxnID":                 exportRetain,
		"CreateAdmission.ViewSchemaID":                exportRetain,
		"FieldValue":                                  exportRetain,
		"MutationResult":                              exportRetain,
		"PatchCommand":                                exportRetain,
		"PatchAdmission":                              exportRetain,
		"PatchAdmission.BaseRowVersion":               exportRetain,
		"PatchAdmission.ClientTxnID":                  exportRetain,
		"PatchAdmission.ViewSchemaID":                 exportRetain,
		"IdempotencyCapability":                       exportRetain,
		"IdempotencyKey":                              exportRetain,
		"IncidentStateCapability":                     exportRetain,
		"NewStoredCreateResult":                       exportRetain,
		"NewStoredPatchResult":                        exportRetain,
		"OperationConflictResolve":                    exportRetain,
		"OperationCreate":                             exportRetain,
		"OperationID":                                 exportRetain,
		"OperationPatch":                              exportRetain,
		"RecordEnvelopeCapability":                    exportRetain,
		"RevisionCapability":                          exportRetain,
		"StoredMutationCreate":                        exportRetain,
		"StoredMutationKind":                          exportRetain,
		"StoredMutationPatch":                         exportRetain,
		"StoredMutationPayload":                       exportRetain,
		"StoredMutationResult":                        exportRetain,
		"StoredMutationResult.Kind":                   exportRetain,
		"StoredMutationResult.Payload":                exportRetain,
	},
	"blobref": {
		"ObjectBlobStorageKey":      exportRetain,
		"ObjectBlobStorageRef":      exportRetain,
		"ParseObjectBlobStorageKey": exportRetain,
		"ParseObjectBlobStorageRef": exportRetain,
		"StorageKeyParts":           exportRetain,
		"StorageRefScheme":          exportRetain,
	},
	"internal/policy": {
		"AllowedNonTerminalFinalizeFailures":         exportRetain,
		"AssociationBlobAvailable":                   exportRetain,
		"AssociationBlobDisposition":                 exportRetain,
		"AssociationBlobExpired":                     exportRetain,
		"AssociationBlobFailed":                      exportRetain,
		"AssociationBlobInconsistent":                exportRetain,
		"AssociationBlobNeedsFinalization":           exportRetain,
		"AssociationBlobQuarantined":                 exportRetain,
		"BlobAvailable":                              exportRetain,
		"BlobFailed":                                 exportRetain,
		"BlobPending":                                exportRetain,
		"BlobQuarantined":                            exportRetain,
		"ClassifyBlobForAssociation":                 exportRetain,
		"CleanupDelay":                               exportRetain,
		"EvidenceAvailable":                          exportRetain,
		"EvidencePendingReceipt":                     exportRetain,
		"EvidenceQuarantined":                        exportRetain,
		"EvidenceReceived":                           exportRetain,
		"EvidenceReleased":                           exportRetain,
		"EvidenceRequested":                          exportRetain,
		"FailureSchedule":                            exportRetain,
		"FinalizeAttemptIsTerminal":                  exportRetain,
		"IncidentMutationBlocked":                    exportRetain,
		"InitialEvidenceLifecycleDisposition":        exportRetain,
		"InitialLifecycleAllowed":                    exportRetain,
		"InitialLifecycleDisposition":                exportRetain,
		"InitialLifecycleGuardViolation":             exportRetain,
		"InitialLifecycleIllegalTransition":          exportRetain,
		"InitialLifecycleInvalid":                    exportRetain,
		"IsServerManagedStorageRef":                  exportRetain,
		"LegalBlobTransition":                        exportRetain,
		"LegalEvidenceTransition":                    exportRetain,
		"ObjectBlobStorageKeyIdentityMismatchReason": exportRetain,
		"ObjectBlobStorageKeyMalformedReason":        exportRetain,
		"PersistedObjectBlobStorageKeyError":         exportRetain,
		"PersistedObjectBlobStorageKeyError.Error":   exportRetain,
		"PersistedObjectBlobStorageKeyErrorReason":   exportRetain,
		"QuarantineClearAdmin":                       exportRetain,
		"QuarantineClearContentInspection":           exportRetain,
		"QuarantineTriggerAdmin":                     exportRetain,
		"QuarantineTriggerContentInspection":         exportRetain,
		"ScheduleFailure":                            exportRetain,
		"ServerManagedStorageRefMatchesAssociation":  exportRetain,
		"TerminalFinalizeAttempt":                    exportRetain,
		"ValidBlobUploadState":                       exportRetain,
		"ValidEvidenceLifecycle":                     exportRetain,
		"ValidQuarantineClearTrigger":                exportRetain,
		"ValidQuarantineEntryTrigger":                exportRetain,
		"ValidatePersistedObjectBlobStorageKey":      exportRetain,
		"ViolatesEvidenceBlobBridge":                 exportRetain,
	},
	"projectioncontract": {
		"Contribution":                        exportRetain,
		"Contribution.ProjectionContribution": exportRetain,
		"Contribution.Source":                 exportRetain,
		"Descriptor":                          exportRetain,
		"NewContribution":                     exportRetain,
		"ProjectionInput":                     exportRetain,
		"ProjectionInputPage":                 exportRetain,
		"SourceReader":                        exportRetain,
		"SurfaceIntent":                       exportRetain,
		"ViewRow":                             exportRetain,
	},
	"projectionports": {
		"AssociationEffects":               exportRetain,
		"EvidenceAffectedViewChange":       exportRetain,
		"EvidenceAssociationEffectsInput":  exportRetain,
		"EvidenceAssociationEffectsResult": exportRetain,
		"EvidenceAssociationSubject":       exportRetain,
		"EvidenceSupportRowChange":         exportRetain,
		"EvidenceViewRowPatch":             exportRetain,
		"MutationRows":                     exportRetain,
		"SupportChangeInvalidate":          exportRetain,
		"SupportChangeKind":                exportRetain,
		"SupportChangePatch":               exportRetain,
	},
}

func TestEvidenceExportedSurfaceReachabilityLock(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve Evidence package path")
	}
	root := filepath.Dir(currentFile)
	for _, relative := range []string{
		"deleterestore",
		"projectionprovider",
		"recoveryprovider",
		"reportingprovider",
		"rollbackprovider",
		"workbookprojection",
	} {
		if _, err := os.Stat(filepath.Join(root, relative)); !os.IsNotExist(err) {
			t.Fatalf("retired Evidence package %s remains reachable: %v", relative, err)
		}
	}
	var mismatches []string
	for relative, expected := range evidenceExportClassifications {
		directory := filepath.Join(root, relative)
		actual := exportedPackageDeclarations(t, directory)
		allowed := make([]string, 0, len(expected))
		for name := range expected {
			allowed = append(allowed, name)
		}
		unapproved, missing := exportSurfaceDelta(actual, allowed)
		if len(unapproved) != 0 || len(missing) != 0 {
			mismatches = append(mismatches, relative+": unapproved="+strings.Join(unapproved, ",")+" missing="+strings.Join(missing, ","))
		}
	}
	if len(mismatches) != 0 {
		sort.Strings(mismatches)
		t.Fatalf("Evidence exported surface mismatch:\n%s", strings.Join(mismatches, "\n"))
	}

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
}

func exportedPackageDeclarations(t testing.TB, directory string) []string {
	t.Helper()
	entries, err := filepath.Glob(filepath.Join(directory, "*.go"))
	if err != nil {
		t.Fatalf("list package directory %s: %v", directory, err)
	}
	fset := token.NewFileSet()
	var declarations []string
	for _, path := range entries {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		declarations = append(declarations, exportedDeclarations(file)...)
	}
	return declarations
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
	switch typed := expression.(type) {
	case *ast.Ident:
		if ast.IsExported(typed.Name) {
			return typed.Name
		}
	case *ast.StarExpr:
		return exportedReceiverName(typed.X)
	}
	return ""
}

func exportSurfaceDelta(actual []string, expected []string) ([]string, []string) {
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
