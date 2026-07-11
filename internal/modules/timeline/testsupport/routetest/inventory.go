package routetest

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/testutil/routeinventory"
)

func ControlQuery() []routeinventory.Entry {
	return []routeinventory.Entry{{
		Name: "timeline query", Transport: routeinventory.TransportHTTP, Method: http.MethodPost,
		Template:      "/api/v1/incidents/{incident_id}/views/" + timeline.TimelineViewSchemaID + "/query",
		SuccessStatus: http.StatusOK, SuccessEnvelope: true, AllowedRole: routeinventory.ControlRoleMembershipRequired,
		Body: func(routeinventory.Fixture) map[string]any { return map[string]any{} },
	}}
}

func ControlCreateAndLive() []routeinventory.Entry {
	return []routeinventory.Entry{
		{
			Name: "timeline create", Transport: routeinventory.TransportHTTP, Method: http.MethodPost,
			Template:      "/api/v1/incidents/{incident_id}/views/" + timeline.TimelineViewSchemaID + "/rows",
			SuccessStatus: http.StatusCreated, SuccessEnvelope: true, RequiresCSRF: true, AllowedRole: routeinventory.ControlRoleEditorOrHigher,
			Body: func(fixture routeinventory.Fixture) map[string]any {
				return map[string]any{"client_txn_id": "txn-phase2-control-timeline-create-" + fixture.ClientTxnSuffix, "timeline.activity_synopsis_text": "Control boundary row " + fixture.ClientTxnSuffix}
			},
		},
		{Name: "timeline websocket", Transport: routeinventory.TransportWebSocket, Method: http.MethodGet, Template: "/ws/v1/incidents/{incident_id}", AllowedRole: routeinventory.ControlRoleMembershipRequired},
	}
}
