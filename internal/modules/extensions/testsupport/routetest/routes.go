package routetest

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/testutil/routeinventory"
)

func PublicDiscovery() []routeinventory.Entry {
	return []routeinventory.Entry{{
		Name: "extensions list", Transport: routeinventory.TransportHTTP, Method: http.MethodGet,
		Template: "/api/v1/extensions", SuccessStatus: http.StatusOK, SuccessEnvelope: true,
	}}
}
