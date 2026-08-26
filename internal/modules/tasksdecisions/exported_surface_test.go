package tasksdecisions_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
)

var tasksDecisionsRootInventory = groupedInventory(map[string]string{
	"const": `
		DecisionsViewSchemaID StoredMutationCreate StoredMutationDecisionSupersession
		StoredMutationPatch TaskRequestsViewSchemaID
	`,
	"var": `ErrClientTxnConflict ErrIdempotencyNotFound ErrStoredMutationKindMismatch`,
	"type": `
		AdmissionFailure CollectionAction CollectionActionPayload ConflictClaims ConflictCommand
		ConflictResolveRequest CreateCommand CreateRequest FieldValue IdempotencyCapability
		IdempotencyKey IdempotencyRecord ImportDependencies ImportLinkCapability
		ImportRecordEnvelopeCapability ImportRevisionCapability IncidentStateCapability
		LifecycleValidationError LinkCapability MutationDependencies MutationFacade MutationResult
		OptionalConflictValue PatchChange PatchCommand PatchRequest RecordEnvelopeCapability
		RevisionCapability RowVersionConflictError SameFieldConflict SameFieldConflictError
		StoredDecisionSupersessionResult StoredMutationKind StoredMutationResult StoredRowMutationResult
		SupersedeCommand SupersedeFacts SupersedeMutationResult SupersedeRequest ValidationError
	`,
	"func": `
		AdmitConflictResolveJSON AdmitCreateJSON AdmitPatchJSON AdmitSupersedeJSON
		ConflictResolveRequestHash CreateRequestHash IncidentBundleSubtypeContribution
		NewImportContribution NewIncidentBundleSourcePort NewMutationContribution
		NewProjectionContribution NewReportingContribution NewRevisionContribution
		NewStoredCreateResult NewStoredDecisionSupersessionResult NewStoredPatchResult
		PatchRequestHash RecoveryStateContribution SupersedeRequestHash
	`,
	"method": `
		AdmissionFailure.CollectionFieldKey AdmissionFailure.CountLimit AdmissionFailure.Error
		AdmissionFailure.Field AdmissionFailure.ReasonCode MutationFacade.Create MutationFacade.Patch
		MutationFacade.ResolveConflict MutationFacade.SupersedeDecision RowVersionConflictError.Error
		SameFieldConflictError.Error StoredMutationResult.DecisionSupersessionResult
		StoredMutationResult.Kind StoredMutationResult.RowMutationResult importOwner.CreateImportRowTx
		reportingContribution.CollectFactsTx reportingContribution.ProviderKey
	`,
	"interface-method": `
		IdempotencyCapability.Get IdempotencyCapability.PutTx
		ImportLinkCapability.SyncFieldReferenceWithMutationValuesTx
		ImportRecordEnvelopeCapability.InsertTx ImportRevisionCapability.AppendNonRowMutationTx
		IncidentStateCapability.RequireOpenTx LinkCapability.ApplyRecordRefCollectionWithMutationValuesTx
		LinkCapability.InsertSupersedesCommandTx LinkCapability.SyncFieldReferenceWithMutationValuesTx
		LinkCapability.ValidateRecordRefCollectionTx RecordEnvelopeCapability.AdvanceVersionTx
		RecordEnvelopeCapability.InsertTx RecordEnvelopeCapability.LoadEnvelopeTx
		RevisionCapability.AppendChangeSetTx RevisionCapability.AppendLiveRevisionTx
		RevisionCapability.AppendMutationTx RevisionCapability.AppendRecordMutationTx
		RevisionCapability.CaptureRecordSnapshotTx RevisionCapability.LoadRevisionWindowTx
	`,
	"interface-embed": `ImportRevisionCapability.ownerfacade.LiveRecordRevisionAppender`,
})

var tasksDecisionsProjectionContractInventory = groupedInventory(map[string]string{
	"type": `
		Contribution DecisionProjectionInput DecisionProjectionInputPage DecisionSourceReader
		TaskRequestProjectionInput TaskRequestProjectionInputPage TaskRequestSourceReader
	`,
	"func": `NewContribution`,
	"method": `
		Contribution.DecisionSource Contribution.ProjectionContribution Contribution.TaskRequestSource
	`,
	"interface-method": `
		DecisionSourceReader.ListDecisionProjectionInputsTx
		DecisionSourceReader.LoadDecisionProjectionInputTx
		TaskRequestSourceReader.ListTaskRequestProjectionInputsTx
		TaskRequestSourceReader.LoadTaskRequestProjectionInputTx
	`,
})

var tasksDecisionsProjectionPortsInventory = groupedInventory(map[string]string{
	"type": `DecisionDerivedFact MutationRows ReportingReader TaskDerivedFact`,
	"interface-method": `
		MutationRows.LoadDecisionTx MutationRows.LoadTaskRequestTx MutationRows.RefreshDecisionTx
		MutationRows.RefreshTaskRequestTx ReportingReader.CollectDecisionDerivedFactsTx
		ReportingReader.CollectTaskDerivedFactsTx
	`,
})

func TestTasksDecisionsExportedSurfaceAllowlist_Unit(t *testing.T) {
	t.Parallel()

	root := tasksDecisionsSourceRoot(t)
	t.Run("root exports and methods are exact", func(t *testing.T) {
		_, locations := assertExactDirectoryInventory(t, root, tasksDecisionsRootInventory, true)
		assertDeclarationLocations(t, locations, map[string]string{
			"func:IncidentBundleSubtypeContribution": "incident_bundle_contribution.go",
			"func:NewIncidentBundleSourcePort":       "incident_bundle_contribution.go",
			"func:NewImportContribution":             "import_create.go",
			"func:NewProjectionContribution":         "projection_contribution.go",
			"func:NewReportingContribution":          "reporting_contribution.go",
			"func:RecoveryStateContribution":         "recovery_state.go",
			"method:MutationFacade.ResolveConflict":  "mutation_conflict_resolution.go",
			"type:RowVersionConflictError":           "mutation_contracts.go",
		})
	})
	t.Run("projection contract exports and methods are exact", func(t *testing.T) {
		_, locations := assertExactDirectoryInventory(t, filepath.Join(root, "projectioncontract"), tasksDecisionsProjectionContractInventory, true)
		for declaration := range tasksDecisionsProjectionContractInventory {
			if locations[declaration] != "contribution.go" {
				t.Fatalf("projection contract declaration %s is in %q, want contribution.go", declaration, locations[declaration])
			}
		}
	})
	t.Run("projection ports exports and methods are exact", func(t *testing.T) {
		_, locations := assertExactDirectoryInventory(t, filepath.Join(root, "projectionports"), tasksDecisionsProjectionPortsInventory, true)
		for declaration := range tasksDecisionsProjectionPortsInventory {
			if locations[declaration] != "ports.go" {
				t.Fatalf("projection port declaration %s is in %q, want ports.go", declaration, locations[declaration])
			}
		}
	})
	t.Run("unexpected qualified method is rejected", func(t *testing.T) {
		file, err := parser.ParseFile(token.NewFileSet(), "negative_fixture.go", `
package fixture
type Guarded struct{}
func (Guarded) UnexpectedMethod() {}
`, 0)
		if err != nil {
			t.Fatalf("parse unexpected-method fixture: %v", err)
		}
		got, _ := inventoryFiles(t, map[string]*ast.File{"negative_fixture.go": file}, true)
		unexpected := inventoryDifference(got, groupedInventory(map[string]string{"type": `Guarded`}))
		if !slices.Equal(unexpected, []string{"method:Guarded.UnexpectedMethod"}) {
			t.Fatalf("unexpected-method fixture reported %v", unexpected)
		}
	})
	t.Run("provider files remain cohesive", func(t *testing.T) {
		assertProviderDeclarationCohesion(t, root)
	})
	t.Run("retired topology remains absent", func(t *testing.T) {
		assertTasksDecisionsRetiredTopologyAbsent(t, root)
	})
}

func tasksDecisionsSourceRoot(t testing.TB) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate Tasks/Decisions package")
	}
	return filepath.Dir(currentFile)
}

func groupedInventory(groups map[string]string) map[string]struct{} {
	result := make(map[string]struct{})
	for kind, declarations := range groups {
		for _, declaration := range strings.Fields(declarations) {
			result[kind+":"+declaration] = struct{}{}
		}
	}
	return result
}

func assertExactDirectoryInventory(t testing.TB, directory string, want map[string]struct{}, exportedOnly bool) (map[string]struct{}, map[string]string) {
	t.Helper()
	entries, err := filepath.Glob(filepath.Join(directory, "*.go"))
	if err != nil {
		t.Fatalf("list %s sources: %v", directory, err)
	}
	parsed := make(map[string]*ast.File)
	for _, path := range entries {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		parsed[filepath.Base(path)] = file
	}
	got, locations := inventoryFiles(t, parsed, exportedOnly)
	unexpected := inventoryDifference(got, want)
	missing := inventoryDifference(want, got)
	if len(unexpected) != 0 || len(missing) != 0 {
		t.Fatalf("%s declaration inventory drifted: unexpected=%v missing=%v", directory, unexpected, missing)
	}
	return got, locations
}

func inventoryFiles(t testing.TB, files map[string]*ast.File, exportedOnly bool) (map[string]struct{}, map[string]string) {
	t.Helper()
	inventory := make(map[string]struct{})
	locations := make(map[string]string)
	add := func(key string, path string) {
		if previous, duplicate := locations[key]; duplicate {
			t.Fatalf("declaration %s appears in both %s and %s", key, previous, path)
		}
		inventory[key] = struct{}{}
		locations[key] = path
	}
	for path, file := range files {
		for _, declaration := range file.Decls {
			switch current := declaration.(type) {
			case *ast.GenDecl:
				for _, specification := range current.Specs {
					switch value := specification.(type) {
					case *ast.TypeSpec:
						if value.Name.Name == "_" || exportedOnly && !value.Name.IsExported() {
							continue
						}
						add("type:"+value.Name.Name, path)
						interfaceType, ok := value.Type.(*ast.InterfaceType)
						if !ok {
							continue
						}
						for _, member := range interfaceType.Methods.List {
							if len(member.Names) == 0 {
								add("interface-embed:"+value.Name.Name+"."+expressionName(member.Type), path)
								continue
							}
							for _, name := range member.Names {
								if name.Name != "_" && (!exportedOnly || name.IsExported()) {
									add("interface-method:"+value.Name.Name+"."+name.Name, path)
								}
							}
						}
					case *ast.ValueSpec:
						for _, name := range value.Names {
							if name.Name != "_" && (!exportedOnly || name.IsExported()) {
								add(current.Tok.String()+":"+name.Name, path)
							}
						}
					}
				}
			case *ast.FuncDecl:
				if current.Name.Name == "_" || exportedOnly && !current.Name.IsExported() {
					continue
				}
				if current.Recv == nil {
					add("func:"+current.Name.Name, path)
					continue
				}
				add("method:"+expressionName(current.Recv.List[0].Type)+"."+current.Name.Name, path)
			}
		}
	}
	return inventory, locations
}

func expressionName(expression ast.Expr) string {
	switch current := expression.(type) {
	case *ast.Ident:
		return current.Name
	case *ast.StarExpr:
		return expressionName(current.X)
	case *ast.SelectorExpr:
		return expressionName(current.X) + "." + current.Sel.Name
	case *ast.IndexExpr:
		return expressionName(current.X)
	case *ast.IndexListExpr:
		return expressionName(current.X)
	default:
		return "<unsupported>"
	}
}

func inventoryDifference(left map[string]struct{}, right map[string]struct{}) []string {
	result := make([]string, 0)
	for declaration := range left {
		if _, ok := right[declaration]; !ok {
			result = append(result, declaration)
		}
	}
	slices.Sort(result)
	return result
}

func assertDeclarationLocations(t testing.TB, got map[string]string, want map[string]string) {
	t.Helper()
	for declaration, path := range want {
		if got[declaration] != path {
			t.Fatalf("declaration %s is in %q, want %q", declaration, got[declaration], path)
		}
	}
}

func assertProviderDeclarationCohesion(t testing.TB, root string) {
	t.Helper()
	cases := []struct {
		path string
		want map[string]struct{}
	}{
		{filepath.Join(root, "internal", "providers", "incidentbundle", "export.go"), groupedInventory(map[string]string{
			"const": `decisionsBundlePath taskRequestsBundlePath`, "func": `exportIncidentBundleFiles`,
		})},
		{filepath.Join(root, "internal", "providers", "incidentbundle", "portable_values.go"), groupedInventory(map[string]string{
			"type": `portableDecision portableTaskRequest`,
			"func": `admittedPortableActor canonicalPortableTime canonicalPortableUUID exactPortableMembers nullableAdmittedPortableActor nullableCanonicalPortableTime nullableCanonicalPortableUUID nullablePortableString portableString`,
		})},
		{filepath.Join(root, "internal", "providers", "incidentbundle", "import_prepare.go"), groupedInventory(map[string]string{
			"type": `preparedTasksDecisionsImport`,
			"func": `preparePortableDecision preparePortableTaskRequest prepareTasksDecisionsImport tasksDecisionsInvariantFailure`,
		})},
		{filepath.Join(root, "internal", "providers", "incidentbundle", "import_apply.go"), groupedInventory(map[string]string{
			"func": `applyPreparedTasksDecisionsImportTx`,
		})},
		{filepath.Join(root, "internal", "providers", "rollback", "task_provider.go"), groupedInventory(map[string]string{
			"type":   `TaskRequestProvider taskLifecycle`,
			"func":   `NewTaskRequestProvider taskSourceForRollbackValue validateTaskReferencesTx validTaskLifecycle validTaskSource`,
			"method": `TaskRequestProvider.RestoreTx TaskRequestProvider.ValidateRollbackValue`,
		})},
		{filepath.Join(root, "internal", "providers", "rollback", "decision_provider.go"), groupedInventory(map[string]string{
			"type":   `DecisionProvider decisionMachine`,
			"func":   `NewDecisionProvider decisionSourceForRollbackValue validDecisionMachine validDecisionSource`,
			"method": `DecisionProvider.RestoreTx DecisionProvider.ValidateRollbackValue`,
		})},
		{filepath.Join(root, "internal", "providers", "rollback", "value_decode.go"), groupedInventory(map[string]string{
			"const": `fieldText fieldTime fieldUUID`, "type": `fieldKind fieldSpec`,
			"func": `nonEmptyText nullableSQLString nullableSQLTime nullableSQLUUID nullableText nullableTime nullableUUID objectMap policyText sourceForRollbackValue typedValues validTypedFields`,
		})},
	}
	for _, current := range cases {
		file, err := parser.ParseFile(token.NewFileSet(), current.path, nil, 0)
		if err != nil {
			t.Fatalf("parse cohesive provider file %s: %v", current.path, err)
		}
		got, _ := inventoryFiles(t, map[string]*ast.File{filepath.Base(current.path): file}, false)
		unexpected := inventoryDifference(got, current.want)
		missing := inventoryDifference(current.want, got)
		if len(unexpected) != 0 || len(missing) != 0 {
			t.Fatalf("provider file %s lost declaration cohesion: unexpected=%v missing=%v", current.path, unexpected, missing)
		}
	}
}

func assertTasksDecisionsRetiredTopologyAbsent(t testing.TB, root string) {
	t.Helper()
	for _, path := range []string{
		filepath.Join(root, "workbookprojection"),
		filepath.Join(root, "workbook_conflict.go"),
		filepath.Join(root, "decision_mutation_store_test.go"),
		filepath.Join(root, "internal", "providers", "incidentbundle", "portability.go"),
		filepath.Join(root, "internal", "providers", "rollback", "provider.go"),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("retired Tasks/Decisions path %s remains: %v", path, err)
		}
	}

	assertOnlyTestEntrypointsExported(t, root)
	repoRoot := filepath.Clean(filepath.Join(root, "..", "..", ".."))
	retiredRootSymbols := map[string]struct{}{
		"Descriptors": {}, "ImportCreateCommand": {}, "IncidentBundleContribution": {},
		"MemberReferenceCapability": {}, "NewIncidentBundleContribution": {},
		"NewMemberReferenceCapability": {}, "NewRecoveryContribution": {}, "Ports": {},
		"Reader": {}, "Rebuilder": {}, "ReportingContribution": {}, "Rows": {},
		"SupersedeRowVersionConflictError": {}, "SurfaceIntents": {}, "TaskDecisionPorts": {},
		"TaskReader": {},
	}
	retiredCoordinationSymbols := map[string]struct{}{
		"RebuildDecisions": {}, "RebuildDecisionsTx": {}, "RebuildTaskRequests": {},
		"RebuildTaskRequestsTx": {}, "TaskDecisionPorts": {},
	}
	oldProjectionImport := "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
	rootImport := "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	internalRoot := filepath.Join(repoRoot, "internal")
	err := filepath.WalkDir(internalRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		rootAlias := ""
		for _, importSpec := range file.Imports {
			importPath, err := strconv.Unquote(importSpec.Path.Value)
			if err != nil {
				return err
			}
			if importPath == oldProjectionImport {
				return &retiredTopologyError{path: path, detail: "imports retired Tasks/Decisions workbookprojection package"}
			}
			if importPath == rootImport {
				rootAlias = "tasksdecisions"
				if importSpec.Name != nil {
					rootAlias = importSpec.Name.Name
				}
			}
		}
		var violation string
		ast.Inspect(file, func(node ast.Node) bool {
			if violation != "" {
				return false
			}
			if identifier, ok := node.(*ast.Ident); ok {
				if _, retired := retiredCoordinationSymbols[identifier.Name]; retired {
					violation = "reaches retired task-specific projection coordination symbol " + identifier.Name
					return false
				}
			}
			selector, ok := node.(*ast.SelectorExpr)
			if !ok || rootAlias == "" {
				return true
			}
			qualifier, ok := selector.X.(*ast.Ident)
			if !ok || qualifier.Name != rootAlias {
				return true
			}
			if _, retired := retiredRootSymbols[selector.Sel.Name]; retired {
				violation = "reaches retired tasksdecisions." + selector.Sel.Name
				return false
			}
			return true
		})
		if violation != "" {
			return &retiredTopologyError{path: path, detail: violation}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("audit retired Tasks/Decisions topology: %v", err)
	}
}

func assertOnlyTestEntrypointsExported(t testing.TB, root string) {
	t.Helper()
	entries, err := filepath.Glob(filepath.Join(root, "*_test.go"))
	if err != nil {
		t.Fatalf("list Tasks/Decisions external tests: %v", err)
	}
	for _, path := range entries {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		if file.Name.Name != "tasksdecisions_test" {
			continue
		}
		for _, declaration := range file.Decls {
			switch current := declaration.(type) {
			case *ast.FuncDecl:
				if current.Recv == nil && current.Name.IsExported() && !strings.HasPrefix(current.Name.Name, "Test") {
					t.Fatalf("external test file %s exposes helper function %s", path, current.Name.Name)
				}
			case *ast.GenDecl:
				for _, specification := range current.Specs {
					switch value := specification.(type) {
					case *ast.TypeSpec:
						if value.Name.IsExported() {
							t.Fatalf("external test file %s exposes helper type %s", path, value.Name.Name)
						}
					case *ast.ValueSpec:
						for _, name := range value.Names {
							if name.IsExported() {
								t.Fatalf("external test file %s exposes helper value %s", path, name.Name)
							}
						}
					}
				}
			}
		}
	}
}

type retiredTopologyError struct {
	path   string
	detail string
}

func (failure *retiredTopologyError) Error() string {
	return failure.path + ": " + failure.detail
}
