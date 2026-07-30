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
				return map[string]any{"client_txn_id": "txn-incident-membership-control-timeline-create-" + fixture.ClientTxnSuffix, "timeline.activity_synopsis_text": "Control boundary row " + fixture.ClientTxnSuffix}
			},
		},
		{Name: "timeline websocket", Transport: routeinventory.TransportWebSocket, Method: http.MethodGet, Template: "/ws/v1/incidents/{incident_id}", AllowedRole: routeinventory.ControlRoleMembershipRequired},
	}
}

func ControlMutations() []routeinventory.Entry {
	return []routeinventory.Entry{
		{
			Name: "record patch", Transport: routeinventory.TransportHTTP, Method: http.MethodPatch,
			Template: "/api/v1/records/{record_id}", SuccessStatus: http.StatusOK,
			SuccessEnvelope: true, RequiresCSRF: true, AllowedRole: routeinventory.ControlRoleEditorOrHigher,
			Body: func(fixture routeinventory.Fixture) map[string]any {
				return map[string]any{
					"view_schema_id": timeline.TimelineViewSchemaID, "base_row_version": fixture.BaseRecordVersion,
					"client_txn_id": "txn-incident-membership-control-record-patch-" + fixture.ClientTxnSuffix,
					"changes": []map[string]any{{
						"field_key": "timeline.activity_synopsis_text",
						"value":     "Control boundary patch " + fixture.ClientTxnSuffix,
					}},
				}
			},
		},
		{
			Name: "mark reviewed", Transport: routeinventory.TransportHTTP, Method: http.MethodPost,
			Template: "/api/v1/records/{record_id}/mark-reviewed", SuccessStatus: http.StatusOK,
			SuccessEnvelope: true, RequiresCSRF: true, AllowedRole: routeinventory.ControlRoleReviewerOrHigher,
			Body: func(fixture routeinventory.Fixture) map[string]any {
				return map[string]any{
					"base_row_version": fixture.BaseRecordVersion,
					"client_txn_id":    "txn-incident-membership-control-mark-reviewed-" + fixture.ClientTxnSuffix,
				}
			},
		},
		{
			Name: "supersede", Transport: routeinventory.TransportHTTP, Method: http.MethodPost,
			Template: "/api/v1/records/{record_id}/supersede", SuccessStatus: http.StatusOK,
			SuccessEnvelope: true, RequiresCSRF: true, AllowedRole: routeinventory.ControlRoleReviewerOrHigher,
			Body: func(fixture routeinventory.Fixture) map[string]any {
				return map[string]any{
					"base_row_version":      fixture.BaseRecordVersion,
					"client_txn_id":         "txn-incident-membership-control-supersede-" + fixture.ClientTxnSuffix,
					"reason":                "Control boundary supersede " + fixture.ClientTxnSuffix,
					"replacement_record_id": fixture.ReplacementRecordID,
				}
			},
		},
	}
}
