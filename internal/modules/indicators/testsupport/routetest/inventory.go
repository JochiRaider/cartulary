package routetest

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/testutil/routeinventory"
)

func ControlQuery() []routeinventory.Entry {
	return []routeinventory.Entry{{
		Name: "indicators query", Transport: routeinventory.TransportHTTP, Method: http.MethodPost,
		Template:      "/api/v1/incidents/{incident_id}/views/" + indicators.ViewSchemaID + "/query",
		SuccessStatus: http.StatusOK, SuccessEnvelope: true, AllowedRole: routeinventory.ControlRoleMembershipRequired,
		Body: func(routeinventory.Fixture) map[string]any { return map[string]any{} },
	}}
}
