package phase4test

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/testutil/golden"
)

type RouteSurface struct {
	Name   string
	Method string
	Path   string
	Body   any
}

func OwnedRouteSurfaceInventory(incidentID string, timelineRecordID string, mentionID string, mergeSurvivorRecordID string) []RouteSurface {
	return []RouteSurface{
		{
			Name:   "mention resolve route",
			Method: http.MethodPost,
			Path:   "/api/v1/entity-mentions/" + mentionID + "/resolve",
		},
		{
			Name:   "explicit merge route",
			Method: http.MethodPost,
			Path:   "/api/v1/records/" + mergeSurvivorRecordID + "/merge",
		},
		{
			Name:   "hosts create route",
			Method: http.MethodPost,
			Path:   "/api/v1/incidents/" + incidentID + "/views/" + golden.Phase4HostsViewSchemaID + "/rows",
		},
		{
			Name:   "hosts query route",
			Method: http.MethodPost,
			Path:   "/api/v1/incidents/" + incidentID + "/views/" + golden.Phase4HostsViewSchemaID + "/query",
			Body:   map[string]any{},
		},
		{
			Name:   "identities create route",
			Method: http.MethodPost,
			Path:   "/api/v1/incidents/" + incidentID + "/views/" + golden.Phase4IdentitiesViewSchemaID + "/rows",
		},
		{
			Name:   "identities query route",
			Method: http.MethodPost,
			Path:   "/api/v1/incidents/" + incidentID + "/views/" + golden.Phase4IdentitiesViewSchemaID + "/query",
			Body:   map[string]any{},
		},
		{
			Name:   "indicators create route",
			Method: http.MethodPost,
			Path:   "/api/v1/incidents/" + incidentID + "/views/" + golden.Phase4IndicatorsViewSchemaID + "/rows",
		},
		{
			Name:   "indicators query route",
			Method: http.MethodPost,
			Path:   "/api/v1/incidents/" + incidentID + "/views/" + golden.Phase4IndicatorsViewSchemaID + "/query",
			Body:   map[string]any{},
		},
		{
			Name:   "timeline query route",
			Method: http.MethodPost,
			Path:   "/api/v1/incidents/" + incidentID + "/views/" + golden.Phase4TimelineViewSchemaID + "/query",
			Body:   map[string]any{},
		},
		{
			Name:   "timeline patch route",
			Method: http.MethodPatch,
			Path:   "/api/v1/records/" + timelineRecordID,
		},
	}
}
