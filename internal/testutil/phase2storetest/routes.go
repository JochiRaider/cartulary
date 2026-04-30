package phase2storetest

import (
	"fmt"
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

type MutationOwner string

const (
	MutationOwnerIncidentResource   MutationOwner = "incident resource mutation"
	MutationOwnerIncidentMembership MutationOwner = "incident membership mutation"
)

type ControlRoleTier string

const (
	ControlRoleMembershipRequired ControlRoleTier = "membership-required"
	ControlRoleEditorOrHigher     ControlRoleTier = "editor|reviewer|admin"
	ControlRoleReviewerOrHigher   ControlRoleTier = "reviewer|admin"
	ControlRoleAdminOnly          ControlRoleTier = "admin"
)

type RouteInventoryFixture struct {
	IncidentID            string
	AdminUserID           string
	CandidateUserID       string
	MemberUserID          string
	PrimaryRecordID       string
	ReplacementRecordID   string
	BaseIncidentVersion   int64
	BaseRecordVersion     int64
	BaseMembershipVersion int64
	ClientTxnSuffix       string
}

type RouteInventoryEntry struct {
	Name            string
	Transport       RouteTransport
	Method          string
	Template        string
	SuccessStatus   int
	SuccessEnvelope bool
	RequiresCSRF    bool
	AllowedRole     ControlRoleTier
	MutationOwners  []MutationOwner
	Body            func(RouteInventoryFixture) map[string]any
}

func BuildRoutePath(template string, fixture RouteInventoryFixture) string {
	replacer := strings.NewReplacer(
		"{incident_id}", fixture.IncidentID,
		"{admin_user_id}", fixture.AdminUserID,
		"{candidate_user_id}", fixture.CandidateUserID,
		"{member_user_id}", fixture.MemberUserID,
		"{record_id}", fixture.PrimaryRecordID,
		"{replacement_record_id}", fixture.ReplacementRecordID,
	)
	return replacer.Replace(template)
}

func PublicRouteInventory() []RouteInventoryEntry {
	return []RouteInventoryEntry{
		{
			Name:            "incident list",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodGet,
			Template:        "/api/v1/incidents",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
		},
		{
			Name:            "incident create",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodPost,
			Template:        "/api/v1/incidents",
			SuccessStatus:   http.StatusCreated,
			SuccessEnvelope: true,
			RequiresCSRF:    true,
			MutationOwners: []MutationOwner{
				MutationOwnerIncidentResource,
				MutationOwnerIncidentMembership,
			},
			Body: func(fixture RouteInventoryFixture) map[string]any {
				return map[string]any{
					"client_txn_id": "txn-phase2-public-incident-create-" + fixture.ClientTxnSuffix,
					"incident_key":  "IR-PUBLIC-" + strings.ToUpper(fixture.ClientTxnSuffix),
					"title":         "Public inventory incident " + fixture.ClientTxnSuffix,
				}
			},
		},
		{
			Name:            "incident get",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodGet,
			Template:        "/api/v1/incidents/{incident_id}",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
		},
		{
			Name:            "incident patch",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodPatch,
			Template:        "/api/v1/incidents/{incident_id}",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
			RequiresCSRF:    true,
			MutationOwners:  []MutationOwner{MutationOwnerIncidentResource},
			Body: func(fixture RouteInventoryFixture) map[string]any {
				return map[string]any{
					"base_incident_version": fixture.BaseIncidentVersion,
					"tlp":                   "amber",
				}
			},
		},
		{
			Name:            "memberships list",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodGet,
			Template:        "/api/v1/incidents/{incident_id}/memberships",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
		},
		{
			Name:            "membership create",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodPost,
			Template:        "/api/v1/incidents/{incident_id}/memberships",
			SuccessStatus:   http.StatusCreated,
			SuccessEnvelope: true,
			RequiresCSRF:    true,
			MutationOwners:  []MutationOwner{MutationOwnerIncidentMembership},
			Body: func(fixture RouteInventoryFixture) map[string]any {
				return map[string]any{
					"client_txn_id": "txn-phase2-public-membership-create-" + fixture.ClientTxnSuffix,
					"user_id":       fixture.CandidateUserID,
					"role":          "viewer",
				}
			},
		},
		{
			Name:            "membership patch",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodPatch,
			Template:        "/api/v1/incidents/{incident_id}/memberships/{member_user_id}",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
			RequiresCSRF:    true,
			MutationOwners:  []MutationOwner{MutationOwnerIncidentMembership},
			Body: func(fixture RouteInventoryFixture) map[string]any {
				return map[string]any{
					"base_membership_version": fixture.BaseMembershipVersion,
					"role":                    "reviewer",
				}
			},
		},
		{
			Name:            "membership delete",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodDelete,
			Template:        "/api/v1/incidents/{incident_id}/memberships/{member_user_id}",
			SuccessStatus:   http.StatusNoContent,
			SuccessEnvelope: false,
			RequiresCSRF:    true,
			MutationOwners:  []MutationOwner{MutationOwnerIncidentMembership},
			Body: func(fixture RouteInventoryFixture) map[string]any {
				return map[string]any{
					"base_membership_version": fixture.BaseMembershipVersion,
				}
			},
		},
		{
			Name:            "default workbook preferences",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodGet,
			Template:        "/api/v1/incidents/{incident_id}/workbook-preferences/default",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
		},
		{
			Name:            "default workbook preferences update",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodPut,
			Template:        "/api/v1/incidents/{incident_id}/workbook-preferences/default",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
			RequiresCSRF:    true,
			Body: func(RouteInventoryFixture) map[string]any {
				return map[string]any{
					"default_sheet_ref": map[string]any{
						"kind": "view_schema",
						"id":   timeline.TimelineViewSchemaID,
					},
				}
			},
		},
		{
			Name:            "user workbook preferences",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodGet,
			Template:        "/api/v1/incidents/{incident_id}/workbook-preferences/me",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
		},
		{
			Name:            "user workbook preferences update",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodPut,
			Template:        "/api/v1/incidents/{incident_id}/workbook-preferences/me",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
			RequiresCSRF:    true,
			Body: func(RouteInventoryFixture) map[string]any {
				return map[string]any{
					"home_sheet_ref": map[string]any{
						"kind": "view_schema",
						"id":   timeline.TimelineViewSchemaID,
					},
				}
			},
		},
		{
			Name:            "extensions list",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodGet,
			Template:        "/api/v1/extensions",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
		},
	}
}

func PublicRouteInventoryIncidentCore() []RouteInventoryEntry {
	return routeInventoryByName(PublicRouteInventory(),
		"incident list",
		"incident create",
		"incident get",
		"incident patch",
	)
}

func PublicRouteInventoryMembershipAdmin() []RouteInventoryEntry {
	return routeInventoryByName(PublicRouteInventory(),
		"memberships list",
		"membership create",
		"membership patch",
		"membership delete",
	)
}

func PublicRouteInventoryWorkbookPreferences() []RouteInventoryEntry {
	return routeInventoryByName(PublicRouteInventory(),
		"default workbook preferences",
		"default workbook preferences update",
		"user workbook preferences",
		"user workbook preferences update",
	)
}

func PublicRouteInventoryExtensionDiscovery() []RouteInventoryEntry {
	return routeInventoryByName(PublicRouteInventory(),
		"extensions list",
	)
}

func ControlBoundaryInventory() []RouteInventoryEntry {
	return []RouteInventoryEntry{
		{
			Name:            "incident get",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodGet,
			Template:        "/api/v1/incidents/{incident_id}",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
			AllowedRole:     ControlRoleMembershipRequired,
		},
		{
			Name:            "incident patch",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodPatch,
			Template:        "/api/v1/incidents/{incident_id}",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
			RequiresCSRF:    true,
			AllowedRole:     ControlRoleReviewerOrHigher,
			Body: func(fixture RouteInventoryFixture) map[string]any {
				return map[string]any{
					"base_incident_version": fixture.BaseIncidentVersion,
					"tlp":                   "amber",
				}
			},
		},
		{
			Name:            "memberships list",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodGet,
			Template:        "/api/v1/incidents/{incident_id}/memberships",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
			AllowedRole:     ControlRoleMembershipRequired,
		},
		{
			Name:            "membership create",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodPost,
			Template:        "/api/v1/incidents/{incident_id}/memberships",
			SuccessStatus:   http.StatusCreated,
			SuccessEnvelope: true,
			RequiresCSRF:    true,
			AllowedRole:     ControlRoleAdminOnly,
			Body: func(fixture RouteInventoryFixture) map[string]any {
				return map[string]any{
					"client_txn_id": "txn-phase2-control-membership-create-" + fixture.ClientTxnSuffix,
					"user_id":       fixture.CandidateUserID,
					"role":          "viewer",
				}
			},
		},
		{
			Name:            "membership patch",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodPatch,
			Template:        "/api/v1/incidents/{incident_id}/memberships/{member_user_id}",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
			RequiresCSRF:    true,
			AllowedRole:     ControlRoleAdminOnly,
			Body: func(fixture RouteInventoryFixture) map[string]any {
				return map[string]any{
					"base_membership_version": fixture.BaseMembershipVersion,
					"role":                    "reviewer",
				}
			},
		},
		{
			Name:            "membership delete",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodDelete,
			Template:        "/api/v1/incidents/{incident_id}/memberships/{member_user_id}",
			SuccessStatus:   http.StatusNoContent,
			SuccessEnvelope: false,
			RequiresCSRF:    true,
			AllowedRole:     ControlRoleAdminOnly,
			Body: func(fixture RouteInventoryFixture) map[string]any {
				return map[string]any{
					"base_membership_version": fixture.BaseMembershipVersion,
				}
			},
		},
		{
			Name:            "default workbook preferences",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodGet,
			Template:        "/api/v1/incidents/{incident_id}/workbook-preferences/default",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
			AllowedRole:     ControlRoleMembershipRequired,
		},
		{
			Name:            "default workbook preferences update",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodPut,
			Template:        "/api/v1/incidents/{incident_id}/workbook-preferences/default",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
			RequiresCSRF:    true,
			AllowedRole:     ControlRoleAdminOnly,
			Body: func(RouteInventoryFixture) map[string]any {
				return map[string]any{
					"default_sheet_ref": map[string]any{
						"kind": "view_schema",
						"id":   timeline.TimelineViewSchemaID,
					},
				}
			},
		},
		{
			Name:            "user workbook preferences",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodGet,
			Template:        "/api/v1/incidents/{incident_id}/workbook-preferences/me",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
			AllowedRole:     ControlRoleMembershipRequired,
		},
		{
			Name:            "user workbook preferences update",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodPut,
			Template:        "/api/v1/incidents/{incident_id}/workbook-preferences/me",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
			RequiresCSRF:    true,
			AllowedRole:     ControlRoleMembershipRequired,
			Body: func(RouteInventoryFixture) map[string]any {
				return map[string]any{
					"home_sheet_ref": map[string]any{
						"kind": "view_schema",
						"id":   timeline.TimelineViewSchemaID,
					},
				}
			},
		},
		{
			Name:            "timeline query",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodPost,
			Template:        "/api/v1/incidents/{incident_id}/views/" + timeline.TimelineViewSchemaID + "/query",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
			AllowedRole:     ControlRoleMembershipRequired,
			Body: func(RouteInventoryFixture) map[string]any {
				return map[string]any{}
			},
		},
		{
			Name:            "hosts query",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodPost,
			Template:        "/api/v1/incidents/{incident_id}/views/" + entities.HostsViewSchemaID + "/query",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
			AllowedRole:     ControlRoleMembershipRequired,
			Body: func(RouteInventoryFixture) map[string]any {
				return map[string]any{}
			},
		},
		{
			Name:            "identities query",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodPost,
			Template:        "/api/v1/incidents/{incident_id}/views/" + entities.IdentitiesViewSchemaID + "/query",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
			AllowedRole:     ControlRoleMembershipRequired,
			Body: func(RouteInventoryFixture) map[string]any {
				return map[string]any{}
			},
		},
		{
			Name:            "indicators query",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodPost,
			Template:        "/api/v1/incidents/{incident_id}/views/" + entities.IndicatorsViewSchemaID + "/query",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
			AllowedRole:     ControlRoleMembershipRequired,
			Body: func(RouteInventoryFixture) map[string]any {
				return map[string]any{}
			},
		},
		{
			Name:            "timeline create",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodPost,
			Template:        "/api/v1/incidents/{incident_id}/views/" + timeline.TimelineViewSchemaID + "/rows",
			SuccessStatus:   http.StatusCreated,
			SuccessEnvelope: true,
			RequiresCSRF:    true,
			AllowedRole:     ControlRoleEditorOrHigher,
			Body: func(fixture RouteInventoryFixture) map[string]any {
				return map[string]any{
					"client_txn_id":    "txn-phase2-control-timeline-create-" + fixture.ClientTxnSuffix,
					"timeline.summary": "Control boundary row " + fixture.ClientTxnSuffix,
				}
			},
		},
		{
			Name:        "timeline websocket",
			Transport:   RouteTransportWebSocket,
			Method:      http.MethodGet,
			Template:    "/ws/v1/incidents/{incident_id}",
			AllowedRole: ControlRoleMembershipRequired,
		},
		{
			Name:            "record patch",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodPatch,
			Template:        "/api/v1/records/{record_id}",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
			RequiresCSRF:    true,
			AllowedRole:     ControlRoleEditorOrHigher,
			Body: func(fixture RouteInventoryFixture) map[string]any {
				return map[string]any{
					"view_schema_id":   timeline.TimelineViewSchemaID,
					"base_row_version": fixture.BaseRecordVersion,
					"client_txn_id":    "txn-phase2-control-record-patch-" + fixture.ClientTxnSuffix,
					"changes": []map[string]any{
						{
							"field_key": "timeline.summary",
							"value":     "Control boundary patch " + fixture.ClientTxnSuffix,
						},
					},
				}
			},
		},
		{
			Name:            "mark reviewed",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodPost,
			Template:        "/api/v1/records/{record_id}/mark-reviewed",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
			RequiresCSRF:    true,
			AllowedRole:     ControlRoleReviewerOrHigher,
			Body: func(fixture RouteInventoryFixture) map[string]any {
				return map[string]any{
					"base_row_version": fixture.BaseRecordVersion,
					"client_txn_id":    "txn-phase2-control-mark-reviewed-" + fixture.ClientTxnSuffix,
				}
			},
		},
		{
			Name:            "supersede",
			Transport:       RouteTransportHTTP,
			Method:          http.MethodPost,
			Template:        "/api/v1/records/{record_id}/supersede",
			SuccessStatus:   http.StatusOK,
			SuccessEnvelope: true,
			RequiresCSRF:    true,
			AllowedRole:     ControlRoleReviewerOrHigher,
			Body: func(fixture RouteInventoryFixture) map[string]any {
				return map[string]any{
					"base_row_version":      fixture.BaseRecordVersion,
					"client_txn_id":         "txn-phase2-control-supersede-" + fixture.ClientTxnSuffix,
					"reason":                "Control boundary supersede " + fixture.ClientTxnSuffix,
					"replacement_record_id": fixture.ReplacementRecordID,
				}
			},
		},
	}
}

func ControlBoundaryInventoryIncidentCore() []RouteInventoryEntry {
	return routeInventoryByName(ControlBoundaryInventory(),
		"incident get",
		"incident patch",
	)
}

func ControlBoundaryInventoryMembershipAdmin() []RouteInventoryEntry {
	return routeInventoryByName(ControlBoundaryInventory(),
		"memberships list",
		"membership create",
		"membership patch",
		"membership delete",
	)
}

func ControlBoundaryInventoryWorkbookPreferences() []RouteInventoryEntry {
	return routeInventoryByName(ControlBoundaryInventory(),
		"default workbook preferences",
		"default workbook preferences update",
		"user workbook preferences",
		"user workbook preferences update",
	)
}

func ControlBoundaryInventoryWorkbookQueries() []RouteInventoryEntry {
	return routeInventoryByName(ControlBoundaryInventory(),
		"timeline query",
		"hosts query",
		"identities query",
		"indicators query",
	)
}

func ControlBoundaryInventoryTimelineRecordAndLive() []RouteInventoryEntry {
	return routeInventoryByName(ControlBoundaryInventory(),
		"timeline create",
		"timeline websocket",
		"record patch",
		"mark reviewed",
		"supersede",
	)
}

func routeInventoryByName(inventory []RouteInventoryEntry, names ...string) []RouteInventoryEntry {
	byName := make(map[string]RouteInventoryEntry, len(inventory))
	for _, route := range inventory {
		byName[route.Name] = route
	}

	routes := make([]RouteInventoryEntry, 0, len(names))
	for _, name := range names {
		route, ok := byName[name]
		if !ok {
			panic(fmt.Sprintf("missing phase2 route inventory entry %q", name))
		}
		routes = append(routes, route)
	}
	return routes
}
