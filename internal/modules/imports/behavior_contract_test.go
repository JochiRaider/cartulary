package imports

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

// TestImportSessionMemberRouteContract requires one handler for each of the 15
// generated Imports operations, each bound for central admission (Core 01
// REQ-01-707). Route resolution, the declared security and the path
// parameters are therefore admitted from the generated catalog, and the six
// state-changing operations of Table 17.2-A are exactly those whose cookie
// alternative pairs CSRF.
func TestImportSessionMemberRouteContract(t *testing.T) {
	t.Parallel()

	operations := httpapi.ContractOperationsForOwner("module.imports")
	handlers := (&service{}).importOperationHandlers()
	if len(operations) != 15 || len(handlers) != 15 {
		t.Fatalf("Imports operation count changed: contracts=%d handlers=%d", len(operations), len(handlers))
	}
	stateChanging := 0
	for _, operation := range operations {
		if handlers[operation.OperationID] == nil {
			t.Fatalf("generated Imports operation %q has no handler", operation.OperationID)
		}
		if operation.StateChanging != operationRequiresCookieCSRF(operation.Security) {
			t.Fatalf("generated operation %q state-changing=%t disagrees with its declared security %v", operation.OperationID, operation.StateChanging, operation.Security)
		}
		if operation.StateChanging {
			stateChanging++
		}
	}
	if stateChanging != 6 {
		t.Fatalf("state-changing Imports operations = %d, want 6", stateChanging)
	}
	registry, err := httpapi.NewRouteRegistry(nil)
	if err != nil {
		t.Fatal(err)
	}
	deps := httpapi.DependencySet{PublicRoutes: registry, Authenticator: httpauth.NewAuthenticator(httpauth.Options{Now: time.Now})}
	if err := bindRoutes(http.NewServeMux(), deps, &service{}); err != nil {
		t.Fatalf("bind Imports routes: %v", err)
	}
	modes := registry.AdmissionModes()
	if len(modes) != 15 {
		t.Fatalf("bound Imports operations = %d, want 15", len(modes))
	}
	for operationID, mode := range modes {
		if mode != httpapi.AdmissionCentral {
			t.Fatalf("Imports operation %q admission mode = %q, want central", operationID, mode)
		}
	}
}

func operationRequiresCookieCSRF(security [][]string) bool {
	for _, alternative := range security {
		hasCookie := false
		hasCSRFCookie := false
		hasCSRFHeader := false
		for _, scheme := range alternative {
			switch scheme {
			case "sessionCookie":
				hasCookie = true
			case "csrfCookie":
				hasCSRFCookie = true
			case "csrfHeader":
				hasCSRFHeader = true
			}
		}
		if hasCookie && hasCSRFCookie && hasCSRFHeader {
			return true
		}
	}
	return false
}

func TestGeneratedImportTargetRegistryContract(t *testing.T) {
	t.Parallel()

	type expectedTarget struct {
		applyStatus string
		facade      string
	}
	expectedViews := map[string]expectedTarget{
		"cartulary.view.timeline.v2":              {applyStatus: applyStatusSupported, facade: "timeline.import_create"},
		"cartulary.view.hosts.v1":                 {applyStatus: applyStatusSupported, facade: "entities.host.import_create"},
		"cartulary.view.identities.v1":            {applyStatus: applyStatusSupported, facade: "entities.identity.import_create"},
		"cartulary.view.indicators.v1":            {applyStatus: applyStatusSupported, facade: "indicators.import_create"},
		"cartulary.view.evidence.v1":              {applyStatus: applyStatusSupported, facade: "evidence.import_create"},
		"cartulary.view.notes.v1":                 {applyStatus: applyStatusSupported, facade: "artifacts.note.import_create"},
		"cartulary.view.assessments.v1":           {applyStatus: applyStatusSupported, facade: "assessments.import_create"},
		"cartulary.view.task_requests.v1":         {applyStatus: applyStatusSupported, facade: "tasksdecisions.task_request.import_create"},
		"cartulary.view.decisions.v1":             {applyStatus: applyStatusSupported, facade: "tasksdecisions.decision.import_create"},
		"cartulary.view.parties.v1":               {applyStatus: applyStatusSupported, facade: "parties.import_create"},
		"cartulary.view.comm_log.v1":              {applyStatus: applyStatusSupported, facade: "artifacts.import_create"},
		"cartulary.view.handoff.v1":               {applyStatus: applyStatusSupported, facade: "artifacts.import_create"},
		"cartulary.view.status_review.v1":         {applyStatus: applyStatusSupported, facade: "artifacts.import_create"},
		"cartulary.view.lesson.v1":                {applyStatus: applyStatusSupported, facade: "artifacts.import_create"},
		"cartulary.view.findings.v1":              {applyStatus: applyStatusSupportedWhenAvailable},
		"cartulary.view.investigative_queries.v1": {applyStatus: applyStatusSupportedWhenAvailable},
		"cartulary.view.forensic_keywords.v1":     {applyStatus: applyStatusSupportedWhenAvailable},
	}
	if len(importTargets) != len(expectedViews) {
		t.Fatalf("current view target count changed: got %d want %d", len(importTargets), len(expectedViews))
	}
	for selector, expected := range expectedViews {
		target, ok := importTargets[selector]
		if !ok {
			t.Fatalf("missing current target %q", selector)
		}
		if target.ViewSchemaID != selector ||
			target.TargetKind != ImportTargetKindViewSchema ||
			target.ApplyStatus != expected.applyStatus ||
			target.CreateFacade != expected.facade {
			t.Fatalf("unexpected current target %q: got %#v want %#v", selector, target, expected)
		}
		if expected.applyStatus == applyStatusSupported && !target.importable(nil) {
			t.Fatalf("supported target %q is not importable", selector)
		}
		if expected.applyStatus == applyStatusSupportedWhenAvailable && target.importable(func(string) bool { return true }) {
			t.Fatalf("reserved target %q became importable", selector)
		}
	}

	key := analyticalImportTargetKey{
		TargetKind:         importTargetKindNetworkFlowTable,
		ExtensionProfileID: networkFlowExtensionProfileID,
	}
	if len(analyticalImportTargets) != 1 {
		t.Fatalf("current analytical target count changed: got %d want 1", len(analyticalImportTargets))
	}
	analytical := analyticalImportTargets[key]
	if analytical.ApplyStatus != applyStatusSupportedWhenClaimed ||
		analytical.ApplyFacade != "network_flow_import_facade_v1" ||
		analytical.importable(nil) ||
		!analytical.importable(func(profileID string) bool {
			return profileID == networkFlowExtensionProfileID
		}) {
		t.Fatalf("unexpected current analytical target: %#v", analytical)
	}
	assertFileMappingCompatibilityContract(t)
}

func TestInternalImportErrorDoesNotEchoRawMessage(t *testing.T) {
	t.Parallel()

	apiErr := internalAPIError(errors.New("sensitive owner detail"))
	if apiErr.Status != http.StatusInternalServerError ||
		apiErr.Code != "internal_error" ||
		apiErr.Message != "The import operation could not be completed." ||
		strings.Contains(apiErr.Message, "sensitive") {
		t.Fatalf("unsafe internal error behavior: %#v", apiErr)
	}
}

func TestXLSXIndexIncludesHiddenSheetWithoutPresentationSemantics(t *testing.T) {
	t.Parallel()

	workbook, apiErr := indexXLSXWorkbook(
		contractWorkbook(t),
		Limits{MaxRows: 100, MaxColumns: 20, MaxCells: 2_000},
		ArchiveLimits{
			DefaultMaxExtractedBytes: 1 << 20,
			MaxCompressionRatio:      100,
			MaxMembers:               20,
		},
	)
	if apiErr != nil {
		t.Fatalf("index contract workbook: %#v", apiErr)
	}
	if len(workbook.ranges) != 2 ||
		workbook.ranges[0].sheet.name != "Visible" ||
		workbook.ranges[1].sheet.name != "Hidden" {
		t.Fatalf("all semantic sheets must be indexed in workbook order, got %#v", workbook.ranges)
	}
	decoded, decodeErr := workbook.decodeRectangle(
		workbook.ranges[1].sheet,
		workbook.ranges[1].rect,
		Limits{MaxRows: 100, MaxColumns: 20, MaxCells: 2_000},
	)
	if decodeErr != nil {
		t.Fatalf("decode hidden range: %#v", decodeErr)
	}
	if len(decoded.rows) != 2 ||
		decoded.rows[1][0].DisplayText != "hidden" ||
		!slices.Contains(decoded.warningCodes, "filtered_or_hidden_state_ignored") {
		t.Fatalf("hidden presentation changed semantic cells or warning: %#v", decoded)
	}
}

func contractWorkbook(t testing.TB) []byte {
	t.Helper()

	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	writeContractZipText(t, writer, "xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="Visible" sheetId="1" r:id="rId1"/>
    <sheet name="Hidden" sheetId="2" state="hidden" r:id="rId2"/>
  </sheets>
</workbook>`)
	writeContractZipText(t, writer, "xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet2.xml"/>
</Relationships>`)
	for index, value := range []string{"visible", "hidden"} {
		writeContractZipText(t, writer, "xl/worksheets/sheet"+string(rune('1'+index))+".xml", `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1"><c r="A1" t="inlineStr"><is><t>header</t></is></c></row>
    <row r="2"><c r="A2" t="inlineStr"><is><t>`+value+`</t></is></c></row>
  </sheetData>
</worksheet>`)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close contract workbook: %v", err)
	}
	return buffer.Bytes()
}

func writeContractZipText(t testing.TB, writer *zip.Writer, name string, content string) {
	t.Helper()

	entry, err := writer.Create(name)
	if err != nil {
		t.Fatalf("create %s: %v", name, err)
	}
	if _, err := io.WriteString(entry, content); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// TestImportRouteBindingsFailClosed requires the registry binding of the
// Imports handlers to fail closed: an unknown operation, an operation bound
// under another owner and a duplicate binding are bind errors, and an
// operation left unbound is reported missing by route parity.
func TestImportRouteBindingsFailClosed(t *testing.T) {
	for _, scenario := range []string{"missing", "duplicate", "unknown", "wrong_owner"} {
		t.Run(scenario, func(t *testing.T) {
			registry, err := httpapi.NewRouteRegistry([]httpapi.ExtensionClaim{{ProfileID: "import", Claimed: true}})
			if err != nil {
				t.Fatal(err)
			}
			deps := httpapi.DependencySet{PublicRoutes: registry, Authenticator: httpauth.NewAuthenticator(httpauth.Options{Now: time.Now})}
			handlers := (&service{}).importOperationHandlers()
			owner := "module.imports"
			switch scenario {
			case "missing":
				delete(handlers, "getImportUnitPreview")
			case "unknown":
				handlers["unknownImportOperation"] = handlers["getImportUnit"]
			case "wrong_owner":
				owner = "module.workbook"
			}
			bindErr := httpapi.BindAdmittedOwnerRoutes(http.NewServeMux(), deps, owner, handlers)
			if scenario == "duplicate" && bindErr == nil {
				bindErr = httpapi.BindAdmittedOwnerRoutes(http.NewServeMux(), deps, owner, handlers)
			}
			if scenario == "missing" {
				if bindErr != nil {
					t.Fatalf("bind remaining Imports operations: %v", bindErr)
				}
				if err := registry.ValidateActive(); err == nil || !strings.Contains(err.Error(), "[getImportUnitPreview]") {
					t.Fatalf("route parity did not report the unbound operation: %v", err)
				}
				return
			}
			if bindErr == nil {
				t.Fatal("invalid binding admitted")
			}
		})
	}
}
