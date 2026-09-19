package networkflow

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

func assertProductionSurface(t *testing.T) {
	allowed := map[string]bool{
		"const GraphViewMaterializationJobKind":                                false,
		"const ImportApplyParticipantID":                                       false,
		"const IndicatorLinkParticipantID":                                     false,
		"const KeyRingManifestMaximumSize":                                     false,
		"const KeyRingsOverrideKey":                                            false,
		"const ManifestTooLarge":                                               false,
		"const ManifestUnreadable":                                             false,
		"const ManifestUnsafe":                                                 false,
		"const ProfileID":                                                      false,
		"const RouteContributionID":                                            false,
		"const SavedGraphCutoverAlgorithmID":                                   false,
		"const WorkspaceContributionID":                                        false,
		"func ApplyConfigurationOverlay":                                       false,
		"func CloneConfiguration":                                              false,
		"func ConfigurationFindingFromError":                                   false,
		"func ExtensionStateFamilyCounters":                                    false,
		"func KeyRingManifestReadError":                                        false,
		"func NewGraphRestoreSourceRegistration":                               false,
		"func NewModule":                                                       false,
		"func NewPortabilityStateBinding":                                      false,
		"func NormalizeAndValidateConfiguration":                               false,
		"func ParseKeyRingsWithRegistry":                                       false,
		"func ReconcileGraphRestoreJobsTx":                                     false,
		"func RecoveryPostgresTables":                                          false,
		"func RecoveryStateContribution":                                       false,
		"func ValidateExtensionState":                                          false,
		"func ValidateRetainedExtensionState":                                  false,
		"func ValidateSavedGraphAdmission":                                     false,
		"method Configuration.EffectiveResourceLimits":                         false,
		"method GraphResultCleanupDispatcher.Close":                            false,
		"method GraphResultCleanupDispatcher.Start":                            false,
		"method GraphViewReceiptReconciler.ReconcileFinalIdempotencyOutcomeTx": false,
		"method Module.ImportOwner":                                            false,
		"method Module.InstallCrossOwnerCoordinator":                           false,
		"method Module.NewGraphResultCleanupDispatcher":                        false,
		"method Module.RegisterGraphViewWorker":                                false,
		"method Module.RegisterRoutes":                                         false,
		"method Module.ReportingGraphSource":                                   false,
		"method Module.TransactionCapabilities":                                false,
		"method PortabilityStateBinding.RetainedAuthoritativeStatePresentTx":   false,
		"method ReportingGraphSource.ReadAndRenewLeasedResult":                 false,
		"method ReportingGraphSource.ReleaseJobLeases":                         false,
		"method ReportingGraphSource.ReleaseJobLeasesTx":                       false,
		"method ReportingGraphSource.SourceOwnerID":                            false,
		"method ReportingGraphSource.ValidateAndLeaseResultTx":                 false,
		"type AdministrativeAuditPort":                                         false,
		"type Configuration":                                                   false,
		"type ConfigurationFinding":                                            false,
		"type EffectiveLimits":                                                 false,
		"type ExtensionStateReader":                                            false,
		"type GraphCleanupTelemetryObservation":                                false,
		"type GraphPhaseTelemetryObservation":                                  false,
		"type GraphResultCleanupDispatcher":                                    false,
		"type GraphResultTelemetryObservation":                                 false,
		"type GraphTelemetryObserver":                                          false,
		"type GraphViewJobFailureFinalization":                                 false,
		"type GraphViewJobFinalizer":                                           false,
		"type GraphViewJobManager":                                             false,
		"type GraphViewJobMutation":                                            false,
		"type GraphViewJobRunner":                                              false,
		"type GraphViewJobSuccessFinalization":                                 false,
		"type GraphViewJobTransactions":                                        false,
		"type GraphViewReceiptReconciler":                                      false,
		"type ImportSourcePort":                                                false,
		"type IncidentLockPort":                                                false,
		"type IndicatorParticipationPort":                                      false,
		"type KeyRings":                                                        false,
		"type ManifestReadFailure":                                             false,
		"type Module":                                                          false,
		"type ModuleDependencies":                                              false,
		"type PortabilityStateBinding":                                         false,
		"type PortabilityStateQuery":                                           false,
		"type ReportingGraphSource":                                            false,
		"type ResourceIntent":                                                  false,
		"type ResourceIntentAppender":                                          false,
		"type ResourceLimitOverrides":                                          false,
	}
	check := func(key string) {
		if _, ok := allowed[key]; !ok {
			t.Errorf("unreviewed Network Flow production export: %s", key)
			return
		}
		allowed[key] = true
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), entry.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			switch d := declaration.(type) {
			case *ast.FuncDecl:
				if !d.Name.IsExported() {
					continue
				}
				if d.Recv == nil {
					check("func " + d.Name.Name)
					continue
				}
				receiver := d.Recv.List[0].Type
				if star, ok := receiver.(*ast.StarExpr); ok {
					receiver = star.X
				}
				name, ok := receiver.(*ast.Ident)
				if !ok {
					t.Fatal("unsupported receiver")
				}
				if name.IsExported() {
					check("method " + name.Name + "." + d.Name.Name)
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						if spec.Name.IsExported() {
							check("type " + spec.Name.Name)
						}
					case *ast.ValueSpec:
						for _, name := range spec.Names {
							if name.IsExported() {
								check(d.Tok.String() + " " + name.Name)
							}
						}
					}
				}
			}
		}
	}
	for name, seen := range allowed {
		if !seen {
			t.Errorf("missing retained Network Flow capability: %s", name)
		}
	}
}
