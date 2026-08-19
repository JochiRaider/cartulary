package httpapi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
)

func TestIncidentCreateLocationDerivesFromApplicationResult_Unit(t *testing.T) {
	result := incidents.CreateIncidentResult{
		Incident: incidents.IncidentRecord{
			ID: uuid.MustParse("00000000-0000-0000-0000-000000000203"),
		},
		Created: true,
	}
	if got, want := incidentLocation(result.Incident.ID), "/api/v1/incidents/00000000-0000-0000-0000-000000000203"; got != want {
		t.Fatalf("incident location = %q, want %q", got, want)
	}
}
