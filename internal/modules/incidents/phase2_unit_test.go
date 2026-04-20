package incidents

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestPhase2_U_2_02_IncidentCreateBootstrapPlanIncludesCreatorAdminAndWorkbookPreferences(t *testing.T) {
	bootstrap := DefaultIncidentCreateBootstrap()
	if bootstrap.CreatorRole != "admin" {
		t.Fatalf("unexpected bootstrap creator role: got %q want admin", bootstrap.CreatorRole)
	}
	if !bootstrap.CreatesIncidentWorkbookPreferences {
		t.Fatal("incident create bootstrap must plan incident workbook preferences")
	}
	if !bootstrap.CreatesUserWorkbookPreferences {
		t.Fatal("incident create bootstrap must plan user workbook preferences")
	}
}

func TestPhase2_U_2_03_IncidentCreateLocationIsRootedAtIncidentMember(t *testing.T) {
	incidentID := uuid.MustParse("00000000-0000-0000-0000-000000000203")
	if got := incidentLocation(incidentID); got != "/api/v1/incidents/"+incidentID.String() {
		t.Fatalf("unexpected incident Location header root: got %q", got)
	}
}

func TestPhase2_U_2_04_IncidentCreateIdempotencyScopesByActorAndNormalizedRequest(t *testing.T) {
	firstRequest, apiErr := DecodeIncidentCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-04",
		"incident_key":"  IR-U204  ",
		"title":"  Replay Incident  "
	}`))
	if apiErr != nil {
		t.Fatalf("decode first create request: %v", apiErr)
	}
	replayedRequest, apiErr := DecodeIncidentCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-04",
		"incident_key":"IR-U204",
		"title":"Replay Incident"
	}`))
	if apiErr != nil {
		t.Fatalf("decode replayed create request: %v", apiErr)
	}
	if got := IncidentCreateRequestHash(firstRequest); !hashesEqual(got, IncidentCreateRequestHash(replayedRequest)) {
		t.Fatal("normalized incident create replay must reuse the original request hash")
	}

	divergentRequest, apiErr := DecodeIncidentCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-04",
		"incident_key":"IR-U204",
		"title":"Different title"
	}`))
	if apiErr != nil {
		t.Fatalf("decode divergent create request: %v", apiErr)
	}
	if hashesEqual(IncidentCreateRequestHash(firstRequest), IncidentCreateRequestHash(divergentRequest)) {
		t.Fatal("divergent replay must not reuse the original request hash")
	}

	firstActorScope := IncidentCreateIdempotencyScope(uuid.MustParse("00000000-0000-0000-0000-000000000204"))
	secondActorScope := IncidentCreateIdempotencyScope(uuid.MustParse("00000000-0000-0000-0000-000000000205"))
	if firstActorScope == secondActorScope {
		t.Fatal("incident create idempotency must scope by actor")
	}
}

func TestPhase2_U_2_08_DeploymentAdminWithoutMembershipDoesNotGainIncidentAccess(t *testing.T) {
	requireErrorContract(t, "incident_not_found", http.StatusNotFound)
	readDenied := IncidentAccessError(nil, true)
	if readDenied == nil || readDenied.Status != http.StatusNotFound || readDenied.Code != "incident_not_found" {
		t.Fatalf("deployment admin without membership must not gain incident read access: %#v", readDenied)
	}

	writeDenied := IncidentAccessError(nil, true, "reviewer", "admin")
	if writeDenied == nil || writeDenied.Status != http.StatusNotFound || writeDenied.Code != "incident_not_found" {
		t.Fatalf("deployment admin without membership must not gain incident write access: %#v", writeDenied)
	}

	membership := &MembershipRecord{Role: "viewer"}
	guard := IncidentAccessError(membership, true, "reviewer", "admin")
	if guard == nil || guard.Status != http.StatusForbidden || guard.Code != "authorization_denied" {
		t.Fatalf("deployment admin membership checks must still enforce incident role gates: %#v", guard)
	}
}
