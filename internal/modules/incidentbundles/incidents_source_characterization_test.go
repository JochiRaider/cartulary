package incidentbundles_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
)

func TestIncidentsSourcePortV3Characterization_Unit(t *testing.T) {
	t.Parallel()

	incidentID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	actorID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	port, err := incidents.NewIncidentBundleSourcePort()
	if err != nil {
		t.Fatalf("construct Incidents source port: %v", err)
	}
	descriptor := port.Descriptor()

	if descriptor.FamilyID != "incident" || descriptor.ContractMajor != sourceport.ContractMajor ||
		descriptor.OwnerID != "module.incidents" ||
		!slices.Equal(descriptor.OwnerRelationIDs, []string{"incident-core"}) {
		t.Fatalf("incident source descriptor drifted: %#v", descriptor)
	}
	if len(descriptor.Paths) != 1 {
		t.Fatalf("incident source paths = %d, want 1", len(descriptor.Paths))
	}
	path := descriptor.Paths[0]
	if path.LogicalPath != "data/incident.json" || path.ContentRole != "singleton_json" ||
		!slices.Equal(path.Versions, []int{3}) || !slices.Equal(path.StableIdentity, []string{"id"}) ||
		path.StableIdentityInvariantID != "incident.source_identity_admitted" {
		t.Fatalf("incident source path drifted: %#v", path)
	}
	wantInvariants := []string{
		"incident.source_identity_admitted",
		"incident.exact_shape",
		"incident.identity_key_lifecycle",
		"incident.attribution_version",
	}
	for _, invariantID := range wantInvariants {
		if !slices.Contains(descriptor.InvariantIDs, invariantID) {
			t.Fatalf("incident source descriptor does not declare %q: %#v", invariantID, descriptor.InvariantIDs)
		}
	}

	const portableRow = `{"closed_at":null,"created_at":"2026-08-28T12:34:56.123456+00:00","created_by_user_id":"22222222-2222-4222-8222-222222222222","current_phase":"triage","description":"Portable description","id":"11111111-1111-4111-8111-111111111111","incident_key":"IR-PORTABLE-1","incident_key_canonical":"IR-PORTABLE-1","incident_version":7,"primary_external_case_ref":null,"severity":"high","status":"active","title":"Portable incident","tlp":"TLP:AMBER","updated_at":"2026-08-28T12:34:56.123456+00:00","updated_by_user_id":"22222222-2222-4222-8222-222222222222"}`
	wantPayload := []byte(portableRow + "\n")
	files, err := port.Export(context.Background(), sourceport.ExportContext{
		Query:      incidentSourceQueryFake{raw: []byte(portableRow), incidentKey: "IR-PORTABLE-1"},
		IncidentID: incidentID,
	})
	if err != nil {
		t.Fatalf("export incident source: %v", err)
	}
	if len(files) != 1 || files[0].Path != "data/incident.json" || !slices.Equal(files[0].Payload, wantPayload) {
		t.Fatalf("incident source export = %#v, want exact v3 payload %q", files, wantPayload)
	}

	actors, err := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{{SourceActorID: actorID.String()}})
	if err != nil {
		t.Fatalf("construct actor catalog: %v", err)
	}
	operationID := "incident-source-characterization"
	prepared, err := port.PrepareImport(context.Background(), sourceport.MapBundle{
		"data/incident.json": wantPayload,
	}, sourceport.ImportContext{
		IncidentID: incidentID, ActorUserID: actorID, BundleVersion: 3,
		OperationID: operationID, Actors: actors,
	})
	if err != nil {
		t.Fatalf("prepare characterized valid incident source: %v", err)
	}
	if _, err := prepared.ValueFor("module.incidents:incident", operationID); err != nil {
		t.Fatalf("prepared incident source binding: %v", err)
	}
}

type incidentSourceQueryFake struct {
	raw         []byte
	incidentKey string
}

func (incidentSourceQueryFake) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected incident source Query call")
}

func (q incidentSourceQueryFake) QueryRow(context.Context, string, ...any) pgx.Row {
	return incidentSourceRowFake(q)
}

type incidentSourceRowFake struct {
	raw         []byte
	incidentKey string
}

func (r incidentSourceRowFake) Scan(destinations ...any) error {
	if len(destinations) != 2 {
		return errors.New("unexpected incident source scan shape")
	}
	raw, ok := destinations[0].(*[]byte)
	if !ok {
		return errors.New("unexpected incident source JSON destination")
	}
	incidentKey, ok := destinations[1].(*string)
	if !ok {
		return errors.New("unexpected incident source key destination")
	}
	*raw = slices.Clone(r.raw)
	*incidentKey = r.incidentKey
	return nil
}

var _ incidentportability.Queryer = incidentSourceQueryFake{}
var _ pgx.Row = incidentSourceRowFake{}
