package entities

import (
	"encoding/json"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const entitiesRepoImportPrefix = "github.com/JochiRaider/cartulary/"

func TestEntitiesProductionImportBoundaries(t *testing.T) {
	allowedSiblingImports := map[string]map[string]bool{
		entitiesRepoImportPrefix + "internal/modules/entities/hostidentity/deleterestore": {
			"revision_provider_contribution.go": true,
		},
		entitiesRepoImportPrefix + "internal/modules/entities/hostidentity/rollbackprovider": {
			"revision_provider_contribution.go": true,
		},
		entitiesRepoImportPrefix + "internal/modules/entities/merge": {
			"http_helpers.go": true,
			"routes.go":       true,
		},
		entitiesRepoImportPrefix + "internal/modules/entities/mentions": {
			"routes.go": true,
		},
		entitiesRepoImportPrefix + "internal/modules/entities/mentions/rollbackprovider": {
			"revision_provider_contribution.go": true,
		},
		entitiesRepoImportPrefix + "internal/modules/incidentbundles/sourceport": {
			"incident_bundle_portability.go":       true,
			"incident_bundle_portable_apply.go":    true,
			"incident_bundle_portable_encode.go":   true,
			"incident_bundle_portable_model.go":    true,
			"incident_bundle_portable_prepare.go":  true,
			"incident_bundle_portable_validate.go": true,
			"incident_bundle_source_port.go":       true,
		},
		entitiesRepoImportPrefix + "internal/modules/incidentportability": {
			"incident_bundle_portability.go":      true,
			"incident_bundle_portable_encode.go":  true,
			"incident_bundle_portable_prepare.go": true,
		},
		entitiesRepoImportPrefix + "internal/modules/incidents/admission": {
			"routes.go": true,
		},
		entitiesRepoImportPrefix + "internal/modules/records/subtypepresence": {
			"incident_bundle_subtype_presence.go": true,
		},
		entitiesRepoImportPrefix + "internal/modules/revisions": {
			"revision_provider_contribution.go": true,
		},
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read entities package directory: %v", err)
	}
	for _, entry := range entries {
		fileName := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
			continue
		}
		for _, importPath := range entitiesProductionImports(t, fileName) {
			if !strings.HasPrefix(importPath, entitiesRepoImportPrefix+"internal/modules/") {
				continue
			}
			allowedFiles, ok := allowedSiblingImports[importPath]
			if !ok {
				t.Fatalf("%s imports unapproved sibling module %s", fileName, importPath)
			}
			if !allowedFiles[fileName] {
				t.Fatalf("%s imports %s; allowed files are %v", fileName, importPath, entitiesAllowedFileNames(allowedFiles))
			}
		}
	}

	t.Run("active tests have one exact owner selector", func(t *testing.T) {
		entitiesReconcileExactTestSelectors(t)
	})

	t.Run("production exports have exact dispositions", func(t *testing.T) {
		entitiesReconcileProductionExports(t)
	})

	t.Run("merge uses consumer-owned cross-owner ports", func(t *testing.T) {
		entries, err := filepath.Glob(filepath.Join("merge", "*.go"))
		if err != nil {
			t.Fatalf("list merge package: %v", err)
		}
		for _, path := range entries {
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			for _, importPath := range entitiesProductionImports(t, path) {
				if importPath == entitiesRepoImportPrefix+"internal/modules/assessments" {
					t.Fatalf("%s imports concrete Assessments instead of the merge-owned port", path)
				}
			}
		}
	})
}

type entitiesExportDisposition string

const (
	entitiesExportRetain entitiesExportDisposition = "retain"
)

// entitiesExportDispositions is the final reviewed production boundary. Every
// entry is retained and inherits the package's independently reviewed role.
var entitiesExportDispositions = map[string]map[string]entitiesExportDisposition{
	".": entitiesExportInventory(`
		IncidentBundleSubtypeContribution NewIncidentBundleSourcePort RecoveryStateContribution
		RegisterRoutes RevisionProviderContribution RouteOptions
	`),
	"entitycontract": entitiesExportInventory(`
		EntityTypeHost EntityTypeIdentity HostsViewSchemaID IdentitiesViewSchemaID
	`),
	"hostidentity": entitiesExportInventory(`
		AliasAppliedMutation AliasMutationValue AliasMutationValue.MutationValue AliasSyncResult AliasSyncResult.Changed AliasValue
		BuildClipboardPastePlan
		ClipboardPasteRequest ClipboardPasteRequest.RequestHash ClipboardPasteResult ClipboardPasteRowResult CollectionAction
		ConflictCommand CreateRequest CreateRequestHash DecodeClipboardPasteRequest DecodeCreateRequest DecodePatchRequest
		DecodeWorkbookConflictResolveRequest EligibleAlias
		ErrHostIdentityRecordNotFound ErrInvalidAliasReference ErrInvalidCreateRequest ErrNoEffectivePatchChange
		ActiveIdentifierTransitionConflict ExactMatchConflictError ExactMatchConflictError.EntityMatchConflictDetails ExactMatchConflictError.Error
		HostRecord HostsViewSchemaID IdentitiesViewSchemaID IdentityRecord ImportDependencies MergeCapability
		MergeCapability.HostCanonicalNormalized MergeCapability.HostExactMatchPrecedence MergeCapability.HostRevisionConflictFacts MergeCapability.IdentityCanonicalNormalized
		MergeCapability.IdentityExactMatchPrecedence MergeCapability.IdentityRevisionConflictFacts MergeCapability.LoadHostTx MergeCapability.LoadIdentityTx MergeCapability.PrepareIdentifierClaimsTx MergeCapability.SyncAliasesTx
		MergeCapability.SyncPreservedIdentifierTx MergeCapability.UpdateHostTx MergeCapability.UpdateIdentityTx MutationResult
		NewImportCreateFacade NewMergeCapability NewSourceFacts NewStore PatchChange PatchMutationResult PatchRequest PatchRequestHash PrepareActiveIdentifierStateTransitionTx
		ReusableIdentifier RowVersionConflictError RowVersionConflictError.Error Store Store.ApplyClipboardPastePlan Store.CreateHostRow
		Store.CreateIdentityRow Store.PatchEntityRow
		Store.QueryHostRowsPage Store.QueryIdentityRowsPage
		Store.ResolveWorkbookConflict StoreDependencies SourceFacts SourceFacts.ListEligibleAliasesTx SourceFacts.ValidateResolvedTargetTx
		WorkbookConflictClaims WorkbookConflictResolveRequest WorkbookConflictResolveRequestHash
	`),
	"hostidentity/deleterestore": entitiesExportInventory(`
		HostSource HostSource.PrepareStateTransitionTx HostSource.SnapshotTx HostSource.SyncEnvelopeMirrorTx HostSource.UpdateSourceDeleteStateTx HostSource.ViewSchemaID
		IdentitySource IdentitySource.PrepareStateTransitionTx IdentitySource.SnapshotTx IdentitySource.SyncEnvelopeMirrorTx IdentitySource.UpdateSourceDeleteStateTx IdentitySource.ViewSchemaID
		NewHostSource NewIdentitySource
	`),
	"hostidentity/projectionprovider": entitiesExportInventory(`
		NewSource Source Source.ListHostProjectionInputsTx Source.ListIdentityProjectionInputsTx Source.LoadHostProjectionInputTx Source.LoadIdentityProjectionInputTx
	`),
	"hostidentity/reportingprovider": entitiesExportInventory(`
		New Provider Provider.CollectFactsTx Provider.CollectFieldsTx Provider.ProviderKey
	`),
	"hostidentity/rollbackprovider": entitiesExportInventory(`
		CollectionProvider CollectionProvider.ApplyInverseTx CollectionProvider.DescribeTx CollectionProvider.IdentifierClaimRecordTx
		HostProvider HostProvider.FinalizeIdentifierClaimRestoreTx HostProvider.PrepareIdentifierClaimRestoreTx HostProvider.RestoreTx HostProvider.ValidateRollbackValue
		IdentityProvider IdentityProvider.FinalizeIdentifierClaimRestoreTx IdentityProvider.PrepareIdentifierClaimRestoreTx IdentityProvider.RestoreTx IdentityProvider.ValidateRollbackValue
		NewCollectionProvider NewHostProvider NewIdentityProvider
	`),
	"mentions": entitiesExportInventory(`
		CreateParams DecodeMentionActionRequest ErrEntityMentionNotFound ErrInvalidMentionResolution
		ErrRecordDeletedUseRestore ErrResolvedRecordNotFound ErrSourceRecordNotFound
		LinkCommand LinkCommandResult LinkMutation LinkOperationsPort LinkType LinkTypeObservedAsIdentity LinkTypeObservedOnHost
		MentionActionAccess MentionActionRequest MentionActionRequestHash MentionActionResult MentionEntityInvalidation MentionResolutionResult
		MentionRowVersionConflictError MentionRowVersionConflictError.Error MentionTargetValidationError MentionTargetValidationError.Error
		MentionTargetValidationError.InvalidMutationTarget MentionTransitionError MentionTransitionError.Error
		MentionTransitionError.MutationTransitionDetails MergeMutation NewStore
		RepointMergedMentionsCommand RepointMergedMentionsResult
		Store Store.ApplyMentionAction Store.ApplyMentionLifecycleTx Store.GetMentionActionAccess Store.InsertTx
		Store.LoadTimelineCollectionFieldsChangedTx Store.NextOrdinalTx Store.RepointMergedMentionsTx Store.ResolveExistingFromMentionTx
		StoreDependencies TimelineEffectsPort TombstoneLinkCommand
	`),
	"mentions/reportingprovider": entitiesExportInventory(`CollectFactsTx CollectFieldsTx`),
	"mentions/rollbackprovider": entitiesExportInventory(`
		MentionProvider MentionProvider.ApplyInverseTx MentionProvider.DescribeTx NewMentionProvider
	`),
	"merge": entitiesExportInventory(`
		AssessmentEffectsPort AssessmentMutation AssessmentProtectedSetChangedError AssessmentProtectedSetChangedError.Error
		AssessmentProtectedSetCommand AssessmentRepointCommand AssessmentRepointResult DecodeMergeRequest
		ErrMergeTargetNotFound LinkEffectMutation LinkEffectsPort MentionEffectsPort MergeExactMatchClassSummary MergePreconditionError MergePreconditionError.Error
		MergeRecordLockedError MergeRecordLockedError.Error MergeRequest MergeRequestHash MergeResult MergeRowVersionConflictError
		MergeRowVersionConflictError.Error MergeSummary NewStore RepointLinksCommand
		RepointLinksResult RepointTagsCommand RepointTagsResult Store Store.GetMergeRouteIncident Store.MergeEntity StoreDependencies
		TimelineEffectsPort
	`),
	"timelinefacts": entitiesExportInventory(`Reader Reader.LoadMentionsTx`),
	"workbookprojection": entitiesExportInventory(`
		Contribution Contribution.ProjectionContribution Contribution.Source DerivedFact Descriptors HostProjectionInput HostProjectionPage
		HostQueryProjection HostSurfaceIntent IdentityProjectionInput IdentityProjectionPage IdentityQueryProjection IdentitySurfaceIntent
		NewContribution Ports Reader Rebuilder SourceReader Writer
	`),
}

var entitiesExportRoles = map[string]string{
	".":                               "bounded HTTP facade and source-owner contribution assembly",
	"entitycontract":                  "shared Entities-owned semantic identifiers",
	"hostidentity":                    "Workbook, import, source-fact, and merge application capabilities",
	"hostidentity/deleterestore":      "Recovery source contribution",
	"hostidentity/projectionprovider": "projection source contribution",
	"hostidentity/reportingprovider":  "Reporting source contribution",
	"hostidentity/rollbackprovider":   "Revisions rollback source contribution",
	"mentions":                        "mention application operations and injected effect language",
	"mentions/reportingprovider":      "Reporting source contribution",
	"mentions/rollbackprovider":       "Revisions rollback source contribution",
	"merge":                           "merge application operation and consumer-owned effects",
	"timelinefacts":                   "Entities-owned Timeline source facts",
	"workbookprojection":              "Entities and Projections typed projection contract",
}

func entitiesExportInventory(entries string) map[string]entitiesExportDisposition {
	inventory := make(map[string]entitiesExportDisposition)
	for _, entry := range strings.Fields(entries) {
		name, rawDisposition, hasDisposition := strings.Cut(entry, "=")
		disposition := entitiesExportRetain
		if hasDisposition {
			disposition = entitiesExportDisposition(rawDisposition)
		}
		inventory[name] = disposition
	}
	return inventory
}

func entitiesReconcileProductionExports(t *testing.T) {
	t.Helper()

	var mismatches []string
	actualPackages := entitiesProductionPackageDirectories(t)
	approvedPackages := make([]string, 0, len(entitiesExportDispositions))
	for relative := range entitiesExportDispositions {
		approvedPackages = append(approvedPackages, relative)
	}
	unapprovedPackages, missingPackages := entitiesExportSurfaceDelta(actualPackages, approvedPackages)
	if len(unapprovedPackages) != 0 || len(missingPackages) != 0 {
		mismatches = append(mismatches, "package inventory: unapproved="+strings.Join(unapprovedPackages, ",")+" missing="+strings.Join(missingPackages, ","))
	}
	for relative, expected := range entitiesExportDispositions {
		if strings.TrimSpace(entitiesExportRoles[relative]) == "" {
			mismatches = append(mismatches, relative+": missing production-role justification")
		}
		actual := entitiesExportedPackageDeclarations(t, relative)
		approved := make([]string, 0, len(expected))
		for name, disposition := range expected {
			if disposition != entitiesExportRetain {
				mismatches = append(mismatches, relative+": final inventory contains non-retained export "+name+": "+string(disposition))
				continue
			}
			approved = append(approved, name)
		}
		unapproved, missing := entitiesExportSurfaceDelta(actual, approved)
		if len(unapproved) != 0 || len(missing) != 0 {
			mismatches = append(mismatches, relative+": unapproved="+strings.Join(unapproved, ",")+" missing="+strings.Join(missing, ","))
		}
	}
	if len(mismatches) != 0 {
		sort.Strings(mismatches)
		t.Fatalf("Entities production export disposition mismatch:\n%s", strings.Join(mismatches, "\n"))
	}

	t.Run("negative fixture detects an unapproved export", func(t *testing.T) {
		file, err := parser.ParseFile(token.NewFileSet(), "negative_fixture.go", "package negativefixture\nfunc UnexpectedExport() {}\n", 0)
		if err != nil {
			t.Fatalf("parse negative fixture: %v", err)
		}
		unexpected, missing := entitiesExportSurfaceDelta(entitiesExportedDeclarations(file), nil)
		if !reflect.DeepEqual(unexpected, []string{"UnexpectedExport"}) || len(missing) != 0 {
			t.Fatalf("negative fixture was not rejected: unexpected=%v missing=%v", unexpected, missing)
		}
	})
}

func entitiesProductionPackageDirectories(t testing.TB) []string {
	t.Helper()
	packages := map[string]struct{}{}
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != "." && entry.Name() == "testsupport" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}
		directory := filepath.ToSlash(filepath.Dir(path))
		packages[directory] = struct{}{}
		return nil
	})
	if err != nil {
		t.Fatalf("discover Entities production packages: %v", err)
	}
	result := make([]string, 0, len(packages))
	for directory := range packages {
		result = append(result, directory)
	}
	sort.Strings(result)
	return result
}

func entitiesExportedPackageDeclarations(t testing.TB, directory string) []string {
	t.Helper()
	entries, err := filepath.Glob(filepath.Join(directory, "*.go"))
	if err != nil {
		t.Fatalf("list package directory %s: %v", directory, err)
	}
	fileSet := token.NewFileSet()
	var declarations []string
	for _, path := range entries {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		declarations = append(declarations, entitiesExportedDeclarations(file)...)
	}
	return declarations
}

func entitiesExportedDeclarations(file *ast.File) []string {
	var names []string
	for _, declaration := range file.Decls {
		switch typed := declaration.(type) {
		case *ast.FuncDecl:
			if !ast.IsExported(typed.Name.Name) {
				continue
			}
			name := typed.Name.Name
			if typed.Recv != nil {
				receiver := entitiesExportedReceiverName(typed.Recv.List[0].Type)
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

func entitiesExportedReceiverName(expression ast.Expr) string {
	switch typed := expression.(type) {
	case *ast.Ident:
		if ast.IsExported(typed.Name) {
			return typed.Name
		}
	case *ast.StarExpr:
		return entitiesExportedReceiverName(typed.X)
	}
	return ""
}

func entitiesExportSurfaceDelta(actual []string, expected []string) ([]string, []string) {
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

func TestEntitiesDoNotRegisterWorkbookRowCreateRoutes(t *testing.T) {
	body, err := os.ReadFile(filepath.Clean("routes.go"))
	if err != nil {
		t.Fatalf("read routes.go: %v", err)
	}
	content := string(body)
	for _, route := range []string{
		"views/cartulary.view.hosts.v1/rows",
		"views/cartulary.view.identities.v1/rows",
		"views/cartulary.view.indicators.v1/rows",
	} {
		if strings.Contains(content, route) {
			t.Fatalf("entities routes.go still registers workbook row-create route %s", route)
		}
	}
}

func TestEntitiesDoNotBuildClipboardPastePlans(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read entities package directory: %v", err)
	}
	for _, entry := range entries {
		fileName := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
			continue
		}
		body, err := os.ReadFile(filepath.Clean(fileName))
		if err != nil {
			t.Fatalf("read %s: %v", fileName, err)
		}
		if strings.Contains(string(body), "BuildBatchPlan(") || strings.Contains(string(body), "BuildTabularRowPlanV1(") {
			t.Fatalf("%s builds clipboard paste plans inside entities", fileName)
		}
	}
}

func TestEntitiesRoutesUseCollaborationPublisher(t *testing.T) {
	imports := entitiesProductionImports(t, "routes.go")
	for _, importPath := range imports {
		if importPath == entitiesRepoImportPrefix+"internal/platform/ws" {
			t.Fatalf("routes.go imports platform/ws directly instead of collaboration publisher")
		}
	}
}

func TestEntitiesRootDoesNotImportHostIdentity(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read entities package directory: %v", err)
	}
	for _, entry := range entries {
		fileName := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
			continue
		}
		for _, importPath := range entitiesProductionImports(t, fileName) {
			if importPath == entitiesRepoImportPrefix+"internal/modules/entities/hostidentity" {
				t.Fatalf("%s imports the hostidentity application surface; root imports must remain contribution-specific", fileName)
			}
		}
	}
}

func TestMergeDoesNotWriteMentionOrProjectionTablesDirectly(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("merge", "*.go"))
	if err != nil {
		t.Fatalf("list merge production files: %v", err)
	}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		body, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		content := string(body)
		for _, disallowed := range []string{
			"UPDATE entity_mentions",
			"INSERT INTO entity_mentions",
			"DELETE FROM entity_mentions",
			"DELETE FROM host_grid_projection",
			"DELETE FROM identity_grid_projection",
		} {
			if strings.Contains(content, disallowed) {
				t.Fatalf("%s contains direct cross-owner write %q", path, disallowed)
			}
		}
	}

	t.Run("carry-forward planning is read-only", func(t *testing.T) {
		path := filepath.Clean(filepath.Join("merge", "source_carry_forward.go"))
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, declaration := range parsed.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil || !strings.HasPrefix(function.Name.Name, "plan") {
				continue
			}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				for _, mutationPrefix := range []string{"Sync", "Update", "Insert", "Delete", "Exec"} {
					if strings.HasPrefix(selector.Sel.Name, mutationPrefix) {
						t.Fatalf("%s calls mutation-capable method %s", function.Name.Name, selector.Sel.Name)
					}
				}
				return true
			})
		}
	})
}

func TestMentionsUseCommandLevelTimelineEffectsPort(t *testing.T) {
	body, err := os.ReadFile(filepath.Clean(filepath.Join("mentions", "ports.go")))
	if err != nil {
		t.Fatalf("read mentions/ports.go: %v", err)
	}
	content := string(body)
	if !strings.Contains(content, "type TimelineEffectsPort interface") {
		t.Fatalf("mentions ports must expose command-level timelineEffectsPort")
	}
	for _, disallowed := range []string{
		"type timelinePort interface",
		"LoadSourceRecordTx(context.Context, pgx.Tx, uuid.UUID)",
		"UpdateSourceRecordTx(context.Context, pgx.Tx",
		"BuildRecordRowTx(context.Context, pgx.Tx, uuid.UUID)",
		"RebuildTimelineProjectionTx(context.Context, pgx.Tx, uuid.UUID)",
		"VersionID(uuid.UUID, int64)",
	} {
		if strings.Contains(content, disallowed) {
			t.Fatalf("mentions timeline effects port still exposes timeline mechanics %q", disallowed)
		}
	}
}

func entitiesProductionImports(t testing.TB, fileName string) []string {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Clean(fileName), nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse imports for %s: %v", fileName, err)
	}
	imports := make([]string, 0, len(parsed.Imports))
	for _, spec := range parsed.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			t.Fatalf("unquote import for %s: %v", fileName, err)
		}
		imports = append(imports, importPath)
	}
	return imports
}

func entitiesAllowedFileNames(files map[string]bool) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

type entitiesTestFamilyManifest struct {
	Rows []struct {
		Runner   string `json:"runner"`
		Selector struct {
			Package string   `json:"package"`
			Tests   []string `json:"tests"`
		} `json:"selector"`
	} `json:"rows"`
}

func entitiesReconcileExactTestSelectors(t testing.TB) {
	t.Helper()

	discovered := map[string]bool{}
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}
		active, err := build.Default.MatchFile(filepath.Dir(path), entry.Name())
		if err != nil {
			return err
		}
		if !active {
			return nil
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		packagePath := "./internal/modules/entities"
		if dir := filepath.ToSlash(filepath.Dir(path)); dir != "." {
			packagePath += "/" + dir
		}
		for _, declaration := range parsed.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Recv != nil || !entitiesIsTopLevelTestName(function.Name.Name) {
				continue
			}
			discovered[packagePath+"\x00"+function.Name.Name] = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("discover active entities tests: %v", err)
	}

	manifestPath := filepath.Join("..", "..", "..", "tools", "test_families", "module.entities.json")
	body, err := os.ReadFile(filepath.Clean(manifestPath))
	if err != nil {
		t.Fatalf("read entities test-family manifest: %v", err)
	}
	var manifest entitiesTestFamilyManifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		t.Fatalf("decode entities test-family manifest: %v", err)
	}

	selected := map[string]int{}
	for _, row := range manifest.Rows {
		if row.Runner != "go" || (row.Selector.Package != "./internal/modules/entities" &&
			!strings.HasPrefix(row.Selector.Package, "./internal/modules/entities/")) {
			continue
		}
		for _, testName := range row.Selector.Tests {
			selected[row.Selector.Package+"\x00"+testName]++
		}
	}

	var anomalies []string
	for key := range discovered {
		switch selected[key] {
		case 0:
			anomalies = append(anomalies, "missing exact selector: "+entitiesDisplayTestKey(key))
		case 1:
		default:
			anomalies = append(anomalies, "duplicate exact selector: "+entitiesDisplayTestKey(key))
		}
	}
	for key, count := range selected {
		if !discovered[key] {
			anomalies = append(anomalies, "stale or inactive selector: "+entitiesDisplayTestKey(key))
		} else if count > 1 {
			anomalies = append(anomalies, "selector appears more than once: "+entitiesDisplayTestKey(key))
		}
	}
	sort.Strings(anomalies)
	if len(anomalies) > 0 {
		t.Fatalf("entities exact-selector reconciliation failed (discovered=%d selected=%d):\n%s",
			len(discovered), len(selected), strings.Join(anomalies, "\n"))
	}
	if len(discovered) != len(selected) {
		t.Fatalf("entities exact-selector count mismatch: discovered=%d selected=%d", len(discovered), len(selected))
	}
}

func entitiesIsTopLevelTestName(name string) bool {
	if !strings.HasPrefix(name, "Test") || len(name) == len("Test") {
		return false
	}
	next := name[len("Test")]
	return next < 'a' || next > 'z'
}

func entitiesDisplayTestKey(key string) string {
	return strings.Replace(key, "\x00", "::", 1)
}
