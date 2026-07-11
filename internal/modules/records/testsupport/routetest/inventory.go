package routetest

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/testutil/routeinventory"
)

func ControlMutations() []routeinventory.Entry {
	return []routeinventory.Entry{
		{
			Name: "record patch", Transport: routeinventory.TransportHTTP, Method: http.MethodPatch, Template: "/api/v1/records/{record_id}",
			SuccessStatus: http.StatusOK, SuccessEnvelope: true, RequiresCSRF: true, AllowedRole: routeinventory.ControlRoleEditorOrHigher,
			Body: func(fixture routeinventory.Fixture) map[string]any {
				return map[string]any{
					"view_schema_id": timeline.TimelineViewSchemaID, "base_row_version": fixture.BaseRecordVersion,
					"client_txn_id": "txn-phase2-control-record-patch-" + fixture.ClientTxnSuffix,
					"changes":       []map[string]any{{"field_key": "timeline.activity_synopsis_text", "value": "Control boundary patch " + fixture.ClientTxnSuffix}},
				}
			},
		},
		{
			Name: "mark reviewed", Transport: routeinventory.TransportHTTP, Method: http.MethodPost, Template: "/api/v1/records/{record_id}/mark-reviewed",
			SuccessStatus: http.StatusOK, SuccessEnvelope: true, RequiresCSRF: true, AllowedRole: routeinventory.ControlRoleReviewerOrHigher,
			Body: func(fixture routeinventory.Fixture) map[string]any {
				return map[string]any{"base_row_version": fixture.BaseRecordVersion, "client_txn_id": "txn-phase2-control-mark-reviewed-" + fixture.ClientTxnSuffix}
			},
		},
		{
			Name: "supersede", Transport: routeinventory.TransportHTTP, Method: http.MethodPost, Template: "/api/v1/records/{record_id}/supersede",
			SuccessStatus: http.StatusOK, SuccessEnvelope: true, RequiresCSRF: true, AllowedRole: routeinventory.ControlRoleReviewerOrHigher,
			Body: func(fixture routeinventory.Fixture) map[string]any {
				return map[string]any{
					"base_row_version": fixture.BaseRecordVersion, "client_txn_id": "txn-phase2-control-supersede-" + fixture.ClientTxnSuffix,
					"reason": "Control boundary supersede " + fixture.ClientTxnSuffix, "replacement_record_id": fixture.ReplacementRecordID,
				}
			},
		},
	}
}
