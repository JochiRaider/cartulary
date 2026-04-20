package phase2test

import (
	"net/http"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
)

type RouteTransport string

const (
	RouteTransportHTTP      RouteTransport = "http"
	RouteTransportWebSocket RouteTransport = "websocket"
)

type RouteCheck string

const (
	RouteCheckAuthorization    RouteCheck = "authorization"
	RouteCheckEnvelope         RouteCheck = "envelope"
	RouteCheckMutationAudit    RouteCheck = "mutation_audit"
	RouteCheckReservedDispatch RouteCheck = "reserved_dispatch"
)

type RouteInventoryFixture struct {
	IncidentID           string
	AdminUserID          string
	PrimaryRecordID      string
	ReplacementRecordID  string
	BaseRecordVersion    int64
	BaseMembershipVersion int64
}

type RouteInventoryEntry struct {
	Name         string
	Transport    RouteTransport
	Method       string
	Template     string
	RequiresCSRF bool
	Checks       []RouteCheck
	Body         func(RouteInventoryFixture) map[string]any
}

func BuildRoutePath(template string, fixture RouteInventoryFixture) string {
	replacer := strings.NewReplacer(
		"{incident_id}", fixture.IncidentID,
		"{admin_user_id}", fixture.AdminUserID,
		"{record_id}", fixture.PrimaryRecordID,
		"{replacement_record_id}", fixture.ReplacementRecordID,
	)
	return replacer.Replace(template)
}

func DeploymentAdminBoundaryInventory() []RouteInventoryEntry {
	return []RouteInventoryEntry{
		{
			Name:      "incident get",
			Transport: RouteTransportHTTP,
			Method:    http.MethodGet,
			Template:  "/api/v1/incidents/{incident_id}",
			Checks:    []RouteCheck{RouteCheckAuthorization, RouteCheckEnvelope},
		},
		{
			Name:         "incident patch",
			Transport:    RouteTransportHTTP,
			Method:       http.MethodPatch,
			Template:     "/api/v1/incidents/{incident_id}",
			RequiresCSRF: true,
			Checks:       []RouteCheck{RouteCheckAuthorization, RouteCheckEnvelope},
			Body: func(RouteInventoryFixture) map[string]any {
				return map[string]any{
					"base_incident_version": 1,
					"tlp":                   "amber",
				}
			},
		},
		{
			Name:      "memberships list",
			Transport: RouteTransportHTTP,
			Method:    http.MethodGet,
			Template:  "/api/v1/incidents/{incident_id}/memberships",
			Checks:    []RouteCheck{RouteCheckAuthorization, RouteCheckEnvelope},
		},
		{
			Name:         "membership create",
			Transport:    RouteTransportHTTP,
			Method:       http.MethodPost,
			Template:     "/api/v1/incidents/{incident_id}/memberships",
			RequiresCSRF: true,
			Checks:       []RouteCheck{RouteCheckAuthorization, RouteCheckEnvelope},
			Body: func(fixture RouteInventoryFixture) map[string]any {
				return map[string]any{
					"client_txn_id": "txn-support-phase2-boundary-membership-create",
					"user_id":       fixture.AdminUserID,
					"role":          "viewer",
				}
			},
		},
		{
			Name:         "membership patch",
			Transport:    RouteTransportHTTP,
			Method:       http.MethodPatch,
			Template:     "/api/v1/incidents/{incident_id}/memberships/{admin_user_id}",
			RequiresCSRF: true,
			Checks:       []RouteCheck{RouteCheckAuthorization, RouteCheckEnvelope},
			Body: func(fixture RouteInventoryFixture) map[string]any {
				return map[string]any{
					"base_membership_version": fixture.BaseMembershipVersion,
					"role":                    "viewer",
				}
			},
		},
		{
			Name:         "membership delete",
			Transport:    RouteTransportHTTP,
			Method:       http.MethodDelete,
			Template:     "/api/v1/incidents/{incident_id}/memberships/{admin_user_id}",
			RequiresCSRF: true,
			Checks:       []RouteCheck{RouteCheckAuthorization, RouteCheckEnvelope},
			Body: func(fixture RouteInventoryFixture) map[string]any {
				return map[string]any{
					"base_membership_version": fixture.BaseMembershipVersion,
				}
			},
		},
		{
			Name:      "default workbook preferences",
			Transport: RouteTransportHTTP,
			Method:    http.MethodGet,
			Template:  "/api/v1/incidents/{incident_id}/workbook-preferences/default",
			Checks:    []RouteCheck{RouteCheckAuthorization, RouteCheckEnvelope},
		},
		{
			Name:      "user workbook preferences",
			Transport: RouteTransportHTTP,
			Method:    http.MethodGet,
			Template:  "/api/v1/incidents/{incident_id}/workbook-preferences/me",
			Checks:    []RouteCheck{RouteCheckAuthorization, RouteCheckEnvelope},
		},
		{
			Name:      "timeline query",
			Transport: RouteTransportHTTP,
			Method:    http.MethodPost,
			Template:  "/api/v1/incidents/{incident_id}/views/" + timeline.TimelineViewSchemaID + "/query",
			Checks:    []RouteCheck{RouteCheckAuthorization, RouteCheckEnvelope},
			Body: func(RouteInventoryFixture) map[string]any {
				return map[string]any{}
			},
		},
		{
			Name:      "hosts query",
			Transport: RouteTransportHTTP,
			Method:    http.MethodPost,
			Template:  "/api/v1/incidents/{incident_id}/views/" + entities.HostsViewSchemaID + "/query",
			Checks:    []RouteCheck{RouteCheckAuthorization, RouteCheckEnvelope},
			Body: func(RouteInventoryFixture) map[string]any {
				return map[string]any{}
			},
		},
		{
			Name:      "identities query",
			Transport: RouteTransportHTTP,
			Method:    http.MethodPost,
			Template:  "/api/v1/incidents/{incident_id}/views/" + entities.IdentitiesViewSchemaID + "/query",
			Checks:    []RouteCheck{RouteCheckAuthorization, RouteCheckEnvelope},
			Body: func(RouteInventoryFixture) map[string]any {
				return map[string]any{}
			},
		},
		{
			Name:      "indicators query",
			Transport: RouteTransportHTTP,
			Method:    http.MethodPost,
			Template:  "/api/v1/incidents/{incident_id}/views/" + entities.IndicatorsViewSchemaID + "/query",
			Checks:    []RouteCheck{RouteCheckAuthorization, RouteCheckEnvelope},
			Body: func(RouteInventoryFixture) map[string]any {
				return map[string]any{}
			},
		},
		{
			Name:         "timeline create",
			Transport:    RouteTransportHTTP,
			Method:       http.MethodPost,
			Template:     "/api/v1/incidents/{incident_id}/views/" + timeline.TimelineViewSchemaID + "/rows",
			RequiresCSRF: true,
			Checks:       []RouteCheck{RouteCheckAuthorization, RouteCheckEnvelope},
			Body: func(RouteInventoryFixture) map[string]any {
				return map[string]any{
					"client_txn_id":    "txn-support-phase2-boundary-row-create",
					"timeline.summary": "Deployment admin denied",
				}
			},
		},
		{
			Name:      "timeline websocket",
			Transport: RouteTransportWebSocket,
			Method:    http.MethodGet,
			Template:  "/ws/v1/incidents/{incident_id}/views/" + timeline.TimelineViewSchemaID + "/changes",
			Checks:    []RouteCheck{RouteCheckAuthorization},
		},
		{
			Name:         "record patch",
			Transport:    RouteTransportHTTP,
			Method:       http.MethodPatch,
			Template:     "/api/v1/records/{record_id}",
			RequiresCSRF: true,
			Checks:       []RouteCheck{RouteCheckAuthorization, RouteCheckEnvelope},
			Body: func(fixture RouteInventoryFixture) map[string]any {
				return map[string]any{
					"view_schema_id":   timeline.TimelineViewSchemaID,
					"base_row_version": fixture.BaseRecordVersion,
					"client_txn_id":    "txn-support-phase2-boundary-record-patch",
					"changes": []map[string]any{
						{
							"field_key": "timeline.summary",
							"value":     "Denied patch",
						},
					},
				}
			},
		},
		{
			Name:         "mark reviewed",
			Transport:    RouteTransportHTTP,
			Method:       http.MethodPost,
			Template:     "/api/v1/records/{record_id}/mark-reviewed",
			RequiresCSRF: true,
			Checks:       []RouteCheck{RouteCheckAuthorization, RouteCheckEnvelope},
			Body: func(fixture RouteInventoryFixture) map[string]any {
				return map[string]any{
					"base_row_version": fixture.BaseRecordVersion,
					"client_txn_id":    "txn-support-phase2-boundary-mark-reviewed",
				}
			},
		},
		{
			Name:         "supersede",
			Transport:    RouteTransportHTTP,
			Method:       http.MethodPost,
			Template:     "/api/v1/records/{record_id}/supersede",
			RequiresCSRF: true,
			Checks:       []RouteCheck{RouteCheckAuthorization, RouteCheckEnvelope},
			Body: func(fixture RouteInventoryFixture) map[string]any {
				return map[string]any{
					"base_row_version":      fixture.BaseRecordVersion,
					"client_txn_id":         "txn-support-phase2-boundary-supersede",
					"reason":                "Denied supersede",
					"replacement_record_id": fixture.ReplacementRecordID,
				}
			},
		},
	}
}
