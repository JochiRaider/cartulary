package routetest

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/testutil/routeinventory"
)

func PublicPreferences() []routeinventory.Entry {
	return preferenceRoutes("")
}

func ControlPreferences() []routeinventory.Entry {
	return preferenceRoutes("control")
}

func preferenceRoutes(posture string) []routeinventory.Entry {
	membershipRole := routeinventory.ControlRoleTier("")
	adminRole := routeinventory.ControlRoleTier("")
	if posture == "control" {
		membershipRole = routeinventory.ControlRoleMembershipRequired
		adminRole = routeinventory.ControlRoleAdminOnly
	}
	return []routeinventory.Entry{
		{Name: "default workbook preferences", Transport: routeinventory.TransportHTTP, Method: http.MethodGet, Template: "/api/v1/incidents/{incident_id}/workbook-preferences/default", SuccessStatus: http.StatusOK, SuccessEnvelope: true, AllowedRole: membershipRole},
		{
			Name: "default workbook preferences update", Transport: routeinventory.TransportHTTP, Method: http.MethodPut, Template: "/api/v1/incidents/{incident_id}/workbook-preferences/default",
			SuccessStatus: http.StatusOK, SuccessEnvelope: true, RequiresCSRF: true, AllowedRole: adminRole,
			Body: func(routeinventory.Fixture) map[string]any {
				return map[string]any{"default_sheet_ref": map[string]any{"kind": "view_schema", "id": timeline.TimelineViewSchemaID}}
			},
		},
		{Name: "user workbook preferences", Transport: routeinventory.TransportHTTP, Method: http.MethodGet, Template: "/api/v1/incidents/{incident_id}/workbook-preferences/me", SuccessStatus: http.StatusOK, SuccessEnvelope: true, AllowedRole: membershipRole},
		{
			Name: "user workbook preferences update", Transport: routeinventory.TransportHTTP, Method: http.MethodPut, Template: "/api/v1/incidents/{incident_id}/workbook-preferences/me",
			SuccessStatus: http.StatusOK, SuccessEnvelope: true, RequiresCSRF: true, AllowedRole: membershipRole,
			Body: func(routeinventory.Fixture) map[string]any {
				return map[string]any{"home_sheet_ref": map[string]any{"kind": "view_schema", "id": timeline.TimelineViewSchemaID}}
			},
		},
	}
}
