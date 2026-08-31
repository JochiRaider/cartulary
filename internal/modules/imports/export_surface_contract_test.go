package imports

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

type exportRoleGroup struct {
	role string
	keys []string
}

var importsRootExportRoles = mustExportRoles(
	exportRoleGroup{
		role: "module composition and admitted runtime publication",
		keys: []string{
			"const:ApplyJobKind",
			"const:DiscoveryJobKind",
			"const:ImportTargetKindViewSchema",
			"const:ProfileID",
			"field:Limits.MaxCSVSourceBytes",
			"field:Limits.MaxCells",
			"field:Limits.MaxColumns",
			"field:Limits.MaxRows",
			"field:Limits.MaxXLSXSourceBytes",
			"field:ModuleDependencies.ArchiveLimits",
			"field:ModuleDependencies.CursorCodec",
			"field:ModuleDependencies.Env",
			"field:ModuleDependencies.ExtensionImportFacades",
			"field:ModuleDependencies.ExtensionProfileAdmission",
			"field:ModuleDependencies.JobOperations",
			"field:ModuleDependencies.JobRunner",
			"field:ModuleDependencies.JobSuccessFinalizer",
			"field:ModuleDependencies.JobTransactions",
			"field:ModuleDependencies.Limits",
			"field:ModuleDependencies.Now",
			"field:ModuleDependencies.OwnerCreateRegistry",
			"field:ModuleDependencies.Postgres",
			"field:ModuleDependencies.RevisionAppender",
			"func:NewModule",
			"method:Module.RegisterRoutes",
			"method:Module.RegisterWorkers",
			"type:ArchiveLimits",
			"type:Limits",
			"type:Module",
			"type:ModuleDependencies",
		},
	},
	exportRoleGroup{
		role: "typed analytical-owner facade protocol",
		keys: []string{
			"field:ExtensionImportApplyRequest.ActorUserID",
			"field:ExtensionImportApplyRequest.ClientTxnID",
			"field:ExtensionImportApplyRequest.ExpectedSourceContentSHA256",
			"field:ExtensionImportApplyRequest.ExtensionProfileID",
			"field:ExtensionImportApplyRequest.ImportSessionID",
			"field:ExtensionImportApplyRequest.ImportUnitID",
			"field:ExtensionImportApplyRequest.IncidentID",
			"field:ExtensionImportApplyRequest.MappingFingerprint",
			"field:ExtensionImportApplyRequest.OwnerMapping",
			"field:ExtensionImportApplyRequest.OwnerMappingSchemaID",
			"field:ExtensionImportApplyRequest.SourceCapability",
			"field:ExtensionImportApplyRequest.TargetKind",
			"field:ExtensionImportApplyResult.OwnerResponse",
			"field:ExtensionImportApplyResult.ResourceRefs",
			"field:ExtensionImportErrorTranslation.CoreReasonCode",
			"field:ExtensionImportErrorTranslation.ErrorSchemaID",
			"field:ExtensionImportErrorTranslation.ErrorTranslationID",
			"field:ExtensionImportErrorTranslation.OwnerError",
			"field:ExtensionImportFacadeBinding.ApplyRequestSchemaID",
			"field:ExtensionImportFacadeBinding.ApplyResultSchemaID",
			"field:ExtensionImportFacadeBinding.CommitProtocolID",
			"field:ExtensionImportFacadeBinding.ContractMajor",
			"field:ExtensionImportFacadeBinding.ErrorSchemaID",
			"field:ExtensionImportFacadeBinding.ErrorTranslationID",
			"field:ExtensionImportFacadeBinding.ExtensionProfileID",
			"field:ExtensionImportFacadeBinding.FacadeID",
			"field:ExtensionImportFacadeBinding.MappingSchemaID",
			"field:ExtensionImportFacadeBinding.OwnerContractRef",
			"field:ExtensionImportFacadeBinding.PreviewRequestSchemaID",
			"field:ExtensionImportFacadeBinding.PreviewResultSchemaID",
			"field:ExtensionImportFacadeBinding.SchemaID",
			"field:ExtensionImportFacadeBinding.TargetKind",
			"field:ExtensionImportMappingRequest.ActorUserID",
			"field:ExtensionImportMappingRequest.ClientTxnID",
			"field:ExtensionImportMappingRequest.ExtensionProfileID",
			"field:ExtensionImportMappingRequest.ImportSessionID",
			"field:ExtensionImportMappingRequest.ImportUnitID",
			"field:ExtensionImportMappingRequest.IncidentID",
			"field:ExtensionImportMappingRequest.OwnerMapping",
			"field:ExtensionImportMappingRequest.OwnerMappingSchemaID",
			"field:ExtensionImportMappingRequest.SourceCapability",
			"field:ExtensionImportMappingRequest.TargetKind",
			"field:ExtensionImportMappingResult.MappingFingerprint",
			"field:ExtensionImportMappingResult.OwnerMapping",
			"field:ExtensionImportMappingResult.OwnerResult",
			"field:ExtensionImportMappingResult.OwnerResultSchemaID",
			"field:ExtensionImportOwnerError.OwnerCode",
			"field:ExtensionImportOwnerError.Retryable",
			"field:ExtensionImportOwnerError.SafeDetails",
			"field:ExtensionImportOwnerError.SchemaID",
			"interface:ExtensionImportFacade.ApplyImportUnitTx",
			"interface:ExtensionImportFacade.Binding",
			"interface:ExtensionImportFacade.PrepareImportUnitMapping",
			"interface:ExtensionImportFacade.TranslateImportUnitError",
			"interface:ExtensionImportFacade.ValidateImportUnitError",
			"interface:ExtensionImportFacade.ValidateImportUnitMappingResult",
			"type:ExtensionImportApplyRequest",
			"type:ExtensionImportApplyResult",
			"type:ExtensionImportErrorTranslation",
			"type:ExtensionImportFacade",
			"type:ExtensionImportFacadeBinding",
			"type:ExtensionImportMappingRequest",
			"type:ExtensionImportMappingResult",
			"type:ExtensionImportOwnerError",
		},
	},
	exportRoleGroup{
		role: "exact analytical source capability",
		keys: []string{
			"field:ImportSourceCapability.ImportSessionID",
			"field:ImportSourceCapability.ImportUnitID",
			"field:ImportSourceCapability.SourceByteSize",
			"field:ImportSourceCapability.SourceContentSHA256",
			"field:ImportSourceCapability.SourceMediaType",
			"field:ImportSourceCapability.SourceStreamRef",
			"field:ImportSourceStream.ImportSourceCapability",
			"field:ImportSourceStream.OriginalFilename",
			"field:ImportSourceStream.Reader",
			"func:NewExtensionSourcePort",
			"interface:ExtensionSourcePort.OpenSourceStream",
			"interface:ExtensionSourcePort.ValidateExtensionApplyPreconditionsTx",
			"method:extensionSourcePort.OpenSourceStream",
			"method:extensionSourcePort.ValidateExtensionApplyPreconditionsTx",
			"type:ExtensionSourcePort",
			"type:ImportSourceCapability",
			"type:ImportSourceStream",
		},
	},
	exportRoleGroup{
		role: "Jobs terminal-state finalization protocol",
		keys: []string{
			"field:JobCancellationFinalization.Completion",
			"field:JobCancellationFinalization.Execution",
			"field:JobCancellationFinalization.Mutate",
			"field:JobFailureFinalization.Completion",
			"field:JobFailureFinalization.Execution",
			"field:JobFailureFinalization.Mutate",
			"field:JobSuccessFinalization.Completion",
			"field:JobSuccessFinalization.Execution",
			"field:JobSuccessFinalization.FinalCommitID",
			"field:JobSuccessFinalization.Mutate",
			"interface:JobSuccessFinalizer.FinalizeImportJobCancellation",
			"interface:JobSuccessFinalizer.FinalizeImportJobFailure",
			"interface:JobSuccessFinalizer.FinalizeImportJobSuccess",
			"type:JobCancellationFinalization",
			"type:JobFailureFinalization",
			"type:JobSuccessFinalization",
			"type:JobSuccessFinalizer",
			"type:JobSuccessMutation",
		},
	},
	exportRoleGroup{
		role: "owner-create registry join",
		keys: []string{
			"func:NewOwnerCreateRegistry",
		},
	},
	exportRoleGroup{
		role: "recovery and revision contributions",
		keys: []string{
			"func:RecoveryStateContribution",
			"func:VNextRecoveryObjectInventory",
			"method:revisionAppendAdapter.AppendChangeSetTx",
		},
	},
	exportRoleGroup{
		role: "standard error interface implementation",
		keys: []string{
			"method:applyBlockedError.Error",
			"method:applyBlockedError.Unwrap",
			"method:stateConflictError.Error",
			"method:stateConflictError.Unwrap",
			"method:translatedImportUnitError.Error",
			"method:translatedImportUnitError.Unwrap",
		},
	},
)

var ownerFacadeExportRoles = mustExportRoles(
	exportRoleGroup{
		role: "cross-owner scalar closed union",
		keys: []string{
			"const:ImportScalarBool",
			"const:ImportScalarCollectionToken",
			"const:ImportScalarNull",
			"const:ImportScalarNumber",
			"const:ImportScalarText",
			"const:ImportScalarTimestamp",
			"const:ImportScalarUUID",
			"field:ImportCollectionToken.NormalizedText",
			"field:ImportCollectionToken.RawText",
			"func:IndexImportFieldValues",
			"func:NewBoolImportScalar",
			"func:NewCollectionTokenImportScalar",
			"func:NewNullImportScalar",
			"func:NewNumberImportScalar",
			"func:NewTextImportScalar",
			"func:NewTimestampImportScalar",
			"func:NewUUIDImportScalar",
			"func:NormalizeImportScalar",
			"method:ImportScalarValue.Bool",
			"method:ImportScalarValue.CollectionToken",
			"method:ImportScalarValue.IsValid",
			"method:ImportScalarValue.Kind",
			"method:ImportScalarValue.Number",
			"method:ImportScalarValue.Text",
			"method:ImportScalarValue.Timestamp",
			"method:ImportScalarValue.UUID",
			"type:ImportCollectionToken",
			"type:ImportScalarKind",
			"type:ImportScalarValue",
		},
	},
	exportRoleGroup{
		role: "cross-owner normalized create request",
		keys: []string{
			"field:ImportFieldValue.CellKind",
			"field:ImportFieldValue.EmptyValuePolicy",
			"field:ImportFieldValue.EntityBindingMode",
			"field:ImportFieldValue.FieldKey",
			"field:ImportFieldValue.NormalizedValue",
			"field:ImportFieldValue.RawValue",
			"field:ImportFieldValue.SourceColumnOrdinal",
			"field:ImportFieldValue.SourceHeaderText",
			"field:ImportFieldValue.TransformID",
			"field:ImportOwnerCreateRequest.ActorUserID",
			"field:ImportOwnerCreateRequest.ClientTxnID",
			"field:ImportOwnerCreateRequest.FieldValues",
			"field:ImportOwnerCreateRequest.ImportSessionID",
			"field:ImportOwnerCreateRequest.ImportUnitID",
			"field:ImportOwnerCreateRequest.IncidentID",
			"field:ImportOwnerCreateRequest.Locator",
			"field:ImportOwnerCreateRequest.LocatorKind",
			"field:ImportOwnerCreateRequest.MappingFingerprint",
			"field:ImportOwnerCreateRequest.ParserProfileID",
			"field:ImportOwnerCreateRequest.ParserVersion",
			"field:ImportOwnerCreateRequest.SourceContentSHA256",
			"field:ImportOwnerCreateRequest.SourceFileKind",
			"field:ImportOwnerCreateRequest.SourceRectA1",
			"field:ImportOwnerCreateRequest.SourceRowProvenance",
			"field:ImportOwnerCreateRequest.SourceRowRef",
			"field:ImportOwnerCreateRequest.TargetViewSchemaID",
			"field:ImportOwnerCreateRequest.UnknownValues",
			"field:ImportSourceRowProvenance.SourceRowRef",
			"field:ImportUnknownValue.CellKind",
			"field:ImportUnknownValue.RawValue",
			"field:ImportUnknownValue.SourceColumnOrdinal",
			"field:ImportUnknownValue.SourceHeaderText",
			"type:ImportFieldValue",
			"type:ImportOwnerCreateRequest",
			"type:ImportSourceRowProvenance",
			"type:ImportUnknownValue",
		},
	},
	exportRoleGroup{
		role: "cross-owner create response",
		keys: []string{
			"field:ImportOwnerCreateResponse.ChangeSetMutationRef",
			"field:ImportOwnerCreateResponse.CreatedOrReused",
			"field:ImportOwnerCreateResponse.OwnerResultCode",
			"field:ImportOwnerCreateResponse.RecordID",
			"field:ImportOwnerCreateResponse.RowRefresh",
			"field:ImportOwnerCreateResponse.RowVersion",
			"type:ImportOwnerCreateResponse",
		},
	},
	exportRoleGroup{
		role: "cross-owner create error protocol",
		keys: []string{
			"const:ImportOwnerCreateValidationFailed",
			"field:ImportOwnerCreateError.ConflictingFieldKeys",
			"field:ImportOwnerCreateError.Field",
			"field:ImportOwnerCreateError.Guard",
			"field:ImportOwnerCreateError.OwnerCode",
			"field:ImportOwnerCreateError.PartyReasonCode",
			"field:ImportOwnerCreateError.ReasonCode",
			"field:ImportOwnerCreateError.Retryable",
			"func:ImportOwnerCreateErrorDetail",
			"func:NewImportOwnerCreateValidationError",
			"func:NewPartyMatchConflictError",
			"method:ImportOwnerCreateError.Error",
			"method:ImportOwnerCreateError.Unwrap",
			"type:ImportOwnerCreateError",
		},
	},
	exportRoleGroup{
		role: "cross-owner create facade binding and dispatch",
		keys: []string{
			"field:ImportOwnerCreateBinding.FacadeID",
			"field:ImportOwnerCreateBinding.TargetViewSchemaID",
			"field:ImportOwnerCreateCommand.ChangeSetID",
			"field:ImportOwnerCreateCommand.MutationSequencer",
			"field:ImportOwnerCreateCommand.Now",
			"field:ImportOwnerCreateCommand.Request",
			"field:ImportOwnerCreateCommand.SequenceNo",
			"func:NewImportOwnerCreateFacade",
			"func:NewImportOwnerCreateFacadeWithNormalizer",
			"func:NewImportOwnerCreateRegistry",
			"interface:ImportOwnerCreateFacade.CreateImportRowTx",
			"interface:ImportOwnerCreateFacade.ImportOwnerCreateBinding",
			"interface:ImportOwnerCreateFacade.NormalizeImportField",
			"interface:ImportOwnerCreateTx.CreateImportRowTx",
			"method:ImportOwnerCreateRegistry.Bindings",
			"method:ImportOwnerCreateRegistry.Resolve",
			"method:boundImportOwnerCreateFacade.CreateImportRowTx",
			"method:boundImportOwnerCreateFacade.ImportOwnerCreateBinding",
			"method:boundImportOwnerCreateFacade.NormalizeImportField",
			"type:ImportOwnerCreateBinding",
			"type:ImportOwnerCreateCommand",
			"type:ImportOwnerCreateFacade",
			"type:ImportOwnerCreateFunc",
			"type:ImportOwnerCreateRegistry",
			"type:ImportOwnerCreateTx",
			"type:ImportOwnerNormalizeFunc",
		},
	},
	exportRoleGroup{
		role: "cross-owner mutation sequencing",
		keys: []string{
			"func:NewImportMutationSequencer",
			"method:ImportMutationSequencer.Allocate",
			"method:ImportOwnerCreateCommand.AllocateMutationSequence",
			"type:ImportMutationSequencer",
		},
	},
)

func mustExportRoles(groups ...exportRoleGroup) map[string]string {
	roles := make(map[string]string)
	for _, group := range groups {
		if strings.TrimSpace(group.role) == "" {
			panic("export role must not be empty")
		}
		for _, key := range group.keys {
			if prior, exists := roles[key]; exists {
				panic(fmt.Sprintf("export %q has duplicate roles %q and %q", key, prior, group.role))
			}
			roles[key] = group.role
		}
	}
	return roles
}

func TestImportsExportSurfaceHasExactProductionRoles(t *testing.T) {
	assertImportsExportSurfaceHasExactProductionRoles(t)
}

func assertImportsExportSurfaceHasExactProductionRoles(t *testing.T) {
	t.Helper()
	tests := []struct {
		name  string
		path  string
		roles map[string]string
	}{
		{name: "root", path: ".", roles: importsRootExportRoles},
		{name: "ownerfacade", path: "ownerfacade", roles: ownerFacadeExportRoles},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := collectExportSurface(test.path)
			if err != nil {
				t.Fatalf("collect export surface: %v", err)
			}
			if err := validateExportRoles(actual, test.roles); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestImportsExportSurfaceGuardRejectsUnclassifiedSyntheticExport(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "fixture.go", `
package fixture
type Allowed struct { Field string }
type Surprise struct { Value string }
`, 0)
	if err != nil {
		t.Fatalf("parse synthetic export fixture: %v", err)
	}
	actual := map[string]struct{}{}
	collectFileExportSurface(file, actual)
	roles := map[string]string{
		"type:Allowed":        "fixture role",
		"field:Allowed.Field": "fixture role",
	}
	if err := validateExportRoles(actual, roles); err == nil || !strings.Contains(err.Error(), "Surprise") {
		t.Fatalf("synthetic unclassified export was not rejected: %v", err)
	}
}

func collectExportSurface(path string) (map[string]struct{}, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	actual := map[string]struct{}{}
	files := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(files, filepath.Join(path, entry.Name()), nil, 0)
		if err != nil {
			return nil, err
		}
		collectFileExportSurface(parsed, actual)
	}
	return actual, nil
}

func collectFileExportSurface(file *ast.File, actual map[string]struct{}) {
	for _, declaration := range file.Decls {
		switch typed := declaration.(type) {
		case *ast.GenDecl:
			for _, specification := range typed.Specs {
				switch spec := specification.(type) {
				case *ast.TypeSpec:
					if !spec.Name.IsExported() {
						continue
					}
					actual["type:"+spec.Name.Name] = struct{}{}
					switch shape := spec.Type.(type) {
					case *ast.StructType:
						for _, field := range shape.Fields.List {
							if len(field.Names) == 0 {
								if name := embeddedFieldName(field.Type); ast.IsExported(name) {
									actual["field:"+spec.Name.Name+"."+name] = struct{}{}
								}
							}
							for _, name := range field.Names {
								if name.IsExported() {
									actual["field:"+spec.Name.Name+"."+name.Name] = struct{}{}
								}
							}
						}
					case *ast.InterfaceType:
						for _, method := range shape.Methods.List {
							for _, name := range method.Names {
								if name.IsExported() {
									actual["interface:"+spec.Name.Name+"."+name.Name] = struct{}{}
								}
							}
						}
					}
				case *ast.ValueSpec:
					kind := "var:"
					if typed.Tok == token.CONST {
						kind = "const:"
					}
					for _, name := range spec.Names {
						if name.IsExported() {
							actual[kind+name.Name] = struct{}{}
						}
					}
				}
			}
		case *ast.FuncDecl:
			if !typed.Name.IsExported() {
				continue
			}
			if typed.Recv == nil {
				actual["func:"+typed.Name.Name] = struct{}{}
				continue
			}
			if receiver := receiverTypeName(typed.Recv.List[0].Type); receiver != "" {
				actual["method:"+receiver+"."+typed.Name.Name] = struct{}{}
			}
		}
	}
}

func embeddedFieldName(expression ast.Expr) string {
	switch typed := expression.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return embeddedFieldName(typed.X)
	case *ast.SelectorExpr:
		return typed.Sel.Name
	default:
		return ""
	}
}

func receiverTypeName(expression ast.Expr) string {
	switch typed := expression.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return receiverTypeName(typed.X)
	default:
		return ""
	}
}

func validateExportRoles(actual map[string]struct{}, roles map[string]string) error {
	unclassified := []string{}
	stale := []string{}
	empty := []string{}
	for key := range actual {
		role, present := roles[key]
		if !present {
			unclassified = append(unclassified, key)
		} else if strings.TrimSpace(role) == "" {
			empty = append(empty, key)
		}
	}
	for key := range roles {
		if _, present := actual[key]; !present {
			stale = append(stale, key)
		}
	}
	sort.Strings(unclassified)
	sort.Strings(stale)
	sort.Strings(empty)
	if len(unclassified) != 0 || len(stale) != 0 || len(empty) != 0 {
		return fmt.Errorf("export role mismatch: unclassified=%v stale=%v empty=%v", unclassified, stale, empty)
	}
	return nil
}
