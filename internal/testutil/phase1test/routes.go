package phase1test

import "github.com/JochiRaider/cartulary/internal/testutil/phase1routes"

type RouteTransport = phase1routes.RouteTransport

const (
	RouteTransportHTTP      = phase1routes.RouteTransportHTTP
	RouteTransportWebSocket = phase1routes.RouteTransportWebSocket
)

type RouteCheck = phase1routes.RouteCheck

const (
	RouteCheckBootstrapBoundary = phase1routes.RouteCheckBootstrapBoundary
	RouteCheckCSRF              = phase1routes.RouteCheckCSRF
	RouteCheckEnvelope          = phase1routes.RouteCheckEnvelope
	RouteCheckMutationAudit     = phase1routes.RouteCheckMutationAudit
	RouteCheckReplay            = phase1routes.RouteCheckReplay
	RouteCheckSecretSafePayload = phase1routes.RouteCheckSecretSafePayload
	RouteCheckSessionRevocation = phase1routes.RouteCheckSessionRevocation
)

type RouteID = phase1routes.RouteID

const (
	RouteLogin              = phase1routes.RouteLogin
	RouteSession            = phase1routes.RouteSession
	RouteLogout             = phase1routes.RouteLogout
	RouteCredentialState    = phase1routes.RouteCredentialState
	RoutePasswordChange     = phase1routes.RoutePasswordChange
	RouteTOTPBegin          = phase1routes.RouteTOTPBegin
	RouteTOTPComplete       = phase1routes.RouteTOTPComplete
	RouteUsersList          = phase1routes.RouteUsersList
	RouteUsersCreate        = phase1routes.RouteUsersCreate
	RouteUsersGet           = phase1routes.RouteUsersGet
	RouteUsersPatch         = phase1routes.RouteUsersPatch
	RouteUsersPasswordReset = phase1routes.RouteUsersPasswordReset
	RouteUsersTOTPReset     = phase1routes.RouteUsersTOTPReset
	RouteUsersRevokeAll     = phase1routes.RouteUsersRevokeAll
	RouteSessionLifecycleWS = phase1routes.RouteSessionLifecycleWS
)

type RouteInventoryFixture = phase1routes.RouteInventoryFixture
type RouteInventoryEntry = phase1routes.RouteInventoryEntry

func BuildRoutePath(template string, fixture RouteInventoryFixture) string {
	return phase1routes.BuildRoutePath(template, fixture)
}

func PublicRouteInventory() []RouteInventoryEntry {
	return phase1routes.PublicRouteInventory()
}

func HasRouteCheck(route RouteInventoryEntry, check RouteCheck) bool {
	return phase1routes.HasRouteCheck(route, check)
}
