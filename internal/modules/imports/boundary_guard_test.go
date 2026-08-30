package imports

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportsTargetsUseOnlyGeneratedRegistry(t *testing.T) {
	body, err := os.ReadFile(filepath.Clean("targets.go"))
	if err != nil {
		t.Fatalf("read targets.go: %v", err)
	}
	content := string(body)
	if !strings.Contains(
		content,
		"internal/gen/importtargetregistry",
	) || !strings.Contains(content, "importtargetregistry.Targets") {
		t.Fatal("targets.go must project the generated import target registry")
	}
	for _, forbidden := range []string{
		"internal/modules/artifacts",
		"internal/modules/assessments",
		"internal/modules/entities",
		"internal/modules/evidence",
		"internal/modules/indicators",
		"internal/modules/parties",
		"internal/modules/tasksdecisions",
		"internal/modules/timeline",
		`"cartulary.view.hosts.v1"`,
		`"cartulary.view.indicators.v1"`,
		`"network_flow_import_facade_v1"`,
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("targets.go contains duplicated owner fact %q", forbidden)
		}
	}
}

func TestImportsOwnerApplyUsesInjectedFacadeRegistry(t *testing.T) {
	body, err := os.ReadFile(filepath.Clean("owner_apply.go"))
	if err != nil {
		t.Fatalf("read owner_apply.go: %v", err)
	}
	content := string(body)
	if !strings.Contains(content, "s.ownerCreateRegistry.Resolve(") ||
		!strings.Contains(content, "owner.CreateImportRowTx(") {
		t.Fatal("owner_apply.go must resolve and invoke the injected owner facade")
	}
	for _, forbidden := range []string{
		"internal/modules/artifacts",
		"internal/modules/assessments",
		"internal/modules/entities",
		"internal/modules/evidence",
		"internal/modules/indicators",
		"internal/modules/parties",
		"internal/modules/tasksdecisions",
		"artifacts.NewStore(",
		"assessments.NewStore(",
		"hostidentity.NewStore(",
		"evidence.NewStore(",
		"indicators.NewApplication(",
		"parties.NewStore(",
		"tasksdecisions.NewStore(",
		"FROM users",
		"FROM records",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("owner_apply.go retains cross-owner dependency %q", forbidden)
		}
	}
}

func TestImportsAndClipboardUseTheSharedMappingKernel(t *testing.T) {
	files := []struct {
		path     string
		required []string
	}{
		{
			path:     "owner_apply.go",
			required: []string{"tabularingest.BuildMappingKernelPlanV1(ctx", "importMappingKernelRequest("},
		},
		{
			path:     filepath.Join("..", "tabularingest", "tabularingest.go"),
			required: []string{"BuildMappingKernelPlanV1(context.Background()", "clipboardMappingKernelRequest("},
		},
	}
	for _, file := range files {
		body, err := os.ReadFile(filepath.Clean(file.path))
		if err != nil {
			t.Fatalf("read %s: %v", file.path, err)
		}
		for _, required := range file.required {
			if !strings.Contains(string(body), required) {
				t.Fatalf("%s does not invoke the shared mapping kernel through %q", file.path, required)
			}
		}
	}

	for _, path := range []string{"apply_coordination.go", "owner_apply.go"} {
		body, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, duplicate := range []string{"func transformImportValue(", "strings.Split(result, delimiter)", "strings.Join(strings.Fields(result), \" \")"} {
			if strings.Contains(string(body), duplicate) {
				t.Fatalf("%s retains duplicate mapping behavior %q", path, duplicate)
			}
		}
	}
}

func TestAnalyticalMappingRemainsTargetOwned(t *testing.T) {
	extensionBody, err := os.ReadFile(filepath.Clean("extension_facade.go"))
	if err != nil {
		t.Fatalf("read extension_facade.go: %v", err)
	}
	if strings.Contains(string(extensionBody), "tabularingest") || strings.Contains(string(extensionBody), "BuildMappingKernelPlanV1") {
		t.Fatal("analytical extension mapping must bypass the view-target mapping kernel")
	}
	coordinationBody, err := os.ReadFile(filepath.Clean("apply_coordination.go"))
	if err != nil {
		t.Fatalf("read apply_coordination.go: %v", err)
	}
	content := string(coordinationBody)
	for _, required := range []string{
		"unit.ApprovedMapping.targetKindOrDefault() != ImportTargetKindViewSchema",
		"s.applyExtensionOwnerUnitTx(",
		"s.applyGenericOwnerUnitTx(",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("apply coordination no longer proves target-owned analytical dispatch through %q", required)
		}
	}
}

func TestImportsProductionPackageHasNoConcretePeerStoresOrPeerTableSQL(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob imports package files: %v", err)
	}
	forbiddenImports := []string{
		"internal/modules/artifacts",
		"internal/modules/assessments",
		"internal/modules/entities/hostidentity",
		"internal/modules/evidence",
		"internal/modules/indicators",
		"internal/modules/parties",
		"internal/modules/tasksdecisions",
		"internal/modules/timeline",
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		content := string(data)
		for _, forbidden := range forbiddenImports {
			if strings.Contains(content, forbidden) {
				t.Fatalf("%s imports peer owner %q", file, forbidden)
			}
		}
	}
}

func TestImportsResponsibilitiesRemainSeparated(t *testing.T) {
	assertAggregateFinalizerRetired(t)

	if _, err := os.Stat("routes.go"); !os.IsNotExist(err) {
		t.Fatalf("routes.go must remain decomposed, stat error: %v", err)
	}

	files := []struct {
		name      string
		required  []string
		forbidden []string
	}{
		{
			name:      "service.go",
			required:  []string{"type Service struct", "func RegisterRoutes(", "func newService("},
			forbidden: []string{"handleImportSessionsCollection", "executeApplyJob", "discoverImportUnits"},
		},
		{
			name:      "http_handlers.go",
			required:  []string{"handleImportSessionsCollection", "handleImportSessionsMember", "handleRegion", "writeAPIError"},
			forbidden: []string{"RegisterHandler(", "parseXLSXTables(", "applyGenericOwnerUnit(", "prepareApprovedMapping(", "s.store.", "s.jobRunner.", "indexXLSXWorkbook(", "discoverImportUnits("},
		},
		{
			name:      "application_operations.go",
			required:  []string{"applicationCreateSession", "applicationGetSession", "applicationPrepareMappingPreview", "applicationExecuteUnitAction", "applicationStartApply", "applicationCreateOperatorRegion"},
			forbidden: []string{"http.ResponseWriter", "*http.Request", "http.HandlerFunc", "BindOwnerRoutes("},
		},
		{
			name:      "mapping.go",
			required:  []string{"prepareApprovedMapping", "validateApprovedMapping", "extensionFacadeAPIError"},
			forbidden: []string{"http.ResponseWriter", "RegisterHandler(", "parseXLSXTables("},
		},
		{
			name:      "jobs.go",
			required:  []string{"registerJobHandlers", "executeDiscoveryJob", "prepareClaimedJob"},
			forbidden: []string{"handleImportSessionsCollection", "discoverImportUnits"},
		},
		{
			name:      "apply_jobs.go",
			required:  []string{"executeApplyJob", "completeApplyJob", "finalizeApplyJob", "importUnitFailure"},
			forbidden: []string{"http.ResponseWriter", "parseXLSXTables("},
		},
		{
			name:      "apply_coordination.go",
			required:  []string{"applyUnit", "sourceRowCellsByOrdinal"},
			forbidden: []string{"http.ResponseWriter", "RegisterHandler(", "transformImportValue", "split_delimited_v1"},
		},
		{
			name:      "unit_outcomes.go",
			required:  []string{"lockApplyUnitTx", "insertAppliedUnitOutcomeTx", "finalizeApplyFromOutcomesTx"},
			forbidden: []string{"http.ResponseWriter", "RegisterHandler(", "CreateImportRowTx"},
		},
		{
			name:      "discovery.go",
			required:  []string{"discoverImportUnits", "discoveredImportUnit", "detectSourceFileKind"},
			forbidden: []string{"http.ResponseWriter", "jobManager"},
		},
		{
			name:      "store.go",
			required:  []string{"type Store struct", "func NewStore(", "CreateAcceptedSession"},
			forbidden: []string{"http.ResponseWriter", "RegisterHandler("},
		},
		{
			name:      "selection.go",
			required:  []string{"validateProposedSelectionDoesNotOverlapTx", "validateSelectedUnitsDoNotOverlap", "statusAfterSelection"},
			forbidden: []string{"http.ResponseWriter", "RegisterHandler(", "CreateImportRowTx"},
		},
		{
			name:      "xlsx.go",
			required:  []string{"indexXLSXWorkbook", "decodeRectangle", "archiveCompressionRatioExceeded"},
			forbidden: []string{"http.ResponseWriter", "jobManager"},
		},
		{
			name:      "regions.go",
			required:  []string{"CreateOperatorRegion", "operatorRegionWithinLimits"},
			forbidden: []string{"handleRegion", "http.ResponseWriter", "RegisterHandler(", "applyGenericOwnerUnit("},
		},
	}
	for _, file := range files {
		t.Run(file.name, func(t *testing.T) {
			body, err := os.ReadFile(filepath.Clean(file.name))
			if err != nil {
				t.Fatalf("read %s: %v", file.name, err)
			}
			content := string(body)
			for _, required := range file.required {
				if !strings.Contains(content, required) {
					t.Fatalf("%s is missing responsibility marker %q", file.name, required)
				}
			}
			for _, forbidden := range file.forbidden {
				if strings.Contains(content, forbidden) {
					t.Fatalf("%s contains foreign responsibility marker %q", file.name, forbidden)
				}
			}
		})
	}
}

func assertAggregateFinalizerRetired(t testing.TB) {
	t.Helper()
	retiredPath := filepath.Join("ownerfacade", "finalize.go")
	if _, err := os.Stat(retiredPath); !os.IsNotExist(err) {
		t.Fatalf("aggregate finalizer source remains: %s (%v)", retiredPath, err)
	}
	forbidden := []string{
		"Finalize" + "Command",
		"FinalizeLive" + "RecordTx",
		"FinalizeHistorical" + "RecordTx",
		"LiveRecordRevision" + "Appender",
		"HistoricalRecordRevision" + "Appender",
		"AppendRecordRevision" + "AndIntentTx",
		"FinalizeRecord" + "AndPublish",
	}
	if err := filepath.WalkDir("..", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		body, readErr := os.ReadFile(filepath.Clean(path))
		if readErr != nil {
			return readErr
		}
		for _, token := range forbidden {
			if strings.Contains(string(body), token) {
				t.Fatalf("%s reintroduces aggregate finalizer symbol %q", filepath.ToSlash(path), token)
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("scan source-owner modules for retired aggregate finalizer: %v", err)
	}
}
