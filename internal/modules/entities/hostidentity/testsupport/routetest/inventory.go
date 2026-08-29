package routetest

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
	"github.com/JochiRaider/cartulary/internal/testutil/routeinventory"
)

func ControlQueries() []routeinventory.Entry {
	return []routeinventory.Entry{
		query("hosts query", entitycontract.HostsViewSchemaID),
		query("identities query", entitycontract.IdentitiesViewSchemaID),
	}
}

func query(name string, schemaID string) routeinventory.Entry {
	return routeinventory.Entry{
		Name: name, Transport: routeinventory.TransportHTTP, Method: http.MethodPost,
		Template:      "/api/v1/incidents/{incident_id}/views/" + schemaID + "/query",
		SuccessStatus: http.StatusOK, SuccessEnvelope: true, AllowedRole: routeinventory.ControlRoleMembershipRequired,
		Body: func(routeinventory.Fixture) map[string]any { return map[string]any{} },
	}
}
